package parser

import (
	caramelerrors "github.com/caramelang/caramel/internal/errors"
	"github.com/caramelang/caramel/internal/parser/token"
	"github.com/caramelang/caramel/internal/types"
)

const (
	Lowest = iota
	Sum
	Product
	Power
	Prefix
)

type InvalidExpr struct {
	Tok token.Token
}

func (InvalidExpr) node()                   {}
func (InvalidExpr) expr()                   {}
func (self InvalidExpr) Token() token.Token { return self.Tok }

type IdentExpr struct {
	Tok  token.Token
	Name token.Token
}

func (IdentExpr) node()                   {}
func (IdentExpr) expr()                   {}
func (self IdentExpr) Token() token.Token { return self.Tok }

type LiteralExpr struct {
	Tok   token.Token
	Value token.Token
}

func (LiteralExpr) node()                   {}
func (LiteralExpr) expr()                   {}
func (self LiteralExpr) Token() token.Token { return self.Tok }

type UnaryExpr struct {
	Tok token.Token
	Op  token.Token
	X   *Expr
}

func (UnaryExpr) node()                   {}
func (UnaryExpr) expr()                   {}
func (self UnaryExpr) Token() token.Token { return self.Tok }

type BinaryExpr struct {
	Tok   token.Token
	Left  *Expr
	Op    token.Token
	Right *Expr
}

func (BinaryExpr) node()                   {}
func (BinaryExpr) expr()                   {}
func (self BinaryExpr) Token() token.Token { return self.Tok }

// CompositeExpr represents typed collection and model values such as
// []string{"one", "two"} and User{name: "Caramel"}.
type CompositeExpr struct {
	Tok      token.Token
	Type     *TypeExpr
	Tok_s    token.Token
	Tok_e    token.Token
	Elements []*CompositeElement
}

func (CompositeExpr) node()                   {}
func (CompositeExpr) expr()                   {}
func (self CompositeExpr) Token() token.Token { return self.Tok }

type CompositeElement struct {
	Tok   token.Token
	Name  *token.Token
	Key   *Expr
	Value *Expr
}

func (CompositeElement) node()                   {}
func (self CompositeElement) Token() token.Token { return self.Tok }

func precedence(kind token.Kind) int {
	switch kind {
	case token.ADD, token.SUB:
		return Sum
	case token.MUL, token.DIV, token.MOD:
		return Product
	case token.POW:
		return Power
	default:
		return Lowest
	}
}

func (self *Parser) ParseExpr() *Expr {
	return self.parseExpr(Lowest)
}

func (self *Parser) parseExpr(pre int) *Expr {
	left := self.parsePrefix()

	for {
		op := self.curTk
		if op.Kind == token.EOF {
			break
		}

		pred := precedence(op.Kind)
		if pred <= pre {
			break
		}

		self.next()
		right := self.parseExpr(pred)

		left = ptr[Expr](BinaryExpr{
			Tok:   op,
			Left:  left,
			Op:    op,
			Right: right,
		})
	}

	return left
}

