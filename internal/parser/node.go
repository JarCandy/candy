package parser

import "github.com/CandyCrafts/candy/internal/parser/token"

type AST struct {
	Root  Node
	Decls []Decl
}

type Node interface {
	node()
	Token() token.Token
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
