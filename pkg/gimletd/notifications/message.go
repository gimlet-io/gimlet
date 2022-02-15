package notifications

import githubLib "github.com/google/go-github/v37/github"

type Message interface {
	AsSlackMessage() (*slackMessage, error)
	AsGithubStatus() (*githubLib.RepoStatus, error)
	AsDiscordMessage() (*discordMessage, error)
	Env() string
	RepositoryName() string
	SHA() string
}
