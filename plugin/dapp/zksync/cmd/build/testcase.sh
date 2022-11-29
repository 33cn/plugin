#!/usr/bin/env bash
# shellcheck source=/dev/null
# shellcheck disable=SC2128
# shellcheck disable=SC2034

set -x
set +e

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

#1ks returner chain31
ZKSYNC_CLI31="docker exec ${NODE1} /root/chain33-cli "
#1jr  authorize chain32
ZKSYNC_CLI32="docker exec ${NODE2} /root/chain33-cli "
#1nl receiver  chain30
ZKSYNC_CLI30="docker exec ${NODE4} /root/chain33-cli "

TOKENID_0="0"
TOKENID_1="1"
TOKENID_SYMBOL_0="ETH"
TOKENID_SYMBOL_1="USDT"
queueId=0

management_key="4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"

ZKSYNC_ACCOUNT_4="4"
account1="1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
accountKey1="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
ZKSYNC_ACCOUNT_5="5"
account2="1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
accountKey2="0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
ZKSYNC_ACCOUNT_6="6"
account3="1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
accountKey3="0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"

# fee
# token id 0 ETH le18
withdrawFee=200000000000000
transferFee=100000000000000
proxyExitFee=200000000000000
tree2contractFee=100000000000000
contract2treeFee=10000
contractChain33Fee=0.0001
# token id 1 le8
withdrawFee1=10000
transferFee1=10000
proxyExitFee1=10000
tree2contractFee1=10000
contract2treeFee1=10000

# 判断结果 $1 和 $2 是否相等
function is_equal() {
    set +x
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit 1
    fi

    if [[ $1 != "$2" ]]; then
        echo -e "${RED}$1 != ${2}${NOC}"
        exit 1
    fi

    set -x
}

# 检查交易是否执行成功 $1:cli 路径  $2:交易hash
function check_tx() {
    set +x
    local CLI=${1}

    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong check_tx parameters${NOC}"
        exit 1
    fi

    if [[ ${2} == "" ]]; then
        echo -e "${RED}wrong check_tx txHash is empty${NOC}"
        exit 1
    fi

    local count=0
    while true; do
        ty=$(${CLI} tx query -s "${2}" | jq .receipt.ty)
        if [[ ${ty} != "" ]]; then
            break
        fi

        count=$((count + 1))
        sleep 1

        if [[ ${count} -ge 100 ]]; then
            echo "chain33 query tx for too long"
            break
        fi
    done

    set -x

    ty=$(${CLI} tx query -s "${2}" | jq .receipt.ty)
    if [[ ${ty} != 2 ]]; then
        echo -e "${RED}check tx error, hash is ${2}${NOC}"
        exit 1
    fi
}

function zksync_set_wallet() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    zksync_import_wallet "${ZKSYNC_CLI31}" "${accountKey1}" "account1"
    #1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
    zksync_import_wallet "${ZKSYNC_CLI32}" "${accountKey2}" "account2"
    #1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
    zksync_import_wallet "${ZKSYNC_CLI30}" "${accountKey3}" "account3"
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
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    #account1
    ${CLI} send coins transfer -a 200 -n transfer -t "${account1}" -k "${management_key}"
    #account2
    ${CLI} send coins transfer -a 200 -n transfer -t "${account2}" -k "${management_key}"
    #account3
    ${CLI} send coins transfer -a 200 -n transfer -t "${account3}" -k "${management_key}"
}

function zksync_deposit() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 deposit amount 1000000000000
    chain33Addr=$(${CLI} zksync l2 l2addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    hash=$(${CLI} send zksync l2 deposit -t "${TOKENID_0}" -a 8000000000000000000 -e 12a0E25E62C1dBD32E505446062B26AECB65F028 -c "$chain33Addr" -i ${queueId} -k "${management_key}")
    check_tx "${CLI}" "${hash}"
    queueId=$((queueId + 1))
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_4}" 8000000000000000000

    #1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR deposit amount 1000000000
    chain33Addr=$(${CLI} zksync l2 l2addr -k 19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4)
    hash=$(${CLI} send zksync l2 deposit -t "${TOKENID_1}" -a 6000000000000000000 -e abcd68033A72978C1084E2d44D1Fa06DdC4A2d57 -c "$chain33Addr" -i ${queueId} -k "${management_key}")
    check_tx "${CLI}" "${hash}"
    queueId=$((queueId + 1))
    query_account_balance "${TOKENID_1}" "${ZKSYNC_ACCOUNT_5}" 6000000000000000000
}

