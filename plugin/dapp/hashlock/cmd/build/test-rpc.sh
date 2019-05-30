#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""
addr_A=14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
addr_B=12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv

eventId=""
txhash=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# $2=0 means true, other false
echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="err"
    fi
}
set -x
saveSeed() {

    seed="journey notable narrow few bar stuff notable custom miss brother attend tongue price theme resist"
    req='{"method":"Chain33.SaveSeed", "params":[{"seed":"'"$seed"'", "passwd": "1314fuzamei"}]}'
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    #    echo "#response: $resp"
    ok=$(jq '(((.error|not) and (.result| has("isOK"))) or (.error and (.result and .result.msg=="ErrSeedExist")))' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

unlock() {
    ok=$(curl -ksd '{"method":"Chain33.UnLock","params":[{"passwd":"1314fuzamei","timeout":0}]}' ${MAIN_HTTP} | jq -r ".result.isOK")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

importPrivkey1() {
    req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944", "label":"genesis11"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    #        echo "#response: $resp"
    ok=$(jq '(((.error|not) and (.result.label=="genesis11") and (.result.acc.addr == "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")) or (.error=="ErrPrivkeyExist"))' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

importPrivkey2() {

    req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01", "label":"genesis12"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    #        echo "#response: $resp"
    ok=$(jq '(((.error|not) and (.result.label=="genesis12") and (.result.acc.addr == "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv")) or (.error=="ErrPrivkeyExist"))' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}


sendTransaction() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        #sendTransaction11
        Chain33_SendToAddress 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv 12000000000
    fi

    Chain33_SendToAddress 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt  12000000000
}

sendToExec() {
    from=$1
    Chain33_SendToAddress "$from" "$hashlock_addr" 2000000000
}


queryBalance() {
    addr=$1
    req='"method":"Chain33.GetAllExecBalance","params":[{"addr":"'"${addr}"'"}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    echo $resp|jq -r ".result"
}

hashlock_lock() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"hashlock","actionName":"HashlockLock", "payload":{"secret":"abc","amount":1000000000, "time":360,"toAddr":"'"${addr_A}"'", "returnAddr":"'"${addr_B}"'","fee":100000000}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signAndSendRawTx "$rawtx" "${addr_B}"
    eventId="${txhash}"
    echo "eventId $eventId"
}

hashlock_unlock() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"hashlock","actionName":"HashlockUnlock", "payload":{"secret":"abc","fee":100000000}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signAndSendRawTx "$rawtx" "${addr_B}"
    #eventId="${txhash}"
    #echo "eventId $eventId"
    echo "txHash ${txhash}"
}

hashlock_send() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"hashlock","actionName":"HashlockSend", "payload":{"secret":"abc","fee":100000000}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signAndSendRawTx "$rawtx" "${addr_A}"
    #eventId="${txhash}"
    #echo "eventId $eventId"
    echo "txHash ${txhash}"
}

# 签名并发送
signAndSendRawTx() {
    unsignedTx=$1
    addr=$2
    req='"method":"Chain33.SignRawTx","params":[{"addr":"'${addr}'","txHex":"'${unsignedTx}'","expire":"120s"}]'
    signedTx=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result")
    if [ "$signedTx" == "null" ]; then
        echo "An error occurred while signing"
    else
        sendSignedTx "$signedTx"
    fi
}

sendSignedTx() {
    signedTx=$1
    local req='"method":"Chain33.SendTransaction","params":[{"token":"","data":"'"$signedTx"'"}]'
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    ok=$(echo "${resp}" | jq -r ".error")
    [ "$ok" == null ]
    rst=$?
    #echo_rst "$FUNCNAME" "$rst"
    txhash=$(echo "${resp}" | jq -r ".result")
    echo "tx hash is $txhash"
}

function block_wait() {
    local req='"method":"Chain33.GetLastHeader","params":[]'
    cur_height=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq ".result.height")
    expect=$((cur_height + ${1}))
    local count=0
    while true; do
        new_height=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq ".result.height")
        if [ "${new_height}" -ge "${expect}" ]; then
            break
        fi
        count=$((count + 1))
        sleep 1
    done
    echo "wait new block $count s, cur height=$expect,old=$cur_height"
}

function queryTransaction() {
    block_wait 1
    local tx_hash="$1"
    local req='"method":"Chain33.QueryTransaction","params":[{"hash":"'"$tx_hash"'"}]'
    local times=10
    while true; do
        ret=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.tx.hash")
        if [ "${ret}" != "${1}" ]; then
            block_wait 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "====query tx=$1 failed"
                echo "req=$req"
                curl -ksd "{$req}" ${MAIN_HTTP}
                return 1
                exit 1
            fi
        else
            echo "====query tx=$1  success"
            ret=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.tx")
            echo $ret
            return 0
            break
        fi
    done
}

Chain33_SendToAddress() {
    local from="$1"
    local to="$2"
    local amount=$3
    local req='"method":"Chain33.SendToAddress", "params":[{"from":"'"$from"'","to":"'"$to"'", "amount":'"$amount"', "note":"test\n"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.hash|length==66)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    hash=$(jq '(.result.hash)' <<<"$resp")
    echo "hash=$hash"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        hashlock_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.hashlock"}]}' ${MAIN_HTTP} | jq -r ".result")
        hashlock_exec="user.p.para.guess"
    else
        hashlock_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"hashlock"}]}' ${MAIN_HTTP} | jq -r ".result")
        hashlock_exec="hashlock"
    fi
    echo "hashlock_addr=$hashlock_addr"
}


function run_test() {

    #保存seed
    saveSeed

    #unlock wallet
    unlock

    #导入admin地址私钥
    importPrivkey1

    #导入用户1地址私钥
    importPrivkey2

    #向管理地址转账，确保有钱执行交易
    sendTransaction
    block_wait 1

    queryBalance $addr_A
    queryBalance $addr_B

    #用户地址向合约转账，确保可以参与游戏
    sendToExec $addr_A
    sendToExec $addr_B

    block_wait 2

    #lock
    hashlock_lock

    #等待2个区块
    block_wait 2

    #查询交易
    queryTransaction $txhash
    queryBalance $addr_A
    queryBalance $addr_B

    #send
    hashlock_send

    #等待2个区块
    block_wait 2

    #查询交易
    queryTransaction $txhash
    queryBalance $addr_A
    queryBalance $addr_B

    #unlock failed
    hashlock_unlock

    #等待2个区块
    block_wait 2

    #查询交易
    queryTransaction $txhash
    queryBalance $addr_A
    queryBalance $addr_B


    #lock
    hashlock_lock

    #等待2个区块
    block_wait 2
    queryTransaction $txhash

    #unlock failed
    hashlock_unlock


    #等待2个区块
    block_wait 2

    #查询交易
    queryTransaction $txhash
}

function main() {

    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    init
    echo "=========== # hashlock rpc test start============="
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============HashLock Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============HashLock Rpc Test Pass==============${NOC}"
    fi
    echo "=========== # hashlock rpc test end============="
}

main "$1"

set +x