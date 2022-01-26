#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck disable=SC2034
# shellcheck disable=SC2154
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./mainPubilcRelayerTest.sh"
source "./proxyVerifyTest.sh"

function AllRelayerMainTest() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    set +e

    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    if [[ $# -ge 2 ]]; then
        chain33ID="${2}"
    fi

    get_cli

    # init
    Chain33Cli=${MainCli}
    InitChain33Validator
    # para add
    initPara

    StartDockerRelayerDeploy
    test_all
    TestRelayerProxy

    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
