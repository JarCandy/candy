package highlight

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/rp1s/colorista"
	diagnostics "github.com/rp1s/digreyt"
)

type DiagnosticRenderer struct {
	colorista  *colorista.Colorista
	syntax     Theme
	diagnostic diagnostics.RenderTheme
}

func NewDiagnosticRenderer() *DiagnosticRenderer {
	return NewDiagnosticRendererWithTheme(TerminalTheme())
}

func NewDiagnosticRendererWithTheme(theme Theme) *DiagnosticRenderer {
	return &DiagnosticRenderer{
		colorista:  colorista.NewColorista(colorista.ThemeAuto),
		syntax:     theme,
		diagnostic: diagnostics.DefaultRenderTheme(),
	}
}

func (self *DiagnosticRenderer) Render(w io.Writer, source string, err diagnostics.Error) error {
	if self == nil {
		self = NewDiagnosticRenderer()
	}

	var out strings.Builder
	self.writeHeader(&out, err)
	if err.IsShowSnippet {
		self.writeSnippet(&out, source, err)
	}
	self.writeDescription(&out, err)

	_, writeErr := fmt.Fprintln(w, out.String())
	return writeErr
}

func (self *DiagnosticRenderer) writeHeader(out *strings.Builder, err diagnostics.Error) {
	view := self.severityView(err.Severity)
	codeName := err.CodeName
	if codeName == "" {
		codeName = view.Label
	}

	if view.Symbol != "" {
		out.WriteString(self.apply(view.Symbol, view.SymbolStyles))
		out.WriteByte(' ')
	}
	out.WriteString(self.apply(codeName, view.CodeStyles))
	out.WriteString(self.apply(": ", self.diagnostic.MutedStyles))
	out.WriteString(err.Message)
	out.WriteByte('\n')

	if err.Pos.Line == 0 {
		out.WriteString(self.apply(self.diagnostic.ModuleLabel, self.diagnostic.LocationStyles))
		out.WriteString(self.apply(err.Pos.FileName, self.diagnostic.LocationStyles))
		out.WriteByte('\n')
		return
	}

	out.WriteString(self.apply(self.diagnostic.LocationArrow, self.diagnostic.LocationStyles))
	out.WriteString(self.apply(err.Pos.FileName, self.diagnostic.LocationStyles))
	out.WriteByte(' ')
	out.WriteString(self.apply(strconv.FormatUint(err.Pos.Line, 10), self.diagnostic.LocationStyles))
	out.WriteString(self.apply(":", self.diagnostic.LocationStyles))
	out.WriteString(self.apply(strconv.FormatUint(err.Pos.Column, 10), self.diagnostic.LocationStyles))
	out.WriteString("\n\n")
}

func (self *DiagnosticRenderer) writeSnippet(out *strings.Builder, source string, err diagnostics.Error) {
	lines := contextSourceLines(source, int(err.Pos.Line), self.diagnostic.ContextLines)
	if len(lines) == 0 {
		return
	}

	width := len(strconv.Itoa(strings.Count(source, "\n") + 1))
	spans := HighlightWithTheme(source, self.syntax)
	for _, line := range lines {
		out.WriteString(self.apply(fmt.Sprintf("%*d │ ", width, line.number), self.diagnostic.MutedStyles))
		self.writeHighlightedLine(out, source, line, spans)
		out.WriteByte('\n')

		if line.number == int(err.Pos.Line) {
			self.writeCaret(out, source, line.text, err, width)
		}
	}
	out.WriteByte('\n')
}

func (self *DiagnosticRenderer) writeHighlightedLine(out *strings.Builder, source string, line sourceLine, spans []Span) {
	cursor := uint64(line.start)
	lineEnd := uint64(line.end)

	for _, span := range spans {
		if span.End <= cursor {
			continue
		}
		if span.Start >= lineEnd {
			break
		}

		start := max(span.Start, cursor)
		end := min(span.End, lineEnd)
		if start > cursor {
			out.WriteString(source[cursor:start])
		}
		if end > start {
			out.WriteString(self.color(source[start:end], span.Color))
			cursor = end
		}
	}

	if cursor < lineEnd {
		out.WriteString(source[cursor:lineEnd])
	}
}

