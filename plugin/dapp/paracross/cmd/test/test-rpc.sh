#!/usr/bin/env bash
# shellcheck disable=SC2128

CASE_ERR=""
UNIT_HTTP=""

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

function paracross_QueryBalance() {
    local req
    local resp
    local balance

    req='{"method":"Chain33.GetBalance", "params":[{"addresses" : ["'"$1"'"], "execer" : "paracross","asset_exec":"paracross","asset_symbol":"coins.bty"}]}'
    resp=$(curl -ksd "$req" "${UNIT_HTTP}")
    balance=$(jq -r '.result[0].balance' <<<"$resp")
    echo "$balance"
    return $?
}

function paracross_Transfer_Withdraw() {

    # 计数器，资产转移操作和取钱操作都成功才算成功，也就是 counter == 2
    local count=0
    #fromAddr  跨链资产转移地址
    local fromAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    #privkey 地址签名
    local privkey="0x4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
    #paracrossAddr 合约地址
    local paracrossAddr="1HPkPopVe3ERfvaAgedDtJQ792taZFEHCe"
    #amount_save 存钱到合约地址
    local amount_save=100000000
    #amount_should 应转移金额
    local amount_should=27000000
    #withdraw_should 应取款金额
    local withdraw_should=13000000
    #fee 交易费
    local fee=1000000
    #转移前余额
    local para_balance_before
    #转移后余额
    local para_balance_after
    #取钱后余额
    local para_balance_withdraw_after
    #构造交易哈希
    local tx_hash
    #实际转移金额
    local amount_real
    #实际取钱金额
    local withdraw_real

    #1. 查询资产转移前余额状态
    para_balance_before=$(paracross_QueryBalance "$fromAddr")
    echo "before transferring:$para_balance_before"

    #2  存钱到合约地址
    tx_hash=$(curl -ksd '{"method":"Chain33.CreateRawTransaction","params":[{"to":"'"$paracrossAddr"'","amount":'$amount_save'}]}' ${UNIT_HTTP} | jq -r ".result")
    ##echo "tx:$tx"
    chain33_SignRawTx "$tx_hash" "$privkey" ${UNIT_HTTP}
    #paracross_SignAndSend $fee "$privkey" "$tx_hash"

    #3  资产从主链转移到平行链
    tx_hash=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"paracross","actionName":"ParacrossAssetTransfer","payload":{"execer":"user.p.para.paracross","execName":"user.p.para.paracross","to":"'"$fromAddr"'","amount":'$amount_should'}}]}' ${UNIT_HTTP} | jq -r ".result")
    #echo "rawTx:$rawTx"
    chain33_SignRawTx "$tx_hash" "$privkey" ${UNIT_HTTP}
    #paracross_SignAndSend $fee "$privkey" "$tx_hash"

    #4 查询转移后余额状态
    local times=100
    while true; do
        para_balance_after=$(paracross_QueryBalance "$fromAddr")
        echo "after transferring:$para_balance_after"
        #real_amount  实际转移金额
        amount_real=$((para_balance_after - para_balance_before))
        #echo $amount_real
        if [ "$amount_real" != "$amount_should" ]; then
            chain33_BlockWait 2 ${UNIT_HTTP}
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                exit 1
            fi
        else
            #echo "para_cross_transfer_withdraw success"
            let count+=1
            break
        fi
    done

    #5 取钱
    tx_hash=$(curl -ksd '{"method":"Chain33.CreateTransaction","params":[{"execer":"paracross","actionName":"ParacrossAssetWithdraw","payload":{"IsWithdraw":true,"execer":"user.p.para.paracross","execName":"user.p.para.paracross","to":"'"$fromAddr"'","amount":'$withdraw_should'}}]}' ${UNIT_HTTP} | jq -r ".result")
    #echo "rawTx:$rawTx"
    chain33_SignRawTx "$tx_hash" "$privkey" ${UNIT_HTTP}
    #paracross_SignAndSend $fee "$privkey" "$tx_hash"

    #6 查询取钱后余额状态
    local times=100
    while true; do
        para_balance_withdraw_after=$(paracross_QueryBalance "$fromAddr")
        echo "after withdrawing :$para_balance_withdraw_after"
        #实际取钱金额
        withdraw_real=$((para_balance_after - para_balance_withdraw_after))
        #echo $withdraw_real
        if [ "$withdraw_should" != "$withdraw_real" ]; then
            chain33_BlockWait 2 ${UNIT_HTTP}
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                exit 1
            fi
        else
            #echo "para_cross_transfer_withdraw success"
            let count+=1
            break
        fi
    done

    [ "$count" -eq 2 ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

function paracross_IsSync() {
    local ok

    ok=$(curl -ksd '{"method":"paracross.IsSync","params":[]}' ${UNIT_HTTP} | jq -r ".result")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_ListTitles() {

    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.ListTitles","params":[]}' ${UNIT_HTTP})
    echo $resp
    ok=$(jq '(.error|not) and (.result| [has("titles"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_GetHeight() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.GetHeight","params":[]}' ${UNIT_HTTP})
    echo $resp
    ok=$(jq '(.error|not) and (.result| [has("consensHeight"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_GetNodeGroupAddrs() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.GetNodeGroupAddrs","params":[{"title":"user.p.para."}]}' ${UNIT_HTTP})
    echo $resp
    ok=$(jq '(.error|not) and (.result| [has("key","value"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_GetNodeGroupStatus() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.GetNodeGroupStatus","params":[{"title":"user.p.para."}]}' ${UNIT_HTTP})
    echo $resp
    ok=$(jq '(.error|not) and (.result| [has("status"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_ListNodeGroupStatus() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.ListNodeGroupStatus","params":[{"title":"user.p.para.","status":2}]}' ${UNIT_HTTP})
    echo $resp
    ok=$(jq '(.error|not) and (.result| [has("status"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function paracross_ListNodeStatus() {
    local resp
    local ok

    resp=$(curl -ksd '{"method":"paracross.ListNodeStatus","params":[{"title":"user.p.para.","status":4}]}' ${UNIT_HTTP})
    echo $resp
    ok=$(jq '(.error|not) and (.result| [has("status"),true])' <<<"$resp")
    [ "$ok" == true ]
    local rst=$?
    echo_rst "$FUNCNAME" "$rst"
}

function run_testcases() {
    chain33_lock
    chain33_unlock
    paracross_GetBlock2MainInfo

    local ispara=$(echo '"'"${UNIT_HTTP}"'"' | jq '.|contains("8901")')
    if [ "$ispara" == true ]; then
        paracross_IsSync
        paracross_GetHeight
        paracross_ListTitles
        paracross_GetNodeGroupAddrs
        paracross_GetNodeGroupStatus
        paracross_ListNodeGroupStatus
        paracross_ListNodeStatus
        paracross_Transfer_Withdraw
    fi
}

function main() {
    UNIT_HTTP=$1
    echo "=========== # paracross rpc test ============="

    run_testcases

    if [ -n "$CASE_ERR" ]; then
        echo "paracross there some case error"
        exit 1
    fi
}

main "$1"
