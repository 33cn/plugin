#!/usr/bin/env bash
set -x

source "./../ebrelayer/publicTest.sh"

CLI="./ebcli_A"
docker_chain33_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' build_chain33_1)
Chain33_CLI="$GOPATH/src/github.com/33cn/plugin/build/chain33-cli --rpc_laddr http://${docker_chain33_ip}:8801"

Ethsender="0xa4ea64a583f6e51c3799335b28a8f0529570a635"
tokenAddr="0x9C3D40A44a2F61Ef8D46fa8C7A731C08FB16cCEF"
testcAddr="0xb43393f9f588fC18Bbd8E99716c25291dB804b41"

ethSender0PrivateKey="3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"

privateKeys[0]="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
privateKeys[1]="3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
privateKeys[2]="a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
privateKeys[3]="bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
privateKeys[4]="c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"
privateKeys[5]="1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695"
privateKeys[6]="4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf"
privateKeys[7]="62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9"
privateKeys[8]="355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"
privateKeys[9]="9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"

ethAddress[0]="0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a"
ethAddress[1]="0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f"
ethAddress[2]="0x0df9a824699Bc5878232C9e612fE1A5346a5A368"
ethAddress[3]="0xcB074CB21cdDDF3ce9c3C0a7AC4497d633C9D9f1"
ethAddress[4]="0xd9dAb021e74EcF475788ed7b61356056B2095830"
ethAddress[5]="0xdb15E7327aDc83F2878624bBD6307f5Af1B477b4"
ethAddress[6]="0x9cBA1fF8D0b0c9Bc95d5762533F8CddBE795f687"
ethAddress[7]="0x1919203bA8b325278d28Fb8fFeac49F2CD881A4e"
ethAddress[8]="0xA4Ea64a583F6e51C3799335b28a8F0529570A635"
ethAddress[9]="0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF"

