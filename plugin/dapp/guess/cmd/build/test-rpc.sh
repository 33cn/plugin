#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
CASE_ERR=""
guess_admin_addr=12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
guess_user1_addr=1PUiGcbsccfxW3zuvHXZBJfznziph5miAo
guess_user2_addr=1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX
guess_addr=""
guess_exec=""

eventId=""
txhash=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

guess_game_start() {
    echo "========== # guess start tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Start", "payload":{"topic":"WorldCup Final","options":"A:France;B:Claodia","category":"football","maxBetsOneTime":10000000000,"maxBetsNumber":100000000000,"devFeeFactor":5,"devFeeAddr":"1D6RFZNp2rh6QdbcZ1d7RWuBUz61We6SD7","platFeeFactor":5,"platFeeAddr":"1PHtChNt3UcfssR7v7trKSk3WJtAWjKjjX"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" ${MAIN_HTTP}

    eventId="${txhash}"
    echo "eventId $eventId"
    echo "========== # guess start tx end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}
}

guess_game_bet() {
    local priv=$1
    local opt=$2

    echo "========== # guess bet tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Bet", "payload":{"gameID":"'"${eventId}"'","option":"'"${opt}"'", "betsNum":500000000}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "${priv}" ${MAIN_HTTP}

    echo "========== # guess bet tx end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}
}

guess_game_stop() {
    echo "========== # guess stop tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"StopBet", "payload":{"gameID":"'"${eventId}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" ${MAIN_HTTP}

    echo "========== # guess stop tx end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}
}

guess_game_publish() {
    echo "========== # guess publish tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Publish", "payload":{"gameID":"'"${eventId}"'","result":"A"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" ${MAIN_HTTP}

    echo "========== # guess publish tx end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}
}

guess_game_abort() {
    echo "========== # guess abort tx begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Abort", "payload":{"gameID":"'"${eventId}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignRawTx "$tx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" ${MAIN_HTTP}

    echo "========== # guess abort tx end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}
}

guess_QueryGameByID() {
    local event_id=$1
    local status=$2
    echo "========== # guess QueryGameByID begin =========="
    local req='"method":"Chain33.Query", "params":[{"execer":"guess","funcName":"QueryGameByID","payload":{"gameID":"'"$event_id"'"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.result|has("game")) and (.result.game.status == '"$status"')' <<<"$resp")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    echo "========== # guess QueryGameByID end =========="
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

    local from="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    chain33_SendToAddress "$from" "$guess_addr" 10000000000 ${MAIN_HTTP}

    from="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    chain33_SendToAddress "$from" "$guess_addr" 10000000000 ${MAIN_HTTP}

    chain33_BlockWait 1 "${MAIN_HTTP}"
}


function run_test() {

    #导入地址私钥
    chain33_ImportPrivkey "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo" "user1" "$MAIN_HTTP"
    chain33_ImportPrivkey "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" "1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX" "user2" "$MAIN_HTTP"
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
    guess_game_bet "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" "A"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12


    #用户2下注
    guess_game_bet "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" "B"

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
    guess_game_bet "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" "A"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12

    #用户2下注
    guess_game_bet "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" "B"

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
    guess_game_bet "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" "A"

    #查询游戏状态
    guess_QueryGameByID "$eventId" 12

    #用户2下注
    guess_game_bet "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" "B"

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

    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    echo "=========== # guess rpc test start============="
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Guess Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Guess Rpc Test Pass==============${NOC}"
    fi
    echo "=========== # guess rpc test end============="
}

main "$1"
