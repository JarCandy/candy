package parser

import "testing"

func TestParseIdentReturnsPointer(t *testing.T) {
	p := New([]byte("value"), "test.cm")
	expr := p.parseIdent()

	if expr == nil {
		t.Fatal("expected ident pointer, got nil")
	}
	ident, ok := (*expr).(IdentExpr)
	if !ok {
		t.Fatalf("expected IdentExpr, got %T", *expr)
	}
	if string(ident.Name.Literal(&p.Lex.Input)) != "value" {
		t.Fatalf("expected value ident, got %q", ident.Name.Literal(&p.Lex.Input))
	}
}
