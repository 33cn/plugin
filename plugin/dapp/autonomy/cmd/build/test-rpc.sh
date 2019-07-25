#!/usr/bin/env bash
# shellcheck disable=SC2128

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
#txhash=""

function run_testcases() {
    echo "run_testcases"
}

function debug_function() {
    set -x
    eval "$@"
    set +x
}

function rpc_test() {
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases

    if [ -n "$CASE_ERR" ]; then
        echo "=======autonomy rpc test  error ==========="
        exit 1
    else
        echo "====== autonomy rpc test  pass ==========="
    fi
}

debug_function rpc_test "$1"
