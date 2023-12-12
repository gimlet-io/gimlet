package worker

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/gitops"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/notifications"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/git/nativeGit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type weeklyReporter struct {
	store                *store.Store
	repoCache            *nativeGit.RepoCache
	notificationsManager *notifications.ManagerImpl
	dynamicConfig        *dynamicconfig.DynamicConfig
	perf                 *prometheus.HistogramVec
}

func NewWeeklyReporter(
	store *store.Store,
	repoCache *nativeGit.RepoCache,
	notificationsManager *notifications.ManagerImpl,
	dynamicConfig *dynamicconfig.DynamicConfig,
	perf *prometheus.HistogramVec,
) weeklyReporter {
	return weeklyReporter{
		store:                store,
		repoCache:            repoCache,
		notificationsManager: notificationsManager,
		dynamicConfig:        dynamicConfig,
		perf:                 perf,
	}
}

func (w *weeklyReporter) Run() {
	for {
		year, week := time.Now().ISOWeek()
		yearAndWeek := fmt.Sprintf("%d-%d", year, week)
		_, err := w.store.KeyValue(yearAndWeek)
		if err == sql.ErrNoRows {
			since, until := weekRange(year, week-1)
			deploys, rollbacks, mostTriggeredBy := w.deploymentActivity(since, until)
			alertSeconds, alertsPercentageChange := w.alertMetrics(since, until)
			serviceLag, repos := w.serviceInformations()

			msg := notifications.WeeklySummary(deploys, rollbacks, mostTriggeredBy, alertSeconds, alertsPercentageChange, serviceLag, repos, w.dynamicConfig.ScmURL())
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

func (w *weeklyReporter) deploymentActivity(since, until time.Time) (deploys int, rollbacks int, overallMostTriggeredBy string) {
	envs, _ := w.store.GetEnvironments()
	maxCount := 0
	for _, env := range envs {
		repo, pathToCleanUp, err := w.repoCache.InstanceForWriteWithHistory(env.AppsRepo) // using a copy of the repo to avoid concurrent map writes error
		defer w.repoCache.CleanupWrittenRepo(pathToCleanUp)
		if err != nil {
			logrus.Errorf("cannot get gitops repo for write: %s", err)
			continue
		}

		releases, err := gitops.Releases(repo, "", env.Name, env.RepoPerEnv, &since, &until, -1, "", w.perf)
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

func (w *weeklyReporter) alertMetrics(since, until time.Time) (alertSeconds int, change float64) {
	alerts, err := w.store.AlertsInterval(since, until)
	if err != nil {
		logrus.Errorf("cannot get alerts: %s", err)
	}

	minusOneWeek := -7 * time.Hour * 24
	alertsBetweenPreviousTwoWeeks, err := w.store.AlertsInterval(since.Add(minusOneWeek), until.Add(minusOneWeek))
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

	stagingReleases, err := appReleases(w.store, w.repoCache, w.perf, "staging")
	if err != nil {
		logrus.Errorf("cannot get releases for staging: %s", err)
		return serviceLag, stagingBehindProdRepos
	}

	prodReleases, err := appReleases(w.store, w.repoCache, w.perf, "production")
	if err != nil {
		logrus.Errorf("cannot get releases for production: %s", err)
		return serviceLag, stagingBehindProdRepos
	}

	for stagingApp, stagingRelease := range stagingReleases {
		prodRelease, exists := prodReleases[stagingApp]
		if !exists {
			continue
		}

		if stagingRelease.Version.Created > prodRelease.Version.Created {
			serviceLag[stagingApp] = float64(stagingRelease.Version.Created - prodRelease.Version.Created)
		}

		if stagingRelease.Version.Created < prodRelease.Version.Created {
			stagingBehindProdRepos = append(stagingBehindProdRepos, stagingRelease.Version.RepositoryName)
		}
	}

	return serviceLag, stagingBehindProdRepos
}

func appReleases(store *store.Store, repoCache *nativeGit.RepoCache, perf *prometheus.HistogramVec, envName string) (map[string]*dx.Release, error) {
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

func weekRange(year, week int) (time.Time, time.Time) {
	t := time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)

	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)

	return t, t.AddDate(0, 0, 7)
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
