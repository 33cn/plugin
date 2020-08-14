#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
# shellcheck disable=SC2178
set -x

source "./publicTest.sh"
source "./allRelayerTest.sh"
Ethsender="0xa4ea64a583f6e51c3799335b28a8f0529570a635"
privateKeys[0]="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
privateKeys[1]="3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
privateKeys[2]="1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695"
privateKeys[3]="4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf"
privateKeys[4]="62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9"
privateKeys[5]="355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"
privateKeys[6]="9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"
ethAddress[0]="0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a"
ethAddress[1]="0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f"
ethAddress[2]="0xdb15E7327aDc83F2878624bBD6307f5Af1B477b4"
ethAddress[3]="0x9cBA1fF8D0b0c9Bc95d5762533F8CddBE795f687"
ethAddress[4]="0x1919203bA8b325278d28Fb8fFeac49F2CD881A4e"
ethAddress[5]="0xA4Ea64a583F6e51C3799335b28a8F0529570A635"
ethAddress[6]="0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF"

maturityDegree=10
tokenAddr=""
tokenAddrBty=""
CLIA=""
ethUrl=""
Chain33Cli=""

loop_send_lock_eth() {
    # while 遍历数组
    echo -e "${GRE}=========== Ethereum Lock begin ===========${NOC}"
    #shellcheck disable=SC2154
    preChain33Balance=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]" | jq ".balance" | sed 's/\"//g')

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(curl -ksd '{"jsonrpc":"2.0","method":"eth_getBalance","params":["'${ethAddress[i]}'", "latest"],"id":1}' "${ethUrl}" | jq -r ".result")
        ethTxHash=$(${CLIA} relayer ethereum lock-async -m 1 -k "${privateKeys[i]}" -r 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
        echo ${i} "lock-async tx hash:" "${ethTxHash}"
        # shellcheck disable=SC2219
        let i++
    done

    #shellcheck disable=SC2154
    eth_block_wait $((maturityDegree + 2)) "${ethUrl}"

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(curl -ksd '{"jsonrpc":"2.0","method":"eth_getBalance","params":["'${ethAddress[i]}'", "latest"],"id":1}' "${ethUrl}" | jq -r ".result")
        res=$((preEthBalance[i] - nowEthBalance))
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" ${res}
        if [[ $res -le 100000000000000000 ]]; then
            echo -e "${RED}error number, expect greater than 100000000000000000, get ${res}${NOC}"
            exit 1
        fi
        # shellcheck disable=SC2219
        let i++
    done
    nowChain33Balance=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]" | jq ".balance" | sed 's/\"//g')
    diff=$(echo "$nowChain33Balance - $preChain33Balance" | bc)
    check_number "${diff}" 7
}

loop_send_burn_eth() {
    echo -e "${GRE}=========== Chain33 Burn begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]" | jq ".balance" | sed 's/\"//g')

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(curl -ksd '{"jsonrpc":"2.0","method":"eth_getBalance","params":["'${ethAddress[i]}'", "latest"],"id":1}' "${ethUrl}" | jq -r ".result")
        ethTxHash=$(${Chain33Cli} send x2ethereum burn -a 1 -r ${ethAddress[i]} -t eth --node_addr "${ethUrl}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
        echo ${i} "burn chain33 tx hash:" "${ethTxHash}"
        # shellcheck disable=SC2219
        let i++
    done

    eth_block_wait $((maturityDegree + 2)) "${ethUrl}"

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(curl -ksd '{"jsonrpc":"2.0","method":"eth_getBalance","params":["'${ethAddress[i]}'", "latest"],"id":1}' "${ethUrl}" | jq -r ".result")
        res=$((nowEthBalance - preEthBalance[i]))
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" ${res}
        if [[ $res -gt 1000000000000000000 ]]; then
            echo -e "${RED}error number, expect greater than 1000000000000000000, get ${res}${NOC}"
            exit 1
        fi
        # shellcheck disable=SC2219
        let i++
    done
    nowChain33Balance=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]" | jq ".balance" | sed 's/\"//g')
    diff=$(echo "$preChain33Balance - $nowChain33Balance" | bc)
    check_number "${diff}" 7
}

loop_send_lock_bty() {
    echo -e "${GRE}=========== Chain33 Lock begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum | jq -r ".balance" | sed 's/\"//g')

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(${CLIA} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddrBty}" | jq -r ".balance")
        ethTxHash=$(${Chain33Cli} send x2ethereum lock -q "${tokenAddrBty}" -a 1 -r ${ethAddress[i]} -t coins.bty --node_addr "${ethUrl}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
        echo ${i} "lock chain33 tx hash:" "${ethTxHash}"
        # shellcheck disable=SC2219
        let i++
    done

    eth_block_wait $((maturityDegree + 2)) "${ethUrl}"

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddrBty}" | jq -r ".balance")
        res=$((nowEthBalance - preEthBalance[i]))
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" ${res}
        check_number "${res}" 1
        # shellcheck disable=SC2219
        let i++
    done
    nowChain33Balance=$(${Chain33Cli} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum | jq -r ".balance" | sed 's/\"//g')
    diff=$(echo "$preChain33Balance - $nowChain33Balance" | bc)
    check_number "${diff}" 7
}

