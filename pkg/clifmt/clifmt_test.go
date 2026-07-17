package clifmt

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/rp1s/digreyt/translate"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestRenderUsesSelectedLanguage(t *testing.T) {
	doc := Document{
		Title: T("Candy help", Lang("ru", "Справка Candy")),
		Usage: T("Usage: candy", Lang("ru", "Использование: candy")),
		Sections: []Section{
			{
				Title: T("Commands", Lang("ru", "Команды")),
				Rows: []Row{
					{Label: "build", Description: T("Build a file.", Lang("ru", "Собрать файл."))},
				},
			},
		},
	}

	out := stripANSI(Sprint(doc, "ru"))
	for _, part := range []string{"Справка Candy", "Использование: candy", "Команды", "Собрать файл."} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected output to contain %q, got %q", part, out)
		}
	}
}

func TestRenderFallsBackToEnglish(t *testing.T) {
	prevTranslator := translate.AutoTranslatorProvider()
	translate.SetAutoTranslator(nil)
	t.Cleanup(func() {
		translate.SetAutoTranslator(prevTranslator)
	})

	doc := Document{
		Title: T("Candy help"),
		Sections: []Section{
			{
				Title: T("Commands"),
				Rows:  []Row{{Label: "help", Description: T("Show help.")}},
			},
		},
	}

	out := stripANSI(Sprint(doc, "missing"))
	for _, part := range []string{"Candy help", "Commands", "Show help."} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected output to contain %q, got %q", part, out)
		}
	}
}

func TestRenderAutoTranslatesMissingLanguage(t *testing.T) {
	prevTranslator := translate.AutoTranslatorProvider()
	translate.SetAutoTranslator(fakeTranslator{prefix: true})
	t.Cleanup(func() {
		translate.SetAutoTranslator(prevTranslator)
	})

	doc := Document{
		Title: T("Candy help"),
		Usage: T("Usage: candy"),
		Sections: []Section{
			{
				Title: T("Commands"),
				Rows:  []Row{{Label: "build", Description: T("Build a file.")}},
			},
		},
	}

	out := stripANSI(Sprint(doc, "uk"))
	for _, part := range []string{"eng|uk:Candy help", "eng|uk:Usage: candy", "eng|uk:Commands", "eng|uk:Build a file."} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected output to contain %q, got %q", part, out)
		}
	}
}

func TestRenderFallsBackWhenAutoTranslatorReturnsProviderError(t *testing.T) {
	prevTranslator := translate.AutoTranslatorProvider()
	translate.SetAutoTranslator(fakeTranslator{text: "'FG' IS AN INVALID TARGET LANGUAGE . EXAMPLE: LANGPAIR=EN|IT"})
	t.Cleanup(func() {
		translate.SetAutoTranslator(prevTranslator)
	})

	doc := Document{
		Title: T("Candy help"),
		Sections: []Section{
			{
				Title: T("Commands"),
				Rows:  []Row{{Label: "help", Description: T("Show help.")}},
			},
		},
	}

	out := stripANSI(Sprint(doc, "fg"))
	for _, part := range []string{"Candy help", "Commands", "Show help."} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected english fallback to contain %q, got %q", part, out)
		}
	}
	if strings.Contains(out, "INVALID TARGET LANGUAGE") {
		t.Fatalf("expected provider error to be hidden, got %q", out)
	}
}

func stripANSI(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}

type fakeTranslator struct {
	prefix bool
	text   string
}

func (self fakeTranslator) Translate(ctx context.Context, sourceLanguage, targetLanguage, text string) (string, error) {
	if self.text != "" {
		return self.text, nil
	}
	if !self.prefix {
		return text, nil
	}
	return sourceLanguage + "|" + targetLanguage + ":" + text, nil
}
