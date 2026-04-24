package formatter

import (
	"context"
	"fmt"
	"time"

	"hsanalyzer/internal/analyzer"
	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/player"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

type ReplayJSONFormatter struct {
	*JSONFormatter
}

func NewReplayJSONFormatter(cardDB carddb.CardDatabase) *ReplayJSONFormatter {
	return &ReplayJSONFormatter{JSONFormatter: NewJSONFormatter(cardDB)}
}

func (f *ReplayJSONFormatter) Name() string {
	return "json-replay"
}

type replayTurn struct {
	Number       int                      `json:"number"`
	ActivePlayer string                   `json:"active_player"`
	StartedAt    string                   `json:"started_at"`
	EndedAt      string                   `json:"ended_at,omitempty"`
	EventCount   int                      `json:"event_count"`
	OpeningState replaySnapshot           `json:"opening_state"`
	Events       []map[string]interface{} `json:"events"`
}

type replaySnapshot struct {
	CurrentTurn   int                             `json:"current_turn"`
	CurrentPlayer string                          `json:"current_player"`
	Players       map[string]replayPlayerSnapshot `json:"players"`
}

type replayPlayerSnapshot struct {
	ID             int         `json:"id"`
	Name           string      `json:"name,omitempty"`
	Mana           int         `json:"mana"`
	MaxMana        int         `json:"max_mana"`
	Overload       int         `json:"overload"`
	HandCount      int         `json:"hand_count"`
	DeckCount      int         `json:"deck_count"`
	BoardCount     int         `json:"board_count"`
	GraveyardCount int         `json:"graveyard_count"`
	Hand           []*jsonCard `json:"hand,omitempty"`
	Deck           []*jsonCard `json:"known_deck,omitempty"`
	Board          []*jsonCard `json:"board,omitempty"`
	Graveyard      []*jsonCard `json:"graveyard,omitempty"`
	DeckList       []*jsonCard `json:"deck_list,omitempty"`
}

type replayDelta struct {
	ResourceChanges map[string]map[string]int `json:"resource_changes,omitempty"`
	ZoneChanges     map[string]map[string]int `json:"zone_changes,omitempty"`
}

func (f *ReplayJSONFormatter) Format(gameState *state.State) (string, error) {
	if gameState == nil {
		return "", fmt.Errorf("game state is nil")
	}

	matchStates := extractMatchStates(gameState)
	documents := make([]map[string]interface{}, 0, len(matchStates))
	for _, matchState := range matchStates {
		if matchState == nil {
			continue
		}
		document, err := f.buildReplayMatchDocument(matchState)
		if err != nil {
			return "", err
		}
		documents = append(documents, document)
	}

	encoded, err := marshalIndentedJSON(wrapMatchDocuments("replay", documents))
	if err != nil {
		return "", fmt.Errorf("marshal replay json: %w", err)
	}
	return encoded, nil
}

func (f *ReplayJSONFormatter) buildReplayMatchDocument(gameState *state.State) (map[string]interface{}, error) {
	replayState := state.New(gameState.GameID, gameState.GameMode)
	pipeline := analyzer.NewPipeline()
	if err := pipeline.RegisterHandler(analyzer.NewApplyEventHandler()); err != nil {
		return nil, fmt.Errorf("register replay pipeline: %w", err)
	}
	if err := f.applyReplayPrelude(pipeline, replayState, gameState.Events); err != nil {
		return nil, fmt.Errorf("apply replay prelude: %w", err)
	}

	turns := make([]replayTurn, 0, len(gameState.Turns))
	globalIndex := 0

	for _, turn := range gameState.Turns {
		if turn == nil {
			continue
		}

		replayTurnEntry := replayTurn{
			Number:       turn.Number,
			ActivePlayer: turn.Player.String(),
			StartedAt:    turn.StartTime.Format(time.RFC3339Nano),
			EventCount:   len(turn.Events),
			Events:       make([]map[string]interface{}, 0, len(turn.Events)),
		}
		if turn.EndTime != nil {
			replayTurnEntry.EndedAt = turn.EndTime.Format(time.RFC3339Nano)
		}

		for index, rawEvent := range turn.Events {
			event, ok := rawEvent.(game.Event)
			if !ok || event == nil {
				continue
			}

			before := f.buildReplaySnapshot(replayState)
			playCost := f.resolvePlayCost(replayState, event)

			if err := pipeline.Process(context.Background(), replayState, event); err != nil {
				return nil, fmt.Errorf("replay process event %d: %w", globalIndex, err)
			}
			if playCost > 0 {
				f.applyPlayCost(replayState, event, playCost)
			}

			after := f.buildReplaySnapshot(replayState)
			replayTurnEntry.Events = append(replayTurnEntry.Events, f.buildReplayEvent(globalIndex, index, event, before, after, replayState))

			if replayTurnEntry.OpeningState.Players == nil && shouldCaptureOpeningState(turn.Events, index, event) {
				replayTurnEntry.OpeningState = after
			}

			globalIndex++
		}

		if replayTurnEntry.OpeningState.Players == nil {
			replayTurnEntry.OpeningState = f.buildReplaySnapshot(replayState)
		}
		turns = append(turns, replayTurnEntry)
	}

	return map[string]interface{}{
		"game": map[string]interface{}{
			"id":          gameState.GameID,
			"mode":        gameState.GameMode,
			"winner":      winnerString(gameState),
			"turn_count":  gameState.GetTurnCount(),
			"event_count": gameState.GetEventCount(),
		},
		"turns":       turns,
		"final_state": f.buildReplaySnapshot(replayState),
	}, nil
}

func (f *ReplayJSONFormatter) FormatEvent(event game.Event) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event is nil")
	}

	replayState := state.New("", state.GameModeUnknown)
	before := f.buildReplaySnapshot(replayState)
	pipeline := analyzer.NewPipeline()
	if err := pipeline.RegisterHandler(analyzer.NewApplyEventHandler()); err != nil {
		return "", err
	}
	if err := pipeline.Process(context.Background(), replayState, event); err != nil {
		return "", err
	}
	after := f.buildReplaySnapshot(replayState)
	return marshalIndentedJSON(f.buildReplayEvent(0, 0, event, before, after, replayState))
}

