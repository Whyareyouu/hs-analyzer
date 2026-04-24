package formatter

import (
	"encoding/json"
	"fmt"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
)

// JSONFormatter сериализует итоговое состояние игры в JSON-хронологию по ходам.
type JSONFormatter struct{}

// NewJSONFormatter создает JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Name() string {
	return "json"
}

func (f *JSONFormatter) Format(gameState *state.State) (string, error) {
	if gameState == nil {
		return "", fmt.Errorf("game state is nil")
	}

	timeline := make([]jsonTurn, 0, len(gameState.Turns))
	for _, turn := range gameState.Turns {
		if turn == nil {
			continue
		}

		item := jsonTurn{
			Number:     turn.Number,
			Player:     turn.Player.String(),
			StartedAt:  turn.StartTime.Format(time.RFC3339Nano),
			EventCount: len(turn.Events),
			Events:     make([]jsonEvent, 0, len(turn.Events)),
		}
		if turn.EndTime != nil {
			item.EndedAt = turn.EndTime.Format(time.RFC3339Nano)
		}

		for index, rawEvent := range turn.Events {
			event, ok := rawEvent.(game.Event)
			if !ok {
				continue
			}

			item.Events = append(item.Events, buildJSONEvent(index, event))
		}

		timeline = append(timeline, item)
	}

	encoded, err := json.MarshalIndent(timeline, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal turns json: %w", err)
	}

	return string(encoded), nil
}

func (f *JSONFormatter) FormatEvent(event game.Event) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event is nil")
	}

	encoded, err := json.MarshalIndent(buildJSONEvent(0, event), "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal event json: %w", err)
	}

	return string(encoded), nil
}

type jsonTurn struct {
	Number     int         `json:"number"`
	Player     string      `json:"player"`
	StartedAt  string      `json:"started_at"`
	EndedAt    string      `json:"ended_at,omitempty"`
	EventCount int         `json:"event_count"`
	Events     []jsonEvent `json:"events"`
}

type jsonEvent struct {
	Index      int                    `json:"index"`
	Type       string                 `json:"type"`
	Timestamp  string                 `json:"timestamp"`
	Summary    string                 `json:"summary,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func buildJSONEvent(index int, event game.Event) jsonEvent {
	item := jsonEvent{
		Index:      index,
		Type:       event.Type().String(),
		Timestamp:  event.Timestamp().Format(time.RFC3339Nano),
		Attributes: make(map[string]interface{}),
	}

	switch e := event.(type) {
	case *game.GameStartEvent:
		item.Summary = "game_start"
		item.Attributes["game_id"] = e.GameID
		item.Attributes["game_mode"] = e.GameMode
	case *game.GameEndEvent:
		item.Summary = "game_end"
		item.Attributes["winner"] = e.Winner.String()
	case *game.TurnStartEvent:
		item.Summary = "turn_start"
		item.Attributes["player"] = e.Player.String()
	case *game.TurnEndEvent:
		item.Summary = "turn_end"
	case *game.DeckListEvent:
		item.Summary = "deck_list"
		item.Attributes["player"] = e.Player.String()
		item.Attributes["name"] = e.Name
		item.Attributes["deck_id"] = e.DeckID
		item.Attributes["code"] = e.Code
		item.Attributes["card_count"] = len(e.Cards)
	case *game.CardPlayedEvent:
		item.Summary = "card_played"
		item.Attributes["player"] = e.Player.String()
		item.Attributes["position"] = e.Position
		item.Attributes["card"] = buildJSONCard(e.Card)
	case *game.CardDrawnEvent:
		item.Summary = "card_drawn"
		item.Attributes["player"] = e.Player.String()
		item.Attributes["card"] = buildJSONCard(e.Card)
	case *game.AttackEvent:
		item.Summary = "attack"
		item.Attributes["attacker_entity_id"] = e.AttackerEntityID
		item.Attributes["target_entity_id"] = e.TargetEntityID
	case *game.DamageEvent:
		item.Summary = "damage"
		item.Attributes["target_entity_id"] = e.TargetEntityID
		item.Attributes["amount"] = e.Amount
	case *game.HealEvent:
		item.Summary = "heal"
		item.Attributes["target_entity_id"] = e.TargetEntityID
		item.Attributes["amount"] = e.Amount
	case *game.DeathEvent:
		item.Summary = "death"
		item.Attributes["entity_id"] = e.EntityID
	case *game.EffectTriggeredEvent:
		item.Summary = "effect_triggered"
		item.Attributes["source_entity_id"] = e.SourceEntityID
		item.Attributes["effect_name"] = e.EffectName
	case *game.DiscoverEvent:
		item.Summary = "discover"
		item.Attributes["player"] = e.Player.String()
		item.Attributes["options"] = append([]string(nil), e.Options...)
		item.Attributes["chosen_card"] = e.ChosenCard
		item.Attributes["chosen_card_id"] = e.ChosenCardID
	default:
		item.Summary = event.Type().String()
	}

	if len(item.Attributes) == 0 {
		item.Attributes = nil
	}

	return item
}

type jsonCard struct {
	ID       string `json:"id"`
	EntityID int    `json:"entity_id"`
	Owner    string `json:"owner,omitempty"`
	Zone     string `json:"zone,omitempty"`
	Position int    `json:"position"`
}

func buildJSONCard(c *card.Card) *jsonCard {
	if c == nil {
		return nil
	}

	return &jsonCard{
		ID:       c.ID,
		EntityID: c.EntityID,
		Owner:    c.Owner.String(),
		Zone:     string(c.Zone),
		Position: c.Position,
	}
}
