# Configuration file

Your enclave's configuration is defined in JSON and applied when signing an executable via `ego sign`.

Here's an example configuration:

```json
{
    "exe": "helloworld",
    "key": "private.pem",
    "debug": true,
    "heapSize": 512,
    "executableHeap": false,
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
    ],
    "env": [
        {
            "name": "LANG",
            "fromHost": true
        },
        {
            "name": "PWD",
            "value": "/data"
        }
    ],
    "files": [
        {
            "source": "some_datafile",
            "target": "/some/path/to/datafile"
        },
        {
            "source": "/etc/ssl/certs/ca-certificates.crt",
            "target": "/etc/ssl/certs/ca-certificates.crt"
        }
    ]
}
```

## Basic settings

`exe` is the (relative or absolute) path to the executable that should be signed.

`key` is the path to the private RSA key of the signer. When invoking `ego sign` and the key file doesn't exist, a key with the required parameters is automatically generated. You can also generate it yourself with:

```bash
openssl genrsa -out private.pem -3 3072
```

If `debug` is true, the enclave can be inspected with a debugger.

`heapSize` specifies the heap size available to the enclave in MB. It should be at least 512 MB.

If `executableHeap` is true, the enclave heap will be executable. This is required if you use libraries that JIT-compile code.

`productID` is assigned by the developer and enables the attester to distinguish between different enclaves signed with the same key.

`securityVersion` should be incremented by the developer whenever a security fix is made to the enclave code.

## Mounts

`mounts` define custom mount points that apply to the file system presented to the enclave. This can be omitted if no mounts other than the default mounts should be performed, or you can define multiple entries with the following parameters:

* `source` (required for `hostfs`): The directory from the host file system that should be mounted in the enclave when using `hostfs`. If this is a relative path, it will be relative to the working directory of the ego host process. For `memfs`, this value will be ignored and can be omitted.
* `target` (required): Defines the mount path in the enclave.
* `type` (required): Either `hostfs` if you want to mount a path from the host's file system in the enclave, or `memfs` if you want to use a temporary file system similar to *tmpfs* on UNIX systems, with your data stored in the secure memory environment of the enclave.
* `readOnly`: Can be `true` or `false` depending on if you want to mount the path as read-only or read-write. When omitted, will default to read-write.

By default, `/` is initialized as an empty `memfs` file system. To expose certain directories to the enclave, you can use the `hostfs` mounts with the options mentioned above. You can also choose to define additional `memfs` mount points, but note that there is no explicit isolation between them. They can be accessed either via the path specified in `target` or also via `/edg/mnt/<target>`, which is where the files of the additional `memfs` mount are stored internally.

:::caution

It's not recommended to use the mount options to remount `/` as `hostfs` other than for testing purposes. You might involuntarily expose files to the host which should stay inside the enclave, risking the confidentiality of your data.

:::

## Environment variables

To protect against manipulations by the untrusted host, all environment variables not starting with `EDG_` are dropped when entering the enclave.

To securely provide environment variables to your application (not starting with `EDG_`), define them in the `env` section. You can either set a static value or whitelist a variable to be taken over from the host.

* `name` (required): The name of the environment variable
* `value` (required if not `fromHost`): The value of the environment variable
* `fromHost`: When set to `true`, the current value of the requested environment variable will be copied over if it exists on the host. If the host doesn't hold this variable, it will either fall back to the value set in `value` (if it exists), or won't be created at all.

A special environment variable is `PWD`. Depending on the mount options you have set in your configuration, you can set the initial working directory of your enclave by specifying your desired path as a value for `PWD`. Note that this directory needs to exist in the context of the enclave, not your host file system.

## Embedded files

`files` specifies files that should be embedded into the enclave. Embedded files are included in the enclave measurement and thus can't be manipulated. At runtime they're accessible via the in-enclave-memory filesystem.

`source` is the path to the file that should be embedded. `target` Is the path within the in-enclave-memory filesystem where the file will reside at runtime.

A common use case is to embed CA certificates so that an app can make secure TLS connections from inside the enclave.

## Advanced users: Tweak underlying enclave configuration

:::warning

The EGo enclave configuration described above covers all settings relevant for most users.

Changing the following settings can negatively impact the stability of your app.

:::

<details>
<summary>Open Enclave configuration file</summary>

EGo is based on Open Enclave.
You can apply your own Open Enclave configuration as follows:

1. Create a file `enclave.conf`. Start with the following settings:

   ```
   Debug=1
   NumHeapPages=131072
   NumStackPages=1024
   NumTCS=32
   ProductID=1
   SecurityVersion=1
   ```

2. Adapt the configuration as needed. See the [Open Enclave documentation](https://github.com/openenclave/openenclave/blob/v0.19.x/docs/GettingStartedDocs/buildandsign.md#signing-an-sgx-enclave) for details.

3. Sign your app with `ego sign`

4. Sign your app with `ego-oesign`:

   ```bash
   /opt/ego/bin/ego-oesign sign -e /opt/ego/share/ego-enclave -c enclave.conf -k private.pem --payload helloworld
   ```

</details>
