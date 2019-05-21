#!/usr/bin/env bash

strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${1}/$strapp"

PARACLI="${OUT_DIR}/chain33-para-cli"
PARANAME=para
SRC_CLI=github.com/33cn/plugin/cli

go build -v -o "${PARACLI}" -ldflags "-X ${SRC_CLI}/buildflags.ParaName=user.p.${PARANAME}. -X ${SRC_CLI}/buildflags.RPCAddr=http://localhost:8901" "${SRC_CLI}"
# shellcheck disable=SC2086
cp ./build/* "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./test/* "${OUT_TESTDIR}"
