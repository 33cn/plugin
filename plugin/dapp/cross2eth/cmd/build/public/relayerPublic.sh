#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

source "./publicTest.sh"

# ETH 部署合约者的私钥 用于部署合约时签名使用
ethDeployAddr="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
ethDeployKey="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

# validatorsAddr=["0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]
#ethValidatorAddrKeyA="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
# shellcheck disable=SC2034
{
    ethValidatorAddrA="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
    ethValidatorAddrB="0x0df9a824699bc5878232c9e612fe1a5346a5a368"
    ethValidatorAddrC="0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1"
    ethValidatorAddrD="0xd9dab021e74ecf475788ed7b61356056b2095830"
    ethValidatorAddrKeyA="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
    ethValidatorAddrKeyB="a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
    ethValidatorAddrKeyC="bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
    ethValidatorAddrKeyD="c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"
}

# chain33 部署合约者的私钥 用于部署合约时签名使用
#chain33DeployAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
#chain33DeployKey="0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"
chain33DeployAddr="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"
chain33DeployKey="0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae"

chain33ReceiverAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
chain33ReceiverAddrKey="4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"

# 新增地址 chain33 需要导入地址 转入 10 bty当收费费
# shellcheck disable=SC2034
chain33ValidatorA="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"
chain33ValidatorB="155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6"
chain33ValidatorC="13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv"
chain33ValidatorD="113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG"
# shellcheck disable=SC2034
{
    chain33ValidatorKeyA="0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae"
    chain33ValidatorKeyB="0x9d539bc5fd084eb7fe86ad631dba9aa086dba38418725c38d9751459f567da66"
    chain33ValidatorKeyC="0x0a6671f101e30a2cc2d79d77436b62cdf2664ed33eb631a9c9e3f3dd348a23be"
    chain33ValidatorKeyD="0x3818b257b05ee75b6e43ee0e3cfc2d8502342cf67caed533e3756966690b62a5"
}

maturityDegree=10

Chain33Cli="../../chain33-cli"
BridgeRegistryOnChain33=""
chain33BridgeBank=""
BridgeRegistryOnEth=""
ethBridgeBank=""
chain33BtyTokenAddr="1111111111111111111114oLvT2"

#
chain33EthTokenAddr=""
ethereumBtyTokenAddr=""

# etheruem erc20 ycc
ethereumYccTokenAddr=""
chain33YccTokenAddr=""

# chain33 erc20 ycc
chain33YccErc20Addr=""
ethBridgeToeknYccAddr=""

CLIA="./ebcli_A"
# shellcheck disable=SC2034
CLIB="./ebcli_B"
CLIC="./ebcli_C"
CLID="./ebcli_D"
chain33ID=0

function kill_ebrelayerC() {
    kill_ebrelayer ./relayer_C/ebrelayer
    sleep 1
}
function kill_ebrelayerD() {
    kill_ebrelayer ./relayer_D/ebrelayer
    sleep 1
}

function start_ebrelayerC() {
    nohup ./relayer_C/ebrelayer ./relayer_C/relayer.toml >./relayer_C/cross2eth_C.log 2>&1 &
    sleep 2
    ${CLIC} unlock -p 123456hzj
    ${Chain33Cli} send coins transfer -a 1 -n note -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    ${Chain33Cli} send coins transfer -a 1 -n note -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    sleep ${maturityDegree}
    eth_block_wait 12
}

function start_ebrelayerD() {
    nohup ./relayer_D/ebrelayer ./relayer_D/relayer.toml >./relayer_D/cross2eth_D.log 2>&1 &
    sleep 2
    ${CLID} unlock -p 123456hzj
    ${Chain33Cli} send coins transfer -a 1 -n note -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    ${Chain33Cli} send coins transfer -a 1 -n note -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    sleep ${maturityDegree}
    eth_block_wait 12
}

function InitAndDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

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
    chain33BridgeBank=$(${Chain33Cli} evm query -c "${chain33DeployAddr}" -b "bridgeBank()" -a "${BridgeRegistryOnChain33}")
    cp Chain33BridgeBank.abi "${chain33BridgeBank}.abi"

    # 在 Eth 上部署合约
    result=$(${CLIA} ethereum deploy)
    cli_ret "${result}" "ethereum deploy"
    BridgeRegistryOnEth=$(echo "${result}" | jq -r ".msg")

    # 拷贝 BridgeRegistry.abi 和 BridgeBank.abi
    cp BridgeRegistry.abi "${BridgeRegistryOnEth}.abi"
    result=$(${CLIA} ethereum bridgeBankAddr)
    ethBridgeBank=$(echo "${result}" | jq -r ".addr")
    cp EthBridgeBank.abi "${ethBridgeBank}.abi"

    # 修改 relayer.toml 字段
    updata_relayer "BridgeRegistryOnChain33" "${BridgeRegistryOnChain33}" "./relayer.toml"

    line=$(delete_line_show "./relayer.toml" "BridgeRegistry=")
    if [ "${line}" ]; then
        sed -i ''"${line}"' a BridgeRegistry="'"${BridgeRegistryOnEth}"'"' "./relayer.toml"
    fi

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function create_bridge_token_eth_BTY() {
    # 在 Ethereum 上创建 bridgeToken BTY
    echo -e "${GRE}======= 在 Ethereum 上创建 bridgeToken BTY ======${NOC}"
    result=$(${CLIA} ethereum token create-bridge-token -s BTY)
    cli_ret "${result}" "ethereum token create-bridge-token -s BTY"
    # shellcheck disable=SC2034
    ethereumBtyTokenAddr=$(echo "${result}" | jq -r .addr)
}

