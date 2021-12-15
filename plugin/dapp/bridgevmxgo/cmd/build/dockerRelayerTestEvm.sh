#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./mainPubilcRelayerTest.sh"

# shellcheck disable=SC2034
{
    chain33BridgeBank=""
    ethBridgeBank=""

    chain33BtyERC20TokenAddr="1111111111111111111114oLvT2"
    ethereumBtyBridgeTokenAddr=""
    chain33EthBridgeTokenAddr=""

    ethereumBycERC20TokenAddr=""
    chain33BycBridgeTokenAddr=""

    ethereumUSDTERC20TokenAddr=""
    chain33USDTBridgeTokenAddr=""

    chain33YccERC20TokenAddr=""
    ethereumYccBridgeTokenAddr=""

    chain33ZbcERC20TokenAddr=""
    ethereumZbcBridgeTokenAddr=""

    BridgeRegistryOnChain33=""
    BridgeRegistryOnEth=""

    multisignChain33Addr=""
    multisignEthAddr=""

    chain33ID=0
    maturityDegree=10

    # ETH 部署合约者的私钥 用于部署合约时签名使用
    ethDeployAddr="0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a"
    ethDeployKey="0x8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

    # chain33 部署合约者的私钥 用于部署合约时签名使用
    chain33DeployAddr="1JxhYLYsrscjTaQfaMoVUrnSdrejP7XRQD"
    chain33DeployKey="0x9ef82623a5e9aac58d3a6b06392af66ec77289522b28896aed66abaaede66903"

    # validatorsAddr=["0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]# shellcheck disable=SC2034
    # eth 验证者私钥
    ethValidatorAddra="0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f"
    ethValidatorAddrb="0x0df9a824699bc5878232c9e612fe1a5346a5a368"
    ethValidatorAddrc="0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1"
    ethValidatorAddrd="0xd9dab021e74ecf475788ed7b61356056b2095830"
    ethValidatorAddrKeya="3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
    ethValidatorAddrKeyb="a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
    ethValidatorAddrKeyc="bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
    ethValidatorAddrKeyd="c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"

    # 新增地址 chain33 需要导入地址 转入 10 bty当收费费
    chain33Validatora="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"
    chain33Validatorb="155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6"
    chain33Validatorc="13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv"
    chain33Validatord="113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG"
    chain33ValidatorKeya="0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae"
    chain33ValidatorKeyb="0x9d539bc5fd084eb7fe86ad631dba9aa086dba38418725c38d9751459f567da66"
    chain33ValidatorKeyc="0x0a6671f101e30a2cc2d79d77436b62cdf2664ed33eb631a9c9e3f3dd348a23be"
    chain33ValidatorKeyd="0x3818b257b05ee75b6e43ee0e3cfc2d8502342cf67caed533e3756966690b62a5"

    ethTestAddr1=0xbc333839E37bc7fAAD0137aBaE2275030555101f
    ethTestAddrKey1=0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2
    ethTestAddr2=0x495953A743ef169EC5D4aC7b5F786BF2Bd56aFd5
    ethTestAddrKey2=0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697

    ethReceiverAddr1="0xa4ea64a583f6e51c3799335b28a8f0529570a635"
    #ethReceiverAddrKey1="355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"

    chain33TestAddr1="1Cj1rqUenPmkeD6A8MGEzkBKQFN2H9yL3x"
    chain33TestAddrKey1="0x7269a7a87d476310da37a9ca1ddc9333c9d7a0dfe1f2998b84758843a895433b"
    chain33TestAddr2="1BCGLhdcdthNutQowV2YShuuN9fJRRGLxu"
    chain33TestAddrKey2="0xb74acfd4eebbbd07bcae212baa7f094235ab8dc04f2f1d828681477b98b24008"

    chain33ReceiverAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    chain33ReceiverAddrKey="4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
}

function chain33_offline_send_evm() {
    # shellcheck disable=SC2154
    result=$(${EvmxgoBoss4xCLI} chain33 offline send -f "${1}")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    # shellcheck disable=SC2154
    check_tx "${Chain33Cli}" "${hash}"
}

