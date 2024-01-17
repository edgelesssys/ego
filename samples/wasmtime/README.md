# Wasmtime sample

This sample shows how to run WebAssembly inside EGo using [Wasmtime](https://pkg.go.dev/github.com/bytecodealliance/wasmtime-go).

By default, *wasmtime-go* comes with a shared library. EGo only supports static linking. To this end, download the wasmtime static library and tell the Go compiler to use it:
```sh
mkdir wasmtime
wget -O- https://github.com/bytecodealliance/wasmtime/releases/download/v15.0.0/wasmtime-v15.0.0-x86_64-linux-c-api.tar.xz | tar xvJf - -C ./wasmtime --strip-components=1
CGO_CFLAGS="-I$PWD/wasmtime/include" CGO_LDFLAGS="$PWD/wasmtime/lib/libwasmtime.a -ldl -lm -static-libgcc" ego-go build
```

Then you can sign and run as usual:
```sh
ego sign wasmtime_sample
ego run wasmtime_sample
```

You should see an output similar to:
```
[erthost] loading enclave ...
[erthost] entering enclave ...
[ego] starting application ...
Results of `sum`: 3
```

Note that `executableHeap` is enabled in `enclave.json` so that Wasmtime can JIT-compile the WebAssembly.
