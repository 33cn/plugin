#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""

relay_parallel_exec=""

relay_CreateRawRelaySaveBTCHeadTx() {
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${relay_parallel_exec}"'","actionName":"BtcHeaders","payload":{"btcHeader":[{"hash":"5e7d9c599cd040ec2ba53f4dee28028710be8c135e779f65c56feadaae34c3f2","height":10,"version":536870912,"merkleRoot":"ab91cd4160e1379c337eee6b7a4bdbb7399d70268d86045aba150743c00c90b6","time":1530862108,"nonce":0,"bits":545259519,"previousHash":"604efe53975ab06cad8748fd703ad5bc960e8b752b2aae98f0f871a4a05abfc7","isReset":true}]}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "0x22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962" "${MAIN_HTTP}"
}

relay_CreateRawRelaySaveBTCHeadTx_11() {
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${relay_parallel_exec}"'","actionName":"BtcHeaders","payload":{"btcHeader":[{"hash":"7b7a4a9b49db5a1162be515d380cd186e98c2bf0bb90f1145485d7c43343fc7c","height":11,"version":536870912,"merkleRoot":"cfa9b66696aea63b7266ffaa1cb4b96c8dd6959eaabf2eb14173f4adaa551f6f","time":1530862108,"nonce":1,"bits":545259519,"previousHash":"5e7d9c599cd040ec2ba53f4dee28028710be8c135e779f65c56feadaae34c3f2","isReset":false}]}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "0x22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962" "${MAIN_HTTP}"
}

relay_CreateRawRelayOrderTx() {
    localCoinSymbol="$1"
    localCoinExec="$2"
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${relay_parallel_exec}"'","actionName":"Create","payload":{"operation":0,"xCoin":"BTC","xAmount":299000000,"xAddr":"1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT","localCoinAmount":1000000000,"localCoinSymbol":"'"$localCoinSymbol"'","localCoinExec":"'"$localCoinExec"'","xBlockWaits":6}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "0x22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962" "${MAIN_HTTP}"
}

relay_CreateRawRelayAcceptTx() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetSellRelayOrder","payload":{"addr":"1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3","status":"pending","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    #    echo "#request: $req"
    id=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.relayorders[0].id")
    if [ "$id" == null ]; then
        echo "id is null"
        echo_rst "$FUNCNAME" "$?"
        exit 1
    fi

    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${relay_parallel_exec}"'","actionName":"Accept","payload":{"orderId":"'"$id"'","xAddr":"1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "0xec9162ea5fc2f473ab8240619a0a0f495ba9e9e5d4d9c434b8794a68280236c4" "${MAIN_HTTP}"
}

relay_CreateRawRelayRevokeTx() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetSellRelayOrder","payload":{"addr":"1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3","status":"pending","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    #    echo "#request: $req"
    id=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.relayorders[0].id")
    if [ "$id" == null ]; then
        echo "id is null"
        echo_rst "$FUNCNAME" "$?"
        exit 1
    fi

    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${relay_parallel_exec}"'","actionName":"Revoke","payload":{"orderId":"'"$id"'","target":0,"action":1}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "0x22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962" "${MAIN_HTTP}"
}

relay_CreateRawRelayConfirmTx() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetRelayOrderByStatus","payload":{"addr":"","status":"locking","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    #    echo "#request: $req"
    id=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.relayorders[0].id")
    if [ "$id" == null ]; then
        echo "id is null"
        echo_rst "$FUNCNAME" "$?"
        exit 1
    fi

    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${relay_parallel_exec}"'","actionName":"ConfirmTx","payload":{"orderId":"'"$id"'","txHash":"6359f0868171b1d194cbee1af2f16ea598ae8fad666d9b012c8ed2b79a236ec4"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "$RETURN_RESP" "0xec9162ea5fc2f473ab8240619a0a0f495ba9e9e5d4d9c434b8794a68280236c4" "${MAIN_HTTP}"
}

query_GetRelayOrderByStatus() {
    status="$1"
    local req='{"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetRelayOrderByStatus","payload":{"addr":"","status":"'"$status"'","coins":["BTC"],"pageNumber":0,"pageSize":0}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result.relayorders[0].id != null)' "$FUNCNAME"
}

