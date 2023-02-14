package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	eventsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gimletd_event_processed_total",
		Help: "The total number of processed events",
	})

	releases = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gimletd_releases",
		Help: "Release status",
	}, []string{"env", "app", "sourceCommit", "commitMessage", "gitopsCommit", "gitopsCommitCreated"})

	perf = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gimletd_perf",
		Help: "Performance of functions",
	}, []string{"function"})
)
