#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

MAIN_HTTP=""
ISSU_ID=""
DEBT_ID=""
COLL_ID=""
BORROW_ID=""
issuance_addr=""
collateralize_addr=""
SystemManager="0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01"
TokenSuperManager="0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc"
TokenAddr="1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"

IssuanceAddr1="1C9t6uNcmbUgebt9HZfKweNb58hUcq5MZY"
IssuancePriv1="0x6bbfe2c8933ad56d244b68f267da576e9df539bcabc160d2ef29acc2838d5d81"
IssuanceAddr2="16pjXn7vMVPqKjuVnYV44ANQGD1TRaw3ct"
IssuancePriv2="0xa099f50ca616017000f338fb00c8eda133b1616f6d62f2ea3e361cc71e6c92d6"
IssuanceAddr3="1CQMn9B5Rh6s8wtnYEhuQwtVxPjcXSC4qC"
IssuancePriv3="0xb4c158903f373636765a0c6226b91980c5a403a37f21bc8248d45c09569e6ad3"
CollateralizeAddr="1BLfkPaAGqSiXyovx3Pm9xUTMHmusLXtLZ"
CollateralizePriv="0xf860db5178e6436cabd499c142084515966f88f581936ee9ccbfc0a6a77c0e70"

# shellcheck source=/dev/null
source ../dapp-test-common.sh

