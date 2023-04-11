#!/bin/bash
# 官方ci集成脚本
strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${1}/$strapp"
#FLAG=$2
OUT_TESTDIR="${1}/dapptest/$strapp"

mkdir -p "${OUT_TESTDIR}"
cp ./test/test-rpc.sh "${OUT_TESTDIR}"

mkdir -p "${OUT_DIR}"
cp ./ci/* "${OUT_DIR}"


PLUGIN_PATH=$(go list -f "{{.Dir}}" github.com/33cn/plugin)
# copy chain33 toml
cp "${PLUGIN_PATH}/chain33.test.toml" "${OUT_DIR}"
