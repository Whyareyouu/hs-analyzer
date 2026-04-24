package logreader

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileReaderReadsLinesSequentially(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "Power.log")
	content := "line one\nline two\nline three\n"
	if err := os.WriteFile(logPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write log file: %v", err)
	}

	reader := NewFileReader(0)
	if err := reader.Open(logPath); err != nil {
		t.Fatalf("open log file: %v", err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
	})

	var lines []string
	for {
		line, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("read next line: %v", err)
		}
		lines = append(lines, line)
	}

	want := []string{"line one", "line two", "line three"}
	if len(lines) != len(want) {
		t.Fatalf("unexpected line count: got %d want %d", len(lines), len(want))
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("unexpected line %d: got %q want %q", i, lines[i], want[i])
		}
	}
}

func TestFileReaderReturnsReadableErrorForMissingFile(t *testing.T) {
	t.Parallel()

	reader := NewFileReader(0)
	err := reader.Open(filepath.Join(t.TempDir(), "missing.log"))
	if err == nil {
		t.Fatalf("expected open error")
	}
	if !strings.Contains(err.Error(), "open log file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileReaderRejectsReadBeforeOpen(t *testing.T) {
	t.Parallel()

	reader := NewFileReader(0)
	_, err := reader.Next()
	if !errors.Is(err, errReaderNotOpened) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileReaderSupportsLongLines(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "Zone.log")
	longLine := strings.Repeat("x", 70*1024)
	if err := os.WriteFile(logPath, []byte(longLine+"\n"), 0o600); err != nil {
		t.Fatalf("write log file: %v", err)
	}

	reader := NewFileReader(128 * 1024)
	if err := reader.Open(logPath); err != nil {
		t.Fatalf("open log file: %v", err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
	})

	line, err := reader.Next()
	if err != nil {
		t.Fatalf("read long line: %v", err)
	}
	if line != longLine {
		t.Fatalf("unexpected long line length: got %d want %d", len(line), len(longLine))
	}

	_, err = reader.Next()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestFileReaderCanReopenAnotherFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	firstPath := filepath.Join(tmpDir, "LoadingScreen.log")
	secondPath := filepath.Join(tmpDir, "Decks.log")

	if err := os.WriteFile(firstPath, []byte("first\n"), 0o600); err != nil {
		t.Fatalf("write first log: %v", err)
	}
	if err := os.WriteFile(secondPath, []byte("second\n"), 0o600); err != nil {
		t.Fatalf("write second log: %v", err)
	}

	reader := NewFileReader(0)
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
	})

	if err := reader.Open(firstPath); err != nil {
		t.Fatalf("open first log: %v", err)
	}
	firstLine, err := reader.Next()
	if err != nil {
		t.Fatalf("read first line: %v", err)
	}
	if firstLine != "first" {
		t.Fatalf("unexpected first line: %q", firstLine)
	}

	if err := reader.Open(secondPath); err != nil {
		t.Fatalf("open second log: %v", err)
	}
	secondLine, err := reader.Next()
	if err != nil {
		t.Fatalf("read second line: %v", err)
	}
	if secondLine != "second" {
		t.Fatalf("unexpected second line: %q", secondLine)
	}
}
