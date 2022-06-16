#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck disable=SC2086
# shellcheck disable=SC2154
# shellcheck disable=SC2034
# shellcheck disable=SC2219
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./mainPubilcRelayerTest.sh"
IYellow="\033[0;93m"
le8=100000000

function start_docker_ebrelayerProxy() {
    cp './relayer.toml' "./relayerproxy.toml"

    sed -i 's/^pushName=.*/pushName="x2ethproxy"/g' "./relayerproxy.toml"

    pushHost=$(get_docker_addr "${dockerNamePrefix}_ebrelayerproxy_1")
    sed -i 's/^pushHost=.*/pushHost="http:\/\/'"${pushHost}"':20000"/' "./relayerproxy.toml"
    sed -i 's/^pushBind=.*/pushBind="'"${pushHost}"':20000"/' "./relayerproxy.toml"

    # 代理转账中继器中的标志位ProcessWithDraw设置为true
    sed -i 's/^ProcessWithDraw=.*/ProcessWithDraw=true/' "./relayerproxy.toml"

    docker cp "./relayerproxy.toml" "${dockerNamePrefix}_ebrelayerproxy_1":/root/relayer.toml
    start_docker_ebrelayer "${dockerNamePrefix}_ebrelayerproxy_1" "/root/ebrelayer" "./ebrelayerproxy.log"
    sleep 1

    init_validator_relayer "${CLIP}" "${validatorPwd}" "${chain33ValidatorKeyp}" "${ethValidatorAddrKeyp}"
    sleep 20
}

function setWithdraw_ethereum() {
    result=$(${CLIP} ethereum cfgWithdraw -f 1 -s ETH -a 100 -d 18)
    cli_ret "${result}" "cfgWithdraw"
    result=$(${CLIP} ethereum cfgWithdraw -f 1 -s USDT -a 100 -d 6)
    cli_ret "${result}" "cfgWithdraw"

    # 在chain33上的bridgeBank合约中设置proxyReceiver
    ${Boss4xCLI} chain33 offline set_withdraw_proxy -c "${chain33BridgeBank}" -a "${chain33Validatorsp}" -k "${chain33DeployKey}" -n "set_withdraw_proxy:${chain33Validatorsp}"
    chain33_offline_send "set_withdraw_proxy.txt"
}

function setWithdraw_bsc() {
    result=$(${CLIP} ethereum cfgWithdraw -f 1 -s BNB -a 100 -d 18)
    cli_ret "${result}" "cfgWithdraw"
    result=$(${CLIP} ethereum cfgWithdraw -f 1 -s BUSDT -a 100 -d 6)
    cli_ret "${result}" "cfgWithdraw"

    # 在chain33上的bridgeBank合约中设置proxyReceiver
    ${Boss4xCLI} chain33 offline set_withdraw_proxy -c "${chain33BridgeBank}" -a "${chain33Validatorsp}" -k "${chain33DeployKey}" -n "set_withdraw_proxy:${chain33Validatorsp}"
    chain33_offline_send "set_withdraw_proxy.txt"
}

# eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 withdraw
function TestETH2Chain33Assets_proxy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 withdraw ===========${NOC}"

    echo -e "${IYellow} lockAmount1 $1 ${NOC}"
    local lockAmount1=$1

    echo -e "${IYellow} ethereumBridgeBank 初始金额 ${NOC}"
    ethBridgeBankBalancebf=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}" | jq -r ".balance")

    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址初始金额 ${NOC}"
    chain33RBalancebf=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址初始金额 ${NOC}"
    chain33VspBalancebf=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33Validatorsp})")

    echo -e "${IYellow} lock ${NOC}"
    result=$(${CLIP} ethereum lock -m "${lockAmount1}" -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 10

    echo -e "${IYellow} ethereumBridgeBank lock 后金额 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}")

    let ethBridgeBankBalanceEnd=ethBridgeBankBalancebf+lockAmount1
    cli_ret "${result}" "balance" ".balance" "${ethBridgeBankBalanceEnd}"

    sleep "${maturityDegree}"

    # chain33 chain33MainBridgeTokenAddr（ETH合约中）查询 lock 金额
    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址 lock 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    let chain33RBalancelock=lockAmount1*le8+chain33RBalancebf
    is_equal "${result}" "${chain33RBalancelock}"

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址 lock 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33Validatorsp})")
    is_equal "${result}" "${chain33VspBalancebf}"

    echo -e "${IYellow} ethTestAddr2 ethereum withdraw 接收地址初始金额 ${NOC}"
    ethT2Balancebf=$(${CLIP} ethereum balance -o "${ethTestAddr2}" | jq -r ".balance")

    echo -e "${IYellow} ethValidatorAddrp ethereum 代理地址初始金额 ${NOC}"
    ethPBalancebf=$(${CLIP} ethereum balance -o "${ethValidatorAddrp}" | jq -r ".balance")

    echo -e "${IYellow} withdraw ${NOC}"
    result=$(${CLIP} chain33 withdraw -m "${lockAmount1}" -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33MainBridgeTokenAddr}")
    cli_ret "${result}" "withdraw"

    sleep "${maturityDegree}"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "${ethBridgeBankBalanceEnd}"

    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址 withdraw 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "${chain33RBalancebf}"

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址 withdraw 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33Validatorsp})")
    let chain33VspBalancewithdraw=lockAmount1*le8+chain33VspBalancebf
    is_equal "${result}" "${chain33VspBalancewithdraw}"

    echo -e "${IYellow} ethTestAddr2 ethereum withdraw 接收地址 withdraw 后金额 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethTestAddr2}" | jq -r ".balance")
    ethT2BalanceEnd=$(echo "${ethT2Balancebf}+${lockAmount1}-1" | bc)
    is_equal "${result}" "${ethT2BalanceEnd}"

    echo -e "${IYellow} ethValidatorAddrp ethereum 代理地址 withdraw 后金额 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethValidatorAddrp}" | jq -r ".balance")

    if [[ $(echo "${ethPBalancebf}-${lockAmount1}+1 < $result" | bc) == 1 ]]; then
        echo -e "${RED}error $ethPBalanceEnd 小于 $result, 应该大于 $ethPBalanceEnd 扣了一点点手续费 ${NOC}"
        exit 1
    fi

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 withdraw
function TestETH2Chain33Assets_proxy_excess() {
    echo -e "${GRE}=========== $FUNCNAME 超额 begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 withdraw ===========${NOC}"

    echo -e "${IYellow} lockAmount1 $1 ${NOC}"
    local lockAmount1=$1

    echo -e "${IYellow} ethereumBridgeBank 初始金额 ${NOC}"
    ethBridgeBankBalancebf=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}" | jq -r ".balance")

    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址初始金额 ${NOC}"
    chain33RBalancebf=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址初始金额 ${NOC}"
    chain33VspBalancebf=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33Validatorsp})")

    echo -e "${IYellow} lock ${NOC}"
    result=$(${CLIP} ethereum lock -m "${lockAmount1}" -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 10

    echo -e "${IYellow} ethereumBridgeBank lock 后金额 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}")
    let ethBridgeBankBalanceEnd=ethBridgeBankBalancebf+lockAmount1
    cli_ret "${result}" "balance" ".balance" "${ethBridgeBankBalanceEnd}"

    sleep "${maturityDegree}"

    # chain33 chain33MainBridgeTokenAddr（ETH合约中）查询 lock 金额
    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址 lock 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    let chain33RBalancelock=lockAmount1*le8+chain33RBalancebf
    is_equal "${result}" "${chain33RBalancelock}"

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址 lock 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33Validatorsp})")
    is_equal "${result}" "${chain33VspBalancebf}"

    echo -e "${IYellow} ethTestAddr2 ethereum withdraw 接收地址初始金额 ${NOC}"
    ethT2Balancebf=$(${CLIP} ethereum balance -o "${ethTestAddr2}" | jq -r ".balance")

    echo -e "${IYellow} ethValidatorAddrp ethereum 代理地址初始金额 ${NOC}"
    ethPBalancebf=$(${CLIP} ethereum balance -o "${ethValidatorAddrp}" | jq -r ".balance")

    echo -e "${IYellow} withdraw ${NOC}"
    result=$(${CLIP} chain33 withdraw -m "${lockAmount1}" -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33MainBridgeTokenAddr}")
    cli_ret "${result}" "withdraw"

    sleep "${maturityDegree}"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "${ethBridgeBankBalanceEnd}"

    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址 withdraw 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "${chain33RBalancebf}"

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址 withdraw 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33Validatorsp})")
    let chain33VspBalancewithdraw=lockAmount1*le8+chain33VspBalancebf
    is_equal "${result}" "${chain33VspBalancewithdraw}"

    echo -e "${IYellow} ethTestAddr2 ethereum withdraw 接收地址 withdraw 后金额 超额了金额跟之前一样${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethTestAddr2}" | jq -r ".balance")
    is_equal "${result}" "${ethT2Balancebf}"

    echo -e "${IYellow} ethValidatorAddrp ethereum 代理地址 withdraw 后金额 超额了金额跟之前一样${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethValidatorAddrp}" | jq -r ".balance")
    is_equal "${result}" "${ethPBalancebf}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33USDT_proxy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 USDT 资产,然后在 chain33 上 withdraw ===========${NOC}"

    echo -e "${IYellow} lockAmount1 $1 ${NOC}"
    local lockAmount1=$1

    echo -e "${IYellow} ethereumBridgeBank 初始金额 ${NOC}"
    ethBridgeBankBalancebf=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")

    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址初始金额 ${NOC}"
    chain33RBalancebf=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址初始金额 ${NOC}"
    chain33VspBalancebf=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33Validatorsp})")

    echo -e "${IYellow} ETH 这端 lock $lockAmount1 个 USDT ${NOC}"
    result=$(${CLIP} ethereum lock -m "${lockAmount1}" -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 10

    echo -e "${IYellow} 查询 ETH 这端 ethereumBridgeBank lock 后金额 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")

    let ethBridgeBankBalanceEnd=ethBridgeBankBalancebf+lockAmount1
    cli_ret "${result}" "balance" ".balance" "${ethBridgeBankBalanceEnd}"

    sleep "${maturityDegree}"

    # chain33 chain33MainBridgeTokenAddr（ETH合约中）查询 lock 金额
    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址 lock 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")

    let chain33RBalancelock=lockAmount1*le8+chain33RBalancebf
    is_equal "${result}" "${chain33RBalancelock}"

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址 lock 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33Validatorsp})")
    is_equal "${result}" "${chain33VspBalancebf}"

    echo -e "${IYellow} ethTestAddr2 ethereum withdraw 接收地址初始金额 ${NOC}"
    ethT2Balancebf=$(${CLIP} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")

    echo -e "${IYellow} ethValidatorAddrp ethereum 代理地址初始金额 ${NOC}"
    ethPBalancebf=$(${CLIP} ethereum balance -o "${ethValidatorAddrp}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")

    echo -e "${IYellow} withdraw ${NOC}"
    result=$(${CLIP} chain33 withdraw -m "${lockAmount1}" -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33USDTBridgeTokenAddr}")
    cli_ret "${result}" "withdraw"

    sleep "${maturityDegree}"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "${ethBridgeBankBalanceEnd}"

    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址 withdraw 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "${chain33RBalancebf}"

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址 withdraw 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33Validatorsp})")

    let chain33VspBalancewithdraw=lockAmount1*le8+chain33VspBalancebf
    is_equal "${result}" "${chain33VspBalancewithdraw}"

    echo -e "${IYellow} ethTestAddr2 ethereum withdraw 接收地址 withdraw 后金额 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")
    ethT2BalanceEnd=$(echo "${ethT2Balancebf}+${lockAmount1}-1" | bc)
    is_equal "${result}" "${ethT2BalanceEnd}"

    echo -e "${IYellow} ethValidatorAddrp ethereum 代理地址 withdraw 后金额 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethValidatorAddrp}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")

    let ethPBalanceEnd=ethPBalancebf-lockAmount1+1
    is_equal "${result}" "${ethPBalanceEnd}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33USDT_proxy_excess() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 USDT 资产,然后在 chain33 上 withdraw ===========${NOC}"

    echo -e "${IYellow} lockAmount1 $1 ${NOC}"
    local lockAmount1=$1

    echo -e "${IYellow} ethereumBridgeBank 初始金额 ${NOC}"
    ethBridgeBankBalancebf=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")

    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址初始金额 ${NOC}"
    chain33RBalancebf=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址初始金额 ${NOC}"
    chain33VspBalancebf=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33Validatorsp})")

    echo -e "${IYellow} ETH 这端 lock $lockAmount1 个 USDT ${NOC}"
    result=$(${CLIP} ethereum lock -m "${lockAmount1}" -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 10

    echo -e "${IYellow} 查询 ETH 这端 ethereumBridgeBank lock 后金额 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")

    let ethBridgeBankBalanceEnd=ethBridgeBankBalancebf+lockAmount1
    cli_ret "${result}" "balance" ".balance" "${ethBridgeBankBalanceEnd}"

    sleep "${maturityDegree}"

    # chain33 chain33MainBridgeTokenAddr（ETH合约中）查询 lock 金额
    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址 lock 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")

    let chain33RBalancelock=lockAmount1*le8+chain33RBalancebf
    is_equal "${result}" "${chain33RBalancelock}"

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址 lock 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33Validatorsp})")
    is_equal "${result}" "${chain33VspBalancebf}"

    echo -e "${IYellow} ethTestAddr2 ethereum withdraw 接收地址初始金额 ${NOC}"
    ethT2Balancebf=$(${CLIP} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")

    echo -e "${IYellow} ethValidatorAddrp ethereum 代理地址初始金额 ${NOC}"
    ethPBalancebf=$(${CLIP} ethereum balance -o "${ethValidatorAddrp}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")

    echo -e "${IYellow} withdraw ${NOC}"
    result=$(${CLIP} chain33 withdraw -m "${lockAmount1}" -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33USDTBridgeTokenAddr}")
    cli_ret "${result}" "withdraw"

    sleep "${maturityDegree}"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIP} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "${ethBridgeBankBalanceEnd}"

    echo -e "${IYellow} chain33ReceiverAddr chain33 端 lock 后接收地址 withdraw 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "${chain33RBalancebf}"

    echo -e "${IYellow} chain33Validatorsp chain33 代理地址 withdraw 后金额 ${NOC}"
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33Validatorsp})")

    let chain33VspBalancewithdraw=lockAmount1*le8+chain33VspBalancebf
    is_equal "${result}" "${chain33VspBalancewithdraw}"

    echo -e "${IYellow} ethTestAddr2 ethereum withdraw 接收地址 withdraw 后金额  超额了金额跟之前一样 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")
    is_equal "${result}" "${ethT2Balancebf}"

    echo -e "${IYellow} ethValidatorAddrp ethereum 代理地址 withdraw 后金额  超额了金额跟之前一样 ${NOC}"
    result=$(${CLIP} ethereum balance -o "${ethValidatorAddrp}" -t "${ethereumUSDTERC20TokenAddr}" | jq -r ".balance")
    is_equal "${result}" "${ethPBalancebf}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestProxy() {
    # 增加一笔转帐交易测试 nonce 不一致结果是否成功
    ${CLIP} ethereum transfer -k "${ethValidatorAddrKeyp}" -m 10 -r "${ethereumMultisignAddr}"
    TestETH2Chain33Assets_proxy 20
    TestETH2Chain33Assets_proxy 30
    TestETH2Chain33Assets_proxy_excess 100

    ${CLIP} ethereum token token_transfer -k "${ethTestAddrKey1}" -m 500 -r "${ethValidatorAddrp}" -t "${ethereumUSDTERC20TokenAddr}"
    TestETH2Chain33USDT_proxy 20
    TestETH2Chain33USDT_proxy 40
    TestETH2Chain33USDT_proxy_excess 100
}

