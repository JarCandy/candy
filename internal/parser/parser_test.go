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

func TestParseAttrWithTrailingColonCall(t *testing.T) {
	p := New([]byte(`db::sqlite::table::("User")`), "test.cm")
	attr := p.parseAttr()

	if attr == nil {
		t.Fatal("expected attr, got nil")
	}
	if got := tokenLiterals(p, attr.Path); !equalStrings(got, []string{"db", "sqlite", "table"}) {
		t.Fatalf("expected path [db sqlite table], got %#v", got)
	}
	if len(attr.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(attr.Args))
	}
	arg := argFromExpr(t, attr.Args[0])
	if literal(p, arg.Vaule.Value) != "User" {
		t.Fatalf("expected arg User, got %q", literal(p, arg.Vaule.Value))
	}
	if p.curTk.Kind != token.EOF {
		t.Fatalf("expected EOF after attr, got %s", p.curTk.Kind)
	}
}

func TestParseAttrWithDirectCall(t *testing.T) {
	p := New([]byte(`db::sqlite::table("User")`), "test.cm")
	attr := p.parseAttr()

	if attr == nil {
		t.Fatal("expected attr, got nil")
	}
	if got := tokenLiterals(p, attr.Path); !equalStrings(got, []string{"db", "sqlite", "table"}) {
		t.Fatalf("expected path [db sqlite table], got %#v", got)
	}
	if len(attr.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(attr.Args))
	}
	if p.curTk.Kind != token.EOF {
		t.Fatalf("expected EOF after attr, got %s", p.curTk.Kind)
	}
}

func TestParseAttrAssignmentCreatesMapEntry(t *testing.T) {
	p := New([]byte(`#[lang=custom("github.com/CandyCrafts/LangEngines/Go@latest")];`), "test.cm")
	attrs := p.parseAttrs()

	if attrs == nil {
		t.Fatal("expected attrs, got nil")
	}
	attr := attrs.Map["lang"]
	if attr == nil {
		t.Fatalf("expected lang attr in map, got %#v", attrs.Map)
	}
	if attr.Value == nil {
		t.Fatal("expected lang attr assignment value")
	}
	value, ok := (*attr.Value).(*Attr)
	if !ok {
		t.Fatalf("expected attr assignment value, got %T", *attr.Value)
	}
	if got := tokenLiterals(p, value.Path); !equalStrings(got, []string{"custom"}) {
		t.Fatalf("expected custom value path, got %#v", got)
	}
	if len(value.Args) != 1 {
		t.Fatalf("expected 1 custom arg, got %d", len(value.Args))
	}
}

func TestParseAccessExpressionWithCallChain(t *testing.T) {
	p := New([]byte(`go::lib("github.com/google/uuid")::NewString()`), "test.cm")
	expr := p.ParseExpr()

	if expr == nil {
		t.Fatal("expected expr, got nil")
	}
	attr, ok := (*expr).(*Attr)
	if !ok {
		t.Fatalf("expected access attr expr, got %T", *expr)
	}
	if got := tokenLiterals(p, attr.Path); !equalStrings(got, []string{"go", "lib", "NewString"}) {
		t.Fatalf("expected path [go lib NewString], got %#v", got)
	}
	if len(attr.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(attr.Args))
	}
	if p.curTk.Kind != token.EOF {
		t.Fatalf("expected EOF after access expr, got %s", p.curTk.Kind)
	}
}

