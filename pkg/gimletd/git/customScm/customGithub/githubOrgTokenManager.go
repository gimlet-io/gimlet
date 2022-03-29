package customGithub

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/gimletd/config"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v37/github"
	"github.com/sirupsen/logrus"
)

// GithubOrgTokenManager maintains a valid git org/non-impersonated token
type GithubOrgTokenManager struct {
	appId          string
	privateKey     string
	installationId int64

	orgUser   string
	orgToken  string
	expiresAt *time.Time
}

func NewGithubOrgTokenManager(config *config.Config) (*GithubOrgTokenManager, error) {
	installID, err := strconv.ParseInt(config.Github.InstallationID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse installationId: %s", err)
	}

	manager := &GithubOrgTokenManager{
		appId:          config.Github.AppID,
		privateKey:     config.Github.PrivateKey.String(),
		installationId: installID,
	}
	err = manager.refreshOrgToken()
	if err != nil {
		return nil, fmt.Errorf("could refresh org token: %s", err)
	}

	go manager.refreshLoop()

	return manager, nil
}

// Token returns a valid cached orgToken
func (tm *GithubOrgTokenManager) Token() (string, string, error) {
	if tm.orgToken == "" {
		return "", "", fmt.Errorf("no valid orgToken available")
	}
	return tm.orgToken, tm.orgUser, nil
}

func (tm *GithubOrgTokenManager) refreshLoop() {
	for {
		time.Sleep(5 * time.Minute)
		if tm.orgToken != "" && tm.expiresAt != nil && tm.expiresAt.Before(time.Now().Add(time.Minute*10)) {
			err := tm.refreshOrgToken()
			if err != nil {
				logrus.Errorf("could not refresh orgToken %v", err)
				continue
			}
		}
	}
}

func (tm *GithubOrgTokenManager) refreshOrgToken() error {
	installationToken, err := tm.installationToken()
	if err != nil {
		return err
	}

	tm.orgUser = "abc123"
	tm.orgToken = *installationToken.Token
	tm.expiresAt = installationToken.ExpiresAt

	return nil
}

func (tm *GithubOrgTokenManager) installationToken() (*github.InstallationToken, error) {
	appToken, err := tm.appToken()
	if err != nil {
		return nil, err
	}

	client := github.NewClient(&http.Client{Transport: &transport{underlyingTransport: http.DefaultTransport, token: appToken}})

	token, _, err := client.Apps.CreateInstallationToken(context.Background(), tm.installationId, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// appToken returns a signed JWT apptoken for the Github app
// Use it to have unimpersonated access of Github resources
func (tm *GithubOrgTokenManager) appToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Local().Add(time.Minute * 5).Unix(),
		"iss": tm.appId,
	})

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(tm.privateKey))
	if err != nil {
		return "", err
	}

	tokenString, _ := token.SignedString(signKey)
	return tokenString, nil
}

type transport struct {
	underlyingTransport http.RoundTripper
	token               string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.token)
	return t.underlyingTransport.RoundTrip(req)
}
