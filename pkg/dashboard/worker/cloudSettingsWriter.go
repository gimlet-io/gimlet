package worker

import (
	"time"

	"github.com/gimlet-io/gimlet/cmd/dashboard/config"
	"github.com/gimlet-io/gimlet/pkg/dashboard/model"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server"
	"github.com/gimlet-io/gimlet/pkg/dashboard/server/streaming"
	"github.com/gimlet-io/gimlet/pkg/dashboard/store"
	"github.com/gimlet-io/gimlet/pkg/git/customScm"
	"github.com/gimlet-io/gimlet/pkg/git/nativeGit"
	"github.com/sirupsen/logrus"
)

type CloudSettingsWriter struct {
	dao          *store.Store
	gitRepoCache *nativeGit.RepoCache
	tokenManager customScm.NonImpersonatedTokenManager
	gitUser      *model.User
	config       *config.Config
	agentHub     *streaming.AgentHub
}

func NewCloudSettingsWriter(
	dao *store.Store,
	gitRepoCache *nativeGit.RepoCache,
	tokenManager customScm.NonImpersonatedTokenManager,
	gitUser *model.User,
	config *config.Config,
	agentHub *streaming.AgentHub,
) *CloudSettingsWriter {
	return &CloudSettingsWriter{
		dao:          dao,
		gitRepoCache: gitRepoCache,
		tokenManager: tokenManager,
		gitUser:      gitUser,
		config:       config,
		agentHub:     agentHub,
	}
}

func (c *CloudSettingsWriter) Run() {
	if c.config.Instance == "" {
		return
	}

	for {
		_, err := c.dao.KeyValue(model.EnsuredCustomRegistry)
		if err == nil { // gimlet cloud settings set already
			logrus.Info("Gimlet Cloud settings set")
			return
		}

		for _, agent := range c.agentHub.Agents {
			if string(agent.SealedSecretsCertificate) != "" {
				logrus.Info("Ensuring Gimlet Cloud settings")
				server.EnsureGimletCloudSettings(c.dao, c.gitRepoCache, c.tokenManager, c.gitUser, agent.Name, agent.SealedSecretsCertificate, c.config.Instance)
			}
		}

		c.agentHub.ForceStateSend()
		time.Sleep(3 * time.Second)
	}
}
