#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck disable=SC2154
# shellcheck disable=SC2034
# shellcheck disable=SC2120
# shellcheck disable=SC2086
# shellcheck source=/dev/null
set -x
set +e

source "./offlinePublic.sh"

{
    chain33BridgeBank=""
    chain33BridgeRegistry=""
    chain33MultisignAddr=""
    chain33BtyERC20TokenAddr="1111111111111111111114oLvT2"

    chain33USDTBridgeTokenAddr=""
    chain33USDTBridgeTokenAddrOnETH=""
    chain33USDTBridgeTokenAddrOnBSC=""

    chain33MainBridgeTokenAddr=""
    chain33MainBridgeTokenAddrETH=""
    chain33MainBridgeTokenAddrBNB=""

    ethereumBridgeBank=""
    ethereumBridgeRegistry=""
    ethereumMultisignAddr=""
    ethereumUSDTERC20TokenAddr=""
    ethereumBtyBridgeTokenAddr=""

    ethereumBridgeBankOnETH=""
    ethereumBridgeRegistryOnETH=""
    ethereumMultisignAddrOnETH=""
    ethereumUSDTERC20TokenAddrOnETH=""
    ethereumBtyBridgeTokenAddrOnETH=""

    ethereumBridgeBankOnBSC=""
    ethereumBridgeRegistryOnBSC=""
    ethereumMultisignAddrOnBSC=""
    ethereumUSDTERC20TokenAddrOnBSC=""
    ethereumBtyBridgeTokenAddrOnBSC=""

    # ETH 部署合约者的私钥 用于部署合约时签名使用
    ethDeployAddr="0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a"
    ethDeployKey="0x8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

    # chain33 部署合约者的私钥 用于部署合约时签名使用
    chain33DeployAddr="1JxhYLYsrscjTaQfaMoVUrnSdrejP7XRQD"
    chain33DeployKey="0x9ef82623a5e9aac58d3a6b06392af66ec77289522b28896aed66abaaede66903"

    # validatorsAddr=["0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]
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

    chain33MultisignA="168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9"
    chain33MultisignB="13KTf57aCkVVJYNJBXBBveiA5V811SrLcT"
    chain33MultisignC="1JQwQWsShTHC4zxHzbUfYQK4kRBriUQdEe"
    chain33MultisignD="1NHuKqoKe3hyv52PF8XBAyaTmJWAqA2Jbb"
    chain33MultisignKeyA="0xcd284cd17456b73619fa609bb9e3105e8eff5d059c5e0b6eb1effbebd4d64144"
    chain33MultisignKeyB="0xe892212221b3b58211b90194365f4662764b6d5474ef2961ef77c909e31eeed3"
    chain33MultisignKeyC="0x9d19a2e9a440187010634f4f08ce36e2bc7b521581436a99f05568be94dc66ea"
    chain33MultisignKeyD="0x45d4ce009e25e6d5e00d8d3a50565944b2e3604aa473680a656b242d9acbff35"

    ethMultisignA=0x4c85848a7E2985B76f06a7Ed338FCB3aF94a7DCf
    ethMultisignB=0x6F163E6daf0090D897AD7016484f10e0cE844994
    ethMultisignC=0x0921948C0d25BBbe85285CB5975677503319F02A
    ethMultisignD=0x69921517970a28b73ac5E4C8ac8Fd135A80D2be1
    ethMultisignKeyA=0x5e8aadb91eaa0fce4df0bcc8bd1af9e703a1d6db78e7a4ebffd6cf045e053574
    ethMultisignKeyB=0x0504bcb22b21874b85b15f1bfae19ad62fc2ad89caefc5344dc669c57efa60db
    ethMultisignKeyC=0x5a43f2c8724f60ea5d6b87ad424daa73639a5fc76702edd3e5eaed37aaffdf49
    ethMultisignKeyD=0x03b28c0fc78c6ebae719b559b0781db24644b655d4bd58e5cf2311c9f03baa3d

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

    # 代理人地址
    ethValidatorAddrp="0x0c05ba5c230fdaa503b53702af1962e08d0c60bf"
    ethValidatorAddrKeyp="9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"
    chain33Validatorp="1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"
    chain33ValidatorKeyp="0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"

    # 代理收币地址
    chain33Validatorsp="1Hf1wnnr6XaYy5Sf3HhAfT4N8JYV4sMh9J"
    chain33ValidatorKeysp="0x1dadb7cbad8ea3f968cfad40ac32981def6215690618e62c48e816e7c732a8c2"

    chain33ID=0
    maturityDegree=20
    validatorPwd="123456fzm"
}

