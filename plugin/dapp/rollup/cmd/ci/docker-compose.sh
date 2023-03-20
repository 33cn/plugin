#!/usr/bin/env bash

# debug mode
#set -x
# Exit immediately if a command exits with a non-zero status.
set -e
set -o pipefail
#set -o verbose
#set -o xtrace

# os: ubuntu16.04 x64
# first, you must install jq tool of json
# sudo apt-get install jq
# sudo apt-get install shellcheck, in order to static check shell script
# sudo apt-get install parallel
# ./docker-compose.sh build

PWD=$(cd "$(dirname "$0")" && pwd)
export PATH="$PWD:$PATH"

MAIN_NODE="${1}_main_1"
MAIN_CLI="docker exec ${MAIN_NODE} /root/chain33-cli"

PARA_NODE="${1}_para1_1"
CLI="docker exec ${PARA_NODE} /root/chain33-cli --paraName=user.p.para."

PARA_NODE2="${1}_para2_1"
CLI2="docker exec ${PARA_NODE2} /root/chain33-cli --paraName=user.p.para."

# shellcheck disable=SC2034

containers=("${MAIN_NODE}" "${PARA_NODE}" "${PARA_NODE2}")
export COMPOSE_PROJECT_NAME="$1"
## global config ###
sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi

echo "=========== # env setting ============="
echo "COMPOSE_FILE=$COMPOSE_FILE"
echo "COMPOSE_PROJECT_NAME=$COMPOSE_PROJECT_NAME"
echo "CLI=$CLI"
####################

# 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt  0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944
gensisKey="0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"
# 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn  982797e30031e9d1d00e0f1edf6da86c6a55338165f12efee7ff77e0d485f4d7
# 13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw  c9d5e87e94677df823e1cce0eab7de2e7bb4f724a12592821f84e12b72c639c2
testKey1="982797e30031e9d1d00e0f1edf6da86c6a55338165f12efee7ff77e0d485f4d7"
testKey2="c9d5e87e94677df823e1cce0eab7de2e7bb4f724a12592821f84e12b72c639c2"

