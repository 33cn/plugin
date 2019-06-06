#!/usr/bin/env bash
# shellcheck disable=SC2128

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
CASE_ERR=""
trade_addr=""
tradeAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
tradeBuyerAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
tokenSymbol="TOKEN"

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

function updateConfig() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "manage","actionName":"Modify","payload":{ "key": "token-blacklist","value": "BTY","op": "add","addr": ""}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "update config create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "update config queryExecRes" "$?"
}

function token_preCreate() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenPreCreateTx","params":[{"name": "yinhebib", "symbol": "'"${tokenSymbol}"'", "total": 1000000000000, "price": 100, "category": 1,"owner":"'${tradeAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token preCreate create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "token preCreate queryExecRes" "$?"
}

function token_finish() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenFinishTx","params":[{"symbol": "'"${tokenSymbol}"'", "owner":"'${tradeAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token finish create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "token finish queryExecRes" "$?"
}

function token_balance() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.GetTokenBalance","params":[{"addresses": ["'"${tradeAddr}"'"],"tokenSymbol":"'"${tokenSymbol}"'","execer": "'"${tokenExecName}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})

    if [ "${res}" == "" ]; then
        echo_rst "token get balance tx" 1
        return
    fi

    addr=$(echo "${res}" | jq -r ".result[0].addr")
    balance=$(echo "${res}" | jq -r ".result[0].balance")

    if [ "${addr}" == "${tradeAddr}" ] && [ "${balance}" -eq 1000000000000 ]; then
        echo_rst "token get balance tx" 0
    else
        echo_rst "token get balance tx" 1
    fi
}

function token_transfer() {
    addr=$1
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${tokenExecName}"'","actionName":"Transfer","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "100000000000", "note": "", "to": "'"${addr}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token transfer create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "token transfer queryExecRes" "$?"
}

function token_sendExec() {
    addr=$1
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${tokenExecName}"'","actionName":"TransferToExec","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "10000000000", "note": "", "to": "'"${trade_addr}"'", "execName": "'"${tradeExecName}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token sendExec create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "${addr}" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "token sendExec queryExecRes" "$?"
}

function trade_createSellTx() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeSellTx","params":[{"tokenSymbol": "'"${tokenSymbol}"'", "amountPerBoardlot": 1000000, "minBoardlot": 1, "pricePerBoardlot": 100000000,"totalBoardlot":100, "fee": 10000000, "assetExec":"token"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade createSellTx create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"

     queryTransaction ".error | not" "true"
    echo_rst "trade createSellTx queryExecRes" "$?"
}

function trade_getSellOrder() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesSellOrder","payload":{"addr": "'"${tradeAddr}"'","token":["'"${tokenSymbol}"'"]}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    result=$(echo "${res}" | jq -r ".error | not")
    if [ "${result}" == true ]; then
        sellID=$(echo "${res}" | jq -r ".result.orders[0].sellID" | awk -F '-' '{print $4}')
        echo_rst "trade getSellOrder" 0
    else
        echo_rst "trade getSellOrder" 1
    fi
}

function trade_createBuyTx() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeBuyTx","params":[{"sellID": "'"${sellID}"'", "boardlotCnt": 1, "fee": 1}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade createBuyTx create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "0xCC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "trade createBuyTx queryExecRes" "$?"
}

function trade_getBuyOrder() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesBuyOrder","payload":{"addr": "'"${tradeBuyerAddr}"'","token":["'"${tokenSymbol}"'"]}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${res}" == true ]; then
        echo_rst "trade getBuyOrder" 0
    else
        echo_rst "trade getBuyOrder" 1
    fi
}

function trade_statusBuyOrder() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesBuyOrderWithStatus","payload":{"addr": "'"${tradeBuyerAddr}"'","status":6}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${res}" == true ]; then
        echo_rst "trade getStatusBuyOrder" 0
    else
        echo_rst "trade getStatusBuyOrder" 1
    fi
}

function trade_statusOrder() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesOrderWithStatus","payload":{"addr": "'"${tradeAddr}"'","status":1}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${res}" == true ]; then
        echo_rst "trade getStatusOrder" 0
    else
        echo_rst "trade getStatusOrder" 1
    fi
}

function trade_statusSellOrder() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesSellOrderWithStatus","payload":{"addr": "'"${tradeAddr}"'", "status":1}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${res}" == true ]; then
        echo_rst "trade getStatusSellOrder" 0
    else
        echo_rst "trade getStatusSellOrder" 1
    fi
}

