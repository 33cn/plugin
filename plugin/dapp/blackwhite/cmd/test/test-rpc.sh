#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set +e
set -o pipefail

MAIN_HTTP=""
source ../dapp-test-common.sh

gID=""
gResp=""

glAddr=""
gameAddr1=""
gameAddr2=""
gameAddr3=""
bwExecAddr=""

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    if [ "$ispara" == true ]; then
        bwExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.blackwhite"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        bwExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"blackwhite"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "bwExecAddr=$bwExecAddr"
}

chain33_NewAccount() {
    label=$1
    req='{"method":"Chain33.NewAccount","params":[{"label":"'"$label"'"}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result.acc.addr|length > 0)' "$FUNCNAME" ".result.acc.addr"
    glAddr=$RETURN_RESP
}

chain33_SendTransaction() {
    rawTx=$1
    addr=$2
    #签名交易
    req='{"method":"Chain33.SignRawTx","params":[{"addr":"'"$addr"'","txHex":"'"$rawTx"'","expire":"120s","fee":10000000,"index":0}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "Chain33.SignRawTx" ".result"
    signTx=$RETURN_RESP

    req='{"method":"Chain33.SendTransaction","params":[{"data":"'"$signTx"'"}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result"

    gResp=$RETURN_RESP
    #返回交易
    chain33_QueryTx "$RETURN_RESP" "${MAIN_HTTP}"
}

blackwhite_BlackwhiteCreateTx() {
    #创建交易
    addr=$1
    req='{"method":"blackwhite.BlackwhiteCreateTx","params":[{"PlayAmount":100000000,"PlayerCount":3,"GameName":"hello","Timeout":600,"Fee":1000000}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result"
    #发送交易
    chain33_SendTransaction "$RETURN_RESP" "${addr}"
    gID="${gResp}"
}

blackwhite_BlackwhitePlayTx() {
    addr=$1
    round1=$2
    round2=$3
    round3=$4
    req='{"method":"blackwhite.BlackwhitePlayTx","params":[{"gameID":"'"$gID"'","amount":100000000,"Fee":1000000,"hashValues":["'"$round1"'","'"$round2"'","'"$round3"'"]}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result"

    #发送交易
    chain33_SendTransaction "$RETURN_RESP" "${addr}"
}

blackwhite_BlackwhiteShowTx() {
    addr=$1
    sec=$2
    req='{"method":"blackwhite.BlackwhiteShowTx","params":[{"gameID":"'"$gID"'","secret":"'"$sec"'","Fee":1000000}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME" ".result"
    chain33_SendTransaction "$RETURN_RESP" "${addr}"
}

blackwhite_BlackwhiteTimeoutDoneTx() {
    gameID=$1
    req='{"method":"blackwhite.BlackwhiteTimeoutDoneTx","params":[{"gameID":"'"$gameID"'","Fee":1000000}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not)' "$FUNCNAME"
}

blackwhite_GetBlackwhiteRoundInfo() {
    gameID=$1
    req='{"method":"Chain33.Query","params":[{"execer":"blackwhite","funcName":"GetBlackwhiteRoundInfo","payload":{"gameID":"'"$gameID"'"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result.round | [has("gameID", "status", "playAmount", "playerCount", "curPlayerCount", "loop", "curShowCount", "timeout"),true] | unique | length == 1)' "$FUNCNAME"
}

