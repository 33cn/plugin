#!/usr/bin/env bash
# shellcheck disable=SC2154
set -e

PWD=$(cd "$(dirname "$0")" && pwd)
export PATH="$PWD:$PATH"

buildpath="temp"

NODE1="$buildpath""_parachain1_1"
CLI="docker exec ${NODE1} /root/chain33-cli"

NODE2="$buildpath""_parachain2_1"

NODE3="$buildpath""_parachain3_1"

NODE4="$buildpath""_parachain4_1"

CHAIN33_CLI="chain33-cli"
containers=("${NODE1}" "${NODE2}" "${NODE3}" "${NODE4}")
## global config ###
sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi

# shellcheck source=/dev/null
source config
####################

function para_init() {
    local index=1
    for auth in "${authAccount[@]}"; do
        tomlfile="chain33.para.$index.toml"
        para_set_toml "$tomlfile"
        sed -i $sedfix 's/^authAccount=.*/authAccount="'''"$auth"'''"/g' "$tomlfile"
        ((index++))
    done
}

function para_set_toml() {
    cp chain33.para.toml "${1}"

    sed -i $sedfix 's/^Title.*/Title="user.p.'''"$paraName"'''."/g' "${1}"
    sed -i $sedfix 's/^startHeight=.*/startHeight='''"$mainStartHeight"'''/g' "${1}"
    sed -i $sedfix 's/^mainChainGrpcAddr=.*/mainChainGrpcAddr='''"$mainChainGrpcAddr"'''/g' "${1}"
    sed -i $sedfix 's/^mainLoopCheckCommitTxDoneForkHeight=.*/mainLoopCheckCommitTxDoneForkHeight='''"$mainLoopCheckCommitTxDoneForkHeight"'''/g' "${1}"

    # rpc
    sed -i $sedfix 's/^jrpcBindAddr=.*/jrpcBindAddr="0.0.0.0:8901"/g' "${1}"
    sed -i $sedfix 's/^grpcBindAddr=.*/grpcBindAddr="0.0.0.0:8902"/g' "${1}"
    sed -i $sedfix 's/^whitelist=.*/whitelist=["localhost","127.0.0.1","0.0.0.0"]/g' "${1}"

    if [ -n "$superManager" ]; then
        # shellcheck disable=SC1004
        sed -i $sedfix 's/^superManager=.*/superManager='''"$superManager"'''/g' "${1}"
    fi

    if [ -n "$tokenApprs" ]; then
        # shellcheck disable=SC1004
        sed -i $sedfix 's/^tokenApprs=.*/tokenApprs='''"$tokenApprs"'''/g' "${1}"
    fi
}

function para_set_wallet() {
    echo "=========== # para set wallet ============="
    for ((i = 0; i < ${#authAccount[@]}; i++)); do
        para_import_wallet "${authPort[$i]}" "${authPrikey[$i]}"
    done
}

function para_import_wallet() {
    local key=$2
    local port=$1
    echo "=========== # save seed to wallet ============="
    ./$CHAIN33_CLI --rpc_laddr "http://localhost:$port" seed save -p 1314fuzamei -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib"

    echo "=========== # unlock wallet ============="
    ./$CHAIN33_CLI --rpc_laddr "http://localhost:$port" wallet unlock -p 1314fuzamei -t 0

    echo "=========== # import private key ============="
    echo "key: ${key}"
    ./$CHAIN33_CLI --rpc_laddr "http://localhost:$port" account import_key -k "${key}" -l "paraAuthAccount"

    echo "=========== # close auto mining ============="
    ./$CHAIN33_CLI --rpc_laddr "http://localhost:$port" wallet auto_mine -f 0

    echo "=========== # wallet status ============="
    ./$CHAIN33_CLI --rpc_laddr "http://localhost:$port" wallet status
}

function para_unlock_wallet() {
    for ((i = 0; i < ${#authPort[@]}; i++)); do
        echo "=========== # para unlock wallet ${authPort[$i]}============="
        ./$CHAIN33_CLI --rpc_laddr "http://localhost:${authPort[$i]}" wallet unlock -p 1314fuzamei -t 0
    done

}
function start() {
    echo "=========== # docker-compose ps ============="
    docker-compose ps
    docker-compose down

    # create and run docker-compose container
    docker-compose up --build -d

    local SLEEP=10
    echo "=========== sleep ${SLEEP}s ============="
    sleep ${SLEEP}

    docker-compose ps

    # query node run status
    echo "status"
    check_docker_status

    #    ./chain33-cli --rpc_laddr http://localhost:18901 block last_header
    $CLI --rpc_laddr http://localhost:8901 block last_header

}

function check_docker_status() {
    status=$(docker-compose ps | grep parachain1_1 | awk '{print $6}')
    statusPara=$(docker-compose ps | grep parachain1_1 | awk '{print $3}')
    if [ "${status}" == "Exit" ] || [ "${statusPara}" == "Exit" ]; then
        echo "=========== chain33 service Exit logs ========== "
        docker-compose logs parachain1
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

function query_tx() {
    sleep 5

    local times=100
    while true; do
        ret=$(${CLI} --rpc_laddr http://localhost:8901 tx query -s "${1}" | jq -r ".tx.hash")
        echo "query hash is ${1}, return ${ret} "
        if [ "${ret}" != "${1}" ]; then
            sleep 5
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "query tx=$1 failed"
                exit 1
            fi
        else
            echo "query tx=$1  success"
            break
        fi
    done
}

function create_yml() {
    touch docker-compose.yml
    cat >>docker-compose.yml <<EOF
version: '3'

services:
EOF

    for ((i = 1; i <= ${#authAccount[@]}; i++)); do
        cat >>docker-compose.yml <<EOF
  parachain$i:
    build:
      context: .
    entrypoint: /root/entrypoint.sh
    environment:
      PARAFILE: "/root/chain33.para.$i.toml"
    ports:
     - "1890$i:8901"
    volumes:
      - "../storage/parachain$i/paradatadir:/root/paradatadir"
      - "../storage/parachain$i/logs:/root/logs"
      - "../storage/parachain$i/parawallet:/root/parawallet"
EOF
    done

}

function create_storage() {
    mkdir -p storage
    cd storage

    for ((i = 0; i < ${#authAccount[@]}; i++)); do
        dirfile="parachain$i"
        mkdir -p "$dirfile"
    done

    cd ..
}

function create_build() {
    rm -rf $buildpath
    mkdir -p $buildpath
    cp chain33* Dockerfile ./*.sh "$buildpath"/
    cd $buildpath
    create_yml
}

function para_create_nodegroup() {
    echo "=========== # para chain create node group ============="
    local auths=""
    for auth in "${authAccount[@]}"; do
        if [ -z $auths ]; then
            auths="$auth"
        else
            auths="$auths,$auth"
        fi
    done
    echo "auths=$auths"

    ##apply
    txhash=$(${CLI} --rpc_laddr http://localhost:8901 --paraName "user.p.$paraName." send para nodegroup apply -a "$auths" -c "${authFrozenCoins}" -k "$applierPrikey")
    echo "tx=$txhash"
    query_tx "${txhash}"
    id=$txhash
    echo "need super manager approve id=$txhash"

    if [ -n "$superManagerPrikey" ]; then
        echo "=========== # para chain approve node group ============="
        ##approve
        txhash=$(${CLI} --rpc_laddr http://localhost:8901 --paraName "user.p.$paraName." send para nodegroup approve -i "$id" -c "${authFrozenCoins}" -k "$superManagerPrikey")
        echo "tx=$txhash"
        query_tx "${CLI}" "${txhash}"

        status=$(${CLI} --rpc_laddr http://localhost:8901 --paraName "user.p.$paraName." para nodegroup status | jq -r ".status")
        if [ "$status" != 2 ]; then
            echo "status not approve status=$status"
            exit 1
        fi

        ${CLI} --rpc_laddr http://localhost:8901 --paraName "user.p.$paraName." para nodegroup addrs
    fi

    echo "======== super node group config end ==================="
}

function main() {
    echo "==============================parachain startup op=$1========================================================"
    ### init para ####
    if [ "$1" == "start" ]; then
        create_storage
        create_build
        para_init

        ### start docker ####
        start
        ### finish ###
        check_docker_container
    fi

    if [ "$1" == "nodegroup" ]; then
        para_create_nodegroup
    fi

    if [ "$1" == "wallet" ]; then
        para_set_wallet
    fi

    if [ "$1" == "miner" ]; then
        para_unlock_wallet
    fi

    echo "===============================parachain startup end========================================================="
}

# run script
main "$1"
