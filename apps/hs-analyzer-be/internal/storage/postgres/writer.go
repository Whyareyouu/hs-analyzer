package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"hsanalyzer/pkg/contracts/matchupload"
)

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Writer struct {
	db SQLExecutor
}

func NewWriter(db SQLExecutor) *Writer {
	return &Writer{db: db}
}

func (w *Writer) SaveMatch(ctx context.Context, payload matchupload.Payload, idempotencyKey string) error {
	if w == nil || w.db == nil {
		return fmt.Errorf("postgres writer is not configured")
	}
	if err := payload.Validate(); err != nil {
		return err
	}
	if err := matchupload.ValidateIdempotencyKey(idempotencyKey); err != nil {
		return err
	}

	if _, err := w.db.ExecContext(ctx, `
INSERT INTO matches (match_id, idempotency_key, payload_hash, schema_version, completion_status, completion_reason, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (match_id) DO NOTHING`,
		payload.MatchID,
		idempotencyKey,
		payload.Hash(),
		payload.SchemaVersion,
		payload.CompletionStatus,
		payload.CompletionReason,
		mustMarshalPayload(payload),
	); err != nil {
		return fmt.Errorf("insert match: %w", err)
	}

	for _, event := range payload.Events {
		eventPayload := []byte(event.Payload)
		if len(eventPayload) == 0 {
			eventPayload = []byte(`{}`)
		}
		if !json.Valid(eventPayload) {
			return fmt.Errorf("event %d payload is not valid json", event.Index)
		}

		if _, err := w.db.ExecContext(ctx, `
INSERT INTO match_events (match_id, event_index, event_type, event_timestamp, event_payload)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (match_id, event_index) DO NOTHING`,
			payload.MatchID,
			event.Index,
			event.Type,
			event.Timestamp,
			string(eventPayload),
		); err != nil {
			return fmt.Errorf("insert match event %d: %w", event.Index, err)
		}
	}

	return nil
}

func mustMarshalPayload(payload matchupload.Payload) string {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
