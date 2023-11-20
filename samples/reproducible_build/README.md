# Reproducible build

This sample shows how to reproducibly build an EGo application using Docker. The Dockerfile builds the helloworld sample.

First generate a signing key:

```sh
openssl genrsa -out private.pem -3 3072
```

The Dockerfile builds an executable for Ubuntu 22.04.
If you need one for Ubuntu 20.04, edit the Dockerfile and switch to the other base image.

Build the `helloworld` executable:

```console
$ DOCKER_BUILDKIT=1 docker build --secret id=signingkey,src=private.pem -o. .
$ ego uniqueid helloworld
59e89da6161853814f6916462fed1be4727461bf502ce5865a80c8712cc2f548
```

You should see the same UniqueID as above if you didn't modify the Dockerfile.

To run the executable, you need to have the same version of EGo installed that is used in the Dockerfile:

```bash
ego run helloworld
```

Alternatively, you can use `helloworld-bundle`, which doesn't require an EGo installation:

```bash
./helloworld-bundle
```

You can also build a docker image for deployment:

```sh
DOCKER_BUILDKIT=1 docker build --secret id=signingkey,src=private.pem --target deploy -t ego-helloworld .
```
