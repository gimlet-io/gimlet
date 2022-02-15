package notifications

import (
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet-cli/pkg/dx"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/model"
	"github.com/gimlet-io/gimlet-cli/pkg/gimletd/worker/events"
)

func TestSendingFluxMessage(t *testing.T) {

	msgHealthCheckPassed := fluxMessage{
		gitopsCommit: &model.GitopsCommit{
			ID:         200,
			Sha:        "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			Status:     "Progressing",
			StatusDesc: "Health check passed",
		},
		gitopsRepo: "gimlet",
		env:        "staging",
	}

	discordMessageHealthCheckPassed, err := msgHealthCheckPassed.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageHealthCheckPassed.Embed.Description, "Applied resources from") {
		t.Errorf("Flux health check passed message must contain 'Applied resources from'")
	}

	msgHealthCheckProgressing := fluxMessage{
		gitopsCommit: &model.GitopsCommit{
			ID:         200,
			Sha:        "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			Status:     "Progressing",
			StatusDesc: "progressing",
		},
		gitopsRepo: "gimlet",
		env:        "staging",
	}

	discordMessageHealthCheckProgressing, err := msgHealthCheckProgressing.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageHealthCheckProgressing.Embed.Description, "Applying gitops changes from") {
		t.Errorf("Flux health check progressing message must contain 'Applying gitops changes from'")
	}

	msgHealthCheckFailed := fluxMessage{
		gitopsCommit: &model.GitopsCommit{
			ID:         200,
			Sha:        "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			Status:     "ReconciliationFailed",
			StatusDesc: "progressing",
		},
		gitopsRepo: "gimlet",
		env:        "staging",
	}

	discordMessageHealthCheckFailed, err := msgHealthCheckFailed.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageHealthCheckFailed.Embed.Description, "failed to apply") {
		t.Errorf("Flux health check failed message message must contain 'failed to apply'")
	}

}

func TestSendingGitopsDeleteMessage(t *testing.T) {

	msgDeleteFailed := gitopsDeleteMessage{
		event: &events.DeleteEvent{
			Env:         "staging",
			App:         "myapp",
			TriggeredBy: "Gimlet",
			StatusDesc:  "cannot delete",
			GitopsRef:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			GitopsRepo:  "testrepo",
			Status:      1,
		},
	}

	discordMessageDeleteFailed, err := msgDeleteFailed.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageDeleteFailed.Embed.Description, "cannot delete") {
		t.Errorf("Gitops deletion failed message must contain 'cannot delete'")
	}

	msgPolicyDeletion := gitopsDeleteMessage{
		event: &events.DeleteEvent{
			Env:         "staging",
			App:         "myapp",
			TriggeredBy: "policy",
			StatusDesc:  "cannot delete",
			GitopsRef:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			GitopsRepo:  "testrepo",
			Status:      2,
		},
	}

	discordMessagePolicyDeletion, err := msgPolicyDeletion.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessagePolicyDeletion.Text, "Policy based deletion") {
		t.Errorf("Gitops policy deletion message must contain 'Policy based deletion'")
	}

}

func TestSendingGitopsDeployMessage(t *testing.T) {

	version := &dx.Version{
		RepositoryName: "testrepo",
		URL:            "https://gimlet.io",
	}

	msgSendFailure := gitopsDeployMessage{
		event: &events.DeployEvent{
			Manifest: &dx.Manifest{
				App: "myapp",
				Env: "staging",
			},
			Artifact: &dx.Artifact{
				Version: *version,
			},
			TriggeredBy: "Gimlet",
			Status:      1,
			StatusDesc:  "",
			GitopsRef:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			GitopsRepo:  "testrepo",
		},
	}

	discordMessageSendFailure, err := msgSendFailure.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageSendFailure.Text, "Failed to roll out") {
		t.Errorf("Gitops reploy failure message must contain 'Failed to roll back'")
	}

	msgSendByGimlet := gitopsDeployMessage{
		event: &events.DeployEvent{
			Manifest: &dx.Manifest{
				App: "myapp",
				Env: "staging",
			},
			Artifact: &dx.Artifact{
				Version: *version,
			},
			TriggeredBy: "Gimlet",
			Status:      0,
			StatusDesc:  "",
			GitopsRef:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			GitopsRepo:  "testrepo",
		},
	}

	discordMessageSendByGimlet, err := msgSendByGimlet.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageSendByGimlet.Text, "is rolling out") {
		t.Errorf("Gitops deploy message which triggered by someone, must contain 'is rolling out'")
	}

}

func TestSendingGitopsRollbackMessage(t *testing.T) {

	msgRollbackFailed := gitopsRollbackMessage{
		event: &events.RollbackEvent{
			RollbackRequest: &dx.RollbackRequest{
				Env:         "staging",
				App:         "myapp",
				TargetSHA:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
				TriggeredBy: "Gimlet",
			},
			Status:     1,
			StatusDesc: "success",
			GitopsRefs: []string{"76ab7d611242f7c6742f0ab662133e02b2ba2b1c", "76ab7d611242f7c6742f0ab662133e02b2ba2bbb", "76ab7d611242f7c6742f0ab662133e02b2ba2lll"},
			GitopsRepo: "gimlet-io",
		},
	}

	discordMessageRollbackFailed, err := msgRollbackFailed.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageRollbackFailed.Text, "Failed to roll back") {
		t.Errorf("Rollback failed message must contain 'Failed to roll back'")
	}

	msgRollbackSuccess := gitopsRollbackMessage{
		event: &events.RollbackEvent{
			RollbackRequest: &dx.RollbackRequest{
				Env:         "staging",
				App:         "myapp",
				TargetSHA:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
				TriggeredBy: "Gimlet",
			},
			Status:     0,
			StatusDesc: "success",
			GitopsRefs: []string{"76ab7d611242f7c6742f0ab662133e02b2ba2b1c", "76ab7d611242f7c6742f0ab662133e02b2ba2bbb", "76ab7d611242f7c6742f0ab662133e02b2ba2lll"},
			GitopsRepo: "gimlet-io",
		},
	}

	discordMessageRollbackSuccess, err := msgRollbackSuccess.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageRollbackSuccess.Text, "is rolling back") {
		t.Errorf("Rollback success message must contain 'is rolling back'")
	}

}
