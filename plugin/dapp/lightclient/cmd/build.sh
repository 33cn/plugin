#!/bin/bash
# 官方ci集成脚本
strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${1}/$strapp"
#FLAG=$2

mkdir -p "${OUT_DIR}"
# 本 dapp 没有 ./build 目录(历史遗留), 原来的 cp 必然失败; depends 以前用
# `find -exec` 会静默吞掉失败, 现在 depends 并行化并用 xargs 透出失败, 这里显式判断
if [ -d ./build ]; then
    cp ./build/* "${OUT_DIR}"
fi
