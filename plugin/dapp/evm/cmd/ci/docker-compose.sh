#!/usr/bin/env bash

# debug mode
#set -x
# Exit immediately if a command exits with a non-zero status.
set -e
set -o pipefail
#set -o verbose
#set -o xtrace

# os: ubuntu18.04 x64
# first, you must install jq tool of json
# sudo apt-get install jq
# sudo apt-get install shellcheck, in order to static check shell script
# sudo apt-get install parallel
# ./docker-compose.sh build

PWD=$(cd "$(dirname "$0")" && pwd)
export PATH="$PWD:$PATH"

SOLO_NODE="${1}_main_1"
SOLO_CLI="docker exec ${SOLO_NODE} /root/chain33-cli --rpc_laddr http://localhost:8545"
Chain33_CLI="docker exec ${SOLO_NODE} /root/chain33-cli"
DAPP="evm"
# shellcheck disable=SC2034
CLI=$SOLO_CLI
containers=("${SOLO_NODE}")
export COMPOSE_PROJECT_NAME="$1"
## global config ###
sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi

echo "=========== # env setting ============="
echo "COMPOSE_FILE=$COMPOSE_FILE"
echo "COMPOSE_PROJECT_NAME=$COMPOSE_PROJECT_NAME"
echo "CLI=SOLO_CLI"
####################
#0xd83b69C56834E85e023B1738E69BFA2F0dd52905
genesisKey="c8729f05b10cc74d40feeb00376e11aa5b50e92b369d778b31b6e902c528f141"
genesisAddr="0xd83b69c56834e85e023b1738e69bfa2f0dd52905"
testAddr="0xDe79A84DD3A16BB91044167075dE17a1CA4b1d6b"


function set_main_config() {
    echo "====== ${FUNCNAME[0]} ======"
    conf="chain33.toml"
    sed -i $sedfix '0,/^Title.*/s//Title="local"/' "${conf}"
    sed -i $sedfix '0,/^CoinSymbol=.*/s//CoinSymbol="abc"/' "${conf}"
    #address defaultDriverb
     sed -i $sedfix '0,/^defaultDriver=.*/s//defaultDriver="eth"/' "${conf}"
    #blockchain singleMode
    sed -i $sedfix '0,/^singleMode=.*/s//singleMode=true/' "${conf}"
    # rpc
    sed -i $sedfix '0,/^whitelist=.*/s//whitelist=["*"]/' "${conf}"
    sed -i $sedfix '0,/^httpAddr=.*/s//httpAddr="0.0.0.0:8545"/' "${conf}"
    # evm ethMapFromSymbol addressDriver
    sed -i $sedfix '0,/^ethMapFromExecutor=.*/s//ethMapFromExecutor="coins"/' "${conf}"
    sed -i $sedfix '0,/^ethMapFromSymbol=.*/s//ethMapFromSymbol="abc"/' "${conf}"
    #consensus
    sed -i $sedfix '0,/^name="ticket"$/s//name="solo"/' "${conf}"
    #genesis
    sed -i $sedfix '0,/^genesis=.*/s//genesis="0xd83b69c56834e85e023b1738e69bfa2f0dd52905"/' "${conf}"

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

function testcase_coinsTransfer(){
    echo "====== ${FUNCNAME[0]} start ======"
    #coins 转账
    #构造交易
    echo "============= create eth tx ============="
    echo "cli:${CLI}"
    rawTx=$(${CLI} coins transfer_eth -f ${genesisAddr}  -t ${testAddr} -a 12)
    echo "${rawTx}"
     #如果返回空
     if [ -z "${rawTx}" ]; then
        exit 1
     fi
     echo "============= sign eth tx ============="
     #签名交易
    signData=$(${CLI} wallet sign -d ${rawTx} -c 2999 -p 2 -k ${genesisKey})
    #如果返回空
     if [ -z "${signData}" ]; then
        exit 1
     fi
    echo "${signData}"
    echo "============= send eth tx ============="
    hash=$(${CLI} wallet send -d ${signData} -e)
    if [ -z "${signData}" ]; then
        exit 1
    fi
    echo "${hash}"

    balance=$(${Chain33_CLI} account balance -a ${testAddr} -e coins | jq -r ".balance")
    if [ "${balance}" != "12.0000" ]; then
        echo " balance  not correct, balance=${balance}"
        exit 1
    fi

    echo "^_^check eth-evm-coins transfer success! ^_^ "

}

fnction testcase_deployErc20(){
  echo "=========== #start deployErc20 contract ============="

}

function testcase_evmPrecompile(){
  echo "=========== # evmPrecompile ============="
}


function run_tesstcase(){
  testcase_coinsTransfer
}
function main() {
     echo "====================DAPP=${DAPP} main begin==================="
    ### init config ####
    echo "#### set main_config"
    set_main_config
    ### start docker
    echo "#### start docker"
    start_docker
    run_tesstcase
    check_docker_container
    #finish
    echo "===============DAPP=$DAPP main end==============="
}

# start
main