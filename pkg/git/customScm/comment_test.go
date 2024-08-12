package customScm

import (
	"fmt"
	"testing"

	"github.com/gimlet-io/gimlet/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
)

func TestComment(t *testing.T) {
	repoName := "gimlet-io/getting-started-app"
	appName := "preview-app"
	hash := "abc123"
	pullNumber := 0
	commentId := int64(0)
	previewUrl := "https://gimlet.io"
	processing := fmt.Sprintf(BodyProcessing, appName, hash)
	ready := fmt.Sprintf(BodyReady, appName, hash, previewUrl, previewUrl)
	failed := fmt.Sprintf(BodyFailed, appName, hash)
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

	err := gitSvc.CreateComment(token, repoName, pullNumber, processing)
	if err != nil {
		panic(err)
	}

	err = gitSvc.UpdateComment(token, repoName, commentId, ready)
	if err != nil {
		panic(err)
	}

	err = gitSvc.UpdateComment(token, repoName, commentId, failed)
	if err != nil {
		panic(err)
	}
}
