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
7cb7cc41b6d45b5f5d8517c51650d00961104da2daab93732f7a90909e5f5136
```
You should see the same UniqueID as above.

Or build a docker image for deployment:
```sh
DOCKER_BUILDKIT=1 docker build --secret id=signingkey,src=private.pem --target deploy -t ego-helloworld .
```
