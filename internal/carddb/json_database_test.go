package carddb

import (
	"encoding/json"
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
			"dbfId":315,
			"name":"Огненный шар",
			"text":"Наносит 6 ед. урона.",
			"cost":4,
			"cardClass":"MAGE",
			"type":"SPELL",
			"set":"CORE",
			"rarity":"FREE"
		},
		{
			"id":"CS2_122",
			"dbfId":122,
			"name":"Вождь рейда",
			"cost":3,
			"attack":2,
			"health":2,
			"cardClass":"NEUTRAL",
			"type":"MINION",
			"set":"CORE",
			"rarity":"COMMON",
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
	if cardByID.Name != "Огненный шар" {
		t.Fatalf("unexpected name: %s", cardByID.Name)
	}

	cardByName, ok := db.GetCardByName("огненный шар")
	if !ok {
		t.Fatalf("card must be found by localized name")
	}
	if cardByName.ID != "CS2_029" {
		t.Fatalf("unexpected id: %s", cardByName.ID)
	}

	cardByUpper, ok := db.GetCardByName("  ОГНЕННЫЙ ШАР ")
	if !ok {
		t.Fatalf("card must be found by case-insensitive localized name")
	}
	if cardByUpper.ID != "CS2_029" {
		t.Fatalf("unexpected id for localized lookup: %s", cardByUpper.ID)
	}

	cardByDBFID, ok := db.GetCardByDBFID(315)
	if !ok {
		t.Fatalf("card must be found by dbf id")
	}
	if cardByDBFID.ID != "CS2_029" {
		t.Fatalf("unexpected id by dbf id: %s", cardByDBFID.ID)
	}
}

func TestJSONDatabase_LoadMapFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "cards-map.json")
	payload := `{
		"CS2_188": {"name":"Одержимый сержант","dbfId":188,"cost":1,"cardClass":"NEUTRAL","type":"MINION","set":"CORE"}
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
	if cardByID.Name != "Одержимый сержант" {
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
		Name: "Огненный шар",
	}

	if got := card.GetLocalizedName("ru"); got != "Огненный шар" {
		t.Fatalf("expected ru name, got %s", got)
	}
	if got := card.GetLocalizedName("en"); got != "Огненный шар" {
		t.Fatalf("expected locale-agnostic name, got %s", got)
	}
}

func TestJSONDatabase_Load_RealCardsJSONFormat(t *testing.T) {
	t.Parallel()

	realCardsPath := filepath.Join("..", "..", "data", "cards.json")
	if _, err := os.Stat(realCardsPath); err != nil {
		t.Skipf("real cards db is unavailable: %v", err)
	}

	rawData, err := os.ReadFile(realCardsPath)
	if err != nil {
		t.Fatalf("read real cards db: %v", err)
	}

	var rawCards []map[string]interface{}
	if err := json.Unmarshal(rawData, &rawCards); err != nil {
		t.Fatalf("unmarshal real cards db: %v", err)
	}
	if len(rawCards) == 0 {
		t.Fatalf("real cards db must contain cards")
	}

	var sampleID, sampleName string
	for _, c := range rawCards {
		id, _ := c["id"].(string)
		name, _ := c["name"].(string)
		if id != "" && name != "" {
			sampleID = id
			sampleName = name
			break
		}
	}
	if sampleID == "" || sampleName == "" {
		t.Fatalf("failed to find sample card in real cards db")
	}

	db := NewJSONDatabase()
	if err := db.Load(realCardsPath); err != nil {
		t.Fatalf("load real cards db: %v", err)
	}

	cardByID, ok := db.GetCard(sampleID)
	if !ok {
		t.Fatalf("card must be found by id: %s", sampleID)
	}
	if cardByID.Name == "" {
		t.Fatalf("card name must not be empty for id %s", sampleID)
	}

	cardByName, ok := db.GetCardByName(sampleName)
	if !ok {
		t.Fatalf("card must be found by name: %s", sampleName)
	}
	if cardByName.ID != sampleID {
		t.Fatalf("unexpected card id by name: got %s, want %s", cardByName.ID, sampleID)
	}
}
