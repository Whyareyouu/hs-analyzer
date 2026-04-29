package postgres

import (
	"encoding/json"
	"sync"
	"testing"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
	"hsanalyzer/pkg/contracts/matchupload"

	"gorm.io/gorm/schema"
)

func TestPostgresModelsUseGORMMappings(t *testing.T) {
	t.Parallel()

	parsed, err := schema.Parse(&MatchModel{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse gorm schema: %v", err)
	}
	if parsed.Table != "matches" {
		t.Fatalf("unexpected table: %s", parsed.Table)
	}
	if parsed.PrioritizedPrimaryField == nil || parsed.PrioritizedPrimaryField.DBName != "match_id" {
		t.Fatalf("expected match_id primary field, got %#v", parsed.PrioritizedPrimaryField)
	}
}

func TestBuildMatchModelStoresFullPayloadAndSortedEvents(t *testing.T) {
	t.Parallel()

	payload := testPayload("match-1")
	payload.Events = []matchupload.Event{
		{Index: 2, Type: "turn_end"},
		{Index: 1, Type: "game_start", Payload: json.RawMessage(`{"ok":true}`)},
	}

	model, err := buildMatchModel(payload, "match-1")
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	if !json.Valid(model.Payload) {
		t.Fatalf("expected payload json: %s", string(model.Payload))
	}
	if len(model.Events) != 2 || model.Events[0].EventIndex != 1 || model.Events[1].EventIndex != 2 {
		t.Fatalf("events were not sorted by index: %#v", model.Events)
	}

	stored, err := model.ToStoredMatch()
	if err != nil {
		t.Fatalf("convert model: %v", err)
	}
	if stored.Payload.MatchID != "match-1" || stored.IdempotencyKey != "match-1" {
		t.Fatalf("unexpected stored match: %#v", stored)
	}
}

func TestBuildMatchModelRejectsInvalidEventPayload(t *testing.T) {
	t.Parallel()

	payload := testPayload("match-1")
	payload.Events = []matchupload.Event{{Index: 0, Type: "bad", Payload: json.RawMessage(`{`)}}
	if _, err := buildMatchModel(payload, "match-1"); err == nil {
		t.Fatalf("expected invalid event payload error")
	}
}

func TestPostgresStoreImplementsMatchStore(t *testing.T) {
	t.Parallel()

	var _ api.MatchStore = (*Store)(nil)
}

func testPayload(matchID string) matchupload.Payload {
	return matchupload.Payload{
		SchemaVersion:    matchupload.SchemaVersionV1,
		MatchID:          matchID,
		CompletionStatus: matchupload.CompletionStatusComplete,
		CompletionReason: matchupload.CompletionReasonGameEnd,
		Source:           matchupload.Source{ClientName: "hs-analyzer-cli"},
		Events:           []matchupload.Event{{Index: 0, Type: "game_start"}},
	}
}
