# custom file sample
This sample shows how to include custom files into your binary.
The file will encode into the binary. In the sample, we chose `/etc/ssl/certs/ca-certificates.crt`, which contains the certificates of common CA's and allows us to make secure tls connections without relying on the certifiactes provided by the host system. To specify files, just add its source and destination path to your `enclave.json` as done in this sample. Directories will be created as needed.


The sample can be built as follows:
```sh
ego-go build
ego sign custom_file
```

To run it inside the enclave:
```sh
ego run custom_file
```

To run it in simulation mode:
```sh
OE_SIMULATION=1 ego run custom_file
```

You should see the html code of https://www.edgeless.systems/ as output.
