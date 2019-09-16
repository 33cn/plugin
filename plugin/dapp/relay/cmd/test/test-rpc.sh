#!/usr/bin/env bash
# shellcheck disable=SC2128

# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""
CASE_ERR=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# $2=0 means true, other false
echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="err"
    fi

}

function block_wait() {
    local req='"method":"Chain33.GetLastHeader","params":[]'
    cur_height=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq ".result.height")
    expect=$((cur_height + ${1}))
    local count=0
    while true; do
        new_height=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq ".result.height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 1
    done
    echo "wait new block $count s, cur height=$expect,old=$cur_height"
}

function query_tx() {
    block_wait 1
    local txhash="$1"
    local req='"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]'
    # echo "req=$req"
    local times=10
    while true; do
        ret=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.tx.hash")
        echo "====query tx= ${1}, return=$ret "
        if [ "${ret}" != "${1}" ]; then
            block_wait 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "====query tx=$1 failed"
                echo "req=$req"
                curl -ksd "{$req}" ${MAIN_HTTP}
                exit 1
            fi
        else
            echo "====query tx=$1  success"
            break
        fi
    done
}




signrawtx() {
    txHex="$1"
    priKey="$2"
    local req='"method":"Chain33.SignRawTx","params":[{"privkey":"'"$priKey"'","txHex":"'"$txHex"'","expire":"120s"}]'
    #    echo "#request SignRawTx: $req"
    signedTx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")
    #    echo "signedTx=$signedTx"
    if [ "$signedTx" != null ]; then
        sendTx "$signedTx"
    else
        echo "signedTx null error"
    fi
}

sendTx() {
    signedTx=$1
    local req='"method":"Chain33.SendTransaction","params":[{"token":"BTY","data":"'"$signedTx"'"}]'
    #    echo "#request sendTx: $req"
    #    curl -ksd "{$req}" ${MAIN_HTTP}
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    err=$(jq '(.error)' <<<"$resp")
    txhash=$(jq -r ".result" <<<"$resp")
    if [ "$err" == null ]; then
        #   echo "tx hash: $txhash"
        query_tx "$txhash"
    else
        echo "send tx error:$err"
    fi

}

relay_CreateRawRelayOrderTx() {
    local req='"method":"relay.CreateRawRelayOrderTx","params":[{"operation":0,"coin":"BTC","amount":299000000,"addr":"1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT","btyAmount":1000000000,"coinWaits":6}]'
    # echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signrawtx "$rawtx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
}

relay_CreateRawRelayAcceptTx() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetSellRelayOrder","payload":{"addr":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv","status":"pending","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    #    echo "#request: $req"
    id=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.relayorders[0].id")
    if [ "$id" == null ]; then
        echo "id is null"
        echo_rst "$FUNCNAME" "$?"
        exit 1
    fi

    local req='"method":"relay.CreateRawRelayAcceptTx","params":[{"orderId":"'"$id"'","coinAddr":"1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT"}]'
    # echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signrawtx "$rawtx" "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588"

}

relay_CreateRawRelayRevokeTx() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetSellRelayOrder","payload":{"addr":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv","status":"pending","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    #    echo "#request: $req"
    id=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.relayorders[0].id")
    if [ "$id" == null ]; then
        echo "id is null"
        echo_rst "$FUNCNAME" "$?"
        exit 1
    fi

    local req='"method":"relay.CreateRawRelayRevokeTx","params":[{"orderId":"'"$id"'","target":0,"action":1}]'
    # echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signrawtx "$rawtx" "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"

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

    local req='"method":"relay.CreateRawRelayConfirmTx","params":[{"orderId":"'"$id"'","rawTx":"6359f0868171b1d194cbee1af2f16ea598ae8fad666d9b012c8ed2b79a236ec4"}]'
    # echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signrawtx "$rawtx" "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588"

}

relay_CreateRawRelaySaveBTCHeadTx() {
    local req='"method":"relay.CreateRawRelaySaveBTCHeadTx","params":[{"hash":"5e7d9c599cd040ec2ba53f4dee28028710be8c135e779f65c56feadaae34c3f2","height":10,"version":536870912,"merkleRoot":"ab91cd4160e1379c337eee6b7a4bdbb7399d70268d86045aba150743c00c90b6","time":1530862108,"nonce":0,"bits":545259519,"previousHash":"604efe53975ab06cad8748fd703ad5bc960e8b752b2aae98f0f871a4a05abfc7","isReset":true}]'
    # echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signrawtx "$rawtx" "0x4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"

}

relay_CreateRawRelaySaveBTCHeadTx_11() {
    local req='"method":"relay.CreateRawRelaySaveBTCHeadTx","params":[{"hash":"7b7a4a9b49db5a1162be515d380cd186e98c2bf0bb90f1145485d7c43343fc7c","height":11,"version":536870912,"merkleRoot":"cfa9b66696aea63b7266ffaa1cb4b96c8dd6959eaabf2eb14173f4adaa551f6f","time":1530862108,"nonce":1,"bits":545259519,"previousHash":"5e7d9c599cd040ec2ba53f4dee28028710be8c135e779f65c56feadaae34c3f2","isReset":false}]'
    # echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    # echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signrawtx "$rawtx" "0x4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"

}

query_GetRelayOrderByStatus() {
    status="$1"
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetRelayOrderByStatus","payload":{"addr":"","status":"'"$status"'","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.relayorders[0].id != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

query_GetSellRelayOrder() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetSellRelayOrder","payload":{"addr":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv","status":"pending","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.relayorders[0].status == "pending") and (.result.relayorders[0].coinOperation == 0) and (.result.relayorders[0].id != "")  ' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

query_GetBuyRelayOrder() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBuyRelayOrder","payload":{"addr":"1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY","status":"locking","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    #   echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.relayorders[0].status == "locking")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

query_GetBTCHeaderList() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBTCHeaderList","payload":{"reqHeight":"10","counts":10,"direction":0}}]'
    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.heights|length == 2)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

query_GetBTCHeaderCurHeight() {
    local req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBTCHeaderCurHeight","payload":{"baseHeight":"0"}}]'
    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.baseHeight == "10") and (.result.curHeight == "10")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    local relay_addr=""
    if [ "$ispara" == true ]; then
        relay_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.relay"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
#        chain33_ImportPrivkey "${MAIN_HTTP}" "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588" "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
        relay_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"relay"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "relayaddr=$relay_addr"

    local testAddr="1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3"
    chain33_getMainChainCoins "${MAIN_HTTP}" "$testAddr" 10000000000
    chain33_getParaChainCoins "${MAIN_HTTP}" "$testAddr" 10000000000
    chain33_SendToAddress "$testAddr" "$relay_addr" 8000000000 "${MAIN_HTTP}"

    block_wait 1

}
function run_testcases() {
    relay_CreateRawRelaySaveBTCHeadTx
    query_GetBTCHeaderCurHeight

    relay_CreateRawRelayOrderTx
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
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_testcases

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Relay Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Relay Rpc Test Pass==============${NOC}"
    fi
}

rpc_test "$1"
