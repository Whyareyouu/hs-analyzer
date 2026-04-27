package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/monitor"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
	"hsanalyzer/pkg/contracts/matchupload"

	"go.uber.org/zap"
)

func TestDeliveryMatchPersisterUploadsCompletedMatch(t *testing.T) {
	t.Parallel()

	var received matchupload.Payload
	var idempotencyKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		idempotencyKey = r.Header.Get(matchupload.HTTPHeaderIdempotencyKey)
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode upload payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(matchupload.Response{Status: matchupload.ResultAccepted, MatchID: received.MatchID})
	}))
	defer server.Close()

	persister, err := newMonitorMatchPersister(monitorOptions{
		matchUploadURL: server.URL,
		matchSpoolDir:  filepath.Join(t.TempDir(), "spool"),
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("create persister: %v", err)
	}

	match := monitor.PersistedMatch{
		Key:   "game:upload",
		State: state.New("upload", state.GameModeStandard),
		Events: []game.Event{
			game.NewGameStartEvent("upload", state.GameModeStandard),
			game.NewGameEndEvent(types.PlayerTypePlayer),
		},
		Complete: true,
		Reason:   monitor.CompletionReasonGameEnd,
	}

	if err := persister.SaveMatch(context.Background(), match); err != nil {
		t.Fatalf("save match: %v", err)
	}
	if persister.SavedCount() != 1 {
		t.Fatalf("unexpected saved count: %d", persister.SavedCount())
	}
	if received.MatchID != "game:upload" {
		t.Fatalf("unexpected uploaded match id: %s", received.MatchID)
	}
	if idempotencyKey != received.IdempotencyKey() {
		t.Fatalf("unexpected idempotency key: %q want %q", idempotencyKey, received.IdempotencyKey())
	}
}