func (self *DiagnosticRenderer) writeCaret(out *strings.Builder, source, line string, err diagnostics.Error, width int) {
	view := self.severityView(err.Severity)
	out.WriteString(self.apply(fmt.Sprintf("%*s │ ", width, ""), self.diagnostic.MutedStyles))

	runes := []rune(line)
	column := clamp(int(err.Pos.Column)-1, 0, len(runes))
	out.WriteString(strings.Repeat(" ", column))

	length := diagnosticRuneLength(source, err)
	if column+length > len(runes) {
		length = len(runes) - column
	}
	if length < 1 {
		length = 1
	}
	out.WriteString(self.apply(strings.Repeat("^", length), view.CaretStyles))
	if err.Arrow != "" {
		out.WriteByte(' ')
		out.WriteString(self.apply(err.Arrow, view.ArrowStyles))
	}
	out.WriteByte('\n')
}

func (self *DiagnosticRenderer) writeDescription(out *strings.Builder, err diagnostics.Error) {
	view := self.severityView(err.Severity)
	for _, description := range err.Description {
		out.WriteString(self.apply(self.diagnostic.DescriptionBullet, view.BulletStyles))
		out.WriteString(self.apply(description, view.DescriptionStyles))
		out.WriteByte('\n')
	}
}

func (self *DiagnosticRenderer) severityView(severity diagnostics.Severity) diagnostics.SeverityView {
	if view, ok := self.diagnostic.SeverityViews[severity]; ok {
		return view
	}
	return diagnostics.SeverityView{
		Symbol:       "?",
		Label:        severity.String(),
		SymbolStyles: []colorista.Style{colorista.Bold, colorista.BrightWhite},
		CodeStyles:   []colorista.Style{colorista.Bold, colorista.BrightWhite},
		CaretStyles:  []colorista.Style{colorista.Bold, colorista.BrightWhite},
		ArrowStyles:  []colorista.Style{colorista.Bold, colorista.BrightWhite},
		BulletStyles: []colorista.Style{colorista.Bold, colorista.BrightWhite},
	}
}

func (self *DiagnosticRenderer) apply(text string, styles []colorista.Style) string {
	return self.colorista.Apply(text, styles...)
}

func (self *DiagnosticRenderer) color(text string, color Color) string {
	return self.colorista.Apply(text, colorista.Rgb(colorista.RGB{
		R: color.R,
		G: color.G,
		B: color.B,
	}))
}

type sourceLine struct {
	number int
	start  int
	end    int
	text   string
}

func contextSourceLines(source string, target, context int) []sourceLine {
	if target < 1 {
		return nil
	}

	lines := make([]sourceLine, 0, strings.Count(source, "\n")+1)
	start := 0
	for i := 0; i <= len(source); i++ {
		if i != len(source) && source[i] != '\n' {
			continue
		}
		end := i
		if end > start && source[end-1] == '\r' {
			end--
		}
		lines = append(lines, sourceLine{
			number: len(lines) + 1,
			start:  start,
			end:    end,
			text:   source[start:end],
		})
		start = i + 1
	}

	targetIndex := target - 1
	if targetIndex >= len(lines) {
		return nil
	}
	from := clamp(targetIndex-context+1, 0, targetIndex)
	return lines[from : targetIndex+1]
}

func diagnosticRuneLength(source string, err diagnostics.Error) int {
	if err.Start >= err.End || err.End > uint64(len(source)) {
		return 1
	}
	length := utf8.RuneCountInString(source[err.Start:err.End])
	if length < 1 {
		return 1
	}
	return length
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
