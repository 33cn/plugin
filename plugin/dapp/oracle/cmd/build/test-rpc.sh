#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""
oracle_addPublisher_unsignedTx="0a066d616e61676512410a3f0a146f7261636c652d7075626c6973682d6576656e741222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630e6b685d696ee9394163a223151344e687572654a784b4e4266373164323642394a336642516f5163666d657a32"
oracle_publisher_addr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
oracle_publishers_addr=""
eventId=""
txhash=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# $2=0 means true, other false
echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="err"
    fi
}

oracle_AddPublisher(){
    echo "=============== # Add publisher ==============="
    signAndSendRawTx "${oracle_addPublisher_unsignedTx}" "${oracle_publisher_addr}"
}

oracle_publish_transaction() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"EventPublish","payload":{"type":"football", "subType":"Premier League","time":1747814996,"content":"test","introduction":"test"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
        echo_rst "$FUNCNAME" "$?"
        rawtx=$(jq -r ".result" <<<"$resp")
        signAndSendRawTx "$rawtx" "${oracle_publisher_addr}"
        eventId="${txhash}"
        echo "eventId $eventId"
}

oracle_prePublishResult_transaction() {
    event_id=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"ResultPrePublish","payload":{"eventID":"'"$event_id"'", "source":"sina sport","result":"0:1"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
        echo_rst "$FUNCNAME" "$?"
        rawtx=$(jq -r ".result" <<<"$resp")
        signAndSendRawTx "$rawtx" "${oracle_publisher_addr}"
}

oracle_eventAbort_transaction() {
    event_id=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"EventAbort","payload":{"eventID":"'"$event_id"'"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
        echo_rst "$FUNCNAME" "$?"
        rawtx=$(jq -r ".result" <<<"$resp")
        signAndSendRawTx "$rawtx" "${oracle_publisher_addr}"
}

oracle_resultAbort_transaction() {
    event_id=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"ResultAbort","payload":{"eventID":"'"$event_id"'"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
        echo_rst "$FUNCNAME" "$?"
        rawtx=$(jq -r ".result" <<<"$resp")
        signAndSendRawTx "$rawtx" "${oracle_publisher_addr}"
}

oracle_publishResult_transaction() {
    event_id=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"ResultPublish","payload":{"eventID":"'"$event_id"'", "source":"sina sport","result":"1:1"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
        echo_rst "$FUNCNAME" "$?"
        rawtx=$(jq -r ".result" <<<"$resp")
        signAndSendRawTx "$rawtx" "${oracle_publisher_addr}"
}

# 签名并发送
signAndSendRawTx() {
    unsignedTx=$1
    addr=$2
    req='"method":"Chain33.SignRawTx","params":[{"addr":"'${addr}'","txHex":"'${unsignedTx}'","expire":"120s"}]'
    signedTx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")
    if [ "$signedTx" == "null" ]; then
        echo "An error occurred while signing"
    else
        sendSignedTx "$signedTx"
    fi
}

sendSignedTx() {
    signedTx=$1
    local req='"method":"Chain33.SendTransaction","params":[{"token":"","data":"'"$signedTx"'"}]'
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [ "$ok" == null ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    txhash=$(echo "${resp}" | jq -r ".result")
    echo "tx hash is $txhash"
}

oracle_QueryOraclesByID() {
    event_id=$1
    local req='"method":"Chain33.Query", "params":[{"execer":"oracle","funcName":"QueryOraclesByIDs","payload":{"eventID":["'"$event_id"'"]}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.status[0] | [has("eventID", "status", "type", "subType", "source"),true] | unique | length == 1)' <<<"$resp")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
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

function queryTransaction() {
    block_wait 1
    local txhash="$1"
    local req='"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]'
    local times=10
    while true; do
        ret=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.tx.hash")
        if [ "${ret}" != "${1}" ]; then
            block_wait 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "====query tx=$1 failed"
                echo "req=$req"
                curl -ksd "{$req}" ${MAIN_HTTP}
                return 1
                exit 1
            fi
        else
            echo "====query tx=$1  success"
            return 0
            break
        fi
    done
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
    block_wait 1
    oracle_QueryOraclesByID "$eventId"

    # 生成发布事件的交易
    oracle_publish_transaction
    # 取消事件发布
    oracle_eventAbort_transaction "$eventId"
    # 根据ID查询事件
    block_wait 1
    oracle_QueryOraclesByID "$eventId"

    # 生成发布事件的交易
    oracle_publish_transaction
    # 预发布事件结果交易
    oracle_prePublishResult_transaction "$eventId"
    # 取消事件预发布
    oracle_resultAbort_transaction "$eventId"
    # 根据ID查询事件
    block_wait 1
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