package logreader

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultLivePollInterval = 250 * time.Millisecond
const liveReaderAnchorBytes = 256

var errLiveReaderNotOpened = errors.New("live log reader is not opened")

// LiveReaderConfig controls incremental reading from a growing log file.
type LiveReaderConfig struct {
	BufferSize   int
	PollInterval time.Duration
	StartOffset  int64
}

// LiveFileReader incrementally reads appended lines from a growing file.
//
// It keeps track of the current byte offset, waits for new data when EOF is
// reached, and reopens/reset state when the active file is truncated or
// replaced. This is the behavior needed for long-running observation of
// Hearthstone logs, which continue to grow while the player is in game.
type LiveFileReader struct {
	bufferSize   int
	pollInterval time.Duration
	startOffset  int64

	path    string
	offset  int64
	pending string
	waiting bool
	anchor  []byte

	lastInfo os.FileInfo
	file     *os.File
	reader   *bufio.Reader
}

// NewLiveFileReader creates a reader for growing log files.
func NewLiveFileReader(cfg LiveReaderConfig) *LiveFileReader {
	bufferSize := cfg.BufferSize
	if bufferSize <= 0 {
		bufferSize = defaultScannerBufferSize
	}

	pollInterval := cfg.PollInterval
	if pollInterval <= 0 {
		pollInterval = defaultLivePollInterval
	}

	return &LiveFileReader{
		bufferSize:   bufferSize,
		pollInterval: pollInterval,
		startOffset:  cfg.StartOffset,
	}
}

// Ensure interface expectations remain explicit at compile time.
var _ LiveLogReader = (*LiveFileReader)(nil)

// Open sets the path to monitor. The file may be absent at open time.
func (r *LiveFileReader) Open(path string) error {
	if r == nil {
		return errors.New("live file reader is nil")
	}
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return errors.New("log path is empty")
	}

	if err := r.Close(); err != nil {
		return err
	}

	cleaned := filepath.Clean(trimmed)
	if info, err := os.Stat(cleaned); err == nil {
		if info.IsDir() {
			return fmt.Errorf("open live log file %q: path is a directory", cleaned)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("open live log file %q: %w", cleaned, err)
	}

	r.path = cleaned
	r.offset = r.startOffset
	if r.offset < 0 {
		r.offset = 0
	}
	r.pending = ""
	r.waiting = false
	r.anchor = nil
	r.lastInfo = nil

	return nil
}

// Next returns the next complete line, waiting for appended data when needed.
func (r *LiveFileReader) Next(ctx context.Context) (string, error) {
	if r == nil {
		return "", errors.New("live file reader is nil")
	}
	if strings.TrimSpace(r.path) == "" {
		return "", errLiveReaderNotOpened
	}

	for {
		present, err := r.syncState()
		if err != nil {
			return "", err
		}
		if !present {
			if err := r.wait(ctx); err != nil {
				return "", err
			}
			continue
		}

		if err := r.ensureOpen(); err != nil {
			if os.IsNotExist(err) {
				if err := r.wait(ctx); err != nil {
					return "", err
				}
				continue
			}
			return "", err
		}

		line, complete, err := r.readNextLine()
		if err != nil {
			return "", err
		}
		if complete {
			return line, nil
		}

		if err := r.wait(ctx); err != nil {
			return "", err
		}
	}
}

// State returns the current live reader state.
func (r *LiveFileReader) State() LiveReaderState {
	if r == nil {
		return LiveReaderState{}
	}
	return LiveReaderState{
		Path:         r.path,
		Offset:       r.offset,
		PendingBytes: len(r.pending),
		Waiting:      r.waiting,
	}
}

// Close releases open file resources while keeping no pending state.
func (r *LiveFileReader) Close() error {
	if r == nil {
		return errors.New("live file reader is nil")
	}

	var err error
	if r.file != nil {
		err = r.file.Close()
	}

	r.file = nil
	r.reader = nil
	r.path = ""
	r.offset = 0
	r.pending = ""
	r.waiting = false
	r.anchor = nil
	r.lastInfo = nil

	if err != nil {
		return fmt.Errorf("close live log file: %w", err)
	}

	return nil
}

