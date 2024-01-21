package main

// The wasmtime library links a few symbols that are not available in EGo, but it's sufficient to provide stub implementations of them.

/*
#cgo LDFLAGS: -static-libgcc
int gnu_get_libc_version() { return 0; }
int __register_atfork() { return 0; }
int __res_init() { return -1; }

// TODO remove with next EGo release that supports this syscall
#include <time.h>
int clock_getres(int clockid, struct timespec* res) {
	res->tv_sec = 0;
	res->tv_nsec = 4000000;
	return 0;
}
*/
import "C"
