#!/usr/bin/env bash

strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${1}/$strapp"
SRC_RELAYD=github.com/33cn/plugin/plugin/dapp/relay/cmd/relayd
FLAG=$2

# shellcheck disable=SC2086,1072
go build ${FLAG} -v -o "${OUT_DIR}/relayd" "${SRC_RELAYD}"
cp ./relayd/relayd.toml "${OUT_DIR}/relayd.toml"
cp ./build/* "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./test/* "${OUT_TESTDIR}"
