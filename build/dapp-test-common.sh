#!/usr/bin/env bash

echo_rst() {
    if [ "$2" == 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="FAIL"
    fi

}

function block_wait() {
    local req='"method":"Chain33.GetLastHeader","params":[]'
    cur_height=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq ".result.height")
    expect=$((cur_height + ${1}))
    local count=0
    while true; do
        new_height=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq ".result.height")
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
    local txhash="$1"
    local req='"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]'
    # echo "req=$req"
    local times=10
    while true; do
        ret=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.tx.hash")
        echo "====query tx= ${1}, return=$ret "
        if [ "${ret}" != "${1}" ]; then
            block_wait 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "====query tx=$1 failed"
                echo "req=$req"
                curl -ksd "{$req}" ${MAIN_HTTP}
                exit 1
            fi
        else
            echo "====query tx=$1  success"
            break
        fi
    done
}

sendTx() {
    echo "send tx '$1' begin"
    signedTx=$1
    type=$2
    local req='"method":"Chain33.SendTransaction","params":[{"token":"BTY","data":"'"$signedTx"'"}]'
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    err=$(jq '(.error)' <<<"$resp")
    txhash=$(jq -r ".result" <<<"$resp")
    if [ "$err" == null ]; then
        query_tx "$txhash"
        if [ "$type" == "start" ]; then
            GAME_ID=$txhash
        fi
    else
        echo "send tx error:$err"
    fi
    echo "send tx '$1' end"
}

Chain33_SendToAddress() {
    echo "send '$3' from '$1' to '$2' begin"
    local from="$1"
    local to="$2"
    local amount=$3
    local req='"method":"Chain33.SendToAddress", "params":[{"from":"'"$from"'","to":"'"$to"'", "amount":'"$amount"', "note":"test\n"}]'
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    ok=$(jq '(.error|not) and (.result.hash|length==66)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    hash=$(jq -r ".result.hash" <<<"$resp")
    query_tx "$hash"
    echo "send '$3' from '$1' to '$2' end"
}

chain33_ImportPrivkey() {
    local pri=$2
    local acc=$3
    local req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"'"$pri"'", "label":"pokerbullimportkey1"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$1")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.label=="pokerbullimportkey1") and (.result.acc.addr == "'"$acc"'")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

signrawtx() {
    echo "sign tx '$1' begin"
    txHex="$1"
    priKey="$2"
    type="$3"
    local req='"method":"Chain33.SignRawTx","params":[{"privkey":"'"$priKey"'","txHex":"'"$txHex"'","expire":"120s"}]'
    signedTx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")
    if [ "$signedTx" != null ]; then
        sendTx "$signedTx" "$3"
    else
        echo "signedTx null error"
    fi
    echo "sign tx '$1' end"
}