function zksync_setPubKey() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 setPubKey
    hash=$(${CLI} send zksync l2 setpubkey -a "${ZKSYNC_ACCOUNT_4}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"
    #    hash=$(${CLI} send zksync l2 setpubkey -a "${ZKSYNC_ACCOUNT_5}" -k "${accountKey2}")
    #    check_tx "${CLI}" "${hash}"
}

function zksync_withdraw() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf=$(${CLI} zksync query account balance id -a "${ZKSYNC_ACCOUNT_4}" -t "${TOKENID_0}" | jq ".tokenBalances[0].balance" | sed 's/\"//g')
    hash=$(${CLI} send zksync l2 withdraw -t "${TOKENID_0}" -a 100000000000000000 -i "${ZKSYNC_ACCOUNT_4}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"
    balanceAf=$((balanceBf - 100000000000000000 - withdrawFee))
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_4}" ${balanceAf}
}

function zksync_set_symbol() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    hash=$(${CLI} send zksync l2 symbol -s "${TOKENID_SYMBOL_0}" -t "${TOKENID_0}" -d 18 -k "${management_key}")
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 symbol -s "${TOKENID_SYMBOL_1}" -t "${TOKENID_1}" -d 8 -k "${management_key}")
    check_tx "${CLI}" "${hash}"
}

function zksync_treeToContract() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf=$(${CLI} zksync query account balance id -a "${ZKSYNC_ACCOUNT_4}" -t "${TOKENID_0}" | jq ".tokenBalances[0].balance" | sed 's/\"//g')
    ${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}"

    hash=$(${CLI} send zksync l2 tree2contract -t "${TOKENID_0}" -a 2000000000000000000 --accountId "${ZKSYNC_ACCOUNT_4}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"
    balanceAf=$((balanceBf - 2000000000000000000 - tree2contractFee))
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_4}" ${balanceAf}

    balance=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    is_equal "${balance}" "2.0000"
}

function zksync_contractToTree() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBfAsset=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    balanceBf=$(${CLI} zksync query account balance id -a "${ZKSYNC_ACCOUNT_4}" -t "${TOKENID_0}" | jq ".tokenBalances[0].balance" | sed 's/\"//g')
    hash=$(${CLI} send zksync l2 contract2tree -t "${TOKENID_SYMBOL_0}" -a 10000000 --accountId "${ZKSYNC_ACCOUNT_4}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"

    balanceAf=$((balanceBf + 100000000000000000))
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_4}" ${balanceAf}
    balance=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    balanceAf=$(echo "${balanceBfAsset} - 0.1 - ${contractChain33Fee}" | bc)
    is_equal "${balance}" "${balanceAf}"
}

function zksync_transfer() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 transfer to 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR amount 100000000

    balanceBf5=$(${CLI} zksync query account balance id -a "${ZKSYNC_ACCOUNT_4}" -t "${TOKENID_0}" | jq ".tokenBalances[0].balance" | sed 's/\"//g')
    balanceBf6=$(${CLI} zksync query account balance id -a "${ZKSYNC_ACCOUNT_5}" -t "${TOKENID_0}" | jq ".tokenBalances[0].balance" | sed 's/\"//g')

    hash=$(${CLI} send zksync l2 transfer -i "${TOKENID_0}" -a 100000000000000000 -f "${ZKSYNC_ACCOUNT_4}" -t "${ZKSYNC_ACCOUNT_5}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"

    balanceAf5=$((balanceBf5 - 100000000000000000 - transferFee))
    balanceAf6=$((balanceBf6 + 100000000000000000))
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_4}" ${balanceAf5}
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_5}" ${balanceAf6}
}

