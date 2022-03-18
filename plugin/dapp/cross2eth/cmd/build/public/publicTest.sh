#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
# shellcheck disable=SC2155
set -x
set -e

# 公用测试函数

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# 出错退出前拷贝日志文件
function exit_cp_file() {
    exit 1

    set -x
    # shellcheck disable=SC2116
    dirNameFa=$(echo ~)
    dirName="${dirNameFa}/x2ethereumlogs"

    if [ ! -d "${dirName}" ]; then
        # shellcheck disable=SC2086
        mkdir -p ${dirName}
    fi

    for name in a b c d; do
        # shellcheck disable=SC2154
        docker cp "${dockerNamePrefix}_ebrelayer${name}_1":/root/logs/x2Ethereum_relayer.log "${dirName}/ebrelayer${name}.log"
        docker exec "${dockerNamePrefix}_ebrelayer${name}_1" tail -n 1000 /root/logs/x2Ethereum_relayer.log
    done

    docker cp "${dockerNamePrefix}_chain33_1":/root/logs/chain33.log "${dirName}/chain33.log"
    docker logs "${dockerNamePrefix}_chain33_1" | tail -n 1000

    exit 1
}

# 判断结果是否正确
function cli_ret() {
    set +x
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_cp_file
    fi

    ok=$(echo "${1}" | jq -r .isOK)
    if [[ ${ok} != "true" ]]; then
        echo -e "${RED}failed to ${2}${NOC}"
        exit_cp_file
    fi

    local jqMsg=".msg"
    if [[ $# -ge 3 ]]; then
        jqMsg="${3}"
    fi

    msg=$(echo "${1}" | jq -r "${jqMsg}")
    if [[ $# -eq 4 ]]; then
        if [ "$(echo "$msg < $4" | bc)" -eq 1 ] || [ "$(echo "$msg > $4" | bc)" -eq 1 ]; then
            echo -e "${RED}The balance is not correct${NOC}"
            exit_cp_file
        fi
    fi

    set -x
    echo "${msg}"
}

# 判断 chain33 金额是否正确
function balance_ret() {
    set +x
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_cp_file
    fi

    local balance=$(echo "${1}" | jq -r ".balance")
    if [ "$(echo "$balance < $2" | bc)" -eq 1 ] || [ "$(echo "$balance > $2" | bc)" -eq 1 ]; then
        echo -e "${RED}The balance is not correct${NOC}"
        exit_cp_file
    fi

    set -x
    echo "${balance}"
}

# 后台启动 ebrelayer 进程 $1 docker 名称 $2进程名称 $3进程信息输出重定向文件
function start_docker_ebrelayer() {
    # 参数如果小于 3 直接报错
    if [[ $# -lt 3 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_cp_file
    fi

    # 后台启动程序
    docker exec "$1" nohup "${2}" >"${3}" 2>&1 &
    sleep 10

    # shellcheck disable=SC2009
    pid=$(docker exec "$1" ps -ef | grep "$2" | grep -v 'grep' | awk '{print $2}' | xargs)
    local count=0
    while [ "${pid}" == "" ]; do
        docker exec "$1" nohup "${2}" >"${3}" 2>&1 &
        sleep 10

        count=$((count + 1))
        if [[ ${count} -ge 20 ]]; then
            echo -e "${RED}start ${1} ${2} failed${NOC}"
            exit_cp_file
        fi

        # shellcheck disable=SC2009
        pid=$(docker exec "$1" ps -ef | grep "$2" | grep -v 'grep' | awk '{print $2}' | xargs)
    done
}

# 杀死进程 ebrelayer 进程 docker ebrelayer 名称
function kill_docker_ebrelayer() {
    # shellcheck disable=SC2009
    pid=$(docker exec "$1" ps -ef | grep "ebrelayer" | grep -v 'grep' | awk '{print $2}' | xargs)
    if [ "${pid}" == "" ]; then
        echo "not find ${1} pid"
        return
    fi

    docker exec "$1" kill "${pid}"
    # shellcheck disable=SC2009
    pid=$(docker exec "$1" ps -ef | grep "ebrelayer" | grep -v 'grep' | awk '{print $2}' | xargs)
    if [ "${pid}" != "" ]; then
        echo "kill ${1} failed"
        docker exec "$1" kill -9 "${pid}"
    fi
    sleep 1
}

# chain33 区块等待 $1:cli 路径  $2:等待高度
function block_wait() {
    set +x
    local CLI=${1}

    if [[ $# -lt 1 ]]; then
        echo -e "${RED}wrong block_wait parameter${NOC}"
        exit_cp_file
    fi

    local cur_height=$(${CLI} block last_header | jq ".height")
    local expect=$((cur_height + ${2}))
    local count=0
    while true; do
        new_height=$(${CLI} block last_header | jq ".height")
        if [[ ${new_height} -ge ${expect} ]]; then
            break
        fi

        count=$((count + 1))
        sleep 1
    done

    count=$((count + 1))
    set -x
    echo -e "${GRE}chain33 wait new block $count s, cur height=$expect,old=$cur_height${NOC}"
}

# 检查交易是否执行成功 $1:cli 路径  $2:交易hash
function check_tx() {
    set +x
    local CLI=${1}

    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong check_tx parameters${NOC}"
        exit_cp_file
    fi

    if [[ ${2} == "" ]]; then
        echo -e "${RED}wrong check_tx txHash is empty${NOC}"
        exit_cp_file
    fi

    local count=0
    while true; do
        ty=$(${CLI} tx query -s "${2}" | jq .receipt.ty)
        if [[ ${ty} != "" ]]; then
            break
        fi

        count=$((count + 1))
        sleep 1

        if [[ ${count} -ge 100 ]]; then
            echo "chain33 query tx for too long"
            break
        fi
    done

    set -x

    ty=$(${CLI} tx query -s "${2}" | jq .receipt.ty)
    if [[ ${ty} != 2 ]]; then
        echo -e "${RED}check tx error, hash is ${2}${NOC}"
        exit_cp_file
    fi
}

# 检查地址是否匹配 $1返回结果 $2匹配地址
function check_addr() {
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong check number parameters${NOC}"
        exit_cp_file
    fi

    addr=$(echo "${1}" | jq -r ".acc.addr")
    if [[ ${addr} != "${2}" ]]; then
        echo -e "${RED}error addr, expect ${1}, get ${2}${NOC}"
        exit_cp_file
    fi
}

# $1 dockerName
function get_docker_addr() {
    local dockerAddr=$(docker inspect "${1}" | jq ".[].NetworkSettings.Networks" | grep "IPAddress" | awk '{ print $2}' | sed 's/\"//g' | sed 's/,//g')
    echo "${dockerAddr}"
}

# 判断结果 $1 和 $2 是否相等
function is_equal() {
    set +x
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_cp_file
    fi

    if [[ $1 != "$2" ]]; then
        echo -e "${RED}$1 != ${2}${NOC}"
        exit_cp_file
    fi

    set -x
}

# 获取本机 IP
function get_inet_addr() {
    inetAddr=$(ifconfig wlp2s0 | grep "inet " | awk '{ print $2}' | awk -F: '{print $2}')
    if [[ ${inetAddr} == "" ]]; then
        inetAddr=$(ifconfig wlp2s0 | grep "inet " | awk '{ print $2}')
        if [[ ${inetAddr} == "" ]]; then
            inetAddr=$(ifconfig eth0 | grep "inet " | awk '{ print $2}' | awk -F: '{print $2}')
            if [[ ${inetAddr} == "" ]]; then
                inetAddr=$(ifconfig eth0 | grep "inet " | awk '{ print $2}')
                if [[ ${inetAddr} == "" ]]; then
                    ip addr show eth0
                    inetAddr=$(ip addr show eth0 | grep "inet " | awk '{ print $2}' | head -c-4)
                fi
            fi
        fi
    fi

    echo "${inetAddr}"
}

# 启动 eth
function start_trufflesuite() {
    # 如果原来存在先删除
    local ganacheName=ganachetest
    local isExit=$(docker inspect ${ganacheName} | jq ".[]" | jq ".Id")
    if [[ ${isExit} != "" ]]; then
        docker stop ${ganacheName}
        docker rm ${ganacheName}
    fi

    # 启动 eth
    docker run -d -e 10000 --name ${ganacheName} -p 7545:8545 -l eth_test trufflesuite/ganache-cli:latest -a 20 --debug -b 1 -m "coast bar giraffe art venue decide symbol law visual crater vital fold" -l 0x7a1200
    sleep 1
}

# eth 区块等待 $1:等待高度  $2:url地址，默认为 http://localhost:7545,测试网络用 https://ropsten-rpc.linkpool.io/
function eth_block_wait() {
    set +x
    if [[ $# -lt 0 ]]; then
        echo -e "${RED}wrong block_wait parameter${NOC}"
        exit_cp_file
    fi

    local cur_height=""
    local new_height=""
    local url=${2}
    if [ "${url}" == "" ]; then
        cur_height=$(curl -ksd '{"id":1,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}' http://localhost:7545 | jq -r ".result")
    else
        cur_height=$(curl -H "Content-Type: application/json" -X POST --data '{"id":1,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}' "${url}" | jq -r ".result")
    fi

    local expect=$((cur_height + ${1} + 1))
    local count=0
    while true; do
        if [ "${url}" == "" ]; then
            new_height=$(curl -ksd '{"id":1,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}' http://localhost:7545 | jq -r ".result")
        else
            new_height=$(curl -H "Content-Type: application/json" -X POST --data '{"id":1,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}' "${url}" | jq -r ".result")
        fi

        if [[ ${new_height} -ge ${expect} ]]; then
            break
        fi

        count=$((count + 1))
        sleep 1
    done

    count=$((count + 1))
    sleep 1
    set -x
    echo -e "${GRE}eth wait new block $count s, cur height=$expect,old=$((cur_height))${NOC}"
}
