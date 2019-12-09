#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
CASE_ERR=""

#eventId=""
#txhash=""

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
}

function run_test() {
    echo "run_test"
}

function main() {
    chain33_RpcTestBegin dposvote
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_test
    chain33_RpcTestRst dposvote "$CASE_ERR"
}

chain33_debug_function main "$1"
