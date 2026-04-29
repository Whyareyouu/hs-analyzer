package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"hsanalyzer/pkg/contracts/matchupload"
)

type recordingExecutor struct {
	queries []string
	args    [][]interface{}
}

func (e *recordingExecutor) ExecContext(_ context.Context, query string, args ...interface{}) (sql.Result, error) {
	e.queries = append(e.queries, query)
	e.args = append(e.args, args)
	return nil, nil
}

func TestPostgresWriterStoresMatchAndEventsInStableOrder(t *testing.T) {
	t.Parallel()

	executor := &recordingExecutor{}
	writer := NewWriter(executor)

	payload := matchupload.Payload{
		SchemaVersion:    matchupload.SchemaVersionV1,
		MatchID:          "match-1",
		CompletionStatus: matchupload.CompletionStatusComplete,
		CompletionReason: matchupload.CompletionReasonGameEnd,
		Source:           matchupload.Source{ClientName: "hs-analyzer-cli"},
		Events: []matchupload.Event{
			{Index: 0, Type: "game_start"},
			{Index: 1, Type: "game_end"},
		},
	}

	if err := writer.SaveMatch(context.Background(), payload, "match-1"); err != nil {
		t.Fatalf("save match: %v", err)
	}
	if len(executor.queries) != 3 {
		t.Fatalf("unexpected query count: %d", len(executor.queries))
	}
	if !strings.Contains(executor.queries[0], "INSERT INTO matches") {
		t.Fatalf("expected first query to insert match: %s", executor.queries[0])
	}
	rawPayload, ok := executor.args[0][6].(string)
	if !ok || !json.Valid([]byte(rawPayload)) {
		t.Fatalf("expected full match payload json in match insert args: %#v", executor.args[0])
	}
	if !strings.Contains(executor.queries[1], "INSERT INTO match_events") || !strings.Contains(executor.queries[2], "INSERT INTO match_events") {
		t.Fatalf("expected event insert queries: %#v", executor.queries)
	}
	if executor.args[1][1] != 0 || executor.args[2][1] != 1 {
		t.Fatalf("events were not written in stable order: %#v", executor.args)
	}
}
