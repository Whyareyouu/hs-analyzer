package formatter

import (
	"fmt"
	"strings"

	"hsanalyzer/internal/card"
	"hsanalyzer/internal/carddb"
	"hsanalyzer/internal/game"
	"hsanalyzer/internal/state"
	"hsanalyzer/internal/types"
)

// TextFormatterConfig задает настройки текстового форматтера.
type TextFormatterConfig struct {
	Language string
	Verbose  bool
}

// TextFormatter форматирует состояние игры и отдельные события в текст.
type TextFormatter struct {
	config TextFormatterConfig
}

// NewTextFormatter создает текстовый форматтер.
func NewTextFormatter(cfg TextFormatterConfig) *TextFormatter {
	if strings.TrimSpace(cfg.Language) == "" {
		cfg.Language = "ru"
	}

	return &TextFormatter{config: cfg}
}

func (f *TextFormatter) Name() string {
	return "text"
}

func (f *TextFormatter) Format(gameState *state.State) (string, error) {
	if gameState == nil {
		return "", fmt.Errorf("game state is nil")
	}

	var sections []string
	sections = append(sections, f.formatHeader(gameState))

	if len(gameState.Turns) > 0 {
		sections = append(sections, f.formatTurns(gameState))
	}

	sections = append(sections, f.formatHands(gameState))
	sections = append(sections, f.formatSummary(gameState))

	return strings.TrimSpace(strings.Join(sections, "\n\n")), nil
}

func (f *TextFormatter) FormatEvent(event game.Event) (string, error) {
	if event == nil {
		return "", fmt.Errorf("event is nil")
	}

	switch e := event.(type) {
	case *game.GameStartEvent:
		return f.gameStartLine(e), nil
	case *game.GameEndEvent:
		return f.gameEndLine(e), nil
	case *game.TurnStartEvent:
		return f.turnStartLine(0, e), nil
	case *game.TurnEndEvent:
		return f.turnEndLine(), nil
	case *game.DeckListEvent:
		return f.deckListLine(e), nil
	case *game.CardPlayedEvent:
		return f.cardPlayedLine(e), nil
	case *game.CardDrawnEvent:
		return f.cardDrawnLine(e), nil
	case *game.AttackEvent:
		return f.attackLine(e), nil
	case *game.DamageEvent:
		return f.damageLine(e), nil
	case *game.HealEvent:
		return f.healLine(e), nil
	case *game.DeathEvent:
		return f.deathLine(e), nil
	case *game.EffectTriggeredEvent:
		return f.effectLine(e), nil
	case *game.DiscoverEvent:
		return f.discoverLine(e), nil
	default:
		return f.unknownEventLine(event), nil
	}
}

func (f *TextFormatter) formatHeader(gameState *state.State) string {
	lines := []string{
		f.sectionTitle("Игра", "Game"),
		fmt.Sprintf("%s %s", f.label("ID игры:", "Game ID:"), emptyFallback(gameState.GameID, "n/a")),
		fmt.Sprintf("%s %s", f.label("Режим:", "Mode:"), gameState.GameMode),
	}

	if gameState.Player != nil && gameState.Player.Name != "" {
		lines = append(lines, fmt.Sprintf("%s %s", f.label("Игрок:", "Player:"), gameState.Player.Name))
	}
	if gameState.Opponent != nil && gameState.Opponent.Name != "" {
		lines = append(lines, fmt.Sprintf("%s %s", f.label("Оппонент:", "Opponent:"), gameState.Opponent.Name))
	}

	return strings.Join(lines, "\n")
}

func (f *TextFormatter) formatTurns(gameState *state.State) string {
	lines := []string{f.sectionTitle("Ходы", "Turns")}

	for _, turn := range gameState.Turns {
		lines = append(lines, fmt.Sprintf("%s %s", f.indent(0), f.turnHeader(turn.Number, turn.Player)))

		for _, rawEvent := range turn.Events {
			event, ok := rawEvent.(game.Event)
			if !ok {
				continue
			}

			line, err := f.FormatEvent(event)
			if err != nil {
				return strings.Join(append(lines, f.indent(1)+"[format error]"), "\n")
			}

			if start, ok := event.(*game.TurnStartEvent); ok {
				line = f.turnStartLine(turn.Number, start)
			}

			lines = append(lines, f.indent(1)+line)
		}

		handChanges := f.turnHandChanges(turn.Events)
		if len(handChanges) > 0 {
			lines = append(lines, f.indent(1)+f.label("Рука:", "Hand:")+" "+strings.Join(handChanges, "; "))
		}
	}

	return strings.Join(lines, "\n")
}

