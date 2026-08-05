package telemetry

import (
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/openmtg/edh-go/pkg/deckimport"
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

// exerciseAndGather builds a private prometheus.NewRegistry(), constructs
// Collectors against it, exercises every collector once with
// representative values — including the guest session, game create, game
// join, and board activation families whose real emit sites are in Phases
// 2 and 3 — and returns reg.Gather().
//
// This helper is mandatory rather than a convenience: a CounterVec or
// HistogramVec with no observations reports no child metrics, so a test
// that gathers from a registry it did not exercise is vacuously green for
// exactly the families nothing in this phase touches. All four
// TestMetrics_* tests below call this same helper, so "the same gather"
// is guaranteed across four separate test functions rather than each one
// risking a fresh, unexercised registry.
func exerciseAndGather(t *testing.T) []*dto.MetricFamily {
	t.Helper()

	reg := prometheus.NewRegistry()
	c := NewCollectors(reg)

	c.ObserveDeckImport(deckimport.SourcePlainText, "success", 10*time.Millisecond)
	c.ObserveGuestSession("success", 5*time.Millisecond)
	c.ObserveGameCreate("success", 5*time.Millisecond)
	c.ObserveGameJoin("success", 5*time.Millisecond)
	c.ObserveBoardActivation(RoleHost, "success", 5*time.Millisecond)
	c.RecordProductEventWritten("recorded")
	c.RecordProductEventDropped(RejectionUnknownEvent)
	c.ObserveDeckSuggestion("success", 10*time.Millisecond)
	c.IncDeckSuggestionTruncated()

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	return families
}

// TestMetrics_LabelAllowlist asserts that every label name on every
// vedh_-prefixed family is a member of AllowedLabelNames.
func TestMetrics_LabelAllowlist(t *testing.T) {
	for _, fam := range exerciseAndGather(t) {
		if !hasVedhPrefix(fam.GetName()) {
			continue
		}
		for _, m := range fam.GetMetric() {
			for _, lp := range m.GetLabel() {
				if _, ok := AllowedLabelNames[lp.GetName()]; !ok {
					t.Errorf("metric %s carries disallowed label %q (value %q)",
						fam.GetName(), lp.GetName(), lp.GetValue())
				}
			}
		}
	}
}

// criterionFourFamilies names the counter/histogram pair for each of the
// five families ROADMAP Phase 1 criterion 4 requires. Create and join are
// separate families, per pkg/telemetry/metrics.go's design note.
var criterionFourFamilies = map[string][2]string{
	"guest session":    {"vedh_guest_session_total", "vedh_guest_session_duration_seconds"},
	"deck import":      {"vedh_deck_import_total", "vedh_deck_import_duration_seconds"},
	"game create":      {"vedh_game_create_total", "vedh_game_create_duration_seconds"},
	"game join":        {"vedh_game_join_total", "vedh_game_join_duration_seconds"},
	"board activation": {"vedh_board_activation_total", "vedh_board_activation_duration_seconds"},
}

// TestMetrics_AllCriterionFourFamiliesExist asserts that a family exists
// for each of guest session, deck import, game create, game join, and
// board activation, and that each has both a counter and a histogram. This
// is the test that goes red if a later refactor deletes a
// declared-but-not-yet-observed collector as apparently dead code.
func TestMetrics_AllCriterionFourFamiliesExist(t *testing.T) {
	families := exerciseAndGather(t)
	byName := map[string]*dto.MetricFamily{}
	for _, fam := range families {
		byName[fam.GetName()] = fam
	}

	for label, names := range criterionFourFamilies {
		counterFam, ok := byName[names[0]]
		if !ok {
			t.Errorf("%s: missing counter family %s", label, names[0])
			continue
		}
		if counterFam.GetType() != dto.MetricType_COUNTER {
			t.Errorf("%s: family %s type = %v, want COUNTER", label, names[0], counterFam.GetType())
		}

		histFam, ok := byName[names[1]]
		if !ok {
			t.Errorf("%s: missing histogram family %s", label, names[1])
			continue
		}
		if histFam.GetType() != dto.MetricType_HISTOGRAM {
			t.Errorf("%s: family %s type = %v, want HISTOGRAM", label, names[1], histFam.GetType())
		}
	}
}

// TestMetrics_SourceLabelValuesAreEnumMembers asserts that every label pair
// named source carries a value in deckimport.AllSourceTypes(), and every
// label pair named role a value in the declared Role constants.
func TestMetrics_SourceLabelValuesAreEnumMembers(t *testing.T) {
	validSource := map[string]struct{}{}
	for _, s := range deckimport.AllSourceTypes() {
		validSource[string(s)] = struct{}{}
	}
	validRole := map[string]struct{}{
		string(RoleHost):    {},
		string(RoleInvitee): {},
	}

	for _, fam := range exerciseAndGather(t) {
		if !hasVedhPrefix(fam.GetName()) {
			continue
		}
		for _, m := range fam.GetMetric() {
			for _, lp := range m.GetLabel() {
				switch lp.GetName() {
				case "source":
					if _, ok := validSource[lp.GetValue()]; !ok {
						t.Errorf("metric %s carries source=%q, not a deckimport.AllSourceTypes() member",
							fam.GetName(), lp.GetValue())
					}
				case "role":
					if _, ok := validRole[lp.GetValue()]; !ok {
						t.Errorf("metric %s carries role=%q, not a declared Role constant",
							fam.GetName(), lp.GetValue())
					}
				}
			}
		}
	}
}

// allBoundedRejectionStrings is every Rejection's String() form, the
// closed set of values the reason label may ever carry.
func allBoundedRejectionStrings() map[string]struct{} {
	out := map[string]struct{}{}
	for _, r := range []Rejection{
		RejectionNone,
		RejectionUnknownEvent,
		RejectionUnknownKey,
		RejectionOversizedValue,
		RejectionClientAuthoritative,
		RejectionMissingSession,
		RejectionWriteError,
	} {
		out[r.String()] = struct{}{}
	}
	return out
}

// TestMetrics_ReasonLabelValuesAreBounded asserts every reason value is one
// of the declared Rejection strings — which includes missing_session.
func TestMetrics_ReasonLabelValuesAreBounded(t *testing.T) {
	bounded := allBoundedRejectionStrings()

	for _, fam := range exerciseAndGather(t) {
		if !hasVedhPrefix(fam.GetName()) {
			continue
		}
		for _, m := range fam.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() != "reason" {
					continue
				}
				if _, ok := bounded[lp.GetValue()]; !ok {
					t.Errorf("metric %s carries reason=%q, not a declared Rejection string",
						fam.GetName(), lp.GetValue())
				}
			}
		}
	}
}

func hasVedhPrefix(name string) bool {
	return len(name) >= len("vedh_") && name[:len("vedh_")] == "vedh_"
}