issuance_Create() {
    echo "========== # issuance create begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"issuance","actionName":"IssuanceCreate","payload":{"debtCeiling":1000.1, "totalBalance":10000.1}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv1}" ${MAIN_HTTP}
    ISSU_ID=$RAW_TX_HASH
    chain33_BlockWait 1 ${MAIN_HTTP}

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"issuance","funcName":"IssuanceInfoByID","payload":{"issuanceId":"'"$ISSU_ID"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"issuance","funcName":"IssuanceByStatus","payload":{"status":1}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # issuance create end =========="
}

issuance_Manage() {
    echo "========== # issuance manage begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"issuance","actionName":"IssuanceManage","payload":{"addr":["'"${IssuanceAddr3}"'"]}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv1}" ${MAIN_HTTP}
    chain33_BlockWait 1 ${MAIN_HTTP}

    echo "========== # issuance manage end =========="
}

issuance_Feed() {
    echo "========== # issuance feed begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"issuance","actionName":"IssuancePriceFeed","payload":{"Price":[1], "Volume":[100]}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv2}" ${MAIN_HTTP}
    chain33_BlockWait 1 "${MAIN_HTTP}"

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"issuance","funcName":"IssuancePrice","payload":{}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # issuance feed end =========="
}

issuance_Debt() {
    echo "========== # issuance debt begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"issuance","actionName":"IssuanceDebt","payload":{"issuanceId":"'"${ISSU_ID}"'", "value":10}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv3}" ${MAIN_HTTP}
    DEBT_ID=$RAW_TX_HASH
    chain33_BlockWait 1 "${MAIN_HTTP}"

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"issuance","funcName":"IssuanceRecordByID","payload":{"issuanceId": "'"${ISSU_ID}"'", "debtId": "'"${DEBT_ID}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"issuance","funcName":"IssuanceRecordsByAddr","payload":{"issuanceId": "'"${ISSU_ID}"'", "addr": "'"${IssuanceAddr3}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"issuance","funcName":"IssuanceRecordsByStatus","payload":{"issuanceId": "'"${ISSU_ID}"'", "status": 1}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # issuance debt end =========="
}

issuance_Repay() {
    echo "========== # issuance repay begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"issuance","actionName":"IssuanceRepay","payload":{"issuanceId":"'"${ISSU_ID}"'", "debtId":"'"${DEBT_ID}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv3}" ${MAIN_HTTP}
    chain33_BlockWait 1 "${MAIN_HTTP}"
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"issuance","funcName":"IssuanceRecordsByStatus","payload":{"issuanceId": "'"${ISSU_ID}"'", "status": 6}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # issuance repay end =========="
}

issuance_Close() {
    echo "========== # issuance close begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"issuance","actionName":"IssuanceClose","payload":{"issuanceId":"'"${ISSU_ID}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv1}" ${MAIN_HTTP}
    chain33_BlockWait 1 "${MAIN_HTTP}"
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"issuance","funcName":"IssuanceByStatus","payload":{"status": 2}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # issuance close end =========="
}

collateralize_Manage() {
    echo "========== # collateralize manage begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"collateralize","actionName":"CollateralizeManage","payload":{"debtCeiling":1000.1, "totalBalance":10000.1}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv1}" ${MAIN_HTTP}
    ISSU_ID=$RAW_TX_HASH
    chain33_BlockWait 1 ${MAIN_HTTP}

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeConfig","payload":{}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # collateralize manage end =========="
}

collateralize_Create() {
    echo "========== # collateralize create begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"collateralize","actionName":"CollateralizeCreate","payload":{"debtCeiling":1000.1, "totalBalance":10000.1}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv3}" ${MAIN_HTTP}
    COLL_ID=$RAW_TX_HASH
    chain33_BlockWait 1 ${MAIN_HTTP}

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeInfoByID","payload":{"collateralizeId":"'"${COLL_ID}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeByStatus","payload":{"status":1}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeByAddr","payload":{"addr":"'"${IssuanceAddr3}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # collateralize create end =========="
}

collateralize_Feed() {
    echo "========== # collateralize feed begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"collateralize","actionName":"CollateralizePriceFeed","payload":{"Price":[1], "Volume":[100]}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv2}" ${MAIN_HTTP}
    chain33_BlockWait 1 "${MAIN_HTTP}"

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizePrice","payload":{}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # collateralize feed end =========="
}

collateralize_Borrow() {
    echo "========== # collateralize borrow begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"collateralize","actionName":"CollateralizeBorrow","payload":{"collateralizeId":"'"${COLL_ID}"'", "value":10.1}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${CollateralizePriv}" ${MAIN_HTTP}
    BORROW_ID=$RAW_TX_HASH
    chain33_BlockWait 1 ${MAIN_HTTP}

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeRecordByID","payload":{"collateralizeId":"'"${COLL_ID}"'", "recordId":"'"${BORROW_ID}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeRecordByStatus","payload":{"collateralizeId":"'"${COLL_ID}"'", "status":1}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeRecordByAddr","payload":{"collateralizeId":"'"${COLL_ID}"'", "addr":"'"${CollateralizeAddr}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo_rst "$FUNCNAME" "$?"
    echo "========== # collateralize borrow end =========="
}

collateralize_Append() {
    echo "========== # collateralize append begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"collateralize","actionName":"CollateralizeAppend","payload":{"collateralizeId":"'"${COLL_ID}"'", "recordID":"'"${BORROW_ID}"'", "value":10}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${CollateralizePriv}" ${MAIN_HTTP}
    chain33_BlockWait 1 ${MAIN_HTTP}
    echo "========== # collateralize append end =========="
}

collateralize_Repay() {
    echo "========== # collateralize repay begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"collateralize","actionName":"CollateralizeRepay","payload":{"collateralizeId":"'"${COLL_ID}"'", "recordID":"'"${BORROW_ID}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${CollateralizePriv}" ${MAIN_HTTP}
    chain33_BlockWait 1 ${MAIN_HTTP}

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeRecordByStatus","payload":{"collateralizeId":"'"${COLL_ID}"'", "status":6}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo "========== # collateralize repay end =========="
}

collateralize_Retrieve() {
    echo "========== # collateralize retrieve begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"collateralize","actionName":"CollateralizeRetrieve","payload":{"collateralizeId":"'"${COLL_ID}"'", "balance":100.1}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" "${IssuancePriv3}" ${MAIN_HTTP}
    chain33_BlockWait 1 ${MAIN_HTTP}

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"collateralize","funcName":"CollateralizeRecordByStatus","payload":{"collateralizeId":"'"${COLL_ID}"'", "status":6}}]}' ${MAIN_HTTP} | jq -r ".result")
    [ "$data" != null ]
    echo "========== # collateralize retrieve end =========="
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        issuance_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.issuance"}]}' ${MAIN_HTTP} | jq -r ".result")
        collateralize_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.collateralize"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        issuance_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"issuance"}]}' ${MAIN_HTTP} | jq -r ".result")
        collateralize_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"collateralize"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi

    local main_ip=${MAIN_HTTP//8901/8801}
    #main chain import pri key
    #1C9t6uNcmbUgebt9HZfKweNb58hUcq5MZY
    chain33_ImportPrivkey ${IssuancePriv1} ${IssuanceAddr1} "issuance1" "${main_ip}"
    #16pjXn7vMVPqKjuVnYV44ANQGD1TRaw3ct
    chain33_ImportPrivkey ${IssuancePriv2} ${IssuanceAddr2} "issuance2" "${main_ip}"
    #1CQMn9B5Rh6s8wtnYEhuQwtVxPjcXSC4qC
    chain33_ImportPrivkey ${IssuancePriv3} ${IssuanceAddr3} "issuance3" "${main_ip}"
    #1BLfkPaAGqSiXyovx3Pm9xUTMHmusLXtLZ
    chain33_ImportPrivkey ${CollateralizePriv} ${CollateralizeAddr} "coll" "${main_ip}"

    if [ "$ispara" == false ]; then
        chain33_applyCoins "${IssuanceAddr1}" 12000000000 "${main_ip}"
        chain33_QueryBalance "${IssuanceAddr1}" "$main_ip"

        chain33_applyCoins "${IssuanceAddr2}" 12000000000 "${main_ip}"
        chain33_QueryBalance "${IssuanceAddr2}" "$main_ip"

        chain33_applyCoins "${IssuanceAddr3}" 12000000000 "${main_ip}"
        chain33_QueryBalance "${IssuanceAddr3}" "$main_ip"

        chain33_applyCoins "${CollateralizeAddr}" 12000000000 "${main_ip}"
        chain33_QueryBalance "${CollateralizeAddr}" "$main_ip"

        chain33_applyCoins "${TokenAddr}" 12000000000 "${main_ip}"
        chain33_QueryBalance "${TokenAddr}" "$main_ip"
    else
        # tx fee
        chain33_applyCoins ${IssuanceAddr1} 1000000000 "${main_ip}"
        chain33_QueryBalance ${IssuanceAddr1} "$main_ip"

        chain33_applyCoins "${IssuanceAddr2}" 1000000000 "${main_ip}"
        chain33_QueryBalance "${IssuanceAddr2}" "$main_ip"

        chain33_applyCoins "${IssuanceAddr3}" 1000000000 "${main_ip}"
        chain33_QueryBalance "${IssuanceAddr3}" "$main_ip"

        chain33_applyCoins "${CollateralizeAddr}" 1000000000 "${main_ip}"
        chain33_QueryBalance "${CollateralizeAddr}" "$main_ip"

        chain33_applyCoins "${TokenAddr}" 1000000000 "${main_ip}"
        chain33_QueryBalance "${TokenAddr}" "$main_ip"

        local para_ip="${MAIN_HTTP}"
        #para chain import pri key
        #1C9t6uNcmbUgebt9HZfKweNb58hUcq5MZY
        chain33_ImportPrivkey ${IssuancePriv1} ${IssuanceAddr1} "issuance1" "${para_ip}"
        #16pjXn7vMVPqKjuVnYV44ANQGD1TRaw3ct
        chain33_ImportPrivkey ${IssuancePriv2} ${IssuanceAddr2} "issuance2" "${para_ip}"
        #1CQMn9B5Rh6s8wtnYEhuQwtVxPjcXSC4qC
        chain33_ImportPrivkey ${IssuancePriv3} ${IssuanceAddr3} "issuance3" "${para_ip}"
        #1BLfkPaAGqSiXyovx3Pm9xUTMHmusLXtLZ
        chain33_ImportPrivkey ${CollateralizePriv} ${CollateralizeAddr} "coll" "${para_ip}"

        chain33_applyCoins "${IssuanceAddr3}" 12000000000 "${para_ip}"
        chain33_QueryBalance "${IssuanceAddr3}" "$para_ip"
        chain33_applyCoins "${CollateralizeAddr}" 12000000000 "${para_ip}"
        chain33_QueryBalance "${CollateralizeAddr}" "$para_ip"
    fi

    chain33_SendToAddress "${IssuanceAddr3}" "$issuance_addr" 10000000000 ${MAIN_HTTP}
    chain33_QueryExecBalance "${IssuanceAddr3}" "issuance" "$MAIN_HTTP"
    chain33_BlockWait 1 "${MAIN_HTTP}"

    chain33_SendToAddress "${CollateralizeAddr}" "$collateralize_addr" 10000000000 ${MAIN_HTTP}
    chain33_QueryExecBalance "${CollateralizeAddr}" "collateralize" "$MAIN_HTTP"
    chain33_BlockWait 1 "${MAIN_HTTP}"
}

manage() {
    echo "========== # issuance add issuance-manage begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key": "issuance-manage", "value":"'"${IssuanceAddr1}"'", "op":"add"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" ${SystemManager} ${MAIN_HTTP}
    echo "========== # issuance add issuance-manage end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}

    echo "========== # issuance add issuance-fund begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key": "issuance-fund", "value":"'"${IssuanceAddr1}"'", "op":"add"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" ${SystemManager} ${MAIN_HTTP}
    echo "========== # issuance add issuance-fund end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}

    echo "========== # issuance add issuance-price-feed begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key": "issuance-price-feed", "value":"'"${IssuanceAddr2}"'", "op":"add"}}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "$tx" ${SystemManager} ${MAIN_HTTP}
    echo "========== # issuance add issuance-price-feed end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}

    #    echo "========== # issuance add issuance-guarantor begin =========="
    #    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key": "issuance-guarantor", "value":"'"${IssuanceAddr3}"'", "op":"add"}}]}' ${MAIN_HTTP} | jq -r ".result")
    #
    #    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    #    ok=$(jq '(.execer != "")' <<<"$data")
    #
    #    [ "$ok" == true ]
    #    echo_rst "$FUNCNAME" "$?"
    #
    #    chain33_SignAndSendTx "$tx" ${SystemManager} ${MAIN_HTTP}
    #    echo "========== # issuance add issuance-guarantor end =========="
    #    chain33_BlockWait 1 ${MAIN_HTTP}
}

