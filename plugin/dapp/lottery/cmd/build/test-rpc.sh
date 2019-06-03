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

# $2=0 means true, other false
echo_rst() {
    if [[ $2 -eq 0 ]]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="err"
    fi
}

lottery_addCreator_unsignedTx="0a066d616e616765123c0a3a0a0f6c6f74746572792d63726561746f721222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630c788b8f7ccbadbc0703a223151344e687572654a784b4e4266373164323642394a336642516f5163666d657a32"
lottery_addCreator_unsignedTx_para="0a12757365722e702e706172612e6d616e616765123c0a3a0a0f6c6f74746572792d63726561746f721222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630a8bba1b887e7dade2b3a22314469484633317577783977356a6a733571514269474a6b4e686e71656564763157"

lottery_creator_addr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
lottery_creator_priv="0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"

gID=""
gResp=""
lottExecAddr=""
luckyNumber=""

purNum=30
drawNum=40
opRatio=5
devRatio=5

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ispara=$ispara"

    if [[ $ispara == true ]]; then
        lottExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.lottery"}]}' ${MAIN_HTTP} | jq -r ".result")
        chain33_SendTransaction "${lottery_addCreator_unsignedTx_para}" "${lottery_creator_priv}"
    else
        lottExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"lottery"}]}' ${MAIN_HTTP} | jq -r ".result")
        chain33_SendTransaction "${lottery_addCreator_unsignedTx}" "${lottery_creator_priv}"
    fi
    echo "lottExecAddr=$lottExecAddr"
}