function DeployEvmxgo() {
    # 在 chain33 上部署合约
    # shellcheck disable=SC2154
    ${EvmxgoBoss4xCLI} chain33 offline create -f 1 -k "${chain33DeployKey}" -n "deploy crossx to chain33" -r "${chain33DeployAddr}, [${chain33Validatora}, ${chain33Validatorb}, ${chain33Validatorc}, ${chain33Validatord}], [96, 1, 1, 1]" --chainID "${chain33ID}"
    result=$(${EvmxgoBoss4xCLI} chain33 offline send -f "deployBridgevmxgo2Chain33.txt")

    for i in {0..6}; do
        hash=$(echo "${result}" | jq -r ".[$i].TxHash")
        check_tx "${Chain33Cli}" "${hash}"
    done
    XgoBridgeRegistryOnChain33=$(echo "${result}" | jq -r ".[6].ContractAddr")

    # 拷贝 BridgeRegistry.abi 和 BridgeBank.abi
    cp XgoBridgeRegistryOnChain33.abi "${XgoBridgeRegistryOnChain33}.abi"
    XgoChain33BridgeBank=$(${Chain33Cli} evm query -c "${chain33DeployAddr}" -b "bridgeBank()" -a "${XgoBridgeRegistryOnChain33}")
    cp XgoChain33BridgeBank.abi "${XgoChain33BridgeBank}.abi"

    ${EvmxgoBoss4xCLI} chain33 offline create_add_lock_list -s ETH -t "${chain33EthBridgeTokenAddr}" -c "${XgoChain33BridgeBank}" -k "${chain33DeployKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "create_add_lock_list.txt"

    ${EvmxgoBoss4xCLI} chain33 offline create_add_lock_list -s BYC -t "${chain33BycBridgeTokenAddr}" -c "${XgoChain33BridgeBank}" -k "${chain33DeployKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "create_add_lock_list.txt"

    ${EvmxgoBoss4xCLI} chain33 offline create_add_lock_list -s USDT -t "${chain33USDTBridgeTokenAddr}" -c "${XgoChain33BridgeBank}" -k "${chain33DeployKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "create_add_lock_list.txt"

    # 重启,需要重新启动relayer,更新nonce
    restart_ebrelayerA
}

# $1 symbol $2 bridgeTokenAddr
function updateConfig() {
    local symbol=$1
    local bridgeTokenAddr=$2
    tx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key":"evmxgo-mint-'"${symbol}"'","value":"{\"address\":\"'"${bridgeTokenAddr}"'\",\"precision\":8,\"introduction\":\"symbol:'"${symbol}"', bridgeTokenAddr:'"${bridgeTokenAddr}"'\"}","op":"add","addr":""}}]}' -H 'content-type:text/plain;' "http://${docker_chain33_ip}:8901" | jq -r ".result")
    if [ "${tx}" == "" ]; then
        echo -e "${RED}update config create tx 1${NOC}"
        exit 1
    fi

    sign=$(${Chain33Cli} wallet sign -k "$chain33ReceiverAddrKey" -d "${tx}")
    hash=$(${Chain33Cli} wallet send -d "${sign}")
    check_tx "${Chain33Cli}" "${hash}"
}

function configbridgevmxgoAddr() {
    local bridgevmxgoAddr=$1
    tx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key":"bridgevmxgo-contract-addr","value":"{\"address\":\"'"${bridgevmxgoAddr}"'\"}","op":"add","addr":""}}]}' -H 'content-type:text/plain;' "http://${docker_chain33_ip}:8901" | jq -r ".result")
    if [ "${tx}" == "" ]; then
        echo -e "${RED}update config create tx 1${NOC}"
        exit 1
    fi

    sign=$(${Chain33Cli} wallet sign -k "$chain33ReceiverAddrKey" -d "${tx}")
    hash=$(${Chain33Cli} wallet send -d "${sign}")
    check_tx "${Chain33Cli}" "${hash}"
}

