package matchpackage

import (
	"testing"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/monitor"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
	"hsanalyzer/pkg/contracts/matchupload"
)

func TestBuildMatchUploadPayloadAfterGameEnd(t *testing.T) {
	t.Parallel()

	gameState := state.New("game-1", state.GameModeStandard)
	match := monitor.PersistedMatch{
		Key: "game:game-1",
		Events: []game.Event{
			game.NewGameStartEvent("game-1", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypePlayer),
			game.NewGameEndEvent(types.PlayerTypePlayer),
		},
		State:    gameState,
		Complete: true,
		Reason:   monitor.CompletionReasonGameEnd,
	}

	payload := BuildMatchUploadPayload(match, "test")
	if payload.SchemaVersion != matchupload.SchemaVersionV1 {
		t.Fatalf("unexpected schema version: %s", payload.SchemaVersion)
	}
	if payload.MatchID != "game:game-1" {
		t.Fatalf("unexpected match id: %s", payload.MatchID)
	}
	if payload.GameID != "game-1" {
		t.Fatalf("unexpected game id: %s", payload.GameID)
	}
	if payload.CompletionStatus != matchupload.CompletionStatusComplete {
		t.Fatalf("unexpected completion status: %s", payload.CompletionStatus)
	}
	if payload.CompletionReason != matchupload.CompletionReasonGameEnd {
		t.Fatalf("unexpected completion reason: %s", payload.CompletionReason)
	}
	if len(payload.Events) != 3 {
		t.Fatalf("unexpected events count: %d", len(payload.Events))
	}
	if err := payload.Validate(); err != nil {
		t.Fatalf("validate payload: %v", err)
	}
}

func TestBuildMatchUploadPayloadForIncompleteFallback(t *testing.T) {
	t.Parallel()

	match := monitor.PersistedMatch{
		Key: "events:fallback",
		Events: []game.Event{
			game.NewGameStartEvent("", state.GameModeStandard),
			game.NewTurnStartEvent(types.PlayerTypeOpponent),
		},
		State:    state.New("", state.GameModeStandard),
		Complete: false,
		Reason:   monitor.CompletionReasonNextGameStart,
	}

	payload := BuildMatchUploadPayload(match, "")
	if payload.CompletionStatus != matchupload.CompletionStatusIncomplete {
		t.Fatalf("unexpected completion status: %s", payload.CompletionStatus)
	}
	if payload.CompletionReason != matchupload.CompletionReasonNextGameStart {
		t.Fatalf("unexpected completion reason: %s", payload.CompletionReason)
	}
	if err := payload.Validate(); err != nil {
		t.Fatalf("validate incomplete payload: %v", err)
	}
}
