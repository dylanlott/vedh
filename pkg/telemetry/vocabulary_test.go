package telemetry

import "testing"

// TestValidateEvent_ConstructionSmoke is a minimal smoke test written by
// this plan's task 1. Task 3 owns the full TestValidateEvent_Rejections
// table and the closed-vocabulary / forbidden-key structural tests; this
// test only proves ValidateEvent accepts a known-good event with a
// non-blank session ID before those land.
func TestValidateEvent_ConstructionSmoke(t *testing.T) {
	got := ValidateEvent("quick_start_viewed", "session-1", nil, false)
	if got != RejectionNone {
		t.Fatalf("ValidateEvent() = %v, want RejectionNone", got)
	}
}
