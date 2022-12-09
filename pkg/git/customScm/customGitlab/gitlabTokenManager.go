package customGitlab

type GitlabTokenManager struct {
}

func (tm *GitlabTokenManager) Token() (string, string, error) {
	return "", "", nil
}

func (tm *GitlabTokenManager) AppToken() (string, error) {
	return "", nil
}
