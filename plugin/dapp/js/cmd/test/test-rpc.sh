#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')

    beneficiary_key=0xf146df80206194c81e0b3171db6aa40c7ad6182a24560698d4871d4dc75223ce
    beneficiary=1DwHQp8S7RS9krQTyrqePxRyvaLcuoQGks
    chain33_applyCoins "${beneficiary}" 10000000000 "${MAIN_HTTP}"
    echo "ipara=$ispara"
    manager_name="manage"
    exec_name="jsvm"
    game="game"
    user_game="user.${exec_name}.${game}"
    super_manager=0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01
    if [ "$ispara" == true ]; then
        exec_name="user.p.para."${exec_name}
        manager_name="user.p.para."${manager_name}
        user_game="user.p.para."${user_game}
        super_manager=0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc
        ## fee
        local main_ip=${MAIN_HTTP//8901/8801}
        chain33_applyCoins "${beneficiary}" 10000000000 "${main_ip}"
    fi
    exec_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'${exec_name}'"}]}' ${MAIN_HTTP} | jq -r ".result")
    echo "exec_addr=${exec_addr}"

    # json 中 \n \t 需要转意, " 影响json的结构， 需要转意
    jsCode=$(sed 's/"/\\"/g' <./js/test.js | sed ':a;N;s/\n/\\n/g;ta' | sed 's/\t/\\t/g')
}

function configJSCreator() {
    req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'${manager_name}'","actionName":"Modify","payload":{"key":"js-creator","op":"add","value":"'${beneficiary}'"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "${super_manager}" "${MAIN_HTTP}"
}

function createJSContract() {
    req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'${exec_name}'","actionName":"Create","payload":{"name":"'${game}'","code":"'${jsCode}'"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "${beneficiary_key}" "${MAIN_HTTP}"
}

function callJS() {
    req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'${user_game}'","actionName":"Call","payload":{"name":"'${game}'","funcname":"hello","args":"{}"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "${beneficiary_key}" "${MAIN_HTTP}"
}

function queryJS() {
    req='{"method":"Chain33.Query","params":[{"execer":"'${user_game}'","funcName":"Query","payload":{"name":"'${game}'","funcname":"hello","args":"{}"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME"
}

function run_testcases() {
    configJSCreator
    createJSContract
    callJS
    queryJS
}

function rpc_test() {
    chain33_RpcTestBegin js
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases
    chain33_RpcTestRst js "$CASE_ERR"
}

chain33_debug_function rpc_test "$1"
