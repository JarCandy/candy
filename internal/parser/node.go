package parser

import (
	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

type AST struct {
	Root  Node
	Decls []Decl
}

type Node interface {
	node()
	Token() token.Token
}

type Decl interface {
	Node
	decl()
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmt()
}

//
// high-level functions
//

type Type interface {
	Node
	typ()
}

type Let struct {
	Tok     token.Token
	Name    token.Token
	Type    Type
	Defualt *Expr
	Pub     bool
}

func (Let) node()                {}
func (n Let) Token() token.Token { return n.Tok }

func (self *Parser) parseLetVar() *Let {
	pub := false

	if self.match(token.PUB) {
		pub = true
		self.next()
	}

	if !self.match(token.LET) {
		self.report(candyerrors.ParserLetStart(span(self.curTk)))
		self.synchronizeTopLevel()
		return nil
	}

	tk := self.curTk
	self.next()

	if !self.match(token.IDENTIFIER) {
		self.report(candyerrors.ParserLetName(span(self.curTk)))
		self.synchronizeTopLevel()
		return nil
	}

	decl := &Let{
		Tok:  tk,
		Name: self.curTk,
		Pub:  pub,
	}
	self.next()

	hasType := false
	hasDefault := false

	if self.match(token.COLON) {
		self.next()
		typ := self.parseType()
		if typ == nil {
			self.synchronizeTopLevel()
			return decl
		}
		decl.Type = typ
		hasType = true
	}

	if self.match(token.ASSIGN) {
		hasDefault = true
		self.next()
		if self.match(token.END, token.EOF) {
			self.report(candyerrors.ParserLetValue(span(self.curTk)))
		} else {
			decl.Defualt = self.ParseExpr()
		}
	}

	if !hasType && !hasDefault {
		self.report(candyerrors.ParserLetBody(span(self.curTk)))
	}

	if self.match(token.END) {
		self.next()
	} else if !self.match(token.EOF) {
		self.report(candyerrors.ParserOptionalSemicolon(span(self.curTk)))
	}

	return decl
}

// helpers func
func ptr[T any](value T) *T {
	return &value
}

func span(tk token.Token) candyerrors.Span {
	return candyerrors.Span{
		Start: tk.Start,
		End:   tk.End,
		Pos: candyerrors.Position{
			FileName: tk.Pos.FileName,
			Line:     tk.Pos.Line,
			Column:   tk.Pos.Column,
			Offset:   tk.Pos.Offset,
		},
	}
}
