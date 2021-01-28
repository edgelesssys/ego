# EGo
![EGo logo](doc/logo.svg)

[EGo](https://ego.dev) is an SDK for building confidential SGX enclaves in Go. It makes enclave development simple by providing two user-friendly tools:
* `ego-go`, a modified Go compiler that automatically compiles a Go project to an enclave - while providing the same CLI as the original Go compiler.
* `ego`, a CLI tool that handles all enclave-related tasks such as signing and running.

Building and running a confidential Go app is as easy as:
```sh
ego-go build hello.go
ego sign hello
ego run hello
```

## Quick Start
If you are on Ubuntu 18.04 and do not want to build the SDK from source, you can install the binary release:
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
To use the SDK, the EGo `bin` directory must be in the PATH:
```sh
export PATH="$PATH:/opt/ego/bin"
```
Now you are ready to build applications with EGo!

### Samples
* [helloworld](samples/helloworld) is a minimal example of an enclave application.
* [remote_attestation](samples/remote_attestation) shows how to do remote attestation in EGo.
* [vault](samples/vault) demonstrates how to port a Go application exemplified by Hashicorp Vault.

## Documentation
* [`ego` command reference](doc/ego_cli.md)
* The [EGo API](https://pkg.go.dev/github.com/edgelesssys/ertgolib) provides remote attestation and sealing
* [Debugging](doc/debugging.md)
