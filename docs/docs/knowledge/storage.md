# Secure storage

EGo provides several ways to store data securely, which are described in the following sections.

## In-enclave-memory filesystem

By default, when an EGo app writes a file, it's stored in the enclave's memory.
This is secure, but the data is lost when the enclave is terminated.

For persistence, you can [mount a host directory](../reference/config.md#mounts) into the enclave.
You should encrypt the data before writing it to the untrusted host filesystem.
You can use one of the following methods for this.

## Sealing

Sealing is the process of encrypting data with a key derived from the enclave and the CPU it's running on.
Use the [EGo API](https://pkg.go.dev/github.com/edgelesssys/ego/ecrypto) to seal and unseal data.

## EStore

[EStore](https://github.com/edgelesssys/estore) is a key-value store with authenticated encryption for data at rest.
It's particularly well suited for use inside an enclave.
Check out the [EStore sample](https://github.com/edgelesssys/ego/tree/master/samples/estore) for a demonstration.
