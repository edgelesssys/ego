#!/bin/bash
source "$(dirname "$0")/setup.sh"

run ego-go build helloworld.go
run ego sign helloworld
run ego sign
run ego sign enclave.json

export OE_SIMULATION=1
run "ego run helloworld"
