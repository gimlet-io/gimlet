package nativeGit

import (
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/go-git/go-git/v5/plumbing/storer"
)

type commitDirIter struct {
	sourceIter    object.CommitIter
	currentCommit *object.Commit
	dir           string
	r             *git.Repository
}

// NewCommitPathIterFromIter returns a commit iterator which performs diffTree between
// successive trees returned from the commit iterator from the argument. The purpose of this is
// to find the commits that explain how the files that match the path came to be.
// If checkParent is true then the function double checks if potential parent (next commit in a path)
// is one of the parents in the tree (it's used by `git log --all`).
// pathFilter is a function that takes path of file as argument and returns true if we want it
func NewCommitDirIterFromIter(dir string, commitIter object.CommitIter, r *git.Repository) object.CommitIter {
	iterator := new(commitDirIter)
	iterator.sourceIter = commitIter
	iterator.dir = dir
	iterator.r = r
	return iterator
}

func (c *commitDirIter) Next() (*object.Commit, error) {
	if c.currentCommit == nil {
		var err error
		c.currentCommit, err = c.sourceIter.Next()
		if err != nil {
			return nil, err
		}
	}
	commit, commitErr := c.getNextFileCommit()

	// Setting current-commit to nil to prevent unwanted states when errors are raised
	if commitErr != nil {
		c.currentCommit = nil
	}
	return commit, commitErr
}

func (c *commitDirIter) getNextFileCommit() (*object.Commit, error) {
	for {
		// Parent-commit can be nil
		// - if the current-commit is the initial commit
		// - if the current-commit is the first after the `since` limit
		parentCommit, parentCommitErr := c.sourceIter.Next()
		if parentCommitErr != nil {
			if parentCommitErr != io.EOF { // real irregularity
				return nil, parentCommitErr
			}

			parentCommit = nil
		}

		// Fetch the trees of the current commit
		currentTree, currTreeErr := c.currentCommit.Tree()
		if currTreeErr != nil {
			return nil, currTreeErr
		}
		limitedTree, limitedTreeErr := currentTree.Tree(c.dir)
		if c.dir == "" {
			limitedTree = currentTree
			limitedTreeErr = nil
		}

		var found bool
		var err error
		if parentCommit != nil {
			found, err = c.hasFileChanges(parentCommit, limitedTreeErr, limitedTree)
			if err != nil {
				return nil, err
			}
		} else {
			// parent commit is nil, but maybe only due to the cutoff by `since` limit
			// that case we peak ahead behind `since` and check the diff
			for _, peakAheadParentHash := range c.currentCommit.ParentHashes {
				peakAheadCommit, err := c.r.CommitObject(peakAheadParentHash)
				if err != nil {
					return nil, err
				}
				found, err = c.hasFileChanges(peakAheadCommit, limitedTreeErr, limitedTree)
				if found {
					break
				}
			}
		}

		// Storing the current-commit in-case a change is found, and
		// Updating the current-commit for the next-iteration
		prevCommit := c.currentCommit
		c.currentCommit = parentCommit

		if found {
			return prevCommit, nil
		}

		// If not matches found and if parent-commit is beyond the initial commit, then return with EOF
		if parentCommit == nil {
			return nil, io.EOF
		}
	}
}

func (c *commitDirIter) hasFileChanges(parentCommit *object.Commit, limitedTreeErr error, limitedTree *object.Tree) (bool, error) {
	// Fetch the trees of the parent commit
	parentTree, parentTreeErr := parentCommit.Tree()
	if parentTreeErr != nil {
		return false, parentTreeErr
	}
	limitedParentTree, limitedParentTreeErr := parentTree.Tree(c.dir)
	if c.dir == "" {
		limitedParentTree = parentTree
		limitedParentTreeErr = nil
	}

	var found bool
	if limitedTreeErr == object.ErrDirectoryNotFound &&
		limitedParentTreeErr == nil {
		// folder was deleted
		found = true
	} else if limitedTreeErr == nil &&
		limitedParentTreeErr == object.ErrDirectoryNotFound {
		// folder was created
		found = true
	} else if limitedTreeErr == nil &&
		limitedParentTreeErr == nil {
		// Find diff between current and parent trees
		changes, diffErr := object.DiffTree(limitedTree, limitedParentTree)
		if diffErr != nil {
			return false, diffErr
		}

		found = len(changes) > 0
	}
	return found, nil
}

func (c *commitDirIter) ForEach(cb func(*object.Commit) error) error {
	for {
		commit, nextErr := c.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return nextErr
		}
		err := cb(commit)
		if err == storer.ErrStop {
			return nil
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (c *commitDirIter) Close() {
	c.sourceIter.Close()
}
