# Vault
```sh
git clone https://github.com/hashicorp/vault
cd vault

# either
ego-env make GO_VERSION_MIN=1.14.6

# or
ego-go build -o bin/vault

cd bin
ego sign vault
ego-run vault server -dev
```
