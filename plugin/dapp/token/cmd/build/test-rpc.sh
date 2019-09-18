#!/usr/bin/env bash
# shellcheck disable=SC2128

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
CASE_ERR=""
tokenAddr="1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq"
tokenSymbol="ABCDE"
token_addr=""
execName="token"

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# $2=0 means true, other false
function echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="FAIL"
    fi
}

function chain33_unlock() {
    ok=$(curl -k -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.UnLock","params":[{"passwd":"1314fuzamei","timeout":0}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.isOK")
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

function signRawTx() {
    unsignedTx=$1
    addr=$2
    signedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SignRawTx","params":[{"addr":"'"${addr}"'","txHex":"'"${unsignedTx}"'","expire":"120s"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "$signedTx" == "null" ]; then
        return 1
    else
        return 0
    fi
}

function sendSignedTx() {
    txHash=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"token":"","data":"'"${signedTx}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "$txHash" == "null" ]; then
        return 1
    else
        return 0
    fi
}

# 查询交易的执行结果
# 根据传入的规则，校验查询的结果 （参数1: 校验规则 参数2: 预期匹配结果）
function queryTransaction() {
    validator=$1
    expectRes=$2
    echo "txhash=${txHash}"
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${txHash}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r "${validator}")
    if [ "${res}" != "${expectRes}" ]; then
        return 1
    else
        return 0
    fi
}

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    local main_ip=${MAIN_HTTP//8901/8801}
    #main chain import pri key
    #1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq
    chain33_ImportPrivkey "0x882c963ce2afbedc2353cb417492aa9e889becd878a10f2529fc9e6c3b756128" "1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq" "token1" "${main_ip}"
    local token1="1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq"
    if [ "$ispara" == false ]; then
        chain33_applyCoins "$token1" 12000000000 "${main_ip}"
        chain33_QueryBalance "${token1}" "$main_ip"
    else
        # tx fee
        chain33_applyCoins "$token1" 1000000000 "${main_ip}"
        chain33_QueryBalance "${token1}" "$main_ip"

        local para_ip="${MAIN_HTTP}"
        #para chain import pri key
        chain33_ImportPrivkey "0x882c963ce2afbedc2353cb417492aa9e889becd878a10f2529fc9e6c3b756128" "1CLrYLNhHfCfMUV7mtdqhbMSF6vGmtTvzq" "token1"  "$para_ip"

        chain33_applyCoins "$token1" 12000000000 "${para_ip}"
        chain33_QueryBalance "${token1}" "$para_ip"
    fi

    if [ "$ispara" == true ]; then
        execName="user.p.para.token"
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.token"}]}' ${MAIN_HTTP} | jq -r ".result")
        Chain33_SendToAddress "$tokenAddr" "$token_addr" 1000000000
        block_wait 2
    else
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"token"}]}' ${MAIN_HTTP} | jq -r ".result")
        Chain33_SendToAddress "$tokenAddr" "$token_addr" 1000000000
        block_wait 2
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

    signRawTx "${unsignedTx}" "${tokenAddr}"
    echo_rst "update config signRawTx" "$?"

    sendSignedTx
    echo_rst "update config sendSignedTx" "$?"

    block_wait 2

    queryTransaction ".error | not" "true"
    echo_rst "update config queryExecRes" "$?"
}
function token_preCreate() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenPreCreateTx","params":[{"name": "yinhebib", "symbol": "'"${tokenSymbol}"'", "total": 100000000000, "price": 100, "category": 1,"owner":"'${tokenAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token preCreate create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    echo_rst "token preCreate signRawTx" "$?"

    sendSignedTx
    echo_rst "token preCreate sendSignedTx" "$?"

    block_wait 2

    queryTransaction ".error | not" "true"
    echo_rst "token preCreate queryExecRes" "$?"
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

    signRawTx "${unsignedTx}" "${tokenAddr}"
    echo_rst "token finish signRawTx" "$?"

    sendSignedTx
    echo_rst "token finish sendSignedTx" "$?"

    block_wait 2

    queryTransaction ".error | not" "true"
    echo_rst "token finish queryExecRes" "$?"
}

function token_getFinishCreated() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${execName}"'","funcName":"GetTokens","payload":{"queryAll":true,"status":1,"tokens":[],"symbolOnly":false}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.tokens" | grep "symbol")

    if [ "${res}" != "" ]; then
        echo_rst "token get finishCreated create tx" 0
    else
        echo_rst "token get finishCreated create tx" 1
    fi
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

    signRawTx "${unsignedTx}" "${tokenAddr}"
    echo_rst "token burn signRawTx" "$?"

    sendSignedTx
    echo_rst "token burn sendSignedTx" "$?"

    block_wait 2

    queryTransaction ".error | not" "true"
    echo_rst "token burn queryExecRes" "$?"
}

function token_mint() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenMintTx","params":[{"symbol": "'"${tokenSymbol}"'","amount": 10000}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token mint create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    echo_rst "token mint signRawTx" "$?"

    sendSignedTx
    echo_rst "token mint sendSignedTx" "$?"

    block_wait 2

    queryTransaction ".error | not" "true"
    echo_rst "token mint queryExecRes" "$?"
}
function token_transfer() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"Transfer","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "1000000000", "note": "", "to": "'"${recvAddr}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token transfer create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    echo_rst "token transfer signRawTx" "$?"

    sendSignedTx
    echo_rst "token transfer sendSignedTx" "$?"

    block_wait 2

    queryTransaction ".error | not" "true"
    echo_rst "token transfer queryExecRes" "$?"
}

function token_sendExec() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"TransferToExec","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "10", "note": "", "to": "'"${token_addr}"'", "execName": "'"${execName}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token sendExec create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    echo_rst "token sendExec signRawTx" "$?"

    sendSignedTx
    echo_rst "token sendExec sendSignedTx" "$?"

    block_wait 2

    queryTransaction ".error | not" "true"
    echo_rst "token sendExec queryExecRes" "$?"
}

function token_withdraw() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"Withdraw","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "10", "note": "", "to": "'"${token_addr}"'", "execName": "'"${execName}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token withdraw create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    echo_rst "token withdraw signRawTx" "$?"

    sendSignedTx
    echo_rst "token withdraw sendSignedTx" "$?"

    block_wait 2

    queryTransaction ".error | not" "true"
    echo_rst "token withdraw queryExecRes" "$?"
}

function run_test() {
    local ip=$1
    set -x
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
    set +x
}

function main() {
    local ip=$1
    MAIN_HTTP=$ip
    echo "=========== # token rpc test ============="
    echo "main_ip=$MAIN_HTTP"

    init
    run_test "$ip"

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Token Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Token Rpc Test Pass==============${NOC}"
    fi
}

chain33_debug_function main "$1"