function zksync_transferToNew() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 transferToNew to 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k amount 100000000

    balanceBf=$(${CLI} zksync query account balance id -a "${ZKSYNC_ACCOUNT_4}" -t "${TOKENID_0}" | jq ".tokenBalances[0].balance" | sed 's/\"//g')
    chain33Addr=$(${CLI} zksync l2 l2addr -k 7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    hash=$(${CLI} send zksync l2 transfer2new -t "${TOKENID_0}" -a 100000000000000000 -f "${ZKSYNC_ACCOUNT_4}" -e 12a0E25E62C1dBD32E505446062B26AECB65F027 -c "$chain33Addr" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"

    balanceAf=$((balanceBf - 100000000000000000 - transferFee))
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_4}" ${balanceAf}
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_6}" 100000000000000000
}

function zksync_asset_transfer() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    hash=$(${CLI} send zksync l2 tree2contract -t "${TOKENID_0}" -a 3000000000000000000 --accountId "${ZKSYNC_ACCOUNT_4}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"

    # transfer
    balanceBf=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    hash=$(${CLI} send zksync asset transfer -a 1.1 -s "${TOKENID_SYMBOL_0}" -t "${account2}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"
    balanceAf=$(echo "${balanceBf} - 1.1" | bc)
    balance=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    is_equal "${balance}" "${balanceAf}"
    balance=$(${CLI} asset balance -a "${account2}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    is_equal "${balance}" "1.1000"

    # transfer_exec
    balanceBf=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    hash=$(${CLI} send zksync asset transfer_exec -a 1.2 -e paracross -s "${TOKENID_SYMBOL_0}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"
    balanceAf=$(echo "${balanceBf} - 1.2" | bc)
    balance=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    is_equal "${balance}" "${balanceAf}"
    balance=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" -e paracross | jq ".balance" | sed 's/\"//g')
    is_equal "${balance}" "1.2000"

    # withdraw
    balanceBf=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    balanceBfP=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" -e paracross | jq ".balance" | sed 's/\"//g')
    hash=$(${CLI} send zksync asset withdraw -a 0.1 -e paracross -s "${TOKENID_SYMBOL_0}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"
    balanceAf=$(echo "${balanceBf} + 0.1" | bc)
    balance=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" | jq ".balance" | sed 's/\"//g')
    is_equal "${balance}" "${balanceAf}"

    balanceAf=$(echo "${balanceBfP} - 0.1" | bc)
    balance=$(${CLI} asset balance -a "${account1}" --asset_exec zksync --asset_symbol "${TOKENID_SYMBOL_0}" -e paracross | jq ".balance" | sed 's/\"//g')
    is_equal "${balance}" "${balanceAf}"
}

function zksync_proxyExit() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    hash=$(${CLI} send zksync l2 proxyexit -a "${ZKSYNC_ACCOUNT_4}" -i "${TOKENID_0}" -t "${ZKSYNC_ACCOUNT_5}" -k "${accountKey1}")
    check_tx "${CLI}" "${hash}"

    balance=$(${CLI} zksync query account balance id -a "${ZKSYNC_ACCOUNT_5}" -t "${TOKENID_0}" | jq ".tokenBalances[0].balance" | sed 's/\"//g')
    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_5}" 0
}

function zksync_fullExit() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    #1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k setPubKey 不支持
    hash=$(${CLI} send zksync l2 setpubkey -a "${ZKSYNC_ACCOUNT_6}" -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    check_tx "${CLI}" "${hash}"

    hash=$(${CLI} send zksync l2 fullexit -t "${TOKENID_0}" -a "${ZKSYNC_ACCOUNT_6}" -i ${queueId} -k "${management_key}")
    check_tx "${CLI}" "${hash}"
    queueId=$((queueId + 1))

    query_account_balance "${TOKENID_0}" "${ZKSYNC_ACCOUNT_6}" 0
}

function zksync_setVerifyKey() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    #set verify key
    verify_key="9bacd15739de2797c5712ba1bb4b04770c792bc4f0b07ba413a3be104c730d0999f4db226287bede62c82c2013cb4998a812081a6953dfb0ce5d61702e89a75cca66711a9deb2bea5d86a1f5ca9ec3aa59fe6d4754ce9335f4719dd3d3549acc2773cf1a9af35365661ffd9230c6f9686463c0d9012db3e261f539a68b3dba4edbdcc88d1567c5ace42099bcb784bf4d95ce329b819ab9abf5ae868cf43f3ad6026f98d38ea6f20a4dffab047049537100448622c1a37b324285a5d2fe5e9b60eb41673c17eb70f8a2d83a9057ebc7864afcc3a059f2f86630c62536d1c98587cf206c7a2ebd29849ac4932243f5cbb3c218d7c4f515dcfe165a2298a55bce26029e64c87c78557f5c22a74577043a695e4576055c7ca65c96b9b7265740c6ab000000039310846b7cd5a7bd2da482f3ec8d72ed029102686ef2dc6781c5979e7b4338c9ad5cb48c1ef34c50b016772988e95d6508aad5c4f2e9b5d3e53acd854d4fc242df0aba967c2dc815e1af26e27a9a64913b1192e9dd17fcf4817c08a807ca7907"
    hash=$(${CLI} send zksync l2 vkey -v "${verify_key}" -k "${management_key}")
    check_tx "${CLI}" "${hash}"
}

function query_account_balance() {
    balance=$(${CLI} zksync query account balance id -a "${2}" -t "${1}" | jq ".tokenBalances[0].balance" | sed 's/\"//g')
    if [[ $balance != "$3" ]]; then
        echo -e "${RED}$balance != ${3}${NOC}"
        exit 1
    fi
}

function create_tx() {
    local accountId=8
    while true; do
        #loop deposit amount 1000000000000
        echo "=========== # zksync add new account test ============="
        privateKey=$(${CLI} account rand -l 1 | jq -r ".privateKey")
        chain33Addr=$(${CLI} zksync l2 l2addr -k "$privateKey")

        hash=$(${CLI} send zksync l2 deposit -t "${TOKENID_0}" -a 1000000000000 -e abcd68033A72978C1084E2d44D1Fa06DdC4A2d57 -c "$chain33Addr" -i ${queueId} -k "${management_key}")
        check_tx "${CLI}" "${hash}"
        query_account_balance "${TOKENID_0}" $accountId 1000000000000
        accountId=$((accountId + 1))
        queueId=$((queueId + 1))
    done
}

function zkrelayer_set_fee() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    #withdraw:2,transfer:3,transfer2new:4,proxyExit:5,contract2tree:9,tree2contract:10
    hash=$(${CLI} send zksync l2 fee -a 2 -t 0 -f "${withdrawFee}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 3 -t 0 -f "${transferFee}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 4 -t 0 -f "${transferFee}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 5 -t 0 -f "${proxyExitFee}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 9 -t 0 -f "${contract2treeFee}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 10 -t 0 -f "${tree2contractFee}" -k ${management_key})
    check_tx "${CLI}" "${hash}"

    hash=$(${CLI} send zksync l2 fee -a 2 -t 1 -f "${withdrawFee1}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 3 -t 1 -f "${transferFee1}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 4 -t 1 -f "${transferFee1}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 5 -t 1 -f "${proxyExitFee1}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 9 -t 1 -f "${contract2treeFee1}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
    hash=$(${CLI} send zksync l2 fee -a 10 -t 1 -f "${tree2contractFee1}" -k ${management_key})
    check_tx "${CLI}" "${hash}"
}

function zksync_test() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    zksync_deposit
    zksync_set_symbol
    zkrelayer_set_fee
    zksync_setPubKey
    zksync_withdraw
    zksync_treeToContract
    zksync_contractToTree
    zksync_transfer
    zksync_transferToNew
    zksync_asset_transfer
    zksync_proxyExit
    zksync_setVerifyKey
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
