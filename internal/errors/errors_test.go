package errors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rp1s/digreyt/translate"
)

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("langpair"); got != "en|candy" {
			t.Errorf("expected MyMemory langpair en|candy, got %q", got)
			http.Error(w, "bad langpair", http.StatusBadRequest)
			return
		}
		text := r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"responseData":{"translatedText":"en|candy:` + text + `"}}`))
	}))
	defer server.Close()

	translator := translate.NewMyMemoryTranslator()
	translator.Endpoint = server.URL

	translate.SetLanguage("candy")
	translate.SetAutoTranslator(translator)
	defer func() {
		translate.SetLanguage(prevLanguage)
		translate.SetAutoTranslator(prevTranslator)
	}()

	err := ParserArgValue(Span{})
	localized, translateErr := err.LocalizeAuto(context.Background())
	if translateErr != nil {
		t.Fatalf("LocalizeAuto() failed: %v", translateErr)
	}

	if localized.Message != "en|candy:parse error" {
		t.Fatalf("expected auto-translated message, got %q", localized.Message)
	}
	if localized.Arrow != "en|candy:Expected argument value" {
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
