#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
CASE_ERR=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

echo_rst() {
    if [ "$2" == 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="FAIL"
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

retrieve_Backup() {
    echo "========== # retrieve backup begin =========="

    req='"method":"retrieve.CreateRawRetrieveBackupTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo","delayPeriod": 61}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138"
    echo "========== # retrieve backup end =========="

    block_wait 1
}

retrieve_Prepare() {
    echo "========== # retrieve prepare begin =========="

    req='"method":"retrieve.CreateRawRetrievePrepareTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989"
    echo "========== # retrieve prepare end =========="

    block_wait 1
}

retrieve_Perform() {
    echo "========== # retrieve perform begin =========="

    req='"method":"retrieve.CreateRawRetrievePerformTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989"
    echo "========== # retrieve perform end =========="

    block_wait 1
}

retrieve_Cancel() {
    echo "========== # retrieve cancel begin =========="

    req='"method":"retrieve.CreateRawRetrieveCancelTx","params":[{"backupAddr":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX","defaultAddr":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}]'
    tx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")

    reqDecode='"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]'
    data=$(curl -ksd "{$reqDecode}" ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138"
    echo "========== # retrieve cancel end =========="

    block_wait 1
}

retrieve_QueryResult() {
    echo "========== # retrieve query result begin =========="

    local status=$1

    req='"method":"Chain33.Query","params":[{"execer":"retrieve","funcName":"GetRetrieveInfo","payload":{"backupAddress":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX", "defaultAddress":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"}}]'
    data=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")
    ok=$(jq '(.status == '"$status"')' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # retrieve query result end =========="
}

chain33_ImportPrivkey() {
    local pri=$2
    local acc=$3
    local req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"'"$pri"'", "label":"retrieveimportkey1"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$1")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.label=="retrieveimportkey1") and (.result.acc.addr == "'"$acc"'")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

signrawtx() {
    txHex="$1"
    priKey="$2"
    local req='"method":"Chain33.SignRawTx","params":[{"privkey":"'"$priKey"'","txHex":"'"$txHex"'","expire":"120s"}]'
    echo "#request SignRawTx: $req"
    signedTx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")
    echo "signedTx=$signedTx"
    if [ "$signedTx" != null ]; then
        sendTx "$signedTx"
    else
        echo "signedTx null error"
    fi
}

sendTx() {
    signedTx=$1
    local req='"method":"Chain33.SendTransaction","params":[{"token":"BTY","data":"'"$signedTx"'"}]'
    echo "#request sendTx: $req"
    #    curl -ksd "{$req}" ${MAIN_HTTP}
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    err=$(jq '(.error)' <<<"$resp")
    txhash=$(jq -r ".result" <<<"$resp")
    if [ "$err" == null ]; then
        echo "tx hash: $txhash"
        query_tx "$txhash"
    else
        echo "send tx error:$err"
    fi

}

Chain33_SendToAddress() {
    local from="$1"
    local to="$2"
    local amount=$3
    local req='"method":"Chain33.SendToAddress", "params":[{"from":"'"$from"'","to":"'"$to"'", "amount":'"$amount"', "note":"test\n"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.hash|length==66)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    hash=$(jq -r ".result.hash" <<<"$resp")
    query_tx "$hash"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        retrieve_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.retrieve"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        retrieve_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"retrieve"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "retrieveaddr=$retrieve_addr"

    from="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    Chain33_SendToAddress "$from" "$retrieve_addr" 1000000000

    from="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    Chain33_SendToAddress "$from" "$retrieve_addr" 1000000000
    block_wait 1
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
}

function main() {
    MAIN_HTTP="$1"
    echo "=========== # retrieve rpc test ============="
    echo "ip=$MAIN_HTTP"

    init
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============retrieve Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============retrieve Rpc Test Pass==============${NOC}"
    fi
}

main "$1"
