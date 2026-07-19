package parser

import (
	caramelerrors "github.com/caramelang/caramel/internal/errors"
	"github.com/caramelang/caramel/internal/parser/token"
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

func (Let) node()                   {}
func (self Let) Token() token.Token { return self.Tok }

type Attrs struct {
	Tok_s token.Token // ATTR_S
	Tok_e token.Token // ATTR_E

	Attrs []*Attr
	Map   map[string]*Attr
}

func (Attrs) node()                     {}
func (self Attrs) Token() token.Token   { return self.Tok_s }
func (self Attrs) Token_s() token.Token { return self.Tok_s }
func (self Attrs) Token_e() token.Token { return self.Tok_e }

func (self *Parser) parseLetVar(letKw bool) *Let {
	pub := false
	tk := self.curTk

	if self.match(token.PUB) {
		pub = true
		tk = self.curTk
		self.next()
	}

	if self.match(token.LET) {
		tk = self.curTk
		self.next()
	} else if letKw {
		self.report(caramelerrors.ParserLetStart(span(self.curTk)))
		self.synchronizeTopLevel()
		return nil
	}

	if !self.match(token.IDENTIFIER) {
		self.report(caramelerrors.ParserLetName(span(self.curTk)))
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
		if self.match(token.ILLEGAL, token.R_PAREN, token.R_BRACE, token.EOF) {
			self.report(caramelerrors.ParserLetValue(span(self.curTk)))
		} else {
			decl.Defualt = self.ParseExpr()
		}
	}

	if !hasType && !hasDefault {
		self.report(caramelerrors.ParserLetBody(span(self.curTk)))
	}

	if self.match(token.ILLEGAL) {
		self.next()
	}

	return decl
}

func (self *Parser) parseAttrs() *Attrs {
	if !self.match(token.ATTR_S) {
		self.report(caramelerrors.ParserAttrsStart(span(self.curTk)))
		return nil
	}

	attrs := &Attrs{
		Tok_s: self.curTk,
		Attrs: make([]*Attr, 0),
		Map:   make(map[string]*Attr),
	}
	self.next()

	for !self.match(token.ATTR_E, token.EOF) {
		if self.consumeUnsupportedComma() {
			continue
		}
		if self.match(token.ILLEGAL) {
			self.next()
			continue
		}

		attr := self.parseAttr()
		if attr == nil {
			self.synchronizeAttrs()
			if self.match(token.ILLEGAL) {
				self.next()
			}
			continue
		}
		attrs.Attrs = append(attrs.Attrs, attr)
		self.addAttrToMap(attrs, attr)

		if self.consumeUnsupportedComma() {
			continue
		}
		if self.match(token.ILLEGAL) {
			self.next()
			continue
		}
		if self.matchAttrPathSegment() {
			continue
		}

		if !self.match(token.ATTR_E, token.EOF) {
			self.report(caramelerrors.ParserAttrsSeparator(span(self.curTk)))
			self.synchronizeAttrs()
			if self.match(token.ILLEGAL) {
				self.next()
			}
		}
	}

	if self.match(token.ATTR_E) {
		attrs.Tok_e = self.curTk
		self.next()
	} else {
		self.report(caramelerrors.ParserAttrsClosing(span(self.curTk)))
	}

	return attrs

}

func (self *Parser) addAttrToMap(attrs *Attrs, attr *Attr) {
	if attrs == nil || attr == nil || attr.Value == nil || len(attr.Path) == 0 {
		return
	}
	attrs.Map[self.tokenText(attr.Path[0])] = attr
}

// helpers func
func ptr[T any](value T) *T {
	return &value
}

func span(tk token.Token) caramelerrors.Span {
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
