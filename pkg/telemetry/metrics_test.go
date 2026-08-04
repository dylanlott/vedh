package telemetry

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// TestNewCollectors_ConstructionSmoke is a minimal smoke test written by
// this plan's task 1. Task 3 owns exerciseAndGather and the full
// TestMetrics_* suite; this test only proves NewCollectors builds
// successfully against a private registry.
func TestNewCollectors_ConstructionSmoke(t *testing.T) {
	reg := prometheus.NewRegistry()
	c := NewCollectors(reg)
	if c == nil {
		t.Fatal("NewCollectors() = nil")
	}
}
