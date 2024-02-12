package customGithub

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gimlet-io/gimlet-cli/cmd/dashboard/dynamicconfig"
	"github.com/gimlet-io/gimlet-cli/pkg/dashboard/model"
	"github.com/google/go-github/v37/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubClient struct {
}

// FetchCommits fetches Github commits and their statuses
// Only the following fields are used:
// url, author, author_pic, message, created, tags, status
/* Getting multiple commits by hash
query {
  viewer {
    login
  }
  rateLimit {
    limit
    cost
    remaining
    resetAt
  }
  repository(owner: "laszlocph", name: "aedes") {
     a: object(oid: "25a913a5e052d3f5b9c4880377542f3ed8389d2b") {
      ... on Commit {
        oid
        message
        authoredDate
        status {
          state
          contexts {
            context
            createdAt
            state
            targetUrl
          }
        }
      }
    }
    b: object(oid: "3396bc4fae754b5f55de23f49f973ddca70295d7") {
      ... on Commit {
        oid
        message
        authoredDate
        status {
          state
          contexts {
            context
            createdAt
            state
            targetUrl
          }
        }
        checkSuites(first: 100){
          nodes {
            checkRuns (first: 100) {
              nodes {
                permalink
                name
                status
                startedAt
                completedAt
              }
            }
          }
        }
        statusCheckRollup{
          state
          contexts(first: 100) {
            nodes {
              __typename
              ... on CheckRun {
                name
                detailsUrl
                completedAt
                status
              }
              ... on StatusContext {
                context
                createdAt
                state
                targetUrl
              }
            }
          }
        }
      }
    }
  }
}
*/
func (c *GithubClient) FetchCommits(
	owner string,
	repo string,
	token string,
	hashesToFetch []string,
) ([]*model.Commit, error) {
	if len(hashesToFetch) > 10 {
		return nil, fmt.Errorf("can only fetch 10 commits at a time")
	}

	// since the query takes 10 hashes
	// we pad it with the first hash
	// getting that multiple times in the result set
	// should be idempotent
	toPadWidth := 10 - len(hashesToFetch)
	if len(hashesToFetch) < 10 {
		for i := 0; i < toPadWidth; i++ {
			hashesToFetch = append(hashesToFetch, hashesToFetch[0])
		}
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	graphQLClient := githubv4.NewClient(httpClient)

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(repo),
		"sha0":  githubv4.GitObjectID(hashesToFetch[0]),
		"sha1":  githubv4.GitObjectID(hashesToFetch[1]),
		"sha2":  githubv4.GitObjectID(hashesToFetch[2]),
		"sha3":  githubv4.GitObjectID(hashesToFetch[3]),
		"sha4":  githubv4.GitObjectID(hashesToFetch[4]),
		"sha5":  githubv4.GitObjectID(hashesToFetch[5]),
		"sha6":  githubv4.GitObjectID(hashesToFetch[6]),
		"sha7":  githubv4.GitObjectID(hashesToFetch[7]),
		"sha8":  githubv4.GitObjectID(hashesToFetch[8]),
		"sha9":  githubv4.GitObjectID(hashesToFetch[9]),
	}

	var query queryObjects

	// q, _ := json.Marshal(query)
	// logrus.Infof("Github query: %s", q)
	// logrus.Infof("Github variables: %s", variables)
	err := graphQLClient.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, err
	}

	var commits []*model.Commit
	commits = append(commits, translateCommit(query.Repository.Object0.Commit))
	commits = append(commits, translateCommit(query.Repository.Object1.Commit))
	commits = append(commits, translateCommit(query.Repository.Object2.Commit))
	commits = append(commits, translateCommit(query.Repository.Object3.Commit))
	commits = append(commits, translateCommit(query.Repository.Object4.Commit))
	commits = append(commits, translateCommit(query.Repository.Object5.Commit))
	commits = append(commits, translateCommit(query.Repository.Object6.Commit))
	commits = append(commits, translateCommit(query.Repository.Object7.Commit))
	commits = append(commits, translateCommit(query.Repository.Object8.Commit))
	commits = append(commits, translateCommit(query.Repository.Object9.Commit))

	// response, _ := json.Marshal(query)
	// logrus.Infof("Github response: %s", response)

	return commits[:10-toPadWidth], nil
}

