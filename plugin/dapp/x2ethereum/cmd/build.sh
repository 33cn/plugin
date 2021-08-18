#!/usr/bin/env bash

# 官方ci集成脚本
strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

#SRC_EBCLI=github.com/33cn/plugin/plugin/dapp/x2ethereum/ebcli
#SRC_EBRELAYER=github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer
#OUT_DIR="${1}/$strapp"
#FLAG=$2
#
## shellcheck disable=SC2086,1072
#go build -i ${FLAG} -v -o "${OUT_DIR}/ebrelayer" "${SRC_EBRELAYER}"
## shellcheck disable=SC2086,1072
#go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_A" "${SRC_EBCLI}"
## shellcheck disable=SC2086,1072
#go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_B" -ldflags "-X ${SRC_EBCLI}/buildflags.RPCAddr=http://localhost:9902" "${SRC_EBCLI}"
## shellcheck disable=SC2086,1072
#go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_C" -ldflags "-X ${SRC_EBCLI}/buildflags.RPCAddr=http://localhost:9903" "${SRC_EBCLI}"
## shellcheck disable=SC2086,1072
#go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_D" -ldflags "-X ${SRC_EBCLI}/buildflags.RPCAddr=http://localhost:9904" "${SRC_EBCLI}"
#
#cp ../ebrelayer/relayer.toml "${OUT_DIR}/relayer.toml"
#cp ./build/* "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./test/* "${OUT_TESTDIR}"
