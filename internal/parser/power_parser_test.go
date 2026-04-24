package parser

import (
	"context"
	"strings"
	"testing"

	"hsanalyzer/internal/game"
	"hsanalyzer/internal/types"
)

func TestPowerParserMetadata(t *testing.T) {
	t.Parallel()

	parser := NewPowerParser()
	if parser.Name() != "power_parser" {
		t.Fatalf("unexpected parser name: %s", parser.Name())
	}
	if parser.LogType() != LogTypePower {
		t.Fatalf("unexpected log type: %s", parser.LogType())
	}
}

func TestPowerParserParseBaseStructure(t *testing.T) {
	t.Parallel()

	gameTypeLine := findLogLine(t, "Power.log", "GameType=GT_RANKED")
	attackLine := findLogLine(t, "Power.log", "BLOCK_START BlockType=ATTACK Entity=[entityName=Хранитель стаи id=29")

	input := strings.Join([]string{
		gameTypeLine,
		attackLine,
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	if _, ok := events[0].(*game.AttackEvent); !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
}

func TestPowerParserRecognizesRealPowerLogFragments(t *testing.T) {
	t.Parallel()

	parser := NewPowerParser()
	gameTypeLine := findLogLine(t, "Power.log", "GameType=GT_RANKED")
	attackLine := findLogLine(t, "Power.log", "BLOCK_START BlockType=ATTACK Entity=[entityName=Хранитель стаи id=29")
	turnLine := findLogLine(t, "Power.log", "TAG_CHANGE Entity=GameEntity tag=TURN value=6")

	if !parser.isRelevantLine(gameTypeLine) {
		t.Fatalf("expected game type line to be relevant")
	}
	if !parser.isRelevantLine(attackLine) {
		t.Fatalf("expected attack line to be relevant")
	}
	if !parser.isRelevantLine(turnLine) {
		t.Fatalf("expected turn line to be relevant")
	}
}

func TestPowerParserParseTurnTransitionsFromRealLogFragments(t *testing.T) {
	t.Parallel()

	turnLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=GameEntity tag=TURN value=6")
	opponentTurnLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=HappY#23785 tag=TURN value=3")
	turnDupLine := findLogLine(t, "Power.log", "PowerTaskList.DebugPrintPower() -     TAG_CHANGE Entity=UNKNOWN HUMAN PLAYER tag=TURN value=4")
	nextTurnGameEntityLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=GameEntity tag=TURN value=7")
	playerTurnLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=UNKNOWN HUMAN PLAYER tag=TURN value=4")

	input := strings.Join([]string{
		turnLine,
		opponentTurnLine,
		turnDupLine,
		nextTurnGameEntityLine,
		playerTurnLine,
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	startOpponent, ok := events[0].(*game.TurnStartEvent)
	if !ok {
		t.Fatalf("unexpected first event type: %T", events[0])
	}
	if startOpponent.Player != types.PlayerTypeOpponent {
		t.Fatalf("unexpected first turn player: %s", startOpponent.Player)
	}

	if _, ok := events[1].(*game.TurnEndEvent); !ok {
		t.Fatalf("unexpected second event type: %T", events[1])
	}

	startPlayer, ok := events[2].(*game.TurnStartEvent)
	if !ok {
		t.Fatalf("unexpected third event type: %T", events[2])
	}
	if startPlayer.Player != types.PlayerTypePlayer {
		t.Fatalf("unexpected second turn player: %s", startPlayer.Player)
	}
}

func TestPowerParserParseCardPlayedFromRealLogFragments(t *testing.T) {
	t.Parallel()

	playStartLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() - BLOCK_START BlockType=PLAY Entity=[entityName=Карта сектантов id=60")
	playResolveLine := findLogLine(t, "Power.log", "PowerTaskList.DebugPrintPower() - BLOCK_START BlockType=POWER Entity=[entityName=Карта сектантов id=60 zone=PLAY")

	input := strings.Join([]string{
		playStartLine,
		playResolveLine,
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	cardPlayed, ok := events[0].(*game.CardPlayedEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if cardPlayed.Player != types.PlayerTypeOpponent {
		t.Fatalf("unexpected player: %s", cardPlayed.Player)
	}
	if cardPlayed.Card == nil {
		t.Fatalf("expected played card")
	}
	if cardPlayed.Card.ID != "TLC_515" {
		t.Fatalf("unexpected card id: %s", cardPlayed.Card.ID)
	}
	if cardPlayed.Card.EntityID != 60 {
		t.Fatalf("unexpected entity id: %d", cardPlayed.Card.EntityID)
	}
	if cardPlayed.Position != 0 {
		t.Fatalf("unexpected position: %d", cardPlayed.Position)
	}
}

func TestPowerParserParseAttackDamageAndDeathFromRealLogFragments(t *testing.T) {
	t.Parallel()

	attackLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() - BLOCK_START BlockType=ATTACK Entity=[entityName=Сумеречное яйцо id=103")
	damageTargetLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=[entityName=Темный наездник id=27 zone=PLAY zonePos=2 cardId=EDR_456 player=1] tag=DAMAGE value=2")
	damageAttackerLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=[entityName=Сумеречное яйцо id=103 zone=PLAY zonePos=2 cardId=CATA_210 player=2] tag=DAMAGE value=2")
	deathTargetLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=[entityName=Темный наездник id=27 zone=PLAY zonePos=2 cardId=EDR_456 player=1] tag=ZONE value=GRAVEYARD")
	deathAttackerLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() -     TAG_CHANGE Entity=[entityName=Сумеречное яйцо id=103 zone=PLAY zonePos=2 cardId=CATA_210 player=2] tag=ZONE value=GRAVEYARD")

	input := strings.Join([]string{
		attackLine,
		damageTargetLine,
		damageAttackerLine,
		deathTargetLine,
		deathAttackerLine,
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}

	if len(events) != 5 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	attack, ok := events[0].(*game.AttackEvent)
	if !ok {
		t.Fatalf("unexpected first event type: %T", events[0])
	}
	if attack.AttackerEntityID != 103 || attack.TargetEntityID != 27 {
		t.Fatalf("unexpected attack event: attacker=%d target=%d", attack.AttackerEntityID, attack.TargetEntityID)
	}

	damageOne, ok := events[1].(*game.DamageEvent)
	if !ok {
		t.Fatalf("unexpected second event type: %T", events[1])
	}
	if damageOne.TargetEntityID != 27 || damageOne.Amount != 2 {
		t.Fatalf("unexpected first damage event: target=%d amount=%d", damageOne.TargetEntityID, damageOne.Amount)
	}

	damageTwo, ok := events[2].(*game.DamageEvent)
	if !ok {
		t.Fatalf("unexpected third event type: %T", events[2])
	}
	if damageTwo.TargetEntityID != 103 || damageTwo.Amount != 2 {
		t.Fatalf("unexpected second damage event: target=%d amount=%d", damageTwo.TargetEntityID, damageTwo.Amount)
	}

	deathOne, ok := events[3].(*game.DeathEvent)
	if !ok {
		t.Fatalf("unexpected fourth event type: %T", events[3])
	}
	if deathOne.EntityID != 27 {
		t.Fatalf("unexpected first death event: %d", deathOne.EntityID)
	}

	deathTwo, ok := events[4].(*game.DeathEvent)
	if !ok {
		t.Fatalf("unexpected fifth event type: %T", events[4])
	}
	if deathTwo.EntityID != 103 {
		t.Fatalf("unexpected second death event: %d", deathTwo.EntityID)
	}
}

func TestPowerParserParseEffectsFromRealLogFragments(t *testing.T) {
	t.Parallel()

	triggerLine := findLogLine(t, "Power.log", "GameState.DebugPrintPower() - BLOCK_START BlockType=TRIGGER Entity=[entityName=Сумеречное яйцо id=95 zone=PLAY zonePos=1 cardId=CATA_210 player=2]")
	statLine := findLogLine(t, "Power.log", "TAG_CHANGE Entity=[entityName=Гаррош Адский Крик id=64 zone=PLAY zonePos=0 cardId=HERO_01 player=1] tag=ATK value=0")

	input := strings.Join([]string{
		triggerLine,
		statLine,
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	triggerEvent, ok := events[0].(*game.EffectTriggeredEvent)
	if !ok {
		t.Fatalf("unexpected first event type: %T", events[0])
	}
	if triggerEvent.SourceEntityID != 95 {
		t.Fatalf("unexpected trigger source: %d", triggerEvent.SourceEntityID)
	}
	if triggerEvent.SourceCardID != "CATA_210" {
		t.Fatalf("unexpected trigger source card: %s", triggerEvent.SourceCardID)
	}
	if triggerEvent.SourceName != "Сумеречное яйцо" {
		t.Fatalf("unexpected trigger source name: %s", triggerEvent.SourceName)
	}
	if triggerEvent.EffectName != "TRIGGER_VISUAL" {
		t.Fatalf("unexpected trigger effect name: %s", triggerEvent.EffectName)
	}

	statEvent, ok := events[1].(*game.EffectTriggeredEvent)
	if !ok {
		t.Fatalf("unexpected second event type: %T", events[1])
	}
	if statEvent.SourceEntityID != 64 {
		t.Fatalf("unexpected stat change source: %d", statEvent.SourceEntityID)
	}
	if statEvent.SourceCardID != "HERO_01" {
		t.Fatalf("unexpected stat source card: %s", statEvent.SourceCardID)
	}
	if statEvent.SourceName != "Гаррош Адский Крик" {
		t.Fatalf("unexpected stat source name: %s", statEvent.SourceName)
	}
	if statEvent.EffectName != "ATK_CHANGE:0" {
		t.Fatalf("unexpected stat effect name: %s", statEvent.EffectName)
	}
}

func TestPowerParserParseDiscoverFromRealLogFragments(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		findLogLine(t, "Power.log", "GameState.DebugPrintEntityChoices() - id=6 Player=HappY#23785"),
		findLogLine(t, "Power.log", "GameState.DebugPrintEntityChoices() -   Entities[0]=[entityName=Безумный приспешник id=117 zone=SETASIDE zonePos=0 cardId=CATA_158 player=2]"),
		findLogLine(t, "Power.log", "GameState.DebugPrintEntityChoices() -   Entities[1]=[entityName=Морхи id=118 zone=SETASIDE zonePos=0 cardId=END_036 player=2]"),
		findLogLine(t, "Power.log", "GameState.DebugPrintEntityChoices() -   Entities[2]=[entityName=Синестра id=119 zone=SETASIDE zonePos=0 cardId=CATA_154 player=2]"),
		findLogLine(t, "Power.log", "GameState.SendChoices() - id=6 ChoiceType=GENERAL"),
		findLogLine(t, "Power.log", "GameState.SendChoices() -   m_chosenEntities[0]=[entityName=Морхи id=118 zone=SETASIDE zonePos=0 cardId=END_036 player=2]"),
	}, "\n")

	parser := NewPowerParser()
	events, err := parser.Parse(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse power log: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("unexpected events count: %d", len(events))
	}

	discover, ok := events[0].(*game.DiscoverEvent)
	if !ok {
		t.Fatalf("unexpected event type: %T", events[0])
	}
	if discover.Player != types.PlayerTypeOpponent {
		t.Fatalf("unexpected player: %s", discover.Player)
	}
	if len(discover.Options) != 3 {
		t.Fatalf("unexpected options count: %d", len(discover.Options))
	}
	if discover.Options[0] != "CATA_158" || discover.Options[1] != "END_036" || discover.Options[2] != "CATA_154" {
		t.Fatalf("unexpected options: %#v", discover.Options)
	}
	if discover.ChosenCard != "END_036" {
		t.Fatalf("unexpected chosen card: %s", discover.ChosenCard)
	}
	if discover.ChosenCardID != 118 {
		t.Fatalf("unexpected chosen entity id: %d", discover.ChosenCardID)
	}
}
