package matchpackage

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/monitor"
	"hsanalyzer/internal/types"
	"hsanalyzer/pkg/contracts/matchupload"
)

const defaultMatchUploadClientName = "hs-analyzer-cli"

// BuildMatchUploadPayload converts a finalized live match into the server wire contract.
func BuildMatchUploadPayload(match monitor.PersistedMatch, clientVersion string) matchupload.Payload {
	events := buildMatchUploadEvents(match.Events)
	startedAt, endedAt := resolveMatchTimeBounds(match.Events)

	payload := matchupload.Payload{
		SchemaVersion:    matchupload.SchemaVersionV1,
		MatchID:          strings.TrimSpace(match.Key),
		CompletionStatus: matchUploadCompletionStatus(match.Complete),
		CompletionReason: matchUploadCompletionReason(match.Reason),
		Source: matchupload.Source{
			ClientName:    defaultMatchUploadClientName,
			ClientVersion: strings.TrimSpace(clientVersion),
		},
		Metadata: map[string]string{},
		Events:   events,
	}

	if match.State != nil {
		payload.GameID = strings.TrimSpace(match.State.GameID)
		if match.State.GameMode != "" {
			payload.Metadata["game_mode"] = string(match.State.GameMode)
		}
	}
	if !startedAt.IsZero() {
		payload.StartedAt = startedAt.UTC().Format(time.RFC3339Nano)
	}
	if !endedAt.IsZero() {
		payload.EndedAt = endedAt.UTC().Format(time.RFC3339Nano)
	}
	if len(payload.Metadata) == 0 {
		payload.Metadata = nil
	}
	if payload.MatchID == "" {
		payload.MatchID = payload.Hash()
	}
	if summary, ok := buildDashboardSummary(match); ok {
		_ = payload.SetDashboardSummary(summary)
	}

	return payload
}

func buildMatchUploadEvents(events []game.Event) []matchupload.Event {
	result := make([]matchupload.Event, 0, len(events))
	for index, event := range events {
		if event == nil {
			continue
		}

		payload, _ := json.Marshal(map[string]string{
			"type": event.Type().String(),
		})
		result = append(result, matchupload.Event{
			Index:     index,
			Type:      event.Type().String(),
			Timestamp: event.Timestamp().UTC().Format(time.RFC3339Nano),
			Payload:   payload,
		})
	}
	return result
}

func resolveMatchTimeBounds(events []game.Event) (time.Time, time.Time) {
	var startedAt time.Time
	var endedAt time.Time
	for _, event := range events {
		if event == nil {
			continue
		}
		timestamp := event.Timestamp()
		if timestamp.IsZero() {
			continue
		}
		if startedAt.IsZero() || timestamp.Before(startedAt) {
			startedAt = timestamp
		}
		if endedAt.IsZero() || timestamp.After(endedAt) {
			endedAt = timestamp
		}
	}
	return startedAt, endedAt
}

func matchUploadCompletionStatus(complete bool) string {
	if complete {
		return matchupload.CompletionStatusComplete
	}
	return matchupload.CompletionStatusIncomplete
}

func matchUploadCompletionReason(reason monitor.CompletionReason) string {
	switch reason {
	case monitor.CompletionReasonGameEnd:
		return matchupload.CompletionReasonGameEnd
	case monitor.CompletionReasonNextGameStart:
		return matchupload.CompletionReasonNextGameStart
	case monitor.CompletionReasonFlush:
		return matchupload.CompletionReasonFlush
	case monitor.CompletionReasonStreamEnd:
		return matchupload.CompletionReasonStreamEnd
	default:
		return string(reason)
	}
}

func eventTypeString(eventType types.EventType) string {
	return eventType.String()
}

func buildDashboardSummary(match monitor.PersistedMatch) (matchupload.DashboardSummary, bool) {
	deckEvent := findPlayerDeckListEvent(match.Events)
	if deckEvent == nil {
		return matchupload.DashboardSummary{}, false
	}

	summary := matchupload.DashboardSummary{
		DeckID:        strings.TrimSpace(deckEvent.DeckID),
		DeckName:      resolveDeckName(match, deckEvent),
		PlayerClass:   deriveDeckClass(deckEvent.Cards),
		OpponentClass: resolveOpponentClass(match),
		Result:        resolveDashboardResult(match),
		DeckCards:     buildDashboardDeckCards(deckEvent.Cards),
	}
	if summary.DeckID == "" {
		summary.DeckID = summary.DeckName
	}
	if summary.PlayerClass == "" {
		summary.PlayerClass = string(carddb.CardClassUnknown)
	}
	return summary, true
}

func findPlayerDeckListEvent(events []game.Event) *game.DeckListEvent {
	for _, current := range events {
		deckEvent, ok := current.(*game.DeckListEvent)
		if !ok || deckEvent == nil || deckEvent.Player != types.PlayerTypePlayer {
			continue
		}
		return deckEvent
	}
	return nil
}

