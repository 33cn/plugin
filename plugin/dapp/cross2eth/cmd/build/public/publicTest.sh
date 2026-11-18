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

# 轮询间隔(秒), 只在等待期间生效
POLL_INTERVAL="${POLL_INTERVAL:-0.3}"

# 判断两个值是否相等, 数值类型兼容 "16" 与 "16.0000" 的写法
function is_expect_value() {
    local got="${1}"
    local expect="${2}"
    if [ "${got}" == "${expect}" ]; then
        return 0
    fi
    if [[ "${got}" =~ ^-?[0-9]+(\.[0-9]+)?$ ]] && [[ "${expect}" =~ ^-?[0-9]+(\.[0-9]+)?$ ]]; then
        if [ "$(echo "${got} == ${expect}" | bc)" -eq 1 ]; then
            return 0
        fi
    fi
    return 1
}

# 有界轮询: 命令输出(最后一行)等于期望值时立即返回, 不再固定 sleep。
# 跨链/出块正常情况下 1~3s 内完成, 固定等待只会白白拖慢 CI;
# 超时后仍返回最后一次结果, 由调用方的 cli_ret/is_equal 报错并收集日志。
# $1 期望值 $2 超时秒 $3 描述 $4... 命令
function wait_output() {
    local expect="${1}"
    local timeout="${2}"
    local desc="${3}"
    shift 3
    # 轮询期间关掉 xtrace, 否则每次重试都会刷日志(所有调用点都在 $(...) 子 shell 里)
    set +x
    local deadline=$((SECONDS + timeout))
    local got=""
    while true; do
        got=$("$@" 2>/dev/null | tail -n 1 | tr -d '\r' || true)
        if is_expect_value "${got}" "${expect}"; then
            printf '%s' "${got}"
            return 0
        fi
        if [ "${SECONDS}" -ge "${deadline}" ]; then
            echo -e "${RED}wait_output timeout(${timeout}s) ${desc}: expect=${expect}, last=${got}${NOC}" >&2
            printf '%s' "${got}"
            return 0
        fi
        sleep "${POLL_INTERVAL}"
    done
}

# 有界轮询: 命令返回的 jq 字段等于期望值时立即返回, 输出最后一次完整结果供 cli_ret 复用。
# $1 jq 表达式 $2 期望值 $3 超时秒 $4 描述 $5... 命令
function wait_cli() {
    local expr="${1}"
    local expect="${2}"
    local timeout="${3}"
    local desc="${4}"
    shift 4
    # 轮询期间关掉 xtrace, 否则每次重试都会刷日志(所有调用点都在 $(...) 子 shell 里)
    set +x
    local deadline=$((SECONDS + timeout))
    local out=""
    local got=""
    while true; do
        out=$("$@" 2>/dev/null | tr -d '\r' || true)
        got=$(echo "${out}" | jq -r "${expr}" 2>/dev/null | tail -n 1 || true)
        if is_expect_value "${got}" "${expect}"; then
            printf '%s' "${out}"
            return 0
        fi
        if [ "${SECONDS}" -ge "${deadline}" ]; then
            echo -e "${RED}wait_cli timeout(${timeout}s) ${desc}: expect ${expr}=${expect}, last=${got}${NOC}" >&2
            printf '%s' "${out}"
            return 0
        fi
        sleep "${POLL_INTERVAL}"
    done
}

