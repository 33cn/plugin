#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./publicTest.sh"
source "./relayerPublic.sh"

# $1 send file
function chain33_offline_send() {
    # shellcheck disable=SC2154
    result=$(${Boss4xCLI} chain33 offline send -f "${1}")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    # shellcheck disable=SC2154
    check_tx "${Chain33Cli}" "${hash}"
}

# $1 send file $2 key
function ethereum_offline_sign_send() {
    # shellcheck disable=SC2154
    key="${ethDeployKey}"
    if [[ ${2} != "" ]]; then
        key="${2}"
    fi
    ${Boss4xCLI} ethereum offline sign -f "${1}" -k "${key}"
    result=$(${Boss4xCLI} ethereum offline send -f "deploysigntxs.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_eth_tx "${hash}"
}

# shellcheck disable=SC2120
function InitAndOfflineDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    result=$(${CLIA} set_pwd -p 123456hzj)
    cli_ret "${result}" "set_pwd"

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    # shellcheck disable=SC2154
    result=$(${CLIA} chain33 import_privatekey -k "${chain33ValidatorKeya}")
    cli_ret "${result}" "chain33 import_privatekey"

    # shellcheck disable=SC2154
    result=$(${CLIA} ethereum import_privatekey -k "${ethValidatorAddrKeya}")
    cli_ret "${result}" "ethereum import_privatekey"

    # 在 chain33 上部署合约
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create -f 1 -k "${chain33DeployKey}" -n "deploy crossx to chain33" -r "${chain33DeployAddr}, [${chain33Validatora}, ${chain33Validatorb}, ${chain33Validatorc}, ${chain33Validatord}], [25, 25, 25, 25]" --chainID "${chain33ID}"
    result=$(${Boss4xCLI} chain33 offline send -f "deployCrossX2Chain33.txt")

    for i in {0..7}; do
        hash=$(echo "${result}" | jq -r ".[$i].TxHash")
        check_tx "${Chain33Cli}" "${hash}"
    done
    BridgeRegistryOnChain33=$(echo "${result}" | jq -r ".[6].ContractAddr")
    # shellcheck disable=SC2034
    multisignChain33Addr=$(echo "${result}" | jq -r ".[7].ContractAddr")
    ${CLIA} chain33 multisign set_multiSign -a "${multisignChain33Addr}"

    # 拷贝 BridgeRegistry.abi 和 BridgeBank.abi
    cp BridgeRegistry.abi "${BridgeRegistryOnChain33}.abi"
    chain33BridgeBank=$(${Chain33Cli} evm query -c "${chain33DeployAddr}" -b "bridgeBank()" -a "${BridgeRegistryOnChain33}")
    cp Chain33BridgeBank.abi "${chain33BridgeBank}.abi"

    # 在 Eth 上部署合约
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline create -p "25,25,25,25" -o "${ethDeployAddr}" -v "${ethValidatorAddra},${ethValidatorAddrb},${ethValidatorAddrc},${ethValidatorAddrd}"
    ${Boss4xCLI} ethereum offline sign -k "${ethDeployKey}"
    result=$(${Boss4xCLI} ethereum offline send -f "deploysigntxs.txt")
    for i in {0..7}; do
        hash=$(echo "${result}" | jq -r ".[$i].TxHash")
        check_eth_tx "${hash}"
    done
    BridgeRegistryOnEth=$(echo "${result}" | jq -r ".[6].ContractAddr")
    ethBridgeBank=$(echo "${result}" | jq -r ".[3].ContractAddr")
    # shellcheck disable=SC2034
    multisignEthAddr=$(echo "${result}" | jq -r ".[7].ContractAddr")
    ${CLIA} ethereum multisign set_multiSign -a "${multisignEthAddr}"

    # 拷贝 BridgeRegistry.abi 和 BridgeBank.abi
    cp BridgeRegistry.abi "${BridgeRegistryOnEth}.abi"
    cp EthBridgeBank.abi "${ethBridgeBank}.abi"

    # 修改 relayer.toml 字段
    updata_relayer "BridgeRegistryOnChain33" "${BridgeRegistryOnChain33}" "./relayer.toml"

    line=$(delete_line_show "./relayer.toml" "BridgeRegistry=")
    if [ "${line}" ]; then
        sed -i ''"${line}"' a BridgeRegistry="'"${BridgeRegistryOnEth}"'"' "./relayer.toml"
    fi

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_deploy_erc20_eth_BYC() {
    # eth 上 铸币 BYC
    echo -e "${GRE}======= 在 ethereum 上创建 ERC20 BYC ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline create_erc20 -m 33000000000000000000 -s BYC -o "${ethTestAddr1}" -d "${ethDeployAddr}"
    ${Boss4xCLI} ethereum offline sign -f "deployErc20BYC.txt" -k "${ethDeployKey}"
    sleep 10
    result=$(${Boss4xCLI} ethereum offline send -f "deploysigntxs.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_eth_tx "${hash}"
    ethereumBycERC20TokenAddr=$(echo "${result}" | jq -r ".[0].ContractAddr")

    ${Boss4xCLI} ethereum offline create_add_lock_list -s BYC -t "${ethereumBycERC20TokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_add_lock_list.txt"
}

function offline_deploy_erc20_eth_USDT() {
    # eth 上 铸币 USDT
    echo -e "${GRE}======= 在 ethereum 上创建 ERC20 USDT ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline create_erc20 -m 33000000000000000000 -s USDT -o "${ethTestAddr1}" -d "${ethDeployAddr}"
    ${Boss4xCLI} ethereum offline sign -f "deployErc20USDT.txt" -k "${ethDeployKey}"
    sleep 10
    result=$(${Boss4xCLI} ethereum offline send -f "deploysigntxs.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_eth_tx "${hash}"
    ethereumUSDTERC20TokenAddr=$(echo "${result}" | jq -r ".[0].ContractAddr")

    ${Boss4xCLI} ethereum offline create_add_lock_list -s USDT -t "${ethereumUSDTERC20TokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_add_lock_list.txt"
}

function offline_deploy_erc20_create_tether_usdt_USDT() {
    # eth 上 铸币 USDT
    echo -e "${GRE}======= 在 ethereum 上创建 ERC20 USDT ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline create_tether_usdt -m 33000000000000000000 -s USDT -d "${ethTestAddr1}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline sign -f "deployTetherUSDT.txt" -k "${ethTestAddrKey1}"
    sleep 10
    result=$(${Boss4xCLI} ethereum offline send -f "deploysigntxs.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_eth_tx "${hash}"
    ethereumUSDTERC20TokenAddr=$(echo "${result}" | jq -r ".[0].ContractAddr")

    ${Boss4xCLI} ethereum offline create_add_lock_list -s USDT -t "${ethereumUSDTERC20TokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_add_lock_list.txt"
}

function offline_create_bridge_token_chain33_BYC() {
    # 在chain33上创建bridgeToken BYC
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken BYC ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create_bridge_token -c "${chain33BridgeBank}" -s BYC -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_bridge_token.txt"

    chain33BycBridgeTokenAddr=$(${Chain33Cli} evm query -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(BYC)")
    echo "BYC Bridge Token Addr = ${chain33BycBridgeTokenAddr}"
    cp BridgeToken.abi "${chain33BycBridgeTokenAddr}.abi"

    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33BycBridgeTokenAddr}" -b "symbol()")
    is_equal "${result}" "BYC"
}

function offline_create_bridge_token_chain33_USDT() {
    # 在chain33上创建bridgeToken USDT
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken USDT ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create_bridge_token -c "${chain33BridgeBank}" -s USDT -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_bridge_token.txt"

    chain33USDTBridgeTokenAddr=$(${Chain33Cli} evm query -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(USDT)")
    echo "USDT Bridge Token Addr = ${chain33USDTBridgeTokenAddr}"
    cp BridgeToken.abi "${chain33USDTBridgeTokenAddr}.abi"

    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33USDTBridgeTokenAddr}" -b "symbol()")
    is_equal "${result}" "USDT"
}

function offline_create_bridge_token_chain33_BUSD() {
    # 在chain33上创建bridgeToken BUSD
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken BUSD ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create_bridge_token -c "${chain33BridgeBank}" -s BUSD -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_bridge_token.txt"

    chain33BUSDBridgeTokenAddr=$(${Chain33Cli} evm query -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(BUSD)")
    echo "BUSD Bridge Token Addr = ${chain33BUSDBridgeTokenAddr}"
    cp BridgeToken.abi "${chain33BUSDBridgeTokenAddr}.abi"

    result=$(${Chain33Cli} evm query -a "${chain33BUSDBridgeTokenAddr}" -c "${chain33BUSDBridgeTokenAddr}" -b "symbol()")
    is_equal "${result}" "BUSD"
}

function offline_deploy_erc20_chain33_YCC() {
    # chain33 token create YCC
    echo -e "${GRE}======= 在 chain33 上创建 ERC20 YCC ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create_erc20 -s YCC -k "${chain33DeployKey}" -o "${chain33TestAddr1}" --chainID "${chain33ID}"
    result=$(${Boss4xCLI} chain33 offline send -f "deployErc20YCCChain33.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_tx "${Chain33Cli}" "${hash}"
    chain33YccERC20TokenAddr=$(echo "${result}" | jq -r ".[0].ContractAddr")
    cp ERC20.abi "${chain33YccERC20TokenAddr}.abi"

    # echo 'YCC.1:增加allowance的设置,或者使用relayer工具进行'
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${chain33BridgeBank}" -c "${chain33YccERC20TokenAddr}" -k "${chain33TestAddrKey1}" --chainID "${chain33ID}"
    chain33_offline_send "approve_erc20.txt"

    # echo 'YCC.2:#执行add lock操作:addToken2LockList'
    ${Boss4xCLI} chain33 offline create_add_lock_list -c "${chain33BridgeBank}" -s YCC -t "${chain33YccERC20TokenAddr}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_add_lock_list.txt"
}

function offline_deploy_erc20_chain33_ZBC() {
    # chain33 token create ZBC
    echo -e "${GRE}======= 在 chain33 上创建 ERC20 ZBC ======${NOC}"
    ${Boss4xCLI} chain33 offline create_erc20 -s ZBC -k "${chain33DeployKey}" -o "${chain33TestAddr1}" --chainID "${chain33ID}"
    result=$(${Boss4xCLI} chain33 offline send -f "deployErc20ZBCChain33.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_tx "${Chain33Cli}" "${hash}"
    chain33ZbcERC20TokenAddr=$(echo "${result}" | jq -r ".[0].ContractAddr")
    cp ERC20.abi "${chain33ZbcERC20TokenAddr}.abi"

    # echo 'ZBC.1:增加allowance的设置,或者使用relayer工具进行'
    ${Boss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${chain33BridgeBank}" -c "${chain33ZbcERC20TokenAddr}" -k "${chain33TestAddrKey1}" --chainID "${chain33ID}"
    chain33_offline_send "approve_erc20.txt"

    # echo 'ZBC.2:#执行add lock操作:addToken2LockList'
    ${Boss4xCLI} chain33 offline create_add_lock_list -c "${chain33BridgeBank}" -s ZBC -t "${chain33ZbcERC20TokenAddr}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_add_lock_list.txt"
}

function offline_create_bridge_token_eth_BTY() {
    # 在 Ethereum 上创建 bridgeToken BTY
    echo -e "${GRE}======= 在 Ethereum 上创建 bridgeToken BTY ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s BTY -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    # shellcheck disable=SC2034
    ethereumBtyBridgeTokenAddr=$(${CLIA} ethereum receipt -s "${hash}" | jq -r .logs[0].address)
}

function offline_create_bridge_token_chain33_ETH() {
    # 在 chain33 上创建 bridgeToken ETH
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken $1 ======${NOC}"
    local symbolName="$1"
    ${Boss4xCLI} chain33 offline create_bridge_token -c "${chain33BridgeBank}" -s "${symbolName}" -k "${chain33DeployKey}" --chainID "${chain33ID}" -n "create_bridge_token:${symbolName}"
    chain33_offline_send "create_bridge_token.txt"

    chain33EthBridgeTokenAddr=$(${Chain33Cli} evm query -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(${symbolName})")
    echo "${symbolName} Token Addr= ${chain33EthBridgeTokenAddr}"
    cp BridgeToken.abi "${chain33EthBridgeTokenAddr}.abi"

    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33EthBridgeTokenAddr}" -b "symbol()")
    is_equal "${result}" "${symbolName}"
}

function offline_create_bridge_token_eth_YCC() {
    # ethereum create-bridge-token YCC
    echo -e "${GRE}======= 在 ethereum 上创建 bridgeToken ycc ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s YCC -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    ethereumYccBridgeTokenAddr=$(${CLIA} ethereum receipt -s "${hash}" | jq -r .logs[0].address)
    cp BridgeToken.abi "${ethereumYccBridgeTokenAddr}.abi"
}

function offline_create_bridge_token_eth_ZBC() {
    # ethereum create-bridge-token ZBC
    echo -e "${GRE}======= 在 ethereum 上创建 bridgeToken ZBC ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s ZBC -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    ethereumZbcBridgeTokenAddr=$(${CLIA} ethereum receipt -s "${hash}" | jq -r .logs[0].address)
    cp BridgeToken.abi "${ethereumZbcBridgeTokenAddr}.abi"
}

function offline_setupChain33Multisign() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== 设置 chain33 离线钱包合约 ===========${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_setup -m "${multisignChain33Addr}" -o "${chain33MultisignA},${chain33MultisignB},${chain33MultisignC},${chain33MultisignD}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_setup.txt"

    ${Boss4xCLI} chain33 offline set_offline_addr -a "${multisignChain33Addr}" -c "${chain33BridgeBank}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "chain33_set_offline_addr.txt"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_setupEthMultisign() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== 设置 ETH 离线钱包合约 ===========${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline multisign_setup -m "${multisignEthAddr}" -d "${ethDeployAddr}" -o "${ethMultisignA},${ethMultisignB},${ethMultisignC},${ethMultisignD}"
    ethereum_offline_sign_send "multisign_setup.txt"

    ${Boss4xCLI} ethereum offline set_offline_addr -a "${multisignEthAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_addr.txt"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_Eth_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer
    # shellcheck disable=SC2154
    #    ${CLIA} ethereum multisign transfer -a 3 -r "${ethBridgeBank}" -o "${ethValidatorAddrKeyB}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 3 -r "${ethBridgeBank}" -c "${multisignEthAddr}" -d "${ethTestAddr1}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    # shellcheck disable=SC2154
    ethereum_offline_sign_send create_multisign_tx.txt "${ethTestAddrKey1}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
        cli_ret "${result}" "balance" ".balance" "16"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}")
        cli_ret "${result}" "balance" ".balance" "20"

    # transfer
    # shellcheck disable=SC2154
    #    ${CLIA} ethereum multisign transfer -a 5 -r "${ethMultisignA}" -o "${ethValidatorAddrKeyB}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 5 -r "${ethMultisignA}" -c "${multisignEthAddr}" -d "${ethTestAddr1}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethTestAddrKey1}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethMultisignA}")
        cli_ret "${result}" "balance" ".balance" "1005"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}")
        cli_ret "${result}" "balance" ".balance" "15"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_EthByc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 8 -r "${ethBridgeBank}" -c "${multisignEthAddr}" -d "${ethTestAddr1}" -t "${ethereumBycERC20TokenAddr}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethTestAddrKey1}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"

    # transfer
    #    ${CLIA} ethereum multisign transfer -a 10 -r "${ethMultisignA}" -o "${ethValidatorAddrKeyB}" -t "${ethereumBycERC20TokenAddr}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 10 -r "${ethMultisignA}" -c "${multisignEthAddr}" -d "${ethTestAddr1}" -t "${ethereumBycERC20TokenAddr}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethTestAddrKey1}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethMultisignA}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "10"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "70"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_EthUSDT() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 8 -r "${ethBridgeBank}" -c "${multisignEthAddr}" -d "${ethTestAddr1}" -t "${ethereumUSDTERC20TokenAddr}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethTestAddrKey1}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"

    # transfer
    #    ${CLIA} ethereum multisign transfer -a 10 -r "${ethMultisignA}" -o "${ethValidatorAddrKeyB}" -t "${ethereumUSDTERC20TokenAddr}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 10 -r "${ethMultisignA}" -c "${multisignEthAddr}" -d "${ethTestAddr1}" -t "${ethereumUSDTERC20TokenAddr}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethTestAddrKey1}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethMultisignA}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "10"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "70"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_Chain33Ycc_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer test
    #    hash=$(${CLIA} chain33 multisign transfer -a 10 -r "${chain33BridgeBank}" -t "${chain33YccERC20TokenAddr}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 10 -r "${chain33BridgeBank}" -m "${multisignChain33Addr}" -t "${chain33YccERC20TokenAddr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} evm query -a "${chain33YccERC20TokenAddr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "6200000000"
    result=$(${Chain33Cli} evm query -a "${chain33YccERC20TokenAddr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
    is_equal "${result}" "30800000000"

    #    hash=$(${CLIA} chain33 multisign transfer -a 5 -r "${chain33MultisignA}" -t "${chain33YccERC20TokenAddr}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 5 -r "${chain33MultisignA}" -m "${multisignChain33Addr}" -t "${chain33YccERC20TokenAddr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} evm query -a "${chain33YccERC20TokenAddr}" -c "${chain33MultisignA}" -b "balanceOf(${chain33MultisignA})")
    is_equal "${result}" "500000000"
    result=$(${Chain33Cli} evm query -a "${chain33YccERC20TokenAddr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
    is_equal "${result}" "30300000000"

    # 判断 ETH 这端是否金额一致
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumYccBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "370"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
