#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""
tokenAddr="1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"
recvAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
superManager="0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc"
tokenSymbol="ABE"
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

function chain33_ImportPrivkey() {
    local pri=$2
    local acc=$3
    local req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"'"$pri"'", "label":"tokenAddr"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$1")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.label=="tokenAddr") and (.result.acc.addr == "'"$acc"'")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

function Chain33_SendToAddress() {
    local from="$1"
    local to="$2"
    local amount=$3
    local req='"method":"Chain33.SendToAddress", "params":[{"from":"'"$from"'","to":"'"$to"'", "amount":'"$amount"', "note":"test\n"}]'
    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.hash|length==66)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    hash=$(jq '(.result.hash)' <<<"$resp")
    echo "hash=$hash"
    #    query_tx "$hash"
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
    txHash=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"token":"","data":"'${signedTx}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
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
    res=`curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'${txHash}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r "${validator}"`
    if [ "${res}" != "${expectRes}" ]; then
        return 1
    else
        token_addr=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'${txHash}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.receipt.logs[1].log.contractName")
        return 0
    fi
}

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    chain33_ImportPrivkey "${MAIN_HTTP}" "${superManager}" "${tokenAddr}"

    if [ "$ispara" == true ]; then
        execName="user.p.para.token"
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.token"}]}' ${MAIN_HTTP} | jq -r ".result")
        Chain33_SendToAddress "$recvAddr" "$tokenAddr" 100000000000
        block_wait 1
        Chain33_SendToAddress "$tokenAddr" "$token_addr" 1000000000
        block_wait 1
    else
        token_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"token"}]}' ${MAIN_HTTP} | jq -r ".result")
        from="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
        Chain33_SendToAddress "$from" "$tokenAddr" 10000000000
        block_wait 1
        Chain33_SendToAddress "$tokenAddr" "$token_addr" 1000000000
        block_wait 1
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
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "update config signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "update config sendSignedTx" "$rst"
        return
    fi

    block_wait 1

    queryTransaction "${validator}" "${expectRes}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "update config queryExecRes" "$rst"
    fi
}
function token_preCreate() {
    validator=$1
    expectRes=$2

    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenPreCreateTx","params":[{"name": "yinhebib", "symbol": "'"${tokenSymbol}"'", "total": 100000000000, "price": 100, "category": 1,"owner":"'${tokenAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token preCreate create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token preCreate signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token preCreate sendSignedTx" "$rst"
        return
    fi

    block_wait 1

    queryTransaction "${validator}" "${expectRes}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token preCreate queryExecRes" "$rst"
    fi
}

function token_getPreCreated() {
    validator=$1
    expectRes=$2
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${execName}"'","funcName":"GetTokens","payload":{"queryAll":true,"status":0,"tokens":[],"symbolOnly":false}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${res}" != "" -a "" ]; then
        echo_rst "token preCreate create tx" 1
        return
    fi

}

function token_finish() {
    validator=$1
    expectRes=$2

    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenFinishTx","params":[{"symbol": "'"${tokenSymbol}"'", "owner":"'${tokenAddr}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token finish create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token finish signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token finish sendSignedTx" "$rst"
        return
    fi

    block_wait 1

    queryTransaction "${validator}" "${expectRes}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token finish queryExecRes" "$rst"
    fi
}

function token_getFinishCreated() {
    validator=$1
    expectRes=$2

    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"${execName}"'","funcName":"GetTokens","payload":{"queryAll":true,"status":1,"tokens":[],"symbolOnly":false}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.tokens" | grep "symbol")

    if [[ "${res}" =~ "${tokenSymbol}" ]]; then
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

    tokenInfo=`echo ${res} | jq -r '.result.tokenAssets'  | grep -A 6 -B 1 "${tokenSymbol}"`
    addr=`echo ${tokenInfo} | awk -F '"' '{print $20}'`
    balance=`echo ${tokenInfo} | awk -F '"' '{print $12}'`

    if [ "${addr}" == "${recvAddr}" ] && [ ${balance} -eq 1000000000 ]; then
        echo_rst "token get assets tx" 0
    else
        echo_rst "token get assets tx" 1
    fi

}
function token_balance() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.GetTokenBalance","params":[{"addresses": ["'${tokenAddr}'"],"tokenSymbol":"'"${tokenSymbol}"'","execer": "'"${execName}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} )

    if [ "${res}" == "" ]; then
        echo_rst "token get balance tx" 1
        return
    fi

    addr=`echo ${res} | jq -r ".result[0].addr"`
    balance=`echo ${res} | jq -r ".result[0].balance"`

    if [ "${addr}" == "${tokenAddr}" -a ${balance} -eq 100000000000 ]; then
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
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token burn signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token burn sendSignedTx" "$rst"
        return
    fi

    block_wait 1

    queryTransaction ".error | not" "true"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token burn queryExecRes" "$rst"
    fi
}

function token_mint() {
    validator=$1
    expectRes=$2
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenMintTx","params":[{"symbol": "'"${tokenSymbol}"'","amount": 10000}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token mint create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token mint signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token mint sendSignedTx" "$rst"
        return
    fi

    block_wait 1

    queryTransaction ".error | not" "true"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token mint queryExecRes" "$rst"
    fi
}
function token_transfer() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"Transfer","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "1000000000", "note": "", "to": "'"${recvAddr}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token transfer create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token transfer signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token transfer sendSignedTx" "$rst"
        return
    fi

    block_wait 1

    queryTransaction "${validator}" "${expectRes}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token transfer queryExecRes" "$rst"
    fi
}

function token_sendExec() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"TransferToExec","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "10", "note": "", "to": "'"${token_addr}"'", "execName": "'"${execName}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token sendExec create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token sendExec signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token sendExec sendSignedTx" "$rst"
        return
    fi

    block_wait 1

    queryTransaction "${validator}" "${expectRes}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token sendExec queryExecRes" "$rst"
    fi
}


function token_withdraw() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${execName}"'","actionName":"Withdraw","payload": {"cointoken":"'"${tokenSymbol}"'", "amount": "10", "note": "", "to": "'"${token_addr}"'", "execName": "'"${execName}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token withdraw create tx" 1
        return
    fi

    signRawTx "${unsignedTx}" "${tokenAddr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token withdraw signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token withdraw sendSignedTx" "$rst"
        return
    fi

    block_wait 1

    queryTransaction "${validator}" "${expectRes}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "token withdraw queryExecRes" "$rst"
    fi
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

main "$1"