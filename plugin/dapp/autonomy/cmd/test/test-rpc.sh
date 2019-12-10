#!/usr/bin/env bash
# shellcheck disable=SC2128

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
#txhash=""

function run_testcases() {
    echo "run_testcases"
}

function rpc_test() {
    chain33_RpcTestBegin autonomy

    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases

    chain33_RpcTestRst autonomy "$CASE_ERR"

}

chain33_debug_function rpc_test "$1"
