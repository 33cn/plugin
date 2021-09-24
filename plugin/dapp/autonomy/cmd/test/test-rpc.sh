#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh
set -x

HTTP=""

EXECTOR=""
EXECTOR_ADDR=""

propKey="0xfd0c4a8a1efcd221ee0f36b7d4f57d8ff843cb8bc193b39c7863332d355acafa"
propAddr="15VUiygdxMSZ3rykwe742yomp2cPJ9Tfve"
#votePrKey="1c3e6cac2f887e1ab9180e2d5772dc4ba01accb8d4df434faba097003eb35482"
#voteAddr="1Q9sQwothzM1gKSzkVZ8Dt1tqKX1uzSagx"

voteAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
votePrKey="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" #14KEKbYtKKQm4wMthSK9J4La4nAiidGozt

voteAddr2="1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"
votePrKey2="B0BB75BC49A787A71F4834DA18614763B53A18291ECE6B5EDEC3AD19D150C3E7" #1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF

voteAddr3="1KcCVZLSQYRUwE5EXTsAoQs9LuJW6xwfQa"
votePrKey3="2AFF1981291355322C7A6308D46A9C9BA311AA21D94F36B43FC6A6021A1334CF"

proposalRuleID=""
proposalBoardID=""
proposalProjectID=""
proposalChangeID=""

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
    "168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9"
    "13KTf57aCkVVJYNJBXBBveiA5V811SrLcT"
    "1JQwQWsShTHC4zxHzbUfYQK4kRBriUQdEe"
    "1NHuKqoKe3hyv52PF8XBAyaTmJWAqA2Jbb"
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
    "0xcd284cd17456b73619fa609bb9e3105e8eff5d059c5e0b6eb1effbebd4d64144"
    "0xe892212221b3b58211b90194365f4662764b6d5474ef2961ef77c909e31eeed3"
    "0x9d19a2e9a440187010634f4f08ce36e2bc7b521581436a99f05568be94dc66ea"
    "0x45d4ce009e25e6d5e00d8d3a50565944b2e3604aa473680a656b242d9acbff35"
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
    chain33_QueryBalance "${targetAddr}" "${ip}"
}

