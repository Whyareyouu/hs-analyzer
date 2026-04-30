package overlay

import (
	"fmt"
	"strings"
	"time"
)

const (
	SchemaVersionV1 = "overlay_snapshot/v1"
)

const (
	MatchStatusIdle      = "idle"
	MatchStatusActive    = "active"
	MatchStatusFinalized = "finalized"
)

const (
	HealthStateHealthy  = "healthy"
	HealthStateDegraded = "degraded"
)

const (
	MessageTypeSnapshot  = "overlay_snapshot"
	MessageTypeDelta     = "overlay_delta"
	MessageTypeReset     = "overlay_reset"
	MessageTypeTransport = "transport_health"
)

type Snapshot struct {
	SchemaVersion       string               `json:"schema_version"`
	Match               Match                `json:"match"`
	PlayerDeckRemaining []PlayerDeckCard     `json:"player_deck_remaining"`
	OpponentPlayedCards []OpponentPlayedCard `json:"opponent_played_cards"`
	Health              Health               `json:"health"`
}

type Match struct {
	MatchID       string `json:"match_id,omitempty"`
	GameID        string `json:"game_id,omitempty"`
	Status        string `json:"status"`
	PlayerClass   string `json:"player_class,omitempty"`
	OpponentClass string `json:"opponent_class,omitempty"`
	DeckID        string `json:"deck_id,omitempty"`
	DeckName      string `json:"deck_name,omitempty"`
}

type Health struct {
	Monitor   string `json:"monitor"`
	Transport string `json:"transport"`
}

type PlayerDeckCard struct {
	CardID          string `json:"card_id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Image           string `json:"image,omitempty"`
	ManaCost        int    `json:"mana_cost"`
	Rarity          string `json:"rarity,omitempty"`
	CardType        string `json:"card_type,omitempty"`
	IsLegendary     bool   `json:"is_legendary,omitempty"`
	InitialCopies   int    `json:"initial_copies"`
	RemainingCopies int    `json:"remaining_copies"`
	DrawnCopies     int    `json:"drawn_copies"`
	PlayedCopies    int    `json:"played_copies"`
}

type OpponentPlayedCard struct {
	CardID        string `json:"card_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Image         string `json:"image,omitempty"`
	ManaCost      int    `json:"mana_cost"`
	PlayCount     int    `json:"play_count"`
	FirstPlayedAt string `json:"first_played_at,omitempty"`
	LastPlayedAt  string `json:"last_played_at,omitempty"`
}

type Envelope struct {
	Type          string    `json:"type"`
	SchemaVersion string    `json:"schema_version"`
	Snapshot      *Snapshot `json:"snapshot,omitempty"`
	Delta         *Snapshot `json:"delta,omitempty"`
	Health        *Health   `json:"health,omitempty"`
	MatchID       string    `json:"match_id,omitempty"`
}

func NewSnapshot() Snapshot {
	return Snapshot{
		SchemaVersion:       SchemaVersionV1,
		Match:               Match{Status: MatchStatusIdle},
		PlayerDeckRemaining: make([]PlayerDeckCard, 0),
		OpponentPlayedCards: make([]OpponentPlayedCard, 0),
		Health: Health{
			Monitor:   HealthStateHealthy,
			Transport: HealthStateHealthy,
		},
	}
}

func (s Snapshot) Clone() Snapshot {
	cloned := s
	cloned.PlayerDeckRemaining = append(make([]PlayerDeckCard, 0, len(s.PlayerDeckRemaining)), s.PlayerDeckRemaining...)
	cloned.OpponentPlayedCards = append(make([]OpponentPlayedCard, 0, len(s.OpponentPlayedCards)), s.OpponentPlayedCards...)
	return cloned
}

func (s *Snapshot) Normalize() {
	if s == nil {
		return
	}
	if strings.TrimSpace(s.SchemaVersion) == "" {
		s.SchemaVersion = SchemaVersionV1
	}
	if strings.TrimSpace(s.Match.Status) == "" {
		s.Match.Status = MatchStatusIdle
	}
	if s.PlayerDeckRemaining == nil {
		s.PlayerDeckRemaining = make([]PlayerDeckCard, 0)
	}
	if s.OpponentPlayedCards == nil {
		s.OpponentPlayedCards = make([]OpponentPlayedCard, 0)
	}
	if strings.TrimSpace(s.Health.Monitor) == "" {
		s.Health.Monitor = HealthStateHealthy
	}
	if strings.TrimSpace(s.Health.Transport) == "" {
		s.Health.Transport = HealthStateHealthy
	}
}