func (f *TextFormatter) formatHands(gameState *state.State) string {
	lines := []string{
		f.sectionTitle("Состояние руки", "Hand State"),
		f.playerHandLine(f.label("Игрок", "Player"), gameState.Player),
		f.playerHandLine(f.label("Оппонент", "Opponent"), gameState.Opponent),
	}

	return strings.Join(lines, "\n")
}

func (f *TextFormatter) formatSummary(gameState *state.State) string {
	lines := []string{
		f.sectionTitle("Итог", "Summary"),
		fmt.Sprintf("%s %d", f.label("Ходов:", "Turns:"), gameState.GetTurnCount()),
		fmt.Sprintf("%s %d", f.label("Событий:", "Events:"), gameState.GetEventCount()),
	}

	if gameState.Winner != nil {
		lines = append(lines, fmt.Sprintf("%s %s", f.label("Победитель:", "Winner:"), f.playerLabel(*gameState.Winner)))
	}

	if f.config.Verbose {
		lines = append(lines,
			fmt.Sprintf("%s %d", f.label("Карт в руке игрока:", "Player hand size:"), handCount(gameState.Player)),
			fmt.Sprintf("%s %d", f.label("Карт в руке оппонента:", "Opponent hand size:"), handCount(gameState.Opponent)),
			fmt.Sprintf("%s %d", f.label("Карт в колоде игрока:", "Player deck size:"), deckCount(gameState.Player)),
			fmt.Sprintf("%s %d", f.label("Карт в колоде оппонента:", "Opponent deck size:"), deckCount(gameState.Opponent)),
		)
	}

	return strings.Join(lines, "\n")
}

func (f *TextFormatter) gameStartLine(event *game.GameStartEvent) string {
	return fmt.Sprintf("▶️ %s %s", f.label("Начало игры:", "Game started:"), event.GameMode)
}

func (f *TextFormatter) gameEndLine(event *game.GameEndEvent) string {
	return fmt.Sprintf("🏆 %s %s", f.label("Победитель:", "Winner:"), f.playerLabel(event.Winner))
}

func (f *TextFormatter) turnStartLine(turnNumber int, event *game.TurnStartEvent) string {
	if turnNumber > 0 {
		return fmt.Sprintf("🔄 %s %d (%s)", f.label("Начало хода", "Turn start"), turnNumber, f.playerLabel(event.Player))
	}
	return fmt.Sprintf("🔄 %s %s", f.label("Начало хода:", "Turn start:"), f.playerLabel(event.Player))
}

func (f *TextFormatter) turnEndLine() string {
	return fmt.Sprintf("⏹️ %s", f.label("Конец хода", "Turn end"))
}

func (f *TextFormatter) deckListLine(event *game.DeckListEvent) string {
	return fmt.Sprintf("📚 %s %s (%d)", f.label("Колода:", "Deck:"), emptyFallback(event.Name, event.DeckID), len(event.Cards))
}

func (f *TextFormatter) cardPlayedLine(event *game.CardPlayedEvent) string {
	return fmt.Sprintf("🃏 %s %s", f.label("Розыгрыш:", "Played:"), formatCard(event.Card))
}

func (f *TextFormatter) cardDrawnLine(event *game.CardDrawnEvent) string {
	return fmt.Sprintf("✋ %s %s", f.label("Добор:", "Draw:"), formatCard(event.Card))
}

func (f *TextFormatter) attackLine(event *game.AttackEvent) string {
	return fmt.Sprintf("⚔️ %s #%d → #%d", f.label("Атака:", "Attack:"), event.AttackerEntityID, event.TargetEntityID)
}

func (f *TextFormatter) damageLine(event *game.DamageEvent) string {
	return fmt.Sprintf("💥 %s #%d (%d)", f.label("Урон:", "Damage:"), event.TargetEntityID, event.Amount)
}

func (f *TextFormatter) healLine(event *game.HealEvent) string {
	return fmt.Sprintf("💚 %s #%d (%d)", f.label("Лечение:", "Heal:"), event.TargetEntityID, event.Amount)
}

func (f *TextFormatter) deathLine(event *game.DeathEvent) string {
	return fmt.Sprintf("☠️ %s #%d", f.label("Смерть:", "Death:"), event.EntityID)
}

func (f *TextFormatter) effectLine(event *game.EffectTriggeredEvent) string {
	return fmt.Sprintf("✨ %s #%d (%s)", f.label("Эффект:", "Effect:"), event.SourceEntityID, event.EffectName)
}

func (f *TextFormatter) discoverLine(event *game.DiscoverEvent) string {
	return fmt.Sprintf("🔍 %s %s → %s", f.label("Discover:", "Discover:"), strings.Join(event.Options, ", "), event.ChosenCard)
}

