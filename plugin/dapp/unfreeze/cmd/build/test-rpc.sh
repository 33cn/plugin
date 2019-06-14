#!/usr/bin/env bash
# shellcheck disable=SC2128

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
txhash=""

function query_unfreezeID() {
    chain33_BlockWait 1 "$MAIN_HTTP"

    # echo "req=$req"
    local times=10
    while true; do
        req='{"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]}'
        ret=$(curl -ksd "$req" ${MAIN_HTTP})
        tx=$(jq -r ".result.tx.hash" <<<"$ret")
        echo "====query tx= ${txhash}, return=$ret "
        if [ "${tx}" != "${txhash}" ]; then
            chain33_BlockWait  1 "${MAIN_HTTP}"
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "====query tx=$txhash failed"
                echo "req=$req"
                curl -ksd "$req" ${MAIN_HTTP}
                exit 1
            fi
        else
            unfreeze_id=$(jq '(.result.receipt.logs['"$uid_index"'].log.current.unfreezeID)' <<<"$ret")
            #echo "${unfreeze_id}"
            unfreeze_id2=${unfreeze_id#\"mavl-unfreeze-}
            uid=${unfreeze_id2%\"}
            echo "====query tx=$txhash  success"
            break
        fi
    done
}

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    exec_name="unfreeze"
    uid_index=2
    if [ "$ispara" == true ]; then
        exec_name="user.p.para."${exec_name}
        uid_index=1
    fi
    exec_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'${exec_name}'"}]}' ${MAIN_HTTP} | jq -r ".result")
    echo "exec_addr=${exec_addr}"

    beneficiary=12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
    beneficiary_key=0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01
    owner=14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    owner_key=CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944
    #unfreeze_exec_addr=15YsqAuXeEXVHgm6RVx4oJaAAnhtwqnu3H

    chain33_SendToAddress "$owner" "$exec_addr" 500000000 "${MAIN_HTTP}"
    chain33_SendToAddress "$beneficiary" "$exec_addr" 500000000 "${MAIN_HTTP}"
    chain33_BlockWait  1 "${MAIN_HTTP}"
}

function CreateRawUnfreezeCreate() {
    req='{"jsonrpc": "2.0", "method" :  "unfreeze.CreateRawUnfreezeCreate" , "params":[{"startTime":10000,"assetExec":"coins","assetSymbol":"bty","totalCount":400000000,"beneficiary":"'$beneficiary'","means":"FixAmount","fixAmount": {"period":10,"amount":1000000}}]}'
    # echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "$owner_key" "${MAIN_HTTP}"
    query_unfreezeID
}

function CreateRawUnfreezeWithdraw() {
    sleep 10
    req='{"method":"unfreeze.CreateRawUnfreezeWithdraw","params":[{"unfreezeID":"'${uid}'"}]}'
    # echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${beneficiary_key}" "${MAIN_HTTP}"
}

function CreateRawUnfreezeTerminate() {
    req='{"method":"unfreeze.CreateRawUnfreezeTerminate","params":[{"unfreezeID":"'${uid}'"}]}'
    # echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "$owner_key" "${MAIN_HTTP}"
    chain33_BlockWait 2 "${MAIN_HTTP}"
}

function GetUnfreeze() {
    req='{"method":"unfreeze.GetUnfreeze","params":[{"data":"'${uid}'"}]}'
    # echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

function GetUnfreezeWithdraw() {
    req='{"method":"unfreeze.GetUnfreezeWithdraw","params":[{"data":"'${uid}'"}]}'
    # echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

function run_testcases() {
    CreateRawUnfreezeCreate

    CreateRawUnfreezeWithdraw
    GetUnfreeze
    GetUnfreezeWithdraw

    CreateRawUnfreezeTerminate
}

function debug_function() {
    set -x
    eval "$@"
    set +x
}

function rpc_test() {
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases

    if [ -n "$CASE_ERR" ]; then
        echo "=======unfreeze rpc test  error ==========="
        exit 1
    else
        echo "====== unfreeze rpc test  pass ==========="
    fi
}

debug_function rpc_test "$1"