func translateCommit(commit commit) *model.Commit {
	var contexts []model.CommitStatus
	for _, c := range commit.Status.Contexts {
		contexts = append(contexts, model.CommitStatus{
			State:       c.State,
			Context:     c.Context,
			CreatedAt:   c.CreatedAt,
			TargetUrl:   c.TargetUrl,
			Description: c.Description,
		})
	}

	for _, checkSuite := range commit.CheckSuits.Nodes {
		for _, checkRun := range checkSuite.CheckRuns.Nodes {
			status := checkRun.Status
			if checkRun.Conclusion != "" {
				status = checkRun.Conclusion
			}
			contexts = append(contexts, model.CommitStatus{
				State:     status,
				Context:   checkRun.Name,
				CreatedAt: checkRun.CompletedAt,
				TargetUrl: checkRun.Permalink,
			})
		}
	}

	return &model.Commit{
		SHA:       commit.OID,
		Message:   commit.Message,
		Author:    commit.Author.User.Login,
		AuthorPic: commit.Author.User.AvatarURL,
		URL:       commit.URL,
		Status: model.CombinedStatus{
			State:    commit.Status.State,
			Contexts: contexts,
		},
	}
}

type ctx struct {
	Context     string
	CreatedAt   string
	State       string
	TargetUrl   string
	Description string
}

type commit struct {
	URL     string
	OID     string
	Message string
	Author  struct {
		User struct {
			Login     string
			AvatarURL string
		}
	}
	Status struct {
		State    string
		Contexts []ctx
	}
	CheckSuits struct {
		Nodes []CheckSuite
	} `graphql:"checkSuites(first: 100)"`
}

type obj struct {
	Commit commit `graphql:"... on Commit"`
}

type CheckSuite struct {
	CheckRuns checkRuns `graphql:"checkRuns (first: 100)"`
}

type checkRuns struct {
	Nodes []checkRun
}

type checkRun struct {
	Permalink   string
	Name        string
	Status      string
	Conclusion  string
	StartedAt   string
	CompletedAt string
}

