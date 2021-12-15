#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./mainPubilcRelayerTest.sh"

function AllRelayerMainTest() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    set +e

    get_cli

    if [[ ${1} != "" ]]; then
        maturityDegree=${1}
        echo -e "${GRE}maturityDegree is ${maturityDegree} ${NOC}"
    fi

    # shellcheck disable=SC2120
    if [[ $# -ge 2 ]]; then
        # shellcheck disable=SC2034
        chain33ID="${2}"
    fi

    # init
    # shellcheck disable=SC2154
    Chain33Cli=${MainCli}
    InitChain33Validator
    # para add
    initPara

    # shellcheck disable=SC2154
    # shellcheck disable=SC2034
    Chain33Cli=${Para8901Cli}
    StartDockerRelayerDeploy

    test_all

    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
