package sdk

import (
	"errors"
	"testing"
)

type errorWriter struct {
	err error
}

func (self errorWriter) Write([]byte) (int, error) {
	return 0, self.err
}

func TestPrintErrorPropagatesWriterError(t *testing.T) {
	want := errors.New("write failed")
	err := PrintError(errorWriter{err: want}, NewError("analysis failed"))
	if !errors.Is(err, want) {
		t.Fatalf("PrintError() error = %v, want %v", err, want)
	}
}
