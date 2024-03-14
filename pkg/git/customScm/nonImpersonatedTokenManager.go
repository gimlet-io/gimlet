package customScm

import (
	"github.com/gimlet-io/gimlet/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet/pkg/git/customScm/customGithub"
	"github.com/gimlet-io/gimlet/pkg/git/customScm/customGitlab"
)

type NonImpersonatedTokenManager interface {
	Token() (string, string, error)
	AppToken() (string, error)
}

type TokenManager struct {
	tokenManagerImpl NonImpersonatedTokenManager
}

func NewTokenManager(dc *dynamicconfig.DynamicConfig) NonImpersonatedTokenManager {
	tm := &TokenManager{}
	tm.Configure(dc)
	return tm
}

func (t *TokenManager) Configure(dc *dynamicconfig.DynamicConfig) {
	if dc.IsGithub() {
		var err error
		t.tokenManagerImpl, err = customGithub.NewGithubOrgTokenManager(
			dc.Github.AppID,
			dc.Github.InstallationID,
			dc.Github.PrivateKey.String(),
		)
		if err != nil {
			panic(err)
		}
	} else if dc.IsGitlab() {
		t.tokenManagerImpl = customGitlab.NewGitlabTokenManager(dc.Gitlab.AdminToken)
	} else {
		t.tokenManagerImpl = NewDummyTokenManager()
	}
}

func (t *TokenManager) Token() (string, string, error) {
	return t.tokenManagerImpl.Token()
}

func (t *TokenManager) AppToken() (string, error) {
	return t.tokenManagerImpl.AppToken()
}
