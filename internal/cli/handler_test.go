package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	diagnostics "github.com/rp1s/digreyt"
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

func TestHandlerCmdReturnsUnknownCommandError(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"caramel", "missing"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	err := HandlerCmd()
	if err == nil || !strings.Contains(err.Error(), `unknown command "missing"`) {
		t.Fatalf("HandlerCmd() error = %v", err)
	}
}

func TestBuildPropagatesFileError(t *testing.T) {
	err := Build([]string{filepath.Join(t.TempDir(), "missing.cm")})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Build() error = %v, want os.ErrNotExist", err)
	}
}

func TestBuildPropagatesParserDiagnostics(t *testing.T) {
	err := Build([]string{writeSource(t, `package(`)})
	assertDiagnosticError(t, err)
}

func TestBuildPropagatesAnalyzerDiagnostics(t *testing.T) {
	err := Build([]string{writeSource(t, `package("main")
let age: string = 10
`)})
	assertDiagnosticError(t, err)
}

func writeSource(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.cm")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertDiagnosticError(t *testing.T, err error) {
	t.Helper()
	var arena *diagnostics.Arena
	if !errors.As(err, &arena) {
		t.Fatalf("error = %T %v, want *digreyt.Arena", err, err)
	}
	if !arena.HasFatalErrors() {
		t.Fatal("expected fatal diagnostics")
	}
}
