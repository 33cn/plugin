#!/usr/bin/env bash
set -x
set -e

source "./public.sh"

CLI="docker exec build_chain33_1 /root/chain33-cli"
managerPrivkey="4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" #对应的chain33地址为: 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
ethAddr="12a0E25E62C1dBD32E505446062B26AECB65F028"

# 准备地址数量
addrInit=20
# 测试地址数量
addrTest=$((addrInit - 10))
idMax=$((addrTest + 3))
queueId=0
le18zero="000000000000000000"
le8zero="00000000"
l2addrs=""
keys=""
accountIDs=""
TOKENID_0="0"
TOKENID_1="1"

function create_addr_all() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    for ((i = 0; i < ${addrInit}; i++)); do
        addr[$i]=$(${CLI} account create -l "zkAddr${i}" | jq -r ".acc.addr")
        key[$i]=$(${CLI} account dump_key -a "${addr[i]}" | jq -r ".data")
        l2addr[$i]=$(${CLI} zksync l2addr -k "${key[i]}")
        hash=$(${CLI} send coins transfer -a 20 -n test -t "${addr[i]}" -k ${managerPrivkey})
    done

    l2addrs="${l2addr[0]}"
    keys="${key[0]}"
    accountIDs="3"
    for ((i = 1; i < ${addrTest}; i++)); do
        id=$((i + 3))
        l2addrs="${l2addrs},${l2addr[i]}"
        keys="${keys},${key[i]}"
        accountIDs="${accountIDs},${id}"
    done
    echo -e "${IYellow} accountIDs: ${accountIDs} ${NOC}"
    echo -e "${IYellow} l2addrs: ${l2addrs} ${NOC}"
    echo -e "${IYellow} keys: ${keys} ${NOC}"

    for ((i = 0; i < ${addrInit}; i++)); do
        echo -e "${IYellow} zkAddr${i}: ${addr[i]} *** l2addr${i}: ${l2addr[i]} *** key: ${key[i]} ${NOC}"
    done
}

function zksync_deposit_init() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    echo -e "${GRE} deposit 初始交易 确定了 account id ${NOC}"

    for ((i = 0; i < ${addrTest}; i++)); do
        hash=$(${CLI} send zksync deposit -t 0 -a 1000000000000008000000 -e ${ethAddr} -c "${l2addr[i]}" -i ${queueId} -k ${managerPrivkey})
        check_tx "${CLI}" "${hash}"
        queueId=$((queueId + 1))

        id=$((i + 3))
        balance0=$(${CLI} zksync query account token -a "${id}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "1000000000000008000000"
    done
}

function zksync_many_deposit() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    echo -e "${IYellow} deposit many 部分哈希成功 部分失败失败原因: queueId 不对 eth last priority queue id=9,new=11: ErrNotAllow${NOC}"

    ${CLI} zksync sendl2 deposit_many -e ${ethAddr} -m "1${le8zero}" -a ${l2addrs} -t 0 -k ${managerPrivkey} -q ${queueId}
    queueId=$((queueId + 10))

    ${CLI} zksync sendl2 deposit_many -e ${ethAddr} -m "2${le8zero}" -a ${l2addrs} -t 1 -k ${managerPrivkey} -q ${queueId}
    queueId=$((queueId + 10))

    for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        balance1=$(${CLI} zksync query account token -a "${i}" -t 1 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  balance1=$balance1 ${NOC}"
    done
}

function zksync_setPubKeys() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    for ((i = 0; i < ${addrTest}; i++)); do
        id=$((i + 3))
        hash=$(${CLI} send zksync pubkey -a "${id}" -k "${key[i]}")
        check_tx "${CLI}" "${hash}"
    done
}

function zksync_many_withdraw() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    ${CLI} zksync sendl2 withdraw_many -t "${TOKENID_0}" -m "2${le8zero}" -a ${accountIDs} -k ${keys}
    sleep 10

    for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
#        balance1=$(${CLI} zksync query account token -a "${i}" -t 1 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  balance1=$balance1 ${NOC}"
        is_equal "${balance0}" "999999999999807000000"
    done
}

function zksync_tree2contract_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    ${CLI} zksync sendl2 tree2contract_many -t "${TOKENID_0}" -m "3${le8zero}" -a ${accountIDs} -k ${keys}

    for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  ${NOC}"
    done
}

function zksync_contract2tree_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    ${CLI} zksync sendl2 contract2tree_many -t "${TOKENID_0}" -m "3${le8zero}" -a ${accountIDs} -k ${keys}

    for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  ${NOC}"
    done
}

function zksync_transfer_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    echo -e "${IYellow} 3转给4 4转给5 ... 最后大家都是只扣了手续费 ${NOC}"
    ${CLI} zksync sendl2 transfer_many -t "${TOKENID_0}" -m "4${le8zero}" -f 3,4,5,6,7,8,9,10,11,12 -d 4,5,6,7,8,9,10,11,12,3 \
     -k ${key[0]},${key[1]},${key[2]},${key[3]},${key[4]},${key[5]},${key[6]},${key[7]},${key[8]},${key[9]}

    sleep 2

    for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  ${NOC}"
    done

    echo -e "${IYellow} 都转给最后一个12 ${NOC}"
    ${CLI} zksync sendl2 transfer_many -t "${TOKENID_0}" -m "5${le8zero}" -f 3,4,5,6,7,8,9,10,11 -d 12,12,12,12,12,12,12,12,12 \
     -k ${key[0]},${key[1]},${key[2]},${key[3]},${key[4]},${key[5]},${key[6]},${key[7]},${key[8]}

    sleep 2

    for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  ${NOC}"
    done
}

function zksync_transfer2new_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    echo -e "${IYellow} 3转给13 4转给14 ... ${NOC}"
    ${CLI} zksync sendl2 transfer2new_many -t "${TOKENID_0}" -m "6${le8zero}" -e ${ethAddr} -f 3,4,5,6,7,8,9,10,11,12 \
     -d ${l2addr[10]},${l2addr[11]},${l2addr[12]},${l2addr[13]},${l2addr[14]},${l2addr[15]},${l2addr[16]},${l2addr[17]},${l2addr[18]},${l2addr[19]} \
     -k ${key[0]},${key[1]},${key[2]},${key[3]},${key[4]},${key[5]},${key[6]},${key[7]},${key[8]},${key[9]}

    sleep 2

    for ((i = 3; i < 23; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  ${NOC}"
    done
}

function zksync_forceexit_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    ${CLI} zksync sendl2 forceexit_many -t "${TOKENID_0}" -a ${accountIDs} -k ${keys}

    sleep 2

     for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "0"
    done
}

function zksync_test_all() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    create_addr_all
    zksync_deposit_init
#    zksync_many_deposit
    zksync_setPubKeys
#    zksync_many_withdraw
#    zksync_tree2contract_many
#    zksync_contract2tree_many
#    zksync_transfer_many
#    zksync_transfer2new_many
    zksync_forceexit_many
}
