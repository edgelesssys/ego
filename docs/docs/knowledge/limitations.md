# Known limitations

Most Go apps can be compiled and run with EGo without modifications. However, there are some limitations:

* An EGo app is a single process. Spawning other processes isn't possible.
* A nil pointer dereference aborts the process without calling deferred functions and without printing an error message.
* cgo support is experimental
  * Libraries must be statically linked. Shared objects are unsupported.
  * Libraries must be compiled with `-fPIC`
* Stripped executables (e.g., `ego-go build -ldflags -s`) are unsupported

## (Partially) unsupported packages

These usually compile, but return an error at runtime.

* `os`: `Pipe` and `StartProcess` are unsupported
* `os/exec`: spawning processes is unsupported

## cgo: Unsupported libc functions

Using these functions causes a build, sign, or runtime error.

* `exec` family
* `fork`
* `pipe`
* `posix_spawn`
* `mmap`: partially supported; memory-mapped files are unsupported

## Different behavior

### `statfs`

`statfs` returns dummy values, regardless of the path.
This ensures that existing code that unconditionally trusts the call's result can be used securely.

You can request the values from the host by prefixing the path with `/edg/hostfs`, e.g.

* `statfs("/edg/hostfs/", &statbuf)` requests info for host path `/`
* `statfs("/edg/hostfs/foo/bar", &statbuf)` requests info for host path `/foo/bar`

Note that this ignores [configured mounts](../reference/config.md#mounts).

Be aware that the returned result isn't trustworthy.
You must check the values before taking actions on them.
