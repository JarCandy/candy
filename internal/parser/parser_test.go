package parser

import (
	"testing"

	candyerrors "github.com/CandyCrafts/candy/internal/errors"
	"github.com/CandyCrafts/candy/internal/parser/token"
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
	if literal(p, args[0].Vaule.Value) != "User" {
		t.Fatalf("expected first arg value User, got %q", literal(p, args[0].Vaule.Value))
	}

	if args[1].Name == nil || literal(p, *args[1].Name) != "limit" {
		t.Fatalf("expected second arg name limit, got %v", args[1].Name)
	}
	if args[1].Vaule.Type != types.Int {
		t.Fatalf("expected second arg type int, got %v", args[1].Vaule.Type)
	}
	if literal(p, args[1].Vaule.Value) != "10" {
		t.Fatalf("expected second arg value 10, got %q", literal(p, args[1].Vaule.Value))
	}
}

func TestParseAccessAttr(t *testing.T) {
	p := New([]byte(`db::sqlite("main", table: "User")`), "test.cm")
	attr := p.parseAttr()

	if attr == nil {
		t.Fatal("expected access attr, got nil")
	}
	if got := tokenLiterals(p, attr.Path); !equalStrings(got, []string{"db", "sqlite"}) {
		t.Fatalf("expected path [db sqlite], got %#v", got)
	}
	if len(attr.Args) != 2 {
		t.Fatalf("expected 2 access attr args, got %d", len(attr.Args))
	}

	first := argFromExpr(t, attr.Args[0])
	if first.Name != nil {
		t.Fatalf("expected first arg name to be nil, got %q", *first.Name)
	}
	if literal(p, first.Vaule.Value) != "main" {
		t.Fatalf("expected first arg value main, got %q", literal(p, first.Vaule.Value))
	}

	second := argFromExpr(t, attr.Args[1])
	if second.Name == nil || literal(p, *second.Name) != "table" {
		t.Fatalf("expected second arg name table, got %v", second.Name)
	}
	if literal(p, second.Vaule.Value) != "User" {
		t.Fatalf("expected second arg value User, got %q", literal(p, second.Vaule.Value))
	}
}

func TestParseNestedAccessAttrArg(t *testing.T) {
	p := New([]byte(`db::sqlite(db::std::name())`), "test.cm")
	attr := p.parseAttr()

	if attr == nil {
		t.Fatal("expected access attr, got nil")
	}
	if got := tokenLiterals(p, attr.Path); !equalStrings(got, []string{"db", "sqlite"}) {
		t.Fatalf("expected outer path [db sqlite], got %#v", got)
	}
	if len(attr.Args) != 1 {
		t.Fatalf("expected 1 outer arg, got %d", len(attr.Args))
	}

	arg := argFromExpr(t, attr.Args[0])
	if arg.Vaule.Type != types.Expr {
		t.Fatalf("expected nested attr arg type expr, got %v", arg.Vaule.Type)
	}
	if arg.Vaule.AccessAttr == nil {
		t.Fatal("expected nested access attr, got nil")
	}
	if got := tokenLiterals(p, arg.Vaule.AccessAttr.Path); !equalStrings(got, []string{"db", "std", "name"}) {
		t.Fatalf("expected nested path [db std name], got %#v", got)
	}
	if len(arg.Vaule.AccessAttr.Args) != 0 {
		t.Fatalf("expected nested attr to have no args, got %d", len(arg.Vaule.AccessAttr.Args))
	}
}

func TestParseArgsReportsDiagnostic(t *testing.T) {
	p := New([]byte(`(,)`), "test.cm")
	args := p.parseArgs()

	if len(args) != 0 {
		t.Fatalf("expected empty args, got %#v", args)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].CodeName != "ParsingError" {
		t.Fatalf("expected ParsingError, got %s", p.Diagnostics.Errors[0].CodeName)
	}
}