# 有界轮询: 命令返回 isOK=true 时立即返回(用于中继器 RPC 尚未就绪的场景)。
# $1 超时秒 $2 描述 $3... 命令
function wait_cli_ok() {
    local timeout="${1}"
    local desc="${2}"
    shift 2
    # 轮询期间关掉 xtrace, 否则每次重试都会刷日志(所有调用点都在 $(...) 子 shell 里)
    set +x
    local deadline=$((SECONDS + timeout))
    local out=""
    local ok=""
    while true; do
        out=$("$@" 2>/dev/null | tr -d '\r' || true)
        ok=$(echo "${out}" | jq -r '.isOK' 2>/dev/null || true)
        if [ "${ok}" == "true" ]; then
            printf '%s' "${out}"
            return 0
        fi
        if [ "${SECONDS}" -ge "${deadline}" ]; then
            echo -e "${RED}wait_cli_ok timeout(${timeout}s) ${desc}, last=${ok}${NOC}" >&2
            printf '%s' "${out}"
            return 0
        fi
        sleep "${POLL_INTERVAL}"
    done
}

# 出错退出前拷贝日志文件
function exit_cp_file() {
    set -x
    # shellcheck disable=SC2116
    dirNameFa=$(echo ~)
    dirName="${dirNameFa}/x2ethereumlogs"

    if [ ! -d "${dirName}" ]; then
        # shellcheck disable=SC2086
        mkdir -p ${dirName}
    fi

    # 单个容器取不到日志不能中断收集, 否则后面的日志会丢
    for name in a b c d; do
        # shellcheck disable=SC2154
        docker cp "${dockerNamePrefix}-ebrelayer${name}-1":/root/logs/x2Ethereum_relayer.log "${dirName}/ebrelayer${name}.log" || true
        docker exec "${dockerNamePrefix}-ebrelayer${name}-1" tail -n 1000 /root/logs/x2Ethereum_relayer.log || true
    done

    docker cp "${dockerNamePrefix}-chain33-1":/root/logs/chain33.log "${dirName}/chain33.log" || true
    docker logs "${dockerNamePrefix}-chain33-1" | tail -n 1000 || true

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

    # 进程通常 1s 内就能起来, 原来固定 sleep 10s 再查进程会白等;
    # 这里 0.5s 轮询一次, 连续查不到(约 5s)才重拉, 最多等 60s 后报错。
    local pid="" count=0
    while true; do
        sleep 0.5
        # shellcheck disable=SC2009
        pid=$(docker exec "$1" ps -ef | grep "$2" | grep -v 'grep' | awk '{print $2}' | xargs || true)
        if [ "${pid}" != "" ]; then
            # 进程已起, 短暂等待其完成 RPC/配置初始化
            sleep 1
            return 0
        fi

        count=$((count + 1))
        if [[ ${count} -ge 120 ]]; then
            echo -e "${RED}start ${1} ${2} failed${NOC}"
            exit_cp_file
        fi

        # 进程可能启动即退出(端口/配置问题), 每 5s 重拉一次, 与原来的重试语义一致
        if [ $((count % 10)) -eq 0 ]; then
            docker exec "$1" nohup "${2}" >"${3}" 2>&1 &
        fi
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
        sleep 0.2
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
        ty=$(${CLI} tx query -s "${2}" | jq .receipt.ty || true)
        if [[ ${ty} != "" ]]; then
            break
        fi

        count=$((count + 1))
        # 交易写入索引通常几百毫秒, 原来 1s 轮询平均白等约 0.5s; 0.2s 轮询, 次数放大保持同样的超时预算
        sleep 0.2

        if [[ ${count} -ge 500 ]]; then
            echo "chain33 query tx for too long"
            break
        fi
    done

    set -x

    ty=$(${CLI} tx query -s "${2}" | jq .receipt.ty || true)
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
    # 原来无上限轮询, eth url 不通时会一直卡到 job 超时; 这里加一个上限(默认 300 轮)
    local maxCount="${3:-300}"
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
        if [ "${count}" -ge "${maxCount}" ]; then
            echo -e "${RED}eth_block_wait timeout, expect height ${expect}, last ${new_height}${NOC}"
            break
        fi
        sleep 1
    done

    count=$((count + 1))
    sleep 1
    set -x
    echo -e "${GRE}eth wait new block $count s, cur height=$expect,old=$((cur_height))${NOC}"
}
