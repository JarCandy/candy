package parser

import (
	"testing"

	"github.com/caramelang/caramel/internal/parser/token"
)

func TestParseExprPrecedence(t *testing.T) {
	expr := New([]byte("1 + 2 * 3"), "expr.cm").ParseExpr()

	root, ok := (*expr).(BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if root.Op.Kind != token.ADD {
		t.Fatalf("expected root ADD, got %s", root.Op.Kind)
	}

	right, ok := (*root.Right).(BinaryExpr)
	if !ok {
		t.Fatalf("expected right BinaryExpr, got %T", root.Right)
	}
	if right.Op.Kind != token.MUL {
		t.Fatalf("expected right MUL, got %s", right.Op.Kind)
	}
}

func TestParseExprPrefix(t *testing.T) {
	expr := New([]byte("-value"), "expr.cm").ParseExpr()

	root, ok := (*expr).(UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}
	if root.Op.Kind != token.SUB {
		t.Fatalf("expected SUB, got %s", root.Op.Kind)
	}
}

func TestParseExprReportsMissingClosingParen(t *testing.T) {
	p := New([]byte("(value"), "expr.cm")
	_ = p.ParseExpr()

	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].CodeName != "ParsingError" {
		t.Fatalf("expected ParsingError, got %s", p.Diagnostics.Errors[0].CodeName)
	}
}
