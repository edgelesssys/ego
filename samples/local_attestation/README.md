# Local attestation sample

This sample shows how to do local attestation between EGo enclaves.
Local attestation doesn't require additional infrastructure to be set up, but it's limited to enclaves running on the same host.

**Note: This sample doesn't work in simulation mode.**

The sample consists of a server and a client, both running in enclaves.
The enclaves first exchange certificates and reports over an insecure channel, verify these objects, and then establish a mutal TLS connection.

Unlike remote reports, local reports can only be verified by a predetermined target enclave.
This works as follows:

1. The verifier creates a report: `targetReport = GetLocalReport(nil, nil)`
2. The verifier sends the `targetReport` to the attester
3. The attester creates a targeted report: `report = GetLocalReport(someData, targetReport)`
4. The attester sends the `report` to the verifier
5. The verifier verifies the `report`: `VerifyLocalReport(report)`

## Implementation details

The server creates an unencrypted HTTP endpoint (the *attest server*) to serve the following:

* `/cert` returns the certificate of the *secure server*
* `/report` returns a local report for the certificate
* `/client` checks the client's report and returns a certificate for the client's public key

The server creates an encrypted, mutually authenticated HTTPS endpoint (the *secure server*) to serve the following:

* `/ping` replies with `pong`

The *secure server* only accepts clients that can provide a certificate obtained by `/client`.
In other words, the server only accepts attested clients.

The client uses the *attest server* to get and verify the server's certificate and to obtain a client certificate.
The client then uses these certificates to establish a mutual TLS connection to the *secure server*.

Some error handling in this sample is omitted for brevity.

## Usage

Build the enclaves:

```sh
cd server
ego-go build
ego sign
cd ../client
ego-go build
ego sign
cd ..
```

Run the server:

```
$ ego run server/server

...
[ego] starting application ...
listening ...
```

In another terminal, run the client:

```
$ ego run client/client

...
[ego] starting application ...
GET http://localhost:8080/cert
GET http://localhost:8080/report
GET http://localhost:8080/client
GET https://localhost:8081/ping
server responded: pong
```
