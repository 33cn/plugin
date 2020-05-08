#!/usr/bin/env bash
source "./allRelayerTest.sh"

function x2Ethereum() {
    if [ "${2}" == "init" ]; then
        return
    elif [ "${2}" == "config" ]; then
        return
    elif [ "${2}" == "test" ]; then
        echo "========================== x2Ethereum test =========================="
        set +e
        set -x
        MainTest 1
    fi
}