function wait_height() {

    if [ "$#" -lt 3 ]; then
        echo "wait_height not enough params"
        exit 1
    fi
    cli="${1}"
    expect="${2}"
    timeout="${3}"
    echo "wait height ${expect}"
    while true; do
        new_height=$(${cli} block last_header | jq ".height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi

        timeout=$((timeout - 1))
        if [ "${timeout}" -lt 0 ]; then
            echo "wait height timeout, expect=${expect}, curr=${new_height}"
            exit 1
        fi
        sleep 0.1
    done

}

function set_main_config() {
    echo "====== ${FUNCNAME[0]} ======"
    conf="chain33.test.toml"
    rm -rf chain33.toml
    sed -i $sedfix '0,/^Title.*/s//Title="local"/' "${conf}"
    sed -i $sedfix '0,/^minTxFeeRate=.*/s//minTxFeeRate=0/' "${conf}"

    # rpc
    sed -i $sedfix '0,/^grpcBindAddr=.*/s//grpcBindAddr="0.0.0.0:8802"/' "${conf}"
    sed -i $sedfix '0,/^whitelist=.*/s//whitelist=["*"]/' "${conf}"
}

function set_para_config() {
    echo "====== ${FUNCNAME[0]}======"
    conf="chain33.para.toml"
    # blockchain
    sed -i $sedfix '0,/^isParaChain.*/s//isParaChain=false/' "${conf}"

    # mempool
    sed -i $sedfix '/^\[mempool\]/,/^name/s/^name.*/name="timeline"/' "${conf}"
    sed -i $sedfix '0,/^minTxFeeRate=.*/s//minTxFeeRate=0/' "${conf}"

    # p2p
    sed -i $sedfix '/^\[p2p\]/,/^enable=.*/s/^enable.*/enable=true/' "${conf}"
    sed -i $sedfix '0,/^waitPid=.*/{//d}' "${conf}"

    # rpc
    sed -i $sedfix '0,/^jrpcBindAddr=.*/s//jrpcBindAddr="localhost:8801"/' "${conf}"
    sed -i $sedfix '0,/^mainChainGrpcAddr=.*/s//mainChainGrpcAddr="main:8802"/' "${conf}"
    sed -i $sedfix '0,/^forwardExecs=.*/s//forwardExecs=["paracross"]/' "${conf}"
    sed -i $sedfix '0,/^forwardActionNames=.*/s//forwardActionNames=["crossAssetTransfer"]/' "${conf}"

    # consensus
    sed -i $sedfix '/^\[consensus\]/,/^name=.*/s/^name.*/name="solo"/' "${conf}"
    sed -i $sedfix '/^\[consensus\]/,/^committer=/{/^committer="/d}' "${conf}"
    sed -i $sedfix '/^\[consensus\]/,/^name=.*/!b;/^name/a committer="rollup"' "${conf}"
    sed -i $sedfix '0,/^minerstart=.*/{//d}' "${conf}"
    # rollup
    sed -i $sedfix '0,/^maxCommitInterval=.*/s//maxCommitInterval=60/' "${conf}"
    sed -i $sedfix '0,/^reservedMainHeight=.*/s//reservedMainHeight=1/' "${conf}"

    # fork
    sed -i $sedfix 's/^ForkBlockHash=.*/ForkBlockHash=0/' "${conf}"
    sed -i $sedfix 's/^ForkRootHash=.*/ForkRootHash=0/' "${conf}"

    cp "${conf}" chain33.para1.toml
    cp "${conf}" chain33.para2.toml

    sed -i $sedfix '/^\[consensus\]/,/^name/!b;/^name/a minerstart=true' chain33.para1.toml
    sed -i $sedfix '/^\[consensus.sub.rollup\]/,/^authKey/s/^authKey.*/authKey="982797e30031e9d1d00e0f1edf6da86c6a55338165f12efee7ff77e0d485f4d7"/' chain33.para1.toml

    sed -i $sedfix '0,/^singleMode=.*/{//d}' chain33.para2.toml
    sed -i $sedfix '/^\[consensus.sub.rollup\]/,/^authKey/s/^authKey.*/authKey="c9d5e87e94677df823e1cce0eab7de2e7bb4f724a12592821f84e12b72c639c2"/' chain33.para2.toml

}

function node_group_config() {

    echo "====== ${FUNCNAME[0]} ======"

    ${MAIN_CLI} send coins transfer -t 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn -a 1000 -k "${gensisKey}"
    ${MAIN_CLI} send coins transfer -t 13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw -a 1000 -k "${gensisKey}"
    ${MAIN_CLI} send coins send_exec -e paracross -a 100 -k "${testKey1}"
    ${MAIN_CLI} send coins send_exec -e paracross -a 100 -k "${testKey2}"

    ${MAIN_CLI} account balance -a 13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw
    nodeAddrs="1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn,13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw"
    blsPubs="99ae786cac6a6ee65718d3bc036a57f90f92f38b5d14e6f61a7cfe2468ff96eed9e162aa0c8bda1f770b6f78513583eb,\
    811575ddef2eee83d9f702f50268f5e85c0999518d52cecedd357799547904faa162a10bc517ca87d50393e3225e3cae"
    # frozen 10
    hash=$(${MAIN_CLI} send para nodegroup apply -a "${nodeAddrs}" -p "${blsPubs}" -c 10 --paraName=user.p.para. -k "${testKey1}")

    echo "apply hash =${hash}"
    ${MAIN_CLI} tx query -s "${hash}"
    res=$(${MAIN_CLI} send para nodegroup approve -c 10 -i "${hash}" --paraName=user.p.para. -k "${testKey1}")
    echo "approve hash =${res}"
    ${MAIN_CLI} tx query -s "${res}"

    if [ "$(${MAIN_CLI} rollup validator -t user.p.para. | jq -r ".blsPubs|length == 2")" != "true" ]; then
        echo "fail to config parachain node group"
        exit 1
    fi

}

function test_forward_tx() {
    echo "====== ${FUNCNAME[0]} start ======"
    lastHeight=$(${MAIN_CLI} block last_header | jq -r ".height")
    rawTx=$(${MAIN_CLI} coins transfer -t 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn -a 1)
    signTx=$(${MAIN_CLI} wallet sign -d "${rawTx}" -k "${gensisKey}")
    echo "== send main tx in parachain =="

    ${CLI} wallet send -d "${signTx}"
    sleep 1
    currHeight=$(${MAIN_CLI} block last_header | jq -r ".height")
    if [ "${currHeight}" -eq "${lastHeight}" ]; then
        echo "wait new block error, last=${lastHeight}, curr=${currHeight}"
        exit 1
    fi

    echo "====== ${FUNCNAME[0]} end ======"
}

function test_rollup_commit() {

    echo "====== ${FUNCNAME[0]} start ======"
    local height=0

    mainHeight=$(${MAIN_CLI} block last_header | jq -r ".height")
    echo "=== add blocks to parachain ==="
    while true; do

        ${CLI} send coins transfer -t 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn -a 1 -k "${gensisKey}"
        height=$((height + 1))
        wait_height "${CLI}" "${height}" 50
        if [ "${height}" -ge 32 ]; then
            break
        fi
    done

    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 300

    echo "=== check rollup commit status ==="
    res=$(${MAIN_CLI} rollup status -t user.p.para. | jq -r "(.commitRound == 1) and (.commitBlockHeight == 32)")
    if [ "${res}" != "true" ]; then
        echo "== ${FUNCNAME[0]} failed =="
        exit 1
    fi

    echo "====== ${FUNCNAME[0]} end ======"
}

function test_cross_chain() {

    echo "====== ${FUNCNAME[0]} start ======"
    mainHeight=$(${MAIN_CLI} block last_header | jq -r ".height")
    paraHeight=$(${CLI} block last_header | jq -r ".height")
    echo "=== mainchain bty -> parachain ==="
    ${CLI} send para cross_transfer -e coins -s bty --paraName=user.p.para. -a 2 -t 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn -k "${testKey1}"
    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 50
    echo "=== parachain para -> mainchain (tx group) ==="
    tx1=$(${CLI} coins send_exec -e paracross -a 10)
    tx2=$(${CLI} para cross_transfer -e user.p.para.coins -s para --paraName=user.p.para. -a 2 -t 13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw)
    gtx=$(${CLI} coins txgroup -t "${tx1} ${tx2}")
    signTx=$(${CLI} wallet sign -d "${gtx}" -k "${testKey1}")
    ${CLI} wallet send -d "${signTx}"

    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 50
    paraHeight=$((paraHeight + 1))
    wait_height "${CLI}" "${paraHeight}" 50

    echo "=== wait main block ==="
    ${MAIN_CLI} send coins transfer -t 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn -a 1 -k "${gensisKey}"
    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 50
    paraHeight=$((paraHeight + 1))
    wait_height "${CLI}" "${paraHeight}" 50

    echo "=== add blocks to parachain ==="
    for ((i = 1; i <= 30; i++)); do
        ${CLI} send coins transfer -t 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn -a 1 -k "${gensisKey}"
        paraHeight=$((paraHeight + 1))
        wait_height "${CLI}" "${paraHeight}" 50
    done

    if [ "$(${CLI} block last_header | jq -r ".height == 64")" != "true" ]; then

        echo "parachain block height should be 64"
        exit 1
    fi

    echo "=== wait rollup commit ==="
    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 300

    echo "=== check cross chain asset balance step 1==="

    balance=$(${CLI} asset balance -e paracross -a 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn \
        --asset_exec=paracross --asset_symbol=coins.bty | jq -r ".balance")
    if [ "${balance}" != "2.0000" ]; then
        echo "bty balance in parachain not correct, balance=${balance}"
        exit 1
    fi

    balance=$(${MAIN_CLI} asset balance -e paracross -a 13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw \
        --asset_exec=paracross --asset_symbol=user.p.para.coins.para | jq -r ".balance")
    if [ "${balance}" != "2.0000" ]; then
        echo "para balance in mainchain not correct, balance=${balance}"
        exit 1
    fi

    echo "=== para mainchain -> parachain ==="
    ${CLI} send para cross_transfer -e paracross -s user.p.para.coins.para \
        --paraName=user.p.para. -a 1 -t 13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw -k "${testKey2}"
    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 50

    frozen=$(${MAIN_CLI} asset balance -e paracross -a 13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw \
        --asset_exec=paracross --asset_symbol=user.p.para.coins.para | jq -r ".balance")
    if [ "${frozen}" != "1.0000" ]; then
        echo "para frozen in mainchain not correct, frozen=${frozen}"
        exit 1
    fi

    echo "=== bty parachain -> mainchain ==="
    ${CLI} send para cross_transfer -e user.p.para.paracross -s coins.bty \
        --paraName=user.p.para. -a 1 -t 14BQdkMhuVgJCRyPeUczUCB2BowGrbF3wK -k "${testKey1}"
    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 50

    ${MAIN_CLI} send coins transfer -t 1CqsBFa8KMGG9yjY4XcCUWaqdscBw6eipn -a 1 -k "${gensisKey}"
    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 50

    echo "=== wait for timeout rollup commit ... ==="
    sleep 60
    mainHeight=$((mainHeight + 1))
    wait_height "${MAIN_CLI}" "${mainHeight}" 300

    echo "=== check rollup commit status ==="
    round=$(${MAIN_CLI} rollup status -t user.p.para. | jq -r ".commitRound")
    if [ "${round}" -ne 3 ]; then
        echo "== check rollup status failed ==, round=${round}"
        exit 1
    fi

    echo "=== check cross chain asset balance step 2==="
    balance=$(${MAIN_CLI} asset balance -e paracross -a 14BQdkMhuVgJCRyPeUczUCB2BowGrbF3wK \
        --asset_exec=coins --asset_symbol=bty | jq -r ".balance")

    if [ "${balance}" != "1.0000" ]; then
        echo "bty balance in mainchain not correct, balance=${balance}"
        exit 1
    fi

    balance=$(${CLI} asset balance -e user.p.para.paracross -a 13mBGpucgALNZkqnb22NeQA5gZ1E1VpSjw \
        --asset_exec=coins --asset_symbol=para | jq -r ".balance")
    if [ "${balance}" != "1.0000" ]; then
        echo "para balance in parachain not correct, balance=${balance}"
        exit 1
    fi

    echo "====== ${FUNCNAME[0]} end ======"
}

function test_rollup() {

    test_forward_tx
    test_rollup_commit
    test_cross_chain
}

function start_docker() {
    echo "=========== # docker-compose ps ============="
    docker-compose ps

    # remove exsit container
    docker-compose down

    # create and run docker-compose container
    docker-compose up --build -d

    local SLEEP=5
    echo "=========== sleep ${SLEEP}s ============="
    sleep ${SLEEP}

    docker-compose ps

    # query node run status
    #    check_docker_status
    ${CLI} net peer
    local height=1000
    while [ $height -gt 0 ]; do
        peersCount=$(${CLI} net peer | jq '.[] | length')
        if [ "${peersCount}" -ge 2 ]; then
            break
        fi
        sleep 1
        ((height--))
        echo "peers error: peersCount=${peersCount}"
    done

    sync_status "${CLI}"
    sync_status "${CLI2}"

}

function check_docker_container() {
    echo "===== check_docker_container ======"
    for con in "${containers[@]}"; do
        runing=$(docker inspect "${con}" | jq '.[0].State.Running')
        if [ ! "${runing}" ]; then
            docker inspect "${con}"
            echo "check ${con} not actived!"
            exit 1
        fi
    done
}

function sync_status() {
    echo "=========== query sync status========== "
    local height=1000
    local wait_sec=0
    while [ $height -gt 0 ]; do
        status=$(${1} net is_sync)
        if [ "${status}" = "true" ]; then
            echo "=========== query clock sync status========== "
            status=$(${1} net is_clock_sync)
            if [ "${status}" = "true" ]; then
                break
            fi
        fi
        ((height--))
        wait_sec=$((wait_sec + 1))
        sleep 0.1
    done
    echo "sync wait  ${wait_sec}/10 s"
}

function main() {
    echo "====================DAPP=$DAPP main begin==================="
    ### init config ####
    set_main_config
    set_para_config

    ### start docker ####
    start_docker

    node_group_config

    test_rollup

    ### finish ###
    check_docker_container
    echo "===============DAPP=$DAPP main end==============="
}

# run script
main