function TestETH2EVMToChain33() {
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    #    cli_ret "${result}" "balance" ".balance" "16"

    # ETH 这端 lock 11个
    result=$(${CLIA} ethereum lock -m 11 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}")
    #    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    # 查询 ETH 这端 bridgeBank 地址 11 原来16
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    #    cli_ret "${result}" "balance" ".balance" "27"

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 11 * le8
    #    is_equal "${result}" "4700000000"

    updateConfig "ETH" "${chain33EthBridgeTokenAddr}"
    configbridgevmxgoAddr "${XgoChain33BridgeBank}"

    ${EvmxgoBoss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${XgoChain33BridgeBank}" -c "${chain33EthBridgeTokenAddr}" -k "${chain33ReceiverAddrKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "approve_erc20.txt"

    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33ReceiverAddr}" -e "${XgoChain33BridgeBank}" -p "lock(${chain33TestAddr2}, ${chain33EthBridgeTokenAddr}, 500000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    #    is_equal "${result}" "4200000000"

    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${XgoChain33BridgeBank})")
    #    is_equal "${result}" "500000000"
}

function Testethereum2EVMToChain33_byc() {
    # 查询 ETH 这端 bridgeBank 地址原来是
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
    #    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 7个
    result=$(${CLIA} ethereum lock -m 7 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    # 查询 ETH 这端 bridgeBank 地址 7
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
    #    cli_ret "${result}" "balance" ".balance" "7"

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 7 * le8
    #    is_equal "${result}" "700000000"

    updateConfig "BYC" "${chain33BycBridgeTokenAddr}"
    configbridgevmxgoAddr "${XgoChain33BridgeBank}"

    ${EvmxgoBoss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${XgoChain33BridgeBank}" -c "${chain33BycBridgeTokenAddr}" -k "${chain33ReceiverAddrKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "approve_erc20.txt"

    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33ReceiverAddr}" -e "${XgoChain33BridgeBank}" -p "lock(${chain33TestAddr2}, ${chain33BycBridgeTokenAddr}, 500000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    #    is_equal "${result}" "4200000000"

    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${XgoChain33BridgeBank})")
    #    is_equal "${result}" "500000000"
}

function Testethereum2EVMToChain33_usdt() {
    # 查询 ETH 这端 bridgeBank 地址原来是
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    #    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 12个
    result=$(${CLIA} ethereum lock -m 12 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    # 查询 ETH 这端 bridgeBank 地址 12
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    #    cli_ret "${result}" "balance" ".balance" "12"

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 7 * le8
    #    is_equal "${result}" "700000000"

    updateConfig "USDT" "${chain33USDTBridgeTokenAddr}"
    configbridgevmxgoAddr "${XgoChain33BridgeBank}"

    ${EvmxgoBoss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${XgoChain33BridgeBank}" -c "${chain33USDTBridgeTokenAddr}" -k "${chain33ReceiverAddrKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "approve_erc20.txt"

    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33ReceiverAddr}" -e "${XgoChain33BridgeBank}" -p "lock(${chain33TestAddr2}, ${chain33USDTBridgeTokenAddr}, 500000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    #    is_equal "${result}" "4200000000"

    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${XgoChain33BridgeBank})")
    #    is_equal "${result}" "500000000"
}

function get_evm_cli() {
    # shellcheck disable=SC2034
    {
        # shellcheck disable=SC2154
        docker_chain33_ip=$(get_docker_addr "${dockerNamePrefix}_chain33_1")
        MainCli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8801"
        Para8801Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para."
        Para8901Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para."

        CLIA="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/ebcli_A"
        CLIB="docker exec ${dockerNamePrefix}_ebrelayerb_1 /root/ebcli_A"
        CLIC="docker exec ${dockerNamePrefix}_ebrelayerc_1 /root/ebcli_A"
        CLID="docker exec ${dockerNamePrefix}_ebrelayerd_1 /root/ebcli_A"

        docker_ganachetest_ip=$(get_docker_addr "${dockerNamePrefix}_ganachetest_1")
        Boss4xCLI="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/boss4x --rpc_laddr http://${docker_chain33_ip}:8901 --rpc_laddr_ethereum http://${docker_ganachetest_ip}:8545 --paraName user.p.para."
        EvmxgoBoss4xCLI="./evmxgoboss4x --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para."
    }
}

function test_evm_all() {
    # test
    Chain33Cli=${Para8901Cli}
    TestChain33ToEthAssets
    TestETH2Chain33Assets
    TestChain33ToEthZBCAssets
    TestETH2Chain33Byc
    TestETH2Chain33USDT

    Chain33Cli=${Para8901Cli}
    lockBty
    lockChain33Ycc
    lockEth
    lockEthByc
    lockEthUSDT

    # 离线多签地址转入阈值设大
    offline_set_offline_token_Bty 100000000000000 10
    offline_set_offline_token_Chain33Ycc 100000000000000 10
    offline_set_offline_token_Eth 100000000000000 10
    offline_set_offline_token_EthByc 100000000000000 10
    offline_set_offline_token_EthUSDT 100000000000000 10

    DeployEvmxgo
    TestETH2EVMToChain33
    Testethereum2EVMToChain33_byc
    Testethereum2EVMToChain33_usdt
}

function AllRelayerMainTest() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    set +e
    get_evm_cli

    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    # shellcheck disable=SC2120
    if [[ $# -ge 2 ]]; then
        chain33ID="${2}"
    fi

    # init
    Chain33Cli=${MainCli}
    InitChain33Validator
    # para add
    initPara

    Chain33Cli=${Para8901Cli}
    StartDockerRelayerDeploy

    test_evm_all

    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
