#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -e
set -o pipefail

MAIN_HTTP=""
source ../dapp-test-common.sh

lottery_addCreator_unsignedTx="0a066d616e616765123c0a3a0a0f6c6f74746572792d63726561746f721222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630c788b8f7ccbadbc0703a223151344e687572654a784b4e4266373164323642394a336642516f5163666d657a32"
lottery_addCreator_unsignedTx_para="0a12757365722e702e706172612e6d616e616765123c0a3a0a0f6c6f74746572792d63726561746f721222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a0361646420a08d0630a8bba1b887e7dade2b3a22314469484633317577783977356a6a733571514269474a6b4e686e71656564763157"

lottery_creator_addr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
lottery_creator_priv="0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"

gID=""
lottExecAddr=""
luckyNumber=""

#设置较小可能导致投注交易执行失败
purNum=500
drawNum=520
opRatio=5
devRatio=5

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ispara=$ispara"

    if [[ $ispara == true ]]; then
        lottExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.lottery"}]}' ${MAIN_HTTP} | jq -r ".result")
        chain33_SignAndSendTx "${lottery_addCreator_unsignedTx_para}" "${lottery_creator_priv}" ${MAIN_HTTP}
    else
        lottExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"lottery"}]}' ${MAIN_HTTP} | jq -r ".result")
        chain33_SignAndSendTx "${lottery_addCreator_unsignedTx}" "${lottery_creator_priv}" ${MAIN_HTTP}
    fi
    echo "lottExecAddr=$lottExecAddr"

    local main_ip=${MAIN_HTTP//8901/8801}
    chain33_ImportPrivkey "0x8223b757a5d0f91b12e7af3b9666ca33be47fe63e1502987b0537089aaf90bc1" "1FLh9wBS2rat1mUS4G95hRpJt6yHYy5nHF" "lottery1" "${main_ip}"
    chain33_ImportPrivkey "0xbfccb96690e0a1f89748b321f85b03e14bda0cb3d5d19f255ff0b9b0ffb624b3" "1UWE6NfXPR7eNAjYgT4HMERp7cMMi486E" "lottery2" "$main_ip"

    local ACCOUNT_A="1FLh9wBS2rat1mUS4G95hRpJt6yHYy5nHF"
    local ACCOUNT_B="1UWE6NfXPR7eNAjYgT4HMERp7cMMi486E"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$ACCOUNT_A" 12000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$main_ip"

        chain33_applyCoins "$ACCOUNT_B" 12000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_B}" "$main_ip"
    else
        chain33_applyCoins "$ACCOUNT_A" 1000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$main_ip"

        chain33_applyCoins "$ACCOUNT_B" 1000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_B}" "$main_ip"
        local para_ip="${MAIN_HTTP}"
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
    req='{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryCreate","payload":{"purBlockNum":'"$purNum"',"drawBlockNum":'"$drawNum"', "opRewardRatio":'"$opRatio"',"devRewardRatio":'"$devRatio"',"fee":1000000}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result"

    #发送交易
    chain33_SignAndSendTx "${RETURN_RESP}" "${priv}" ${MAIN_HTTP}

    gID="${RAW_TX_HASH}"
    echo "gameID $gID"
}

lottery_LotteryBuy() {
    #创建交易
    priv=$1
    amount=$2
    number=$3
    way=$4
    req='{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryBuy","payload":{"lotteryId":"'"$gID"'","amount":'"$amount"',"number":'"$number"',"way":'"$way"',"fee":1000000}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result"

    #发送交易
    chain33_SignAndSendTx "${RETURN_RESP}" "${priv}" ${MAIN_HTTP}
}

lottery_LotteryDraw() {
    #创建交易
    priv=$1
    req='{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryDraw","payload":{"lotteryId":"'"$gID"'","fee":1000000}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result"
    #发送交易
    chain33_SignAndSendTx "${RETURN_RESP}" "${priv}" ${MAIN_HTTP}
}