func collectReplayTurnOpenings(f *JSONFormatter, gameState *state.State) (map[int][]*jsonCard, error) {
	replay := NewReplayJSONFormatter(f.cardDB)
	replayState := state.New(gameState.GameID, gameState.GameMode)
	pipeline := analyzer.NewPipeline()
	if err := pipeline.RegisterHandler(analyzer.NewApplyEventHandler()); err != nil {
		return nil, err
	}
	if err := replay.applyReplayPrelude(pipeline, replayState, gameState.Events); err != nil {
		return nil, err
	}

	result := make(map[int][]*jsonCard)
	for _, turn := range gameState.Turns {
		if turn == nil {
			continue
		}
		for index, rawEvent := range turn.Events {
			event, ok := rawEvent.(game.Event)
			if !ok || event == nil {
				continue
			}
			playCost := replay.resolvePlayCost(replayState, event)
			if err := pipeline.Process(context.Background(), replayState, event); err != nil {
				return nil, err
			}
			if playCost > 0 {
				replay.applyPlayCost(replayState, event, playCost)
			}
			if _, exists := result[turn.Number]; !exists && shouldCaptureOpeningState(turn.Events, index, event) {
				result[turn.Number] = replay.buildReplaySnapshot(replayState).Players[turn.Player.String()].Hand
			}
		}
	}

	return result, nil
}

func shouldCaptureOpeningState(turnEvents []interface{}, index int, event game.Event) bool {
	switch current := event.(type) {
	case *game.TurnStartEvent:
		if index+1 >= len(turnEvents) {
			return true
		}
		next, ok := turnEvents[index+1].(game.Event)
		return !ok || !isOpeningDraw(next, current.Player)
	case *game.CardDrawnEvent:
		if index+1 >= len(turnEvents) {
			return true
		}
		next, ok := turnEvents[index+1].(game.Event)
		return !ok || !isOpeningDraw(next, current.Player)
	default:
		return false
	}
}

func isOpeningDraw(event game.Event, playerType types.PlayerType) bool {
	draw, ok := event.(*game.CardDrawnEvent)
	return ok && draw.Player == playerType
}

func (f *ReplayJSONFormatter) resolvePlayCost(gameState *state.State, event game.Event) int {
	played, ok := event.(*game.CardPlayedEvent)
	if !ok || played == nil {
		return 0
	}
	playerState := gameState.GetPlayer(played.Player)
	if playerState != nil {
		if current := playerState.Hand.GetCard(played.Card.EntityID); current != nil {
			return currentCardCost(current, f.resolveCardInfo(current))
		}
	}
	return currentCardCost(played.Card, f.resolveCardInfo(played.Card))
}

