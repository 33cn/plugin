#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

source "./publicTest.sh"
source "./relayerPublic.sh"

# ETH 部署合约者的私钥 用于部署合约时签名使用
#ethDeployAddr="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
ethDeployKey="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"


# chain33 部署合约者的私钥 用于部署合约时签名使用
chain33DeployAddr="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"
chain33DeployKey="0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"

ethValidatorAddrKeyA="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

chain33ReceiverAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
#chain33ReceiverAddrKey="4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"

Chain33Cli="../../chain33-cli"
chain33BridgeBank=""
ethBridgeBank=""
#chain33BtyTokenAddr="1111111111111111111114oLvT2"
#chain33EthTokenAddr=""
#ethereumBtyTokenAddr=""
#chain33YccTokenAddr=""
ethereumYccTokenAddr=""
multisignChain33Addr=""
multisignEthAddr=""

CLIA="./ebcli_A"
chain33ID=33

# shellcheck disable=SC2034
{
bseMultisignA=0x0f2e821517D4f64a012a04b668a6b1aa3B262e08
bseMultisignB=0x21B5f4C2F6Ff418fa0067629D9D76AE03fB4a2d2
bseMultisignC=0xee760B2E502244016ADeD3491948220B3b1dd789
bseMultisignKeyA=f934e9171c5cf13b35e6c989e95f5e95fa471515730af147b66d60fbcd664b7c
bseMultisignKeyB=2bcf3e23a17d3f3b190a26a098239ad2d20267a673440e0f57a23f44f94b77b9
bseMultisignKeyC=e5f8caae6468061c17543bc2205c8d910b3c71ad4d055105cde94e88ccb4e650
TestNodeAddr="https://data-seed-prebsc-1-s1.binance.org:8545/"
}

