#!/usr/bin/env bash

NODE3=build_chain33_1
PARA_CLI="docker exec ${NODE3} /root/chain33-para-cli"
CLI="docker exec ${NODE3} /root/chain33-cli"

PRIVACY_ADDR_A="1HGPrjc6H7yBzFV5yCbibvnSUGUgdDNQi3"
PRIVACY_ADDR_B="19WGov4b7wLf4f8JHMRDnJsGVNMDzap38w"
KEY_ADDR_A="0x227df96a414e26e85c7d87a12296344e6a731ce73e424ba9845cb305ec963843"
KEY_ADDR_B="0x64259075bf2e5a74334442f5048ceaf8427f6097e2dac99c0e463785c7768550"
# 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

# shellcheck disable=SC2128
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

function block_wait() {
    if [ "$#" -lt 2 ]; then
        echo "wrong block_wait params"
        exit 1
    fi
    cur_height=$(${1} block last_header | jq ".height")
    expect=$((cur_height + ${2}))
    count=0
    while true; do
        new_height=$(${1} block last_header | jq ".height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 1
    done
    echo "wait new block $count s"
}

function query_tx() {
    block_wait "${1}" 1

    local times=100
    while true; do
        ret=$(${1} tx query -s "${2}" | jq -r ".tx.hash")
        echo "query hash is ${2} return ${ret} "
        if [ "${ret}" != "${2}" ]; then
            block_wait "${1}" 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "query tx= $2 failed"
                exit 1
            fi
        else
            echo "query tx= $2 success"
            break
        fi
    done
}

function signAndSendPrivacyTx() {
    local rawtxdata=${1}
    local addr=${2}

    local signdata=$(${PARA_CLI} wallet sign -a "${addr}" -d "${rawtxdata}")
    local priavacyData=$(${PARA_CLI} para privacy create -d "${signdata}")
    local signData2=$(${PARA_CLI} wallet sign -a "${addr}" -d "${priavacyData}")
    txHash=$(${PARA_CLI} wallet send -d "${signData2}")
    query_tx "${PARA_CLI}" "${txHash}"
}

function signAndSendPrivacyTx2() {
    local rawtxdata=${1}
    local private=${2}

    local signdata=$(${PARA_CLI} wallet sign -k "${private}" -d "${rawtxdata}")
    local priavacyData=$(${PARA_CLI} para privacy create -d "${signdata}")
    local signData2=$(${PARA_CLI} wallet sign -k "${private}" -d "${priavacyData}")
    txHash=$(${PARA_CLI} wallet send -d "${signData2}")
    query_tx "${PARA_CLI}" "${txHash}"
}

function show() {
    ${PARA_CLI} para privacy show -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b -s ${1}
}

function echoRed() {
    echo -e "${RED}${1} ${NOC}"
}

function echoGre() {
    echo -e "${GRE}${1} ${NOC}"
}

function init() {
    echoGre "=========== # import_key and coins transfer ============="
    local cli=${1}
    local amount=${2}
    ${cli} account import_key -k ${KEY_ADDR_A} -l "privacy_test_A"
    ${cli} account import_key -k ${KEY_ADDR_B} -l "privacy_test_B"
    txHash=$(${cli} send coins transfer -a ${amount} -n test -t ${PRIVACY_ADDR_A} -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    query_tx "${cli}" "${txHash}"
    txHash=$(${cli} send coins transfer -a ${amount} -n test -t ${PRIVACY_ADDR_B} -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    query_tx "${cli}" "${txHash}"

    echoGre "=========== # init end ============="
}

function coinsPrivacyTransfer() {
    echoGre "=========== # coins transfer ============="
    rawtxdata=$(${PARA_CLI} coins transfer -a 10 -n test -t "${PRIVACY_ADDR_B}")
    signAndSendPrivacyTx "${rawtxdata}" "${PRIVACY_ADDR_A}"
    show ${txHash}

    ${PARA_CLI} account balance -a ${PRIVACY_ADDR_B} -e coins
    local balance=$(${PARA_CLI} account balance -a ${PRIVACY_ADDR_B} -e coins | jq -r '.balance')
    if [ "${balance}" != "20.0000" ]; then
        echoRed "wrong, should be 10010.0000"
        exit 1
    fi

    echoGre "=========== # coins transfer end ============="
}

function tokenPrivacyTransfer() {
    echoGre "=========== # 0.token precreate ============="
    rawtxdata=$(${PARA_CLI} token precreate -f 0.001 -i test -n privacyTest -a ${PRIVACY_ADDR_A} -p 0 -s PTC -t 100000)
    signAndSendPrivacyTx "${rawtxdata}" "${PRIVACY_ADDR_A}"
    ${PARA_CLI} token get_precreated

    echoGre "=========== # 1.token finish ============="
    rawtxdata=$(${PARA_CLI} token finish -f 0.001 -a ${PRIVACY_ADDR_A} -s PTC)
    signAndSendPrivacyTx2 "${rawtxdata}" "0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc"
    ${PARA_CLI} token get_finish_created

    ${PARA_CLI} token token_balance -a ${PRIVACY_ADDR_A} -e token -s PTC
    balance=$(${PARA_CLI} token token_balance -a ${PRIVACY_ADDR_A} -e token -s PTC | jq -r '.[]|.balance')
    if [ "${balance}" != "100000.0000" ]; then
        echoRed "wrong para token genesis create, should be 100000.0000"
        exit 1
    fi

    echoGre "=========== # 2.token transfer ============="
    rawtxdata=$(${PARA_CLI} token transfer -a 10 -s PTC -t ${PRIVACY_ADDR_B})
    signAndSendPrivacyTx "${rawtxdata}" "${PRIVACY_ADDR_A}"

    ${PARA_CLI} token token_balance -a ${PRIVACY_ADDR_B} -e token -s PTC
    balance=$(${PARA_CLI} token token_balance -a ${PRIVACY_ADDR_B} -e token -s PTC | jq -r '.[]|.balance')
    if [ "${balance}" != "10.0000" ]; then
        echoRed "wrong para token transfer, should be 10.0000"
        exit 1
    fi

    echoGre "=========== # 3.token send exec ============="
    rawtxdata=$(${PARA_CLI} token send_exec -a 10100 -s PTC -e trade)
    signAndSendPrivacyTx "${rawtxdata}" "${PRIVACY_ADDR_A}"

    ${PARA_CLI} token token_balance -a 12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13 -e token -s PTC
    balance=$(${PARA_CLI} token token_balance -a 12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13 -e token -s PTC | jq -r '.[]|.balance')
    if [ "${balance}" != "10100.0000" ]; then
        echoRed "wrong para token send exec, should be 10100.0000"
        exit 1
    fi

    echoGre "=========== # 4.token withdraw ============="
    rawtxdata=$(${PARA_CLI} token withdraw -a 100 -s PTC -e trade)
    signAndSendPrivacyTx "${rawtxdata}" "${PRIVACY_ADDR_A}"

    ${PARA_CLI} token token_balance -a 12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13 -e token -s PTC
    balance=$(${PARA_CLI} token token_balance -a 12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13 -e token -s PTC | jq -r '.[]|.balance')
    if [ "${balance}" != "10000.0000" ]; then
        echoRed "wrong para token withdraw, should be 10000.0000"
        exit 1
    fi

    echoGre "=========== # token end ============="
}

function tradePrivacyTransfer() {
    echoGre "=========== # 0.trade sell ============="
    rawtxdata=$(${PARA_CLI} trade sell -m 1 -p 0.001 -s PTC -t 100)
    signAndSendPrivacyTx "${rawtxdata}" "${PRIVACY_ADDR_A}"
    sellID=$(${PARA_CLI} para privacy show -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b -s ${txHash} | jq -r '.txs[0].receipt.logs[1].log.base.sellID' | awk -F '-' '{print $4}')
    if [ "${sellID}" == "" ]; then
        echoRed "wrong, sellID empty txHash = ${txHash}"
        exit 1
    fi

    ${PARA_CLI} token token_balance -a ${PRIVACY_ADDR_A} -s PTC -e trade
    frozenBalance=$(${PARA_CLI} token token_balance -a ${PRIVACY_ADDR_A} -s PTC -e trade | jq -r '.[0].frozen')
    if [ "${frozenBalance}" != "100.0000" ]; then
        echoRed "para token send exec, should be 100.0000"
        exit 1
    fi

    echoGre "=========== # 1.trade buy ============="
    #购买前需要在trade合约中有足额的BTY
    rawtxdata=$(${PARA_CLI} coins send_exec -a 10 -e trade)
    signAndSendPrivacyTx "${rawtxdata}" "${PRIVACY_ADDR_B}"

    ${PARA_CLI} account balance -a ${PRIVACY_ADDR_B} -e user.p.para.trade
    local balance=$(${PARA_CLI} account balance -a ${PRIVACY_ADDR_B} -e user.p.para.trade | jq -r '.balance')
    if [ "${balance}" != "10.0000" ]; then
        echoRed "wrong, should be 10.0000"
        exit 1
    fi

    # 购买指定卖单的token
    rawtxdata=$(${PARA_CLI} trade buy -c 200 -s ${sellID})
    signAndSendPrivacyTx "${rawtxdata}" "${PRIVACY_ADDR_B}"

    ${PARA_CLI} token token_balance -a ${PRIVACY_ADDR_B} -s PTC -e trade
    balance=$(${PARA_CLI} token token_balance -a ${PRIVACY_ADDR_B} -s PTC -e trade | jq -r '.[].balance')
    if [ "${balance}" != "2.0000" ]; then
        echoRed "wrong, should be 2.0000"
        exit 1
    fi

    echoGre "=========== # trade end ============="
}

function privacy_test() {
    init "${CLI}" 5
    init "${PARA_CLI}" 10

    coinsPrivacyTransfer
    tokenPrivacyTransfer
    tradePrivacyTransfer
}
