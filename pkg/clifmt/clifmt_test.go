package clifmt

import (
	"context"
	"regexp"
	"strings"
	"testing"

	cdb "github.com/caramelang/caramel/internal/database"
	"github.com/rp1s/digreyt/translate"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestRenderUsesSelectedLanguage(t *testing.T) {
	doc := Document{
		Title: T("Caramel help", Lang("ru", "Справка Caramel")),
		Usage: T("Usage: caramel", Lang("ru", "Использование: caramel")),
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
	for _, part := range []string{"Справка Caramel", "Использование: caramel", "Команды", "Собрать файл."} {
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
		Title: T("Caramel help"),
		Sections: []Section{
			{
				Title: T("Commands"),
				Rows:  []Row{{Label: "help", Description: T("Show help.")}},
			},
		},
	}

	out := stripANSI(Sprint(doc, "missing"))
	for _, part := range []string{"Caramel help", "Commands", "Show help."} {
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
		Title: T("Caramel help"),
		Usage: T("Usage: caramel"),
		Sections: []Section{
			{
				Title: T("Commands"),
				Rows:  []Row{{Label: "build", Description: T("Build a file.")}},
			},
		},
	}

	out := stripANSI(Sprint(doc, "uk"))
	for _, part := range []string{"eng|uk:Caramel help", "eng|uk:Usage: caramel", "eng|uk:Commands", "eng|uk:Build a file."} {
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
		Title: T("Caramel help"),
		Sections: []Section{
			{
				Title: T("Commands"),
				Rows:  []Row{{Label: "help", Description: T("Show help.")}},
			},
		},
	}

	out := stripANSI(Sprint(doc, "fg"))
	for _, part := range []string{"Caramel help", "Commands", "Show help."} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected english fallback to contain %q, got %q", part, out)
		}
	}
	if strings.Contains(out, "INVALID TARGET LANGUAGE") {
		t.Fatalf("expected provider error to be hidden, got %q", out)
	}
}

func TestRenderUsesPersistedCacheEntries(t *testing.T) {
	prevTranslator := translate.AutoTranslatorProvider()
	translate.SetAutoTranslator(fakeTranslator{prefix: true})
	t.Cleanup(func() {
		translate.SetAutoTranslator(prevTranslator)
	})

	store := &fakeCLITextCache{entries: make(map[string]string)}
	doc := Document{Title: T("Caramel help")}

	renderer := New("uk")
	renderer.Auto = true
	renderer.cacheStore = store
	out := stripANSI(renderer.Render(doc))
	if !strings.Contains(out, "eng|uk:Caramel help") {
		t.Fatalf("expected first render to use auto translation, got %q", out)
	}

	translate.SetAutoTranslator(fakeTranslator{text: "should-not-be-used"})
	renderer2 := New("uk")
	renderer2.Auto = true
	renderer2.cacheStore = store
	out2 := stripANSI(renderer2.Render(doc))
	if strings.Contains(out2, "should-not-be-used") {
		t.Fatalf("expected cached translation to be reused, got %q", out2)
	}
	if !strings.Contains(out2, "eng|uk:Caramel help") {
		t.Fatalf("expected second render to reuse persisted cache value, got %q", out2)
	}
}

func stripANSI(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}

type fakeCLITextCache struct {
	entries map[string]string
}

func (self *fakeCLITextCache) Init(ctx context.Context) error { return nil }
func (self *fakeCLITextCache) Create(ctx context.Context, entry cdb.CLITextEntry) error {
	self.entries[entry.Lang+"\x00"+entry.OriginalText] = entry.Text
	return nil
}
func (self *fakeCLITextCache) Get(ctx context.Context, filters map[string]any) (cdb.CLITextEntry, error) {
	if originalText, ok := filters["OriginalText"].(string); ok {
		if lang, ok := filters["Lang"].(string); ok {
			if text, ok := self.entries[lang+"\x00"+originalText]; ok {
				return cdb.CLITextEntry{Lang: lang, OriginalText: originalText, Text: text}, nil
			}
		}
	}
	return cdb.CLITextEntry{}, context.DeadlineExceeded
}
func (self *fakeCLITextCache) Update(ctx context.Context, entry cdb.CLITextEntry, filters map[string]any) error {
	if originalText, ok := filters["OriginalText"].(string); ok {
		if lang, ok := filters["Lang"].(string); ok {
			self.entries[lang+"\x00"+originalText] = entry.Text
		}
	}
	return nil
}
func (self *fakeCLITextCache) Delete(ctx context.Context, filters map[string]any) error { return nil }
func (self *fakeCLITextCache) List(ctx context.Context, limit int, offset int) ([]cdb.CLITextEntry, error) {
	return nil, nil
}
func (self *fakeCLITextCache) Search(ctx context.Context, term string, limit int, offset int) ([]cdb.CLITextEntry, error) {
	return nil, nil
}
func (self *fakeCLITextCache) ReadText(ctx context.Context, originalText string, lang string) (string, bool, error) {
	text, ok := self.entries[lang+"\x00"+originalText]
	return text, ok, nil
}
func (self *fakeCLITextCache) WriteText(ctx context.Context, originalText string, lang string, text string) error {
	self.entries[lang+"\x00"+originalText] = text
	return nil
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
