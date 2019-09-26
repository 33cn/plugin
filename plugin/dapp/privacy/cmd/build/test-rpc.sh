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
    if [ "$2" == true ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="FAIL"
    fi

}

privacy_CreateRawTransaction() {

    local ip=$1
    req='"method":"privacy.CreateRawTransaction","params":[{"pubkeypair":"0a9d212b2505aefaa8da370319088bbccfac097b007f52ed71d8133456c8185823c8eac43c5e937953d7b6c8e68b0db1f4f03df4946a29f524875118960a35fb", "assetExec":"coins", "tokenname":"BTY", "type":1, "amount":100000000}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '.error|not' <<<"$resp")
    echo_rst "$FUNCNAME" "$ok"
}

privacy_GetPrivacyTxByAddr() {

    local ip=$1
    req='"method":"privacy.GetPrivacyTxByAddr","params":[{"tokenname":"BTY","sendRecvFlag":0,"from":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", "direction":1, "count":1}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '.error|not' <<<"$resp")

    echo_rst "$FUNCNAME" "$ok"
}

privacy_ShowPrivacyKey() {

    local ip=$1
    req='"method":"privacy.ShowPrivacyKey", "params":[{"data":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and .result.showSuccessful and (.result.pubkeypair=="0a9d212b2505aefaa8da370319088bbccfac097b007f52ed71d8133456c8185823c8eac43c5e937953d7b6c8e68b0db1f4f03df4946a29f524875118960a35fb")' <<<"$resp")

    echo_rst "$FUNCNAME" "$ok"
}

privacy_ShowPrivacyAccountInfo() {

    local ip=$1
    req='"method":"privacy.ShowPrivacyAccountInfo", "params":[{"addr":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", "token":"BTY", "displaymode":1}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result|[has("utxos", "ftxos", "displaymode"), true] | unique | length == 1)' <<<"$resp")

    echo_rst "$FUNCNAME" "$ok"
}

privacy_ShowPrivacyAccountSpend() {

    local ip=$1
    req='"method":"privacy.ShowPrivacyAccountSpend", "params":[{"addr":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", "token":"BTY"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and .result.utxoHaveTxHashs' <<<"$resp")

    echo_rst "$FUNCNAME" "$ok"
}

privacy_RescanUtxos() {

    local ip=$1
    req='"method":"privacy.RescanUtxos", "params":[{"addrs":["12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"], "flag":0}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result|[has("flag", "repRescanResults"), true] | unique | length == 1)' <<<"$resp")

    echo_rst "$FUNCNAME" "$ok"
}

privacy_EnablePrivacy() {

    local ip=$1
    req='"method":"privacy.EnablePrivacy", "params":[{"addrs":["12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"]}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$ip")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and .result.results[0].IsOK' <<<"$resp")
    echo_rst "$FUNCNAME" "$ok"
}

function run_test() {
    local ip=$1
    privacy_EnablePrivacy "$ip"
    privacy_ShowPrivacyKey "$ip"
    privacy_CreateRawTransaction "$ip"
    privacy_ShowPrivacyAccountInfo "$ip"
    privacy_ShowPrivacyAccountSpend "$ip"
    privacy_RescanUtxos "$ip"
    privacy_GetPrivacyTxByAddr "$ip"

}
function main() {
    MAIN_HTTP="$1"
    echo "=========== # privacy rpc test ============="
    echo "ip=$MAIN_HTTP"

    run_test "$MAIN_HTTP"

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Privacy Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Prviacy Rpc Test Pass==============${NOC}"
    fi
}

main "$1"
