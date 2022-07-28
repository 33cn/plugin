#!/usr/bin/env bash
# shellcheck disable=SC2154
# shellcheck disable=SC2034
# shellcheck disable=SC2128
set -x
set -e

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'
IYellow="\033[0;93m"

CLI="docker exec build_chain33_1 /root/chain33-cli"

function GetChain33Addr() {
    chain33Addr1=$(${CLI} zksync l2addr -k "$1")
    echo "${chain33Addr1}"
}

function block_wait() {
    if [ "$#" -lt 2 ]; then
        echo "wrong block_wait params"
        exit 1
    fi
    cur_height=$(${1} block last_header | jq ".height")
    expect=$((cur_height + ${2}))
    local count=0
    while true; do
        new_height=$(${1} block last_header | jq ".height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 0.1
    done
    echo "wait new block $count/10 s, cur height=$expect,old=$cur_height"
}

# 检查交易是否执行成功 $1:cli 路径  $2:交易hash
function check_tx() {
    set +x
    local CLI=${1}

    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong check_tx parameters${NOC}"
        exit 1
    fi

    if [[ ${2} == "" ]]; then
        echo -e "${RED}wrong check_tx txHash is empty${NOC}"
        exit 1
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
        exit 1
    fi
}

function Chain33ImportKey() {
    local key="${1}"
    local addr="${2}"
    local label="${3}"
    local amount="${4}"
    # 转帐到 DeployAddr 需要手续费
    result=$(${Chain33Cli} account import_key -k "${key}" -l "${label}")
    check_addr "${result}" "${addr}"
    hash=$(${Chain33Cli} send coins transfer -a "${amount}" -n test -t "${addr}" -k 4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01)
    check_tx "${Chain33Cli}" "${hash}"
}

function create_addr() {
    local label="${1}"
#    local amount="${2}"
    local amount=1
    # 转帐到 DeployAddr 需要手续费
    addr=$(${Chain33Cli} account create -l "${label}" | jq -r ".acc.addr")
    hash=$(${Chain33Cli} send coins transfer -a "${amount}" -n test -t "${addr}" -k 4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01)
#    check_tx "${Chain33Cli}" "${hash}"
    key=$(${Chain33Cli} account dump_key -a "${addr}" | jq -r ".data")
}

function init_seed() {
    seed=$(${Chain33Cli} seed generate -l 0)
    seed=$(${Chain33Cli} seed generate -l 0)
    seed=$(${Chain33Cli} seed generate -l 0)
    seed=$(${Chain33Cli} seed generate -l 0)
    seed=$(${Chain33Cli} seed generate -l 0)
    ${Chain33Cli} seed save -p 123456fzm -s "${seed}"
    ${Chain33Cli} wallet unlock -p 123456fzm
}

function transfer_bty_init() {
    echo -e "${GRE}=========== $FUNCNAME ===========${NOC}"
#    # 会丢包 少量地址一直没有到帐
#    parallel --jobs 10 ${CLI} send coins transfer -a 1 -n test -t {} -k 4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01 :::  \
#    ${addr[0]} ${addr[1]} ${addr[2]} ${addr[3]} ${addr[4]} ${addr[5]} ${addr[6]} ${addr[7]} ${addr[8]} ${addr[9]} \
#    ${addr[10]} ${addr[11]} ${addr[12]} ${addr[13]} ${addr[14]} ${addr[15]} ${addr[16]} ${addr[17]} ${addr[18]} ${addr[19]} \
#    ${addr[20]} ${addr[21]} ${addr[22]} ${addr[23]} ${addr[24]} ${addr[25]} ${addr[26]} ${addr[27]} ${addr[28]} ${addr[29]} \
#    ${addr[30]} ${addr[31]} ${addr[32]} ${addr[33]} ${addr[34]} ${addr[35]} ${addr[36]} ${addr[37]} ${addr[38]} ${addr[39]} \
#    ${addr[40]} ${addr[41]} ${addr[42]} ${addr[43]} ${addr[44]} ${addr[45]} ${addr[46]} ${addr[47]} ${addr[48]} ${addr[49]} \
#    ${addr[50]} ${addr[51]} ${addr[52]} ${addr[53]} ${addr[54]} ${addr[55]} ${addr[56]} ${addr[57]} ${addr[58]} ${addr[59]} \
#    ${addr[60]} ${addr[61]} ${addr[62]} ${addr[63]} ${addr[64]} ${addr[65]} ${addr[66]} ${addr[67]} ${addr[68]} ${addr[69]} \
#    ${addr[70]} ${addr[71]} ${addr[72]} ${addr[73]} ${addr[74]} ${addr[75]} ${addr[76]} ${addr[77]} ${addr[78]} ${addr[79]} \
#    ${addr[80]} ${addr[81]} ${addr[82]} ${addr[83]} ${addr[84]} ${addr[85]} ${addr[86]} ${addr[87]} ${addr[88]} ${addr[89]} \
#    ${addr[90]} ${addr[91]} ${addr[92]} ${addr[93]} ${addr[94]} ${addr[95]} ${addr[96]} ${addr[97]} ${addr[98]} ${addr[99]} \
#    ${addr[100]} ${addr[101]} ${addr[102]} ${addr[103]} ${addr[104]} ${addr[105]} ${addr[106]} ${addr[107]} ${addr[108]} ${addr[109]}
}

# 判断结果 $1 和 $2 是否相等
function is_equal() {
    set +x
    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong parameter${NOC}"
        exit 1
    fi

    if [[ $1 != "$2" ]]; then
        echo -e "${RED}$1 != ${2}${NOC}"
        exit 1
    fi

    set -x
}

