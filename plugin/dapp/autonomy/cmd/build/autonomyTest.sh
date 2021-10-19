#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

source "./publicTest.sh"

# shellcheck disable=SC2034
{
    propKey="0xfd0c4a8a1efcd221ee0f36b7d4f57d8ff843cb8bc193b39c7863332d355acafa"
    propAddr="15VUiygdxMSZ3rykwe742yomp2cPJ9Tfve"

    votePrKey="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" #14KEKbYtKKQm4wMthSK9J4La4nAiidGozt

    voteAddr2="1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"
    votePrKey2="B0BB75BC49A787A71F4834DA18614763B53A18291ECE6B5EDEC3AD19D150C3E7" #1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF

    voteAddr3="1KcCVZLSQYRUwE5EXTsAoQs9LuJW6xwfQa"
    votePrKey3="2AFF1981291355322C7A6308D46A9C9BA311AA21D94F36B43FC6A6021A1334CF"

    boardsAddr="1N578zmVzVR7RxLfnp7XAeDmAy499Jw3q2
, 1DZ1kL9x3rRwz7EZjcLt1kMYu6Zdp3MjGR
, 1HUYR1Mzb91m3dmsEE1vPrv7BsAHmEtVzM
, 1JHmVgchSLjszN9LAYa3gds811c4BH2J51
, 14TDcn95hxHbpySPtxGDK3aY6qDNuk5idg
, 1EpbYadEAcwbrxuh6Ph2qyJZhMY9F9CCCv
, 1NqfXb3YotDTPuShgSAhrBH28ETzCrsZx2
, 138xSezQUi9kynnAZR23kWBwHvDSRX8JDK
, 18bncBidkwRaUYPaqa9Rc58tWeRT2C5tfL
, 16z3gEvRGof8cnQRvicV8BdY54fUfnBo8A
, 1DpzRWk9SYnEvuGrFsJPFvMVRSYRHaCbtU
, 1JSQbsb1hfB6CLzwzVZZEtai4da8m59obK
, 12bRYKMsLSbN5upxoSenypWrygVuQh8rM5
, 1Ew4Cpcfh1u1EeRkYPwVc4PRaSnz4D2eph
, 156yPWN3eeZ4yPiwHRifu41FmrTVPavXQ5
, 192YKhdAkFm18KLcGumM8JeDgWrnpSdo93
, 12kXgBxsKzUmhfX5Mnv9EKRNGjQXhZzDFd
, 1HmyNGYj2xyMmQDTpcuheLKXKBe5rj3QpQ
, 1HGPrjc6H7yBzFV5yCbibvnSUGUgdDNQi3
, 19WGov4b7wLf4f8JHMRDnJsGVNMDzap38w
, 1HmRa1jAnzJ5SpJRrUWqUki6hx1u33Nbq4
, 168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9
, 13KTf57aCkVVJYNJBXBBveiA5V811SrLcT
, 1JQwQWsShTHC4zxHzbUfYQK4kRBriUQdEe
, 1NHuKqoKe3hyv52PF8XBAyaTmJWAqA2Jbb"
    # shellcheck disable=SC2206
    arrayAddr=(${boardsAddr//,/ })

    lenAddr=${#arrayAddr[@]}
    for ((i = 0; i < lenAddr - 1; i++)); do
        boards=$(echo "${boards}${arrayAddr[i]}-")
    done
    # shellcheck disable=SC2116
    boards=$(echo "${boards}${arrayAddr[lenAddr - 1]}")

    boardsPrKey="fa54751118c8159ade22c253f85945a4dd2030b1cf2502eaf785d0a4f5ad7e35
,da9371ea52f1fc9d72e75dbc9836774895cd0966fd53c83f5e2c92d878903693
,ad9731261c40c68fee96f7b846408fa33d1f3dc2a27bdb3694ec8f3aa153a98b
,e902d23ad26052cec64e9ed9055853327787b3bf26eb4646a6d6c1bb516f9fcd
,b3af59368bcc6779aecb4e7fbf0cc4040f8f8329eb7632d580be4b6d3be15357
,fd9284c11707b571e347b8f44c54aae89c0810d410a7c1f9613326358c564a5e
,b013948c123986aadc2525bfe9935dd07972e14b250b938153e763f917ada8d3
,5641e3aba9ce0660665cd1a816d1b50267b5bc5f337f73e6ceed0cc5ececa7d2
,5a7631b7101252d685bf2b4ba2f11a72fb867faeaa545064ce8d0901ffc3cb17
,c9d0d2639c0c0f2b275e5bf8a2797122af110316a4e8e1fe03c39693f5028c93
,00b4e3ac1365b89d68c8d5b07b3505cd614408b0d9c8c88df564d4e072deb401
,aa139d3f16c1785b2a9171d7863bc4ee9d45115cc0fa71a14c43536c933d1659
,33f111573a4613477f8291928a9bae012a74fc9858acb22c9c65ecf7844b63f2
,44e43dd0f769bb99638b9cc3c7468225ff703f69c17e9c4751d04e45fdc6c4a0
,9a75d6c779846fe2ad4a36021fe9d08652ca69ce09ab39ad874d921a5ec41716
,f5ca6b2ad545bd4b854871b18c8d37d2fe8c3625e91be86a204c4086f28e8d0f
,89504608a03590e5a4d8c1c82b75e908b28f9963587c85b96b628d238d3a4d1a
,fa653545ae52403665fb803ef410c4d3d7f74628b6d3f92218968ad496e4f81b
,227df96a414e26e85c7d87a12296344e6a731ce73e424ba9845cb305ec963843
,64259075bf2e5a74334442f5048ceaf8427f6097e2dac99c0e463785c7768550
,3b812d92b5c365d698255f55c3f0dca027c2f89f2409c51a6297bcf3343c11e6
,cd284cd17456b73619fa609bb9e3105e8eff5d059c5e0b6eb1effbebd4d64144
,e892212221b3b58211b90194365f4662764b6d5474ef2961ef77c909e31eeed3
,9d19a2e9a440187010634f4f08ce36e2bc7b521581436a99f05568be94dc66ea
,45d4ce009e25e6d5e00d8d3a50565944b2e3604aa473680a656b242d9acbff35"
    # shellcheck disable=SC2206
    arrayKey=(${boardsPrKey//,/ })

    changeKey="0x7503333e74190abdfa10b9e5e3a225136ea7ed1f5beeeb710b1f3933a083b2c5"
    changeAddr="15MP8oQXW9UuzbfMsUyVy4ThStm9ZH7hgG"
    changeKey2="0x38d12b36d8d84a80131db15f2a067e2706f8367c933c075b3b3cefa3a864ad15"
    changeAddr2="1Kubv93zSatYRq6eHdBa4BkMBwtB2UpQum"

    minerAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"

    Chain33Cli="../../chain33-cli"
    proposalRuleID=""
    proposalBoardID=""
    proposalProjectID=""
    start=0
    end=0
    last_header=0

    boardApproveRatio=50
    pubOpposeRatio=33
    proposalAmount=2000000000
    largeProjectAmount=100000000000000
    publicPeriod=172800
    pubAttendRatio=60
    pubApproveRatio=65

    start_block=50
}

function update_last_header() {
    last_header=$(${Chain33Cli} block last_header | jq -r ".height")

    local h=100
    if [ "$#" -eq 1 ]; then
        h=$1
    fi

    start=$((last_header + h))
    end=$((start + 20 + 720))
}

function sign_and_send() {
    local raw="$1"
    local key="$2"
    data=$(${Chain33Cli} wallet sign -d "${raw}" -k "${key}")
    hash=$(${Chain33Cli} wallet send -d "${data}")
    check_tx "${Chain33Cli}" "${hash}"
    echo "${hash}"
}

function proposalRuleTx() {
    raw=$(${Chain33Cli} autonomy proposalRule -e "${end}" -s "${start}" -l 1000000 -p 20 -r ${boardApproveRatio} -a ${pubAttendRatio} -v ${pubApproveRatio} -o ${pubOpposeRatio} -u ${publicPeriod})
    sign_and_send "${raw}" "${propKey}"
    proposalRuleID="${hash}"
}

# $1 status
function showRule_status() {
    local status=$1
    result=$(${Chain33Cli} autonomy showRule -p "${proposalRuleID}" -y 0 | jq -r ".propRules[0].status")
    is_equal "${result}" "${status}"

    result=$(${Chain33Cli} autonomy showRule -p "${proposalRuleID}" -y 1 -s "${status}")
    is_not_equal "${result}" "ErrNotFound"
}

function check_activeRule() {
    result=$(${Chain33Cli} autonomy showActiveRule | jq -r ".boardApproveRatio")
    is_equal "${result}" "${boardApproveRatio}"
    result=$(${Chain33Cli} autonomy showActiveRule | jq -r ".pubOpposeRatio")
    is_equal "${result}" "${pubOpposeRatio}"
    result=$(${Chain33Cli} autonomy showActiveRule | jq -r ".publicPeriod")
    is_equal "${result}" "${publicPeriod}"
    result=$(${Chain33Cli} autonomy showActiveRule | jq -r ".pubAttendRatio")
    is_equal "${result}" "${pubAttendRatio}"
    result=$(${Chain33Cli} autonomy showActiveRule | jq -r ".pubApproveRatio")
    is_equal "${result}" "${pubApproveRatio}"
}

function testProposalRule() {
    # proposal
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    update_last_header "${start_block}"
    proposalRuleTx

    #vote
    block_wait "${Chain33Cli}" "${start_block}"
    raw=$(${Chain33Cli} autonomy voteRule -r 1 -p "${proposalRuleID}")
    sign_and_send "${raw}" "${votePrKey}"
    raw=$(${Chain33Cli} autonomy voteRule -r 1 -p "${proposalRuleID}")
    sign_and_send "${raw}" "${votePrKey2}"
    raw=$(${Chain33Cli} autonomy voteRule -r 1 -p "${proposalRuleID}")
    sign_and_send "${raw}" "${votePrKey3}"

    showRule_status 4
    check_activeRule

    #test revoke
    update_last_header
    proposalRuleTx
    raw=$(${Chain33Cli} autonomy revokeRule -p "${proposalRuleID}")
    sign_and_send "${raw}" "${propKey}"
    showRule_status 2

    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function proposalBoardTx() {
    raw=$(${Chain33Cli} autonomy proposalBoard -e "${end}" -s "${start}" -u 3 -b "${boards}")
    sign_and_send "${raw}" "${propKey}"
    proposalBoardID="${hash}"
}

# $1 status
function showBoard_status() {
    local status=$1
    result=$(${Chain33Cli} autonomy showBoard -p "${proposalBoardID}" -y 0 | jq -r ".propBoards[0].status")
    is_equal "${result}" "${status}"

    result=$(${Chain33Cli} autonomy showBoard -p "${proposalBoardID}" -y 1 -s "${status}")
    is_not_equal "${result}" "ErrNotFound"
}

function check_activeBoard() {
    ${Chain33Cli} autonomy showActiveBoard
    for ((i = 0; i < lenAddr; i++)); do
        ret=$(${Chain33Cli} autonomy showActiveBoard | jq -r ".boards[$i]")
        if [ "${arrayAddr[i]}" != "${ret}" ]; then
            exit 1
        fi
    done
}

function testProposalBoard() {
    #proposal
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    update_last_header "${start_block}"
    proposalBoardTx

    #vote
    block_wait "${Chain33Cli}" "${start_block}"
    raw=$(${Chain33Cli} autonomy voteBoard -r 1 -p "${proposalBoardID}")
    sign_and_send "${raw}" "${votePrKey}"
    raw=$(${Chain33Cli} autonomy voteBoard -r 1 -p "${proposalBoardID}")
    sign_and_send "${raw}" "${votePrKey2}"
    #query
    showBoard_status 4
    check_activeBoard

    #test revoke
    update_last_header
    proposalBoardTx
    raw=$(${Chain33Cli} autonomy revokeBoard -p "${proposalBoardID}")
    sign_and_send "${raw}" "${propKey}"
    showBoard_status 2
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function proposalProjectTx() {
    raw=$(${Chain33Cli} autonomy proposalProject -e "${end}" -s "${start}" -a 1 -o "${propAddr}")
    sign_and_send "${raw}" "${propKey}"
    proposalProjectID="${hash}"
}

function showProject_status() {
    local status=$1
    result=$(${Chain33Cli} autonomy showProject -p "${proposalProjectID}" -y 0 | jq -r ".propProjects[0].status")
    is_equal "${result}" "${status}"

    result=$(${Chain33Cli} autonomy showProject -p "${proposalProjectID}" -y 1 -s "${status}")
    is_not_equal "${result}" "ErrNotFound"
}

function testProposalProject() {
    # proposal
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    update_last_header "${start_block}"
    proposalProjectTx

    #vote
    block_wait "${Chain33Cli}" "${start_block}"
    for ((i = 0; i < 13; i++)); do
        raw=$(${Chain33Cli} autonomy voteProject -r 1 -p "${proposalProjectID}")
        sign_and_send "${raw}" "${arrayKey[$i]}"
    done
    #query
    showProject_status 5

    #test revoke
    update_last_header
    proposalProjectTx

    raw=$(${Chain33Cli} autonomy revokeProject -p "${proposalProjectID}")
    sign_and_send "${raw}" "${propKey}"

    showProject_status 2
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function proposalChangeTx() {
    local addr=$1
    local key=$2
    raw=$(${Chain33Cli} autonomy proposalChange -e "${end}" -s "${start}" -c "${addr}")
    sign_and_send "${raw}" "${key}"
    proposalChangeID="${hash}"
}

function showChange_status() {
    local status=$1

    result=$(${Chain33Cli} autonomy showChange -p "${proposalChangeID}" -y 0 | jq -r ".propChanges[0].status")
    is_equal "${result}" "${status}"

    result=$(${Chain33Cli} autonomy showChange -p "${proposalChangeID}" -y 1 -s "${status}")
    is_not_equal "${result}" "ErrNotFound"
}

function testProposalChange() {
    # proposal
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    #init
    autonomyAddr=$(${Chain33Cli} exec addr -e autonomy)
    hash=$(${Chain33Cli} send coins transfer -a 20 -n test -t "${autonomyAddr}" -k "${arrayKey[20]}")
    check_tx "${Chain33Cli}" "${hash}"
    hash=$(${Chain33Cli} send coins transfer -a 20 -n test -t "${autonomyAddr}" -k "${arrayKey[21]}")
    check_tx "${Chain33Cli}" "${hash}"

    update_last_header "${start_block}"
    proposalChangeTx "${changeAddr}" "${arrayKey[20]}"

    ret=$(${Chain33Cli} autonomy showActiveBoard | jq -r ".boards[20]")
    if [ "${arrayAddr[20]}" != "${ret}" ]; then
        exit 1
    fi

    #vote
    block_wait "${Chain33Cli}" "${start_block}"
    for ((i = 0; i < 13; i++)); do
        raw=$(${Chain33Cli} autonomy voteChange -r 1 -p "${proposalChangeID}")
        sign_and_send "${raw}" "${arrayKey[$i]}"
    done

    ret=$(${Chain33Cli} autonomy showActiveBoard | jq -r ".boards[20]")
    if [ "${changeAddr}" != "${ret}" ]; then
        exit 1
    fi

    #query
    showChange_status 4

    #test revoke
    update_last_header
    proposalChangeTx "${changeAddr2}" "${arrayKey[21]}"

    raw=$(${Chain33Cli} autonomy revokeChange -p "${proposalChangeID}")
    sign_and_send "${raw}" "${arrayKey[21]}"

    showChange_status 2
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function testProposalTerminate() {
    #test terminate
    echo -e "${GRE}=========== $FUNCNAME begin ===========${NOC}"
    update_last_header "${start_block}"
    proposalRuleTx

    update_last_header "${start_block}"
    proposalBoardTx

    update_last_header "${start_block}"
    proposalProjectTx

    update_last_header "${start_block}"
    proposalChangeTx "${changeAddr2}" "${arrayKey[21]}"

    block_wait "${Chain33Cli}" 900

    raw=$(${Chain33Cli} autonomy terminateRule -p "${proposalRuleID}")
    sign_and_send "${raw}" "${propKey}"
    showRule_status 4

    raw=$(${Chain33Cli} autonomy terminateBoard -p "${proposalBoardID}")
    sign_and_send "${raw}" "${propKey}"
    showBoard_status 4

    raw=$(${Chain33Cli} autonomy terminateProject -p "${proposalProjectID}")
    sign_and_send "${raw}" "${propKey}"
    showProject_status 5

    raw=$(${Chain33Cli} autonomy terminateChange -p "${proposalChangeID}")
    sign_and_send "${raw}" "${arrayKey[21]}"
    showChange_status 4
    echo -e "${GRE}=========== $FUNCNAME end ===========${NOC}"
}

function mainTest() {
    # shellcheck disable=SC2154
    docker_chain33_ip=$(get_docker_addr "${dockerNamePrefix}_chain33_1")
    Chain33Cli="./chain33-cli --rpc_laddr http://${docker_chain33_ip}:8801"
    InitChain33Account

    testProposalRule
    testProposalBoard
    testProposalProject
    testProposalChange
    testProposalTerminate
}
