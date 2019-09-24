#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

lottery_addCreator_unsignedTx="0a066d616e616765123c0a3a0a0f6c6f74746572792d63726561746f721222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630c788b8f7ccbadbc0703a223151344e687572654a784b4e4266373164323642394a336642516f5163666d657a32"
lottery_addCreator_unsignedTx_para="0a12757365722e702e706172612e6d616e616765123c0a3a0a0f6c6f74746572792d63726561746f721222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630a8bba1b887e7dade2b3a22314469484633317577783977356a6a733571514269474a6b4e686e71656564763157"

lottery_creator_addr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
lottery_creator_priv="0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"

gID=""
lottExecAddr=""
luckyNumber=""

purNum=200
drawNum=220
opRatio=5
devRatio=5

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ispara=$ispara"

    if [[ $ispara == true ]]; then
        lottExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.lottery"}]}' ${MAIN_HTTP} | jq -r ".result")
        chain33_SignRawTx "${lottery_addCreator_unsignedTx_para}" "${lottery_creator_priv}" ${MAIN_HTTP}
    else
        lottExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"lottery"}]}' ${MAIN_HTTP} | jq -r ".result")
        chain33_SignRawTx "${lottery_addCreator_unsignedTx}" "${lottery_creator_priv}" ${MAIN_HTTP}
    fi
    echo "lottExecAddr=$lottExecAddr"

    local main_ip=${MAIN_HTTP//8901/8801}
    #main chain import pri key
    #1FLh9wBS2rat1mUS4G95hRpJt6yHYy5nHF
    chain33_ImportPrivkey "0x8223b757a5d0f91b12e7af3b9666ca33be47fe63e1502987b0537089aaf90bc1" "1FLh9wBS2rat1mUS4G95hRpJt6yHYy5nHF" "lottery1" "${main_ip}"
    #1UWE6NfXPR7eNAjYgT4HMERp7cMMi486E
    chain33_ImportPrivkey "0xbfccb96690e0a1f89748b321f85b03e14bda0cb3d5d19f255ff0b9b0ffb624b3" "1UWE6NfXPR7eNAjYgT4HMERp7cMMi486E" "lottery2" "$main_ip"

    local ACCOUNT_A="1FLh9wBS2rat1mUS4G95hRpJt6yHYy5nHF"
    local ACCOUNT_B="1UWE6NfXPR7eNAjYgT4HMERp7cMMi486E"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$ACCOUNT_A" 12000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$main_ip"

        chain33_applyCoins "$ACCOUNT_B" 12000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_B}" "$main_ip"
    else
        # tx fee
        chain33_applyCoins "$ACCOUNT_A" 1000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$main_ip"

        chain33_applyCoins "$ACCOUNT_B" 1000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_B}" "$main_ip"
        local para_ip="${MAIN_HTTP}"
        #para chain import pri key
        chain33_ImportPrivkey "0x8223b757a5d0f91b12e7af3b9666ca33be47fe63e1502987b0537089aaf90bc1" "1FLh9wBS2rat1mUS4G95hRpJt6yHYy5nHF" "lottery1" "$para_ip"
        chain33_ImportPrivkey "0xbfccb96690e0a1f89748b321f85b03e14bda0cb3d5d19f255ff0b9b0ffb624b3" "1UWE6NfXPR7eNAjYgT4HMERp7cMMi486E" "lottery2" "$para_ip"

        chain33_applyCoins "$ACCOUNT_A" 12000000000 "${para_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$para_ip"
        chain33_applyCoins "$ACCOUNT_B" 12000000000 "${para_ip}"
        chain33_QueryBalance "${ACCOUNT_B}" "$para_ip"
    fi
}

lottery_LotteryCreate() {
    #创建交易
    priv=$1
    resp=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryCreate",
    "payload":{"purBlockNum":'"$purNum"',"drawBlockNum":'"$drawNum"', "opRewardRatio":'"$opRatio"',"devRewardRatio":'"$devRatio"',"fee":1000000}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SignRawTx "${rawTx}" "${priv}" ${MAIN_HTTP}

    gID="${RAW_TX_HASH}"
    echo "gameID $gID"
}

lottery_LotteryBuy() {
    #创建交易
    priv=$1
    amount=$2
    number=$3
    way=$4
    resp=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryBuy",
    "payload":{"lotteryId":"'"$gID"'","amount":'"$amount"',"number":'"$number"',"way":'"$way"',"fee":1000000}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SignRawTx "${rawTx}" "${priv}" ${MAIN_HTTP}
}

lottery_LotteryDraw() {
    #创建交易
    priv=$1
    resp=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryDraw",
    "payload":{"lotteryId":"'"$gID"'","fee":1000000}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SignRawTx "${rawTx}" "${priv}" ${MAIN_HTTP}
}

