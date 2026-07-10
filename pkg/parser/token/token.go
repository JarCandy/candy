package token

type Kind uint8

const (
	ILLEGAL   Kind = iota
	COMMENT        // comment
	M_COMMENT      // /* comment */
	SPACING        // whitespace
	EOF

	// Groups
	G_NUMBER
	G_STRING
	G_LITERAL
	G_ARITHMETIC

	INTEGER    // 123
	IMAGINARY  // 123i
	FLOATING   // 12.3
	STRING     // "abc"
	RAW_STRING // `abc`
	CHARACTER  // 'a'
	IDENTIFIER // NameVar
	TRUE
	FALSE

	PACKAGE // package
	USE     // use

	PUB // pub
	LET // let

	ASSIGN             // =
	TRANSITION         // ->
	REVERSE_TRANSITION // <-

	SUB // -
	ADD // +
	MUL // *
	DIV // /
	MOD // %
	POW // ^

	ATTR_S // attribute #[ tokens
	ATTR_E // attribute ]

	TEMPLATE_S // &{ tokens
	TEMPLATE_E // }

	REF // &

	L_PAREN // (
	R_PAREN // )
	L_BRACE // {
	R_BRACE // }
	L_BRACK // [
	R_BRACK // ]

	COLON        // :
	DOUBLE_COLON // ::
	END          // ;
	COMMA        // ,
	DOT          // .

)

type Token struct {
	Kind  Kind
	Pos   Position
	Start uint64
	End   uint64
}

type Position struct {
	FileName string
	Line     uint64
	Column   uint64
	Offset   uint64
}

func (tk Token) Literal(source *[]byte) []byte {
	b := (*source)[tk.Start:tk.End]

	switch tk.Kind {
	case STRING, RAW_STRING:
		if len(b) >= 1 && b[0] == '"' {
			if len(b) >= 2 && b[len(b)-1] == '"' {
				return b[1 : len(b)-1]
			}
			return b[1:]
		}

		if len(b) >= 1 && b[0] == '`' {
			if len(b) >= 2 && b[len(b)-1] == '`' {
				return b[1 : len(b)-1]
			}
			return b[1:]
		}

		return b

	case CHARACTER:
		if len(b) >= 2 && b[0] == '\'' && b[len(b)-1] == '\'' {
			return b[1 : len(b)-1]
		}
		return b

	case COMMENT:
		if len(b) >= 2 && b[0] == '/' && b[1] == '/' {
			return b[2:]
		}
		return b

	case M_COMMENT:
		if len(b) >= 4 &&
			b[0] == '/' && b[1] == '*' &&
			b[len(b)-2] == '*' && b[len(b)-1] == '/' {
			return b[2 : len(b)-2]
		}
		return b

	default:
		return b
	}
}

var keywords = map[string]Kind{
	"package": PACKAGE,
	"use":     USE,
	"pub":     PUB,
	"let":     LET,
	"true":    TRUE,
	"false":   FALSE,
}

func SearchKeyword(ident []byte) Kind {
	if kind, ok := keywords[string(ident)]; ok {
		return kind
	}
	return IDENTIFIER
}
