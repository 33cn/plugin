#!/usr/bin/env bash
# shellcheck disable=SC2034
# shellcheck disable=SC2154
# shellcheck disable=SC2155
# shellcheck disable=SC2086
# shellcheck disable=SC2004
# shellcheck disable=SC1091

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

M1_NODE="${1}_chain33_1"
M2_NODE="${1}_chain32_1"
M3_NODE="${1}_chain31_1"
M4_NODE="${1}_chain30_1"
#测试节点
M1_NODE_CLI="docker exec ${M1_NODE} /root/chain33-cli"
ETH_CLI="docker exec ${M1_NODE} /root/chain33-cli --rpc_laddr http://localhost:8545"
#挖矿节点
M2_NODE_CLI="docker exec ${M2_NODE} /root/chain33-cli"
#M3_NODE_CLI="docker exec ${M3_NODE} /root/chain33-cli"
#M4_NODE_CLI="docker exec ${M4_NODE} /root/chain33-cli"

DAPP="evm2"
MAIN_HTTP=""
ETH_HTTP=""
# shellcheck disable=SC2034
CLI=${M1_NODE_CLI}
containers=("${M1_NODE}" "${M2_NODE}" "${M3_NODE}" "${M4_NODE}")

export COMPOSE_PROJECT_NAME="$1"

## global config ###
sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi

echo "=========== # env setting ============="
echo "COMPOSE_FILE=$COMPOSE_FILE"
echo "COMPOSE_PROJECT_NAME=$COMPOSE_PROJECT_NAME"
echo "ETHCLI=ETH_CLI"
echo "CLI=${CLI}"
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
minerkeys=("${m4minerkey}" "${m1minerkey}" "${m2minerkey}" "${m3minerkey}")
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
    echo "init_nodes containerName:${containerName}"
    cli="docker exec ${containerName} /root/chain33-cli"
    ip=$( ${cli} net info | jq -r ".externalAddr")
    ip=$(echo "$ip" | cut -d':' -f 1)
    if [ "$ip" == "127.0.0.1" ]; then
          ip=$(${cli} net info | jq -r ".localAddr")
          ip=$(echo "$ip" | cut -d':' -f 1)
    fi
     # ip="127.0.0.1"
    MCli=http://${ip}:8801
    echo "mcli: ${MCli}"
    #miner-1 创建导入seed
    pwd="a1234567"
    seed=$(${cli}  seed generate -l 0)
    echo "seed:$seed"

    ok=$(${cli} seed save -s "${seed}" -p ${pwd}| jq -r ".isOK")
    echo "save seed is ok: ${ok}"
    #wallet unlock 解锁钱包
    ${cli} wallet unlock -p ${pwd}
    islock=$(${cli} wallet status| jq -r ".isWalletLock" )
    echo "wallet islock:${islock}"
    if [ ${islock} == true ]; then
      echo "node wallet is locked"
      exit 1
    fi
    #导入私钥
    echo "key: ${2}"
    ${cli} account import_key -k "${2}" -l ${1} -t ${3}
    #打开自动挖矿
    ${cli} wallet auto_mine -f 1


    echo "init miner-nodes ${containerName} success"

}

