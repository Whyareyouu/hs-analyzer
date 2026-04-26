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

func TestAnalyzeCommandPrintsFormattedResultFromFixtureLogs(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", fixtureLogsDir("standard_game_1"),
		"--cards-db-path", fixtureCardsPath(),
		"--format", "text",
		"--lang", "ru",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	expected := strings.TrimSpace(readFixtureFile(t, fixtureExpectedPath("standard_game_1_text.txt")))
	if output != expected {
		t.Fatalf("unexpected output:\n--- got ---\n%s\n--- want ---\n%s", output, expected)
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

	outputPath := filepath.Join(t.TempDir(), "result.txt")
	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", fixtureLogsDir("standard_game_1"),
		"--cards-db-path", fixtureCardsPath(),
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

func TestAnalyzeCommandAcceptsExplicitPowerLogPathViaLogsPathFlag(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	powerLogPath := filepath.Join(fixtureLogsDir("standard_game_1"), "Power.log")

	cmd := newRootCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", powerLogPath,
		"--cards-db-path", fixtureCardsPath(),
		"--format", "text",
		"--lang", "ru",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	expected := strings.TrimSpace(readFixtureFile(t, fixtureExpectedPath("standard_game_1_text.txt")))
	if output != expected {
		t.Fatalf("unexpected output:\n--- got ---\n%s\n--- want ---\n%s", output, expected)
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

func TestAnalyzeCommandWritesJSONTurnsArrayToFileFromFixtureLogs(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "timeline.json")
	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", fixtureLogsDir("standard_game_1"),
		"--cards-db-path", fixtureCardsPath(),
		"--format", "json",
		"--json-mode", "replay",
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
	if !strings.Contains(string(data), `"mode": "replay"`) || !strings.Contains(string(data), `"opening_state"`) || !strings.Contains(string(data), `"state_after"`) {
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
	if len(discover.Options) == 0 || discover.Options[0].Name == "" {
		t.Fatalf("expected human-readable discover options in json output")
	}
	if discover.ChosenCard == nil || discover.ChosenCard.Name == "" {
		t.Fatalf("expected human-readable chosen discover card in json output")
	}

	var replayDocument jsonTimelineDocument
	if err := json.Unmarshal(data, &replayDocument); err != nil {
		t.Fatalf("unmarshal replay document: %v", err)
	}
	if replayDocument.Mode != "replay" {
		t.Fatalf("unexpected mode: %s", replayDocument.Mode)
	}
	if len(replayDocument.Matches) == 0 || len(replayDocument.Matches[0].Turns) == 0 {
		t.Fatalf("expected matches in replay json")
	}
	for _, want := range strings.Split(readFixtureFile(t, fixtureExpectedPath("standard_game_1_json_contains.txt")), "\n") {
		want = strings.TrimSpace(want)
		if want == "" {
			continue
		}
		if !strings.Contains(string(data), want) {
			t.Fatalf("json output does not contain %q:\n%s", want, string(data))
		}
	}

	outputEntries, err := os.ReadDir(projectOutputDirForTest())
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if !containsTimelineJSON(outputEntries) {
		t.Fatalf("expected auto-saved timeline json in output")
	}
	if !containsTimelineJSONMode(outputEntries, "replay") {
		t.Fatalf("expected replay auto-saved timeline json in output")
	}
}

func TestAnalyzeCommandWritesHumanJSONToFile(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "timeline-human.json")
	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", fixtureLogsDir("standard_game_1"),
		"--cards-db-path", fixtureCardsPath(),
		"--format", "json",
		"--json-mode", "human",
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
	if !strings.Contains(string(data), `"mode": "human"`) || !strings.Contains(string(data), `"opening_hand"`) {
		t.Fatalf("unexpected human json output:\n%s", string(data))
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

func TestAnalyzeCommandRejectsMissingCardsDatabase(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", fixtureLogsDir("standard_game_1"),
		"--cards-db-path", filepath.Join(t.TempDir(), "missing-cards.json"),
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected missing cards db error")
	}
	if !strings.Contains(err.Error(), "invalid cards database path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeCommandRejectsMissingLogsDirectory(t *testing.T) {
	t.Parallel()

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"analyze",
		"--logs-path", filepath.Join(t.TempDir(), "missing-logs"),
		"--cards-db-path", fixtureCardsPath(),
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected missing logs error")
	}
	if !strings.Contains(err.Error(), "invalid logs path") {
		t.Fatalf("unexpected error: %v", err)
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

func containsTimelineJSONMode(entries []os.DirEntry, mode string) bool {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	suffix := "_" + mode + ".json"
	return slices.ContainsFunc(names, func(name string) bool {
		return strings.HasPrefix(name, "timeline_") && strings.HasSuffix(name, suffix)
	})
}

func projectOutputDirForTest() string {
	return filepath.Clean(filepath.Join("..", "..", "output"))
}

type jsonTimelineDocument struct {
	Mode    string              `json:"mode"`
	Matches []jsonTimelineMatch `json:"matches"`
}

type jsonTimelineMatch struct {
	Turns []jsonTimelineTurn `json:"turns"`
}

type jsonTimelineTurn struct {
	OpeningHand []jsonTimelineCard  `json:"opening_hand"`
	Events      []jsonTimelineEvent `json:"events"`
}

type jsonTimelineEvent struct {
	Type       string                     `json:"type"`
	Action     map[string]json.RawMessage `json:"action"`
	Attributes map[string]json.RawMessage `json:"attributes"`
}

type jsonTimelineAttackEvent struct {
	Attacker *jsonTimelineEntity `json:"attacker"`
	Target   *jsonTimelineEntity `json:"target"`
}

type jsonTimelineDiscoverEvent struct {
	Options    []jsonTimelineCard `json:"options"`
	ChosenCard *jsonTimelineCard  `json:"chosen_card"`
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
	var document jsonTimelineDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, false, err
	}

	for _, match := range document.Matches {
		for _, turn := range match.Turns {
			for _, event := range turn.Events {
				if event.Type != "card_played" {
					continue
				}

				rawCard, ok := firstAvailableField(event, "card")
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
	}

	return nil, false, nil
}

func findFirstAttackEvent(data []byte) (*jsonTimelineAttackEvent, bool, error) {
	var document jsonTimelineDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, false, err
	}

	for _, match := range document.Matches {
		for _, turn := range match.Turns {
			for _, event := range turn.Events {
				if event.Type != "attack" {
					continue
				}

				payload, err := marshalEventPayload(event)
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
	}

	return nil, false, nil
}

func findFirstDiscoverEvent(data []byte) (*jsonTimelineDiscoverEvent, bool, error) {
	var document jsonTimelineDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, false, err
	}

	for _, match := range document.Matches {
		for _, turn := range match.Turns {
			for _, event := range turn.Events {
				if event.Type != "discover" {
					continue
				}

				payload, err := marshalEventPayload(event)
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
	}

	return nil, false, nil
}

func marshalEventPayload(event jsonTimelineEvent) ([]byte, error) {
	if len(event.Action) > 0 {
		return json.Marshal(event.Action)
	}
	return json.Marshal(event.Attributes)
}

func firstAvailableField(event jsonTimelineEvent, key string) (json.RawMessage, bool) {
	if raw, ok := event.Action[key]; ok {
		return raw, true
	}
	if raw, ok := event.Attributes[key]; ok {
		return raw, true
	}
	return nil, false
}
