#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x

# eth 和 chain33 两端都启动
# 启动4个 relayer 每个权重一样

source "./publicTest.sh"

CLIA="../build/ebcli_A"
CLIB="../build/ebcli_B"
CLIC="../build/ebcli_C"
CLID="../build/ebcli_D"

docker_chain33_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' build_chain33_1)
Chain33Cli="$GOPATH/src/github.com/33cn/plugin/build/chain33-cli --rpc_laddr http://${docker_chain33_ip}:8801"

chain33SenderAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
chain33SenderAddrKey="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"

# validatorsAddr=["0x92c8b16afd6d423652559c6e266cbe1c29bfd84f", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]
ethValidatorAddrKeyA="3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
ethValidatorAddrKeyB="a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
ethValidatorAddrKeyC="bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
ethValidatorAddrKeyD="c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"

# 新增地址 chain33 需要导入地址 转入 10 bty当收费费
chain33Validator1="1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"
chain33Validator2="155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6"
chain33Validator3="13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv"
chain33Validator4="113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG"
chain33ValidatorKey1="0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
chain33ValidatorKey2="0x9d539bc5fd084eb7fe86ad631dba9aa086dba38418725c38d9751459f567da66"
chain33ValidatorKey3="0x0a6671f101e30a2cc2d79d77436b62cdf2664ed33eb631a9c9e3f3dd348a23be"
chain33ValidatorKey4="0x3818b257b05ee75b6e43ee0e3cfc2d8502342cf67caed533e3756966690b62a5"

ethReceiverAddr1="0xa4ea64a583f6e51c3799335b28a8f0529570a635"
ethReceiverAddrKey1="355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"
ethReceiverAddr2="0x0c05ba5c230fdaa503b53702af1962e08d0c60bf"
ethReceiverAddrKey2="9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"
maturityDegree=10

function InitAndDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    result=$(${CLIA} relayer set_pwd -n 123456hzj -o kk)
    cli_ret "${result}" "set_pwd"

    result=$(${CLIA} relayer unlock -p 123456hzj)
    cli_ret "${result}" "unlock"

    result=$(${CLIA} relayer ethereum deploy)
    cli_ret "${result}" "deploy"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function EthImportKey() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # 重启 ebrelayer 并解锁
    for name in A B C D; do
        start_ebrelayer "../build/"$name"/ebrelayer" "../build/"$name"/ebrelayer.log"

        # 导入测试地址私钥
        CLI="../build/ebcli_$name"

        result=$(${CLI} relayer set_pwd -n 123456hzj -o kk)
        #cli_ret "${result}" "set_pwd"

        result=$(${CLI} relayer unlock -p 123456hzj)
        cli_ret "${result}" "unlock"
    done

    result=$(${CLIA} relayer ethereum import_chain33privatekey -k "${chain33ValidatorKey1}")
    cli_ret "${result}" "import_chain33privatekey"
    result=$(${CLIB} relayer ethereum import_chain33privatekey -k "${chain33ValidatorKey2}")
    cli_ret "${result}" "import_chain33privatekey"
    result=$(${CLIC} relayer ethereum import_chain33privatekey -k "${chain33ValidatorKey3}")
    cli_ret "${result}" "import_chain33privatekey"
    result=$(${CLID} relayer ethereum import_chain33privatekey -k "${chain33ValidatorKey4}")
    cli_ret "${result}" "import_chain33privatekey"

    result=$(${CLIA} relayer ethereum import_ethprivatekey -k "${ethValidatorAddrKeyA}")
    cli_ret "${result}" "import_ethprivatekey"
    result=$(${CLIB} relayer ethereum import_ethprivatekey -k "${ethValidatorAddrKeyB}")
    cli_ret "${result}" "import_ethprivatekeyB"
    result=$(${CLIC} relayer ethereum import_ethprivatekey -k "${ethValidatorAddrKeyC}")
    cli_ret "${result}" "import_ethprivatekeyC"
    result=$(${CLID} relayer ethereum import_ethprivatekey -k "${ethValidatorAddrKeyD}")
    cli_ret "${result}" "import_ethprivatekeyD"

    result=$(${CLIA} relayer chain33 import_privatekey -k "${ethValidatorAddrKeyA}")
    cli_ret "${result}" "A relayer chain33 import_privatekey"
    result=$(${CLIB} relayer chain33 import_privatekey -k "${ethValidatorAddrKeyB}")
    cli_ret "${result}" "B relayer chain33 import_privatekey"
    result=$(${CLIC} relayer chain33 import_privatekey -k "${ethValidatorAddrKeyC}")
    cli_ret "${result}" "C relayer chain33 import_privatekey"
    result=$(${CLID} relayer chain33 import_privatekey -k "${ethValidatorAddrKeyD}")
    cli_ret "${result}" "D relayer chain33 import_privatekey"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function StartRelayerAndDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    for name in A B C D; do
        local ebrelayer="../build/$name/ebrelayer"
        kill_ebrelayer "${ebrelayer}"
    done
    kill_ebrelayer "../build/ebrelayer"
    sleep 1

    rm -rf '../build/A' '../build/B' '../build/C' '../build/D' '../build/datadir' '../build/ebrelayer.log' '../build/logs'
    mkdir '../build/A' '../build/B' '../build/C' '../build/D'
    cp '../ebrelayer/relayer.toml' '../build/A/relayer.toml'
    cp '../build/ebrelayer' '../build/A/ebrelayer'

    start_trufflesuite

    start_ebrelayer "../build/A/ebrelayer" "../build/A/ebrelayer.log"
    # 部署合约
    InitAndDeploy

    # 获取 BridgeRegistry 地址
    result=$(${CLIA} relayer ethereum bridgeRegistry)
    BridgeRegistry=$(cli_ret "${result}" "bridgeRegistry" ".addr")
    #    BridgeRegistry="0x5331F912027057fBE8139D91B225246e8159232f"

    kill_ebrelayer "../build/A/ebrelayer"
    # 修改 relayer.toml 配置文件
    updata_relayer_toml ${BridgeRegistry} ${maturityDegree} "../build/A/relayer.toml"
    updata_all_relayer_toml

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# chian33 添加验证着及权重
function InitChain33Vilators() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # 导入 chain33Validators 私钥生成地址
    result=$(${Chain33Cli} account import_key -k ${chain33ValidatorKey1} -l validator1)
    check_addr "${result}" ${chain33Validator1}
    result=$(${Chain33Cli} account import_key -k ${chain33ValidatorKey2} -l validator2)
    check_addr "${result}" ${chain33Validator2}
    result=$(${Chain33Cli} account import_key -k ${chain33ValidatorKey3} -l validator3)
    check_addr "${result}" ${chain33Validator3}
    result=$(${Chain33Cli} account import_key -k ${chain33ValidatorKey4} -l validator4)
    check_addr "${result}" ${chain33Validator4}

    # SetConsensusThreshold
    hash=$(${Chain33Cli} send x2ethereum setconsensus -p 80 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"

    # add a validator
    hash=$(${Chain33Cli} send x2ethereum add -a ${chain33Validator1} -p 25 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"
    hash=$(${Chain33Cli} send x2ethereum add -a ${chain33Validator2} -p 25 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"
    hash=$(${Chain33Cli} send x2ethereum add -a ${chain33Validator3} -p 25 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"
    hash=$(${Chain33Cli} send x2ethereum add -a ${chain33Validator4} -p 25 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"

    # query Validators
    totalPower=$(${Chain33Cli} send x2ethereum query totalpower -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq .totalPower | sed 's/\"//g')
    check_number 100 ${totalPower}

    # cions 转帐到 x2ethereum 合约地址
    hash=$(${Chain33Cli} send coins send_exec -e x2ethereum -a 200 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)

    check_tx "${Chain33Cli}" "${hash}"
    result=$(${Chain33Cli} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum)
    balance_ret "${result}" "200.0000"

    # chain33Validator 要有手续费
    hash=$(${Chain33Cli} send coins transfer -a 10 -t "${chain33Validator1}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"
    result=$(${Chain33Cli} account balance -a "${chain33Validator1}" -e coins)
    balance_ret "${result}" "10.0000"

    hash=$(${Chain33Cli} send coins transfer -a 10 -t "${chain33Validator2}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"
    result=$(${Chain33Cli} account balance -a "${chain33Validator2}" -e coins)
    balance_ret "${result}" "10.0000"

    hash=$(${Chain33Cli} send coins transfer -a 10 -t "${chain33Validator3}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"
    result=$(${Chain33Cli} account balance -a "${chain33Validator3}" -e coins)
    balance_ret "${result}" "10.0000"

    hash=$(${Chain33Cli} send coins transfer -a 10 -t "${chain33Validator4}" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    check_tx "${Chain33Cli}" "${hash}"
    result=$(${Chain33Cli} account balance -a "${chain33Validator4}" -e coins)
    balance_ret "${result}" "10.0000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestChain33ToEthAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # token4chain33 在 以太坊 上先有 bty
    result=$(${CLIA} relayer ethereum token4chain33 -s coins.bty)
    tokenAddr=$(cli_ret "${result}" "token4chain33" ".addr")

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr1}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "0"

    # chain33 lock bty
    hash=$(${Chain33Cli} send x2ethereum lock -a 5 -t coins.bty -r ${ethReceiverAddr1} -q ${tokenAddr} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    block_wait "${Chain33Cli}" $((maturityDegree + 2))
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} account balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e x2ethereum)
    balance_ret "${result}" "195.0000"

    eth_block_wait $((maturityDegree + 2))

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr1}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "5"

    # eth burn
    result=$(${CLIA} relayer ethereum burn -m 5 -k "${ethReceiverAddrKey1}" -r "${chain33SenderAddr}" -t "${tokenAddr}")
    cli_ret "${result}" "burn"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr1}" -t "coins.bty")
    cli_ret "${result}" "balance" ".balance" "0"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    result=$(${Chain33Cli} account balance -a "${chain33SenderAddr}" -e x2ethereum)
    balance_ret "${result}" "5.0000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# eth to chain33
# 在以太坊上锁定资产,然后在 chain33 上铸币,针对 eth 资产
function TestETH2Chain33Assets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    ${CLIA} relayer unlock -p 123456hzj

    result=$(${CLIA} relayer ethereum bridgeBankAddr)
    bridgeBankAddr=$(cli_ret "${result}" "bridgeBankAddr" ".addr")

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    # eth lock 10
    result=$(${CLIA} relayer ethereum lock -m 10 -k "${ethReceiverAddrKey1}" -r 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    cli_ret "${result}" "lock"

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "10"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    result=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]")
    balance_ret "${result}" "10"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}")
    balance=$(cli_ret "${result}" "balance" ".balance")

    hash=$(${Chain33Cli} send x2ethereum burn -a 10 -t eth -r ${ethReceiverAddr2} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    block_wait "${Chain33Cli}" $((maturityDegree + 2))
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]")
    balance_ret "${result}" "0"

    eth_block_wait 2

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}")
    cli_ret "${result}" "balance" ".balance" $(echo "${balance}+10" | bc)

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33Erc20() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    ${CLIA} relayer unlock -p 123456hzj

    # token4erc20 在 chain33 上先有 token,同时 mint
    tokenSymbol="testc"
    result=$(${CLIA} relayer ethereum token4erc20 -s "${tokenSymbol}")
    tokenAddr=$(cli_ret "${result}" "token4erc20" ".addr")

    # 先铸币 1000
    result=$(${CLIA} relayer ethereum mint -m 1000 -o "${ethReceiverAddr1}" -t "${tokenAddr}")
    cli_ret "${result}" "mint"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "1000"

    result=$(${CLIA} relayer ethereum bridgeBankAddr)
    bridgeBankAddr=$(cli_ret "${result}" "bridgeBankAddr" ".addr")

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "0"

    # lock 100
    result=$(${CLIA} relayer ethereum lock -m 100 -k "${ethReceiverAddrKey1}" -r "${chain33Validator1}" -t "${tokenAddr}")
    cli_ret "${result}" "lock"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "900"

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "100"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    result=$(${Chain33Cli} x2ethereum balance -s "${chain33Validator1}" -t "${tokenSymbol}" -a "${tokenAddr}" | jq ".res" | jq ".[]")
    balance_ret "${result}" "100"

    # chain33 burn 100
    hash=$(${Chain33Cli} send x2ethereum burn -a 100 -t "${tokenSymbol}" -r ${ethReceiverAddr2} -q ${tokenAddr} -k "${chain33Validator1}")
    block_wait "${Chain33Cli}" $((maturityDegree + 2))
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} x2ethereum balance -s "${chain33Validator1}" -t "${tokenSymbol}" -a "${tokenAddr}" | jq ".res" | jq ".[]")
    balance_ret "${result}" "0"

    eth_block_wait 2

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "100"

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenSymbol}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestChain33ToEthAssetsKill() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    if [ "${tokenAddrBty}" == "" ]; then
        # token4chain33 在 以太坊 上先有 bty
        result=$(${CLIA} relayer ethereum token4chain33 -s bty)
        tokenAddrBty=$(cli_ret "${result}" "token4chain33" ".addr")
    fi

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenAddrBty}")
    cli_ret "${result}" "balance" ".balance" "0"

    kill_ebrelayerC
    kill_ebrelayerD

    # chain33 lock bty
    hash=$(${Chain33Cli} send x2ethereum lock -a 5 -t bty -r ${ethReceiverAddr2} -q ${tokenAddrBty} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    block_wait "${Chain33Cli}" $((maturityDegree + 2))
    check_tx "${Chain33Cli}" "${hash}"

    eth_block_wait $((maturityDegree + 2))

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}" -t "${tokenAddrBty}")
    cli_ret "${result}" "balance" ".balance" "0"

    start_ebrelayerC

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}" -t "${tokenAddrBty}")
    cli_ret "${result}" "balance" ".balance" "5"

    # eth burn
    result=$(${CLIA} relayer ethereum burn -m 5 -k "${ethReceiverAddrKey2}" -r "${chain33Validator1}" -t "${tokenAddrBty}")
    cli_ret "${result}" "burn"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}" -t "${tokenAddrBty}")
    cli_ret "${result}" "balance" ".balance" "0"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    result=$(${Chain33Cli} account balance -a "${chain33Validator1}" -e x2ethereum)
    balance_ret "${result}" "0"

    start_ebrelayerD

    result=$(${Chain33Cli} account balance -a "${chain33Validator1}" -e x2ethereum)
    balance_ret "${result}" "5"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# eth to chain33
