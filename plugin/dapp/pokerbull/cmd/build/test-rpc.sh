#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
GAME_ID=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

pokerbull_PlayRawTx() {
    echo "========== # pokerbull play tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Play","payload":{"gameId":"pokerbull-abc", "value":"1000000000", "round":1}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" ${MAIN_HTTP}
    echo "========== # pokerbull play tx end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

pokerbull_QuitRawTx() {
    echo "========== # pokerbull quit tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Quit","payload":{"gameId":"'$GAME_ID'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" ${MAIN_HTTP}
    echo "========== # pokerbull quit tx end =========="

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

pokerbull_ContinueRawTx() {
    echo "========== # pokerbull continue tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Continue","payload":{"gameId":"'$GAME_ID'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "0xa26038cbdd9e6fbfb85f2c3d032254755e75252b9edccbecc16d9ba117d96705" ${MAIN_HTTP}
    echo "========== # pokerbull continue tx end =========="

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

pokerbull_StartRawTx() {
    echo "========== # pokerbull start tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Start","payload":{"value":"1000000000", "playerNum":"2"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" ${MAIN_HTTP}
    GAME_ID=$RAW_TX_HASH
    echo "========== # pokerbull start tx end =========="

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

pokerbull_QueryResult() {
    echo "========== # pokerbull query result begin =========="
    local req='"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByID","payload":{"gameId":"'$GAME_ID'"}}]'
    data=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    ok=$(jq '(.game.gameId == "'"$GAME_ID"'")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByAddr","payload":{"addr":"14VkqML8YTRK4o15Cf97CQhpbnRUa6sJY4"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByStatus","payload":{"status":"3"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # pokerbull query result end =========="
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
    #main chain import pri key
    #14VkqML8YTRK4o15Cf97CQhpbnRUa6sJY4
    chain33_ImportPrivkey "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" "14VkqML8YTRK4o15Cf97CQhpbnRUa6sJY4" "pokerbull1" "${main_ip}"
    #1MuVM87DLigWhJxLJKvghTa1po4ZdWtDv1
    chain33_ImportPrivkey "0xa26038cbdd9e6fbfb85f2c3d032254755e75252b9edccbecc16d9ba117d96705" "1MuVM87DLigWhJxLJKvghTa1po4ZdWtDv1" "pokerbull2" "$main_ip"

    local pokerbull1="14VkqML8YTRK4o15Cf97CQhpbnRUa6sJY4"
    local pokerbull2="1MuVM87DLigWhJxLJKvghTa1po4ZdWtDv1"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$pokerbull1" 12000000000 "${main_ip}"
        chain33_QueryBalance "${pokerbull1}" "$main_ip"

        chain33_applyCoins "$pokerbull2" 12000000000 "${main_ip}"
        chain33_QueryBalance "${pokerbull2}" "$main_ip"
    else
        # tx fee
        chain33_applyCoins "$pokerbull1" 1000000000 "${main_ip}"
        chain33_QueryBalance "${pokerbull1}" "$main_ip"

        chain33_applyCoins "$pokerbull2" 1000000000 "${main_ip}"
        chain33_QueryBalance "${pokerbull2}" "$main_ip"
        local para_ip="${MAIN_HTTP}"
        #para chain import pri key
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
    MAIN_HTTP="$1"
    echo "=========== # pokerbull rpc test ============="
    echo "ip=$MAIN_HTTP"

    init
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Pokerbull Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Pokerbull Rpc Test Pass==============${NOC}"
    fi
}

chain33_debug_function main "$1"
