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

    local backupaddr=$1
    local defaultaddr=$2
    local delayPeriod=$3

    tx=$(curl -ksd '{"method":"retrieve.CreateRawRetrieveBackupTx","params":[{"backupAddr":"$backupaddr","defaultAddr":"$defaultaddr","delayPeriod": $delayPeriod}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer == "retrieve")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
    echo "========== # retrieve backup end =========="

    block_wait 1
}

retrieve_Prepare() {
    echo "========== # retrieve prepare begin =========="

    local backupaddr=$1
    local defaultaddr=$2

    tx=$(curl -ksd '{"method":"retrieve.CreateRawRetrievePrepareTx","params":[{"backupAddr":"$backupaddr","defaultAddr":"$defaultaddr"}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer == "retrieve")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588"
    echo "========== # retrieve prepare end =========="

    block_wait 1
}

retrieve_Perform() {
    echo "========== # retrieve perform begin =========="

    local backupaddr=$1
    local defaultaddr=$2

    tx=$(curl -ksd '{"method":"retrieve.CreateRawRetrievePerformTx","params":[{"backupAddr":"$backupaddr","defaultAddr":"$defaultaddr"}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer == "retrieve")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588"
    echo "========== # retrieve perform end =========="

    block_wait 1
}

retrieve_Cancel() {
    echo "========== # retrieve cancel begin =========="

    local backupaddr=$1
    local defaultaddr=$2

    tx=$(curl -ksd '{"method":"retrieve.CreateRawRetrieveCancelTx","params":[{"backupAddr":"$backupaddr","defaultAddr":"$defaultaddr"}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer == "retrieve")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    signrawtx "$tx" "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
    echo "========== # retrieve cancel end =========="

    block_wait 1
}

retrieve_QueryResult() {
    echo "========== # retrieve query result begin =========="

    local backupaddr=$1
    local defaultaddr=$2
    local status=$3

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"retrieve","funcName":"GetRetrieveInfo","payload":{"backupAddress":"$backupaddr", "defaultAddress":"defaultaddr"}}]}' ${MAIN_HTTP} | jq -r ".result")
    ok=$(jq '(.status == $3)' <<<"$data")

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
    hash=$(jq '(.result.hash)' <<<"$resp")
    echo "hash=$hash"
    query_tx "$hash"
}

function run_test() {
    retrieve_Backup "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"  61
    retrieve_QueryResult  "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" 1

    retrieve_Prepare "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
    retrieve_QueryResult  "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" 2

    #retrieve_Perform "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
    #retrieve_QueryResult  "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" 3

    retrieve_Cancel "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
    retrieve_QueryResult  "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" 4
}

function main() {
    MAIN_HTTP="$1"
    echo "=========== # retrieve rpc test ============="
    echo "ip=$MAIN_HTTP"

    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============retrieve Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============retrieve Rpc Test Pass==============${NOC}"
    fi
}

main "$1"