func (self *Parser) parsePrefix() *Expr {
	switch self.curTk.Kind {
	case token.SUB, token.RA:
		op := self.curTk
		self.next()
		x := self.parseExpr(Prefix)
		return ptr[Expr](UnaryExpr{
			Tok: op,
			Op:  op,
			X:   x,
		})

	case token.INTEGER,
		token.FLOATING,
		token.IMAGINARY,
		token.STRING,
		token.RAW_STRING,
		token.CHARACTER,
		token.TRUE,
		token.FALSE:
		return self.parseLiteral()

	case token.IDENTIFIER:
		if self.match_peek(token.L_BRACK) {
			tk := self.curTk
			typ := self.parseType()
			if typ == nil {
				return ptr[Expr](InvalidExpr{Tok: tk})
			}
			return self.parseCompositeBody(typ, tk)
		}
		if self.match_peek(token.L_BRACE) {
			tk := self.curTk
			self.next()
			return self.parseCompositeBody(&TypeExpr{Tok: tk, Path: []token.Token{tk}}, tk)
		}
		if self.match_peek(token.D_COLON, token.DOT, token.L_PAREN) {
			attr := self.parseAttr()
			if attr == nil {
				return ptr[Expr](InvalidExpr{Tok: self.curTk})
			}
			if self.match(token.L_BRACE) && attrCanBeCompositeType(attr) {
				return self.parseCompositeBody(&TypeExpr{Tok: attr.Path[len(attr.Path)-1], Path: attr.Path}, attr.Tok)
			}
			return ptr[Expr](attr)
		}
		return self.parseIdent()

	case token.L_BRACK:
		return self.parseSliceComposite()

	case token.L_PAREN:
		self.next()
		expr := self.parseExpr(Lowest)
		if self.curTk.Kind == token.R_PAREN {
			self.next()
		} else {
			self.report(caramelerrors.ParserMissingClosingParen(span(self.curTk)))
		}
		return expr

	default:
		tk := self.curTk
		if tk.Kind != token.ILLEGAL {
			self.report(caramelerrors.ParserUnexpectedExprToken(span(tk)))
		}
		self.next()
		return ptr[Expr](InvalidExpr{Tok: tk})
	}
}

func (self *Parser) parseSliceComposite() *Expr {
	tk := self.curTk
	typ := &TypeExpr{Tok: tk, Modifiers: make([]TypeModifier, 0), Path: make([]token.Token, 0)}

	for {
		switch {
		case self.match(token.L_BRACK):
			modifier := TypeModifier{Kind: TypeSlice, Tok_s: self.curTk}
			self.next()
			if !self.match(token.R_BRACK) {
				self.report(caramelerrors.ParserTypeSliceClosing(span(self.curTk)))
				return ptr[Expr](InvalidExpr{Tok: tk})
			}
			modifier.Tok_e = self.curTk
			typ.Modifiers = append(typ.Modifiers, modifier)
			self.next()

		case self.match(token.RA):
			typ.Modifiers = append(typ.Modifiers, TypeModifier{
				Kind:  TypePointer,
				Tok_s: self.curTk,
				Tok_e: self.curTk,
			})
			self.next()

		default:
			goto path
		}
	}

path:
	if !self.match(token.IDENTIFIER) {
		self.report(caramelerrors.ParserCompositeType(span(self.curTk)))
		return ptr[Expr](InvalidExpr{Tok: tk})
	}

	for {
		typ.Path = append(typ.Path, self.curTk)
		typ.Tok = self.curTk
		self.next()
		if !self.match(token.D_COLON) {
			break
		}
		self.next()
		if !self.match(token.IDENTIFIER) {
			self.report(caramelerrors.ParserCompositeType(span(self.curTk)))
			return ptr[Expr](InvalidExpr{Tok: tk})
		}
	}

	return self.parseCompositeBody(typ, tk)
}

func (self *Parser) parseCompositeBody(typ *TypeExpr, tk token.Token) *Expr {
	if !self.match(token.L_BRACE) {
		self.report(caramelerrors.ParserCompositeBody(span(self.curTk)))
		return ptr[Expr](InvalidExpr{Tok: tk})
	}

	composite := CompositeExpr{
		Tok:      tk,
		Type:     typ,
		Tok_s:    self.curTk,
		Elements: make([]*CompositeElement, 0),
	}
	self.next()

	for !self.match(token.R_BRACE, token.EOF) {
		if self.match(token.COMMA) {
			self.next()
			continue
		}
		if self.match(token.ILLEGAL) {
			self.next()
			continue
		}

		element := &CompositeElement{Tok: self.curTk}
		first := self.ParseExpr()
		if self.match(token.COLON) {
			if ident, ok := exprIdent(first); ok {
				name := ident.Name
				element.Name = &name
			} else {
				element.Key = first
			}
			self.next()
			if self.match(token.R_BRACE, token.COMMA, token.EOF) {
				self.report(caramelerrors.ParserCompositeValue(span(self.curTk)))
				if self.match(token.COMMA) {
					self.next()
				}
				continue
			}
			element.Value = self.ParseExpr()
		} else {
			element.Value = first
		}
		composite.Elements = append(composite.Elements, element)
		if self.match(token.COMMA) {
			self.next()
		}
	}

	if self.match(token.R_BRACE) {
		composite.Tok_e = self.curTk
		self.next()
	} else {
		self.report(caramelerrors.ParserCompositeClosing(span(self.curTk)))
	}

	return ptr[Expr](composite)
}

