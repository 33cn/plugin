#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null

UNIT_HTTP=""
IS_PARA=false

source ../dapp-test-common.sh

paracross_GetBlock2MainInfo() {
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName": "GetBlock2MainInfo", "payload" : {"start":1,"end":3}}]}' ${UNIT_HTTP} '(.result.items[1].height == "2")' "$FUNCNAME"
}

function paracross_QueryParaBalance() {
    local req
    local resp
    local balance
    local ip_http
    local para_http

    ip_http=${UNIT_HTTP%:*}
    para_http="$ip_http:8901"
    local exec=$2
    local symbol="coins.bty"
    if [ -n "$3" ]; then
        symbol="$3"
    fi

    req='{"method":"Chain33.GetBalance", "params":[{"addresses" : ["'"$1"'"], "execer" : "'"${exec}"'","asset_exec":"paracross","asset_symbol":"'"${symbol}"'"}]}'
    resp=$(curl -ksd "$req" "${para_http}")
    balance=$(jq -r '.result[0].balance' <<<"$resp")
    echo "$balance"
    return $?
}

function paracross_QueryMainBalance() {
    local req
    local resp
    local balance
    local ip_http
    local main_http

    ip_http=${UNIT_HTTP%:*}
    main_http="$ip_http:8801"

    req='{"method":"Chain33.GetBalance", "params":[{"addresses" : ["'"$1"'"], "execer" : "paracross"}]}'
    resp=$(curl -ksd "$req" "${main_http}")
    balance=$(jq -r '.result[0].balance' <<<"$resp")
    echo "$balance"
    return $?
}

function paracross_QueryMainAssetBalance() {
    local req
    local resp
    local balance
    local ip_http
    local main_http

    ip_http=${UNIT_HTTP%:*}
    main_http="$ip_http:8801"
    local exec=$2
    local symbol="bty"
    if [ -n "$3" ]; then
        symbol="$3"
    fi

    req='{"method":"Chain33.GetBalance", "params":[{"addresses" : ["'"$1"'"], "execer" : "'"${exec}"'","asset_exec":"paracross","asset_symbol":"'"${symbol}"'"}]}'
    resp=$(curl -ksd "$req" "${main_http}")
    balance=$(jq -r '.result[0].balance' <<<"$resp")
    echo "$balance"
    return $?
}

