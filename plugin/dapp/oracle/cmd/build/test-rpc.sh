#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""
oracle_addPublisher_unsignedTx="0a066d616e61676512410a3f0a146f7261636c652d7075626c6973682d6576656e741222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630e6b685d696ee9394163a223151344e687572654a784b4e4266373164323642394a336642516f5163666d657a32"
oracle_publisher_addr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
oracle_publishers_addr=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# $2=0 means true, other false
echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo "$1 ok"
    else
        echo "$1 err"
        CASE_ERR="err"
    fi

}

create_publish_transaction() {
    local ip=$1
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"EventPublish","payload":{"type":"football", "subType":"Premier League","time":?????????,"content":"test","introduction":"test"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '.error|not' <<<"$resp")
}

create_prePublishResult_transaction() {
    local ip=$1
    event_id=$2
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"ResultPrePublish","payload":{"eventID":"${event_id}", "source":"sina sport","result":"0:1"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '.error|not' <<<"$resp")
}

create_publishResult_transaction() {
    local ip=$1
    event_id=$2
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"oracle","actionName":"ResultPublish","payload":{"eventID":"${event_id}", "source":"sina sport","result":"1:1"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '.error|not' <<<"$resp")
}

oracle_AddPublisher(){
    echo "=============== # add publisher ==============="
    local ip=$1
    signRawTx "${oracle_addPublisher_unsignedTx}" "${oracle_publisher_addr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "AddPublisher signRawTx" "$rst"
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "AddPublisher sendSignedTx" "$rst"
    fi
    block_wait 1
    queryTransaction ".result.receipt.tyName" "ExecOk"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "AddPublisher queryExecRes" "$rst"
    fi
}

oracle_sendPublishEvent() {
    echo "=============== # add sendPublishEvent ==============="
    local ip=$1
    create_publish_transaction "$ip"
}

signRawTx() {
    unsignedTx=$1
    addr=$2
    signedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SignRawTx","params":[{"addr":"'${addr}'","txHex":"'${unsignedTx}'","expire":"120s"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ $signedTx == "null" ]; then
        return 1
    else
        return 0
    fi
}

sendSignedTx() {
    txHash=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"token":"","data":"'${signedTx}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ $txHash == "null" ]; then
        return 1
    else
        return 0
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

# 查询交易的执行结果
# 根据传入的规则，校验查询的结果 （参数1: 校验规则 参数2: 预期匹配结果）
function queryTransaction() {
    validator=$1
    expectRes=$2
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'${txHash}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r "${validator}")
    if [ ${res} != ${expectRes} ]; then
        return 1
    else
        oracle_publishers_addr=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'${txHash}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.receipt.logs[1].log.current.arr.value")
        echo $oracle_publishers_addr
        return 0
    fi
}

function run_test() {
    local ip=$1
    oracle_AddPublisher "$ip"
    oracle_sendPublishEvent "$ip"
}

function main() {
    local ip=$1
    MAIN_HTTP="http://$ip:8801"
    echo "=========== # oracle rpc test ============="
    echo "main_ip=$MAIN_HTTP"
    run_test "$MAIN_HTTP"

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Oracle Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Oracle Rpc Test Pass==============${NOC}"
    fi
}

main "$1"