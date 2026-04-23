package timeline

import (
	"fmt"

	"hs-analyzer/internal/state"
)

type TimelineTurn struct {
	TurnNumber int      `json:"turn_number"`
	Player     string   `json:"player"`
	Actions    []string `json:"actions"`
}

type Timeline struct {
	Turns []TimelineTurn `json:"turns"`
}

func Build(match state.MatchState) Timeline {
	out := Timeline{Turns: make([]TimelineTurn, 0, len(match.Turns))}
	for _, turn := range match.Turns {
		actions := make([]string, 0, len(turn.Actions))
		for _, action := range turn.Actions {
			actions = append(actions, formatAction(action))
		}
		out.Turns = append(out.Turns, TimelineTurn{
			TurnNumber: turn.TurnNumber,
			Player:     turn.Player,
			Actions:    actions,
		})
	}
	return out
}

func formatAction(a state.TurnAction) string {
	switch a.Type {
	case "Draw":
		return fmt.Sprintf("%s draws %s", a.Player, safe(a.Card))
	case "PlayCard":
		return fmt.Sprintf("%s plays %s", a.Player, safe(a.Card))
	case "Discover":
		return fmt.Sprintf("%s discovers %s from %v", a.Player, safe(a.Chosen), a.Options)
	default:
		return fmt.Sprintf("%s unknown action (%s)", a.Player, safe(a.Reason))
	}
}

func safe(v string) string {
	if v == "" {
		return "Unknown"
	}
	return v
}
