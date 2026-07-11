package test

import (
	"testing"

	"github.com/CandyCrafts/candy/internal/parser"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

func TestParseExprPrecedence(t *testing.T) {
	expr := parser.New([]byte("1 + 2 * 3"), "expr.cm").ParseExpr()

	root, ok := expr.(parser.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if root.Op.Kind != token.ADD {
		t.Fatalf("expected root ADD, got %s", root.Op.Kind)
	}

	right, ok := root.Right.(parser.BinaryExpr)
	if !ok {
		t.Fatalf("expected right BinaryExpr, got %T", root.Right)
	}
	if right.Op.Kind != token.MUL {
		t.Fatalf("expected right MUL, got %s", right.Op.Kind)
	}
}

func TestParseExprPrefix(t *testing.T) {
	expr := parser.New([]byte("-value"), "expr.cm").ParseExpr()

	root, ok := expr.(parser.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}
	if root.Op.Kind != token.SUB {
		t.Fatalf("expected SUB, got %s", root.Op.Kind)
	}
}
