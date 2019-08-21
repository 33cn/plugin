#!/usr/bin/env bash

set -e
set -o pipefail
#set -o verbose
#set -o xtrace

CHAIN33_PATH=$1

sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi
echo "=====chain33_path: ${CHAIN33_PATH} ========"
AutoTestMain="${CHAIN33_PATH}/cmd/autotest/main.go"
ImportPlugin='"github.com/33cn/plugin/plugin"'

function build_auto_test() {

    cp "${AutoTestMain}" ./
    sed -i $sedfix "/^package/a import _ ${ImportPlugin}" main.go
    go build -v -i -o autotest

}

function clean_auto_test() {
    rm -f ../autotest/main.go
}

trap "clean_auto_test" INT TERM EXIT

build_auto_test
