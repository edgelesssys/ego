# Examples 🧪

Just installed EGo? Here are some examples to start your confidential development process.

## [Hello world](https://github.com/edgelesssys/ego/blob/master/samples/helloworld)

An enclave saying "hello" to the world; works with any CPU, no special hardware required.

## [Attested HTTPS server](https://github.com/edgelesssys/ego/tree/master/samples/attested_tls)

A minimal HTTPS server running in an enclave. The server uses the EGo library to create a `tls.Config`, which automatically obtains and sends a remote-attestation statement to a client. Consequently, the client knows that the server is indeed running in an enclave and POSTs its secret. This example showcases the easy-to-use yet powerful attestation features of the EGo library.

## [Attested HTTPS server (manually)](https://github.com/edgelesssys/ego/blob/master/samples/remote_attestation)

Similar to the above, but the server manages remote attestation by itself. This example showcases the raw attestation features of the EGo library. Use this as a starting point if you want to use an existing (HTTPS) client application and thus need to perform attestation separately.

## [HashiCorp Vault](https://github.com/edgelesssys/ego/tree/master/samples/vault)

Vault is a common way to store secrets and share them on dynamic infrastructures. With EGo, you can build a confidential version of unmodified Vault.

## [WebAssembly with Wasmer](https://github.com/edgelesssys/ego/tree/master/samples/wasmer)
You can run WebAssembly inside EGo using Wasmer.

## [Microsoft Azure Attestation (MAA)](https://github.com/edgelesssys/ego/tree/master/samples/azure_attestation)

Azure offers MAA as a public service. Clients can send remote-attestation statements to MAA via a REST API. MAA verifies such statements and returns a corresponding JSON Web Token (JWT). This example demonstrates how to use MAA with EGo.

## [cgo](https://github.com/edgelesssys/ego/tree/master/samples/cgo)

Go apps can use C and C++ libraries through cgo. EGo also supports cgo and this example shows you how.
