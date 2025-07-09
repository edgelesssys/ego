# Installing EGo

## Install the snap

The easiest way to install EGo is via the snap:

```bash
sudo snap install ego-dev --classic
```

You also need `gcc` and `libcrypto`. On Ubuntu install them with:

```bash
sudo apt install build-essential libssl-dev
```

## Install the DEB package

If you're on Ubuntu 20.04, 22.04, or 24.04, you can install the DEB package:

```bash
sudo mkdir -p /etc/apt/keyrings
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo tee /etc/apt/keyrings/intel-sgx-keyring.asc > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/intel-sgx-keyring.asc arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/intel-sgx.list
sudo apt update
EGO_DEB=ego_1.7.1_amd64_ubuntu-$(lsb_release -rs).deb
wget https://github.com/edgelesssys/ego/releases/download/v1.7.1/$EGO_DEB
sudo apt install ./$EGO_DEB build-essential libssl-dev
```

## Build from source

You can also build EGo yourself, with the following steps.

*Prerequisite* : [Edgeless RT](https://github.com/edgelesssys/edgelessrt) is installed and sourced.

```bash
mkdir build
cd build
cmake ..
make
make install
```
