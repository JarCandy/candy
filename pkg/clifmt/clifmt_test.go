package clifmt

import (
	"regexp"
	"strings"
	"testing"
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

func stripANSI(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}
