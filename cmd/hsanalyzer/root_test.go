package main

import (
	"bytes"
	"os"
	"path/filepath"
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
	copyFileForTest(t, filepath.Join("..", "..", "logs", "LoadingScreen_old.log"), filepath.Join(logsDir, "LoadingScreen_old.log"))

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
}

func TestAnalyzeCommandWritesOutputToFile(t *testing.T) {
	t.Parallel()

	logsDir := t.TempDir()
	copyFileForTest(t, filepath.Join("..", "..", "logs", "Decks.log"), filepath.Join(logsDir, "Decks.log"))
	copyFileForTest(t, filepath.Join("..", "..", "logs", "LoadingScreen_old.log"), filepath.Join(logsDir, "LoadingScreen_old.log"))

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
		format:      "json",
	})
	if err == nil {
		t.Fatalf("expected unsupported format error")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("unexpected error: %v", err)
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
