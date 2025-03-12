# PCCS Server

## Build your ready to use PCCS Server using docker

### Prerequisites
In order to establish a connection to Intels PCS API, the PCCS needs to be configured with an API key.
To get your free API key, go to https://api.portal.trustedservices.intel.com/provisioning-certification, create an account and click on "Subscribe".

You should then see two keys. Use either the primary or the second one as your API key in the following.

### Build via Docker

Build the pccs image:
```bash
docker build --tag ghcr.io/edgelesssys/pccs pccs
```

### Run the docker image

After you've build the image, run it using docker. It is important that you paste your API key in the run command.

*Note*: Optionally you can configure your PCCS with a custom user password (`-e USERPASS=<user-pwd>`)
and a custom admin password <br/>(`-e ADMINPASS=<admin-pwd>`), but in most cases there is no need to do that.
```bash
docker run -e APIKEY=<your-API-key> -p 8081:8081 --name pccs -d pccs
```

The PCCS is now available on port 8081. Verify that your PCCS Server runs correctly:
```bash
curl --noproxy "*" -v -k -G "https://localhost:8081/sgx/certification/v3/rootcacrl"
```
You should see a 200 status code. This means your PCCS Server is able to deliver data for your applications!
