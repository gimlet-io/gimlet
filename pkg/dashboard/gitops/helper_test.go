package gitops

import (
	"testing"

	"github.com/gimlet-io/gimlet/pkg/dx"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stretchr/testify/assert"
)

func Test_Releases(t *testing.T) {
	repo := initHistory()

	perf := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "a",
		Help: "a",
	}, []string{"function"})

	releases, err := Releases(repo, "my-app", "staging", false, nil, nil, 10, "", perf)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(releases), "should get all releases")

	releases, err = Releases(repo, "my-app3", "staging", true, nil, nil, 10, "", perf)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases), "should get all releases")

	releases, err = Releases(repo, "", "staging", true, nil, nil, 10, "", perf)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases), "should get all releases")
}

func Test_ReleasesLimit(t *testing.T) {
	repo := initHistory()

	perf := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "b",
		Help: "b",
	}, []string{"function"})

	releases, err := Releases(repo, "my-app", "staging", false, nil, nil, 1, "", perf)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases), "should get only one release")
}

func Test_ReleasesGitRepo(t *testing.T) {
	repo := initHistory()

	perf := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "c",
		Help: "c",
	}, []string{"function"})

	releases, err := Releases(repo, "my-app2", "staging", false, nil, nil, -1, "laszlocph/gimletd-test2", perf)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases), "should get the commit from the gitrepo")
	assert.Equal(t, "xxx", releases[0].App, "should get the commit from the gitrepo")

	releases, err = Releases(repo, "my-app3", "staging", true, nil, nil, -1, "laszlocph/gimletd-test3", perf)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases), "should get the commit from the gitrepo")
	assert.Equal(t, "fosdem-2024", releases[0].App, "should get the commit from the gitrepo")
}

func Test_Status(t *testing.T) {
	repo := initHistory()

	perf := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gimletd_perf",
		Help: "Performance of functions",
	}, []string{"function"})

	status, err := Status(repo, "my-app", "staging", false, perf)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(status), "should get release status for app")

	status, err = Status(repo, "my-app3", "staging", true, perf)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(status), "should get release status for app")

	status, err = Status(repo, "", "staging", false, perf)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(status), "should get release status for all apps")
}

func initHistory() *git.Repository {
	repo, _ := git.Init(memory.NewStorage(), memfs.New())

	nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app2/file":         `1`,
			"staging/my-app2/release.json": "{}",
			"staging/release.json":         "{}",
		},
		[]string{"staging/my-app2"},
		"First commit is not read - it's a bug",
	)

	releaseJson := `{"app":"xxx","env":"staging","artifactId":"my-app-94578d91-ef9a-413d-9afb-602256d2b124","triggeredBy":"policy","gitopsRef":"","gitopsRepo":"", "version":{"repositoryName":"laszlocph/gimletd-test2","sha":"d7aa20d7055999200b52c4ffd146d5c7c415e3e7","created":1622792757,"branch":"master","event":"pr"}}`
	nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app2/file":         `2`,
			"staging/my-app2/release.json": releaseJson,
			"staging/release.json":         releaseJson,
		},
		[]string{"staging/my-app2"},
		"1st commit",
	)

	releaseJson = `{"app":"fosdem-2021","env":"staging","artifactId":"my-app-94578d91-ef9a-413d-9afb-602256d2b124","triggeredBy":"policy","gitopsRef":"","gitopsRepo":"", "version":{"repositoryName":"laszlocph/gimletd-test","sha":"d7aa20d7055999200b52c4ffd146d5c7c415e3e7","created":1622792757,"branch":"master","event":"pr"}}`
	nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app/file":         `3`,
			"staging/my-app/release.json": releaseJson,
			"staging/release.json":        releaseJson,
		},
		[]string{"staging/my-app"},
		"1st commit",
	)

	releaseJson = `{"app":"fosdem-2022","env":"staging","artifactId":"my-app-94578d91-ef9a-413d-9afb-602256d2b124","triggeredBy":"policy","gitopsRef":"","gitopsRepo":"", "version":{"repositoryName":"laszlocph/gimletd-test","sha":"d7aa20d7055999200b52c4ffd146d5c7c415e3e7","created":1622792757,"branch":"master","event":"pr"}}`
	nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app/file":         `4`,
			"staging/my-app/release.json": releaseJson,
			"staging/release.json":        releaseJson,
		},
		[]string{"staging/my-app"},
		"2nd commit",
	)

	releaseJson = `{"app":"fosdem-2023","env":"staging","artifactId":"my-app-94578d91-ef9a-413d-9afb-602256d2b124","triggeredBy":"policy","gitopsRef":"","gitopsRepo":"", "version":{"repositoryName":"laszlocph/gimletd-test","sha":"d7aa20d7055999200b52c4ffd146d5c7c415e3e7","created":1622792757,"branch":"master","event":"pr"}}`
	nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"staging/my-app/file":         `5`,
			"staging/my-app/release.json": releaseJson,
			"staging/release.json":        releaseJson,
		},
		[]string{"staging/my-app"},
		"3rd commit",
	)

	releaseJson = `{"app":"fosdem-2024","env":"staging","artifactId":"my-app-94578d91-ef9a-413d-9afb-602256d2b124","triggeredBy":"policy","gitopsRef":"","gitopsRepo":"", "version":{"repositoryName":"laszlocph/gimletd-test3","sha":"d7aa20d7055999200b52c4ffd146d5c7c415e3e7","created":1622792757,"branch":"master","event":"pr"}}`
	nativeGit.CommitFilesToGit(
		repo,
		map[string]string{
			"my-app3/file":         `5`,
			"my-app3/release.json": releaseJson,
			"release.json":         releaseJson,
		},
		[]string{"my-app3"},
		"4th commit",
	)

	return repo
}

