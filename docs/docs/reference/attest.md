# Remote attestation

Use *remote attestation* to verify that an EGo app is indeed running inside an enclave and to identify it by its hash.

Attestation relies on external SGX services:

* The *Provisioning Certificate Caching Service (PCCS)* caches attestation data from Intel. It's either operated by the cloud provider or must be hosted by yourself.
* The *quote provider* helps EGo to connect to the PCCS. Both the attester and the verifier must install it.

## Set up the quote provider

Install the quote provider on all machines that should host or verify EGo apps:

```bash
sudo mkdir -p /etc/apt/keyrings
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo tee /etc/apt/keyrings/intel-sgx-keyring.asc > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/intel-sgx-keyring.asc arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/intel-sgx.list
sudo apt update
sudo apt install libsgx-dcap-default-qpl
```

Locate its configuration file at `/etc/sgx_default_qcnl.conf`.
If the file doesn't exist or is outdated, download it with the following command:

```bash
wget -qO- https://raw.githubusercontent.com/intel/SGXDataCenterAttestationPrimitives/master/QuoteGeneration/qcnl/linux/sgx_default_qcnl.conf | sudo tee /etc/sgx_default_qcnl.conf > /dev/null
```

You can configure the quote provider to get the collaterals from the Intel PCS, the PCCS of your cloud service provider (CSP), or your own PCCS.

### Intel PCS (verifier only)

:::note

This method only supports verification. For configuring machines that host EGo apps, choose one of the other methods.

:::

Using the Intel PCS is the simplest and most generic way for verifying, but it may be slower and less reliable than a PCCS.
Configure it by uncommenting the `"collateral_service"` key:

```json
  ,"collateral_service": "https://api.trustedservices.intel.com/sgx/certification/v4/"
```

### PCCS of your CSP

If you're running an EGo app in the cloud, it's recommended to use the PCCS of your CSP.
Set the `"pccs_url"` value to the respective address:

* Azure: `https://global.acccache.azure.net/sgx/certification/v4/`

  See the [Azure documentation](https://learn.microsoft.com/en-us/azure/security/fundamentals/trusted-hardware-identity-management#how-do-i-use-intel-qpl-with-trusted-hardware-identity-management) for more information on configuring the quote provider.

* Alibaba: `https://sgx-dcap-server.[Region-ID].aliyuncs.com/sgx/certification/v4/`

  See the [Alibaba documentation](https://www.alibabacloud.com/help/en/ecs/user-guide/build-an-sgx-encrypted-computing-environment) for supported Region-ID values and more information on configuring the quote provider.

### Your own PCCS

If you're running an EGo app on premises, you must host a PCCS yourself.

#### Set up the PCCS

To set up a PCCS, you can either follow [the instructions from Intel](https://www.intel.com/content/www/us/en/developer/articles/guide/intel-software-guard-extensions-data-center-attestation-primitives-quick-install-guide.html) or use our provided Docker image.

To do the latter, follow these steps:

1. Register with [Intel](https://api.portal.trustedservices.intel.com/provisioning-certification) to get a PCCS API key
2. Run the PCCS:

   ```bash
   docker run -e APIKEY=<your-API-key> -p 8081:8081 --name pccs -d ghcr.io/edgelesssys/pccs
   ```

3. Verify that the PCCS is running:

   ```bash
   curl -kv https://localhost:8081/sgx/certification/v4/rootcacrl
   ```

   You should see a 200 status code.

#### Configure the quote provider

In the configuration file, set the `"pccs_url"` value to the address of your PCCS.

If your PCCS runs with a certificate not signed by a trusted CA, you need to set `"use_secure_cert"` to `false`.
This instructs the quote provider to accept a self-signed certificate of the PCCS.
It doesn't affect the security of the remote attestation process itself.
