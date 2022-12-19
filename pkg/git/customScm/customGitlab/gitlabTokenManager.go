package customGitlab

type GitlabTokenManager struct {
	adminToken string
}

func NewGitlabTokenManager(adminToken string) *GitlabTokenManager {
	return &GitlabTokenManager{
		adminToken: adminToken,
	}
}

func (tm *GitlabTokenManager) Token() (string, string, error) {
	return tm.adminToken, "oauth2", nil
}

func (tm *GitlabTokenManager) AppToken() (string, error) {
	return tm.adminToken, nil
}
