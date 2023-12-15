package notifications

type Message interface {
	AsSlackMessage() (*slackMessage, error)
	AsStatus() (*status, error)
	AsDiscordMessage() (*discordMessage, error)
	Env() string
	RepositoryName() string
	SHA() string
	CustomChannel() string
	Silenced() bool
}

type status struct {
	state       string
	context     string
	description string
	repo        string
	sha         string
}
