package main

import (
	"encoding/json"
	"unsafe"
)

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

func main() {}

//export alloc
func alloc(size uint32) *byte {
	buf := make([]byte, size)
	return &buf[0]
}

//export free
func free(ptr *byte) {
}

//export SayHello
func SayHello(ptr *byte, size uint32) uint64 {
	input := unsafe.Slice(ptr, size)

	var req Request
	if err := json.Unmarshal(input, &req); err != nil {
		return 0
	}

	resp := Response{
		Message: "Hello, " + req.Name + "!",
	}

	output, err := json.Marshal(resp)
	if err != nil {
		return 0
	}

	// Выделяем память
	outPtr := alloc(uint32(len(output)))
	copy(unsafe.Slice(outPtr, len(output)), output)
	return (uint64(uintptr(unsafe.Pointer(outPtr))) << 32) | uint64(len(output))
}