function block_wait() {
    if [[ $# -lt 1 ]]; then
        echo "wrong block_wait params"
        exit 1
    fi
    http="$2"
    if [[ ! $http ]]; then
        http=${MAIN_HTTP}
    fi
    cur_height=$(curl -ksd '{"method":"Chain33.GetLastHeader","params":[{}]}' -H 'content-type:text/plain;' ${http} | jq -r ".result.height")
    expect=$((cur_height + ${1}))
    local count=0
    while true; do
        new_height=$(curl -ksd '{"method":"Chain33.GetLastHeader","params":[{}]}' -H 'content-type:text/plain;' ${http} | jq -r ".result.height")
        if [[ ${new_height} -ge ${expect} ]]; then
            break
        fi
        count=$((count + 1))
        sleep 1
    done
    echo "wait new block $count s, cur height=$expect,old=$cur_height"
}

function query_tx() {
    if [[ $# -lt 1 ]]; then
        echo "wrong block_wait params"
        exit 1
    fi
    http="$2"
    if [[ ! $http ]]; then
        http=${MAIN_HTTP}
    fi
    txhash="$1"
    req='{"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]}'
    local times=10
    block_wait 1 ${http}
    while true; do
        ret=$(curl -ksd "$req" ${http})
        tx=$(jq -r ".result.tx.hash" <<<"$ret")
        if [[ ${tx} != "${1}" ]]; then
            block_wait 1 ${http}
            times=$((times - 1))
            if [[ $times -le 0 ]]; then
                echo_rst "query tx=$1" 1
                exit 1
            fi
        else
            exec_ok=$(jq '(.result.receipt.tyName == "ExecOk")' <<<"$ret")
            [[ $exec_ok == true ]]
            echo_rst "query tx=$1" $?
            echo -e "######\\n  response: $ret  \\n######"
            break
        fi
    done
}

chain33_SendToAddress() {
    from=$1
    to=$2
    amount=$3
    http=$4
    note="test"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.SendToAddress","params":[{"from":"'"$from"'","to":"'"$to"'","amount":'"$amount"',"note":"'"$note"'"}]}' -H 'content-type:text/plain;' "${http}")
    ok=$(jq '(.error|not)' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

    #查询交易
    txhash=$(jq -r ".result.hash" <<<"$resp")
    query_tx "$txhash" "$http"
}

chain33_SendTransaction() {
    rawTx=$1
    priv=$2
    #签名交易
    set -x
    resp=$(curl -ksd '{"method":"Chain33.SignRawTx","params":[{"privkey":"'"$priv"'","txHex":"'"$rawTx"'","expire":"120s","fee":10000000,"index":0}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    set +x
    [[ $ok == null ]]
    rst=$?
    echo_rst "chain33_SignRawTx" "$rst"

    signTx=$(echo "${resp}" | jq -r ".result")
    set -x
    resp=$(curl -ksd '{"method":"Chain33.SendTransaction","params":[{"data":"'"$signTx"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    set +x
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

    #查询交易
    gResp=$(jq -r ".result" <<<"$resp")
    query_tx "$gResp"
}

lottery_LotteryCreate() {
    #创建交易
    priv=$1
    set -x
    resp=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryCreate",
    "payload":{"purBlockNum":'"$purNum"',"drawBlockNum":'"$drawNum"', "opRewardRatio":'"$opRatio"',"devRewardRatio":'"$devRatio"',"fee":1000000}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    set +x
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SendTransaction "${rawTx}" "${priv}"
    gID="${gResp}"
    echo "gameID $gID"
}

lottery_LotteryBuy() {
    #创建交易
    priv=$1
    amount=$2
    number=$3
    way=$4
    set -x
    resp=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryBuy",
    "payload":{"lotteryId":"'"$gID"'","amount":'"$amount"',"number":'"$number"',"way":'"$way"',"fee":1000000}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    set +x
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SendTransaction "${rawTx}" "${priv}"
}

lottery_LotteryDraw() {
    #创建交易
    priv=$1
    set -x
    resp=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryDraw",
    "payload":{"lotteryId":"'"$gID"'","fee":1000000}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    set +x
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SendTransaction "${rawTx}" "${priv}"
}

lottery_LotteryClose() {
    #创建交易
    priv=$1
    set -x
    resp=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryClose",
    "payload":{"lotteryId":"'"$gID"'","fee":1000000}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    set +x
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SendTransaction "${rawTx}" "${priv}"
}

lottery_GetLotteryNormalInfo() {
    gameID=$1
    addr=$2
    execer="lottery"
    funcName="GetLotteryNormalInfo"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.purBlockNum == "'"$purNum"'") and (.result.drawBlockNum == "'"$drawNum"'") and (.result.createAddr == "'"$addr"'") and (.result.opRewardRatio == "'"$opRatio"'") and (.result.devRewardRatio == "'"$devRatio"'") and (.result | [has("createHeight"), true] | unique | length == 1)' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

lottery_GetLotteryCurrentInfo() {
    gameID=$1
    status=$2
    amount=$3
    execer="lottery"
    funcName="GetLotteryCurrentInfo"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.status == '"$status"') and (.result.buyAmount == "'"$amount"'") and (.result | [has("lastTransToPurState", "lastTransToDrawState", "totalPurchasedTxNum", "round", "luckyNumber", "lastTransToPurStateOnMain", "lastTransToDrawStateOnMain", "purBlockNum", "drawBlockNum", "missingRecords", "totalAddrNum"), true] | unique | length == 1)' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

    if [[ $status == 3 ]]; then
        luckyNumber=$(echo "${resp}" | jq -r ".result.luckyNumber")
        echo -e "######\\n  luckyNumber is $luckyNumber  \\n######"
    fi
}

lottery_GetLotteryPurchaseAddr() {
    gameID=$1
    count=$2
    execer="lottery"
    funcName="GetLotteryPurchaseAddr"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.address | length == '"$count"')' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

lottery_GetLotteryHistoryLuckyNumber() {
    gameID=$1
    count=$2
    lucky=$3
    execer="lottery"
    funcName="GetLotteryHistoryLuckyNumber"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$lucky"'")' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

lottery_GetLotteryRoundLuckyNumber() {
    gameID=$1
    round=$2
    lucky=$3
    execer="lottery"
    funcName="GetLotteryRoundLuckyNumber"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "round":['"$round"']}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == 1) and (.result.records[0].number == "'"$lucky"'")' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

lottery_GetLotteryHistoryBuyInfo() {
    gameID=$1
    addr=$2
    count=$3
    number=$4
    execer="lottery"
    funcName="GetLotteryHistoryBuyInfo"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$number"'")' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

lottery_GetLotteryBuyRoundInfo() {
    gameID=$1
    addr=$2
    round=$3
    count=$4
    number=$5
    execer="lottery"
    funcName="GetLotteryBuyRoundInfo"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'", "round":'"$round"'}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$number"'")' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

lottery_GetLotteryHistoryGainInfo() {
    gameID=$1
    addr=$2
    count=$3
    amount=$4
    execer="lottery"
    funcName="GetLotteryHistoryGainInfo"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].addr == "'"$addr"'") and (.result.records[0].buyAmount == "'"$amount"'")' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

lottery_GetLotteryRoundGainInfo() {
    gameID=$1
    addr=$2
    round=$3
    amount=$4
    execer="lottery"
    funcName="GetLotteryRoundGainInfo"
    set -x
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'", "round":'"$round"'}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.addr == "'"$addr"'") and (.result.round == "'"$round"'") and (.result.buyAmount == "'"$amount"'") and (.result | [has("fundAmount"), true] | unique | length == 1)' <<<"$resp")
    set +x
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function run_testcases() {
    #账户地址
    gameAddr1="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    gamePriv1="0x56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138"
    gameAddr2="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    gamePriv2="0x2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989"
    #该账户需要转账
    gameAddr3="12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d"
    gamePriv3="0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4"

    #主链中相应账户需要转帐
    M_HTTP=${MAIN_HTTP//8901/8801}
    chain33_SendToAddress "${gameAddr1}" "${gameAddr3}" 300000000 "${M_HTTP}"
    #平行链相应账户需要转帐
    chain33_SendToAddress "${gameAddr1}" "${gameAddr3}" 300000000 "${MAIN_HTTP}"

    #给游戏合约中转帐
    chain33_SendToAddress "${gameAddr1}" "${lottExecAddr}" 200000000 "${MAIN_HTTP}"
    chain33_SendToAddress "${gameAddr2}" "${lottExecAddr}" 200000000 "${MAIN_HTTP}"
    chain33_SendToAddress "${gameAddr3}" "${lottExecAddr}" 200000000 "${MAIN_HTTP}"

    #创建游戏
    lottery_LotteryCreate "${lottery_creator_priv}"
    lottery_GetLotteryNormalInfo "$gID" "${lottery_creator_addr}"
    lottery_GetLotteryCurrentInfo "$gID" 1 0

    #第一次投注
    lottery_LotteryBuy "${gamePriv1}" 1 12345 1
    lottery_LotteryBuy "${gamePriv2}" 1 66666 2
    lottery_LotteryBuy "${gamePriv3}" 1 56789 5
    #查询
    lottery_GetLotteryCurrentInfo "$gID" 2 3
    lottery_GetLotteryPurchaseAddr "$gID" 3
    lottery_GetLotteryHistoryBuyInfo "$gID" "${gameAddr1}" 1 "12345"
    lottery_GetLotteryBuyRoundInfo "$gID" "${gameAddr2}" 1 1 "66666"

    #第二次投注
    lottery_LotteryBuy "${gamePriv1}" 1 12321 1
    lottery_LotteryBuy "${gamePriv3}" 1 78987 5
    #查询
    lottery_GetLotteryCurrentInfo "$gID" 2 5
    lottery_GetLotteryPurchaseAddr "$gID" 3
    lottery_GetLotteryHistoryBuyInfo "$gID" "${gameAddr1}" 2 "12321"
    lottery_GetLotteryBuyRoundInfo "$gID" "${gameAddr3}" 1 2 "78987"

    #游戏开奖
    block_wait ${drawNum} "${M_HTTP}"
    lottery_LotteryDraw "${lottery_creator_priv}"
    lottery_GetLotteryCurrentInfo "$gID" 3 0

    #游戏查询
    lottery_GetLotteryHistoryLuckyNumber "$gID" 1 "${luckyNumber}"
    lottery_GetLotteryRoundLuckyNumber "$gID" 1 "${luckyNumber}"
    lottery_GetLotteryHistoryGainInfo "$gID" "${gameAddr1}" 1 2
    lottery_GetLotteryHistoryGainInfo "$gID" "${gameAddr2}" 1 1
    lottery_GetLotteryRoundGainInfo "$gID" "${gameAddr1}" 1 2
    lottery_GetLotteryRoundGainInfo "$gID" "${gameAddr3}" 1 2

    #关闭游戏
    lottery_LotteryClose "${lottery_creator_priv}"
    lottery_GetLotteryCurrentInfo "$gID" 4 0
}

function main() {
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    echo "=========== # lottery rpc test start============="
    init
    run_testcases

    if [[ -n $CASE_ERR ]]; then
        echo -e "${RED}=============Lottery Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Lottery Rpc Test Pass==============${NOC}"
    fi
    echo "=========== # lottery rpc test end============="
}

main "$1"
