package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRootCommandContainsExpectedSubcommands(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	if cmd == nil {
		t.Fatalf("expected root command")
	}

	subcommands := cmd.Commands()
	if len(subcommands) < 2 {
		t.Fatalf("expected at least two subcommands")
	}

	names := []string{subcommands[0].Name(), subcommands[1].Name()}
	joined := strings.Join(names, ",")
	if !strings.Contains(joined, "analyze") || !strings.Contains(joined, "update-cards") {
		t.Fatalf("unexpected subcommands: %v", names)
	}
}

func TestAnalyzeCommandPrintsFormattedResultFromRealLogs(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	copyFileForTest(t, filepath.Join("..", "..", "logs", "Decks.log"), filepath.Join(logsDir, "Decks.log"))
	copyFileForTest(t, filepath.Join("..", "..", "logs", "LoadingScreen.log"), filepath.Join(logsDir, "LoadingScreen.log"))

	var stdout bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", logsDir,
		"--cards-db-path", filepath.Join("..", "..", "data", "cards.json"),
		"--format", "text",
		"--lang", "ru",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{"Игра", "Режим: STANDARD", "Игрок: ZA | LE | PA", "Итог"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output does not contain %q:\n%s", want, output)
		}
	}

	entries, err := os.ReadDir(projectOutputDirForTest())
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if !containsTimelineJSON(entries) {
		t.Fatalf("expected auto-saved timeline json in output")
	}
}

func TestAnalyzeCommandWritesOutputToFile(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	copyFileForTest(t, filepath.Join("..", "..", "logs", "Decks.log"), filepath.Join(logsDir, "Decks.log"))
	copyFileForTest(t, filepath.Join("..", "..", "logs", "LoadingScreen.log"), filepath.Join(logsDir, "LoadingScreen.log"))

	outputPath := filepath.Join(t.TempDir(), "result.txt")
	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", logsDir,
		"--cards-db-path", filepath.Join("..", "..", "data", "cards.json"),
		"--output", "file",
		"--output-file", outputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(data), "Игра") {
		t.Fatalf("unexpected file output:\n%s", string(data))
	}
}

