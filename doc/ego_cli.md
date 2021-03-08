# ego command reference
`ego` is a tool for managing EGo enclaves.

Usage:
```
ego <command> [arguments]
```

Commands:
```
sign        Sign an executable built with ego-go.
run         Run a signed executable in standalone mode.
marblerun   Run a signed executable as a Marblerun Marble.
signerid    Print the SignerID of a signed executable.
uniqueid    Print the UniqueID of a signed executable.
env         Run a command in the EGo environment.
```

## sign
Usage:
```
ego sign [executable | config.json]
```
Sign an executable built with ego-go. Executables must be signed before they can be run in an enclave.

This command can be used in different modes:
* `ego sign <executable>`\
  Generates a new key `private.pem` and a default configuration `enclave.json` in the current directory and signs the executable.

* `ego sign`\
  Searches in the current directory for `enclave.json` and signs the therein provided executable.

* `ego sign <config.json>`\
  Signs an executable according to a given configuration.

See [this section](#enclave-configuration-file) for more information on the configuration file.

## run
Usage:
```
ego run <executable> [args...]
```
Run a signed executable in an enclave. You can pass arbitrary arguments to the enclave.

Environment variables are only readable from within the enclave if they start with "EDG_".

You need an SGX-enabled machine to run an enclave. For development, you can also enable simulation mode by setting OE_SIMULATION=1, e.g.:
```
OE_SIMULATION=1 ego run helloworld
```

## marblerun
Usage:
```
ego marblerun <executable>
```
Run a signed executable as a Marblerun Marble.
Requires a running Marblerun Coordinator instance.

Environment variables are only readable from within the enclave if they start with "EDG_" and will be extended/overwritten with the ones specified in the manifest.

Requires the following configuration environment variables:
* EDG_MARBLE_COORDINATOR_ADDR\
  The Coordinator address
* EDG_MARBLE_TYPE\
  The type of this Marble (as specified in the manifest)
* EDG_MARBLE_DNS_NAMES\
  The alternative DNS names for this Marble's TLS certificate
* EDG_MARBLE_UUID_FILE\
  The location where this Marble will store its UUID

Set OE_SIMULATION=1 to run in simulation mode.

## signerid
Usage:
```
ego signerid <executable | key.pem>
```
Print the SignerID either from a signed executable or by reading a keyfile.`

## uniqueid
Usage:
```
ego uniqueid <executable>
```
Print the UniqueID of a signed executable.

## env
Usage:
```
ego env ...
```
Run a command within the ego build environment. For example, run
```
ego env make
```
to build a Go project that uses a Makefile.

## Enclave configuration file
An enclave configuration is defined in JSON and applied when signing an executable.

Here is an example configuration:
```json
{
    "exe": "helloworld",
    "key": "private.pem",
    "debug": true,
    "heapSize": 512,
    "productID": 1,
    "securityVersion": 1,
    "mounts": [
        {
            "source": "/home/user",
            "target": "/data",
            "type": "hostfs",
            "readOnly": false
        },
        {
            "target": "/tmp",
            "type": "memfs"
        }
    ]
}
```

`exe` is the (relative or absolute) path to the executable that should be signed.

`key` is the path to the private RSA key of the signer. When invoking `ego sign` and the key file does not exist, a key with the required parameters is automatically generated. You can also generate it yourself with:
```
openssl genrsa -out private.pem -3 3072
```

If `debug` is true, the enclave will be debuggable.

`heapSize` specifies the heap size available to the enclave in MB. It should be at least 512 MB.

A `productID` (SGX: ISVPRODID) is assigned by the developer and enables the attester to distinguish between different enclaves signed with the same key.

The developer should increment the `securityVersion` (SGX: ISVSVN) whenever a security fix is made to the enclave code.

`mounts` defines custom mount points which apply to the file system presented to the enclave. This can be `null` if no mounts other than the default mounts should be performed, or you can define multiple entries with the following parameters:

  * `source` (required for `hostfs`): The directory from host file system which should be mounted in the enclave when using `hostfs`. For `memfs`, this value will be ignored and can be omitted.
  * `target` (required): Defines the mount path in the enclave.
  * `type` (required): Either `hostfs` if you want to mount a path from the host's file system in the enclave, or `memfs` if you want to use a temporary file system similar to *tmpfs* on UNIX systems.
  * `readOnly`: Can be `true` or `false` depending on if you want to mount the path as read-only or read-write. When omitted, will default to read-write.
