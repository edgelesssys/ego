package main

import (
	"fmt"

	"github.com/bytecodealliance/wasmtime-go/v15"
)

// https://pkg.go.dev/github.com/bytecodealliance/wasmtime-go#example-Config-Fuel
func main() {
	// Let's assume we don't have WebAssembly bytes at hand. We
	// will write WebAssembly manually.
	wasmBytes, err := wasmtime.Wat2Wasm(`
	(module
	  (type (func (param i32 i32) (result i32)))
	  (func (type 0)
	    local.get 0
	    local.get 1
	    i32.add)
	  (export "sum" (func 0)))
`)
	if err != nil {
		fmt.Println("Failed to parse WebAssembly code:", err)
		return
	}

	engine := wasmtime.NewEngine()
	
	module, err := wasmtime.NewModule(engine, wasmBytes)
	if err != nil {
		fmt.Println("failed to compile module ",err)
	}

	linker := wasmtime.NewLinker(engine)
	err = linker.DefineWasi()
	if err != nil {
		fmt.Println("failed to define wasi ",err)
	}

	wasiConfig := wasmtime.NewWasiConfig()

	store := wasmtime.NewStore(engine)
	store.SetWasi(wasiConfig)

	instance,err := linker.Instantiate(store, module)
	if err != nil {
		fmt.Println("failed to instantiate ",err)
	}

	sumFunc := instance.GetFunc(store,"sum")
	if sumFunc != nil {
		result ,err := sumFunc.Call(store,1,2)
		if err != nil {
			fmt.Println("failed to call sum ",err)
		}
		fmt.Println("Results of `sum`:", result)
	}
}
