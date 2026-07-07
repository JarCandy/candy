package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func resolvePluginPath() string {
	candidates := []string{"plugin.wasm", filepath.Join("cmd", "plugin.wasm")}

	if wd, err := os.Getwd(); err == nil {
		for _, candidate := range candidates {
			fullPath := filepath.Join(wd, candidate)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
		}
	}

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, candidate := range []string{filepath.Join(dir, "plugin.wasm"), filepath.Join(dir, "cmd", "plugin.wasm")} {
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}

	return filepath.Join("cmd", "plugin.wasm")
}

func main() {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	// Инициализируем WASI для совместимости с TinyGo
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		log.Fatal(err)
	}

	// Загружаем скомпилированный плагин
	pluginPath := resolvePluginPath()
	fmt.Printf("Using plugin: %s\n", pluginPath)

	wasmBytes, err := os.ReadFile(pluginPath)
	if err != nil {
		log.Fatal(err)
	}

	compiled, err := r.CompileModule(ctx, wasmBytes)
	if err != nil {
		log.Fatal(err)
	}

	mod, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer mod.Close(ctx)

	reqData := map[string]string{"name": "world"}
	jsonReq, err := json.Marshal(reqData)
	if err != nil {
		log.Fatal(err)
	}
	reqSize := uint64(len(jsonReq))

	// Changed from "allocate" to "alloc"
	allocFunc := mod.ExportedFunction("alloc")
	if allocFunc == nil {
		log.Fatal("Function 'alloc' not exported")
	}

	results, err := allocFunc.Call(ctx, reqSize)
	if err != nil {
		log.Fatal(err)
	}
	wasmBufPtr := uint32(results[0])

	if !mod.Memory().Write(wasmBufPtr, jsonReq) {
		log.Fatal("Failed to write to memory")
	}

	helloFunc := mod.ExportedFunction("SayHello")
	if helloFunc == nil {
		log.Fatal("Function 'SayHello' not exported")
	}

	results, err = helloFunc.Call(ctx, uint64(wasmBufPtr), reqSize)
	if err != nil {
		log.Fatal(err)
	}

	packedResult := results[0]
	resPtr := uint32(packedResult >> 32)
	resSize := uint32(packedResult & 0xFFFFFFFF)

	jsonResBytes, ok := mod.Memory().Read(resPtr, resSize)
	if !ok {
		log.Fatal("Failed to read from memory")
	}

	var finalResponse map[string]string
	if err := json.Unmarshal(jsonResBytes, &finalResponse); err != nil {
		log.Fatal(err)
	}

	fmt.Println(finalResponse["message"])

}
