package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	diagnostics "github.com/rp1s/digreyt"
)

func TestPrintErrorRendersAllDiagnostics(t *testing.T) {
	arena := diagnostics.New("let value: string = 42")
	arena.Add(diagnostics.Error{
		CodeName:      "FirstError",
		Message:       "first failure",
		Arrow:         "here",
		Severity:      diagnostics.SeverityError,
		IsShowSnippet: true,
		Start:         4,
		End:           9,
		Pos: diagnostics.Position{
			FileName: "model.cm",
			Line:     1,
			Column:   5,
		},
	})
	arena.Add(diagnostics.Error{
		CodeName: "SecondError",
		Message:  "second failure",
		Severity: diagnostics.SeverityError,
		Pos: diagnostics.Position{
			FileName: "model.cm",
			Line:     1,
			Column:   21,
		},
	})

	var out bytes.Buffer
	if err := PrintError(&out, arena); err != nil {
		t.Fatalf("PrintError() failed: %v", err)
	}

	for _, expected := range []string{"FirstError", "first failure", "SecondError", "second failure", "model.cm"} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("diagnostic output does not contain %q:\n%s", expected, out.String())
		}
	}
}

func TestPrintErrorPrintsEveryJoinedError(t *testing.T) {
	var out bytes.Buffer
	err := errors.Join(errors.New("build failed"), errors.New("close failed"))
	if printErr := PrintError(&out, err); printErr != nil {
		t.Fatalf("PrintError() failed: %v", printErr)
	}

	for _, expected := range []string{"build failed", "close failed"} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("joined error output does not contain %q: %q", expected, out.String())
		}
	}
}

func TestPrintErrorReturnsWriterError(t *testing.T) {
	want := errors.New("write failed")
	err := PrintError(errorWriter{err: want}, errors.New("build failed"))
	if !errors.Is(err, want) {
		t.Fatalf("PrintError() error = %v, want %v", err, want)
	}
}

type errorWriter struct {
	err error
}

func (self errorWriter) Write([]byte) (int, error) {
	return 0, self.err
}