func TestParseArgsRecoversAfterInvalidArg(t *testing.T) {
	p := New([]byte(`(, "ok")`), "test.cm")
	args := p.parseArgs()

	if len(args) != 1 {
		t.Fatalf("expected 1 recovered arg, got %d", len(args))
	}
	if literal(p, args[0].Vaule.Value) != "ok" {
		t.Fatalf("expected recovered arg value ok, got %q", literal(p, args[0].Vaule.Value))
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
}

func TestParseArgsReportsMissingClosingParen(t *testing.T) {
	p := New([]byte(`("ok"`), "test.cm")
	args := p.parseArgs()

	if len(args) != 1 {
		t.Fatalf("expected 1 arg before EOF, got %d", len(args))
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected closing parenthesis for arguments" {
		t.Fatalf("expected closing paren diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParseAccessAttrReportsDiagnostic(t *testing.T) {
	p := New([]byte(`db::123`), "test.cm")
	attr := p.parseAttr()

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

func TestParseAccessAttrReportsEOFPathSegment(t *testing.T) {
	p := New([]byte(`db::`), "test.cm")
	attr := p.parseAttr()

	if attr != nil {
		t.Fatalf("expected nil attr, got %#v", attr)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected attribute path segment" {
		t.Fatalf("expected attr path diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
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

func TestParsePackageReportsEOFPathSegment(t *testing.T) {
	p := New([]byte(`package::`), "test.cm")
	decl := p.parsePackage()

	if decl != nil {
		t.Fatalf("expected nil package decl, got %#v", decl)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected attribute path segment" {
		t.Fatalf("expected attr path diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParsePackageReportsLongPath(t *testing.T) {
	p := New([]byte(`package::module`), "test.cm")
	decl := p.parsePackage()

	if decl == nil {
		t.Fatal("expected package decl, got nil")
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Package path is too long" {
		t.Fatalf("expected package path diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParsePackageReportsOptionalSemicolonWarning(t *testing.T) {
	p := New([]byte(`package use`), "test.cm")
	decl := p.parsePackage()

	if decl == nil {
		t.Fatal("expected package decl, got nil")
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Severity != candyerrors.SeverityWarning {
		t.Fatalf("expected warning, got %s", p.Diagnostics.Errors[0].Severity)
	}
	if p.Diagnostics.Errors[0].Arrow != "Optional semicolon is missing" {
		t.Fatalf("expected optional semicolon warning, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParsePackageConsumesOptionalSemicolon(t *testing.T) {
	p := New([]byte(`package;`), "test.cm")
	decl := p.parsePackage()

	if decl == nil {
		t.Fatal("expected package decl, got nil")
	}
	if len(p.Diagnostics.Errors) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(p.Diagnostics.Errors))
	}
	if p.curTk.Kind != token.EOF {
		t.Fatalf("expected EOF after optional semicolon, got %s", p.curTk.Kind)
	}
}

func TestParseUseImportsWithAlias(t *testing.T) {
	p := New([]byte(`use ("github.com/CandyCrafts/plugins/db" -> d,)`), "test.cm")
	decl := p.parseUse()

	if decl == nil {
		t.Fatal("expected use decl, got nil")
	}
	if len(decl.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(decl.Imports))
	}
	item := decl.Imports[0]
	if literal(p, item.Link) != "github.com/CandyCrafts/plugins/db" {
		t.Fatalf("expected db link, got %q", literal(p, item.Link))
	}
	if item.Alias == nil || literal(p, *item.Alias) != "d" {
		t.Fatalf("expected alias d, got %v", item.Alias)
	}
	if len(decl.AliasMap) != 1 {
		t.Fatalf("expected 1 alias map entry, got %d", len(decl.AliasMap))
	}
	if got := literal(p, decl.AliasMap[*item.Alias]); got != "github.com/CandyCrafts/plugins/db" {
		t.Fatalf("expected alias map to point to db link, got %q", got)
	}
}

func TestParseUseImportsWithoutAlias(t *testing.T) {
	p := New([]byte(`use ("github.com/CandyCrafts/plugins/db", "github.com/CandyCrafts/plugins/http")`), "test.cm")
	decl := p.parseUse()

	if decl == nil {
		t.Fatal("expected use decl, got nil")
	}
	if len(decl.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(decl.Imports))
	}
	if decl.Imports[0].Alias != nil {
		t.Fatalf("expected first import alias nil, got %v", decl.Imports[0].Alias)
	}
	if len(decl.AliasMap) != 0 {
		t.Fatalf("expected empty alias map, got %d", len(decl.AliasMap))
	}
}

func TestParseUseReportsInvalidImportPath(t *testing.T) {
	p := New([]byte(`use (123, "ok")`), "test.cm")
	decl := p.parseUse()

	if decl == nil {
		t.Fatal("expected use decl, got nil")
	}
	if len(decl.Imports) != 1 {
		t.Fatalf("expected parser to recover 1 import, got %d", len(decl.Imports))
	}
	if literal(p, decl.Imports[0].Link) != "ok" {
		t.Fatalf("expected recovered link ok, got %q", literal(p, decl.Imports[0].Link))
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected import path" {
		t.Fatalf("expected import path diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParseUseReportsMissingClosingParen(t *testing.T) {
	p := New([]byte(`use ("github.com/CandyCrafts/plugins/db"`), "test.cm")
	decl := p.parseUse()

	if decl == nil {
		t.Fatal("expected use decl, got nil")
	}
	if len(decl.Imports) != 1 {
		t.Fatalf("expected 1 import before EOF, got %d", len(decl.Imports))
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected closing parenthesis for use" {
		t.Fatalf("expected closing paren diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParseUseReportsEOFAfterAliasArrow(t *testing.T) {
	p := New([]byte(`use ("github.com/CandyCrafts/plugins/db" ->`), "test.cm")
	decl := p.parseUse()

	if decl == nil {
		t.Fatal("expected use decl, got nil")
	}
	if len(decl.Imports) != 1 {
		t.Fatalf("expected partial import to be kept, got %d", len(decl.Imports))
	}
	if len(p.Diagnostics.Errors) != 2 {
		t.Fatalf("expected alias and closing diagnostics, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected import alias" {
		t.Fatalf("expected alias diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
	if p.Diagnostics.Errors[1].Arrow != "Expected closing parenthesis for use" {
		t.Fatalf("expected closing paren diagnostic, got %q", p.Diagnostics.Errors[1].Arrow)
	}
}

func TestRunDoesNotFailOnWarningsOnly(t *testing.T) {
	p := New([]byte(`package use ("github.com/CandyCrafts/plugins/db")`), "test.cm")
	_, err := p.Run()

	if err != nil {
		t.Fatalf("expected nil error for warning-only diagnostics, got %v", err)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 warning diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Severity != candyerrors.SeverityWarning {
		t.Fatalf("expected warning, got %s", p.Diagnostics.Errors[0].Severity)
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

func literal(p *Parser, tk token.Token) string {
	return string(tk.Literal(&p.Lex.Input))
}

func tokenLiterals(p *Parser, tokens []token.Token) []string {
	values := make([]string, 0, len(tokens))
	for _, tk := range tokens {
		values = append(values, literal(p, tk))
	}
	return values
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