lottery_LotteryClose() {
    #创建交易
    priv=$1
    req='{"method":"Chain33.CreateTransaction","params":[{"execer":"lottery","actionName":"LotteryClose","payload":{"lotteryId":"'"$gID"'","fee":1000000}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result"
    #发送交易
    chain33_SignAndSendTx "${RETURN_RESP}" "${priv}" ${MAIN_HTTP}
}

lottery_GetLotteryNormalInfo() {
    gameID=$1
    addr=$2
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryNormalInfo","payload":{"lotteryId":"'"$gameID"'"}}]}'
    resok='(.error|not) and (.result.purBlockNum == "'"$purNum"'") and (.result.drawBlockNum == "'"$drawNum"'") and (.result.createAddr == "'"$addr"'") and (.result.opRewardRatio == "'"$opRatio"'") and (.result.devRewardRatio == "'"$devRatio"'") and (.result | [has("createHeight"), true] | unique | length == 1)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

lottery_GetLotteryCurrentInfo() {
    gameID=$1
    status=$2
    amount=$3
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryCurrentInfo","payload":{"lotteryId":"'"$gameID"'"}}]}'
    resok='(.error|not) and (.result.status == '"$status"') and (.result.buyAmount == "'"$amount"'") and (.result | [has("lastTransToPurState", "lastTransToDrawState", "totalPurchasedTxNum", "round", "luckyNumber", "lastTransToPurStateOnMain", "lastTransToDrawStateOnMain", "purBlockNum", "drawBlockNum", "missingRecords", "totalAddrNum"), true] | unique | length == 1)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME" ".result.luckyNumber"

    if [[ $status == 3 ]]; then
        luckyNumber=$RETURN_RESP
        echo -e "######\\n  luckyNumber is $luckyNumber  \\n######"
    fi
    echo "end"
}

lottery_GetLotteryPurchaseAddr() {
    gameID=$1
    count=$2
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryPurchaseAddr","payload":{"lotteryId":"'"$gameID"'"}}]}'
    resok='(.error|not) and (.result.address | length == '"$count"')'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

lottery_GetLotteryHistoryLuckyNumber() {
    gameID=$1
    count=$2
    lucky=$3
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryHistoryLuckyNumber","payload":{"lotteryId":"'"$gameID"'"}}]}'
    resok='(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$lucky"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

lottery_GetLotteryRoundLuckyNumber() {
    gameID=$1
    round=$2
    lucky=$3
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryRoundLuckyNumber","payload":{"lotteryId":"'"$gameID"'", "round":['"$round"']}}]}'
    resok='(.error|not) and (.result.records | length == 1) and (.result.records[0].number == "'"$lucky"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

lottery_GetLotteryHistoryBuyInfo() {
    gameID=$1
    addr=$2
    count=$3
    number=$4
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryHistoryBuyInfo","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'"}}]}'
    resok='(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$number"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

lottery_GetLotteryBuyRoundInfo() {
    gameID=$1
    addr=$2
    round=$3
    count=$4
    number=$5
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryBuyRoundInfo","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'", "round":'"$round"'}}]}'
    resok='(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].number == "'"$number"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

lottery_GetLotteryHistoryGainInfo() {
    gameID=$1
    addr=$2
    count=$3
    amount=$4
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryHistoryGainInfo","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'"}}]}'
    resok='(.error|not) and (.result.records | length == '"$count"') and (.result.records[0].addr == "'"$addr"'") and (.result.records[0].buyAmount == "'"$amount"'")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

lottery_GetLotteryRoundGainInfo() {
    gameID=$1
    addr=$2
    round=$3
    amount=$4
    req='{"method":"Chain33.Query","params":[{"execer":"lottery","funcName":"GetLotteryRoundGainInfo","payload":{"lotteryId":"'"$gameID"'", "addr":"'"$addr"'", "round":'"$round"'}}]}'
    resok='(.error|not) and (.result.addr == "'"$addr"'") and (.result.round == "'"$round"'") and (.result.buyAmount == "'"$amount"'") and (.result | [has("fundAmount"), true] | unique | length == 1)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
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
    chain33_RpcTestBegin lottery
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases
    chain33_RpcTestRst lottery "$CASE_ERR"
}

chain33_debug_function main "$1"
