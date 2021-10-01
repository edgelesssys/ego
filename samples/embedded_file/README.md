# Embedded file sample
This sample shows how to embed a file into an EGo enclave.

Embedded files are included in the enclave measurement and thus can't be manipulated. At runtime they are accessible via the in-enclave-memory filesystem.

Specify files to embed in the enclave configuration file:
```js
    "files": [
        {
            "source": "/etc/ssl/certs/ca-certificates.crt",
            "target": "/etc/ssl/certs/ca-certificates.crt"
        }
    ]
```
`source` is the path to the file that should be embedded. `target` Is the path within the in-enclave-memory filesystem where the file will reside at runtime. Actual embedding into the enclave binary happens on signing.

In this sample, we chose `/etc/ssl/certs/ca-certificates.crt`, which contains the certificates of common CA's and allows us to make secure TLS connections from inside the enclave. Remember that we can't trust the certificates provided by the host that will run the enclave.

The sample can be built and run as follows:
```sh
ego-go build
ego sign
ego run embedded_file
```

You should see an output similar to:
```
[erthost] loading enclave ...
[erthost] entering enclave ...
[ego] starting application ...
Getting https://www.edgeless.systems/
200 OK
```
