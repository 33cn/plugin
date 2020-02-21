#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh

HTTP=""

EXECTOR=""
EXECTOR_ADDR=""
TICKET_EXECTOR=""
TICKET_ADDR=""

propKey="0xfd0c4a8a1efcd221ee0f36b7d4f57d8ff843cb8bc193b39c7863332d355acafa"
propAddr="15VUiygdxMSZ3rykwe742yomp2cPJ9Tfve"
votePrKey="1c3e6cac2f887e1ab9180e2d5772dc4ba01accb8d4df434faba097003eb35482"
voteAddr="1Q9sQwothzM1gKSzkVZ8Dt1tqKX1uzSagx"

proposalID=""

boardsAddr=(
    "1N578zmVzVR7RxLfnp7XAeDmAy499Jw3q2"
    "1DZ1kL9x3rRwz7EZjcLt1kMYu6Zdp3MjGR"
    "1HUYR1Mzb91m3dmsEE1vPrv7BsAHmEtVzM"
    "1JHmVgchSLjszN9LAYa3gds811c4BH2J51"
    "14TDcn95hxHbpySPtxGDK3aY6qDNuk5idg"
    "1EpbYadEAcwbrxuh6Ph2qyJZhMY9F9CCCv"
    "1NqfXb3YotDTPuShgSAhrBH28ETzCrsZx2"
    "138xSezQUi9kynnAZR23kWBwHvDSRX8JDK"
    "18bncBidkwRaUYPaqa9Rc58tWeRT2C5tfL"
    "16z3gEvRGof8cnQRvicV8BdY54fUfnBo8A"
    "1DpzRWk9SYnEvuGrFsJPFvMVRSYRHaCbtU"
    "1JSQbsb1hfB6CLzwzVZZEtai4da8m59obK"
    "12bRYKMsLSbN5upxoSenypWrygVuQh8rM5"
    "1Ew4Cpcfh1u1EeRkYPwVc4PRaSnz4D2eph"
    "156yPWN3eeZ4yPiwHRifu41FmrTVPavXQ5"
    "192YKhdAkFm18KLcGumM8JeDgWrnpSdo93"
    "12kXgBxsKzUmhfX5Mnv9EKRNGjQXhZzDFd"
    "1HmyNGYj2xyMmQDTpcuheLKXKBe5rj3QpQ"
    "1HGPrjc6H7yBzFV5yCbibvnSUGUgdDNQi3"
    "19WGov4b7wLf4f8JHMRDnJsGVNMDzap38w"
    "1HmRa1jAnzJ5SpJRrUWqUki6hx1u33Nbq4"
)
boardsPrKey=(
    "fa54751118c8159ade22c253f85945a4dd2030b1cf2502eaf785d0a4f5ad7e35"
    "da9371ea52f1fc9d72e75dbc9836774895cd0966fd53c83f5e2c92d878903693"
    "ad9731261c40c68fee96f7b846408fa33d1f3dc2a27bdb3694ec8f3aa153a98b"
    "e902d23ad26052cec64e9ed9055853327787b3bf26eb4646a6d6c1bb516f9fcd"
    "b3af59368bcc6779aecb4e7fbf0cc4040f8f8329eb7632d580be4b6d3be15357"
    "fd9284c11707b571e347b8f44c54aae89c0810d410a7c1f9613326358c564a5e"
    "b013948c123986aadc2525bfe9935dd07972e14b250b938153e763f917ada8d3"
    "5641e3aba9ce0660665cd1a816d1b50267b5bc5f337f73e6ceed0cc5ececa7d2"
    "5a7631b7101252d685bf2b4ba2f11a72fb867faeaa545064ce8d0901ffc3cb17"
    "c9d0d2639c0c0f2b275e5bf8a2797122af110316a4e8e1fe03c39693f5028c93"
    "00b4e3ac1365b89d68c8d5b07b3505cd614408b0d9c8c88df564d4e072deb401"
    "aa139d3f16c1785b2a9171d7863bc4ee9d45115cc0fa71a14c43536c933d1659"
    "33f111573a4613477f8291928a9bae012a74fc9858acb22c9c65ecf7844b63f2"
    "44e43dd0f769bb99638b9cc3c7468225ff703f69c17e9c4751d04e45fdc6c4a0"
    "9a75d6c779846fe2ad4a36021fe9d08652ca69ce09ab39ad874d921a5ec41716"
    "f5ca6b2ad545bd4b854871b18c8d37d2fe8c3625e91be86a204c4086f28e8d0f"
    "89504608a03590e5a4d8c1c82b75e908b28f9963587c85b96b628d238d3a4d1a"
    "fa653545ae52403665fb803ef410c4d3d7f74628b6d3f92218968ad496e4f81b"
    "227df96a414e26e85c7d87a12296344e6a731ce73e424ba9845cb305ec963843"
    "64259075bf2e5a74334442f5048ceaf8427f6097e2dac99c0e463785c7768550"
    "3b812d92b5c365d698255f55c3f0dca027c2f89f2409c51a6297bcf3343c11e6"
)

