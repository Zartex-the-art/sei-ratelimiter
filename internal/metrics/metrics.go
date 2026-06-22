package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var RequestsTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total requests",
	},
)

var RequestsAllowed = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "requests_allowed_total",
		Help: "Allowed requests",
	},
)

var RequestsBlocked = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "requests_blocked_total",
		Help: "Blocked requests",
	},
)
