#!/usr/bin/env bash
set -x
set -e

source "./public.sh"

CLI="docker exec build_chain33_1 /root/chain33-cli"
managerPrivkey="4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" #对应的chain33地址为: 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv

# 测试地址数量
addrCount=10
queueId=0

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
    check_tx "${CLI}" "${hash}"
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
    check_tx "${CLI}" "${hash}"
}

#zksync_transfer tokenId amount fromAccountId toAccountId privkey
function zksync_transfer() {
    local tokenId=$1
    local amount=$2
    local fromAccountId=$3
    local toAccountId=$4
    local privkey=$5
    echo "=========== # zksync transfer test ============="
    hash=$(${CLI} send zksync transfer -t "${tokenId}" -a "${amount}" -f "${fromAccountId}" -o "${toAccountId}" -k "${privkey}")
    check_tx "${CLI}" "${hash}"
}

#zksync_forcexit tokenId amount privkey
function zksync_forcexit() {
    local tokenId=$1
    local amount=$2
    local privkey=$3
    echo "=========== # zksync forceexit test ============="
    hash=$(${CLI} send zksync forceexit -t "${tokenId}" -a "${amount}" -k "${privkey}")
    check_tx "${CLI}" "${hash}"
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
    check_tx "${CLI}" "${hash}"
}

#zksync_withdraw_nft fromId tokenId amount privkey
function zksync_withdraw_nft() {
    local fromId=$1
    local tokenId=$2
    local amount=$3
    local privkey=$4

    echo "=========== # zksync withdrawnft test ============="
    hash=$(${CLI} send zksync nft withdraw -a "${fromId}" -i "${tokenId}" -n "${amount}" -k "${privkey}")
    check_tx "${CLI}" "${hash}"
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
    check_tx "${CLI}" "${hash}"
}

#ZKSYNC_ACCOUNT_3 ---> 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
#ZKSYNC_ACCOUNT_4 ---> 19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4

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
    local queueId=1
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

function create_addr_all() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    for ((i = 0; i < ${addrCount}; i++)); do
        addr[$i]=$(${CLI} account create -l "zkAddr${i}" | jq -r ".acc.addr")
        key[$i]=$(${CLI} account dump_key -a "${addr[i]}" | jq -r ".data")
        l2addr[$i]=$(${CLI} zksync l2addr -k "${key[i]}")
        hash=$(${CLI} send coins transfer -a 100 -n test -t "${addr[i]}" -k 4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01)
    done

    for ((i = 0; i < ${addrCount}; i++)); do
        echo -e "${IYellow} zkAddr${i}: ${addr[i]} *** l2addr${i}: ${l2addr[i]} *** key: ${key[i]} ${NOC}"
    done
}

function zksync_deposit_init() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    echo -e "${GRE} deposit 初始交易 确定了 account id ${NOC}"

    for ((i = 0; i < ${addrCount}; i++)); do
#        echo -e "${IYellow} zkAddr${i}: ${addr[i]} *** l2addr${i}: ${l2addr[i]} *** key: ${key[i]} ${NOC}"
        hash=$(${CLI} send zksync deposit -t 0 -a 1000000000000000000000 -e 12a0E25E62C1dBD32E505446062B26AECB65F028 -c "${l2addr[i]}" -i ${queueId} -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
        check_tx "${CLI}" "${hash}"
        queueId=$((queueId + 1))

        id=$((i + 3))
        balance0=$(${CLI} zksync query account token -a "${id}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "1000000000000000000000"
    done
}

function zksync_many_deposit() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    echo -e "${IYellow} deposit many 部分哈希成功 部分失败失败原因: queueId 不对 eth last priority queue id=9,new=11: ErrNotAllow${NOC}"

    ${CLI} zksync sendl2 deposit_many -e 12a0E25E62C1dBD32E505446062B26AECB65F028 -m 1000 -a ${l2addr[0]},${l2addr[1]},${l2addr[2]},${l2addr[3]},${l2addr[4]},${l2addr[5]},${l2addr[6]},${l2addr[7]},${l2addr[8]},${l2addr[9]} \
     -t 0 -k 4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01 -q ${queueId}
    queueId=$((queueId + 10))

    ${CLI} zksync sendl2 deposit_many -e 12a0E25E62C1dBD32E505446062B26AECB65F028 -m 2000 -a ${l2addr[0]},${l2addr[1]},${l2addr[2]},${l2addr[3]},${l2addr[4]},${l2addr[5]},${l2addr[6]},${l2addr[7]},${l2addr[8]},${l2addr[9]} \
     -t 1 -k 4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01 -q ${queueId}
    queueId=$((queueId + 10))

    for ((i = 3; i < 13; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        balance1=$(${CLI} zksync query account token -a "${i}" -t 1 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  balance1=$balance1 ${NOC}"
    done
}

function zksync_setPubKeys() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    for ((i = 0; i < ${addrCount}; i++)); do
        id=$((i + 3))
        hash=$(${CLI} send zksync pubkey -a "${id}" -k "${key[i]}")
        check_tx "${CLI}" "${hash}"
    done
}

function zksync_many_withdraw() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    ${CLI} zksync sendl2 withdraw_many -t "${TOKENID_0}" -m 200000 -a 3,4,5,6,7,8,9,10,11,12 \
     -k ${key[0]},${key[1]},${key[2]},${key[3]},${key[4]},${key[5]},${key[6]},${key[7]},${key[8]},${key[9]}

    for ((i = 3; i < 13; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
#        balance1=$(${CLI} zksync query account token -a "${i}" -t 1 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  balance1=$balance1 ${NOC}"
    done
}

function zksync_test_all() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    create_addr_all
    zksync_deposit_init
#    zksync_many_deposit
    zksync_setPubKeys
    zksync_many_withdraw
}
