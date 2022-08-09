#!/usr/bin/env bash
# shellcheck disable=SC2034
# shellcheck disable=SC2128
# shellcheck disable=SC2154
# shellcheck disable=SC2004
# shellcheck disable=SC2002
# shellcheck disable=SC2116
set -x
set -e

source "./public.sh"
source "./testAddress.sh"

CLI="docker exec build_chain33_1 /root/chain33-cli"
managerPrivkey="4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" #对应的chain33地址为: 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
ethAddr="12a0E25E62C1dBD32E505446062B26AECB65F028"
contentHash="4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"

# 测试地址数量
addrTest=100
# 测试 transfer2new 需要的地址 < addrTest
transfer2newAddr=2
# 共需准备地址数量
addrInit=$((addrTest + transfer2newAddr))
idMax=$((addrTest + 3))
endId=$((idMax - 1))
queueId=0
initBalance=8000000000000000000
withdrawFee=1000000
transferFee=100000
forceExitFee=1000000
nftFee=100
le18zero="000000000000000000"
le8zero="00000000"
l2addrs=""
keys=""
accountIDs=""
tokenIds=""
nftTokenId=256
TOKENID_0="0"
TOKENID_1="1"
sleepTime=20

function create_addr_all() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"    
    for ((i = 0; i < ${addrInit}; i++)); do
        ${CLI} account import_key -l "zkAddr${i}" -k "${key[i]}"
        hash=$(${CLI} send coins transfer -a 20 -n test -t "${addr[i]}" -k ${managerPrivkey})
    done
    ${CLI} send coins transfer -a 100 -n test -t "${addr[0]}" -k ${managerPrivkey}
    sleep 5

    l2addrs="${l2addr[0]}"
    accountIDs="3"
    tokenIds="258"
    for ((i = 1; i < ${addrTest}; i++)); do
        id=$((i + 3))
        l2addrs="${l2addrs},${l2addr[i]}"
        accountIDs="${accountIDs},${id}"
        tid=$((i + 258))
        tokenIds="${tokenIds},${tid}"
    done
    echo -e "${IYellow} accountIDs: ${accountIDs} ${NOC}"
    echo -e "${IYellow} l2addrs: ${l2addrs} ${NOC}"
    echo -e "${IYellow} keys: ${keys} ${NOC}"

    for ((i = 0; i < ${addrInit}; i++)); do
        echo -e "${IYellow} zkAddr${i}: ${addr[i]} *** l2addr${i}: ${l2addr[i]} ${NOC}"
    done
}

