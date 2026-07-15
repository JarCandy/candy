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
