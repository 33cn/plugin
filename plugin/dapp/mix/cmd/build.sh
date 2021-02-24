#!/usr/bin/env bash

strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${1}/$strapp"

mkdir -p "${OUT_DIR}"/gnark
# shellcheck disable=SC2086
cp ./build/* "${OUT_DIR}"
cp -r ./gnark/circuit "${OUT_DIR}"/gnark

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./test/* "${OUT_TESTDIR}"
