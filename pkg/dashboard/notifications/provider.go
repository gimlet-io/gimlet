package notifications

type Provider interface {
	send(msg Message) error
}