function trade_statusTokenBuyOrder() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetTokenBuyOrderByStatus","payload":{"tokenSymbol": "'"${tokenSymbol}"'", "count" :1 , "direction": 1,"status":6}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${res}" == true ]; then
        echo_rst "trade getTokenBuyOrder" 0
    else
        echo_rst "trade getTokenBuyOrder" 1
    fi

}

function trade_statusTokenSellOrder() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetTokenSellOrderByStatus","payload":{"tokenSymbol": "'"${tokenSymbol}"'", "count" :1 , "direction": 1,"status":1}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${res}" == true ]; then
        echo_rst "trade getTokenSellOrder" 0
    else
        echo_rst "trade getTokenSellOrder" 1
    fi
}

function trade_buyLimit() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeBuyLimitTx","params":[{"tokenSymbol":"'"${tokenSymbol}"'","amountPerBoardlot":1000000,"minBoardlot":1, "pricePerBoardlot":100000, "totalBoardlot":200, "fee": 1, "assetExec":"'"${tokenExecName}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade buyLimit create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "trade buyLimit queryExecRes" "$?"
    buyID=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${txHash}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.receipt.logs[1].log.base.buyID" | awk -F '-' '{print $4}')
}

function trade_sellMarket() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeSellMarketTx","params":[{"buyID":"'"${buyID}"'","boardlotCnt":10, "fee": 1}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade sellMarket create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "${tradeAddr}" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "trade sellMarket queryExecRes" "$?"
}

function trade_revokeBuy() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeRevokeTx","params":[{"sellID":"'"${sellID}"'","fee": 1}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade revokeBuy create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "trade revokeBuy queryExecRes" "$?"
}

function trade_revoke() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeRevokeBuyTx","params":[{"buyID":"'"${buyID}"'","fee": 1}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade revoke create tx" 1
        return
    fi

    chain33_SignRawTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"

    queryTransaction ".error | not" "true"
    echo_rst "trade revoke queryExecRes" "$?"
}

# 查询交易的执行结果
# 根据传入的规则，校验查询的结果 （参数1: 校验规则 参数2: 预期匹配结果）
function queryTransaction() {
    validator=$1
    expectRes=$2
    echo "txhash=${RAW_TX_HASH}"
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${RAW_TX_HASH}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r "${validator}")
    if [ "${res}" != "${expectRes}" ]; then
        return 1
    else
        return 0
    fi
}

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    tokenExecName="token"
    tradeExecName="trade"
    from="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    local trade_addr=""
    if [ "$ispara" == "true" ]; then
        tokenExecName="user.p.para.token"
        tradeExecName="user.p.para.trade"
        trade_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${tradeExecName}"'"}]}' ${MAIN_HTTP} | jq -r ".result")
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${tokenExecName}"'"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        trade_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${tradeExecName}"'"}]}' ${MAIN_HTTP} | jq -r ".result")
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${tokenExecName}"'"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi

    chain33_SendToAddress "$tradeAddr" "$tradeBuyerAddr" 10000000000 "${MAIN_HTTP}"
    chain33_SendToAddress "$tradeAddr" "$trade_addr" 10000000000 "${MAIN_HTTP}"
    chain33_SendToAddress "$tradeAddr" "$token_addr" 10000000000 "${MAIN_HTTP}"
    chain33_BlockWait 2 "${MAIN_HTTP}"
    chain33_SendToAddress "$tradeBuyerAddr" "$trade_addr" 10000000000 "${MAIN_HTTP}"
    chain33_BlockWait 2 "${MAIN_HTTP}"

    echo "trade=$trade_addr"

    updateConfig
    token_preCreate
    token_finish
    token_balance
    token_transfer "${tradeBuyerAddr}"
    token_sendExec "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
    token_sendExec "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944 "
}

function run_test() {
    local ip=$1

    trade_createSellTx
    trade_getSellOrder
    trade_createBuyTx
    trade_getBuyOrder
    trade_statusBuyOrder
    trade_statusOrder
    trade_statusSellOrder

    trade_buyLimit
    trade_statusTokenBuyOrder
    trade_sellMarket
    trade_statusTokenSellOrder
    trade_revokeBuy
    trade_revoke
}

function main() {
    local ip=$1
    MAIN_HTTP=$ip
    echo "=========== # trade rpc test ============="
    echo "main_ip=$MAIN_HTTP"

    init
    run_test "$MAIN_HTTP"

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============trade Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============trade Rpc Test Pass==============${NOC}"
    fi
}

main "$1"