func resolveDeckName(match monitor.PersistedMatch, deckEvent *game.DeckListEvent) string {
	if deckEvent != nil && strings.TrimSpace(deckEvent.Name) != "" {
		return strings.TrimSpace(deckEvent.Name)
	}
	if match.State != nil && match.State.Player != nil && strings.TrimSpace(match.State.Player.Name) != "" {
		return strings.TrimSpace(match.State.Player.Name)
	}
	if deckEvent != nil && strings.TrimSpace(deckEvent.DeckID) != "" {
		return strings.TrimSpace(deckEvent.DeckID)
	}
	return "Unknown Deck"
}

func resolveDashboardResult(match monitor.PersistedMatch) string {
	if match.State == nil || match.State.Winner == nil {
		return matchupload.DashboardResultUnknown
	}
	switch *match.State.Winner {
	case types.PlayerTypePlayer:
		return matchupload.DashboardResultWin
	case types.PlayerTypeOpponent:
		return matchupload.DashboardResultLoss
	default:
		return matchupload.DashboardResultUnknown
	}
}

func deriveDeckClass(cards []*card.Card) string {
	for _, current := range cards {
		info := cardInfo(current)
		if info == nil {
			continue
		}
		if info.Class != "" && info.Class != carddb.CardClassNeutral {
			return string(info.Class)
		}
	}
	return ""
}

func resolveOpponentClass(match monitor.PersistedMatch) string {
	if match.State == nil || match.State.Opponent == nil || match.State.Opponent.Hero == nil {
		return string(carddb.CardClassUnknown)
	}
	return heroCardIDToClass(match.State.Opponent.Hero.CardID)
}

func heroCardIDToClass(cardID string) string {
	cardID = strings.ToUpper(strings.TrimSpace(cardID))
	switch {
	case strings.HasPrefix(cardID, "HERO_01"):
		return string(carddb.CardClassWarrior)
	case strings.HasPrefix(cardID, "HERO_02"):
		return string(carddb.CardClassShaman)
	case strings.HasPrefix(cardID, "HERO_03"):
		return string(carddb.CardClassRogue)
	case strings.HasPrefix(cardID, "HERO_04"):
		return string(carddb.CardClassPaladin)
	case strings.HasPrefix(cardID, "HERO_05"):
		return string(carddb.CardClassHunter)
	case strings.HasPrefix(cardID, "HERO_06"):
		return string(carddb.CardClassDruid)
	case strings.HasPrefix(cardID, "HERO_07"):
		return string(carddb.CardClassWarlock)
	case strings.HasPrefix(cardID, "HERO_08"):
		return string(carddb.CardClassMage)
	case strings.HasPrefix(cardID, "HERO_09"):
		return string(carddb.CardClassPriest)
	case strings.HasPrefix(cardID, "HERO_10"):
		return string(carddb.CardClassDemonHunter)
	case strings.HasPrefix(cardID, "HERO_11"):
		return string(carddb.CardClassDeathKnight)
	default:
		return string(carddb.CardClassUnknown)
	}
}

func buildDashboardDeckCards(cards []*card.Card) []matchupload.DashboardDeckCard {
	type aggregate struct {
		card matchupload.DashboardDeckCard
	}

	grouped := make(map[string]*aggregate)
	order := make([]string, 0, len(cards))
	for _, current := range cards {
		if current == nil {
			continue
		}
		cardID := strings.TrimSpace(current.ID)
		if cardID == "" {
			continue
		}
		entry, ok := grouped[cardID]
		if !ok {
			entry = &aggregate{card: newDashboardDeckCard(current)}
			grouped[cardID] = entry
			order = append(order, cardID)
		}
		entry.card.Copies++
	}

	result := make([]matchupload.DashboardDeckCard, 0, len(order))
	for _, cardID := range order {
		result = append(result, grouped[cardID].card)
	}
	return result
}

func newDashboardDeckCard(current *card.Card) matchupload.DashboardDeckCard {
	result := matchupload.DashboardDeckCard{
		CardID: strings.TrimSpace(current.ID),
	}
	if info := cardInfo(current); info != nil {
		result.Name = strings.TrimSpace(info.Name)
		result.ManaCost = info.Cost
		result.Rarity = string(info.Rarity)
		result.IsLegendary = info.Rarity == carddb.RarityLegendary
		if result.CardID != "" {
			result.Image = fmt.Sprintf("https://art.hearthstonejson.com/v1/render/latest/ruRU/256x/%s.png", result.CardID)
		}
	} else {
		result.ManaCost = current.Cost
	}
	if result.Name == "" {
		result.Name = result.CardID
	}
	return result
}

func cardInfo(current *card.Card) *carddb.CardInfo {
	if current == nil || current.Info == nil {
		return nil
	}
	switch info := current.Info.(type) {
	case *carddb.CardInfo:
		return info
	case carddb.CardInfo:
		return &info
	default:
		return nil
	}
}
