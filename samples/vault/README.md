# Hashicorp Vault sample
This sample shows how to port an existing Go application to EGo.

First clone the repo:
```sh
git clone https://github.com/hashicorp/vault
cd vault
```

Then you can build the Vault executable with `ego-go build`:
```sh
ego-go build -o bin/vault
```

Alternatively, you can also use `make` by running it within the `ego env`:
```sh
ego env make GO_VERSION_MIN=1.14.6
```

Either way, you need to sign the resulting executable:
```sh
cd bin
ego sign vault
```

Then you can run Vault:
```
ego run vault server -dev
```
