#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
PARA_HTTP=""
CASE_ERR=""
GAME_ID=""
PASSWD="ABCD"
HASH_VALUE=$(echo -n "ABCD1" | sha256sum | awk '{print $1}')
create_txHash=""
match_txHash=""
close_txHash=""
ACCOUNT_A="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
PRIVA_A="cc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"
ACCOUNT_B="19MJmA7GcE1NfMwdGqgLJioBjVbzQnVYvR"
PRIVA_B="5072a3b6ed612845a7c00b88b38e4564093f57ce652212d6e26da9fded83e951"
EXECTOR=""
#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

function echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="FAIL"
    fi

}

function chain33_GetExecAddr() {
    #获取GAME合约地址
    local exector=$1
    local req='"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${exector}"'"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    GAME_ADDR=$(echo "${res}" | jq -r ".result")
    echo_rst "$FUNCNAME" "$?"
}

function CreateGameTx() {
    local amount=$1
    local hash_value=$2
    local exector=$3
    local addr=$4
    local req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${exector}"'", "actionName":"createGame", "payload":{"amount": '"${amount}"',"hashType":"sha256","hashValue":"'"${hash_value}"'"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    rawTx=$(echo "${resp}" | jq -r ".result")
    if [ "$rawTx" == "null" ]; then
        echo_rst "CreateGame createRawTx" 1
    fi
    signRawTx "${rawTx}" "${ACCOUNT_A}"
    echo_rst "CreateGame signRawTx" "$?"
    sendSignedTx
    echo_rst "CreateGame sendSignedTx" "$?"
    GAME_ID="${txHash}"
    create_txHash="${txHash}"
    query_tx "${txHash}"
    echo_rst "CreateGame query_tx" "$?"
}

function MatchGameTx() {
    local gameId=$1
    local exector=$2
    local req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${exector}"'", "actionName":"matchGame", "payload":{"gameId": "'"${gameId}"'","guess":2}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    rawTx=$(echo "${resp}" | jq -r ".result")
    if [ "$rawTx" == "null" ]; then
        echo_rst "MatchGame createRawTx" 1
    fi
    signRawTx "${rawTx}" "${ACCOUNT_B}"
    echo_rst "MatchGame signRawTx" "$?"
    sendSignedTx
    echo_rst "MatchGame sendSignedTx" "$?"
    match_txHash="${txHash}"
    query_tx "${txHash}"
    echo_rst "MatchGame query_tx" "$?"
}

function CloseGameTx() {
    local gameId=$1
    local secret=$2
    local exector=$3
    local req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${exector}"'", "actionName":"closeGame", "payload":{"gameId": "'"${gameId}"'","secret":"'"${secret}"'","result":1}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    rawTx=$(echo "${resp}" | jq -r ".result")
    if [ "$rawTx" == "null" ]; then
        echo_rst "CloseGame createRawTx" 1
    fi
    signRawTx "${rawTx}" "${ACCOUNT_A}"
    echo_rst "CloseGame signRawTx" "$?"
    sendSignedTx
    echo_rst "CloseGame sendSignedTx" "$?"
    close_txHash="${txHash}"
    query_tx "${txHash}"
    echo_rst "CloseGame query_tx" "$?"
}

function CancleGameTx() {
    local gameId=$1
    local exector=$2
    local req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${exector}"'", "actionName":"cancelGame", "payload":{"gameId": "'"${gameId}"'"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    rawTx=$(echo "${resp}" | jq -r ".result")
    if [ "$rawTx" == "null" ]; then
        echo_rst "CancleGame createRawTx" 1
    fi
    signRawTx "${rawTx}" "${ACCOUNT_A}"
    echo_rst "CancleGame signRawTx" "$?"
    sendSignedTx
    echo_rst "CancleGame sendSignedTx" "$?"
    close_txHash="${txHash}"
    query_tx "${txHash}"
    echo_rst "CancleGame query_tx" "$?"
}

function QueryGameByStatus() {
    local exector=$1
    local req='"method":"Chain33.Query","params":[{"execer":"'"${exector}"'","funcName":"QueryGameListByStatusAndAddr","payload":{"status":1,"address":""}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    GAMES=$(echo "${resp}" | jq -r ".result.games")
    echo_rst "$FUNCNAME" "$?"
}

function QueryGameByGameId() {
    local gameId=$1
    local exector=$2
    local status=$3
    local req='"method":"Chain33.Query","params":[{"execer":"'"${exector}"'","funcName":"QueryGameById","payload":{"gameId":"'"${gameId}"'"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    STATUS=$(echo "${resp}" | jq -r ".result.game.status")
    if [ "${STATUS}" -ne "${status}" ]; then
        echo "status is not equal"
        echo_rst "QueryGameByGameId" 1
        return 0
    fi
    echo_rst "QueryGameByGameId" 0
}

function chain33_ImportPrivkey() {
    local pri=$2
    local acc=$3
    local req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"'"$pri"'", "label":"gameB"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$1")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.label=="gameB") and (.result.acc.addr == "'"$acc"'")' <<<"$resp")
    [ "$ok" == true ]
    # echo_rst "$FUNCNAME" "$?"
}

function Chain33_SendToAddress() {
    local from="$1"
    local to="$2"
    local amount=$3
    local req='"method":"Chain33.SendToAddress", "params":[{"from":"'"$from"'","to":"'"$to"'", "amount":'"$amount"', "note":"test\n"}]'
    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.hash|length==66)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    hash=$(jq '(.result.hash)' <<<"$resp")
    echo "hash=$hash"
    # query_tx "$hash"
}

function chain33_unlock() {
    ok=$(curl -k -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.UnLock","params":[{"passwd":"1314fuzamei","timeout":0}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.isOK")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
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

function signRawTx() {
    unsignedTx=$1
    addr=$2
    signedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SignRawTx","params":[{"addr":"'${addr}'","txHex":"'${unsignedTx}'","expire":"120s","fee":10000000}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "$signedTx" == "null" ]; then
        return 1
    else
        return 0
    fi
}

function sendSignedTx() {
    txHash=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"token":"","data":"'${signedTx}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "$txHash" == "null" ]; then
        return 1
    else
        return 0
    fi
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

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    from="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"

    chain33_ImportPrivkey "${MAIN_HTTP}" "${PRIVA_B}" "${ACCOUNT_B}"

    local game_addr=""
    if [ "$ispara" == "true" ]; then
        EXECTOR="user.p.para.game"
        game_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.game"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        EXECTOR="game"
        game_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"game"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "gameAddr=${game_addr}"

    Chain33_SendToAddress "${ACCOUNT_B}" "$game_addr" 5000000000

    Chain33_SendToAddress "${ACCOUNT_A}" "$game_addr" 5000000000

    block_wait 1
}

function run_test() {
    local ip=$1
    CreateGameTx 1000000000 "${HASH_VALUE}" "${EXECTOR}"

    QueryGameByGameId "${GAME_ID}" "${EXECTOR}" 1

    MatchGameTx "${GAME_ID}" "${EXECTOR}"

    QueryGameByGameId "${GAME_ID}" "${EXECTOR}" 2

    CloseGameTx "${GAME_ID}" 1 "${EXECTOR}"

    QueryGameByGameId "${GAME_ID}" "${EXECTOR}" 4

    CreateGameTx 500000000 "${HASH_VALUE}" "${EXECTOR}"

    QueryGameByGameId "${GAME_ID}" "${EXECTOR}" 1

    CancleGameTx "${GAME_ID}" "${EXECTOR}"

    QueryGameByGameId "${GAME_ID}" "${EXECTOR}" 3
}

function main() {
    local ip=$1
    MAIN_HTTP=$ip
    echo "=========== # game rpc test ============="
    echo "main_ip=$MAIN_HTTP"

    Chain33_SendToAddress "${ACCOUNT_A}" "${ACCOUNT_B}" 20000000000

    block_wait 1

    init

    run_test "$MAIN_HTTP"

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============game Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============game Rpc Test Pass==============${NOC}"
    fi
}

main "$1"
