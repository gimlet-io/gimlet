package worker

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/api"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	helper "github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/sirupsen/logrus"
	giturl "github.com/whilp/git-urls"
	"sigs.k8s.io/yaml"
)

type ChartVersionUpdater struct {
	config            *config.Config
	dynamicConfig     *dynamicconfig.DynamicConfig
	tokenManager      customScm.NonImpersonatedTokenManager
	repoCache         *helper.RepoCache
	chartUpdatePrList *map[string]interface{}
}

func NewChartVersionUpdater(
	config *config.Config,
	dynamicConfig *dynamicconfig.DynamicConfig,
	tokenManager customScm.NonImpersonatedTokenManager,
	repoCache *helper.RepoCache,
	chartUpdatePrList *map[string]interface{},
) *ChartVersionUpdater {
	return &ChartVersionUpdater{
		config:            config,
		dynamicConfig:     dynamicConfig,
		tokenManager:      tokenManager,
		repoCache:         repoCache,
		chartUpdatePrList: chartUpdatePrList,
	}
}

func (c *ChartVersionUpdater) Run() {
	for {
		(*c.chartUpdatePrList) = map[string]interface{}{}
		token, _, _ := c.tokenManager.Token()
		gitSvc := customScm.NewGitService(c.dynamicConfig)

		repos, err := gitSvc.OrgRepos(token)
		if err != nil {
			logrus.Errorf("cannot get org repos: %s", err)
		}
		for _, repoName := range repos {
			err = c.updateRepoEnvConfigsChartVersion(token, repoName)
			if err != nil {
				logrus.Errorf("cannot update chart versions for %s: %s", repoName, err)
			}
		}

		logrus.Info("chart version update process completed")
		time.Sleep(24 * time.Hour)
	}
}

func (c *ChartVersionUpdater) updateRepoEnvConfigsChartVersion(token string, repoName string) error {
	logrus.Infof("evaluating %s for chart version update", repoName)
	goScmHelper := genericScm.NewGoScmHelper(c.dynamicConfig, nil)
	prList, err := goScmHelper.ListOpenPRs(token, repoName)
	if err != nil {
		return fmt.Errorf("cannot list pull requests: %s", err)
	}
	for _, pullRequest := range prList {
		if strings.HasPrefix(pullRequest.Source, "gimlet-chart-update") {
			(*c.chartUpdatePrList)[repoName] = &api.PR{
				Sha:   pullRequest.Sha,
				Link:  pullRequest.Link,
				Title: pullRequest.Title,
			}
			return nil
		}
	}

	repo, tmpPath, err := c.repoCache.InstanceForWrite(repoName)
	if err != nil {
		os.RemoveAll(tmpPath)
		return fmt.Errorf("could not open %s: %s", repoName, err)
	}

	headBranch, err := helper.HeadBranch(repo)
	if err != nil {
		return fmt.Errorf("cannot get head branch: %s", err)
	}

	files, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, headBranch, ".gimlet")
	if err != nil {
		if !strings.Contains(err.Error(), "directory not found") {
			return fmt.Errorf("cannot list files in .gimlet/: %s", err)
		}
	}

	configsPerEnv, err := configsPerEnv(files)
	if err != nil {
		return fmt.Errorf("cannot extract configs per env: %s", err)
	}

	for envName, configs := range configsPerEnv {
		// Checkout the head branch before creating a new branch
		err = helper.Checkout(repo, fmt.Sprintf("refs/heads/%s", headBranch))
		if err != nil {
			logrus.Warnf("cannot checkout head branch: %s", err)
			continue
		}

		sourceBranch, err := server.GenerateBranchNameWithUniqueHash("gimlet-chart-update", 4)
		if err != nil {
			logrus.Warnf("cannot generate branch name: %s", err)
			continue
		}

		err = helper.Branch(repo, fmt.Sprintf("refs/heads/%s", sourceBranch))
		if err != nil {
			logrus.Warnf("cannot checkout branch: %s", err)
			continue
		}

		for fileName, content := range configs {
			latestVersion := findLatestVersion(content, c.config.Charts)

			updatedContent := updateChartVersion(content, latestVersion)

			_ = os.MkdirAll(filepath.Join(tmpPath, ".gimlet"), helper.Dir_RWX_RX_R)
			err = os.WriteFile(filepath.Join(tmpPath, fmt.Sprintf(".gimlet/%s", fileName)), []byte(updatedContent), helper.Dir_RWX_RX_R)
			if err != nil {
				logrus.Warnf("cannot write file in %s: %s", repoName, err)
				continue
			}
		}

		empty, err := helper.NothingToCommit(repo)
		if err != nil {
			logrus.Warnf("cannot get git state: %s", err)
			continue
		}
		if empty {
			continue
		}

		err = server.StageCommitAndPush(repo, tmpPath, token, "[Gimlet] Deployment template update")
		if err != nil {
			logrus.Warnf("cannot stage, commit and push: %s", err)
			continue
		}

		createdPr, _, err := goScmHelper.CreatePR(token, repoName, sourceBranch, headBranch,
			fmt.Sprintf("[Gimlet] Deployment template update for %s", envName),
			"This is an automated Pull Request that updates the Helm chart version in Gimlet manifests.")
		if err != nil {
			logrus.Warnf("cannot create pull request: %s", err)
			continue
		}
		(*c.chartUpdatePrList)[repoName] = &api.PR{
			Sha:   createdPr.Sha,
			Link:  createdPr.Link,
			Title: createdPr.Title,
		}
		logrus.Infof("pull request created for %s, %s with chart version update", repoName, envName)
	}
	return nil
}

