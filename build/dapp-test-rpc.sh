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

        dapps=$(find . -maxdepth 1 -type d ! -name dapptest ! -name ticket ! -name . | sed 's/^\.\///' | sort)
        echo "dapps list: $dapps"
        parallel -k --retries 3 --verbose --joblog ./testlog ./{}/"${RPC_TESTFILE}" "$ip" ::: "$dapps"
        echo "check dapps test log"
        cat ./testlog

        ##ticket用例最后执行
        ./ticket/"${RPC_TESTFILE}" "$ip"
    fi
    echo "============ # dapp rpc test end ============="
}
