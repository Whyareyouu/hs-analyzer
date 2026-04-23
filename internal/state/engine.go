package state

import "hs-analyzer/internal/domain"

type TurnAction struct {
	Type    string   `json:"type"`
	Player  string   `json:"player"`
	Card    string   `json:"card,omitempty"`
	Options []string `json:"options,omitempty"`
	Chosen  string   `json:"chosen,omitempty"`
	Reason  string   `json:"reason,omitempty"`
}

type TurnState struct {
	TurnNumber int          `json:"turn_number"`
	Player     string       `json:"player"`
	Actions    []TurnAction `json:"actions"`
}

type MatchState struct {
	Started bool        `json:"started"`
	Ended   bool        `json:"ended"`
	Turns   []TurnState `json:"turns"`
}

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Build(events []domain.GameEvent) MatchState {
	state := MatchState{
		Started: false,
		Ended:   false,
		Turns:   make([]TurnState, 0, 32),
	}

	currentTurn := -1
	for _, ge := range events {
		switch ev := ge.(type) {
		case domain.TurnStartEvent:
			state.Started = true
			state.Turns = append(state.Turns, TurnState{
				TurnNumber: ev.Turn,
				Player:     normalizePlayer(ev.Player),
				Actions:    []TurnAction{},
			})
			currentTurn = len(state.Turns) - 1
		case domain.DrawEvent:
			idx := ensureTurn(&state, &currentTurn, ev.Player)
			state.Turns[idx].Actions = append(state.Turns[idx].Actions, TurnAction{
				Type:   "Draw",
				Player: normalizePlayer(ev.Player),
				Card:   ev.Card,
			})
		case domain.PlayCardEvent:
			idx := ensureTurn(&state, &currentTurn, ev.Player)
			state.Turns[idx].Actions = append(state.Turns[idx].Actions, TurnAction{
				Type:   "PlayCard",
				Player: normalizePlayer(ev.Player),
				Card:   ev.Card,
			})
		case domain.DiscoverEvent:
			idx := ensureTurn(&state, &currentTurn, ev.Player)
			state.Turns[idx].Actions = append(state.Turns[idx].Actions, TurnAction{
				Type:    "Discover",
				Player:  normalizePlayer(ev.Player),
				Options: ev.Options,
				Chosen:  ev.Chosen,
			})
		case domain.UnknownEvent:
			idx := ensureTurn(&state, &currentTurn, ev.Player)
			state.Turns[idx].Actions = append(state.Turns[idx].Actions, TurnAction{
				Type:   "Unknown",
				Player: normalizePlayer(ev.Player),
				Reason: ev.Reason,
			})
		}
	}

	if state.Started && len(state.Turns) > 0 {
		state.Ended = true
	}

	return state
}

func ensureTurn(state *MatchState, currentTurn *int, player string) int {
	if *currentTurn >= 0 {
		return *currentTurn
	}
	state.Turns = append(state.Turns, TurnState{
		TurnNumber: 0,
		Player:     normalizePlayer(player),
		Actions:    []TurnAction{},
	})
	*currentTurn = len(state.Turns) - 1
	return *currentTurn
}

func normalizePlayer(player string) string {
	if player == "" {
		return "Unknown"
	}
	return player
}
