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

    resp=$(curl -ksd '{"method":"paracross.ListTitles","params":[]}' ${UNIT_HTTP})
    #echo $resp
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

function paracross_testTxGroup() {
    local para_ip=$1

    ispara=$(echo '"'"${para_ip}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    local paracross_addr=""
    local main_ip=${para_ip//8901/8801}

    paracross_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"paracross"}]}' "${main_ip}" | jq -r ".result")
    echo "paracross_addr=$paracross_addr"

    #main chain import pri key
    local test_addr="1MAuE8QSbbech3bVKK2JPJJxYxNtT95oSU"
    local test_prikey="0x24d1fad138be98eebee31440f144aa38c404533f40862995282162bc538e91c8"
    #execer
    local paracross_execer_name="user.p.para.paracross"
    local trade_exec_name="user.p.para.trade"
    local trade_exec_addr="12bihjzbaYWjcpDiiy9SuAWeqNksQdiN13"
    local amount_trade=100000000
    local amount_transfer=800000000
    chain33_ImportPrivkey "${test_prikey}" "${test_addr}" "paracross-transfer6" "${main_ip}"

    # tx fee + transfer 10 coins
    chain33_applyCoins "${test_addr}" 1000000000 "${main_ip}"
    chain33_QueryBalance "${test_addr}" "$main_ip"

    #deposit 8 coins to paracross
    chain33_SendToAddress "${test_addr}" "$paracross_addr" "$amount_transfer" "${main_ip}"
    chain33_QueryExecBalance "${test_addr}" "paracross" "${main_ip}"

    #  资产从主链转移到平行链
    tx_hash_asset=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${paracross_execer_name}"'","actionName":"ParacrossAssetTransfer","payload":{"execName":"'"${paracross_execer_name}"'","to":"'"$test_addr"'","amount":'${amount_transfer}'}}]}' "${para_ip}" | jq -r ".result")
    #curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"user.p.para.paracross","actionName":"ParacrossAssetTransfer","payload":{"execName":"user.p.para.paracross","to":"1MAuE8QSbbech3bVKK2JPJJxYxNtT95oSU","amount":100000000}}]}' http://172.20.0.5:8901

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
    req='"method":"Chain33.CreateNoBlanaceTxs","params":[{"txHexs":["'"${tx_hash_asset}"'","'"${tx_hash_transferExec}"'"],"privkey":"'"${test_prikey}"'","expire":"120s"}]'
    resp=$(curl -ksd "{$req}" "${para_ip}")
    err=$(jq '(.error)' <<<"$resp")
    if [ "$err" != null ]; then
        echo "$resp"
        exit 1
    fi
    tx_hash_group=$(jq -r ".result" <<<"$resp")

    #sign 1
    tx_sign=$(curl -ksd '{"method":"Chain33.SignRawTx","params":[{"privkey":"'"$test_prikey"'","txHex":"'"$tx_hash_group"'","index":2,"expire":"120s"}]}' "${para_ip}" | jq -r ".result")
    #curl -ksd '{"method":"Chain33.SignRawTx","params":[{"privkey":"0x24d1fad138be98eebee31440f144aa38c404533f40862995282162bc538e91c8","txHex":"0a10757365722e702e706172612e6e6f6e6512126e6f2d6665652d7472616e73616374696f6e1a6e080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a473045022100fe7763b2fa5b42eddccf1a3434cb6d6bb60a5c5c32e5498219e99be01bb94ad302201ecb9931a2bb4e1b0d49ec50ee552a774c1db8a1eb9b2dff47e4b931625e3af220e0a71230a89780b2a5b187990a3a2231466a58697076505142754e6f7a4133506150724b6846703854343166717141707640034a8c050a8e020a10757365722e702e706172612e6e6f6e6512126e6f2d6665652d7472616e73616374696f6e1a6e080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a473045022100fe7763b2fa5b42eddccf1a3434cb6d6bb60a5c5c32e5498219e99be01bb94ad302201ecb9931a2bb4e1b0d49ec50ee552a774c1db8a1eb9b2dff47e4b931625e3af220e0a71230a89780b2a5b187990a3a2231466a58697076505142754e6f7a4133506150724b6846703854343166717141707640034a2054ba4451fb8f226dd54b52ec086f4eaa4990d66876899b1badec8ce96ef55208522000f2e5970b5b1854da143f4f3e91bf7eb87d1d62869ef08a2ce3b109994ec3650abb010a15757365722e702e706172612e7061726163726f7373122e10904e22291080c2d72f2222314d41754538515362626563683362564b4b324a504a4a7859784e745439356f535530d195faf7d3a1ec9a4c3a223139574a4a7639366e4b4155347348465771476d7371666a786433376a617a71696940034a2054ba4451fb8f226dd54b52ec086f4eaa4990d66876899b1badec8ce96ef5520852208fb7939bf2701e5af9cef05da33020be682e10c84fde9bd8d24765031fad530b0aba010a15757365722e702e706172612e7061726163726f7373124f1004424b0a09636f696e732e6274791080ade2042215757365722e702e706172612e7061726163726f73732a2231326269686a7a626159576a637044696979395375415765714e6b735164694e313330a8c984ebb2bb90a5613a223139574a4a7639366e4b4155347348465771476d7371666a786433376a617a71696940034a2054ba4451fb8f226dd54b52ec086f4eaa4990d66876899b1badec8ce96ef55208522000f2e5970b5b1854da143f4f3e91bf7eb87d1d62869ef08a2ce3b109994ec365","index":2,"expire":"120s"}]}' http://172.20.0.5:8901
    #sign 2
    tx_sign2=$(curl -ksd '{"method":"Chain33.SignRawTx","params":[{"privkey":"'"$test_prikey"'","txHex":"'"$tx_sign"'","index":3,"expire":"120s"}]}' "${para_ip}" | jq -r ".result")
    #curl -ksd '{"method":"Chain33.SignRawTx","params":[{"privkey":"0x24d1fad138be98eebee31440f144aa38c404533f40862995282162bc538e91c8","txHex":"0a10757365722e702e706172612e6e6f6e6512126e6f2d6665652d7472616e73616374696f6e1a6e080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a473045022100fe7763b2fa5b42eddccf1a3434cb6d6bb60a5c5c32e5498219e99be01bb94ad302201ecb9931a2bb4e1b0d49ec50ee552a774c1db8a1eb9b2dff47e4b931625e3af220e0a71230a89780b2a5b187990a3a2231466a58697076505142754e6f7a4133506150724b6846703854343166717141707640034afc050a8e020a10757365722e702e706172612e6e6f6e6512126e6f2d6665652d7472616e73616374696f6e1a6e080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a473045022100fe7763b2fa5b42eddccf1a3434cb6d6bb60a5c5c32e5498219e99be01bb94ad302201ecb9931a2bb4e1b0d49ec50ee552a774c1db8a1eb9b2dff47e4b931625e3af220e0a71230a89780b2a5b187990a3a2231466a58697076505142754e6f7a4133506150724b6846703854343166717141707640034a2054ba4451fb8f226dd54b52ec086f4eaa4990d66876899b1badec8ce96ef55208522000f2e5970b5b1854da143f4f3e91bf7eb87d1d62869ef08a2ce3b109994ec3650aab020a15757365722e702e706172612e7061726163726f7373122e10904e22291080c2d72f2222314d41754538515362626563683362564b4b324a504a4a7859784e745439356f53551a6e0801122103589ebf581958aeb8a72ff517f823c878aee16139ecbf0001a4611e9c004fecdf1a473045022100da5ad2bdc6e1e43a01d32c44f116e5d0bf96aa4c16debad49381ea5d11a49835022055a510460df9b63f8f585393d6603abf1388fac0e122b53ef3533f242287915730d195faf7d3a1ec9a4c3a223139574a4a7639366e4b4155347348465771476d7371666a786433376a617a71696940034a2054ba4451fb8f226dd54b52ec086f4eaa4990d66876899b1badec8ce96ef5520852208fb7939bf2701e5af9cef05da33020be682e10c84fde9bd8d24765031fad530b0aba010a15757365722e702e706172612e7061726163726f7373124f1004424b0a09636f696e732e6274791080ade2042215757365722e702e706172612e7061726163726f73732a2231326269686a7a626159576a637044696979395375415765714e6b735164694e313330a8c984ebb2bb90a5613a223139574a4a7639366e4b4155347348465771476d7371666a786433376a617a71696940034a2054ba4451fb8f226dd54b52ec086f4eaa4990d66876899b1badec8ce96ef55208522000f2e5970b5b1854da143f4f3e91bf7eb87d1d62869ef08a2ce3b109994ec365","index":3,"expire":"120s"}]}' http://172.20.0.5:8901

    #send
    chain33_SendTx "${tx_sign2}" "${para_ip}"
    #tx_rst=$(curl -ksd '{"method":"Chain33.SendTransaction","params":[{"token":"BTY","data":"${tx_sign2}"}]'}  "${para_ip}" | jq -r ".result" )

    local transfer_expect="700000000"
    local exec_expect="100000000"
    transfer_val=$(paracross_QueryParaBalance "${test_addr}" "$paracross_execer_name")
    transfer_exec_val=$(paracross_QueryParaBalance "${test_addr}" "$trade_exec_name")
    if [ "${transfer_val}" != $transfer_expect ]; then
        echo "paracross_testTxGroup trasfer failed, get=$transfer_val,expec=$transfer_expect"
        exit 1
    fi
    if [ "${transfer_exec_val}" != $exec_expect ]; then
        echo "paracross_testTxGroup toexec failed, get=$transfer_exec_val,expec=$exec_expect"
        exit 1
    fi

    echo_rst "$FUNCNAME" 0

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
