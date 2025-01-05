package main

import (
	"log"
)

func main() {
	wasmbin := createWasmBinary(readWatFile("./add.wat"))
	err := validateWASMHeader(wasmbin)
	if err != nil {
		log.Fatal(err)
	}

	err = parseWASMCodeSection(wasmbin)
	if err != nil {
		log.Fatal(err)
	}
}
