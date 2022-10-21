# Installing EGo 📦

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
If you're on Ubuntu 18.04 or above, you can install the DEB package:
```bash
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo apt-key add
sudo add-apt-repository "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu `lsb_release -cs` main"
wget https://github.com/edgelesssys/ego/releases/download/v1.0.1/ego_1.0.1_amd64.deb
sudo apt install ./ego_1.0.1_amd64.deb build-essential libssl-dev
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
