#!/usr/bin/env bash

strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${1}/$strapp"
mkdir -p "${OUT_DIR}"
cp ./build/* "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./test/* "${OUT_TESTDIR}"