function start_docker_ebrelayerA() {
    docker cp "./relayer.toml" "${dockerNamePrefix}_ebrelayera_1":/root/relayer.toml
    start_docker_ebrelayer "${dockerNamePrefix}_ebrelayera_1" "/root/ebrelayer" "./ebrelayera.log"
    sleep 5
}

# start ebrelayer B C D
function updata_toml_start_bcd() {
    for name in b c d; do
        local file="./relayer$name.toml"
        cp './relayer.toml' "${file}"

        # 修改 relayer.toml 配置文件 pushName 字段
        sed -i 's/^pushName=.*/pushName="x2eth'"${name}"'"/g' "${file}"
        pushHost=$(get_docker_addr "${dockerNamePrefix}_ebrelayer${name}_1")
        sed -i 's/^pushHost=.*/pushHost="http:\/\/'"${pushHost}"':20000"/' "${file}"
        sed -i 's/^pushBind=.*/pushBind="'"${pushHost}"':20000"/' "${file}"
        if [[ ${name} == "d" ]]; then
            sed -i 's/^DelayedSendTime=.*/DelayedSendTime=180000/' "${file}"
        fi

        docker cp "${file}" "${dockerNamePrefix}_ebrelayer${name}_1":/root/relayer.toml
        start_docker_ebrelayer "${dockerNamePrefix}_ebrelayer${name}_1" "/root/ebrelayer" "./ebrelayer${name}.log"

        CLI="docker exec ${dockerNamePrefix}_ebrelayer${name}_1 /root/ebcli_A"
        eval chain33ValidatorKey=\$chain33ValidatorKey${name}
        eval ethValidatorAddrKey=\$ethValidatorAddrKey${name}

        init_validator_relayer "${CLI}" "${validatorPwd}" "${chain33ValidatorKey}" "${ethValidatorAddrKey}"
    done
}

function restart_ebrelayerA() {
    # 重启
    kill_docker_ebrelayer "${dockerNamePrefix}_ebrelayera_1"
    sleep 1
    start_docker_ebrelayerA

    result=$(${CLIA} unlock -p "${validatorPwd}")
    cli_ret "${result}" "unlock"
    sleep 20
}

function restart_ebrelayer_bcd() {
    # 重启
    local name=$1
    kill_docker_ebrelayer "${dockerNamePrefix}_ebrelayer${name}_1"
    sleep 1
    start_docker_ebrelayer "${dockerNamePrefix}_ebrelayer${name}_1" "/root/ebrelayer" "./ebrelayer${name}.log"
    sleep 5

    result=$(docker exec "${dockerNamePrefix}_ebrelayer${name}_1" "/root/ebcli_A" unlock -p "${validatorPwd}")
    cli_ret "${result}" "unlock"
    sleep 20
}

