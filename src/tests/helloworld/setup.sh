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
}

trap onexit EXIT

run()
{
    echo "run test: $@"
    $@
}


parent_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )
egoPath=$parent_path/../../..

tPath=$(mktemp -d)
cd $tPath

cmake -DCMAKE_INSTALL_PREFIX=$tPath/install $egoPath
make
make install
export PATH="$tPath/install/bin:$PATH"
cp $egoPath/samples/helloworld/helloworld.go .
