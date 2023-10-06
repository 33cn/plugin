#!/usr/bin/env bash
# shellcheck disable=SC2034
# shellcheck disable=SC2154
# shellcheck disable=SC2155
# shellcheck disable=SC2086

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
source ../dapp-test-common.sh

PWD=$(cd "$(dirname "$0")" && pwd)
export PATH="$PWD:$PATH"

M1_NODE="${1}_main_1"
M2_NODE="${1}_miner2_1"
M3_NODE="${1}_miner3_1"
M4_NODE="${1}_miner4_1"
#测试节点
M1_EtH_CLI="docker exec ${M1_NODE} /root/chain33-cli --rpc_laddr http://localhost:8545"
M1_NODE_CLI="docker exec ${M1_NODE} /root/chain33-cli"
#挖矿节点
M2_NODE_CLI="docker exec ${M2_NODE} /root/chain33-cli"
M3_NODE_CLI="docker exec ${M3_NODE} /root/chain33-cli"
M4_NODE_CLI="docker exec ${M4_NODE} /root/chain33-cli"

DAPP="evm"
evm_contractAddr=""
MAIN_HTTP=""
ETH_HTTP=""
# shellcheck disable=SC2034
CLI=M1_NODE_CLI
ETHCLI= ${M1_EtH_CLI}
containers=("${M1_NODE}" "${M2_NODE}" "${M3_NODE}" "${M4_NODE}")
minerkeys=("${m4minerAddr}" "${m1minerkey}" "${m2minerkey}" "${m3minerkey}")
export COMPOSE_PROJECT_NAME="$1"
## global config ###
sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi
echo "=========== # env setting ============="
echo "COMPOSE_FILE=$COMPOSE_FILE"
echo "COMPOSE_PROJECT_NAME=$COMPOSE_PROJECT_NAME"
echo "ETHCLI=M1_EtH_CLI"
####################

m1minerAddr="0xd83b69c56834e85e023b1738e69bfa2f0dd52905"
m1minerkey="c8729f05b10cc74d40feeb00376e11aa5b50e92b369d778b31b6e902c528f141"
m2minerAddr="1HxQwHcVKKkRkb8kFASpXXdWNdqbj6ZwEw"
m2minerkey="0xf515bd419155f9251a4bca0df2ab815c268360c8dcb4b2c5b34bfc39fdbb5161"
m3minerAddr="0xe65ff24e1d175a5773374ad59beb153e0c111bc9"
m3minerkey="0xe953fb8476f65352f8e56ad0d3d0d90a893c7dc03f342d5768b628af688b8ad6"
m4normalAddr="0x66aeea39231d48701569df8b59bf5edaf8194e40"
m4addrkey="0x7999f44dfede5c94bba48d9cbabf8eb1cfccc65b7648e9cb2a172d96656cdc63"

m4minerAddr="0x1b4318867ed409897d095090c358c4256a01d22f"
m4minerkey="0x0c88e123c54439e31aabaffa97c39628fd597e9b60d7fe810fef1dfa72dff4ad"

genesis="0xDe79A84DD3A16BB91044167075dE17a1CA4b1d6b"
genesiskey="18f5bf55d3500d216ed3cba3bc1e417507c6c3daf951a3386a5477716d33a160"

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

function init_nodes() {
    ### test cases ###
    containerName=${1}
    cli="docker exec ${containerName} /root/chain33-cli"
    ip=$( ${cli} net info | jq -r ".externalAddr")
    ip=$(echo "$ip" | cut -d':' -f 1)
    if [ "$ip" == "127.0.0.1" ]; then
          ip=$(${cli} net info | jq -r ".localAddr")
          ip=$(echo "$ip" | cut -d':' -f 1)
    fi
     # ip="127.0.0.1"
    MCli=http://${ip}:8801

      #miner-1 创建导入seed
    pwd="a1234567"
    seed=${cli}  seed generate -l 0
    echo "$seed"
    ok=${cli} seed save -s $seed -p ${pwd} |jq -r ".isok"
    echo "save seed is ok: ${ok}"
    #wallet unlock 解锁钱包
    ${cli} wallet unlock -p ${pwd}
    islock=${cli} wallet status |jq -r ".isWalletLock"
    echo "wallet islock:${islock}"
    if [ ${islock} ]; then
      echo "node wallet is locked"
      exit 1
    fi
    #导入私钥
    ${cli} wallet import_key -k ${2} -l
    #打开自动挖矿
    $${cli} wallet auto_mine -f 1 -l ${1}
    #等待2个区块
    chain33_BlockWait ${2} ${MCli}
    echo "init miner-nodes ${containerName} success"

}

#run_testcase 跑测试用例，绑定挖矿
function run_testcase(){
  #普通节点导入私钥,十六进制地址
 # CLI account import_key -k "${m4addrkey}" -p 2 -l "test_miner"
  CLI account import_key -k "${m4minerkey}" -p 2 -l "mineraddr"
  CLI account list


  #从创世地址导出部分币用于挖矿 5000000 coins
  rawtx=$(${CLI} coins withdraw  -e ticket -a 5000000  -n "take test money")
  #签名sign
  signedRawTx=$(${CLI} wallet sign -d ${rawtx} -k ${genesiskey} -p 2)
  echo "signedTx:${signedRawTx}"
  #发送交易
  hash=${${CLI} wallet  send -d ${signedRawTx}}
  check_tx "${CLI}" "${hash}"

  #发送一部分coins到测试地址上
  rawtx=$(${CLI}  coins  coins transfer -a 4500000 -t ${m4normalAddr})
  signedRawTx=$(${CLI} wallet sign -d ${rawtx} -k ${genesiskey} -p 2)
  #发送交易
  hash=${${CLI} wallet  send -d ${signedRawTx}}
  check_tx "${CLI}" "${hash}"
  result=$(${CLI} account balance -a "${m4normalAddr}" -e coins)
  balance_ret "${result}" "4500000.0000"

  #发送一部分币到测试挖矿地址上
  rawtx=$(${CLI}  coins  coins transfer -a 10 -t ${m4minerAddr})
  signedRawTx=$(${CLI} wallet sign -d ${rawtx} -k ${m4addrkey} -p 2)
  #发送交易
  hash=${${CLI} wallet  send -d ${signedRawTx}}
  check_tx "${CLI}" "${hash}"
  result=$(${CLI} account balance -a "${m4minerAddr}" -e coins)
  balance_ret "${result}" "10.0000"

  #开始创建绑定交易
  rawtx=$(${CLI} ticket bind -b ${m4minerAddr} -o ${m4normalAddr})
  #用代理交易处理
  ethrawtx=$(${EthCLI} coins transfer_eth -d ${rawtx} -t "0x0000000000000000000000000000000000200005" -f ${m4normalAddr})

  #签名
  signedRawTx=$(${EthCLI} wallet sign -d ${ethrawtx} -p 2 -c 1999 -k ${m4addrkey})

  hash=$(${EthCLI}  wallet send -d ${signedRawTx} -e)
  if [ -z "${hash}" ]; then
    exit 1
  fi

  check_tx "${CLI}" "${hash}"



}



function main() {
     echo "====================DAPP=${DAPP} main begin==================="
    ### start docker
    echo "#### start docker"
    start_docker
    i=0
    for con in "${containers[@]}"; do
      #初始化各个容器  参数1:容器名称，参数2: 挖矿地址的私钥
      init_nodes con ${minerkeys[$i]}
      i++
     done


    run_testcase
    check_docker_container
    #finish
    docker-compose down
    echo "===============DAPP=$DAPP main end==============="
}

# start
main

