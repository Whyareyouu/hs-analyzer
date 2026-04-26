package logreader

import (
	"context"
	"io"
)

const defaultScannerBufferSize = 1024 * 1024

// LogReader определяет интерфейс для последовательного чтения строк из логов.
type LogReader interface {
	Open(path string) error
	Next() (string, error)
	Close() error
}

// LiveLogReader defines a reader for growing log files.
type LiveLogReader interface {
	Open(path string) error
	Next(ctx context.Context) (string, error)
	State() LiveReaderState
	Close() error
}

// LiveReaderState exposes the current read position and wait state.
type LiveReaderState struct {
	Path         string
	Offset       int64
	PendingBytes int
	Waiting      bool
}

// Ensure interface expectations remain explicit at compile time.
var _ LogReader = (*FileReader)(nil)

// IsEOF reports whether the returned error means that the log has been fully read.
func IsEOF(err error) bool {
	return err == io.EOF
}