func exprIdent(expr *Expr) (IdentExpr, bool) {
	if expr == nil || *expr == nil {
		return IdentExpr{}, false
	}
	switch ident := (*expr).(type) {
	case IdentExpr:
		return ident, true
	case *IdentExpr:
		return *ident, true
	default:
		return IdentExpr{}, false
	}
}

func attrCanBeCompositeType(attr *Attr) bool {
	if attr == nil || len(attr.Path) == 0 || len(attr.Args) != 0 || attr.Value != nil {
		return false
	}
	for _, separator := range attr.Separators {
		if separator.Kind != token.D_COLON {
			return false
		}
	}
	return true
}

func (self *Parser) parseLiteral() *Expr {
	tk := self.curTk
	self.next()
	return ptr[Expr](LiteralExpr{Tok: tk, Value: tk})
}

func (self *Parser) parseIdent() *Expr {
	tk := self.curTk
	self.next()
	return ptr[Expr](IdentExpr{Tok: tk, Name: tk})
}

// Expr Call

type Vaule struct {
	Type       types.Type
	Value      token.Token
	AccessAttr *Attr
}

type Attr struct {
	Tok        token.Token
	Path       []token.Token // db::sqlite() -> [tk:"db", tk:"sqlite"]
	Separators []token.Token // db::sqlite().name -> [tk:"::", tk:"."]
	Args       []*Expr       // Arg && Call db::sqlite(db::std::name())
	Value      *Expr         // lang=custom(...)
}

func (Attr) node()                   {}
func (Attr) expr()                   {}
func (self Attr) Token() token.Token { return self.Tok }

type Arg struct {
	Tok   token.Token
	Name  *token.Token // nil -> ("string")
	Vaule Vaule
}

func (Arg) node()                   {}
func (Arg) expr()                   {}
func (self Arg) Token() token.Token { return self.Tok }

// supports db::sqlite("", conn: "")
// works only with IDENTIFIER
// eats all the tokens, you get a new one in the parser state
func (self *Parser) parseAttr() *Attr {
	if !self.matchAttrPathSegment() {
		self.report(caramelerrors.ParserAttrStart(span(self.curTk)))
		return nil
	}

	attr := &Attr{
		Tok:        self.curTk,
		Path:       make([]token.Token, 0),
		Separators: make([]token.Token, 0),
	}

	for !self.match(token.R_PAREN, token.ATTR_E, token.EOF) {
		if !self.matchAttrPathSegment() {
			self.report(caramelerrors.ParserAttrPathSegment(span(self.curTk)))
			self.synchronizeArgs()
			return nil
		}

		attr.Path = append(attr.Path, self.curTk)
		self.next()

		if self.match(token.L_PAREN) {
			args := self.parseArgs()
			attr.Args = append(attr.Args, argsToExprs(args)...)
		}

		if self.match(token.ASSIGN) {
			value, ok := self.parseAttrAssignmentValue()
			if ok {
				attr.Value = value
			}
			return attr
		}

		if !self.match(token.D_COLON, token.DOT) {
			return attr
		}
		attr.Separators = append(attr.Separators, self.curTk)
		self.next()

		if self.match(token.EOF) {
			self.report(caramelerrors.ParserAttrPathSegment(span(self.curTk)))
			return nil
		}

		if self.match(token.L_PAREN) {
			args := self.parseArgs()
			attr.Args = append(attr.Args, argsToExprs(args)...)
			if !self.match(token.D_COLON, token.DOT) {
				return attr
			}
			attr.Separators = append(attr.Separators, self.curTk)
			self.next()
		}
	}

	return attr
}