#run_testcase 跑测试用例，绑定挖矿,向ticket 合约转账,解绑挖矿
function run_testcase(){
  #普通节点导入私钥,十六进制地址
  ${CLI} account list
  ip=$( ${CLI} net info | jq -r ".externalAddr")
  MCli=http://${ip}:8801
  ECli=http://${ip}:8545
  chain33_BlockWait 1  ${MCli}
  #先给genesisaddr 打一点币
  hash=$( ${M2_NODE_CLI} send coins transfer -t ${genesis} -k ${m2minerkey} -a 100 -n "send to genesisaddr" )
  echo "run_testcase hash:${hash},genesisaddr: ${genesis}"
  if [ -z "${hash}" ]; then
    echo "hash empty"
    exit 1
  fi
  chain33_BlockWait 1  ${MCli}
  chain33_QueryTx ${hash}  ${MCli}
  #从创世地址导出部分币用于挖矿 110000 coins
  chain33_BlockWait 10 ${MCli}
  targeBalance=110000
  while true; do
  result=$(${CLI} account balance -a "${genesis}" -e ticket | jq -r ".balance")
  balance=$(printf "%.0f\n" $result)
  echo "balance: $(("${balance}")) ,timestamp:$(date +"%s")"
  if [ $((balance))  -ge ${targeBalance} ];then
    break
  fi
  sleep 1
  done

  local rawtx=$(${CLI} coins withdraw  -e ticket -a 110000  -n "take test money")
  echo "withdraw coins rawtx:${rawtx}"
  #签名sign
  local signedRawTx=$(${CLI} wallet sign -d ${rawtx} -k ${genesiskey} -p 2)
  echo "withdraw coins signedTx:${signedRawTx}"
  #发送交易
  local hash=$(${CLI} wallet  send -d ${signedRawTx})
  echo "withdraw coins hash:${hash}"
  chain33_QueryTx ${hash} ${MCli}

  #发送一部分coins到测试地址上
  echo "send coins to testaddress:${m4normalAddr}"

  local rawtx=$(${CLI} coins transfer -a 100001 -t ${m4normalAddr})
  local signedRawTx=$(${CLI} wallet sign -d ${rawtx} -k ${genesiskey} -p 2)
  #发送交易
  local hash=$(${CLI} wallet  send -d ${signedRawTx})
  echo "send coins hash:${hash}"
  chain33_QueryTx  ${hash} ${MCli}
  result=$(${CLI} account balance -a "${m4normalAddr}" -e coins)
  echo "query balance---->:${result}"
  balance_ret "${result}" "100001.0000"
  echo "balance check success"
  #发送一部分币到测试挖矿地址上
  local hash=$(${CLI}  send coins transfer -t ${m4minerAddr} -a 10 -k ${m2minerkey})
  chain33_QueryTx ${hash} ${MCli}
  result=$(${CLI} account balance -a "${m4minerAddr}" -e coins)
  balance_ret "${result}" "10.0000"


  #代理绑定验证
  bindMiner
  #代理买票挖矿
  trans2Ticket
  #关闭票数
  closeColdAddrTicket
  #取回ticket合约下的余额到coins
  withdrawTicketBalance
  #解绑
  closeBindMiner
  echo "=============test close bindminer end ^_^============="
  echo "========== all test finished ,end=========="

}

function bindMiner(){
      ip=$( ${CLI} net info | jq -r ".externalAddr")
      MCli=http://${ip}:8801
      ECli=http://${ip}:8545
      #开始创建绑定交易
      echo "===========start test bindminer ============"
      local rawtx=$(${CLI} ticket bind -b ${m4minerAddr} -o ${m4normalAddr})
      echo "bindminer rawtx:${rawtx}"
      #用代理交易处理
      local ethrawtx=$(${CLI} coins transfer_eth -d ${rawtx} -t "0x0000000000000000000000000000000000200005" -f ${m4normalAddr} --rpc_laddr ${ECli} )
      echo "binminer ethrawtx:${ethrawtx}"
      #签名
      local signedRawTx=$(${CLI} wallet sign -d ${ethrawtx} -p 2 -c 1999 -k ${m4addrkey} --rpc_laddr ${ECli} )
      echo "bindminer signedrawTx:${signedRawTx}"
      local hash=$(${CLI}  wallet send -d ${signedRawTx} -e --rpc_laddr ${ECli} )
      if [ -z "${hash}" ]; then
        exit 1
      fi

      chain33_QueryTx ${hash} ${MCli}
      #检查节点的绑定挖矿的冷钱包地址
      coldAddrs=$(${CLI} ticket cold -m ${m4minerAddr} |jq ".datas")
      preCut=${coldAddrs#*[}
      realColdAddr=$( echo ${preCut%]*}|tr -d '\n' |tr -d '"')
      echo "coldAddr:${realColdAddr}"
      echo "m4normal:${m4normalAddr}"

      if [ "${realColdAddr}" != "${m4normalAddr}" ]; then
        echo "addr should equal"
         exit 1
      fi
}
function closeBindMiner() {
    ip=$( ${CLI} net info | jq -r ".externalAddr")
    MCli=http://${ip}:8801
    ECli=http://${ip}:8545
    #测试解除绑定挖矿
      echo "=============test close bindminer============="
      local rawtx=$(${CLI} ticket bind -b "" -o ${m4normalAddr})
      echo "close bindminer rawtx:${rawtx}"
      #用代理交易处理
      local ethrawtx=$(${CLI} coins transfer_eth -d ${rawtx} -t "0x0000000000000000000000000000000000200005" -f ${m4normalAddr} --rpc_laddr ${ECli} )
      echo "close binminer ethrawtx:${ethrawtx}"
      #签名
      local signedRawTx=$(${CLI} wallet sign -d ${ethrawtx} -p 2 -c 1999 -k ${m4addrkey} --rpc_laddr ${ECli} )
      echo "close bindminer signedrawTx:${signedRawTx}"
      local hash=$(${CLI}  wallet send -d ${signedRawTx} -e --rpc_laddr ${ECli} )

      if [ -z "${hash}" ]; then
          exit 1
      fi

      chain33_QueryTx ${hash} ${MCli}
      #检查节点的绑定挖矿的冷钱包地址
      coldAddrs=$(${CLI} ticket cold -m ${m4minerAddr} |jq ".datas")
      preCut=${coldAddrs#*[}
      realColdAddr=$( echo ${preCut%]*}|tr -d '\n' |tr -d '"')
      echo "coldAddr:${realColdAddr}"
      if [ -n "${realColdAddr}" ];then
          exit 1

      fi
}

