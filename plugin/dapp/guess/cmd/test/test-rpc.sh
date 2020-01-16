#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -e
set -o pipefail

MAIN_HTTP=""
source ../dapp-test-common.sh

MAIN_HTTP=""
guess_admin_addr=12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
guess_user1_addr=1NrfEBfdFJUUqgbw5ZbHXhdew6NNQumYhM
guess_user2_addr=17tRkBrccmFiVcLPXgEceRxDzJ2WaDZumN
guess_addr=""
guess_exec=""

eventId=""
txhash=""

guess_game_start() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Start", "payload":{"topic":"WorldCup Final","options":"A:France;B:Claodia","category":"football","maxBetsOneTime":10000000000,"maxBetsNumber":100000000000,"devFeeFactor":5,"devFeeAddr":"1D6RFZNp2rh6QdbcZ1d7RWuBUz61We6SD7","platFeeFactor":5,"platFeeAddr":"1PHtChNt3UcfssR7v7trKSk3WJtAWjKjjX"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" ${MAIN_HTTP} "$FUNCNAME"
    eventId="${txhash}"
}

guess_game_bet() {
    local priv=$1
    local opt=$2
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Bet", "payload":{"gameID":"'"${eventId}"'","option":"'"${opt}"'", "betsNum":500000000}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "${priv}" ${MAIN_HTTP} "$FUNCNAME"
}

guess_game_stop() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"StopBet", "payload":{"gameID":"'"${eventId}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" ${MAIN_HTTP} "$FUNCNAME"
}

guess_game_publish() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Publish", "payload":{"gameID":"'"${eventId}"'","result":"A"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" ${MAIN_HTTP} "$FUNCNAME"
}

guess_game_abort() {
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Abort", "payload":{"gameID":"'"${eventId}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" ${MAIN_HTTP} "$FUNCNAME"
}

guess_QueryGameByID() {
    local event_id=$1
    local status=$2
    local req='{"method":"Chain33.Query", "params":[{"execer":"guess","funcName":"QueryGameByID","payload":{"gameID":"'"$event_id"'"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.result|has("game")) and (.result.game.status == '"$status"')' "$FUNCNAME"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        guess_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.guess"}]}' ${MAIN_HTTP} | jq -r ".result")
        guess_exec="user.p.para.guess"
    else
        guess_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"guess"}]}' ${MAIN_HTTP} | jq -r ".result")
        guess_exec="guess"
    fi
    echo "guess_addr=$guess_addr"

    local main_ip=${MAIN_HTTP//8901/8801}
    chain33_ImportPrivkey "0xc889d2958843fc96d4bd3f578173137d37230e580d65e9074545c61e7e9c1932" "1NrfEBfdFJUUqgbw5ZbHXhdew6NNQumYhM" "guess11" "${main_ip}"
    chain33_ImportPrivkey "0xf10c79470dc74c229c4ee73b05d14c58322b771a6c749d27824f6a59bb6c2d73" "17tRkBrccmFiVcLPXgEceRxDzJ2WaDZumN" "guess22" "$main_ip"

    local guess1="1NrfEBfdFJUUqgbw5ZbHXhdew6NNQumYhM"
    local guess2="17tRkBrccmFiVcLPXgEceRxDzJ2WaDZumN"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$guess1" 12000000000 "${main_ip}"
        chain33_QueryBalance "${guess1}" "$main_ip"

        chain33_applyCoins "$guess2" 12000000000 "${main_ip}"
        chain33_QueryBalance "${guess2}" "$main_ip"
    else
        chain33_applyCoins "$guess1" 1000000000 "${main_ip}"
        chain33_QueryBalance "${guess1}" "$main_ip"

        chain33_applyCoins "$guess2" 1000000000 "${main_ip}"
        chain33_QueryBalance "${guess2}" "$main_ip"
        local para_ip="${MAIN_HTTP}"
        chain33_ImportPrivkey "0xc889d2958843fc96d4bd3f578173137d37230e580d65e9074545c61e7e9c1932" "1NrfEBfdFJUUqgbw5ZbHXhdew6NNQumYhM" "guess11" "$para_ip"
        chain33_ImportPrivkey "0xf10c79470dc74c229c4ee73b05d14c58322b771a6c749d27824f6a59bb6c2d73" "17tRkBrccmFiVcLPXgEceRxDzJ2WaDZumN" "guess22" "$para_ip"

        chain33_applyCoins "$guess1" 12000000000 "${para_ip}"
        chain33_QueryBalance "${guess1}" "$para_ip"
        chain33_applyCoins "$guess2" 12000000000 "${para_ip}"
        chain33_QueryBalance "${guess2}" "$para_ip"
    fi

    chain33_SendToAddress "$guess1" "$guess_addr" 10000000000 ${MAIN_HTTP}
    chain33_QueryExecBalance "${guess1}" "guess" "$MAIN_HTTP"
    chain33_SendToAddress "$guess2" "$guess_addr" 10000000000 ${MAIN_HTTP}
    chain33_QueryExecBalance "${guess2}" "guess" "$MAIN_HTTP"

    chain33_BlockWait 1 "${MAIN_HTTP}"
}

