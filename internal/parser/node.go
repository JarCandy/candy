package parser

type AST struct {
	Root  Node
	Decls []Decl
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
