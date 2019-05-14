#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
PARA_HTTP=""
CASE_ERR=""
UNIT_HTTP=""

# $2=0 means true, other false
echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo "$1 ok"
    else
        echo "$1 err"
        CASE_ERR="err"
    fi

}

gID=""
gResp=""


glAddr=""
gameAddr1=""
gameAddr2=""
gameAddr3=""

chain33_NewAccount() {
    label=$1
    result=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.NewAccount","params":[{"label":"'"$label"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.acc.addr")
    [ "$result" != "" ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    glAddr=$result
    echo $glAddr
}

function block_wait() {
    if [ "$#" -lt 1 ]; then
        echo "wrong block_wait params"
        exit 1
    fi
    cur_height=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.GetLastHeader","params":[{}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.height")
    expect=$((cur_height + ${1}))
    local count=0
    while true; do
        new_height=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.GetLastHeader","params":[{}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 1
    done
    echo "wait new block $count s, cur height=$expect,old=$cur_height"
}

chain33_SendToAddress() {
    from=$1
    to=$2
    amount=$3
    note="test"
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.SendToAddress","params":[{"from":"'"$from"'","to":"'"$to"'","amount":'"$amount"',"note":"'"$note"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    ok=$(jq '(.error|not)' <<<"$resp")
	[ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

chain33_SendTransaction() {
    rawTx=$1
    addr=$2
    #签名交易
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.SignRawTx","params":[{"addr":"'"$addr"'","txHex":"'"$rawTx"'","expire":"120s","fee":10000000,"index":0}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [ "$ok" == null ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

    signTx=$(echo "${resp}" | jq -r ".result")
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"data":"'"$signTx"'"}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [ "$ok" == null ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #返回交易
    gResp=$(echo "${resp}" | jq -r ".result")
    echo "tx hash is $gResp"
}


blackwhite_BlackwhiteCreateTx() {
    #创建交易
    addr=$1
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"blackwhite.BlackwhiteCreateTx","params":[{"PlayAmount":100000000,"PlayerCount":3,"GameName":"hello","Timeout":600,"Fee":1000000}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [ "$ok" == null ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SendTransaction  "${rawTx}" "${addr}"
    gID="${gResp}"
    echo "gameID $gID"
}

blackwhite_BlackwhitePlayTx() {
    addr=$1
    round1=$2
    round2=$3
    round3=$4
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"blackwhite.BlackwhitePlayTx","params":[{"gameID":"'"$gID"'","amount":100000000,"Fee":1000000,"hashValues":["'"$round1"'","'"$round2"'","'"$round3"'"]}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [ "$ok" == null ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SendTransaction  "${rawTx}" "${addr}"
}

blackwhite_BlackwhiteShowTx() {
    addr=$1
    sec=$2
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"blackwhite.BlackwhiteShowTx","params":[{"gameID":"'"$gID"'","secret":"'"$sec"'","Fee":1000000}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [ "$ok" == null ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    #发送交易
    rawTx=$(echo "${resp}" | jq -r ".result")
    chain33_SendTransaction  "${rawTx}" "${addr}"
}

blackwhite_BlackwhiteTimeoutDoneTx() {
    gameID=$1
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"blackwhite.BlackwhiteTimeoutDoneTx","params":[{"gameID":"'"$gameID"'","Fee":1000000}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [ "$ok" == null ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

blackwhite_GetBlackwhiteRoundInfo() {
    gameID=$1
    execer="blackwhite"
    funcName="GetBlackwhiteRoundInfo"
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"gameID":"'"$gameID"'"}}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(jq '(.error|not) and (.result.round | [has("gameID", "status", "playAmount", "playerCount", "curPlayerCount", "loop", "curShowCount", "timeout"),true] | unique | length == 1)' <<<"$resp")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

blackwhite_GetBlackwhiteByStatusAndAddr() {
    gameID=$1
    addr=$2
    execer="blackwhite"
    funcName="GetBlackwhiteByStatusAndAddr"
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"status":5,"address":"'"$addr"'","count":1,"direction":0,"index":-1}}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(jq '(.error|not) and (.result.round[0].createAddr == "'"$addr"'") and (.result.round[0].status == 5) and (.result.round[0] | [has("gameID", "status", "playAmount", "playerCount", "curPlayerCount", "loop", "curShowCount", "timeout", "winner"),true] | unique | length == 1)' <<<"$resp")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

blackwhite_GetBlackwhiteloopResult() {
    gameID=$1
    execer="blackwhite"
    funcName="GetBlackwhiteloopResult"
    resp=$(curl -ksd '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"'"$execer"'","funcName":"'"$funcName"'","payload":{"gameID":"'"$gameID"'","loopSeq":0}}]}' -H 'content-type:text/plain;' ${UNIT_HTTP})
    ok=$(jq '(.error|not) and (.result.gameID == "'"$gameID"'") and (.result.results|length >= 1)' <<<"$resp")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function run_testcases() {
    #密钥
    sect1="123"
    #结果base64.StdEncoding.EncodeToString(common.Sha256([]byte("0"+secret+black)))
    # black == "1" white := "0"
    #black0="O3LD8NyaeeSCc8xDfvBoacTrQlrY91FHT9ceEOXgs18="
    black1="6vm6gJ2wvEIxC8Yc6r/N6lIU5OZk633YMnIfwcZBD0o="
    black2="6FXx5aeDSCaq1UrhLO8u0H31Hl8TpvzxuHrgGo9WeFk="
    white0="DrNPzA68XiGimZE/igx70kTPJxnIJnVf8NCGnb7XoYU="
    white1="SB5Pnf6Umf2Wba0dqyNOezq5FEqTd22WPVYAhSA6Lxs="
    #white2="OiexKDzIlS1CKr3KBNWEY1k5uXzDI/ou6Dd+x0ByQCM="

    #先创建账户地址
    chain33_NewAccount "label189"
    gameAddr1="${glAddr}"
    chain33_NewAccount "label289"
    gameAddr2="${glAddr}"
    chain33_NewAccount "label389"
    gameAddr3="${glAddr}"

    #给每个账户分别转帐
    origAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    chain33_SendToAddress  "${origAddr}" "${gameAddr1}" 1000000000
    chain33_SendToAddress  "${origAddr}" "${gameAddr2}" 1000000000
    chain33_SendToAddress  "${origAddr}" "${gameAddr3}" 1000000000

    block_wait 1

    #给游戏合约中转帐
    bwExecAddr="146wei89zoX5TNQKATBJmduNPEtSKTXi1z"
    chain33_SendToAddress  "${gameAddr1}" "${bwExecAddr}" 500000000
    chain33_SendToAddress  "${gameAddr2}" "${bwExecAddr}" 500000000
    chain33_SendToAddress  "${gameAddr3}" "${bwExecAddr}" 500000000

    block_wait 1
    blackwhite_BlackwhiteCreateTx "${gameAddr1}"

    block_wait 1
    blackwhite_BlackwhitePlayTx  "${gameAddr1}"  "${white0}" "${white1}" "${black2}"
    blackwhite_BlackwhitePlayTx  "${gameAddr2}"  "${white0}" "${black1}" "${black2}"
    blackwhite_BlackwhitePlayTx  "${gameAddr3}"  "${white0}" "${black1}" "${black2}"

    block_wait 1
    blackwhite_BlackwhiteShowTx "${gameAddr1}"  "${sect1}"
    blackwhite_BlackwhiteShowTx "${gameAddr2}"  "${sect1}"
    blackwhite_BlackwhiteShowTx "${gameAddr3}"  "${sect1}"

    blackwhite_BlackwhiteTimeoutDoneTx "$gID"
    #查询部分
    block_wait 1
    blackwhite_GetBlackwhiteRoundInfo "$gID"
    blackwhite_GetBlackwhiteByStatusAndAddr "$gID" "${gameAddr1}"
    blackwhite_GetBlackwhiteloopResult  "$gID"


}
function paracross_rpc_test() {
    local ip=$1
    MAIN_HTTP="http://$ip:8801"
    PARA_HTTP="http://$ip:8901"
    echo "=========== # paracross rpc test ============="
    echo "main_ip=$MAIN_HTTP,para_ip=$PARA_HTTP"

    UNIT_HTTP=$MAIN_HTTP
    run_testcases

    if [ -n "$CASE_ERR" ]; then
        echo "paracross there some case error"
        exit 1
    fi
}

paracross_rpc_test 172.19.0.6

#UNIT_HTTP="http://172.19.0.6:8801"

#blackwhite_GetBlackwhiteRoundInfo  "0xb7b61ea3412d528b956c91df76c26bb8e3adcb7ffce11b0a1008bb7d0f52d729"
#blackwhite_GetBlackwhiteByStatusAndAddr "0xb7b61ea3412d528b956c91df76c26bb8e3adcb7ffce11b0a1008bb7d0f52d729" "18Y87cw2hiYC71bvpD872oYMYXtw66Qp6o"
#blackwhite_GetBlackwhiteloopResult  "0xb7b61ea3412d528b956c91df76c26bb8e3adcb7ffce11b0a1008bb7d0f52d729"

#blackwhite_BlackwhiteTimeoutDoneTx  "0xb7b61ea3412d528b956c91df76c26bb8e3adcb7ffce11b0a1008bb7d0f52d729"