package game

import (
	"testing"
	"time"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/entity"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

func TestBaseEvent(t *testing.T) {
	before := time.Now()
	e := NewBaseEvent(types.EventTypeTurnStart)
	after := time.Now()

	if e.Type() != types.EventTypeTurnStart {
		t.Fatalf("unexpected type: %s", e.Type())
	}
	if e.Timestamp().Before(before) || e.Timestamp().After(after) {
		t.Fatalf("timestamp must be set to now")
	}
}

func TestGameEventsApply(t *testing.T) {
	s := state.New("old-id", state.GameModeUnknown)

	start := NewGameStartEvent("new-id", state.GameModeStandard)
	if err := start.Apply(s); err != nil {
		t.Fatalf("game start apply failed: %v", err)
	}
	if s.GameID != "new-id" {
		t.Fatalf("unexpected game id: %s", s.GameID)
	}
	if s.GameMode != state.GameModeStandard {
		t.Fatalf("unexpected game mode: %s", s.GameMode)
	}

	end := NewGameEndEvent(types.PlayerTypePlayer)
	if err := end.Apply(s); err != nil {
		t.Fatalf("game end apply failed: %v", err)
	}
	if s.Winner == nil || *s.Winner != types.PlayerTypePlayer {
		t.Fatalf("winner must be player")
	}
}

func TestTurnEventsApply(t *testing.T) {
	s := state.New("game-id", state.GameModeStandard)

	start := NewTurnStartEvent(types.PlayerTypePlayer)
	if err := start.Apply(s); err != nil {
		t.Fatalf("turn start apply failed: %v", err)
	}
	if s.CurrentTurn != 1 {
		t.Fatalf("expected turn 1, got %d", s.CurrentTurn)
	}

	end := NewTurnEndEvent()
	if err := end.Apply(s); err != nil {
		t.Fatalf("turn end apply failed: %v", err)
	}
	if s.GetCurrentTurn() == nil || !s.GetCurrentTurn().IsEnded() {
		t.Fatalf("current turn must be ended")
	}
}

func TestCardEventsApply(t *testing.T) {
	s := state.New("game-id", state.GameModeStandard)
	player := s.GetPlayer(types.PlayerTypePlayer)

	played := card.New("CARD_PLAY", 100, types.PlayerTypePlayer)
	played.Zone = card.ZonePlay
	player.Hand.AddCard(played)

	playEvent := NewCardPlayedEvent(types.PlayerTypePlayer, played, 0)
	if err := playEvent.Apply(s); err != nil {
		t.Fatalf("card played apply failed: %v", err)
	}
	if player.Hand.Count() != 0 {
		t.Fatalf("expected empty hand after play")
	}
	if player.Board.Count() != 1 {
		t.Fatalf("expected one minion on board")
	}

	drawn := card.New("CARD_DRAW", 101, types.PlayerTypePlayer)
	drawEvent := NewCardDrawnEvent(types.PlayerTypePlayer, drawn)
	if err := drawEvent.Apply(s); err != nil {
		t.Fatalf("card drawn apply failed: %v", err)
	}
	if player.Hand.Count() != 1 {
		t.Fatalf("expected one card in hand after draw")
	}
}

func TestCombatEventsApply(t *testing.T) {
	s := state.New("game-id", state.GameModeStandard)
	s.Player.Hero = entity.New(1, "HERO_01", types.PlayerTypePlayer)
	s.Player.Hero.SetStats(0, 30)

	attack := NewAttackEvent(1, 1)
	if err := attack.Apply(s); err != nil {
		t.Fatalf("attack apply failed: %v", err)
	}

	damage := NewDamageEvent(1, 5)
	if err := damage.Apply(s); err != nil {
		t.Fatalf("damage apply failed: %v", err)
	}
	if s.Player.Hero.Health != 25 {
		t.Fatalf("expected hero health 25, got %d", s.Player.Hero.Health)
	}

	heal := NewHealEvent(1, 3)
	if err := heal.Apply(s); err != nil {
		t.Fatalf("heal apply failed: %v", err)
	}
	if s.Player.Hero.Health != 28 {
		t.Fatalf("expected hero health 28, got %d", s.Player.Hero.Health)
	}

	death := NewDeathEvent(1)
	if err := death.Apply(s); err != nil {
		t.Fatalf("death apply failed: %v", err)
	}
	if s.Player.Hero.Health != 0 {
		t.Fatalf("expected hero health 0, got %d", s.Player.Hero.Health)
	}
}

func TestEffectEventsApply(t *testing.T) {
	s := state.New("game-id", state.GameModeStandard)

	effect := NewEffectTriggeredEvent(1, "battlecry")
	if err := effect.Apply(s); err != nil {
		t.Fatalf("effect apply failed: %v", err)
	}

	discover := NewDiscoverEvent(
		types.PlayerTypePlayer,
		[]string{"CARD_A", "CARD_B", "CARD_C"},
		"CARD_B",
		200,
	)
	if err := discover.Apply(s); err != nil {
		t.Fatalf("discover apply failed: %v", err)
	}
}
