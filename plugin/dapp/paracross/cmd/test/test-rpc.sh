#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""

# $2=0 means true, other false
echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo "$1 ok"
    else
        echo "$1 err"
        CASE_ERR="err"
    fi

}

chain33_lock() {
    ok=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Lock","params":[]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.isOK")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

chain33_unlock() {
    ok=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.UnLock","params":[{"passwd":"1314fuzamei","timeout":0}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.isOK")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

function run_testcases() {
    chain33_lock
    chain33_unlock

}
function paracross_rpc_test() {
    MAIN_HTTP="$1"
    echo "=========== # paracross rpc test ============="
    echo "ip=$MAIN_HTTP"

    run_testcases

    if [ -n "$CASE_ERR" ]; then
        echo "paracross there some case error"
        exit 1
    fi
}

paracross_rpc_test "$1"
