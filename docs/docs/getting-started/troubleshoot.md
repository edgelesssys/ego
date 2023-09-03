# Troubleshooting

## SGX

If EGo works in simulation mode (with `OE_SIMULATION=1`), but not with SGX, check the following.

### Operating system

EGo currently supports Ubuntu 20.04 and 22.04.

### Hardware

The hardware must support SGX and it must be enabled in the BIOS:

```shell-session
$ sudo apt install cpuid
$ cpuid | grep SGX
      SGX: Software Guard Extensions supported = true
      SGX_LC: SGX launch config supported      = true
   SGX capability (0x12/0):
      SGX1 supported                         = true
```

* `SGX: Software Guard Extensions supported` is true if the hardware supports it.
* `SGX_LC: SGX launch config supported` is true if the hardware also supports FLC. This is required for attestation.
* `SGX1 supported` is true if it's enabled in the BIOS.

### Driver

The SGX driver exposes a device:

```bash
ls /dev/*sgx*
```

If the output is empty, install the driver.

If your system [supports FLC](#hardware), make sure your Linux kernel is version 5.11 or newer.
You can check with `uname -r`.
If you can't upgrade your kernel, you may [install the *DCAP driver*](https://download.01.org/intel-sgx/latest/linux-latest/docs/Intel_SGX_SW_Installation_Guide_for_Linux.pdf) instead.

On systems without FLC support, you need the SGX *out-of-tree driver*.
Note that Intel deprecated this kind of SGX implementation and EGo doesn't support remote attestation on such systems.
To install the driver, follow the [SGX installation guide from Intel](https://download.01.org/intel-sgx/latest/linux-latest/docs/Intel_SGX_SW_Installation_Guide_for_Linux.pdf).

### Required packages

#### non-FLC system

If your system doesn't support FLC, install the `libsgx-launch` package:

```bash
sudo mkdir -p /etc/apt/keyrings
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo tee /etc/apt/keyrings/intel-sgx-keyring.asc > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/intel-sgx-keyring.asc arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/intel-sgx.list
sudo apt update
sudo apt install libsgx-launch
```

#### SGX device issues

If the [SGX device exists](#driver), but you get one of these errors:

* `Failed to open Intel SGX device.`
* `ERROR: enclave_load_data failed (addr=0x..., prot=0x1, err=0x1001) (oe_result_t=OE_PLATFORM_ERROR)`

Install the `libsgx-enclave-common` package:

```bash
sudo mkdir -p /etc/apt/keyrings
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo tee /etc/apt/keyrings/intel-sgx-keyring.asc > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/intel-sgx-keyring.asc arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/intel-sgx.list
sudo apt update
sudo apt install --no-install-recommends libsgx-enclave-common
```

## Attestation

If EGo works in SGX mode (i.e., without `OE_SIMULATION`), but attestation fails, check the following.

### FLC

Attestation only works on [SGX-FLC](#hardware) systems.

### Quote provider

You must install a [quote provider](../reference/attest.md).

## Out of memory

The amount of available memory to an SGX enclave is set when signing the binary.
If you get a memory allocation error, try to increase the `heapSize` in [enclave.json](../reference/config.md) and sign the binary again.
Note that the runtime itself also occupies memory and that the Go allocator may pre-allocate more memory than is currently in use.
Thus, you usually have to give your enclave more memory than actually used by your app.
