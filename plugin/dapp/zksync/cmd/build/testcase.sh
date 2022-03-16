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
    chain33Addr=$(${CLI} zksync getChain33Addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    rawData=$(${CLI} zksync deposit -t 1 -a 100000000000000000000 -e abcd68033A72978C1084E2d44D1Fa06DdC4A2d57 -c "$chain33Addr")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 1
}

function zksync_setPubKey() {
    echo "=========== # zksync setPubKey test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 setPubKey
    rawData=$(${CLI} zksync setPubKey -a 1)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 1

}

function zksync_withdraw() {
    echo "=========== # zksync withdraw test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 withdraw amount 100000000
    rawData=$(${CLI} zksync withdraw -t 1 -a 100000000 --accountId 1)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 1
}

function zksync_treeToContract() {
    echo "=========== # zksync treeToContract test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 treeToContract amount 1000000000
    rawData=$(${CLI} zksync treeToContract -t 1 -a 1000000000000000000 --accountId 1)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 1
}

function zksync_contractToTree() {
    echo "=========== # zksync contractToTree test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 contractToTree to self amount 100000000
    chain33Addr=$(${CLI} zksync getChain33Addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    rawData=$(${CLI} zksync contractToTree -t 1 -a 100000000 --accountId 1)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 1

    #1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR deposit amount 1000000000
    chain33Addr=$(${CLI} zksync getChain33Addr -k 19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4)
    rawData=$(${CLI} zksync deposit -t 1 -a 1000000000 -e abcd68033A72978C1084E2d44D1Fa06DdC4A2d57 -c "$chain33Addr")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 2
}

function zksync_transfer() {
    echo "=========== # zksync transfer test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 transfer to 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR amount 100000000
    rawData=$(${CLI} zksync transfer -t 1 -a 100000000 --accountId 1 --toAccountId 2)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 1
    query_account "${CLI}" 2
}

function zksync_transferToNew() {
    echo "=========== # zksync transferToNew test ============="
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 transferToNew to 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k amount 100000000
    chain33Addr=$(${CLI} zksync getChain33Addr -k 7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    rawData=$(${CLI} zksync transferToNew -t 1 -a 100000000 --accountId 1 -e 12a0E25E62C1dBD32E505446062B26AECB65F028 -c "$chain33Addr")
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 3
}

function zksync_forceExit() {
    echo "=========== # zksync forceExit test ============="
    #1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR setPubKey
    rawData=$(${CLI} zksync setPubKey -a 2)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 2

    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 help 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR forceExit
    rawData=$(${CLI} zksync forceExit -t 1 --accountId 2)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 2
}

function zksync_fullExit() {
    echo "=========== # zksync fullExit test ============="
    #1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k setPubKey
    rawData=$(${CLI} zksync setPubKey -a 3)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 3

    #1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k fullExit
    rawData=$(${CLI} zksync fullExit -t 1 --accountId 3)
    echo "${rawData}"

    signData=$(${CLI} wallet sign -d "$rawData" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    query_account "${CLI}" 3
}

function zksync_setVerifyKey() {
    echo "=========== # zksync setVerifyKey test ============="
    #set verify key
    rawData=$(${CLI} zksync vkey -v e793f2c53d7c6880722c1902cc8c582c582843fb58acfa0cf16483ab6e324705a49d1a0e604ab22a0421b442aefcfa790e34902d0f7aa37676e6e71aff5e8ab0ec6396a99eaba0f3e6d41ebe5d5146871107461d14432d09d70fc7117e38f1992b81373cc779af2d826b2938cfafb65f5bf4068ec352aa3b5c5ce80752df9083d9a2c1781128d7c31ebd5c7a6c205987d2e19c77725e34870044907cd18d3a3c268be12a7805ee3622356f1f85f5bd7d140d1e2a0328155959c299aa6eef9a828dd3b14a6743281d1d3ba54b1590522f5f505cf3544ca49bf1ff28b456e6b2b5d0b10d6f9d4ae1a06e118e2f59db1cf4fd041581ace0d8bbcd2184e6587e001c1d304c69fa59bf583f91897a5bea5a3b059cbef7d554822bfbf6dac4cd64438100000003817edbbf30d409be1aaa7d13ecf6b90b3e981bc69fdae6dd61d55926ad04776aaf6bf1362a21ac4d30efed12a0657a91852d807fad7f34f15d634d03334ee8eb86cd48ae958b07063a449f55163c90791a85f9ec563b2e4fb7512089c2f87e10)
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
    ret=$(${1} zksync account -a "${2}")
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

    local accountId=4
    while true; do
         #loop deposit amount 1000000000000
         echo "=========== # zksync setVerifyKey test ============="
         privateKey=$(${CLI} account rand -l 1 | jq -r ".privateKey")
         echo "${privateKey}"
         chain33Addr=$(${CLI} zksync getChain33Addr -k "$privateKey")

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
    create_tx
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
