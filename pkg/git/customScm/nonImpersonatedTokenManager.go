package customScm

type NonImpersonatedTokenManager interface {
	Token() (string, string, error)
	AppToken() (string, error)
}

type DummyTokenManager struct {
}

func NewDummyTokenManager() *DummyTokenManager {
	return &DummyTokenManager{}
}

func (t *DummyTokenManager) Token() (string, string, error) {
	return "", "", nil
}

func (t *DummyTokenManager) AppToken() (string, error) {
	return "", nil
}
