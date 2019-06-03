#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
GAME_ID=""

source ../dapp-test-common.sh

pokerbull_PlayRawTx() {
    echo "========== # pokerbull play tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Play","payload":{"gameId":"pokerbull-abc", "value":"1000000000", "round":1}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" ${MAIN_HTTP}
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

    chain33_SignRawTx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" ${MAIN_HTTP}
    echo "========== # pokerbull quit tx end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

pokerbull_ContinueRawTx() {
    echo "========== # pokerbull continue tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Continue","payload":{"gameId":"'$GAME_ID'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" ${MAIN_HTTP}
    echo "========== # pokerbull continue tx end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

pokerbull_StartRawTx() {
    echo "========== # pokerbull start tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Start","payload":{"value":"1000000000", "playerNum":"2"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" ${MAIN_HTTP}
    GAME_ID=$RAW_TX_HASH
    echo "========== # pokerbull start tx end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

pokerbull_QueryResult() {
    echo "========== # pokerbull query result begin =========="
    local req='"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByID","payload":{"gameId":"'$GAME_ID'"}}]'
    data=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")
    ok=$(jq '(.game.gameId == "$GAME_ID")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByAddr","payload":{"addr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}}]}' ${MAIN_HTTP} | jq -r ".result")
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

    local from="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    chain33_SendToAddress "$from" "$pokerbull_addr" 10000000000 ${MAIN_HTTP}

    from="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    chain33_SendToAddress "$from" "$pokerbull_addr" 10000000000 ${MAIN_HTTP}
    chain33_BlockWait 1 ${MAIN_HTTP}
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

main "$1"
