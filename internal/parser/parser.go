package parser

import (
	caramelerrors "github.com/caramelang/caramel/internal/errors"
	"github.com/caramelang/caramel/internal/parser/lexer"
	"github.com/caramelang/caramel/internal/parser/token"
	diagnostics "github.com/rp1s/digreyt"
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
		case token.PUB:
			if self.match_peek(token.L_PAREN) {
				ast.Decls = append(ast.Decls, self.parsePubDeclGroup()...)
			} else if decl := self.parseLetVar(true); decl != nil {
				ast.Decls = append(ast.Decls, &LetDecl{Let: decl})
			}
		case token.LET:
			if decl := self.parseLetVar(true); decl != nil {
				ast.Decls = append(ast.Decls, &LetDecl{Let: decl})
			}
		case token.IDENTIFIER:
			if decl := self.parseQualifiedDecl(); decl != nil {
				ast.Decls = append(ast.Decls, decl)
			}
		case token.ATTR_S:
			if attrs := self.parseAttrs(); attrs != nil {
				ast.Decls = append(ast.Decls, &AttrsDecl{Attrs: attrs})
			}
		case token.COMMA:
			self.consumeUnsupportedComma()
		default:
			if self.curTk.Kind != token.ILLEGAL {
				self.report(caramelerrors.ParserUnexpectedTopLevel(span(self.curTk)))
			}
			self.synchronizeTopLevel()
		}
	}

	if self.Diagnostics.HasFatalErrors() {
		return &ast, self.Diagnostics
	}

	return &ast, nil
}

func (self *Parser) consumeUnsupportedComma() bool {
	if !self.match(token.COMMA) {
		return false
	}
	self.report(caramelerrors.LexerUnexpectedComma(span(self.curTk)))
	self.next()
	return true
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

func (self *Parser) tokenText(tk token.Token) string {
	return string(tk.Literal(&self.Lex.Input))
}

func (self *Parser) synchronizeTopLevel() {
	self.next()
	for self.curTk.Kind != token.EOF {
		switch self.curTk.Kind {
		case token.PACKAGE, token.USE, token.PUB, token.LET, token.ATTR_S, token.IDENTIFIER:
			return
		case token.R_BRACE:
			self.next()
			return
		}
		self.next()
	}
}

func (self *Parser) synchronizeBlock() {
	self.next()
	for self.curTk.Kind != token.EOF {
		switch self.curTk.Kind {
		case token.ATTR_S, token.PUB, token.LET, token.IDENTIFIER, token.R_PAREN, token.R_BRACE, token.ILLEGAL, token.COMMA:
			return
		}
		self.next()
	}
}

func (self *Parser) synchronizeArgs() {
	for self.curTk.Kind != token.EOF {
		if self.match_group(token.G_LITERAL) || self.match(token.R_PAREN, token.ATTR_E, token.ILLEGAL, token.COMMA) {
			return
		}
		self.next()
	}
}

func (self *Parser) synchronizeAttrs() {
	for self.curTk.Kind != token.EOF {
		if self.matchAttrPathSegment() || self.match(token.ATTR_E, token.ILLEGAL, token.COMMA) {
			return
		}
		self.next()
	}
}

func (self *Parser) synchronizeUse() {
	for self.curTk.Kind != token.EOF {
		if self.match(token.STRING, token.RAW_STRING, token.R_PAREN, token.ILLEGAL, token.COMMA) {
			return
		}
		self.next()
	}
}

func (self *Parser) synchronizePubGroup() {
	self.next()
	for self.curTk.Kind != token.EOF {
		switch self.curTk.Kind {
		case token.PUB, token.LET, token.ILLEGAL, token.COMMA, token.R_PAREN:
			return
		}
		self.next()
	}
}
