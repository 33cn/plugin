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
    # shellcheck disable=SC2116
    dirNameFa=$(echo ~)
    dirName="$dirNameFa/x2ethereumlogs"

    if [ ! -d "${dirName}" ]; then
        # shellcheck disable=SC2086
        mkdir -p ${dirName}
    fi

    for name in a b c d; do
        # shellcheck disable=SC2154
        docker cp "${dockerNamePrefix}_ebrelayer${name}_1":/root/logs/x2Ethereum_relayer.log "$dirName/ebrelayer$name.log"
    done
    docker cp "${dockerNamePrefix}_chain33_1":/root/logs/chain33.log "$dirName/chain33.log"

    exit 1
}

function copyErrLogs() {
    if [ -n "$CASE_ERR" ]; then
        # /var/lib/jenkins
        # shellcheck disable=SC2116
        dirNameFa=$(echo ~)
        dirName="$dirNameFa/x2ethereumlogs"

        if [ ! -d "${dirName}" ]; then
            # shellcheck disable=SC2086
            mkdir -p ${dirName}
        fi

        for name in a b c d; do
            # shellcheck disable=SC2154
            docker cp "${dockerNamePrefix}_ebrelayer${name}_rpc_1":/root/logs/x2Ethereum_relayer.log "$dirName/ebrelayer$name_rpc.log"
        done
        docker cp "${dockerNamePrefix}_chain33_1":/root/logs/chain33.log "$dirName/chain33_rpc.log"
    fi
}

