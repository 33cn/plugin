#!/usr/bin/env bash
#shellcheck disable=SC2128
#shellcheck source=/dev/null

source ../dapp-test-common.sh
source "../x2ethereum/publicTest.sh"

sendAddress="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
sendPriKey="0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
MAIN_HTTP=""
chain33SenderAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
# validatorsAddr=["0x92c8b16afd6d423652559c6e266cbe1c29bfd84f", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]
ethValidatorAddrKeyA="3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
ethValidatorAddrKeyB="a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
ethValidatorAddrKeyC="bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
ethValidatorAddrKeyD="c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"
# 新增地址 chain33 需要导入地址 转入 10 bty当收费费
chain33Validator1="1H4zzzQEQQR2FxXwppiMRXcvqLvqzxK2nv"
chain33Validator2="1Nq5AhTgVNvYaWQqih8ZQQEaRk3CFhTDHp"
chain33Validator3="16nmxjF58z5oKK9m44cGy241zMSJWPN1Ty"
chain33Validator4="182nAEMxF1JWWxEWdu4jvd68aZhQumS97H"
chain33ValidatorKey1="0x260124d9c619b0088241ffe2f1d7dc56b0b6100c88c342040387cd62b8ba35a3"
chain33ValidatorKey2="0x7812f8c688048943f1c168f8f2f76f44912de1f0ff8b12358b213118081869b2"
chain33ValidatorKey3="0xd44c8f3d8cac5d9c7fef7b0a0bf7be0909372ec6368064f742193de0bddeb2d1"
chain33ValidatorKey4="0xaad36689ca332026d4a4ceee62c8a91bac7bc100906b25a181a7f28b8552b53e"
ethReceiverAddr1="0xa4ea64a583f6e51c3799335b28a8f0529570a635"
ethReceiverAddrKey1="355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"
ethReceiverAddr2="0x0c05ba5c230fdaa503b53702af1962e08d0c60bf"
#ethReceiverAddrKey2="9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"
maturityDegree=1
CLIA_HTTP="http://127.0.0.1:9901"
CLIB_HTTP="http://127.0.0.1:9902"
CLIC_HTTP="http://127.0.0.1:9903"
CLID_HTTP="http://127.0.0.1:9904"

# 更新 B C D 的配置文件
function updata_all_relayer_toml_rpc() {
    local port=9901
    local port2=20000

    for name in B C D; do
        local file="./x2ethereum/$name/relayer.toml"
        cp './x2ethereum/A/relayer.toml' "${file}"
        cp './x2ethereum/A/ebrelayer' "./x2ethereum/$name/ebrelayer"

        # 删除配置文件中不需要的字段
        for deleteName in "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers" "deployerPrivateKey" "deploy"; do
            delete_line "${file}" "${deleteName}"
        done

        # 替换端口
        port=$((port + 1))
        sed -i 's/localhost:9901/localhost:'${port}'/g' "${file}"

        port2=$((port2 + 1))
        sed -i 's/20000/'${port2}'/g' "${file}"

        sed -i 's/x2ethereum/x2ethereum'${name}'/g' "${file}"
    done
}

# 更新配置文件 $1 为 BridgeRegistry 合约地址; $2 等待区块 默认10; $3 relayer.toml 地址
function updata_relayer_toml_rpc() {
    local BridgeRegistry=${1}
    local maturityDegree=${2}
    local file=${3}

    # shellcheck disable=SC2155
    local pushHost=$(ifconfig wlp2s0 | grep "inet " | awk '{ print $2}' | awk -F: '{print $2}')
    if [[ ${pushHost} == "" ]]; then
        pushHost=$(ifconfig wlp2s0 | grep "inet " | awk '{ print $2}')
        if [[ ${pushHost} == "" ]]; then
            pushHost=$(ifconfig eth0 | grep "inet " | awk '{ print $2}' | awk -F: '{print $2}')
            if [[ ${pushHost} == "" ]]; then
                pushHost=$(ifconfig eth0 | grep "inet " | awk '{ print $2}')
                if [[ ${pushHost} == "" ]]; then
                    ip addr show eth0
                    pushHost=$(ip addr show eth0 | grep "inet " | awk '{ print $2}' | head -c-4)
                fi
            fi
        fi
    fi

    if [[ ${pushHost} == "" ]]; then
        echo -e "${RED}pushHost is empty${NOC}"
        exit 1
    fi

    # shellcheck disable=SC2155
    local line=$(delete_line_show "${file}" "chain33Host")
    # 在第 line 行后面 新增合约地址
    sed -i ''"${line}"' a chain33Host="'"${MAIN_HTTP}"'"' "${file}"

    line=$(delete_line_show "${file}" "pushHost")
    sed -i ''"${line}"' a pushHost="http://'"${pushHost}"':20000"' "${file}"

    line=$(delete_line_show "${file}" "BridgeRegistry")
    sed -i ''"${line}"' a BridgeRegistry="'"${BridgeRegistry}"'"' "${file}"

    sed -i 's/EthMaturityDegree=10/'EthMaturityDegree="${maturityDegree}"'/g' "${file}"
    sed -i 's/maturityDegree=10/'maturityDegree="${maturityDegree}"'/g' "${file}"

    sed -i 's/^EthBlockFetchPeriod=.*/EthBlockFetchPeriod=500/g' "${file}"
    sed -i 's/^fetchHeightPeriodMs=.*/fetchHeightPeriodMs=500/g' "${file}"
}