func (r *LiveFileReader) syncState() (bool, error) {
	info, err := os.Stat(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			r.waiting = true
			r.resetTracking()
			return false, nil
		}
		return false, fmt.Errorf("stat live log file %q: %w", r.path, err)
	}
	if info.IsDir() {
		return false, fmt.Errorf("stat live log file %q: path is a directory", r.path)
	}

	r.waiting = false

	if r.lastInfo != nil {
		if !os.SameFile(r.lastInfo, info) {
			r.resetTracking()
			return true, nil
		}
		if info.Size() < r.offset {
			r.resetTracking()
			return true, nil
		}
		if fileInfoChanged(r.lastInfo, info) {
			matches, err := r.anchorMatchesCurrentFile()
			if err != nil {
				return false, err
			}
			if !matches {
				r.resetTracking()
				return true, nil
			}
		}
	}

	r.lastInfo = info

	return true, nil
}

func (r *LiveFileReader) ensureOpen() error {
	if r.file != nil && r.reader != nil {
		return nil
	}

	file, err := os.Open(r.path)
	if err != nil {
		return fmt.Errorf("open live log file %q: %w", r.path, err)
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("stat live log file %q: %w", r.path, err)
	}

	if _, err := file.Seek(r.offset, io.SeekStart); err != nil {
		_ = file.Close()
		return fmt.Errorf("seek live log file %q: %w", r.path, err)
	}

	r.file = file
	r.reader = bufio.NewReaderSize(file, r.bufferSize)
	r.lastInfo = info

	return nil
}

func (r *LiveFileReader) readNextLine() (string, bool, error) {
	segment, err := r.reader.ReadString('\n')
	r.offset += int64(len(segment))
	r.appendAnchor(segment)

	switch {
	case err == nil:
		line := r.pending + trimTrailingNewline(segment)
		r.pending = ""
		return line, true, nil
	case errors.Is(err, io.EOF):
		if len(segment) > 0 {
			r.pending += segment
		}
		if closeErr := r.closeHandle(); closeErr != nil {
			return "", false, closeErr
		}
		return "", false, nil
	default:
		_ = r.closeHandle()
		return "", false, fmt.Errorf("read live log file %q: %w", r.path, err)
	}
}

func (r *LiveFileReader) wait(ctx context.Context) error {
	r.waiting = true
	timer := time.NewTimer(r.pollInterval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (r *LiveFileReader) resetTracking() {
	_ = r.closeHandle()
	r.offset = 0
	r.pending = ""
	r.anchor = nil
	r.lastInfo = nil
}

func (r *LiveFileReader) closeHandle() error {
	var err error
	if r.file != nil {
		err = r.file.Close()
	}
	r.file = nil
	r.reader = nil
	if err != nil {
		return fmt.Errorf("close live log file: %w", err)
	}
	return nil
}

func trimTrailingNewline(value string) string {
	value = strings.TrimSuffix(value, "\n")
	value = strings.TrimSuffix(value, "\r")
	return value
}

func (r *LiveFileReader) appendAnchor(segment string) {
	if len(segment) == 0 {
		return
	}

	r.anchor = append(r.anchor, segment...)
	if len(r.anchor) > liveReaderAnchorBytes {
		r.anchor = append([]byte(nil), r.anchor[len(r.anchor)-liveReaderAnchorBytes:]...)
	}
}

func (r *LiveFileReader) anchorMatchesCurrentFile() (bool, error) {
	if len(r.anchor) == 0 || r.offset == 0 {
		return true, nil
	}

	file, err := os.Open(r.path)
	if err != nil {
		return false, fmt.Errorf("open live log file %q for anchor check: %w", r.path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	start := r.offset - int64(len(r.anchor))
	if start < 0 {
		start = 0
	}

	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return false, fmt.Errorf("seek live log file %q for anchor check: %w", r.path, err)
	}

	want := r.anchor
	if int64(len(want)) > r.offset {
		want = want[len(want)-int(r.offset):]
	}

	buffer := make([]byte, len(want))
	if _, err := io.ReadFull(file, buffer); err != nil {
		return false, fmt.Errorf("read live log file %q for anchor check: %w", r.path, err)
	}

	return string(buffer) == string(want), nil
}

func fileInfoChanged(previous, current os.FileInfo) bool {
	if previous == nil || current == nil {
		return false
	}
	return previous.Size() != current.Size() || !previous.ModTime().Equal(current.ModTime())
}
