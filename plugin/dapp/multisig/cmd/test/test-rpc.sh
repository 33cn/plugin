#!/usr/bin/env bash
# shellcheck disable=SC2128
set +e
set -o pipefail

MAIN_HTTP=""

# shellcheck source=/dev/null
source ../dapp-test-common.sh

Symbol="BTY"
Asset="coins"
#PrivKeyA="0x06c0fa653c719275d1baa365c7bc0b9306447287499a715b541b930482eaa504"
#PrivKeyB="0x4c8663cded61093af20339ae038b3c6bfa58a33e65874a655022f82eaf3f2fa0"
#PrivKeyC="0x9abcf378b397682109c174b37a45bfc8a459c9514dd2ef719e22a9815373047d"
#PrivKeyD="0xbf8f865a03fec64f30d2243847807e88d2dbc8104e77925e4fc11c4d4380f3da"
#PrivKeyE="0x5b8ca316cf073aa94f1056a9e3f6e0b9a9ec11ae45862d58c7a09640b4d55302"
#PrivKeyGen="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
AddrA="1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK"
AddrB="1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj"
#AddrC="1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd"
#AddrD="166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf"
AddrE="1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo"
#GenAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"

GenAddr="15wcitPEu1X1TBfrGfwN8GTkNTJoCmGc75"
PrivKeyGen="0x295710fa409bd0b0bf928efa0994645edfe80a247d89c1e1637f90dc5e303f5e"

