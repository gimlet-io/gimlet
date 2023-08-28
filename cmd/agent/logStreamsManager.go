package main

import "sync"

type logStreamsManager struct {
	runningLogStreams map[string]chan int
	lock              sync.Mutex
}

func NewLogStreamsManager() *logStreamsManager {
	return &logStreamsManager{
		runningLogStreams: make(map[string]chan int),
	}
}

func (l *logStreamsManager) Open(channel chan int, namespace string, serviceName string) {
	pod := namespace + "/" + serviceName

	l.lock.Lock()
	l.runningLogStreams[pod] = channel
	l.lock.Unlock()
}

func (l *logStreamsManager) Stop(namespace string, serviceName string) {
	l.lock.Lock()
	for svc, stopCh := range l.runningLogStreams {
		if svc == namespace+"/"+serviceName {
			stopCh <- 0
		}
	}
	l.lock.Unlock()
}

func (l *logStreamsManager) StopAll() {
	l.lock.Lock()
	for _, stopCh := range l.runningLogStreams {
		stopCh <- 0
	}
	l.lock.Unlock()
}
