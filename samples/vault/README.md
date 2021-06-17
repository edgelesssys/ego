# Hashicorp Vault sample
This sample shows how to port an existing Go application to EGo.

First clone the repo:
```sh
git clone --depth=1 --branch=v1.6.3 https://github.com/hashicorp/vault
cd vault
```

Then you can build the Vault executable with `ego-go build`:
```sh
ego-go build -v -o bin/vault
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
```sh
ego run vault server -dev -dev-no-store-token -dev-root-token-id mytoken
```

Open another terminal and export the environment variables:
```sh
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=mytoken
```

Then use it:
```sh
$ ./vault kv put secret/hello foo=world

Key              Value
---              -----
created_time     2021-03-03T15:30:16.376988478Z
deletion_time    n/a
destroyed        false
version          1

$ ./vault kv get secret/hello

====== Metadata ======
Key              Value
---              -----
created_time     2021-03-03T15:30:16.376988478Z
deletion_time    n/a
destroyed        false
version          1

=== Data ===
Key    Value
---    -----
foo    world
```
