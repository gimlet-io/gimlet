package stack

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/blang/semver/v4"
	"github.com/fluxcd/source-controller/pkg/sourceignore"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/storage/memory"
	giturl "github.com/whilp/git-urls"
)

func GenerateFromStackYaml(stackConfig dx.StackConfig) (map[string]string, error) {
	stackTemplates, err := cloneStackFromRepo(stackConfig.Stack.Repository)
	if err != nil {
		return nil, err
	}

	return generate(stackTemplates, stackConfig.Config)
}

func generate(
	stackTemplate map[string]string,
	values map[string]interface{},
) (map[string]string, error) {
	generatedFiles := map[string]string{}

	for path, fileContent := range stackTemplate {
		if path == "stack-definition.yaml" {
			continue
		}
		templates, err := template.New(path).Funcs(sprig.TxtFuncMap()).Parse(fileContent)
		if err != nil {
			return nil, err
		}

		var templated bytes.Buffer
		err = templates.Execute(&templated, values)
		if err != nil {
			return nil, err
		}

		// filter empty and white space only files
		if len(strings.TrimSpace(templated.String())) != 0 {
			generatedFiles[path] = templated.String()
		}
	}

	return generatedFiles, nil
}

func StackDefinitionFromRepo(repoUrl string) (string, error) {
	stackTemplates, err := cloneStackFromRepo(repoUrl)
	if err != nil {
		return "", err
	}

	return stackTemplates["stack-definition.yaml"], nil
}

// cloneStackFromRepo takes a git repo url, and returns the files of the git reference
// if the repoUrl is a local filesystem location, it loads the files from there
func cloneStackFromRepo(repoURL string) (map[string]string, error) {
	gitAddress, err := giturl.ParseScp(repoURL)
	if err != nil {
		_, err2 := os.Stat(repoURL)
		if err2 != nil {
			return nil, fmt.Errorf("cannot parse stacks's git address: %s", err)
		} else {
			return loadStackFromFS(repoURL)
		}
	}
	gitUrl := strings.ReplaceAll(repoURL, gitAddress.RawQuery, "")
	gitUrl = strings.ReplaceAll(gitUrl, "?", "")

	fs := memfs.New()
	opts := &git.CloneOptions{
		URL: gitUrl,
	}
	repo, err := git.Clone(memory.NewStorage(), fs, opts)
	if err != nil {
		return nil, fmt.Errorf("cannot clone: %s", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("cannot get worktree: %s", err)
	}

	params, _ := url.ParseQuery(gitAddress.RawQuery)
	if v, found := params["sha"]; found {
		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(v[0]),
		})
		if err != nil {
			return nil, fmt.Errorf("cannot checkout sha: %s", err)
		}
	}
	if v, found := params["tag"]; found {
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(v[0]),
		})
		if err != nil {
			return nil, fmt.Errorf("cannot checkout tag: %s", err)
		}
	}
	if v, found := params["branch"]; found {
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewRemoteReferenceName("origin", v[0]),
		})
		if err != nil {
			return nil, fmt.Errorf("cannot checkout branch: %s", err)
		}
	}

	paths, err := util.Glob(worktree.Filesystem, "*/*")
	if err != nil {
		return nil, fmt.Errorf("cannot list files: %s", err)
	}
	paths2, err := util.Glob(worktree.Filesystem, "*")
	if err != nil {
		return nil, fmt.Errorf("cannot list files: %s", err)
	}
	paths = append(paths, paths2...)
	paths3, err := util.Glob(worktree.Filesystem, "*/*/*")
	if err != nil {
		return nil, fmt.Errorf("cannot list files: %s", err)
	}
	paths = append(paths, paths3...)
	paths4, err := util.Glob(worktree.Filesystem, "*/*/*/*")
	if err != nil {
		return nil, fmt.Errorf("cannot list files: %s", err)
	}
	paths = append(paths, paths4...)

	fs = worktree.Filesystem

	var stackIgnorePatterns []gitignore.Pattern
	const stackIgnoreFile = ".stackignore"
	_, err = fs.Stat(stackIgnoreFile)
	if err == nil {
		if f, err := fs.Open(stackIgnoreFile); err == nil {
			defer f.Close()
			stackIgnorePatterns = sourceignore.ReadPatterns(f, []string{})
		} else {
			return nil, fmt.Errorf("cannot read .stackignore file: %s", err)
		}
	}

	ignore := gitignore.NewMatcher(stackIgnorePatterns)

	files := map[string]string{}
	for _, path := range paths {
		info, err := fs.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("cannot get file: %s", err)
		}

		if info.IsDir() {
			continue
		}

		if ignore.Match(strings.Split(path, "/"), false) {
			continue
		}

		f, err := fs.Open(path)
		if err != nil {
			return nil, fmt.Errorf("cannot get file: %s", err)
		}
		defer f.Close()

		content, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("cannot get file: %s", err)
		}

		files[path] = string(content)
	}

	return files, nil
}

