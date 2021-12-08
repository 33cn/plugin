#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./offlinePublic.sh"

maturityDegree=10

chain33BridgeBank=""
ethBridgeBank=""
chain33BtyERC20TokenAddr="1111111111111111111114oLvT2"
ethereumBtyBridgeTokenAddr=""
chain33EthBridgeTokenAddr=""

ethereumBycERC20TokenAddr=""
chain33BycBridgeTokenAddr=""

ethereumUSDTERC20TokenAddr=""
chain33USDTBridgeTokenAddr=""

ethereumBUSDERC20TokenAddr="0xe9e7cea3dedca5984780bafc599bd69add087d56"
chain33BUSDBridgeTokenAddr=""

chain33YccERC20TokenAddr=""
ethereumYccBridgeTokenAddr=""

chain33ZbcERC20TokenAddr=""
ethereumZbcBridgeTokenAddr=""

BridgeRegistryOnChain33=""
BridgeRegistryOnEth=""

multisignChain33Addr=""
multisignEthAddr=""
chain33ID=0

EvmxgoBoss4xCLI="./evmxgoboss4x"
BscProvider="wss://bsc-ws-node.nariox.org:443"
BscProviderUrl="https://bsc-dataseed.binance.org/"
#BscProvider="wss://ws-testnet.hecochain.com"
#BscProviderUrl="https://http-testnet.hecochain.com"

function start_docker_ebrelayerA() {
    # shellcheck disable=SC2154
    docker cp "./relayer.toml" "${dockerNamePrefix}_ebrelayera_1":/root/relayer.toml
    start_docker_ebrelayer "${dockerNamePrefix}_ebrelayera_1" "/root/ebrelayer" "./ebrelayera.log"
    sleep 5
}

# start ebrelayer B C D
function updata_toml_start_bcd() {
    for name in b c d; do
        local file="./relayer$name.toml"
        cp './relayer.toml' "${file}"

        # 删除配置文件中不需要的字段
        for deleteName in "deploy4chain33" "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers" "deploy" "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers"; do
            delete_line "${file}" "${deleteName}"
        done

        pushNameChange "${file}"

        pushHost=$(get_docker_addr "${dockerNamePrefix}_ebrelayer${name}_1")
        line=$(delete_line_show "${file}" "pushHost")
        sed -i ''"${line}"' a pushHost="http://'"${pushHost}"':20000"' "${file}"

        line=$(delete_line_show "${file}" "pushBind")
        sed -i ''"${line}"' a pushBind="'"${pushHost}"':20000"' "${file}"

        docker cp "${file}" "${dockerNamePrefix}_ebrelayer${name}_1":/root/relayer.toml
        start_docker_ebrelayer "${dockerNamePrefix}_ebrelayer${name}_1" "/root/ebrelayer" "./ebrelayer${name}.log"

        CLI="docker exec ${dockerNamePrefix}_ebrelayer${name}_1 /root/ebcli_A"
        result=$(${CLI} set_pwd -p 123456hzj)
        cli_ret "${result}" "set_pwd"

        result=$(${CLI} unlock -p 123456hzj)
        cli_ret "${result}" "unlock"

        eval chain33ValidatorKey=\$chain33ValidatorKey${name}
        # shellcheck disable=SC2154
        result=$(${CLI} chain33 import_privatekey -k "${chain33ValidatorKey}")
        cli_ret "${result}" "chain33 import_privatekey"

        eval ethValidatorAddrKey=\$ethValidatorAddrKey${name}
        # shellcheck disable=SC2154
        result=$(${CLI} ethereum import_privatekey -k "${ethValidatorAddrKey}")
        cli_ret "${result}" "ethereum import_privatekey"
    done
}

function restart_ebrelayerA() {
    # 重启
    kill_docker_ebrelayer "${dockerNamePrefix}_ebrelayera_1"
    sleep 1
    start_docker_ebrelayerA

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"
}

