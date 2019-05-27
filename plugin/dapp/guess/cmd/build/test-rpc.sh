#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""
guess_admin_addr=14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
guess_user1_addr=12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
guess_user2_addr=1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF

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

saveSeed() {
    seed="journey notable narrow few bar stuff notable custom miss brother attend tongue price theme resist"
    req='{"method":"Chain33.SaveSeed", "params":[{"seed":'"$seed"', "passwd": "1314fuzamei"}]}'
    resp=$(curl -ksd "$req" "${MAIN_HTTP}")
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result| has("isOK"))' <<<"$resp")
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

    req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944", "label":"genesis1"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$1")
    #        echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.label=="genesis1") and (.result.acc.addr == "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

importPrivkey2() {

    req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01", "label":"genesis2"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$1")
    #        echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.label=="genesis2") and (.result.acc.addr == "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

importPrivkey3() {

    req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"B0BB75BC49A787A71F4834DA18614763B53A18291ECE6B5EDEC3AD19D150C3E7", "label":"genesis3"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$1")
    #        echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.label=="genesis3") and (.result.acc.addr == "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}


sendTransaction1() {
    local fee=1000000
    local exec="coins"
    local to="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    local from="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
    local privkey="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"

    tx1="0a05636f696e73122f18010a2b1080a094a58d1d2222313271796f6361794e46374c7636433971573461767873324537553431664b53667620a08d0630e9a6e48dd5ddab82393a22313271796f6361794e46374c7636433971573461767873324537553431664b536676"
    tx=$(curl -ksd '{"method":"Chain33.SignRawTx","params":[{"expire":"120s","fee":'$fee',"privkey":"'$privkey'","txHex":"'$tx1'"}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.SendTransaction","params":[{"data":"'"$tx"'"}]}' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result != null)' <<<"$data")

    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}
sendTransaction2() {
    local fee=1000000
    local exec="coins"
    local to="1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"
    local from="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
    local privkey="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"

    tx1="0a05636f696e73122f18010a2b1080a094a58d1d22223145444467684174674273616d724e45744e6d5964517a43315145684c6b7238377420a08d0630c1cc9c89dfafd8b4493a223145444467684174674273616d724e45744e6d5964517a43315145684c6b72383774"
    tx=$(curl -ksd '{"method":"Chain33.SignRawTx","params":[{"expire":"120s","fee":'$fee',"privkey":"'$privkey'","txHex":"'$tx1'"}]}' ${MAIN_HTTP} | jq -r ".result")

    data=$(curl -ksd '{"method":"Chain33.SendTransaction","params":[{"data":"'"$tx"'"}]}' ${MAIN_HTTP})
    ok=$(jq '(.error|not) and (.result != null)' <<<"$data")

    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
}


guess_game_start() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Start", "payload":{"topic":"WorldCup Final","options":"A:France;B:Claodia","category":"football","maxBetsOneTime":10000000000,"maxBetsNumber":100000000000,"devFeeFactor":5,"devFeeAddr":"1D6RFZNp2rh6QdbcZ1d7RWuBUz61We6SD7","platFeeFactor":5,"platFeeAddr":"1PHtChNt3UcfssR7v7trKSk3WJtAWjKjjX"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signAndSendRawTx "$rawtx" "${guess_admin_addr}"
    eventId="${txhash}"
    echo "eventId $eventId"
}

guess_game_bet1() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Bet", "payload":{"gameID":"${eventId}","option":"A", "betsNum":500000000}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signAndSendRawTx "$rawtx" "${guess_user1_addr}"
    #eventId="${txhash}"
    #echo "eventId $eventId"
    echo "txHash ${txhash}"
}

guess_game_bet2() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Bet", "payload":{"gameID":"${eventId}","option":"B", "betsNum":500000000}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signAndSendRawTx "$rawtx" "${guess_user2_addr}"
    #eventId="${txhash}"
    #echo "eventId $eventId"
    echo "txHash ${txhash}"
}

guess_game_stop() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"StopBet", "payload":{"gameID":"${eventId}"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signAndSendRawTx "$rawtx" "${guess_admin_addr}"
    #eventId="${txhash}"
    #echo "eventId $eventId"
    echo "txHash ${txhash}"
}

guess_game_publish() {
    req='"method":"Chain33.CreateTransaction","params":[{"execer":"guess","actionName":"Publish", "payload":{"gameID":"${eventId}","result":"A"}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    signAndSendRawTx "$rawtx" "${guess_admin_addr}"
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

guess_QueryGameByID() {
    event_id=$1
    status=$2
    local req='"method":"Chain33.Query", "params":[{"execer":"guess","funcName":"QueryGameByID","payload":{"gameID":["'"$event_id"'"]}}]'
    #echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result. | [has("gameId", "status", "result"),true] | unique | length == 1) and (.result.game.status == "'$status'")' <<<"$resp")
    [ "$ok" == true ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"
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
    local txhash="$1"
    local req='"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]'
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
            return 0
            break
        fi
    done
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

    #导入用户2地址私钥
    importPrivkey3

    #向用户1地址转账
    sendTransaction1

    #向用户2地址转账
    sendTransaction2

    #管理员创建游戏
    guess_game_start

    #等待2个区块
    block_wait 2

    #查询游戏状态
    guess_QueryGameByID $eventId 11

    #用户1下注
    guess_game_bet1

    #等待1个区块
    block_wait 2

    #查询游戏状态
    guess_QueryGameByID $eventId 12

    #用户2下注
    guess_game_bet2

    #等待2个区块
    block_wait 2

    #查询游戏状态
    guess_QueryGameByID $eventId 12


    #管理员停止下注
    guess_game_stop

    #等待2个区块
    block_wait 2

    #查询游戏状态
    guess_QueryGameByID $eventId 13

    #管理员发布结果
    guess_game_publish

    #等待2个区块
    block_wait 2

    #查询游戏状态
    guess_QueryGameByID $eventId 15
}

function main() {

    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    echo "=========== # guess rpc test start============="
    run_test

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Guess Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Guess Rpc Test Pass==============${NOC}"
    fi
    echo "=========== # guess rpc test end============="
}

main "$1"
