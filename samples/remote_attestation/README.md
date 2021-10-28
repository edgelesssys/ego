# Go remote attestation sample
This sample shows how to do remote attestation of a EGo enclave. It consists of a server running in the enclave and a client that attests the server before sending a secret.

**Note: This sample only works on SGX-FLC systems with a [quote provider](https://docs.edgeless.systems/ego/#/reference/attest) installed.**

The server generates a self-signed certificate and a report for remote attestation using [GetRemoteReport()](https://pkg.go.dev/github.com/edgelesssys/ego/enclave#GetRemoteReport) that includes the certificate's hash. It thereby binds the certificate to the enclave's identity.

The server runs HTTPS and serves the following:
* `/cert` makes the TLS certificate explicitly available on the HTTP layer (for simplicity for this sample).
* `/report` returns the remote report.
* `/secret` receives the secret via a query parameter named `s`.

The client first gets the certificate and the report, skipping TLS certificate verification. Then it verifies the certificate using the report via [VerifyRemoteReport()](https://pkg.go.dev/github.com/edgelesssys/ego/eclient#VerifyRemoteReport). From there on it can establish a secure connection to the enclave server and send its secret.

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
