#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null

source "./dockerRelayerTest.sh"
source "./paracrosstestcase.sh"

function cross2eth() {
    if [ "${2}" == "init" ]; then
        para_init "${3}"
    elif [ "${2}" == "config" ]; then
        para_set_wallet
        para_transfer
    elif [ "${2}" == "test" ]; then
        echo "========================== cross2eth test =========================="
        set +e
        set -x
        para_create_nodegroup
        AllRelayerMainTest 10
        #        perf_test_main 1
        echo "========================== cross2eth test end =========================="
    fi
}
