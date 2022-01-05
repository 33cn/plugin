#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

# 主要在平行链上测试

source "./mainPubilcRelayerTest.sh"
source "./proxyVerifyTest.sh"

# shellcheck disable=SC2154
# shellcheck disable=SC2034
function test_test() {
    Boss4xCLI=${Boss4xCLIeth}
    CLIA=${CLIAeth}
    ethereumBridgeBank="${ethereumBridgeBankOnETH}"
    ethereumMultisignAddr="${ethereumMultisignAddrOnETH}"
    chain33MainBridgeTokenAddr="${chain33MainBridgeTokenAddrETH}"
    ethereumBtyBridgeTokenAddr="${ethereumBtyBridgeTokenAddrOnETH}"
    ethereumUSDTERC20TokenAddr="${ethereumUSDTERC20TokenAddrOnETH}"
    chain33USDTBridgeTokenAddr="${chain33USDTBridgeTokenAddrOnETH}"

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0"

    for (( i = 0; i < 1000; i++ )); do
        ${CLIA} ethereum lock -m 0.002 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}"
    done

    sleep 4

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0.002"

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
    test_test
#    test_all
#    TestRelayerProxy

    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
