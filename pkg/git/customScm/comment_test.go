package customScm

import (
	"fmt"
	"testing"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
)

const (
	bodyProcessing = `### <span aria-hidden="true">👷</span> Deploy Preview for *%s* processing.

| Name | Link |
|:-:|------------------------|
|<span aria-hidden="true">🔨</span> Latest commit | %s |
|<span aria-hidden="true">🔍</span> Latest deploy log | %s |
	`

	bodyReady = `### <span aria-hidden="true">✅</span> Deploy Preview for *%s* ready!

| Name | Link |
|:-:|------------------------|
|<span aria-hidden="true">🔨</span> Latest commit | %s |
|<span aria-hidden="true">🔍</span> Latest deploy log | %s |
|<span aria-hidden="true">😎</span> Deploy Preview | %s |
`

	bodyFailed = `### <span aria-hidden="true">❌</span> Deploy Preview for *%s* failed.

|  Name | Link |
|:-:|------------------------|
|<span aria-hidden="true">🔨</span> Latest commit | %s |
|<span aria-hidden="true">🔍</span> Latest deploy log | %s |
`
)

func TestComment(t *testing.T) {
	owner := "gimlet-io"
	repo := "getting-started-app"
	appName := "preview-app"
	hash := "abc123"
	pullNumber := 0
	logUrl := "https://gimlet.io"
	previewUrl := "https://gimlet.io"
	processing := fmt.Sprintf(bodyProcessing, appName, hash, logUrl)
	ready := fmt.Sprintf(bodyReady, appName, hash, logUrl, previewUrl)
	failed := fmt.Sprintf(bodyFailed, appName, hash, logUrl)
	config := &dynamicconfig.DynamicConfig{
		Github: config.Github{
			AppID:          "",
			InstallationID: "",
			PrivateKey:     "",
		},
	}

	tokenManager := NewTokenManager(config)
	token, _, _ := tokenManager.Token()
	gitSvc := NewGitService(config)

	commentId, err := gitSvc.CreateComment(token, owner, repo, pullNumber, &processing)
	if err != nil {
		panic(err)
	}

	err = gitSvc.UpdateComment(token, owner, repo, commentId, &ready)
	if err != nil {
		panic(err)
	}

	err = gitSvc.UpdateComment(token, owner, repo, commentId, &failed)
	if err != nil {
		panic(err)
	}
}
