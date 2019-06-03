#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
oracle_addPublisher_unsignedTx="0a066d616e61676512410a3f0a146f7261636c652d7075626c6973682d6576656e741222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630e6b685d696ee9394163a223151344e687572654a784b4e4266373164323642394a336642516f5163666d657a32"
oracle_addPublisher_unsignedTx_para="0a12757365722e702e706172612e6d616e61676512410a3f0a146f7261636c652d7075626c6973682d6576656e741222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630a186de8894c9aa864d3a22314469484633317577783977356a6a733571514269474a6b4e686e71656564763157"
oracle_publisher_key="56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138"
eventId=""
txhash=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

oracle_AddPublisher() {
    echo "=============== # Add publisher ==============="
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ispara=$ispara"
    if [ "$ispara" == true ]; then
        chain33_SignRawTx "${oracle_addPublisher_unsignedTx_para}" "${oracle_publisher_key}" "${MAIN_HTTP}"
    else
        chain33_SignRawTx "${oracle_addPublisher_unsignedTx}" "${oracle_publisher_key}" "${MAIN_HTTP}"
    fi
}

oracle_publish_transaction() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"EventPublish","payload":{"type":"football", "subType":"Premier League","time":1747814996,"content":"test","introduction":"test"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${oracle_publisher_key}" "${MAIN_HTTP}"
    eventId="${txhash}"
    echo "eventId $eventId"
}

oracle_prePublishResult_transaction() {
    event_id=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"ResultPrePublish","payload":{"eventID":"'"$event_id"'", "source":"sina sport","result":"0:1"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${oracle_publisher_key}" "${MAIN_HTTP}"
}

oracle_eventAbort_transaction() {
    event_id=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"EventAbort","payload":{"eventID":"'"$event_id"'"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${oracle_publisher_key}" "${MAIN_HTTP}"
}

oracle_resultAbort_transaction() {
    event_id=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"ResultAbort","payload":{"eventID":"'"$event_id"'"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${oracle_publisher_key}" "${MAIN_HTTP}"
}

oracle_publishResult_transaction() {
    event_id=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"ResultPublish","payload":{"eventID":"'"$event_id"'", "source":"sina sport","result":"1:1"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${oracle_publisher_key}" "${MAIN_HTTP}"
}

oracle_QueryOraclesByID() {
    event_id=$1
    local req='"method":"Chain33.Query", "params":[{"execer":"oracle","funcName":"QueryOraclesByIDs","payload":{"eventID":["'"$event_id"'"]}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.status[0] | [has("eventID", "status", "type", "subType", "source"),true] | unique | length == 1)' <<<"$resp")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function run_test() {
    # 增加发布人
    oracle_AddPublisher
    # 生成发布事件的交易
    oracle_publish_transaction
    # 预发布事件结果交易
    oracle_prePublishResult_transaction "$eventId"
    # 事件正式发布
    oracle_publishResult_transaction "$eventId"
    # 根据ID查询事件
    chain33_BlockWait 2 "${MAIN_HTTP}"
    oracle_QueryOraclesByID "$eventId"

    # 生成发布事件的交易
    oracle_publish_transaction
    # 取消事件发布
    oracle_eventAbort_transaction "$eventId"
    # 根据ID查询事件
    chain33_BlockWait 2 "${MAIN_HTTP}"
    oracle_QueryOraclesByID "$eventId"

    # 生成发布事件的交易
    oracle_publish_transaction
    # 预发布事件结果交易
    oracle_prePublishResult_transaction "$eventId"
    # 取消事件预发布
    oracle_resultAbort_transaction "$eventId"
    # 根据ID查询事件
    chain33_BlockWait 2 "${MAIN_HTTP}"
    oracle_QueryOraclesByID "$eventId"

}

function main() {

    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    echo "=========== # oracle rpc test start============="
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Oracle Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Oracle Rpc Test Pass==============${NOC}"
    fi
    echo "=========== # oracle rpc test end============="
}

main "$1"
