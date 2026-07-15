package parser

import (
	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

type TypeExpr struct {
	Tok  token.Token
	Path []token.Token // last item = Type.Tok
}

func (TypeExpr) node()                {}
func (TypeExpr) typ()                 {}
func (n TypeExpr) Token() token.Token { return n.Tok }

func (self *Parser) parseType() *TypeExpr {
	if !self.match(token.IDENTIFIER) {
		self.report(candyerrors.ParserTypeStart(span(self.curTk)))
		return nil
	}

	te := &TypeExpr{
		Tok:  self.curTk,
		Path: make([]token.Token, 0),
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
			self.report(candyerrors.ParserTypePathSegment(span(self.curTk)))
			return nil
		}
	}
}

func (te *TypeExpr) lastPathToken() token.Token {
	if te == nil || len(te.Path) == 0 {
		return te.Tok
	}
	return te.Path[len(te.Path)-1]
}