function check_balance() {
    local balanceAf=$1
    local tokenId=0
    if [[ $# -eq 2 ]]; then
        tokenId=$2
    fi

    for ((i = 3; i < ${idMax}; i++)); do
        balance=$(${CLI} zksync query account token -a "${i}" -t "$tokenId" | jq -r ".tokenBalances[].balance")
        is_equal "${balance}" "${balanceAf}"
    done
}

function zksync_deposit_init() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    echo -e "${GRE} deposit 初始交易 确定了 account id ${NOC}"

#    for ((i = 0; i < ${addrTest}; i++)); do
#        hash=$(${CLI} send zksync deposit -t 0 -a ${initBalance} -e ${ethAddr} -c "${l2addr[i]}" -i ${queueId} -k ${managerPrivkey})
#        check_tx "${CLI}" "${hash}"
#        queueId=$((queueId + 1))
#    done

    # 第一个地址先 deposit
    hash=$(${CLI} send zksync deposit -t 0 -a 2400000000000000000000 -e ${ethAddr} -c "${l2addr[0]}" -i ${queueId} -k ${managerPrivkey})
    check_tx "${CLI}" "${hash}"
    queueId=$((queueId + 1))

    ${CLI} zksync sendl2 pubkey_many -a 3 -k "${key[0]}"
    sleep 5

    tol2addrs="${l2addr[1]}"
    for ((i = 2; i < ${addrTest}; i++)); do
        tol2addrs="${tol2addrs},${l2addr[i]}"
    done

    # transfer2new_to_many 其他地址创建 account id
    ${CLI} zksync sendl2 transfer2new_to_many -t "${TOKENID_0}" -m "${initBalance}" -e ${ethAddr} -f 3 -d "${tol2addrs}" -k "${key[0]}"
    sleep $sleepTime

    # 第一个地址 forceexit_many 再 deposit -a "${initBalance}" 为了跟其他地址金额一样 方便测试
    ${CLI} zksync sendl2 forceexit_many -t "${TOKENID_0}" -a 3 -k "${key[0]}"
    sleep 2
    hash=$(${CLI} send zksync deposit -t 0 -a "${initBalance}" -e ${ethAddr} -c "${l2addr[0]}" -i ${queueId} -k ${managerPrivkey})
    check_tx "${CLI}" "${hash}"
    queueId=$((queueId + 1))

    check_balance "${initBalance}"

    # 获取 keys
    keys="${key[0]}"
    for ((i = 1; i < ${addrTest}; i++)); do
        id=$((i + 3))
        gkey=$( get_key $id )
        keys="${keys},${gkey}"
    done
}

function zksync_many_deposit() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    echo -e "${IYellow} deposit many 部分哈希成功 部分失败失败原因: queueId 不对 eth last priority queue id=9,new=11: ErrNotAllow${NOC}"

    ${CLI} zksync sendl2 deposit_many -e ${ethAddr} -m "1${le8zero}" -a "${l2addrs}" -t 0 -k ${managerPrivkey} -q ${queueId}
    queueId=$((queueId + addrTest))

    ${CLI} zksync sendl2 deposit_many -e ${ethAddr} -m "2${le8zero}" -a "${l2addrs}" -t 1 -k ${managerPrivkey} -q ${queueId}
    queueId=$((queueId + addrTest))

    for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        balance1=$(${CLI} zksync query account token -a "${i}" -t 1 | jq -r ".tokenBalances[].balance")
        echo -e "${IYellow} balance0=$balance0  balance1=$balance1 ${NOC}"
    done
}

function zksync_setPubKeys() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    ${CLI} zksync sendl2 pubkey_many -a ${accountIDs} -k "${keys}"
    sleep 5
}

function zksync_many_withdraw() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf=$(${CLI} zksync query account token -a "3" -t 0 | jq -r ".tokenBalances[].balance")
    ${CLI} zksync sendl2 withdraw_many -t "${TOKENID_0}" -m "2${le8zero}" -a ${accountIDs} -k "${keys}"
    sleep $sleepTime

    balanceAf=$(( $balanceBf - 2${le8zero} - ${withdrawFee} ))
    check_balance "${balanceAf}"
}

function zksync_tree2contract_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf=$(${CLI} zksync query account token -a "3" -t 0 | jq -r ".tokenBalances[].balance")
    ${CLI} zksync sendl2 tree2contract_many -t "${TOKENID_0}" -m "300${le8zero}" -a ${accountIDs} -k "${keys}"
    sleep $sleepTime

    balanceAf=$(( $balanceBf - 300${le8zero} ))
    check_balance "${balanceAf}"
}

function zksync_contract2tree_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf=$(${CLI} zksync query account token -a "3" -t 0 | jq -r ".tokenBalances[].balance")
    ${CLI} zksync sendl2 contract2tree_many -t "${TOKENID_0}" -m "300${le8zero}" -a ${accountIDs} -k "${keys}"
    sleep $sleepTime

    balanceAf=$(( $balanceBf + 300${le8zero} ))
    check_balance "${balanceAf}"
}