blackwhite_GetBlackwhiteByStatusAndAddr() {
    addr=$1
    req='{"method":"Chain33.Query","params":[{"execer":"blackwhite","funcName":"GetBlackwhiteByStatusAndAddr","payload":{"status":5,"address":"'"$addr"'","count":1,"direction":0,"index":-1}}]}'
    resok='(.error|not) and (.result.round[0].createAddr == "'"$addr"'") and (.result.round[0].status == 5) and (.result.round[0] | [has("gameID", "status", "playAmount", "playerCount", "curPlayerCount", "loop", "curShowCount", "timeout", "winner"),true] | unique | length == 1)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

blackwhite_GetBlackwhiteloopResult() {
    gameID=$1
    req='{"method":"Chain33.Query","params":[{"execer":"blackwhite","funcName":"GetBlackwhiteloopResult","payload":{"gameID":"'"$gameID"'","loopSeq":0}}]}'
    resok='(.error|not) and (.result.gameID == "'"$gameID"'") and (.result.results|length >= 1)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

function run_testcases() {
    #密钥
    sect1="123"
    black1="6vm6gJ2wvEIxC8Yc6r/N6lIU5OZk633YMnIfwcZBD0o="
    black2="6FXx5aeDSCaq1UrhLO8u0H31Hl8TpvzxuHrgGo9WeFk="
    white0="DrNPzA68XiGimZE/igx70kTPJxnIJnVf8NCGnb7XoYU="
    white1="SB5Pnf6Umf2Wba0dqyNOezq5FEqTd22WPVYAhSA6Lxs="

    #先创建账户地址
    chain33_NewAccount "label188"
    gameAddr1="${glAddr}"
    chain33_NewAccount "label288"
    gameAddr2="${glAddr}"
    chain33_NewAccount "label388"
    gameAddr3="${glAddr}"

    #给每个账户分别转帐
    origAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"

    chain33_GetAccounts "${MAIN_HTTP}"

    #主链中相应账户需要转帐
    M_HTTP=${MAIN_HTTP//8901/8801}
    chain33_SendToAddress "${origAddr}" "${gameAddr1}" 1000000000 "${M_HTTP}"
    chain33_SendToAddress "${origAddr}" "${gameAddr2}" 1000000000 "${M_HTTP}"
    chain33_SendToAddress "${origAddr}" "${gameAddr3}" 1000000000 "${M_HTTP}"

    #平行链相应账户需要转帐
    chain33_SendToAddress "${origAddr}" "${gameAddr1}" 1000000000 "${MAIN_HTTP}"
    chain33_SendToAddress "${origAddr}" "${gameAddr2}" 1000000000 "${MAIN_HTTP}"
    chain33_SendToAddress "${origAddr}" "${gameAddr3}" 1000000000 "${MAIN_HTTP}"

    #给游戏合约中转帐
    chain33_SendToAddress "${gameAddr1}" "${bwExecAddr}" 500000000 "${MAIN_HTTP}"
    chain33_SendToAddress "${gameAddr2}" "${bwExecAddr}" 500000000 "${MAIN_HTTP}"
    chain33_SendToAddress "${gameAddr3}" "${bwExecAddr}" 500000000 "${MAIN_HTTP}"

    blackwhite_BlackwhiteCreateTx "${gameAddr1}"

    blackwhite_BlackwhitePlayTx "${gameAddr1}" "${white0}" "${white1}" "${black2}"
    blackwhite_BlackwhitePlayTx "${gameAddr2}" "${white0}" "${black1}" "${black2}"
    blackwhite_BlackwhitePlayTx "${gameAddr3}" "${white0}" "${black1}" "${black2}"

    blackwhite_BlackwhiteShowTx "${gameAddr1}" "${sect1}"
    blackwhite_BlackwhiteShowTx "${gameAddr2}" "${sect1}"
    blackwhite_BlackwhiteShowTx "${gameAddr3}" "${sect1}"

    blackwhite_BlackwhiteTimeoutDoneTx "$gID"
    #查询部分
    blackwhite_GetBlackwhiteRoundInfo "$gID"
    blackwhite_GetBlackwhiteByStatusAndAddr "${gameAddr1}"
    blackwhite_GetBlackwhiteloopResult "$gID"
}

function main() {
    chain33_RpcTestBegin blackwhite
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases
    chain33_RpcTestRst blackwhite "$CASE_ERR"
}

chain33_debug_function main "$1"
