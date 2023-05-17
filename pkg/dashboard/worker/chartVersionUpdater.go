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
	gitSvc       customScm.CustomGitService
	tokenManager customScm.NonImpersonatedTokenManager
	repoCache    *helper.RepoCache
	goScm        *genericScm.GoScmHelper
	chart        config.Chart
}

func NewChartVersionUpdater(
	gitSvc customScm.CustomGitService,
	tokenManager customScm.NonImpersonatedTokenManager,
	repoCache *helper.RepoCache,
	goScm *genericScm.GoScmHelper,
	chart config.Chart,
) *ChartVersionUpdater {
	return &ChartVersionUpdater{
		gitSvc:       gitSvc,
		tokenManager: tokenManager,
		repoCache:    repoCache,
		goScm:        goScm,
		chart:        chart,
	}
}

func (c *ChartVersionUpdater) Run() {
	for {
		token, _, _ := c.tokenManager.Token()
		repos, err := c.gitSvc.OrgRepos(token)
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
	prList, err := c.goScm.ListOpenPRs(token, repoName)
	if err != nil {
		return fmt.Errorf("cannot list pull requests: %s", err)
	}
	for _, pullRequest := range prList {
		if strings.HasPrefix(pullRequest.Source, "gimlet-chart-update") {
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

	sourceBranch, err := server.GenerateBranchNameWithUniqueHash("gimlet-chart-update", 4)
	if err != nil {
		return fmt.Errorf("cannot generate branch name: %s", err)
	}

	err = helper.Branch(repo, fmt.Sprintf("refs/heads/%s", sourceBranch))
	if err != nil {
		return fmt.Errorf("cannot checkout branch: %s", err)
	}

	files, err := helper.RemoteFolderOnBranchWithoutCheckout(repo, headBranch, ".gimlet")
	if err != nil {
		if !strings.Contains(err.Error(), "directory not found") {
			return fmt.Errorf("cannot list files in .gimlet/: %s", err)
		}
	}

	for fileName, content := range files {
		latestVersion := c.chart.Version

		chartFromGitRepo, err := isChartFromGitRepo(content)
		if err != nil {
			logrus.Warnf("cannot parse manifest string: %s", err)
			continue
		}
		if chartFromGitRepo {
			latestVersion = c.chart.Name
		}
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
		return fmt.Errorf("cannot get git state: %s", err)
	}
	if empty {
		return nil
	}

	err = server.StageCommitAndPush(repo, tmpPath, token, "[Gimlet] Deployment template update")
	if err != nil {
		return fmt.Errorf("cannot stage, commit and push: %s", err)
	}

	_, _, err = c.goScm.CreatePR(token, repoName, sourceBranch, headBranch,
		"[Gimlet] Deployment template update",
		"This is an automated Pull Request that updates the Helm chart version in Gimlet manifests.")
	if err != nil {
		return fmt.Errorf("cannot create pull request: %s", err)
	}
	logrus.Infof("pull request created for %s with chart version update", repoName)
	return nil
}

func updateChartVersion(raw string, latestVersion string) string {
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

func isChartFromGitRepo(content string) (bool, error) {
	var manifest dx.Manifest
	err := yaml.Unmarshal([]byte(content), &manifest)
	if err != nil {
		return false, err
	}
	return strings.HasPrefix(manifest.Chart.Name, "git@") || strings.Contains(manifest.Chart.Name, ".git"), nil
}
