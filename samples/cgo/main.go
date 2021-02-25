package main

// #cgo LDFLAGS: -l:libz.a
// #include <zlib.h>
import "C"

import "fmt"

func main() {
	fmt.Println(C.GoString(C.zlibVersion()))
}
