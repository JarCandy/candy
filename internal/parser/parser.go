package parser

import (
	diagnostics "github.com/CandyCrafts/candy/internal/digerr"
	digerrors "github.com/CandyCrafts/candy/internal/digerr/errors"
	"github.com/CandyCrafts/candy/internal/parser/lexer"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

type Parser struct {
	Lex         *lexer.Lexer
	Diagnostics *diagnostics.Arena

	curTk  token.Token
	peekTk token.Token

	pos uint32
}

func New(input []byte, filename string) *Parser {
	lex := lexer.New(input, filename)
	self := &Parser{
		Lex:         lex,
		Diagnostics: lex.Diagnostics,
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
	var ast AST

	for self.curTk.Kind != token.EOF {
		switch self.curTk.Kind {
		case token.PACKAGE:
			if decl := self.parsePackage(); decl != nil {
				ast.Decls = append(ast.Decls, decl)
			}
		case token.USE:
			if decl := self.parseUse(); decl != nil {
				ast.Decls = append(ast.Decls, decl)
			}
		case token.IDENTIFIER:
			self.next()
		case token.PUB:
			self.next()
		case token.LET:
			self.next()
		case token.ATTR_S:
			self.next()
		default:
			if self.curTk.Kind != token.ILLEGAL {
				self.report(digerrors.ParserUnexpectedTopLevel(self.curTk))
			}
			self.next()
		}
	}

	if self.Diagnostics.HasErrors() {
		return &ast, self.Diagnostics
	}

	return &ast, nil
}

func (self *Parser) match(kinds ...token.Kind) bool {
	for _, k := range kinds {
		if k == self.curTk.Kind {
			return true
		}
	}
	return false
}

func (self *Parser) match_group(kinds ...token.Kind) bool {
	kind := self.curTk.Kind
	for {
		for _, k := range kinds {
			if k == kind {
				return true
			}
		}

		group := token.Group(kind)
		if group == kind {
			return false
		}
		kind = group
	}
}

func (self *Parser) match_peek(kinds ...token.Kind) bool {
	for _, k := range kinds {
		if k == self.peekTk.Kind {
			return true
		}
	}
	return false
}

// helpers func
func ptr[T any](value T) *T {
	return &value
}
