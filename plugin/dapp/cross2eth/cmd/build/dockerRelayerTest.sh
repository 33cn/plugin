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


#0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697 0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2 0x1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695 0x4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf 0x62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9 0x355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71 0x72dff1c863631208a3d4f67a5fb0b7ebe69f827a75f332e2069dc7c825cb2202 0xecbc20b02e1ffd321e31c2a6d7d35a69715ba43ef2b0048a27de4f67b8249bde 0x1649955b3f2852a9cd71e50237b5a6f717539cffbe336bfcd95eb19a1b5c6f1b 0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0"

    for (( i = 0; i < 10; i++ )); do
        parallel --jobs 10 ${CLIA} ethereum lock -m 0.002 -k {} -r "${chain33ReceiverAddr}" ::: 0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697 0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2 0x1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695 0x4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf 0x62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9 0x355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71 0x72dff1c863631208a3d4f67a5fb0b7ebe69f827a75f332e2069dc7c825cb2202 0xecbc20b02e1ffd321e31c2a6d7d35a69715ba43ef2b0048a27de4f67b8249bde 0x1649955b3f2852a9cd71e50237b5a6f717539cffbe336bfcd95eb19a1b5c6f1b 0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd

#        ${CLIA} ethereum lock -m 0.002 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}"
    done

    sleep 4

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0.002"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    for (( i = 0; i < 10; i++ )); do
        parallel --jobs 10 ${CLIA} ethereum lock -m 0.002 -k {} -r "${chain33ReceiverAddr}" ::: 0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697 0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2 0x1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695 0x4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf 0x62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9 0x355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71 0x72dff1c863631208a3d4f67a5fb0b7ebe69f827a75f332e2069dc7c825cb2202 0xecbc20b02e1ffd321e31c2a6d7d35a69715ba43ef2b0048a27de4f67b8249bde 0x1649955b3f2852a9cd71e50237b5a6f717539cffbe336bfcd95eb19a1b5c6f1b 0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd
        parallel --jobs 10 ${CLIA} chain33 burn -m {} -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33MainBridgeTokenAddr}" ::: 0.001 0.002 0.003  0.001 0.002 0.003  0.001 0.002 0.003  0.002
#        ${CLIA} ethereum lock -m 0.002 -k "${ethTestAddrKey1}" -r "${chain33ReceiverAddr}"
    done

    sleep 4

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0.002"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")
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
