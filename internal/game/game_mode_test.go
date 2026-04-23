package game

import "testing"

func TestGameMode_String(t *testing.T) {
	tests := []struct {
		name string
		g    GameMode
		want string
	}{
		{"Standard", GameModeStandard, "STANDARD"},
		{"Wild", GameModeWild, "WILD"},
		{"Arena", GameModeArena, "ARENA"},
		{"Battlegrounds", GameModeBattlegrounds, "BATTLEGROUNDS"},
		{"TavernBrawl", GameModeTavernBrawl, "TAVERN_BRAWL"},
		{"Duels", GameModeDuels, "DUELS"},
		{"Mercenaries", GameModeMercenaries, "MERCENARIES"},
		{"Unknown", GameModeUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.String(); got != tt.want {
				t.Errorf("GameMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGameMode_IsValid(t *testing.T) {
	tests := []struct {
		name string
		g    GameMode
		want bool
	}{
		{"Valid Standard", GameModeStandard, true},
		{"Valid Wild", GameModeWild, true},
		{"Valid Arena", GameModeArena, true},
		{"Valid Battlegrounds", GameModeBattlegrounds, true},
		{"Valid TavernBrawl", GameModeTavernBrawl, true},
		{"Valid Duels", GameModeDuels, true},
		{"Valid Mercenaries", GameModeMercenaries, true},
		{"Valid Unknown", GameModeUnknown, true},
		{"Invalid", GameMode("INVALID"), false},
		{"Empty", GameMode(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.IsValid(); got != tt.want {
				t.Errorf("GameMode.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGameMode_IsSupported(t *testing.T) {
	tests := []struct {
		name string
		g    GameMode
		want bool
	}{
		{"Supported Standard", GameModeStandard, true},
		{"Unsupported Wild", GameModeWild, false},
		{"Unsupported Arena", GameModeArena, false},
		{"Unsupported Battlegrounds", GameModeBattlegrounds, false},
		{"Unsupported TavernBrawl", GameModeTavernBrawl, false},
		{"Unsupported Duels", GameModeDuels, false},
		{"Unsupported Mercenaries", GameModeMercenaries, false},
		{"Unsupported Unknown", GameModeUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.IsSupported(); got != tt.want {
				t.Errorf("GameMode.IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGameMode_DisplayName(t *testing.T) {
	tests := []struct {
		name string
		g    GameMode
		want string
	}{
		{"Standard", GameModeStandard, "Стандарт"},
		{"Wild", GameModeWild, "Дикий"},
		{"Arena", GameModeArena, "Арена"},
		{"Battlegrounds", GameModeBattlegrounds, "Поля сражений"},
		{"TavernBrawl", GameModeTavernBrawl, "Потасовка"},
		{"Duels", GameModeDuels, "Дуэли"},
		{"Mercenaries", GameModeMercenaries, "Наемники"},
		{"Unknown", GameModeUnknown, "Неизвестный"},
		{"Invalid", GameMode("INVALID"), "Неизвестный"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.DisplayName(); got != tt.want {
				t.Errorf("GameMode.DisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}
