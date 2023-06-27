package customScm

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
