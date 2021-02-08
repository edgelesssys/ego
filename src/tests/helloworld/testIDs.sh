#!/bin/bash
source "$(dirname "$0")/setup.sh"

run ego-go build helloworld.go
run ego sign helloworld

run ego signerid helloworld
res1=$(ego signerid helloworld)
run ego signerid public.pem
res2=$(ego signerid public.pem)

run ego uniqueid helloworld
res3=$(ego uniqueid helloworld)

if [[ "$res1" != "$res2" ]]; then
    exit 1
fi
if [[ "$res1" == "$res3" ]]; then
    exit 1
fi

if [[ "${#res1}" != "64" ]]; then
    exit 1
fi
if [[ "${#res3}"  != "64" ]]; then
    exit 1
fi
