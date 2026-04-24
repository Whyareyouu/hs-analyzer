package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readLogLines(t *testing.T, fileName string) []string {
	t.Helper()

	path := filepath.Join("..", "..", "logs", fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log %s: %v", path, err)
	}

	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	return strings.Split(normalized, "\n")
}

func findLogLine(t *testing.T, fileName, contains string) string {
	t.Helper()

	for _, line := range readLogLines(t, fileName) {
		if strings.Contains(line, contains) {
			return line
		}
	}

	t.Fatalf("line containing %q not found in %s", contains, fileName)
	return ""
}
