package lexer

import (
	"unicode/utf8"

	caramelerrors "github.com/caramelang/caramel/internal/errors"
	. "github.com/caramelang/caramel/internal/parser/token"
	diagnostics "github.com/rp1s/digreyt"
)

type Lexer struct {
	Input       []byte
	fileName    string
	Diagnostics *diagnostics.Arena

	pos    int
	curPos int
	rn     rune
	rnSize int

	line int
	col  int

	savePos  int
	saveLine int
	saveCol  int
	saveRn   rune
	saveSize int

	tokPos  int
	tokLine int
	tokCol  int

	interpStack    []int
	attrDepth      int
	inStringResume bool
}

func New(input []byte, fileName string) *Lexer {
	if input == nil {
		input = input[:0]
	}
	self := &Lexer{
		Input:       input,
		fileName:    fileName,
		Diagnostics: diagnostics.New(string(input)),
		line:        1,
		col:         0,
	}
	self.advance()
	return self
}

func (self *Lexer) advance() {
	if self.curPos >= len(self.Input) {
		self.pos = self.curPos
		self.rn = 0
		self.rnSize = 0
		return
	}

	self.pos = self.curPos

	if b := self.Input[self.curPos]; b < utf8.RuneSelf {
		self.rn = rune(b)
		self.rnSize = 1
	} else {
		r, size := utf8.DecodeRune(self.Input[self.curPos:])
		self.rn = r
		self.rnSize = size
	}

	self.curPos += self.rnSize

	if self.rn == '\n' {
		self.line++
		self.col = 0
	} else {
		self.col++
	}
}

func (self *Lexer) peek() rune {
	if self.curPos >= len(self.Input) {
		return 0
	}
	if b := self.Input[self.curPos]; b < utf8.RuneSelf {
		return rune(b)
	}
	r, _ := utf8.DecodeRune(self.Input[self.curPos:])
	return r
}

// снимки состояния

func (self *Lexer) freeze() {
	self.savePos = self.pos
	self.saveLine = self.line
	self.saveCol = self.col
	self.saveRn = self.rn
	self.saveSize = self.rnSize
}

func (self *Lexer) unfreeze() {
	self.pos = self.savePos
	self.curPos = self.savePos + self.saveSize
	self.line = self.saveLine
	self.col = self.saveCol
	self.rn = self.saveRn
	self.rnSize = self.saveSize
}

func (self *Lexer) tok(kind Kind) Token {
	return Token{
		Kind: kind,
		Pos: Position{
			FileName: self.fileName,
			Line:     uint64(self.tokLine),
			Column:   uint64(self.tokCol),
			Offset:   uint64(self.tokPos),
		},
		Start: uint64(self.tokPos),
		End:   uint64(self.pos),
	}
}