loop_send_lock_eth() {
    #while 遍历数组
    #    ======================== Ethereum Lock =========================================
    echo -e "${GRE}=========== Ethereum Lock begin ===========${NOC}"
    preChain33Balance=$(${Chain33_CLI} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]" | jq ".balance")

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(curl -ksd '{"jsonrpc":"2.0","method":"eth_getBalance","params":["'${ethAddress[i]}'", "latest"],"id":1}' http://localhost:7545 | jq -r ".result")
        ethTxHash=$(${CLI} relayer ethereum lock-async -m 0.1 -k "${privateKeys[i]}" -r 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
        echo ${i} "lock-async tx hash:" ${ethTxHash}
        let i++
    done

    eth_block_wait 12

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(curl -ksd '{"jsonrpc":"2.0","method":"eth_getBalance","params":["'${ethAddress[i]}'", "latest"],"id":1}' http://localhost:7545 | jq -r ".result")
        res=$(gawk -M 'BEGIN{printf "%d\n", \
      '${preEthBalance[i]}' - \
      '${nowEthBalance}'}')
        echo ${i} "preBalance" ${preEthBalance[i]} "nowBalance" ${nowEthBalance} "diff" ${res}
        if [[ $res -le 100000000000000000 ]]; then
            echo -e "${RED}error number, expect greater than 100000000000000000, get ${res}${NOC}"
            exit 1
        fi
        let i++
    done
    nowChain33Balance=$(${Chain33_CLI} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]" | jq ".balance")
    diff=$(gawk -M 'BEGIN{printf "%d\n", \
    '${nowChain33Balance}' - \
    '${preChain33Balance}'}')
    check_number "${diff}" 1

}

loop_send_burn_eth() {

    #   =========================== Chain33 Burn ========================================
    echo -e "${GRE}=========== Chain33 Burn begin ===========${NOC}"

    preChain33Balance=$(${Chain33_CLI} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]" | jq ".balance")

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(curl -ksd '{"jsonrpc":"2.0","method":"eth_getBalance","params":["'${ethAddress[i]}'", "latest"],"id":1}' http://localhost:7545 | jq -r ".result")
        ethTxHash=$(${Chain33_CLI} send x2ethereum burn -a 0.1 -r ${ethAddress[i]} -t eth -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
        echo ${i} "burn chain33 tx hash:" ${ethTxHash}
        let i++
    done

    eth_block_wait 12

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(curl -ksd '{"jsonrpc":"2.0","method":"eth_getBalance","params":["'${ethAddress[i]}'", "latest"],"id":1}' http://localhost:7545 | jq -r ".result")
        res=$(gawk -M 'BEGIN{printf "%d\n", \
      '${nowEthBalance}' - \
      '${preEthBalance[i]}'}')
        echo ${i} "preBalance" ${preEthBalance[i]} "nowBalance" ${nowEthBalance} "diff" ${res}
        if [[ $res -gt 100000000000000000 ]]; then
            echo -e "${RED}error number, expect greater than 100000000000000000, get ${res}${NOC}"
            exit 1
        fi
        let i++
    done
    nowChain33Balance=$(${Chain33_CLI} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]" | jq ".balance")
    diff=$(gawk -M 'BEGIN{printf "%d\n", \
    '${preChain33Balance}' - \
    '${nowChain33Balance}'}')
    check_number "${diff}" 1

}

loop_send_lock_bty() {

    #   =========================== Chain33 Lock =========================================
    echo -e "${GRE}=========== Chain33 Lock begin ===========${NOC}"

    preChain33Balance=$(${Chain33_CLI} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum | jq -r ".balance")

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(${CLI} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddr}" | jq -r ".balance")
        ethTxHash=$(${Chain33_CLI} send x2ethereum lock -q "${tokenAddr}" -a 1 -r ${ethAddress[i]} -t bty -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
        echo ${i} "lock chain33 tx hash:" ${ethTxHash}
        let i++
    done

    eth_block_wait 12

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLI} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddr}" | jq -r ".balance")
        res=$(gawk -M 'BEGIN{printf "%d\n", \
      '${nowEthBalance}' - \
      '${preEthBalance[i]}'}')
        echo ${i} "preBalance" ${preEthBalance[i]} "nowBalance" ${nowEthBalance} "diff" ${res}
        check_number "${res}" 1
        let i++
    done
    nowChain33Balance=$(${Chain33_CLI} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum | jq -r ".balance")
    diff=$(gawk -M 'BEGIN{printf "%d\n", \
    '${preChain33Balance}' - \
    '${nowChain33Balance}'}')
    check_number "${diff}" 10

}

loop_send_burn_bty() {

    #   =========================== Ethereum Burn ========================================
    echo -e "${GRE}=========== Ethereum Burn begin ===========${NOC}"

    preChain33Balance=$(${Chain33_CLI} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum | jq -r ".balance")

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[$i]=$(${CLI} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddr}" | jq -r ".balance")
        approveTxHash=$(${CLI} relayer ethereum approve -m 1 -k "${privateKeys[i]}" -t "${tokenAddr}")
        ethTxHash=$(${CLI} relayer ethereum burn-async -m 1 -k "${privateKeys[i]}" -r 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t "${tokenAddr}")
        echo ${i} "burn-async tx hash:" ${ethTxHash}
        let i++
    done

    eth_block_wait 12

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLI} relayer ethereum balance -o "${ethAddress[i]}" -t "${tokenAddr}" | jq -r ".balance")
        res=$(gawk -M 'BEGIN{printf "%d\n", \
      '${preEthBalance[i]}' - \
      '${nowEthBalance}'}')
        echo ${i} "preBalance" ${preEthBalance[i]} "nowBalance" ${nowEthBalance} "diff" ${res}
        check_number "${res}" 1
        let i++
    done
    nowChain33Balance=$(${Chain33_CLI} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum | jq -r ".balance")
    diff=$(gawk -M 'BEGIN{printf "%d\n", \
    '${nowChain33Balance}' - \
    '${preChain33Balance}'}')
    check_number "${diff}" 10

}

loop_send_lock_erc20() {

    #    ======================== Ethereum Lock Erc20 =========================================
    echo -e "${GRE}=========== Ethereum Lock Erc20 begin ===========${NOC}"
    preChain33Balance=$(${Chain33_CLI} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t testc | jq ".res" | jq ".[]" | jq ".balance")

    i=0
    preEthBalance=$(${CLI} relayer ethereum balance -o "${Ethsender}" -t "${testcAddr}" | jq -r ".balance")

    approveTxHash=$(${CLI} relayer ethereum approve -m 10 -k "${privateKeys[8]}" -t "${testcAddr}")
    echo ${i} "lock-async erc20 approve tx hash:" ${approveTxHash}

    while [[ i -lt ${#privateKeys[@]} ]]; do
        ethTxHash=$(${CLI} relayer ethereum lock-async -m 1 -k "${privateKeys[8]}" -r 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t "${testcAddr}")
        echo ${i} "lock-async erc20 tx hash:" ${ethTxHash}
        let i++
    done

    eth_block_wait 12

    nowEthBalance=$(${CLI} relayer ethereum balance -o "${Ethsender}" -t "${testcAddr}" | jq -r ".balance")
    res=$(gawk -M 'BEGIN{printf "%d\n", \
      '${preEthBalance}' - \
      '${nowEthBalance}'}')
    echo ${i} "preBalance" ${preEthBalance} "nowBalance" ${nowEthBalance} "diff" ${res}
    check_number "${diff}" 10

    nowChain33Balance=$(${Chain33_CLI} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t testc | jq ".res" | jq ".[]" | jq ".balance")
    diff=$(gawk -M 'BEGIN{printf "%d\n", \
    '${nowChain33Balance}' - \
    '${preChain33Balance}'}')
    check_number "${diff}" 10

}

loop_send_burn_erc20() {

    #   =========================== Chain33 Burn Erc20 ========================================
    echo -e "${GRE}=========== Chain33 Burn Erc20 begin ===========${NOC}"

    preChain33Balance=$(${Chain33_CLI} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t testc | jq ".res" | jq ".[]" | jq ".balance")

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        preEthBalance[i]=$(${CLI} relayer ethereum balance -o "${ethAddress[i]}" -t "${testcAddr}" | jq -r ".balance")
        ethTxHash=$(${Chain33_CLI} send x2ethereum burn -a 1 -r ${ethAddress[i]} -t testc -q "${testcAddr}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
        echo ${i} "burn chain33 tx hash:" ${ethTxHash}
        let i++
    done

    eth_block_wait 12

    i=0
    while [[ i -lt ${#privateKeys[@]} ]]; do
        nowEthBalance=$(${CLI} relayer ethereum balance -o "${ethAddress[i]}" -t "${testcAddr}" | jq -r ".balance")
        res=$(gawk -M 'BEGIN{printf "%d\n", \
      '${nowEthBalance}' - \
      '${preEthBalance[i]}'}')
        echo ${i} "preBalance" ${preEthBalance[i]} "nowBalance" ${nowEthBalance} "diff" ${res}
        check_number "${res}" 1
        let i++
    done

    nowChain33Balance=$(${Chain33_CLI} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t testc | jq ".res" | jq ".[]" | jq ".balance")
    diff=$(gawk -M 'BEGIN{printf "%d\n", \
    '${preChain33Balance}' - \
    '${nowChain33Balance}'}')
    check_number "${diff}" 10

}

main() {
    loop_send_lock_eth
    loop_send_burn_eth
    loop_send_lock_bty
    loop_send_burn_bty
    loop_send_lock_erc20
    loop_send_burn_erc20
}

main
