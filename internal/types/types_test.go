package types

import "testing"

func TestPlayerType_String(t *testing.T) {
	tests := []struct {
		name string
		p    PlayerType
		want string
	}{
		{"Player", PlayerTypePlayer, "PLAYER"},
		{"Opponent", PlayerTypeOpponent, "OPPONENT"},
		{"Unknown", PlayerTypeUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("PlayerType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlayerType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		p    PlayerType
		want bool
	}{
		{"Valid Player", PlayerTypePlayer, true},
		{"Valid Opponent", PlayerTypeOpponent, true},
		{"Valid Unknown", PlayerTypeUnknown, true},
		{"Invalid", PlayerType("INVALID"), false},
		{"Empty", PlayerType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsValid(); got != tt.want {
				t.Errorf("PlayerType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZoneType_String(t *testing.T) {
	tests := []struct {
		name string
		z    ZoneType
		want string
	}{
		{"Hand", ZoneHand, "HAND"},
		{"Deck", ZoneDeck, "DECK"},
		{"Play", ZonePlay, "PLAY"},
		{"Graveyard", ZoneGraveyard, "GRAVEYARD"},
		{"Secret", ZoneSecret, "SECRET"},
		{"SetAside", ZoneSetAside, "SETASIDE"},
		{"Unknown", ZoneUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.z.String(); got != tt.want {
				t.Errorf("ZoneType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZoneType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		z    ZoneType
		want bool
	}{
		{"Valid Hand", ZoneHand, true},
		{"Valid Deck", ZoneDeck, true},
		{"Valid Play", ZonePlay, true},
		{"Valid Graveyard", ZoneGraveyard, true},
		{"Valid Secret", ZoneSecret, true},
		{"Valid SetAside", ZoneSetAside, true},
		{"Valid Unknown", ZoneUnknown, true},
		{"Invalid", ZoneType("INVALID"), false},
		{"Empty", ZoneType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.z.IsValid(); got != tt.want {
				t.Errorf("ZoneType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardType_String(t *testing.T) {
	tests := []struct {
		name string
		c    CardType
		want string
	}{
		{"Minion", CardTypeMinion, "MINION"},
		{"Spell", CardTypeSpell, "SPELL"},
		{"Weapon", CardTypeWeapon, "WEAPON"},
		{"Hero", CardTypeHero, "HERO"},
		{"HeroPower", CardTypeHeroPower, "HERO_POWER"},
		{"Unknown", CardTypeUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("CardType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		c    CardType
		want bool
	}{
		{"Valid Minion", CardTypeMinion, true},
		{"Valid Spell", CardTypeSpell, true},
		{"Valid Weapon", CardTypeWeapon, true},
		{"Valid Hero", CardTypeHero, true},
		{"Valid HeroPower", CardTypeHeroPower, true},
		{"Valid Unknown", CardTypeUnknown, true},
		{"Invalid", CardType("INVALID"), false},
		{"Empty", CardType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.IsValid(); got != tt.want {
				t.Errorf("CardType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventType_String(t *testing.T) {
	tests := []struct {
		name string
		e    EventType
		want string
	}{
		{"GameStart", EventTypeGameStart, "game_start"},
		{"TurnStart", EventTypeTurnStart, "turn_start"},
		{"TurnEnd", EventTypeTurnEnd, "turn_end"},
		{"DeckList", EventTypeDeckList, "deck_list"},
		{"CardPlayed", EventTypeCardPlayed, "card_played"},
		{"CardDrawn", EventTypeCardDrawn, "card_drawn"},
		{"Attack", EventTypeAttack, "attack"},
		{"Damage", EventTypeDamage, "damage"},
		{"Heal", EventTypeHeal, "heal"},
		{"Death", EventTypeDeath, "death"},
		{"CardCostChanged", EventTypeCardCostChanged, "card_cost_changed"},
		{"EffectTriggered", EventTypeEffectTriggered, "effect_triggered"},
		{"Discover", EventTypeDiscover, "discover"},
		{"GameEnd", EventTypeGameEnd, "game_end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("EventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		e    EventType
		want bool
	}{
		{"Valid GameStart", EventTypeGameStart, true},
		{"Valid TurnStart", EventTypeTurnStart, true},
		{"Valid DeckList", EventTypeDeckList, true},
		{"Valid CardPlayed", EventTypeCardPlayed, true},
		{"Valid Attack", EventTypeAttack, true},
		{"Valid CardCostChanged", EventTypeCardCostChanged, true},
		{"Valid GameEnd", EventTypeGameEnd, true},
		{"Invalid", EventType("invalid_event"), false},
		{"Empty", EventType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.IsValid(); got != tt.want {
				t.Errorf("EventType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
