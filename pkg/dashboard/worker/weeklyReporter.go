package worker

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/customScm"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var perf = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "a",
	Help: "a",
}, []string{"function"})

type weeklyReporter struct {
	store                *store.Store
	repoCache            *nativeGit.RepoCache
	notificationsManager *notifications.ManagerImpl
	dynamicConfig        *dynamicconfig.DynamicConfig
	tokenManager         customScm.NonImpersonatedTokenManager
}

func NewWeeklyReporter(
	store *store.Store,
	repoCache *nativeGit.RepoCache,
	notificationsManager *notifications.ManagerImpl,
	dynamicConfig *dynamicconfig.DynamicConfig,
	tokenManager customScm.NonImpersonatedTokenManager,
) weeklyReporter {
	return weeklyReporter{
		store:                store,
		repoCache:            repoCache,
		notificationsManager: notificationsManager,
		dynamicConfig:        dynamicConfig,
		tokenManager:         tokenManager,
	}
}

func (w *weeklyReporter) Run() {
	for {
		year, week := time.Now().ISOWeek()
		yearAndWeek := fmt.Sprintf("%d-%d", year, week)
		_, err := w.store.KeyValue(yearAndWeek)
		if err == sql.ErrNoRows {
			deploys, rollbacks, mostTriggeredBy := w.deployments()
			seconds, change := w.alerts()
			lagSeconds := w.lag()
			repos := w.stagingBehindProdRepos()
			msg := notifications.WeeklySummary(deploys, rollbacks, mostTriggeredBy, seconds, change, lagSeconds, repos)
			w.notificationsManager.Broadcast(msg)
			w.store.SaveKeyValue(&model.KeyValue{Key: yearAndWeek})
			logrus.Info("newsletter notification sent")
		} else if err != nil {
			logrus.Errorf("cannot get key value")
		}
		logrus.Info("weekly report completed")
		time.Sleep(24 * time.Hour)
	}
}

func (w *weeklyReporter) deployments() (deploys int, rollbacks int, overallMostTriggeredBy string) {
	envs, _ := w.store.GetEnvironments()
	oneWeekAgo := time.Now().Add(-7 * time.Hour * 24)

	maxCount := 0
	for _, env := range envs {
		repo, pathToCleanUp, err := w.repoCache.InstanceForWriteWithHistory(env.AppsRepo) // using a copy of the repo to avoid concurrent map writes error
		defer w.repoCache.CleanupWrittenRepo(pathToCleanUp)
		if err != nil {
			logrus.Errorf("cannot get gitops repo for write: %s", err)
			continue
		}

		releases, err := gitops.Releases(repo, "", env.Name, env.RepoPerEnv, &oneWeekAgo, nil, -1, "", perf)
		if err != nil {
			logrus.Errorf("cannot get releases: %s", err)
			continue
		}

		mostTriggered := mostTriggeredBy(releases)
		count := len(releases)
		if count > maxCount {
			maxCount = count
			overallMostTriggeredBy = mostTriggered
		}

		for _, r := range releases {
			if r.RolledBack {
				rollbacks++
			} else {
				deploys++
			}
		}
	}

	return deploys, rollbacks, overallMostTriggeredBy
}

func (w *weeklyReporter) alerts() (alertSeconds int, change float64) {
	alerts, err := w.store.AlertsInWeek()
	if err != nil {
		logrus.Errorf("cannot get alerts: %s", err)
	}

	alertsBetweenPreviousTwoWeeks, err := w.store.AlertsBetweenPreviousTwoWeeks()
	if err != nil {
		logrus.Errorf("cannot get alerts: %s", err)
	}

	alertSeconds = calcDuration(alerts)
	alertsBetweenPreviousTwoWeeksSeconds := calcDuration(alertsBetweenPreviousTwoWeeks)
	change = percentageChange(alertsBetweenPreviousTwoWeeksSeconds, alertSeconds)
	if math.IsNaN(change) || math.IsInf(change, 1) {
		change = 0
	}

	return alertSeconds, change
}

func (w *weeklyReporter) lag() (lagSeconds map[string]int64) {
	apps := []string{"getting-started-app"}

	for _, app := range apps {
		prod, err := w.latestRelease(app, "production", "")
		if err != nil {
			logrus.Errorf("cannot get latest release: %s", err)
			continue
		}

		staging, err := w.latestRelease(app, "staging", "")
		if err != nil {
			logrus.Errorf("cannot get latest release: %s", err)
			continue
		}

		if prod == nil || staging == nil {
			continue
		}

		lagSeconds[app] = staging.Created - prod.Created
	}
	return
}

func (w *weeklyReporter) stagingBehindProdRepos() (filtered []string) {
	token, _, _ := w.tokenManager.Token()
	gitSvc := customScm.NewGitService(w.dynamicConfig)
	repos, err := gitSvc.OrgRepos(token)
	if err != nil {
		logrus.Errorf("cannot get repos: %s", err)
		return
	}

	for _, repo := range repos {
		prod, err := w.latestRelease("", "production", repo)
		if err != nil {
			logrus.Errorf("cannot get latest release: %s", err)
			continue
		}

		staging, err := w.latestRelease("", "staging", repo)
		if err != nil {
			logrus.Errorf("cannot get latest release: %s", err)
			continue
		}

		if prod == nil || staging == nil {
			continue
		}

		if staging.Created < prod.Created {
			filtered = append(filtered, repo)
		}
	}
	return
}

func (w *weeklyReporter) latestRelease(app, envName, gitRepo string) (*dx.Release, error) {
	env, err := w.store.GetEnvironment(envName)
	if err != nil {
		return nil, err
	}

	repo, pathToCleanUp, err := w.repoCache.InstanceForWriteWithHistory(env.AppsRepo) // using a copy of the repo to avoid concurrent map writes error
	defer w.repoCache.CleanupWrittenRepo(pathToCleanUp)
	if err != nil {
		return nil, err
	}

	releases, err := gitops.Releases(repo, app, env.Name, env.RepoPerEnv, nil, nil, 1, gitRepo, perf)
	if err != nil {
		return nil, err
	}
	return releases[0], nil
}

func calcDuration(alerts []*model.Alert) (seconds int) {
	for _, a := range alerts {
		if a.FiredAt == 0 {
			continue
		}

		resolved := time.Now().Unix()
		if a.ResolvedAt != 0 {
			resolved = a.ResolvedAt
		}

		seconds += (int(resolved) - int(a.FiredAt))
	}
	return seconds
}

func percentageChange(old, new int) (delta float64) {
	diff := float64(new - old)
	delta = (diff / float64(old)) * 100
	return
}

func mostTriggeredBy(releases []*dx.Release) (mostTriggeredBy string) {
	triggerCount := map[string]int{}

	for _, release := range releases {
		triggerCount[release.TriggeredBy]++
	}

	maxCount := 0
	for name, count := range triggerCount {
		if count > maxCount {
			maxCount = count
			mostTriggeredBy = name
		}
	}

	return
}
