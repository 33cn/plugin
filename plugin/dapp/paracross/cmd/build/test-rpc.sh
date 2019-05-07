#!/usr/bin/env bash

MAIN_HTTP=""

# $2=0 means true, other false
echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo "$1 ok"
    else
        echo "$1 err"
    fi

}

paracross_GetBlock2MainInfo() {
    height=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"paracross.GetBlock2MainInfo","params":[{"start":1,"end":3}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.items[1].height")
    [ "$height" -eq 2 ]
    rst=$?
    echo_rst $FUNCNAME $rst
}

chain33_lock() {
    ok=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Lock","params":[]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.isOK")
    [ "$ok" == true ]
    rst=$?
    echo_rst $FUNCNAME $rst
}

chain33_unlock() {
    ok=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.UnLock","params":[{"passwd":"1314fuzamei","timeout":0}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.isOK")
    [ "$ok" == true ]
    rst=$?
    echo_rst $FUNCNAME $rst

}
function dapp_rpc_test() {
    local ip=$1
    MAIN_HTTP="http://$ip:8901"
    echo "=========== # paracross rpc test ============="
    echo "ip=$MAIN_HTTP"
    paracross_GetBlock2MainInfo
    chain33_unlock
}

#dapp_rpc_test $1
