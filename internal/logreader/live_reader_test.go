package logreader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLiveFileReaderRejectsReadBeforeOpen(t *testing.T) {
	t.Parallel()

	reader := NewLiveFileReader(LiveReaderConfig{PollInterval: 10 * time.Millisecond})
	_, err := reader.Next(context.Background())
	if !errors.Is(err, errLiveReaderNotOpened) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLiveFileReaderReadsAppendedLinesIncrementally(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "Power.log")
	if err := os.WriteFile(logPath, []byte("first\n"), 0o600); err != nil {
		t.Fatalf("write initial log: %v", err)
	}

	reader := NewLiveFileReader(LiveReaderConfig{PollInterval: 10 * time.Millisecond})
	if err := reader.Open(logPath); err != nil {
		t.Fatalf("open log: %v", err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
	})

	line, err := reader.Next(context.Background())
	if err != nil {
		t.Fatalf("read first line: %v", err)
	}
	if line != "first" {
		t.Fatalf("unexpected first line: %q", line)
	}

	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		line, err := reader.Next(context.Background())
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- line
	}()

	time.Sleep(30 * time.Millisecond)
	if err := appendToFile(logPath, "second\n"); err != nil {
		t.Fatalf("append second line: %v", err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("unexpected read error: %v", err)
	case line := <-resultCh:
		if line != "second" {
			t.Fatalf("unexpected appended line: %q", line)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for appended line")
	}

	state := reader.State()
	if state.Offset <= int64(len("first\n")) {
		t.Fatalf("expected offset to advance, got %d", state.Offset)
	}
}

func TestLiveFileReaderWaitsForCompleteLine(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "Power.log")
	if err := os.WriteFile(logPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}

	reader := NewLiveFileReader(LiveReaderConfig{PollInterval: 10 * time.Millisecond})
	if err := reader.Open(logPath); err != nil {
		t.Fatalf("open log: %v", err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
	})

	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		line, err := reader.Next(context.Background())
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- line
	}()

	time.Sleep(30 * time.Millisecond)
	if err := appendToFile(logPath, "partial"); err != nil {
		t.Fatalf("append partial line: %v", err)
	}

	select {
	case line := <-resultCh:
		t.Fatalf("reader returned incomplete line %q", line)
	case err := <-errCh:
		t.Fatalf("unexpected read error: %v", err)
	case <-time.After(80 * time.Millisecond):
	}

	if err := appendToFile(logPath, " line\n"); err != nil {
		t.Fatalf("append line tail: %v", err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("unexpected read error: %v", err)
	case line := <-resultCh:
		if line != "partial line" {
			t.Fatalf("unexpected reconstructed line: %q", line)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for complete line")
	}
}

func TestLiveFileReaderHandlesTruncateByRestartingFromBeginning(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "Power.log")
	if err := os.WriteFile(logPath, []byte("first\n"), 0o600); err != nil {
		t.Fatalf("write initial log: %v", err)
	}

	reader := NewLiveFileReader(LiveReaderConfig{PollInterval: 10 * time.Millisecond})
	if err := reader.Open(logPath); err != nil {
		t.Fatalf("open log: %v", err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
	})

	line, err := reader.Next(context.Background())
	if err != nil {
		t.Fatalf("read first line: %v", err)
	}
	if line != "first" {
		t.Fatalf("unexpected first line: %q", line)
	}

	if err := os.WriteFile(logPath, []byte("second\n"), 0o600); err != nil {
		t.Fatalf("truncate and rewrite log: %v", err)
	}

	line, err = reader.Next(context.Background())
	if err != nil {
		t.Fatalf("read second line after truncate: %v", err)
	}
	if line != "second" {
		t.Fatalf("unexpected line after truncate: %q", line)
	}
}

func TestLiveFileReaderWaitsForMissingFileAndResumesWhenItAppears(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "Power.log")

	reader := NewLiveFileReader(LiveReaderConfig{PollInterval: 10 * time.Millisecond})
	if err := reader.Open(logPath); err != nil {
		t.Fatalf("open missing log path: %v", err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
	})

	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		line, err := reader.Next(context.Background())
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- line
	}()

	time.Sleep(40 * time.Millisecond)
	if err := os.WriteFile(logPath, []byte("appeared\n"), 0o600); err != nil {
		t.Fatalf("create log file: %v", err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("unexpected read error: %v", err)
	case line := <-resultCh:
		if line != "appeared" {
			t.Fatalf("unexpected line after file appears: %q", line)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for created file")
	}
}

func TestLiveFileReaderCanBeCanceledWhileWaiting(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "Power.log")
	if err := os.WriteFile(logPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}

	reader := NewLiveFileReader(LiveReaderConfig{PollInterval: 50 * time.Millisecond})
	if err := reader.Open(logPath); err != nil {
		t.Fatalf("open log: %v", err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Fatalf("close reader: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	_, err := reader.Next(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}

func TestLiveFileReaderStartsFromConfiguredOffset(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "Power.log")
	if err := os.WriteFile(logPath, []byte("skip-me\nread-me\n"), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	reader := NewLiveFileReader(LiveReaderConfig{
		PollInterval: 10 * time.Millisecond,
		StartOffset:  int64(len("skip-me\n")),
	})
	if err := reader.Open(logPath); err != nil {
		t.Fatalf("open live reader: %v", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	line, err := reader.Next(context.Background())
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if line != "read-me" {
		t.Fatalf("unexpected line: %q", line)
	}
}

func appendToFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.WriteString(content); err != nil {
		return err
	}

	return file.Sync()
}