handleBoards() {
    local ip=$1
    for ((i = 0; i < ${#boardsPrKey[*]}; i++)); do
        echo "${boardsPrKey[$i]}"
        lab="board_"${i}
        chain33_ImportPrivkey "${boardsPrKey[$i]}" "${boardsAddr[$i]}" "${lab}" "${ip}"
        chain33_applyCoins "${boardsAddr[$i]}" 5000000000 "${ip}"
        chain33_SendToAddress "${boardsAddr[$i]}" "$EXECTOR_ADDR" 4000000000 "$HTTP"
    done
    # 金额要转入合约中
}

txQuery() {
    ty=$(curl -ksd '{"method":"Chain33.QueryTransaction","params":[{"hash":"'"$RAW_TX_HASH"'"}]}' "${HTTP}" | jq -r ".result.receipt.ty")
    if [[ ${ty} != 2 ]]; then
        txQueryShow=$(curl -ksd '{"method":"Chain33.QueryTransaction","params":[{"hash":"'"$RAW_TX_HASH"'"}]}' "${HTTP}" | jq -r ".result")
        echo "$txQueryShow"
        echo_rst "$1 query_tx" 1
    fi
}

proposalBoardTx() {
    local start=$1
    local end=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropBoard", "payload":{"boardUpdate": 3,"boards": ['"${boards}"'],"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    proposalBoardID=$RAW_TX_HASH
    echo "proposalBoardID = $proposalBoardID"
    txQuery "$FUNCNAME"
}

voteBoardTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropBoard", "payload":{"proposalID": "'"${ID}"'","voteOption":1}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${HTTP}"
    echo "$RAW_TX_HASH"
    txQuery "$FUNCNAME"
}

revokeProposalTx() {
    local ID=$1
    local funcName=$2
    local key=$3
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"'"${funcName}"'", "payload":{"proposalID": "'"${ID}"'"}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${key}" "${HTTP}"
    echo "$RAW_TX_HASH"
    txQuery "$FUNCNAME"
}

terminateProposalTx() {
    local ID=$1
    local funcName=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"'"${funcName}"'", "payload":{"proposalID": "'"${ID}"'"}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    echo "$RAW_TX_HASH"
    txQuery "$FUNCNAME"
}

queryProposal() {
    local ID=$1
    local funcName=$2
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"'"${funcName}"'","payload":{"data":"'"${ID}"'"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" "${HTTP}" "$resok" "$FUNCNAME"
}

listProposal() {
    local status=$1
    local funcName=$2
    local addr=""
    local direct=0
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"'"${funcName}"'","payload":{"status":"'"${status}"'", "addr":"'"${addr}"'", "count":1, "direction":"'"${direct}"'"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" "${HTTP}" "$resok" "$FUNCNAME ${funcName}"
}

queryActivePropBoard() {
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"GetActiveBoard","payload":{"data":"1"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" "${HTTP}" "$resok" "$FUNCNAME"
}

testProposalBoard() {
    #proposal
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 20 + 720))
    proposalBoardTx ${start} ${end}
    #vote
    chain33_BlockWait 100 "$HTTP"
    voteBoardTx "${proposalBoardID}" "${votePrKey}"
    voteBoardTx "${proposalBoardID}" "${votePrKey2}"
    #query
    queryProposal "${proposalBoardID}" "GetProposalBoard"
    listProposal 4 "ListProposalBoard"
    queryActivePropBoard

    #test revoke
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalBoardTx ${start} ${end}
    revokeProposalTx "${proposalBoardID}" "RvkPropBoard" "${propKey}"
    queryProposal "${proposalBoardID}" "GetProposalBoard"
    listProposal 2 "ListProposalBoard"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

proposalRuleTx() {
    local start=$1
    local end=$2
    local propAmount=$3
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropRule", "payload":{"ruleCfg": {"proposalAmount" : '"${propAmount}"',"boardApproveRatio":50,"pubOpposeRatio":33,"largeProjectAmount":100000000000000,"publicPeriod":172800,"pubAttendRatio":60,"pubApproveRatio":60},"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    proposalRuleID=$RAW_TX_HASH
    echo "proposalRuleID = $proposalRuleID"
    txQuery "$FUNCNAME"
}

voteRuleTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropRule", "payload":{"proposalID": "'"${ID}"'", "vote":1}}]}' # "approve": true,
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${HTTP}"
    echo "$RAW_TX_HASH"
    txQuery "$FUNCNAME"
}

queryActivePropRule() {
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"GetActiveRule","payload":{"data":"1"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" "${HTTP}" "$resok" "$FUNCNAME"
}

testProposalRule() {
    # proposal
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 20 + 720))
    proposalRuleTx ${start} ${end} 2000000000
    #vote
    chain33_BlockWait 100 "$HTTP"
    voteRuleTx "${proposalRuleID}" ${votePrKey}
    voteRuleTx "${proposalRuleID}" "${votePrKey2}"
    voteRuleTx "${proposalRuleID}" "${votePrKey3}"
    #query
    queryProposal "${proposalRuleID}" "GetProposalRule"
    listProposal 4 "ListProposalRule"
    queryActivePropRule

    #test revoke
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalRuleTx ${start} ${end} 2000000000
    revokeProposalTx "${proposalRuleID}" "RvkPropRule" "${propKey}"
    queryProposal "${proposalRuleID}" "GetProposalRule"
    listProposal 2 "ListProposalRule"

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

proposalProjectTx() {
    local start=$1
    local end=$2
    local amount=$3
    local toAddr=$4
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropProject", "payload":{"amount" : '"${amount}"', "toAddr" : "'"${toAddr}"'","startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${HTTP}"
    proposalProjectID=$RAW_TX_HASH
    echo "proposalProjectID = $proposalProjectID"
    txQuery "$FUNCNAME"
}

voteProjectTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropProject", "payload":{"proposalID": "'"${ID}"'","vote":1}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${HTTP}"
    echo "$RAW_TX_HASH"
    txQuery "$FUNCNAME"
}

testProposalProject() {
    # proposal
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 20 + 720))
    proposalProjectTx ${start} ${end} 100000000 ${propAddr}
    chain33_BlockWait 100 "$HTTP"
    #vote
    for ((i = 0; i < 11; i++)); do
        voteProjectTx "${proposalProjectID}" "${boardsPrKey[$i]}"
    done
    #query
    queryProposal "${proposalProjectID}" "GetProposalProject"
    listProposal 5 "ListProposalProject"
    #test revoke
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalProjectTx ${start} ${end} 100000000 ${propAddr}
    revokeProposalTx "${proposalProjectID}" "RvkPropProject" "${propKey}"
    queryProposal "${proposalProjectID}" "GetProposalProject"
    listProposal 2 "ListProposalProject"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

proposalChangeTx() {
    local start=$1
    local end=$2
    local addr=$3
    local cancel=$4
    local key=$5
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropChange", "payload":{"changes" : [{"cancel": '"${cancel}"', "addr":"'"${addr}"'"}],"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${key}" "${HTTP}"
    proposalChangeID=$RAW_TX_HASH
    echo "proposalChangeID = $proposalChangeID"
    txQuery "$FUNCNAME"
}

voteChangeTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropChange", "payload":{"proposalID": "'"${ID}"'","vote":1}}]}'
    echo "${req}"
    chain33_Http "$req" "${HTTP}" '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${HTTP}"
    echo "$RAW_TX_HASH"
    txQuery "$FUNCNAME"
}

