# Build your app with EGo

## `ego-go` replaces `go`

EGo provides a simple way to build your Go project: `ego-go`, an adapted Go compiler. It builds enclave-compatible executables from a given Go project—while providing the same CLI as the original Go compiler. You can build your app with

```bash
ego-go build
```

## Sign and run

After you've built your app, sign it with the [`ego sign`](../reference/cli.md#ego-sign) command. Just insert the name of your binary in the following command:

```bash
ego sign <executable>
```

Run the executable with

```bash
ego run <executable>
```

You can set the `OE_SIMULATION=1` environment variable if you want to simulate the enclave, e.g. for development on hardware that doesn't support enclaves.

Look at the [example collection](../getting-started/examples.md) if you want to see the build process in action.

## FIPS 140 cryptography

EGo optionally supports FIPS 140-compliant cryptography.

Build your app with the Go FIPS 140 module enabled:

```bash
GOFIPS140=latest ego-go build
```

Then sign and run as usual.
EGo automatically uses the corresponding FIPS 140-enabled runtime.
