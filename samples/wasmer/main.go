package main

import (
	"fmt"

	"github.com/wasmerio/wasmer-go/wasmer"
)

// https://pkg.go.dev/github.com/wasmerio/wasmer-go@v1.0.4/wasmer#hdr-Examples
func main() {
	// Let's assume we don't have WebAssembly bytes at hand. We
	// will write WebAssembly manually.
	wasmBytes := []byte(`
	(module
	  (type (func (param i32 i32) (result i32)))
	  (func (type 0)
	    local.get 0
	    local.get 1
	    i32.add)
	  (export "sum" (func 0)))
`)

	// Create an Engine
	engine := wasmer.NewEngine()

	// Create a Store
	store := wasmer.NewStore(engine)

	// Let's compile the module.
	module, err := wasmer.NewModule(store, wasmBytes)

	if err != nil {
		fmt.Println("Failed to compile module:", err)
	}

	// Create an empty import object.
	importObject := wasmer.NewImportObject()

	// Let's instantiate the WebAssembly module.
	instance, err := wasmer.NewInstance(module, importObject)

	if err != nil {
		panic(fmt.Sprintln("Failed to instantiate the module:", err))
	}

	// Now let's execute the `sum` function.
	sum, err := instance.Exports.GetFunction("sum")

	if err != nil {
		panic(fmt.Sprintln("Failed to get the `add_one` function:", err))
	}

	result, err := sum(1, 2)

	if err != nil {
		panic(fmt.Sprintln("Failed to call the `add_one` function:", err))
	}

	fmt.Println("Results of `sum`:", result)
}
