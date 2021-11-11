package main

// The wasmer library links a few symbols that are not available in EGo, but it's sufficient to provide stub implementations of them.

/*
int __longjmp_chk() { return 0; }
int __register_atfork() { return -1; }
int __res_init() { return -1; }
int __sigsetjmp() { return 0; }
*/
import "C"
