package token

import (
	"github.com/CandyCrafts/candy/internal/types"
)

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
	AS      // as

	PUB // pub
	LET // let

	ASSIGN     // =
	TRANSITION // ->
	RRT        // <-

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

	RA // &

	L_PAREN // (
	R_PAREN // )
	L_BRACE // {
	R_BRACE // }
	L_BRACK // [
	R_BRACK // ]

	COLON   // :
	D_COLON // ::
	DOT     // .

)

func kindPtr(k Kind) *Kind {
	return &k
}

func Group(tk Kind) Kind {
	k := kindPtr(tk.group())
	return k.group()
}

func (self *Kind) Group() Kind {
	return Group(*self)
}

func (self *Kind) group() Kind {
	switch *self {
	case INTEGER, IMAGINARY, FLOATING:
		return G_NUMBER
	case STRING, RAW_STRING, CHARACTER:
		return G_STRING
	case SUB, ADD, MUL, DIV, MOD, POW:
		return G_ARITHMETIC
	case G_ARITHMETIC, G_STRING, IDENTIFIER, G_NUMBER, TRUE, FALSE:
		return G_LITERAL
	default:
		return *self
	}
}

func Expand(tk Kind) []Kind {
	switch tk {
	case G_LITERAL:
		return []Kind{
			G_NUMBER,
			G_STRING,
			INTEGER,
			IMAGINARY,
			FLOATING,
			STRING,
			TRUE,
			FALSE,
			RAW_STRING,
			CHARACTER,
			IDENTIFIER,
		}
	case G_NUMBER:
		return []Kind{
			INTEGER,
			IMAGINARY,
			FLOATING,
		}
	case G_STRING:
		return []Kind{
			STRING,
			RAW_STRING,
			CHARACTER,
		}
	case G_ARITHMETIC:
		return []Kind{
			SUB,
			ADD,
			MUL,
			DIV,
			MOD,
			POW,
		}
	default:
		return nil
	}
}

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

func (self Token) Literal(source *[]byte) []byte {
	b := (*source)[self.Start:self.End]

	switch self.Kind {
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
	"as":      AS,
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

func (self Kind) String() string {
	switch self {
	case ILLEGAL:
		return "ILLEGAL"
	case COMMENT:
		return "COMMENT"
	case M_COMMENT:
		return "M_COMMENT"
	case SPACING:
		return "SPACING"
	case EOF:
		return "EOF"
	case INTEGER:
		return "INTEGER"
	case IMAGINARY:
		return "IMAGINARY"
	case FLOATING:
		return "FLOATING"
	case STRING:
		return "STRING"
	case RAW_STRING:
		return "RAW_STRING"
	case CHARACTER:
		return "CHARACTER"
	case IDENTIFIER:
		return "IDENTIFIER"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case PACKAGE:
		return "PACKAGE"
	case USE:
		return "USE"
	case AS:
		return "AS"
	case PUB:
		return "PUB"
	case LET:
		return "LET"
	case ASSIGN:
		return "ASSIGN"
	case TRANSITION:
		return "TRANSITION"
	case RRT:
		return "RRT"
	case SUB:
		return "SUB"
	case ADD:
		return "ADD"
	case MUL:
		return "MUL"
	case DIV:
		return "DIV"
	case MOD:
		return "MOD"
	case POW:
		return "POW"
	case ATTR_S:
		return "ATTR_S"
	case ATTR_E:
		return "ATTR_E"
	case TEMPLATE_S:
		return "TEMPLATE_S"
	case TEMPLATE_E:
		return "TEMPLATE_E"
	case RA:
		return "RA"
	case L_PAREN:
		return "L_PAREN"
	case R_PAREN:
		return "R_PAREN"
	case L_BRACE:
		return "L_BRACE"
	case R_BRACE:
		return "R_BRACE"
	case L_BRACK:
		return "L_BRACK"
	case R_BRACK:
		return "R_BRACK"
	case COLON:
		return "COLON"
	case D_COLON:
		return "D_COLON"
	case DOT:
		return "DOT"
	default:
		return "UNKNOWN"
	}
}

func (self Kind) TypeFromKind() types.Type {
	switch self {
	case TRUE, FALSE:
		return types.Bool
	case CHARACTER:
		return types.Char
	case INTEGER:
		return types.Int
	case FLOATING:
		return types.Float
	case IMAGINARY:
		return types.Complex
	case STRING, RAW_STRING:
		return types.String
	case IDENTIFIER:
		return types.Decl
	default:
		return types.Null
	}
}