function get_transfer_many_ids() {
    toIDs="4"
    accountIDs2="3"
    end=$((addrTest - 1))
    endIDs=${endId}
    keys2="${key[0]}"
    for ((i = 1; i < ${end}; i++)); do
        tid=$((i + 4))
        toIDs="${toIDs},${tid}"
        id=$((i + 3))
        accountIDs2="${accountIDs2},${id}"
        endIDs="${endIDs},${endId}"
        gkey=$( get_key $id )
        keys2="${keys2},${gkey}"
    done
    toIDs="${toIDs},3"
}

function zksync_transfer_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf=$(${CLI} zksync query account token -a "3" -t 0 | jq -r ".tokenBalances[].balance")
    echo -e "${IYellow} 3转给4 4转给5 ... 最后大家都是只扣了手续费 ${NOC}"
    ${CLI} zksync sendl2 transfer_many -t "${TOKENID_0}" -m "4${le8zero}" -f ${accountIDs} -d ${toIDs} -k "${keys}"
    sleep $sleepTime

    balanceAf=$(( balanceBf - transferFee ))
    check_balance "${balanceAf}"

    balanceBf=$(${CLI} zksync query account token -a "3" -t 0 | jq -r ".tokenBalances[].balance")
    echo -e "${IYellow} 都转给最后一个12 ${NOC}"
    ${CLI} zksync sendl2 transfer_many -t "${TOKENID_0}" -m "5${le8zero}" -f ${accountIDs2} -d ${endIDs} -k "${keys2}"
    sleep $sleepTime

    balanceAf=$(( balanceBf - 5${le8zero} - transferFee ))
    echo -e "${IYellow} 3 4 5 ... 都少了金额 扣了手续费 ${NOC}"
    for ((i = 3; i < ${endId}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "${balanceAf}"
    done

    balanceEnd=$(( balanceBf + 5${le8zero} * end))
    echo -e "${IYellow} 最后一个金额增加 ${NOC}"
    balance0=$(${CLI} zksync query account token -a "${endId}" -t 0 | jq -r ".tokenBalances[].balance")
    is_equal "${balance0}" "${balanceEnd}"
}

function get_transfer_many_2_ids() {
    fromIDs="3"
    toIDs="4"
    end=$((addrTest-2))
    keys2="${key[0]}"
    for ((i = 2; i < ${end};)); do
        id=$((i + 3))
        fromIDs="${fromIDs},${id}"
        tid=$((i + 4))
        toIDs="${toIDs},${tid}"
        gkey=$( get_key $id )
        keys2="${keys2},${gkey}"
        i=$((i + 2))
    done
}

function zksync_transfer_many_2() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf3=$(${CLI} zksync query account token -a "3" -t 0 | jq -r ".tokenBalances[].balance")
    balanceBf4=$(${CLI} zksync query account token -a "4" -t 0 | jq -r ".tokenBalances[].balance")
    echo -e "${IYellow} 3转给4 5转给6 7转给8 ... ${NOC}"
    ${CLI} zksync sendl2 transfer_many -t "${TOKENID_0}" -m "4${le8zero}" -f ${fromIDs} -d ${toIDs} -k "${keys2}"
    sleep $sleepTime

    endId2=$(( endId - 1 ))
    balanceAf3=$(( balanceBf3 - transferFee - 4${le8zero} ))
    echo -e "${IYellow} 3 5 7 ... 都少了金额 扣了手续费 ${NOC}"
    for ((i = 3; i < ${endId2};)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "${balanceAf3}"
        i=$((i + 2))
    done

    balanceAf4=$(( balanceBf4 + 4${le8zero} ))
    echo -e "${IYellow} 4 6 8 ... 都多了金额 ${NOC}"
    for ((i = 4; i < ${endId2};)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "${balanceAf4}"
        i=$((i + 2))
    done
}

function get_transfer2new_many_ids() {
    accountIDs3="3"
    toIDs="${idMax}"
    tol2addrs="${l2addr[addrTest]}"
    keys2="${key[0]}"
    for ((i = 1; i < ${transfer2newAddr}; i++)); do
        id=$((i + 3))
        accountIDs3="${accountIDs3},${id}"
        tid=$((i + idMax))
        toIDs="${toIDs},${tid}"
        toAddr=$((i + addrTest))
        tol2addrs="${tol2addrs},${l2addr[toAddr]}"
        gkey=$( get_key $id )
        keys2="${keys2},${gkey}"
    done
}

function zksync_transfer2new_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf=$(${CLI} zksync query account token -a "3" -t 0 | jq -r ".tokenBalances[].balance")
    echo -e "${IYellow} 3转给13 4转给14 ... ${NOC}"
    ${CLI} zksync sendl2 transfer2new_many -t "${TOKENID_0}" -m "6${le8zero}" -e ${ethAddr} -f ${accountIDs3} -d "${tol2addrs}" -k "${keys2}"
    sleep $sleepTime

    balanceAf=$(( balanceBf - 6${le8zero} - transferFee ))
    echo -e "${IYellow} 3 4 5 ... 11 12 都扣了手续费和金额 最后几个地址增加金额 ${NOC}"
    for ((i = 0; i < ${transfer2newAddr}; i++)); do
        id=$((i + 3))
        balance0=$(${CLI} zksync query account token -a "${id}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "${balanceAf}"

        tid=$((i + idMax))
        balance0=$(${CLI} zksync query account token -a "${tid}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "6${le8zero}"
    done
}

function zksync_forceexit_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    ${CLI} zksync sendl2 forceexit_many -t "${TOKENID_0}" -a ${accountIDs} -k "${keys}"
    sleep $sleepTime

    check_balance "0"
}

function zksync_mintNFT_ERC1155_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf1=$(${CLI} zksync query account token -a 1 -t 0 | jq -r ".tokenBalances[].balance")
    balanceBf=$(${CLI} zksync query account token -a 3 -t 0 | jq -r ".tokenBalances[].balance")
    ${CLI} zksync sendl2 nft mint -p 1 -m 1000 -e "${keys}" -t "${accountIDs}" -f ${accountIDs} -k "${keys}"
    sleep $sleepTime

    balance=$(${CLI} zksync query account token -a 1 -t 0 | jq -r ".tokenBalances[].balance")
    balanceAf1=$((balanceBf1 + ${addrTest} * ${nftFee}))
    is_equal "${balance}" "${balanceAf1}"

    balance=$(${CLI} zksync query account token -a 2 -t ${nftTokenId} | jq -r ".tokenBalances[].balance")
    balance256=$((addrTest + nftTokenId + 1))
    is_equal "${balance}" "${balance256}"

    balanceAf=$((balanceBf - nftFee))
    check_balance "${balanceAf}"
}

function zksync_mintNFT_ERC721_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf1=$(${CLI} zksync query account token -a 1 -t 0 | jq -r ".tokenBalances[].balance")
    balanceBf=$(${CLI} zksync query account token -a 3 -t 0 | jq -r ".tokenBalances[].balance")
    ${CLI} zksync sendl2 nft mint -p 2 -m 1 -e "${keys}" -t "${accountIDs}" -f ${accountIDs} -k "${keys}"
    sleep $sleepTime

    balance=$(${CLI} zksync query account token -a 1 -t 0 | jq -r ".tokenBalances[].balance")
    balanceAf1=$((balanceBf1 + ${addrTest} * ${nftFee}))
    is_equal "${balance}" "${balanceAf1}"

    balance=$(${CLI} zksync query account token -a 2 -t ${nftTokenId} | jq -r ".tokenBalances[].balance")
    balance256=$((addrTest + nftTokenId + 1))
    is_equal "${balance}" "${balance256}"

    balanceAf=$((balanceBf - nftFee))
    check_balance "${balanceAf}"
    check_balance "1" ${nftTokenId}

#    # 结果不再匹配 因为258 259 260 ... 执行顺序没有按照id 3 4 5 ... 结果不能匹配查询
#    for ((i = ${$((nftTokenId + 2))}; i < ${$((addrTest + nftTokenId + 1))}; i++)); do
#        balance256=$(${CLI} zksync query account token -a 2 -t ${i} | jq -r ".tokenBalances[].balance")
#    done
}

