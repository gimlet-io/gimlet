package worker

import (
	"database/sql"
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
		yearAndWeek := fmt.Sprintf("%d-%d", year, week)
		_, err := w.store.KeyValue(yearAndWeek)
		if err == sql.ErrNoRows {
			deploys, rollbacks, mostTriggeredBy := w.deploymentActivity()
			alertSeconds, alertsPercentageChange := w.alertMetrics()
			serviceLag, repos := w.serviceInformations()

			msg := notifications.WeeklySummary(deploys, rollbacks, mostTriggeredBy, alertSeconds, alertsPercentageChange, serviceLag, repos)
			w.notificationsManager.Broadcast(msg)

			// w.store.SaveKeyValue(&model.KeyValue{Key: yearAndWeek})
			logrus.Info("newsletter notification sent")
		} else if err != nil {
			logrus.Errorf("cannot get key value")
		}
		logrus.Info("weekly report completed")
		time.Sleep(24 * time.Hour)
	}
}

func (w *weeklyReporter) deploymentActivity() (deploys int, rollbacks int, overallMostTriggeredBy string) {
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

func (w *weeklyReporter) alertMetrics() (alertSeconds int, change float64) {
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

	return alertSeconds, percentageChange(alertsBetweenPreviousTwoWeeksSeconds, alertSeconds)
}

func (w *weeklyReporter) serviceInformations() (map[string]float64, []string) {
	serviceLag := map[string]float64{}
	stagingBehindProdRepos := []string{}

	stagingReleases, err := appReleases(w.store, w.repoCache, "staging")
	if err != nil {
		logrus.Errorf("cannot get releases: %s", err)
		return serviceLag, stagingBehindProdRepos
	}

	prodReleases, err := appReleases(w.store, w.repoCache, "preview") // production
	if err != nil {
		logrus.Errorf("cannot get releases: %s", err)
		return serviceLag, stagingBehindProdRepos
	}

	for stagingApp, stagingRelease := range stagingReleases {
		for prodApp, prodRelease := range prodReleases {
			if stagingApp != prodApp {
				continue
			}

			if stagingRelease.Version.Created > prodRelease.Version.Created {
				serviceLag[stagingApp] = float64(stagingRelease.Version.Created - prodRelease.Version.Created)
			}

			if stagingRelease.Version.Created < prodRelease.Version.Created {
				stagingBehindProdRepos = append(stagingBehindProdRepos, stagingRelease.Version.RepositoryName)
			}
		}
	}

	return serviceLag, stagingBehindProdRepos
}

func appReleases(store *store.Store, repoCache *nativeGit.RepoCache, envName string) (map[string]*dx.Release, error) {
	env, err := store.GetEnvironment(envName)
	if err != nil {
		return nil, err
	}

	repo, pathToCleanUp, err := repoCache.InstanceForWriteWithHistory(env.AppsRepo) // using a copy of the repo to avoid concurrent map writes error
	defer repoCache.CleanupWrittenRepo(pathToCleanUp)
	if err != nil {
		return nil, err
	}

	appReleases, err := gitops.Status(repo, "", envName, env.RepoPerEnv, perf)
	if err != nil {
		return nil, err
	}

	for app, release := range appReleases { // decorate release with created time
		r, err := gitops.Releases(repo, app, envName, env.RepoPerEnv, nil, nil, 1, release.Version.RepositoryName, perf)
		if err != nil {
			logrus.Errorf("cannot get releases")
			continue
		}

		if len(r) == 0 {
			continue
		}

		release.Created = r[0].Created
	}

	return appReleases, nil
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
