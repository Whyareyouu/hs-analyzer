package formatter

import (
	"encoding/json"
	"strings"
	"testing"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

type stubCardDatabase struct {
	cards map[string]*carddb.CardInfo
}

func (db *stubCardDatabase) Load(string) error {
	return nil
}

func (db *stubCardDatabase) GetCard(id string) (*carddb.CardInfo, bool) {
	cardInfo, ok := db.cards[id]
	return cardInfo, ok
}

func (db *stubCardDatabase) GetCardByDBFID(int) (*carddb.CardInfo, bool) {
	return nil, false
}

func (db *stubCardDatabase) GetCardByName(string) (*carddb.CardInfo, bool) {
	return nil, false
}

func TestJSONFormatterFormatStateAsTurnsArray(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter(nil)
	gameState := state.New("game-1", state.GameModeStandard)

	turnStart := game.NewTurnStartEvent(types.PlayerTypePlayer)
	if err := turnStart.Apply(gameState); err != nil {
		t.Fatalf("apply turn start: %v", err)
	}
	gameState.AddEvent(turnStart)

	drawn := card.New("CARD_DRAW", 101, types.PlayerTypePlayer)
	drawEvent := game.NewCardDrawnEvent(types.PlayerTypePlayer, drawn)
	if err := drawEvent.Apply(gameState); err != nil {
		t.Fatalf("apply draw: %v", err)
	}
	gameState.AddEvent(drawEvent)

	turnEnd := game.NewTurnEndEvent()
	if err := turnEnd.Apply(gameState); err != nil {
		t.Fatalf("apply turn end: %v", err)
	}
	gameState.AddEvent(turnEnd)

	raw, err := formatter.Format(gameState)
	if err != nil {
		t.Fatalf("format state: %v", err)
	}

	var turns []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &turns); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	if len(turns) != 1 {
		t.Fatalf("unexpected turns count: %d", len(turns))
	}
	if turns[0]["number"].(float64) != 1 {
		t.Fatalf("unexpected turn number: %#v", turns[0]["number"])
	}
	events, ok := turns[0]["events"].([]interface{})
	if !ok || len(events) != 3 {
		t.Fatalf("unexpected turn events: %#v", turns[0]["events"])
	}
}

func TestJSONFormatterFormatEvent(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter(nil)
	raw, err := formatter.FormatEvent(game.NewAttackEvent(10, 20))
	if err != nil {
		t.Fatalf("format event: %v", err)
	}

	if !strings.Contains(raw, `"type": "attack"`) {
		t.Fatalf("unexpected json: %s", raw)
	}
	if !strings.Contains(raw, `"attacker_entity_id": 10`) {
		t.Fatalf("unexpected json: %s", raw)
	}
}

