# TLS inside the enclave

Accepting or establishing TLS connections inside the enclave has peculiarities you should be aware of.

## Hosting a TLS server

You need to consider how to securely provide the enclave with a private key.
A common approach is to let the enclave generate the key itself and bind it to a remote attestation statement.
In practice, this means that the enclave includes the hash of the public key in the attestation report.
See the [remote attestation sample](https://github.com/edgelesssys/ego/tree/master/samples/remote_attestation) on how to do this.

EGo offers an API that simplifies this common pattern.
See the [attested TLS sample](https://github.com/edgelesssys/ego/tree/master/samples/attested_tls) for details.

## Connecting to a TLS server

A client needs a set of root certificates to verify a TLS server.
The enclave can't trust the host's root certificates.
Thus, you need to consider how to securely provide the enclave with these certificates.

### Embed root certificates

You can embed root certificates in the enclave binary.
To embed the certificates of your trusted development/build machine, add to [enclave.json](../reference/config.md#embedded-files):

```json
    "files": [
        {
            "source": "/etc/ssl/certs/ca-certificates.crt",
            "target": "/etc/ssl/certs/ca-certificates.crt"
        }
    ]
```

The [embedded files sample](https://github.com/edgelesssys/ego/tree/master/samples/embedded_file) demonstrates this use case.

### Skip verification

There may be cases where you don't need trust in the server:

* You don't send secret data to the server
* and you don't need trust in the replied data or you verify the data by other means.

Then you can skip verification:

```go
tlsConfig := &tls.Config{InsecureSkipVerify: true}
client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
resp, err := client.Get("https://...")
```