boards='
"'${boardsAddr[0]}'",
"'${boardsAddr[1]}'",
"'${boardsAddr[2]}'",
"'${boardsAddr[3]}'",
"'${boardsAddr[4]}'",
"'${boardsAddr[5]}'",
"'${boardsAddr[6]}'",
"'${boardsAddr[7]}'",
"'${boardsAddr[8]}'",
"'${boardsAddr[9]}'",
"'${boardsAddr[10]}'",
"'${boardsAddr[11]}'",
"'${boardsAddr[12]}'",
"'${boardsAddr[13]}'",
"'${boardsAddr[14]}'",
"'${boardsAddr[15]}'",
"'${boardsAddr[16]}'",
"'${boardsAddr[17]}'",
"'${boardsAddr[18]}'",
"'${boardsAddr[19]}'",
"'${boardsAddr[20]}'"
'
chain33_para_init() {
    ip=$1
    chain33_ImportPrivkey "${votePrKey}" "${voteAddr}" "autonomytest" "${ip}"
    chain33_SendToAddress "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv" "$voteAddr" 630000000000 "${ip}"
}
chain33_applyCoinsNOLimit() {
    echo "chain33_getMainChainCoins"
    if [ "$#" -lt 3 ]; then
        echo "chain33_getMainCoins wrong params"
        exit 1
    fi
    local targetAddr=$1
    local count=$2
    local ip=$3

    local poolAddr="1PcGKYYoLn1PLLJJodc1UpgWGeFAQasAkx"
    chain33_SendToAddress "${poolAddr}" "${targetAddr}" "$count" "${ip}"
}