func (f *ReplayJSONFormatter) applyPlayCost(gameState *state.State, event game.Event, cost int) {
	played, ok := event.(*game.CardPlayedEvent)
	if !ok || played == nil {
		return
	}
	playerState := gameState.GetPlayer(played.Player)
	if playerState != nil {
		_ = playerState.SpendMana(cost)
	}
}

func (f *ReplayJSONFormatter) buildReplayEvent(globalIndex, turnIndex int, event game.Event, before, after replaySnapshot, gameState *state.State) map[string]interface{} {
	entry := map[string]interface{}{
		"global_index": globalIndex,
		"turn_index":   turnIndex,
		"type":         event.Type().String(),
		"timestamp":    event.Timestamp().Format(time.RFC3339Nano),
		"action":       f.buildReplayAction(event, gameState),
		"delta":        buildReplayDelta(before, after),
	}
	if isStatefulReplayEvent(event) {
		entry["state_before"] = before
		entry["state_after"] = after
	}
	return entry
}

func (f *ReplayJSONFormatter) buildReplayAction(event game.Event, gameState *state.State) map[string]interface{} {
	action := map[string]interface{}{"summary": event.Type().String()}
	switch e := event.(type) {
	case *game.TurnStartEvent:
		action["player"] = e.Player.String()
	case *game.CardDrawnEvent:
		action["player"] = e.Player.String()
		action["card"] = f.buildJSONCard(resolveCardFromState(gameState, e.Card))
	case *game.CardPlayedEvent:
		action["player"] = e.Player.String()
		action["position"] = e.Position
		action["card"] = f.buildJSONCard(resolveCardFromState(gameState, e.Card))
	case *game.CardCostChangedEvent:
		action["entity_id"] = e.EntityID
		action["card_id"] = e.CardID
		action["cost"] = e.Cost
	case *game.AttackEvent:
		action["attacker"] = f.buildJSONEventEntityRef(gameState, e.AttackerEntityID, e.AttackerCardID, e.AttackerName)
		action["target"] = f.buildJSONEventEntityRef(gameState, e.TargetEntityID, e.TargetCardID, e.TargetName)
	case *game.DamageEvent:
		action["target"] = f.buildJSONEventEntityRef(gameState, e.TargetEntityID, e.TargetCardID, e.TargetName)
		action["amount"] = e.Amount
	case *game.HealEvent:
		action["target"] = f.buildJSONEventEntityRef(gameState, e.TargetEntityID, e.TargetCardID, e.TargetName)
		action["amount"] = e.Amount
	case *game.DeathEvent:
		action["entity"] = f.buildJSONEventEntityRef(gameState, e.EntityID, e.CardID, e.Name)
	case *game.EffectTriggeredEvent:
		sourceName, sourceText := f.resolveEffectSource(e)
		details := describeEffect(e.EffectName, sourceName, sourceText, "ru")
		action["source"] = f.buildJSONEventEntityRef(gameState, e.SourceEntityID, e.SourceCardID, e.SourceName)
		action["effect_name"] = e.EffectName
		action["effect_label"] = details.Label
		action["effect_description"] = details.Description
	case *game.DiscoverEvent:
		action["player"] = e.Player.String()
		action["options"] = f.buildJSONCardRefs(e.Options)
		action["chosen_card"] = f.buildJSONCardByID(e.ChosenCard)
	}
	return action
}

func (f *ReplayJSONFormatter) buildReplaySnapshot(gameState *state.State) replaySnapshot {
	snapshot := replaySnapshot{
		CurrentTurn:   0,
		CurrentPlayer: types.PlayerTypeUnknown.String(),
		Players: map[string]replayPlayerSnapshot{
			types.PlayerTypePlayer.String():   {},
			types.PlayerTypeOpponent.String(): {},
		},
	}
	if gameState == nil {
		return snapshot
	}
	snapshot.CurrentTurn = gameState.CurrentTurn
	snapshot.CurrentPlayer = gameState.CurrentPlayer.String()
	snapshot.Players[types.PlayerTypePlayer.String()] = f.buildReplayPlayerSnapshot(gameState.Player)
	snapshot.Players[types.PlayerTypeOpponent.String()] = f.buildReplayPlayerSnapshot(gameState.Opponent)
	return snapshot
}

