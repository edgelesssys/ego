# Reproducible build
This sample shows how to reproducibly build an EGo application using Docker. The Dockerfile builds the helloworld sample.

First generate a signing key:
```sh
openssl genrsa -out private.pem -3 3072
```

Either build the `helloworld` executable:
```console
$ DOCKER_BUILDKIT=1 docker build --secret id=signingkey,src=private.pem -o. .
$ ego uniqueid helloworld
cfe6779c87fcaef7005dd848391e89c26156c3c51ef36eeea6d6db18cfea29bd
```
You should see the same UniqueID as above.

Or build a docker image for deployment:
```sh
DOCKER_BUILDKIT=1 docker build --secret id=signingkey,src=private.pem --target deploy -t ego-helloworld .
```
