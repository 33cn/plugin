#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./mainPubilcRelayerTest.sh"

function start_docker_ebrelayerProxy() {
    # shellcheck disable=SC2154
    cp './relayer.toml' "./relayerproxy.toml"

    # 删除配置文件中不需要的字段
    for deleteName in "deploy4chain33" "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers" "deploy" "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers"; do
        delete_line "./relayerproxy.toml" "${deleteName}"
    done

    pushNameChange "./relayerproxy.toml"

    # 代理转账中继器中的标志位ProcessWithDraw设置为true
    sed -i 's/^ProcessWithDraw=.*/ProcessWithDraw=true/g' "./relayerproxy.toml"

    # shellcheck disable=SC2154
    docker cp "./relayerproxy.toml" "${dockerNamePrefix}_ebrelayerproxy_1":/root/relayer.toml
    start_docker_ebrelayer "${dockerNamePrefix}_ebrelayerproxy_1" "/root/ebrelayer" "./ebrelayerproxy.log"
    sleep 1

    # shellcheck disable=SC2154
    init_validator_relayer "${CLIP}" "${validatorPwd}" "${chain33ValidatorKeyp}" "${ethValidatorAddrKeyp}"
}

#strle18="000000000000000000"
#strle6="000000"

function setWithdraw() {
    result=$(${CLIP} ethereum cfgWithdraw -f 1 -s ETH -a "500000000000000000000")
    cli_ret "${result}" "cfgWithdraw"
    result=$(${CLIP} ethereum cfgWithdraw -f 1 -s USDT -a "500000000")
    cli_ret "${result}" "cfgWithdraw"

    # 在chain33上的bridgeBank合约中设置proxyReceiver
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline set_withdraw_proxy -c "${chain33BridgeBank}" -a "${chain33Validatorsp}" -k "${chain33DeployKey}" -n "set_withdraw_proxy:${chain33Validatorsp}"
    chain33_offline_send "set_withdraw_proxy.txt"
}

# eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn
function TestETH2Chain33Assets_proxy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn ===========${NOC}"
    # shellcheck disable=SC2154
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0"

    # shellcheck disable=SC2154
    result=$(${CLIA} ethereum lock -m 0.002 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0.002"

    # shellcheck disable=SC2086
    sleep "${maturityDegree}"

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    # shellcheck disable=SC2154
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    #    is_equal "${result}" "2000000000000000"

    # 原来的数额
    # shellcheck disable=SC2154
    result=$(${CLIA} ethereum balance -o "${ethTestAddr2}")

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum'
    # shellcheck disable=SC2154
    result=$(${CLIA} chain33 withdraw -m 0.0003 -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33EthBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep "${maturityDegree}"

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    #    is_equal "${result}" "1700000000000000"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0.0017"

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum 6'
    result=$(${CLIA} chain33 burn -m 0.0017 -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33EthBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep "${maturityDegree}"

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
#    is_equal "${result}" "0"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0"

    result=$(${CLIA} ethereum balance -o "${ethTestAddr2}")

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33USDT_proxy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 USDT 资产,然后在 chain33 上 burn ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    # shellcheck disable=SC2154
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 12个 USDT
    result=$(${CLIA} ethereum lock -m 12 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    # 查询 ETH 这端 bridgeBank 地址 12 USDT
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "12"

    sleep "${maturityDegree}"

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    # shellcheck disable=SC2154
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 12 * le8
    is_equal "${result}" "1200000000"

    # 原来的数额 0
    # shellcheck disable=SC2154
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 5 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33USDTBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep "${maturityDegree}"

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 12-5 * le8
    is_equal "${result}" "700000000"

    # 查询 ETH 这端 bridgeBank 地址 7
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "7"

    # 更新后的金额 5
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "5"

    echo '#5.burn USDT from Chain33 USDT(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 7 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33USDTBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep "${maturityDegree}"

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "0"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 更新后的金额 12
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "12"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestRelayerProxy() {
    start_docker_ebrelayerProxy
    setWithdraw

    TestETH2Chain33Assets_proxy
#    TestETH2Chain33USDT_proxy
}

function AllRelayerMainTest() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    set +e

    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    # shellcheck disable=SC2120
    if [[ $# -ge 2 ]]; then
        # shellcheck disable=SC2034
        chain33ID="${2}"
    fi

    get_cli

    # init
    # shellcheck disable=SC2154
    # shellcheck disable=SC2034
    Chain33Cli=${MainCli}
    InitChain33Validator
    # para add
    initPara

    StartDockerRelayerDeploy
    test_all

    TestRelayerProxy
    
    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
