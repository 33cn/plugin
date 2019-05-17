#!/usr/bin/env bash

strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}
OUT_DIR="${1}/$strapp"
#FLAG=$2

mkdir -p "${OUT_DIR}"
cp ./build/* "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
chmod +x ./build/test-rpc.sh
cp ./build/test-rpc.sh "${OUT_TESTDIR}"
