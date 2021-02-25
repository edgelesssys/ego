# cgo sample
This sample demonstrates the experimental cgo support.

The sample can be built and run as follows:
```sh
sudo apt install zlib1g-dev
ego-go build
ego sign cgo
ego run cgo
```

Make sure that you link libraries statically, e.g., use `-l:libz.a` instead of `-lz`. However, libc must be linked dynamically.

You can check with `ldd`:
```sh
$ ldd cgo
linux-vdso.so.1
libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0
libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
/lib64/ld-linux-x86-64.so.2
```
Here we only have libc and friends, so this is fine.
