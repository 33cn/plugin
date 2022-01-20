#!/usr/bin/env bash

# 官方ci集成脚本
strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

SRC_EBCLI=github.com/33cn/plugin/plugin/dapp/cross2eth/ebcli
SRC_EBRELAYER=github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer
SRC_BOSS4XCLI=github.com/33cn/plugin/plugin/dapp/cross2eth/boss4x

OUT_DIR="${1}/$strapp"
FLAG=$2

# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/ebrelayer" "${SRC_EBRELAYER}"
# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/ebcli_A" "${SRC_EBCLI}"
# shellcheck disable=SC2086,1072
go build -i ${FLAG} -v -o "${OUT_DIR}/boss4x" "${SRC_BOSS4XCLI}"

cp ../../../../chain33.para.toml "${OUT_DIR}"
cp ../ebrelayer/relayer.toml "${OUT_DIR}/relayer.toml"
cp ./build/* "${OUT_DIR}"
cp ./build/abi/* "${OUT_DIR}"
cp ./build/public/* "${OUT_DIR}"
cp ../../cross2eth/boss4x/chain33/deploy_chain33.toml "${OUT_DIR}"
cp ../../cross2eth/boss4x/ethereum/deploy_ethereum.toml "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./test/* "${OUT_TESTDIR}"
