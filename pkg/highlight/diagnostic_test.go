package highlight

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	diagnostics "github.com/rp1s/digreyt"
)

func TestDiagnosticRendererHighlightsSource(t *testing.T) {
	source := `pub name: string = "none" // default`
	err := diagnostics.Error{
		CodeName:      "TypeMismatch",
		Message:       "type mismatch",
		Arrow:         "expected int",
		Severity:      diagnostics.SeverityError,
		IsShowSnippet: true,
		Start:         4,
		End:           8,
		Pos: diagnostics.Position{
			FileName: "model.cm",
			Line:     1,
			Column:   5,
		},
	}

	var out bytes.Buffer
	if renderErr := NewDiagnosticRenderer().Render(&out, source, err); renderErr != nil {
		t.Fatalf("Render() failed: %v", renderErr)
	}

	for _, color := range []Color{TerminalTheme().Keyword, TerminalTheme().Type, TerminalTheme().String, TerminalTheme().Comment} {
		sequence := fmt.Sprintf("38;2;%d;%d;%d", color.R, color.G, color.B)
		if !strings.Contains(out.String(), sequence) {
			t.Errorf("output does not contain syntax color %s:\n%s", sequence, out.String())
		}
	}
	if !strings.Contains(out.String(), "^^^^") {
		t.Fatalf("output does not contain the diagnostic caret:\n%s", out.String())
	}
}

func TestContextSourceLinesPreservesByteOffsets(t *testing.T) {
	source := "let имя = true\r\nlet next = false"
	lines := contextSourceLines(source, 2, 5)
	if len(lines) != 2 {
		t.Fatalf("contextSourceLines() returned %d lines, want 2", len(lines))
	}
	if got := source[lines[1].start:lines[1].end]; got != "let next = false" {
		t.Fatalf("second line = %q", got)
	}
}
