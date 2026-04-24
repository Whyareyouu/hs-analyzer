package logreader

import "io"

const defaultScannerBufferSize = 1024 * 1024

// LogReader определяет интерфейс для последовательного чтения строк из логов.
type LogReader interface {
	Open(path string) error
	Next() (string, error)
	Close() error
}

// Ensure interface expectations remain explicit at compile time.
var _ LogReader = (*FileReader)(nil)

// IsEOF reports whether the returned error means that the log has been fully read.
func IsEOF(err error) bool {
	return err == io.EOF
}
