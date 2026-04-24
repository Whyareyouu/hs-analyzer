package logreader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

var errReaderNotOpened = errors.New("log reader is not opened")

// FileReader реализует LogReader для чтения строк из файла.
type FileReader struct {
	bufferSize int
	file       *os.File
	scanner    *bufio.Scanner
	path       string
}

// NewFileReader создает новый FileReader с заданным размером буфера сканера.
func NewFileReader(bufferSize int) *FileReader {
	if bufferSize <= 0 {
		bufferSize = defaultScannerBufferSize
	}

	return &FileReader{
		bufferSize: bufferSize,
	}
}

// Open открывает лог-файл и подготавливает последовательное чтение строк.
func (r *FileReader) Open(path string) error {
	if r == nil {
		return errors.New("file reader is nil")
	}

	if err := r.Close(); err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open log file %q: %w", path, err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), r.bufferSize)

	r.file = file
	r.scanner = scanner
	r.path = path

	return nil
}

// Next возвращает следующую строку лога.
func (r *FileReader) Next() (string, error) {
	if r == nil {
		return "", errors.New("file reader is nil")
	}
	if r.scanner == nil {
		return "", errReaderNotOpened
	}

	if r.scanner.Scan() {
		return r.scanner.Text(), nil
	}

	if err := r.scanner.Err(); err != nil {
		return "", fmt.Errorf("read log file %q: %w", r.path, err)
	}

	return "", io.EOF
}

// Close освобождает ресурсы, связанные с открытым лог-файлом.
func (r *FileReader) Close() error {
	if r == nil {
		return errors.New("file reader is nil")
	}

	var err error
	if r.file != nil {
		err = r.file.Close()
	}

	r.file = nil
	r.scanner = nil
	r.path = ""

	if err != nil {
		return fmt.Errorf("close log file: %w", err)
	}

	return nil
}
