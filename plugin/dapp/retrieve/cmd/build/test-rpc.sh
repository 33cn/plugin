#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
CASE_ERR=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

source ../dapp-test-common.sh

retrieve_Backup() {
    echo "========== # retrieve backup begin =========="

    req='"method":"retrieve.CreateRawRetrieveBackupTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo","delayPeriod": 61}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138"
    echo "========== # retrieve backup end =========="

    block_wait 1
}

retrieve_Prepare() {
    echo "========== # retrieve prepare begin =========="

    req='"method":"retrieve.CreateRawRetrievePrepareTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989"
    echo "========== # retrieve prepare end =========="

    block_wait 1
}

retrieve_Perform() {
    echo "========== # retrieve perform begin =========="

    req='"method":"retrieve.CreateRawRetrievePerformTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989"
    echo "========== # retrieve perform end =========="

    block_wait 1
}

retrieve_Cancel() {
    echo "========== # retrieve cancel begin =========="

    req='"method":"retrieve.CreateRawRetrieveCancelTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138"
    echo "========== # retrieve cancel end =========="

    block_wait 1
}

retrieve_QueryResult() {
    echo "========== # retrieve query result begin =========="

    local status=$1

    req='"method":"Chain33.Query","params":[{"execer":"retrieve","funcName":"GetRetrieveInfo","payload":{"backupAddress":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX", "defaultAddress":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}}]'
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
    echo "retrieveaddr=$retrieve_addr"

    from="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    Chain33_SendToAddress "$from" "$retrieve_addr" 1000000000

    from="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    Chain33_SendToAddress "$from" "$retrieve_addr" 1000000000
    block_wait 1
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
