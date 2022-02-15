#!/usr/bin/env bash

#1ks returner chain31
ZKSYNC_CLI31="docker exec ${NODE1} /root/chain33-cli "
#1jr  authorize chain32
ZKSYNC_CLI32="docker exec ${NODE2} /root/chain33-cli "
#1nl receiver  chain30
ZKSYNC_CLI30="docker exec ${NODE4} /root/chain33-cli "

# shellcheck source=/dev/null
#source test-rpc.sh

function zksync_set_wallet() {
    echo "=========== # zksync set wallet ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    zksync_import_wallet "${ZKSYNC_CLI31}" "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" "account1"
    #1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
    zksync_import_wallet "${ZKSYNC_CLI32}" "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" "account2"
    #1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
    zksync_import_wallet "${ZKSYNC_CLI30}" "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" "account3"
}

function zksync_import_wallet() {
    local lable=$3
    echo "=========== # save seed to wallet ============="
    result=$(${1} seed save -p 1314fuzamei -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib" | jq ".isok")
    if [ "${result}" = "false" ]; then
        echo "save seed to wallet error seed, result: ${result}"
        exit 1
    fi

    echo "=========== # unlock wallet ============="
    result=$(${1} wallet unlock -p 1314fuzamei -t 0 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi

    echo "=========== # import private key ============="
    echo "key: ${2}"
    result=$(${1} account import_key -k "${2}" -l "$lable" | jq ".label")
    if [ -z "${result}" ]; then
        exit 1
    fi

    echo "=========== # wallet status ============="
    ${1} wallet status
}

function zksync_transfer() {
    echo "=========== # zksync chain transfer ============="

    #account1
    ${CLI} send coins transfer -a 200 -n transfer -t 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #account2
    ${CLI} send coins transfer -a 200 -n transfer -t 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #account3
    ${CLI} send coins transfer -a 200 -n transfer -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

}

function zksync_deposit() {
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 deposit amount 10000000000
    rawData=$(${CLI} send zksync deposit -t 1 -a 10000000000 -e abcd68033A72978C1084E2d44D1Fa06DdC4A2d57 -c 1ac911ce06f25973cb33f7dbadf5ba2eda083fa02962371a070069000a96e4eb)
    echo "${rawData}"

    query_proof "${MIX_CLI31}" 1

    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    query_proof "${MIX_CLI31}" 1
}

function query_proof() {
    block_wait "${1}" 1

    local times=200
    ret=$(${1} zksync proofs --height "${2}" | jq -r ".operationInfos[0]")
    echo "query proofs addr=${2}, return ${ret} "

}

function query_tx() {
    block_wait "${1}" 1

    local times=200
    while true; do
        ret=$(${1} tx query -s "${2}" | jq -r ".tx.hash")
        echo "query hash is ${2}, return ${ret} "
        if [ "${ret}" != "${2}" ]; then
            block_wait "${1}" 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "query tx=$2 failed"
                exit 1
            fi
        else
            echo "query tx=$2  success"
            break
        fi
    done
}

function zksync_test() {
    echo "=========== # mix chain test ============="
    zksync_deposit
}

function zksync() {
    if [ "${2}" == "init" ]; then
        echo "zksync init"
    elif [ "${2}" == "config" ]; then
        zksync_set_wallet
        zksync_transfer
    elif [ "${2}" == "test" ]; then
        zksync_test "${1}"
    fi

}