function zksync_transferNFT_ERC721_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    balanceBf=$(${CLI} zksync query account token -a 3 -t 0 | jq -r ".tokenBalances[].balance")
    ${CLI} zksync sendl2 nft transfer -t ${nftTokenId} -m 1 -t "${tokenIds}" -a "${accountIDs}" -r ${accountIDs} -k "${keys}"
    sleep $sleepTime

    balanceAf=$((balanceBf - nftFee))
    check_balance "${balanceAf}"
    check_balance "1" ${nftTokenId}
}

function zksync_withdrawNFT_many() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    local amount=$1
    ${CLI} zksync sendl2 nft withdraw -a "${accountIDs}" -k "${keys}" -t "${tokenIds}" -m "$amount"

    sleep 10

     for ((i = 3; i < ${idMax}; i++)); do
        balance0=$(${CLI} zksync query account token -a "${i}" -t 0 | jq -r ".tokenBalances[].balance")
        is_equal "${balance0}" "0"
    done
}

function zksync_deposit_transfer() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    # forceexit 后 让所有id金额为 initBalance 方便后面测试
    # 第一个地址先 deposit
    hash=$(${CLI} send zksync deposit -t 0 -a 2400000000000000000000 -e ${ethAddr} -c "${l2addr[0]}" -i ${queueId} -k ${managerPrivkey})
    check_tx "${CLI}" "${hash}"
    queueId=$((queueId + 1))

    toIDs="4"
    end=$((addrTest - 1))
    for ((i = 1; i < ${end}; i++)); do
        tid=$((i + 4))
        toIDs="${toIDs},${tid}"
    done
    ${CLI} zksync sendl2 transfer_many_2 -t "${TOKENID_0}" -m "${initBalance}" -f 3 -d ${toIDs} -k "${key[0]}"
    sleep $sleepTime

    # 第一个地址 forceexit_many 再 deposit -a "${initBalance}" 为了跟其他地址金额一样 方便测试
    ${CLI} zksync sendl2 forceexit_many -t "${TOKENID_0}" -a 3 -k "${key[0]}"
    sleep 2
    hash=$(${CLI} send zksync deposit -t 0 -a "${initBalance}" -e ${ethAddr} -c "${l2addr[0]}" -i ${queueId} -k ${managerPrivkey})
    check_tx "${CLI}" "${hash}"
    queueId=$((queueId + 1))

    check_balance "${initBalance}"
}

