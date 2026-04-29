package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hsanalyzer/pkg/contracts/matchupload"
)

func TestPostMatchesStoresValidPayload(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	server := NewServer(store).Routes()

	resp := postPayload(t, server, backendPayload(), "match-1")
	if resp.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d body=%s", resp.Code, resp.Body.String())
	}

	if _, ok := store.GetMatch("match-1"); !ok {
		t.Fatalf("expected stored match")
	}
}

func TestPostMatchesRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	payload := backendPayload()
	payload.SchemaVersion = "invalid"

	resp := postPayload(t, NewServer(NewMemoryStore()).Routes(), payload, "match-1")
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestPostMatchesHandlesDuplicateSamePayload(t *testing.T) {
	t.Parallel()

	server := NewServer(NewMemoryStore()).Routes()
	payload := backendPayload()

	first := postPayload(t, server, payload, "match-1")
	second := postPayload(t, server, payload, "match-1")
	if first.Code != http.StatusAccepted {
		t.Fatalf("unexpected first status: %d", first.Code)
	}
	if second.Code != http.StatusOK {
		t.Fatalf("unexpected second status: %d body=%s", second.Code, second.Body.String())
	}
}

func TestPostMatchesRejectsDuplicateDifferentPayload(t *testing.T) {
	t.Parallel()

	server := NewServer(NewMemoryStore()).Routes()
	firstPayload := backendPayload()
	secondPayload := backendPayload()
	secondPayload.Events = append(secondPayload.Events, matchupload.Event{Index: 1, Type: "game_end"})

	first := postPayload(t, server, firstPayload, "match-1")
	second := postPayload(t, server, secondPayload, "match-1")
	if first.Code != http.StatusAccepted {
		t.Fatalf("unexpected first status: %d", first.Code)
	}
	if second.Code != http.StatusConflict {
		t.Fatalf("unexpected second status: %d body=%s", second.Code, second.Body.String())
	}
}

func TestPostMatchesReturnsStableErrorShape(t *testing.T) {
	t.Parallel()

	resp := postPayload(t, NewServer(NewMemoryStore()).Routes(), backendPayload(), "")
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", resp.Code)
	}

	var decoded matchupload.ErrorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if decoded.Error.Code != matchupload.ErrorValidationFailed {
		t.Fatalf("unexpected error code: %s", decoded.Error.Code)
	}
}

func TestDeckDashboardEndpointsReturnTypedPayloads(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	server := NewServer(store).Routes()
	_ = postPayload(t, server, backendDashboardPayload(), "match-1")

	for _, path := range []string{
		"/api/v1/decks",
		"/api/v1/decks/deck-spell-dh/overview",
		"/api/v1/decks/deck-spell-dh/cards",
		"/api/v1/decks/deck-spell-dh/matchups",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("unexpected status for %s: %d body=%s", path, resp.Code, resp.Body.String())
		}
		if !json.Valid(resp.Body.Bytes()) {
			t.Fatalf("invalid json for %s: %s", path, resp.Body.String())
		}
	}
}

func TestDeckDashboardReturnsNotFoundForUnknownDeck(t *testing.T) {
	t.Parallel()

	server := NewServer(NewMemoryStore()).Routes()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/decks/unknown/overview", nil)
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d body=%s", resp.Code, resp.Body.String())
	}
}

func postPayload(t *testing.T, handler http.Handler, payload matchupload.Payload, idempotencyKey string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches", bytes.NewReader(body))
	if idempotencyKey != "" {
		req.Header.Set(matchupload.HTTPHeaderIdempotencyKey, idempotencyKey)
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func backendPayload() matchupload.Payload {
	return matchupload.Payload{
		SchemaVersion:    matchupload.SchemaVersionV1,
		MatchID:          "match-1",
		CompletionStatus: matchupload.CompletionStatusComplete,
		CompletionReason: matchupload.CompletionReasonGameEnd,
		Source:           matchupload.Source{ClientName: "hs-analyzer-cli"},
		Events: []matchupload.Event{
			{Index: 0, Type: "game_start"},
		},
	}
}

func backendDashboardPayload() matchupload.Payload {
	payload := backendPayload()
	payload.MatchID = "game:1"
	payload.EndedAt = "2026-04-29T10:15:03.000000000Z"
	_ = payload.SetDashboardSummary(matchupload.DashboardSummary{
		DeckID:      "deck-spell-dh",
		DeckName:    "Spell Demon Hunter",
		PlayerClass: "DEMONHUNTER",
		Result:      matchupload.DashboardResultWin,
		DeckCards: []matchupload.DashboardDeckCard{
			{CardID: "CARD_1", Name: "Taste of Chaos", ManaCost: 1, Copies: 2, Rarity: "COMMON"},
		},
	})
	return payload
}