func TestJSONFormatterAddsFullCardInfoForPlayedCard(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter(&stubCardDatabase{
		cards: map[string]*carddb.CardInfo{
			"TLC_515": {
				ID:          "TLC_515",
				DBFID:       12345,
				Name:        "Cultist Map",
				Text:        "Deal damage.",
				Cost:        2,
				Type:        carddb.CardTypeSpell,
				Rarity:      carddb.RarityCommon,
				Class:       carddb.CardClassRogue,
				Set:         "THE_LOST_CITY",
				Collectible: true,
			},
		},
	})

	playedCard := card.New("TLC_515", 60, types.PlayerTypePlayer)
	playedCard.MoveTo(card.ZonePlay, 1)

	raw, err := formatter.FormatEvent(game.NewCardPlayedEvent(types.PlayerTypePlayer, playedCard, 1))
	if err != nil {
		t.Fatalf("format event: %v", err)
	}

	var event struct {
		Attributes struct {
			Card struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Info struct {
					ID          string `json:"id"`
					DBFID       int    `json:"dbf_id"`
					Name        string `json:"name"`
					Text        string `json:"text"`
					Cost        int    `json:"cost"`
					Type        string `json:"type"`
					Rarity      string `json:"rarity"`
					CardClass   string `json:"card_class"`
					Set         string `json:"set"`
					Collectible bool   `json:"collectible"`
				} `json:"info"`
			} `json:"card"`
		} `json:"attributes"`
	}
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}

	if event.Attributes.Card.ID != "TLC_515" {
		t.Fatalf("unexpected card id: %#v", event.Attributes.Card.ID)
	}
	if event.Attributes.Card.Name != "Cultist Map" {
		t.Fatalf("unexpected card name: %#v", event.Attributes.Card.Name)
	}
	if event.Attributes.Card.Info.ID != "TLC_515" {
		t.Fatalf("unexpected card info id: %#v", event.Attributes.Card.Info.ID)
	}
	if event.Attributes.Card.Info.DBFID != 12345 {
		t.Fatalf("unexpected dbf id: %#v", event.Attributes.Card.Info.DBFID)
	}
	if event.Attributes.Card.Info.Name != "Cultist Map" {
		t.Fatalf("unexpected info name: %#v", event.Attributes.Card.Info.Name)
	}
	if event.Attributes.Card.Info.Text != "Deal damage." {
		t.Fatalf("unexpected info text: %#v", event.Attributes.Card.Info.Text)
	}
	if event.Attributes.Card.Info.Cost != 2 {
		t.Fatalf("unexpected info cost: %#v", event.Attributes.Card.Info.Cost)
	}
	if event.Attributes.Card.Info.Type != "SPELL" {
		t.Fatalf("unexpected info type: %#v", event.Attributes.Card.Info.Type)
	}
	if event.Attributes.Card.Info.Rarity != "COMMON" {
		t.Fatalf("unexpected info rarity: %#v", event.Attributes.Card.Info.Rarity)
	}
	if event.Attributes.Card.Info.CardClass != "ROGUE" {
		t.Fatalf("unexpected info class: %#v", event.Attributes.Card.Info.CardClass)
	}
	if event.Attributes.Card.Info.Set != "THE_LOST_CITY" {
		t.Fatalf("unexpected info set: %#v", event.Attributes.Card.Info.Set)
	}
	if !event.Attributes.Card.Info.Collectible {
		t.Fatalf("expected collectible card")
	}
}

func TestJSONFormatterRejectsNilInputs(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter(nil)
	if _, err := formatter.Format(nil); err == nil {
		t.Fatalf("expected nil state error")
	}
	if _, err := formatter.FormatEvent(nil); err == nil {
		t.Fatalf("expected nil event error")
	}
}

func TestJSONFormatterHumanizesEffectAndCardText(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter(&stubCardDatabase{
		cards: map[string]*carddb.CardInfo{
			"TLC_515": {
				ID:   "TLC_515",
				Name: "Карта сектантов",
				Text: "[x]Вы <b>раскапываете</b> карту\nиз колоды.",
			},
		},
	})

	raw, err := formatter.FormatEvent(game.NewDetailedEffectTriggeredEvent(60, "TLC_515", "Карта сектантов", "TRIGGER_VISUAL"))
	if err != nil {
		t.Fatalf("format event: %v", err)
	}

	if strings.Contains(raw, `\u003c`) {
		t.Fatalf("html should not be escaped: %s", raw)
	}
	if !strings.Contains(raw, `"summary": "Срабатывание эффекта"`) {
		t.Fatalf("unexpected summary: %s", raw)
	}
	if !strings.Contains(raw, `"source_card_text": "Вы раскапываете карту\nиз колоды."`) {
		t.Fatalf("unexpected card text: %s", raw)
	}
}

