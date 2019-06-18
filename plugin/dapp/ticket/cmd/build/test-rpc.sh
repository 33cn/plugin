#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

ticketId=""
price=$((10000 * 100000000))

ticket_CreateBindMiner() {
    #创建交易
    minerAddr=$1
    returnAddr=$2
    returnPriv=$3
    amount=$4
    set -x
    resp=$(curl -ksd '{"method":"ticket.CreateBindMiner","params":[{"bindAddr":"'"$minerAddr"'", "originAddr":"'"$returnAddr"'", "amount":'"$amount"', "checkBalance":true}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result.txHex")
    chain33_SignRawTx "${rawTx}" "${returnPriv}" ${MAIN_HTTP}
    set +x
}

ticket_SetAutoMining() {
    flag=$1
    set -x
    resp=$(curl -ksd '{"method":"ticket.SetAutoMining","params":[{"flag":'"$flag"'}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.isOK == true)' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

ticket_GetTicketCount() {
    set -x
    resp=$(curl -ksd '{"method":"ticket.GetTicketCount","params":[{}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result > 0)' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

ticket_CloseTickets() {
    addr=$1
    set -x
    resp=$(curl -ksd '{"method":"ticket.CloseTickets","params":[{"minerAddress":"'"$addr"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.hashes | length > 0)' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

ticket_TicketInfos() {
    tid=$1
    minerAddr=$2
    returnAddr=$3
    status=$4
    execer="ticket"
    funcName="TicketInfos"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"ticketIds":["'"$tid"'"]}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.tickets | length > 0) and (.result.tickets[0].minerAddress == "'"$minerAddr"'") and (.result.tickets[0].returnAddress == "'"$returnAddr"'") and (.result.tickets[0].status == '"$status"')' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

ticket_TicketList() {
    minerAddr=$1
    returnAddr=$2
    status=$3
    execer="ticket"
    funcName="TicketList"
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"addr":"'"$minerAddr"'", "status":'"$status"'}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    set -x
    ok=$(jq '(.error|not) and (.result.tickets | length > 0) and (.result.tickets[0].minerAddress == "'"$minerAddr"'") and (.result.tickets[0].returnAddress == "'"$returnAddr"'") and (.result.tickets[0].status == '"$status"')' <<<"$resp")
    set +x
    ticket0=$(echo "${resp}" | jq -r ".result.tickets[0]")
    echo -e "######\\n  ticket[0] is $ticket0)  \\n######"
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

    if [[ $status == 1 ]]; then
        ticketId=$(echo "${resp}" | jq -r ".result.tickets[0].ticketId")
        echo -e "######\\n  ticketId is $ticketId  \\n######"
    fi
}

ticket_MinerAddress() {
    returnAddr=$1
    minerAddr=$2
    execer="ticket"
    funcName="MinerAddress"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"data":"'"$returnAddr"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.data == "'"$minerAddr"'")' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

ticket_MinerSourceList() {
    minerAddr=$1
    returnAddr=$2
    execer="ticket"
    funcName="MinerSourceList"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"data":"'"$minerAddr"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.datas | length > 0) and (.result.datas[0] == "'"$returnAddr"'")' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

ticket_RandNumHash() {
    hash=$1
    blockNum=$2
    execer="ticket"
    funcName="RandNumHash"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"hash":"'"$hash"'", "blockNum":'"$blockNum"'}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.hash != "")' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function run_testcases() {
    #账户地址
    minerAddr1="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    returnAddr1="1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"

    minerAddr2="12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d"

    returnAddr2="1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB"
    returnPriv2="0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d"

    set -x
    chain33_SendToAddress "${minerAddr1}" "${minerAddr2}" 100000000 "${MAIN_HTTP}"
    chain33_SendToAddress "${minerAddr1}" "${returnAddr2}" $((price + 5 * 100000000)) "${MAIN_HTTP}"
    set +x

    ticket_SetAutoMining 0
    ticket_GetTicketCount
    ticket_TicketList "${minerAddr1}" "${returnAddr1}" 1
    ticket_TicketInfos "${ticketId}" "${minerAddr1}" "${returnAddr1}" 1
    #购票
    ticket_CreateBindMiner "${minerAddr2}" "${returnAddr2}" "${returnPriv2}" ${price}
    ticket_MinerAddress "${returnAddr2}" "${minerAddr2}"
    ticket_MinerSourceList "${minerAddr2}" "${returnAddr2}"
    #关闭
    ticket_CloseTickets "${minerAddr2}"

    chain33_LastBlockhash "${MAIN_HTTP}"
    ticket_RandNumHash "${LAST_BLOCK_HASH}" 5
}

function main() {
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    echo "=========== # ticket rpc test start============="

    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    if [[ $ispara == true ]]; then
        echo "***skip ticket test on parachain***"
    else
        run_testcases
    fi

    if [[ -n $CASE_ERR ]]; then
        echo -e "${RED}=============Ticket Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Ticket Rpc Test Pass==============${NOC}"
    fi
    echo "=========== # ticket rpc test end============="
}

main "$1"
