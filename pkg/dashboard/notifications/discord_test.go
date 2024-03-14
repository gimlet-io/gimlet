package notifications

import (
	"strings"
	"testing"

	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dx"
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
		result: model.Result{
			Manifest: &dx.Manifest{
				Env: "staging",
				App: "myapp",
			},
			TriggeredBy: "Gimlet",
			StatusDesc:  "cannot delete",
			GitopsRef:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			GitopsRepo:  "testrepo",
			Status:      model.Failure,
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
		result: model.Result{
			Manifest: &dx.Manifest{
				Env: "staging",
				App: "myapp",
			},
			TriggeredBy: "policy",
			StatusDesc:  "cannot delete",
			GitopsRef:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			GitopsRepo:  "testrepo",
			Status:      model.Success,
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
		event: model.Result{
			Manifest: &dx.Manifest{
				App: "myapp",
				Env: "staging",
			},
			Artifact: &dx.Artifact{
				Version: *version,
			},
			TriggeredBy: "Gimlet",
			Status:      model.Failure,
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
		event: model.Result{
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

	msgRollback := gitopsRollbackMessage{
		event: model.Event{
			Status:     model.Failure.String(),
			StatusDesc: "failed",
			Results: []model.Result{
				{
					GitopsRef:  "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
					GitopsRepo: "gimlet-io",
				},
				{
					GitopsRef:  "76ab7d611242f7c6742f0ab662133e02b2ba2bbb",
					GitopsRepo: "gimlet-io",
				},
				{
					GitopsRef:  "76ab7d611242f7c6742f0ab662133e02b2ba2lll",
					GitopsRepo: "gimlet-io",
				},
			},
		},
		rollbackRequest: dx.RollbackRequest{
			Env:         "staging",
			App:         "myapp",
			TargetSHA:   "76ab7d611242f7c6742f0ab662133e02b2ba2b1c",
			TriggeredBy: "Gimlet",
		},
	}

	msgRollback.event.Status = model.Success.String()
	discordMessageRollbackSuccess, err := msgRollback.AsDiscordMessage()
	if err != nil {
		t.Errorf("Failed to create Discord message!")
	}

	if !strings.Contains(discordMessageRollbackSuccess.Text, "is rolling back") {
		t.Errorf("Rollback success message must contain 'is rolling back'")
	}

}
