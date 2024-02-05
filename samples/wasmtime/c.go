package main

// The wasmtime library links a few symbols that are not available in EGo, but it's sufficient to provide stub implementations of them.

/*
#cgo LDFLAGS: -static-libgcc
int gnu_get_libc_version() { return 0; }
int __register_atfork() { return 0; }
*/
import "C"
