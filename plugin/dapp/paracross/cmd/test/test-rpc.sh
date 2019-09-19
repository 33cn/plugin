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

chain33_lock() {
    local ok

    ok=$(curl -ksd '{"method":"Chain33.Lock","params":[]}' ${UNIT_HTTP} | jq -r ".result.isOK")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

chain33_unlock() {
    local ok
    ok=$(curl -ksd '{"method":"Chain33.UnLock","params":[{"passwd":"1314fuzamei","timeout":0}]}' ${UNIT_HTTP} | jq -r ".result.isOK")
    [ "$ok" == true ]
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

    req='{"method":"Chain33.GetBalance", "params":[{"addresses" : ["'"$1"'"], "execer" : "paracross","asset_exec":"paracross","asset_symbol":"coins.bty"}]}'
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
    para_balance_before=$(paracross_QueryParaBalance "$from_addr")
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
        para_balance_after=$(paracross_QueryParaBalance "$from_addr")
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
        para_balance_withdraw_after=$(paracross_QueryParaBalance "$from_addr")
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

function run_testcases() {
    chain33_lock
    chain33_unlock
    paracross_GetBlock2MainInfo
    paracross_IsSync
    paracross_GetHeight
    paracross_ListTitles
    paracross_GetNodeGroupAddrs
    paracross_GetNodeGroupStatus
    paracross_ListNodeGroupStatus
    paracross_ListNodeStatus
    paracross_Transfer_Withdraw
}

function main() {

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

    if [ -n "$CASE_ERR" ]; then
        echo "paracross there some case error"
        exit 1
    fi
}

main "$1" "$2" "$3" "$4"
#main http://127.0.0.1:8801
#main http://47.98.253.127:8801 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b user.p.fzmtest.paracross
