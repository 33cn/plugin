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
#(20) 0xd80fb30b4a8768a3a3841e8abEf606051725c0BB (1000 ETH)
#(21) 0x9eCD9837e8FB1FFACeCAd82ceAAfEc501A1CD8b6 (1000 ETH)
#(22) 0xbD5EDD0F081A3856FA9777E53cE6d000d97f41f2 (1000 ETH)
#(23) 0x558686dCa4659269f9F6DaA9aeA4B34a7EB78Ef8 (1000 ETH)
#(24) 0x82F986715Ead6BFA574335c4335f3eDFfd5C1Ff0 (1000 ETH)
#(25) 0x799Cf59a23D5277eee53A80eb788Fa6cd31D61dE (1000 ETH)
#(26) 0xb845B5820AB1Cb0Bbb64c23e4129351305941a11 (1000 ETH)
#(27) 0x0F36961190cD281750eFD635164F353FD79B4Bdc (1000 ETH)
#(28) 0x1b0A9eAc49c76aC436F3ccaD7625b263899ba726 (1000 ETH)
#(29) 0x3fD56e8bBbb051638824BaFFe5078Ed0b488a81B (1000 ETH)
#(20) 0x4bbe8e10a0987b6d56c505dfe41f74d357ac8211e91c4a179f9c5e38c181aaf0
#(21) 0xc2d896509c89c56d2365f40c5da5680174314884118fea024013818f792fcd64
#(22) 0x56e4f5884dbd4248e649bb5163815dd6fcae7a656f43419becd183e49fe2b514
#(23) 0x6ae0d6e5f14c1719f170100b6a84f3d7be14c2623404191d854ef98739e813a6
#(24) 0x0e8b158dae56bc58c69ea9cac8737317dc47d2fb50d40bea4f1629414a8f7846
#(25) 0x3d5ac5f963568544b8b855c8950030ec60cd0a0e9a5a8be725e790c625404fcd
#(26) 0x4cca1a474c3a9789fd11987d19ae13a3553c47c7b441e77739af5b2d48f01371
#(27) 0x8f23fc9a39795a1687e4f756312015d96b448829c32114981038ea74438361a0
#(28) 0xbf667b0227522bd3fd220e748c43d5529e323eca9bcb005ac58e6f2431577b28
#(29) 0x2e81ad3c49763fa9c0ea79b27e6cf584e6e073b43524c151ec1ade03706c1a00
    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0"

    for (( i = 0; i < 10; i++ )); do
        parallel --jobs 10 ${CLIA} ethereum lock -m 0.002 -k {} -r "${chain33ReceiverAddr}" ::: 0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697 0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2 0x1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695 0x4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf 0x62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9 0x355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71 0x72dff1c863631208a3d4f67a5fb0b7ebe69f827a75f332e2069dc7c825cb2202 0xecbc20b02e1ffd321e31c2a6d7d35a69715ba43ef2b0048a27de4f67b8249bde 0x1649955b3f2852a9cd71e50237b5a6f717539cffbe336bfcd95eb19a1b5c6f1b 0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd
    done

    sleep 4

    result=$(${CLIA} ethereum balance -o "${ethereumBridgeBank}")
#    cli_ret "${result}" "balance" ".balance" "0.002"
    result=$(${Chain33Cli} evm query -a "${chain33MainBridgeTokenAddr}" -c "${chain33DeployAddr}" -b "balanceOf(${chain33ReceiverAddr})")

    for (( i = 0; i < 10; i++ )); do
        parallel --jobs 10 ${CLIA} ethereum lock -m 0.002 -k {} -r "${chain33ReceiverAddr}" ::: 0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697 0x0c61f5a879d70807686e43eccc1f52987a15230ae0472902834af4d1933674f2 0x1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695 0x4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf 0x62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9 0x355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71 0x72dff1c863631208a3d4f67a5fb0b7ebe69f827a75f332e2069dc7c825cb2202 0xecbc20b02e1ffd321e31c2a6d7d35a69715ba43ef2b0048a27de4f67b8249bde 0x1649955b3f2852a9cd71e50237b5a6f717539cffbe336bfcd95eb19a1b5c6f1b 0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd
        parallel --jobs 10 ${CLIA} chain33 burn_increase -m {} -k "${chain33ReceiverAddrKey}" -r "${ethTestAddr2}" -t "${chain33MainBridgeTokenAddr}" ::: 0.001 0.002 0.003  0.001 0.002 0.003  0.001 0.002 0.003  0.002
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
#    test_test
    test_all
    TestRelayerProxy

    echo_addrs
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}
