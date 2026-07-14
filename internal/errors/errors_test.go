package errors

import (
	"context"
	"testing"

	"github.com/rp1s/digreyt/translate"
)

type fakeAutoTranslator struct{}

func (fakeAutoTranslator) Translate(_ context.Context, sourceLanguage, targetLanguage, text string) (string, error) {
	return sourceLanguage + "->" + targetLanguage + ":" + text, nil
}

func TestErrorUsesSelectedLanguage(t *testing.T) {
	translate.SetLanguage("ru")
	defer translate.SetLanguage(translate.DefaultLanguage)

	err := LexerUnexpectedLess(Span{})

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

	err := ParserArgValue(Span{})

	if err.Message != "parse error" {
		t.Fatalf("expected english fallback message, got %q", err.Message)
	}
	if err.Arrow != "Expected argument value" {
		t.Fatalf("expected english fallback arrow, got %q", err.Arrow)
	}
}

func TestErrorAutoTranslatesMissingLanguage(t *testing.T) {
	prevLanguage := translate.Language()
	prevTranslator := translate.AutoTranslatorProvider()
	translate.SetLanguage("candy")
	translate.SetAutoTranslator(fakeAutoTranslator{})
	defer func() {
		translate.SetLanguage(prevLanguage)
		translate.SetAutoTranslator(prevTranslator)
	}()

	err := ParserArgValue(Span{})
	localized, translateErr := err.LocalizeAuto(context.Background())
	if translateErr != nil {
		t.Fatalf("LocalizeAuto() failed: %v", translateErr)
	}

	if localized.Message != "eng->candy:parse error" {
		t.Fatalf("expected auto-translated message, got %q", localized.Message)
	}
	if localized.Arrow != "eng->candy:Expected argument value" {
		t.Fatalf("expected auto-translated arrow, got %q", localized.Arrow)
	}
}

func TestTextCanUseEnglishOnlyForAutoTranslation(t *testing.T) {
	values := text("single source text")

	if len(values) != 1 {
		t.Fatalf("expected only source translation, got %d", len(values))
	}
	if values[0].Language != translate.DefaultLanguage {
		t.Fatalf("expected default language, got %q", values[0].Language)
	}
	if values[0].Text != "single source text" {
		t.Fatalf("expected source text, got %q", values[0].Text)
	}
}
