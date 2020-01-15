#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
source ../dapp-test-common.sh

MAIN_HTTP=""

EXECTOR=""
EXECTOR_ADDR=""

propKey="0xfd0c4a8a1efcd221ee0f36b7d4f57d8ff843cb8bc193b39c7863332d355acafa"
propAddr="15VUiygdxMSZ3rykwe742yomp2cPJ9Tfve"
votePrKey="1c3e6cac2f887e1ab9180e2d5772dc4ba01accb8d4df434faba097003eb35482"
voteAddr="1Q9sQwothzM1gKSzkVZ8Dt1tqKX1uzSagx"

proposalID=""

#boards='
#"1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj",
#"12cjnN5D4DPdBQSwh6vjwJbtsW4EJALTMv",
#"1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH",
#"1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB",
#"1L1puAUjfmtDECKo2C1qLWsAMZtDGTBWf6",
#"1LNf9AVXzUMQkQM5hgGLhkdrVtD8UMBSUm",
#"1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu",
#"1DyR84CU5AHbGXLEnhHMwMvWNMeunLZsuJ",
#"132pBvrgSYgHASxzoeL3bqnsqUpaBbUktm",
#"1DEV4XwdBUWRkMuy4ARRiEAoxQ2LoDByNG",
#"18Y87cw2hiYC71bvpD872oYMYXtw66Qp6o",
#"1Fghq6cgdJEDr6gQBmvba3t6aXAkyZyjr2",
#"142KsfJLvEA5FJxAgKm9ZDtFVjkRaPdu82",
#"1MAuE8QSbbech3bVKK2JPJJxYxNtT95oSU",
#"14Cuq8Ltx8a88PCLbSVPKZqcsY7Xi7bKLg",
#"155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6",
#"1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3",
#"13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv",
#"1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum",
#"113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG",
#"1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"
#'

boardsAddr=(
"1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj"
"12cjnN5D4DPdBQSwh6vjwJbtsW4EJALTMv"
"1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH"
"1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB"
"1L1puAUjfmtDECKo2C1qLWsAMZtDGTBWf6"
"1LNf9AVXzUMQkQM5hgGLhkdrVtD8UMBSUm"
"1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu"
"1DyR84CU5AHbGXLEnhHMwMvWNMeunLZsuJ"
"132pBvrgSYgHASxzoeL3bqnsqUpaBbUktm"
"1DEV4XwdBUWRkMuy4ARRiEAoxQ2LoDByNG"
"18Y87cw2hiYC71bvpD872oYMYXtw66Qp6o"
"1Fghq6cgdJEDr6gQBmvba3t6aXAkyZyjr2"
"142KsfJLvEA5FJxAgKm9ZDtFVjkRaPdu82"
"1MAuE8QSbbech3bVKK2JPJJxYxNtT95oSU"
"14Cuq8Ltx8a88PCLbSVPKZqcsY7Xi7bKLg"
"155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6"
"1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3"
"13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv"
"1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum"
"113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG"
"1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"
)
boardsPrKey=(
"d165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5"
"c21d38be90493512a5c2417d565269a8b23ce8152010e404ff4f75efead8183a"
"fdf2bbff853ecff2e7b86b2a8b45726c6538ca7d1403dc94e50131ef379bdca0"
"794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d"
"f2cc48d30560e4c92e84821df68cf1086de82ee6a5725fc2a590a58d6ffe4fc5"
"eb4738a7c685a7ccf5471c3335a2d7ebe284b11d8a1717d814904b8d1ba936d9"
"128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1"
"4c9691bf6acc908ef5c07abcad23cf7f98e46e84101aa5059322aa53eb4dc471"
"50b9c6a4358ef8ffc96d5831a8dfd5e0fae504d49e20c5eafd12b6015423de33"
"96e3c766850a915fe4718b890d96208d5d1a3694b2597e08165480b5b48b84cb"
"eac5e45243c3920cf8a98f3d3a2e3a9b43f30a21769b57f734213913511e7575"
"d2aaa6f050a4db13fbd2c8bf87cbb96e217289172baca6c12e8a8b0680b9aa1a"
"33b3b977c657435a49773b5605a704ad5fdca0fa34fe36a02ea0f13a49099832"
"24d1fad138be98eebee31440f144aa38c404533f40862995282162bc538e91c8"
"15915d94b5e0f112b8e4002b545a7c230011e54b40dce665ce843ee1a3f577ad"
"9d539bc5fd084eb7fe86ad631dba9aa086dba38418725c38d9751459f567da66"
"22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962"
"0a6671f101e30a2cc2d79d77436b62cdf2664ed33eb631a9c9e3f3dd348a23be"
"ec9162ea5fc2f473ab8240619a0a0f495ba9e9e5d4d9c434b8794a68280236c4"
"3818b257b05ee75b6e43ee0e3cfc2d8502342cf67caed533e3756966690b62a5"
"d627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
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
}

