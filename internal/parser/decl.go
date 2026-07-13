package parser

import "github.com/CandyCrafts/candy/internal/parser/token"

type Package struct {
	Name string
}

func (Package) node() {}
func (Package) decl() {}

type Use struct {
	Path string
}

func (Use) node() {}
func (Use) decl() {}

func (self *Parser) parsePackage() *Package {
	if !self.match(token.PACKAGE) {
		return nil
	}

	self.next()
	return &Package{}
}

func (self *Parser) parseUse() *Use {
	if !self.match(token.USE) {
		return nil
	}

	self.next()
	return &Use{}
}
