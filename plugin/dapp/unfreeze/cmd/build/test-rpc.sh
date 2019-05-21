#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# base functions
# $2=0 means true, other false
function echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="err"
    fi

}

function Chain33_SendToAddress() {
    local from="$1"
    local to="$2"
    local amount=$3
    local req='{"method":"Chain33.SendToAddress", "params":[{"from":"'"$from"'","to":"'"$to"'", "amount":'"$amount"', "note":"test\n"}]}'
    #    echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.hash|length==66)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    hash=$(jq '(.result.hash)' <<<"$resp")
    echo "hash=$hash"
    #    query_tx "$hash"

}

function sign_raw_tx() {
    txHex="$1"
    priKey="$2"
    req='{"method":"Chain33.SignRawTx","params":[{"privkey":"'"$priKey"'","txHex":"'"$txHex"'","expire":"120s"}]}'
    #    echo "#request SignRawTx: $req"
    signedTx=$(curl -ksd "$req" ${MAIN_HTTP} | jq -r ".result")
    #    echo "signedTx=$signedTx"
    if [ "$signedTx" != null ]; then
        send_tx "$signedTx"
    else
        echo "signedTx null error"
    fi
}

function send_tx() {
    signedTx=$1
    req='{"method":"Chain33.SendTransaction","params":[{"token":"BTY","data":"'"$signedTx"'"}]}'
    #    echo "#request sendTx: $req"
    #    curl -ksd "$req" ${MAIN_HTTP}
    resp=$(curl -ksd "$req" ${MAIN_HTTP})
    err=$(jq '(.error)' <<<"$resp")
    txhash=$(jq -r ".result" <<<"$resp")
    if [ "$err" == null ]; then
        #   echo "tx hash: $txhash"
        query_tx "$txhash"
    else
        echo "send tx error:$err"
    fi

}

function block_wait() {
    req='{"method":"Chain33.GetLastHeader","params":[{}]}'
    cur_height=$(curl -ksd "$req" ${MAIN_HTTP} | jq ".result.height")
    expect=$((cur_height + ${1}))
    local count=0
    while true; do
        new_height=$(curl -ksd "$req" ${MAIN_HTTP} | jq ".result.height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 1
    done
    echo "wait new block $count s, cur height=$expect,old=$cur_height"
}

function query_tx() {
    block_wait 1
    txhash="$1"
    # echo "req=$req"
    local times=10
    while true; do
        req='{"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]}'
        ret=$(curl -ksd "$req" ${MAIN_HTTP})
        tx=$(jq -r ".result.tx.hash" <<<"$ret")
        echo "====query tx= ${1}, return=$ret "
        if [ "${tx}" != "${1}" ]; then
            block_wait 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "====query tx=$1 failed"
                echo "req=$req"
                curl -ksd "$req" ${MAIN_HTTP}
                exit 1
            fi
        else
            exec_err=$(jq '(.result.receipt.logs[0].tyName == "LogErr")' <<<"$ret")
            [ "$exec_err" != true ]
            echo "====query tx=$1  success"
            break
        fi
    done
}

function query_unfreezeID() {
    block_wait 1

    # echo "req=$req"
    local times=10
    while true; do
        req='{"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]}'
        ret=$(curl -ksd "$req" ${MAIN_HTTP})
        tx=$(jq -r ".result.tx.hash" <<<"$ret")
        echo "====query tx= ${txhash}, return=$ret "
        if [ "${tx}" != "${txhash}" ]; then
            block_wait 1
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

    Chain33_SendToAddress "$owner" "$exec_addr" 500000000
    Chain33_SendToAddress "$beneficiary" "$exec_addr" 500000000
    block_wait 1
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
    sign_raw_tx "$rawtx" "$owner_key"
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
    sign_raw_tx "$rawtx" "${beneficiary_key}"
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
    sign_raw_tx "$rawtx" "$owner_key"
    block_wait 2
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
