#!/usr/bin/env bash

strpwd=$(pwd)
strcmd=${strpwd##*dapp/}
strapp=${strcmd%/cmd*}

OUT_DIR="${1}/$strapp"
mkdir -p "${OUT_DIR}"
# shellcheck disable=SC2086
cp -r ./build/* "${OUT_DIR}"

OUT_TESTDIR="${1}/dapptest/$strapp"
mkdir -p "${OUT_TESTDIR}"
cp ./test/* "${OUT_TESTDIR}"

#Get files related to jvm exec env
cd "${strpwd}" || exit
if ! [ -d bigfile ]; then
    echo "Going to get files related to jvm exec env"
    git clone https://gitlab.33.cn/root/bigfile.git
fi
cp bigfile/jvm/contract_loader/Chain33Loader.jar "${OUT_DIR}"
mkdir -p "${OUT_DIR}"/jarlib
cp bigfile/jvm/jarlib/Gson.jar "${OUT_DIR}"/jarlib
cp bigfile/jvm/java_contract/* "${OUT_DIR}"
cp -r bigfile/jvm/j2sdk-image "${OUT_DIR}"
cp bigfile/jvm/jli_static_lib/libjli.a ../openjdk/
cd - || exit
