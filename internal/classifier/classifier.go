package classifier

import (
	"strconv"
	"strings"

	"hs-analyzer/internal/domain"
)

type Classifier struct{}

func New() *Classifier {
	return &Classifier{}
}

func (c *Classifier) Classify(rawEvents []domain.RawEvent) []domain.GameEvent {
	out := make([]domain.GameEvent, 0, len(rawEvents))
	for _, raw := range rawEvents {
		out = append(out, c.classifyOne(raw))
	}
	return out
}

func (c *Classifier) classifyOne(raw domain.RawEvent) domain.GameEvent {
	player := strOrUnknown(raw.Data, "player")

	switch strings.ToUpper(raw.Type) {
	case "TURN_START":
		turn := intFromData(raw.Data, "turn")
		if turn <= 0 {
			return domain.UnknownEvent{
				Player:   player,
				Reason:   "turn_start_missing_turn",
				RawEvent: raw,
			}
		}
		return domain.TurnStartEvent{
			Player:   player,
			Turn:     turn,
			RawEvent: raw,
		}
	case "DRAW":
		card := strOrUnknown(raw.Data, "card")
		if card == "Unknown" {
			return domain.UnknownEvent{
				Player:   player,
				Reason:   "draw_missing_card",
				RawEvent: raw,
			}
		}
		return domain.DrawEvent{Player: player, Card: card, RawEvent: raw}
	case "PLAY":
		card := strOrUnknown(raw.Data, "card")
		if card == "Unknown" {
			return domain.UnknownEvent{
				Player:   player,
				Reason:   "play_missing_card",
				RawEvent: raw,
			}
		}
		return domain.PlayCardEvent{Player: player, Card: card, RawEvent: raw}
	case "DISCOVER":
		options := splitCSV(strOrUnknown(raw.Data, "options"))
		chosen := strOrUnknown(raw.Data, "chosen")
		return domain.DiscoverEvent{
			Player:   player,
			Options:  options,
			Chosen:   chosen,
			RawEvent: raw,
		}
	default:
		return domain.UnknownEvent{
			Player:   player,
			Reason:   "unsupported_raw_type",
			RawEvent: raw,
		}
	}
}

func strOrUnknown(data map[string]interface{}, key string) string {
	v, ok := data[key]
	if !ok {
		return "Unknown"
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return "Unknown"
	}
	return s
}

func intFromData(data map[string]interface{}, key string) int {
	v, ok := data[key]
	if !ok {
		return 0
	}
	s, ok := v.(string)
	if !ok {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

func splitCSV(value string) []string {
	if value == "Unknown" || strings.TrimSpace(value) == "" {
		return []string{"Unknown"}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return []string{"Unknown"}
	}
	return out
}
