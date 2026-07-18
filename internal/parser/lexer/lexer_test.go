package lexer

import (
	"testing"

	"github.com/CandyCrafts/candy/internal/parser/token"

	"github.com/k0kubun/pp"
)

func TestLexer(t *testing.T) {
	lex := New([]byte(`
package("main")

use (
    "github.com/CandyCrafts/plugins/db" // import NEW db
)

// global attr
#[lang=custom("github.com/CandyCrafts/LangEngines/Go@latest")] // import NEW engine

#[db::sqlite::table("User")]
go::model User {
    #[db::sqlite::index] // behavior tag
    pub Id: strings = go::lib("github.com/google/uuid")::NewString()
    pub Name: string = "none"
}

#[composer::file::no_edit(true)]
go::impl User {
    #[db::go::func::delete_rec]
    pub banned() -> go::type::error
}
	`), "model.cp")

	for {
		tk := lex.NextToken()
		pp.Println(tk.Kind.String())
		if tk.Kind == token.EOF {
			break
		}
	}
}

func TestLexerRejectsSemicolon(t *testing.T) {
	lex := New([]byte(";"), "test.cm")
	tk := lex.NextToken()

	if tk.Kind != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL token, got %s", tk.Kind)
	}
	if len(lex.Diagnostics.Errors) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(lex.Diagnostics.Errors))
	}
	if lex.Diagnostics.Errors[0].Arrow != "Semicolons are not supported" {
		t.Fatalf("expected semicolon diagnostic, got %q", lex.Diagnostics.Errors[0].Arrow)
	}
}

func TestLexerRejectsComma(t *testing.T) {
	lex := New([]byte(","), "test.cm")
	tk := lex.NextToken()

	if tk.Kind != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL token, got %s", tk.Kind)
	}
	if len(lex.Diagnostics.Errors) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(lex.Diagnostics.Errors))
	}
	if lex.Diagnostics.Errors[0].Arrow != "Commas are not supported" {
		t.Fatalf("expected comma diagnostic, got %q", lex.Diagnostics.Errors[0].Arrow)
	}
}

func TestLexerReportsIllegalDiagnostic(t *testing.T) {
	lex := New([]byte("<"), "test.cm")
	tk := lex.NextToken()

	if tk.Kind != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL token, got %s", tk.Kind)
	}
	if len(lex.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(lex.Diagnostics.Errors))
	}
	if lex.Diagnostics.Errors[0].CodeName != "LexerIllegal" {
		t.Fatalf("expected LexerIllegal, got %s", lex.Diagnostics.Errors[0].CodeName)
	}
}

func TestLexerReportsNoClosingDiagnostic(t *testing.T) {
	lex := New([]byte(`"unterminated`), "test.cm")
	tk := lex.NextToken()

	if tk.Kind != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL token, got %s", tk.Kind)
	}
	if len(lex.Diagnostics.Errors) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(lex.Diagnostics.Errors))
	}
	if lex.Diagnostics.Errors[0].CodeName != "LexerNoClosing" {
		t.Fatalf("expected LexerNoClosing, got %s", lex.Diagnostics.Errors[0].CodeName)
	}
}