function deployMultisign() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    for name in A B C D; do
        eval chain33MultisignKey=\$chain33MultisignKey${name}
        eval chain33Multisign=\$chain33Multisign${name}
        # shellcheck disable=SC2154
        result=$(${Chain33Cli} account import_key -k "${chain33MultisignKey}" -l multisignAddr$name)
        # shellcheck disable=SC2154
        check_addr "${result}" "${chain33Multisign}"

        # chain33Multisign 要有手续费
        hash=$(${Chain33Cli} send coins transfer -a 10 -t "${chain33Multisign}" -k "${chain33DeployAddr}")
        check_tx "${Chain33Cli}" "${hash}"
        result=$(${Chain33Cli} account balance -a "${chain33Multisign}" -e coins)
        balance_ret "${result}" "10.0000"
    done

    echo -e "${GRE}=========== 部署 chain33 离线钱包合约 ===========${NOC}"
    result=$(${CLIA} chain33 multisign deploy)
    cli_ret "${result}" "chain33 multisign deploy"
    multisignChain33Addr=$(echo "${result}" | jq -r ".msg")

    # shellcheck disable=SC2154
    result=$(${CLIA} chain33 multisign setup -k "${chain33DeployKey}" -o "${chain33MultisignA},${chain33MultisignB},${chain33MultisignC},${chain33MultisignD}")
    cli_ret "${result}" "chain33 multisign setup"

    # multisignChain33Addr 要有手续费
    hash=$(${Chain33Cli} send coins transfer -a 10 -t "${multisignChain33Addr}" -k "${chain33DeployAddr}")
    check_tx "${Chain33Cli}" "${hash}"
    result=$(${Chain33Cli} account balance -a "${multisignChain33Addr}" -e coins)
    balance_ret "${result}" "10.0000"

    hash=$(${Chain33Cli} evm call -f 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "configOfflineSaveAccount(${multisignChain33Addr})" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

#    echo -e "${GRE}=========== 部署 ETH 离线钱包合约 ===========${NOC}"
#    result=$(${CLIA} ethereum multisign deploy)
#    cli_ret "${result}" "ethereum multisign deploy"
#    multisignEthAddr=$(echo "${result}" | jq -r ".msg")
#
#    result=$(${CLIA} ethereum multisign setup -k "${ethDeployKey}" -o "${bseMultisignA},${bseMultisignB},${bseMultisignC}")
#    cli_ret "${result}" "ethereum multisign setup"
#
#    result=$(${CLIA} ethereum multisign set_offline_addr -s "${multisignEthAddr}")
#    cli_ret "${result}" "set_offline_addr"
#
#    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lock_eth_balance() {
    local lockAmount=$1
    local bridgeBankBalance=$2
    local multisignBalance=$3

    result=$(${CLIA} ethereum lock -m "${lockAmount}" -k "${ethValidatorAddrKeyA}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

     # eth 等待 10 个区块
    eth_block_wait 2

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" )
#    cli_ret "${result}" "balance" ".balance" "${bridgeBankBalance}"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" )
#    cli_ret "${result}" "balance" ".balance" "${multisignBalance}"
}

function lockEth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # echo '2:#配置自动转离线钱包(BNB, 4, 50%)'
#    result=$(${CLIA} ethereum multisign set_offline_token -s ETH -m 4)
#    cli_ret "${result}" "set_offline_token -s ETH -m 4"

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" )
#    cli_ret "${result}" "balance" ".balance" "0"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" )
#    cli_ret "${result}" "balance" ".balance" "0"

#    lock_eth_balance 4 2 2
    lock_eth_balance 2 4 2
#    lock_eth_balance 1 10 10
#    lock_eth_balance 16 13 23

    # transfer
    hash=$(./ebcli_A ethereum multisign transfer -a 1 -r "${ethBridgeBank}" -k "${bseMultisignKeyA},${bseMultisignKeyB},${bseMultisignKeyC}")
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" )
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" )

    result=$(${CLIA} ethereum balance -o "${bseMultisignA}" )
    hash=$(./ebcli_A ethereum multisign transfer -a 1 -r "${bseMultisignA}" -k "${bseMultisignKeyA},${bseMultisignKeyB},${bseMultisignKeyC}")
    result=$(${CLIA} ethereum balance -o "${bseMultisignA}" )
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" )

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lock_eth_ycc_balance() {
    local lockAmount=$1
    local bridgeBankBalance=$2
    local multisignBalance=$3

    result=$(${CLIA} ethereum lock -m "${lockAmount}" -k "${ethDeployKey}" -r "${chain33ReceiverAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 10 个区块
    eth_block_wait 2

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "${bridgeBankBalance}"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "${multisignBalance}"
}

function lockEthYcc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # echo '2:#配置自动转离线钱包(ycc, 100, 40%)'
    result=$(${CLIA} ethereum multisign set_offline_token -s YCC -m 100 -p 40 -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "set_offline_token -s YCC -m 100"

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"
    result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    lock_eth_ycc_balance 70 70 0
    lock_eth_ycc_balance 30 60 40
    lock_eth_ycc_balance 60 72 88

    # transfer
    # multisignEthAddr 要有手续费
    ./ebcli_A ethereum transfer -k "${ethDeployKey}" -m 10 -r "${multisignEthAddr}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function StartRelayerOnBsc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"


    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

    BridgeRegistryOnEth=0x5331F912027057fBE8139D91B225246e8159232f
    ethBridgeBank=0xC65B02a22B714b55D708518E2426a22ffB79113d
    multisignEthAddr=0xbf271b2B23DA4fA8Dc93Ce86D27dd09796a7Bf54

# shellcheck disable=SC2120
function mainTest() {
    if [[ $# -ge 1 && "${1}" != "" ]]; then
        chain33ID="${1}"
    fi
    StartChain33

    # shellcheck disable=SC2009
    pid=$(ps -ef | grep "ebrelayer" | grep -v 'grep' | awk '{print $2}' | xargs)
    if [ "${pid}" == "" ]; then
        kill_ebrelayer ebrelayer
        sleep 2
        rm datadir/ logs/ -rf
    fi

    # shellcheck disable=SC2155
    line=$(delete_line_show "./relayer.toml" "EthProviderCli=\"http://127.0.0.1:7545\"")
    if [ "${line}" ]; then
        sed -i ''"${line}"' a EthProviderCli="https://data-seed-prebsc-1-s1.binance.org:8545/"' "./relayer.toml"
    fi

    line=$(delete_line_show "./relayer.toml" "EthProvider=\"ws://127.0.0.1:7545/\"")
    if [ "${line}" ]; then
        sed -i ''"${line}"' a EthProvider="wss://data-seed-prebsc-1-s1.binance.org:8545/"' "./relayer.toml"
    fi



#    StartRelayer_A
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # 修改 relayer.toml 配置文件 pushName 字段
    pushNameChange "./relayer.toml"

    # 启动 ebrelayer
    start_ebrelayerA

    # 导入私钥 部署合约 设置 bridgeRegistry 地址
#    InitAndDeploy
{
    echo -e "${GRE}=========== InitAndDeploy begin ===========${NOC}"

    result=$(${CLIA} set_pwd -p 123456hzj)
    cli_ret "${result}" "set_pwd"

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    result=$(${CLIA} chain33 import_privatekey -k "${chain33DeployKey}")
    cli_ret "${result}" "chain33 import_privatekey"

    result=$(${CLIA} ethereum import_privatekey -k "${ethDeployKey}")
    cli_ret "${result}" "ethereum import_privatekey"

    # 在 chain33 上部署合约
    result=$(${CLIA} chain33 deploy)
    cli_ret "${result}" "chain33 deploy"
    BridgeRegistryOnChain33=$(echo "${result}" | jq -r ".msg")

    # 拷贝 BridgeRegistry.abi 和 BridgeBank.abi
    cp BridgeRegistry.abi "${BridgeRegistryOnChain33}.abi"
    chain33BridgeBank=$(${Chain33Cli} evm abi call -c "${chain33DeployAddr}" -b "bridgeBank()" -a "${BridgeRegistryOnChain33}")
    cp Chain33BridgeBank.abi "${chain33BridgeBank}.abi"

    # 在 Eth 上部署合约
#    result=$(${CLIA} ethereum deploy)
#    cli_ret "${result}" "ethereum deploy"
#    BridgeRegistryOnEth=$(echo "${result}" | jq -r ".msg")

    # 拷贝 BridgeRegistry.abi 和 BridgeBank.abi
#    cp BridgeRegistry.abi "${BridgeRegistryOnEth}.abi"
#    result=$(${CLIA} ethereum bridgeBankAddr)
#    ethBridgeBank=$(echo "${result}" | jq -r ".addr")
    cp EthBridgeBank.abi "${ethBridgeBank}.abi"

    # 修改 relayer.toml 字段
    updata_relayer "BridgeRegistryOnChain33" "${BridgeRegistryOnChain33}" "./relayer.toml"

    line=$(delete_line_show "./relayer.toml" "BridgeRegistry=")
    if [ "${line}" ]; then
        sed -i ''"${line}"' a BridgeRegistry="'"${BridgeRegistryOnEth}"'"' "./relayer.toml"
    fi

    echo -e "${GRE}=========== InitAndDeploy end ===========${NOC}"
}

    # 重启
    kill_ebrelayer ebrelayer
    start_ebrelayerA

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"

    deployMultisign

    lockEth
#    lockEthYcc
}

mainTest "${1}"

#lockEth
