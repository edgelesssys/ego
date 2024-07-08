# EStore sample

This sample shows how to use [EStore](https://github.com/edgelesssys/estore) with EGo.
EStore is a key-value store with authenticated encryption for data at rest.
It's particularly well suited for use inside an enclave.

The sample uses SGX sealing to securely store the encryption key of the database.
Alternatively, you can use [MarbleRun](https://github.com/edgelesssys/marblerun) to manage the key.

You can build and run the sample as follows:

```bash
ego-go build
ego sign estore-sample
ego run estore-sample
```

You should see an output similar to:

```shell-session
$ ego run estore-sample
[erthost] loading enclave ...
[erthost] entering enclave ...
[ego] starting application ...
Creating new DB
hello=world

$ ego run estore-sample
[erthost] loading enclave ...
[erthost] entering enclave ...
[ego] starting application ...
Found existing DB
hello=world
```
