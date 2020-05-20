#!/usr/bin/env bash
set -x
# 只启动 eth 这一端的测试
# 只启动一个 relayer 第一个地址权重设置超过2/3

source "./publicTest.sh"

CLI="../build/ebcli_A"

chain33SenderAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
chain33SenderAddrKey="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"

# 0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f
ethValidatorAddrKey="3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"

ethReceiverAddr1="0xa4ea64a583f6e51c3799335b28a8f0529570a635"
ethReceiverAddrKey1="355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"

ethReceiverAddr2="0x0c05ba5c230fdaa503b53702af1962e08d0c60bf"
ethReceiverAddrKey2="9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"

ethReceiverAddr3="0x1919203bA8b325278d28Fb8fFeac49F2CD881A4e"
#ethReceiverAddrKey3="62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9"

prophecyTx0="0x112260c98aec81b3e235af47c355db720f60e751cce100fed6f334e1b1530bde"
prophecyTx1="0x222260c98aec81b3e235af47c355db720f60e751cce100fed6f334e1b1530bde"
prophecyTx2="0x332260c98aec81b3e235af47c355db720f60e751cce100fed6f334e1b1530bde"
prophecyTx3="0x442260c98aec81b3e235af47c355db720f60e751cce100fed6f334e1b1530bde"
prophecyTx4="0x552260c98aec81b3e235af47c355db720f60e751cce100fed6f334e1b1530bde"
prophecyTx5="0x662260c98aec81b3e235af47c355db720f60e751cce100fed6f334e1b1530bde"
prophecyTx6="0x772260c98aec81b3e235af47c355db720f60e751cce100fed6f334e1b1530bde"

