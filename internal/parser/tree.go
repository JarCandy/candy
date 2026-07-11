package parser

type AST struct {
	Root Node
}

type Node interface {
	node()
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

type Package struct {
	Name string
}

func (self Package) node() {}
func (self Package) decl() {}