function paracross_Transfer_Withdraw_Inner() {
    # 计数器，资产转移操作和取钱操作都成功才算成功，也就是 counter == 2
    local count=0
    #fromAddr  跨链资产转移地址
    local from_addr="$1"
    #privkey 地址签名
    local privkey="$2"
    #paracrossAddr 合约地址
    local paracross_addr="$3"
    #标题
    local execer_name="$4"
    #amount_save 存钱到合约地址
    local amount_save=1000000
    #amount_should 应转移金额
    local amount_should=27000
    #withdraw_should 应取款金额
    local withdraw_should=13000
    #平行链转移前余额
    local para_balance_before
    #平行链转移后余额
    local para_balance_after
    #平行链取钱后余额
    local para_balance_withdraw_after
    #主链转移前余额
    local main_balance_before
    #主链转移后余额
    local main_balance_after
    #主链取钱后余额
    local main_balance_withdraw_after

    #构造交易哈希
    local tx_hash
    #平行链实际转移金额
    local para_amount_real
    #平行链实际取钱金额
    local para_withdraw_real
    #主链实际转移金额
    local main_amount_real
    #主链实际取钱金额
    local main_withdraw_real

    #2  存钱到合约地址
    tx_hash=$(curl -ksd '{"method":"Chain33.CreateRawTransaction","params":[{"to":"'"$paracross_addr"'","amount":'$amount_save'}]}' ${UNIT_HTTP} | jq -r ".result")
    chain33_SignAndSendTx "$tx_hash" "$privkey" ${UNIT_HTTP}

    #1. 查询资产转移前余额状态
    para_balance_before=$(paracross_QueryParaBalance "$from_addr" "paracross")
    echo "para before transferring:$para_balance_before"
    main_balance_before=$(paracross_QueryMainBalance "$from_addr")
    echo "main before transferring:$main_balance_before"

    #3  资产从主链转移到平行链
    tx_hash=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"'"$execer_name"'","actionName":"ParacrossAssetTransfer","payload":{"execName":"'"$execer_name"'","to":"'"$from_addr"'","amount":'$amount_should'}}]}' ${UNIT_HTTP} | jq -r ".result")
    chain33_SignAndSendTx "$tx_hash" "$privkey" ${UNIT_HTTP}

    #4 查询转移后余额状态
    local times=200
    while true; do
        para_balance_after=$(paracross_QueryParaBalance "$from_addr" "paracross")
        echo "para after transferring:$para_balance_after"
        main_balance_after=$(paracross_QueryMainBalance "$from_addr")
        echo "main after transferring:$main_balance_after"
        #real_amount  实际转移金额
        para_amount_real=$((para_balance_after - para_balance_before))
        main_amount_real=$((main_balance_before - main_balance_after))
        #echo $amount_real
        if [ "$para_amount_real" != "$amount_should" ] || [ "$main_amount_real" != "$amount_should" ]; then
            chain33_BlockWait 2 ${UNIT_HTTP}
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                exit 1
            fi
        else
            count=$((count + 1))
            break
        fi
    done

    #5 取钱
    tx_hash=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"'"$execer_name"'","actionName":"ParacrossAssetWithdraw","payload":{"IsWithdraw":true,"execName":"'"$execer_name"'","to":"'"$from_addr"'","amount":'$withdraw_should'}}]}' ${UNIT_HTTP} | jq -r ".result")
    chain33_SignAndSendTx "$tx_hash" "$privkey" ${UNIT_HTTP}

    #6 查询取钱后余额状态
    local times=200
    while true; do
        para_balance_withdraw_after=$(paracross_QueryParaBalance "$from_addr" "paracross")
        echo "para after withdrawing :$para_balance_withdraw_after"
        main_balance_withdraw_after=$(paracross_QueryMainBalance "$from_addr")
        echo "main after withdrawing :$main_balance_withdraw_after"
        #实际取钱金额
        para_withdraw_real=$((para_balance_after - para_balance_withdraw_after))
        main_withdraw_real=$((main_balance_withdraw_after - main_balance_after))
        if [ "$withdraw_should" != "$para_withdraw_real" ] && [ "$withdraw_should" != "$main_withdraw_real" ]; then
            chain33_BlockWait 2 ${UNIT_HTTP}
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                exit 1
            fi
        else
            count=$((count + 1))
            break
        fi
    done

    [ "$count" -eq 2 ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

function paracross_Transfer_Withdraw() {
    #fromAddr  跨链资产转移地址
    local from_addr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    #privkey 地址签名
    local privkey="4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
    #paracrossAddr 合约地址
    local paracross_addr="1HPkPopVe3ERfvaAgedDtJQ792taZFEHCe"
    #execer
    local execer_name="user.p.para.paracross"

    paracross_Transfer_Withdraw_Inner "$from_addr" "$privkey" "$paracross_addr" "$execer_name"
}

function paracross_IsSync() {
    if [ "$IS_PARA" == "true" ]; then
        req='{"method":"paracross.IsSync","params":[]}'
    else
        req='{"method":"Chain33.IsSync","params":[]}'
    fi
    chain33_Http "$req" ${UNIT_HTTP} '(.error|not)' "$FUNCNAME"
}

function paracross_ListTitles() {
    local main_ip=${UNIT_HTTP//8901/8801}
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName": "ListTitles", "payload" : {}}]}' ${main_ip} '(.error|not) and (.result| [has("titles"),true])' "$FUNCNAME"
}

function paracross_GetHeight() {
    if [ "$IS_PARA" == "true" ]; then
        chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName": "GetHeight", "payload" : {}}]}' ${UNIT_HTTP} '(.error|not) and (.result| [has("consensHeight"),true])' "$FUNCNAME"
    fi
}

function paracross_GetNodeGroupAddrs() {
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetNodeGroupAddrs","payload":{"title":"user.p.para."}}]}' ${UNIT_HTTP} '(.error|not) and (.result| [has("key","value"),true])' "$FUNCNAME"
}

function paracross_GetNodeGroupStatus() {
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetNodeGroupStatus","payload":{"title":"user.p.para."}}]}' ${UNIT_HTTP} '(.error|not) and (.result| [has("status"),true])' "$FUNCNAME"
}

function paracross_ListNodeGroupStatus() {
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"ListNodeGroupStatus","payload":{"title":"user.p.para.","status":2}}]}' ${UNIT_HTTP} '(.error|not) and (.result| [has("status"),true])' "$FUNCNAME"
}

function paracross_ListNodeStatus() {
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"ListNodeStatusInfo","payload":{"title":"user.p.para.","status":3}}]}' ${UNIT_HTTP} '(.error|not) and (.result| [has("status"),true])' "$FUNCNAME"
}

