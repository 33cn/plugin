#!/usr/bin/env bash

output_dir=${1}
strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${output_dir}/$strapp"
[ ! -e "${OUT_DIR}" ] && mkdir -p "${OUT_DIR}"

# shellcheck disable=SC2086
cp ./build/* "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./build/test-rpc.sh "${OUT_TESTDIR}"
