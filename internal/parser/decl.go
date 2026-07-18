package parser

import (
	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

type Package struct {
	Tok  token.Token
	Args []*Expr
}

func (Package) node()                   {}
func (Package) decl()                   {}
func (self Package) Token() token.Token { return self.Tok }

type Use struct {
	Tok      token.Token
	Imports  []UseImport
	AliasMap map[token.Token]token.Token
}

type UseImport struct {
	Link  token.Token
	Alias *token.Token
}

func (Use) node()                   {}
func (Use) decl()                   {}
func (self Use) Token() token.Token { return self.Tok }

type LetDecl struct {
	*Let
}

func (LetDecl) node() {}
func (LetDecl) decl() {}
func (self LetDecl) Token() token.Token {
	if self.Let == nil {
		return token.Token{}
	}
	return self.Let.Token()
}

type AttrsDecl struct {
	*Attrs
}

func (AttrsDecl) node() {}
func (AttrsDecl) decl() {}
func (self AttrsDecl) Token() token.Token {
	if self.Attrs == nil {
		return token.Token{}
	}
	return self.Attrs.Token()
}

type QualifiedDecl struct {
	Tok  token.Token
	Path []token.Token
	Name token.Token
	Body []Stmt
}

func (QualifiedDecl) node()                   {}
func (QualifiedDecl) decl()                   {}
func (self QualifiedDecl) Token() token.Token { return self.Tok }

func (self *Parser) parsePackage() *Package {
	if !self.match(token.PACKAGE) {
		return nil
	}
	tk := self.curTk
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

	return &Package{Tok: tk, Args: attr.Args}
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
		if self.match(token.ILLEGAL) {
			self.next()
			continue
		}

		item, ok := self.parseUseImport()
		if ok {
			use.Imports = append(use.Imports, item)
			if item.Alias != nil {
				use.AliasMap[*item.Alias] = item.Link
			}
		} else {
			self.synchronizeUse()
		}

		if self.match(token.STRING, token.RAW_STRING) {
			continue
		}
		if self.match(token.ILLEGAL) {
			continue
		}

		if !self.match(token.R_PAREN, token.EOF) {
			self.report(candyerrors.ParserUseSeparator(span(self.curTk)))
			self.synchronizeUse()
		}
	}

	if self.match(token.R_PAREN) {
		self.next()
	} else {
		self.report(candyerrors.ParserUseClosingParen(span(self.curTk)))
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

	if self.match(token.AS, token.TRANSITION) {
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

func (self *Parser) parsePubDeclGroup() []Decl {
	if !self.match(token.PUB) {
		return nil
	}

	decls := make([]Decl, 0)
	self.next()

	if !self.match(token.L_PAREN) {
		self.report(candyerrors.ParserPubGroupStart(span(self.curTk)))
		self.synchronizeTopLevel()
		return decls
	}
	self.next()

	for !self.match(token.R_PAREN, token.EOF) {
		if self.match(token.ILLEGAL) {
			self.next()
			continue
		}

		if !self.match(token.PUB, token.LET) {
			self.report(candyerrors.ParserLetStart(span(self.curTk)))
			self.synchronizePubGroup()
			continue
		}

		let := self.parseLetVar(true)
		if let != nil {
			let.Pub = true
			decls = append(decls, &LetDecl{Let: let})
		}

	}

	if self.match(token.R_PAREN) {
		self.next()
	} else {
		self.report(candyerrors.ParserPubGroupClosing(span(self.curTk)))
	}

	return decls
}

func (self *Parser) parseQualifiedDecl() Decl {
	path, ok := self.parseQualifiedDeclPath()
	if !ok {
		return nil
	}

	if !self.match(token.IDENTIFIER) {
		self.report(candyerrors.ParserDeclName(span(self.curTk)))
		self.synchronizeTopLevel()
		return nil
	}
	name := self.curTk
	self.next()

	body := self.parseDeclBody()
	return &QualifiedDecl{Tok: path[0], Path: path, Name: name, Body: body}
}

func (self *Parser) parseQualifiedDeclPath() ([]token.Token, bool) {
	if !self.match(token.IDENTIFIER) {
		self.report(candyerrors.ParserDeclKind(span(self.curTk)))
		self.synchronizeTopLevel()
		return nil, false
	}

	path := []token.Token{self.curTk}
	self.next()

	for self.match(token.D_COLON) {
		self.next()
		if !self.match(token.IDENTIFIER) {
			self.report(candyerrors.ParserDeclKind(span(self.curTk)))
			self.synchronizeTopLevel()
			return nil, false
		}
		path = append(path, self.curTk)
		self.next()
	}

	return path, true
}

func (self *Parser) parseDeclBody() []Stmt {
	if !self.match(token.L_BRACE) {
		self.report(candyerrors.ParserDeclBodyStart(span(self.curTk)))
		self.synchronizeTopLevel()
		return nil
	}
	self.next()

	body := make([]Stmt, 0)
	for !self.match(token.R_BRACE, token.EOF) {
		if self.match(token.ILLEGAL) {
			self.next()
			continue
		}

		if self.match(token.PUB) && self.match_peek(token.L_PAREN) {
			body = append(body, self.parsePubMemberGroup()...)
			continue
		}

		stmt := self.parseBlockStmt()
		if stmt == nil {
			self.synchronizeBlock()
			continue
		}
		body = append(body, stmt)
	}

	if self.match(token.R_BRACE) {
		self.next()
	} else {
		self.report(candyerrors.ParserDeclBodyClosing(span(self.curTk)))
	}

	return body
}
