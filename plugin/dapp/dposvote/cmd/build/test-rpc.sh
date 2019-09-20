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

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
}

function run_test() {
    echo "run_test"
}

function main() {

    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    echo "=========== # dposvote rpc test start============="
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Dposvote Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Dposvote Rpc Test Pass==============${NOC}"
    fi
    echo "=========== # dposvote rpc test end============="
}

main "$1"