func TestBuildContainerRejectsUnsupportedFormat(t *testing.T) {
	t.Parallel()

	_, err := buildContainer(analyzeOptions{
		logsPath:    filepath.Join("..", "..", "logs"),
		cardsDBPath: filepath.Join("..", "..", "data", "cards.json"),
		format:      "xml",
	})
	if err == nil {
		t.Fatalf("expected unsupported format error")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeCommandWritesJSONTurnsArrayToFile(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	copyFileForTest(t, filepath.Join("..", "..", "logs", "Decks.log"), filepath.Join(logsDir, "Decks.log"))
	copyFileForTest(t, filepath.Join("..", "..", "logs", "LoadingScreen.log"), filepath.Join(logsDir, "LoadingScreen.log"))
	copyFileForTest(t, filepath.Join("..", "..", "logs", "Power.log"), filepath.Join(logsDir, "Power.log"))

	outputPath := filepath.Join(t.TempDir(), "timeline.json")
	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", logsDir,
		"--cards-db-path", filepath.Join("..", "..", "data", "cards.json"),
		"--format", "json",
		"--output", "file",
		"--output-file", outputPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(data), `"number"`) || !strings.Contains(string(data), `"events"`) || !strings.Contains(string(data), `"player"`) {
		t.Fatalf("unexpected json file output:\n%s", string(data))
	}

	card, ok, err := findFirstPlayedCard(data)
	if err != nil {
		t.Fatalf("find first played card: %v", err)
	}
	if !ok {
		t.Fatalf("expected at least one card_played event in json output")
	}
	if card.ID == "" {
		t.Fatalf("expected played card id in json output")
	}
	if card.Name == "" {
		t.Fatalf("expected played card name in json output")
	}
	if card.Info == nil {
		t.Fatalf("expected played card info in json output")
	}
	if card.Info.ID != card.ID {
		t.Fatalf("card info id does not match card id: %s vs %s", card.Info.ID, card.ID)
	}
	if card.Info.Name == "" {
		t.Fatalf("expected played card info name in json output")
	}
	if card.Info.Type == "" {
		t.Fatalf("expected played card info type in json output")
	}

	attack, ok, err := findFirstAttackEvent(data)
	if err != nil {
		t.Fatalf("find first attack event: %v", err)
	}
	if !ok {
		t.Fatalf("expected at least one attack event in json output")
	}
	if attack.Attacker == nil || attack.Attacker.Name == "" {
		t.Fatalf("expected human-readable attacker in attack event")
	}
	if attack.Target == nil || attack.Target.Name == "" {
		t.Fatalf("expected human-readable target in attack event")
	}

	discover, ok, err := findFirstDiscoverEvent(data)
	if err != nil {
		t.Fatalf("find first discover event: %v", err)
	}
	if !ok {
		t.Fatalf("expected at least one discover event in json output")
	}
	if len(discover.OptionCards) == 0 || discover.OptionCards[0].Name == "" {
		t.Fatalf("expected human-readable discover options in json output")
	}
	if discover.ChosenCardInfo == nil || discover.ChosenCardInfo.Name == "" {
		t.Fatalf("expected human-readable chosen discover card in json output")
	}

	outputEntries, err := os.ReadDir(projectOutputDirForTest())
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if !containsTimelineJSON(outputEntries) {
		t.Fatalf("expected auto-saved timeline json in output")
	}
}

func TestRootCommandRejectsInvalidLogLevel(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--log-level", "trace", "analyze"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected invalid log level error")
	}
	if !strings.Contains(err.Error(), "invalid log level") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func copyFileForTest(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func containsTimelineJSON(entries []os.DirEntry) bool {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return slices.ContainsFunc(names, func(name string) bool {
		return strings.HasPrefix(name, "timeline_") && strings.HasSuffix(name, ".json")
	})
}

func projectOutputDirForTest() string {
	return filepath.Clean(filepath.Join("..", "..", "output"))
}

type jsonTimelineTurn struct {
	Events []jsonTimelineEvent `json:"events"`
}

type jsonTimelineEvent struct {
	Type       string                     `json:"type"`
	Attributes map[string]json.RawMessage `json:"attributes"`
}

type jsonTimelineAttackEvent struct {
	Attacker *jsonTimelineEntity `json:"attacker"`
	Target   *jsonTimelineEntity `json:"target"`
}

type jsonTimelineDiscoverEvent struct {
	OptionCards    []jsonTimelineCard `json:"option_cards"`
	ChosenCardInfo *jsonTimelineCard  `json:"chosen_card_info"`
}

type jsonTimelineEntity struct {
	EntityID int    `json:"entity_id"`
	Name     string `json:"name"`
	CardID   string `json:"card_id"`
}

type jsonTimelineCard struct {
	ID   string                `json:"id"`
	Name string                `json:"name"`
	Info *jsonTimelineCardInfo `json:"info"`
}

type jsonTimelineCardInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func findFirstPlayedCard(data []byte) (*jsonTimelineCard, bool, error) {
	var turns []jsonTimelineTurn
	if err := json.Unmarshal(data, &turns); err != nil {
		return nil, false, err
	}

	for _, turn := range turns {
		for _, event := range turn.Events {
			if event.Type != "card_played" {
				continue
			}

			rawCard, ok := event.Attributes["card"]
			if !ok {
				continue
			}

			var card jsonTimelineCard
			if err := json.Unmarshal(rawCard, &card); err != nil {
				return nil, false, err
			}

			return &card, true, nil
		}
	}

	return nil, false, nil
}

func findFirstAttackEvent(data []byte) (*jsonTimelineAttackEvent, bool, error) {
	var turns []jsonTimelineTurn
	if err := json.Unmarshal(data, &turns); err != nil {
		return nil, false, err
	}

	for _, turn := range turns {
		for _, event := range turn.Events {
			if event.Type != "attack" {
				continue
			}

			payload, err := marshalEventAttributes(event.Attributes)
			if err != nil {
				return nil, false, err
			}

			var attack jsonTimelineAttackEvent
			if err := json.Unmarshal(payload, &attack); err != nil {
				return nil, false, err
			}

			return &attack, true, nil
		}
	}

	return nil, false, nil
}

func findFirstDiscoverEvent(data []byte) (*jsonTimelineDiscoverEvent, bool, error) {
	var turns []jsonTimelineTurn
	if err := json.Unmarshal(data, &turns); err != nil {
		return nil, false, err
	}

	for _, turn := range turns {
		for _, event := range turn.Events {
			if event.Type != "discover" {
				continue
			}

			payload, err := marshalEventAttributes(event.Attributes)
			if err != nil {
				return nil, false, err
			}

			var discover jsonTimelineDiscoverEvent
			if err := json.Unmarshal(payload, &discover); err != nil {
				return nil, false, err
			}

			return &discover, true, nil
		}
	}

	return nil, false, nil
}

func marshalEventAttributes(attributes map[string]json.RawMessage) ([]byte, error) {
	return json.Marshal(attributes)
}
