package formatter

import (
	"encoding/json"
	"strings"
	"testing"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

func TestJSONFormatterFormatStateAsTurnsArray(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter()
	gameState := state.New("game-1", state.GameModeStandard)

	turnStart := game.NewTurnStartEvent(types.PlayerTypePlayer)
	if err := turnStart.Apply(gameState); err != nil {
		t.Fatalf("apply turn start: %v", err)
	}
	gameState.AddEvent(turnStart)

	drawn := card.New("CARD_DRAW", 101, types.PlayerTypePlayer)
	drawEvent := game.NewCardDrawnEvent(types.PlayerTypePlayer, drawn)
	if err := drawEvent.Apply(gameState); err != nil {
		t.Fatalf("apply draw: %v", err)
	}
	gameState.AddEvent(drawEvent)

	turnEnd := game.NewTurnEndEvent()
	if err := turnEnd.Apply(gameState); err != nil {
		t.Fatalf("apply turn end: %v", err)
	}
	gameState.AddEvent(turnEnd)

	raw, err := formatter.Format(gameState)
	if err != nil {
		t.Fatalf("format state: %v", err)
	}

	var turns []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &turns); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	if len(turns) != 1 {
		t.Fatalf("unexpected turns count: %d", len(turns))
	}
	if turns[0]["number"].(float64) != 1 {
		t.Fatalf("unexpected turn number: %#v", turns[0]["number"])
	}
	events, ok := turns[0]["events"].([]interface{})
	if !ok || len(events) != 3 {
		t.Fatalf("unexpected turn events: %#v", turns[0]["events"])
	}
}

func TestJSONFormatterFormatEvent(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter()
	raw, err := formatter.FormatEvent(game.NewAttackEvent(10, 20))
	if err != nil {
		t.Fatalf("format event: %v", err)
	}

	if !strings.Contains(raw, `"type": "attack"`) {
		t.Fatalf("unexpected json: %s", raw)
	}
	if !strings.Contains(raw, `"attacker_entity_id": 10`) {
		t.Fatalf("unexpected json: %s", raw)
	}
}

func TestJSONFormatterRejectsNilInputs(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter()
	if _, err := formatter.Format(nil); err == nil {
		t.Fatalf("expected nil state error")
	}
	if _, err := formatter.FormatEvent(nil); err == nil {
		t.Fatalf("expected nil event error")
	}
}
