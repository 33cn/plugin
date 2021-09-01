#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 只测试 lock 操作，不测试 burn

source "./publicTest.sh"
source "./relayerPublic.sh"

# ETH 部署合约者的私钥 用于部署合约时签名使用
ethDeployAddr="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
ethDeployKey="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

# chain33 部署合约者的私钥 用于部署合约时签名使用
chain33DeployAddr="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"

Chain33Cli="../../chain33-cli"
chain33BridgeBank=""
ethBridgeBank=""
chain33BtyTokenAddr="1111111111111111111114oLvT2"
ethereumYccTokenAddr=""
multisignChain33Addr=""
multisignEthAddr=""
ethBridgeToeknYccAddr=""
chain33YccErc20Addr=""

CLIA="./ebcli_A"
chain33ID=0

function set_offline_token_Bty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== chain33 端 configLockedTokenOfflineSave BTY ======${NOC}"
    #    echo '2:#配置自动转离线钱包(bty, 1000, 50%)'
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "configLockedTokenOfflineSave(${chain33BtyTokenAddr},BTY,100000000000,50)" --khainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lock_multisign_Bty_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    lock_bty_multisign 330 "330.0000" "0.0000"
    lock_bty_multisign 800 "565.0000" "565.0000"
    lock_bty_multisign 500 "532.5000" "1097.5000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function transfer_multisign_Bty_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer test
    # shellcheck disable=SC2154
    hash=$(${CLIA} chain33 multisign transfer -a 100 -r "${chain33BridgeBank}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    check_tx "${Chain33Cli}" "${hash}"
    sleep 2
    result=$(${Chain33Cli} account balance -a "${multisignChain33Addr}" -e evm)
    balance_ret "${result}" "997.5000"
    result=$(${Chain33Cli} account balance -a "${chain33BridgeBank}" -e evm)
    balance_ret "${result}" "632.5000"

    # shellcheck disable=SC2154
    hash=$(${CLIA} chain33 multisign transfer -a 100 -r "${chain33MultisignA}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    check_tx "${Chain33Cli}" "${hash}"
    sleep 2
    result=$(${Chain33Cli} account balance -a "${multisignChain33Addr}" -e evm)
    balance_ret "${result}" "897.5000"
    result=$(${Chain33Cli} account balance -a "${chain33MultisignA}" -e evm)
    balance_ret "${result}" "100.0000"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockBty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    set_offline_token_Bty
    lock_multisign_Bty_test
    transfer_multisign_Bty_test

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function set_offline_token_Chain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    echo -e "${GRE}===== chain33 端 configLockedTokenOfflineSave ERC20 YCC ======${NOC}"
    #    echo '2:#配置自动转离线钱包(YCC, 100, 60%)'
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "configLockedTokenOfflineSave(${chain33YccErc20Addr},YCC,10000000000,60)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lock_multisign_Chain33Ycc_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    lock_chain33_ycc_multisign 30 30 0
    lock_chain33_ycc_multisign 70 40 60
    lock_chain33_ycc_multisign 260 120 240
    lock_chain33_ycc_multisign 10 52 318
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function transfer_multisign_Chain33Ycc_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # transfer test
    hash=$(${CLIA} chain33 multisign transfer -a 10 -r "${chain33BridgeBank}" -t "${chain33YccErc20Addr}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    check_tx "${Chain33Cli}" "${hash}"
    sleep 2
    result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "6200000000"
    result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
    is_equal "${result}" "30800000000"

    hash=$(${CLIA} chain33 multisign transfer -a 5 -r "${chain33MultisignA}" -t "${chain33YccErc20Addr}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    check_tx "${Chain33Cli}" "${hash}"
    sleep 2
    result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${chain33MultisignA}" -b "balanceOf(${chain33MultisignA})")
    is_equal "${result}" "500000000"
    result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
    is_equal "${result}" "30300000000"

    # 判断 ETH 这端是否金额一致
    result=$(${CLIA} ethereum balance -o "${ethDeployAddr}" -t "${ethBridgeToeknYccAddr}")
    cli_ret "${result}" "balance" ".balance" "370"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockChain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    set_offline_token_Chain33Ycc
    lock_multisign_Chain33Ycc_test
    transfer_multisign_Chain33Ycc_test

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function set_offline_token_Eth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # echo '2:#配置自动转离线钱包(eth, 20, 50%)'
    result=$(${CLIA} ethereum multisign set_offline_token -s ETH -m 20)
    cli_ret "${result}" "set_offline_token -s ETH -m 20"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lock_multisign_Eth_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    lock_eth_multisign 19 19 0
    lock_eth_multisign 1 10 10
    lock_eth_multisign 16 13 23
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function transfer_multisign_Eth_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer
    # shellcheck disable=SC2154
    ${CLIA} ethereum multisign transfer -a 3 -r "${ethBridgeBank}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    sleep 2
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "16"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}")
    cli_ret "${result}" "balance" ".balance" "20"

    # transfer
    # shellcheck disable=SC2154
    ${CLIA} ethereum multisign transfer -a 5 -r "${ethMultisignA}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    sleep 2
    result=$(${CLIA} ethereum balance -o "${ethMultisignA}")
    cli_ret "${result}" "balance" ".balance" "105"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}")
    cli_ret "${result}" "balance" ".balance" "15"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockEth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock multisign ETH ======${NOC}"
    set_offline_token_Eth
    lock_multisign_Eth_test
    transfer_multisign_Eth_test
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function set_offline_token_EthYcc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # echo '2:#配置自动转离线钱包(ycc, 100, 40%)'
    result=$(${CLIA} ethereum multisign set_offline_token -s YCC -m 100 -p 40 -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "set_offline_token -s YCC -m 100"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lock_multisign_EthYcc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    lock_ethereum_ycc_multisign 70 70 0
    lock_ethereum_ycc_multisign 30 60 40
    lock_ethereum_ycc_multisign 60 72 88

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function transfer_multisign_EthYcc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer
    # multisignEthAddr 要有手续费
    ${CLIA} ethereum transfer -k "${ethDeployKey}" -m 10 -r "${multisignEthAddr}"
    sleep 2

    # transfer
    ${CLIA} ethereum multisign transfer -a 8 -r "${ethBridgeBank}" -t "${ethereumYccTokenAddr}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    sleep 2
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"

    # transfer
    ${CLIA} ethereum multisign transfer -a 10 -r "${ethMultisignA}" -t "${ethereumYccTokenAddr}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    sleep 2
    result=$(${CLIA} ethereum balance -o "${ethMultisignA}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "10"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "70"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockEthYcc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock multisign YCC ======${NOC}"
    set_offline_token_EthYcc
    lock_multisign_EthYcc
    transfer_multisign_EthYcc
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
