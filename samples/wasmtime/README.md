# Wasmtime sample

This sample shows how to run WebAssembly inside EGo using [Wasmtime](https://pkg.go.dev/github.com/bytecodealliance/wasmtime-go).

By default, *wasmtime-go* comes with a library that makes direct syscalls.
EGo only supports syscall via libc.
To this end, build the wasmtime library with libc backend:

```sh
git clone -bv15.0.1 --depth=1 https://github.com/bytecodealliance/wasmtime
cd wasmtime
git submodule update --init --depth=1
RUSTFLAGS=--cfg=rustix_use_libc cargo build --release -p wasmtime-c-api
cd ..
```

Tell the Go compiler to use it:

```sh
CGO_LDFLAGS=wasmtime/target/release/libwasmtime.a ego-go build
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
