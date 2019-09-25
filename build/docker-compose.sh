#!/usr/bin/env bash

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

NODE3="${1}_chain33_1"
CLI="docker exec ${NODE3} /root/chain33-cli"

NODE2="${1}_chain32_1"

NODE1="${1}_chain31_1"

NODE4="${1}_chain30_1"
#CLI4="docker exec ${NODE4} /root/chain33-cli"

NODE5="${1}_chain29_1"
CLI5="docker exec ${NODE5} /root/chain33-cli"

containers=("${NODE1}" "${NODE2}" "${NODE3}" "${NODE4}")
export COMPOSE_PROJECT_NAME="$1"
## global config ###
sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi
CLI_IP=""
PARA_CLI_IP=""
DAPP=""
if [ -n "${2}" ]; then
    DAPP=$2
fi

DAPP_TEST_FILE=""
if [ -n "${DAPP}" ]; then
    DAPP_TEST_FILE="testcase.sh"
    if [ -e "$DAPP_TEST_FILE" ]; then
        # shellcheck source=/dev/null
        source "${DAPP_TEST_FILE}"
    fi

    DAPP_COMPOSE_FILE="docker-compose-${DAPP}.yml"
    if [ -e "$DAPP_COMPOSE_FILE" ]; then
        export COMPOSE_FILE="docker-compose.yml:${DAPP_COMPOSE_FILE}"

    fi

fi

if [ -z "$DAPP" ] || [ "$DAPP" == "paracross" ]; then
    # shellcheck source=/dev/null
    source system-test-rpc.sh
    # shellcheck source=/dev/null
    source dapp-test-rpc.sh
fi

source dapp-test-common.sh

echo "=========== # env setting ============="
echo "DAPP=$DAPP"
echo "DAPP_TEST_FILE=$DAPP_TEST_FILE"
echo "COMPOSE_FILE=$COMPOSE_FILE"
echo "COMPOSE_PROJECT_NAME=$COMPOSE_PROJECT_NAME"
echo "CLI=$CLI"
####################

function base_init() {

    # update test environment
    sed -i $sedfix 's/^Title.*/Title="local"/g' chain33.toml
    sed -i $sedfix 's/^TestNet=.*/TestNet=true/g' chain33.toml

    sed -i $sedfix 's/^powLimitBits=.*/powLimitBits="0x1f2fffff"/g' chain33.toml
    sed -i $sedfix 's/^targetTimePerBlock=.*/targetTimePerBlock=1/g' chain33.toml
    sed -i $sedfix 's/^targetTimespan=.*/targetTimespan=10000000/g' chain33.toml

    # p2p
    sed -i $sedfix 's/^seeds=.*/seeds=["chain33:13802","chain32:13802","chain31:13802"]/g' chain33.toml
    #sed -i $sedfix 's/^enable=.*/enable=true/g' chain33.toml
    sed -i $sedfix '0,/^enable=.*/s//enable=true/' chain33.toml
    sed -i $sedfix 's/^isSeed=.*/isSeed=true/g' chain33.toml
    sed -i $sedfix 's/^innerSeedEnable=.*/innerSeedEnable=false/g' chain33.toml
    sed -i $sedfix 's/^useGithub=.*/useGithub=false/g' chain33.toml

    # rpc
    sed -i $sedfix 's/^jrpcBindAddr=.*/jrpcBindAddr="0.0.0.0:8801"/g' chain33.toml
    sed -i $sedfix 's/^grpcBindAddr=.*/grpcBindAddr="0.0.0.0:8802"/g' chain33.toml
    sed -i $sedfix 's/^whitelist=.*/whitelist=["localhost","127.0.0.1","0.0.0.0"]/g' chain33.toml

    # wallet
    sed -i $sedfix 's/^minerdisable=.*/minerdisable=false/g' chain33.toml

    sed -i $sedfix 's/^nodeGroupFrozenCoins=.*/nodeGroupFrozenCoins=20/g' chain33.toml
    sed -i $sedfix 's/^paraConsensusStopBlocks=.*/paraConsensusStopBlocks=100/g' chain33.toml

    # ticket
    sed -i $sedfix 's/^ticketPrice =.*/ticketPrice = 10000/g' chain33.toml

    #relay genesis
    sed -i $sedfix 's/^genesis="12qyocayNF7.*/genesis="1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3"/g' chain33.toml

}

