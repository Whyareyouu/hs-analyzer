package projections

import (
	"reflect"
	"testing"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
	"hsanalyzer/pkg/contracts/matchupload"
)

func TestBackendProjectionBuildUsesCanonicalStorage(t *testing.T) {
	t.Parallel()

	graph := BuildGraph([]api.StoredMatch{{Payload: graphPayload("match-1")}})
	if len(graph.Nodes) != 3 {
		t.Fatalf("unexpected nodes: %#v", graph.Nodes)
	}
	if len(graph.Edges) != 2 {
		t.Fatalf("unexpected edges: %#v", graph.Edges)
	}
	if graph.Edges[0].From != "match-1" {
		t.Fatalf("unexpected edge source: %#v", graph.Edges[0])
	}
}

func TestBackendProjectionRebuildIsDeterministic(t *testing.T) {
	t.Parallel()

	first := BuildGraph([]api.StoredMatch{{Payload: graphPayload("match-b")}, {Payload: graphPayload("match-a")}})
	second := BuildGraph([]api.StoredMatch{{Payload: graphPayload("match-a")}, {Payload: graphPayload("match-b")}})
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic graph\nfirst=%#v\nsecond=%#v", first, second)
	}
}

func graphPayload(matchID string) matchupload.Payload {
	return matchupload.Payload{
		SchemaVersion:    matchupload.SchemaVersionV1,
		MatchID:          matchID,
		CompletionStatus: matchupload.CompletionStatusComplete,
		CompletionReason: matchupload.CompletionReasonGameEnd,
		Source:           matchupload.Source{ClientName: "hs-analyzer-cli"},
		Events: []matchupload.Event{
			{Index: 0, Type: "game_start"},
			{Index: 1, Type: "game_end"},
		},
	}
}
