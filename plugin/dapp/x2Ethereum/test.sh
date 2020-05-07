#!/usr/bin/env bash
set -x
CLI="/opt/src/github.com/33cn/plugin/build/chain33-cli"
EBCLI="/opt/src/github.com/33cn/plugin/plugin/dapp/x2Ethereum/build/ebcli_A"

ETHContractAddr="0x0000000000000000000000000000000000000000"
BTYContractAddr="0x40BFE5eD039A9a2Eb42ece2E2CA431bFa7Cf4c42"
Ethsender="0xcbfddc6ae318970ba3feeb0541624f95822e413a"
BtyReceiever="1BqP2vHkYNjSgdnTqm7pGbnphLhtEhuJFi"
BtyTokenContractAddr="0xbAf2646b8DaD8776fc74Bf4C8d59E6fB3720eddf"

Validator1="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
Validator2="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
Validator3="1BqP2vHkYNjSgdnTqm7pGbnphLhtEhuJFi"

# 测试流程
#
# bankAddr = 1BqP2vHkYNjSgdnTqm7pGbnphLhtEhuJFi
#
#
# 1.Exec_EthBridgeClaim, 测试chain33这端
#
#
#

block_wait() {
    if [[ $# -lt 1 ]]; then
        echo "wrong #block_wait parameter"
        exit 1
    fi
    cur_height=$(${CLI} block last_header | jq ".height")
    expect=$((cur_height + ${1}))
    local count=0
    while true; do
        new_height=$(${CLI} block last_header | jq ".height")
        if [[ ${new_height} -ge ${expect} ]]; then
            break
        fi
        count=$((count + 1))
        sleep 1
    done
    sleep 1
    count=$((count + 1))
    echo "wait new block $count s, cur height=$expect,old=$cur_height"
}

# check_tx(ty, hash)
# ty：交易执行结果
#   1：交易打包
#   2：交易执行成功
# hash：交易hash
check_tx() {
    if [[ $# -lt 2 ]]; then
        echo "wrong check_tx parameters"
        exit 1
    fi
    ty=$(${CLI} tx query -s ${2} | jq .receipt.ty)
    if [[ ${ty} != ${1} ]]; then
        echo "check tx error, hash is ${2}"
        exit 1
    fi
}

check_Number() {
    if [[ $# -lt 2 ]]; then
        echo "wrong check_Number parameters"
        exit 1
    fi
    if [[ ${1} != ${2} ]]; then
        echo "error Number, expect ${1}, get ${2}"
        exit 1
    fi
}

# SetConsensusThreshold
hash=$(${CLI} send x2ethereum setconsensus -p 80 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 2
check_tx 2 ${hash}

# query ConsensusThreshold
nowConsensus=$(${CLI} send x2ethereum query consensus -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .nowConsensusThreshold | sed 's/\"//g')
check_Number 80 ${nowConsensus}

# add a validator
hash=$(${CLI} send x2ethereum add -a ${Validator3} -p 7 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 2
check_tx 2 ${hash}

# query Validators
validators=$(${CLI} send x2ethereum query validators -v ${Validator3} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .validators | sed 's/\"//g')
totalPower=$(${CLI} send x2ethereum query validators -v ${Validator3} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .totalPower | sed 's/\"//g')
check_Number 7 ${totalPower}

# add a validator again
hash=$(${CLI} send x2ethereum add -a ${Validator2} -p 6 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 2
check_tx 2 ${hash}

# query Validators
validators=$(${CLI} send x2ethereum query validators -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .validators | sed 's/\"//g')
totalPower=$(${CLI} send x2ethereum query validators -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .totalPower | sed 's/\"//g')
check_Number 13 ${totalPower}

# remove a validator
hash=$(${CLI} send x2ethereum remove -a ${Validator2} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 2
check_tx 2 ${hash}

# query Validators
validators=$(${CLI} send x2ethereum query validators -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .validators | sed 's/\"//g')
totalPower=$(${CLI} send x2ethereum query validators -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .totalPower | sed 's/\"//g')
check_Number 7 ${totalPower}

# add a validator again
hash=$(${CLI} send x2ethereum add -a ${Validator2} -p 6 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 2
check_tx 2 ${hash}

# query Validators
validators=$(${CLI} send x2ethereum query validators -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .validators | sed 's/\"//g')
totalPower=$(${CLI} send x2ethereum query validators -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .totalPower | sed 's/\"//g')
check_Number 13 ${totalPower}

# add a validator
hash=$(${CLI} send x2ethereum add -a ${Validator1} -p 87 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 2
check_tx 2 ${hash}

# query Validators
validators=$(${CLI} send x2ethereum query validators -v 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .validators | sed 's/\"//g')
totalPower=$(${CLI} send x2ethereum query validators -v 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .totalPower | sed 's/\"//g')
check_Number 6 ${totalPower}

totalPower=$(${CLI} send x2ethereum query totalpower -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .totalPower | sed 's/\"//g')
check_Number 100 ${totalPower}

##############################################################
######################## 测试交易 #############################
##############################################################

# chain33 -> ethereum

# send bty to mavl-x2ethereum-bty-12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
${CLI} send coins send_exec -e x2ethereum -a 200 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv

# send a chain33eth tx，在chain33侧lock 5个bty
hash=$(${CLI} send x2ethereum lock --amount 5 -e x2ethereum -t bty -r ${Ethsender} -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -q ${BtyTokenContractAddr} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 4
check_tx 2 ${hash}

#在以太坊侧burn掉1个bty
hash=$(${EBCLI} relayer ethereum burn -m 1 -t ${BtyTokenContractAddr} -r ${BtyReceiever} -k 772db14fc5ae829b155e1eda09431a0b566833f2de3b50b2d35625542b3ae52f | jq .msg | sed 's/\"//g')

# ethereum -> chain33，在以太坊侧lock 0.1 eth
hash=$(${EBCLI} relayer ethereum lock -m 0.1 -r ${BtyReceiever} -k 772db14fc5ae829b155e1eda09431a0b566833f2de3b50b2d35625542b3ae52f)

#在chain33 burn 0.1eth
hash=$(${CLI} send x2ethereum burn --amount 0.1 -e x2ethereum -t eth -r ${Ethsender} -s ${BtyReceiever} -q ${ETHContractAddr} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 4
check_tx 2 ${hash}

# ethereum -> chain33
hash=$(${EBCLI} relayer ethereum lock -m 0.1 -t ${BtyTokenContractAddr} -r ${BtyReceiever} -k 772db14fc5ae829b155e1eda09431a0b566833f2de3b50b2d35625542b3ae52f)

hash=$(${CLI} send x2ethereum burn --amount 0.1 -t ${BtyTokenContractAddr} -e x2ethereum -t eth -r ${Ethsender} -s ${BtyReceiever} -q ${ETHContractAddr} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
#block_wait 4
check_tx 2 ${hash}
