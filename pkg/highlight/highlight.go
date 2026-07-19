package highlight

import (
	"github.com/caramelang/caramel/internal/parser/lexer"
	"github.com/caramelang/caramel/internal/parser/token"
)

type Color struct {
	R uint8
	G uint8
	B uint8
}

type Span struct {
	Start uint64
	End   uint64
	Color Color
}

type Theme struct {
	Keyword     Color
	Type        Color
	Identifier  Color
	String      Color
	Number      Color
	Boolean     Color
	Comment     Color
	Operator    Color
	Punctuation Color
	Invalid     Color
}

func DefaultTheme() Theme {
	return Theme{
		Keyword:     Color{R: 124, G: 58, B: 237},
		Type:        Color{R: 3, G: 105, B: 161},
		Identifier:  Color{R: 31, G: 41, B: 55},
		String:      Color{R: 4, G: 120, B: 87},
		Number:      Color{R: 180, G: 83, B: 9},
		Boolean:     Color{R: 194, G: 65, B: 12},
		Comment:     Color{R: 107, G: 114, B: 128},
		Operator:    Color{R: 190, G: 24, B: 93},
		Punctuation: Color{R: 71, G: 85, B: 105},
		Invalid:     Color{R: 220, G: 38, B: 38},
	}
}

func TerminalTheme() Theme {
	return Theme{
		Keyword:     Color{R: 198, G: 120, B: 221},
		Type:        Color{R: 86, G: 156, B: 214},
		Identifier:  Color{R: 220, G: 220, B: 220},
		String:      Color{R: 106, G: 153, B: 85},
		Number:      Color{R: 181, G: 206, B: 168},
		Boolean:     Color{R: 206, G: 145, B: 120},
		Comment:     Color{R: 128, G: 128, B: 128},
		Operator:    Color{R: 212, G: 212, B: 212},
		Punctuation: Color{R: 180, G: 180, B: 180},
		Invalid:     Color{R: 244, G: 71, B: 71},
	}
}

func Highlight(text string) []Span {
	return HighlightWithTheme(text, DefaultTheme())
}

func HighlightWithTheme(text string, theme Theme) []Span {
	source := []byte(text)
	lex := lexer.New(source, "")
	spans := make([]Span, 0)

	for {
		tk := lex.NextToken()
		if tk.Kind == token.EOF {
			return spans
		}

		color, ok := colorFor(source, tk, theme)
		if !ok || tk.Start >= tk.End || tk.End > uint64(len(source)) {
			continue
		}
		spans = append(spans, Span{Start: tk.Start, End: tk.End, Color: color})
	}
}

func colorFor(source []byte, tk token.Token, theme Theme) (Color, bool) {
	switch tk.Kind {
	case token.PACKAGE, token.USE, token.AS, token.PUB, token.LET:
		return theme.Keyword, true
	case token.STRING, token.RAW_STRING, token.CHARACTER:
		return theme.String, true
	case token.INTEGER, token.IMAGINARY, token.FLOATING:
		return theme.Number, true
	case token.TRUE, token.FALSE:
		return theme.Boolean, true
	case token.COMMENT, token.M_COMMENT:
		return theme.Comment, true
	case token.IDENTIFIER:
		if isBuiltinType(source, tk) {
			return theme.Type, true
		}
		return theme.Identifier, true
	case token.ASSIGN, token.TRANSITION, token.RRT,
		token.SUB, token.ADD, token.MUL, token.DIV, token.MOD, token.POW, token.RA:
		return theme.Operator, true
	case token.ATTR_S, token.ATTR_E, token.TEMPLATE_S, token.TEMPLATE_E,
		token.L_PAREN, token.R_PAREN, token.L_BRACE, token.R_BRACE,
		token.L_BRACK, token.R_BRACK, token.COLON, token.D_COLON, token.DOT, token.COMMA:
		return theme.Punctuation, true
	case token.ILLEGAL:
		return theme.Invalid, true
	default:
		return Color{}, false
	}
}

func isBuiltinType(source []byte, tk token.Token) bool {
	if tk.Start >= tk.End || tk.End > uint64(len(source)) {
		return false
	}

	switch string(source[tk.Start:tk.End]) {
	case "bool", "char", "complex", "float", "int", "string":
		return true
	default:
		return false
	}
}