# $1 sendAddress, $2 balance
function queryExecBalance() {
    local resp=""
    chain33_QueryExecBalance "${1}" "x2ethereum" "$MAIN_HTTP"
    # shellcheck disable=SC2155
    local balance=$(echo "$resp" | jq -r ".result" | jq ".[].balance")
    if [ "${balance}" != "${2}" ]; then
        echo_rst "queryExecBalance" "1" "${balance} != ${2}"
    fi
}

# $1 chain33Address, $2 balance
function queryChain33Balance() {
    local resp=""
    chain33_QueryBalance "${1}" "${MAIN_HTTP}"
    # shellcheck disable=SC2155
    local balance=$(echo $resp | jq -r ".result.execAccount" | jq ".[].account.balance")
    if [ "${balance}" != "${2}" ]; then
        echo_rst "queryChain33Balance" "1" "${balance} != ${2}"
    fi
}

# $1 req , $2 balance
function queryRelayerBalance() {
    chain33_Http "${1}" ${CLIA_HTTP} '(.error|not) and (.result != null)' "GetBalance" ".result.balance"
    if [ "${RETURN_RESP}" != "${2}" ]; then
        echo_rst "queryRelayerBalance" "1" "${RETURN_RESP} != ${2}"
    fi
}

# $1 req , $2 balance
function queryChain33X2ethBalance() {
    chain33_Http "${req}" ${MAIN_HTTP} '(.error|not) and (.result != null)' "GetBalance" ".result"
    # shellcheck disable=SC2155
    local balance=$(echo "${RETURN_RESP}" | jq -r ".res" | jq ".[].balance" | sed 's/\"//g')
    if [ "${balance}" != "${2}" ]; then
        echo_rst "queryChain33X2ethBalance" "1" "${balance} != ${2}"
    fi
}

function StartRelayerAndDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    for name in A B C D; do
        local ebrelayer="./x2ethereum/$name/ebrelayer"
        kill_ebrelayer "${ebrelayer}"
    done
    sleep 1

    rm -rf './x2ethereum/A' './x2ethereum/B' './x2ethereum/C' './x2ethereum/D'
    mkdir './x2ethereum/A' './x2ethereum/B' './x2ethereum/C' './x2ethereum/D'
    cp '../x2ethereum/relayer.toml' './x2ethereum/A/relayer.toml'
    cp '../x2ethereum/ebrelayer' './x2ethereum/A/ebrelayer'

    start_trufflesuite

    start_ebrelayer "./x2ethereum/A/ebrelayer" "./x2ethereum/A/ebrelayer.log"
    # 部署合约
    InitAndDeploy

    # 获取 BridgeRegistry 地址
    local req='{"method":"Manager.ShowBridgeRegistryAddr","params":[{}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result.addr"
    local BridgeRegistry="$RETURN_RESP"

    kill_ebrelayer "./x2ethereum/A/ebrelayer"
    # 修改 relayer.toml 配置文件
    updata_relayer_toml_rpc "${BridgeRegistry}" "${maturityDegree}" "./x2ethereum/A/relayer.toml"
    updata_all_relayer_toml_rpc

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function InitAndDeploy() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    local req='{"method":"Manager.SetPassphase","params":[{"Passphase":"123456hzj"}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "SetPassphase" ".result"

    local req='{"method":"Manager.Unlock","params":["123456hzj"]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "Unlock" ".result"

    local req='{"method":"Manager.DeployContrcts","params":[{}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# chian33 添加验证着及权重
