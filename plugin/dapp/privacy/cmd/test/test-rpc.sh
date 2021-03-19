#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -e
set -o pipefail

source ../dapp-test-common.sh

MAIN_HTTP=""

privacy_CreateRawTransaction() {
    req='{"method":"privacy.CreateRawTransaction","params":[{"pubkeypair":"0a9d212b2505aefaa8da370319088bbccfac097b007f52ed71d8133456c8185823c8eac43c5e937953d7b6c8e68b0db1f4f03df4946a29f524875118960a35fb", "assetExec":"coins", "tokenname":"BTY", "actionType":101, "amount":100000000}]}'
    chain33_Http "$req" ${MAIN_HTTP} '.error|not' "$FUNCNAME"
}

privacy_GetPrivacyTxByAddr() {
    chain33_Http '{"method":"privacy.GetPrivacyTxByAddr","params":[{"assetExec":"coins", "assetSymbol":"BTY","sendRecvFlag":0,"address":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", "direction":1, "count":1}]}' ${MAIN_HTTP} '.error|not' "$FUNCNAME"
}

privacy_ShowPrivacyKey() {
    req='{"method":"privacy.ShowPrivacyKey", "params":[{"data":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"}]}'
    resok='(.error|not) and .result.showSuccessful and (.result.pubkeypair=="0a9d212b2505aefaa8da370319088bbccfac097b007f52ed71d8133456c8185823c8eac43c5e937953d7b6c8e68b0db1f4f03df4946a29f524875118960a35fb")'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

privacy_ShowPrivacyAccountInfo() {
    req='{"method":"privacy.ShowPrivacyAccountInfo", "params":[{"addr":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", "assetExec":"coins", "token":"BTY", "displaymode":1}]}'
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result|[has("utxos", "ftxos", "displaymode"), true] | unique | length == 1)' "$FUNCNAME"
}

privacy_ShowPrivacyAccountSpend() {
    chain33_Http '{"method":"privacy.ShowPrivacyAccountSpend", "params":[{"addr":"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", "assetExec":"coins", "token":"BTY"}]}' ${MAIN_HTTP} '(.error|not) and .result.utxoHaveTxHashs' "$FUNCNAME"
}

privacy_RescanUtxos() {
    chain33_Http '{"method":"privacy.RescanUtxos", "params":[{"addrs":["12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"], "flag":0}]}' ${MAIN_HTTP} '(.error|not) and (.result|[has("flag", "repRescanResults"), true] | unique | length == 1)' "$FUNCNAME"
}

privacy_EnablePrivacy() {
    chain33_Http '{"method":"privacy.EnablePrivacy", "params":[{"addrs":["12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"]}]}' ${MAIN_HTTP} '(.error|not) and .result.results[0].IsOK' "$FUNCNAME"
}

function run_test() {
    privacy_EnablePrivacy
    privacy_ShowPrivacyKey
    privacy_CreateRawTransaction
    privacy_ShowPrivacyAccountInfo
    privacy_ShowPrivacyAccountSpend
    privacy_RescanUtxos
    privacy_GetPrivacyTxByAddr
}

function main() {
    chain33_RpcTestBegin privacy
    MAIN_HTTP="$1"
    echo "ip=$MAIN_HTTP"

    run_test
    chain33_RpcTestRst privacy "$CASE_ERR"
}

chain33_debug_function main "$1"
