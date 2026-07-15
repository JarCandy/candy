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
