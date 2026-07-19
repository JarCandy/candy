package analyzer

import (
	"testing"

	"github.com/caramelang/caramel/internal/composer"
	"github.com/caramelang/caramel/internal/parser"
)

func TestProjectChecksLetTypesBySimpleComparison(t *testing.T) {
	source := []byte(`package("main")

let title: string = "Caramel"
let age: string = 10
let active: bool = true
let badFlag: bool = "true"
let sum: int = 1 + 2
let negative: int = -1
let namespaced: go::type::string = "x"

go::model User {
	pub (
		Name: string = "none"
		Count: int = 1
	)
}
`)

	result := analyzeSource(t, source)
	if len(result.Files) != 1 {
		t.Fatalf("expected one file, got %d", len(result.Files))
	}

	file := result.Files[0]
	if file.LetCount != 9 {
		t.Fatalf("expected 9 lets, got %d", file.LetCount)
	}
	if file.TypeChecks != 9 {
		t.Fatalf("expected 9 type checks, got %d", file.TypeChecks)
	}
	if len(file.TypeErrors) != 3 {
		t.Fatalf("expected three type errors, got %d", len(file.TypeErrors))
	}

	assertTypeError(t, file.TypeErrors, "age", "string", "int")
	assertTypeError(t, file.TypeErrors, "badFlag", "bool", "string")
	assertTypeError(t, file.TypeErrors, "namespaced", "go::type::string", "string")
	if !result.Diagnostics.HasFatalErrors() {
		t.Fatal("expected fatal diagnostic for type mismatch")
	}
}

func TestProjectSkipsUnknownExprTypes(t *testing.T) {
	source := []byte(`package("main")

let id: string = go::lib("github.com/google/uuid")::NewString()
`)

	result := analyzeSource(t, source)
	file := result.Files[0]

	if file.LetCount != 1 {
		t.Fatalf("expected one let, got %d", file.LetCount)
	}
	if file.TypeChecks != 0 {
		t.Fatalf("expected unknown expression to be skipped, got %d checks", file.TypeChecks)
	}
	if len(file.TypeErrors) != 0 {
		t.Fatalf("expected no type errors, got %d", len(file.TypeErrors))
	}
}

func TestTypeCheckerKeepsTypeModifiers(t *testing.T) {
	source := []byte(`let name: *[]*string = "none"`)
	result := analyzeSource(t, source)

	if len(result.Files[0].TypeErrors) != 1 {
		t.Fatalf("expected one type error, got %#v", result.Files[0].TypeErrors)
	}
	assertTypeError(t, result.Files[0].TypeErrors, "name", "*[]*string", "string")
}

func analyzeSource(t *testing.T, source []byte) *Result {
	t.Helper()

	ast, err := parser.New(source, "test.cm").Run()
	if err != nil {
		t.Fatal(err)
	}

	project := &composer.Project{
		Name: "test",
		AstFile: []composer.AstFile{
			{
				FileName: "test.cm",
				Path:     "test.cm",
				Source:   source,
				Ast:      *ast,
			},
		},
	}

	result, err := New().Project(project)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertTypeError(t *testing.T, errors []TypeError, name string, declared string, got string) {
	t.Helper()

	for _, err := range errors {
		if err.Name == name && err.Declared == declared && err.Got == got {
			return
		}
	}

	t.Fatalf("expected type error %s %s/%s, got %#v", name, declared, got, errors)
}
