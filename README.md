# EGo
<img src="doc/logo.svg" alt="EGo logo" width="25%"/>

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
If you are on Ubuntu 18.04 and do not want to build EGo from source, you can install the binary release:
```bash
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo apt-key add
sudo add-apt-repository 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu bionic main'
wget https://github.com/edgelesssys/ego/releases/download/v0.1.0/ego_0.1.0_amd64.deb
sudo apt install ./ego_0.1.0_amd64.deb
```
Then proceed with [Use](#use).

## Build and Install
*Prerequisite*: [Edgeless RT](https://github.com/edgelesssys/edgelessrt) is installed and sourced.

```sh
mkdir build
cd build
cmake ..
make
make install
```

## Use
To use EGo, its `bin` directory must be in the PATH:
```sh
export PATH="$PATH:/opt/ego/bin"
```
Now you are ready to build applications with EGo!

### Samples
* [helloworld](samples/helloworld) is a minimal example of an enclave application.
* [remote_attestation](samples/remote_attestation) shows how to do remote attestation in EGo.
* [vault](samples/vault) demonstrates how to port a Go application exemplified by Hashicorp Vault.

## Documentation
* The [EGo API](https://pkg.go.dev/github.com/edgelesssys/ego) provides access to *remote attestation* and *sealing* to your confidential app at runtime.
* [`ego` command reference](doc/ego_cli.md)
* [Debugging](doc/debugging.md)
