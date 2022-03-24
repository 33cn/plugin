#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./mainPubilcRelayerTest.sh"
source "./proxyVerifyTest.sh"

# shellcheck disable=SC2154
function StartDockerRelayerDeploy_onlyETH() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # 修改 relayer.toml
    up_relayer_toml

    # 删除行
    sed -i "16,23"'d' "./relayer.toml"

    # 启动 ebrelayer
    start_docker_ebrelayerA

    docker cp "./deploy_chain33.toml" "${dockerNamePrefix}_ebrelayera_1":/root/deploy_chain33.toml
    docker cp "./deploy_ethereum.toml" "${dockerNamePrefix}_ebrelayera_1":/root/deploy_ethereum.toml

    # 部署合约 设置 bridgeRegistry 地址
    OfflineDeploy_chain33
    # 修改 relayer.toml 字段
    sed -i 's/^BridgeRegistryOnChain33=.*/BridgeRegistryOnChain33="'"${chain33BridgeRegistry}"'"/g' "./relayer.toml"

    # shellcheck disable=SC2154
    # shellcheck disable=SC2034
    {
        Boss4xCLI=${Boss4xCLIeth}
        CLIA=${CLIAeth}
        OfflineDeploy_ethereum "./deploy_ethereum.toml"
        ethereumBridgeBankOnETH="${ethereumBridgeBank}"
        ethereumBridgeRegistryOnETH="${ethereumBridgeRegistry}"
        ethereumMultisignAddrOnETH="${ethereumMultisignAddr}"

        sed -i '12,18s/BridgeRegistry=.*/BridgeRegistry="'"${ethereumBridgeRegistryOnETH}"'"/g' "./relayer.toml"
    }

    # 向离线多签地址打点手续费
    Chain33Cli=${MainCli}
    initMultisignChain33Addr
    transferChain33MultisignFee
    Chain33Cli=${Para8901Cli}

    docker cp "./relayer.toml" "${dockerNamePrefix}_ebrelayera_1":/root/relayer.toml
    InitRelayerA

    # 设置 token 地址
    # shellcheck disable=SC2154
    # shellcheck disable=SC2034
    {

        Boss4xCLI=${Boss4xCLIeth}
        CLIA=${CLIAeth}
        ethereumBridgeBank="${ethereumBridgeBankOnETH}"
        offline_create_bridge_token_chain33_symbol "USDT"
        chain33USDTBridgeTokenAddrOnETH="${chain33MainBridgeTokenAddr}"
        offline_create_bridge_token_chain33_symbol "ETH"
        chain33MainBridgeTokenAddrETH="${chain33MainBridgeTokenAddr}"
        offline_create_bridge_token_eth_BTY
        ethereumBtyBridgeTokenAddrOnETH="${ethereumBtyBridgeTokenAddr}"
        offline_deploy_erc20_create_tether_usdt_USDT "USDT"
        ethereumUSDTERC20TokenAddrOnETH="${ethereumUSDTERC20TokenAddr}"
    }

    # shellcheck disable=SC2086
    {
        docker cp "${chain33BridgeBank}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33BridgeBank}.abi
        docker cp "${chain33BridgeRegistry}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33BridgeRegistry}.abi
        docker cp "${chain33USDTBridgeTokenAddrOnETH}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33USDTBridgeTokenAddrOnETH}.abi
        docker cp "${chain33MainBridgeTokenAddrETH}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33MainBridgeTokenAddrETH}.abi
        docker cp "${ethereumBridgeBankOnETH}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${ethereumBridgeBankOnETH}.abi
        docker cp "${ethereumBridgeRegistryOnETH}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${ethereumBridgeRegistryOnETH}.abi
    }

    # start ebrelayer B C D
    updata_toml_start_bcd
    restart_ebrelayerA

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2034
# shellcheck disable=SC2154
function AllRelayerMainTest() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    set +e

    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    # shellcheck disable=SC2120
    if [[ $# -ge 2 ]]; then
        chain33ID="${2}"
    fi

    get_cli

    # init
    Chain33Cli=${MainCli}
    InitChain33Validator
    # para add
    initPara

    StartDockerRelayerDeploy_onlyETH
    #  test_all_onlyETH
    Boss4xCLI=${Boss4xCLIeth}
    CLIA=${CLIAeth}
    ethereumBridgeBank="${ethereumBridgeBankOnETH}"
    ethereumMultisignAddr="${ethereumMultisignAddrOnETH}"
    chain33MainBridgeTokenAddr="${chain33MainBridgeTokenAddrETH}"
    ethereumBtyBridgeTokenAddr="${ethereumBtyBridgeTokenAddrOnETH}"
    ethereumUSDTERC20TokenAddr="${ethereumUSDTERC20TokenAddrOnETH}"
    chain33USDTBridgeTokenAddr="${chain33USDTBridgeTokenAddrOnETH}"
    test_lock_and_burn "ETH" "USDT"

    # TestRelayerProxy_onlyETH
    start_docker_ebrelayerProxy
    setWithdraw_ethereum
    TestProxy

    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