loop_send_burn_bty() {
    echo -e "${GRE}=========== Ethereum Burn begin ===========${NOC}"

    preChain33Balance=$(${Chain33Cli} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum | jq -r ".balance" | sed 's/\"//g')

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(${CLIA} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddrBty}" | jq -r ".balance")
        approveTxHash=$(${CLIA} relayer ethereum approve -m 1 -k "${privateKeys[i]}" -t "${tokenAddrBty}")
        ethTxHash=$(${CLIA} relayer ethereum burn-async -m 1 -k "${privateKeys[i]}" -r 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t "${tokenAddrBty}")
        echo ${i} "burn-async tx hash:" "${ethTxHash}"
        # shellcheck disable=SC2219
        let i++
    done

    eth_block_wait $((maturityDegree + 2)) "${ethUrl}"

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddrBty}" | jq -r ".balance")
        res=$((preEthBalance[i] - nowEthBalance))
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" ${res}
        check_number "${res}" 1
        # shellcheck disable=SC2219
        let i++
    done
    nowChain33Balance=$(${Chain33Cli} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum | jq -r ".balance" | sed 's/\"//g')
    diff=$(echo "$nowChain33Balance - $preChain33Balance" | bc)
    check_number "${diff}" 7
}

loop_send_lock_erc20() {
    echo -e "${GRE}=========== Ethereum Lock Erc20 begin ===========${NOC}"
    preChain33Balance=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t testc | jq ".res" | jq ".[]" | jq ".balance" | sed 's/\"//g')
    preEthBalance=$(${CLIA} relayer ethereum balance -o "${Ethsender}" -t "${tokenAddr}" | jq -r ".balance")
    approveTxHash=$(${CLIA} relayer ethereum approve -m 10 -k "${privateKeys[5]}" -t "${tokenAddr}")

    i=0
    echo ${i} "lock-async erc20 approve tx hash:" "${approveTxHash}"
    while [[ i -lt ${#privateKeys[@]} ]]; do
        ethTxHash=$(${CLIA} relayer ethereum lock-async -m 1 -k "${privateKeys[5]}" -r 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t "${tokenAddr}")
        echo ${i} "lock-async erc20 tx hash:" "${ethTxHash}"
        # shellcheck disable=SC2219
        let i++
    done
    eth_block_wait $((maturityDegree + 2)) "${ethUrl}"

    nowEthBalance=$(${CLIA} relayer ethereum balance -o "${Ethsender}" -t "${tokenAddr}" | jq -r ".balance")
    res=$((preEthBalance - nowEthBalance))
    echo ${i} "preBalance" "${preEthBalance}" "nowBalance" "${nowEthBalance}" "diff" ${res}
    check_number "${diff}" 7

    nowChain33Balance=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t testc | jq ".res" | jq ".[]" | jq ".balance" | sed 's/\"//g')
    diff=$((nowChain33Balance - preChain33Balance))
    check_number "${diff}" 7
}

loop_send_burn_erc20() {
    echo -e "${GRE}=========== Chain33 Burn Erc20 begin ===========${NOC}"
    preChain33Balance=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t testc | jq ".res" | jq ".[]" | jq ".balance" | sed 's/\"//g')

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[i]=$(${CLIA} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddr}" | jq -r ".balance")
        ethTxHash=$(${Chain33Cli} send x2ethereum burn -a 1 -r ${ethAddress[i]} -t testc -q "${tokenAddr}" --node_addr "${ethUrl}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
        echo ${i} "burn chain33 tx hash:" "${ethTxHash}"
        # shellcheck disable=SC2219
        let i++
    done

    eth_block_wait $((maturityDegree + 2)) "${ethUrl}"

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLIA} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddr}" | jq -r ".balance")
        res=$((nowEthBalance - preEthBalance[i]))
        echo ${i} "preBalance" "${preEthBalance[i]}" "nowBalance" "${nowEthBalance}" "diff" ${res}
        check_number "${res}" 1
        # shellcheck disable=SC2219
        let i++
    done
    nowChain33Balance=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t testc | jq ".res" | jq ".[]" | jq ".balance" | sed 's/\"//g')
    diff=$(echo "$preChain33Balance - $nowChain33Balance" | bc)
    check_number "${diff}" 7
}

perf_test_main() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    loop_send_lock_eth
    loop_send_burn_eth
    loop_send_lock_bty
    loop_send_burn_bty
    loop_send_lock_erc20
    loop_send_burn_erc20

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