func (self *Lexer) NextToken() Token {
	self.tokPos = self.pos
	if isSpace(self.rn) {
		for isSpace(self.rn) {
			self.advance()
		}
		return self.tok(SPACING)
	}

	self.tokPos = self.pos
	self.tokLine = self.line
	self.tokCol = self.col

	if self.rn == 0 {
		return self.tok(EOF)
	}

	if self.inStringResume {
		self.inStringResume = false
		self.tokPos, self.tokLine, self.tokCol = self.pos, self.line, self.col
		return self.readString()
	}

	ch := self.rn
	self.advance()

	switch ch {
	case '/':
		switch self.rn {
		case '/':
			return self.lineComment()
		case '*':
			return self.multiLineComment()
		}
		return self.tok(DIV)

	case '.':
		return self.tok(DOT)

	case '<':
		if self.rn == '-' {
			self.advance()
			return self.tok(RRT)
		}
		tk := self.tok(ILLEGAL)
		self.report(caramelerrors.LexerUnexpectedLess(span(tk)))
		return tk

	case '-':
		if self.rn == '>' {
			self.advance()
			return self.tok(TRANSITION)
		}
		return self.tok(SUB)

	case '+':
		return self.tok(ADD)

	case '*':
		return self.tok(MUL)

	case '%':
		return self.tok(MOD)

	case '^':
		return self.tok(POW)

	case '&':
		if self.rn == '{' {
			self.advance()
			if n := len(self.interpStack); n > 0 {
				self.interpStack[n-1]++
			}
			return self.tok(L_BRACE)
		}
		return self.tok(RA)

	case '#':
		if self.rn == '[' {
			self.advance()
			self.attrDepth++
			return self.tok(ATTR_S)
		}
		tk := self.tok(ILLEGAL)
		self.report(caramelerrors.LexerUnexpectedSharp(span(tk)))
		return tk

	case '=':
		return self.tok(ASSIGN)

	case ':':
		if self.rn == ':' {
			self.advance()
			return self.tok(D_COLON)
		}
		return self.tok(COLON)
	case '(':
		return self.tok(L_PAREN)
	case ')':
		return self.tok(R_PAREN)
	case '{':
		if n := len(self.interpStack); n > 0 {
			self.interpStack[n-1]++
		}
		return self.tok(L_BRACE)
	case '}':
		if n := len(self.interpStack); n > 0 {
			self.interpStack[n-1]--
			if self.interpStack[n-1] == 0 {
				self.interpStack = self.interpStack[:n-1]
				self.inStringResume = true
				return self.tok(TEMPLATE_E)
			}
		}
		return self.tok(R_BRACE)
	case '[':
		if self.attrDepth > 0 {
			self.report(caramelerrors.LexerNestedAttribute(span(self.tok(ILLEGAL))))
			self.attrDepth++
		}
		return self.tok(L_BRACK)
	case ']':
		if self.attrDepth > 0 {
			self.attrDepth--
			if self.attrDepth == 0 {
				return self.tok(ATTR_E)
			}
		}
		return self.tok(R_BRACK)
	case ';':
		tk := self.tok(ILLEGAL)
		self.report(caramelerrors.LexerUnexpectedSemicolon(span(tk)))
		return tk
	case ',':
		tk := self.tok(ILLEGAL)
		self.report(caramelerrors.LexerUnexpectedComma(span(tk)))
		return tk

	case '"':
		return self.readString()
	case '`':
		return self.readRawString()
	case '\'':
		return self.readChar()

	default:
		if ch >= '0' && ch <= '9' {
			return self.readNumber(ch)
		}

		if isIdentStart(ch) {
			return self.readIdent()
		}
		tk := self.tok(ILLEGAL)
		self.report(caramelerrors.LexerUnknownCharacter(span(tk)))
		return tk
	}
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return r > 0x7F && isSpaceUnicode(r)
}

func isSpaceUnicode(r rune) bool {
	switch r {
	case 0x00A0,
		0x1680,
		0x2000, 0x2001, 0x2002, 0x2003,
		0x2004, 0x2005, 0x2006, 0x2007,
		0x2008, 0x2009, 0x200A,
		0x2028, 0x2029,
		0x202F, 0x205F,
		0x3000,
		0xFEFF:
		return true
	}
	return false
}

func isIdentStart(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		r > 0x7F && isLetterUnicode(r)
}

func isIdentContinue(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r > 0x7F && (isLetterUnicode(r) || isDigitUnicode(r))
}

func isLetterUnicode(r rune) bool {
	return r >= 0xAA
}

func isDigitUnicode(r rune) bool {
	return r >= 0x0660
}

func (self *Lexer) lineComment() Token {
	self.advance() // '/'
	for self.rn != '\n' && self.rn != 0 {
		self.advance()
	}
	return self.tok(COMMENT)
}

func (self *Lexer) multiLineComment() Token {
	self.advance() // '*'
	self.freeze()
	for {
		if self.rn == 0 {
			tk := self.tok(ILLEGAL)
			self.report(caramelerrors.LexerUnterminatedMultilineComment(span(tk)))
			self.unfreeze()
			self.stabilize()
			return tk
		}
		if self.rn == '*' && self.peek() == '/' {
			self.advance()
			self.advance()
			break
		}
		self.advance()
	}
	return self.tok(M_COMMENT)
}

