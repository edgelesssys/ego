# EGo
<img src="doc/logo.svg" alt="EGo logo" width="40%"/>

[![GitHub Actions Status][github-actions-badge]][github-actions]
[![GitHub license][license-badge]](LICENSE)
[![Go Report Card][go-report-card-badge]][go-report-card]
[![PkgGoDev][go-pkg-badge]][go-pkg]
[![Gitter Chat][gitter-badge]][gitter]

[EGo](https://ego.dev) is a framework for building *confidential apps* in Go. Confidential apps run in always-encrypted and verifiable enclaves on Intel SGX-enabled hardware. EGo simplifies enclave development by providing two user-friendly tools:

* `ego-go`, an adapted Go compiler that builds enclave-compatible executables from a given Go project - while providing the same CLI as the original Go compiler.
* `ego`, a CLI tool that handles all enclave-related tasks such as signing and enclave creation.

Building and running a confidential Go app is as easy as:
```sh
ego-go build hello.go
ego sign hello
ego run hello
```

## Quick Start
If you are on Ubuntu 18.04 or above and do not want to build EGo from source, you can install the binary release:
```bash
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo apt-key add
sudo add-apt-repository 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu bionic main'
wget https://github.com/edgelesssys/ego/releases/download/v0.1.0/ego_0.1.0_amd64.deb
sudo apt install ./ego_0.1.0_amd64.deb
```
Now you are ready to build applications with EGo! To start, check out the [samples](#samples).

## Build and Install
*Prerequisite*: [Edgeless RT](https://github.com/edgelesssys/edgelessrt) is installed and sourced.

```sh
mkdir build
cd build
cmake ..
make
make install
```

## Samples
* [helloworld](samples/helloworld) is a minimal example of an enclave application.
* [remote_attestation](samples/remote_attestation) shows how to do remote attestation in EGo.
* [vault](samples/vault) demonstrates how to port a Go application exemplified by Hashicorp Vault.

## Documentation
* The [EGo API](https://pkg.go.dev/github.com/edgelesssys/ego) provides access to *remote attestation* and *sealing* to your confidential app at runtime.
* [`ego` command reference](doc/ego_cli.md)
* [Debugging](doc/debugging.md)

<!-- refs -->
[github-actions]: https://github.com/edgelesssys/ego/actions
[github-actions-badge]: https://github.com/edgelesssys/ego/workflows/Unit%20Tests/badge.svg
[go-pkg]: https://pkg.go.dev/github.com/edgelesssys/ego
[go-pkg-badge]: https://pkg.go.dev/badge/github.com/edgelesssys/ego
[go-report-card]: https://goreportcard.com/report/github.com/edgelesssys/ego
[go-report-card-badge]: https://goreportcard.com/badge/github.com/edgelesssys/ego
[license-badge]: https://img.shields.io/github/license/edgelesssys/ego
[gitter]: https://gitter.im/edgeless-systems/community
[gitter-badge]: https://badges.gitter.im/edgeless-systems/Marblerun.svg
