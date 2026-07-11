package parser

import (
	"github.com/CandyCrafts/candy/internal/parser/lexer"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

type Parser struct {
	Lex *lexer.Lexer

	curTk  token.Token
	peekTk token.Token

	pos uint32
}

func New(input []byte, filename string) *Parser {
	self := &Parser{
		Lex: lexer.New(input, filename),
	}
	self.next().next()
	self.pos = 0
	return self
}

func (self *Parser) next() *Parser {
	if self.curTk.Kind == token.EOF {
		return self
	}

	self.curTk = self.peekTk
	self.pos++

	for {
		self.peekTk = self.Lex.NextToken()
		if !isTrivia(self.peekTk.Kind) {
			break
		}
	}

	return self
}

func isTrivia(kind token.Kind) bool {
	return kind == token.SPACING || kind == token.COMMENT || kind == token.M_COMMENT
}

func (self *Parser) Run() (*AST, error) {
	switch self.curTk.Kind {
	case token.PACKAGE:
		break
	case token.USE:
		break
	case token.IDENTIFIER:
		break
	case token.PUB:
		break
	case token.LET:
		break
	case token.ATTR_S:
		break
	}
	return &AST{}, nil
}

func (self *Parser) match(kinds ...token.Kind) bool {
	for _, k := range kinds {
		if k == self.curTk.Kind {
			return true
		}
	}
	return false
}

func (self *Parser) match_peek(kinds ...token.Kind) bool {
	for _, k := range kinds {
		if k == self.peekTk.Kind {
			return true
		}
	}
	return false
}
