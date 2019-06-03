#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
GAME_ID=""

addr_A=1PUiGcbsccfxW3zuvHXZBJfznziph5miAo
addr_B=1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX

# shellcheck source=/dev/null
source ../dapp-test-common.sh
set -x
hashlock_lock() {

    local secret=$1
    echo "========== # hashlock lock tx begin =========="

    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"hashlock","actionName":"HashlockLock", "payload":{"secret":"'"${secret}"'","amount":1000000000, "time":75,"toAddr":"'"${addr_B}"'", "returnAddr":"'"${addr_A}"'","fee":100000000}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" ${MAIN_HTTP}
    echo "txHash ${txhash}"
    echo "========== # hashlock lock tx end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

hashlock_send() {
    local secret=$1

    echo "========== # hashlock send tx begin =========="

    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"hashlock","actionName":"HashlockSend", "payload":{"secret":"'"${secret}"'","fee":100000000}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" ${MAIN_HTTP}
    echo "txHash ${txhash}"
    echo "========== # hashlock send tx end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

hashlock_unlock() {
    local secret=$1
    echo "========== # hashlock unlock tx begin =========="

    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"hashlock","actionName":"HashlockUnlock", "payload":{"secret":"'"${secret}"'","fee":100000000}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" ${MAIN_HTTP}
    echo "txHash ${txhash}"
    echo "========== # hashlock unlock tx end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        hashlock_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.hashlock"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        hashlock_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"hashlock"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi

    local from="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    chain33_SendToAddress "$from" "$hashlock_addr" 10000000000 ${MAIN_HTTP}

    from="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    chain33_SendToAddress "$from" "$hashlock_addr" 10000000000 ${MAIN_HTTP}
    chain33_BlockWait 1 "${MAIN_HTTP}"
}

function run_test() {
    chain33_QueryBalance "$addr_A" "${MAIN_HTTP}"
    chain33_QueryBalance "$addr_B" "${MAIN_HTTP}"
    hashlock_lock "abc"
    chain33_QueryBalance "$addr_A" "${MAIN_HTTP}"
    hashlock_send "abc"
    chain33_QueryBalance "$addr_B" "${MAIN_HTTP}"
    hashlock_unlock "abc"

    hashlock_lock "aef"
    chain33_QueryBalance "$addr_A" "${MAIN_HTTP}"

    sleep 75
    hashlock_unlock "aef"
    chain33_BlockWait 1 ${MAIN_HTTP}
    chain33_QueryBalance "$addr_A" "${MAIN_HTTP}"
}

function main() {
    MAIN_HTTP="$1"
    echo "=========== # Hashlock rpc test ============="
    echo "ip=$MAIN_HTTP"

    init
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Hashlock Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Hashlock Rpc Test Pass==============${NOC}"
    fi
}

main "$1"
set +x
