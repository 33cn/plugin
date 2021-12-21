#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./mainPubilcRelayerTest.sh"

function start_docker_ebrelayerProxy() {
    # shellcheck disable=SC2154
    cp './relayer.toml' "./relayerproxy.toml"

    # 删除配置文件中不需要的字段
    for deleteName in "deploy4chain33" "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers" "deploy" "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers"; do
        delete_line "./relayerproxy.toml" "${deleteName}"
    done

    pushNameChange "./relayerproxy.toml"

    sed -i 's/^ProcessWithDraw=.*/ProcessWithDraw=true/g' "./relayerproxy.toml"

    # shellcheck disable=SC2154
    docker cp "./relayerproxy.toml" "${dockerNamePrefix}_ebrelayerproxy_1":/root/relayer.toml
    start_docker_ebrelayer "${dockerNamePrefix}_ebrelayerproxy_1" "/root/ebrelayer" "./ebrelayerproxy.log"
    sleep 1

    # shellcheck disable=SC2154
    init_validator_relayer "${CLIP}" "${validatorPwd}" "${chain33ValidatorKeyp}" "${ethValidatorAddrKeyp}"
}

#1.通过以下命令设置代理提币
#./ebcli_A ethereum cfgWithdraw
#./ebcli_A ethereum cfgWithdraw 设置的数量是以wei为单位的 USDT 精度为６ eth/bnb　或者ｈｔ就是１８
#
#２．在chain33上的bridgeBank合约中设置proxyReceiver
#
#3.将其中代理转账中继器中的标志位ProcessWithDraw设置为true，其余中继器保持false

function setWithdraw() {
    start_docker_ebrelayerProxy
}

function AllRelayerMainTest() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    set +e

    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    # shellcheck disable=SC2120
    if [[ $# -ge 2 ]]; then
        # shellcheck disable=SC2034
        chain33ID="${2}"
    fi

    get_cli

    # init
    # shellcheck disable=SC2154
    # shellcheck disable=SC2034
    Chain33Cli=${MainCli}
    InitChain33Validator
    # para add
    initPara

    StartDockerRelayerDeploy

    test_all

    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