func TestJSONFormatterAddsHumanReadableDataForEntityAndDiscoverEvents(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter(&stubCardDatabase{
		cards: map[string]*carddb.CardInfo{
			"ATTACKER_001": {
				ID:     "ATTACKER_001",
				Name:   "Учебный атакующий",
				Text:   "Наносит урон.",
				Cost:   3,
				Attack: 4,
				Health: 5,
				Type:   carddb.CardTypeMinion,
			},
			"TARGET_001": {
				ID:     "TARGET_001",
				Name:   "Учебная цель",
				Text:   "Получает урон.",
				Cost:   2,
				Attack: 2,
				Health: 3,
				Type:   carddb.CardTypeMinion,
			},
			"DISCOVER_001": {
				ID:   "DISCOVER_001",
				Name: "Первая находка",
				Text: "Первая карта для выбора.",
				Type: carddb.CardTypeSpell,
			},
			"DISCOVER_002": {
				ID:   "DISCOVER_002",
				Name: "Вторая находка",
				Text: "Вторая карта для выбора.",
				Type: carddb.CardTypeSpell,
			},
		},
	})

	gameState := state.New("game-1", state.GameModeStandard)
	turnStart := game.NewTurnStartEvent(types.PlayerTypePlayer)
	if err := turnStart.Apply(gameState); err != nil {
		t.Fatalf("apply turn start: %v", err)
	}
	gameState.AddEvent(turnStart)

	attacker := card.New("ATTACKER_001", 10, types.PlayerTypePlayer)
	attacker.MoveTo(card.ZonePlay, 0)
	gameState.Player.Board.AddCardAt(attacker, 0)
	attackerEntity := gameState.TrackCardEntity(attacker)
	attackerEntity.SetStats(4, 5)

	target := card.New("TARGET_001", 20, types.PlayerTypeOpponent)
	target.MoveTo(card.ZonePlay, 1)
	gameState.Opponent.Board.AddCardAt(target, 1)
	targetEntity := gameState.TrackCardEntity(target)
	targetEntity.SetStats(2, 3)

	attackEvent := game.NewAttackEvent(10, 20)
	gameState.AddEvent(attackEvent)

	damageEvent := game.NewDamageEvent(20, 4)
	gameState.AddEvent(damageEvent)

	deathEvent := game.NewDeathEvent(20)
	gameState.AddEvent(deathEvent)

	discoverEvent := game.NewDiscoverEvent(types.PlayerTypePlayer, []string{"DISCOVER_001", "DISCOVER_002"}, "DISCOVER_002", 20)
	gameState.AddEvent(discoverEvent)

	raw, err := formatter.Format(gameState)
	if err != nil {
		t.Fatalf("format state: %v", err)
	}

	if !strings.Contains(raw, `"attacker"`) || !strings.Contains(raw, `"Учебный атакующий"`) {
		t.Fatalf("expected human-readable attacker in json: %s", raw)
	}
	if !strings.Contains(raw, `"target"`) || !strings.Contains(raw, `"Учебная цель"`) {
		t.Fatalf("expected human-readable target in json: %s", raw)
	}
	if !strings.Contains(raw, `"chosen_card_info"`) || !strings.Contains(raw, `"Вторая находка"`) {
		t.Fatalf("expected human-readable discover result in json: %s", raw)
	}
	if !strings.Contains(raw, `"option_cards"`) || !strings.Contains(raw, `"Первая находка"`) {
		t.Fatalf("expected human-readable discover options in json: %s", raw)
	}
}

func TestJSONFormatterAddsDeckListCards(t *testing.T) {
	t.Parallel()

	formatter := NewJSONFormatter(&stubCardDatabase{
		cards: map[string]*carddb.CardInfo{
			"DECK_001": {
				ID:   "DECK_001",
				Name: "Карта из колоды",
				Text: "Описание карты колоды.",
				Cost: 1,
				Type: carddb.CardTypeSpell,
			},
		},
	})

	cardInDeck := card.New("DECK_001", 0, types.PlayerTypePlayer)
	raw, err := formatter.FormatEvent(game.NewDeckListEvent(types.PlayerTypePlayer, []*card.Card{cardInDeck}, "Учебная колода", "deck-1", "code"))
	if err != nil {
		t.Fatalf("format event: %v", err)
	}

	if !strings.Contains(raw, `"cards"`) {
		t.Fatalf("expected cards list in deck event: %s", raw)
	}
	if !strings.Contains(raw, `"Карта из колоды"`) {
		t.Fatalf("expected human-readable deck card in json: %s", raw)
	}
}
