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

function offline_deploy_erc20_eth_YCC() {
    # eth 上 铸币 YCC
    echo -e "${GRE}======= 在 ethereum 上创建 ERC20 ycc ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline create_erc20 -m 33000000000000000000 -s YCC -o "${ethTestAddr1}" -d "${ethDeployAddr}"
    ${Boss4xCLI} ethereum offline sign -f "deployErc20YCC.txt" -k "${ethDeployKey}"
    sleep 10
    result=$(${Boss4xCLI} ethereum offline send -f "deploysigntxs.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_eth_tx "${hash}"
    ethereumYccTokenAddr=$(echo "${result}" | jq -r ".[0].ContractAddr")

    ${Boss4xCLI} ethereum offline create_add_lock_list -s YCC -t "${ethereumYccTokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_add_lock_list.txt"
}

function offline_create_bridge_token_chain33_YCC() {
    # 在chain33上创建bridgeToken YCC
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken YCC ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create_bridge_token -c "${chain33BridgeBank}" -s YCC -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_bridge_token.txt"

    chain33YccTokenAddr=$(${Chain33Cli} evm query -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(YCC)")
    echo "YCC Token Addr = ${chain33YccTokenAddr}"
    cp BridgeToken.abi "${chain33YccTokenAddr}.abi"

    result=$(${Chain33Cli} evm query -a "${chain33YccTokenAddr}" -c "${chain33YccTokenAddr}" -b "symbol()")
    is_equal "${result}" "YCC"

    ${CLIA} chain33 token set -t "${chain33YccTokenAddr}" -s YCC
}

function offline_deploy_erc20_chain33_YCC() {
    # chain33 token create YCC
    echo -e "${GRE}======= 在 chain33 上创建 ERC20 YCC ======${NOC}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create_erc20 -s YCC -k "${chain33DeployKey}" -o "${chain33TestAddr1}" --chainID "${chain33ID}"
    result=$(${Boss4xCLI} chain33 offline send -f "deployErc20YCCChain33.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_tx "${Chain33Cli}" "${hash}"
    chain33YccErc20Addr=$(echo "${result}" | jq -r ".[0].ContractAddr")
    cp ERC20.abi "${chain33YccErc20Addr}.abi"

    # echo 'YCC.1:增加allowance的设置,或者使用relayer工具进行'
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${chain33BridgeBank}" -c "${chain33YccErc20Addr}" -k "${chain33TestAddrKey1}" --chainID "${chain33ID}"
    chain33_offline_send "approve_erc20.txt"

    # echo 'YCC.2:#执行add lock操作:addToken2LockList'
    ${Boss4xCLI} chain33 offline create_add_lock_list -c "${chain33BridgeBank}" -s YCC -t "${chain33YccErc20Addr}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_add_lock_list.txt"
}

function offline_deploy_erc20_chain33_ZBC() {
    # chain33 token create ZBC
    echo -e "${GRE}======= 在 chain33 上创建 ERC20 ZBC ======${NOC}"
    ${Boss4xCLI} chain33 offline create_erc20 -s ZBC -k "${chain33DeployKey}" -o "${chain33TestAddr1}" --chainID "${chain33ID}"
    result=$(${Boss4xCLI} chain33 offline send -f "deployErc20ZBCChain33.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_tx "${Chain33Cli}" "${hash}"
    chain33ZBCErc20Addr=$(echo "${result}" | jq -r ".[0].ContractAddr")
    cp ERC20.abi "${chain33ZBCErc20Addr}.abi"

    # echo 'ZBC.1:增加allowance的设置,或者使用relayer工具进行'
    ${Boss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${chain33BridgeBank}" -c "${chain33ZBCErc20Addr}" -k "${chain33TestAddrKey1}" --chainID "${chain33ID}"
    chain33_offline_send "approve_erc20.txt"

    # echo 'ZBC.2:#执行add lock操作:addToken2LockList'
    ${Boss4xCLI} chain33 offline create_add_lock_list -c "${chain33BridgeBank}" -s ZBC -t "${chain33ZBCErc20Addr}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_add_lock_list.txt"
}

function offline_create_bridge_token_eth_BTY() {
    # 在 Ethereum 上创建 bridgeToken BTY
    echo -e "${GRE}======= 在 Ethereum 上创建 bridgeToken BTY ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s BTY -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    ethereumBtyTokenAddr=$(${CLIA} ethereum receipt -s "${hash}" | jq -r .logs[0].address)
    ${CLIA} ethereum token set -t "${ethereumBtyTokenAddr}" -s BTY
}

function offline_create_bridge_token_chain33_ETH() {
    # 在 chain33 上创建 bridgeToken ETH
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken ETH ======${NOC}"
    ${Boss4xCLI} chain33 offline create_bridge_token -c "${chain33BridgeBank}" -s ETH -k "${chain33DeployKey}" --chainID "${chain33ID}" -n "create_bridge_token:ETH"
    chain33_offline_send "create_bridge_token.txt"

    chain33EthTokenAddr=$(${Chain33Cli} evm query -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(ETH)")
    echo "ETH Token Addr= ${chain33EthTokenAddr}"
    cp BridgeToken.abi "${chain33EthTokenAddr}.abi"

    result=$(${Chain33Cli} evm query -a "${chain33EthTokenAddr}" -c "${chain33EthTokenAddr}" -b "symbol()")
    is_equal "${result}" "ETH"

    ${CLIA} chain33 token set -t "${chain33EthTokenAddr}" -s ETH
}

function offline_create_bridge_token_eth_YCC() {
    # ethereum create-bridge-token YCC
    echo -e "${GRE}======= 在 ethereum 上创建 bridgeToken ycc ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s YCC -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    ethBridgeToeknYccAddr=$(${CLIA} ethereum receipt -s "${hash}" | jq -r .logs[0].address)
    ${CLIA} ethereum token set -t "${ethBridgeToeknYccAddr}" -s YCC
    cp BridgeToken.abi "${ethBridgeToeknYccAddr}.abi"
}

function offline_create_bridge_token_eth_ZBC() {
    # ethereum create-bridge-token ZBC
    echo -e "${GRE}======= 在 ethereum 上创建 bridgeToken ZBC ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s ZBC -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    ethBridgeToeknZBCAddr=$(${CLIA} ethereum receipt -s "${hash}" | jq -r .logs[0].address)
    ${CLIA} ethereum token set -t "${ethBridgeToeknZBCAddr}" -s ZBC
    cp BridgeToken.abi "${ethBridgeToeknZBCAddr}.abi"
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

function offline_transfer_multisign_EthYcc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 8 -r "${ethBridgeBank}" -c "${multisignEthAddr}" -d "${ethTestAddr1}" -t "${ethereumYccTokenAddr}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethTestAddrKey1}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"

    # transfer
    #    ${CLIA} ethereum multisign transfer -a 10 -r "${ethMultisignA}" -o "${ethValidatorAddrKeyB}" -t "${ethereumYccTokenAddr}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 10 -r "${ethMultisignA}" -c "${multisignEthAddr}" -d "${ethTestAddr1}" -t "${ethereumYccTokenAddr}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethTestAddrKey1}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethMultisignA}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "10"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "70"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_Chain33Ycc_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer test
    #    hash=$(${CLIA} chain33 multisign transfer -a 10 -r "${chain33BridgeBank}" -t "${chain33YccErc20Addr}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 10 -r "${chain33BridgeBank}" -m "${multisignChain33Addr}" -t "${chain33YccErc20Addr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "6200000000"
    result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
    is_equal "${result}" "30800000000"

    #    hash=$(${CLIA} chain33 multisign transfer -a 5 -r "${chain33MultisignA}" -t "${chain33YccErc20Addr}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 5 -r "${chain33MultisignA}" -m "${multisignChain33Addr}" -t "${chain33YccErc20Addr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${chain33MultisignA}" -b "balanceOf(${chain33MultisignA})")
    is_equal "${result}" "500000000"
    result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
    is_equal "${result}" "30300000000"

    # 判断 ETH 这端是否金额一致
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethBridgeToeknYccAddr}")
    cli_ret "${result}" "balance" ".balance" "370"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
