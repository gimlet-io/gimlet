package notifications

import (
	"github.com/sirupsen/logrus"
)

type Manager interface {
	Broadcast(msg Message)
	AddProvider(provider Provider)
}

type ManagerImpl struct {
	provider  []Provider
	broadcast chan Message
}

type DummyManagerImpl struct {
}

func NewManager() *ManagerImpl {
	return &ManagerImpl{
		provider:  []Provider{},
		broadcast: make(chan Message),
	}
}

func NewDummyManager() *DummyManagerImpl {
	return &DummyManagerImpl{}
}

func (m *ManagerImpl) Broadcast(msg Message) {
	if msg.Silenced() {
		return
	}

	m.broadcast <- msg
}

func (m *DummyManagerImpl) Broadcast(msg Message) {
}

func (m *DummyManagerImpl) AddProvider(provider Provider) {
}

func (m *ManagerImpl) AddProvider(provider Provider) {
	m.provider = append(m.provider, provider)
}

func (m *ManagerImpl) Run() {
	for {
		select {
		case message := <-m.broadcast:
			for _, p := range m.provider {
				go func(p Provider) {
					err := p.send(message)
					if err != nil {
						logrus.Warnf("cannot send notification: %s ", err)
					}
				}(p)
			}
		}
	}
}
