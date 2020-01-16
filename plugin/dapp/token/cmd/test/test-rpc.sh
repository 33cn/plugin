#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
tokenAddr="1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"
recvAddr="1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq"
superManager="0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc"
tokenSymbol="ABCDE"
token_addr=""
execName="token"

# 查询交易的执行结果
# 根据传入的规则，校验查询的结果 （参数1: 校验规则 参数2: 预期匹配结果）
function queryTransaction() {
    validator=$1
    expectRes=$2
    #  echo "txhash=${RAW_TX_HASH}"
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${RAW_TX_HASH}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r "${validator}")
    if [ "${res}" != "${expectRes}" ]; then
        return 1
    else
        return 0
    fi
}

function signRawTxAndQuery() {
    chain33_SignAndSendTx "${unsignedTx}" "${superManager}" "${MAIN_HTTP}"
    queryTransaction ".error | not" "true"
    echo_rst "$1 queryExecRes" "$?"
}

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    chain33_ImportPrivkey "${superManager}" "${tokenAddr}" "tokenAddr" "${MAIN_HTTP}"

    local main_ip=${MAIN_HTTP//8901/8801}
    chain33_ImportPrivkey "0x882c963ce2afbedc2353cb417492aa9e889becd878a10f2529fc9e6c3b756128" "1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq" "token1" "${main_ip}"

    local ACCOUNT_A="1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$ACCOUNT_A" 12000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$main_ip"
    else
        chain33_applyCoins "$ACCOUNT_A" 1000000000 "${main_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$main_ip"

        local para_ip="${MAIN_HTTP}"
        chain33_ImportPrivkey "0x882c963ce2afbedc2353cb417492aa9e889becd878a10f2529fc9e6c3b756128" "1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq" "token1" "$para_ip"

        chain33_applyCoins "$ACCOUNT_A" 12000000000 "${para_ip}"
        chain33_QueryBalance "${ACCOUNT_A}" "$para_ip"
    fi

    if [ "$ispara" == true ]; then
        execName="user.p.para.token"
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.token"}]}' ${MAIN_HTTP} | jq -r ".result")
        chain33_SendToAddress "$recvAddr" "$tokenAddr" 10000000000 "${MAIN_HTTP}"
        chain33_BlockWait 2 "${MAIN_HTTP}"
        chain33_SendToAddress "$tokenAddr" "$token_addr" 1000000000 "${MAIN_HTTP}"
        chain33_BlockWait 2 "${MAIN_HTTP}"
    else
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"token"}]}' ${MAIN_HTTP} | jq -r ".result")
        from="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
        chain33_SendToAddress "$from" "$tokenAddr" 10000000000 "${MAIN_HTTP}"
        chain33_BlockWait 2 "${MAIN_HTTP}"
        chain33_SendToAddress "$tokenAddr" "$token_addr" 1000000000 "${MAIN_HTTP}"
        chain33_BlockWait 2 "${MAIN_HTTP}"
    fi
    echo "token=$token_addr"
    updateConfig
}

function updateConfig() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "manage","actionName":"Modify","payload":{ "key": "token-blacklist","value": "BTY","op": "add","addr": ""}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "update config create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME"
}

function token_preCreate() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenPreCreateTx","params":[{"name": "yinhebib", "symbol": "'"${tokenSymbol}"'", "total": 100000000000, "price": 100, "category": 1,"owner":"'${tokenAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token preCreate create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME"
}

function token_getPreCreated() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${execName}"'","funcName":"GetTokens","payload":{"queryAll":true,"status":0,"tokens":[],"symbolOnly":false}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".error | not")
    if [ "${res}" != "true" ]; then
        echo_rst "token preCreate create tx" 1
        return
    fi
}

function token_finish() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenFinishTx","params":[{"symbol": "'"${tokenSymbol}"'", "owner":"'${tokenAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token finish create tx" 1
        return
    fi

    signRawTxAndQuery "$FUNCNAME"
}

function token_getFinishCreated() {
    req='{"method":"Chain33.Query","params":[{"execer":"'"${execName}"'","funcName":"GetTokens","payload":{"queryAll":true,"status":1,"tokens":[],"symbolOnly":false}}]}'
    chain33_Http "$req" ${MAIN_HTTP} "(.result.tokens[0].symbol != null)" "$FUNCNAME"
}

function token_assets() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer": "'"${execName}"'","funcName":"GetAccountTokenAssets","payload": {"address":"'"${recvAddr}"'", "execer": "token"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})

    if [ "${res}" == "" ]; then
        echo_rst "token get balance tx" 1
        return
    fi

    tokenInfo=$(echo "${res}" | jq -r '.result.tokenAssets' | grep -A 6 -B 1 "${tokenSymbol}")
    addr=$(echo "${tokenInfo}" | grep "addr" | awk -F '"' '{print $4}')
    balance=$(echo "${tokenInfo}" | grep "balance" | awk -F '"' '{print $4}')

    if [ "${addr}" == "${recvAddr}" ] && [ "${balance}" -eq 1000000000 ]; then
        echo_rst "token get assets tx" 0
    else
        echo_rst "token get assets tx" 1
    fi
}

function token_balance() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.GetTokenBalance","params":[{"addresses": ["'${tokenAddr}'"],"tokenSymbol":"'"${tokenSymbol}"'","execer": "'"${execName}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})

    if [ "${res}" == "" ]; then
        echo_rst "token get balance tx" 1
        return
    fi

    addr=$(echo "${res}" | jq -r ".result[0].addr")
    balance=$(echo "${res}" | jq -r ".result[0].balance")

    if [ "${addr}" == "${tokenAddr}" ] && [ "${balance}" -eq 100000000000 ]; then
        echo_rst "token get balance tx" 0
    else
        echo_rst "token get balance tx" 1
    fi
}

function token_burn() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenBurnTx","params":[{"symbol": "'"${tokenSymbol}"'","amount": 10000}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token burn create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME"
}

function token_mint() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenMintTx","params":[{"symbol": "'"${tokenSymbol}"'","amount": 10000}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token mint create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME"
}
function token_transfer() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"Transfer","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "1000000000", "note": "", "to": "'"${recvAddr}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token transfer create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME"
}

function token_sendExec() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"TransferToExec","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "10", "note": "", "to": "'"${token_addr}"'", "execName": "'"${execName}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token sendExec create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME"
}

function token_withdraw() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"Withdraw","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "10", "note": "", "to": "'"${token_addr}"'", "execName": "'"${execName}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token withdraw create tx" 1
        return
    fi
    signRawTxAndQuery "$FUNCNAME"
}

function run_test() {
    local ip=$1
    token_preCreate
    token_getPreCreated

    token_finish
    token_getFinishCreated

    token_balance
    token_burn
    token_mint
    token_transfer
    token_sendExec
    token_assets
    token_withdraw
}

function main() {
    chain33_RpcTestBegin token
    local ip=$1
    MAIN_HTTP=$ip
    echo "main_ip=$MAIN_HTTP"

    init
    run_test "$ip"
    chain33_RpcTestRst token "$CASE_ERR"
}

chain33_debug_function main "$1"
