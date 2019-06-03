#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""

source ../dapp-test-common.sh

retrieve_Backup() {
    echo "========== # retrieve backup begin =========="

    local req='"method":"retrieve.CreateRawRetrieveBackupTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo","delayPeriod": 61}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    local reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" ${MAIN_HTTP}
    echo "========== # retrieve backup end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

retrieve_Prepare() {
    echo "========== # retrieve prepare begin =========="

    local req='"method":"retrieve.CreateRawRetrievePrepareTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    local reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" ${MAIN_HTTP}
    echo "========== # retrieve prepare end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

retrieve_Perform() {
    echo "========== # retrieve perform begin =========="

    local req='"method":"retrieve.CreateRawRetrievePerformTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    local reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" ${MAIN_HTTP}
    echo "========== # retrieve perform end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

retrieve_Cancel() {
    echo "========== # retrieve cancel begin =========="

    local req='"method":"retrieve.CreateRawRetrieveCancelTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    local reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" ${MAIN_HTTP}
    echo "========== # retrieve cancel end =========="

    chain33_BlockWait 1 ${MAIN_HTTP}
}

retrieve_QueryResult() {
    echo "========== # retrieve query result begin =========="

    local status=$1

    local req='"method":"Chain33.Query","params":[{"execer":"retrieve","funcName":"GetRetrieveInfo","payload":{"backupAddress":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX", "defaultAddress":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}}]'
    data=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")
    ok=$(jq '(.status == '"$status"')' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # retrieve query result end =========="
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        retrieve_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.retrieve"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        retrieve_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"retrieve"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi

    local from="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    chain33_SendToAddress "$from" "$retrieve_addr" 1000000000 ${MAIN_HTTP}

    from="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    chain33_SendToAddress "$from" "$retrieve_addr" 1000000000 ${MAIN_HTTP}
    chain33_BlockWait 1 ${MAIN_HTTP}
}

function run_test() {
    retrieve_Backup
    retrieve_QueryResult 1

    retrieve_Prepare
    retrieve_QueryResult 2

    retrieve_Cancel
    retrieve_QueryResult 4

    retrieve_Backup
    retrieve_QueryResult 1

    retrieve_Prepare
    retrieve_QueryResult 2

    sleep 61
    retrieve_Perform
    retrieve_QueryResult 3
}

function main() {
    MAIN_HTTP="$1"
    echo "=========== # retrieve rpc test ============="
    echo "ip=$MAIN_HTTP"

    init
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============retrieve Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============retrieve Rpc Test Pass==============${NOC}"
    fi
}

main "$1"
