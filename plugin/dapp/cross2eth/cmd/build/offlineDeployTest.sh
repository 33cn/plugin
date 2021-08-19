#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

source "./publicTest.sh"
source "./relayerPublic.sh"
source "./multisignPublicTest.sh"

# ETH 部署合约者的私钥 用于部署合约时签名使用
ethDeployAddr="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
ethDeployKey="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

# validatorsAddr=["0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]
ethValidatorAddrKeyA="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

# chain33 部署合约者的私钥 用于部署合约时签名使用
chain33DeployAddr="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"

chain33ReceiverAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
chain33ReceiverAddrKey="4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"

ethReceiverAddr1="0xa4ea64a583f6e51c3799335b28a8f0529570a635"
#ethReceiverAddrKey1="355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"

maturityDegree=10

Chain33Cli="../../chain33-cli"
chain33BridgeBank=""
ethBridgeBank=""
chain33BtyTokenAddr="1111111111111111111114oLvT2"
chain33EthTokenAddr=""
ethereumBtyTokenAddr=""
chain33YccTokenAddr=""
ethereumYccTokenAddr=""
chain33ZBCErc20Addr=""
ethBridgeToeknZBCAddr=""

CLIA="./ebcli_A"
Boss4xCLI="./boss4x"
chain33ID=33

# $1 send file
function chain33_offline_send() {
    result=$(${Boss4xCLI} chain33 offline send -f "${1}")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_tx "${Chain33Cli}" "${hash}"
}

