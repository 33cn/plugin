#!/usr/bin/env bash
# shellcheck disable=SC2128

RPC_TESTFILE=test-rpc.sh
DAPP_TEST_COMMON=dapp-test-common.sh

function dapp_test_rpc() {
    local ip=$1
    echo "============ # dapp rpc test begin ============="
    if [ -d dapptest ]; then
        cp $DAPP_TEST_COMMON dapptest/
        cd dapptest || return

        dapps=$(find . -maxdepth 1 -type d ! -name dapptest ! -name evm ! -name game ! -name guess ! -name hashlock ! -name ticket ! -name lottery ! -name pokerbull ! -name token ! -name trade ! -name . | sed 's/^\.\///' | sort)
        echo "dapps list: $dapps"
        parallel -k --retries 3 --verbose --joblog ./testlog ./{}/test-rpc.sh "$ip" ::: "$dapps"
        echo "check dapps test log"
        cat ./testlog
    fi
    echo "============ # dapp rpc test end ============="
}

#dapp_test_rpc $1
