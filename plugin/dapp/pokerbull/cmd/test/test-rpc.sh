#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -e
set -o pipefail

MAIN_HTTP=""
GAME_ID=""

source ../dapp-test-common.sh

pokerbull_PlayRawTx() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Play","payload":{"gameId":"pokerbull-abc", "value":"1000000000", "round":1}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" ${MAIN_HTTP} "$FUNCNAME"
}

pokerbull_QuitRawTx() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Quit","payload":{"gameId":"'$GAME_ID'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" ${MAIN_HTTP} "$FUNCNAME"
}

pokerbull_ContinueRawTx() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Continue","payload":{"gameId":"'$GAME_ID'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "0xa26038cbdd9e6fbfb85f2c3d032254755e75252b9edccbecc16d9ba117d96705" ${MAIN_HTTP} "$FUNCNAME"
}

pokerbull_StartRawTx() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Start","payload":{"value":"1000000000", "playerNum":"2"}}]}' ${MAIN_HTTP} | jq -r ".result")
    req='{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.result.txs[0].execer != null)' "$FUNCNAME"
    chain33_SignAndSendTx "$tx" "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" ${MAIN_HTTP}
    GAME_ID=$RAW_TX_HASH
    chain33_BlockWait 1 "${MAIN_HTTP}"
}

pokerbull_QueryResult() {
    req='{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByID","payload":{"gameId":"'$GAME_ID'"}}]}'
    resok='(.result.game.gameId == "'"$GAME_ID"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"

    req='{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByAddr","payload":{"addr":"14VkqML8YTRK4o15Cf97CQhpbnRUa6sJY4"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.result != null)' "$FUNCNAME"

    req='{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByStatus","payload":{"status":"3"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.result != null)' "$FUNCNAME"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        pokerbull_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.pokerbull"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        pokerbull_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"pokerbull"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi

    local main_ip=${MAIN_HTTP//8901/8801}
    chain33_ImportPrivkey "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" "14VkqML8YTRK4o15Cf97CQhpbnRUa6sJY4" "pokerbull1" "${main_ip}"
    chain33_ImportPrivkey "0xa26038cbdd9e6fbfb85f2c3d032254755e75252b9edccbecc16d9ba117d96705" "1MuVM87DLigWhJxLJKvghTa1po4ZdWtDv1" "pokerbull2" "$main_ip"

    local pokerbull1="14VkqML8YTRK4o15Cf97CQhpbnRUa6sJY4"
    local pokerbull2="1MuVM87DLigWhJxLJKvghTa1po4ZdWtDv1"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$pokerbull1" 12000000000 "${main_ip}"
        chain33_QueryBalance "${pokerbull1}" "$main_ip"

        chain33_applyCoins "$pokerbull2" 12000000000 "${main_ip}"
        chain33_QueryBalance "${pokerbull2}" "$main_ip"
    else
        chain33_applyCoins "$pokerbull1" 1000000000 "${main_ip}"
        chain33_QueryBalance "${pokerbull1}" "$main_ip"

        chain33_applyCoins "$pokerbull2" 1000000000 "${main_ip}"
        chain33_QueryBalance "${pokerbull2}" "$main_ip"
        local para_ip="${MAIN_HTTP}"
        chain33_ImportPrivkey "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" "14VkqML8YTRK4o15Cf97CQhpbnRUa6sJY4" "pokerbull1" "$para_ip"
        chain33_ImportPrivkey "0xa26038cbdd9e6fbfb85f2c3d032254755e75252b9edccbecc16d9ba117d96705" "1MuVM87DLigWhJxLJKvghTa1po4ZdWtDv1" "pokerbull2" "$para_ip"

        chain33_applyCoins "$pokerbull1" 12000000000 "${para_ip}"
        chain33_QueryBalance "${pokerbull1}" "$para_ip"
        chain33_applyCoins "$pokerbull2" 12000000000 "${para_ip}"
        chain33_QueryBalance "${pokerbull2}" "$para_ip"
    fi

    chain33_SendToAddress "$pokerbull1" "$pokerbull_addr" 10000000000 ${MAIN_HTTP}
    chain33_QueryExecBalance "${pokerbull1}" "pokerbull" "$MAIN_HTTP"
    chain33_SendToAddress "$pokerbull2" "$pokerbull_addr" 10000000000 ${MAIN_HTTP}
    chain33_QueryExecBalance "${pokerbull2}" "pokerbull" "$MAIN_HTTP"

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

function run_test() {
    pokerbull_StartRawTx

    pokerbull_ContinueRawTx

    pokerbull_QuitRawTx

    pokerbull_PlayRawTx

    pokerbull_QueryResult
}

function main() {
    chain33_RpcTestBegin pokerbull
    MAIN_HTTP="$1"
    echo "ip=$MAIN_HTTP"

    init
    run_test
    chain33_RpcTestRst pokerbull "$CASE_ERR"
}

chain33_debug_function main "$1"
