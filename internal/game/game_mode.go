package game

// GameMode определяет режим игры
type GameMode string

const (
	// GameModeStandard - стандартный режим
	GameModeStandard GameMode = "STANDARD"

	// GameModeWild - дикий режим
	GameModeWild GameMode = "WILD"

	// GameModeArena - арена
	GameModeArena GameMode = "ARENA"

	// GameModeBattlegrounds - поля сражений
	GameModeBattlegrounds GameMode = "BATTLEGROUNDS"

	// GameModeTavernBrawl - потасовка
	GameModeTavernBrawl GameMode = "TAVERN_BRAWL"

	// GameModeDuels - дуэли
	GameModeDuels GameMode = "DUELS"

	// GameModeMercenaries - наемники
	GameModeMercenaries GameMode = "MERCENARIES"

	// GameModeUnknown - неизвестный режим
	GameModeUnknown GameMode = "UNKNOWN"
)

// String возвращает строковое представление режима игры
func (g GameMode) String() string {
	return string(g)
}

// IsValid проверяет валидность режима игры
func (g GameMode) IsValid() bool {
	switch g {
	case GameModeStandard, GameModeWild, GameModeArena,
		GameModeBattlegrounds, GameModeTavernBrawl, GameModeDuels,
		GameModeMercenaries, GameModeUnknown:
		return true
	default:
		return false
	}
}

// IsSupported проверяет, поддерживается ли режим игры (для первой версии только стандарт)
func (g GameMode) IsSupported() bool {
	return g == GameModeStandard
}

// DisplayName возвращает отображаемое имя режима игры
func (g GameMode) DisplayName() string {
	switch g {
	case GameModeStandard:
		return "Стандарт"
	case GameModeWild:
		return "Дикий"
	case GameModeArena:
		return "Арена"
	case GameModeBattlegrounds:
		return "Поля сражений"
	case GameModeTavernBrawl:
		return "Потасовка"
	case GameModeDuels:
		return "Дуэли"
	case GameModeMercenaries:
		return "Наемники"
	default:
		return "Неизвестный"
	}
}
