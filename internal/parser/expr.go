package parser

import (
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
	Token token.Token
}

func (InvalidExpr) node() {}
func (InvalidExpr) expr() {}

type IdentExpr struct {
	Name token.Token
}

func (IdentExpr) node() {}
func (IdentExpr) expr() {}

type LiteralExpr struct {
	Value token.Token
}

func (LiteralExpr) node() {}
func (LiteralExpr) expr() {}

type UnaryExpr struct {
	Op token.Token
	X  *Expr
}

func (UnaryExpr) node() {}
func (UnaryExpr) expr() {}

type BinaryExpr struct {
	Left  *Expr
	Op    token.Token
	Right *Expr
}

func (BinaryExpr) node() {}
func (BinaryExpr) expr() {}

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
			Op: op,
			X:  x,
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
		return self.parseIdent()

	case token.L_PAREN:
		self.next()
		expr := self.parseExpr(Lowest)
		if self.curTk.Kind == token.R_PAREN {
			self.next()
		} else {
			// TODO: report an error here for a missing closing ')'.
		}
		return expr

	default:
		tk := self.curTk
		// TODO: report an error here for an unexpected expression token.
		self.next()
		return ptr[Expr](InvalidExpr{Token: tk})
	}
}

func (self *Parser) parseLiteral() *Expr {
	tk := self.curTk
	self.next()
	return ptr[Expr](LiteralExpr{Value: tk})
}

func (self *Parser) parseIdent() *Expr {
	tk := self.curTk
	self.next()
	return ptr[Expr](IdentExpr{Name: tk})
}

// Expr Call

type Vaule struct {
	Type       types.Type
	Value      string // "" -> AccessAttr
	AccessAttr *AccessAttr
}

type AccessAttr struct {
	Path []string // db::sqlite()
	Args []*Expr  // Arg && Call db::sqlite(db::std::name())
}

func (AccessAttr) node() {}
func (AccessAttr) expr() {}

type Arg struct {
	Name  *string // nil -> ("string")
	Vaule Vaule
}

func (Arg) node() {}
func (Arg) expr() {}

func (self *Parser) parseAccessAttr() *AccessAttr {
	if !self.match(token.IDENTIFIER) {
		panic("The *Parser.parseAccessAttr function was used incorrectly.")
	}

	for !self.match(token.L_PAREN, token.EOF) {
		if !self.match(token.IDENTIFIER) {
			// todo error
			return nil
		}
		aa := &AccessAttr{
			Path: make([]string, 0),
		}
		aa.Path = append(aa.Path, string(self.curTk.Literal(&self.Lex.Input)))
		self.next()

		if !self.match(token.D_COLON) {
			// todo error
			return nil
		}
		self.next()

		args := self.parseArgs()
		aa.Args = argsToExprs(args)

	}

	return nil
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
		panic("The *Parser.parseArgs function was used incorrectly.")
	}
	self.next()

	isComma := false
	args := make([]*Arg, 0)

	for !self.match(token.R_PAREN, token.EOF) {
		if !isComma {
			// todo error
			return nil
		} else {
			isComma = false
		}

		if self.match(token.IDENTIFIER) && self.match_peek(token.COLON) {
			tk := self.curTk
			self.next().next()

			if !self.match_group(token.G_LITERAL) {
				arg := self.parseAccessAttr()
				if arg == nil {
					// todo error
					return nil
				}

				args = append(args, &Arg{
					Name: ptr(string(tk.Literal(&self.Lex.Input))),
					Vaule: Vaule{
						Type:       types.Expr,
						Value:      "",
						AccessAttr: arg,
					},
				})
			}
			args = append(args, &Arg{
				Name: ptr(string(tk.Literal(&self.Lex.Input))),
				Vaule: Vaule{
					Type:       self.curTk.Kind.TypeFromKind(),
					Value:      string(self.curTk.Literal(&self.Lex.Input)),
					AccessAttr: nil,
				},
			})
		} else if self.match(token.IDENTIFIER) && !self.match_peek(token.COLON) {
			if !self.match_group(token.G_LITERAL) {
				arg := self.parseAccessAttr()
				if arg == nil {
					// todo error
					return nil
				}

				args = append(args, &Arg{
					Name: nil,
					Vaule: Vaule{
						Type:       types.Expr,
						Value:      "",
						AccessAttr: arg,
					},
				})
			}
			args = append(args, &Arg{
				Name: ptr(string(self.curTk.Literal(&self.Lex.Input))),
				Vaule: Vaule{
					Type:       self.curTk.Kind.TypeFromKind(),
					Value:      string(self.curTk.Literal(&self.Lex.Input)),
					AccessAttr: nil,
				},
			})
		}

		if self.match(token.COMMA) {
			isComma = true
			self.next()
			continue
		}
	}

	return args
}
