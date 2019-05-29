#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
CASE_ERR=""
GAME_ID=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

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

pokerbull_PlayRawTx() {
    echo "========== # pokerbull play tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Play","payload":{"gameId":"pokerbull-abc", "value":"1000000000", "round":1}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" "play"
    echo "========== # pokerbull play tx end =========="

    block_wait 1
}

pokerbull_QuitRawTx() {
    echo "========== # pokerbull quit tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Quit","payload":{"gameId":"'$GAME_ID'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" "quit"
    echo "========== # pokerbull quit tx end =========="

    block_wait 1
}

pokerbull_ContinueRawTx() {
    echo "========== # pokerbull continue tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Continue","payload":{"gameId":"'$GAME_ID'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588" "continue"
    echo "========== # pokerbull continue tx end =========="

    block_wait 1
}

pokerbull_StartRawTx() {
    echo "========== # pokerbull start tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"pokerbull","actionName":"Start","payload":{"value":"1000000000", "playerNum":"2"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" "start"
    echo "========== # pokerbull start tx end =========="

    block_wait 1
}

pokerbull_QueryResult() {
    echo "========== # pokerbull query result begin =========="
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByID","payload":{"gameId":"'$GAME_ID'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    ok=$(jq '(.game.gameId == "'$GAME_ID'")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByAddr","payload":{"addr":"14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"pokerbull","funcName":"QueryGameByStatus","payload":{"status":"3"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # pokerbull query result end =========="
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
    hash=$(jq '(.result.hash)' <<<"$resp")
    echo "hash=$hash"
    echo "send '$3' from '$1' to '$2' end"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        pokerbull_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.pokerbull"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        pokerbull_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"pokerbull"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "pokerbulladdr=$pokerbull_addr"

    from="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    Chain33_SendToAddress "$from" "$pokerbull_addr" 10000000000

    from="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    Chain33_SendToAddress "$from" "$pokerbull_addr" 10000000000
    block_wait 1
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
