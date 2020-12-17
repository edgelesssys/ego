# Vault
```sh
git clone https://github.com/hashicorp/vault
cd vault

# either
sed -i 's GO_VERSION_MIN=.* GO_VERSION_MIN=1.14.6 ' Makefile
ego-env make

# or
ego-go build -o bin/vault

cd bin
ego sign vault
ego-run vault server -dev
```
