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
55c37e50951559004bffe9dffa4c21779e0783f76ed76fe2b7a43ff6ab8d3fdb
```
You should see the same UniqueID as above.

Or build a docker image for deployment:
```sh
DOCKER_BUILDKIT=1 docker build --secret id=signingkey,src=private.pem --target deploy -t ego-helloworld .
```
