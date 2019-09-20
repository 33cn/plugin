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
        dir=$(find . -maxdepth 1 -type d ! -name dapptest ! -name multisig ! -name paracross ! -name ticket ! -name . | sed 's/^\.\///' | sort)
        echo "dapps list: $dir"
        for app in $dir; do
            echo "=========== # $app rpc test ============="
            ./"$app/${RPC_TESTFILE}" "$ip"
            echo "=========== # $app rpc end ============="
        done

        ##ticket用例最后执行
         ./ticket/"${RPC_TESTFILE}" "$ip"
    fi
    echo "============ # dapp rpc test end ============="
}
