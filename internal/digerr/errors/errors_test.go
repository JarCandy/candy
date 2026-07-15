package errors

import (
	"testing"

	"github.com/CandyCrafts/candy/internal/digerr/translate"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

func TestErrorUsesSelectedLanguage(t *testing.T) {
	translate.SetLanguage("ru")
	defer translate.SetLanguage(translate.DefaultLanguage)

	err := LexerUnexpectedLess(token.Token{})

	if err.Message != "недопустимый символ" {
		t.Fatalf("expected russian message, got %q", err.Message)
	}
	if err.Arrow != "Неожиданный символ '<'" {
		t.Fatalf("expected russian arrow, got %q", err.Arrow)
	}
}

func TestErrorFallsBackToEnglish(t *testing.T) {
	translate.SetLanguage("missing")
	defer translate.SetLanguage(translate.DefaultLanguage)

	err := ParserArgValue(token.Token{})

	if err.Message != "parse error" {
		t.Fatalf("expected english fallback message, got %q", err.Message)
	}
	if err.Arrow != "Expected argument value" {
		t.Fatalf("expected english fallback arrow, got %q", err.Arrow)
	}
}