func (f *TextFormatter) unknownEventLine(event game.Event) string {
	return fmt.Sprintf("%s %s", f.label("Событие:", "Event:"), event.Type())
}

func (f *TextFormatter) turnHeader(number int, playerType types.PlayerType) string {
	return fmt.Sprintf("%s %d (%s)", f.label("Ход", "Turn"), number, f.playerLabel(playerType))
}

func (f *TextFormatter) turnHandChanges(rawEvents []interface{}) []string {
	changes := make([]string, 0)

	for _, rawEvent := range rawEvents {
		event, ok := rawEvent.(game.Event)
		if !ok {
			continue
		}

		switch e := event.(type) {
		case *game.CardDrawnEvent:
			changes = append(changes, fmt.Sprintf("+ %s", formatCardShort(e.Card)))
		case *game.CardPlayedEvent:
			changes = append(changes, fmt.Sprintf("- %s", formatCardShort(e.Card)))
		}
	}

	return changes
}

func (f *TextFormatter) playerHandLine(title string, p interface{ GetHandSize() int }) string {
	return fmt.Sprintf("%s: %d", title, p.GetHandSize())
}

func (f *TextFormatter) sectionTitle(ru, en string) string {
	if f.config.Language == "en" {
		return en
	}
	return ru
}

func (f *TextFormatter) label(ru, en string) string {
	if f.config.Language == "en" {
		return en
	}
	return ru
}

func (f *TextFormatter) playerLabel(playerType types.PlayerType) string {
	if f.config.Language == "en" {
		switch playerType {
		case types.PlayerTypePlayer:
			return "Player"
		case types.PlayerTypeOpponent:
			return "Opponent"
		default:
			return "Unknown"
		}
	}

	switch playerType {
	case types.PlayerTypePlayer:
		return "Игрок"
	case types.PlayerTypeOpponent:
		return "Оппонент"
	default:
		return "Неизвестно"
	}
}

func (f *TextFormatter) indent(level int) string {
	return strings.Repeat("  ", level)
}

func formatCard(c *card.Card) string {
	if c == nil {
		return "UNKNOWN_CARD"
	}

	parts := []string{cardDisplayName(c)}
	if c.ID != "" {
		parts = append(parts, fmt.Sprintf("[%s]", c.ID))
	}

	stats := make([]string, 0, 3)
	cost := cardCost(c)
	if cost >= 0 {
		stats = append(stats, fmt.Sprintf("💎%d", cost))
	}
	attack, ok := cardAttack(c)
	if ok {
		stats = append(stats, fmt.Sprintf("⚔️%d", attack))
	}
	health, ok := cardHealth(c)
	if ok {
		stats = append(stats, fmt.Sprintf("❤️%d", health))
	}
	if len(stats) > 0 {
		parts = append(parts, strings.Join(stats, " "))
	}

	return strings.Join(parts, " ")
}

func formatCardShort(c *card.Card) string {
	if c == nil {
		return "UNKNOWN_CARD"
	}
	if name := cardDisplayName(c); name != "" {
		return name
	}
	return emptyFallback(c.ID, "UNKNOWN_CARD")
}

func cardDisplayName(c *card.Card) string {
	if c == nil {
		return ""
	}
	if info, ok := c.Info.(*carddb.CardInfo); ok && info != nil && info.Name != "" {
		return info.GetLocalizedName("ru")
	}
	return emptyFallback(c.ID, "UNKNOWN_CARD")
}

func cardCost(c *card.Card) int {
	if c == nil {
		return -1
	}
	if c.Cost > 0 {
		return c.Cost
	}
	if info, ok := c.Info.(*carddb.CardInfo); ok && info != nil {
		return info.Cost
	}
	return 0
}

func cardAttack(c *card.Card) (int, bool) {
	if c == nil {
		return 0, false
	}
	if info, ok := c.Info.(*carddb.CardInfo); ok && info != nil {
		switch info.Type {
		case carddb.CardTypeMinion, carddb.CardTypeWeapon, carddb.CardTypeHero:
			return info.Attack, true
		}
	}
	return 0, false
}

func cardHealth(c *card.Card) (int, bool) {
	if c == nil {
		return 0, false
	}
	if info, ok := c.Info.(*carddb.CardInfo); ok && info != nil {
		switch info.Type {
		case carddb.CardTypeMinion, carddb.CardTypeHero:
			return info.Health, true
		}
	}
	return 0, false
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func handCount(p interface{ GetHandSize() int }) int {
	if p == nil {
		return 0
	}
	return p.GetHandSize()
}

func deckCount(p interface{ GetDeckSize() int }) int {
	if p == nil {
		return 0
	}
	return p.GetDeckSize()
}
