#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
#source ../dapp-test-common.sh
#
#set -x
#
#MAIN_HTTP=""
#chain33SenderAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
## validatorsAddr=["0x92c8b16afd6d423652559c6e266cbe1c29bfd84f", "0x0df9a824699bc5878232c9e612fe1a5346a5a368", "0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1", "0xd9dab021e74ecf475788ed7b61356056b2095830"]
#ethValidatorAddrKeyA="3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
#ethValidatorAddrKeyB="a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
#ethValidatorAddrKeyC="bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
#ethValidatorAddrKeyD="c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"
## 新增地址 chain33 需要导入地址 转入 10 bty当收费费
#chain33Validator1="1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"
#chain33Validator2="155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6"
#chain33Validator3="13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv"
#chain33Validator4="113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG"
#chain33ValidatorKey1="0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
#chain33ValidatorKey2="0x9d539bc5fd084eb7fe86ad631dba9aa086dba38418725c38d9751459f567da66"
#chain33ValidatorKey3="0x0a6671f101e30a2cc2d79d77436b62cdf2664ed33eb631a9c9e3f3dd348a23be"
#chain33ValidatorKey4="0x3818b257b05ee75b6e43ee0e3cfc2d8502342cf67caed533e3756966690b62a5"
#ethReceiverAddr1="0xa4ea64a583f6e51c3799335b28a8f0529570a635"
#ethReceiverAddrKey1="355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"
#ethReceiverAddr2="0x0c05ba5c230fdaa503b53702af1962e08d0c60bf"
#ethReceiverAddrKey2="9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"
#
#maturityDegree=10
#
#function init() {
#    chain33_ImportPrivkey "${chain33ValidatorKey1}" "${chain33Validator1}" "tokenAddr" "${MAIN_HTTP}"
#    chain33_ImportPrivkey "${chain33ValidatorKey2}" "${chain33Validator2}" "tokenAddr" "${MAIN_HTTP}"
#    chain33_ImportPrivkey "${chain33ValidatorKey3}" "${chain33Validator3}" "tokenAddr" "${MAIN_HTTP}"
#    chain33_ImportPrivkey "${chain33ValidatorKey4}" "${chain33Validator4}" "tokenAddr" "${MAIN_HTTP}"
#
#    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"x2ethereum","actionName":"SetConsensusThreshold","payload":{"consensusThreshold":"80", }}]}' ${MAIN_HTTP} | jq -r ".result")
#    chain33_SignAndSendTxWait "$tx" "0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01" ${MAIN_HTTP} "$FUNCNAME"
#}