#关闭代理挖矿源地址的票
function closeColdAddrTicket() {
      echo "==========closeColdAddrTicket============"
      ip=$( ${CLI} net info | jq -r ".externalAddr")
      MCli=http://${ip}:8801
      ECli=http://${ip}:8545

      ticketNum=$(${CLI} ticket count)
      echo "ticketNum:${ticketNum} time:$(date +"%s") "
      result=$(${CLI} account balance -a "${m4normalAddr}" -e ticket)
      echo "ticket balance:${result}"
      ${CLI} wallet auto_mine -f 0
      sleep 10
      hash=$(${CLI} ticket close | jq ".hashes")
      preCut=${hash#*[}
      realhash=$( echo ${preCut%]*}|tr -d '\n' |tr -d '"')
      echo "ticket close hash:${realhash}"
      if [ -z "${realhash}" ]; then
          exit 1
      fi
      chain33_QueryTx ${realhash} ${MCli}
      #Check ticket balance/frozen
      while true; do
        result=$(${CLI} account balance -a "${m4normalAddr}" -e ticket)
        local balance=$(echo "${result}" | jq -r ".balance")
        local frozen=$(echo "${result}" | jq -r ".frozen")
        echo "ticket balance:${balance},addr:${m4normalAddr},timestamp:$(date +"%s")"
        echo "ticket frozen:${frozen},addr:${m4normalAddr},timestamp:$(date +"%s")"
        checkBalance=$(printf "%.0f\n" $frozen)
        echo "checkBalance: $(("${checkBalance}")) "
        if [ $((checkBalance))  -eq 0 ];then
            break
        fi
        sleep 1
      done


}

