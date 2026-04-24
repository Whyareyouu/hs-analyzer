package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixturePath(parts ...string) string {
	all := append([]string{"..", "..", "tests", "testdata"}, parts...)
	return filepath.Join(all...)
}

func fixtureLogsDir(name string) string {
	return fixturePath(name, "logs")
}

func fixtureCardsPath() string {
	return fixturePath("cards", "cards_sample.json")
}

func fixtureExpectedPath(name string) string {
	return fixturePath("expected", name)
}

func readFixtureFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	return strings.ReplaceAll(string(data), "\r\n", "\n")
}
