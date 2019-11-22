#!/usr/bin/env bash
# shellcheck disable=SC2128

RPC_TESTFILE=test-rpc.sh
DAPP_TEST_COMMON=dapp-test-common.sh

function dapp_test_rpc() {
    local ip=$1
    echo "============ # dapp rpc test begin ============="
    if [ -d dapptest ]; then
        cp "$DAPP_TEST_COMMON" dapptest/
        cd dapptest || return
        rm -f "retries.log"
        rm -f "jobs.log"

        dapps=$(find . -maxdepth 1 -type d ! -name dapptest ! -name . | sed 's/^\.\///' | sort)
        echo "dapps list: $dapps"
        set +e
        parallel -k --joblog ./jobs.log 'echo tried {} >>./retries.log; ./{}/"'"${RPC_TESTFILE}"'" "'"$ip"'"' ::: "$dapps"
        local ret=$?
        # retries 3 times if one dapp fail
        echo "============ # retried dapps log: ============="
        cat ./retries.log
        echo "============ # check dapps test log: ============="
        cat ./jobs.log
        set -e
        if [ $ret -ne 0 ]; then
            exit 1
        fi

    fi
    echo "============ # dapp rpc test end ============="
}
