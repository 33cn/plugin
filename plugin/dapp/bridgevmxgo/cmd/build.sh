#!/usr/bin/env bash

# 官方ci集成脚本
strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

SRC_EBCLI=github.com/33cn/plugin/plugin/dapp/cross2eth/ebcli
SRC_EBRELAYER=github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer
SRC_BOSS4XCLI=github.com/33cn/plugin/plugin/dapp/cross2eth/boss4x
SRC_EVMXGOBOSS4XCLI=github.com/33cn/plugin/plugin/dapp/bridgevmxgo/boss4x

OUT_DIR="${1}/$strapp"
FLAG=$2

# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/ebrelayer" "${SRC_EBRELAYER}"
# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_A" "${SRC_EBCLI}"
# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_B" -ldflags "-X ${SRC_EBCLI}/buildflags.RPCAddr=http://localhost:9902" "${SRC_EBCLI}"
# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_C" -ldflags "-X ${SRC_EBCLI}/buildflags.RPCAddr=http://localhost:9903" "${SRC_EBCLI}"
# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_D" -ldflags "-X ${SRC_EBCLI}/buildflags.RPCAddr=http://localhost:9904" "${SRC_EBCLI}"
# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/boss4x" "${SRC_BOSS4XCLI}"
# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/evmxgoboss4x" "${SRC_EVMXGOBOSS4XCLI}"

cp ../../../../chain33.para.toml "${OUT_DIR}"
cp ../../cross2eth/ebrelayer/relayer.toml "${OUT_DIR}/relayer.toml"
cp ./build/* "${OUT_DIR}"
cp ./build/abi/* "${OUT_DIR}"
cp ../../cross2eth/cmd/build/public/* "${OUT_DIR}"
cp ../../cross2eth/cmd/build/abi/* "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./test/* "${OUT_TESTDIR}"