function create_bridge_token_chain33_ETH() {
    # 在 chain33 上创建 bridgeToken ETH
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken ETH ======${NOC}"
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "createNewBridgeToken(ETH)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"
    chain33EthTokenAddr=$(${Chain33Cli} evm query -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(ETH)")
    echo "ETH Token Addr= ${chain33EthTokenAddr}"
    cp BridgeToken.abi "${chain33EthTokenAddr}.abi"

    result=$(${Chain33Cli} evm query -a "${chain33EthTokenAddr}" -c "${chain33EthTokenAddr}" -b "symbol()")
    is_equal "${result}" "ETH"
}

function deploy_erc20_eth_YCC() {
    # eth 上 铸币 YCC
    echo -e "${GRE}======= 在 ethereum 上创建 ERC20 ycc ======${NOC}"
    result=$(${CLIA} ethereum deploy_erc20 -c "${ethDeployAddr}" -n YCC -s YCC -m 33000000000000000000)
    cli_ret "${result}" "ethereum deploy_erc20 -s YCC"
    ethereumYccTokenAddr=$(echo "${result}" | jq -r .msg)

    result=$(${CLIA} ethereum token add_lock_list -s YCC -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "add_lock_list"
}

function create_bridge_token_chain33_YCC() {
    # 在chain33上创建bridgeToken YCC
    echo -e "${GRE}======= 在 chain33 上创建 bridgeToken YCC ======${NOC}"
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "createNewBridgeToken(YCC)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"
    chain33YccTokenAddr=$(${Chain33Cli} evm query -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(YCC)")
    echo "YCC Token Addr = ${chain33YccTokenAddr}"
    cp BridgeToken.abi "${chain33YccTokenAddr}.abi"

    result=$(${Chain33Cli} evm query -a "${chain33YccTokenAddr}" -c "${chain33YccTokenAddr}" -b "symbol()")
    is_equal "${result}" "YCC"
}

function deploy_erc20_chain33_YCC() {
    # chain33 token create YCC
    echo -e "${GRE}======= 在 chain33 上创建 ERC20 YCC ======${NOC}"
    result=$(${CLIA} chain33 token create -s YCC -o "${chain33DeployAddr}")
    cli_ret "${result}" "chain33 token create -s YCC"
    chain33YccErc20Addr=$(echo "${result}" | jq -r .msg)
    cp ERC20.abi "${chain33YccErc20Addr}.abi"

    # echo 'YCC.1:增加allowance的设置,或者使用relayer工具进行'
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33YccErc20Addr}" -p "approve(${chain33BridgeBank}, 330000000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    # echo 'YCC.2:#执行add lock操作:addToken2LockList'
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "addToken2LockList(${chain33YccErc20Addr}, YCC)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"
}

function create_bridge_token_eth_YCC() {
    # ethereum create-bridge-token YCC
    echo -e "${GRE}======= 在 ethereum 上创建 bridgeToken ycc ======${NOC}"
    result=$(${CLIA} ethereum token create-bridge-token -s YCC)
    cli_ret "${result}" "ethereum token create -s YCC"
    ethBridgeToeknYccAddr=$(echo "${result}" | jq -r .addr)
    cp BridgeToken.abi "${ethBridgeToeknYccAddr}.abi"
}

function deploy_erc20_chain33_ZBC() {
    # chain33 token create ZBC
    echo -e "${GRE}======= 在 chain33 上创建 ERC20 ZBC ======${NOC}"
    result=$(${CLIA} chain33 token create -s ZBC -o "${chain33DeployAddr}")
    cli_ret "${result}" "chain33 token create -s ZBC"
    chain33ZBCErc20Addr=$(echo "${result}" | jq -r .msg)
    cp ERC20.abi "${chain33ZBCErc20Addr}.abi"

    # echo 'ZBC.1:增加allowance的设置,或者使用relayer工具进行'
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33ZBCErc20Addr}" -p "approve(${chain33BridgeBank}, 330000000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    # echo 'ZBC.2:#执行add lock操作:addToken2LockList'
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "addToken2LockList(${chain33ZBCErc20Addr}, ZBC)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"
}

function create_bridge_token_eth_ZBC() {
    # ethereum create-bridge-token ZBC
    echo -e "${GRE}======= 在 ethereum 上创建 bridgeToken ZBC ======${NOC}"
    result=$(${CLIA} ethereum token create-bridge-token -s ZBC)
    cli_ret "${result}" "ethereum token create -s ZBC"
    ethBridgeToeknZBCAddr=$(echo "${result}" | jq -r .addr)
    cp BridgeToken.abi "${ethBridgeToeknZBCAddr}.abi"
}

# shellcheck disable=SC2120
function InitTokenAddr() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    create_bridge_token_eth_BTY
    create_bridge_token_chain33_ETH
    deploy_erc20_eth_YCC
    create_bridge_token_chain33_YCC
    deploy_erc20_chain33_YCC
    create_bridge_token_eth_YCC
    deploy_erc20_chain33_ZBC
    create_bridge_token_eth_ZBC
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function start_ebrelayerA() {
    nohup ./ebrelayer ./relayer.toml >cross2ethA.log 2>&1 &
    sleep 2
}

# start ebrelayer B C D
function updata_toml_start_BCD() {
    bind_port=9901
    push_port=20000
    for name in B C D; do
        local file="./relayer_$name/relayer.toml"
        cp './relayer.toml' "${file}"

        # 删除配置文件中不需要的字段
        for deleteName in "deploy4chain33" "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers" "deploy" "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers"; do
            delete_line "${file}" "${deleteName}"
        done

        bind_port=$((bind_port + 1))
        line=$(delete_line_show "./relayer_$name/relayer.toml" "JrpcBindAddr")
        if [ "${line}" ]; then
            sed -i ''"${line}"' a JrpcBindAddr="localhost:'${bind_port}'"' "./relayer_$name/relayer.toml"
        fi

        push_port=$((push_port + 1))
        line=$(delete_line_show "./relayer_$name/relayer.toml" "pushHost")
        if [ "${line}" ]; then
            sed -i ''"${line}"' a pushHost="http://localhost:'${push_port}'"' "./relayer_$name/relayer.toml"
        fi
        line=$(delete_line_show "./relayer_$name/relayer.toml" "pushBind")
        if [ "${line}" ]; then
            sed -i ''"${line}"' a pushBind="0.0.0.0:'${push_port}'"' "./relayer_$name/relayer.toml"
        fi

        sleep 1
        pushNameChange "./relayer_$name/relayer.toml"

        nohup ./relayer_$name/ebrelayer ./relayer_$name/relayer.toml >./relayer_$name/cross2eth_$name.log 2>&1 &
        sleep 2

        CLI="./ebcli_$name"
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

function validators_config() {
    # 修改 relayer.toml 配置文件 initPowers
    # shellcheck disable=SC2155
    line=$(delete_line_show "./relayer.toml" 'initPowers=\[96, 1, 1, 1\]')
    if [ "${line}" ]; then
        sed -i ''"${line}"' a initPowers=[25, 25, 25, 25]' "./relayer.toml"
    fi

    line=$(delete_line_show "./relayer.toml" 'initPowers=\[96, 1, 1, 1\]')
    if [ "${line}" ]; then
        sed -i ''"${line}"' a initPowers=[25, 25, 25, 25]' "./relayer.toml"
    fi

    line=$(delete_line_show "./relayer.toml" 'operatorAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"')
    if [ "${line}" ]; then
        sed -i ''"${line}"' a operatorAddr='\""${chain33DeployAddr}"\"'' "./relayer.toml"
    fi

    line=$(delete_line_show "./relayer.toml" 'deployerPrivateKey="0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"')
    if [ "${line}" ]; then
        sed -i ''"${line}"' a deployerPrivateKey='\""${chain33DeployKey}"\"'' "./relayer.toml"
    fi

    line=$(delete_line_show "./relayer.toml" 'validatorsAddr=\["14KEKbYtKKQm4wMthSK9J4La4nAiidGozt')
    if [ "${line}" ]; then
        sed -i ''"${line}"' a validatorsAddr=['\""${chain33ValidatorA}"\"', '\""${chain33ValidatorB}"\"', '\""${chain33ValidatorC}"\"', '\""${chain33ValidatorD}"\"']' "./relayer.toml"
    fi
}

function StartRelayerAndDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # 修改 relayer.toml 配置文件 pushName 字段
    pushNameChange "./relayer.toml"
    validators_config

    # 启动 ebrelayer
    start_ebrelayerA

    # 导入私钥 部署合约 设置 bridgeRegistry 地址
    InitAndDeploy

    # 重启
    kill_ebrelayer ebrelayer
    start_ebrelayerA

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    # start ebrelayer B C D
    updata_toml_start_BCD

    # 设置 token 地址
    InitTokenAddr

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# chian33 初始化准备
function InitChain33() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # init
    ${Chain33Cli} seed save -p 1314fuzamei -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib"
    ${Chain33Cli} wallet unlock -p 1314fuzamei -t 0
    ${Chain33Cli} account import_key -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944 -l returnAddr
    ${Chain33Cli} account import_key -k "${chain33ReceiverAddrKey}" -l minerAddr
    hash=$(${Chain33Cli} send coins transfer -a 10000 -n test -t "${chain33ReceiverAddr}" -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    check_tx "${Chain33Cli}" "${hash}"

    InitChain33Validator

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# chian33 初始化准备
function InitChain33Validator() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # 转帐到 DeployAddr
    result=$(${Chain33Cli} account import_key -k "${chain33DeployKey}" -l "DeployAddr")
    check_addr "${result}" "${chain33DeployAddr}"
    hash=$(${Chain33Cli} send coins transfer -a 6000 -n test -t "${chain33DeployAddr}" -k 4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01)
    check_tx "${Chain33Cli}" "${hash}"

    # 转账到 EVM  合约中
    hash=$(${Chain33Cli} send coins send_exec -e evm -a 3000 -k "${chain33DeployAddr}")
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} account balance -a "${chain33DeployAddr}" -e evm)
    #    balance_ret "${result}" "4000.0000"

    # 导入 chain33Validators 私钥生成地址
    for name in B C D; do
        eval chain33ValidatorKey=\$chain33ValidatorKey${name}
        eval chain33Validator=\$chain33Validator${name}
        result=$(${Chain33Cli} account import_key -k "${chain33ValidatorKey}" -l validator$name)
        # shellcheck disable=SC2154
        check_addr "${result}" "${chain33Validator}"

        # chain33Validator 要有手续费
        hash=$(${Chain33Cli} send coins transfer -a 100 -t "${chain33Validator}" -k "${chain33DeployAddr}")
        check_tx "${Chain33Cli}" "${hash}"
        #        result=$(${Chain33Cli} account balance -a "${chain33Validator}" -e coins)
        #        balance_ret "${result}" "100.0000"
    done

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function StartChain33() {
    kill_ebrelayer chain33
    sleep 2

    # delete chain33 datadir
    rm ../../datadir ../../logs -rf

    nohup ../../chain33 -f ./ci/cross2eth/test.toml >chain33log.log 2>&1 &

    sleep 1

    InitChain33
}

function AllRelayerStart() {
    kill_all_ebrelayer
    StartRelayerAndDeploy
}

function StartOneRelayer() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    kill_ebrelayer ebrelayer
    sleep 10
    rm datadir/ logs/ -rf

    # 修改 relayer.toml 配置文件 pushName 字段
    pushNameChange "./relayer.toml"

    # 启动 ebrelayer
    start_ebrelayerA

    # 导入私钥 部署合约 设置 bridgeRegistry 地址
    InitAndDeploy

    # 重启
    kill_ebrelayer ebrelayer
    start_ebrelayerA

    result=$(${CLIA} unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    # 设置 token 地址
    InitTokenAddr

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function StartRelayerOnRopsten() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# shellcheck disable=SC2034
{
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
    ethMultisignC=0xbc333839E37bc7fAAD0137aBaE2275030555101f
    ethMultisignD=0x495953A743ef169EC5D4aC7b5F786BF2Bd56aFd5
    ethMultisignKeyA=0x5e8aadb91eaa0fce4df0bcc8bd1af9e703a1d6db78e7a4ebffd6cf045e053574
    ethMultisignKeyB=0x0504bcb22b21874b85b15f1bfae19ad62fc2ad89caefc5344dc669c57efa60db
    ethMultisignKeyC=0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2
    ethMultisignKeyD=0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697
}

function initMultisignChain33Addr() {
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

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function deployChain33AndEthMultisign() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    echo -e "${GRE}=========== 部署 chain33 离线钱包合约 ===========${NOC}"
    result=$(${CLIA} chain33 multisign deploy)
    cli_ret "${result}" "chain33 multisign deploy"
    multisignChain33Addr=$(echo "${result}" | jq -r ".msg")

    echo -e "${GRE}=========== 部署 ETH 离线钱包合约 ===========${NOC}"
    result=$(${CLIA} ethereum multisign deploy)
    cli_ret "${result}" "ethereum multisign deploy"
    multisignEthAddr=$(echo "${result}" | jq -r ".msg")

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function setupChain33Multisign() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    echo -e "${GRE}=========== 设置 chain33 离线钱包合约 ===========${NOC}"
    result=$(${CLIA} chain33 multisign setup -k "${chain33DeployKey}" -o "${chain33MultisignA},${chain33MultisignB},${chain33MultisignC},${chain33MultisignD}")
    cli_ret "${result}" "chain33 multisign setup"

    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "configOfflineSaveAccount(${multisignChain33Addr})" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function setupEthMultisign() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    echo -e "${GRE}=========== 设置 ETH 离线钱包合约 ===========${NOC}"
    result=$(${CLIA} ethereum multisign setup -k "${ethDeployKey}" -o "${ethMultisignA},${ethMultisignB},${ethMultisignC},${ethMultisignD}")
    cli_ret "${result}" "ethereum multisign setup"

    result=$(${CLIA} ethereum multisign set_offline_addr -s "${multisignEthAddr}")
    cli_ret "${result}" "set_offline_addr"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function transferChain33MultisignFee() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # multisignChain33Addr 要有手续费
    hash=$(${Chain33Cli} send coins transfer -a 10 -t "${multisignChain33Addr}" -k "${chain33DeployAddr}")
    check_tx "${Chain33Cli}" "${hash}"
    result=$(${Chain33Cli} account balance -a "${multisignChain33Addr}" -e coins)
    balance_ret "${result}" "10.0000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function deployMultisign() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    initMultisignChain33Addr
    deployChain33AndEthMultisign
    setupChain33Multisign
    setupEthMultisign
    transferChain33MultisignFee
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# lock bty 判断是否转入多签地址金额是否正确
function lock_bty_multisign() {
    local lockAmount=$1
    local lockAmount2="${1}00000000"
    hash=$(${Chain33Cli} send evm call -f 1 -a "${lockAmount}" -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "lock(${ethDeployAddr}, ${chain33BtyTokenAddr}, ${lockAmount2})" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    if [[ $# -eq 3 ]]; then
        local bridgeBankBalance=$2
        local multisignBalance=$3

        result=$(${Chain33Cli} account balance -a "${chain33BridgeBank}" -e evm)
        balance_ret "${result}" "${bridgeBankBalance}"
        result=$(${Chain33Cli} account balance -a "${multisignChain33Addr}" -e evm)
        balance_ret "${result}" "${multisignBalance}"
    fi
}

# lock chain33 ycc erc20 判断是否转入多签地址金额是否正确
function lock_chain33_ycc_multisign() {
    local lockAmount="${1}00000000"
    hash=$(${Chain33Cli} send evm call -f 1 -k "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "lock(${ethDeployAddr}, ${chain33YccErc20Addr}, ${lockAmount})" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"

    if [[ $# -eq 3 ]]; then
        local bridgeBankBalance="${2}00000000"
        local multisignBalance="${3}00000000"
        if [[ ${3} == "0" ]]; then
            multisignBalance="0"
        fi

        result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${chain33BridgeBank}" -b "balanceOf(${chain33BridgeBank})")
        is_equal "${result}" "${bridgeBankBalance}"
        result=$(${Chain33Cli} evm query -a "${chain33YccErc20Addr}" -c "${multisignChain33Addr}" -b "balanceOf(${multisignChain33Addr})")
        is_equal "${result}" "${multisignBalance}"
    fi
}

# lock eth 判断是否转入多签地址金额是否正确
function lock_eth_multisign() {
    local lockAmount=$1
    result=$(${CLIA} ethereum lock -m "${lockAmount}" -k "${ethDeployKey}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"

    if [[ $# -eq 3 ]]; then
        local bridgeBankBalance=$2
        local multisignBalance=$3
        # eth 等待 2 个区块
        sleep 4
        #        eth_block_wait 2

        result=$(${CLIA} ethereum balance -o "${ethBridgeBank}")
        cli_ret "${result}" "balance" ".balance" "${bridgeBankBalance}"
        result=$(${CLIA} ethereum balance -o "${multisignEthAddr}")
        cli_ret "${result}" "balance" ".balance" "${multisignBalance}"
    fi
}

# lock ethereum ycc erc20 判断是否转入多签地址金额是否正确
function lock_ethereum_ycc_multisign() {
    local lockAmount=$1
    result=$(${CLIA} ethereum lock -m "${lockAmount}" -k "${ethDeployKey}" -r "${chain33ReceiverAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "lock"

    if [[ $# -eq 3 ]]; then
        local bridgeBankBalance=$2
        local multisignBalance=$3

        # eth 等待 2 个区块
        sleep 4
        #        eth_block_wait 2

        result=$(${CLIA} ethereum balance -o "${ethBridgeBank}" -t "${ethereumYccTokenAddr}")
        cli_ret "${result}" "balance" ".balance" "${bridgeBankBalance}"
        result=$(${CLIA} ethereum balance -o "${multisignEthAddr}" -t "${ethereumYccTokenAddr}")
        cli_ret "${result}" "balance" ".balance" "${multisignBalance}"
    fi
}

# 检查交易是否执行成功 $1:交易hash
function check_eth_tx() {
    local tx=${1}
    ty=$(${CLIA} ethereum receipt -s "${tx}" | jq .status | sed 's/\"//g')
    if [[ ${ty} != 0x1 ]]; then
        echo -e "${RED}check eth tx error, hash is ${tx}${NOC}"
        exit_cp_file
    fi
}