#withdrawTicketBalance 从ticket合约中取走coins
function withdrawTicketBalance(){
   echo "==========withdrawTicketBalance============"
   ip=$( ${CLI} net info | jq -r ".externalAddr")
   MCli=http://${ip}:8801
   ECli=http://${ip}:8545
   result=$(${CLI} account balance -a "${m4normalAddr}" -e ticket)
   local balance=$(echo "${result}" | jq -r ".balance")
   tbalance=$(printf "%.0f\n" $balance)
   echo "ticket balance:${tbalance},addr:${m4normalAddr},timestamp:$(date +"%s")"

   local withdrawTx=$( ${CLI} coins withdraw  -e ticket -a 100000  -n "take test money")
   echo "withdraw rawTx:${withdrawTx}"
    #用代理交易处理
   local ethrawtx=$(${CLI} coins transfer_eth -d ${withdrawTx} -t "0x0000000000000000000000000000000000200005" -f ${m4normalAddr} --rpc_laddr ${ECli} )
   echo "withdraw ethrawtx:${ethrawtx}"
   #签名
   local signedRawTx=$(${CLI} wallet sign -d ${ethrawtx} -p 2 -c 1999 -k ${m4addrkey} --rpc_laddr ${ECli} )
   echo "withdraw signedrawTx:${signedRawTx}"
   local hash=$(${CLI}  wallet send -d ${signedRawTx} -e --rpc_laddr ${ECli} )
   if [ -z "${hash}" ]; then
      exit 1
   fi
   echo "withdraw hash:${hash}"
   chain33_QueryTx ${hash} ${MCli}
   #check coins balance
   local result=$(${CLI} account balance -a "${m4normalAddr}" -e coins)
   local balance=$(echo "${result}" | jq -r ".balance")
   echo "coldminer addr coins balance:${balance}"
   local checkBalance=$(printf "%.0f\n" $balance)
   echo "checkBalance: $(("${checkBalance}")) "
   if [ $((checkBalance))  -lt $((tbalance)) ]; then
               exit 1
   fi

}
#trans2Ticket 转账到ticket合约下进行挖矿操作
function trans2Ticket() {
      echo "==========test transfer coins to ticket contractor============"
      ip=$( ${CLI} net info | jq -r ".externalAddr")
      MCli=http://${ip}:8801
      ECli=http://${ip}:8545
      sleep 5
      local rawtx=$(${CLI} coins send_exec -e ticket -a 100000)
      echo "send_exec 2 ticket rawTx:${rawtx}"
      #用代理交易处理
      local ethrawtx=$(${CLI} coins transfer_eth -d ${rawtx} -t "0x0000000000000000000000000000000000200005" -f ${m4normalAddr} --rpc_laddr ${ECli} )
      echo "send_exec 2 ticket ethrawtx:${ethrawtx}"
      #签名
      local signedRawTx=$(${CLI} wallet sign -d ${ethrawtx} -p 2 -c 1999 -k ${m4addrkey} --rpc_laddr ${ECli} )
      echo "send_exec 2 ticket signedrawTx:${signedRawTx}"
      local hash=$(${CLI}  wallet send -d ${signedRawTx} -e --rpc_laddr ${ECli} )
      if [ -z "${hash}" ]; then
          exit 1
      fi
      chain33_QueryTx ${hash} ${MCli}
      #检查 m4normalAddr ticket余额
      result=$(${CLI} account balance -a "${m4normalAddr}" -e ticket)
      balance_ret "${result}" "100000.0000"
      #等待挖矿
      #获取tiket票数

      while true; do
        ticketNum=$(${CLI} ticket count)
        echo "ticketNum:${ticketNum} time:$(date +"%s") "
        if [ ${ticketNum} -gt 0 ]; then
          break
        fi
        sleep 1
        done
}

# 判断 chain33 金额是否正确
function balance_ret() {

    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_cp_file
    fi

    local balance=$(echo "${1}" | jq -r ".balance")
    if [ "$(echo "$balance < $2" | bc)" -eq 1 ] || [ "$(echo "$balance > $2" | bc)" -eq 1 ]; then
        echo -e "${RED}The balance is not correct${NOC}"
        exit_cp_file
    fi


    echo "${balance}"
}





function main() {
    echo "====================DAPP=${DAPP} main begin==================="


    ### start docker
    echo "#### start docker"
    start_docker
    i=0

    for con in "${containers[@]}"; do
      t=2
      #初始化各个容器  参数1:容器名称，参数2: 挖矿地址的私钥
      if [ $i == 2 ] ; then
        t=0
      fi
      init_nodes ${con} ${minerkeys[$i]} ${t}
      i=$(($i+1))
      echo "i:${i}"
     done


    run_testcase
    check_docker_container
    #finish
    docker-compose down
    echo "===============DAPP=$DAPP main end==============="
}

# start
main

