#!/bin/bash
set -e

onexit()
{
    if [ $? -ne 0 ]; then
        echo "failed"
    else
        echo "All tests passed!"
    fi
    rm -r "$tPath"
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
cd "$tPath"

cmake -DCMAKE_INSTALL_PREFIX="$tPath/install" "$egoPath"
make -j`nproc`
make install
export PATH="$tPath/install/bin:$PATH"

# Setup integration test
mkdir -p /tmp/ego-integration-test/relative/path
echo -n 'It works!' > /tmp/ego-integration-test/test-file.txt
echo -n 'It relatively works!' > /tmp/ego-integration-test/relative/path/test-file.txt
echo -n 'i should be in memfs' > /tmp/ego-integration-test/file-host.txt

# Build integration test
cd "$egoPath/ego/cmd/integration-test"
cp enclave.json /tmp/ego-integration-test/enclave.json
export CGO_ENABLED=0  # test that ego-go ignores this
run ego-go build -o /tmp/ego-integration-test/integration-test

# Sign intergration test
cd /tmp/ego-integration-test
run ego sign

# Test id commands
dump=$(ego-oesign dump -e integration-test)
run echo "$dump" | grep "^mrenclave=$(ego uniqueid integration-test)$"
run echo "$dump" | grep "^mrsigner=$(ego signerid integration-test)$"
run echo "$dump" | grep "^mrsigner=$(ego signerid public.pem)$"
export OE_LOG_LEVEL=INFO # regression: id commands were broken with OE_LOG_LEVEL set
run echo "$dump" | grep "^mrenclave=$(ego uniqueid integration-test)$"
run echo "$dump" | grep "^mrsigner=$(ego signerid integration-test)$"
run echo "$dump" | grep "^mrsigner=$(ego signerid public.pem)$"
unset OE_LOG_LEVEL

# Run integration test
run ego run integration-test

# Test heap size check on sign
sed -i 's/"heapSize": 16,/"heapSize": 16385,/' enclave.json
run ego sign |& grep "heapSize is set to more than"

# Test ego_largeheap
cd "$egoPath/ego/cmd/integration-test"
run ego-go build -o /tmp/ego-integration-test/integration-test -tags ego_largeheap
cd /tmp/ego-integration-test
run ego sign  # sign with 16385 heapSize should succeed now
sed -i 's/"heapSize": 16385,/"heapSize": 511,/' enclave.json
run ego sign |& grep "heapSize is set to less than"
# Run integration test built with ego_largeheap and heapSize of 512 MB
sed -i 's/"heapSize": 511,/"heapSize": 512,/' enclave.json
run ego sign
run ego run integration-test

# Test unsupported import detection on sign & run
mkdir "$tPath/unsupported-import-test"
cd "$egoPath/ego/cmd/unsupported-import-test"
run ego-go build -o "$tPath/unsupported-import-test/unsupported-import"
cd "$tPath/unsupported-import-test"
run ego sign unsupported-import |& grep "You cannot import the github.com/edgelesssys/ego/eclient package"
run ego run unsupported-import |& grep "You cannot import the github.com/edgelesssys/ego/eclient package"

# Test GetSealKeyID
mkdir "$tPath/sealkeyid-test"
cd "$egoPath/ego/cmd/sealkeyid-test"
run ego-go build -o "$tPath/sealkeyid-test"
cp enclave?.json "$tPath/sealkeyid-test"
cd "$tPath/sealkeyid-test"
run ego sign enclave1.json
keyid1=$(ego run sealkeyid-test)
run ego sign enclave2.json
keyid2=$(ego run sealkeyid-test)
echo 'test keyid1 = keyid2'
test "$keyid1" = "$keyid2"