handleBoards() {
    local main_ip=${MAIN_HTTP//8901/8801}
    chain33_ImportPrivkey "${propKey}" "${propAddr}" "prop" "${main_ip}"
    for ((i = 0; i < ${#boardsPrKey[*]}; i++)); do
      echo "${boardsPrKey[$i]}"
      lab="board_"${i}
      chain33_ImportPrivkey "${boardsPrKey[$i]}" "${boardsAddr[$i]}" "${lab}" "${main_ip}"
      chain33_applyCoins "${boardsAddr[$i]}" 100000000 "${main_ip}"
    done
}

proposalBoardTx() {
    local start=$1
    local end=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropBoard", "payload":{"boards": ['"${boards}"'],"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${MAIN_HTTP}"
    proposalID=$RAW_TX_HASH
    echo $proposalID
    echo_rst "proposalBoard query_tx" "$?"
}

voteBoardTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropBoard", "payload":{"proposalID": "'"${ID}"'","approve": true}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${MAIN_HTTP}"
    echo $RAW_TX_HASH
    echo_rst "voteBoard query_tx" "$?"
}

revokeProposalTx() {
    local ID=$1
    local funcName=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"'"${funcName}"'", "payload":{"proposalID": "'"${ID}"'"}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${MAIN_HTTP}"
    echo $RAW_TX_HASH
    echo_rst "revoke Proposal $funcName query_tx" "$?"
}

terminateProposalTx() {
    local ID=$1
    local funcName=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"'"${funcName}"'", "payload":{"proposalID": "'"${ID}"'"}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${MAIN_HTTP}"
    echo $RAW_TX_HASH
    echo_rst "terminate Proposal $funcName query_tx" "$?"
}

queryProposal() {
    local ID=$1
    local funcName=$2
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"'"${funcName}"'","payload":{"data":"'"${ID}"'"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

listProposal() {
    local status=$1
    local funcName=$2
    local addr=""
    local direct=0
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"'"${funcName}"'","payload":{"status":"'"${status}"'", "addr":"'"${addr}"'", "count":1, "direction":"'"${direct}"'"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

queryActivePropBoard() {
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"GetActiveBoard","payload":{"data":"1"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

testProposalBoard() {
    #proposal
    chain33_LastBlockHeight ${MAIN_HTTP}
    start=$((LAST_BLOCK_HEIGHT + 10))
    end=$((start + 20 + 720))
    proposalBoardTx ${start} ${end}
    #vote
    chain33_BlockWait 10 "$MAIN_HTTP"
    voteBoardTx ${proposalID} ${votePrKey}
    #query
    queryProposal ${proposalID} "GetProposalBoard"
    listProposal 4 "ListProposalBoard"
    queryActivePropBoard
    #test revoke
    chain33_LastBlockHeight ${MAIN_HTTP}
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalBoardTx ${start} ${end}
    revokeProposalTx ${proposalID} "RvkPropBoard"
    terminateProposalTx ${proposalID} "TmintPropBoard"
    queryProposal ${proposalID} "GetProposalBoard"
}

proposalRuleTx() {
    local start=$1
    local end=$2
    local propAmount=$3
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropRule", "payload":{"ruleCfg": {"proposalAmount" : '"${propAmount}"'},"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${MAIN_HTTP}"
    proposalID=$RAW_TX_HASH
    echo $proposalID
    echo_rst "proposalRule query_tx" "$?"
}

voteRuleTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropRule", "payload":{"proposalID": "'"${ID}"'","approve": true}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${MAIN_HTTP}"
    echo $RAW_TX_HASH
    echo_rst "voteRule query_tx" "$?"
}

queryActivePropRule() {
    local req='{"method":"Chain33.Query","params":[{"execer":"'"${EXECTOR}"'","funcName":"GetActiveRule","payload":{"data":"1"}}]}'
    resok='(.error|not)'
    chain33_Http "$req" ${MAIN_HTTP} "$resok" "$FUNCNAME"
}

testProposalRule() {
    # proposal
    chain33_LastBlockHeight ${MAIN_HTTP}
    start=$((LAST_BLOCK_HEIGHT + 10))
    end=$((start + 20 + 720))
    proposalRuleTx ${start} ${end} 2000000000
    #vote
    chain33_BlockWait 10 "$MAIN_HTTP"
    voteRuleTx ${proposalID} ${votePrKey}
    #query
    queryProposal ${proposalID} "GetProposalRule"
    listProposal 4 "ListProposalRule"
    queryActivePropRule
    #test revoke
    chain33_LastBlockHeight ${MAIN_HTTP}
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalRuleTx ${start} ${end} 2000000000
    revokeProposalTx ${proposalID} "RvkPropRule"
    terminateProposalTx ${proposalID} "TmintPropRule"
    queryProposal ${proposalID} "GetProposalRule"
}

proposalProjectTx() {
    local start=$1
    local end=$2
    local amount=$3
    local toAddr=$4
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropProject", "payload":{"amount" : '"${amount}"', "toAddr" : "'"${toAddr}"'","startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${MAIN_HTTP}"
    proposalID=$RAW_TX_HASH
    echo $proposalID
    echo_rst "proposalRule query_tx" "$?"
}

voteProjectTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropProject", "payload":{"proposalID": "'"${ID}"'","approve": true}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${MAIN_HTTP}"
    echo $RAW_TX_HASH
    echo_rst "voteRule query_tx" "$?"
}

testProposalProject() {
    # proposal
    chain33_LastBlockHeight ${MAIN_HTTP}
    start=$((LAST_BLOCK_HEIGHT + 10))
    end=$((start + 20 + 720))
    proposalProjectTx ${start} ${end} 100000000 ${propAddr}
    chain33_BlockWait 10 "$MAIN_HTTP"
    #vote
    for ((i = 0; i < 11; i++)); do
      voteProjectTx ${proposalID} "${boardsPrKey[$i]}"
    done
    #query
    queryProposal ${proposalID} "GetProposalProject"
    listProposal 5 "ListProposalProject"
    #test revoke
    chain33_LastBlockHeight ${MAIN_HTTP}
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalProjectTx ${start} ${end} 100000000 ${propAddr}
    revokeProposalTx ${proposalID} "RvkPropProject"
    terminateProposalTx ${proposalID} "TmintPropProject"
    queryProposal ${proposalID} "GetProposalProject"
}

proposalChangeTx() {
    local start=$1
    local end=$2
    local addr=$3
    local cancel=$4
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"PropChange", "payload":{"changes" : [{"cancel": '"${cancel}"', "addr":"'"${addr}"'"}],"startBlockHeight":'"${start}"',"endBlockHeight":'"${end}"'}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${propKey}" "${MAIN_HTTP}"
    proposalID=$RAW_TX_HASH
    echo $proposalID
    echo_rst "proposalChange query_tx" "$?"
}

voteChangeTx() {
    local ID=$1
    local privk=$2
    local req='{"method":"Chain33.CreateTransaction","params":[{"execer":"'"${EXECTOR}"'", "actionName":"VotePropChange", "payload":{"proposalID": "'"${ID}"'","approve": true}}]}'
    echo "${req}"
    chain33_Http "$req" ${MAIN_HTTP} '(.error|not) and (.result != null)' "$FUNCNAME" ".result"
    chain33_SignAndSendTx "${RETURN_RESP}" "${privk}" "${MAIN_HTTP}"
    echo $RAW_TX_HASH
    echo_rst "voteRule query_tx" "$?"
}

testProposalChange() {
    # proposal
    chain33_LastBlockHeight ${MAIN_HTTP}
    start=$((LAST_BLOCK_HEIGHT + 10))
    end=$((start + 20 + 720))
    proposalChangeTx ${start} ${end} "${boardsAddr[20]}" true
    chain33_BlockWait 10 "$MAIN_HTTP"
    #vote
    for ((i = 0; i < 14; i++)); do
      voteChangeTx ${proposalID} "${boardsPrKey[$i]}"
    done
    #query
    queryProposal ${proposalID} "GetProposalChange"
    listProposal 4 "ListProposalChange"
    #test revoke
    chain33_LastBlockHeight ${MAIN_HTTP}
    start=$((LAST_BLOCK_HEIGHT + 100))
    end=$((start + 120 + 720))
    proposalChangeTx ${start} ${end} "${boardsAddr[20]}" false
    revokeProposalTx ${proposalID} "RvkPropChange"
    terminateProposalTx ${proposalID} "TmintPropChange"
    queryProposal ${proposalID} "GetProposalChange"
}

init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"

    if [ "$ispara" == true ]; then
        EXECTOR_ADDR=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.autonomy"}]}' ${MAIN_HTTP} | jq -r ".result")
        EXECTOR="user.p.para.autonomy"
    else
        EXECTOR_ADDR=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"autonomy"}]}' ${MAIN_HTTP} | jq -r ".result")
        EXECTOR="autonomy"
    fi
    echo "EXECTOR_ADDR=$EXECTOR_ADDR"


    local main_ip=${MAIN_HTTP//8901/8801}
    chain33_ImportPrivkey "${propKey}" "${propAddr}" "prop" "${main_ip}"

    if [ "$ispara" == false ]; then
        chain33_applyCoinsNOLimit "$propAddr" 100000000000 "${main_ip}"
        chain33_QueryBalance "${propAddr}" "$main_ip"

    else
        chain33_applyCoins "$propAddr" 1000000000 "${main_ip}"
        chain33_QueryBalance "${propAddr}" "$main_ip"

        local para_ip="${MAIN_HTTP}"
        chain33_ImportPrivkey "${propKey}" "${propAddr}" "prop" "$para_ip"

        #平行链中账户转帐
        chain33_applyCoinsNOLimit "$propAddr" 100000000000 "${para_ip}"
        chain33_QueryBalance "${propAddr}" "$para_ip"
    fi

    # 往合约中转
    chain33_SendToAddress "$propAddr" "$EXECTOR_ADDR" 90000000000 "${MAIN_HTTP}"
    chain33_QueryExecBalance "${propAddr}" "autonomy" "$MAIN_HTTP"

    # 往ticket合约中转帐
    chain33_SendToAddress "$voteAddr" "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp" 300100000000 "${MAIN_HTTP}"
    chain33_QueryExecBalance "$voteAddr" "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp" "$MAIN_HTTP"
}

function run_testcases() {
    echo "run_testcases"
    handleBoards
    testProposalRule
    testProposalBoard
    testProposalProject
    testProposalChange
}

function rpc_test() {
    chain33_RpcTestBegin autonomy

    MAIN_HTTP="$1"
    echo "main_ip=$MAIN_HTTP"

    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    if [ "$ispara" == true ]; then
        return 0
    fi

    init
    run_testcases

    chain33_RpcTestRst autonomy "$CASE_ERR"

}

chain33_debug_function rpc_test "$1"

#chain33_debug_function rpc_test  "http://127.0.0.1:8801"

#ImpBoards