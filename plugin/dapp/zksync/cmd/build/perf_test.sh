#!/usr/bin/env bash
set -x
set -e

RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

CLI="docker exec build_chain33_1 /root/chain33-cli"
managerPrivkey="4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" #对应的chain33地址为: 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv

function GetChain33Addr() {
    chain33Addr1=$(${CLI} zksync l2addr -k $1)
    echo ${chain33Addr1}
}

function block_wait() {
    if [ "$#" -lt 2 ]; then
        echo "wrong block_wait params"
        exit 1
    fi
    cur_height=$(${1} block last_header | jq ".height")
    expect=$((cur_height + ${2}))
    local count=0
    while true; do
        new_height=$(${1} block last_header | jq ".height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 0.1
    done
    echo "wait new block $count/10 s, cur height=$expect,old=$cur_height"
}

function query_tx() {
    set +x
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
    set -x
}

function query_account() {
    block_wait "${1}" 1

    local times=200
    ret=$(${1} zksync query account id -a "${2}")
    echo "query account accountId=${2}, return ${ret} "

}

function signAndSend() {
    local rawData=$1
    local privkey=$2

    signData=$(${CLI} wallet sign -d "$rawData" -k "$privkey")
#    echo "${signData}"
    hash=$(${CLI} wallet send -d "$signData")
    echo "${hash}"
}

# zksync_deposit tokenId amount ethAddr chain33Addr queueId
function zksync_deposit() {
#    echo -e "${RED}wrong parameter${NOC}"
    echo -e "${GRE}=========== # zksync deposit test =============${NOC}"
    local tid=$1
    local amount=$2
    local ethAddr=$3
    local chain33Addr=$4
    local queueId=$5

    local hash=$(${CLI} send zksync deposit -t ${tid} -a ${amount} -e ${ethAddr} -c ${chain33Addr} -i ${queueId} -k ${managerPrivkey})
    query_tx "${CLI}" "${hash}" # bug ErrExecPanic
#    query_account "${CLI}" 1
}

function zksync_deposit_oneShoot() {
    echo -e "${GRE}=========== # zksync deposit test =============${NOC}"
    local tid=$1
    local amount=$2
    local ethAddr=$3
    local chain33Addr=$4
    local queueId=$5

    hash=$(${CLI} send zksync deposit -t ${tid} -a ${amount} -e ${ethAddr} -c ${chain33Addr} -i ${queueId} -k ${managerPrivkey})
    echo "${hash}"
}

# $op: buy/sell $left:$right $price $amount, trader-acc by: $id, $eth, $privkey
function zksync_limitorder() {
  echo "=========== # zksync limitorder test ============="
    local operator=$1
    local left=$2
    local right=$3
    local price=$4
    local amount=$5
    local accID=$6
    local accEth=$7
    local privkey=$8
 
    local hash=$(${CLI} send zksync zkLimitOrder -o ${operator} \
         -l ${left} -r ${right}  -p ${price} -a ${amount}  \
         --accountId ${accID} --ethAddress ${accEth} -k ${privkey})
    query_tx "${CLI}" "${hash}"
}

# query asset by  $acc-id $token-id
function zksync_account2token() {
   echo "=========== # zksync account2token test: id=$1============="
    local accID=$1
    local tid=$2

    ${CLI} zksync token -a ${accID} --token ${tid}
}

function zksync_setPubKey() {
    local accid=$1
    local acckey=$2
    echo "=========== # zksync setPubKey test ============="
    hash=$(${CLI} send zksync pubkey -a "${accid}" -k "${acckey}")
    query_tx "${CLI}" "${hash}"
}

#zksync_transfer2new -t tokenId -a amount -i accountId -e ethAddress -c chain33Addr
#zksync_transfer2new tokenId amount accountId ethAddress chain33Addr privkey
function zksync_transfer2new() {
    local tokenId=$1
    local amount=$2
    local accountId=$3
    local ethAddress=$4
    local chain33Addr=$5
    local privkey=$6
    echo "=========== # zksync transfer2new test ============="
    hash=$(${CLI} send zksync transfer2new -t "${tokenId}" -a "${amount}" -i "${accountId}" -e "${ethAddress}" -c "${chain33Addr}" -k "${privkey}")
    query_tx "${CLI}" "${hash}"
}

#zksync_transfer tokenId amount fromAccountId toAccountId privkey
function zksync_transfer() {
    local tokenId=$1
    local amount=$2
    local fromAccountId=$3
    local toAccountId=$4
    local privkey=$5
<<<<<<< HEAD
    echo -e "${GRE}=========== # zksync transfer test =============${NOC}"
    rawData=$(${CLI} zksync transfer -t "${tokenId}" -a "${amount}" -f "${fromAccountId}" -o "${toAccountId}")
#    echo "${rawData}"

    signAndSend ${rawData} ${privkey}
=======
    echo "=========== # zksync transfer test ============="
    hash=$(${CLI} send zksync transfer -t "${tokenId}" -a "${amount}" -f "${fromAccountId}" -o "${toAccountId}" -k "${privkey}")
    query_tx "${CLI}" "${hash}"
>>>>>>> zksync-opt-testscript-0714
}

function print_raw_data_zksync_transfer() {
    local tokenId=$1
    local amount=$2
    local fromAccountId=$3
    local toAccountId=$4
    local privkey=$5
    echo -e "${GRE}=========== # zksync transfer test =============${NOC}"
    rawData=$(${CLI} zksync transfer -t "${tokenId}" -a "${amount}" -f "${fromAccountId}" -o "${toAccountId}")
#    echo "${rawData}"
    signData=$(${CLI} wallet sign -d "$rawData" -k "$privkey")
    echo "${signData}"
}

function zksync_transfer_oneshoot() {
    local tokenId=$1
    local amount=$2
    local fromAccountId=$3
    local toAccountId=$4
    local privkey=$5
    echo -e "${GRE}=========== # zksync transfer test =============${NOC}"
    hash=$(${CLI} send zksync transfer -t "${tokenId}" -a "${amount}" -f "${fromAccountId}" -o "${toAccountId}" -k ${privkey})
    echo "${hash}"
}

#zksync_forcexit tokenId amount privkey
function zksync_forcexit() {
    local tokenId=$1
    local amount=$2
    local privkey=$3
    echo "=========== # zksync forceexit test ============="
    hash=$(${CLI} send zksync forceexit -t "${tokenId}" -a "${amount}" -k "${privkey}")
    query_tx "${CLI}" "${hash}"
}

#zksync_mint_nft creatorId recipientId contentHash protocol amount privkey
function zksync_mint_nft() {
    local creatorId=$1
    local recipientId=$2
    local contentHash=$3
    local protocol=$4
    local amount=$5
    local privkey=$6
    echo "=========== # zksync mint test ============="
    hash=$(${CLI} send zksync nft mint -f "${creatorId}" -t "${recipientId}" -e "${contentHash}" -p 2 -n "${amount}" -k "${privkey}")
    query_tx "${CLI}" "${hash}"
}

#zksync_withdraw_nft fromId tokenId amount privkey
function zksync_withdraw_nft() {
    local fromId=$1
    local tokenId=$2
    local amount=$3
    local privkey=$4

    echo "=========== # zksync withdrawnft test ============="
    hash=$(${CLI} send zksync nft withdraw -a "${fromId}" -i "${tokenId}" -n "${amount}" -k "${privkey}")
    query_tx "${CLI}" "${hash}"
}

#zksync_transfer_nft fromId toId tokenId amount privkey
function zksync_transfer_nft() {
    local fromId=$1
    local toId=$2
    local tokenId=$3
    local amount=$4
    local privkey=$5

    echo "=========== # zksync transfer nft test ============="
    hash=$(${CLI} send zksync nft transfer -a "${fromId}" -t "${toId}" -i "${tokenId}" -n "${amount}" -k "${privkey}")
    query_tx "${CLI}" "${hash}"
}

#ZKSYNC_ACCOUNT_3 ---> 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
#ZKSYNC_ACCOUNT_4 ---> 19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
function send_l2_deposit() {
    echo "=========== # send_l2_txs ============="
    local tokenId=0
    local ethAddr="12a0E25E62C1dBD32E505446062B26AECB65F028"
    local chain33Addr="2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a"

#    docker exec build_chain33_1 ./chain33-cli zksync l2addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b

    local ethAddr4="abcd68033A72978C1084E2d44D1Fa06DdC4A2d57"
    local chain33AddrAcc4="2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1"
    local privkey="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    local Acc4privkey="19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
    local fromId=3
    local toId=4
    local amount=100000
    local queueId=4
    local contentHash="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    local contentHash_two="0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
# ZKERC1155 = 1
#	ZKERC721  = 2
    local protocol=2
    local nftTokenId=258

    zksync_deposit ${tokenId} ${amount} ${ethAddr4} ${chain33AddrAcc4} ${queueId}
}

function send_l2_txs() {
    echo "=========== # send_l2_txs ============="
    local tokenId=0
    local ethAddr="12a0E25E62C1dBD32E505446062B26AECB65F028"
#    docker exec build_chain33_1 ./chain33-cli zksync l2addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    local chain33Addr="2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a"
    local privkey="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    local Acc4privkey="19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
    local fromId=3
    local toId=4
    local amount=1
    local queueId=3
    local contentHash="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    local contentHash_two="0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
# ZKERC1155 = 1
#	ZKERC721  = 2
    local protocol=2
    local nftTokenId=258

    zksync_deposit ${tokenId} ${amount} ${ethAddr} ${chain33Addr} ${queueId}

    zksync_mint_nft ${fromId} ${toId} ${contentHash_two} ${protocol} ${amount} ${privkey}
    queueId=$((queueId + 1))
    zksync_deposit ${tokenId} ${amount} ${ethAddr} ${chain33Addr} ${queueId}
    zksync_transfer2new ${tokenId} ${amount} ${fromId} ${ethAddr} ${chain33Addr} ${privkey}
    zksync_transfer_nft ${toId} ${fromId} ${nftTokenId} ${amount} ${Acc4privkey}

    zksync_transfer ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
    zksync_transfer ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
    zksync_transfer ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
    zksync_withdraw_nft ${fromId} ${nftTokenId} ${amount} ${privkey}


    local count=0
    while true; do
        count=$((count + 1))
        #sleep 1
        if [[ ${count} -ge 10 ]]; then
            echo "send 10 L2 txs"
            break
        fi
    done
}

function batch_send_l2_txs() {
    echo "=========== # batch_send_l2_txs ============="
    local tokenId=0
    local ethAddr="12a0E25E62C1dBD32E505446062B26AECB65F028"
#    docker exec build_chain33_1 ./chain33-cli zksync l2addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    local chain33Addr="2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a"
    local privkey="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    local Acc4privkey="19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
    local fromId=3
    local toId=4
    local amount=1
    local queueId=5
    local contentHash="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    local contentHash_two="0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
    local protocol=2
    local nftTokenId=258



    local count=0
    while true; do
        #sleep 1
        zksync_deposit ${tokenId} ${amount} ${ethAddr} ${chain33Addr} ${queueId}
        zksync_transfer ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
        zksync_transfer ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
        zksync_transfer ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
        queueId=$((queueId + 1))
        zksync_deposit ${tokenId} ${amount} ${ethAddr} ${chain33Addr} ${queueId}
        count=$((count + 1))
        if [[ ${count} -ge 10 ]]; then
            echo "Finish send 10 L2 txs"
            break
        fi
    done
}

function batch_send_l2_txs_oneshoot() {
    echo "=========== # batch_send_l2_txs ============="
    local tokenId=0
    local ethAddr="12a0E25E62C1dBD32E505446062B26AECB65F028"
#    docker exec build_chain33_1 ./chain33-cli zksync l2addr -k 6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    local chain33Addr="2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a"
    local privkey="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    local Acc4privkey="19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
    local fromId=3
    local toId=4
    local amount=1
    local queueId=15
    local contentHash="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    local contentHash_two="0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
    local protocol=2
    local nftTokenId=258



    local count=0
    while true; do
        #sleep 1
        zksync_deposit_oneShoot ${tokenId} ${amount} ${ethAddr} ${chain33Addr} ${queueId}
        zksync_transfer_oneshoot ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
        zksync_transfer_oneshoot ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
        zksync_transfer_oneshoot ${tokenId} ${amount} ${fromId} ${toId} ${privkey}
        queueId=$((queueId + 1))
        zksync_deposit_oneShoot ${tokenId} ${amount} ${ethAddr} ${chain33Addr} ${queueId}
        count=$((count + 1))
        if [[ ${count} -ge 10 ]]; then
            echo "Finish send 10 L2 txs"
            break
        fi
    done
}
send_l2_txs
