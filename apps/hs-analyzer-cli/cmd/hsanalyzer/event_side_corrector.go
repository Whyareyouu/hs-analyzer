package main

import (
	"strings"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
)

type liveEventSideCorrector struct {
	playerDeckCardIDs   map[string]struct{}
	parserLocalSide     types.PlayerType
	currentTurnPlayer   types.PlayerType
	seenDeckCardsBySide map[types.PlayerType]map[string]struct{}
	pendingDeckDraws    []pendingDeckDrawCorrection
}

type pendingDeckDrawCorrection struct {
	event      *game.CardDrawnEvent
	player     types.PlayerType
	owner      types.PlayerType
	controller types.PlayerType
}

func newLiveEventSideCorrector() *liveEventSideCorrector {
	return &liveEventSideCorrector{
		playerDeckCardIDs: make(map[string]struct{}),
		parserLocalSide:   types.PlayerTypeUnknown,
		currentTurnPlayer: types.PlayerTypeUnknown,
		seenDeckCardsBySide: map[types.PlayerType]map[string]struct{}{
			types.PlayerTypePlayer:   make(map[string]struct{}),
			types.PlayerTypeOpponent: make(map[string]struct{}),
		},
		pendingDeckDraws: make([]pendingDeckDrawCorrection, 0),
	}
}

func (c *liveEventSideCorrector) Correct(event game.Event) game.Event {
	if c == nil || event == nil {
		return event
	}

	switch current := event.(type) {
	case *game.GameStartEvent:
		c.currentTurnPlayer = types.PlayerTypeUnknown
		return current
	case *game.DeckListEvent:
		c.trackPlayerDeck(current)
		return current
	case *game.CardDrawnEvent:
		c.correctCardDrawn(current)
		return current
	case *game.CardPlayedEvent:
		c.learnLocalSideFromKnownDeckCard(current.Player, current.Card)
		c.registerDeckCardEvidence(current.Player, current.Card)
		c.correctPendingDeckDraws()
		current.Player = c.correctPlayer(current.Player)
		c.correctCardOwner(current.Card)
		return current
	case *game.TurnStartEvent:
		current.Player = c.correctPlayer(current.Player)
		c.currentTurnPlayer = current.Player
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
	c.currentTurnPlayer = types.PlayerTypeUnknown
	c.seenDeckCardsBySide = map[types.PlayerType]map[string]struct{}{
		types.PlayerTypePlayer:   make(map[string]struct{}),
		types.PlayerTypeOpponent: make(map[string]struct{}),
	}
	c.pendingDeckDraws = c.pendingDeckDraws[:0]
}

func (c *liveEventSideCorrector) correctCardDrawn(event *game.CardDrawnEvent) {
	if c == nil || event == nil {
		return
	}

	rawPlayer := event.Player
	rawOwner, rawController := rawCardOwnership(event.Card)
	c.registerDeckCardEvidence(rawPlayer, event.Card)
	if c.parserLocalSide == types.PlayerTypeUnknown {
		c.rememberPendingDeckDraw(event, rawPlayer, rawOwner, rawController)
		return
	}

	event.Player = c.correctPlayer(rawPlayer)
	c.correctCardOwnerFromRaw(event.Card, rawOwner, rawController)
	c.correctPendingDeckDraws()
}

func (c *liveEventSideCorrector) learnLocalSideFromKnownDeckCard(player types.PlayerType, current *card.Card) {
	if c.parserLocalSide != types.PlayerTypeUnknown {
		return
	}
	if c.currentTurnPlayer != types.PlayerTypePlayer {
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

func (c *liveEventSideCorrector) registerDeckCardEvidence(player types.PlayerType, current *card.Card) {
	if c == nil || c.parserLocalSide != types.PlayerTypeUnknown {
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

	sideEvidence, ok := c.seenDeckCardsBySide[player]
	if !ok {
		sideEvidence = make(map[string]struct{})
		c.seenDeckCardsBySide[player] = sideEvidence
	}
	sideEvidence[cardID] = struct{}{}

	playerCount := len(c.seenDeckCardsBySide[types.PlayerTypePlayer])
	opponentCount := len(c.seenDeckCardsBySide[types.PlayerTypeOpponent])
	switch {
	case playerCount >= 2 && opponentCount == 0:
		c.parserLocalSide = types.PlayerTypePlayer
	case opponentCount >= 2 && playerCount == 0:
		c.parserLocalSide = types.PlayerTypeOpponent
	}
}

func (c *liveEventSideCorrector) rememberPendingDeckDraw(event *game.CardDrawnEvent, player, owner, controller types.PlayerType) {
	if c == nil || event == nil {
		return
	}
	if player != types.PlayerTypePlayer && player != types.PlayerTypeOpponent {
		return
	}
	cardID := cardIDFromCard(event.Card)
	if cardID == "" {
		return
	}
	if _, ok := c.playerDeckCardIDs[cardID]; !ok {
		return
	}

	const maxPendingDeckDraws = 16
	correction := pendingDeckDrawCorrection{
		event:      event,
		player:     player,
		owner:      owner,
		controller: controller,
	}
	if len(c.pendingDeckDraws) >= maxPendingDeckDraws {
		copy(c.pendingDeckDraws, c.pendingDeckDraws[1:])
		c.pendingDeckDraws[len(c.pendingDeckDraws)-1] = correction
		return
	}
	c.pendingDeckDraws = append(c.pendingDeckDraws, correction)
}

func (c *liveEventSideCorrector) correctPendingDeckDraws() {
	if c == nil || c.parserLocalSide == types.PlayerTypeUnknown || len(c.pendingDeckDraws) == 0 {
		return
	}
	for _, pending := range c.pendingDeckDraws {
		if pending.event == nil {
			continue
		}
		pending.event.Player = c.correctPlayer(pending.player)
		c.correctCardOwnerFromRaw(pending.event.Card, pending.owner, pending.controller)
	}
	c.pendingDeckDraws = c.pendingDeckDraws[:0]
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

func (c *liveEventSideCorrector) correctCardOwnerFromRaw(current *card.Card, owner, controller types.PlayerType) {
	if current == nil {
		return
	}
	current.Owner = c.correctPlayer(owner)
	current.Controller = c.correctPlayer(controller)
}

func cardIDFromCard(current *card.Card) string {
	if current == nil {
		return ""
	}
	return strings.TrimSpace(current.ID)
}

func rawCardOwnership(current *card.Card) (types.PlayerType, types.PlayerType) {
	if current == nil {
		return types.PlayerTypeUnknown, types.PlayerTypeUnknown
	}
	return current.Owner, current.Controller
}
