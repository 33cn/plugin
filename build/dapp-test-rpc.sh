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

        dapps=$(find . -maxdepth 1 -type d ! -name dapptest ! -name . | sed 's/^\.\///' | sort)
        echo "dapps list: $dapps"
        parallel -k --retries 3 --joblog ./testlog 'echo tried {} >>./retries.log; ./{}/"'"${RPC_TESTFILE}"'" "'"$ip"'"' ::: "$dapps"
        echo "============ # retried dapps log: ============="
        cat ./retries.log
        echo "============ # check dapps test log: ============="
        cat ./testlog
    fi
    echo "============ # dapp rpc test end ============="
}