func Test_ExtractImageStrategy(t *testing.T) {
	strategy := ExtractImageStrategy(&dx.Manifest{})
	assert.Equal(t, "dynamic", strategy)

	strategy = ExtractImageStrategy(&dx.Manifest{
		Chart: dx.Chart{
			Repository: "repository: https://chart.onechart.dev",
			Name:       "onechart",
		},
	})
	assert.Equal(t, "dynamic", strategy)

	strategy = ExtractImageStrategy(&dx.Manifest{
		Chart: dx.Chart{
			Repository: "repository: https://chart.onechart.dev",
			Name:       "onechart",
		},
		Values: map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "nginx",
				"tag":        "1.19",
				"strategy":   "static", //breaking change
			},
		},
	})
	assert.Equal(t, "static", strategy)

	strategy = ExtractImageStrategy(&dx.Manifest{
		Chart: dx.Chart{
			Repository: "repository: https://chart.onechart.dev",
			Name:       "static-site",
		},
	})
	assert.Equal(t, "static-site", strategy)

	strategy = ExtractImageStrategy(&dx.Manifest{
		Chart: dx.Chart{
			Repository: "repository: https://chart.onechart.dev",
			Name:       "onechart",
		},
		Values: map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "myimage",
				"tag":        "{{ .SHA }}",
			},
		},
	})
	assert.Equal(t, "dynamic", strategy)

	strategy = ExtractImageStrategy(&dx.Manifest{
		Chart: dx.Chart{
			Repository: "repository: https://chart.onechart.dev",
			Name:       "onechart",
		},
		Values: map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "127.0.0.1:32447",
				"tag":        "{{ .SHA }}",
				"strategy":   "buildpacks", //breaking change
			},
		},
	})
	assert.Equal(t, "buildpacks", strategy)

	strategy = ExtractImageStrategy(&dx.Manifest{
		Chart: dx.Chart{
			Repository: "repository: https://chart.onechart.dev",
			Name:       "onechart",
		},
		Values: map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "127.0.0.1:32447",
				"tag":        "{{ .SHA }}",
				"dockerfile": "Dockerfile",
				"strategy":   "dockerfile",
			},
		},
	})
	assert.Equal(t, "dockerfile", strategy)
}
