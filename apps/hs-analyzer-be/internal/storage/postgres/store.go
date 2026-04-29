package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
	"hsanalyzer/pkg/contracts/matchupload"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MatchModel struct {
	MatchID          string            `gorm:"column:match_id;primaryKey"`
	IdempotencyKey   string            `gorm:"column:idempotency_key;uniqueIndex;not null"`
	PayloadHash      string            `gorm:"column:payload_hash;not null"`
	SchemaVersion    string            `gorm:"column:schema_version;not null"`
	CompletionStatus string            `gorm:"column:completion_status;not null"`
	CompletionReason string            `gorm:"column:completion_reason;not null"`
	Payload          datatypes.JSON    `gorm:"column:payload;type:jsonb;not null"`
	CreatedAt        time.Time         `gorm:"column:created_at;autoCreateTime"`
	Events           []MatchEventModel `gorm:"foreignKey:MatchID;references:MatchID;constraint:OnDelete:CASCADE"`
}

func (MatchModel) TableName() string {
	return "matches"
}

type MatchEventModel struct {
	MatchID        string         `gorm:"column:match_id;primaryKey"`
	EventIndex     int            `gorm:"column:event_index;primaryKey"`
	EventType      string         `gorm:"column:event_type;not null"`
	EventTimestamp string         `gorm:"column:event_timestamp"`
	EventPayload   datatypes.JSON `gorm:"column:event_payload;type:jsonb"`
}

func (MatchEventModel) TableName() string {
	return "match_events"
}

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func ApplyMigrations(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("postgres db is nil")
	}
	if err := db.WithContext(ctx).AutoMigrate(&MatchModel{}, &MatchEventModel{}); err != nil {
		return fmt.Errorf("auto-migrate postgres schema: %w", err)
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

	var stored bool
	err := s.db.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		var existing MatchModel
		err := tx.
			Where("match_id = ? OR idempotency_key = ?", payload.MatchID, idempotencyKey).
			First(&existing).Error
		switch {
		case err == nil:
			if existing.PayloadHash != payload.Hash() {
				return api.ErrDuplicateConflict
			}
			stored = false
			return nil
		case errors.Is(err, gorm.ErrRecordNotFound):
		default:
			return fmt.Errorf("query existing postgres match: %w", err)
		}

		model, err := buildMatchModel(payload, idempotencyKey)
		if err != nil {
			return err
		}
		if err := tx.Create(&model).Error; err != nil {
			return fmt.Errorf("create postgres match: %w", err)
		}
		stored = true
		return nil
	})
	return stored, err
}

func (s *Store) GetMatch(matchID string) (api.StoredMatch, bool) {
	if s == nil || s.db == nil {
		return api.StoredMatch{}, false
	}

	var model MatchModel
	err := s.db.
		Where("match_id = ?", matchID).
		First(&model).Error
	if err != nil {
		return api.StoredMatch{}, false
	}

	stored, err := model.ToStoredMatch()
	if err != nil {
		return api.StoredMatch{}, false
	}
	return stored, true
}

func (s *Store) ListMatches() []api.StoredMatch {
	if s == nil || s.db == nil {
		return nil
	}

	var models []MatchModel
	if err := s.db.Order("match_id ASC").Find(&models).Error; err != nil {
		return nil
	}

	result := make([]api.StoredMatch, 0, len(models))
	for _, model := range models {
		stored, err := model.ToStoredMatch()
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

func buildMatchModel(payload matchupload.Payload, idempotencyKey string) (MatchModel, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return MatchModel{}, fmt.Errorf("marshal match payload: %w", err)
	}

	events := append([]matchupload.Event(nil), payload.Events...)
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Index < events[j].Index
	})

	eventModels := make([]MatchEventModel, 0, len(events))
	for _, event := range events {
		rawEventPayload := []byte(event.Payload)
		if len(rawEventPayload) == 0 {
			rawEventPayload = []byte(`{}`)
		}
		if !json.Valid(rawEventPayload) {
			return MatchModel{}, fmt.Errorf("event %d payload is not valid json", event.Index)
		}
		eventModels = append(eventModels, MatchEventModel{
			MatchID:        payload.MatchID,
			EventIndex:     event.Index,
			EventType:      event.Type,
			EventTimestamp: event.Timestamp,
			EventPayload:   datatypes.JSON(rawEventPayload),
		})
	}

	return MatchModel{
		MatchID:          payload.MatchID,
		IdempotencyKey:   idempotencyKey,
		PayloadHash:      payload.Hash(),
		SchemaVersion:    payload.SchemaVersion,
		CompletionStatus: payload.CompletionStatus,
		CompletionReason: payload.CompletionReason,
		Payload:          datatypes.JSON(rawPayload),
		Events:           eventModels,
	}, nil
}

func (m MatchModel) ToStoredMatch() (api.StoredMatch, error) {
	var payload matchupload.Payload
	if err := json.Unmarshal(m.Payload, &payload); err != nil {
		return api.StoredMatch{}, err
	}
	return api.StoredMatch{
		Payload:        payload,
		IdempotencyKey: m.IdempotencyKey,
		PayloadHash:    m.PayloadHash,
	}, nil
}

func UpsertMatchEvents(db *gorm.DB, events []MatchEventModel) error {
	if len(events) == 0 {
		return nil
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&events).Error
}
