package parser

import (
	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

type LetStmt struct {
	*Let
}

func (LetStmt) node() {}
func (LetStmt) stmt() {}
func (n LetStmt) Token() token.Token {
	if n.Let == nil {
		return token.Token{}
	}
	return n.Let.Token()
}

func (self *Parser) parseLetStmt() *LetStmt {
	if self.match(token.PUB) {
		self.report(candyerrors.ParserLetStart(span(self.curTk)))
		self.synchronizeTopLevel()
		return nil
	}

	let := self.parseLetVar(true)
	if let == nil {
		return nil
	}
	return &LetStmt{Let: let}
}

type AttrsStmt struct {
	*Attrs
}

func (AttrsStmt) node() {}
func (AttrsStmt) stmt() {}
func (n AttrsStmt) Token() token.Token {
	if n.Attrs == nil {
		return token.Token{}
	}
	return n.Attrs.Token()
}

func (self *Parser) parseAttrsStmt() *AttrsStmt {
	attrs := self.parseAttrs()
	if attrs == nil {
		return nil
	}
	return &AttrsStmt{Attrs: attrs}
}

type MethodStmt struct {
	Tok    token.Token
	Name   token.Token
	Args   []*Arg
	Return Type
	Pub    bool
}

func (MethodStmt) node()                {}
func (MethodStmt) stmt()                {}
func (n MethodStmt) Token() token.Token { return n.Tok }

func (self *Parser) parseBlockStmt() Stmt {
	switch self.curTk.Kind {
	case token.ATTR_S:
		return self.parseAttrsStmt()
	case token.PUB, token.LET, token.IDENTIFIER:
		return self.parseMemberStmt()
	default:
		if self.curTk.Kind != token.ILLEGAL {
			self.report(candyerrors.ParserUnexpectedBlockToken(span(self.curTk)))
		}
		return nil
	}
}

func (self *Parser) parseMemberStmt() Stmt {
	return self.parseMemberStmtWithPub(false)
}

func (self *Parser) parseMemberStmtWithPub(defaultPub bool) Stmt {
	pub := false
	tk := self.curTk
	if defaultPub {
		pub = true
	}

	if self.match(token.PUB) {
		pub = true
		tk = self.curTk
		self.next()
	}

	if self.match(token.LET) {
		tk = self.curTk
		self.next()
	}

	if !self.match(token.IDENTIFIER) {
		self.report(candyerrors.ParserMemberName(span(self.curTk)))
		return nil
	}

	if !pub && tk.Kind != token.LET {
		tk = self.curTk
	}
	name := self.curTk
	self.next()

	if self.match(token.L_PAREN) {
		return self.parseMethodTail(tk, name, pub)
	}

	let := self.parseMemberLetTail(tk, name, pub)
	if let == nil {
		return nil
	}
	return &LetStmt{Let: let}
}

func (self *Parser) parsePubMemberGroup() []Stmt {
	if !self.match(token.PUB) {
		return nil
	}

	stmts := make([]Stmt, 0)
	self.next()

	if !self.match(token.L_PAREN) {
		self.report(candyerrors.ParserPubGroupStart(span(self.curTk)))
		self.synchronizeBlock()
		return stmts
	}
	self.next()

	for !self.match(token.R_PAREN, token.EOF) {
		if self.match(token.COMMA, token.END) {
			self.next()
			continue
		}

		if self.match(token.ATTR_S) {
			if attrs := self.parseAttrsStmt(); attrs != nil {
				stmts = append(stmts, attrs)
			}
		} else if self.match(token.PUB, token.LET, token.IDENTIFIER) {
			if stmt := self.parseMemberStmtWithPub(true); stmt != nil {
				stmts = append(stmts, stmt)
			}
		} else {
			if self.curTk.Kind != token.ILLEGAL {
				self.report(candyerrors.ParserUnexpectedBlockToken(span(self.curTk)))
			}
			self.synchronizeBlock()
		}

		if self.match(token.COMMA, token.END) {
			self.next()
		}
	}

	if self.match(token.R_PAREN) {
		self.next()
	} else {
		self.report(candyerrors.ParserPubGroupClosing(span(self.curTk)))
	}

	if self.match(token.END, token.COMMA) {
		self.next()
	}

	return stmts
}

func (self *Parser) parseMemberLetTail(tk token.Token, name token.Token, pub bool) *Let {
	decl := &Let{
		Tok:  tk,
		Name: name,
		Pub:  pub,
	}

	hasType := false
	hasDefault := false

	if self.match(token.COLON) {
		self.next()
		typ := self.parseType()
		if typ == nil {
			self.synchronizeBlock()
			return decl
		}
		decl.Type = typ
		hasType = true
	}

	if self.match(token.ASSIGN) {
		hasDefault = true
		self.next()
		if self.match(token.END, token.COMMA, token.R_PAREN, token.R_BRACE, token.EOF) {
			self.report(candyerrors.ParserLetValue(span(self.curTk)))
		} else {
			decl.Defualt = self.ParseExpr()
		}
	}

	if !hasType && !hasDefault {
		self.report(candyerrors.ParserLetBody(span(self.curTk)))
	}

	self.consumeMemberBoundary()
	return decl
}

func (self *Parser) parseMethodTail(tk token.Token, name token.Token, pub bool) *MethodStmt {
	method := &MethodStmt{
		Tok:  tk,
		Name: name,
		Pub:  pub,
	}

	method.Args = self.parseArgs()

	if !self.match(token.TRANSITION) {
		self.report(candyerrors.ParserMethodReturn(span(self.curTk)))
		self.consumeMemberBoundary()
		return method
	}
	self.next()

	typ := self.parseType()
	if typ == nil {
		self.synchronizeBlock()
		self.consumeMemberBoundary()
		return method
	}
	method.Return = typ

	self.consumeMemberBoundary()
	return method
}

func (self *Parser) consumeMemberBoundary() {
	if self.match(token.END, token.COMMA) {
		self.next()
		return
	}
	if !self.match(token.R_PAREN, token.R_BRACE, token.EOF) {
		self.report(candyerrors.ParserOptionalSemicolon(span(self.curTk)))
	}
}
