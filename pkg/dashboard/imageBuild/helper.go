package imageBuild

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TriggerImagebuild(
	gitRepoCache *nativeGit.RepoCache,
	agentHub *streaming.AgentHub,
	store *store.Store,
	artifact *dx.Artifact,
	imageBuildRequest *dx.ImageBuildRequest,
) (*model.Event, error) {
	sourcePath, err := prepSourceForImageBuild(
		gitRepoCache, artifact.Version.RepositoryName, artifact.Version.SHA,
	)
	if err != nil {
		return nil, err
	}
	imageBuildRequest.SourcePath = sourcePath

	imageBuildEvent, err := imageBuildRequestEvent(imageBuildRequest, artifact.Version.RepositoryName)
	if err != nil {
		return nil, err
	}
	event, err := store.CreateEvent(imageBuildEvent)
	if err != nil {
		return nil, err
	}

	agentHub.TriggerImageBuild(imageBuildEvent.ID, imageBuildRequest)
	return event, nil
}

func prepSourceForImageBuild(gitRepCache *nativeGit.RepoCache, ownerAndRepo string, sha string) (string, error) {
	repo, repoPath, err := gitRepCache.InstanceForWrite(ownerAndRepo)
	if err != nil {
		return "", fmt.Errorf("cannot get repo: %s", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("cannot get worktree: %s", err)
	}
	err = worktree.Reset(&git.ResetOptions{
		Commit: plumbing.NewHash(sha),
		Mode:   git.HardReset,
	})
	if err != nil {
		return "", fmt.Errorf("cannot set version: %s", err)
	}

	tarFile, err := ioutil.TempFile("/tmp", "source-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("cannot get temp file: %s", err)
	}

	err = tartar(tarFile.Name(), []string{repoPath})
	if err != nil {
		return "", fmt.Errorf("cannot tar folder: %s", err)
	}

	return tarFile.Name(), nil
}

// https://github.com/vladimirvivien/go-tar/blob/master/tartar/tartar.go
// tarrer walks paths to create tar file tarName
func tartar(tarName string, paths []string) (err error) {
	tarFile, err := os.Create(tarName)
	if err != nil {
		return err
	}
	defer func() {
		err = tarFile.Close()
	}()

	absTar, err := filepath.Abs(tarName)
	if err != nil {
		return err
	}

	// enable compression if file ends in .gz
	tw := tar.NewWriter(tarFile)
	if strings.HasSuffix(tarName, ".gz") || strings.HasSuffix(tarName, ".gzip") {
		gz := gzip.NewWriter(tarFile)
		defer gz.Close()
		tw = tar.NewWriter(gz)
	}
	defer tw.Close()

	// walk each specified path and add encountered file to tar
	for _, path := range paths {
		// validate path
		path = filepath.Clean(path)
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if absPath == absTar {
			fmt.Printf("tar file %s cannot be the source\n", tarName)
			continue
		}
		if absPath == filepath.Dir(absTar) {
			fmt.Printf("tar file %s cannot be in source %s\n", tarName, absPath)
			continue
		}

		walker := func(file string, finfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// fill in header info using func FileInfoHeader
			hdr, err := tar.FileInfoHeader(finfo, finfo.Name())
			if err != nil {
				return err
			}

			relFilePath := file
			if filepath.IsAbs(path) {
				relFilePath, err = filepath.Rel(path, file)
				if err != nil {
					return err
				}
			}
			// ensure header has relative file path
			hdr.Name = relFilePath

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			// if path is a dir, dont continue
			if finfo.Mode().IsDir() {
				return nil
			}

			// add file to tar
			srcFile, err := os.Open(file)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			_, err = io.Copy(tw, srcFile)
			if err != nil {
				return err
			}
			return nil
		}

		// build tar
		if err := filepath.Walk(path, walker); err != nil {
			fmt.Printf("failed to add %s to tar: %s\n", path, err)
		}
	}
	return nil
}

func imageBuildRequestEvent(imageBuildRequest *dx.ImageBuildRequest, repository string) (*model.Event, error) {
	requestStr, err := json.Marshal(imageBuildRequest)
	if err != nil {
		return nil, err
	}

	event := &model.Event{
		Type:       model.ImageBuildRequestedEvent,
		Blob:       string(requestStr),
		Repository: repository,
		SHA:        imageBuildRequest.Sha,
		Results: []model.Result{
			{
				Status: model.Pending,
			},
		},
	}

	return event, nil
}
