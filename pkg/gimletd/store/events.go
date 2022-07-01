package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/store/sql"
	"github.com/google/uuid"
	"github.com/russross/meddler"
)

// CreateEvent stores a new event in the database
func (db *Store) CreateEvent(event *model.Event) (*model.Event, error) {
	event.ID = uuid.New().String()
	event.Created = time.Now().Unix()
	event.Status = model.StatusNew
	return event, meddler.Insert(db, "events", event)
}

// createEvent stores a new event in the database, but it is able to fake the created date.
// Should be only used in tests
func (db *Store) createEvent(event *model.Event, created int64) (*model.Event, error) {
	event.ID = uuid.New().String()
	event.Created = created
	event.Status = model.StatusNew
	return event, meddler.Insert(db, "events", event)
}

// Artifacts returns all events in the database within the given constraints
func (db *Store) Artifacts(
	repo, branch string,
	gitEvent *dx.GitEvent,
	sourceBranch string,
	sha []string,
	limit, offset int,
	since, until *time.Time) ([]*model.Event, error) {

	filters := []string{}
	args := []interface{}{}

	filters = addFilter(filters, "type = $1")
	args = append(args, model.ArtifactCreatedEvent)

	if since != nil {
		filters = addFilter(filters, fmt.Sprintf("created >= $%d", len(filters)+1))
		args = append(args, since.Unix())
	}
	if until != nil {
		filters = addFilter(filters, fmt.Sprintf("created < $%d", len(filters)+1))
		args = append(args, until.Unix())
	}

	if repo != "" {
		filters = addFilter(filters, fmt.Sprintf("repository = $%d", len(filters)+1))
		args = append(args, repo)
	}
	if branch != "" {
		filters = addFilter(filters, fmt.Sprintf("branch = $%d", len(filters)+1))
		args = append(args, branch)
	}
	if sourceBranch != "" {
		filters = addFilter(filters, fmt.Sprintf("branch = $%d", len(filters)+1))
		args = append(args, sourceBranch)
	}
	if len(sha) != 0 {
		for idx, s := range sha {
			if idx == 0 {
				if len(sha) == 1 {
					filters = append(filters, fmt.Sprintf(" AND sha in ($%d)", len(filters)+1))
				} else {
					filters = append(filters, fmt.Sprintf(" AND sha in ($%d", len(filters)+1))
				}
			} else if idx == len(sha)-1 {
				filters = append(filters, fmt.Sprintf(", $%d)", len(filters)+1))
			} else {
				filters = append(filters, fmt.Sprintf(", $%d", len(filters)+1))
			}
			args = append(args, s)
		}
	}

	if gitEvent != nil {
		var intRep int
		intRep = int(*gitEvent)
		filters = addFilter(filters, fmt.Sprintf(" event = %d", intRep))
	}

	if limit == 0 && offset == 0 {
		limit = 10
	}
	limitAndOffset := fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)

	query := fmt.Sprintf(`
SELECT id, repository, branch, event, source_branch, target_branch, tag, created, blob, status, status_desc, sha, artifact_id
FROM events
%s
ORDER BY created desc
%s;`, strings.Join(filters, " "), limitAndOffset)

	var data []*model.Event
	err := meddler.QueryAll(db, &data, query, args...)
	return data, err
}

// Artifact returns an artifact by id
func (db *Store) Artifact(id string) (*model.Event, error) {
	query := fmt.Sprintf(`
SELECT id, repository, branch, event, source_branch, target_branch, tag, created, blob, status, status_desc, sha, artifact_id
FROM events
WHERE artifact_id = $1;
`)

	var data model.Event
	err := meddler.QueryRow(db, &data, query, id)
	return &data, err
}

// Event returns an event by id
func (db *Store) Event(id string) (*model.Event, error) {
	query := fmt.Sprintf(`
SELECT id, created, blob, status, status_desc, gitops_hashes, results
FROM events
WHERE id = $1;
`)

	var data model.Event
	err := meddler.QueryRow(db, &data, query, id)
	return &data, err
}

// UnprocessedEvents selects an event timeline
func (db *Store) UnprocessedEvents() (events []*model.Event, err error) {
	stmt := sql.Stmt(db.driver, sql.SelectUnprocessedEvents)
	err = meddler.QueryAll(db, &events, stmt)
	return events, err
}

// UpdateEventStatus updates an event status in the database
func (db *Store) UpdateEventStatus(id string, status string, desc string, gitopsStatusString string, results string) error {
	stmt := sql.Stmt(db.driver, sql.UpdateEventStatus)
	_, err := db.Exec(stmt, status, desc, gitopsStatusString, results, id)
	return err
}

func addFilter(filters []string, filter string) []string {
	if len(filters) == 0 {
		return append(filters, "WHERE "+filter)
	}

	return append(filters, "AND "+filter)
}