# chain33 lock BTY, eth burn BTY
function TestChain33ToEthAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== chain33 lock BTY, eth burn BTY ===========${NOC}"
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumBtyBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 原来的地址金额
    result=$(${Chain33Cli} asset balance -a "${chain33TestAddr1}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "500.0000"

    # chain33 lock bty
    hash=$(${Chain33Cli} send evm call -f 1 -a 5 -k "${chain33TestAddr1}" -e "${chain33BridgeBank}" -p "lock(${ethTestAddr1}, ${chain33BtyERC20TokenAddr}, 500000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    # 原来的地址金额 减少了 5
    result=$(${Chain33Cli} asset balance -a "${chain33TestAddr1}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "495.0000"

    # chain33BridgeBank 是否增加了 5
    result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "5.0000"

    sleep 4

    # eth 这端 金额是否增加了 5
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumBtyBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "5"

    # eth burn
    result=$(${CLIA} ethereum burn -m 3 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumBtyBridgeTokenAddr}") #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    sleep 4

    # eth 这端 金额是否减少了 3
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumBtyBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "2"

    sleep ${maturityDegree}

    # 接收的地址金额 变成了 3
    result=$(${Chain33Cli} asset balance -a "${chain33ReceiverAddr}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "3.0000"

    # chain33BridgeBank 是否减少了 3
    result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "2.0000"

    # eth burn 2
    result=$(${CLIA} ethereum burn -m 2 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumBtyBridgeTokenAddr}") #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    sleep 4

    # eth 这端 金额是否减少了
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumBtyBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    sleep ${maturityDegree}

    # 接收的地址金额 变成了 5
    result=$(${Chain33Cli} asset balance -a "${chain33ReceiverAddr}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "5.0000"

    # chain33BridgeBank 是否减少了 3
    result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "0.0000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# chain33 lock ZBC, eth burn ZBC
function TestChain33ToEthZBCAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== chain33 lock ZBC, eth burn ZBC ===========${NOC}"
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumZbcBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 原来的地址金额
    result=$(${Chain33Cli} evm query -a "${chain33ZbcERC20TokenAddr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "0"

    # chain33 lock ZBC
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33TestAddr1}" -e "${chain33BridgeBank}" -p "lock(${ethTestAddr1}, ${chain33ZbcERC20TokenAddr}, 900000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    # chain33BridgeBank 是否增加了 9
    result=$(${Chain33Cli} evm query -a "${chain33ZbcERC20TokenAddr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "900000000"

    sleep 4

    # eth 这端 金额是否增加了 9
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumZbcBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "9"

    # eth burn
    result=$(${CLIA} ethereum burn -m 8 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumZbcBridgeTokenAddr}") #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    sleep 4

    # eth 这端 金额是否减少了 1
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumZbcBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "1"

    sleep ${maturityDegree}

    # 接收的地址金额 变成了 8
    result=$(${Chain33Cli} evm query -a "${chain33ZbcERC20TokenAddr}" -c "${chain33ReceiverAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "800000000"

    # chain33BridgeBank 是否减少了 1
    result=$(${Chain33Cli} evm query -a "${chain33ZbcERC20TokenAddr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
    is_equal "${result}" "100000000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn
function TestETH2Chain33Assets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn ===========${NOC}"
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0"

    result=$(${CLIA} ethereum lock -m 0.002 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0.002"

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
#    is_equal "${result}" "2000000000000000"

    # 原来的数额
    result=$(${CLIA} ethereum balance -o "${ethTestAddr2}")

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 0.0003 -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33EthBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
#    is_equal "${result}" "1700000000000000"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0.0017"

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum 6'
    result=$(${CLIA} chain33 burn -m 0.0017 -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33EthBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "0"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33Byc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 byc 资产,然后在 chain33 上 burn ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 7个 YCC
    result=$(${CLIA} ethereum lock -m 7 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    # 查询 ETH 这端 bridgeBank 地址 7 YCC
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "7"

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 7 * le8
    is_equal "${result}" "700000000"

    # 原来的数额 0
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 5 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33BycBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 7-5 * le8
    is_equal "${result}" "200000000"

    # 查询 ETH 这端 bridgeBank 地址 2
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "2"

    # 更新后的金额 5
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "5"

    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 2 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33BycBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 7-5 * le8
    is_equal "${result}" "0"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 更新后的金额 5
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBycERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "7"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33USDT() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 USDT 资产,然后在 chain33 上 burn ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
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

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33USDTBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 12 * le8
    is_equal "${result}" "1200000000"

    # 原来的数额 0
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 5 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33USDTBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

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

    sleep ${maturityDegree}

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

function TestETH2Chain33BUSD() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 BUSD 资产,然后在 chain33 上 burn ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBUSDERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 12个 BUSD
    result=$(${CLIA} ethereum lock -m 3 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumBUSDERC20TokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    # 查询 ETH 这端 bridgeBank 地址 12 BUSD
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBUSDERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "3"

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33BUSDBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 12 * le8
#    is_equal "${result}" "3000000000000000000"

    # 原来的数额 0
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBUSDERC20TokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "0"

    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 1 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33BUSDBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33BUSDBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 12-5 * le8
#    is_equal "${result}" "2000000000000000000"

    # 查询 ETH 这端 bridgeBank 地址 7
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBUSDERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "2"

    # 更新后的金额 5
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBUSDERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "1"

    echo '#5.burn BUSD from Chain33 BUSD(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 2 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33BUSDBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33BUSDBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "0"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBUSDERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 更新后的金额 12
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBUSDERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "3"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2120
function offline_set_offline_token_Bty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== chain33 端 configLockedTokenOfflineSave BTY ======${NOC}"
    #    echo '2:#配置自动转离线钱包(bty, 100, 50%)'
    local threshold=10000000000
    local percents=50
    if [[ $# -eq 2 ]]; then
        threshold=$1
        percents=$2
    fi
    # shellcheck disable=SC2086
    ${Boss4xCLI} chain33 offline set_offline_token -c "${chain33BridgeBank}" -s BTY -m ${threshold} -p ${percents} -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "chain33_set_offline_token.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2120
function offline_set_offline_token_Chain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== chain33 端 configLockedTokenOfflineSave ERC20 YCC ======${NOC}"
    #    echo '2:#配置自动转离线钱包(YCC, 100, 60%)'
    local threshold=10000000000
    local percents=60
    if [[ $# -eq 2 ]]; then
        threshold=$1
        percents=$2
    fi
    # shellcheck disable=SC2086
    ${Boss4xCLI} chain33 offline set_offline_token -c "${chain33BridgeBank}" -t "${chain33YccERC20TokenAddr}" -s YCC -m ${threshold} -p ${percents} -k "${chain33DeployKey}" --chainID "${chain33ID}"
    chain33_offline_send "chain33_set_offline_token.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2120
function offline_set_offline_token_Eth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # echo '2:#配置自动转离线钱包(eth, 20, 50%)'
    local threshold=0.002
    local percents=50
    if [[ $# -eq 2 ]]; then
        threshold=$1
        percents=$2
    fi
    # shellcheck disable=SC2086
    ${Boss4xCLI} ethereum offline set_offline_token -s BNB -m ${threshold} -p ${percents} -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_token.txt"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2120
function offline_set_offline_token_EthByc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    local threshold=100
    local percents=40
    if [[ $# -eq 2 ]]; then
        threshold=$1
        percents=$2
    fi
    # shellcheck disable=SC2086
    ${Boss4xCLI} ethereum offline set_offline_token -s BYC -m ${threshold} -p ${percents} -t "${ethereumBycERC20TokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_token.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2120
function offline_set_offline_token_EthUSDT() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    local threshold=100
    local percents=40
    if [[ $# -eq 2 ]]; then
        threshold=$1
        percents=$2
    fi
    # shellcheck disable=SC2086
    ${Boss4xCLI} ethereum offline set_offline_token -s USDT -m ${threshold} -p ${percents} -t "${ethereumUSDTERC20TokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_token.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_transfer_multisign_Bty_test() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # transfer test
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 50 -r "${chain33BridgeBank}" -m "${multisignChain33Addr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "103.2500"
    result=$(${Chain33Cli} asset balance -a "${multisignChain33Addr}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "59.7500"

    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 10 -r "${chain33MultisignA}" -m "${multisignChain33Addr}"
    # shellcheck disable=SC2154
    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
    chain33_offline_send "multisign_transfer.txt"
    sleep 10
    result=$(${Chain33Cli} asset balance -a "${chain33MultisignA}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "10.0000"
    result=$(${Chain33Cli} asset balance -a "${multisignChain33Addr}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "49.7500"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# lock bty 判断是否转入多签地址金额是否正确
function lock_bty_multisign_docker() {
    local lockAmount=$1
    local lockAmount2="${1}00000000"
    hash=$(${Chain33Cli} send evm call -f 1 -a "${lockAmount}" -k "${chain33TestAddr1}" -e "${chain33BridgeBank}" -p "lock(${ethTestAddr1}, ${chain33BtyERC20TokenAddr}, ${lockAmount2})" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    if [[ $# -eq 3 ]]; then
        local bridgeBankBalance=$2
        local multisignBalance=$3

        result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
        is_equal "${result}" "${bridgeBankBalance}"
        result=$(${Chain33Cli} asset balance -a "${multisignChain33Addr}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
        is_equal "${result}" "${multisignBalance}"
    fi
}

function lockBty() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== chain33 端 lock BTY ======${NOC}"
    #    echo '2:#配置自动转离线钱包(bty, 100, 50%)'
    offline_set_offline_token_Bty

    lock_bty_multisign_docker 33 "33.0000" "0.0000"
    lock_bty_multisign_docker 80 "56.5000" "56.5000"
    lock_bty_multisign_docker 50 "53.2500" "109.7500"

    # transfer test
    offline_transfer_multisign_Bty_test
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockChain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== chain33 端 lock ERC20 YCC ======${NOC}"
    #    echo '2:#配置自动转离线钱包(YCC, 100, 60%)'
    offline_set_offline_token_Chain33Ycc

    lock_chain33_ycc_multisign 30 30 0
    lock_chain33_ycc_multisign 70 40 60
    lock_chain33_ycc_multisign 260 120 240
    lock_chain33_ycc_multisign 10 52 318

    # transfer test
    offline_transfer_multisign_Chain33Ycc_test
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockEth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock ETH ======${NOC}"
    # echo '2:#配置自动转离线钱包(eth, 20, 50%)'
    offline_set_offline_token_Eth 0.002 50

    # 重启 nonce 会不统一 要重启一下
    restart_ebrelayerA

    lock_eth_multisign 0.001 0.001 0
    lock_eth_multisign 0.001 0.001 0.001
    lock_eth_multisign 0.001 0.001 0.002

    # transfer
    # shellcheck disable=SC2154
    offline_transfer_multisign_Eth_test
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockEthByc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock ERC20 Byc ======${NOC}"
    # echo '2:#配置自动转离线钱包(ycc, 100, 40%)'
    offline_set_offline_token_EthByc
    # 重启 nonce 会不统一 要重启一下
    restart_ebrelayerA

    lock_ethereum_byc_multisign 70 70 0
    lock_ethereum_byc_multisign 30 60 40
    lock_ethereum_byc_multisign 60 72 88

    # multisignEthAddr 要有手续费
    ${CLIA} ethereum transfer -k "${ethDeployKey}" -m 0.0001 -r "${multisignEthAddr}"
    sleep 10

    # transfer
    offline_transfer_multisign_EthByc
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockEthUSDT() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock ERC20 USDT ======${NOC}"
    # echo '2:#配置自动转离线钱包(ycc, 100, 40%)'
    offline_set_offline_token_EthUSDT
    # 重启 nonce 会不统一 要重启一下
    restart_ebrelayerA

    lock_ethereum_usdt_multisign 70 70 0
    lock_ethereum_usdt_multisign 30 60 40
    lock_ethereum_usdt_multisign 60 72 88

    # multisignEthAddr 要有手续费
    ${CLIA} ethereum transfer -k "${ethDeployKey}" -m 0.0001 -r "${multisignEthAddr}"
    sleep 10

    # transfer
    offline_transfer_multisign_EthUSDT
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function up_relayer_toml() {
    # 修改 relayer.toml 配置文件 pushName 字段
    pushNameChange "./relayer.toml"
    # 修改 relayer.toml 配置文件 initPowers
    validators_config

    # change EthProvider url
    dockerAddr=$(get_docker_addr "${dockerNamePrefix}_ganachetest_1")

    # 修改 relayer.toml 配置文件
    updata_relayer_a_toml "${dockerAddr}" "${dockerNamePrefix}_ebrelayera_1" "./relayer.toml"
    # shellcheck disable=SC2155
    local line=$(delete_line_show "./relayer.toml" 'EthProvider="ws:')
    sed -i ''"${line}"' a EthProvider="'"${BscProvider}"'"' "./relayer.toml"

    line=$(delete_line_show "./relayer.toml" 'EthProviderCli="http:')
    sed -i ''"${line}"' a EthProviderCli="'"${BscProviderUrl}"'"' "./relayer.toml"

    # 删除私钥
    delete_line "./relayer.toml" "deployerPrivateKey="
    delete_line "./relayer.toml" "deployerPrivateKey="

    # para
    # shellcheck disable=SC2155
    local line=$(delete_line_show "./relayer.toml" "chain33Host")
    # 在第 line 行后面 新增合约地址
    docker_chain33_ip=$(get_docker_addr "${dockerNamePrefix}_chain33_1")
    sed -i ''"${line}"' a chain33Host="http://'"${docker_chain33_ip}"':8901"' "./relayer.toml"

    # shellcheck disable=SC2155
    local line=$(delete_line_show "./relayer.toml" "ChainName")
    # 在第 line 行后面 新增合约地址
    sed -i ''"${line}"' a ChainName="user.p.para."' "./relayer.toml"

#    # shellcheck disable=SC2155
#    local line=$(delete_line_show "./relayer.toml" "maturityDegree=10")
#    sed -i ''"${line}"' a maturityDegree=1' "./relayer.toml"
#
#    # shellcheck disable=SC2155
#    local line=$(delete_line_show "./relayer.toml" "EthMaturityDegree=10")
#    sed -i ''"${line}"' a EthMaturityDegree=1' "./relayer.toml"
}

function StartDockerRelayerDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # 修改 relayer.toml 配置文件
    up_relayer_toml

    # 启动 ebrelayer
    start_docker_ebrelayerA

    # 部署合约 设置 bridgeRegistry 地址
    InitAndOfflineDeploy

    # 设置 ethereum symbol
    ${Boss4xCLI} ethereum offline set_symbol -s "BNB" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_symbol.txt"

    # 设置离线多签数据
    Chain33Cli=${MainCli}
    initMultisignChain33Addr
    Chain33Cli=${Para8901Cli}
    offline_setupChain33Multisign
    offline_setupEthMultisign
    Chain33Cli=${MainCli}
    transferChain33MultisignFee
    Chain33Cli=${Para8901Cli}

    # shellcheck disable=SC2086
    {
        docker cp "${BridgeRegistryOnChain33}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${BridgeRegistryOnChain33}.abi
        docker cp "${chain33BridgeBank}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33BridgeBank}.abi
        docker cp "${BridgeRegistryOnEth}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${BridgeRegistryOnEth}.abi
        docker cp "${ethBridgeBank}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${ethBridgeBank}.abi
    }

    # 重启
    restart_ebrelayerA

    # start ebrelayer B C D
    updata_toml_start_bcd

    # 设置 token 地址
#    offline_create_bridge_token_eth_BTY
#    offline_deploy_erc20_eth_BYC
#    offline_deploy_erc20_eth_USDT
#    offline_create_bridge_token_eth_YCC
#    offline_create_bridge_token_eth_ZBC
#
#    offline_create_bridge_token_chain33_ETH "BNB"
#    offline_create_bridge_token_chain33_BYC
#    offline_deploy_erc20_chain33_YCC
#    offline_deploy_erc20_chain33_ZBC
#    offline_create_bridge_token_chain33_USDT

    offline_create_bridge_token_chain33_BUSD

    ${Boss4xCLI} ethereum offline create_add_lock_list -s BUSD -t "${chain33BUSDBridgeTokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "create_add_lock_list.txt"

    # shellcheck disable=SC2086
    {
#        docker cp "${chain33EthBridgeTokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33EthBridgeTokenAddr}.abi
#        docker cp "${chain33BycBridgeTokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33BycBridgeTokenAddr}.abi
#        docker cp "${chain33USDTBridgeTokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33USDTBridgeTokenAddr}.abi
        docker cp "${chain33BUSDBridgeTokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33BUSDBridgeTokenAddr}.abi
#        docker cp "${chain33YccERC20TokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33YccERC20TokenAddr}.abi
#        docker cp "${ethereumYccBridgeTokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${ethereumYccBridgeTokenAddr}.abi
    }

    # 重启,因为relayerA的验证人地址和部署人的地址是一样的,所以需要重新启动relayer,更新nonce
    restart_ebrelayerA

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function echo_addrs() {
    echo -e "${GRE}=========== echo contract addrs ===========${NOC}"
    echo -e "${GRE}BridgeRegistryOnChain33: ${BridgeRegistryOnChain33} ${NOC}"
    echo -e "${GRE}BridgeRegistryOnEth: ${BridgeRegistryOnEth} ${NOC}"
    echo -e "${GRE}chain33BridgeBank: ${chain33BridgeBank} ${NOC}"
    echo -e "${GRE}ethBridgeBank: ${ethBridgeBank} ${NOC}"
    echo -e "${GRE}chain33BtyERC20TokenAddr: ${chain33BtyERC20TokenAddr} ${NOC}"
    echo -e "${GRE}chain33EthBridgeTokenAddr: ${chain33EthBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}ethereumBtyBridgeTokenAddr: ${ethereumBtyBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}ethereumBycERC20TokenAddr: ${ethereumBycERC20TokenAddr} ${NOC}"
    echo -e "${GRE}chain33BycBridgeTokenAddr: ${chain33BycBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}ethereumUSDTERC20TokenAddr: ${ethereumUSDTERC20TokenAddr} ${NOC}"
    echo -e "${GRE}chain33USDTBridgeTokenAddr: ${chain33USDTBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}ethereumBUSDERC20TokenAddr: ${ethereumBUSDERC20TokenAddr} ${NOC}"
    echo -e "${GRE}chain33BUSDBridgeTokenAddr: ${chain33BUSDBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}chain33YccERC20TokenAddr: ${chain33YccERC20TokenAddr} ${NOC}"
    echo -e "${GRE}ethereumYccBridgeTokenAddr: ${ethereumYccBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}chain33ZbcERC20TokenAddr: ${chain33ZbcERC20TokenAddr} ${NOC}"
    echo -e "${GRE}ethereumZbcBridgeTokenAddr: ${ethereumZbcBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}multisignChain33Addr: ${multisignChain33Addr} ${NOC}"
    echo -e "${GRE}multisignEthAddr: ${multisignEthAddr} ${NOC}"
    echo -e "${GRE}XgoBridgeRegistryOnChain33: ${XgoBridgeRegistryOnChain33} ${NOC}"
    echo -e "${GRE}XgoChain33BridgeBank: ${XgoChain33BridgeBank} ${NOC}"

    echo -e "${GRE}=========== echo don't use addrs ===========${NOC}"
    echo -e "${GRE}ethDeployAddr: ${ethDeployAddr} ${NOC}"
    echo -e "${GRE}chain33DeployAddr: ${chain33DeployAddr} ${NOC}"
    echo -e "${GRE}chain33ValidatorA: ${chain33Validatora} ${NOC}"
    echo -e "${GRE}chain33ValidatorB: ${chain33Validatorb} ${NOC}"
    echo -e "${GRE}chain33ValidatorC: ${chain33Validatorc} ${NOC}"
    echo -e "${GRE}chain33ValidatorD: ${chain33Validatord} ${NOC}"
    echo -e "${GRE}ethValidatorAddrA: ${ethValidatorAddra} ${NOC}"
    echo -e "${GRE}ethValidatorAddrB: ${ethValidatorAddrb} ${NOC}"
    echo -e "${GRE}ethValidatorAddrC: ${ethValidatorAddrc} ${NOC}"
    echo -e "${GRE}ethValidatorAddrD: ${ethValidatorAddrd} ${NOC}"

    echo -e "${GRE}=========== echo use test addrs and keys===========${NOC}"
    echo -e "${GRE}ethTestAddr1: ${ethTestAddr1} ${NOC}"
    echo -e "${GRE}ethTestAddrKey1: ${ethTestAddrKey1} ${NOC}"
    echo -e "${GRE}chain33TestAddr1: ${chain33TestAddr1} ${NOC}"
    echo -e "${GRE}chain33TestAddrKey1: ${chain33TestAddrKey1} ${NOC}"
}

function chain33_offline_send_evm() {
    # shellcheck disable=SC2154
    result=$(${EvmxgoBoss4xCLI} chain33 offline send -f "${1}")
    hash=$(echo "${result}" | jq -r ".[0].TxHash")
    # shellcheck disable=SC2154
    check_tx "${Chain33Cli}" "${hash}"
}

function DeployEvmxgo() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
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

    ${EvmxgoBoss4xCLI} chain33 offline create_add_lock_list -s BNB -t "${chain33EthBridgeTokenAddr}" -c "${XgoChain33BridgeBank}" -k "${chain33DeployKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "create_add_lock_list.txt"

    ${EvmxgoBoss4xCLI} chain33 offline create_add_lock_list -s BYC -t "${chain33BycBridgeTokenAddr}" -c "${XgoChain33BridgeBank}" -k "${chain33DeployKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "create_add_lock_list.txt"

    ${EvmxgoBoss4xCLI} chain33 offline create_add_lock_list -s USDT -t "${chain33USDTBridgeTokenAddr}" -c "${XgoChain33BridgeBank}" -k "${chain33DeployKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "create_add_lock_list.txt"

    # 重启,需要重新启动relayer,更新nonce
    restart_ebrelayerA

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
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
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")

    # ETH 这端 lock
    result=$(${CLIA} ethereum lock -m 0.001 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    # 查询 ETH 这端 bridgeBank 地址 11 原来16
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    updateConfig "BNB" "${chain33EthBridgeTokenAddr}"
    configbridgevmxgoAddr "${XgoChain33BridgeBank}"

    ${EvmxgoBoss4xCLI} chain33 offline approve_erc20 -a 330000000000 -s "${XgoChain33BridgeBank}" -c "${chain33EthBridgeTokenAddr}" -k "${chain33ReceiverAddrKey}" -f 1 --chainID "${chain33ID}"
    chain33_offline_send_evm "approve_erc20.txt"

    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33ReceiverAddr}" -e "${XgoChain33BridgeBank}" -p "lock(${chain33TestAddr2}, ${chain33EthBridgeTokenAddr}, 100000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${XgoChain33BridgeBank})")

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function Testethereum2EVMToChain33_byc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
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

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function Testethereum2EVMToChain33_usdt() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

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

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function test_all() {
#    TestETH2Chain33Assets
#    TestChain33ToEthAssets
#    TestChain33ToEthZBCAssets
#    TestETH2Chain33Byc
#    TestETH2Chain33USDT
    TestETH2Chain33BUSD

#    lockBty
#    lockChain33Ycc
#    lockEth
#    lockEthByc
#    lockEthUSDT
#
#    # 离线多签地址转入阈值设大
#    offline_set_offline_token_Bty 100000000000000 10
#    offline_set_offline_token_Chain33Ycc 100000000000000 10
#    offline_set_offline_token_Eth 100000000000000 10
#    offline_set_offline_token_EthByc 100000000000000 10
#    offline_set_offline_token_EthUSDT 100000000000000 10

#    DeployEvmxgo
#    TestETH2EVMToChain33
#    Testethereum2EVMToChain33_byc
#    Testethereum2EVMToChain33_usdt
}

function get_cli() {
    # shellcheck disable=SC2034
    {
        docker_chain33_ip=$(get_docker_addr "${dockerNamePrefix}_chain33_1")
        MainCli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8801"
        Para8801Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para."
        Para8901Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para."

        CLIA="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/ebcli_A"
        CLIB="docker exec ${dockerNamePrefix}_ebrelayerb_1 /root/ebcli_A"
        CLIC="docker exec ${dockerNamePrefix}_ebrelayerc_1 /root/ebcli_A"
        CLID="docker exec ${dockerNamePrefix}_ebrelayerd_1 /root/ebcli_A"

#        docker_ganachetest_ip=$(get_docker_addr "${dockerNamePrefix}_ganachetest_1")
        Boss4xCLI="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/boss4x --rpc_laddr http://${docker_chain33_ip}:8901 --rpc_laddr_ethereum ${BscProviderUrl} --paraName user.p.para. --chainEthId 56"
        echo "${Boss4xCLI}"

        EvmxgoBoss4xCLI="./evmxgoboss4x --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para."
    }
}

function AllRelayerMainTest() {
    set +e

    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # 获取配置地址
    init_read_address "./addrTestConfig.toml"

    get_cli

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

    # 部署
    Chain33Cli=${Para8901Cli}
    StartDockerRelayerDeploy

    # test
    test_all

    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

