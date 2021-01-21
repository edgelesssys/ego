#!/bin/bash

. /opt/edgelessrt/share/openenclave/openenclaverc
egoPath=$GOPATH/src/github.com/edgelesssys/ego

tPath=$(mktemp -d)
cd $tPath 

cmake -DCMAKE_INSTALL_PREFIX=$tPath $egoPath 
make
make install
export PATH="$PATH:$tPath/bin"

cp $egoPath/samples/helloworld/helloworld.go .


./bin/ego-go build helloworld.go

./bin/ego sign helloworld
retVal=$?
if [ $retVal -ne 0 ]; then
    echo "Error when executing ./bin/ego sign helloworld"
    cd $egoPath
    rm -r $tPath
    exit 1
fi


./bin/ego sign 
retVal=$?
if [ $retVal -ne 0 ]; then
    echo "Error when executing ./bin/ego sign"
    cd $egoPath
    rm -r $tPath
    exit 1
fi


./bin/ego sign enclave.json
retVal=$?
if [ $retVal -ne 0 ]; then
    echo "Error when executing ./bin/ego sign enclave.json"
    cd $egoPath
    rm -r $tPath
    exit 1
fi



OE_SIMULATION=1 ./bin/ego run helloworld
retVal=$?
if [ $retVal -ne 0 ]; then
    echo "Error when executing 'OE_SIMULATION=1 ./bin/ego run helloworld'"
    cd $egoPath
    rm -r $tPath
    exit 1
fi

cd $egoPath
rm -r $tPath

echo "All tests passed!"