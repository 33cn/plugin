#!/usr/bin/env bash

output_dir=${1}
strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${output_dir}/$strapp"
[ ! -e "${OUT_DIR}" ] && mkdir -p "${OUT_DIR}"

# shellcheck disable=SC2086
cp ./build/* "${OUT_DIR}"
