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
chain33EthBridgeTokenAddr=""
ethereumBtyBridgeTokenAddr=""
chain33BycBridgeTokenAddr=""
ethereumBycERC20TokenAddr=""
BridgeRegistryOnChain33=""
chain33YccERC20TokenAddr=""
BridgeRegistryOnEth=""
ethereumYccBridgeTokenAddr=""
chain33ZbcERC20TokenAddr=""
ethereumZbcBridgeTokenAddr=""
multisignChain33Addr=""
multisignEthAddr=""
chain33ID=0

# shellcheck disable=SC2034
{
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
function offline_set_offline_token_EthYcc() {
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

function MainTest() {
    set +e
    chain33ID=0
    chain33BridgeBank=15Myyvq97WinTWto8zcEdm838zXmvJKfnX
    ethBridgeBank=0xC65B02a22B714b55D708518E2426a22ffB79113d
#    ethereumBtyBridgeTokenAddr=0x9c3d40a44a2f61ef8d46fa8c7a731c08fb16ccef
#    chain33EthBridgeTokenAddr=1JVFbJhFUWUNH41PxbV7NqwUd3F9BJ3nqV

    ethereumBycERC20TokenAddr=0x20a32A5680EBf55740B0C98B54cDE8e6FD5a4FB0
#    ethereumYccBridgeTokenAddr=0x05f3f31c7d53bcb71a6487dff3115d86370698bd
#    chain33BycBridgeTokenAddr=1BdREGqsjbcKkvRheXWYKRq37vJHMs22Uy
    chain33YccERC20TokenAddr=17yu1yULdGFddUz26PEeaHpJtkFGEpzYrA
#    chain33ZbcERC20TokenAddr=1AqRwUa4T3q9DuCyUwn5ucHgtUhbUP2yfu
#    ethereumZbcBridgeTokenAddr=0x89bb32184e466a9c8ea50c31174b575c2bcd64c2

    dockerNamePrefix="build"
    docker_chain33_ip=$(get_docker_addr "${dockerNamePrefix}_chain33_1")
#    MainCli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8801"
#    Para8801Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para."
    Para8901Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8901 --paraName user.p.para."

    # shellcheck disable=SC2034
    {
        CLIA="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/ebcli_A"
        CLIB="docker exec ${dockerNamePrefix}_ebrelayerb_1 /root/ebcli_A"
        CLIC="docker exec ${dockerNamePrefix}_ebrelayerc_1 /root/ebcli_A"
        CLID="docker exec ${dockerNamePrefix}_ebrelayerd_1 /root/ebcli_A"

        docker_ganachetest_ip=$(get_docker_addr "${dockerNamePrefix}_ganachetest_1")
        Boss4xCLI="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/boss4x --rpc_laddr http://${docker_chain33_ip}:8901 --rpc_laddr_ethereum http://${docker_ganachetest_ip}:8545 --paraName user.p.para."

        echo "${Boss4xCLI}"
    }

    Chain33Cli=${Para8901Cli}

    # 离线多签地址转入阈值设大
    offline_set_offline_token_Bty 100000000000000 10
    offline_set_offline_token_Chain33Ycc 100000000000000 10
    offline_set_offline_token_Eth 100000000000000 10
    offline_set_offline_token_EthYcc 100000000000000 10

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

MainTest