# $1 send file $2 key
function ethereum_offline_sign_send() {
    key="${ethDeployKey}"
    if [[ ${2} != "" ]]; then
        key="${2}"
    fi
    ${Boss4xCLI} ethereum offline sign -f "${1}" -k "${key}"
    result=$(${Boss4xCLI} ethereum offline send -f "deploysigntxs.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_eth_tx "${hash}"
}

# chain33 lock BTY, eth burn BTY
function TestChain33ToEthAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== chain33 lock BTY, eth burn BTY ===========${NOC}"
    result=$(${CLIA} ethereum balance -o "${ethDeployAddr}" -t "${ethereumBtyTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 原来的地址金额
    result=$(${Chain33Cli} account balance -a "${chain33DeployAddr}" -e evm)
#    balance=$(cli_ret "${result}" "balance" ".balance")

    # chain33 lock bty
    hash=$(${Chain33Cli} evm call -f 1 -a 5 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "lock(${ethDeployAddr}, ${chain33BtyTokenAddr}, 500000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    # 原来的地址金额 减少了 5
    result=$(${Chain33Cli} account balance -a "${chain33DeployAddr}" -e evm)
#    cli_ret "${result}" "balance" ".balance" "$(echo "${balance}-5" | bc)"
    #balance_ret "${result}" "195.0000"

    # chain33BridgeBank 是否增加了 5
    result=$(${Chain33Cli} account balance -a "${chain33BridgeBank}" -e evm)
    balance_ret "${result}" "5.0000"

    eth_block_wait 10

    # eth 这端 金额是否增加了 5
    result=$(${CLIA} ethereum balance -o "${ethDeployAddr}" -t "${ethereumBtyTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "5"

    # eth burn
    result=$(${CLIA} ethereum burn -m 3 -k "${ethDeployKey}" -r "${chain33ReceiverAddr}" -t "${ethereumBtyTokenAddr}" ) #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    eth_block_wait 10

    # eth 这端 金额是否减少了 3
    result=$(${CLIA} ethereum balance -o "${ethDeployAddr}" -t "${ethereumBtyTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "2"

    sleep ${maturityDegree}

     # 接收的地址金额 变成了 3
    result=$(${Chain33Cli} account balance -a "${chain33ReceiverAddr}" -e evm)
    balance_ret "${result}" "3.0000"

    # chain33BridgeBank 是否减少了 3
    result=$(${Chain33Cli} account balance -a "${chain33BridgeBank}" -e evm)
    balance_ret "${result}" "2.0000"

    # eth burn
    result=$(${CLIA} ethereum burn -m 2 -k "${ethDeployKey}" -r "${chain33ReceiverAddr}" -t "${ethereumBtyTokenAddr}" ) #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    eth_block_wait 10
    sleep ${maturityDegree}

    result=$(${Chain33Cli} account balance -a "${chain33BridgeBank}" -e evm)
    balance_ret "${result}" "0.0000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# chain33 lock ZBC, eth burn ZBC
function TestChain33ToEthZBCAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== chain33 lock ZBC, eth burn ZBC ===========${NOC}"
    result=$(${CLIA} ethereum balance -o "${ethDeployAddr}" -t "${ethBridgeToeknZBCAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 原来的地址金额
    result=$(${Chain33Cli} evm abi call -a "${chain33ZBCErc20Addr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "0"

    # chain33 lock ZBC
    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "lock(${ethDeployAddr}, ${chain33ZBCErc20Addr}, 900000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    # chain33BridgeBank 是否增加了 9
    result=$(${Chain33Cli} evm abi call -a "${chain33ZBCErc20Addr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "900000000"

    eth_block_wait 10

    # eth 这端 金额是否增加了 9
    result=$(${CLIA} ethereum balance -o "${ethDeployAddr}" -t "${ethBridgeToeknZBCAddr}")
    cli_ret "${result}" "balance" ".balance" "9"

    # eth burn
    result=$(${CLIA} ethereum burn -m 8 -k "${ethDeployKey}" -r "${chain33ReceiverAddr}" -t "${ethBridgeToeknZBCAddr}" ) #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    eth_block_wait 10

    # eth 这端 金额是否减少了 1
    result=$(${CLIA} ethereum balance -o "${ethDeployAddr}" -t "${ethBridgeToeknZBCAddr}")
    cli_ret "${result}" "balance" ".balance" "1"

    sleep ${maturityDegree}

     # 接收的地址金额 变成了 8
     result=$(${Chain33Cli} evm abi call -a "${chain33ZBCErc20Addr}" -c "${chain33ReceiverAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "800000000"

    # chain33BridgeBank 是否减少了 1
    result=$(${Chain33Cli} evm abi call -a "${chain33ZBCErc20Addr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "100000000"

    # eth burn
    result=$(${CLIA} ethereum burn -m 1 -k "${ethDeployKey}" -r "${chain33ReceiverAddr}" -t "${ethBridgeToeknZBCAddr}" ) #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    eth_block_wait 10
    sleep ${maturityDegree}

    result=$(${Chain33Cli} evm abi call -a "${chain33ZBCErc20Addr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "0"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn
function TestETH2Chain33Assets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" )
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 11个
    result=$(${CLIA} ethereum lock -m 11 -k "${ethValidatorAddrKeyA}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

     # eth 等待 10 个区块
    eth_block_wait 10

    # 查询 ETH 这端 bridgeBank 地址 11
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" )
    cli_ret "${result}" "balance" ".balance" "11"

    sleep ${maturityDegree}

    # chain33 chain33EthTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm abi call -a "${chain33EthTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 11 * le8
    is_equal "${result}" "1100000000"

    # 原来的数额
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}")
    cli_ret "${result}" "balance" ".balance" "100"

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum'
    ${CLIA} chain33 burn -m 5 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33EthTokenAddr}"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm abi call -a "${chain33EthTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 11-5 * le8
    is_equal "${result}" "600000000"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" )
    cli_ret "${result}" "balance" ".balance" "6"

    # 比之前多 5
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}")
    cli_ret "${result}" "balance" ".balance" "105"

    ${CLIA} chain33 burn -m 6 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33EthTokenAddr}"
    
    sleep ${maturityDegree}

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" )
    cli_ret "${result}" "balance" ".balance" "0"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 ycc 资产,然后在 chain33 上 burn ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 7个 YCC
    result=$(${CLIA} ethereum lock -m 7 -k "${ethDeployKey}" -r "${chain33ReceiverAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "lock"

     # eth 等待 10 个区块
    eth_block_wait 10

    # 查询 ETH 这端 bridgeBank 地址 7 YCC
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "7"

    sleep ${maturityDegree}

    # chain33 chain33EthTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm abi call -a "${chain33YccTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 7 * le8
    is_equal "${result}" "700000000"

    # 原来的数额 0
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
    ${CLIA} chain33 burn -m 5 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33YccTokenAddr}"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm abi call -a "${chain33YccTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 7-5 * le8
    is_equal "${result}" "200000000"

    # 查询 ETH 这端 bridgeBank 地址 2
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "2"

    # 更新后的金额 5
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "5"

    ${CLIA} chain33 burn -m 2 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33YccTokenAddr}"

    sleep ${maturityDegree}

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2120
function InitAndOfflineDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    result=$(${CLIA} set_pwd -p 123456hzj)
    cli_ret "${result}" "set_pwd"

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    # shellcheck disable=SC2154
    result=$(${CLIA} chain33 import_privatekey -k "${chain33ValidatorKeyA}")
    cli_ret "${result}" "chain33 import_privatekey"

    result=$(${CLIA} ethereum import_privatekey -k "${ethValidatorAddrKeyA}")
    cli_ret "${result}" "ethereum import_privatekey"

    # 在 chain33 上部署合约
#    ./boss4x chain33 offline create -f 1 -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -n "deploy crossx to chain33" -r "1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ, [1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ, 155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6, 13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv, 113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG], [25, 25, 25, 25]" --chainID 33
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create -f 1 -k "${chain33DeployKey}" -n "deploy crossx to chain33" -r "${chain33ValidatorA}, [${chain33ValidatorA}, ${chain33ValidatorB}, ${chain33ValidatorC}, ${chain33ValidatorD}], [25, 25, 25, 25]" --chainID 33
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
    chain33BridgeBank=$(${Chain33Cli} evm abi call -c "${chain33DeployAddr}" -b "bridgeBank()" -a "${BridgeRegistryOnChain33}")
    cp Chain33BridgeBank.abi "${chain33BridgeBank}.abi"

    # 在 Eth 上部署合约
#    ./boss4x ethereum offline create -p "25,25,25,25" -o 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a -v "0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a,0x0df9a824699bc5878232c9e612fe1a5346a5a368,0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1,0xd9dab021e74ecf475788ed7b61356056b2095830"
#    ./boss4x ethereum offline sign -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline create -p "25,25,25,25" -o "${ethValidatorAddrA}" -v "${ethValidatorAddrA},${ethValidatorAddrB},${ethValidatorAddrC},${ethValidatorAddrD}"
    ${Boss4xCLI} ethereum offline sign -k "${ethValidatorAddrKeyA}"
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
    ${Boss4xCLI} ethereum offline create_erc20 -m 33000000000000000000 -s YCC -o "${ethValidatorAddrA}" -d "${ethDeployAddr}"
    ${Boss4xCLI} ethereum offline sign -f "deployErc20YCC.txt" -k "${ethDeployKey}"
    sleep 10
    result=$(${Boss4xCLI} ethereum offline send -f "deploysigntxs.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_eth_tx "${hash}"
    ethereumYccTokenAddr=$(echo "${result}" | jq -r ".[0].ContractAddr")

#    result=$(${CLIA} ethereum token add_lock_list -s YCC -t "${ethereumYccTokenAddr}")
#    cli_ret "${result}" "add_lock_list"
    ${Boss4xCLI} ethereum offline create_add_lock_list -s YCC -t "${ethereumYccTokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_add_lock_list.txt"
}

function offline_create_bridge_token_chain33_YCC() {
    # 在chain33上创建bridgeToken YCC
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken YCC ======${NOC}"
#    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "createNewBridgeToken(YCC)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"
    ${Boss4xCLI} chain33 offline create_bridge_token -c "${chain33BridgeBank}" -s YCC -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_bridge_token.txt"

    chain33YccTokenAddr=$(${Chain33Cli} evm abi call -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(YCC)")
    echo "YCC Token Addr = ${chain33YccTokenAddr}"
    cp BridgeToken.abi "${chain33YccTokenAddr}.abi"

    result=$(${Chain33Cli} evm abi call -a "${chain33YccTokenAddr}" -c "${chain33YccTokenAddr}" -b "symbol()")
    is_equal "${result}" "YCC"
}

function offline_deploy_erc20_chain33_YCC() {
    # chain33 token create YCC
    echo -e "${GRE}======= 在 chain33 上创建 ERC20 YCC ======${NOC}"
    ${Boss4xCLI} chain33 offline create_erc20 -s YCC -k "${chain33DeployKey}" -o "${chain33DeployAddr}" --chainID 33
    result=$(${Boss4xCLI} chain33 offline send -f "deployErc20YCCChain33.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_tx "${Chain33Cli}" "${hash}"
    chain33YccErc20Addr=$(echo "${result}" | jq -r ".[0].ContractAddr")
    cp ERC20.abi "${chain33YccErc20Addr}.abi"

    # echo 'YCC.1:增加allowance的设置,或者使用relayer工具进行'
#    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33YccErc20Addr}" -p "approve(${chain33BridgeBank}, 330000000000)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"
    ${Boss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${chain33BridgeBank}" -c "${chain33YccErc20Addr}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "approve_erc20.txt"

    # echo 'YCC.2:#执行add lock操作:addToken2LockList'
#    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "addToken2LockList(${chain33YccErc20Addr}, YCC)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"
    ${Boss4xCLI} chain33 offline create_add_lock_list -c "${chain33BridgeBank}" -s YCC -t "${chain33YccErc20Addr}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_add_lock_list.txt"
}

function offline_deploy_erc20_chain33_ZBC() {
    # chain33 token create ZBC
    echo -e "${GRE}======= 在 chain33 上创建 ERC20 ZBC ======${NOC}"
    ${Boss4xCLI} chain33 offline create_erc20 -s ZBC -k "${chain33DeployKey}" -o "${chain33DeployAddr}" --chainID 33
    result=$(${Boss4xCLI} chain33 offline send -f "deployErc20ZBCChain33.txt")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    check_tx "${Chain33Cli}" "${hash}"
    chain33ZBCErc20Addr=$(echo "${result}" | jq -r ".[0].ContractAddr")
    cp ERC20.abi "${chain33ZBCErc20Addr}.abi"

    # echo 'ZBC.1:增加allowance的设置,或者使用relayer工具进行'
#    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33ZBCErc20Addr}" -p "approve(${chain33BridgeBank}, 330000000000)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"
    ${Boss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${chain33BridgeBank}" -c "${chain33ZBCErc20Addr}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "approve_erc20.txt"

    # echo 'ZBC.2:#执行add lock操作:addToken2LockList'
#    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "addToken2LockList(${chain33ZBCErc20Addr}, ZBC)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"
    ${Boss4xCLI} chain33 offline create_add_lock_list -c "${chain33BridgeBank}" -s ZBC -t "${chain33ZBCErc20Addr}" -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_add_lock_list.txt"
}

function offline_create_bridge_token_eth_BTY() {
    # 在 Ethereum 上创建 bridgeToken BTY
    echo -e "${GRE}======= 在 Ethereum 上创建 bridgeToken BTY ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s BTY -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    ethereumBtyTokenAddr=$(./ebcli_A ethereum receipt -s "${hash}" | jq -r .logs[0].address)
    ./ebcli_A ethereum token set -t "${ethereumBtyTokenAddr}" -s BTY
}

function offline_create_bridge_token_chain33_ETH() {
    # 在 chain33 上创建 bridgeToken ETH
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken ETH ======${NOC}"
#    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "createNewBridgeToken(ETH)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"
    ${Boss4xCLI} chain33 offline create_bridge_token -c "${chain33BridgeBank}" -s ETH -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "create_bridge_token.txt"

    chain33EthTokenAddr=$(${Chain33Cli} evm abi call -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(ETH)")
    echo "ETH Token Addr= ${chain33EthTokenAddr}"
    cp BridgeToken.abi "${chain33EthTokenAddr}.abi"

    result=$(${Chain33Cli} evm abi call -a "${chain33EthTokenAddr}" -c "${chain33EthTokenAddr}" -b "symbol()")
    is_equal "${result}" "ETH"
}

function offline_create_bridge_token_eth_YCC() {
    # ethereum create-bridge-token YCC
    echo -e "${GRE}======= 在 ethereum 上创建 bridgeToken ycc ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s YCC -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    ethBridgeToeknYccAddr=$(./ebcli_A ethereum receipt -s "${hash}" | jq -r .logs[0].address)
    ./ebcli_A ethereum token set -t "${ethBridgeToeknYccAddr}" -s YCC
    cp BridgeToken.abi "${ethBridgeToeknYccAddr}.abi"
}

function offline_create_bridge_token_eth_ZBC() {
    # ethereum create-bridge-token ZBC
    echo -e "${GRE}======= 在 ethereum 上创建 bridgeToken ZBC ======${NOC}"
    ${Boss4xCLI} ethereum offline create_bridge_token -s ZBC -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_bridge_token.txt"

    ethBridgeToeknZBCAddr=$(./ebcli_A ethereum receipt -s "${hash}" | jq -r .logs[0].address)
    ./ebcli_A ethereum token set -t "${ethBridgeToeknZBCAddr}" -s ZBC
    cp BridgeToken.abi "${ethBridgeToeknZBCAddr}.abi"
}

function offline_setupChain33Multisign() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    echo -e "${GRE}=========== 设置 chain33 离线钱包合约 ===========${NOC}"
    # shellcheck disable=SC2154
#    result=$(${CLIA} chain33 multisign setup -k "${chain33DeployKey}" -o "${chain33MultisignA},${chain33MultisignB},${chain33MultisignC},${chain33MultisignD}")
#    cli_ret "${result}" "chain33 multisign setup"
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
#    result=$(${CLIA} ethereum multisign setup -k "${ethDeployKey}" -o "${ethMultisignA},${ethMultisignB},${ethMultisignC},${ethMultisignD}")
#    cli_ret "${result}" "ethereum multisign setup"
    ${Boss4xCLI} ethereum offline multisign_setup -m "${multisignEthAddr}" -d "${ethDeployAddr}" -o "${ethMultisignA},${ethMultisignB},${ethMultisignC},${ethMultisignD}"
    ethereum_offline_sign_send "multisign_setup.txt"

    ${Boss4xCLI} ethereum offline set_offline_addr -a "${multisignEthAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_addr.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function StartRelayerAndOfflineDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    kill_all_ebrelayer

    # 修改 relayer.toml 配置文件 pushName 字段
    pushNameChange "./relayer.toml"
    validators_config

    # 删除私钥
    delete_line "./relayer.toml" "deployerPrivateKey="
    delete_line "./relayer.toml" "deployerPrivateKey="

    # 启动 ebrelayer
    start_ebrelayerA

    # 导入私钥 部署合约 设置 bridgeRegistry 地址
    InitAndOfflineDeploy

    # 设置离线多签数据
    initMultisignChain33Addr
    offline_setupChain33Multisign
    offline_setupEthMultisign
    transferChain33MultisignFee

    # 重启
    kill_ebrelayer ebrelayer
    start_ebrelayerA

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    ethBridgeBank=$(./ebcli_A ethereum bridgeBankAddr | jq -r ".addr")
    cp EthBridgeBank.abi "${ethBridgeBank}.abi"

    # start ebrelayer B C D
    updata_toml_start_BCD

    # 设置 token 地址
    # InitTokenAddr
    offline_create_bridge_token_eth_BTY
    offline_create_bridge_token_chain33_ETH
    offline_deploy_erc20_eth_YCC
    offline_create_bridge_token_chain33_YCC
    offline_deploy_erc20_chain33_YCC
    offline_create_bridge_token_eth_YCC
    offline_deploy_erc20_chain33_ZBC
    offline_create_bridge_token_eth_ZBC

    # 重启
    restart_ebrelayerA

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function restart_ebrelayerA() {
    # 重启
    kill_ebrelayer "./ebrelayer ./relayer.toml"
    start_ebrelayerA

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    sleep ${maturityDegree}
}

function offline_set_offline_token_Eth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # echo '2:#配置自动转离线钱包(eth, 20, 50%)'
    ${Boss4xCLI} ethereum offline set_offline_token -s ETH -m 20 -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_token.txt"

#    result=$(${CLIA} ethereum multisign set_offline_token -s ETH -m 20)
#    cli_ret "${result}" "set_offline_token -s ETH -m 20"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_set_offline_token_EthYcc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    ${Boss4xCLI} ethereum offline set_offline_token -s YCC -m 100 -p 40 -t "${ethereumYccTokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_token.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_Eth_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # transfer
    # shellcheck disable=SC2154
#    ${CLIA} ethereum multisign transfer -a 3 -r "${ethBridgeBank}" -o "${ethValidatorAddrKeyB}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 3 -r "${ethBridgeBank}" -c "${multisignEthAddr}" -d "${ethValidatorAddrB}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    # shellcheck disable=SC2154
    ethereum_offline_sign_send create_multisign_tx.txt "${ethValidatorAddrKeyB}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "16"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}")
    cli_ret "${result}" "balance" ".balance" "20"

    # transfer
    # shellcheck disable=SC2154
#    ${CLIA} ethereum multisign transfer -a 5 -r "${ethMultisignA}" -o "${ethValidatorAddrKeyB}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 5 -r "${ethMultisignA}" -c "${multisignEthAddr}" -d "${ethValidatorAddrB}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethValidatorAddrKeyB}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethMultisignA}")
    cli_ret "${result}" "balance" ".balance" "105"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}")
    cli_ret "${result}" "balance" ".balance" "15"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_EthYcc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer
    # multisignEthAddr 要有手续费
    ${CLIA} ethereum transfer -k "${ethDeployKey}" -m 10 -r "${multisignEthAddr}"
    sleep 10

     # transfer
#    ${CLIA} ethereum multisign transfer -a 8 -r "${ethBridgeBank}" -o "${ethValidatorAddrKeyB}" -t "${ethereumYccTokenAddr}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 8 -r "${ethBridgeBank}" -c "${multisignEthAddr}" -d "${ethValidatorAddrB}" -t "${ethereumYccTokenAddr}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
   ethereum_offline_sign_send create_multisign_tx.txt "${ethValidatorAddrKeyB}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "80"

    # transfer
#    ${CLIA} ethereum multisign transfer -a 10 -r "${ethMultisignA}" -o "${ethValidatorAddrKeyB}" -t "${ethereumYccTokenAddr}" -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline multisign_transfer_prepare -a 10 -r "${ethMultisignA}" -c "${multisignEthAddr}" -d "${ethValidatorAddrB}" -t "${ethereumYccTokenAddr}"
    ${Boss4xCLI} ethereum offline sign_multisign_tx -k "${ethMultisignKeyA},${ethMultisignKeyB},${ethMultisignKeyC},${ethMultisignKeyD}"
    ${Boss4xCLI} ethereum offline create_multisign_tx
    ethereum_offline_sign_send create_multisign_tx.txt "${ethValidatorAddrKeyB}"
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethMultisignA}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "10"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "70"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_set_offline_token_Bty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== chain33 端 configLockedTokenOfflineSave BTY ======${NOC}"
#    echo '2:#配置自动转离线钱包(bty, 1000, 50%)'
#    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "configLockedTokenOfflineSave(${chain33BtyTokenAddr},BTY,100000000000,50)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"
    ${Boss4xCLI} chain33 offline set_offline_token -c "${chain33BridgeBank}" -s BTY -m 100000000000 -p 50 -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "chain33_set_offline_token.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_Bty_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer test
    # shellcheck disable=SC2154
#    hash=$(${CLIA} chain33 multisign transfer -a 100 -r "${chain33BridgeBank}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 100 -r "${chain33BridgeBank}" -m "${multisignChain33Addr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} account balance -a "${multisignChain33Addr}" -e evm)
    balance_ret "${result}" "997.5000"
    result=$(${Chain33Cli} account balance -a "${chain33BridgeBank}" -e evm)
    balance_ret "${result}" "632.5000"

    # shellcheck disable=SC2154
#    hash=$(${CLIA} chain33 multisign transfer -a 100 -r "${chain33MultisignA}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 100 -r "${chain33MultisignA}" -m "${multisignChain33Addr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} account balance -a "${multisignChain33Addr}" -e evm)
    balance_ret "${result}" "897.5000"
    result=$(${Chain33Cli} account balance -a "${chain33MultisignA}" -e evm)
    balance_ret "${result}" "100.0000"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_set_offline_token_Chain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    echo -e "${GRE}===== chain33 端 configLockedTokenOfflineSave ERC20 YCC ======${NOC}"
#    echo '2:#配置自动转离线钱包(YCC, 100, 60%)'
#    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "configLockedTokenOfflineSave(${chain33YccErc20Addr},YCC,10000000000,60)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"

    ${Boss4xCLI} chain33 offline set_offline_token -c "${chain33BridgeBank}" -t "${chain33YccErc20Addr}" -s YCC -m 10000000000 -p 60 -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "chain33_set_offline_token.txt"

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
    result=$(${Chain33Cli} evm abi call -a "${chain33YccErc20Addr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "6200000000"
    result=$(${Chain33Cli} evm abi call -a "${chain33YccErc20Addr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
    is_equal "${result}" "30800000000"

#    hash=$(${CLIA} chain33 multisign transfer -a 5 -r "${chain33MultisignA}" -t "${chain33YccErc20Addr}" -k "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" | jq -r ".msg")
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 5 -r "${chain33MultisignA}" -m "${multisignChain33Addr}" -t "${chain33YccErc20Addr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} evm abi call -a "${chain33YccErc20Addr}" -c "${chain33MultisignA}" -b "balanceOf(${chain33MultisignA})")
    is_equal "${result}" "500000000"
    result=$(${Chain33Cli} evm abi call -a "${chain33YccErc20Addr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
    is_equal "${result}" "30300000000"

    # 判断 ETH 这端是否金额一致
    result=$(${CLIA} ethereum balance -o "${ethDeployAddr}" -t "${ethBridgeToeknYccAddr}" )
    cli_ret "${result}" "balance" ".balance" "370"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockAndtransfer_multisign_chain33Bty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    offline_set_offline_token_Bty
    lock_multisign_Bty_test
    offline_transfer_multisign_Bty_test

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockAndtransfer_multisign_chain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    offline_set_offline_token_Chain33Ycc
    lock_multisign_Chain33Ycc_test
    offline_transfer_multisign_Chain33Ycc_test

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockAndtransfer_multisign_ethereumEth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock multisign ETH ======${NOC}"
    offline_set_offline_token_Eth
    # 重启 nonce 会不统一 要重启一下
    restart_ebrelayerA
    lock_multisign_Eth_test
    offline_transfer_multisign_Eth_test
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockAndtransfer_multisign_ethereumBty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock multisign YCC ======${NOC}"
    offline_set_offline_token_EthYcc
    # 重启
    restart_ebrelayerA
    lock_multisign_EthYcc
    offline_transfer_multisign_EthYcc
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2120
function mainTest() {
    if [[ $# -ge 1 && "${1}" != "" ]]; then
        chain33ID="${1}"
    fi

    StartChain33
    start_trufflesuite
    StartRelayerAndOfflineDeploy

    TestChain33ToEthAssets
    TestChain33ToEthZBCAssets
    TestETH2Chain33Assets
    TestETH2Chain33Ycc

    lockAndtransfer_multisign_chain33Bty
    lockAndtransfer_multisign_chain33Ycc
    lockAndtransfer_multisign_ethereumEth
    lockAndtransfer_multisign_ethereumBty
}

mainTest "${1}"

