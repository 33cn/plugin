#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""

tokenExecName="token"
ExecName="evmxgo"
privateKey="0x4dcb00c7d01a3d377c0d5a14cd7ec91798a74c8b41896c5d21fc8b9bf4b40e42"

function updateConfig() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key":"evmxgo-mint-DOG","value":"{\"address\":\"address1234\",\"precision\":4,\"introduction\":\"介绍\"}","op":"add","addr":""}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "update config create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "$privateKey" "${unsignedTx}"
}

function evmxgo_mint() {
    local data='{"jsonrpc":"2.0","id":1,"method":"Chain33.CreateTransaction","params":[{"execer":"evmxgo","actionName":"Mint","payload":{"symbol":"DOG","amount":10000000}}]}'
    unsignedTx=$(curl -s --data-binary "$data" -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "evmxgo mint create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "$privateKey" "${unsignedTx}"
}

function evmxgo_burn() {
    local data='{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer":"evmxgo","actionName":"Burn","payload":{"symbol":"DOG","amount":10000000}}]}'
    unsignedTx=$(curl -s --data-binary "$data" -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "evmxgo mint create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "$privateKey" "${unsignedTx}"
}

function evmxgo_transfer() {
    addr=$1
    symbol=$2
    local data='{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${ExecName}"'","actionName":"Transfer","payload": {"cointoken":"'"${symbol}"'", "amount": "10000000", "note": "", "to": "'"${addr}"'"}}]}'
    unsignedTx=$(curl -s --data-binary "${data}" -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token transfer create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "$privateKey" "${unsignedTx}"
}

function evmxgo_transfer_exec() {
    addr=$1
    symbol=$2
    local data='{"jsonrpc":"2.0","id":1,"method":"Chain33.CreateTransaction","params":[{"execer":"'"${ExecName}"'","actionName":"TransferToExec","payload":{"cointoken":"'"${symbol}"'","amount":10000,"note":"","execName":"token","to":"12hpJBHybh1mSyCijQ2MQJPk7z7kZ7jnQa"}}]}'
    unsignedTx=$(curl -s --data-binary "${data}" -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token transfer create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "$privateKey" "${unsignedTx}"
}

function evmxgo_withdraw() {
    addr=$1
    symbol=$2
    local data='{"jsonrpc":"2.0","id":1,"method":"Chain33.CreateTransaction","params":[{"execer":"'"${ExecName}"'","actionName":"Withdraw","payload":{"cointoken":"'"${symbol}"'","amount":1000,"note":"","execName":"'"${tokenExecName}"'","to":"'"${addr}"'"}}]}'
    unsignedTx=$(curl -s --data-binary "${data}" -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token transfer create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME" "$privateKey" "${unsignedTx}"
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
    chain33_SignAndSendTx "$3" "$2" "${MAIN_HTTP}"
    queryTransaction ".error | not" "true"
    echo_rst "$1 queryExecRes" "$?"
}

function init() {
    updateConfig
}

function run_test() {
    local ip=$1

    evmxgo_mint
    evmxgo_burn
    evmxgo_transfer "17EVv6tW2HzE73TVB6YXQYThQJxa7kuZb8" "DOG"
    evmxgo_transfer_exec "12hpJBHybh1mSyCijQ2MQJPk7z7kZ7jnQa" "DOG"
    evmxgo_withdraw "12hpJBHybh1mSyCijQ2MQJPk7z7kZ7jnQa" "DOG"
}

function main() {
    chain33_RpcTestBegin evmxgo
    local ip=$1
    MAIN_HTTP=$ip
    echo "main_ip=$MAIN_HTTP"

    init
    run_test "$MAIN_HTTP"

    chain33_RpcTestRst evmxgo "$CASE_ERR"
}

chain33_debug_function main "http://ip:port/"
