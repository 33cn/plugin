#!/usr/bin/env bash
# shellcheck disable=SC2128

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')

    beneficiary=12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
    beneficiary_key=0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01
    #owner=14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    #owner_key=CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944
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
    fi
    exec_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"'${exec_name}'"}]}' ${MAIN_HTTP} | jq -r ".result")
    echo "exec_addr=${exec_addr}"

    # json 中 \n \t 需要转意, " 影响json的结构， 需要转意
    jsCode=$(sed 's/"/\\"/g' <./js/test.js | sed ':a;N;s/\n/\\n/g;ta' | sed 's/\t/\\t/g')
}

function configJSCreator() {
    req='{"jsonrpc": "2.0", "method" :  "Chain33.CreateTransaction" , "params":[{"execer":"'${manager_name}'","actionName":"Modify","payload":{"key":"js-creator","op":"add", "value" : "'${beneficiary}'"}}]}'
    echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${super_manager}" "${MAIN_HTTP}"
}

function createJSContract() {
    req='{"jsonrpc": "2.0", "method" :  "Chain33.CreateTransaction" , "params":[{"execer":"'${exec_name}'","actionName":"Create","payload":{"name":"'${game}'","code":"'${jsCode}'"}}]}'
    echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${beneficiary_key}" "${MAIN_HTTP}"
}

function callJS() {
    #the_exec=
    req='{"jsonrpc": "2.0", "method" :  "Chain33.CreateTransaction" , "params":[{"execer":"'${user_game}'","actionName":"Call","payload":{"name":"'${game}'","funcname":"hello", "args" : "{}"}}]}'
    # echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "${beneficiary_key}" "${MAIN_HTTP}"
}

function queryJS() {
    req='{"jsonrpc": "2.0", "method" :  "Chain33.Query" , "params":[{"execer":"'${user_game}'","funcName":"Query","payload":{"name":"'${game}'","funcname":"hello", "args" : "{}"}}]}'
    # echo "#request: $req"
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

function run_testcases() {
    configJSCreator

    createJSContract
    callJS
    queryJS
}

function debug_function() {
    set -x
    eval "$@"
    set +x
}

function rpc_test() {
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases

    if [ -n "$CASE_ERR" ]; then
        echo "=======js rpc test  error ==========="
        exit 1
    else
        echo "====== js rpc test  pass ==========="
    fi
}

rpc_test "$1"