function TestRelayerProxy() {
    start_docker_ebrelayerProxy

    Boss4xCLI=${Boss4xCLIeth}
    CLIP=${CLIPeth}
    ethereumBridgeBank="${ethereumBridgeBankOnETH}"
    chain33MainBridgeTokenAddr="${chain33MainBridgeTokenAddrETH}"
    ethereumUSDTERC20TokenAddr="${ethereumUSDTERC20TokenAddrOnETH}"
    chain33USDTBridgeTokenAddr="${chain33USDTBridgeTokenAddrOnETH}"
    setWithdraw_ethereum
    TestProxy

    Boss4xCLI=${Boss4xCLIbsc}
    CLIP=${CLIPbsc}
    ethereumBridgeBank="${ethereumBridgeBankOnBSC}"
    chain33MainBridgeTokenAddr="${chain33MainBridgeTokenAddrBNB}"
    ethereumUSDTERC20TokenAddr="${ethereumUSDTERC20TokenAddrOnBSC}"
    chain33USDTBridgeTokenAddr="${chain33USDTBridgeTokenAddrOnBSC}"
    setWithdraw_bsc
    TestProxy
}

function parallel_test() {
    Boss4xCLI=${Boss4xCLIeth}
    CLIA=${CLIAeth}
    ethereumBridgeBank="${ethereumBridgeBankOnETH}"
    ethereumMultisignAddr="${ethereumMultisignAddrOnETH}"
    chain33MainBridgeTokenAddr="${chain33MainBridgeTokenAddrETH}"
    ethereumBtyBridgeTokenAddr="${ethereumBtyBridgeTokenAddrOnETH}"
    ethereumUSDTERC20TokenAddr="${ethereumUSDTERC20TokenAddrOnETH}"
    chain33USDTBridgeTokenAddr="${chain33USDTBridgeTokenAddrOnETH}"

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
    #    cli_ret "${result}" "balance" ".balance" "0"

    for ((i = 0; i < 10; i++)); do
        parallel --jobs 20 ${CLIA} ethereum lock -m 0.005 -k {} -r "${chain33ReceiverAddr}" ::: 0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697 0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2 0x1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695 0x4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf 0x62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9 0x355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71 0x72dff1c863631208a3d4f67a5fb0b7ebe69f827a75f332e2069dc7c825cb2202 0xecbc20b02e1ffd321e31c2a6d7d35a69715ba43ef2b0048a27de4f67b8249bde 0x1649955b3f2852a9cd71e50237b5a6f717539cffbe336bfcd95eb19a1b5c6f1b 0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd 0x4bbe8e10a0987b6d56c505dfe41f74d357ac8211e91c4a179f9c5e38c181aaf0 0xc2d896509c89c56d2365f40c5da5680174314884118fea024013818f792fcd64 0x56e4f5884dbd4248e649bb5163815dd6fcae7a656f43419becd183e49fe2b514 0x6ae0d6e5f14c1719f170100b6a84f3d7be14c2623404191d854ef98739e813a6 0x0e8b158dae56bc58c69ea9cac8737317dc47d2fb50d40bea4f1629414a8f7846 0x3d5ac5f963568544b8b855c8950030ec60cd0a0e9a5a8be725e790c625404fcd 0x4cca1a474c3a9789fd11987d19ae13a3553c47c7b441e77739af5b2d48f01371 0x8f23fc9a39795a1687e4f756312015d96b448829c32114981038ea74438361a0 0xbf667b0227522bd3fd220e748c43d5529e323eca9bcb005ac58e6f2431577b28 0x2e81ad3c49763fa9c0ea79b27e6cf584e6e073b43524c151ec1ade03706c1a00
    done

    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
    #    cli_ret "${result}" "balance" ".balance" "1"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    for ((i = 0; i < 10; i++)); do
        parallel --jobs 20 ${CLIA} ethereum lock -m 0.005 -k {} -r "${chain33ReceiverAddr}" ::: 0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697 0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2 0x1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695 0x4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf 0x62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9 0x355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71 0x72dff1c863631208a3d4f67a5fb0b7ebe69f827a75f332e2069dc7c825cb2202 0xecbc20b02e1ffd321e31c2a6d7d35a69715ba43ef2b0048a27de4f67b8249bde 0x1649955b3f2852a9cd71e50237b5a6f717539cffbe336bfcd95eb19a1b5c6f1b 0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd 0x4bbe8e10a0987b6d56c505dfe41f74d357ac8211e91c4a179f9c5e38c181aaf0 0xc2d896509c89c56d2365f40c5da5680174314884118fea024013818f792fcd64 0x56e4f5884dbd4248e649bb5163815dd6fcae7a656f43419becd183e49fe2b514 0x6ae0d6e5f14c1719f170100b6a84f3d7be14c2623404191d854ef98739e813a6 0x0e8b158dae56bc58c69ea9cac8737317dc47d2fb50d40bea4f1629414a8f7846 0x3d5ac5f963568544b8b855c8950030ec60cd0a0e9a5a8be725e790c625404fcd 0x4cca1a474c3a9789fd11987d19ae13a3553c47c7b441e77739af5b2d48f01371 0x8f23fc9a39795a1687e4f756312015d96b448829c32114981038ea74438361a0 0xbf667b0227522bd3fd220e748c43d5529e323eca9bcb005ac58e6f2431577b28 0x2e81ad3c49763fa9c0ea79b27e6cf584e6e073b43524c151ec1ade03706c1a00
        parallel --jobs 20 ${CLIA} chain33 burn_increase -m {} -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33MainBridgeTokenAddr}" ::: 0.003 0.004 0.005 0.006 0.007 0.003 0.004 0.005 0.006 0.007 0.003 0.004 0.005 0.006 0.007 0.003 0.004 0.005 0.006 0.007
    done

    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
    #    cli_ret "${result}" "balance" ".balance" "1"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
}
