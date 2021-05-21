# EGo
<img src="src/logo.svg" alt="EGo logo" width="40%"/>

[![GitHub Actions Status][github-actions-badge]][github-actions]
[![GitHub license][license-badge]](LICENSE)
[![Go Report Card][go-report-card-badge]][go-report-card]
[![PkgGoDev][go-pkg-badge]][go-pkg]
[![Discord Chat][discord-badge]][discord]

[EGo](https://ego.dev) is a framework for building *confidential apps* in Go. Confidential apps run in always-encrypted and verifiable enclaves on Intel SGX-enabled hardware. EGo simplifies enclave development by providing two user-friendly tools:

* `ego-go`, an adapted Go compiler that builds enclave-compatible executables from a given Go project - while providing the same CLI as the original Go compiler.
* `ego`, a CLI tool that handles all enclave-related tasks such as signing and enclave creation.

Building and running a confidential Go app is as easy as:
```sh
ego-go build hello.go
ego sign hello
ego run hello
```

## Install

### Install the Snap
The easiest way to install EGo is via the Snap:
```sh
sudo snap install ego-dev --classic
```

### Install the DEB package
If you're on Ubuntu 18.04 or above, you can install the DEB package:
```bash
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo apt-key add
sudo add-apt-repository "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu `lsb_release -cs` main"
wget https://github.com/edgelesssys/ego/releases/download/v0.2.2/ego_0.2.2_amd64.deb
sudo apt install ./ego_0.2.2_amd64.deb build-essential
```

### Build from source
*Prerequisite*: [Edgeless RT](https://github.com/edgelesssys/edgelessrt) is installed and sourced.

```sh
mkdir build
cd build
cmake ..
make
make install
```
## Getting started
Now you're ready to build applications with EGo! To start, check out the following samples:
* [helloworld](samples/helloworld) is a minimal example of an enclave application.
* [remote_attestation](samples/remote_attestation) shows how to use the basic remote attestation API of EGo.
* [attested_tls](samples/attested_tls) is similar to the above, but uses a higher level API to establish an attested TLS connection.
* [vault](samples/vault) demonstrates how to port a Go application exemplified by Hashicorp Vault.
* [cgo](samples/cgo) demonstrates the experimental cgo support.
* [azure_attestation](samples/azure_attestation) shows how to use Microsoft Azure Attestation for remote attestation.

## Documentation
* The [EGo documentation](https://ego.dev/docs/introduction/) covers building, signing, running, and debugging confidential apps.
* The [EGo API](https://pkg.go.dev/github.com/edgelesssys/ego) provides access to *remote attestation* and *sealing* to your confidential app at runtime.

## Contribute
To report a problem or suggest a new feature, file a GitHub issue.

To report a security issue, write to security@edgeless.systems

Pull requests are welcome! You need to agree to our [Contributor License Agreement](https://cla-assistant.io/edgelesssys/ego).

<!-- refs -->
[github-actions]: https://github.com/edgelesssys/ego/actions
[github-actions-badge]: https://github.com/edgelesssys/ego/workflows/Unit%20Tests/badge.svg
[go-pkg]: https://pkg.go.dev/github.com/edgelesssys/ego
[go-pkg-badge]: https://pkg.go.dev/badge/github.com/edgelesssys/ego
[go-report-card]: https://goreportcard.com/report/github.com/edgelesssys/ego
[go-report-card-badge]: https://goreportcard.com/badge/github.com/edgelesssys/ego
[license-badge]: https://img.shields.io/github/license/edgelesssys/ego
[discord]: https://discord.gg/rH8QTH56JN
[discord-badge]: https://img.shields.io/badge/chat-on%20Discord-blue