func (self *Lexer) readString() Token {
	self.freeze()

	for self.rn != '"' && self.rn != 0 {
		if self.rn == '\\' {
			self.advance()
			if self.rn != 0 {
				self.advance()
			}
			continue
		}

		if self.rn == '$' && self.peek() == '{' {
			self.interpStack = append(self.interpStack, 0)
			return self.tok(STRING)
		}

		self.advance()
	}

	if self.rn == 0 {
		tk := self.tok(ILLEGAL)
		self.report(caramelerrors.LexerUnterminatedString(span(tk)))
		self.unfreeze()
		self.stabilize()
		return tk
	}

	self.advance() // '"'
	return self.tok(STRING)
}

func (self *Lexer) readRawString() Token {
	self.freeze()
	for self.rn != '`' && self.rn != 0 {
		self.advance()
	}
	if self.rn == 0 {
		tk := self.tok(ILLEGAL)
		self.report(caramelerrors.LexerUnterminatedRawString(span(tk)))
		self.unfreeze()
		self.stabilize()
		return tk
	}
	self.advance()
	return self.tok(RAW_STRING)
}

func (self *Lexer) readChar() Token {
	if self.rn == '\\' {
		self.advance()
		self.advance()
	} else if self.rn != '\'' && self.rn != 0 {
		self.advance()
	}
	if self.rn != '\'' {
		tk := self.tok(ILLEGAL)
		self.report(caramelerrors.LexerInvalidCharacterLiteral(span(tk)))
		return tk
	}
	self.advance()
	return self.tok(CHARACTER)
}

func (self *Lexer) readIdent() Token {
	for isIdentContinue(self.rn) {
		self.advance()
	}
	lit := self.Input[self.tokPos:self.pos]
	return self.tok(SearchKeyword(lit))
}

func (self *Lexer) readNumber(first rune) Token {
	_ = first
	isFloat := false
	isIdent := false

	for {
		ch := self.rn
		if ch >= '0' && ch <= '9' {
			self.advance()
			continue
		}
		if ch == '.' {
			if isIdent || isFloat {
				break
			}
			next := self.peek()
			if next < '0' || next > '9' {
				break
			}
			isFloat = true
			self.advance()
			continue
		}
		if isIdentContinue(ch) {
			isIdent = true
			self.advance()
			continue
		}
		break
	}

	lit := self.Input[self.tokPos:self.pos]

	if isIdent {
		n := len(lit)
		if n >= 2 && lit[n-1] == 'i' {
			onlyDigits := true
			for _, b := range lit[:n-1] {
				if (b < '0' || b > '9') && b != '.' {
					onlyDigits = false
					break
				}
			}
			if onlyDigits {
				return self.tok(IMAGINARY)
			}
		}
		return self.tok(IDENTIFIER)
	}

	if isFloat {
		return self.tok(FLOATING)
	}
	return self.tok(INTEGER)
}

func (self *Lexer) report(err caramelerrors.Error) {
	if self.Diagnostics == nil {
		return
	}

	self.Diagnostics.Add(err)
}

func span(tk Token) caramelerrors.Span {
	return caramelerrors.Span{
		Start: tk.Start,
		End:   tk.End,
		Pos: caramelerrors.Position{
			FileName: tk.Pos.FileName,
			Line:     tk.Pos.Line,
			Column:   tk.Pos.Column,
			Offset:   tk.Pos.Offset,
		},
	}
}

func (self *Lexer) stabilize() {
	k := map[Kind]bool{
		PACKAGE: true, USE: true, PUB: true, LET: true,
	}
	for {
		self.freeze()
		tk := self.NextToken()
		switch {
		case tk.Kind == EOF:
			return
		case tk.Kind == SPACING || tk.Kind == COMMENT || tk.Kind == M_COMMENT:
			continue
		case k[tk.Kind], tk.Kind == R_BRACE:
			self.unfreeze()
			return
		}
	}
}
