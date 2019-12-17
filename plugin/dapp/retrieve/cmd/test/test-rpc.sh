#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -e
set -o pipefail

MAIN_HTTP=""

source ../dapp-test-common.sh

# TODO
# 1. 合约测试的先后顺序 是否可以在指定合约之后测试
# 2. 或将资产类的合约先测试
# 3. 或资产类的合约提供创建的函数 创建一个某某名字的token
function updateConfig() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "manage","actionName":"Modify","payload":{ "key": "token-blacklist","value": "BTY","op": "add","addr": ""}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "update config create tx" 1
        return
    fi

    chain33_SignAndSendTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"
}

function token_preCreate() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenPreCreateTx","params":[{"name": "yinhebib", "symbol": "'"$1"'", "total": 1000000000000, "price": 0, "category": 1,"owner":"'"$2"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token preCreate create tx" 1
        return
    fi

    chain33_SignAndSendTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"
}

function token_finish() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"token.CreateRawTokenFinishTx","params":[{"symbol": "'"$1"'", "owner":"'"$2"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token finish create tx" 1
        return
    fi

    chain33_SignAndSendTx "${unsignedTx}" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" "${MAIN_HTTP}"
}

function token_sendExec() {
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.CreateTransaction","params":[{"execer": "'"${token_ame}"'","actionName":"TransferToExec","payload": {"cointoken":"'"$1"'", "amount": "10000000000", "note": "", "to": "'"${retrieve_addr}"'", "execName": "'"${retrieve_name}"'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        echo_rst "token sendExec create tx" 1
        return
    fi

    chain33_SignAndSendTx "${unsignedTx}" "$3" "${MAIN_HTTP}"
}

function createToken() {
    # symbol owner owner_key
    updateConfig
    token_preCreate "$1" "$2"
    token_finish "$1" "$2"
    token_sendExec "$1" "$2" "$3"
}

retrieve_Backup() {
    local req='{"method":"retrieve.CreateRawRetrieveBackupTx","params":[{"backupAddr":"'$retrieve1'","defaultAddr":"'$retrieve2'","delayPeriod": 61}]}'
    tx=$(curl -ksd "$req" ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$retrieve2_key" ${MAIN_HTTP} "$FUNCNAME"
}

retrieve_Prepare() {
    local req='{"method":"retrieve.CreateRawRetrievePrepareTx","params":[{"backupAddr":"'$retrieve1'","defaultAddr":"'$retrieve2'"}]}'
    tx=$(curl -ksd "$req" ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$retrieve1_key" ${MAIN_HTTP} "$FUNCNAME"
}

retrieve_Perform() {
    local req='{"method":"retrieve.CreateRawRetrievePerformTx","params":[{"backupAddr":"'$retrieve1'","defaultAddr":"'$retrieve2'"}]}'
    tx=$(curl -ksd "$req" ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$retrieve1_key" ${MAIN_HTTP} "$FUNCNAME"
}

retrieve_Perform_Token() {
    local req='{"method":"retrieve.CreateRawRetrievePerformTx","params":[{"backupAddr":"'$retrieve1'","defaultAddr":"'$retrieve2'","assets": [{"exec":"token","symbol":"'"$symbol"'"}] }]}'
    tx=$(curl -ksd "$req" ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$retrieve1_key" ${MAIN_HTTP} "$FUNCNAME"
}

retrieve_Cancel() {
    local req='{"method":"retrieve.CreateRawRetrieveCancelTx","params":[{"backupAddr":"'$retrieve1'","defaultAddr":"'$retrieve2'"}]}'
    tx=$(curl -ksd "$req" ${MAIN_HTTP} | jq -r ".result")
    chain33_SignAndSendTxWait "$tx" "$retrieve2_key" ${MAIN_HTTP} "$FUNCNAME"
}

retrieve_QueryResult() {
    local status=$1
    local req='{"method":"Chain33.Query","params":[{"execer":"retrieve","funcName":"GetRetrieveInfo","payload":{"backupAddress":"'$retrieve1'", "defaultAddress":"'$retrieve2'"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.result.status == '"$status"')' "$FUNCNAME"
}

retrieve_QueryAssetResult() {
    local status=$1
    local req='{"method":"Chain33.Query","params":[{"execer":"retrieve","funcName":"GetRetrieveInfo","payload":{"backupAddress":"'$retrieve1'", "defaultAddress":"'$retrieve2'","assetExec":"token", "assetSymbol":"'"$symbol"'"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.result.status == '"$status"')' "$FUNCNAME"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        token_ame="user.p.para.token"
        retrieve_name="user.p.para.retrieve"
        retrieve_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.retrieve"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        token_ame="token"
        retrieve_name="retrieve"
        retrieve_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"retrieve"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi

    retrieve1_key=0xf54a8ffe50b308a2d37f44a9e595fd2d156c09732b712b8548eccf1dce4d0fde
    retrieve1=19P3ZQg5VYgzTUGvLD4etFSrh74mk6HUWW
    chain33_ImportPrivkey "${retrieve1_key}" "${retrieve1}" "retrieve1" "${MAIN_HTTP}"
    chain33_applyCoins "${retrieve1}" 10000000000 "${MAIN_HTTP}"

    retrieve2_key=0x61d86bf173ed37835fba9ff5b062382249c1b978cb2d3c6e2a3abbdf38314432
    retrieve2=18x7o8Uktqs8RHEcPsMLJvaHKo22swLbqF
    chain33_ImportPrivkey "${retrieve2_key}" "${retrieve2}" "retrieve2" "${MAIN_HTTP}"
    chain33_applyCoins "${retrieve2}" 10000000000 "${MAIN_HTTP}"

    if [ "$ispara" == true ]; then
        local main_ip=${MAIN_HTTP//8901/8801}
        chain33_applyCoins "${retrieve1}" 1000000000 "${main_ip}"
        chain33_applyCoins "${retrieve2}" 1000000000 "${main_ip}"
    fi

    chain33_SendToAddress "$retrieve1" "$retrieve_addr" 1000000000 ${MAIN_HTTP}
    chain33_SendToAddress "$retrieve2" "$retrieve_addr" 1000000000 ${MAIN_HTTP}
    symbol="RETRIEVE"
    createToken "$symbol" "$retrieve2" "$retrieve2_key"
    chain33_BlockWait 1 "${MAIN_HTTP}"
}

function run_test() {
    retrieve_Backup
    retrieve_QueryResult 1

    retrieve_Prepare
    retrieve_QueryResult 2

    retrieve_Cancel
    retrieve_QueryResult 4

    retrieve_Backup
    retrieve_QueryResult 1

    retrieve_Prepare
    retrieve_QueryResult 2

    sleep 61
    retrieve_Perform
    retrieve_QueryResult 3
    retrieve_Perform_Token
    retrieve_QueryAssetResult 3
}

function main() {
    chain33_RpcTestBegin retrieve
    MAIN_HTTP="$1"
    echo "ip=$MAIN_HTTP"

    init
    run_test
    chain33_RpcTestRst retrieve "$CASE_ERR"
}

chain33_debug_function main "$1"