token() {
    echo "========== # issuance add token begin =========="
    echo "========== # issuance add token token-blacklist begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key": "token-blacklist", "value":"BTY", "op":"add"}}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${TokenSuperManager}" "${MAIN_HTTP}"
    echo "========== # issuance add token token-blacklist end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}

    echo "========== # issuance add token token-finisher begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"manage","actionName":"Modify","payload":{"key": "token-finisher", "value":"'${TokenAddr}'", "op":"add"}}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${TokenSuperManager}" "${MAIN_HTTP}"
    echo "========== # issuance add token token-finisher end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}

    echo "========== # issuance add token precreate begin =========="
    tx=$(curl -ksd '{"method":"token.CreateRawTokenPreCreateTx","params":[{"name": "ccny", "symbol": "CCNY", "total": 10000000000000000, "price": 0, "category": 1,"owner":"'${TokenAddr}'"}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${TokenSuperManager}" "${MAIN_HTTP}"
    echo "========== # issuance add token precreate end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}

    echo "========== # issuance add token finish begin =========="
    tx=$(curl -ksd '{"method":"token.CreateRawTokenFinishTx","params":[{"symbol": "CCNY", "owner":"'${TokenAddr}'"}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${TokenSuperManager}" "${MAIN_HTTP}"
    echo "========== # issuance add token finish end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}

    echo "========== # issuance add token transfer begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer": "token","actionName":"Transfer","payload": {"cointoken":"CCNY", "amount": "10000000000000", "note": "", "to": "'"${IssuanceAddr1}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${TokenSuperManager}" "${MAIN_HTTP}"

    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer": "token","actionName":"Transfer","payload": {"cointoken":"CCNY", "amount": "10000000000000", "note": "", "to": "'"${IssuanceAddr3}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${TokenSuperManager}" "${MAIN_HTTP}"

    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer": "token","actionName":"Transfer","payload": {"cointoken":"CCNY", "amount": "100000000000", "note": "", "to": "'"${CollateralizeAddr}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${TokenSuperManager}" "${MAIN_HTTP}"
    echo "========== # issuance add token transfer end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}

    echo "========== # issuance add token transfer to issuance begin =========="
    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer": "token","actionName":"Transfer","payload": {"cointoken":"CCNY", "amount": "10000000000000", "note": "", "to": "'"${issuance_addr}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${IssuancePriv1}" "${MAIN_HTTP}"

    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer": "token","actionName":"Transfer","payload": {"cointoken":"CCNY", "amount": "10000000000000", "note": "", "to": "'"${collateralize_addr}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${IssuancePriv3}" "${MAIN_HTTP}"

    tx=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer": "token","actionName":"Transfer","payload": {"cointoken":"CCNY", "amount": "100000000000", "note": "", "to": "'"${collateralize_addr}"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    data=$(curl -ksd '{"method":"Chain33.DecodeRawTransaction","params":[{"txHex":"'"$tx"'"}]}' ${MAIN_HTTP} | jq -r ".result.txs[0]")
    ok=$(jq '(.execer != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    chain33_SignAndSendTx "${tx}" "${CollateralizePriv}" "${MAIN_HTTP}"
    echo "========== # issuance add token transfer to issuance end =========="
    chain33_BlockWait 1 ${MAIN_HTTP}
}

function issuance_test() {
    issuance_Create

    issuance_Manage

    issuance_Feed

    issuance_Debt

    issuance_Repay

    issuance_Close
}

function collateralize_test() {
    collateralize_Manage

    collateralize_Create

    collateralize_Feed

    collateralize_Borrow

    collateralize_Append

    collateralize_Repay

    collateralize_Retrieve
}

function main() {
    chain33_RpcTestBegin "issuance & collateralize"
    MAIN_HTTP="$1"
    echo "ip=$MAIN_HTTP"

    init
    manage
    token
    issuance_test
    collateralize_test

    chain33_RpcTestRst "issuance & collateralize" "$CASE_ERR"
}

chain33_debug_function main "$1"