multisigExecAddr=""
multisigAccAddr=""
execName=""
#execCoins=""
#symbolCoins=""
function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    if [ "$ispara" == true ]; then
        execName="user.p.para.multisig"
        Symbol="para"
        multisigExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.multisig"}]}' ${MAIN_HTTP} | jq -r ".result")
    else
        execName="multisig"
        Symbol="BTY"
        multisigExecAddr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"multisig"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi

    local main_ip=${MAIN_HTTP//8901/8801}

    if [ "$ispara" == false ]; then
        chain33_applyCoins "$GenAddr" 12000000000 "${main_ip}"
        chain33_QueryBalance "${GenAddr}" "$main_ip"

    else
        # tx fee
        chain33_applyCoins "$GenAddr" 1000000000 "${main_ip}"
        chain33_QueryBalance "${GenAddr}" "$main_ip"

        local para_ip="${MAIN_HTTP}"
        #para chain import pri key
        chain33_ImportPrivkey "0x295710fa409bd0b0bf928efa0994645edfe80a247d89c1e1637f90dc5e303f5e" "15wcitPEu1X1TBfrGfwN8GTkNTJoCmGc75" "gen" "$para_ip"

        chain33_applyCoins "$GenAddr" 12000000000 "${para_ip}"
        chain33_QueryBalance "${GenAddr}" "$para_ip"
    fi

    echo "multisigExecAddr=$multisigExecAddr"
}
# 创建多重签名账户
function multisig_AccCreateTx() {
    echo "========== # multisig_AccCreateTx begin  =========="
    txHex=$(curl -ksd '{"method":"multisig.MultiSigAccCreateTx","params":[{"owners":[{"ownerAddr":"'$AddrA'","weight":20},{"ownerAddr":"'$AddrB'","weight":10},{"ownerAddr":"'$GenAddr'","weight":30}],"requiredWeight":15,"dailyLimit":{"symbol":"'$Symbol'","execer":"'$Asset'","dailyLimit":1000000000}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}

    #查看创建的多重签名地址是否ok
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccCount","payload":{}}]}' ${MAIN_HTTP} | jq -r ".result.data")

    #获取创建的多重签名地址
    multisigAccAddr=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccounts","payload":{"start":"0","end":"0"}}]}' ${MAIN_HTTP} | jq -r ".result.address[0]")
    echo "multisigAccAddr=$multisigAccAddr"
    #多重签名地址查询具体信息
    result1=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccountInfo","payload":{"multiSigAccAddr":"'"$multisigAccAddr"'"}}]}' ${MAIN_HTTP})

    ok1=$(jq '(.result.createAddr == "'$GenAddr'")' <<<"$result1")
    [ "$ok1" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

    result=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccAllAddress","payload":{"multiSigAccAddr":"'$GenAddr'"}}]}' ${MAIN_HTTP})
    ok=$(jq '(.result.address[0] == "'"$multisigAccAddr"'")' <<<"$result")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

    echo "========== # multisig_AccCreateTx ok  =========="
}

#多重签名地址转入操作
function multisig_TransferInTx() {
    echo "========== # multisig_TransferInTx begin =========="

    #首先转账到multisig合约中
    txHex=$(curl -ksd '{"method":"Chain33.CreateRawTransaction","params":[{"to":"'"$multisigExecAddr"'","amount":5000000000,"fee":1,"note":"12312","execName":"'"$execName"'"}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #转账到multisigAccAddr地址中
    txHex=$(curl -ksd '{"method":"multisig.MultiSigAccTransferInTx","params":[{"symbol":"'$Symbol'","amount":4000000000,"note":"test ","execname":"'$Asset'","to":"'"$multisigAccAddr"'"}]}' ${MAIN_HTTP} | jq -r ".result")

    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #查询multisigAccAddr地址资产信息
    accountasset=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccAssets","payload":{"multiSigAddr":"'"$multisigAccAddr"'","assets":{"execer":"'$Asset'","symbol":"'$Symbol'"},"isAll":false}}]}' ${MAIN_HTTP} | jq -r ".result.accAssets[0]")
    echo "multisig_TransferInTx:=${accountasset}"
    ok=$(jq '(.assets.execer == "'$Asset'") and (.assets.symbol == "'$Symbol'") and (.account.frozen == "4000000000")' <<<"$accountasset")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    echo "========== # multisig_TransferInTx end =========="

}

function multisig_TransferOutTx() {
    echo "========== # multisig_TransferOutTx begin =========="
    #由GenAddr账户签名从multisigAccAddr账户转出2000000000到AddrB

    txHex=$(curl -ksd '{"method":"multisig.MultiSigAccTransferOutTx","params":[{"symbol":"'$Symbol'","amount":2000000000,"note":"test ","execname":"coins","to":"'$AddrB'","from":"'"$multisigAccAddr"'"}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #查询AddrB账户在multisig合约下有2000000000
    accountasset=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccAssets","payload":{"multiSigAddr":"1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj","assets":{"execer":"coins","symbol":"'$Symbol'"},"isAll":false}}]}' ${MAIN_HTTP} | jq -r ".result.accAssets[0]")
    echo "multisig_TransferOutTx:=${accountasset}"
    ok=$(jq '(.assets.execer == "'$Asset'") and (.assets.symbol == "'$Symbol'") and (.account.balance == "2000000000")' <<<"$accountasset")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

    #查询multisigAccAddr地址资产信息，减少了2000000000

    accountasset=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccAssets","payload":{"multiSigAddr":"'"$multisigAccAddr"'","assets":{"execer":"'$Asset'","symbol":"'$Symbol'"},"isAll":false}}]}' ${MAIN_HTTP} | jq -r ".result.accAssets[0]")
    ok=$(jq '(.assets.execer == "'$Asset'") and (.assets.symbol == "'$Symbol'") and (.account.frozen == "2000000000")' <<<"$accountasset")
    [ "$ok" == true ]

    rst=$?
    echo_rst "$FUNCNAME" "$rst"
    echo "========== # multisig_TransferOutTx end =========="
}

function multisig_OwnerOperateTx() {
    echo "========== # multisig_OwnerOperateTx begin =========="
    #通过GenAddr账户添加AddrE到多重签名账户的owner

    txHex=$(curl -ksd '{"method":"multisig.MultiSigOwnerOperateTx","params":[{"multiSigAccAddr":"'"$multisigAccAddr"'","newOwner":"'$AddrE'","newWeight":8,"operateFlag":1}]}' ${MAIN_HTTP} | jq -r ".result")

    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #查询多重签名账户的信息中有AddrE
    owner=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccountInfo","payload":{"multiSigAccAddr":"'"$multisigAccAddr"'"}}]}' ${MAIN_HTTP} | jq -r ".result.owners[3]")
    ok=$(jq '(.ownerAddr == "'$AddrE'")' <<<"$owner")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    #删除多重签名账户的信息中owner AddrE

    txHex=$(curl -ksd '{"method":"multisig.MultiSigOwnerOperateTx","params":[{"multiSigAccAddr":"'"$multisigAccAddr"'","oldOwner":"'$AddrE'","operateFlag":2}]}' ${MAIN_HTTP} | jq -r ".result")

    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #修改多重签名账户中owner AddrA的weight为30
    txHex=$(curl -ksd '{"method":"multisig.MultiSigOwnerOperateTx","params":[{"multiSigAccAddr":"'"$multisigAccAddr"'","oldOwner":"'$AddrA'","newWeight":30,"operateFlag":3}]}' ${MAIN_HTTP} | jq -r ".result")

    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #将多重签名账户中owner AddrA的地址替换成AddrE
    txHex=$(curl -ksd '{"method":"multisig.MultiSigOwnerOperateTx","params":[{"multiSigAccAddr":"'"$multisigAccAddr"'","oldOwner":"'$AddrA'","newOwner":"'$AddrE'","operateFlag":4}]}' ${MAIN_HTTP} | jq -r ".result")

    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #查询多重签名账户的信息中有AddrE并且weight为30

    owner=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccountInfo","payload":{"multiSigAccAddr":"'"$multisigAccAddr"'"}}]}' ${MAIN_HTTP} | jq -r ".result.owners[0]")
    ok=$(jq '(.ownerAddr == "'$AddrE'") and (.weight == "30")' <<<"$owner")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

function multisig_AccOperateTx() {
    echo "========== # multisig_AccOperateTx begin =========="
    #修改每日限额的值为 Symbol：Asset dailyLimit=1200000000
    txHex=$(curl -ksd '{"method":"multisig.MultiSigAccOperateTx","params":[{"multiSigAccAddr":"'"$multisigAccAddr"'","dailyLimit":{"symbol":"'$Symbol'","execer":"'$Asset'","dailyLimit":1200000000}}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #增加资产的配置 HYB：token dailyLimit=1000000000

    txHex=$(curl -ksd '{"method":"multisig.MultiSigAccOperateTx","params":[{"multiSigAccAddr":"'"$multisigAccAddr"'","dailyLimit":{"symbol":"HYB","execer":"token","dailyLimit":1000000000}}]}' ${MAIN_HTTP} | jq -r ".result")

    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #修改RequiredWeight=16
    txHex=$(curl -ksd '{"method":"multisig.MultiSigAccOperateTx","params":[{"multiSigAccAddr":"'"$multisigAccAddr"'","newRequiredWeight":16,"operateFlag":true}]}' ${MAIN_HTTP} | jq -r ".result")
    chain33_SignRawTx "$txHex" "$PrivKeyGen" ${MAIN_HTTP}
    #chain33_BlockWait 1 ${MAIN_HTTP}

    #获取本多重签名账户上的交易数，已经对应的交易信息
    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccTxCount","payload":{"multiSigAccAddr":"'"$multisigAccAddr"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    ok=$(jq '(.data != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    #获取本多重签名账户上的交易数，通过交易交易id获取交易信息

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigAccTxCount","payload":{"multiSigAccAddr":"'"$multisigAccAddr"'"}}]}' ${MAIN_HTTP} | jq -r ".result")
    ok=$(jq '(.data != "")' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

    #查询多重签名账户信息中

    data=$(curl -ksd '{"method":"Chain33.Query","params":[{"execer":"multisig","funcName":"MultiSigTxInfo","payload":{"multiSigAddr":"'"$multisigAccAddr"'","txId":"7"}}]}' ${MAIN_HTTP} | jq -r ".result")
    ok=$(jq '(.txid == "7") and (.executed == true) and (.multiSigAddr == "'"$multisigAccAddr"'") ' <<<"$data")

    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

function run_test() {
    multisig_AccCreateTx
    multisig_TransferInTx
    multisig_TransferOutTx
    multisig_OwnerOperateTx
    multisig_AccOperateTx

}

function main() {
    chain33_RpcTestBegin multisi
    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    run_test

    chain33_RpcTestRst multisi "$CASE_ERR"
}

main "$1"
