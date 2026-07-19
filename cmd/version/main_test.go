package main

import (
	"errors"
	"path/filepath"
	"testing"
)

type errorWriter struct {
	err error
}

func (self errorWriter) Write([]byte) (int, error) {
	return 0, self.err
}

func TestRunPropagatesOutputError(t *testing.T) {
	want := errors.New("write failed")
	err := run(errorWriter{err: want}, []string{filepath.Join(t.TempDir(), "version.gen.go")})
	if !errors.Is(err, want) {
		t.Fatalf("run() error = %v, want %v", err, want)
	}
}
