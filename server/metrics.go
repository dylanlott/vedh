package server

import (
	"github.com/openmtg/edh-go/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus"
)

// collectors is the single registration site for every vedh_-prefixed
// Prometheus collector this phase declares. promauto (inside
// telemetry.NewCollectors) registers into prometheus.DefaultRegisterer,
// which the already-mounted promhttp.Handler() serves (server/graphql.go),
// so Serve() needs no edit and the collectors inherit withMetricsAuth,
// METRICS_ENABLED, and METRICS_TOKEN unchanged.
//
// Declaring all four ROADMAP criterion-4 families here does not, by
// itself, close criterion 4: an unobserved CounterVec/HistogramVec exports
// no child series until an emit site calls one of Collectors' Observe*
// methods. See pkg/telemetry/metrics.go for exactly which families this
// plan observes versus declares for a later phase.
var collectors = telemetry.NewCollectors(prometheus.DefaultRegisterer)
