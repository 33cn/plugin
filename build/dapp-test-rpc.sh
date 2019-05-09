#!/usr/bin/env bash
# shellcheck disable=SC2128

RPC_TESTFILE=test-rpc.sh

function dapp_test_rpc() {
    local ip=$1
    echo "============ # dapp rpc test ============="
    if [ -d dapptest ]; then
        cd dapptest || return
        dir=$(find . -maxdepth 1 -type d ! -name dapptest ! -name . | sed 's/^\.\///')
        for app in $dir; do
            echo "=========== # $app rpc test ============="
            ./"$app/${RPC_TESTFILE}" "$ip"
        done

    fi
}

#dapp_test_rpc $1
