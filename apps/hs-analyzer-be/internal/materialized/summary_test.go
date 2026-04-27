package materialized

import (
	"testing"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
	"hsanalyzer/pkg/contracts/matchupload"
)

func TestBackendMaterializedViewReadsFromCanonicalStorage(t *testing.T) {
	t.Parallel()

	summary := BuildMatchSummary(api.StoredMatch{Payload: materializedPayload("match-1")})
	if summary.MatchID != "match-1" {
		t.Fatalf("unexpected match id: %s", summary.MatchID)
	}
	if summary.EventCount != 2 {
		t.Fatalf("unexpected event count: %d", summary.EventCount)
	}
	if summary.GameMode != "standard" {
		t.Fatalf("unexpected game mode: %s", summary.GameMode)
	}
}

func TestBackendHistoryIsDeterministic(t *testing.T) {
	t.Parallel()

	history := BuildHistory([]api.StoredMatch{
		{Payload: materializedPayload("match-b")},
		{Payload: materializedPayload("match-a")},
	})
	if history[0].MatchID != "match-a" || history[1].MatchID != "match-b" {
		t.Fatalf("unexpected history order: %#v", history)
	}
}

func materializedPayload(matchID string) matchupload.Payload {
	return matchupload.Payload{
		SchemaVersion:    matchupload.SchemaVersionV1,
		MatchID:          matchID,
		CompletionStatus: matchupload.CompletionStatusComplete,
		CompletionReason: matchupload.CompletionReasonGameEnd,
		Source:           matchupload.Source{ClientName: "hs-analyzer-cli"},
		Metadata:         map[string]string{"game_mode": "standard"},
		Events: []matchupload.Event{
			{Index: 0, Type: "game_start"},
			{Index: 1, Type: "game_end"},
		},
	}
}
