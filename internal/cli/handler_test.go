package cli

import (
	"testing"

	"github.com/rp1s/digreyt/translate"
)

func TestParseGlobalArgsSetsCurrentLanguage(t *testing.T) {
	prev := CurrentLanguage
	t.Cleanup(func() {
		SetLanguage(prev)
	})

	args, showHelp, err := parseGlobalArgs([]string{"--lang", "ru", "build", "main.cm"})
	if err != nil {
		t.Fatal(err)
	}
	if showHelp {
		t.Fatal("expected command execution, got help")
	}
	if CurrentLanguage != "ru" {
		t.Fatalf("expected current language ru, got %q", CurrentLanguage)
	}
	if translate.Language() != "ru" {
		t.Fatalf("expected translate language ru, got %q", translate.Language())
	}
	if len(args) != 2 || args[0] != "build" || args[1] != "main.cm" {
		t.Fatalf("expected cleaned command args, got %#v", args)
	}
}

func TestParseGlobalArgsSupportsInlineLanguageAndHelp(t *testing.T) {
	prev := CurrentLanguage
	t.Cleanup(func() {
		SetLanguage(prev)
	})

	args, showHelp, err := parseGlobalArgs([]string{"--lang=ru", "--help"})
	if err != nil {
		t.Fatal(err)
	}
	if !showHelp {
		t.Fatal("expected help flag")
	}
	if CurrentLanguage != "ru" {
		t.Fatalf("expected current language ru, got %q", CurrentLanguage)
	}
	if len(args) != 0 {
		t.Fatalf("expected no command args, got %#v", args)
	}
}

func TestParseGlobalArgsRequiresLanguageValue(t *testing.T) {
	_, _, err := parseGlobalArgs([]string{"--lang"})
	if err == nil {
		t.Fatal("expected missing language error")
	}
}