func TestParseAttrs(t *testing.T) {
	p := New([]byte(`#[db::sqlite::table::("User")];`), "test.cm")
	attrs := p.parseAttrs()

	if attrs == nil {
		t.Fatal("expected attrs, got nil")
	}
	if attrs.Token_s().Kind != token.ATTR_S {
		t.Fatalf("expected ATTR_S token, got %s", attrs.Token_s().Kind)
	}
	if attrs.Token_e().Kind != token.ATTR_E {
		t.Fatalf("expected ATTR_E token, got %s", attrs.Token_e().Kind)
	}
	if len(attrs.Attrs) != 1 {
		t.Fatalf("expected 1 attr, got %d", len(attrs.Attrs))
	}
	if got := tokenLiterals(p, attrs.Attrs[0].Path); !equalStrings(got, []string{"db", "sqlite", "table"}) {
		t.Fatalf("expected path [db sqlite table], got %#v", got)
	}
	if len(attrs.Attrs[0].Args) != 1 {
		t.Fatalf("expected 1 attr arg, got %d", len(attrs.Attrs[0].Args))
	}
	if p.curTk.Kind != token.EOF {
		t.Fatalf("expected EOF after attrs, got %s", p.curTk.Kind)
	}
}

func TestParseAttrsReportsOptionalSemicolonWarning(t *testing.T) {
	p := New([]byte(`#[db::sqlite] let name: User`), "test.cm")
	attrs := p.parseAttrs()

	if attrs == nil {
		t.Fatal("expected attrs, got nil")
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
	if p.curTk.Kind != token.LET {
		t.Fatalf("expected parser to stop at LET, got %s", p.curTk.Kind)
	}
}

func TestParseAttrsReportsMissingClosingBracket(t *testing.T) {
	p := New([]byte(`#[db::sqlite`), "test.cm")
	attrs := p.parseAttrs()

	if attrs == nil {
		t.Fatal("expected partial attrs, got nil")
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected closing attribute bracket" {
		t.Fatalf("expected closing attr diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestRunParsesAttrsDeclarations(t *testing.T) {
	p := New([]byte(`#[db::sqlite::table::("User")];`), "test.cm")
	ast, err := p.Run()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ast.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(ast.Decls))
	}
	decl, ok := ast.Decls[0].(*AttrsDecl)
	if !ok {
		t.Fatalf("expected *AttrsDecl, got %T", ast.Decls[0])
	}
	if decl.Attrs == nil {
		t.Fatal("expected attrs payload")
	}
	if len(decl.Attrs.Attrs) != 1 {
		t.Fatalf("expected 1 attr, got %d", len(decl.Attrs.Attrs))
	}
}

func TestParseAttrsStmt(t *testing.T) {
	p := New([]byte(`#[db::sqlite::index];`), "test.cm")
	stmt := p.parseAttrsStmt()

	if stmt == nil {
		t.Fatal("expected attrs stmt, got nil")
	}
	if stmt.Attrs == nil {
		t.Fatal("expected attrs payload")
	}
	if len(stmt.Attrs.Attrs) != 1 {
		t.Fatalf("expected 1 attr, got %d", len(stmt.Attrs.Attrs))
	}
	if got := tokenLiterals(p, stmt.Attrs.Attrs[0].Path); !equalStrings(got, []string{"db", "sqlite", "index"}) {
		t.Fatalf("expected path [db sqlite index], got %#v", got)
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

func TestParseLetVarWithPubTypeAndDefault(t *testing.T) {
	p := New([]byte(`pub let name: db::User = "Candy";`), "test.cm")
	decl := p.parseLetVar(true)

	if decl == nil {
		t.Fatal("expected let decl, got nil")
	}
	if !decl.Pub {
		t.Fatal("expected public let")
	}
	if literal(p, decl.Name) != "name" {
		t.Fatalf("expected name token, got %q", literal(p, decl.Name))
	}

	typ, ok := decl.Type.(*TypeExpr)
	if !ok {
		t.Fatalf("expected *TypeExpr, got %T", decl.Type)
	}
	if got := tokenLiterals(p, typ.Path); !equalStrings(got, []string{"db", "User"}) {
		t.Fatalf("expected type path [db User], got %#v", got)
	}
	if decl.Defualt == nil {
		t.Fatal("expected default expr")
	}
	value, ok := (*decl.Defualt).(LiteralExpr)
	if !ok {
		t.Fatalf("expected literal default, got %T", *decl.Defualt)
	}
	if literal(p, value.Value) != "Candy" {
		t.Fatalf("expected default Candy, got %q", literal(p, value.Value))
	}
	if p.curTk.Kind != token.EOF {
		t.Fatalf("expected EOF after semicolon, got %s", p.curTk.Kind)
	}
}

func TestParseLetVarWithTypeOnly(t *testing.T) {
	p := New([]byte(`let name: User`), "test.cm")
	decl := p.parseLetVar(true)

	if decl == nil {
		t.Fatal("expected let decl, got nil")
	}
	if decl.Pub {
		t.Fatal("expected private let")
	}
	if decl.Type == nil {
		t.Fatal("expected type")
	}
	if decl.Defualt != nil {
		t.Fatalf("expected nil default, got %#v", decl.Defualt)
	}
	if len(p.Diagnostics.Errors) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(p.Diagnostics.Errors))
	}
}

func TestParseLetVarWithDefaultOnly(t *testing.T) {
	p := New([]byte(`let name = 10`), "test.cm")
	decl := p.parseLetVar(true)

	if decl == nil {
		t.Fatal("expected let decl, got nil")
	}
	if decl.Type != nil {
		t.Fatalf("expected nil type, got %#v", decl.Type)
	}
	if decl.Defualt == nil {
		t.Fatal("expected default expr")
	}
}

func TestParseLetVarReportsMissingName(t *testing.T) {
	p := New([]byte(`let : User`), "test.cm")
	decl := p.parseLetVar(true)

	if decl != nil {
		t.Fatalf("expected nil let decl, got %#v", decl)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected variable name" {
		t.Fatalf("expected variable name diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParseLetVarReportsMissingBody(t *testing.T) {
	p := New([]byte(`let name`), "test.cm")
	decl := p.parseLetVar(true)

	if decl == nil {
		t.Fatal("expected partial let decl, got nil")
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected variable type or value" {
		t.Fatalf("expected let body diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestParseLetVarReportsMissingDefaultValue(t *testing.T) {
	p := New([]byte(`let name = ;`), "test.cm")
	decl := p.parseLetVar(true)

	if decl == nil {
		t.Fatal("expected partial let decl, got nil")
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected variable value" {
		t.Fatalf("expected let value diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
	}
}

func TestRunParsesLetDeclarations(t *testing.T) {
	p := New([]byte(`pub let name: User = "Candy"`), "test.cm")
	ast, err := p.Run()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ast.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(ast.Decls))
	}
	decl, ok := ast.Decls[0].(*LetDecl)
	if !ok {
		t.Fatalf("expected *LetDecl decl, got %T", ast.Decls[0])
	}
	if decl.Let == nil {
		t.Fatal("expected let payload")
	}
	if !decl.Let.Pub {
		t.Fatal("expected public let")
	}
}

func TestRunParsesPubLetGroup(t *testing.T) {
	p := New([]byte(`pub (
    let name: string = "Candy"
    let id: go::type::string = go::lib("github.com/google/uuid")::NewString()
)`), "test.cm")
	ast, err := p.Run()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ast.Decls) != 2 {
		t.Fatalf("expected 2 decls, got %d", len(ast.Decls))
	}
	for i, decl := range ast.Decls {
		letDecl, ok := decl.(*LetDecl)
		if !ok {
			t.Fatalf("expected decl %d to be *LetDecl, got %T", i, decl)
		}
		if letDecl.Let == nil || !letDecl.Let.Pub {
			t.Fatalf("expected decl %d to be public let, got %#v", i, letDecl.Let)
		}
	}

	second := ast.Decls[1].(*LetDecl)
	if got := tokenLiterals(p, second.Type.(*TypeExpr).Path); !equalStrings(got, []string{"go", "type", "string"}) {
		t.Fatalf("expected second type path [go type string], got %#v", got)
	}
	if _, ok := (*second.Defualt).(*Attr); !ok {
		t.Fatalf("expected second default attr expr, got %T", *second.Defualt)
	}
}

func TestRunParsesModelAndImplSyntax(t *testing.T) {
	source := []byte(`package("main");

use (
    "github.com/CandyCrafts/plugins/db" -> d,
)

#[lang=custom("github.com/CandyCrafts/LangEngines/Go@latest")];

#[db::sqlite::table("User")];
go::model User {
    #[db::sqlite::index];
    pub Id: strings = go::lib("github.com/google/uuid")::NewString()
    pub Name: string = "none",
}

#[comoser::file::no_edit(true)];
go::impl User {
    #[db::go::func::delete_rec];
    pub banned() -> go::type::error,
}`)
	p := New(source, "test.cm")
	ast, err := p.Run()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ast.Decls) != 7 {
		t.Fatalf("expected 7 top-level decls, got %d", len(ast.Decls))
	}

	langAttrs, ok := ast.Decls[2].(*AttrsDecl)
	if !ok {
		t.Fatalf("expected lang attrs decl, got %T", ast.Decls[2])
	}
	if langAttrs.Attrs.Map["lang"] == nil {
		t.Fatal("expected lang attr map entry")
	}

	model, ok := ast.Decls[4].(*QualifiedDecl)
	if !ok {
		t.Fatalf("expected qualified model decl, got %T", ast.Decls[4])
	}
	if literal(p, model.Name) != "User" {
		t.Fatalf("expected model name User, got %q", literal(p, model.Name))
	}
	if got := tokenLiterals(p, model.Path); !equalStrings(got, []string{"go", "model"}) {
		t.Fatalf("expected model path [go model], got %#v", got)
	}
	if len(model.Body) != 3 {
		t.Fatalf("expected 3 model statements, got %d", len(model.Body))
	}
	field, ok := model.Body[1].(*LetStmt)
	if !ok {
		t.Fatalf("expected field stmt, got %T", model.Body[1])
	}
	if field.Defualt == nil {
		t.Fatal("expected default access expr")
	}
	if _, ok := (*field.Defualt).(*Attr); !ok {
		t.Fatalf("expected default attr expr, got %T", *field.Defualt)
	}

	impl, ok := ast.Decls[6].(*QualifiedDecl)
	if !ok {
		t.Fatalf("expected qualified impl decl, got %T", ast.Decls[6])
	}
	if got := tokenLiterals(p, impl.Path); !equalStrings(got, []string{"go", "impl"}) {
		t.Fatalf("expected impl path [go impl], got %#v", got)
	}
	if len(impl.Body) != 2 {
		t.Fatalf("expected 2 impl statements, got %d", len(impl.Body))
	}
	method, ok := impl.Body[1].(*MethodStmt)
	if !ok {
		t.Fatalf("expected method stmt, got %T", impl.Body[1])
	}
	if literal(p, method.Name) != "banned" {
		t.Fatalf("expected method name banned, got %q", literal(p, method.Name))
	}
	ret, ok := method.Return.(*TypeExpr)
	if !ok {
		t.Fatalf("expected method return type, got %T", method.Return)
	}
	if got := tokenLiterals(p, ret.Path); !equalStrings(got, []string{"go", "type", "error"}) {
		t.Fatalf("expected return path [go type error], got %#v", got)
	}
}

func TestRunParsesPubMemberGroupInModel(t *testing.T) {
	p := New([]byte(`go::model User {
    pub (
        Id: strings = go::lib("github.com/google/uuid")::NewString()
        Name: string = "none"
    )
}`), "test.cm")
	ast, err := p.Run()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ast.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(ast.Decls))
	}
	model, ok := ast.Decls[0].(*QualifiedDecl)
	if !ok {
		t.Fatalf("expected qualified model decl, got %T", ast.Decls[0])
	}
	if len(model.Body) != 2 {
		t.Fatalf("expected 2 model fields, got %d", len(model.Body))
	}
	for i, stmt := range model.Body {
		let, ok := stmt.(*LetStmt)
		if !ok {
			t.Fatalf("expected body %d to be *LetStmt, got %T", i, stmt)
		}
		if let.Let == nil || !let.Let.Pub {
			t.Fatalf("expected body %d to be public let, got %#v", i, let.Let)
		}
	}
}

