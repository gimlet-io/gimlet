package server

import (
	"time"
)

type threshold interface {
	isFired() bool
}

type podStrategy struct {
	timestamp int64
	waitTime  time.Duration
}

type eventStrategy struct {
	timestamp              int64
	count                  int32
	expectedCountPerMinute float64
	expectedCount          int32
}

func (s podStrategy) isFired() bool {
	podAlertTime := time.Unix(s.timestamp, 0)
	managerWaitTime := time.Now().Add(-time.Minute * s.waitTime)

	return podAlertTime.Before(managerWaitTime)
}

func (s eventStrategy) isFired() bool {
	firstTimestampSinceInMinutes := time.Since(time.Unix(s.timestamp, 0)).Minutes()
	countPerMinute := float64(s.count) / firstTimestampSinceInMinutes

	return countPerMinute >= s.expectedCountPerMinute && s.count >= s.expectedCount
}