function start() {
    echo "=========== # docker-compose ps ============="
    docker-compose ps

    # remove exsit container
    docker-compose down

    # create and run docker-compose container
    #docker-compose -f docker-compose.yml -f docker-compose-paracross.yml -f docker-compose-relay.yml up --build -d
    docker-compose up --build -d

    local SLEEP=10
    echo "=========== sleep ${SLEEP}s ============="
    sleep ${SLEEP}

    docker-compose ps

    # query node run status
    check_docker_status
    ${CLI} block last_header
    ${CLI} net info

    get_CLI_IP

    ${CLI} net peer_info
    local count=1000
    while [ $count -gt 0 ]; do
        peersCount=$(${CLI} net peer_info | jq '.[] | length')
        if [ "${peersCount}" -ge 2 ]; then
            break
        fi
        sleep 1
        ((count--))
        echo "peers error: peersCount=${peersCount}"
    done

    miner "${CLI}"
    # miner "${CLI4}"
    block_wait "${CLI}" 1

    echo "=========== check genesis hash ========== "
    ${CLI} block hash -t 0
    res=$(${CLI} block hash -t 0 | jq -r ".hash")
    #in case changes result in genesis change
    if [ "${res}" != "0x67c58d6ba9175313f0468ae4e0ddec946549af7748037c2fdd5d54298afd20b6" ]; then
        echo "genesis hash error!"
        exit 1
    fi

    echo "=========== query height ========== "
    ${CLI} block last_header
    result=$(${CLI} block last_header | jq ".height")
    if [ "${result}" -lt 1 ]; then
        block_wait "${CLI}" 2
    fi

    sync_status "${CLI}"

    ${CLI} wallet status
    ${CLI} account list
    ${CLI} mempool list
}

