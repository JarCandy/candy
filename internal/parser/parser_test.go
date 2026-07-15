package parser

import (
	"reflect"
	"testing"

	"github.com/CandyCrafts/candy/internal/types"
)

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

func TestParseArgs(t *testing.T) {
	p := New([]byte(`("User", limit: 10)`), "test.cm")
	args := p.parseArgs()

	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}

	if args[0].Name != nil {
		t.Fatalf("expected first arg name to be nil, got %q", *args[0].Name)
	}
	if args[0].Vaule.Type != types.String {
		t.Fatalf("expected first arg type string, got %v", args[0].Vaule.Type)
	}
	if args[0].Vaule.Value != "User" {
		t.Fatalf("expected first arg value User, got %q", args[0].Vaule.Value)
	}

	if args[1].Name == nil || *args[1].Name != "limit" {
		t.Fatalf("expected second arg name limit, got %v", args[1].Name)
	}
	if args[1].Vaule.Type != types.Int {
		t.Fatalf("expected second arg type int, got %v", args[1].Vaule.Type)
	}
	if args[1].Vaule.Value != "10" {
		t.Fatalf("expected second arg value 10, got %q", args[1].Vaule.Value)
	}
}

func TestParseAccessAttr(t *testing.T) {
	p := New([]byte(`db::sqlite("main", table: "User")`), "test.cm")
	attr := p.parseAccessAttr()

	if attr == nil {
		t.Fatal("expected access attr, got nil")
	}
	if !reflect.DeepEqual(attr.Path, []string{"db", "sqlite"}) {
		t.Fatalf("expected path [db sqlite], got %#v", attr.Path)
	}
	if len(attr.Args) != 2 {
		t.Fatalf("expected 2 access attr args, got %d", len(attr.Args))
	}

	first := argFromExpr(t, attr.Args[0])
	if first.Name != nil {
		t.Fatalf("expected first arg name to be nil, got %q", *first.Name)
	}
	if first.Vaule.Value != "main" {
		t.Fatalf("expected first arg value main, got %q", first.Vaule.Value)
	}

	second := argFromExpr(t, attr.Args[1])
	if second.Name == nil || *second.Name != "table" {
		t.Fatalf("expected second arg name table, got %v", second.Name)
	}
	if second.Vaule.Value != "User" {
		t.Fatalf("expected second arg value User, got %q", second.Vaule.Value)
	}
}

func TestParseArgsReportsDiagnostic(t *testing.T) {
	p := New([]byte(`(,)`), "test.cm")
	args := p.parseArgs()

	if args != nil {
		t.Fatalf("expected nil args, got %#v", args)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].CodeName != "ParsingError" {
		t.Fatalf("expected ParsingError, got %s", p.Diagnostics.Errors[0].CodeName)
	}
}

func TestParseAccessAttrReportsDiagnostic(t *testing.T) {
	p := New([]byte(`db::123`), "test.cm")
	attr := p.parseAccessAttr()

	if attr != nil {
		t.Fatalf("expected nil attr, got %#v", attr)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow == "" {
		t.Fatal("expected diagnostic arrow text")
	}
}

func TestRunReportsUnexpectedTopLevelToken(t *testing.T) {
	p := New([]byte(`;`), "test.cm")
	_, err := p.Run()

	if err == nil {
		t.Fatal("expected diagnostics error, got nil")
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].CodeName != "ParsingError" {
		t.Fatalf("expected ParsingError, got %s", p.Diagnostics.Errors[0].CodeName)
	}
}

func argFromExpr(t *testing.T, expr *Expr) *Arg {
	t.Helper()
	if expr == nil {
		t.Fatal("expected expr, got nil")
	}
	arg, ok := (*expr).(*Arg)
	if !ok {
		t.Fatalf("expected *Arg expr, got %T", *expr)
	}
	return arg
}
