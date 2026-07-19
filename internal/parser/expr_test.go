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

func TestParsePointerSliceComposite(t *testing.T) {
	p := New([]byte(`&[]&string{"hello", "world"}`), "expr.cm")
	expr := p.ParseExpr()

	if p.Diagnostics.HasFatalErrors() {
		t.Fatalf("unexpected diagnostics: %v", p.Diagnostics)
	}
	pointer, ok := (*expr).(UnaryExpr)
	if !ok || pointer.Op.Kind != token.RA {
		t.Fatalf("expected pointer UnaryExpr, got %T", *expr)
	}
	composite, ok := (*pointer.X).(CompositeExpr)
	if !ok {
		t.Fatalf("expected CompositeExpr, got %T", *pointer.X)
	}
	if len(composite.Type.Modifiers) != 2 || composite.Type.Modifiers[0].Kind != TypeSlice || composite.Type.Modifiers[1].Kind != TypePointer {
		t.Fatalf("unexpected composite type modifiers: %#v", composite.Type.Modifiers)
	}
	if len(composite.Elements) != 2 {
		t.Fatalf("expected two elements, got %d", len(composite.Elements))
	}
}

func TestParseCompositeContainsModel(t *testing.T) {
	p := New([]byte(`[]User{User{name: "first"}, &User{name: "second"}}`), "expr.cm")
	expr := p.ParseExpr()

	if p.Diagnostics.HasFatalErrors() {
		t.Fatalf("unexpected diagnostics: %v", p.Diagnostics)
	}
	composite, ok := (*expr).(CompositeExpr)
	if !ok {
		t.Fatalf("expected CompositeExpr, got %T", *expr)
	}
	if len(composite.Elements) != 2 {
		t.Fatalf("expected two model elements, got %d", len(composite.Elements))
	}
	if _, ok := (*composite.Elements[0].Value).(CompositeExpr); !ok {
		t.Fatalf("expected first model literal, got %T", *composite.Elements[0].Value)
	}
	pointer, ok := (*composite.Elements[1].Value).(UnaryExpr)
	if !ok {
		t.Fatalf("expected pointer model literal, got %T", *composite.Elements[1].Value)
	}
	if _, ok := (*pointer.X).(CompositeExpr); !ok {
		t.Fatalf("expected nested model literal, got %T", *pointer.X)
	}
}

func TestParseMapCompositeWithModelValue(t *testing.T) {
	p := New([]byte(`&map[string]*User{"hello": &User{name: &"world"}}`), "expr.cm")
	expr := p.ParseExpr()

	if p.Diagnostics.HasFatalErrors() {
		t.Fatalf("unexpected diagnostics: %v", p.Diagnostics)
	}
	pointer := (*expr).(UnaryExpr)
	composite, ok := (*pointer.X).(CompositeExpr)
	if !ok {
		t.Fatalf("expected map CompositeExpr, got %T", *pointer.X)
	}
	if composite.Type.Key == nil || composite.Type.Element == nil {
		t.Fatalf("expected map key and element types: %#v", composite.Type)
	}
	if len(composite.Elements) != 1 || composite.Elements[0].Key == nil {
		t.Fatalf("expected keyed map element: %#v", composite.Elements)
	}
}

func TestParseNestedMapComposite(t *testing.T) {
	p := New([]byte(`&map[string]map[string]string{
		"outer": map[string]string{"inner": "value"}
	}`), "expr.cm")
	expr := p.ParseExpr()

	if p.Diagnostics.HasFatalErrors() {
		t.Fatalf("unexpected diagnostics: %v", p.Diagnostics)
	}
	pointer, ok := (*expr).(UnaryExpr)
	if !ok {
		t.Fatalf("expected pointer UnaryExpr, got %T", *expr)
	}
	outer, ok := (*pointer.X).(CompositeExpr)
	if !ok {
		t.Fatalf("expected outer map CompositeExpr, got %T", *pointer.X)
	}
	outerElementType, ok := outer.Type.Element.(*TypeExpr)
	if !ok || outerElementType.Key == nil || outerElementType.Element == nil {
		t.Fatalf("expected nested map element type, got %#v", outer.Type.Element)
	}
	if len(outer.Elements) != 1 || outer.Elements[0].Value == nil {
		t.Fatalf("expected one outer map entry, got %#v", outer.Elements)
	}
	if _, ok := (*outer.Elements[0].Value).(CompositeExpr); !ok {
		t.Fatalf("expected nested map value, got %T", *outer.Elements[0].Value)
	}
}
