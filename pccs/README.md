# PCCS Server

## Build your ready to use PCCS Server using docker

Build the pccs image:

```bash
docker build --tag pccs .
```

## Run the docker image

After you've build the image, run it using docker.

*Note*: Optionally you can configure your PCCS with a custom user password (`-e USERPASS=<user-pwd>`)
and a custom admin password (`-e ADMINPASS=<admin-pwd>`), but in most cases there is no need to do that.

```bash
docker run -p 8081:8081 --name pccs -d pccs
```

The PCCS is now available on port 8081. Verify that your PCCS Server runs correctly:

```bash
curl --noproxy "*" -v -k -G "https://localhost:8081/sgx/certification/v4/rootcacrl"
```

You should see a 200 status code. This means your PCCS Server is able to deliver data for your applications!