func loadStackFromFS(root string) (map[string]string, error) {
	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}

	var stackIgnorePatterns []gitignore.Pattern
	stackIgnoreFile := filepath.Join(root, ".stackignore")
	_, err := os.Stat(stackIgnoreFile)
	if err == nil {
		if f, err := os.Open(stackIgnoreFile); err == nil {
			defer f.Close()
			stackIgnorePatterns = sourceignore.ReadPatterns(f, []string{})
		} else {
			return nil, fmt.Errorf("cannot read .stackignore file: %s", err)
		}
	}

	ignore := gitignore.NewMatcher(stackIgnorePatterns)

	files := map[string]string{}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		path = strings.TrimPrefix(path, root)
		if info.IsDir() {
			return nil
		}

		if ignore.Match(strings.Split(path, "/"), false) {
			return nil
		}

		content, err := ioutil.ReadFile(filepath.Join(root, path))
		if err != nil {
			return fmt.Errorf("cannot get file: %s", err)
		}
		files[path] = string(content)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk %s: %s", root, err)
	}

	return files, nil
}

func IsVersionLocked(stackConfig dx.StackConfig) (bool, error) {
	gitAddress, err := giturl.ParseScp(stackConfig.Stack.Repository)
	if err != nil {
		_, err2 := os.Stat(stackConfig.Stack.Repository)
		if err2 != nil {
			return false, fmt.Errorf("cannot parse stacks's git address: %s", err)
		} else {
			return true, nil
		}
	}

	params, _ := url.ParseQuery(gitAddress.RawQuery)
	if _, found := params["sha"]; found {
		return true, nil
	}
	if _, found := params["tag"]; found {
		return true, nil
	}
	if _, found := params["branch"]; found {
		return true, nil
	}

	return false, nil
}

const DefaultStackURL = "https://github.com/gimlet-io/gimlet-stack-reference.git"

func LatestVersion(repoURL string) (string, error) {
	gitAddress, err := giturl.ParseScp(repoURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse stacks's git address: %s", err)
	}
	gitUrl := strings.ReplaceAll(repoURL, gitAddress.RawQuery, "")
	gitUrl = strings.ReplaceAll(gitUrl, "?", "")

	fs := memfs.New()
	opts := &git.CloneOptions{
		URL: gitUrl,
	}
	repo, err := git.Clone(memory.NewStorage(), fs, opts)
	if err != nil {
		return "", fmt.Errorf("cannot clone: %s", err)
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		return "", fmt.Errorf("cannot get tags: %s", err)
	}

	var latestTag semver.Version
	var latestTagString = ""
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		tagString := strings.TrimPrefix(string(tagRef.Name()), "refs/tags/")
		tagStringWithoutV := strings.TrimPrefix(tagString, "v")

		tag, err := semver.Make(tagStringWithoutV)
		if err != nil {
			return err
		}

		if latestTag.Compare(tag) == -1 {
			latestTag = tag
			latestTagString = tagString
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return latestTagString, nil
}

func VersionsSince(repoURL string, sinceString string) ([]string, error) {
	gitAddress, err := giturl.ParseScp(repoURL)
	if err != nil {
		return []string{}, fmt.Errorf("cannot parse stacks's git address: %s", err)
	}
	gitUrl := strings.ReplaceAll(repoURL, gitAddress.RawQuery, "")
	gitUrl = strings.ReplaceAll(gitUrl, "?", "")

	fs := memfs.New()
	opts := &git.CloneOptions{
		URL: gitUrl,
	}
	repo, err := git.Clone(memory.NewStorage(), fs, opts)
	if err != nil {
		return []string{}, fmt.Errorf("cannot clone: %s", err)
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		return []string{}, fmt.Errorf("cannot get tags: %s", err)
	}

	sinceStringWithoutV := strings.TrimPrefix(sinceString, "v")
	since, err := semver.Make(sinceStringWithoutV)
	if err != nil {
		return []string{}, err
	}

	tagsSince := []string{}
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		tagString := strings.TrimPrefix(string(tagRef.Name()), "refs/tags/")
		tagStringWithoutV := strings.TrimPrefix(tagString, "v")

		tag, err := semver.Make(tagStringWithoutV)
		if err != nil {
			return err
		}

		if since.Compare(tag) == -1 {
			tagsSince = append(tagsSince, tagString)
		}

		return nil
	})
	if err != nil {
		return []string{}, err
	}

	// After this, the versions are properly sorted
	sort.Strings(tagsSince)

	return tagsSince, nil
}

func CurrentVersion(repoURL string) string {
	gitAddress, err := giturl.ParseScp(repoURL)
	if err != nil {
		return ""
	}

	params, _ := url.ParseQuery(gitAddress.RawQuery)
	if tag, found := params["tag"]; found {
		return tag[0]
	}

	return ""
}

func RepoUrlWithoutVersion(repoURL string) string {
	gitAddress, err := giturl.ParseScp(repoURL)
	if err != nil {
		return ""
	}

	gitUrl := strings.ReplaceAll(repoURL, gitAddress.RawQuery, "")
	gitUrl = strings.ReplaceAll(gitUrl, "?", "")

	return gitUrl
}
