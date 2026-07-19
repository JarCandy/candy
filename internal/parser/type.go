package parser

import (
	caramelerrors "github.com/caramelang/caramel/internal/errors"
	"github.com/caramelang/caramel/internal/parser/token"
)

type TypeExpr struct {
	Tok       token.Token
	Modifiers []TypeModifier
	Path      []token.Token // last item = Type.Tok
}

type TypeModifierKind uint8

const (
	TypePointer TypeModifierKind = iota + 1
	TypeSlice
)

type TypeModifier struct {
	Kind  TypeModifierKind
	Tok_s token.Token
	Tok_e token.Token
}

func (TypeExpr) node()                   {}
func (TypeExpr) typ()                    {}
func (self TypeExpr) Token() token.Token { return self.Tok }

func (self *Parser) parseType() *TypeExpr {
	modifiers := make([]TypeModifier, 0)
	for {
		switch {
		case self.match(token.MUL):
			modifiers = append(modifiers, TypeModifier{
				Kind:  TypePointer,
				Tok_s: self.curTk,
				Tok_e: self.curTk,
			})
			self.next()

		case self.match(token.L_BRACK):
			modifier := TypeModifier{Kind: TypeSlice, Tok_s: self.curTk}
			self.next()
			if !self.match(token.R_BRACK) {
				self.report(caramelerrors.ParserTypeSliceClosing(span(self.curTk)))
				return nil
			}
			modifier.Tok_e = self.curTk
			modifiers = append(modifiers, modifier)
			self.next()

		default:
			goto path
		}
	}

path:
	if !self.match(token.IDENTIFIER) {
		self.report(caramelerrors.ParserTypeStart(span(self.curTk)))
		return nil
	}

	te := &TypeExpr{
		Tok:       self.curTk,
		Modifiers: modifiers,
		Path:      make([]token.Token, 0),
	}

	for {
		te.Path = append(te.Path, self.curTk)
		self.next()

		if !self.match(token.D_COLON) {
			te.Tok = te.lastPathToken()
			return te
		}
		self.next()

		if !self.match(token.IDENTIFIER) {
			self.report(caramelerrors.ParserTypePathSegment(span(self.curTk)))
			return nil
		}
	}
}

func (self *TypeExpr) lastPathToken() token.Token {
	if self == nil || len(self.Path) == 0 {
		return self.Tok
	}
	return self.Path[len(self.Path)-1]
}