InitAndDeploy() {
    echo "=========== $FUNCNAME begin ==========="

    result=$(${CLI} relayer set_pwd -n 123456hzj -o kk)
    cli_ret "${result}" "set_pwd"

    result=$(${CLI} relayer unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    result=$(${CLI} relayer ethereum deploy)
    cli_ret "${result}" "deploy"

    result=$(${CLI} relayer ethereum import_chain33privatekey -k "${chain33SenderAddrKey}")
    cli_ret "${result}" "import_chain33privatekey"

    result=$(${CLI} relayer ethereum import_ethprivatekey -k "${ethValidatorAddrKey}")
    cli_ret "${result}" "import_ethprivatekey"

    echo "=========== $FUNCNAME end ==========="
}

# eth to chain33
# 在以太坊上锁定资产,然后在 chain33 上铸币,针对 erc20 资产
# 以太坊 brun 资产,balance 对比是否正确
TestETH2Chain33Erc20() {
    echo "=========== $FUNCNAME begin ==========="

    ${CLI} relayer unlock -p 123456hzj
    # token4erc20 在 chain33 上先有 token,同时 mint
    tokenSymbol="testc"
    result=$(${CLI} relayer ethereum token4erc20 -s "${tokenSymbol}")
    tokenAddr=$(cli_ret "${result}" "token4erc20" ".addr")

    # 先铸币 1000
    result=$(${CLI} relayer ethereum mint -m 1000 -o "${ethReceiverAddr1}" -t "${tokenAddr}")
    cli_ret "${result}" "mint"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "1000"

    result=$(${CLI} relayer ethereum bridgeBankAddr)
    bridgeBankAddr=$(cli_ret "${result}" "bridgeBankAddr" ".addr")

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 2 chain33 lock 前先审批一下
    result=$(${CLI} relayer ethereum approve -m 100 -k "${ethReceiverAddrKey1}" -t "${tokenAddr}")
    cli_ret "${result}" "approve"

    # ETH 2 chain33 lock 100
    # -r chain33 receiver addr
    result=$(${CLI} relayer ethereum lock -m 100 -k "${ethReceiverAddrKey1}" -r "${chain33SenderAddr}" -t "${tokenAddr}")
    cli_ret "${result}" "lock"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "900"

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "100"

    # ETH 2 chain33 withdraw 40
    # -c 1 burn 40
    result=$(${CLI} relayer ethereum prophecy -i "${prophecyTx0}" -m 40 -a "${chain33SenderAddr}" -c 1 -r "${ethReceiverAddr2}" -s "${tokenSymbol}" -t "${tokenAddr}")
    cli_ret "${result}" "prophecy -m 40"

    walitProphecyFinish "${ethReceiverAddr2}" "${tokenAddr}" 40

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "40"

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "60"

    # burn 60
    result=$(${CLI} relayer ethereum prophecy -i "${prophecyTx1}" -m 60 -a "${chain33SenderAddr}" -c 1 -r "${ethReceiverAddr2}" -s "${tokenSymbol}" -t "${tokenAddr}")
    cli_ret "${result}" "prophecy -m 60"

    walitProphecyFinish "${ethReceiverAddr2}" "${tokenAddr}" 100

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "100"

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo "=========== $FUNCNAME end ==========="
}

TestETH2Chain33Erc20_err() {
    echo "=========== $FUNCNAME begin ==========="

    ${CLI} relayer unlock -p 123456hzj
    # token4erc20 在 chain33 上先有 token,同时 mint
    tokenSymbol="errc"
    result=$(${CLI} relayer ethereum token4erc20 -s "${tokenSymbol}")
    tokenAddr=$(cli_ret "${result}" "token4erc20" ".addr")

    # 先铸币 1000
    result=$(${CLI} relayer ethereum mint -m 1000 -o "${ethReceiverAddr1}" -t "${tokenAddr}")
    cli_ret "${result}" "mint"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "1000"

    result=$(${CLI} relayer ethereum bridgeBankAddr)
    bridgeBankAddr=$(cli_ret "${result}" "bridgeBankAddr" ".addr")

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 2 chain33 lock 前先审批一下
    result=$(${CLI} relayer ethereum approve -m 100 -k "${ethReceiverAddrKey1}" -t "${tokenAddr}")
    cli_ret "${result}" "approve"

    # lock 200 err
    result=$(${CLI} relayer ethereum lock -m 200 -k "${ethReceiverAddrKey1}" -r "${chain33SenderAddr}" -t "${tokenAddr}")
    cli_ret_err "${result}"

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "0"

    # lock 1100 err
    result=$(${CLI} relayer ethereum approve -m 1100 -k "${ethReceiverAddrKey1}" -t "${tokenAddr}")
    #cli_ret "${result}" "approve"
    result=$(${CLI} relayer ethereum lock -m 1100 -k "${ethReceiverAddrKey1}" -r "${chain33SenderAddr}" -t "${tokenAddr}")
    cli_ret_err "${result}"

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "0"

    result=$(${CLI} relayer ethereum approve -m 300 -k "${ethReceiverAddrKey1}" -t "${tokenAddr}")
    cli_ret "${result}" "approve"

    # ETH 2 chain33 lock 100
    # -r chain33 receiver addr
    result=$(${CLI} relayer ethereum lock -m 300 -k "${ethReceiverAddrKey1}" -r "${chain33SenderAddr}" -t "${tokenAddr}")
    cli_ret "${result}" "lock"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "700"

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "300"

    # -c 1 burn 500
    result=$(${CLI} relayer ethereum prophecy -i "${prophecyTx2}" -m 500 -a "${chain33SenderAddr}" -c 1 -r "${ethReceiverAddr2}" -s "${tokenSymbol}" -t "${tokenAddr}")
    #cli_ret "${result}" "prophecy -m 40"

    sleep 15

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "0"

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "300"

    echo "=========== $FUNCNAME end ==========="
}

# eth to chain33
# 在以太坊上锁定资产,然后在 chain33 上铸币,针对 eth 资产
TestETH2Chain33Assets() {
    echo "=========== $FUNCNAME begin ==========="
    ${CLI} relayer unlock -p 123456hzj

    result=$(${CLI} relayer ethereum bridgeBankAddr)
    bridgeBankAddr=$(cli_ret "${result}" "bridgeBankAddr" ".addr")

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # chain33 lock eth
    # lock 100
    result=$(${CLI} relayer ethereum lock -m 100 -k "${ethReceiverAddrKey1}" -r "${chain33SenderAddr}")
    cli_ret "${result}" "lock"

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "100"

    # ETH 2 chain33 withdraw 40
    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}")
    balance3=$(cli_ret "${result}" "balance" ".balance")

    result=$(${CLI} relayer ethereum prophecy -i "${prophecyTx3}" -m 40 -a "${chain33SenderAddr}" -c 1 -r "${ethReceiverAddr2}" -s eth)
    cli_ret "${result}" "prophecy -m 40"

    walitProphecyFinish "${ethReceiverAddr2}" $((balance3 + 40))

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}")
    balance4=$(cli_ret "${result}" "balance" ".balance")

    echo "${balance3} ${balance4}"
    if [[ ${balance4} != $((balance3 + 40)) ]]; then
        echo "wrong balance"
        exit 1
    fi

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "60"

    # ETH 2 chain33 withdraw 110 error
    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}")
    balance3=$(cli_ret "${result}" "balance" ".balance")

    result=$(${CLI} relayer ethereum prophecy -i "${prophecyTx4}" -m 110 -a "${chain33SenderAddr}" -c 1 -r "${ethReceiverAddr2}" -s eth)
    #cli_ret "${result}" "prophecy -m 110"

    sleep 15

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}")
    balance4=$(cli_ret "${result}" "balance" ".balance")

    echo "${balance3} ${balance4}"
    if [[ ${balance4} != "${balance4}" ]]; then
        echo "wrong balance"
        exit 1
    fi

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "60"

    # ETH 2 chain33 withdraw 60
    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}")
    balance3=$(cli_ret "${result}" "balance" ".balance")

    result=$(${CLI} relayer ethereum prophecy -i "${prophecyTx5}" -m 60 -a "${chain33SenderAddr}" -c 1 -r "${ethReceiverAddr2}" -s eth)
    cli_ret "${result}" "prophecy -m 60"

    wait_prophecy_finish "${CLI}" "${ethReceiverAddr2}" $((balance3 + 60))

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}")
    balance4=$(cli_ret "${result}" "balance" ".balance")

    echo "${balance3} ${balance4}"
    if [[ ${balance4} != $((balance3 + 60)) ]]; then
        echo "wrong balance"
        exit 1
    fi

    result=$(${CLI} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo "=========== $FUNCNAME end ==========="
}

