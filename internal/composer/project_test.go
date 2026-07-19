package composer

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (self roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return self(request)
}

type closeErrorBody struct {
	io.Reader
	err error
}

func (self closeErrorBody) Close() error {
	return self.err
}

func TestLoadResolvesBuildPathFromWorkingDirectory(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
	})

	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}

	sourcePath := filepath.Join(srcDir, "user.cm")
	if err := os.WriteFile(sourcePath, []byte(`package("main")`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	workPath, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	expectedSourcePath := filepath.Join(workPath, "src", "user.cm")

	project, err := Load(filepath.Join("src", "user.cm"), "")
	if err != nil {
		t.Fatal(err)
	}

	if project.Name != "user" {
		t.Fatalf("expected default project name user, got %q", project.Name)
	}
	if project.WorkPath != workPath {
		t.Fatalf("expected work path %q, got %q", workPath, project.WorkPath)
	}
	if len(project.AstFile) != 1 {
		t.Fatalf("expected one AST file, got %d", len(project.AstFile))
	}
	if project.AstFile[0].Path != expectedSourcePath {
		t.Fatalf("expected source path %q, got %q", expectedSourcePath, project.AstFile[0].Path)
	}
	if project.AstFile[0].FileName != "user.cm" {
		t.Fatalf("expected file name user.cm, got %q", project.AstFile[0].FileName)
	}
	if len(project.AstFile[0].Ast.Decls) == 0 {
		t.Fatal("expected parsed declarations")
	}
}

func TestLoadKeepsExplicitProjectName(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
	})

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "main.cm")
	if err := os.WriteFile(sourcePath, []byte(`package("main")`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	project, err := Load("main.cm", "custom")
	if err != nil {
		t.Fatal(err)
	}

	if project.Name != "custom" {
		t.Fatalf("expected explicit project name custom, got %q", project.Name)
	}
}

func TestDownloadProjectPluginPropagatesResponseCloseError(t *testing.T) {
	want := errors.New("response close failed")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       closeErrorBody{Reader: strings.NewReader("plugin"), err: want},
			Header:     make(http.Header),
		}, nil
	})}

	err := downloadProjectPlugin(client, "https://example.test/plugin", filepath.Join(t.TempDir(), "plugin.wasm"))
	if !errors.Is(err, want) {
		t.Fatalf("downloadProjectPlugin() error = %v, want %v", err, want)
	}
}
