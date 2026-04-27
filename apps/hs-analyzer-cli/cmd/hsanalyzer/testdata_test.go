package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixturePath(parts ...string) string {
	all := append([]string{projectRootForTest(nil), "tests", "testdata"}, parts...)
	return filepath.Join(all...)
}

func projectRootForTest(t *testing.T) string {
	if t != nil {
		t.Helper()
	}
	workingDir, err := os.Getwd()
	if err != nil {
		if t != nil {
			t.Fatalf("get working directory: %v", err)
		}
		panic(err)
	}
	projectRoot, err := findProjectRoot(workingDir)
	if err != nil {
		if t != nil {
			t.Fatalf("find project root: %v", err)
		}
		panic(err)
	}
	return projectRoot
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
