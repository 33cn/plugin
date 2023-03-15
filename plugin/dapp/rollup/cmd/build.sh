#!/bin/bash
# 官方ci集成脚本
strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${1}/$strapp"
#FLAG=$2

mkdir -p "${OUT_DIR}"
cp ./ci/* "${OUT_DIR}"

CHAIN33_PATH=$(go list -f "{{.Dir}}" github.com/33cn/chain33)
PLUGIN_PATH=$(go list -f "{{.Dir}}" github.com/33cn/plugin)
# copy chain33 toml

cp "${CHAIN33_PATH}/cmd/chain33/chain33.test.toml" "${OUT_DIR}"
cp "${PLUGIN_PATH}/chain33.para.toml" "${OUT_DIR}"
