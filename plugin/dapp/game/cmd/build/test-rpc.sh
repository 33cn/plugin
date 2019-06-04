#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
GAME_ID=""
PASSWD="ABCD"
HASH_VALUE=$(echo -n "ABCD1" | sha256sum | awk '{print $1}')

ACCOUNT_A="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
PRIVA_A="56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138"

ACCOUNT_B="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
PRIVA_B="2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989"

EXECTOR=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

function chain33_GetExecAddr() {
    #获取GAME合约地址
    local exector=$1
    local req='"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${exector}"'"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    # GAME_ADDR=$(echo "${res}" | jq -r ".result")
    echo_rst "$FUNCNAME" "$?"
}

function CreateGameTx() {
    local amount=$1
    local hash_value=$2
    local req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"createGame", "payload":{"amount": '"${amount}"',"hashType":"sha256","hashValue":"'"${hash_value}"'"}}]'
    echo "#request: $req"

    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    rawTx=$(echo "${resp}" | jq -r ".result")
    if [ "$rawTx" == "null" ]; then
        echo_rst "CreateGame createRawTx" 1
    fi

    chain33_SignRawTx "${rawTx}" "${PRIVA_A}" "${MAIN_HTTP}"
    GAME_ID=$RAW_TX_HASH

    echo_rst "CreateGame query_tx" "$?"
}

function MatchGameTx() {
    local gameId=$1
    local req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"matchGame", "payload":{"gameId": "'"${gameId}"'","guess":2}}]'

    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"

    rawTx=$(echo "${resp}" | jq -r ".result")
    if [ "$rawTx" == "null" ]; then
        echo_rst "MatchGame createRawTx" 1
    fi

    chain33_SignRawTx "${rawTx}" "${PRIVA_B}" "${MAIN_HTTP}"
    echo_rst "MatchGame query_tx" "$?"
}

function CloseGameTx() {
    local gameId=$1
    local secret=$2
    local req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"closeGame", "payload":{"gameId": "'"${gameId}"'","secret":"'"${secret}"'","result":1}}]'

    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"

    rawTx=$(echo "${resp}" | jq -r ".result")
    if [ "$rawTx" == "null" ]; then
        echo_rst "CloseGame createRawTx" 1
    fi

    chain33_SignRawTx "${rawTx}" "${PRIVA_A}" "${MAIN_HTTP}"
    echo_rst "CloseGame query_tx" "$?"
}

function CancleGameTx() {
    local gameId=$1
    local req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"cancelGame", "payload":{"gameId": "'"${gameId}"'"}}]'

    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"

    rawTx=$(echo "${resp}" | jq -r ".result")
    if [ "$rawTx" == "null" ]; then
        echo_rst "CancleGame createRawTx" 1
    fi

    chain33_SignRawTx "${rawTx}" "${PRIVA_A}" "${MAIN_HTTP}"
    echo_rst "CancleGame query_tx" "$?"
}

function QueryGameByStatus() {
    local status=$1
    local req='"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"QueryGameListByStatusAndAddr","payload":{"status":'"${status}"',"address":""}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    GAMES=$(echo "${resp}" | jq -r ".result.games")
    echo "${GAMES}"
    echo_rst "$FUNCNAME" "$?"
}

function QueryGameByGameId() {
    local gameId=$1
    local status=$2
    local req='"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"QueryGameById","payload":{"gameId":"'"${gameId}"'"}}]'
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

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    local game_addr=""
    if [ "$ispara" == "true" ]; then
        EXECTOR="user.p.para.game"
        game_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.game"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        EXECTOR="game"
        game_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"game"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "gameAddr=${game_addr}"

    chain33_SendToAddress "${ACCOUNT_B}" "$game_addr" 5000000000 "${MAIN_HTTP}"

    chain33_SendToAddress "${ACCOUNT_A}" "$game_addr" 5000000000 "${MAIN_HTTP}"

    chain33_BlockWait 1 "$MAIN_HTTP"
}

function run_test() {
    local ip=$1
    CreateGameTx 1000000000 "${HASH_VALUE}"

    QueryGameByGameId "${GAME_ID}" 1

    QueryGameByStatus 1

    MatchGameTx "${GAME_ID}"

    QueryGameByGameId "${GAME_ID}" 2

    QueryGameByStatus 2

    CloseGameTx "${GAME_ID}" "${PASSWD}"

    QueryGameByGameId "${GAME_ID}" 4

    QueryGameByStatus 4

    CreateGameTx 500000000 "${HASH_VALUE}"

    QueryGameByGameId "${GAME_ID}" 1

    CancleGameTx "${GAME_ID}"

    QueryGameByGameId "${GAME_ID}" 3

    QueryGameByStatus 3
}

function main() {
    local ip=$1
    MAIN_HTTP=$ip
    echo "=========== # game rpc test ============="
    echo "main_ip=$MAIN_HTTP"

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
