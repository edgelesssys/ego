#!/bin/bash
set -e

onexit()
{
    if [ $? -ne 0 ]; then
        echo "failed"
    else
        echo "All tests passed!"
    fi
    rm -r $tPath
    rm -r /tmp/ego-integration-test
}

trap onexit EXIT

run()
{
    echo "run test: $@"
    $@
}


parent_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )
egoPath=$parent_path/..

tPath=$(mktemp -d)
cd $tPath

cmake -DCMAKE_INSTALL_PREFIX=$tPath/install $egoPath
make -j`nproc`
make install
export PATH="$tPath/install/bin:$PATH"

# Setup integration test
mkdir -p /tmp/ego-integration-test
echo -n -e "It works!" > /tmp/ego-integration-test/test-file.txt

# Build integration test
cd $egoPath/cmd/integration-test/
cp enclave.json /tmp/ego-integration-test/enclave.json
run ego-go build -o /tmp/ego-integration-test/integration-test

# Sign & run intergration test
cd /tmp/ego-integration-test
run ego sign
run ego run integration-test