function run_test() {
    #导入地址私钥
    chain33_ImportPrivkey "0xc889d2958843fc96d4bd3f578173137d37230e580d65e9074545c61e7e9c1932" "1NrfEBfdFJUUqgbw5ZbHXhdew6NNQumYhM" "user1" "$MAIN_HTTP"
    chain33_ImportPrivkey "0xf10c79470dc74c229c4ee73b05d14c58322b771a6c749d27824f6a59bb6c2d73" "17tRkBrccmFiVcLPXgEceRxDzJ2WaDZumN" "user2" "$MAIN_HTTP"
    chain33_ImportPrivkey "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv" "admin" "$MAIN_HTTP"

    chain33_QueryBalance "${guess_admin_addr}" "$MAIN_HTTP"
    chain33_QueryBalance "${guess_user1_addr}" "$MAIN_HTTP"
    chain33_QueryBalance "${guess_user2_addr}" "$MAIN_HTTP"
    chain33_QueryExecBalance "${guess_user1_addr}" "${guess_exec}" "$MAIN_HTTP"
    chain33_QueryExecBalance "${guess_user2_addr}" "${guess_exec}" "$MAIN_HTTP"

    #场景1：start -> bet -> bet -> stop -> publish
    #管理员创建游戏
    guess_game_start

    #查询游戏状态
    guess_QueryGameByID "$eventId" 11

    #用户1下注
    guess_game_bet "0xc889d2958843fc96d4bd3f578173137d37230e580d65e9074545c61e7e9c1932" "A"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12

    #用户2下注
    guess_game_bet "0xf10c79470dc74c229c4ee73b05d14c58322b771a6c749d27824f6a59bb6c2d73" "B"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12

    #管理员停止下注
    guess_game_stop

    #查询游戏状态
    guess_QueryGameByID "$eventId" 13

    #管理员发布结果
    guess_game_publish

    #查询游戏状态
    guess_QueryGameByID "$eventId" 15

    #查询余额
    chain33_QueryExecBalance "${guess_user1_addr}" "${guess_exec}" "$MAIN_HTTP"
    chain33_QueryExecBalance "${guess_user2_addr}" "${guess_exec}" "$MAIN_HTTP"

    #场景2：start->stop->abort
    guess_game_start

    #查询游戏状态
    guess_QueryGameByID "$eventId" 11

    #管理员停止下注
    guess_game_stop

    #查询游戏状态
    guess_QueryGameByID "$eventId" 13

    #管理员发布结果
    guess_game_abort

    #查询游戏状态
    guess_QueryGameByID "$eventId" 14

    #场景3：start->abort
    guess_game_start

    #查询游戏状态
    guess_QueryGameByID "$eventId" 11

    #管理员发布结果
    guess_game_abort

    #查询游戏状态
    guess_QueryGameByID "$eventId" 14

    #场景4：start->bet->abort

    #管理员创建游戏
    guess_game_start

    #查询游戏状态
    guess_QueryGameByID "$eventId" 11

    #用户1下注
    guess_game_bet "0xc889d2958843fc96d4bd3f578173137d37230e580d65e9074545c61e7e9c1932" "A"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12

    #用户2下注
    guess_game_bet "0xf10c79470dc74c229c4ee73b05d14c58322b771a6c749d27824f6a59bb6c2d73" "B"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12

    #管理员发布结果
    guess_game_abort

    #查询游戏状态
    guess_QueryGameByID "$eventId" 14

    #场景5：start->bet->stop->abort
    #管理员创建游戏
    guess_game_start

    #查询游戏状态
    guess_QueryGameByID "$eventId" 11

    #用户1下注
    guess_game_bet "0xc889d2958843fc96d4bd3f578173137d37230e580d65e9074545c61e7e9c1932" "A"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12

    #用户2下注
    guess_game_bet "0xf10c79470dc74c229c4ee73b05d14c58322b771a6c749d27824f6a59bb6c2d73" "B"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12

    #管理员停止下注
    guess_game_stop

    #查询游戏状态
    guess_QueryGameByID "$eventId" 13

    #管理员发布结果
    guess_game_abort

    #查询游戏状态
    guess_QueryGameByID "$eventId" 14

    #查询余额
    chain33_QueryExecBalance "${guess_user1_addr}" "${guess_exec}" "$MAIN_HTTP"
    chain33_QueryExecBalance "${guess_user2_addr}" "${guess_exec}" "$MAIN_HTTP"
}

function main() {
    chain33_RpcTestBegin guess
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_test
    chain33_RpcTestRst guess "$CASE_ERR"
}

chain33_debug_function main "$1"