handleBoards() {
    local ip=$1
    for ((i = 0; i < ${#boardsPrKey[*]}; i++)); do
        echo "${boardsPrKey[$i]}"
        lab="board_"${i}
        chain33_ImportPrivkey "${boardsPrKey[$i]}" "${boardsAddr[$i]}" "${lab}" "${ip}"
        chain33_applyCoins "${boardsAddr[$i]}" 100000000 "${ip}"
    done
}

proposalBoardTx() {
    local start=$1
    local end=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropBoard", "payload":{"boards": ['"${boards}"'],"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    proposalID=$RAW_TX_HASH
    echo "$proposalID"
    echo_rst "proposalBoard query_tx" "$?"
}

voteBoardTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropBoard", "payload":{"proposalID": "'"${ID}"'","approve": true}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${HTTP}"
    echo "$RAW_TX_HASH"
    echo_rst "voteBoard query_tx" "$?"
}

revokeProposalTx() {
    local ID=$1
    local funcName=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"'"${funcName}"'", "payload":{"proposalID": "'"${ID}"'"}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    echo "$RAW_TX_HASH"
    echo_rst "revoke Proposal $funcName query_tx" "$?"
}

terminateProposalTx() {
    local ID=$1
    local funcName=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"'"${funcName}"'", "payload":{"proposalID": "'"${ID}"'"}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    echo "$RAW_TX_HASH"
    echo_rst "terminate Proposal $funcName query_tx" "$?"
}

queryProposal() {
    local ID=$1
    local funcName=$2
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"'"${funcName}"'","payload":{"data":"'"${ID}"'"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" ${HTTP} "$resok" "$FUNCNAME"
}

listProposal() {
    local status=$1
    local funcName=$2
    local addr=""
    local direct=0
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"'"${funcName}"'","payload":{"status":"'"${status}"'", "addr":"'"${addr}"'", "count":1, "direction":"'"${direct}"'"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" ${HTTP} "$resok" "$FUNCNAME"
}

queryActivePropBoard() {
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"GetActiveBoard","payload":{"data":"1"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" ${HTTP} "$resok" "$FUNCNAME"
}

testProposalBoard() {
    #proposal
    chain33_LastBlockHeight ${HTTP}
    start=$((LAST_BLOCK_HEIGHT + 10))
    end=$((start + 20 + 720))
    proposalBoardTx ${start} ${end}
    #vote
    chain33_BlockWait 10 "$HTTP"
    voteBoardTx "${proposalID}" "${votePrKey}"
    #query
    queryProposal "${proposalID}" "GetProposalBoard"
    listProposal 4 "ListProposalBoard"
    queryActivePropBoard
    #test revoke
    chain33_LastBlockHeight ${HTTP}
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalBoardTx ${start} ${end}
    revokeProposalTx "${proposalID}" "RvkPropBoard"
    terminateProposalTx "${proposalID}" "TmintPropBoard"
    queryProposal "${proposalID}" "GetProposalBoard"
    listProposal 2 "ListProposalBoard"
}

proposalRuleTx() {
    local start=$1
    local end=$2
    local propAmount=$3
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropRule", "payload":{"ruleCfg": {"proposalAmount" : '"${propAmount}"'},"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    proposalID=$RAW_TX_HASH
    echo "$proposalID"
    echo_rst "proposalRule query_tx" "$?"
}

voteRuleTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropRule", "payload":{"proposalID": "'"${ID}"'","approve": true}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${HTTP}"
    echo "$RAW_TX_HASH"
    echo_rst "voteRule query_tx" "$?"
}

queryActivePropRule() {
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"GetActiveRule","payload":{"data":"1"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" ${HTTP} "$resok" "$FUNCNAME"
}

testProposalRule() {
    # proposal
    chain33_LastBlockHeight ${HTTP}
    start=$((LAST_BLOCK_HEIGHT + 10))
    end=$((start + 20 + 720))
    proposalRuleTx ${start} ${end} 2000000000
    #vote
    chain33_BlockWait 10 "$HTTP"
    voteRuleTx "${proposalID}" ${votePrKey}
    #query
    queryProposal "${proposalID}" "GetProposalRule"
    listProposal 4 "ListProposalRule"
    queryActivePropRule
    #test revoke
    chain33_LastBlockHeight ${HTTP}
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalRuleTx ${start} ${end} 2000000000
    revokeProposalTx "${proposalID}" "RvkPropRule"
    terminateProposalTx "${proposalID}" "TmintPropRule"
    queryProposal "${proposalID}" "GetProposalRule"
    listProposal 2 "ListProposalRule"
}

proposalProjectTx() {
    local start=$1
    local end=$2
    local amount=$3
    local toAddr=$4
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropProject", "payload":{"amount" : '"${amount}"', "toAddr" : "'"${toAddr}"'","startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    proposalID=$RAW_TX_HASH
    echo "$proposalID"
    echo_rst "proposalRule query_tx" "$?"
}

voteProjectTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropProject", "payload":{"proposalID": "'"${ID}"'","approve": true}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${HTTP}"
    echo "$RAW_TX_HASH"
    echo_rst "voteRule query_tx" "$?"
}

testProposalProject() {
    # proposal
    chain33_LastBlockHeight ${HTTP}
    start=$((LAST_BLOCK_HEIGHT + 10))
    end=$((start + 20 + 720))
    proposalProjectTx ${start} ${end} 100000000 ${propAddr}
    chain33_BlockWait 10 "$HTTP"
    #vote
    for ((i = 0; i < 11; i++)); do
        voteProjectTx "${proposalID}" "${boardsPrKey[$i]}"
    done
    #query
    queryProposal "${proposalID}" "GetProposalProject"
    listProposal 5 "ListProposalProject"
    #test revoke
    chain33_LastBlockHeight ${HTTP}
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalProjectTx ${start} ${end} 100000000 ${propAddr}
    revokeProposalTx "${proposalID}" "RvkPropProject"
    terminateProposalTx "${proposalID}" "TmintPropProject"
    queryProposal "${proposalID}" "GetProposalProject"
    listProposal 2 "ListProposalProject"
}

proposalChangeTx() {
    local start=$1
    local end=$2
    local addr=$3
    local cancel=$4
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropChange", "payload":{"changes" : [{"cancel": '"${cancel}"', "addr":"'"${addr}"'"}],"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    proposalID=$RAW_TX_HASH
    echo "$proposalID"
    echo_rst "proposalChange query_tx" "$?"
}

voteChangeTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropChange", "payload":{"proposalID": "'"${ID}"'","approve": true}}]}'
    echo "${req}"
    chain33_Http "$req" ${HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${HTTP}"
    echo "$RAW_TX_HASH"
    echo_rst "voteRule query_tx" "$?"
}

testProposalChange() {
    # proposal
    chain33_LastBlockHeight ${HTTP}
    start=$((LAST_BLOCK_HEIGHT + 10))
    end=$((start + 20 + 720))
    proposalChangeTx ${start} ${end} "${boardsAddr[20]}" true
    chain33_BlockWait 10 "$HTTP"
    #vote
    for ((i = 0; i < 14; i++)); do
        voteChangeTx "${proposalID}" "${boardsPrKey[$i]}"
    done
    #query
    queryProposal "${proposalID}" "GetProposalChange"
    listProposal 4 "ListProposalChange"
    #test revoke
    chain33_LastBlockHeight ${HTTP}
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalChangeTx ${start} ${end} "${boardsAddr[20]}" false
    revokeProposalTx "${proposalID}" "RvkPropChange"
    terminateProposalTx "${proposalID}" "TmintPropChange"
    queryProposal "${proposalID}" "GetProposalChange"
    listProposal 2 "ListProposalChange"
}

init() {
    ispara=$(echo '"'"${HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    if [ "$ispara" == true ]; then
        EXECTOR_ADDR=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.autonomy"}]}' ${HTTP} | jq -r ".result")
        EXECTOR="user.p.para.autonomy"
        TICKET_ADDR=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.ticket"}]}' ${HTTP} | jq -r ".result")
        TICKET_EXECTOR="user.p.para.ticket"
    else
        EXECTOR_ADDR=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"autonomy"}]}' ${HTTP} | jq -r ".result")
        EXECTOR="autonomy"
        TICKET_ADDR=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"ticket"}]}' ${HTTP} | jq -r ".result")
        TICKET_EXECTOR="ticket"
    fi
    echo "EXECTOR_ADDR=$EXECTOR_ADDR"

    local main_ip=${HTTP//8901/8801}
    chain33_ImportPrivkey "${propKey}" "${propAddr}" "prop" "${main_ip}"

    if [ "$ispara" == false ]; then
        chain33_applyCoinsNOLimit "$propAddr" 100000000000 "${main_ip}"
        chain33_QueryBalance "${propAddr}" "$main_ip"

    else
        chain33_applyCoins "$propAddr" 1000000000 "${main_ip}"
        chain33_QueryBalance "${propAddr}" "$main_ip"
        #主链投票账户转帐
        handleBoards "$main_ip"

        local para_ip="${HTTP}"
        chain33_ImportPrivkey "${propKey}" "${propAddr}" "prop" "$para_ip"

        #平行链中账户转帐
        chain33_applyCoinsNOLimit "$propAddr" 100000000000 "$para_ip"
        chain33_QueryBalance "$propAddr" "$para_ip"
        chain33_para_init "$para_ip"
    fi

    # 往合约中转
    chain33_SendToAddress "$propAddr" "$EXECTOR_ADDR" 90000000000 "$HTTP"
    chain33_QueryExecBalance "$propAddr" "$EXECTOR" "$HTTP"

    # 往ticket合约中转帐
    chain33_SendToAddress "$voteAddr" "$TICKET_ADDR" 300100000000 "$HTTP"
    chain33_QueryExecBalance "$voteAddr" "$TICKET_EXECTOR" "$HTTP"
    # 往投票账户中转帐
    handleBoards "$HTTP"
}

function run_testcases() {
    echo "run_testcases"
    testProposalRule
    testProposalBoard
    testProposalProject
    testProposalChange
}

function rpc_test() {
    chain33_RpcTestBegin autonomy

    HTTP="$1"
    echo "main_ip=$HTTP"

    init
    ispara=$(echo '"'"${HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    if [ "$ispara" == true ]; then
        echo "skip autonomy temporary on parachain"
    else
        run_testcases
    fi

    chain33_RpcTestRst autonomy "$CASE_ERR"

}

chain33_debug_function rpc_test "$1"
