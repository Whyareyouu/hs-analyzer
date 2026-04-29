package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
	"hsanalyzer/pkg/contracts/matchupload"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("postgres db is nil")
	}
	for index, migration := range Migrations() {
		if _, err := db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("apply postgres migration %d: %w", index+1, err)
		}
	}
	return nil
}

func (s *Store) SaveMatch(payload matchupload.Payload, idempotencyKey string) (bool, error) {
	if s == nil || s.db == nil {
		return false, fmt.Errorf("postgres store is not configured")
	}
	if err := payload.Validate(); err != nil {
		return false, err
	}
	if err := matchupload.ValidateIdempotencyKey(idempotencyKey); err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin postgres transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	hash := payload.Hash()
	existingHash, ok, err := findExistingPayloadHash(ctx, tx, payload.MatchID, idempotencyKey)
	if err != nil {
		return false, err
	}
	if ok {
		if existingHash != hash {
			return false, api.ErrDuplicateConflict
		}
		if err := tx.Commit(); err != nil {
			return false, fmt.Errorf("commit postgres duplicate transaction: %w", err)
		}
		return false, nil
	}

	writer := NewWriter(tx)
	if err := writer.SaveMatch(ctx, payload, idempotencyKey); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit postgres transaction: %w", err)
	}
	return true, nil
}

func (s *Store) GetMatch(matchID string) (api.StoredMatch, bool) {
	if s == nil || s.db == nil {
		return api.StoredMatch{}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, `
SELECT payload, idempotency_key, payload_hash
FROM matches
WHERE match_id = $1`, matchID)

	stored, err := scanStoredMatch(row)
	if err != nil {
		return api.StoredMatch{}, false
	}
	return stored, true
}

func (s *Store) ListMatches() []api.StoredMatch {
	if s == nil || s.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `
SELECT payload, idempotency_key, payload_hash
FROM matches
ORDER BY match_id`)
	if err != nil {
		return nil
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make([]api.StoredMatch, 0)
	for rows.Next() {
		stored, err := scanStoredMatch(rows)
		if err != nil {
			continue
		}
		result = append(result, stored)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Payload.MatchID < result[j].Payload.MatchID
	})
	return result
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanStoredMatch(row rowScanner) (api.StoredMatch, error) {
	var rawPayload []byte
	var idempotencyKey string
	var payloadHash string
	if err := row.Scan(&rawPayload, &idempotencyKey, &payloadHash); err != nil {
		return api.StoredMatch{}, err
	}

	var payload matchupload.Payload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return api.StoredMatch{}, err
	}

	return api.StoredMatch{
		Payload:        payload,
		IdempotencyKey: idempotencyKey,
		PayloadHash:    payloadHash,
	}, nil
}

func findExistingPayloadHash(ctx context.Context, tx *sql.Tx, matchID, idempotencyKey string) (string, bool, error) {
	var hash string
	err := tx.QueryRowContext(ctx, `
SELECT payload_hash
FROM matches
WHERE match_id = $1 OR idempotency_key = $2
LIMIT 1`, matchID, idempotencyKey).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("query existing postgres match: %w", err)
	}
	return hash, true, nil
}