# chain33 lock BTY, eth burn BTY
function TestChain33ToEthAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== chain33 lock BTY, eth burn BTY ===========${NOC}"
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumBtyBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 原来的地址金额
    result=$(${Chain33Cli} asset balance -a "${chain33TestAddr1}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
    is_equal "${result}" "500.0000"

    # chain33 lock bty
    hash=$(${Chain33Cli} send evm call -f 1 -a 5 -k "${chain33TestAddr1}" -e "${chain33BridgeBank}" -p "lock(${ethTestAddr1}, ${chain33BtyERC20TokenAddr}, 500000000)")
    check_tx "${Chain33Cli}" "${hash}"

    # 原来的地址金额 减少了 5
    result=$(${Chain33Cli} asset balance -a "${chain33TestAddr1}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
    is_equal "${result}" "495.0000"

    # chain33BridgeBank 是否增加了 5
    result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
    is_equal "${result}" "5.0000"

    sleep 10

    # eth 这端 金额是否增加了 5
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumBtyBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "5"

    # eth burn
    result=$(${CLIA} ethereum burn -m 3 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumBtyBridgeTokenAddr}") #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    sleep 10

    # eth 这端 金额是否减少了 3
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumBtyBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "2"

    sleep ${maturityDegree}

    # 接收的地址金额 变成了 3
    result=$(${Chain33Cli} asset balance -a "${chain33ReceiverAddr}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
    is_equal "${result}" "3.0000"

    # chain33BridgeBank 是否减少了 3
    result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
    is_equal "${result}" "2.0000"

    # eth burn 2
    result=$(${CLIA} ethereum burn -m 2 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumBtyBridgeTokenAddr}") #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"

    sleep 10

    # eth 这端 金额是否减少了
    result=$(${CLIA} ethereum balance -o "${ethTestAddr1}" -t "${ethereumBtyBridgeTokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    sleep ${maturityDegree}

    # 接收的地址金额 变成了 5
    result=$(${Chain33Cli} asset balance -a "${chain33ReceiverAddr}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
    is_equal "${result}" "5.0000"

    # chain33BridgeBank 是否减少了 3
    result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
    is_equal "${result}" "0.0000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33Assets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn ===========${NOC}"
    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0"

    result=$(${CLIA} ethereum lock -m 0.002 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 10

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0.002"

    #    restart_ebrelayer_bcd "b"
    #    restart_ebrelayer_bcd "c"

    sleep ${maturityDegree}

    # chain33 chain33MainBridgeTokenAddr（ETH合约中）查询 lock 金额
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "200000"

    # 原来的数额
    result=$(${CLIA} ethereum balance -o "${ethTestAddr2}")

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 0.0003 -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33MainBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    #    restart_ebrelayer_bcd "b"
    #    restart_ebrelayer_bcd "c"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "170000"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0.0017"

    echo '#5.burn ETH from Chain33 ETH(Chain33)-----> Ethereum 6'
    result=$(${CLIA} chain33 burn -m 0.0017 -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33MainBridgeTokenAddr}")
    cli_ret "${result}" "burn"

    sleep ${maturityDegree}

    echo "check the balance on chain33"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
    is_equal "${result}" "0"

    # 查询 ETH 这端 bridgeBank 地址 0
    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
    cli_ret "${result}" "balance" ".balance" "0"

    result=$(${CLIA} ethereum balance -o "${ethTestAddr2}")

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33USDT() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}=========== eth to chain33 在以太坊上锁定 USDT 资产,然后在 chain33 上 burn ===========${NOC}"
    # 查询 ETH 这端 bridgeBank 地址原来是 0
    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # ETH 这端 lock 12个 USDT
    result=$(${CLIA} ethereum lock -m 12 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "lock"

    # eth 等待 2 个区块
    sleep 10

    # 查询 ETH 这端 bridgeBank 地址 12 USDT
    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "12"

    sleep ${maturityDegree}

    # chain33 chain33MainBridgeTokenAddr（ETH合约中）查询 lock 金额
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
    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
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
    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # 更新后的金额 12
    result=$(${CLIA} ethereum balance -o "${ethReceiverAddr1}" -t "${ethereumUSDTERC20TokenAddr}")
    cli_ret "${result}" "balance" ".balance" "12"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

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

    ${Boss4xCLI} chain33 offline set_offline_token -c "${chain33BridgeBank}" -s BTY -m ${threshold} -p ${percents} -k "${chain33DeployKey}"
    chain33_offline_send "chain33_set_offline_token.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_set_offline_token_Eth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # echo '2:#配置自动转离线钱包(eth, 20, 50%)'
    local threshold=20
    local percents=50
    local symbol="ETH"
    if [[ $# -eq 3 ]]; then
        threshold=$1
        percents=$2
        symbol=$3
    fi

    ${Boss4xCLI} ethereum offline set_offline_token -s ${symbol} -m ${threshold} -p ${percents} -c "${ethereumBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_token.txt"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function offline_set_offline_token_EthUSDT() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    local threshold=100
    local percents=40
    local symbol="USDT"
    if [[ $# -eq 3 ]]; then
        threshold=$1
        percents=$2
        symbol=$3
    fi

    ${Boss4xCLI} ethereum offline set_offline_token -s ${symbol} -m ${threshold} -p ${percents} -t "${ethereumUSDTERC20TokenAddr}" -c "${ethereumBridgeBank}" -d "${ethDeployAddr}"
    ethereum_offline_sign_send "set_offline_token.txt"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

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
    hash=$(${Para8901Cli} send para transfer_exec -a "${evm_amount}" -e "${paraName}evm" -s coins.bty -k "${key}")
    check_tx "${Para8901Cli}" "${hash}"
    result=$(${Para8901Cli} asset balance -a "${addr}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
    is_equal "${result}" "${evm_amount}.0000"
}

# lock bty 判断是否转入多签地址金额是否正确
function lock_bty_multisign_docker() {
    local lockAmount=$1
    local lockAmount2="${1}00000000"
    hash=$(${Chain33Cli} send evm call -f 1 -a "${lockAmount}" -k "${chain33TestAddr1}" -e "${chain33BridgeBank}" -p "lock(${ethTestAddr1}, ${chain33BtyERC20TokenAddr}, ${lockAmount2})")
    check_tx "${Chain33Cli}" "${hash}"

    if [[ $# -eq 3 ]]; then
        local bridgeBankBalance=$2
        local multisignBalance=$3

        result=$(${Chain33Cli} asset balance -a "${chain33BridgeBank}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
        is_equal "${result}" "${bridgeBankBalance}"
        result=$(${Chain33Cli} asset balance -a "${chain33MultisignAddr}" --asset_exec paracross --asset_symbol coins.bty -e "${paraName}evm" | jq -r .balance)
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

function lockEth() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock ETH ======${NOC}"
    # echo '2:#配置自动转离线钱包(eth, 20, 50%)'
    local symbol="${1}"
    offline_set_offline_token_Eth 20 50 "${symbol}"

    # 重启 nonce 会不统一 要重启一下
    restart_ebrelayerA

    lock_eth_multisign 19 19 0
    lock_eth_multisign 1 10 10
    lock_eth_multisign 16 13 23

    # transfer

    offline_transfer_multisign_Eth_test
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function lockEthUSDT() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo -e "${GRE}===== ethereum 端 lock ERC20 USDT ======${NOC}"
    # echo '2:#配置自动转离线钱包(ycc, 100, 40%)'
    local symbol="${1}"
    offline_set_offline_token_EthUSDT 100 40 "${symbol}"
    # 重启 nonce 会不统一 要重启一下
    restart_ebrelayerA

    lock_ethereum_usdt_multisign 70 70 0
    lock_ethereum_usdt_multisign 30 60 40
    lock_ethereum_usdt_multisign 60 72 88

    # ethereumMultisignAddr 要有手续费
    ${CLIA} ethereum transfer -k "${ethDeployKey}" -m 10 -r "${ethereumMultisignAddr}"
    sleep 10

    # transfer
    offline_transfer_multisign_EthUSDT
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function up_relayer_toml() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    relaye_file="./relayer.toml"

    # 修改 relayer.toml 配置文件 pushName 字段
    sed -i 's/^pushName=.*/pushName="x2ethA"/g' "${relaye_file}"

    # 替换7到15行
    sed -i '12,18s/ethProvider=.*/ethProvider=["ws:\/\/'"${docker_ganachetesteth_ip}"':8545\/","ws:\/\/'"${docker_ganachetesteth_ip}"':8545\/"]/g' "${relaye_file}"
    sed -i '20,26s/ethProvider=.*/ethProvider=["ws:\/\/'"${docker_ganachetestbsc_ip}"':8545\/","ws:\/\/'"${docker_ganachetestbsc_ip}"':8545\/"]/g' "${relaye_file}"
    sed -i '12,18s/EthProviderCli=.*/EthProviderCli=["http:\/\/'"${docker_ganachetesteth_ip}"':8545\/", "http:\/\/'"${docker_ganachetesteth_ip}"':8545\/"]/g' "${relaye_file}"
    sed -i '20,26s/EthProviderCli=.*/EthProviderCli=["http:\/\/'"${docker_ganachetestbsc_ip}"':8545\/", "http:\/\/'"${docker_ganachetestbsc_ip}"':8545\/"]/g' "${relaye_file}"
    sed -i 's/^pushHost=.*/pushHost="http:\/\/'"${docker_ebrelayera_ip}"':20000"/' "${relaye_file}"
    sed -i 's/^pushBind=.*/pushBind="'"${docker_ebrelayera_ip}"':20000"/' "${relaye_file}"
    sed -i 's/^chain33Host=.*/chain33Host="http:\/\/'"${docker_chain33_ip}"':8901"/' "${relaye_file}"
    sed -i 's/^chain33RpcUrls=.*/chain33RpcUrls=["http:\/\/'"${docker_chain33_ip}"':8901","http:\/\/'"${docker_chain30_ip}"':8901","http:\/\/'"${docker_chain31_ip}"':8901","http:\/\/'"${docker_chain32_ip}"':8901"]/' "${relaye_file}"

    sed -i 's/^EthBlockFetchPeriod=.*/EthBlockFetchPeriod=500/g' "${relaye_file}"
    sed -i 's/^fetchHeightPeriodMs=.*/fetchHeightPeriodMs=500/g' "${relaye_file}"

    sed -i 's/^ChainName=.*/ChainName="'"${paraName}"'"/g' "${relaye_file}"
    sed -i 's/^maturityDegree=.*/maturityDegree=1/g' "${relaye_file}"
    sed -i 's/^EthMaturityDegree=.*/EthMaturityDegree=1/g' "${relaye_file}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function StartDockerRelayerDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # 修改 relayer.toml
    up_relayer_toml
    # 启动 ebrelayer
    start_docker_ebrelayerA

    docker cp "./deploy_chain33.toml" "${dockerNamePrefix}_ebrelayera_1":/root/deploy_chain33.toml
    docker cp "./deploy_ethereum.toml" "${dockerNamePrefix}_ebrelayera_1":/root/deploy_ethereum.toml

    # 部署合约 设置 bridgeRegistry 地址
    OfflineDeploy

    # 向离线多签地址打点手续费
    Chain33Cli=${MainCli}
    initMultisignChain33Addr
    transferChain33MultisignFee
    Chain33Cli=${Para8901Cli}

    # 重启
    kill_docker_ebrelayer "${dockerNamePrefix}_ebrelayera_1"
    sleep 1
    start_docker_ebrelayerA
    InitRelayerA

    # 设置 token 地址
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

        Boss4xCLI=${Boss4xCLIbsc}
        CLIA=${CLIAbsc}
        ethereumBridgeBank="${ethereumBridgeBankOnBSC}"
        offline_create_bridge_token_chain33_symbol "BUSDT"
        chain33USDTBridgeTokenAddrOnBSC="${chain33MainBridgeTokenAddr}"
        offline_create_bridge_token_chain33_symbol "BNB"
        chain33MainBridgeTokenAddrBNB="${chain33MainBridgeTokenAddr}"
        offline_create_bridge_token_eth_BTY
        ethereumBtyBridgeTokenAddrOnBSC="${ethereumBtyBridgeTokenAddr}"
        offline_deploy_erc20_create_tether_usdt_USDT "BUSDT"
        ethereumUSDTERC20TokenAddrOnBSC="${ethereumUSDTERC20TokenAddr}"

        docker cp "${chain33BridgeBank}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33BridgeBank}.abi
        docker cp "${chain33BridgeRegistry}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33BridgeRegistry}.abi
        docker cp "${chain33USDTBridgeTokenAddrOnETH}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33USDTBridgeTokenAddrOnETH}.abi
        docker cp "${chain33USDTBridgeTokenAddrOnBSC}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33USDTBridgeTokenAddrOnBSC}.abi
        docker cp "${chain33MainBridgeTokenAddrETH}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33MainBridgeTokenAddrETH}.abi
        docker cp "${chain33MainBridgeTokenAddrBNB}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${chain33MainBridgeTokenAddrBNB}.abi
        docker cp "${ethereumBridgeBankOnETH}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${ethereumBridgeBankOnETH}.abi
        docker cp "${ethereumBridgeRegistryOnETH}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${ethereumBridgeRegistryOnETH}.abi
        docker cp "${ethereumBridgeBankOnBSC}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${ethereumBridgeBankOnBSC}.abi
        docker cp "${ethereumBridgeRegistryOnBSC}.abi" "${dockerNamePrefix}_ebrelayera_1":/root/${ethereumBridgeRegistryOnBSC}.abi
    }

    # start ebrelayer B C D
    updata_toml_start_bcd
    restart_ebrelayerA

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function echo_addrs() {
    echo -e "${GRE}=========== echo contract addrs ===========${NOC}"
    echo -e "${GRE}chain33BridgeBank: ${chain33BridgeBank} ${NOC}"
    echo -e "${GRE}chain33BridgeRegistry: ${chain33BridgeRegistry} ${NOC}"
    echo -e "${GRE}chain33MultisignAddr: ${chain33MultisignAddr} ${NOC}"
    echo -e "${GRE}chain33USDTBridgeTokenAddr: ${chain33USDTBridgeTokenAddr} ${NOC}"
    echo -e "${GRE}chain33MainBridgeTokenAddrETH: ${chain33MainBridgeTokenAddrETH} ${NOC}"
    echo -e "${GRE}chain33MainBridgeTokenAddrBNB: ${chain33MainBridgeTokenAddrBNB} ${NOC}"
    echo -e "${GRE}ethereumBridgeBankOnETH: ${ethereumBridgeBankOnETH} ${NOC}"
    echo -e "${GRE}ethereumBridgeRegistryOnETH: ${ethereumBridgeRegistryOnETH} ${NOC}"
    echo -e "${GRE}ethereumMultisignAddrOnETH: ${ethereumMultisignAddrOnETH} ${NOC}"
    echo -e "${GRE}ethereumUSDTERC20TokenAddrOnETH: ${ethereumUSDTERC20TokenAddrOnETH} ${NOC}"
    echo -e "${GRE}ethereumBtyBridgeTokenAddrOnETH: ${ethereumBtyBridgeTokenAddrOnETH} ${NOC}"
    echo -e "${GRE}ethereumBridgeBankOnBSC: ${ethereumBridgeBankOnBSC} ${NOC}"
    echo -e "${GRE}ethereumBridgeRegistryOnBSC: ${ethereumBridgeRegistryOnBSC} ${NOC}"
    echo -e "${GRE}ethereumMultisignAddrOnBSC: ${ethereumMultisignAddrOnBSC} ${NOC}"
    echo -e "${GRE}ethereumUSDTERC20TokenAddrOnBSC: ${ethereumUSDTERC20TokenAddrOnBSC} ${NOC}"
    echo -e "${GRE}ethereumBtyBridgeTokenAddrOnBSC: ${ethereumBtyBridgeTokenAddrOnBSC} ${NOC}"

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

function get_cli() {
    paraName="user.p.para."
    docker_chain33_ip=$(get_docker_addr "${dockerNamePrefix}_chain33_1")
    MainCli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8801"
    Para8801Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8901 --paraName ${paraName}"
    Para8901Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8901 --paraName ${paraName}"
    docker_chain31_ip=$(get_docker_addr "${dockerNamePrefix}_chain31_1")
    docker_chain32_ip=$(get_docker_addr "${dockerNamePrefix}_chain32_1")
    docker_chain30_ip=$(get_docker_addr "${dockerNamePrefix}_chain30_1")

    docker_ebrelayera_ip=$(get_docker_addr "${dockerNamePrefix}_ebrelayera_1")
    CLIP="docker exec ${dockerNamePrefix}_ebrelayerproxy_1 /root/ebcli_A"
    CLIA="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/ebcli_A"
    CLIB="docker exec ${dockerNamePrefix}_ebrelayerb_1 /root/ebcli_A"
    CLIC="docker exec ${dockerNamePrefix}_ebrelayerc_1 /root/ebcli_A"
    CLID="docker exec ${dockerNamePrefix}_ebrelayerd_1 /root/ebcli_A"

    docker_ganachetesteth_ip=$(get_docker_addr "${dockerNamePrefix}_ganachetesteth_1")
    docker_ganachetestbsc_ip=$(get_docker_addr "${dockerNamePrefix}_ganachetestbsc_1")
    Boss4xCLI="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/boss4x --rpc_laddr http://${docker_chain33_ip}:8901 --rpc_laddr_ethereum http://${docker_ganachetesteth_ip}:8545 --paraName ${paraName} --chainID ${chain33ID} --chainEthId 1337"

    Boss4xCLIeth="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/boss4x --rpc_laddr http://${docker_chain33_ip}:8901 --rpc_laddr_ethereum http://${docker_ganachetesteth_ip}:8545 --paraName ${paraName} --chainID ${chain33ID} --chainEthId 1337"
    Boss4xCLIbsc="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/boss4x --rpc_laddr http://${docker_chain33_ip}:8901 --rpc_laddr_ethereum http://${docker_ganachetestbsc_ip}:8545 --paraName ${paraName} --chainID ${chain33ID} --chainEthId 1337"

    CLIAeth="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/ebcli_A --node_addr http://${docker_ganachetesteth_ip}:8545 --eth_chain_name Ethereum"
    CLIAbsc="docker exec ${dockerNamePrefix}_ebrelayera_1 /root/ebcli_A --node_addr http://${docker_ganachetestbsc_ip}:8545 --eth_chain_name Binance"

    CLIPeth="docker exec ${dockerNamePrefix}_ebrelayerproxy_1 /root/ebcli_A --node_addr http://${docker_ganachetesteth_ip}:8545 --eth_chain_name Ethereum"
    CLIPbsc="docker exec ${dockerNamePrefix}_ebrelayerproxy_1 /root/ebcli_A --node_addr http://${docker_ganachetestbsc_ip}:8545 --eth_chain_name Binance"
}

function test_lock_and_burn() {
    local symbol1="${1}"
    local symbol2="${2}"
    # test
    Chain33Cli=${Para8901Cli}
    TestETH2Chain33Assets
    TestETH2Chain33USDT

    lockEth "${symbol1}"
    lockEthUSDT "${symbol2}"

    # 离线多签地址转入阈值设大
    offline_set_offline_token_Eth 100000000000000 10 "${symbol1}"
    offline_set_offline_token_EthUSDT 100000000000000 10 "${symbol2}"

    #    TestChain33ToEthAssets
    #    lockBty
    #    offline_set_offline_token_Bty 100000000000000 10
}

function test_all() {
    Boss4xCLI=${Boss4xCLIeth}
    CLIA=${CLIAeth}
    ethereumBridgeBank="${ethereumBridgeBankOnETH}"
    ethereumMultisignAddr="${ethereumMultisignAddrOnETH}"
    chain33MainBridgeTokenAddr="${chain33MainBridgeTokenAddrETH}"
    ethereumBtyBridgeTokenAddr="${ethereumBtyBridgeTokenAddrOnETH}"
    ethereumUSDTERC20TokenAddr="${ethereumUSDTERC20TokenAddrOnETH}"
    chain33USDTBridgeTokenAddr="${chain33USDTBridgeTokenAddrOnETH}"
    test_lock_and_burn "ETH" "USDT"

    Boss4xCLI=${Boss4xCLIbsc}
    CLIA=${CLIAbsc}
    ethereumBridgeBank="${ethereumBridgeBankOnBSC}"
    ethereumMultisignAddr="${ethereumMultisignAddrOnBSC}"
    chain33MainBridgeTokenAddr="${chain33MainBridgeTokenAddrBNB}"
    ethereumBtyBridgeTokenAddr="${ethereumBtyBridgeTokenAddrOnBSC}"
    ethereumUSDTERC20TokenAddr="${ethereumUSDTERC20TokenAddrOnBSC}"
    chain33USDTBridgeTokenAddr="${chain33USDTBridgeTokenAddrOnBSC}"
    test_lock_and_burn "BNB" "BUSDT"
}
