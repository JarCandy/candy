package parser

import (
	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

type Package struct {
	Tok  token.Token
	Name token.Token
	Args []*Expr
}

func (Package) node()                {}
func (Package) decl()                {}
func (n Package) Token() token.Token { return n.Tok }

type Use struct {
	Tok      token.Token
	Imports  []UseImport
	AliasMap map[token.Token]token.Token
}

type UseImport struct {
	Link  token.Token
	Alias *token.Token
}

func (Use) node()                {}
func (Use) decl()                {}
func (n Use) Token() token.Token { return n.Tok }

type LetDecl struct {
	*Let
}

func (LetDecl) node() {}
func (LetDecl) decl() {}
func (n LetDecl) Token() token.Token {
	if n.Let == nil {
		return token.Token{}
	}
	return n.Let.Token()
}

func (self *Parser) parsePackage() *Package {
	if !self.match(token.PACKAGE) {
		return nil
	}
	tk := self.curTk
	self.curTk.Kind = token.IDENTIFIER // so that the function *Parser.parseAttr eats everything correctly
	attr := self.parseAttr()
	if attr == nil {
		return nil
	}
	if len(attr.Path) != 1 {
		errTk := tk
		if len(attr.Path) > 1 {
			errTk = attr.Path[1]
		}
		self.report(candyerrors.ParserPackagePath(span(errTk)))
	}

	if self.match(token.END) {
		self.next()
	} else if !self.match(token.EOF) {
		self.report(candyerrors.ParserOptionalSemicolon(span(self.curTk)))
	}
	return &Package{Tok: tk, Name: attr.Path[0], Args: attr.Args}
}

func (self *Parser) parseUse() *Use {
	if !self.match(token.USE) {
		return nil
	}

	tk := self.curTk
	self.next()

	use := &Use{
		Tok:      tk,
		Imports:  make([]UseImport, 0),
		AliasMap: make(map[token.Token]token.Token),
	}

	if !self.match(token.L_PAREN) {
		self.report(candyerrors.ParserUseStart(span(self.curTk)))
		self.synchronizeTopLevel()
		return use
	}
	self.next()

	for !self.match(token.R_PAREN, token.EOF) {
		item, ok := self.parseUseImport()
		if ok {
			use.Imports = append(use.Imports, item)
			if item.Alias != nil {
				use.AliasMap[*item.Alias] = item.Link
			}
		} else {
			self.synchronizeUse()
		}

		if self.match(token.COMMA) {
			self.next()
			continue
		}

		if !self.match(token.R_PAREN, token.EOF) {
			self.report(candyerrors.ParserUseSeparator(span(self.curTk)))
			self.synchronizeUse()
			if self.match(token.COMMA) {
				self.next()
			}
		}
	}

	if self.match(token.R_PAREN) {
		self.next()
	} else {
		self.report(candyerrors.ParserUseClosingParen(span(self.curTk)))
	}

	if self.match(token.END) {
		self.next()
	} else if !self.match(token.EOF) {
		self.report(candyerrors.ParserOptionalSemicolon(span(self.curTk)))
	}

	return use
}

func (self *Parser) parseUseImport() (UseImport, bool) {
	if !self.match(token.STRING, token.RAW_STRING) {
		self.report(candyerrors.ParserUsePath(span(self.curTk)))
		return UseImport{}, false
	}

	item := UseImport{Link: self.curTk}
	self.next()

	if self.match(token.TRANSITION) {
		self.next()
		if !self.match(token.IDENTIFIER) {
			self.report(candyerrors.ParserUseAlias(span(self.curTk)))
			return item, true
		}
		alias := self.curTk
		item.Alias = &alias
		self.next()
	}

	return item, true
}
