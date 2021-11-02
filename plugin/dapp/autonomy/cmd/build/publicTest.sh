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

function exit_test() {
    exit 1
}

# $1 dockerName
function get_docker_addr() {
    local dockerAddr=$(docker inspect "${1}" | jq ".[].NetworkSettings.Networks" | grep "IPAddress" | awk '{ print $2}' | sed 's/\"//g' | sed 's/,//g')
    echo "${dockerAddr}"
}

# chain33 区块等待 $1:cli 路径  $2:等待高度
function block_wait() {
    set +x
    local CLI=${1}

    if [[ $# -lt 1 ]]; then
        echo -e "${RED}wrong block_wait parameter${NOC}"
        exit_test
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
        exit_test
    fi

    if [[ ${2} == "" ]]; then
        echo -e "${RED}wrong check_tx txHash is empty${NOC}"
        exit_test
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
        hashInfo=$(${CLI} tx query -s "${2}")
        echo -e "${RED}query tx: ${hashInfo}${NOC}"
        exit_test
    fi
}

# 检查地址是否匹配 $1返回结果 $2匹配地址
function check_addr() {
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong check number parameters${NOC}"
        exit_test
    fi

    addr=$(echo "${1}" | jq -r ".acc.addr")
    if [[ ${addr} != "${2}" ]]; then
        echo -e "${RED}error addr, expect ${1}, get ${2}${NOC}"
        exit_test
    fi
}

# 判断结果 $1 和 $2 是否相等
function is_equal() {
    set +x
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_test
    fi

    if [[ $1 != "$2" ]]; then
        echo -e "${RED}$1 != ${2}${NOC}"
        exit_test
    fi

    set -x
}

# 判断结果 $1 和 $2 是否不相等
function is_not_equal() {
    set +x
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit_test
    fi

    if [[ $1 == "$2" ]]; then
        echo -e "${RED}$1 == ${2}${NOC}"
        exit_test
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
    {
        import_addr "${propKey}" "prop" "${propAddr}" 1000
        import_addr "${votePrKey2}" "vote2" "${voteAddr2}" 100
        import_addr "${votePrKey3}" "vote3" "${voteAddr3}" 100
        import_addr "${changeKey}" "changeTest" "${changeAddr}" 10
    }
    autonomyAddr=$(${Chain33Cli} exec addr -e autonomy)
    hash=$(${Chain33Cli} send coins transfer -a 900 -n test -t "${autonomyAddr}" -k "${propKey}")
    check_tx "${Chain33Cli}" "${hash}"

    local count=0
    # shellcheck disable=SC2154
    # shellcheck disable=SC2068
    for key in ${arrayKey[@]}; do
        import_addr "${key}" "board${count}" "${arrayAddr[count]}" 100
        count=$((count + 1))
    done
}
