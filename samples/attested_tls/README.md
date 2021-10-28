# Go attested TLS sample
This sample shows how to establish a TLS connection to an EGo enclave that is transparently attested. It consists of a server running in the enclave and a client that sends a secret.

**Note: This sample only works on SGX-FLC systems with a [quote provider](https://docs.edgeless.systems/ego/#/reference/attest) installed.**

The server creates a `tls.Config` object using [CreateAttestationServerTLSConfig()](https://pkg.go.dev/github.com/edgelesssys/ego/enclave#CreateAttestationServerTLSConfig) that can then be used to create a server.

The server runs HTTPS and serves the following:
* `/secret` receives the secret via a query parameter named `s`.

The client creates a `tls.Config` object using [CreateAttestationClientTLSConfig()](https://pkg.go.dev/github.com/edgelesssys/ego/eclient#CreateAttestationClientTLSConfig). In a callback function properties of the remote report are checked. The validity of the certificate is automatically checked by the `tls.Config`. The client uses this config to send its secret via an `http.Client`.

Some error handling in this sample is omitted for brevity.

The server can be built and run as follows:
```sh
ego-go build
ego sign server
ego run server
```

The client can be built either using `ego-go` or a recent Go compiler:
```sh
CGO_CFLAGS=-I/opt/ego/include CGO_LDFLAGS=-L/opt/ego/lib go build ra_client/client.go
```
Or if using the EGo snap:
```sh
EGOPATH=/snap/ego-dev/current/opt/ego CGO_CFLAGS=-I$EGOPATH/include CGO_LDFLAGS=-L$EGOPATH/lib go build ra_client/client.go
```

The client expects the `signer ID` (`MRSIGNER`) as an argument. The `signer ID` can be derived from the signer's public key using `ego signerid`:
```sh
./client -s `ego signerid public.pem`
```