function zksync_test_all() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
    create_addr_all
    zksync_deposit_init
    zksync_setPubKeys

    # 测试不通过 err: eth last priority queue id=6,new=4 queueId 不按照顺序打包
#    zksync_many_deposit

    # 10 200 300 个地址测试通过
    echo -e "${GRE}=========== withdraw ===========${NOC}"
    zksync_many_withdraw

    # 10 200 300 个地址测试通过
    echo -e "${GRE}=========== tree2contract contract2tree ===========${NOC}"
    zksync_tree2contract_many
    zksync_contract2tree_many

    # 10 200 300 个地址测试通过
    echo -e "${GRE}=========== transfer ===========${NOC}"
    get_transfer_many_ids
    zksync_transfer_many

    get_transfer_many_2_ids
    zksync_transfer_many_2

#    get_transfer2new_many_ids
#    zksync_transfer2new_many

    zksync_forceexit_many
    zksync_deposit_transfer
    echo -e "${GRE}=========== nft ERC1155 和 ERC721 分开测试 ===========${NOC}"
    zksync_mintNFT_ERC1155_many

    # 测试不通过 err: token not exist with Id=259 NFTTokenId不按照顺序生成
#    zksync_mintNFT_ERC721_many
#    zksync_transferNFT_ERC721_many
#    zksync_withdrawNFT_many 1
}
