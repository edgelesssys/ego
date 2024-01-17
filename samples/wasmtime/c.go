package main

/*
#cgo LDFLAGS: -static-libgcc
int gnu_get_libc_version() { return 0; }
int __register_atfork() { return 0; }
int __res_init() { return -1; }
*/
import "C"