func (self *Parser) matchAttrPathSegment() bool {
	return self.match(token.IDENTIFIER, token.PACKAGE, token.USE, token.LET, token.PUB)
}

func (self *Parser) parseAttrAssignmentValue() (*Expr, bool) {
	self.next()
	if self.match(token.IDENTIFIER) && self.match_peek(token.L_PAREN, token.D_COLON, token.DOT) {
		attr := self.parseAttr()
		if attr == nil {
			return nil, false
		}
		return ptr[Expr](attr), true
	}
	if self.match_group(token.G_LITERAL) {
		return self.parseExpr(Lowest), true
	}
	self.report(caramelerrors.ParserArgValue(span(self.curTk)))
	return nil, false
}

func argsToExprs(args []*Arg) []*Expr {
	exprs := make([]*Expr, 0, len(args))
	for _, arg := range args {
		exprs = append(exprs, ptr[Expr](arg))
	}
	return exprs
}

func (self *Parser) parseArgs() []*Arg {
	if !self.match(token.L_PAREN) {
		self.report(caramelerrors.ParserArgsStart(span(self.curTk)))
		return nil
	}
	self.next()

	args := make([]*Arg, 0)

	for !self.match(token.R_PAREN, token.EOF) {
		if self.consumeUnsupportedComma() {
			continue
		}
		if self.match(token.ILLEGAL) {
			self.next()
			continue
		}

		errCount := len(self.Diagnostics.Errors)
		arg := self.parseArg()
		if arg == nil {
			if len(self.Diagnostics.Errors) == errCount {
				self.report(caramelerrors.ParserArg(span(self.curTk)))
			}
			self.synchronizeArgs()
			if self.match(token.ILLEGAL) {
				self.next()
			}
			continue
		}
		args = append(args, arg)

		if self.match(token.ILLEGAL) {
			self.next()
			continue
		}
		if self.match_group(token.G_LITERAL) {
			continue
		}
		if self.consumeUnsupportedComma() {
			continue
		}

		if !self.match(token.R_PAREN, token.EOF) {
			self.report(caramelerrors.ParserArgSeparator(span(self.curTk)))
			self.synchronizeArgs()
			if self.match(token.ILLEGAL) {
				self.next()
			}
		}
	}

	if self.match(token.R_PAREN) {
		self.next()
	} else {
		self.report(caramelerrors.ParserArgsClosingParen(span(self.curTk)))
	}
	return args
}

func (self *Parser) parseArg() *Arg {
	if self.match(token.IDENTIFIER) && self.match_peek(token.COLON) {
		tk := self.curTk
		self.next().next()
		return self.parseArgValue(&tk)
	}

	return self.parseArgValue(nil)
}

func (self *Parser) parseArgValue(name *token.Token) *Arg {
	if self.match(token.IDENTIFIER) && self.match_peek(token.D_COLON, token.DOT, token.L_PAREN) {
		tk := self.curTk
		errCount := len(self.Diagnostics.Errors)
		arg := self.parseAttr()
		if arg == nil {
			if len(self.Diagnostics.Errors) == errCount {
				self.report(caramelerrors.ParserAttrAccess(span(self.curTk)))
			}
			return nil
		}

		return &Arg{
			Tok:  tk,
			Name: name,
			Vaule: Vaule{
				Type:       types.Expr,
				AccessAttr: arg,
			},
		}
	}

	if !self.match_group(token.G_LITERAL) {
		self.report(caramelerrors.ParserArgValue(span(self.curTk)))
		return nil
	}

	tk := self.curTk
	arg := &Arg{
		Tok:  tk,
		Name: name,
		Vaule: Vaule{
			Type:       tk.Kind.TypeFromKind(),
			Value:      tk,
			AccessAttr: nil,
		},
	}
	self.next()
	return arg
}

func (self *Parser) report(err caramelerrors.Error) {
	self.Diagnostics.Add(err)
}
