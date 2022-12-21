package notifications

type Message interface {
	AsSlackMessage() (*slackMessage, error)
	AsStatus() (*status, error)
	AsDiscordMessage() (*discordMessage, error)
	Env() string
	RepositoryName() string
	SHA() string
}

type status struct {
	state       string
	context     string
	description string
	targetURL   string
}