function kill_all_ebrelayer() {
    for name in A B C D; do
        local ebrelayer="./../build/$name/ebrelayer"
        kill_ebrelayer "${ebrelayer}"
    done
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

# 判断结果是否错误
function cli_ret_err() {
    #set +x
    ok=$(echo "${1}" | jq -r .isOK)
    echo "${ok}"
    if [[ ${ok} == "true" ]]; then
        echo -e "${RED}isOK is true${NOC}"
        exit_cp_file
    fi
    #set -x
}

# 查询关键字所在行然后删除 ${1}文件名称 ${2}关键字
function delete_line() {
    line=$(cat -n "${1}" | grep "${2}" | awk '{print $1}')
    if [[ ${line} != "" ]]; then
        sed -i "${line}"'d' "${1}" # 删除行
    fi
}

# 查询关键字所在行然后删除 ${1}文件名称 ${2}关键字
function delete_line_show() {
    local line=$(cat -n "${1}" | grep "${2}" | awk '{print $1}')
    if [[ ${line} != "" ]]; then
        sed -i "${line}"'d' "${1}" # 删除行
        line=$((line - 1))
    fi
    echo "${line}"
}

# 后台启动 ebrelayer 进程 $1进程名称 $2进程信息输出重定向文件
function start_ebrelayer() {
    # 参数如果小于 2 直接报错
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_cp_file
    fi

    # 判断可执行文件是否存在
    if [ ! -x "${1}" ]; then
        echo -e "${RED}${1} not exist${NOC}"
        exit_cp_file
    fi

    # 后台启动程序
    nohup "${1}" >"${2}" 2>&1 &
    sleep 2

    # shellcheck disable=SC2009
    pid=$(ps -ef | grep "${1}" | grep -v 'grep' | awk '{print $2}')
    local count=0
    while [ "${pid}" == "" ]; do
        nohup "${1}" >"${2}" 2>&1 &
        sleep 2

        count=$((count + 1))
        if [[ ${count} -ge 20 ]]; then
            echo -e "${RED}start ${1} failed${NOC}"
            exit_cp_file
        fi

        # shellcheck disable=SC2009
        pid=$(ps -ef | grep "${1}" | grep -v 'grep' | awk '{print $2}')
    done
}

# 后台启动 ebrelayer 进程 $1 A B C D
function start_ebrelayer_and_unlock() {
    start_ebrelayer "./${1}/ebrelayer" "./${1}/ebrelayer.log"
    sleep 2

    local CLI="./ebcli_$1"
    local count=0
    while true; do
        result=$(${CLI} relayer unlock -p 123456hzj | jq -r .isOK)
        if [[ ${result} == "true" ]]; then
            break
        fi

        count=$((count + 1))
        if [[ ${count} == 5 ]]; then
            echo -e "${RED}failed to unlock${NOC}"
            exit_cp_file
        fi

        sleep 1
    done
}

# 后台启动 ebrelayer 进程 $1 A B C D
function start_ebrelayer_and_setpwd_unlock() {
    start_ebrelayer "./${1}/ebrelayer" "./${1}/ebrelayer.log"
    sleep 2

    local CLI="./ebcli_$1"
    local count=0
    while true; do
        result=$(${CLI} relayer set_pwd -p 123456hzj | jq -r .isOK)
        if [[ ${result} == "true" ]]; then
            break
        fi

        count=$((count + 1))
        if [[ ${count} == 5 ]]; then
            echo -e "${RED}failed to set_pwd${NOC}"
            exit_cp_file
        fi

        sleep 1
    done

    count=0
    while true; do
        result=$(${CLI} relayer unlock -p 123456hzj | jq -r .isOK)
        if [[ ${result} == "true" ]]; then
            break
        fi

        count=$((count + 1))
        if [[ ${count} == 5 ]]; then
            echo -e "${RED}failed to unlock${NOC}"
            exit_cp_file
        fi

        sleep 1
    done
}

# 杀死进程ebrelayer 进程 $1进程名称
function kill_ebrelayer() {
    # shellcheck disable=SC2009
    pid=$(ps -ef | grep "${1}" | grep -v 'grep' | awk '{print $2}')
    if [ "${pid}" == "" ]; then
        echo "not find ${1} pid"
        return
    fi

    kill "${pid}"
    # shellcheck disable=SC2009
    pid=$(ps -ef | grep "${1}" | grep -v 'grep' | awk '{print $2}')
    if [ "${pid}" != "" ]; then
        echo "kill ${1} failed"
        kill -9 "${pid}"
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

function check_number() {
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong check number parameters${NOC}"
        exit_cp_file
    fi

    if [ "$(echo "$1 < $2" | bc)" -eq 1 ] || [ "$(echo "$1 > $2" | bc)" -eq 1 ]; then
        echo -e "${RED}error number, expect ${1}, get ${2}${NOC}"
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

function get_inet_addr() {
    local inetAddr=$(ifconfig wlp2s0 | grep "inet " | awk '{ print $2}' | awk -F: '{print $2}')
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

# $1 dockerName
function get_docker_addr() {
    local dockerAddr=$(docker inspect "${1}" | jq ".[].NetworkSettings.Networks" | grep "IPAddress" | awk '{ print $2}' | sed 's/\"//g' | sed 's/,//g')
    echo "${dockerAddr}"
}

# 更新配置文件 $1 为 BridgeRegistry 合约地址; $2 等待区块 默认10; $3 relayer.toml 地址
function updata_relayer_toml() {
    local BridgeRegistry=${1}
    local maturityDegree=${2}
    local file=${3}

    local chain33Host=$(get_docker_addr "${dockerNamePrefix}_chain33_1")
    if [[ ${chain33Host} == "" ]]; then
        echo -e "${RED}chain33Host is empty${NOC}"
        exit_cp_file
    fi

    local pushHost=$(get_inet_addr)
    if [[ ${pushHost} == "" ]]; then
        echo -e "${RED}pushHost is empty${NOC}"
        exit_cp_file
    fi

    local line=$(delete_line_show "${file}" "chain33Host")
    # 在第 line 行后面 新增合约地址
    sed -i ''"${line}"' a chain33Host="http://'"${chain33Host}"':8801"' "${file}"

    line=$(delete_line_show "${file}" "pushHost")
    sed -i ''"${line}"' a pushHost="http://'"${pushHost}"':20000"' "${file}"

#    line=$(delete_line_show "${file}" "pushBind")
#    sed -i ''"${line}"' a pushBind="'"${pushHost}"':20000"' "${file}"

    line=$(delete_line_show "${file}" "BridgeRegistry")
    sed -i ''"${line}"' a BridgeRegistry="'"${BridgeRegistry}"'"' "${file}"

    sed -i 's/EthMaturityDegree=10/'EthMaturityDegree="${maturityDegree}"'/g' "${file}"
    sed -i 's/maturityDegree=10/'maturityDegree="${maturityDegree}"'/g' "${file}"

    sed -i 's/^EthBlockFetchPeriod=.*/EthBlockFetchPeriod=500/g' "${file}"
    sed -i 's/^fetchHeightPeriodMs=.*/fetchHeightPeriodMs=500/g' "${file}"
}

# 更新配置文件 $1 为 BridgeRegistry 合约地址; $2 等待区块 默认10; $3 relayer.toml 地址
function updata_relayer_toml_ropston() {
    local BridgeRegistry=${1}
    local maturityDegree=${2}
    local file=${3}

    local chain33Host=127.0.0.1
    local pushHost=127.0.0.1

    local line=$(delete_line_show "${file}" "chain33Host")
    # 在第 line 行后面 新增合约地址
    sed -i ''"${line}"' a chain33Host="http://'${chain33Host}':8801"' "${file}"

    line=$(delete_line_show "${file}" "pushHost")
    sed -i ''"${line}"' a pushHost="http://'${pushHost}':20000"' "${file}"

    line=$(delete_line_show "${file}" "BridgeRegistry")
    sed -i ''"${line}"' a BridgeRegistry="'"${BridgeRegistry}"'"' "${file}"

    sed -i 's/EthMaturityDegree=10/'EthMaturityDegree="${maturityDegree}"'/g' "${file}"
    sed -i 's/maturityDegree=10/'maturityDegree="${maturityDegree}"'/g' "${file}"
}

# $1 portRelayer
function updata_docker_relayer_toml() {
    local port=$1
    local portRelayer=$1

    if [ "${portRelayer}" != "20000" ]; then
      sed -i 's/20000/'"${portRelayer}"'/g' "./relayer.toml"
      sed -i 's/20000/'"${portRelayer}"'/g' "./Dockerfile-x2ethrelay"
    fi

    for name in b c d; do
        local file="./relayer$name.toml"
        cp './relayer.toml' "${file}"

        # 删除配置文件中不需要的字段
        for deleteName in "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers" "deployerPrivateKey" "deploy"; do
           delete_line "${file}" "${deleteName}"
        done

        port=$((port + 1))
        sed -i 's/'"${portRelayer}"'/'${port}'/g' "${file}"

        sed -i 's/x2ethereum/x2ethereum'${name}'/g' "${file}"

        local dockerfile="./Dockerfile-x2ethrelay$name"
        cp "./Dockerfile-x2ethrelay" "${dockerfile}"

        # shellcheck disable=SC2155
        local line=$(delete_line_show "${dockerfile}" "COPY relayer.toml relayer.toml")
        sed -i ''"${line}"' a COPY relayer'$name'.toml relayer.toml' "${dockerfile}"

        sed -i 's/'"${portRelayer}"'/'"${port}"'/g' "${dockerfile}"

        local dockeryml="./docker-compose-ebrelayer$name.yml"
        cp "./docker-compose-ebrelayer.yml" "${dockeryml}"

        line=$(delete_line_show "${dockeryml}" "ebrelayera")
        sed -i ''"${line}"' a \ \ ebrelayer'$name':' "${dockeryml}"

        line=$(delete_line_show "${dockeryml}" "dockerfile: Dockerfile-x2ethrelay")
        sed -i ''"${line}"' a \ \ \ \ \ \ dockerfile: Dockerfile-x2ethrelay'$name'' "${dockeryml}"

        sed -i 's/'"${portRelayer}"'/'${port}'/g' "${dockeryml}"
    done
}

# 更新配置文件 $1 为 BridgeRegistry 合约地址; $2 等待区块 默认10; $3 MAIN_HTTP; $4 relayer.toml 地址
function updata_relayer_toml_rpc() {
    local BridgeRegistry=${1}
    local maturityDegree=${2}
    local MAIN_HTTP=${3}
    local file=${4}

    # shellcheck disable=SC2155
    local pushHost=$(get_inet_addr)
    if [[ ${pushHost} == "" ]]; then
        echo -e "${RED}pushHost is empty${NOC}"
        copyErrLogs
    fi

    # shellcheck disable=SC2155
    local line=$(delete_line_show "${file}" "chain33Host")
    # 在第 line 行后面 新增合约地址
    sed -i ''"${line}"' a chain33Host="'"${MAIN_HTTP}"'"' "${file}"

    line=$(delete_line_show "${file}" "pushHost")
    sed -i ''"${line}"' a pushHost="http://'"${pushHost}"':20000"' "${file}"

    line=$(delete_line_show "${file}" "BridgeRegistry")
    sed -i ''"${line}"' a BridgeRegistry="'"${BridgeRegistry}"'"' "${file}"

    sed -i 's/EthMaturityDegree=10/'EthMaturityDegree="${maturityDegree}"'/g' "${file}"
    sed -i 's/maturityDegree=10/'maturityDegree="${maturityDegree}"'/g' "${file}"

    sed -i 's/^EthBlockFetchPeriod=.*/EthBlockFetchPeriod=500/g' "${file}"
    sed -i 's/^fetchHeightPeriodMs=.*/fetchHeightPeriodMs=500/g' "${file}"
}

# $1 portRelayer
function updata_docker_relayer_toml_rpc() {
    local portRelayer=$1
    local port=$1
    sed -i 's/20000/'"${portRelayer}"'/g' "./x2ethereum/relayer.toml"
    sed -i 's/20000/'"${portRelayer}"'/g' "./x2ethereum/Dockerfile-x2ethrelay"

    for name in b c d; do
        local file="./x2ethereum/relayer$name.toml"
        cp './x2ethereum/relayer.toml' "${file}"

        # 删除配置文件中不需要的字段
        for deleteName in "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers" "deployerPrivateKey" "deploy"; do
           delete_line "${file}" "${deleteName}"
        done

        port=$((port - 1))
        sed -i 's/'"${portRelayer}"'/'${port}'/g' "${file}"

        sed -i 's/x2ethereum/x2ethereum'${name}'/g' "${file}"

        local dockerfile="./x2ethereum/Dockerfile-x2ethrelay$name"
        cp "./x2ethereum/Dockerfile-x2ethrelay" "${dockerfile}"

        # shellcheck disable=SC2155
        local line=$(delete_line_show "${dockerfile}" "COPY relayer.toml relayer.toml")
        sed -i ''"${line}"' a COPY relayer'$name'.toml relayer.toml' "${dockerfile}"
        sed -i 's/'"${portRelayer}"'/'"${port}"'/g' "${dockerfile}"

        local dockeryml="./x2ethereum/docker-compose-ebrelayer$name.yml"
        cp "./x2ethereum/docker-compose-ebrelayer.yml" "${dockeryml}"

        sed -i 's/ebrelayera/ebrelayer'${name}'/g' "${dockeryml}"
        sed -i 's/Dockerfile-x2ethrelay/Dockerfile-x2ethrelay'${name}'/g' "${dockeryml}"
        sed -i 's/'"${portRelayer}"'/'${port}'/g' "${dockeryml}"
    done
}

# 更新 B C D 的配置文件
function updata_all_relayer_toml() {
    local port=9901
    local port2=20000
    #    local dockername=30

    for name in B C D; do
        local file="./$name/relayer.toml"
        cp './A/relayer.toml' "${file}"
        cp './ebrelayer' "./$name/ebrelayer"

        # 删除配置文件中不需要的字段
        for deleteName in "deployerPrivateKey" "operatorAddr" "validatorsAddr" "initPowers" "deployerPrivateKey" "deploy"; do
            delete_line "${file}" "${deleteName}"
        done

        # 替换端口
        port=$((port + 1))
        sed -i 's/localhost:9901/localhost:'${port}'/g' "${file}"

        port2=$((port2 + 1))
        sed -i 's/20000/'${port2}'/g' "${file}"

        sed -i 's/x2ethereum/x2ethereum'${name}'/g' "${file}"
    done
}

# 启动 eth
function start_trufflesuite() {
    # 如果原来存在先删除
    local ganacheName=ganachetest
    #local isExit=$(docker inspect ${ganacheName} | jq ".[]" | jq ".Id")
    #if [[ ${isExit} != "" ]]; then
    docker stop ${ganacheName}
    docker rm ${ganacheName}
    #fi

    # 启动 eth
    docker run -d --name ${ganacheName} -p 7545:8545 -l eth_test trufflesuite/ganache-cli:latest -a 10 --debug -b 2 -m "coast bar giraffe art venue decide symbol law visual crater vital fold"
    sleep 1
}

# $1 CLI
function wait_prophecy_finish() {
    local CLI=${1}
    set +x
    local count=0
    while true; do
        if [[ $# -eq 4 ]]; then
            ${CLI} relayer ethereum balance -o "${2}" -t "${3}"
            balance=$(${CLI} relayer ethereum balance -o "${2}" -t "${3}" | jq -r .balance)
            if [[ ${balance} == "${4}" ]]; then
                break
            fi
        fi
        if [[ $# -eq 3 ]]; then
            ${CLI} relayer ethereum balance -o "${2}"
            balance=$(${CLI} relayer ethereum balance -o "${2}" | jq -r .balance)
            if [[ ${balance} == "${3}" ]]; then
                break
            fi
        fi
        count=$((count + 1))
        if [[ ${count} == 30 ]]; then
            set -x
            echo -e "${RED}failed to get balance${NOC}"
            exit_cp_file
        fi

        sleep 1
    done
    set -x
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
