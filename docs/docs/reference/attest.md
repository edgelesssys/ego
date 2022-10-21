# Remote attestation
Use *remote attestation* to verify that an EGo app is indeed running inside an enclave and to identify it by its hash.

Attestation relies on external SGX services:
* The *Provisioning Certificate Caching Service (PCCS)* caches attestation data from Intel. It's either operated by the cloud provider or must be hosted by yourself.
* The *quote provider* helps EGo to connect to the PCCS. Both the attester and the verifier must install it.

The required setup varies depending on the environment.

## Azure
Azure operates a PCCS. Use it by installing the Azure DCAP client as quote provider:
```bash
sudo ego install az-dcap-client
```
Or manually:
```bash
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | sudo apt-key add
sudo add-apt-repository "deb [arch=amd64] https://packages.microsoft.com/ubuntu/`lsb_release -rs`/prod `lsb_release -cs` main"
sudo apt install az-dcap-client
```

## Alibaba Cloud
Alibaba operates a PCCS that can be used with the default quote provider.
1. Install the default quote provider:
   ```bash
   sudo ego install libsgx-dcap-default-qpl
   ```
   Or manually by following the [guide from Open Enclave](https://github.com/openenclave/openenclave/blob/master/docs/GettingStartedDocs/Contributors/NonAccMachineSGXLinuxGettingStarted.md#3-set-up-intel-dcap-quote-provider-library-qpl).

1. Set the `PCCS_URL` in `/etc/sgx_default_qcnl.conf` as explained in [Alibaba's documentation](https://www.alibabacloud.com/help/doc-detail/208095.htm#step-fn4-02q-tj4)

## On-premises or other cloud
You must host a PCCS yourself and use the default quote provider to connect to it.

### Set up the PCCS
1. Register with [Intel](https://api.portal.trustedservices.intel.com/provisioning-certification) to get a PCCS API key
1. Run the PCCS:
   ```bash
   docker run -e APIKEY=<your-API-key> -p 8081:8081 --name pccs -d ghcr.io/edgelesssys/pccs
   ```
1. Verify that the PCCS is running:
   ```bash
   curl -kv https://localhost:8081/sgx/certification/v3/rootcacrl
   ```
   You should see a 200 status code.

### Set up the quote provider
1. Install the default quote provider:
   ```bash
   sudo ego install libsgx-dcap-default-qpl
   ```
   Or manually by following the [guide from Open Enclave](https://github.com/openenclave/openenclave/blob/master/docs/GettingStartedDocs/Contributors/NonAccMachineSGXLinuxGettingStarted.md#3-set-up-intel-dcap-quote-provider-library-qpl).

1. If the PCCS runs on another machine than the quote provider, change the host of the `PCCS_URL` in `/etc/sgx_default_qcnl.conf` to the PCCS machine.
