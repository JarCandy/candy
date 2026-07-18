package parser

import (
	"testing"

	"github.com/CandyCrafts/candy/internal/parser/token"
)

func TestParseTypePathStopsBeforeAssign(t *testing.T) {
	p := New([]byte(`pkg::User = value`), "test.cm")
	typ := p.parseType()

	if typ == nil {
		t.Fatal("expected type path, got nil")
	}
	if got := tokenLiterals(p, typ.Path); !equalStrings(got, []string{"pkg", "User"}) {
		t.Fatalf("expected path [pkg User], got %#v", got)
	}
	if literal(p, typ.Token()) != "User" {
		t.Fatalf("expected type token User, got %q", literal(p, typ.Token()))
	}
	if p.curTk.Kind != token.ASSIGN {
		t.Fatalf("expected parser to stop at assign, got %s", p.curTk.Kind)
	}
}

func TestParseTypeWithPointerAndSliceModifiers(t *testing.T) {
	p := New([]byte(`*[]*string = "none"`), "test.cm")
	typ := p.parseType()

	if typ == nil {
		t.Fatal("expected modified type, got nil")
	}
	if got := tokenLiterals(p, typ.Path); !equalStrings(got, []string{"string"}) {
		t.Fatalf("expected path [string], got %#v", got)
	}
	if len(typ.Modifiers) != 3 {
		t.Fatalf("expected 3 type modifiers, got %d", len(typ.Modifiers))
	}
	if typ.Modifiers[0].Kind != TypePointer || typ.Modifiers[1].Kind != TypeSlice || typ.Modifiers[2].Kind != TypePointer {
		t.Fatalf("expected pointer, slice, pointer modifiers, got %#v", typ.Modifiers)
	}
	if literal(p, typ.Modifiers[0].Tok_s) != "*" || literal(p, typ.Modifiers[1].Tok_s) != "[" || literal(p, typ.Modifiers[1].Tok_e) != "]" || literal(p, typ.Modifiers[2].Tok_s) != "*" {
		t.Fatalf("expected modifier tokens for *([])*, got %#v", typ.Modifiers)
	}
	if p.curTk.Kind != token.ASSIGN {
		t.Fatalf("expected parser to stop at assign, got %s", p.curTk.Kind)
	}
}

func TestParseTypeReportsUnclosedSliceModifier(t *testing.T) {
	p := New([]byte(`[*string`), "test.cm")
	typ := p.parseType()

	if typ != nil {
		t.Fatalf("expected nil type, got %#v", typ)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected closing bracket for slice type" {
		t.Fatalf("expected slice closing diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParseTypePathStopsBeforeCallParen(t *testing.T) {
	p := New([]byte(`User()`), "test.cm")
	typ := p.parseType()

	if typ == nil {
		t.Fatal("expected type path, got nil")
	}
	if got := tokenLiterals(p, typ.Path); !equalStrings(got, []string{"User"}) {
		t.Fatalf("expected path [User], got %#v", got)
	}
	if p.curTk.Kind != token.L_PAREN {
		t.Fatalf("expected parser to stop before call paren, got %s", p.curTk.Kind)
	}
}

func TestParseTypeReportsMissingPathSegment(t *testing.T) {
	p := New([]byte(`pkg::`), "test.cm")
	typ := p.parseType()

	if typ != nil {
		t.Fatalf("expected nil type, got %#v", typ)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected type path segment" {
		t.Fatalf("expected type path segment diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParseTypeReportsInvalidStart(t *testing.T) {
	p := New([]byte(`123`), "test.cm")
	typ := p.parseType()

	if typ != nil {
		t.Fatalf("expected nil type, got %#v", typ)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected type path" {
		t.Fatalf("expected type path diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}
