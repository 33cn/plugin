#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

retrieve_Backup() {
    echo "========== # retrieve backup begin =========="

    local req='"method":"retrieve.CreateRawRetrieveBackupTx","params":[{"backupAddr":"13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx","defaultAddr":"1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL","delayPeriod": 61}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    local reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "0x3665fa66d1a17d2fc319a45250c8c8b9302ae0c393c2e39f2ef3b2f6bc40a42d" ${MAIN_HTTP}
    echo "========== # retrieve backup end =========="

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

retrieve_Prepare() {
    echo "========== # retrieve prepare begin =========="

    local req='"method":"retrieve.CreateRawRetrievePrepareTx","params":[{"backupAddr":"13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx","defaultAddr":"1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    local reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "0xed8a078ee44eac473bd1d5c971e231c255badf7f0c2fbdbe31ef34669c441d6f" ${MAIN_HTTP}
    echo "========== # retrieve prepare end =========="

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

retrieve_Perform() {
    echo "========== # retrieve perform begin =========="

    local req='"method":"retrieve.CreateRawRetrievePerformTx","params":[{"backupAddr":"13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx","defaultAddr":"1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    local reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "0xed8a078ee44eac473bd1d5c971e231c255badf7f0c2fbdbe31ef34669c441d6f" ${MAIN_HTTP}
    echo "========== # retrieve perform end =========="

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

retrieve_Cancel() {
    echo "========== # retrieve cancel begin =========="

    local req='"method":"retrieve.CreateRawRetrieveCancelTx","params":[{"backupAddr":"13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx","defaultAddr":"1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    local reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "0x3665fa66d1a17d2fc319a45250c8c8b9302ae0c393c2e39f2ef3b2f6bc40a42d" ${MAIN_HTTP}
    echo "========== # retrieve cancel end =========="

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

retrieve_QueryResult() {
    echo "========== # retrieve query result begin =========="

    local status=$1

    local req='"method":"Chain33.Query","params":[{"execer":"retrieve","funcName":"GetRetrieveInfo","payload":{"backupAddress":"13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx", "defaultAddress":"1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL"}}]'
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

    local main_ip=${MAIN_HTTP//8901/8801}
    #main chain import pri key
    #1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL
    chain33_ImportPrivkey "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" "1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL" "retrieve1" "${main_ip}"
    #13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx
    chain33_ImportPrivkey "0xed8a078ee44eac473bd1d5c971e231c255badf7f0c2fbdbe31ef34669c441d6f" "13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx" "retrieve2" "$main_ip"

    local retrieve1="1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL"
    local retrieve2="13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$retrieve1" 12000000000 "${main_ip}"
        chain33_QueryBalance "${retrieve1}" "$main_ip"

        chain33_applyCoins "$retrieve2" 12000000000 "${main_ip}"
        chain33_QueryBalance "${retrieve2}" "$main_ip"
    else
        # tx fee
        chain33_applyCoins "$retrieve1" 1000000000 "${main_ip}"
        chain33_QueryBalance "${retrieve1}" "$main_ip"

        chain33_applyCoins "$retrieve2" 1000000000 "${main_ip}"
        chain33_QueryBalance "${retrieve2}" "$main_ip"
        local para_ip="${MAIN_HTTP}"
        #para chain import pri key
        chain33_ImportPrivkey "0x0316d5e33e7bce2455413156cb95209f8c641af352ee5d648c647f24383e4d94" "1PdaXiQU994gzh4RcjLir2AbyqcQ3TwnBL" "retrieve1"  "$para_ip"
        chain33_ImportPrivkey "0xed8a078ee44eac473bd1d5c971e231c255badf7f0c2fbdbe31ef34669c441d6f" "13t1hnMNHqQ5K4QPeqq5xmdg2kTbDPtrgx" "retrieve2"  "$para_ip"

        chain33_applyCoins "$retrieve1" 12000000000 "${para_ip}"
        chain33_QueryBalance "${retrieve1}" "$para_ip"
        chain33_applyCoins "$retrieve2" 12000000000 "${para_ip}"
        chain33_QueryBalance "${retrieve2}" "$para_ip"
    fi

    chain33_SendToAddress "$retrieve1" "$retrieve_addr" 1000000000 ${MAIN_HTTP}
    chain33_QueryExecBalance "${retrieve1}" "retrieve" "$MAIN_HTTP"
    chain33_SendToAddress "$retrieve2" "$retrieve_addr" 1000000000 ${MAIN_HTTP}
    chain33_QueryExecBalance "${retrieve2}" "retrieve" "$MAIN_HTTP"

    chain33_BlockWait 1 "${MAIN_HTTP}"
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

chain33_debug_function main "$1"
