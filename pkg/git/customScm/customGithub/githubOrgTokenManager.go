package customGithub

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

func NewGithubOrgTokenManager(appId string, installationId string, privateKey string) (*GithubOrgTokenManager, error) {
	installID, err := strconv.ParseInt(installationId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse installationId: %s", err)
	}

	manager := &GithubOrgTokenManager{
		appId:          appId,
		privateKey:     privateKey,
		installationId: installID,
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
		if tm.orgToken == "" ||
			(tm.orgToken != "" && tm.expiresAt != nil && tm.expiresAt.Before(time.Now().Add(time.Minute*10))) {
			err := tm.refreshOrgToken()
			if err != nil {
				logrus.Errorf("could not refresh orgToken %v", err)
			}
		}
		time.Sleep(5 * time.Minute)
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
	appToken, err := tm.AppToken()
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
func (tm *GithubOrgTokenManager) AppToken() (string, error) {
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
