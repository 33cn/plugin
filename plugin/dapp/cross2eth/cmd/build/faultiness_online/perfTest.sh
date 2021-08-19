#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
# shellcheck disable=SC2178
set -x

source "./publicTest.sh"
source "./relayerPublic.sh"

ethAddress[0]=0xdb15E7327aDc83F2878624bBD6307f5Af1B477b4
ethAddress[1]=0x9cBA1fF8D0b0c9Bc95d5762533F8CddBE795f687
ethAddress[2]=0x1919203bA8b325278d28Fb8fFeac49F2CD881A4e
ethAddress[3]=0xA4Ea64a583F6e51C3799335b28a8F0529570A635
ethAddress[4]=0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF
ethAddress[5]=0x0921948C0d25BBbe85285CB5975677503319F02A
ethAddress[6]=0x69921517970a28b73ac5E4C8ac8Fd135A80D2be1
ethAddress[7]=0x9a8438aeaa86aFA6a18e39551E14078fb2bdea90
ethAddress[8]=0xcF73954c8BaE39cCe0eb633885B6Bcf93f70c867
ethAddress[9]=0x1A03038C468F2520288c352d74232f862BA7c6a6
ethAddress[10]=0xe51a1c7f7C704D8FcC192b242bE656fFB34A70e4
ethAddress[11]=0x4c85848a7E2985B76f06a7Ed338FCB3aF94a7DCf
ethAddress[12]=0x6F163E6daf0090D897AD7016484f10e0cE844994
ethAddress[13]=0xbc333839E37bc7fAAD0137aBaE2275030555101f
ethAddress[14]=0x495953A743ef169EC5D4aC7b5F786BF2Bd56aFd5

privateKeys[0]=0x1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695
privateKeys[1]=0x4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf
privateKeys[2]=0x62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9
privateKeys[3]=0x355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71
privateKeys[4]=0x9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08
privateKeys[5]=0x5a43f2c8724f60ea5d6b87ad424daa73639a5fc76702edd3e5eaed37aaffdf49
privateKeys[6]=0x03b28c0fc78c6ebae719b559b0781db24644b655d4bd58e5cf2311c9f03baa3d
privateKeys[7]=0x72dff1c863631208a3d4f67a5fb0b7ebe69f827a75f332e2069dc7c825cb2202
privateKeys[8]=0xecbc20b02e1ffd321e31c2a6d7d35a69715ba43ef2b0048a27de4f67b8249bde
privateKeys[9]=0x1649955b3f2852a9cd71e50237b5a6f717539cffbe336bfcd95eb19a1b5c6f1b
privateKeys[10]=0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd
privateKeys[11]=0x5e8aadb91eaa0fce4df0bcc8bd1af9e703a1d6db78e7a4ebffd6cf045e053574
privateKeys[12]=0x0504bcb22b21874b85b15f1bfae19ad62fc2ad89caefc5344dc669c57efa60db
privateKeys[13]=0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2
privateKeys[14]=0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697

# ETH 部署合约者的私钥 用于部署合约时签名使用
#ethDeployAddr="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
ethDeployKey="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

# chain33 部署合约者的私钥 用于部署合约时签名使用
chain33DeployAddr="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"
#chain33DeployKey="0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"

chain33ReceiverAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
chain33ReceiverAddrKey="4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"

maturityDegree=10

Chain33Cli="../../chain33-cli"
chain33BridgeBank=""
chain33BtyTokenAddr="1111111111111111111114oLvT2"
chain33EthTokenAddr=""
ethereumBtyTokenAddr=""
chain33YccTokenAddr=""
ethereumYccTokenAddr=""

CLIA="./ebcli_A"
chain33ID=33