function InitChain33Vilators() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # 导入 chain33Validators 私钥生成地址
    chain33_ImportPrivkey "${chain33ValidatorKey1}" "${chain33Validator1}" "tokenAddr" "${MAIN_HTTP}"
    chain33_ImportPrivkey "${chain33ValidatorKey2}" "${chain33Validator2}" "tokenAddr" "${MAIN_HTTP}"
    chain33_ImportPrivkey "${chain33ValidatorKey3}" "${chain33Validator3}" "tokenAddr" "${MAIN_HTTP}"
    chain33_ImportPrivkey "${chain33ValidatorKey4}" "${chain33Validator4}" "tokenAddr" "${MAIN_HTTP}"

    # SetConsensusThreshold
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"SetConsensusThreshold","payload":{"consensusThreshold":"80"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$sendPriKey" ${MAIN_HTTP} "SetConsensusThreshold"

    # add a validator
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"AddValidator","payload":{"address":"'${chain33Validator1}'","power":"25"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$sendPriKey" ${MAIN_HTTP} "AddValidator"
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"AddValidator","payload":{"address":"'${chain33Validator2}'","power":"25"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$sendPriKey" ${MAIN_HTTP} "AddValidator"
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"AddValidator","payload":{"address":"'${chain33Validator3}'","power":"25"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$sendPriKey" ${MAIN_HTTP} "AddValidator"
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"AddValidator","payload":{"address":"'${chain33Validator4}'","power":"25"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$sendPriKey" ${MAIN_HTTP} "AddValidator"

    # query Validators
    chain33_Http '{"method":"Chain33.Query","params":[{"execer":"x2ethereum","funcName":"GetTotalPower","payload":{}}]}' ${MAIN_HTTP} '(.error|not) and (.result != null)' "GetTotalPower" ".result.totalPower"
    if [ "${RETURN_RESP}" != "100" ]; then
        echo -e "${RED}=========== GetTotalPower err: TotalPower = $RETURN_RESP ===========${NOC}"
    fi

    # cions 转帐到 x2ethereum 合约地址
    x2eth_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"x2ethereum"}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SendToAddress "${sendAddress}" "${x2eth_addr}" 20000000000 "${MAIN_HTTP}"
    queryExecBalance "${sendAddress}" "20000000000"

    # chain33Validator 要有手续费
    chain33_applyCoins "${chain33Validator1}" 1000000000 "${MAIN_HTTP}"
    queryChain33Balance "${chain33Validator1}" "1000000000"
    chain33_applyCoins "${chain33Validator2}" 1000000000 "${MAIN_HTTP}"
    queryChain33Balance "${chain33Validator2}" "1000000000"
    chain33_applyCoins "${chain33Validator3}" 1000000000 "${MAIN_HTTP}"
    queryChain33Balance "${chain33Validator3}" "1000000000"
    chain33_applyCoins "${chain33Validator4}" 1000000000 "${MAIN_HTTP}"
    queryChain33Balance "${chain33Validator4}" "1000000000"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function EthImportKey() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    local port=9901
    # 重启 ebrelayer 并解锁
    for name in A B C D; do
        start_ebrelayer "./x2ethereum/$name/ebrelayer" "./x2ethereum/$name/ebrelayer.log"

        local CLI_HTTP="http://127.0.0.1:${port}"

        if [ "${name}" != "A" ]; then
            local req='{"method":"Manager.SetPassphase","params":[{"Passphase":"123456hzj"}]}'
            chain33_Http "$req" ${CLI_HTTP} '(.error|not) and (.result != null)' "SetPassphase" ".result"
        fi

        req='{"method":"Manager.Unlock","params":["123456hzj"]}'
        chain33_Http "$req" ${CLI_HTTP} '(.error|not) and (.result != null)' "Unlock" ".result"

        port=$((port + 1))
    done

    local req='{"method":"Manager.ImportChain33PrivateKey4EthRelayer","params":["'${chain33ValidatorKey1}'"]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "ImportChain33PrivateKey4EthRelayer" ".result"
    req='{"method":"Manager.ImportChain33PrivateKey4EthRelayer","params":["'${chain33ValidatorKey2}'"]}'
    chain33_Http "$req" ${CLIB_HTTP} '(.error|not) and (.result != null)' "ImportChain33PrivateKey4EthRelayer" ".result"
    req='{"method":"Manager.ImportChain33PrivateKey4EthRelayer","params":["'${chain33ValidatorKey3}'"]}'
    chain33_Http "$req" ${CLIC_HTTP} '(.error|not) and (.result != null)' "ImportChain33PrivateKey4EthRelayer" ".result"
    req='{"method":"Manager.ImportChain33PrivateKey4EthRelayer","params":["'${chain33ValidatorKey4}'"]}'
    chain33_Http "$req" ${CLID_HTTP} '(.error|not) and (.result != null)' "ImportChain33PrivateKey4EthRelayer" ".result"

    req='{"method":"Manager.ImportChain33RelayerPrivateKey","params":[{"privateKey":"'${ethValidatorAddrKeyA}'"}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "ImportChain33RelayerPrivateKey" ".result"
    req='{"method":"Manager.ImportChain33RelayerPrivateKey","params":[{"privateKey":"'${ethValidatorAddrKeyB}'"}]}'
    chain33_Http "$req" ${CLIB_HTTP} '(.error|not) and (.result != null)' "ImportChain33RelayerPrivateKey" ".result"
    req='{"method":"Manager.ImportChain33RelayerPrivateKey","params":[{"privateKey":"'${ethValidatorAddrKeyC}'"}]}'
    chain33_Http "$req" ${CLIC_HTTP} '(.error|not) and (.result != null)' "ImportChain33RelayerPrivateKey" ".result"
    req='{"method":"Manager.ImportChain33RelayerPrivateKey","params":[{"privateKey":"'${ethValidatorAddrKeyD}'"}]}'
    chain33_Http "$req" ${CLID_HTTP} '(.error|not) and (.result != null)' "ImportChain33RelayerPrivateKey" ".result"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestChain33ToEthAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # token4chain33 在 以太坊 上先有 bty
    local req='{"method":"Manager.CreateBridgeToken","params":["coins.bty"]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "CreateBridgeToken" ".result.addr"
    tokenAddrBty=${RETURN_RESP}
    #    tokenAddrBty="0x9C3D40A44a2F61Ef8D46fa8C7A731C08FB16cCEF"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${ethReceiverAddr1}'","tokenAddr":"'${tokenAddrBty}'"}]}'
    queryRelayerBalance "$req" "0"

    # chain33 lock bty
    #shellcheck disable=SC2086
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"Chain33ToEthLock","payload":{"TokenContract":"'${tokenAddrBty}'","Chain33Sender":"'${sendPriKey}'","EthereumReceiver":"'${ethReceiverAddr1}'","Amount":"500000000","IssuerDotSymbol":"coins.bty","Decimals":"8"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$sendPriKey" ${MAIN_HTTP} "Chain33ToEthLock"

    queryExecBalance "${sendAddress}" "19500000000"

    eth_block_wait $((maturityDegree + 2))

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${ethReceiverAddr1}'","tokenAddr":"'${tokenAddrBty}'"}]}'
    queryRelayerBalance "$req" "5"

    # eth burn
    req='{"method":"Manager.Burn","params":[{"ownerKey":"'${ethReceiverAddrKey1}'","tokenAddr":"'${tokenAddrBty}'","chain33Receiver":"'${chain33SenderAddr}'","amount":"500000000"}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "Burn" ".result"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${ethReceiverAddr1}'","tokenAddr":"'${tokenAddrBty}'"}]}'
    queryRelayerBalance "$req" "0"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    queryExecBalance "${chain33SenderAddr}" "500000000"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# eth to chain33
# 在以太坊上锁定资产,然后在 chain33 上铸币,针对 eth 资产
function TestETH2Chain33Assets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    local req='{"method":"Manager.ShowBridgeBankAddr","params":[{}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "ShowBridgeBankAddr" ".result.addr"
    bridgeBankAddr="${RETURN_RESP}"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${bridgeBankAddr}'","tokenAddr":""}]}'
    queryRelayerBalance "$req" "0"

    # eth lock 0.1
    req='{"method":"Manager.LockEthErc20Asset","params":[{"ownerKey":"'${ethReceiverAddrKey1}'","tokenAddr":"","amount":"100000000000000000","chain33Receiver":"'${sendAddress}'"}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "LockEthErc20Asset" ".result"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${bridgeBankAddr}'","tokenAddr":""}]}'
    queryRelayerBalance "$req" "0.1"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    req='{"method":"Chain33.Query","params":[{"execer":"x2ethereum","funcName":"GetRelayerBalance","payload":{"tokenSymbol":"eth","address":"'${sendAddress}'","tokenAddr":"0x0000000000000000000000000000000000000000"}}]}'
    queryChain33X2ethBalance "${req}" "0.1"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${ethReceiverAddr2}'","tokenAddr":""}]}'
    chain33_Http "${req}" ${CLIA_HTTP} '(.error|not) and (.result != null)' "GetBalance" ".result.balance"
    local balance=${RETURN_RESP}

    #    burn 0.1
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"Chain33ToEthBurn","payload":{"TokenContract":"0x0000000000000000000000000000000000000000","Chain33Sender":"'${sendPriKey}'","EthereumReceiver":"'${ethReceiverAddr2}'","Amount":"10000000","IssuerDotSymbol":"eth","Decimals":"18"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$sendPriKey" ${MAIN_HTTP} "Chain33ToEthBurn"

    req='{"method":"Chain33.Query","params":[{"execer":"x2ethereum","funcName":"GetRelayerBalance","payload":{"tokenSymbol":"eth","address":"'${sendAddress}'","tokenAddr":"0x0000000000000000000000000000000000000000"}}]}'
    queryChain33X2ethBalance "${req}" "0"

    eth_block_wait $((maturityDegree + 3))

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${bridgeBankAddr}'","tokenAddr":""}]}'
    queryRelayerBalance "$req" "0"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${ethReceiverAddr2}'","tokenAddr":""}]}'
    #queryRelayerBalance "$req" "$(echo "${balance}+0.1" | bc)"
    queryRelayerBalance "$req" "100.1"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function TestETH2Chain33Erc20() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # token4erc20 在 chain33 上先有 token,同时 mint
    local req='{"method":"Manager.CreateERC20Token","params":["testc"]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "CreateERC20Token" ".result.addr"
    tokenAddr="${RETURN_RESP}"

    # 先铸币 1000
    req='{"method":"Manager.MintErc20","params":[{"owner":"'${ethReceiverAddr1}'","tokenAddr":"'${tokenAddr}'","amount":"100000000000"}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "MintErc20" ".result.addr"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${ethReceiverAddr1}'","tokenAddr":"'${tokenAddr}'"}]}'
    queryRelayerBalance "$req" "1000"

    local req='{"method":"Manager.ShowBridgeBankAddr","params":[{}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "ShowBridgeBankAddr" ".result.addr"
    bridgeBankAddr="${RETURN_RESP}"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${bridgeBankAddr}'","tokenAddr":"'${tokenAddr}'"}]}'
    queryRelayerBalance "$req" "0"

    # lock 100
    req='{"method":"Manager.LockEthErc20Asset","params":[{"ownerKey":"'${ethReceiverAddrKey1}'","tokenAddr":"'${tokenAddr}'","amount":"10000000000","chain33Receiver":"'${chain33Validator1}'"}]}'
    chain33_Http "$req" ${CLIA_HTTP} '(.error|not) and (.result != null)' "LockEthErc20Asset" ".result"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${ethReceiverAddr1}'","tokenAddr":"'${tokenAddr}'"}]}'
    queryRelayerBalance "$req" "900"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${bridgeBankAddr}'","tokenAddr":"'${tokenAddr}'"}]}'
    queryRelayerBalance "$req" "100"

    # eth 等待 10 个区块
    eth_block_wait $((maturityDegree + 2))

    req='{"method":"Chain33.Query","params":[{"execer":"x2ethereum","funcName":"GetRelayerBalance","payload":{"tokenSymbol":"testc","address":"'${chain33Validator1}'","tokenAddr":"'${tokenAddr}'"}}]}'
    queryChain33X2ethBalance "${req}" "100"

    # chain33 burn 100
    #shellcheck disable=SC2086
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"Chain33ToEthBurn","payload":{"TokenContract":"'${tokenAddr}'","Chain33Sender":"'${chain33ValidatorKey1}'","EthereumReceiver":"'${ethReceiverAddr2}'","Amount":"10000000000","IssuerDotSymbol":"testc","Decimals":"8"}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$chain33ValidatorKey1" ${MAIN_HTTP} "Chain33ToEthBurn"

    req='{"method":"Chain33.Query","params":[{"execer":"x2ethereum","funcName":"GetRelayerBalance","payload":{"tokenSymbol":"testc","address":"'${chain33Validator1}'","tokenAddr":"'${tokenAddr}'"}}]}'
    queryChain33X2ethBalance "${req}" "0"

    eth_block_wait $((maturityDegree + 3))

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${ethReceiverAddr2}'","tokenAddr":"'${tokenAddr}'"}]}'
    queryRelayerBalance "$req" "100"

    req='{"method":"Manager.GetBalance","params":[{"owner":"'${bridgeBankAddr}'","tokenAddr":"'${tokenAddr}'"}]}'
    queryRelayerBalance "$req" "0"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function rpc_test() {
    set +e
    chain33_RpcTestBegin x2ethereum
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    if [ "$ispara" == false ]; then
        # init
        StartRelayerAndDeploy
        InitChain33Vilators
        EthImportKey

        # test
        TestChain33ToEthAssets
        TestETH2Chain33Assets
        TestETH2Chain33Erc20
    fi
    chain33_RpcTestRst x2ethereum "$CASE_ERR"
}

chain33_debug_function rpc_test "$1"
