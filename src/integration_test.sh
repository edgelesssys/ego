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
    rm -r /tmp/ego-unsupported-import-test
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
mkdir -p /tmp/ego-integration-test/relative/path
echo -n 'It works!' > /tmp/ego-integration-test/test-file.txt
echo -n 'It relatively works!' > /tmp/ego-integration-test/relative/path/test-file.txt
echo -n 'i should be in memfs' > /tmp/ego-integration-test/file-host.txt

# Build integration test
cd $egoPath/ego/cmd/integration-test/
cp enclave.json /tmp/ego-integration-test/enclave.json
export CGO_ENABLED=0  # test that ego-go ignores this
run ego-go build -o /tmp/ego-integration-test/integration-test

# Sign & run intergration test
cd /tmp/ego-integration-test
run ego sign
run ego run integration-test

# Test unsupported import detection on sign & run
mkdir -p /tmp/ego-unsupported-import-test
cd $egoPath/ego/cmd/unsupported-import-test
run ego-go build -o /tmp/ego-unsupported-import-test/unsupported-import
cd /tmp/ego-unsupported-import-test
run ego sign unsupported-import |& grep "You cannot import the github.com/edgelesssys/ego/eclient package"
run ego run unsupported-import |& grep "You cannot import the github.com/edgelesssys/ego/eclient package"