lottery_LotteryClose() {
    #创建交易
    priv=$1
    resp=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryClose",
    "payload":{"lotteryId":"'"$gID"'","fee":1000000}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [[ $ok == null ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SignRawTx "${rawTx}" "${priv}" ${MAIN_HTTP}
}

lottery_GetLotteryNormalInfo() {
    gameID=$1
    addr=$2
    execer="lottery"
    funcName="GetLotteryNormalInfo"
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.purBlockNum == "'"$purNum"'") and (.result.drawBlockNum == "'"$drawNum"'") and (.result.createAddr == "'"$addr"'") and (.result.opRewardRatio == "'"$opRatio"'") and (.result.devRewardRatio == "'"$devRatio"'") and (.result | [has("createHeight"), true] | unique | length == 1)' <<<"$resp")
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
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.status == '"$status"') and (.result.buyAmount == "'"$amount"'") and (.result | [has("lastTransToPurState", "lastTransToDrawState", "totalPurchasedTxNum", "round", "luckyNumber", "lastTransToPurStateOnMain", "lastTransToDrawStateOnMain", "purBlockNum", "drawBlockNum", "missingRecords", "totalAddrNum"), true] | unique | length == 1)' <<<"$resp")
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
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.address | length == '"$count"')' <<<"$resp")
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
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$lucky"'")' <<<"$resp")
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
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "round":['"$round"']}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == 1) and (.result.records[0].number == "'"$lucky"'")' <<<"$resp")
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
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$number"'")' <<<"$resp")
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
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'", "round":'"$round"'}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$number"'")' <<<"$resp")
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
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].addr == "'"$addr"'") and (.result.records[0].buyAmount == "'"$amount"'")' <<<"$resp")
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
    resp=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'", "round":'"$round"'}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result.addr == "'"$addr"'") and (.result.round == "'"$round"'") and (.result.buyAmount == "'"$amount"'") and (.result | [has("fundAmount"), true] | unique | length == 1)' <<<"$resp")
    [[ $ok == true ]]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function run_testcases() {
    #账户地址
    gameAddr1="1FLh9wBS2rat1mUS4G95hRpJt6yHYy5nHF"
    gamePriv1="0x8223b757a5d0f91b12e7af3b9666ca33be47fe63e1502987b0537089aaf90bc1"
    gameAddr2="1UWE6NfXPR7eNAjYgT4HMERp7cMMi486E"
    gamePriv2="0xbfccb96690e0a1f89748b321f85b03e14bda0cb3d5d19f255ff0b9b0ffb624b3"

    #给游戏合约中转帐
    chain33_SendToAddress "${gameAddr1}" "${lottExecAddr}" 500000000 "${MAIN_HTTP}"
    chain33_SendToAddress "${gameAddr2}" "${lottExecAddr}" 500000000 "${MAIN_HTTP}"

    #创建游戏
    lottery_LotteryCreate "${lottery_creator_priv}"
    lottery_GetLotteryNormalInfo "$gID" "${lottery_creator_addr}"
    lottery_GetLotteryCurrentInfo "$gID" 1 0

    #第一次投注
    lottery_LotteryBuy "${gamePriv1}" 1 12345 1
    lottery_LotteryBuy "${gamePriv2}" 2 66666 2
    #查询
    lottery_GetLotteryCurrentInfo "$gID" 2 3
    lottery_GetLotteryPurchaseAddr "$gID" 2
    lottery_GetLotteryHistoryBuyInfo "$gID" "${gameAddr1}" 1 "12345"
    lottery_GetLotteryBuyRoundInfo "$gID" "${gameAddr2}" 1 1 "66666"

    #第二次投注
    lottery_LotteryBuy "${gamePriv1}" 2 12321 1
    lottery_LotteryBuy "${gamePriv2}" 1 78987 5
    #查询
    lottery_GetLotteryCurrentInfo "$gID" 2 6
    lottery_GetLotteryPurchaseAddr "$gID" 2
    lottery_GetLotteryHistoryBuyInfo "$gID" "${gameAddr1}" 2 "12321"
    lottery_GetLotteryBuyRoundInfo "$gID" "${gameAddr2}" 1 2 "78987"

    #游戏开奖
    M_HTTP=${MAIN_HTTP//8901/8801}
    chain33_BlockWait ${drawNum} "${M_HTTP}"
    lottery_LotteryDraw "${lottery_creator_priv}"
    lottery_GetLotteryCurrentInfo "$gID" 3 0

    #游戏查询
    lottery_GetLotteryHistoryLuckyNumber "$gID" 1 "${luckyNumber}"
    lottery_GetLotteryRoundLuckyNumber "$gID" 1 "${luckyNumber}"
    lottery_GetLotteryHistoryGainInfo "$gID" "${gameAddr1}" 1 3
    lottery_GetLotteryHistoryGainInfo "$gID" "${gameAddr2}" 1 3
    lottery_GetLotteryRoundGainInfo "$gID" "${gameAddr1}" 1 3
    lottery_GetLotteryRoundGainInfo "$gID" "${gameAddr2}" 1 3

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

chain33_debug_function main "$1"
