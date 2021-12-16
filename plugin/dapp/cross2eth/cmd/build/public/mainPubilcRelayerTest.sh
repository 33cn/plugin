#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./offlinePublic.sh"

# shellcheck disable=SC2034
{
    chain33BridgeBank=""
    ethBridgeBank=""
    BridgeRegistryOnChain33=""
    BridgeRegistryOnEth=""
    multisignChain33Addr=""
    multisignEthAddr=""

    chain33BtyERC20TokenAddr="1111111111111111111114oLvT2"
    ethereumBtyBridgeTokenAddr=""
    chain33EthBridgeTokenAddr=""
    ethereumUSDTERC20TokenAddr=""
    chain33USDTBridgeTokenAddr=""

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
#function TestChain33ToEthZBCAssets() {
#    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
#    echo -e "${GRE}=========== chain33 lock ZBC, eth burn ZBC ===========${NOC}"
#    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumZbcBridgeTokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "0"
#
#    # 原来的地址金额
#    result=$(${Chain33Cli} evm query -a "${chain33ZbcERC20TokenAddr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
#    is_equal "${result}" "0"
#
#    # chain33 lock ZBC
#    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33TestAddr1}" -e "${chain33BridgeBank}" -p "lock(${ethTestAddr1}, ${chain33ZbcERC20TokenAddr}, 900000000)" --chainID "${chain33ID}")
#    check_tx "${Chain33Cli}" "${hash}"
#
#    # chain33BridgeBank 是否增加了 9
#    result=$(${Chain33Cli} evm query -a "${chain33ZbcERC20TokenAddr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
#    is_equal "${result}" "900000000"
#
#    sleep 4
#
#    # eth 这端 金额是否增加了 9
#    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumZbcBridgeTokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "9"
#
#    # eth burn
#    result=$(${CLIA} ethereum burn -m 8 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumZbcBridgeTokenAddr}") #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
#    cli_ret "${result}" "burn"
#
#    sleep 4
#
#    # eth 这端 金额是否减少了 1
#    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumZbcBridgeTokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "1"
#
#    sleep ${maturityDegree}
#
#    # 接收的地址金额 变成了 8
#    result=$(${Chain33Cli} evm query -a "${chain33ZbcERC20TokenAddr}" -c "${chain33ReceiverAddr}" -b "balanceOf(${chain33ReceiverAddr})")
#    is_equal "${result}" "800000000"
#
#    # chain33BridgeBank 是否减少了 1
#    result=$(${Chain33Cli} evm query -a "${chain33ZbcERC20TokenAddr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
#    is_equal "${result}" "100000000"
#
#    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
#}

# eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn

function TestETH2Chain33Assets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 11个
    result=$(${CLIA} ethereum lock -m 11 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 4

    # 查询 ETH 这端 bridgeBank 地址 11
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "11"

    sleep ${maturityDegree}

    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 11 * le8
    is_equal "${result}" "1100000000"

    # 原来的数额
    result=$(${CLIA} ethereum balance -o "${ethTestAddr2}")
    cli_ret "${result}" "balance" ".balance" "1000"

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 5 -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33EthBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 11-5 * le8
    is_equal "${result}" "600000000"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "6"

    # 比之前多 5
    result=$(${CLIA} ethereum balance -o "${ethTestAddr2}")
    cli_ret "${result}" "balance" ".balance" "1005"

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum 6'
    result=$(${CLIA} chain33 burn -m 6 -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33EthBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33EthBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    # 结果是 11-5 * le8
    is_equal "${result}" "0"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 比之前多 5
    result=$(${CLIA} ethereum balance -o "${ethTestAddr2}")
    cli_ret "${result}" "balance" ".balance" "1011"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

#function TestETH2Chain33Byc() {
#    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
#    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 byc 资产,然后在 chain33 上 burn ===========${NOC}"
#    # 查询 ETH 这端 bridgeBank 地址原来是 0
#    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "0"
#
#    # ETH 这端 lock 7个 YCC
#    result=$(${CLIA} ethereum lock -m 7 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumBycERC20TokenAddr}")
#    cli_ret "${result}" "lock"
#
#    # eth 等待 2 个区块
#    sleep 4
#
#    # 查询 ETH 这端 bridgeBank 地址 7 YCC
#    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "7"
#
#    sleep ${maturityDegree}
#
#    # chain33 chain33EthBridgeTokenAddr（ETH合约中）查询 lock 金额
#    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
#    # 结果是 7 * le8
#    is_equal "${result}" "700000000"
#
#    # 原来的数额 0
#    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBycERC20TokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "0"
#
#    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
#    result=$(${CLIA} chain33 burn -m 5 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33BycBridgeTokenAddr}")
#    cli_ret "${result}" "burn"
#
#    sleep ${maturityDegree}
#
#    echo "check the balance on chain33"
#    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
#    # 结果是 7-5 * le8
#    is_equal "${result}" "200000000"
#
#    # 查询 ETH 这端 bridgeBank 地址 2
#    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "2"
#
#    # 更新后的金额 5
#    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBycERC20TokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "5"
#
#    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
#    result=$(${CLIA} chain33 burn -m 2 -k "${chain33ReceiverAddrKey}" -r "${ethReceiverAddr1}" -t "${chain33BycBridgeTokenAddr}")
#    cli_ret "${result}" "burn"
#
#    sleep ${maturityDegree}
#
#    echo "check the balance on chain33"
#    result=$(${Chain33Cli} evm query -a "${chain33BycBridgeTokenAddr}" -c "${chain33TestAddr1}" -b "balanceOf(${chain33ReceiverAddr})")
#    # 结果是 7-5 * le8
#    is_equal "${result}" "0"
#
#    # 查询 ETH 这端 bridgeBank 地址 0
#    result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumBycERC20TokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "0"
#
#    # 更新后的金额 5
#    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumBycERC20TokenAddr}")
#    cli_ret "${result}" "balance" ".balance" "7"
#
#    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
#}

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
#function offline_set_offline_token_Chain33Ycc() {
#    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
#    echo -e "${GRE}===== chain33 端 configLockedTokenOfflineSave ERC20 YCC ======${NOC}"
#    #    echo '2:#配置自动转离线钱包(YCC, 100, 60%)'
#    local threshold=10000000000
#    local percents=60
#    if [[ $# -eq 2 ]]; then
#        threshold=$1
#        percents=$2
#    fi
#    # shellcheck disable=SC2086
#    ${Boss4xCLI} chain33 offline set_offline_token -c "${chain33BridgeBank}" -t "${chain33YccERC20TokenAddr}" -s YCC -m ${threshold} -p ${percents} -k "${chain33DeployKey}" --chainID "${chain33ID}"
#    chain33_offline_send "chain33_set_offline_token.txt"
#
#    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
#}

# shellcheck disable=SC2120
function offline_set_offline_token_Eth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # echo '2:#配置自动转离线钱包(eth, 20, 50%)'
    local threshold=20
    local percents=50
    if [[ $# -eq 2 ]]; then
        threshold=$1
        percents=$2
    fi
    # shellcheck disable=SC2086
    ${Boss4xCLI} ethereum offline set_offline_token -s ETH -m ${threshold} -p ${percents} -c "${ethBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_token.txt"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2120
#function offline_set_offline_token_EthByc() {
#    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
#    local threshold=100
#    local percents=40
#    if [[ $# -eq 2 ]]; then
#        threshold=$1
#        percents=$2
#    fi
#    # shellcheck disable=SC2086
#    ${Boss4xCLI} ethereum offline set_offline_token -s BYC -m ${threshold} -p ${percents} -t "${ethereumBycERC20TokenAddr}" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
#    ethereum_offline_sign_send "set_offline_token.txt"
#
#    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
#}

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

#function offline_transfer_multisign_Bty_test() {
#    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
#    # transfer test
#    # shellcheck disable=SC2154
#    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 50 -r "${chain33BridgeBank}" -m "${multisignChain33Addr}"
#    # shellcheck disable=SC2154
#    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
#    chain33_offline_send "multisign_transfer.txt"
#    sleep 10
#    result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
#    is_equal "${result}" "103.2500"
#    result=$(${Chain33Cli} asset balance -a "${multisignChain33Addr}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
#    is_equal "${result}" "59.7500"
#
#    # shellcheck disable=SC2154
#    ${Boss4xCLI} chain33 offline create_multisign_transfer -a 10 -r "${chain33MultisignA}" -m "${multisignChain33Addr}"
#    # shellcheck disable=SC2154
#    ${Boss4xCLI} chain33 offline multisign_transfer -k "${chain33DeployKey}" -s "${chain33MultisignKeyA},${chain33MultisignKeyB},${chain33MultisignKeyC},${chain33MultisignKeyD}" --chainID "${chain33ID}"
#    chain33_offline_send "multisign_transfer.txt"
#    sleep 10
#    result=$(${Chain33Cli} asset balance -a "${chain33MultisignA}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
#    is_equal "${result}" "10.0000"
#    result=$(${Chain33Cli} asset balance -a "${multisignChain33Addr}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
#    is_equal "${result}" "49.7500"
#    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
#}

function coins_cross_transfer() {
    local key="${1}"
    local addr="${2}"
    local amount="${3}"
    local para_amount="${4}"
    local evm_amount="${5}"
    # 先把 bty 转入到 paracross 合约中
    hash=$(${MainCli} send coins send_exec -e paracross -a "${amount}" -k "${key}")
    check_tx "${MainCli}" "${hash}"

    # 主链中的 bty 夸链到 平行链中
    hash=$(${Para8801Cli} send para cross_transfer -a "${para_amount}" -e coins -s bty -t "${addr}" -k "${key}")
    check_tx "${Para8801Cli}" "${hash}"
    check_tx "${Para8901Cli}" "${hash}"
    result=$(${Para8901Cli} asset balance -a "${addr}" --asset_exec paracross --asset_symbol coins.bty | jq -r .balance)
    is_equal "${result}" "${para_amount}.0000"

    # 把平行链中的 bty 转入 平行链中的 evm 合约
    hash=$(${Para8901Cli} send para transfer_exec -a "${evm_amount}" -e user.p.para.evm -s coins.bty -k "${key}")
    check_tx "${Para8901Cli}" "${hash}"
    result=$(${Para8901Cli} asset balance -a "${addr}" --asset_exec paracross --asset_symbol coins.bty -e user.p.para.evm | jq -r .balance)
    is_equal "${result}" "${evm_amount}.0000"
}

function initPara() {
    # para add
    hash=$(${Para8901Cli} send coins transfer -a 10000 -n test -t "${chain33ReceiverAddr}" -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    check_tx "${Para8901Cli}" "${hash}"

    Chain33Cli=${Para8901Cli}
    InitChain33Validator

    coins_cross_transfer "${chain33DeployKey}" "${chain33DeployAddr}" 1000 800 500
    coins_cross_transfer "${chain33TestAddrKey1}" "${chain33TestAddr1}" 1000 800 500
    coins_cross_transfer "${chain33TestAddrKey2}" "${chain33TestAddr2}" 1000 800 500

    # 平行链共识节点增加测试币
    ${MainCli} send coins transfer -a 1000 -n test -t "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4" -k "${chain33ReceiverAddrKey}"
    ${MainCli} send coins transfer -a 1000 -n test -t "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR" -k "${chain33ReceiverAddrKey}"
    ${MainCli} send coins transfer -a 1000 -n test -t "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k" -k "${chain33ReceiverAddrKey}"
    ${MainCli} send coins transfer -a 1000 -n test -t "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs" -k "${chain33ReceiverAddrKey}"
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

#function lockChain33Ycc() {
#    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
#    echo -e "${GRE}===== chain33 端 lock ERC20 YCC ======${NOC}"
#    #    echo '2:#配置自动转离线钱包(YCC, 100, 60%)'
#    offline_set_offline_token_Chain33Ycc
#
#    lock_chain33_ycc_multisign 30 30 0
#    lock_chain33_ycc_multisign 70 40 60
#    lock_chain33_ycc_multisign 260 120 240
#    lock_chain33_ycc_multisign 10 52 318
#
#    # transfer test
#    offline_transfer_multisign_Chain33Ycc_test
#    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
#}

function lockEth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock ETH ======${NOC}"
    # echo '2:#配置自动转离线钱包(eth, 20, 50%)'
    offline_set_offline_token_Eth

    # 重启 nonce 会不统一 要重启一下
    restart_ebrelayerA

    lock_eth_multisign 19 19 0
    lock_eth_multisign 1 10 10
    lock_eth_multisign 16 13 23

    # transfer
    # shellcheck disable=SC2154
    offline_transfer_multisign_Eth_test
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

#function lockEthByc() {
#    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
#    echo -e "${GRE}===== ethereum 端 lock ERC20 Byc ======${NOC}"
#    # echo '2:#配置自动转离线钱包(ycc, 100, 40%)'
#    offline_set_offline_token_EthByc
#    # 重启 nonce 会不统一 要重启一下
#    restart_ebrelayerA
#
#    lock_ethereum_byc_multisign 70 70 0
#    lock_ethereum_byc_multisign 30 60 40
#    lock_ethereum_byc_multisign 60 72 88
#
#    # multisignEthAddr 要有手续费
#    ${CLIA} ethereum transfer -k "${ethDeployKey}" -m 10 -r "${multisignEthAddr}"
#    sleep 10
#
#    # transfer
#    offline_transfer_multisign_EthByc
#    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
#}

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
    ${CLIA} ethereum transfer -k "${ethDeployKey}" -m 10 -r "${multisignEthAddr}"
    sleep 10

    # transfer
    offline_transfer_multisign_EthUSDT
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function up_relayer_toml() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # 修改 relayer.toml 配置文件 pushName 字段
    pushNameChange "./relayer.toml"
    # 修改 relayer.toml 配置文件 initPowers
    validators_config

    # change EthProvider url
    dockerAddr=$(get_docker_addr "${dockerNamePrefix}_ganachetest_1")

    # 修改 relayer.toml 配置文件
    updata_relayer_a_toml "${dockerAddr}" "${dockerNamePrefix}_ebrelayera_1" "./relayer.toml"

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

    # shellcheck disable=SC2155
    local line=$(delete_line_show "./relayer.toml" "maturityDegree=10")
    sed -i ''"${line}"' a maturityDegree=1' "./relayer.toml"

    # shellcheck disable=SC2155
    local line=$(delete_line_show "./relayer.toml" "EthMaturityDegree=10")
    sed -i ''"${line}"' a EthMaturityDegree=1' "./relayer.toml"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function StartDockerRelayerDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # 修改 relayer.toml
    up_relayer_toml

    # 启动 ebrelayer
    start_docker_ebrelayerA

    # 部署合约 设置 bridgeRegistry 地址
    InitAndOfflineDeploy

    # 设置 ethereum symbol
#    ${Boss4xCLI} ethereum offline set_symbol -s "ETH" -c "${ethBridgeBank}" -d "${ethDeployAddr}"
#    ethereum_offline_sign_send "set_symbol.txt"

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
    #    InitTokenAddr
    offline_create_bridge_token_eth_BTY
    offline_create_bridge_token_chain33_ETH "ETH"
#    offline_deploy_erc20_eth_BYC
#    offline_create_bridge_token_chain33_BYC
#    offline_deploy_erc20_chain33_YCC
#    offline_create_bridge_token_eth_YCC
#    offline_deploy_erc20_chain33_ZBC
#    offline_create_bridge_token_eth_ZBC
    #    offline_deploy_erc20_eth_USDT
    offline_deploy_erc20_create_tether_usdt_USDT
    offline_create_bridge_token_chain33_USDT

    # shellcheck disable=SC2086
    {
        docker cp "${chain33EthBridgeTokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33EthBridgeTokenAddr}.abi
#        docker cp "${chain33BycBridgeTokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33BycBridgeTokenAddr}.abi
        docker cp "${chain33USDTBridgeTokenAddr}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33USDTBridgeTokenAddr}.abi
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
    echo -e "${GRE}ethereumUSDTERC20TokenAddr: ${ethereumUSDTERC20TokenAddr} ${NOC}"
    echo -e "${GRE}chain33USDTBridgeTokenAddr: ${chain33USDTBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}multisignChain33Addr: ${multisignChain33Addr} ${NOC}"
    echo -e "${GRE}multisignEthAddr: ${multisignEthAddr} ${NOC}"
    # shellcheck disable=SC2154
    echo -e "${GRE}XgoBridgeRegistryOnChain33: ${XgoBridgeRegistryOnChain33} ${NOC}"
    # shellcheck disable=SC2154
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

        docker_ganachetest_ip=$(get_docker_addr "${dockerNamePrefix}_ganachetest_1")
        Boss4xCLI="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/boss4x --rpc_laddr http://${docker_chain33_ip}:8901 --rpc_laddr_ethereum http://${docker_ganachetest_ip}:8545 --paraName user.p.para."
        echo "${Boss4xCLI}"
    }
}

function test_all() {
    # test
    Chain33Cli=${Para8901Cli}
    TestChain33ToEthAssets
    TestETH2Chain33Assets
    TestETH2Chain33USDT

    Chain33Cli=${Para8901Cli}
    lockBty
    lockEth
    lockEthUSDT

    # 离线多签地址转入阈值设大
    offline_set_offline_token_Bty 100000000000000 10
    offline_set_offline_token_Eth 100000000000000 10
    offline_set_offline_token_EthUSDT 100000000000000 10
}
