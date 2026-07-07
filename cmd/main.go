package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func main() {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	// Инициализируем WASI
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		log.Fatal(err)
	}

	// Загружаем скомпилированный плагин
	wasmBytes, err := os.ReadFile("plugin.wasm")
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

	allocFunc := mod.ExportedFunction("allocate")
	if allocFunc == nil {
		log.Fatal("Function 'allocate' not exported")
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