# chain33 to eth
# 在 chain33 上锁定资产,然后在以太坊上铸币
# chain33 brun 资产,balance 对比是否正确
TestChain33ToEthAssets() {
    echo "=========== $FUNCNAME begin ==========="
    result=$(${CLI} relayer unlock -p 123456hzj)
    # token4chain33 在 以太坊 上先有 bty
    result=$(${CLI} relayer ethereum token4chain33 -s coins.bty)
    tokenAddr=$(cli_ret "${result}" "token4chain33" ".addr")

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr1}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "0"

    # -c 2 chain33 lock 100
    result=$(${CLI} relayer ethereum prophecy -i "${prophecyTx6}" -m 100 -a "${chain33SenderAddr}" -c 2 -r "${ethReceiverAddr1}" -s bty -t "${tokenAddr}")
    cli_ret "${result}" "prophecy -m 100"

    walitProphecyFinish "${ethReceiverAddr1}" "${tokenAddr}" "100"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr1}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "100"

    # transfer 10
    result=$(${CLI} relayer ethereum transfer -m 10 -k "${ethReceiverAddrKey1}" -r "${ethReceiverAddr2}" -t "${tokenAddr}")
    cli_ret "${result}" "transfer"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr1}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "90"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "10"

    result=$(${CLI} relayer ethereum transfer -m 10 -k "${ethReceiverAddrKey2}" -r "${ethReceiverAddr3}" -t "${tokenAddr}")
    cli_ret "${result}" "transfer"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr2}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "0"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr3}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "10"

    result=$(${CLI} relayer ethereum transfer -m 10 -k "${ethReceiverAddrKey2}" -r "${ethReceiverAddr3}" -t "${tokenAddr}")
    cli_ret_err "${result}"

    result=$(${CLI} relayer ethereum transfer -m 200 -k "${ethReceiverAddrKey1}" -r "${ethReceiverAddr2}" -t "${tokenAddr}")
    cli_ret_err "${result}"

    # brun 90
    result=$(${CLI} relayer ethereum burn -m 90 -k "${ethReceiverAddrKey1}" -r "${chain33SenderAddr}" -t "${tokenAddr}")
    cli_ret "${result}" "burn"

    result=$(${CLI} relayer ethereum balance -o "${ethReceiverAddr1}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "0"

    # brun 10 err
    result=$(${CLI} relayer ethereum burn -m 10 -k "${ethReceiverAddrKey1}" -r "${chain33SenderAddr}" -t "${tokenAddr}")
    cli_ret_err "${result}"

    echo "=========== $FUNCNAME end ==========="
}

main() {
    # 修改配置文件权重
    kill_ebrelayer "./../build/ebrelayer"
    rm -rf "./../build/datadir"
    sed -i 's/initPowers=\[25, 25, 25, 25\]/initPowers=\[925, 25, 25, 25\]/g' "./../build/relayer.toml"
    start_ebrelayer "./../build/ebrelayer" "./../build/ebrelayer.log"

    InitAndDeploy
    TestETH2Chain33Erc20
    TestETH2Chain33Erc20_err
    TestETH2Chain33Assets
    TestChain33ToEthAssets

    kill_ebrelayer "./../build/ebrelayer"
}
main
