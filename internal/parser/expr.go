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
	X  Expr
}

func (UnaryExpr) node() {}
func (UnaryExpr) expr() {}

type BinaryExpr struct {
	Left  Expr
	Op    token.Token
	Right Expr
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

func exprPtr(expr Expr) *Expr {
	return &expr
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

		left = exprPtr(BinaryExpr{
			Left:  *left,
			Op:    op,
			Right: *right,
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
		return exprPtr(UnaryExpr{
			Op: op,
			X:  *x,
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
		return exprPtr(InvalidExpr{Token: tk})
	}
}

func (self *Parser) parseLiteral() *Expr {
	tk := self.curTk
	self.next()
	return exprPtr(LiteralExpr{Value: tk})
}

func (self *Parser) parseIdent() *Expr {
	tk := self.curTk
	self.next()
	return exprPtr(IdentExpr{Name: tk})
}

// Expr Call

type Vaule struct {
	Type  types.Type
	Value string
}

type Call struct {
	Path []string // db::sqlite()
	Args []Arg
}

func (Call) node() {}
func (Call) expr() {}

type Arg struct {
	Name  *string // nil -> ("string")
	Vaule Vaule
}

func (Arg) node() {}
func (Arg) expr() {}

func (self *Parser) parseCall() *Call {

	return nil
}
