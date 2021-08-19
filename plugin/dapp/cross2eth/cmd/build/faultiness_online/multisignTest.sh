#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 只测试 lock 操作，不测试 burn

source "./publicTest.sh"
source "./relayerPublic.sh"
source "./multisignPublicTest.sh"

## ETH 部署合约者的私钥 用于部署合约时签名使用
#ethDeployAddr="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
#ethDeployKey="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
#
## chain33 部署合约者的私钥 用于部署合约时签名使用
#chain33DeployAddr="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"
#
#Chain33Cli="../../chain33-cli"
#chain33BridgeBank=""
#ethBridgeBank=""
#chain33BtyTokenAddr="1111111111111111111114oLvT2"
#ethereumYccTokenAddr=""
#multisignChain33Addr=""
#multisignEthAddr=""
#ethBridgeToeknYccAddr=""
#chain33YccErc20Addr=""
#
#CLIA="./ebcli_A"
chain33ID=33

function mainTest() {
    if [[ $# -ge 1 && "${1}" != "" ]]; then
        # shellcheck disable=SC2034
        chain33ID="${1}"
    fi
    StartChain33
    start_trufflesuite
    AllRelayerStart

    deployMultisign

    lockBty
    lockChain33Ycc
    lockEth
    lockEthYcc
}

mainTest "${1}"