function loop_send_lock_bty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} account balance -a "${chain33DeployAddr}" -e evm | jq -r ".balance" | sed 's/\"//g')

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(${CLIA} ethereum balance -o "${ethAddress[i]}" -t "${ethereumBtyTokenAddr}" | jq -r ".balance")

        hash=$(${Chain33Cli} evm call -f 1 -a 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "lock(${ethAddress[i]}, ${chain33BtyTokenAddr}, 100000000)" --chainID "${chain33ID}")
        check_tx "${Chain33Cli}" "${hash}"

        i=$((i+1))
    done

    eth_block_wait $((maturityDegree + 2))

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} ethereum balance -o "${ethAddress[i]}" -t "${ethereumBtyTokenAddr}" | jq -r ".balance")
        res=$((nowEthBalance - preEthBalance[i]))
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" ${res}
        check_number "${res}" 1
        i=$((i+1))
    done
    nowChain33Balance=$(${Chain33Cli} account balance -a "${chain33DeployAddr}" -e evm | jq -r ".balance" | sed 's/\"//g')
    diff=$(echo "$preChain33Balance - $nowChain33Balance" | bc)
    check_number "${diff}" "${#privateKeys[@]}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function loop_send_burn_bty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} account balance -a "${chain33ReceiverAddr}" -e evm | jq -r ".balance" | sed 's/\"//g')

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(${CLIA} ethereum balance -o "${ethAddress[i]}" -t "${ethereumBtyTokenAddr}" | jq -r ".balance")
        result=$(${CLIA} ethereum burn -m 1 -k "${privateKeys[i]}" -r "${chain33ReceiverAddr}" -t "${ethereumBtyTokenAddr}" )
        cli_ret "${result}" "burn"
        i=$((i+1))
    done

    eth_block_wait $((maturityDegree + 2))

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} ethereum balance -o "${ethAddress[i]}" -t "${ethereumBtyTokenAddr}" | jq -r ".balance")
        res=$((preEthBalance[i] - nowEthBalance))
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" ${res}
        check_number "${res}" 1
        i=$((i+1))
    done
    nowChain33Balance=$(${Chain33Cli} account balance -a "${chain33ReceiverAddr}" -e evm | jq -r ".balance" | sed 's/\"//g')
    diff=$(echo "$nowChain33Balance - $preChain33Balance" | bc)
    check_number "${diff}" "${#privateKeys[@]}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function loop_send_lock_eth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} evm abi call -a "${chain33EthTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(${CLIA} ethereum balance -o "${ethAddress[i]}" | jq -r ".balance")
        result=$(${CLIA} ethereum lock -m 1 -k "${privateKeys[i]}" -r "${chain33ReceiverAddr}")
        cli_ret "${result}" "lock"
        i=$((i+1))
    done

    eth_block_wait $((maturityDegree + 2))

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} ethereum balance -o "${ethAddress[i]}" | jq -r ".balance")
        res=$(echo "${preEthBalance[i]} - $nowEthBalance" | bc)
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" "${res}"
        diff=$(echo "$res >= 1"| bc) # 浮点数比较 判断是否大于1 大于返回1 小于返回0
        if [ "${diff}" -ne 1 ]; then
            echo -e "${RED}error number, expect greater than 1, get ${res}${NOC}"
            exit 1
        fi
        i=$((i+1))
    done
    nowChain33Balance=$(${Chain33Cli} evm abi call -a "${chain33EthTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    diff=$(echo "$nowChain33Balance - $preChain33Balance" | bc)
    diff=$(echo "$diff / 100000000" | bc)
    check_number "${diff}" "${#privateKeys[@]}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function loop_send_burn_eth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} evm abi call -a "${chain33EthTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(${CLIA} ethereum balance -o "${ethAddress[i]}" | jq -r ".balance")
        ethTxHash=$(${CLIA} chain33 burn -m 1 -k "${chain33ReceiverAddrKey}" -r "${ethAddress[i]}" -t "${chain33EthTokenAddr}" | jq -r ".msg")
        echo ${i} "burn chain33 tx hash:" "${ethTxHash}"
        i=$((i+1))
    done

    eth_block_wait $((maturityDegree + 2))

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} ethereum balance -o "${ethAddress[i]}" | jq -r ".balance")
        res=$(echo "$nowEthBalance - ${preEthBalance[i]}" | bc)
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" "${res}"
        diff=$(echo "$res >= 1"| bc) # 浮点数比较 判断是否大于1 大于返回1 小于返回0
        if [ "${diff}" -ne 1 ]; then
            echo -e "${RED}error number, expect greater than 1, get ${res}${NOC}"
            exit 1
        fi
        i=$((i+1))
    done
    nowChain33Balance=$(${Chain33Cli} evm abi call -a "${chain33EthTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    diff=$(echo "$preChain33Balance - $nowChain33Balance" | bc)
    diff=$(echo "$diff / 100000000" | bc)
    check_number "${diff}" "${#privateKeys[@]}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function loop_send_lock_ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} evm abi call -a "${chain33YccTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    # 先往每个ETH地址中导入token
    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        ethTxHash=$(${CLIA} ethereum transfer -m 10 -k "${ethDeployKey}" -r "${ethAddress[i]}" -t "${ethereumYccTokenAddr}" | jq -r ".msg")
        echo ${i} "burn chain33 tx hash:" "${ethTxHash}"
        i=$((i+1))
    done

    sleep 2

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[i]=$(${CLIA} ethereum balance -o "${ethAddress[i]}" -t "${ethereumYccTokenAddr}" | jq -r ".balance")
        ethTxHash=$(${CLIA} ethereum lock -m 1 -k "${privateKeys[i]}" -r "${chain33ReceiverAddr}" -t "${ethereumYccTokenAddr}" | jq -r ".msg")
        echo ${i} "lock ycc tx hash:" "${ethTxHash}"
        i=$((i+1))
    done
    eth_block_wait $((maturityDegree + 2))

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} ethereum balance -o "${ethAddress[i]}" -t "${ethereumYccTokenAddr}" | jq -r ".balance")
        res=$(echo "${preEthBalance[i]} - $nowEthBalance" | bc)
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" "${res}"
        check_number "${res}" 1
        i=$((i+1))
    done

    nowChain33Balance=$(${Chain33Cli} evm abi call -a "${chain33YccTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    diff=$((nowChain33Balance - preChain33Balance))
    diff=$(echo "$diff / 100000000" | bc)
    check_number "${diff}" "${#privateKeys[@]}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function loop_send_burn_ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} evm abi call -a "${chain33YccTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[i]=$(${CLIA} ethereum balance -o "${ethAddress[i]}" -t "${ethereumYccTokenAddr}" | jq -r ".balance")
        ethTxHash=$(${CLIA} chain33 burn -m 1 -k "${chain33ReceiverAddrKey}" -r "${ethAddress[i]}" -t "${chain33YccTokenAddr}" | jq -r ".msg")
        echo ${i} "burn chain33 tx hash:" "${ethTxHash}"
        i=$((i+1))
    done

    eth_block_wait $((maturityDegree + 2))

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} ethereum balance -o "${ethAddress[i]}" -t "${ethereumYccTokenAddr}" | jq -r ".balance")
        res=$((nowEthBalance - preEthBalance[i]))
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" ${res}
        check_number "${res}" 1
        i=$((i+1))
    done
    nowChain33Balance=$(${Chain33Cli} evm abi call -a "${chain33YccTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    diff=$(echo "$preChain33Balance - $nowChain33Balance" | bc)
    diff=$(echo "$diff / 100000000" | bc)
    check_number "${diff}" "${#privateKeys[@]}"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function perf_test_main() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    StartChain33
    start_trufflesuite
    AllRelayerStart

    loop_send_lock_bty
    loop_send_burn_bty
    loop_send_lock_eth
    loop_send_burn_eth
    loop_send_lock_ycc
    loop_send_burn_ycc

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

perf_test_main 10