# 在以太坊上锁定资产,然后在 chain33 上铸币,针对 eth 资产
function TestETH2Chain33AssetsKill() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    ${CLIA} relayer unlock -p 123456hzj

    result=$(${CLIA} relayer ethereum bridgeBankAddr)
    bridgeBankAddr=$(cli_ret "${result}" "bridgeBankAddr" ".addr")

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    kill_ebrelayerC
    kill_ebrelayerD

    # eth lock 0.1
    result=$(${CLIA} relayer ethereum lock -m 0.1 -k "${ethReceiverAddrKey1}" -r 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    cli_ret "${result}" "lock"

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "0.1"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    result=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]")
    balance_ret "${result}" "0"

    start_ebrelayerC
    start_ebrelayerD

    result=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]")
    balance_ret "${result}" "0.1"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}")
    balance=$(cli_ret "${result}" "balance" ".balance")

    kill_ebrelayerC
    kill_ebrelayerD

    hash=$(${Chain33Cli} send x2ethereum burn -a 0.1 -t eth -r ${ethReceiverAddr2} -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    block_wait "${Chain33Cli}" $((maturityDegree + 2))
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} x2ethereum balance -s 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -t eth | jq ".res" | jq ".[]")
    balance_ret "${result}" "0"

    eth_block_wait 2

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}")
    cli_ret "${result}" "balance" ".balance" "0.1"

    start_ebrelayerC
    start_ebrelayerD

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}")
    cli_ret "${result}" "balance" ".balance" $(echo "${balance}+0.1" | bc)

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33Erc20Kill() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    ${CLIA} relayer unlock -p 123456hzj

    # token4erc20 在 chain33 上先有 token,同时 mint
    tokenSymbol="testcc"
    result=$(${CLIA} relayer ethereum token4erc20 -s "${tokenSymbol}")
    tokenAddr=$(cli_ret "${result}" "token4erc20" ".addr")

    # 先铸币 1000
    result=$(${CLIA} relayer ethereum mint -m 1000 -o "${ethReceiverAddr1}" -t "${tokenAddr}")
    cli_ret "${result}" "mint"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenAddr}")
    cli_ret "${result}" "balance" ".balance" "1000"

    result=$(${CLIA} relayer ethereum bridgeBankAddr)
    bridgeBankAddr=$(cli_ret "${result}" "bridgeBankAddr" ".addr")

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    kill_ebrelayerC
    kill_ebrelayerD

    # lock 100
    result=$(${CLIA} relayer ethereum lock -m 100 -k "${ethReceiverAddrKey1}" -r "${chain33Validator1}" -t "${tokenAddr}")
    cli_ret "${result}" "lock"

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr1}" -t "${tokenAddr}")
    cli_ret "${result}" "balance" ".balance" "900"

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenAddr}")
    cli_ret "${result}" "balance" ".balance" "100"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    result=$(${Chain33Cli} x2ethereum balance -s "${chain33Validator1}" -t "${tokenSymbol}" -a "${tokenAddr}" | jq ".res" | jq ".[]")
    balance_ret "${result}" "0"

    start_ebrelayerC
    start_ebrelayerD

    result=$(${Chain33Cli} x2ethereum balance -s "${chain33Validator1}" -t "${tokenSymbol}" -a "${tokenAddr}" | jq ".res" | jq ".[]")
    balance_ret "${result}" "100"

    kill_ebrelayerC
    kill_ebrelayerD

    # chain33 burn 100
    hash=$(${Chain33Cli} send x2ethereum burn -a 100 -t "${tokenSymbol}" -r ${ethReceiverAddr2} -q ${tokenAddr} -k "${chain33Validator1}")
    block_wait "${Chain33Cli}" $((maturityDegree + 2))
    check_tx "${Chain33Cli}" "${hash}"

    result=$(${Chain33Cli} x2ethereum balance -s "${chain33Validator1}" -t "${tokenSymbol}" -a "${tokenAddr}" | jq ".res" | jq ".[]")
    balance_ret "${result}" "0"

    eth_block_wait 2

    start_ebrelayerC

    result=$(${CLIA} relayer ethereum balance -o "${ethReceiverAddr2}" -t "${tokenAddr}")
    cli_ret "${result}" "balance" ".balance" "100"

    result=$(${CLIA} relayer ethereum balance -o "${bridgeBankAddr}" -t "${tokenAddr}")
    cli_ret "${result}" "balance" ".balance" "0"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function main() {
    echo -e "${GRE}===========allTest $FUNCNAME begin ===========${NOC}"

    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    StartRelayerAndDeploy
    InitChain33Vilators
    EthImportKey

    TestChain33ToEthAssets
    TestETH2Chain33Assets
    TestETH2Chain33Erc20

    # kill relayer and start relayer
    #    TestChain33ToEthAssetsKill
    #    TestETH2Chain33AssetsKill
    #    TestETH2Chain33Erc20Kill

    echo -e "${GRE}===========allTest $FUNCNAME end ===========${NOC}"
}

main "${1}"
