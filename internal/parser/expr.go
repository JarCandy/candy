package parser

import (
	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser/token"
	"github.com/CandyCrafts/candy/internal/types"
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
	case token.SUB:
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
		if self.match_peek(token.D_COLON, token.L_PAREN) {
			attr := self.parseAttr()
			if attr == nil {
				return ptr[Expr](InvalidExpr{Tok: self.curTk})
			}
			return ptr[Expr](attr)
		}
		return self.parseIdent()

	case token.L_PAREN:
		self.next()
		expr := self.parseExpr(Lowest)
		if self.curTk.Kind == token.R_PAREN {
			self.next()
		} else {
			self.report(candyerrors.ParserMissingClosingParen(span(self.curTk)))
		}
		return expr

	default:
		tk := self.curTk
		if tk.Kind != token.ILLEGAL {
			self.report(candyerrors.ParserUnexpectedExprToken(span(tk)))
		}
		self.next()
		return ptr[Expr](InvalidExpr{Tok: tk})
	}
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
	Tok   token.Token
	Path  []token.Token // db::sqlite() -> [tk:"db", tk:"sqlite"]
	Args  []*Expr       // Arg && Call db::sqlite(db::std::name())
	Value *Expr         // lang=custom(...)
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
		self.report(candyerrors.ParserAttrStart(span(self.curTk)))
		return nil
	}

	attr := &Attr{
		Tok:  self.curTk,
		Path: make([]token.Token, 0),
	}

	for !self.match(token.R_PAREN, token.COMMA, token.ATTR_E, token.EOF) {
		if !self.matchAttrPathSegment() {
			self.report(candyerrors.ParserAttrPathSegment(span(self.curTk)))
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

		if !self.match(token.D_COLON) {
			return attr
		}
		self.next()

		if self.match(token.EOF) {
			self.report(candyerrors.ParserAttrPathSegment(span(self.curTk)))
			return nil
		}

		if self.match(token.L_PAREN) {
			args := self.parseArgs()
			attr.Args = append(attr.Args, argsToExprs(args)...)
			if !self.match(token.D_COLON) {
				return attr
			}
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
	if self.match(token.IDENTIFIER) && self.match_peek(token.L_PAREN, token.D_COLON) {
		attr := self.parseAttr()
		if attr == nil {
			return nil, false
		}
		return ptr[Expr](attr), true
	}
	if self.match_group(token.G_LITERAL) {
		return self.parseExpr(Lowest), true
	}
	self.report(candyerrors.ParserArgValue(span(self.curTk)))
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
		self.report(candyerrors.ParserArgsStart(span(self.curTk)))
		return nil
	}
	self.next()

	args := make([]*Arg, 0)

	for !self.match(token.R_PAREN, token.EOF) {
		errCount := len(self.Diagnostics.Errors)
		arg := self.parseArg()
		if arg == nil {
			if len(self.Diagnostics.Errors) == errCount {
				self.report(candyerrors.ParserArg(span(self.curTk)))
			}
			self.synchronizeArgs()
			if self.match(token.COMMA) {
				self.next()
				continue
			}
			continue
		}
		args = append(args, arg)

		if self.match(token.COMMA) {
			self.next()
			continue
		}

		if !self.match(token.R_PAREN, token.EOF) {
			self.report(candyerrors.ParserArgSeparator(span(self.curTk)))
			self.synchronizeArgs()
			if self.match(token.COMMA) {
				self.next()
				continue
			}
		}
	}

	if self.match(token.R_PAREN) {
		self.next()
	} else {
		self.report(candyerrors.ParserArgsClosingParen(span(self.curTk)))
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
	if self.match(token.IDENTIFIER) && self.match_peek(token.D_COLON, token.L_PAREN) {
		tk := self.curTk
		errCount := len(self.Diagnostics.Errors)
		arg := self.parseAttr()
		if arg == nil {
			if len(self.Diagnostics.Errors) == errCount {
				self.report(candyerrors.ParserAttrAccess(span(self.curTk)))
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
		self.report(candyerrors.ParserArgValue(span(self.curTk)))
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

func (self *Parser) report(err candyerrors.Error) {
	self.Diagnostics.Add(err)
}
