package carddb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONDatabase_LoadAndGetters(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "cards.json")
	payload := `[
		{
			"id":"CS2_029",
			"name":"Fireball",
			"nameRu":"Огненный шар",
			"text":"Deal 6 damage.",
			"textRu":"Наносит 6 урона.",
			"cost":4,
			"type":"SPELL",
			"rarity":"FREE",
			"cardClass":"MAGE"
		},
		{
			"id":"CS2_122",
			"name":"Raid Leader",
			"cost":3,
			"attack":2,
			"health":2,
			"type":"MINION",
			"rarity":"COMMON",
			"cardClass":"NEUTRAL",
			"race":"NONE"
		}
	]`

	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatalf("write temp cards file: %v", err)
	}

	db := NewJSONDatabase()
	if err := db.Load(path); err != nil {
		t.Fatalf("load cards: %v", err)
	}

	cardByID, ok := db.GetCard("CS2_029")
	if !ok {
		t.Fatalf("card must be found by id")
	}
	if cardByID.Name != "Fireball" {
		t.Fatalf("unexpected name: %s", cardByID.Name)
	}

	cardByEN, ok := db.GetCardByName("fireball")
	if !ok {
		t.Fatalf("card must be found by en name")
	}
	if cardByEN.ID != "CS2_029" {
		t.Fatalf("unexpected id: %s", cardByEN.ID)
	}

	cardByRU, ok := db.GetCardByName("  ОГНЕННЫЙ ШАР ")
	if !ok {
		t.Fatalf("card must be found by ru name")
	}
	if cardByRU.ID != "CS2_029" {
		t.Fatalf("unexpected id for ru lookup: %s", cardByRU.ID)
	}
}

func TestJSONDatabase_LoadMapFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "cards-map.json")
	payload := `{
		"CS2_188": {"name":"Abusive Sergeant","cost":1,"type":"MINION"}
	}`

	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatalf("write temp cards file: %v", err)
	}

	db := NewJSONDatabase()
	if err := db.Load(path); err != nil {
		t.Fatalf("load cards map: %v", err)
	}

	cardByID, ok := db.GetCard("CS2_188")
	if !ok {
		t.Fatalf("card must be found by id from map format")
	}
	if cardByID.Name != "Abusive Sergeant" {
		t.Fatalf("unexpected card name: %s", cardByID.Name)
	}
}

func TestJSONDatabase_LoadErrors(t *testing.T) {
	t.Parallel()

	db := NewJSONDatabase()
	if err := db.Load("does-not-exist.json"); err == nil {
		t.Fatalf("expected error for missing file")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "broken.json")
	if err := os.WriteFile(path, []byte("{broken"), 0o600); err != nil {
		t.Fatalf("write broken file: %v", err)
	}

	if err := db.Load(path); err == nil {
		t.Fatalf("expected parse error for broken json")
	}
}

func TestCardInfo_GetLocalizedName(t *testing.T) {
	t.Parallel()

	card := &CardInfo{
		Name:   "Fireball",
		NameRu: "Огненный шар",
	}

	if got := card.GetLocalizedName("ru"); got != "Огненный шар" {
		t.Fatalf("expected ru name, got %s", got)
	}
	if got := card.GetLocalizedName("ru-RU"); got != "Огненный шар" {
		t.Fatalf("expected ru-RU name, got %s", got)
	}
	if got := card.GetLocalizedName("en"); got != "Fireball" {
		t.Fatalf("expected fallback english name, got %s", got)
	}

	card.NameRu = ""
	if got := card.GetLocalizedName("ru"); got != "Fireball" {
		t.Fatalf("expected fallback name without NameRu, got %s", got)
	}
}
