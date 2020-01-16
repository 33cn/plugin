#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
trade_addr=""
tradeAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
tradeBuyerAddr="1CvLe1qNaC7tCf5xmfAqJ9UJkMhtmhUKNg"
tokenSymbol="TOKEN"

function updateConfig() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "manage","actionName":"Modify","payload":{ "key": "token-blacklist","value": "BTY","op": "add","addr": ""}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "update config create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
}

function token_preCreate() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenPreCreateTx","params":[{"name": "yinhebib", "symbol": "'"${tokenSymbol}"'", "total": 1000000000000, "price": 100, "category": 1,"owner":"'${tradeAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token preCreate create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
}

function token_finish() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenFinishTx","params":[{"symbol": "'"${tokenSymbol}"'", "owner":"'${tradeAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token finish create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
}

function token_balance() {
    req='{"method":"token.GetTokenBalance","params":[{"addresses": ["'"${tradeAddr}"'"],"tokenSymbol":"'"${tokenSymbol}"'","execer": "'"${tokenExecName}"'"}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error | not) and (.result[0].addr == "'"${tradeAddr}"'") and (.result[0].balance == 1000000000000)' "$FUNCNAME"
}

function token_transfer() {
    addr=$1
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${tokenExecName}"'","actionName":"Transfer","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "100000000000", "note": "", "to": "'"${addr}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token transfer create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
}

function token_sendExec() {
    addr=$1
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${tokenExecName}"'","actionName":"TransferToExec","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "10000000000", "note": "", "to": "'"${trade_addr}"'", "execName": "'"${tradeExecName}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token sendExec create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "${addr}"
}

function trade_createSellTx() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeSellTx","params":[{"tokenSymbol": "'"${tokenSymbol}"'", "amountPerBoardlot": 1000000, "minBoardlot": 1, "pricePerBoardlot": 100000000,"totalBoardlot":100, "fee": 10000000, "assetExec":"token", "priceExec" : "coins", "priceSymbol" : "'"${coinSymbol}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade createSellTx create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
}

function trade_getSellOrder() {
    req='{"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesSellOrder","payload":{"addr": "'"${tradeAddr}"'","token":["'"${tokenSymbol}"'"]}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error | not)' "$FUNCNAME"
    sellID=$(echo "${RETURN_RESP}" | jq -r ".result.orders[0].sellID" | awk -F '-' '{print $4}')
}

function trade_createBuyTx() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeBuyTx","params":[{"sellID": "'"${sellID}"'", "boardlotCnt": 1, "fee": 1}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade createBuyTx create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0xaeef1ad76d43a2056d0dcb57d5bf1ba96471550614ab9e7f611ef9c5ca403f42"
}

function trade_getBuyOrder() {
    req='{"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesBuyOrder","payload":{"addr": "'"${tradeBuyerAddr}"'","token":["'"${tokenSymbol}"'"]}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error | not)' "$FUNCNAME"
}

function trade_statusBuyOrder() {
    req='{"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesBuyOrderWithStatus","payload":{"addr": "'"${tradeBuyerAddr}"'","status":6}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error | not)' "$FUNCNAME"
}

function trade_statusOrder() {
    req='{"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesOrderWithStatus","payload":{"addr": "'"${tradeAddr}"'","status":1}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error | not)' "$FUNCNAME"
}

function trade_statusSellOrder() {
    req='{"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetOnesSellOrderWithStatus","payload":{"addr": "'"${tradeAddr}"'", "status":1}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error | not)' "$FUNCNAME"
}

function trade_statusTokenBuyOrder() {
    req='{"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetTokenBuyOrderByStatus","payload":{"tokenSymbol": "'"${tokenSymbol}"'", "count" :1 , "direction": 1,"status":6}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error | not)' "$FUNCNAME"
}

function trade_statusTokenSellOrder() {
    req='{"method":"Chain33.Query","params":[{"execer":"'"${tradeExecName}"'","funcName":"GetTokenSellOrderByStatus","payload":{"tokenSymbol": "'"${tokenSymbol}"'", "count" :1 , "direction": 1,"status":1}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error | not)' "$FUNCNAME"
}

function trade_buyLimit() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeBuyLimitTx","params":[{"tokenSymbol":"'"${tokenSymbol}"'","amountPerBoardlot":1000000,"minBoardlot":1, "pricePerBoardlot":100000, "totalBoardlot":200, "fee": 1, "assetExec":"'"${tokenExecName}"'", "priceExec" : "coins", "priceSymbol" : "'"${coinSymbol}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade buyLimit create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0xaeef1ad76d43a2056d0dcb57d5bf1ba96471550614ab9e7f611ef9c5ca403f42"
    buyID=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${RAW_TX_HASH}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.receipt.logs[1].log.base.buyID" | awk -F '-' '{print $4}')
}

function trade_sellMarket() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeSellMarketTx","params":[{"buyID":"'"${buyID}"'","boardlotCnt":10, "fee": 1}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade sellMarket create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "${tradeAddr}"
}

function trade_revokeBuy() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeRevokeTx","params":[{"sellID":"'"${sellID}"'","fee": 1}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade revokeBuy create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
}

function trade_revoke() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"trade.CreateRawTradeRevokeBuyTx","params":[{"buyID":"'"${buyID}"'","fee": 1}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "trade revoke create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
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

function signRawTxAndQuery() {
    chain33_SignAndSendTx "${unsignedTx}" "$2" "${MAIN_HTTP}"
    queryTransaction ".error | not" "true"
    echo_rst "$1 queryExecRes" "$?"
}

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    tokenExecName="token"
    tradeExecName="trade"
    local trade_addr=""
    if [ "$ispara" == "true" ]; then
        coinSymbol="para"
        tokenExecName="user.p.para.token"
        tradeExecName="user.p.para.trade"
        trade_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${tradeExecName}"'"}]}' ${MAIN_HTTP} | jq -r ".result")
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${tokenExecName}"'"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        coinSymbol="bty"
        trade_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${tradeExecName}"'"}]}' ${MAIN_HTTP} | jq -r ".result")
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'"${tokenExecName}"'"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi

    local main_ip=${MAIN_HTTP//8901/8801}
    chain33_ImportPrivkey "0xaeef1ad76d43a2056d0dcb57d5bf1ba96471550614ab9e7f611ef9c5ca403f42" "1CvLe1qNaC7tCf5xmfAqJ9UJkMhtmhUKNg" "trade1" "${main_ip}"

    local ACCOUNT_A="1CvLe1qNaC7tCf5xmfAqJ9UJkMhtmhUKNg"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$ACCOUNT_A" 12000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$main_ip"
    else
        chain33_applyCoins "$ACCOUNT_A" 1000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$main_ip"

        local para_ip="${MAIN_HTTP}"
        chain33_ImportPrivkey "0xaeef1ad76d43a2056d0dcb57d5bf1ba96471550614ab9e7f611ef9c5ca403f42" "1CvLe1qNaC7tCf5xmfAqJ9UJkMhtmhUKNg" "trade1" "$para_ip"

        chain33_applyCoins "$ACCOUNT_A" 12000000000 "${para_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$para_ip"
    fi

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
    token_sendExec "0xaeef1ad76d43a2056d0dcb57d5bf1ba96471550614ab9e7f611ef9c5ca403f42 "
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
    chain33_RpcTestBegin trade
    local ip=$1
    MAIN_HTTP=$ip
    echo "main_ip=$MAIN_HTTP"

    init
    run_test "$MAIN_HTTP"
    chain33_RpcTestRst trade "$CASE_ERR"
}

chain33_debug_function main "$1"
