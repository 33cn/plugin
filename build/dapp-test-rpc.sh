#!/usr/bin/env bash
# shellcheck disable=SC2128

RPC_TESTFILE=test-rpc.sh

function dapp_test_rpc() {
    local ip=$1
    echo "============ # dapp rpc test begin ============="
    if [ -d dapptest ]; then
        cd dapptest || return
        dir=$(find . -maxdepth 1 -type d ! -name dapptest ! -name blackwhite ! -name . | sed 's/^\.\///')
        for app in $dir; do
            echo "=========== # $app rpc test ============="
            ./"$app/${RPC_TESTFILE}" "$ip"
            echo "=========== # $app rpc end ============="
        done

    fi
    echo "============ # dapp rpc test end ============="
}

#dapp_test_rpc $1