query_GetSellRelayOrder() {
    local req='{"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetSellRelayOrder","payload":{"addr":"1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3","status":"pending","coins":["BTC"],"pageNumber":0,"pageSize":0}}]}'
    resok='(.error|not) and (.result.relayorders[0].status == "pending") and (.result.relayorders[0].operation == 0) and (.result.relayorders[0].id != null)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

query_GetBuyRelayOrder() {
    local req='{"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBuyRelayOrder","payload":{"addr":"1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum","status":"locking","coins":["BTC"],"pageNumber":0,"pageSize":0}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result.relayorders[0].status == "locking")' "$FUNCNAME"
}

query_GetBTCHeaderList() {
    local req='{"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBTCHeaderList","payload":{"reqHeight":"10","counts":10,"direction":0}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result.heights|length == 2)' "$FUNCNAME"
}

query_GetBTCHeaderCurHeight() {
    local req='{"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBTCHeaderCurHeight","payload":{"baseHeight":"0"}}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result.baseHeight == "10") and (.result.curHeight == "10")' "$FUNCNAME"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    local relay_addr=""
    if [ "$ispara" == true ]; then
        relay_parallel_exec="user.p.para.relay"
        relay_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.relay"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        relay_parallel_exec="relay"
        relay_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"relay"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "relayaddr=$relay_addr"

    local main_ip=${MAIN_HTTP//8901/8801}

    chain33_ImportPrivkey "22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962" "1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3" "relay_sell" "${main_ip}"
    chain33_ImportPrivkey "ec9162ea5fc2f473ab8240619a0a0f495ba9e9e5d4d9c434b8794a68280236c4" "1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum" "relay_acc" "$main_ip"

    local sellAddr="1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3"
    local accepAddr="1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$sellAddr" 12000000000 "${main_ip}"
        chain33_QueryBalance "${sellAddr}" "$main_ip"

        chain33_applyCoins "$accepAddr" 12000000000 "${main_ip}"
        chain33_QueryBalance "${accepAddr}" "$main_ip"
    else
        chain33_applyCoins "$sellAddr" 1000000000 "${main_ip}"
        chain33_QueryBalance "${sellAddr}" "$main_ip"

        chain33_applyCoins "$accepAddr" 1000000000 "${main_ip}"
        chain33_QueryBalance "${accepAddr}" "$main_ip"
        local para_ip="${MAIN_HTTP}"
        chain33_ImportPrivkey "22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962" "1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3" "relay_sell" "$para_ip"
        chain33_ImportPrivkey "ec9162ea5fc2f473ab8240619a0a0f495ba9e9e5d4d9c434b8794a68280236c4" "1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum" "relay_acc" "$para_ip"

        chain33_applyCoins "$sellAddr" 12000000000 "${para_ip}"
        chain33_QueryBalance "${sellAddr}" "$para_ip"
        chain33_applyCoins "$accepAddr" 12000000000 "${para_ip}"
        chain33_QueryBalance "${accepAddr}" "$para_ip"
    fi

    chain33_SendToAddress "$sellAddr" "$relay_addr" 10000000000 "${MAIN_HTTP}"
    chain33_QueryExecBalance "${sellAddr}" "relay" "$MAIN_HTTP"
    chain33_SendToAddress "$accepAddr" "$relay_addr" 10000000000 "${MAIN_HTTP}"
    chain33_QueryExecBalance "${accepAddr}" "relay" "$MAIN_HTTP"

}
function run_testcases() {
    relay_CreateRawRelaySaveBTCHeadTx
    query_GetBTCHeaderCurHeight

    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    if [ "$ispara" == true ]; then
        relay_CreateRawRelayOrderTx "GD"
    else
        relay_CreateRawRelayOrderTx
    fi

    query_GetSellRelayOrder
    query_GetRelayOrderByStatus "pending"

    relay_CreateRawRelayAcceptTx
    query_GetBuyRelayOrder
    query_GetRelayOrderByStatus "locking"

    relay_CreateRawRelayConfirmTx
    query_GetRelayOrderByStatus "confirming"

    relay_CreateRawRelayOrderTx
    relay_CreateRawRelayRevokeTx
    query_GetRelayOrderByStatus "canceled"

    relay_CreateRawRelaySaveBTCHeadTx_11
    query_GetBTCHeaderList
}

function rpc_test() {
    chain33_RpcTestBegin Relay
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases
    chain33_RpcTestRst Relay "$CASE_ERR"
}

chain33_debug_function rpc_test "$1"
