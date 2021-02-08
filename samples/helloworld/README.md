# helloworld sample
This sample shows how to build an enclave application with EGo.

The sample can be built as follows:
```sh
ego-go build
ego sign helloworld
```

To run it inside the enclave:
```sh
ego run helloworld
```

To run it in simulation mode:
```sh
OE_SIMULATION=1 ego run helloworld
```
