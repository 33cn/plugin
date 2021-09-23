#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
# shellcheck disable=SC2155
set -x
set -e

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# 出错退出前拷贝日志文件
function exit_cp_file() {
    exit 1
}

# $1 dockerName
function get_docker_addr() {
    local dockerAddr=$(docker inspect "${1}" | jq ".[].NetworkSettings.Networks" | grep "IPAddress" | awk '{ print $2}' | sed 's/\"//g' | sed 's/,//g')
    echo "${dockerAddr}"
}

#function block_wait() {
#    set +x
#    set +x
#    local block=$1
#    for((i=1;i<=block;i++));do
#        hash=$(${Chain33Cli} send coins transfer -a 0.001 -n test -t "${propAddr}" -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
#        check_tx "${Chain33Cli}" "${hash}"
#        echo "这是第 $i 次调用";
#    done;
#    set -x
#    set -x
#}

# 杀死进程ebrelayer 进程 $1进程名称
function kill_ebrelayer() {
    # shellcheck disable=SC2009
    ps -ef | grep "${1}"
    # shellcheck disable=SC2009
    pid=$(ps -ef | grep "${1}" | grep -v 'grep' | awk '{print $2}' | xargs)
    if [ "${pid}" == "" ]; then
        echo "not find ${1} pid"
        return
    fi

    kill -9 "${pid}"
    sleep 1
    # shellcheck disable=SC2009
    pid=$(ps -ef | grep "${1}" | grep -v 'grep' | awk '{print $2}' | xargs)
    if [ "${pid}" != "" ]; then
        echo "kill ${1} failed"
        kill -9 "${pid}"
    fi
    sleep 1
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

# 查询关键字所在行然后删除 ${1}文件名称 ${2}关键字
function delete_line() {
    line=$(cat -n "${1}" | grep "${2}" | awk '{print $1}' | xargs | awk '{print $1}')
    if [ "${line}" ]; then
        sed -i "${line}"'d' "${1}" # 删除行
    fi
}

# 查询关键字所在行然后删除 ${1}文件名称 ${2}关键字
function delete_line_show() {
    local line=$(cat -n "${1}" | grep "${2}" | awk '{print $1}' | xargs | awk '{print $1}')
    if [ "${line}" ]; then
        sed -i "${line}"'d' "${1}" # 删除行
        line=$((line - 1))
    fi
    echo "${line}"
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

# 判断结果 $1 和 $2 是否不相等
function is_not_equal() {
    set +x
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_cp_file
    fi

    if [[ $1 == "$2" ]]; then
        echo -e "${RED}$1 == ${2}${NOC}"
        exit_cp_file
    fi

    set -x
}

# import_key and transfer $1 key, $2 label, $3 addr, $4 transfer amount
function import_addr() {
    local key="$1"
    local label="$2"
    local addr="$3"
    # shellcheck disable=SC2154
    result=$(${Chain33Cli} account import_key -k "${key}" -l "${label}")
    check_addr "${result}" "${addr}"

    if [ "$#" -eq 4 ]; then
        # shellcheck disable=SC2154
        hash=$(${Chain33Cli} send coins transfer -a "$4" -n test -t "${addr}" -k "${minerAddr}")
        check_tx "${Chain33Cli}" "${hash}"
    fi
}

function InitChain33Account() {
    # shellcheck disable=SC2154
    import_addr "${propKey}" "prop" "${propAddr}" 1000
    # shellcheck disable=SC2154
    import_addr "${votePrKey2}" "vote2" "${voteAddr2}" 100
    # shellcheck disable=SC2154
    import_addr "${votePrKey3}" "vote3" "${voteAddr3}" 100
#    import_addr "${votePrKey}" "vote" "${voteAddr}" 3200

    # shellcheck disable=SC2154
    import_addr "${changeKey}" "changeTest" "${changeAddr}" 10

    autonomyAddr=$(${Chain33Cli} exec addr -e autonomy)
#    ticketAddr=$(${Chain33Cli} exec addr -e ticket)

    hash=$(${Chain33Cli} send coins transfer -a 900 -n test -t "${autonomyAddr}" -k "${propKey}")
    check_tx "${Chain33Cli}" "${hash}"
#    hash=$(${Chain33Cli} send coins transfer -a 3100 -n test -t "${ticketAddr}" -k "${votePrKey}")
#    check_tx "${Chain33Cli}" "${hash}"

    # shellcheck disable=SC2154
#    hash=$(${Chain33Cli} send coins transfer -a 10 -n test -t "${voteAddr2}" -k "${minerAddr}")
#    check_tx "${Chain33Cli}" "${hash}"

    local count=0
    # shellcheck disable=SC2154
    # shellcheck disable=SC2068
    for key in ${arrayKey[@]}
    do
        import_addr "${key}" "board${count}" "${arrayAddr[count]}" 100
        count=$((count + 1))
    done
}

# chian33 初始化准备
function InitChain33() {
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"

    # init
#    ${Chain33Cli} seed save -p 1314fuzamei -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib"
#    ${Chain33Cli} wallet unlock -p 1314fuzamei -t 0
#    ${Chain33Cli} account import_key -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944 -l returnAddr

    InitChain33Account

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function StartChain33() {
#    kill_ebrelayer chain33
#    sleep 2
#
#    # delete chain33 datadir
#    rm ../../datadir ../../logs -rf
#
#    nohup ../../chain33 -f ./ci/autonomy/test.toml >chain33log.log 2>&1 &
#
#    sleep 1

    InitChain33
}