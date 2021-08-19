#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

source "./publicTest.sh"
source "./relayerPublic.sh"

# ETH 部署合约者的私钥 用于部署合约时签名使用
#ethDeployAddr="0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a"
ethDeployKey="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

ethSendAddress=0x495953A743ef169EC5D4aC7b5F786BF2Bd56aFd5
ethSendPrivateKeys=0x2809477ede1261da21270096776ba7dc68b89c9df5f029965eaa5fe7f0b80697

# validatorsAddr=["0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]
#ethValidatorAddrKeyA="8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"

# chain33 部署合约者的私钥 用于部署合约时签名使用
chain33DeployAddr="1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ"
#chain33DeployKey="0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"

chain33ReceiverAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
chain33ReceiverAddrKey="4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"

#maturityDegree=10

Chain33Cli="../../chain33-cli"
chain33BridgeBank=""
#ethBridgeBank=""
chain33BtyTokenAddr="1111111111111111111114oLvT2"
chain33EthTokenAddr=""
ethereumBtyTokenAddr=""
chain33YccTokenAddr=""
ethereumYccTokenAddr=""

CLIA="./ebcli_A"
chain33ID=33

# chain33 lock BTY, eth burn BTY
function LockTestChain33ToEthAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # chain33 lock bty
    hash=$(${Chain33Cli} evm call -f 1 -a 1 -c "${chain33DeployAddr}" -e "${chain33BridgeBank}" -p "lock(${ethSendAddress}, ${chain33BtyTokenAddr}, 100000000)" --chainID "${chain33ID}")
    check_tx "${Chain33Cli}" "${hash}"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# chain33 lock BTY, eth burn BTY
function BurnTestChain33ToEthAssets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # eth burn
    result=$(${CLIA} ethereum burn -m 1 -k "${ethSendPrivateKeys}" -r "${chain33DeployAddr}" -t "${ethereumBtyTokenAddr}" ) #--node_addr https://ropsten.infura.io/v3/9e83f296716142ffbaeaafc05790f26c)
    cli_ret "${result}" "burn"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn
function LockTestETH2Chain33Assets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # ETH 这端 lock 11个
    result=$(${CLIA} ethereum lock -m 2 -k "${ethSendPrivateKeys}" -r "${chain33ReceiverAddr}")
    cli_ret "${result}" "lock"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

# eth to chain33 在以太坊上锁定 ETH 资产,然后在 chain33 上 burn
function BurnTestETH2Chain33Assets() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    result=$(${CLIA} chain33 burn -m 2 -k "${chain33ReceiverAddrKey}" -r "${ethSendAddress}" -t "${chain33EthTokenAddr}")
    cli_ret "${result}" "burn"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function LockTestETH2Chain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    # ETH 这端 lock 7个 YCC
    result=$(${CLIA} ethereum lock -m 3 -k "${ethSendPrivateKeys}" -r "${chain33ReceiverAddr}" -t "${ethereumYccTokenAddr}")
    cli_ret "${result}" "lock"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function BurnTestETH2Chain33Ycc() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    echo '#5.burn YCC from Chain33 YCC(Chain33)-----> Ethereum'
    result=$(${CLIA} chain33 burn -m 3 -k "${chain33ReceiverAddrKey}" -r "${ethSendAddress}" -t "${chain33YccTokenAddr}")
    cli_ret "${result}" "burn"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function mainTest() {
    StartChain33
    start_trufflesuite
    AllRelayerStart

    ${CLIA} ethereum token token_transfer -k "${ethDeployKey}" -m 10000 -r "${ethSendAddress}" -t "${ethereumYccTokenAddr}"

    for (( i = 0; i < 10; i++ )); do
        LockTestChain33ToEthAssets
        LockTestETH2Chain33Assets
        LockTestETH2Chain33Ycc
        sleep 1
    done

    while true ; do
        LockTestChain33ToEthAssets
        LockTestETH2Chain33Assets
        LockTestETH2Chain33Ycc

        eth_block_wait 2

        BurnTestChain33ToEthAssets
        BurnTestETH2Chain33Assets
        BurnTestETH2Chain33Ycc

        eth_block_wait 10
    done
}

mainTest

