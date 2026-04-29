package main

import (
	"strings"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
)

type liveEventSideCorrector struct {
	playerDeckCardIDs map[string]struct{}
	parserLocalSide   types.PlayerType
}

func newLiveEventSideCorrector() *liveEventSideCorrector {
	return &liveEventSideCorrector{
		playerDeckCardIDs: make(map[string]struct{}),
		parserLocalSide:   types.PlayerTypeUnknown,
	}
}

func (c *liveEventSideCorrector) Correct(event game.Event) game.Event {
	if c == nil || event == nil {
		return event
	}

	switch current := event.(type) {
	case *game.DeckListEvent:
		c.trackPlayerDeck(current)
		return current
	case *game.CardDrawnEvent:
		current.Player = c.correctPlayer(current.Player)
		c.correctCardOwner(current.Card)
		return current
	case *game.CardPlayedEvent:
		c.learnLocalSideFromKnownDeckCard(current.Player, current.Card)
		current.Player = c.correctPlayer(current.Player)
		c.correctCardOwner(current.Card)
		return current
	case *game.TurnStartEvent:
		current.Player = c.correctPlayer(current.Player)
		return current
	case *game.GameEndEvent:
		current.Winner = c.correctPlayer(current.Winner)
		return current
	default:
		return current
	}
}

func (c *liveEventSideCorrector) trackPlayerDeck(event *game.DeckListEvent) {
	if event == nil || event.Player != types.PlayerTypePlayer {
		return
	}

	c.playerDeckCardIDs = make(map[string]struct{}, len(event.Cards))
	for _, current := range event.Cards {
		cardID := cardIDFromCard(current)
		if cardID == "" {
			continue
		}
		c.playerDeckCardIDs[cardID] = struct{}{}
	}
	c.parserLocalSide = types.PlayerTypeUnknown
}

func (c *liveEventSideCorrector) learnLocalSideFromKnownDeckCard(player types.PlayerType, current *card.Card) {
	if c.parserLocalSide != types.PlayerTypeUnknown {
		return
	}
	if player != types.PlayerTypePlayer && player != types.PlayerTypeOpponent {
		return
	}
	cardID := cardIDFromCard(current)
	if cardID == "" {
		return
	}
	if _, ok := c.playerDeckCardIDs[cardID]; !ok {
		return
	}
	c.parserLocalSide = player
}

func (c *liveEventSideCorrector) correctPlayer(player types.PlayerType) types.PlayerType {
	if c.parserLocalSide == types.PlayerTypeUnknown || c.parserLocalSide == types.PlayerTypePlayer {
		return player
	}
	if c.parserLocalSide != types.PlayerTypeOpponent {
		return player
	}

	switch player {
	case types.PlayerTypePlayer:
		return types.PlayerTypeOpponent
	case types.PlayerTypeOpponent:
		return types.PlayerTypePlayer
	default:
		return player
	}
}

func (c *liveEventSideCorrector) correctCardOwner(current *card.Card) {
	if current == nil {
		return
	}
	current.Owner = c.correctPlayer(current.Owner)
	current.Controller = c.correctPlayer(current.Controller)
}

func cardIDFromCard(current *card.Card) string {
	if current == nil {
		return ""
	}
	return strings.TrimSpace(current.ID)
}
