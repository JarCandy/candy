package plugin

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caramelang/caramel/pkg/branding"
)

func TestLoadPluginReadsWasmIntoMemory(t *testing.T) {
	for _, relativePath := range []string{
		branding.PluginWasmFileName,
		filepath.Join(branding.PluginOutputDirectory, branding.PluginWasmFileName),
	} {
		t.Run(relativePath, func(t *testing.T) {
			directory := t.TempDir()
			writeTestInterface(t, directory)
			wasmPath := filepath.Join(directory, relativePath)
			if err := os.MkdirAll(filepath.Dir(wasmPath), 0o755); err != nil {
				t.Fatal(err)
			}
			wasm := append(append([]byte(nil), wasmV1Header...), 0x00)
			if err := os.WriteFile(wasmPath, wasm, 0o600); err != nil {
				t.Fatal(err)
			}

			loaded, err := LoadPlugin(LoadPluginArgs{Path: &directory})
			if err != nil {
				t.Fatalf("LoadPlugin() error = %v", err)
			}
			if loaded.PathPluginSave != directory {
				t.Errorf("PathPluginSave = %q, want %q", loaded.PathPluginSave, directory)
			}
			if loaded.PubInfo.Name != "example" || loaded.PubInfo.Author != "caramel" || loaded.PubInfo.Version != "v1.2.3" {
				t.Errorf("PubInfo = %+v, want parsed interface.json metadata", loaded.PubInfo)
			}
			if !bytes.Equal(loaded.WasmBin, wasm) {
				t.Errorf("WasmBin = %v, want %v", loaded.WasmBin, wasm)
			}
			if loaded.IsGlobal {
				t.Error("local plugin marked as global")
			}
		})
	}
}

func TestLoadPluginRejectsMissingWasm(t *testing.T) {
	directory := t.TempDir()
	writeTestInterface(t, directory)
	_, err := LoadPlugin(LoadPluginArgs{Path: &directory})
	if err == nil || !strings.Contains(err.Error(), "plugin.wasm not found") {
		t.Fatalf("LoadPlugin() error = %v, want plugin.wasm not found", err)
	}
}

func TestLoadPluginRejectsInvalidWasm(t *testing.T) {
	directory := t.TempDir()
	writeTestInterface(t, directory)
	if err := os.WriteFile(filepath.Join(directory, branding.PluginWasmFileName), []byte("invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadPlugin(LoadPluginArgs{Path: &directory})
	if err == nil || !strings.Contains(err.Error(), "not a valid WebAssembly module") {
		t.Fatalf("LoadPlugin() error = %v, want invalid WebAssembly error", err)
	}
}

func TestLoadPluginRejectsInvalidInterface(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, branding.PluginInfoFileName), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadPlugin(LoadPluginArgs{Path: &directory})
	if err == nil || !strings.Contains(err.Error(), "parse") {
		t.Fatalf("LoadPlugin() error = %v, want interface parse error", err)
	}
}

func writeTestInterface(t *testing.T, directory string) {
	t.Helper()
	content := []byte(`{"type":1,"name":"example","author":"caramel","version":"v1.2.3"}`)
	if err := os.WriteFile(filepath.Join(directory, branding.PluginInfoFileName), content, 0o600); err != nil {
		t.Fatal(err)
	}
}