func (s Snapshot) Validate() error {
	if s.SchemaVersion != SchemaVersionV1 {
		return fmt.Errorf("unsupported schema version %q", s.SchemaVersion)
	}
	switch s.Match.Status {
	case MatchStatusIdle, MatchStatusActive, MatchStatusFinalized:
	default:
		return fmt.Errorf("unsupported match status %q", s.Match.Status)
	}
	if err := validateHealth(s.Health); err != nil {
		return err
	}
	for _, card := range s.PlayerDeckRemaining {
		if err := validatePlayerDeckCard(card); err != nil {
			return err
		}
	}
	for _, card := range s.OpponentPlayedCards {
		if err := validateOpponentPlayedCard(card); err != nil {
			return err
		}
	}
	return nil
}

func (e Envelope) Validate() error {
	if strings.TrimSpace(e.Type) == "" {
		return fmt.Errorf("message type is empty")
	}
	if strings.TrimSpace(e.SchemaVersion) == "" {
		return fmt.Errorf("schema version is empty")
	}
	if e.SchemaVersion != SchemaVersionV1 {
		return fmt.Errorf("unsupported envelope schema version %q", e.SchemaVersion)
	}

	switch e.Type {
	case MessageTypeSnapshot:
		if e.Snapshot == nil {
			return fmt.Errorf("snapshot payload is required")
		}
		return e.Snapshot.Validate()
	case MessageTypeDelta:
		if e.Delta == nil {
			return fmt.Errorf("delta payload is required")
		}
		return e.Delta.Validate()
	case MessageTypeReset:
		if strings.TrimSpace(e.MatchID) == "" {
			return fmt.Errorf("match id is required for reset")
		}
		return nil
	case MessageTypeTransport:
		if e.Health == nil {
			return fmt.Errorf("health payload is required")
		}
		return validateHealth(*e.Health)
	default:
		return fmt.Errorf("unsupported message type %q", e.Type)
	}
}

func validateHealth(health Health) error {
	switch health.Monitor {
	case HealthStateHealthy, HealthStateDegraded:
	default:
		return fmt.Errorf("unsupported monitor health %q", health.Monitor)
	}
	switch health.Transport {
	case HealthStateHealthy, HealthStateDegraded:
	default:
		return fmt.Errorf("unsupported transport health %q", health.Transport)
	}
	return nil
}

func validatePlayerDeckCard(card PlayerDeckCard) error {
	if strings.TrimSpace(card.CardID) == "" {
		return fmt.Errorf("player deck card id is empty")
	}
	if card.InitialCopies < 0 || card.RemainingCopies < 0 || card.DrawnCopies < 0 || card.PlayedCopies < 0 {
		return fmt.Errorf("player deck copies must be non-negative")
	}
	if card.RemainingCopies > card.InitialCopies {
		return fmt.Errorf("remaining copies exceed initial copies for %s", card.CardID)
	}
	if card.DrawnCopies > card.InitialCopies {
		return fmt.Errorf("drawn copies exceed initial copies for %s", card.CardID)
	}
	if card.PlayedCopies > card.InitialCopies {
		return fmt.Errorf("played copies exceed initial copies for %s", card.CardID)
	}
	return nil
}

func validateOpponentPlayedCard(card OpponentPlayedCard) error {
	if strings.TrimSpace(card.CardID) == "" {
		return fmt.Errorf("opponent card id is empty")
	}
	if card.PlayCount <= 0 {
		return fmt.Errorf("opponent play count must be positive for %s", card.CardID)
	}
	if err := validateOptionalTime(card.FirstPlayedAt); err != nil {
		return fmt.Errorf("invalid first played at for %s: %w", card.CardID, err)
	}
	if err := validateOptionalTime(card.LastPlayedAt); err != nil {
		return fmt.Errorf("invalid last played at for %s: %w", card.CardID, err)
	}
	return nil
}

func validateOptionalTime(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if _, err := time.Parse(time.RFC3339Nano, raw); err != nil {
		return err
	}
	return nil
}
