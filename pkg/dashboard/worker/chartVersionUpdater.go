package worker

import (
	"fmt"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/genericScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/sirupsen/logrus"
	giturl "github.com/whilp/git-urls"
)

type ChartVersionUpdater struct {
	gitSvc       customScm.CustomGitService
	tokenManager customScm.NonImpersonatedTokenManager
	repoCache    *nativeGit.RepoCache
	goScm        *genericScm.GoScmHelper
}

func NewChartVersionUpdater(
	gitSvc customScm.CustomGitService,
	tokenManager customScm.NonImpersonatedTokenManager,
	repoCache *nativeGit.RepoCache,
	goScm *genericScm.GoScmHelper,
) *ChartVersionUpdater {
	return &ChartVersionUpdater{
		gitSvc:       gitSvc,
		tokenManager: tokenManager,
		repoCache:    repoCache,
		goScm:        goScm,
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
			err = server.UpdateRepoEnvConfigsChartVersion(
				c.repoCache,
				c.goScm,
				token,
				repoName,
			)
			if err != nil {
				logrus.Errorf("cannot update chart versions for %s: %s", repoName, err)
			}
		}

		logrus.Info("chart version update process completed")
		time.Sleep(24 * time.Hour)
	}
}

func updateChartVersion(raw string, latestVersion string) string {
	gitAddress, _ := giturl.Parse(latestVersion)
	gitUrl := strings.ReplaceAll(latestVersion, gitAddress.RawQuery, "")
	gitUrl = strings.ReplaceAll(gitUrl, "?", "")

	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, fmt.Sprintf("  name: %s", gitUrl)) {
			lines[i] = fmt.Sprintf("  name: %s", latestVersion)
			break
		}
		if strings.HasPrefix(line, "  version:") {
			lines[i] = fmt.Sprintf("  version: %s", latestVersion)
			break
		}
	}
	return strings.Join(lines, "\n")
}
