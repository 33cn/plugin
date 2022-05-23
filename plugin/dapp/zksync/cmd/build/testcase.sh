#!/usr/bin/env bash

#1ks returner chain31
ZKSYNC_CLI31="docker exec ${NODE1} /root/chain33-cli "
#1jr  authorize chain32
ZKSYNC_CLI32="docker exec ${NODE2} /root/chain33-cli "
#1nl receiver  chain30
ZKSYNC_CLI30="docker exec ${NODE4} /root/chain33-cli "

ZKSYNC_ACCOUNT_3="3"
ZKSYNC_ACCOUNT_4="4"
TOKENID_0="0"
TOKENID_1="1"
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

function zksync_init() {
    echo "=========== # zksync chain init ============="

    #account1
    ${CLI} send coins transfer -a 200 -n transfer -t 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #account2
    ${CLI} send coins transfer -a 200 -n transfer -t 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    #account3
    ${CLI} send coins transfer -a 200 -n transfer -t 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

}

function zksync_deposit() {
    echo "=========== # zksync deposit test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 deposit amount 1000000000000
    chain33Addr=$(${CLI} zksync l2addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    hash=$(${CLI} send zksync deposit -t "${TOKENID_0}" -a 1000000000000000000000 -e 12a0E25E62C1dBD32E505446062B26AECB65F028 -c "$chain33Addr" -i 0 -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_3}"
}

function zksync_setPubKey() {
    echo "=========== # zksync setPubKey test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 setPubKey
    rawData=$(${CLI} zksync pubkey -a "${ZKSYNC_ACCOUNT_3}")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_3}"

}

function zksync_withdraw() {
    echo "=========== # zksync withdraw test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 withdraw amount 100000000
    rawData=$(${CLI} zksync withdraw -t "${TOKENID_0}" -a 100000000 -i "${ZKSYNC_ACCOUNT_3}")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_3}"
}

function zksync_treeToContract() {
    echo "=========== # zksync treeToContract test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 treeToContract amount 1000000000
    rawData=$(${CLI} zksync tree2contract -t "${TOKENID_0}" -a 10000000000000000000 --accountId "${ZKSYNC_ACCOUNT_3}")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_3}"
}

function zksync_contractToTree() {
    echo "=========== # zksync contractToTree test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 contractToTree to self amount 100000000
    chain33Addr=$(${CLI} zksync l2addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    rawData=$(${CLI} zksync contract2tree -t "${TOKENID_0}" -a 1000000000000000000 --accountId "${ZKSYNC_ACCOUNT_3}")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_3}"

    #1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR deposit amount 1000000000
    chain33Addr=$(${CLI} zksync l2addr -k 19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4)
    rawData=$(${CLI} zksync deposit -t "${TOKENID_1}" -a 1000000000 -e abcd68033A72978C1084E2d44D1Fa06DdC4A2d57 -c "$chain33Addr" -i 1)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_4}"
}

function zksync_transfer() {
    echo "=========== # zksync transfer test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 transfer to 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR amount 100000000
    rawData=$(${CLI} zksync transfer -t "${TOKENID_0}" -a 100000000 --accountId "${ZKSYNC_ACCOUNT_3}" --toAccountId "${ZKSYNC_ACCOUNT_4}")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_3}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_4}"
}

function zksync_transferToNew() {
    echo "=========== # zksync transferToNew test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 transferToNew to 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k amount 100000000
    chain33Addr=$(${CLI} zksync l2addr -k 7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    rawData=$(${CLI} zksync transfer2new -t "${TOKENID_0}" -a 100000000 --accountId "${ZKSYNC_ACCOUNT_3}" -e 12a0E25E62C1dBD32E505446062B26AECB65F027 -c "$chain33Addr")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 4
}

function zksync_forceExit() {
    echo "=========== # zksync forceExit test ============="
    #1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR setPubKey
    rawData=$(${CLI} zksync pubkey -a "${ZKSYNC_ACCOUNT_4}")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_4}"

    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 help 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR forceExit
    rawData=$(${CLI} zksync forceexit -t "${TOKENID_0}" -a "${ZKSYNC_ACCOUNT_4}")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" "${ZKSYNC_ACCOUNT_4}"
}

function zksync_fullExit() {
    echo "=========== # zksync fullExit test ============="
    #1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k setPubKey
    rawData=$(${CLI} zksync pubkey -a 4)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 4

    #1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k fullExit
    rawData=$(${CLI} zksync fullexit -t "${TOKENID_0}" -a 4 -i 2)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 4
}

function zksync_setVerifyKey() {
    echo "=========== # zksync setVerifyKey test ============="
    #set verify key
    rawData=$(${CLI} zksync vkey -v 9bacd15739de2797c5712ba1bb4b04770c792bc4f0b07ba413a3be104c730d0999f4db226287bede62c82c2013cb4998a812081a6953dfb0ce5d61702e89a75cca66711a9deb2bea5d86a1f5ca9ec3aa59fe6d4754ce9335f4719dd3d3549acc2773cf1a9af35365661ffd9230c6f9686463c0d9012db3e261f539a68b3dba4edbdcc88d1567c5ace42099bcb784bf4d95ce329b819ab9abf5ae868cf43f3ad6026f98d38ea6f20a4dffab047049537100448622c1a37b324285a5d2fe5e9b60eb41673c17eb70f8a2d83a9057ebc7864afcc3a059f2f86630c62536d1c98587cf206c7a2ebd29849ac4932243f5cbb3c218d7c4f515dcfe165a2298a55bce26029e64c87c78557f5c22a74577043a695e4576055c7ca65c96b9b7265740c6ab000000039310846b7cd5a7bd2da482f3ec8d72ed029102686ef2dc6781c5979e7b4338c9ad5cb48c1ef34c50b016772988e95d6508aad5c4f2e9b5d3e53acd854d4fc242df0aba967c2dc815e1af26e27a9a64913b1192e9dd17fcf4817c08a807ca7907)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
}

function query_account() {
    block_wait "${1}" 1

    local times=200
    ret=$(${1} zksync query account id -a "${2}")
    echo "query account accountId=${2}, return ${ret} "

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

function create_tx() {
    block_wait "${CLI}" 10

    local accountId=5
    while true; do
        #loop deposit amount 1000000000000
        echo "=========== # zksync add new account test ============="
        privateKey=$(${CLI} account rand -l 1 | jq -r ".privateKey")
        echo "${privateKey}"
        chain33Addr=$(${CLI} zksync l2addr -k "$privateKey")

        rawData=$(${CLI} zksync deposit -t 1 -a 1000000000000 -e abcd68033A72978C1084E2d44D1Fa06DdC4A2d57 -c "$chain33Addr")
        echo "${rawData}"

        signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
        echo "${signData}"
        hash=$(${CLI} wallet send -d "$signData")
        echo "${hash}"
        query_tx "${CLI}" "${hash}"
        query_account "${CLI}" $accountId
        accountId=$((accountId + 1))
    done
}

function zksync_test() {
    echo "=========== # zksync chain test ============="
    zksync_deposit
    zksync_setPubKey
    zksync_withdraw
    zksync_treeToContract
    zksync_contractToTree
    zksync_transfer
    zksync_transferToNew
    zksync_forceExit
    zksync_fullExit
    zksync_setVerifyKey
    #    create_tx
}

function zksync() {
    if [ "${2}" == "init" ]; then
        echo "zksync init"
    elif [ "${2}" == "config" ]; then
        zksync_set_wallet
        zksync_init
    elif [ "${2}" == "test" ]; then
        zksync_test "${1}"
    fi

}
