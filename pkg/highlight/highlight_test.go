package highlight

import "testing"

func TestHighlightReturnsTokenRanges(t *testing.T) {
	source := `pub let name: string = "Caramel" // default value`
	spans := Highlight(source)
	if len(spans) == 0 {
		t.Fatal("Highlight() returned no spans")
	}

	want := map[string]Color{
		"pub":              DefaultTheme().Keyword,
		"let":              DefaultTheme().Keyword,
		"name":             DefaultTheme().Identifier,
		"string":           DefaultTheme().Type,
		`"Caramel"`:        DefaultTheme().String,
		"// default value": DefaultTheme().Comment,
	}
	for text, color := range want {
		if !containsSpan(source, spans, text, color) {
			t.Errorf("missing span for %q with color %#v", text, color)
		}
	}

	for _, span := range spans {
		if span.Start >= span.End || span.End > uint64(len(source)) {
			t.Fatalf("invalid span: %#v", span)
		}
	}
}

func TestHighlightUsesUTF8ByteOffsets(t *testing.T) {
	source := `let имя: string = "да"`
	spans := Highlight(source)
	if !containsSpan(source, spans, "имя", DefaultTheme().Identifier) {
		t.Fatalf("identifier range does not use valid UTF-8 byte offsets: %#v", spans)
	}
}

func TestHighlightWithThemeUsesCustomColors(t *testing.T) {
	theme := DefaultTheme()
	theme.Keyword = Color{R: 1, G: 2, B: 3}

	spans := HighlightWithTheme("let value = true", theme)
	if !containsSpan("let value = true", spans, "let", theme.Keyword) {
		t.Fatalf("custom keyword color was not used: %#v", spans)
	}
}

func containsSpan(source string, spans []Span, text string, color Color) bool {
	for _, span := range spans {
		if source[span.Start:span.End] == text && span.Color == color {
			return true
		}
	}
	return false
}
