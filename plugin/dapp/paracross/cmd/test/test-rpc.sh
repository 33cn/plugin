#!/usr/bin/env bash
# shellcheck disable=SC2128

CASE_ERR=""
UNIT_HTTP=""
IS_PARA=false

# shellcheck source=/dev/null
source ../dapp-test-common.sh

paracross_GetBlock2MainInfo() {
    local height

    height=$(curl -ksd '{"method":"paracross.GetBlock2MainInfo","params":[{"start":1,"end":3}]}' ${UNIT_HTTP} | jq -r ".result.items[1].height")
    [ "$height" -eq 2 ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_SignAndSend() {
    local signedTx
    local sendedTx

    signedTx=$(curl -ksd '{"method":"Chain33.SignRawTx","params":[{"expire":"120s","fee":'"$1"',"privkey":"'"$2"'","txHex":"'"$3"'"}]}' ${UNIT_HTTP} | jq -r ".result")
    #echo "signedTx:$signedTx"
    sendedTx=$(curl -ksd '{"method":"Chain33.SendTransaction","params":[{"data":"'"$signedTx"'"}]}' ${UNIT_HTTP} | jq -r ".result")
    echo "sendedTx:$sendedTx"
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

    req='{"method":"Chain33.GetBalance", "params":[{"addresses" : ["'"$1"'"], "execer" : "'"${exec}"'","asset_exec":"paracross","asset_symbol":"coins.bty"}]}'
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
    #fee 交易费
    #local fee=1000000
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
    ##echo "tx:$tx"
    chain33_SignRawTx "$tx_hash" "$privkey" ${UNIT_HTTP}
    #paracross_SignAndSend $fee "$privkey" "$tx_hash"

    #1. 查询资产转移前余额状态
    para_balance_before=$(paracross_QueryParaBalance "$from_addr" "paracross")
    echo "para before transferring:$para_balance_before"
    main_balance_before=$(paracross_QueryMainBalance "$from_addr")
    echo "main before transferring:$main_balance_before"

    #3  资产从主链转移到平行链
    tx_hash=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"'"$execer_name"'","actionName":"ParacrossAssetTransfer","payload":{"execName":"'"$execer_name"'","to":"'"$from_addr"'","amount":'$amount_should'}}]}' ${UNIT_HTTP} | jq -r ".result")
    #echo "rawTx:$rawTx"
    chain33_SignRawTx "$tx_hash" "$privkey" ${UNIT_HTTP}
    #paracross_SignAndSend $fee "$privkey" "$tx_hash"

    #4 查询转移后余额状态
    local times=100
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
            #echo "para_cross_transfer_withdraw success"
            count=$((count + 1))
            break
        fi
    done

    #5 取钱
    tx_hash=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"'"$execer_name"'","actionName":"ParacrossAssetWithdraw","payload":{"IsWithdraw":true,"execName":"'"$execer_name"'","to":"'"$from_addr"'","amount":'$withdraw_should'}}]}' ${UNIT_HTTP} | jq -r ".result")
    #echo "rawTx:$rawTx"
    chain33_SignRawTx "$tx_hash" "$privkey" ${UNIT_HTTP}
    #paracross_SignAndSend $fee "$privkey" "$tx_hash"

    #6 查询取钱后余额状态
    local times=100
    while true; do
        para_balance_withdraw_after=$(paracross_QueryParaBalance "$from_addr" "paracross")
        echo "para after withdrawing :$para_balance_withdraw_after"
        main_balance_withdraw_after=$(paracross_QueryMainBalance "$from_addr")
        echo "main after withdrawing :$main_balance_withdraw_after"
        #实际取钱金额
        para_withdraw_real=$((para_balance_after - para_balance_withdraw_after))
        main_withdraw_real=$((main_balance_withdraw_after - main_balance_after))
        #echo $withdraw_real
        if [ "$withdraw_should" != "$para_withdraw_real" ] || [ "$withdraw_should" != "$main_withdraw_real" ]; then
            chain33_BlockWait 2 ${UNIT_HTTP}
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                exit 1
            fi
        else
            #echo "para_cross_transfer_withdraw success"
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
    local ok

    if [ "$IS_PARA" == "true" ]; then
        ok=$(curl -ksd '{"method":"paracross.IsSync","params":[]}' ${UNIT_HTTP} | jq -r ".result")
    else
        ok=$(curl -ksd '{"method":"Chain33.IsSync","params":[]}' ${UNIT_HTTP} | jq -r ".result")
    fi

    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_ListTitles() {

    local resp
    local ok
    local main_ip=${UNIT_HTTP//8901/8801}
    resp=$(curl -ksd '{"method":"paracross.ListTitles","params":[]}' ${main_ip})
    echo "$resp"
    ok=$(jq '(.error|not) and (.result| [has("titles"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_GetHeight() {
    local resp
    local ok

    if [ "$IS_PARA" == "true" ]; then
        resp=$(curl -ksd '{"method":"paracross.GetHeight","params":[]}' ${UNIT_HTTP})
        #echo $resp
        ok=$(jq '(.error|not) and (.result| [has("consensHeight"),true])' <<<"$resp")
        [ "$ok" == true ]
        local rst=$?
        echo_rst "$FUNCNAME" "$rst"
    fi
}

function paracross_GetNodeGroupAddrs() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.GetNodeGroupAddrs","params":[{"title":"user.p.para."}]}' ${UNIT_HTTP})
    #echo $resp
    ok=$(jq '(.error|not) and (.result| [has("key","value"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_GetNodeGroupStatus() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.GetNodeGroupStatus","params":[{"title":"user.p.para."}]}' ${UNIT_HTTP})
    #echo $resp
    ok=$(jq '(.error|not) and (.result| [has("status"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_ListNodeGroupStatus() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.ListNodeGroupStatus","params":[{"title":"user.p.para.","status":2}]}' ${UNIT_HTTP})
    #echo $resp
    ok=$(jq '(.error|not) and (.result| [has("status"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_ListNodeStatus() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.ListNodeStatus","params":[{"title":"user.p.para.","status":4}]}' ${UNIT_HTTP})
    #echo $resp
    ok=$(jq '(.error|not) and (.result| [has("status"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

#main chain import pri key
para_test_addr="1MAuE8QSbbech3bVKK2JPJJxYxNtT95oSU"
para_test_prikey="0x24d1fad138be98eebee31440f144aa38c404533f40862995282162bc538e91c8"

function paracross_txgroupex() {
    local amount_transfer=$1
    local amount_trade=$2
    local para_ip=$3

    local paracross_execer_name="user.p.para.paracross"
    local trade_exec_name="user.p.para.trade"

    #  资产从主链转移到平行链
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"'"${paracross_execer_name}"'","actionName":"ParacrossAssetTransfer","payload":{"execName":"'"${paracross_execer_name}"'","to":"'"$para_test_addr"'","amount":'${amount_transfer}'}}]'
    echo "$req"
    resp=$(curl -ksd "{$req}" "${para_ip}")
    # echo "$resp"
    err=$(jq '(.error)' <<<"$resp")
    if [ "$err" != null ]; then
        echo "$resp"
        exit 1
    fi
    tx_hash_asset=$(jq -r ".result" <<<"$resp")
    #    tx_hash_asset=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${paracross_execer_name}"'","actionName":"ParacrossAssetTransfer","payload":{"execName":"'"${paracross_execer_name}"'","to":"'"$para_test_addr"'","amount":'${amount_transfer}'}}]}' "${para_ip}" | jq -r ".result")

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

//测试平行链交易组跨链失败,主链自动恢复原值
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
    //测试跨链过去１个,交易组转账８个失败的场景,主链应该还保持原来的
    local amount_trade=800000000
    local amount_transfer=100000000
    local amount_left=500000000

    left_exec_val=$(paracross_QueryMainBalance "${para_test_addr}")
    if [ "${left_exec_val}" != $amount_left ]; then
        echo "paracross_testTxGroupFail left main paracross failed, get=$left_exec_val,expec=$amount_left"
        exit 1
    fi

    paracross_txgroupex "${amount_transfer}" "${amount_trade}" "${para_ip}"

    local count=0
    local times=100
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

    paracross_txgroupex "${amount_transfer}" "${amount_trade}" "${para_ip}"

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
    req='"method":"paracross.GetHeight","params":[]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    err=$(jq '(.error)' <<<"$resp")
    if [ "$err" != null ]; then
        echo "$resp"
        exit 1
    fi
    chainheight=$(jq '(.result.chainHeight)' <<<"$resp")
    newHeight=$((chainheight + 2000))
    echo "apply stage startHeight=$newHeight"
    req='"method":"Chain33.CreateTransaction","params":[{"execer" : "user.p.para.paracross","actionName" : "selfConsStageConfig","payload" : {"title":"user.p.para.","op" : "1", "stage" : {"startHeight":'"$newHeight"',"enable":2} }}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "$para_test_prikey" "${para_ip}"

    echo "get stage apply id"
    req='"method":"paracross.ListSelfStages","params":[{"status":1,"count":1}]'
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

    req='"method":"Chain33.CreateTransaction","params":[{"execer" : "user.p.para.paracross","actionName" : "selfConsStageConfig","payload":{"title":"user.p.para.","op":"2","vote":{"id":"'"$id"'","value":1} }}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    rawtx=$(jq -r ".result" <<<"$resp")
    chain33_SignRawTx "$rawtx" "$KS_PRI" "${para_ip}"
    chain33_SignRawTx "$rawtx" "$JR_PRI" "${para_ip}"
    chain33_SignRawTx "$rawtx" "$NL_PRI" "${para_ip}"

    echo "query status"
    req='"method":"paracross.ListSelfStages","params":[{"status":3,"count":1}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    ok1=$(jq '(.error|not) and (.result| [has("id"),true])' <<<"$resp")

    req='"method":"paracross.GetSelfConsStages","params":[]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    ok2=$(jq '(.error|not) and (.result| [has("startHeight"),true])' <<<"$resp")

    req='"method":"paracross.GetSelfConsOneStage","params":[{"data":1000}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    ok3=$(jq '(.error|not) and (.result.enable==1)' <<<"$resp")

    req='"method":"paracross.GetSelfConsOneStage","params":[{"data":'"$newHeight"'}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    echo "$resp"
    ok4=$(jq '(.error|not) and (.result.enable==2)' <<<"$resp")
    echo "1=$ok1,2=$ok2,3=$ok3,4=$ok4"

    [ "$ok1" == true ] && [ "$ok2" == true ] && [ "$ok3" == true ] && [ "$ok4" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
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
    paracross_Transfer_Withdraw
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
            run_testcases
        fi
    fi

    chain33_RpcTestRst paracross "$CASE_ERR"
}

chain33_debug_function main "$1" "$2" "$3" "$4"
#main http://127.0.0.1:8801
#main http://47.98.253.127:8801 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b user.p.fzmtest.paracross