type queryObjects struct {
	Viewer struct {
		Login string
	}
	RateLimit struct {
		Limit     int
		Cost      int
		Remaining int
		ResetAt   string
	}
	Repository struct {
		Object0 obj `graphql:"obj0: object(oid: $sha0)"`
		Object1 obj `graphql:"obj1: object(oid: $sha1)"`
		Object2 obj `graphql:"obj2: object(oid: $sha2)"`
		Object3 obj `graphql:"obj3: object(oid: $sha3)"`
		Object4 obj `graphql:"obj4: object(oid: $sha4)"`
		Object5 obj `graphql:"obj5: object(oid: $sha5)"`
		Object6 obj `graphql:"obj6: object(oid: $sha6)"`
		Object7 obj `graphql:"obj7: object(oid: $sha7)"`
		Object8 obj `graphql:"obj8: object(oid: $sha8)"`
		Object9 obj `graphql:"obj9: object(oid: $sha9)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// OrgRepos returns all repos of an org using the installation
func (c *GithubClient) OrgRepos(installationToken string) ([]string, error) {
	client := github.NewClient(
		&http.Client{
			Transport: &transport{
				underlyingTransport: http.DefaultTransport,
				token:               installationToken,
			},
		},
	)

	opt := &github.ListOptions{PerPage: 100}
	var allRepos []string
	for {
		repos, resp, err := client.Apps.ListRepos(context.Background(), opt)
		if err != nil {
			return nil, err
		}

		for _, r := range repos.Repositories {
			repo := *r.Owner.Login + "/" + *r.Name
			allRepos = append(allRepos, repo)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (c *GithubClient) GetAppNameAndAppSettingsURLs(appToken string, ctx context.Context) (string, string, string, error) {
	dynamicConfig := ctx.Value("dynamicConfig").(*dynamicconfig.DynamicConfig)

	client := github.NewClient(
		&http.Client{
			Transport: &transport{
				underlyingTransport: http.DefaultTransport,
				token:               appToken,
			},
		},
	)

	appinfo, _, err := client.Apps.Get(ctx, "")
	if err != nil {
		return "", "", "", fmt.Errorf("cannot get info from App : %s", err)
	}

	installationID := dynamicConfig.Github.InstallationID
	installationIDint, err := strconv.ParseInt(installationID, 0, 64)
	if err != nil {
		return "", "", "", fmt.Errorf("cannot parse App Token : %s", err)
	}

	installation, _, err := client.Apps.GetInstallation(ctx, installationIDint)
	if err != nil {
		return "", "", "", fmt.Errorf("cannot get installation: %s", err)
	}

	return *appinfo.Name, *appinfo.HTMLURL, *installation.HTMLURL, err
}

func (c *GithubClient) GetAppOwner(appToken string) (string, error) {
	client := github.NewClient(
		&http.Client{
			Transport: &transport{
				underlyingTransport: http.DefaultTransport,
				token:               appToken,
			},
		},
	)

	appinfo, _, err := client.Apps.Get(context.Background(), "")
	if err != nil {
		return "", fmt.Errorf("cannot get info from App : %s", err)
	}

	return *appinfo.Owner.Login, err
}

func (c *GithubClient) GetUserPrimaryEmail(appToken string) (string, error) {
	client := github.NewClient(
		&http.Client{
			Transport: &transport{
				underlyingTransport: http.DefaultTransport,
				token:               appToken,
			},
		},
	)

	var primaryUserEmail string
	githubUserMails, _, err := client.Users.ListEmails(context.Background(), &github.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, githubUserMail := range githubUserMails {
		if *githubUserMail.Primary {
			primaryUserEmail = githubUserMail.GetEmail()
		}
	}

	return primaryUserEmail, err
}

func (c *GithubClient) CreateRepository(owner string, repo string, loggedInUser string, orgToken string, userToken string) error {
	token := orgToken
	if owner == loggedInUser {
		owner = "" // if the repo is not an org repo, but the logged in user's, the Github API doesn't need an org
		token = userToken
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	var (
		name     = repo
		private  = true
		autoInit = true
	)

	r := &github.Repository{
		Name:     &name,
		Private:  &private,
		AutoInit: &autoInit,
	}
	_, _, err := client.Repositories.Create(context.Background(), owner, r)
	return err
}

func (c *GithubClient) AddDeployKeyToRepo(owner, repo, token, keyTitle, keyValue string, canWrite bool) error {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	keys, _, err := client.Repositories.ListKeys(context.Background(), owner, repo, nil)
	if err != nil {
		return err
	}
	if len(keys) != 0 {
		return nil
	}

	readOnly := !canWrite
	_, _, err = client.Repositories.CreateKey(context.Background(), owner, repo, &github.Key{
		Title:    &keyTitle,
		Key:      &keyValue,
		ReadOnly: &readOnly,
	})
	return err
}

func (c *GithubClient) CreateComment(token, owner, repo string, pullNumber int, body *string) (int64, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	comment, _, err := client.Issues.CreateComment(context.Background(), owner, repo, pullNumber, &github.IssueComment{
		Body: body,
	})

	return *comment.ID, err
}

func (c *GithubClient) UpdateComment(token, owner, repo string, commentId int64, body *string) error {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	_, _, err := client.Issues.EditComment(context.Background(), owner, repo, commentId, &github.IssueComment{
		Body: body,
	})

	return err
}
