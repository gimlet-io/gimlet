package genericScm

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/config"
	"github.com/gimlet-io/go-scm/scm"
	"github.com/gimlet-io/go-scm/scm/driver/github"
	"github.com/gimlet-io/go-scm/scm/transport/oauth2"
	"github.com/sirupsen/logrus"
)

type GoScmHelper struct {
	client *scm.Client
}

func NewGoScmHelper(config *config.Config, tokenUpdateCallback func(token *scm.Token)) *GoScmHelper {
	client, err := github.New("https://api.github.com")
	if err != nil {
		logrus.WithError(err).
			Fatalln("main: cannot create the GitHub client")
	}
	if config.Github.Debug {
		client.DumpResponse = httputil.DumpResponse
	}

	client.Client = &http.Client{
		Transport: &oauth2.Transport{
			Source: &Refresher{
				ClientID:     config.Github.ClientID,
				ClientSecret: config.Github.ClientSecret,
				Endpoint:     "https://github.com/login/oauth/access_token",
				Source:       oauth2.ContextTokenSource(),
				tokenUpdater: tokenUpdateCallback,
			},
		},
	}

	return &GoScmHelper{
		client: client,
	}
}

// defaultTransport provides a default http.Transport. If
// skipVerify is true, the transport will skip ssl verification.
func defaultTransport(skipVerify bool) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipVerify,
		},
	}
}

func (helper *GoScmHelper) Parse(req *http.Request, fn scm.SecretFunc) (scm.Webhook, error) {
	return helper.client.Webhooks.Parse(req, fn)
}

func (helper *GoScmHelper) UserRepos(accessToken string, refreshToken string, expires time.Time) ([]string, error) {
	var repos []string

	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   accessToken,
		Refresh: refreshToken,
		Expires: expires,
	})

	opts := scm.ListOptions{Size: 100}
	for {
		scmRepos, meta, err := helper.client.Repositories.List(ctx, opts)
		if err != nil {
			return []string{}, err
		}
		for _, repo := range scmRepos {
			repos = append(repos, repo.Namespace+"/"+repo.Name)
		}

		opts.Page = meta.Page.Next
		opts.URL = meta.Page.NextURL

		if opts.Page == 0 && opts.URL == "" {
			break
		}
	}

	return repos, nil
}

func (helper *GoScmHelper) User(accessToken string, refreshToken string) (*scm.User, error) {
	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   accessToken,
		Refresh: refreshToken,
	})
	user, _, err := helper.client.Users.Find(ctx)
	return user, err
}

func (helper *GoScmHelper) Organizations(accessToken string, refreshToken string) ([]*scm.Organization, error) {
	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   accessToken,
		Refresh: refreshToken,
	})
	organizations, _, err := helper.client.Organizations.List(ctx, scm.ListOptions{
		Size: 50,
	})

	return organizations, err
}

func (helper *GoScmHelper) CreatePR(
	accessToken string,
	repoPath string,
	sourceBranch string,
	targetBranch string,
) error {
	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   accessToken,
		Refresh: "",
	})

	prSubject := "Title of the PR"
	prDescription := "Description of the PR"
	newPR := &scm.PullRequestInput{
		Title:  prSubject,
		Body:   prDescription,
		Source: sourceBranch,
		Target: targetBranch,
	}

	_, _, err := helper.client.PullRequests.Create(ctx, repoPath, newPR)

	return err
}

func (helper *GoScmHelper) Content(accessToken string, repo string, path string, branch string) (string, string, error) {
	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   accessToken,
		Refresh: "",
	})
	content, _, err := helper.client.Contents.Find(
		ctx,
		repo,
		path,
		branch)

	return string(content.Data), string(content.BlobID), err
}

func (helper *GoScmHelper) CreateContent(
	accessToken string,
	repo string,
	path string,
	content []byte,
	branch string,
	message string,
) error {
	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   accessToken,
		Refresh: "",
	})
	_, err := helper.client.Contents.Create(
		ctx,
		repo,
		path,
		&scm.ContentParams{
			Data:    content,
			Branch:  branch,
			Message: message,
			Signature: scm.Signature{
				Name:  "Gimlet",
				Email: "gimlet-dashboard@gimlet.io",
			},
		})

	return err
}

func (helper *GoScmHelper) UpdateContent(
	accessToken string,
	repo string,
	path string,
	content []byte,
	blobID string,
	branch string,
	message string,
) error {
	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   accessToken,
		Refresh: "",
	})
	_, err := helper.client.Contents.Update(
		ctx,
		repo,
		path,
		&scm.ContentParams{
			Data:    content,
			Message: message,
			Branch:  branch,
			BlobID:  blobID,
			Signature: scm.Signature{
				Name:  "Gimlet",
				Email: "gimlet-dashboard@gimlet.io",
			},
		})

	return err
}

// DirectoryContents returns a map of file paths as keys and their file contents in the values
func (helper *GoScmHelper) DirectoryContents(accessToken string, repo string, directoryPath string) (map[string]string, error) {
	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   accessToken,
		Refresh: "",
	})
	directoryFiles, _, err := helper.client.Contents.List(
		ctx,
		repo,
		directoryPath,
		"HEAD",
		scm.ListOptions{
			Size: 50,
		},
	)

	files := map[string]string{}
	for _, file := range directoryFiles {
		files[file.Path] = file.BlobID
	}

	return files, err
}

func (helper *GoScmHelper) RegisterWebhook(
	host string,
	token string,
	webhookSecret string,
	owner string,
	repo string,
) error {
	ctx := context.WithValue(context.Background(), scm.TokenKey{}, &scm.Token{
		Token:   token,
		Refresh: "",
	})

	hook := &scm.HookInput{
		Name:   "Gimlet Dashboard",
		Target: host + "/hook",
		Secret: webhookSecret,
		Events: scm.HookEvents{
			Push:   true,
			Status: true,
			Branch: true,
			//CheckRun: true,
		},
	}

	return replaceHook(ctx, helper.client, scm.Join(owner, repo), hook)
}
