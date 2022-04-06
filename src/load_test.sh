#!/bin/sh
set -e

onexit()
{
  if [ $? -ne 0 ]; then
    echo fail
  else
    echo pass
  fi
  pkill ego-host
  rm -r "$tmp"
}

tmp=$(mktemp -d)
trap onexit EXIT

cd "$tmp"
wget -qO- https://github.com/bojand/ghz/releases/download/v0.105.0/ghz-linux-x86_64.tar.gz | tar xz
wget -qO- https://github.com/grpc/grpc-go/archive/refs/tags/v1.44.0.tar.gz | tar xz
cd grpc-go-1.44.0/examples/helloworld

# delete line that logs each rpc call because it bloats stdout
sed -i '/Printf("Received/d' greeter_server/main.go

ego-go build -o server ./greeter_server

echo '{
  "exe": "server",
  "key": "private.pem",
  "debug": true,
  "heapSize": 512,
  "env": [
    { "name": "EGOMAXTHREADS", "value": "7" },
    { "name": "GOMAXPROCS", "value": "9" }
  ]
}' > enclave.json

ego sign

# start grpc server
ego run server &
pid=$!
sleep 10

# start load test
timeout 10m "$tmp/ghz" --insecure --proto helloworld/helloworld.proto --call helloworld.Greeter/SayHello -n900000 127.0.0.1:50051

# check if process crashed
kill -0 $pid