func TestRunParsesSingleTokenQualifiedDeclPath(t *testing.T) {
	p := New([]byte(`model User {
    pub (
        name: type = expr,
    )
}`), "test.cm")
	ast, err := p.Run()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ast.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(ast.Decls))
	}
	decl, ok := ast.Decls[0].(*QualifiedDecl)
	if !ok {
		t.Fatalf("expected qualified decl, got %T", ast.Decls[0])
	}
	if got := tokenLiterals(p, decl.Path); !equalStrings(got, []string{"model"}) {
		t.Fatalf("expected path [model], got %#v", got)
	}
	if literal(p, decl.Name) != "User" {
		t.Fatalf("expected decl name User, got %q", literal(p, decl.Name))
	}
	if len(decl.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(decl.Body))
	}
}

func TestRunParsesPubMemberGroupWithTrailingComma(t *testing.T) {
	p := New([]byte(`go::model Name {
    pub (
        name: type = expr,
    )
}`), "test.cm")
	ast, err := p.Run()

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ast.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(ast.Decls))
	}
	decl, ok := ast.Decls[0].(*QualifiedDecl)
	if !ok {
		t.Fatalf("expected qualified model decl, got %T", ast.Decls[0])
	}
	if literal(p, decl.Name) != "Name" {
		t.Fatalf("expected model name Name, got %q", literal(p, decl.Name))
	}
	if got := tokenLiterals(p, decl.Path); !equalStrings(got, []string{"go", "model"}) {
		t.Fatalf("expected model path [go model], got %#v", got)
	}
	if len(decl.Body) != 1 {
		t.Fatalf("expected 1 model field, got %d", len(decl.Body))
	}
	field, ok := decl.Body[0].(*LetStmt)
	if !ok {
		t.Fatalf("expected model field let stmt, got %T", decl.Body[0])
	}
	if field.Let == nil || !field.Let.Pub {
		t.Fatalf("expected public model field, got %#v", field.Let)
	}
	if literal(p, field.Name) != "name" {
		t.Fatalf("expected field name, got %q", literal(p, field.Name))
	}
	if p.Diagnostics.HasFatalErrors() {
		t.Fatalf("expected no fatal diagnostics, got %v", p.Diagnostics)
	}
}

func TestParseLetStmt(t *testing.T) {
	p := New([]byte(`let name: User = "Candy"`), "test.cm")
	stmt := p.parseLetStmt()

	if stmt == nil {
		t.Fatal("expected let stmt, got nil")
	}
	if stmt.Let == nil {
		t.Fatal("expected let payload")
	}
	if stmt.Let.Pub {
		t.Fatal("expected private let stmt")
	}
	if literal(p, stmt.Let.Name) != "name" {
		t.Fatalf("expected name token, got %q", literal(p, stmt.Let.Name))
	}
	if stmt.Let.Type == nil {
		t.Fatal("expected let stmt type")
	}
	if stmt.Let.Defualt == nil {
		t.Fatal("expected let stmt default")
	}
}

func TestParseLetStmtRejectsPub(t *testing.T) {
	p := New([]byte(`pub let name: User`), "test.cm")
	stmt := p.parseLetStmt()

	if stmt != nil {
		t.Fatalf("expected nil let stmt, got %#v", stmt)
	}
	if len(p.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(p.Diagnostics.Errors))
	}
	if p.Diagnostics.Errors[0].Arrow != "Expected let declaration" {
		t.Fatalf("expected let start diagnostic, got %q", p.Diagnostics.Errors[0].Arrow)
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