testProposalChange() {
    # proposal
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 20 + 720))
    proposalChangeTx ${start} ${end} "${boardsAddr[21]}" true "${boardsPrKey[0]}"
    chain33_BlockWait 100 "$HTTP"
    #vote
    for ((i = 0; i < 11; i++)); do
        voteChangeTx "${proposalChangeID}" "${boardsPrKey[$i]}"
    done
    #query
    queryProposal "${proposalChangeID}" "GetProposalChange"
    listProposal 4 "ListProposalChange"

    #test revoke
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalChangeTx ${start} ${end} "${boardsAddr[22]}" true "${boardsPrKey[1]}"
    revokeProposalTx "${proposalChangeID}" "RvkPropChange" "${boardsPrKey[1]}"
    queryProposal "${proposalChangeID}" "GetProposalChange"
    listProposal 2 "ListProposalChange"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

testProposalTerminate() {
    #test terminate
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalRuleTx ${start} ${end} 2000000000

    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalBoardTx ${start} ${end}

    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalProjectTx ${start} ${end} 100000000 ${propAddr}

    chain33_LastBlockHeight "${HTTP}"
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalChangeTx ${start} ${end} "${boardsAddr[23]}" true "${boardsPrKey[2]}"

    chain33_BlockWait 940 "$HTTP"

    terminateProposalTx "${proposalRuleID}" "TmintPropRule"
    queryProposal "${proposalRuleID}" "GetProposalRule"
    listProposal 4 "ListProposalRule"

    terminateProposalTx "${proposalBoardID}" "TmintPropBoard"
    queryProposal "${proposalBoardID}" "GetProposalBoard"
    listProposal 4 "ListProposalBoard"

    terminateProposalTx "${proposalProjectID}" "TmintPropProject"
    queryProposal "${proposalProjectID}" "GetProposalProject"
    listProposal 5 "ListProposalProject"

    terminateProposalTx "${proposalChangeID}" "TmintPropChange"
    queryProposal "${proposalChangeID}" "GetProposalChange"
    listProposal 4 "ListProposalChange"
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function run_testcases() {
    echo "run_testcases"

    testProposalRule
    testProposalBoard
    testProposalProject
    testProposalChange
    testProposalTerminate
}

init() {
    ispara=$(echo '"'"${HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    if [ "$ispara" == true ]; then
        EXECTOR_ADDR=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.autonomy"}]}' "${HTTP}" | jq -r ".result")
        EXECTOR="user.p.para.autonomy"
    else
        EXECTOR_ADDR=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"autonomy"}]}' "${HTTP}" | jq -r ".result")
        EXECTOR="autonomy"
    fi
    echo "EXECTOR_ADDR=$EXECTOR_ADDR"

    local main_ip=${HTTP//8901/8801}
    chain33_ImportPrivkey "${propKey}" "${propAddr}" "prop" "${main_ip}"
    chain33_ImportPrivkey "${votePrKey2}" "${voteAddr2}" "voteAddr2" "${main_ip}"
    chain33_ImportPrivkey "${votePrKey3}" "${voteAddr3}" "voteAddr3" "${main_ip}"

    if [ "$ispara" == false ]; then
        chain33_applyCoinsNOLimit "$propAddr" 100000000000 "${main_ip}"
        chain33_applyCoinsNOLimit "${voteAddr}" 10000000000 "${main_ip}"
        chain33_applyCoinsNOLimit "${voteAddr2}" 10000000000 "${main_ip}"
        chain33_applyCoinsNOLimit "${voteAddr3}" 10000000000 "${main_ip}"
    else
        chain33_applyCoinsNOLimit "$propAddr" 1000000000 "${main_ip}"
        #主链投票账户转帐
        handleBoards "$main_ip"

        local para_ip="${HTTP}"
        chain33_ImportPrivkey "${propKey}" "${propAddr}" "prop" "$para_ip"
        chain33_ImportPrivkey "${votePrKey2}" "${voteAddr2}" "voteAddr2" "${para_ip}"
        chain33_ImportPrivkey "${votePrKey3}" "${voteAddr3}" "voteAddr3" "${para_ip}"

        #平行链中账户转帐
        chain33_applyCoinsNOLimit "$propAddr" 100000000000 "$para_ip"
        chain33_applyCoinsNOLimit "${voteAddr}" 10000000000 "${para_ip}"
        chain33_applyCoinsNOLimit "${voteAddr2}" 10000000000 "${para_ip}"
        chain33_applyCoinsNOLimit "${voteAddr3}" 10000000000 "${para_ip}"
    fi

    # 往合约中转
    chain33_SendToAddress "$propAddr" "$EXECTOR_ADDR" 90000000000 "$HTTP"
    chain33_QueryExecBalance "$propAddr" "$EXECTOR" "$HTTP"

    # 往投票账户中转帐
    handleBoards "$HTTP"
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