function miner() {
    #echo "=========== # create seed for wallet ============="
    #seed=$(${1} seed generate -l 0 | jq ".seed")
    #if [ -z "${seed}" ]; then
    #    exit 1
    #fi

    echo "=========== # save seed to wallet ============="
    result=$(${1} seed save -p 1314fuzamei -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib" | jq ".isok")
    if [ "${result}" = "false" ]; then
        echo "save seed to wallet error seed, result: ${result}"
        exit 1
    fi

    echo "=========== # unlock wallet ============="
    result=$(${1} wallet unlock -p 1314fuzamei -t 0 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi

    echo "=========== # import private key returnAddr ============="
    result=$(${1} account import_key -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944 -l returnAddr | jq ".label")
    echo "${result}"
    if [ -z "${result}" ]; then
        exit 1
    fi

    echo "=========== # import private key mining ============="
    result=$(${1} account import_key -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01 -l minerAddr | jq ".label")
    echo "${result}"
    if [ -z "${result}" ]; then
        exit 1
    fi

    echo "=========== # close auto mining ============="
    result=$(${1} wallet auto_mine -f 1 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi

}
function block_wait() {
    if [ "$#" -lt 2 ]; then
        echo "wrong block_wait params"
        exit 1
    fi
    local hp=$CLI_IP
    #local s1="para"
    #result=$(echo "$1" | grep "${s1}")
    #if [ -n $result ]; then
    #   hp=${CLI_IP//8801/8901}
    #fi
    echo "http=$hp"
    chain33_BlockWait "${2}" "${hp}"
}

function block_wait2height() {
    if [ "$#" -lt 3 ]; then
        echo "wrong block_wait params"
        exit 1
    fi
    local count=0
    local new_height=0
    local expect=${2}
    local isPara=${3}

    while true; do
        new_height=$(${1} block last_header | jq ".height")
        if [ "$isPara" == "1" ]; then
            ${1} para blocks -s "$new_height" -e "$new_height"
            new_height=$(${1} para blocks -s "$new_height" -e "$new_height" | jq ".items[0].mainHeight")
        fi
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 0.1
    done
    echo "wait new block $count/10 s, cur_height=$new_height,expect=$expect"
}

function check_docker_status() {
    status=$(docker-compose ps | grep chain33_1 | awk '{print $6}')
    statusPara=$(docker-compose ps | grep chain33_1 | awk '{print $3}')
    if [ "${status}" == "Exit" ] || [ "${statusPara}" == "Exit" ]; then
        echo "=========== chain33 service Exit logs ========== "
        docker-compose logs chain33
        echo "=========== chain33 service Exit logs End========== "
    fi

}

function check_docker_container() {
    echo "============== check_docker_container ==============================="
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
    local sync_status
    local count=1000
    local wait_sec=0
    while [ $count -gt 0 ]; do
        sync_status=$(${1} net is_sync)
        if [ "${sync_status}" = "true" ]; then
            echo "=========== query clock sync status========== "
            sync_status=$(${1} net is_clock_sync)
            if [ "${sync_status}" = "true" ]; then
                break
            fi
        fi
        ((count--))
        wait_sec=$((wait_sec + 1))
        sleep 0.1
    done
    echo "sync wait  ${wait_sec}/10 s"
}

function sync() {
    echo "=========== stop  ${NODE5} node========== "
    docker stop "${NODE5}"
    sleep 10

    echo "=========== start ${NODE5} node========== "
    docker start "${NODE5}"

    sleep 1
    sync_status "${CLI5}"
}

function transfer() {
    echo "=========== # transfer ============="
    hashes=()
    for ((i = 0; i < 10; i++)); do
        hash=$(${1} send coins transfer -a 1 -n test -t 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
        hashes=("${hashes[@]}" "$hash")
        echo "$hash"
    done
    echo "send end"
    block_wait "${1}" 1
    echo "len: ${#hashes[@]}"
    if [ "${#hashes[@]}" != 10 ]; then
        echo "tx number wrong"
        exit 1
    fi
    echo "query hash start"
    for ((i = 0; i < ${#hashes[*]}; i++)); do
        txs=$(${1} tx query_hash -s "${hashes[$i]}" | jq ".txs")
        if [ -z "${txs}" ]; then
            echo "cannot find tx"
            exit 1
        fi
    done

    echo "=========== # withdraw ============="
    hash=$(${1} send coins transfer -a 2 -n deposit -t 1wvmD6RNHzwhY4eN75WnM6JcaAvNQ4nHx -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    echo "${hash}"
    block_wait "${1}" 1
    before=$(${1} account balance -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -e retrieve | jq -r ".balance")
    if [ "${before}" == "0.0000" ]; then
        echo "wrong ticket balance, should not be zero"
        exit 1
    fi

    hash=$(${1} send coins withdraw -a 1 -n withdraw -e retrieve -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    echo "${hash}"
    block_wait "${1}" 1
    txs=$(${1} tx query_hash -s "${hash}" | jq ".txs")
    if [ "${txs}" == "null" ]; then
        echo "withdraw cannot find tx"
        exit 1
    fi

    hash=$(${1} send coins transfer -a 1000 -n transfer -t 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    block_wait "${1}" 1
}

function dapp_test_address() {
    echo "=========== # import private key dapptest1 mining ============="
    result=$(${1} account import_key -k 56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138 -l dapptest1 | jq ".label")
    echo "${result}"
    if [ -z "${result}" ]; then
        exit 1
    fi

    echo "=========== # import private key dapptest2 mining ============="
    result=$(${1} account import_key -k 2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989 -l dapptest2 | jq ".label")
    echo "${result}"
    if [ -z "${result}" ]; then
        exit 1
    fi
    result=$(${1} account import_key -k 9d315182e56fde7fadb94408d360203894e5134216944e858f9b31f70e9ecf40 -l rpctestpooladdr | jq ".label")
    echo "${result}"
    if [ -z "${result}" ]; then
        exit 1
    fi

    block_wait "${1}" 1

    hash=$(${1} send coins transfer -a 1500 -n transfer -t 1PUiGcbsccfxW3zuvHXZBJfznziph5miAo -k 2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989)
    echo "${hash}"

    #total allocation for rpc test
    hash=$(${1} send coins transfer -a 8000 -n transfer -t 1PcGKYYoLn1PLLJJodc1UpgWGeFAQasAkx -k 2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989)
    echo "${hash}"

    block_wait "${1}" 1
}

function base_config() {
    #    sync
    transfer "${CLI}"
    #    transfer "${CLI4}"
}

function rpc_test() {
    if [ "$DAPP" == "" ]; then
        system_test_rpc "http://${1}:8801"
        dapp_test_address "${CLI}"
        dapp_test_rpc "http://${1}:8801"
    fi
    if [ "$DAPP" == "paracross" ]; then
        system_test_rpc "http://${1}:8901"
        dapp_test_address "${CLI}"

        dapp_test_rpc "http://${1}:8901"
    fi

}
function dapp_run() {
    if [ -e "$DAPP_TEST_FILE" ]; then
        ${DAPP} "${CLI}" "${1}" "${2}"
    fi

}

function get_CLI_IP() {
    local cli_ip=$(${CLI} net info | jq -r ".externalAddr")
    cli_ip=$(echo "$cli_ip" | cut -d':' -f 1)
    CLI_IP="http://${cli_ip}:8801"
    PARA_CLI_IP=${CLI_IP//8801/8901}
    echo "CLI_IP=$CLI_IP,PARA_CLI_IP=$PARA_CLI_IP"
}

function main() {
    echo "==============================DAPP=$DAPP main begin========================================================"
    ### init para ####
    base_init
    dapp_run init

    ### start docker ####
    start

    ### config env ###
    base_config
    dapp_run config

    ### test cases ###
    ip=$(${CLI} net info | jq -r ".externalAddr")
    ip=$(echo "$ip" | cut -d':' -f 1)
    dapp_run test "${ip}"

    ### rpc test  ###
    rpc_test "${ip}"

    ### finish ###
    check_docker_container
    echo "===============================DAPP=$DAPP main end========================================================="
}

# run script
main
