package main

import "sync"

type runningLogStreams struct {
	runningLogStreams map[string]chan int
	lock              sync.Mutex
}

func NewRunningLogStreams() *runningLogStreams {
	return &runningLogStreams{
		runningLogStreams: make(map[string]chan int),
	}
}

func (l *runningLogStreams) Regsiter(channel chan int, namespace string, serviceName string) {
	pod := namespace + "/" + serviceName

	l.lock.Lock()
	l.runningLogStreams[pod] = channel
	l.lock.Unlock()
}

func (l *runningLogStreams) Stop(namespace string, serviceName string) {
	l.lock.Lock()
	for svc, stopCh := range l.runningLogStreams {
		if svc == namespace+"/"+serviceName {
			stopCh <- 0
		}
	}
	l.lock.Unlock()
}

func (l *runningLogStreams) StopAll() {
	l.lock.Lock()
	for _, stopCh := range l.runningLogStreams {
		stopCh <- 0
	}
	l.lock.Unlock()
}