function paracross_GetSupervisionInfo() {
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetSupervisionNodeGroupAddrs","payload":{"title":"user.p.para."}}]}' ${UNIT_HTTP} '(.error|not) and (.result| [has("key","value"),true])' "GetSupervisionNodeGroupAddrs"
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"ListSupervisionNodeStatusInfo","payload":{"title":"user.p.para.","status":0}}]}' ${UNIT_HTTP} '(.error|not) and (.result| [has("status"),true])' "ListSupervisionNodeStatusInfo status:0"
}

para_test_addr="1MAuE8QSbbech3bVKK2JPJJxYxNtT95oSU"
para_test_prikey="0x24d1fad138be98eebee31440f144aa38c404533f40862995282162bc538e91c8"

function paracross_txgroupex() {
    local amount_transfer=$1
    local amount_trade=$2
    local para_ip=$3
    local para_title=$4

    local coins_exec="$5"
    local bty_symbol="$6"
    local paracross_execer_name="$para_title.paracross"
    local trade_exec_name="$para_title.trade"

    #  资产从主链转移到平行链
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${paracross_execer_name}"'","actionName":"CrossAssetTransfer","payload":{"assetExec":"'"${coins_exec}"'","assetSymbol":"'"${bty_symbol}"'","toAddr":"'"${para_test_addr}"'","amount":'${amount_transfer}'}}]'
    echo "$req"
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    err=$(jq '(.error)' <<<"$resp")
    if [ "$err" != null ]; then
        echo "$resp"
        exit 1
    fi
    tx_hash_asset=$(jq -r ".result" <<<"$resp")

    #  资产从平行链转移到平行链合约
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${paracross_execer_name}"'","actionName":"TransferToExec","payload":{"execName":"'"${paracross_execer_name}"'","to":"'"${trade_exec_addr}"'","amount":'${amount_trade}', "cointoken":"coins.bty"}}]'
    echo "$req"
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    err=$(jq '(.error)' <<<"$resp")
    if [ "$err" != null ]; then
        echo "$resp"
        exit 1
    fi
    tx_hash_transferExec=$(jq -r ".result" <<<"$resp")

    #create tx group with none
    req='"method":"Chain33.CreateNoBlanaceTxs","params":[{"txHexs":["'"${tx_hash_asset}"'","'"${tx_hash_transferExec}"'"],"privkey":"'"${para_test_prikey}"'","expire":"120s"}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    err=$(jq '(.error)' <<<"$resp")
    if [ "$err" != null ]; then
        echo "$resp"
        exit 1
    fi
    tx_hash_group=$(jq -r ".result" <<<"$resp")

    #sign 1
    tx_sign=$(curl -ksd '{"method":"Chain33.SignRawTx","params":[{"privkey":"'"$para_test_prikey"'","txHex":"'"$tx_hash_group"'","index":2,"expire":"120s"}]}' "${para_ip}" | jq -r ".result")
    #sign 2
    tx_sign2=$(curl -ksd '{"method":"Chain33.SignRawTx","params":[{"privkey":"'"$para_test_prikey"'","txHex":"'"$tx_sign"'","index":3,"expire":"120s"}]}' "${para_ip}" | jq -r ".result")

    #send
    chain33_SendTx "${tx_sign2}" "${para_ip}"
}

