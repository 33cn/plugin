#!/usr/bin/env bash
#set -x

RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

cli="docker exec build_chain33_1 ./chain33-cli"
#cli="./chain33-cli"

# 判断 chain33 金额是否正确
#addr amount
function balance_enough() {
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
    fi

    while true; do
        result=$(${cli} account balance -a "${1}" -e coins)
        local balance
        balance=$(echo "${result}" | jq -r ".balance")
        if [ "$(echo "$balance > $2" | bc)" -eq 1 ]; then
            echo -e "${GRE}The balance is enough now, and is:${balance}${NOC}"
            return
        fi
        block_wait "${cli}" 1
    done
}

# chain33 区块等待 $1:cli 路径  $2:等待高度
function block_wait() {
    #    set +x
    local CLI=${1}

    if [[ $# -lt 1 ]]; then
        echo -e "${RED}wrong block_wait parameter${NOC}"
        exit_cp_file
    fi

    local cur_height
    local expect
    local count
    cur_height=$(${CLI} block last_header | jq ".height")
    expect=$((cur_height + ${2}))
    count=0
    while true; do
        new_height=$(${CLI} block last_header | jq ".height")
        if [[ ${new_height} -ge ${expect} ]]; then
            break
        fi

        count=$((count + 1))
        sleep 1
    done

    count=$((count + 1))
    #    set -x
    echo -e "${GRE}chain33 wait new block $count s, cur height=$expect,old=$cur_height${NOC}"
}

#创建账户，并充值
function setupAccount() {
    $cli account import_key -l player -k 0x7b2800cdecd978ab0e877f7e3734b9d0b11d864fa51d9b623d7bdbd76c16a40d

    balance_enough 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv 1010
    echo "transfer to 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
    $cli send coins transfer -a 1000 -n "t1000" -t 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
    echo 'transfer to 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU'
    balance_enough 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt 800
    $cli send coins transfer -a 800 -n "t100" -t 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
}

#部署合约
function deployJavaContract() {
    for contract in Guess Dice; do
        echo "Deploy contract for $contract"
        $cli send jvm create -x $contract -d . -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    done
}

function depositAndStartGame() {
    #    for contract in Guess Dice
    for contract in Guess Dice; do
        echo "transfer to user.jvm.$contract"
        $cli send coins send_exec -a 300 -e user.jvm.$contract -n send2exec -k 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU
        #如果不等一个块，有可能区块打包的时候，乱序，导致执行失败
        block_wait "${cli}" 1

        #开始游戏
        echo "send tx to startGame for user.jvm.$contract"
        $cli send jvm call -e $contract -x startGame -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
        block_wait "${cli}" 1

        #投注
        echo "send tx to playGame for user.jvm.$contract"
        $cli send jvm call -e $contract -x playGame -r "6 2" -k 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU
        echo "                 "
    done
    block_wait "${cli}" 12
}

function closeGame() {
    for contract in Guess Dice; do
        echo "close $contract"
        $cli send jvm call -e $contract -x closeGame -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    done
    block_wait "${cli}" 10
}

expectQueryRes[0]="guessNum=6,ticketNum=2"
expectQueryRes[1]="diceNum=6,ticketNum=2"

function queryGame() {
    i=0
    for contract in Guess Dice; do
        echo "query get${contract}RecordByRound"
        result=$($cli jvm query -e user.jvm.$contract -r "get${contract}RecordByRound 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU 1")
        if [ "${result}" == "${expectQueryRes[i]}" ]; then
            echo -e "${GRE}Succeed to do query from user.jvm.$contract, and get ${result}${NOC}"
        else
            echo -e "${RED}error query via get${contract}RecordByRound, expect ${expectQueryRes[i]}, get ${result}${NOC}"
            exit 1
        fi
        let i++

        echo "query getBonusByRound"
        $cli jvm query -e user.jvm.$contract -r "getBonusByRound 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU 1"

        echo "query getLuckNumByRound"
        $cli jvm query -e user.jvm.$contract -r "getLuckNumByRound 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU 1"
    done
}

function playGameInfinite() {
    #    for contract in Guess Dice
    while true; do
        #投注
        echo "send tx to playGame for user.jvm.Guess"
        $cli send jvm call -e Guess -x playGame -r "8 1" -k 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU
        echo "                 "

        echo "send tx to playGame for user.jvm.Dice"
        $cli send jvm call -e Dice -x playGame -r "3 1" -k 1PrTWtT1Bzhg2L8jjVKU7ohxHVXLU4NMEU
        echo "                 "
        block_wait "${cli}" 1
    done
}

function dice_game_test() {
  set -x
    setupAccount
    deployJavaContract
    depositAndStartGame
    closeGame
    queryGame
#    playGameInfinite
}

#dice_game_test
