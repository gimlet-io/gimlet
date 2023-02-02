package store

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet-cli/pkg/dx"

	"github.com/stretchr/testify/assert"
)

func TestEventCRUD(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	artifactStr := `
{
  "version": {
    "repositoryName": "my-app",
    "sha": "ea9ab7cc31b2599bf4afcfd639da516ca27a4780",
    "branch": "master",
	"event": "pr",
    "authorName": "Jane Doe",
    "authorEmail": "jane@doe.org",
    "committerName": "Jane Doe",
    "committerEmail": "jane@doe.org",
    "message": "Bugfix 123",
    "url": "https://github.com/gimlet-io/gimlet-cli/commit/ea9ab7cc31b2599bf4afcfd639da516ca27a4780"
  },
  "items": [
    {
      "name": "CI",
      "url": "https://jenkins.example.com/job/dev/84/display/redirect"
    }
  ]
}
`

	var a dx.Artifact
	json.Unmarshal([]byte(artifactStr), &a)

	aModel, err := model.ToEvent(a)
	assert.Nil(t, err)

	savedEvent, err := s.CreateEvent(aModel)
	assert.Nil(t, err)
	assert.NotEqual(t, savedEvent.Created, 0)
	assert.Equal(t, savedEvent.Event, dx.PR)

	artifacts, err := s.Artifacts("", "", nil, "", []string{}, 0, 0, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(artifacts))
	assert.Equal(t, "ea9ab7cc31b2599bf4afcfd639da516ca27a4780", artifacts[0].SHA)
}

func TestAdvancedArtifactQueries(t *testing.T) {
	s := NewTest()
	defer func() {
		s.Close()
	}()

	err := setupData(s)
	assert.Nil(t, err)

	artifacts, err := s.Artifacts("", "", nil, "", []string{}, 0, 0, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(artifacts))
	assert.Equal(t, "sha1", artifacts[0].SHA)

	threeHoursAgo := time.Now().Add(-3 * time.Hour)
	artifacts, err = s.Artifacts("", "", nil, "", []string{}, 0, 0, &threeHoursAgo, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(artifacts))
	assert.Equal(t, "sha1", artifacts[0].SHA)

	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	artifacts, err = s.Artifacts("", "", nil, "", []string{}, 0, 0, &threeHoursAgo, &twoHoursAgo)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(artifacts))

	artifacts, err = s.Artifacts("", "", nil, "", []string{"sha1", "sha2"}, 0, 0, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(artifacts))
}

func setupData(s *Store) error {
	anHourAgo := time.Now().Add(-1 * time.Hour)
	aModel, _ := model.ToEvent(dx.Artifact{
		Version: dx.Version{
			SHA: "sha1",
		},
	})
	_, err := s.createEvent(aModel, anHourAgo.Unix())
	if err != nil {
		return err
	}

	tenHoursAgo := time.Now().Add(-10 * time.Hour)
	aModel, _ = model.ToEvent(dx.Artifact{
		Version: dx.Version{
			SHA: "sha2",
		},
	})
	_, err = s.createEvent(aModel, tenHoursAgo.Unix())
	return err
}