#测试平行链交易组跨链失败,主链自动恢复原值
function paracross_testTxGroupFail() {
    local para_ip=$1

    ispara=$(echo '"'"${para_ip}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    local paracross_addr=""
    local main_ip=${para_ip//8901/8801}

    paracross_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"paracross"}]}' "${main_ip}" | jq -r ".result")
    echo "paracross_addr=$paracross_addr"

    #execer
    local trade_exec_addr="12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13"
    #测试跨链过去１个,交易组转账８个失败的场景,主链应该还保持原来的
    local amount_trade=800000000
    local amount_transfer=100000000
    local amount_left=500000000

    #当前为５个
    left_exec_val=$(paracross_QueryMainBalance "${para_test_addr}")
    if [ "${left_exec_val}" != $amount_left ]; then
        echo "paracross_testTxGroupFail left main paracross failed, get=$left_exec_val,expec=$amount_left"
        exit 1
    fi

    paracross_txgroupex "${amount_transfer}" "${amount_trade}" "${para_ip}" "user.p.para" "coins" "bty"

    #跨链失败后仍应该有５个，之前transfer到trade的２个应该保持不变
    local count=0
    local times=300
    local paracross_execer_name="user.p.para.paracross"
    local trade_exec_name="user.p.para.trade"
    local transfer_expect="200000000"
    local exec_expect="100000000"
    while true; do
        transfer_val=$(paracross_QueryParaBalance "${para_test_addr}" "$paracross_execer_name")
        transfer_exec_val=$(paracross_QueryParaBalance "${para_test_addr}" "$trade_exec_name")
        left_exec_val=$(paracross_QueryMainBalance "${para_test_addr}")
        if [ "${left_exec_val}" != $amount_left ] || [ "${transfer_val}" != $transfer_expect ] || [ "${transfer_exec_val}" != $exec_expect ]; then
            echo "trans=${transfer_val}-expect=${transfer_expect},trader=${transfer_exec_val}-expect=${exec_expect},left=${left_exec_val}-expect=${amount_left}"
            chain33_BlockWait 2 ${UNIT_HTTP}
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_testfail failed"
                exit 1
            fi
            echo "paracross_testTxGroupFail left main paracross failed, get=$left_exec_val,expec=$amount_left"
        else
            count=$((count + 1))
            break
        fi
    done
    [ "$count" -eq 1 ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

#测试paraAssetWithdraw fail, 但是目前拿不到game　平行链ip构建平行链交易，可以在主链上构建不包含CreateNoBlanaceTxs的交易组来测试，不然需要在平行链上构建
function paracross_testParaAssetWithdrawFail() {
    local para_ip=$1

    ispara=$(echo '"'"${para_ip}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    local paracross_addr=""
    local main_ip=${para_ip//8901/8801}

    local game_token_test_addr="1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu"

    #execer
    local trade_exec_addr="12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13"
    #测试跨链过去１0个,交易组转账8000个失败的场景,主链应该还保持原来的
    local amount_trade=800000000000
    local amount_transfer=1000000000
    local amount_left=10000000000

    #当前为５个
    left_exec_val=$(paracross_QueryMainAssetBalance "${game_token_test_addr}" "paracross") "user.p.game.coins.para"
    if [ "${left_exec_val}" != $amount_left ]; then
        echo "paracross_testTxGroupFail left main paracross failed, get=$left_exec_val,expec=$amount_left"
        exit 1
    fi

    paracross_txgroupex "${amount_transfer}" "${amount_trade}" "${para_ip}" "user.p.game" "paracross" "user.p.game.coins.para"

    #跨链失败后仍应该有５个，之前transfer到trade的２个应该保持不变
    local count=0
    local times=300
    while true; do
        left_exec_val=$(paracross_QueryMainAssetBalance "${game_token_test_addr}" "paracross" "user.p.game.coins.para")
        if [ "${left_exec_val}" != $amount_left ]; then
            echo "left=${left_exec_val}-expect=${amount_left}"
            chain33_BlockWait 2 ${UNIT_HTTP}
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_testfail failed"
                exit 1
            fi
            echo "paracross_testTxGroupFail left main paracross failed, get=$left_exec_val,expec=$amount_left"
        else
            count=$((count + 1))
            break
        fi
    done
    [ "$count" -eq 1 ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_testTxGroup() {
    local para_ip=$1

    ispara=$(echo '"'"${para_ip}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    local paracross_addr=""
    local main_ip=${para_ip//8901/8801}

    paracross_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"paracross"}]}' "${main_ip}" | jq -r ".result")
    echo "paracross_addr=$paracross_addr"

    #execer
    local paracross_execer_name="user.p.para.paracross"
    local trade_exec_name="user.p.para.trade"
    local trade_exec_addr="12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13"
    local amount_trade=100000000
    local amount_deposit=800000000
    local amount_transfer=300000000
    local amount_left=500000000
    chain33_ImportPrivkey "${para_test_prikey}" "${para_test_addr}" "paracross-transfer6" "${main_ip}"

    # tx fee + transfer 10 coins
    chain33_applyCoins "${para_test_addr}" 1000000000 "${main_ip}"
    chain33_QueryBalance "${para_test_addr}" "$main_ip"

    #deposit 8 coins to paracross
    chain33_SendToAddress "${para_test_addr}" "$paracross_addr" "$amount_deposit" "${main_ip}"
    chain33_QueryExecBalance "${para_test_addr}" "paracross" "${main_ip}"

    paracross_txgroupex "${amount_transfer}" "${amount_trade}" "${para_ip}" "user.p.para" "coins" "bty"

    local transfer_expect="200000000"
    local exec_expect="100000000"
    transfer_val=$(paracross_QueryParaBalance "${para_test_addr}" "$paracross_execer_name")
    transfer_exec_val=$(paracross_QueryParaBalance "${para_test_addr}" "$trade_exec_name")
    left_exec_val=$(paracross_QueryMainBalance "${para_test_addr}")
    if [ "${transfer_val}" != $transfer_expect ]; then
        echo "paracross_testTxGroup trasfer failed, get=$transfer_val,expec=$transfer_expect"
        exit 1
    fi
    if [ "${transfer_exec_val}" != $exec_expect ]; then
        echo "paracross_testTxGroup toexec failed, get=$transfer_exec_val,expec=$exec_expect"
        exit 1
    fi
    if [ "${left_exec_val}" != $amount_left ]; then
        echo "paracross_testTxGroup left main paracross failed, get=$left_exec_val,expec=$amount_left"
        exit 1
    fi

    echo_rst "$FUNCNAME" 0

}

paracross_testSelfConsensStages() {
    local para_ip=$1
    req='"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName": "GetHeight", "payload" : {}}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    err=$(jq '(.error)' <<<"$resp")
    if [ "$err" != null ]; then
        echo "$resp"
        exit 1
    fi
    chainheight=$(jq -r '(.result.chainHeight)' <<<"$resp")
    newHeight=$((chainheight + 2000))
    echo "1. apply stage startHeight=$newHeight"
    req='"method":"Chain33.CreateTransaction","params":[{"execer" : "user.p.para.paracross","actionName" : "SelfStageConfig","payload" : {"title":"user.p.para.","ty" : "1", "stage" : {"startHeight":'"$newHeight"',"enable":2} }}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignAndSendTx "$rawtx" "$para_test_prikey" "${para_ip}"

    echo "2. get stage apply id"
    req='"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"ListSelfStages","payload":{"status":1,"count":1}}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    id=$(jq -r ".result.stageInfo[0].id" <<<"$resp")
    if [ -z "$id" ]; then
        echo "paracross stage apply id null"
        exit 1
    fi

    echo "vote id"
    KS_PRI="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    JR_PRI="0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
    NL_PRI="0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"

    req='"method":"Chain33.CreateTransaction","params":[{"execer" : "user.p.para.paracross","actionName" : "SelfStageConfig","payload":{"title":"user.p.para.","ty":"2","vote":{"id":"'"$id"'","value":1} }}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    rawtx=$(jq -r ".result" <<<"$resp")
    echo "send vote 1"
    chain33_SignAndSendTx "$rawtx" "$KS_PRI" "${para_ip}"
    echo "send vote 2"
    chain33_SignAndSendTx "$rawtx" "$JR_PRI" "${para_ip}" "130s"
    echo "send vote 3"
    chain33_SignAndSendTx "$rawtx" "$NL_PRI" "${para_ip}" "140s"

    echo "query status"
    req='"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"ListSelfStages","payload":{"status":3,"count":1}}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    ok1=$(jq '(.error|not) and (.result| [has("id"),true])' <<<"$resp")

    req='"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetSelfConsStages","payload":{}}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    ok2=$(jq '(.error|not) and (.result| [has("startHeight"),true])' <<<"$resp")

    req='"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetSelfConsOneStage","payload":{"data":1000}}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    ok3=$(jq '(.error|not) and (.result.enable==1)' <<<"$resp")

    req='"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetSelfConsOneStage","payload":{"data":'"$newHeight"'}}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    ok4=$(jq '(.error|not) and (.result.enable==2)' <<<"$resp")
    echo "1=$ok1,2=$ok2,3=$ok3,4=$ok4"

    [ "$ok1" == true ] && [ "$ok2" == true ] && [ "$ok3" == true ] && [ "$ok4" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

addr1q9="1Q9sQwothzM1gKSzkVZ8Dt1tqKX1uzSagx"
priv1q9="0x1c3e6cac2f887e1ab9180e2d5772dc4ba01accb8d4df434faba097003eb35482"

paracross_testBind() {
    local para_ip=$1
    echo "bind miner"
    echo "1. create tx"
    req='"method":"Chain33.CreateTransaction","params":[{"execer" : "user.p.para.paracross","actionName" : "ParaBindMiner","payload" : {"bindAction":"1","bindCoins":5, "targetNode":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"}}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignAndSendTxWait "$rawtx" "${priv1q9}" "${para_ip}"

    echo "2. get bind"
    #    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetNodeBindMinerList","payload":{"Node":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"}}]}' "${para_ip}" '(.error|not) and (.result.List| [has("1KSBd17H7Z"),true])' "$FUNCNAME" '(.result.List)'
    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetNodeBindMinerList","payload":{"node":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"}}]}' "${para_ip}" '(.error|not) and (.result.List| [has("1Q9sQw"),true])' "$FUNCNAME" '(.result.List)'
}

paracross_testUnBind() {
    local para_ip=$1
    echo "unBind miner"
    echo "1. create tx"
    req='"method":"Chain33.CreateTransaction","params":[{"execer" : "user.p.para.paracross","actionName" : "ParaBindMiner","payload" : {"bindAction":"2", "targetNode" : "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"}}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignAndSendTxWait "$rawtx" "${priv1q9}" "${para_ip}"

    echo "2. get bind"
    #    req='"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetNodeBindMinerList","payload":{"data":$nodeAddr}]'
    #    resp=$(curl -ksd "{$req}" "${para_ip}")
    #    echo "$resp"
    #    superNode=$(jq -r ".result.List.SuperNode" <<<"$resp")
    #    miners=$(jq -r ".result.List.Miners" <<<"$resp")

    chain33_Http '{"method":"Chain33.Query","params":[{ "execer":"paracross", "funcName":"GetNodeBindMinerList","payload":{"node":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4","miner":"1Q9sQwothzM1gKSzkVZ8Dt1tqKX1uzSagx"}}]}' "${para_ip}" '(.error|not) and (.result.List| [has("1Q9sQw"),true])' "$FUNCNAME" '(.result.List)'
}

paracross_testBindMiner() {
    #bind node
    paracross_testBind "$1"
    #unbind
    paracross_testUnBind "$1"
}

function apply_coins() {
    local main_ip=${UNIT_HTTP//8901/8801}

    chain33_applyCoins "${addr1q9}" 1000000000 "${main_ip}"
    chain33_QueryBalance "${addr1q9}" "$main_ip"

    local para_ip="${UNIT_HTTP}"

    chain33_applyCoins "${addr1q9}" 1000000000 "${para_ip}"
    chain33_QueryBalance "${addr1q9}" "$para_ip"

    chain33_ImportPrivkey "$priv1q9" "$addr1q9" "bindminer" "$para_ip"

    local para_exec_addr=""
    para_exec_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.paracross"}]}' ${para_ip} | jq -r ".result")
    chain33_SendToAddress "$addr1q9" "${para_exec_addr}" 900000000 "${para_ip}"
    chain33_QueryExecBalance "${addr1q9}" "user.p.para.paracross" "$para_ip"
}

function run_testcases() {
    paracross_GetBlock2MainInfo
    paracross_IsSync
    paracross_GetHeight
    paracross_ListTitles
    paracross_GetNodeGroupAddrs
    paracross_GetNodeGroupStatus
    paracross_ListNodeGroupStatus
    paracross_ListNodeStatus
    paracross_GetSupervisionInfo
    paracross_Transfer_Withdraw
    paracross_testBindMiner "$UNIT_HTTP"

    paracross_testTxGroup "$UNIT_HTTP"
    paracross_testTxGroupFail "$UNIT_HTTP"
    paracross_testSelfConsensStages "$UNIT_HTTP"
}

function main() {
    chain33_RpcTestBegin paracross
    UNIT_HTTP=$1
    IS_PARA=$(echo '"'"${UNIT_HTTP}"'"' | jq '.|contains("8901")')

    if [ $# -eq 4 ] && [ -n "$2" ] && [ -n "$3" ] && [ -n "$4" ]; then
        #fromAddr  跨链资产转移地址
        local from_addr="$2"
        #privkey 地址签名
        local privkey="$3"
        #execer
        local execer_name="$4"
        #paracrossAddr 合约地址
        local paracross_addr="1HPkPopVe3ERfvaAgedDtJQ792taZFEHCe"

        echo "=========== # start cross transfer monitor ============="
        while true; do
            paracross_Transfer_Withdraw_Inner "$from_addr" "$privkey" "$paracross_addr" "$execer_name"
            chain33_BlockWait 1 "${UNIT_HTTP}"
        done
    else
        if [ "$IS_PARA" == "true" ]; then
            echo "=========== # paracross rpc test ============="
            apply_coins
            run_testcases
        fi
    fi

    chain33_RpcTestRst paracross "$CASE_ERR"
}

chain33_debug_function main "$1" "$2" "$3" "$4"
#main http://127.0.0.1:8801
#main http://47.98.253.127:8801 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b user.p.fzmtest.paracross
