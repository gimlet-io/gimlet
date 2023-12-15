package notifications

type Message interface {
	AsSlackMessage() (*slackMessage, error)
	AsStatus() (*status, error)
	AsDiscordMessage() (*discordMessage, error)
	Env() string
	RepositoryName() string
	SHA() string
	CustomChannel() string
}

type status struct {
	state       string
	context     string
	description string
	repo        string
	sha         string
}
