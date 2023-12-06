package worker

import (
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
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
}

func NewWeeklyReporter(
	store *store.Store,
	repoCache *nativeGit.RepoCache,
	notificationsManager *notifications.ManagerImpl,
) weeklyReporter {
	return weeklyReporter{
		store:                store,
		repoCache:            repoCache,
		notificationsManager: notificationsManager,
	}
}

func (w *weeklyReporter) Run() {
	for {
		year, week := time.Now().ISOWeek()
		fmt.Printf("%d-%d", year, week)
		// if !sent {
		deploys, rollbacks, mostTriggeredBy := w.deployments()
		seconds, change := w.alerts()
		lagSeconds := w.lag()
		msg := notifications.WeeklySummary(deploys, rollbacks, mostTriggeredBy, seconds, change, lagSeconds, []string{})
		w.notificationsManager.Broadcast(msg)
		//}
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
	alerts, _ := w.store.AlertsWithinAWeek()
	alertSeconds = seconds(alerts)

	// TODO
	// oldAlerts, _ := w.store.AlertsEarlier()
	// oldAlertSeconds := seconds(oldAlerts)
	oldAlertSeconds := 0
	change = percentageChange(oldAlertSeconds, alertSeconds)
	return alertSeconds, change
}

func seconds(alerts []*model.Alert) (seconds int) {
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

func (w *weeklyReporter) lag() (lagSeconds int64) { // TODO nil pointer error
	prod, _ := w.store.GetEnvironment("production")
	latestProdRelease := latestRelease(w.repoCache, prod)
	staging, _ := w.store.GetEnvironment("staging")
	latestStagingRelease := latestRelease(w.repoCache, staging)

	if latestProdRelease == nil || latestStagingRelease == nil {
		return 0
	}

	return latestStagingRelease.Created - latestProdRelease.Created
}

func latestRelease(repoCache *nativeGit.RepoCache, env *model.Environment) *dx.Release {
	repo, pathToCleanUp, err := repoCache.InstanceForWriteWithHistory(env.AppsRepo) // using a copy of the repo to avoid concurrent map writes error
	defer repoCache.CleanupWrittenRepo(pathToCleanUp)
	if err != nil {
		logrus.Errorf("cannot get gitops repo for write: %s", err)
		return nil
	}

	releases, err := gitops.Releases(repo, "", env.Name, env.RepoPerEnv, nil, nil, 1, "", perf)
	if err != nil {
		logrus.Errorf("cannot get releases: %s", err)
		return nil
	}
	return releases[0]
}

func mostTriggeredBy(releases []*dx.Release) (mostTriggerer string) {
	triggerCount := make(map[string]int)

	// Count occurrences of TriggeredBy names
	for _, release := range releases {
		triggerCount[release.TriggeredBy]++
	}

	// Find the name with the highest count
	maxCount := 0
	for name, count := range triggerCount {
		if count > maxCount {
			maxCount = count
			mostTriggerer = name
		}
	}

	return mostTriggerer
}