func (f *ReplayJSONFormatter) buildReplayPlayerSnapshot(playerState *player.Player) replayPlayerSnapshot {
	if playerState == nil {
		return replayPlayerSnapshot{}
	}
	return replayPlayerSnapshot{
		ID:             playerState.ID,
		Name:           playerState.Name,
		Mana:           playerState.Mana,
		MaxMana:        playerState.MaxMana,
		Overload:       playerState.OverloadMana,
		HandCount:      playerState.GetHandSize(),
		DeckCount:      playerState.GetDeckSize(),
		BoardCount:     playerState.GetBoardSize(),
		GraveyardCount: playerState.Graveyard.Count(),
		Hand:           f.buildJSONCards(playerState.Hand.GetCards()),
		Deck:           f.buildJSONCards(playerState.Deck.GetCards()),
		Board:          f.buildJSONCards(playerState.Board.GetCards()),
		Graveyard:      f.buildJSONCards(playerState.Graveyard.GetCards()),
		DeckList:       f.buildJSONCards(playerState.DeckList),
	}
}

func buildReplayDelta(before, after replaySnapshot) *replayDelta {
	delta := &replayDelta{
		ResourceChanges: make(map[string]map[string]int),
		ZoneChanges:     make(map[string]map[string]int),
	}
	for _, playerKey := range []string{types.PlayerTypePlayer.String(), types.PlayerTypeOpponent.String()} {
		b := before.Players[playerKey]
		a := after.Players[playerKey]
		if b.Mana != a.Mana || b.MaxMana != a.MaxMana || b.Overload != a.Overload {
			delta.ResourceChanges[playerKey] = map[string]int{
				"mana_before":     b.Mana,
				"mana_after":      a.Mana,
				"max_mana_before": b.MaxMana,
				"max_mana_after":  a.MaxMana,
				"overload_before": b.Overload,
				"overload_after":  a.Overload,
			}
		}
		if b.HandCount != a.HandCount || b.DeckCount != a.DeckCount || b.BoardCount != a.BoardCount || b.GraveyardCount != a.GraveyardCount {
			delta.ZoneChanges[playerKey] = map[string]int{
				"hand_before":      b.HandCount,
				"hand_after":       a.HandCount,
				"deck_before":      b.DeckCount,
				"deck_after":       a.DeckCount,
				"board_before":     b.BoardCount,
				"board_after":      a.BoardCount,
				"graveyard_before": b.GraveyardCount,
				"graveyard_after":  a.GraveyardCount,
			}
		}
	}
	if len(delta.ResourceChanges) == 0 {
		delta.ResourceChanges = nil
	}
	if len(delta.ZoneChanges) == 0 {
		delta.ZoneChanges = nil
	}
	if delta.ResourceChanges == nil && delta.ZoneChanges == nil {
		return nil
	}
	return delta
}

func isStatefulReplayEvent(event game.Event) bool {
	switch event.(type) {
	case *game.TurnStartEvent, *game.TurnEndEvent, *game.CardDrawnEvent, *game.CardPlayedEvent, *game.CardCostChangedEvent, *game.DamageEvent, *game.HealEvent, *game.DeathEvent, *game.DiscoverEvent:
		return true
	default:
		return false
	}
}

func resolveCardFromState(gameState *state.State, fallback *card.Card) *card.Card {
	if gameState == nil || fallback == nil {
		return fallback
	}
	if current := gameState.GetCard(fallback.EntityID); current != nil {
		return current
	}
	return fallback
}

func (f *ReplayJSONFormatter) applyReplayPrelude(pipeline *analyzer.Pipeline, replayState *state.State, events []interface{}) error {
	if pipeline == nil || replayState == nil {
		return nil
	}

	for _, rawEvent := range events {
		event, ok := rawEvent.(game.Event)
		if !ok || event == nil {
			continue
		}
		if _, isTurnStart := event.(*game.TurnStartEvent); isTurnStart {
			return nil
		}

		playCost := f.resolvePlayCost(replayState, event)

		if err := pipeline.Process(context.Background(), replayState, event); err != nil {
			return err
		}
		if playCost > 0 {
			f.applyPlayCost(replayState, event, playCost)
		}
	}

	return nil
}