func updateChartVersion(raw string, latestVersion string) string {
	if latestVersion == "" {
		return raw
	}

	gitAddress, _ := giturl.Parse(latestVersion)
	gitUrl := strings.ReplaceAll(latestVersion, gitAddress.RawQuery, "")
	gitUrl = strings.ReplaceAll(gitUrl, "?", "")
	params, _ := url.ParseQuery(gitAddress.RawQuery)
	var latestHash string
	if v, found := params["sha"]; found {
		latestHash = fmt.Sprintf("sha=%s", v[0])
	}

	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, fmt.Sprintf("  name: %s", gitUrl)) {
			regex := regexp.MustCompile(`sha=([^& ]+)`)
			lines[i] = regex.ReplaceAllString(line, latestHash)
			break
		}
		if strings.HasPrefix(line, "  version:") {
			lines[i] = fmt.Sprintf("  version: %s", latestVersion)
			break
		}
	}
	return strings.Join(lines, "\n")
}

func configsPerEnv(files map[string]string) (map[string]map[string]string, error) {
	configsPerEnv := map[string]map[string]string{}
	for fileName, content := range files {
		var manifest dx.Manifest
		err := yaml.Unmarshal([]byte(content), &manifest)
		if err != nil {
			return nil, err
		}

		if configsPerEnv[manifest.Env] == nil {
			configsPerEnv[manifest.Env] = map[string]string{}
		}

		configsPerEnv[manifest.Env][fileName] = content
	}
	return configsPerEnv, nil
}

func findLatestVersion(content string, charts config.Charts) string {
	var manifest dx.Manifest
	err := yaml.Unmarshal([]byte(content), &manifest)
	if err != nil {
		logrus.Warnf("cannot parse manifest %s", err)
		return ""
	}

	return findChartInConfig(charts, manifest.Chart.Name)
}

func findChartInConfig(charts config.Charts, chartName string) string {
	if strings.HasPrefix(chartName, "git@") || strings.Contains(chartName, ".git") {
		path := chartPath(chartName)
		return charts.FindGitRepoHTTPSScheme(path)
	}

	return charts.Find(chartName)
}

func chartPath(chartName string) string {
	gitAddress, _ := giturl.Parse(chartName)
	params, _ := url.ParseQuery(gitAddress.RawQuery)
	v := params["path"]

	return v[0]
}
