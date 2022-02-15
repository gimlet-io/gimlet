package server

import (
	"testing"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
)

const staging = "staging"
const production = "production"
const myapp = "myapp"
const perAppLimit = 3

func TestInsertIntoRolloutHistory(t *testing.T) {
	rolloutHistory := []*Env{}

	rolloutHistory = insertIntoRolloutHistory(rolloutHistory, &dx.Release{Env: staging, App: myapp}, perAppLimit)
	rolloutHistory = insertIntoRolloutHistory(rolloutHistory, &dx.Release{Env: staging, App: myapp}, perAppLimit)
	if releasesLength(rolloutHistory, staging, myapp) != 2 {
		t.Errorf("should insert release")
	}

	rolloutHistory = insertIntoRolloutHistory(rolloutHistory, &dx.Release{Env: production, App: myapp}, perAppLimit)
	if releasesLength(rolloutHistory, staging, myapp) != 2 {
		t.Errorf("should keep releases")
	}
	if releasesLength(rolloutHistory, production, myapp) != 1 {
		t.Errorf("should insert release")
	}

	rolloutHistory = insertIntoRolloutHistory(rolloutHistory, &dx.Release{Env: staging, App: myapp}, perAppLimit)
	if releasesLength(rolloutHistory, staging, myapp) != 3 {
		t.Errorf("insertIntoRolloutHistory should have inserted a release")
	}

	rolloutHistory = insertIntoRolloutHistory(rolloutHistory, &dx.Release{Env: staging, App: myapp}, perAppLimit)
	if releasesLength(rolloutHistory, staging, myapp) != 3 {
		t.Errorf("should not have longer release history than per app limit")
	}
}

func TestOrderRolloutHistoryFromAscending(t *testing.T) {

	now := time.Now()

	rolloutHistory := []*Env{
		{
			Name: staging,
			Apps: []*App{
				{
					Name: myapp,
					Releases: []*dx.Release{
						{
							ArtifactID: "newer",
							Created:    now.Add(-5 * time.Second).Unix(),
						},
						{
							ArtifactID: "older",
							Created:    now.Add(-10 * time.Second).Unix(),
						},
					},
				},
			},
		},
	}

	rolloutHistory = orderRolloutHistoryFromAscending(rolloutHistory)
	if rolloutHistory[0].Apps[0].Releases[0].ArtifactID != "older" {
		t.Errorf("should have reversed the release order")
	}

	rolloutHistory = orderRolloutHistoryFromAscending(rolloutHistory)
	if rolloutHistory[0].Apps[0].Releases[0].ArtifactID != "older" {
		t.Errorf("should have kept the order since nothing changed")
	}
}

func releasesLength(rolloutHistory []*Env, envName string, app string) int {
	for _, env := range rolloutHistory {
		if env.Name == envName {
			for _, app := range env.Apps {
				return len(app.Releases)
			}
		}
	}

	return -1
}
