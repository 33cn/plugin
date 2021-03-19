#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -e
set -o pipefail

MAIN_HTTP=""

source ../dapp-test-common.sh

#ticketId=""
price=$((10000 * 100000000))

ticket_CreateBindMiner() {
    #创建交易
    minerAddr=$1
    returnAddr=$2
    returnPriv=$3
    amount=$4
    req='{"method":"ticket.CreateBindMiner","params":[{"bindAddr":"'"$minerAddr"'", "originAddr":"'"$returnAddr"'", "amount":'"$amount"', "checkBalance":true}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result.txHex"
    chain33_SignAndSendTx "$RETURN_RESP" "${returnPriv}" ${MAIN_HTTP}
}

ticket_SetAutoMining() {
    flag=$1
    req='{"method":"ticket.SetAutoMining","params":[{"flag":'"$flag"'}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result.isOK == true)' "$FUNCNAME"
}

ticket_GetTicketCount() {
    chain33_Http '{"method":"ticket.GetTicketCount","params":[{}]}' ${MAIN_HTTP} '(.error|not) and (.result > 0)' "$FUNCNAME"
}

ticket_CloseTickets() {
    addr=$1
    req='{"method":"ticket.CloseTickets","params":[{"minerAddress":"'"$addr"'"}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME"
}

ticket_TicketInfos() {
    tid=$1
    minerAddr=$2
    returnAddr=$3
    req='{"method":"Chain33.Query","params":[{"execer":"ticket","funcName":"TicketInfos","payload":{"ticketIds":["'"$tid"'"]}}]}'
    resok='(.error|not) and (.result.tickets | length > 0) and (.result.tickets[0].minerAddress == "'"$minerAddr"'") and (.result.tickets[0].returnAddress == "'"$returnAddr"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

ticket_TicketList() {
    minerAddr=$1
    returnAddr=$2
    status=$3
    req='{"method":"Chain33.Query","params":[{"execer":"ticket","funcName":"TicketList","payload":{"addr":"'"$minerAddr"'", "status":'"$status"'}}]}'
    resok='(.error|not) and (.result.tickets | length > 0) and (.result.tickets[0].minerAddress == "'"$minerAddr"'") and (.result.tickets[0].returnAddress == "'"$returnAddr"'") and (.result.tickets[0].status == '"$status"')'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"

    ticket0=$(echo "${RETURN_RESP}" | jq -r ".result.tickets[0]")
    echo -e "######\\n  ticket[0] is $ticket0)  \\n######"
    ticketId=$(echo "${RETURN_RESP}" | jq -r ".result.tickets[0].ticketId")
    echo -e "######\\n  ticketId is $ticketId  \\n######"
}

ticket_MinerAddress() {
    returnAddr=$1
    minerAddr=$2
    req='{"method":"Chain33.Query","params":[{"execer":"ticket","funcName":"MinerAddress","payload":{"data":"'"$returnAddr"'"}}]}'
    resok='(.error|not) and (.result.data == "'"$minerAddr"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

ticket_MinerSourceList() {
    minerAddr=$1
    returnAddr=$2
    req='{"method":"Chain33.Query","params":[{"execer":"ticket","funcName":"MinerSourceList","payload":{"data":"'"$minerAddr"'"}}]}'
    resok='(.error|not) and (.result.datas | length > 0) and (.result.datas[0] == "'"$returnAddr"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

ticket_RandNumHash() {
    hash=$1
    blockNum=$2
    req='{"method":"Chain33.Query","params":[{"execer":"ticket","funcName":"RandNumHash","payload":{"hash":"'"$hash"'", "blockNum":'"$blockNum"'}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result.hash != null)' "$FUNCNAME"
}

function run_testcases() {
    #账户地址
    minerAddr1="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    #returnAddr1="1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"

    minerAddr2="12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d"

    returnAddr2="1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB"
    returnPriv2="0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d"

    chain33_QueryBalance "${returnAddr2}" "${MAIN_HTTP}"
    chain33_applyCoins "${minerAddr2}" 1000000000 "${MAIN_HTTP}"

    ticket_SetAutoMining 0
    ticket_GetTicketCount
    #ticket_TicketList "${minerAddr1}" "${returnAddr1}" 1
    #ticket_TicketInfos "${ticketId}" "${minerAddr1}" "${returnAddr1}"
    #购票
    ticket_CreateBindMiner "${minerAddr2}" "${returnAddr2}" "${returnPriv2}" ${price}
    ticket_MinerAddress "${returnAddr2}" "${minerAddr2}"
    ticket_MinerSourceList "${minerAddr2}" "${returnAddr2}"
    #关闭
    ticket_CloseTickets "${minerAddr1}"

    chain33_LastBlockhash "${MAIN_HTTP}"
    ticket_RandNumHash "${LAST_BLOCK_HASH}" 5
}

function main() {
    chain33_RpcTestBegin Ticket
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    if [[ $ispara == true ]]; then
        echo "***skip ticket test on parachain***"
    else
        run_testcases
    fi

    chain33_RpcTestRst Ticket "$CASE_ERR"
}

chain33_debug_function main "$1"
