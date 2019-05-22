#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
CASE_ERR=""
START_TX=""
CONTINUE_TX=""
QUIT_TX=""
PLAY_TX=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

echo_rst() {
    if [ "$2" == true ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="FAIL"
    fi

}

pokerbull_PlayRawTx() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Play","payload":{"gameID":"'$START_TX'", "value":"1000000000", "round":1}}]}' ${MAIN_HTTP} | jq -r ".result")
    PLAY_TX=$tx

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer == "Play")' <<<"$data")

    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

pokerbull_QuitRawTx() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Quit","payload":{"gameID":"'$START_TX'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    QUIT_TX=$tx

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer == "Quit")' <<<"$data")

    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

pokerbull_ContinueRawTx() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Continue","payload":{"gameID":"'$START_TX'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    CONTINUE_TX=$tx

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer == "Continue")' <<<"$data")

    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

pokerbull_StartRawTx() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Start","payload":{"value":"1000000000", "playerNum":"2"}}]}' ${MAIN_HTTP} | jq -r ".result")
    START_TX=$tx

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer == "Start")' <<<"$data")

    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

Chain33_SendRawTx() {
    local req='"method":"Chain33.SignRawTx", "params":[{"data":"'"$1"'", "addr":"'"$2"'", "expire":"3600"}]'

    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")

    #    echo "#response: $resp"
    ok=$(jq '(.error|not)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    local reqSend='"method":"Chain33.SendTransaction", "params":[{"data":"'"$resp"'"}]'

    #    echo "#request: $req"
    respSend=$(curl -ksd "{$req}" "${MAIN_HTTP}")

    #    echo "#response: $resp"
    okSend=$(jq '(.error|not)' <<<"$respSend")
    [ "$okSend" == true ]
    echo_rst "$FUNCNAME" "$?"
}

Chain33_SendToAddress() {
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
    #    query_tx "$hash"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    local relay_addr=""
    if [ "$ispara" == true ]; then
        pokerbull_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.pokerbull"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        chain33_ImportPrivkey "${MAIN_HTTP}" "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588" "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
        pokerbull_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"pokerbull"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "pokerbulladdr=$pokerbull_addr"

    from="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    Chain33_SendToAddress "$from" "$pokerbull_addr" 10000000000

    from="1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
    Chain33_SendToAddress "$from" "$pokerbull_addr" 10000000000
    block_wait 1
}

function run_test() {
    pokerbull_StartRawTx
    Chain33_SendRawTx "$START_TX" "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    ##pokerbull_ContinueRawTx
    ##pokerbull_QuitRawTx
    ##pokerbull_PlayRawTx
    ##pokerbull_QueryResult

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
