package formatter

import (
	"strings"
	"testing"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

func TestTextFormatterFormatEventTurnCardCombatAndEffects(t *testing.T) {
	t.Parallel()

	formatter := NewTextFormatter(TextFormatterConfig{Language: "ru"})

	cardInfo := &carddb.CardInfo{
		ID:     "CARD_001",
		Name:   "Учебный дракон",
		Cost:   4,
		Attack: 3,
		Health: 5,
		Type:   carddb.CardTypeMinion,
	}
	c := card.New("CARD_001", 42, types.PlayerTypePlayer)
	c.Info = cardInfo

	cases := []struct {
		name   string
		event  game.Event
		expect string
	}{
		{name: "turn start", event: game.NewTurnStartEvent(types.PlayerTypePlayer), expect: "Начало хода"},
		{name: "card played", event: game.NewCardPlayedEvent(types.PlayerTypePlayer, c, 0), expect: "Учебный дракон"},
		{name: "card stats", event: game.NewCardPlayedEvent(types.PlayerTypePlayer, c, 0), expect: "⚔️3"},
		{name: "attack", event: game.NewAttackEvent(10, 20), expect: "#10"},
		{name: "damage", event: game.NewDamageEvent(20, 5), expect: "5"},
		{name: "death", event: game.NewDeathEvent(20), expect: "Смерть"},
		{name: "effect", event: game.NewEffectTriggeredEvent(9, "BATTLECRY"), expect: "BATTLECRY"},
		{name: "discover", event: game.NewDiscoverEvent(types.PlayerTypePlayer, []string{"A", "B", "C"}, "B", 77), expect: "A, B, C"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := formatter.FormatEvent(tc.event)
			if err != nil {
				t.Fatalf("format event: %v", err)
			}
			if !strings.Contains(got, tc.expect) {
				t.Fatalf("unexpected output %q, expected substring %q", got, tc.expect)
			}
		})
	}
}

func TestTextFormatterFormatStateWithTurnsHandAndSummary(t *testing.T) {
	t.Parallel()

	formatter := NewTextFormatter(TextFormatterConfig{Language: "ru", Verbose: true})
	gameState := state.New("game-42", state.GameModeStandard)
	gameState.Player.Name = "Игрок"
	gameState.Opponent.Name = "Оппонент"

	start := game.NewGameStartEvent("game-42", state.GameModeStandard)
	if err := start.Apply(gameState); err != nil {
		t.Fatalf("apply game start: %v", err)
	}
	gameState.AddEvent(start)

	turnStart := game.NewTurnStartEvent(types.PlayerTypePlayer)
	if err := turnStart.Apply(gameState); err != nil {
		t.Fatalf("apply turn start: %v", err)
	}
	gameState.AddEvent(turnStart)

	drawn := testCard("DRAW_1", 1, "Доборная карта", 2, 0, 0, carddb.CardTypeSpell)
	drawEvent := game.NewCardDrawnEvent(types.PlayerTypePlayer, drawn)
	if err := drawEvent.Apply(gameState); err != nil {
		t.Fatalf("apply draw: %v", err)
	}
	gameState.AddEvent(drawEvent)

	played := testCard("PLAY_1", 2, "Боевой миньон", 3, 2, 4, carddb.CardTypeMinion)
	gameState.Player.Hand.AddCard(played)
	playEvent := game.NewCardPlayedEvent(types.PlayerTypePlayer, played, 0)
	if err := playEvent.Apply(gameState); err != nil {
		t.Fatalf("apply play: %v", err)
	}
	gameState.AddEvent(playEvent)

	discover := game.NewDiscoverEvent(types.PlayerTypePlayer, []string{"OPT_A", "OPT_B", "OPT_C"}, "OPT_B", 77)
	if err := discover.Apply(gameState); err != nil {
		t.Fatalf("apply discover: %v", err)
	}
	gameState.AddEvent(discover)

	turnEnd := game.NewTurnEndEvent()
	if err := turnEnd.Apply(gameState); err != nil {
		t.Fatalf("apply turn end: %v", err)
	}
	gameState.AddEvent(turnEnd)

	gameEnd := game.NewGameEndEvent(types.PlayerTypePlayer)
	if err := gameEnd.Apply(gameState); err != nil {
		t.Fatalf("apply game end: %v", err)
	}
	gameState.AddEvent(gameEnd)

	got, err := formatter.Format(gameState)
	if err != nil {
		t.Fatalf("format state: %v", err)
	}

	expectSubstrings := []string{
		"Игра",
		"ID игры: game-42",
		"Ходы",
		"Ход 1 (Игрок)",
		"Начало хода 1 (Игрок)",
		"Добор: Доборная карта",
		"Розыгрыш: Боевой миньон",
		"Рука: + Доборная карта; - Боевой миньон",
		"Состояние руки",
		"Игрок: 1",
		"Итог",
		"Победитель: Игрок",
		"Карт в руке игрока: 1",
	}

	for _, want := range expectSubstrings {
		if !strings.Contains(got, want) {
			t.Fatalf("formatted output does not contain %q:\n%s", want, got)
		}
	}
}

func TestTextFormatterFormatSupportsEnglish(t *testing.T) {
	t.Parallel()

	formatter := NewTextFormatter(TextFormatterConfig{Language: "en"})
	got, err := formatter.FormatEvent(game.NewGameEndEvent(types.PlayerTypeOpponent))
	if err != nil {
		t.Fatalf("format event: %v", err)
	}
	if !strings.Contains(got, "Winner: Opponent") {
		t.Fatalf("unexpected english output: %s", got)
	}
}

func TestTextFormatterRejectsNilInputs(t *testing.T) {
	t.Parallel()

	formatter := NewTextFormatter(TextFormatterConfig{})
	if _, err := formatter.Format(nil); err == nil {
		t.Fatalf("expected nil state error")
	}
	if _, err := formatter.FormatEvent(nil); err == nil {
		t.Fatalf("expected nil event error")
	}
}

func testCard(id string, entityID int, name string, cost, attack, health int, cardType carddb.CardType) *card.Card {
	c := card.New(id, entityID, types.PlayerTypePlayer)
	c.Info = &carddb.CardInfo{
		ID:     id,
		Name:   name,
		Cost:   cost,
		Attack: attack,
		Health: health,
		Type:   cardType,
	}
	c.Cost = cost
	return c
}
