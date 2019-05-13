#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
PARA_HTTP=""
CASE_ERR=""
UNIT_HTTP=""

# $2=0 means true, other false
echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo "$1 ok"
    else
        echo "$1 err"
        CASE_ERR="err"
    fi

}

function block_wait() {
    req='"method":"Chain33.GetLastHeader","params":[]'
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

function query_tx() {
    block_wait 1
    txhash="$1"
    local req='"method":"Chain33.QueryTransaction","params":[{"hash":"'"$txhash"'"}]'
    # echo "req=$req"
    local times=100
    while true; do
        ret=$(curl -ksd "{$req}" ${MAIN_HTTP} | jq -r ".result.tx.hash")
        echo "query hash is ${1}, return ${ret} "
        if [ "${ret}" != "${1}" ]; then
            block_wait  1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "query tx=$1 failed"
                exit 1
            fi
        else
            echo "query tx=$1  success"
            break
        fi
    done
}

Chain33_SendToAddress() {
    from="$1"
    to="$2"
    amount=$3
    req='"method":"Chain33.SendToAddress", "params":[{"from":"'"$from"'","to":"'"$to"'", "amount":'"$amount"', "note":"test\n"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
       echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.hash|length==66)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

signrawtx(){
    txHex="$1"
    req='"method":"Chain33.SignRawTx","params":[{"privkey":"CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944","txHex":"'"$txHex"'","expire":"120s"}]'
    echo "#request SignRawTx: $req"
    curl -ksd "{$req}" ${MAIN_HTTP}
    signedTx=$(curl -ksd "{$req}" ${MAIN_HTTP} |jq -r ".result")

    req='"method":"Chain33.SendTransaction","params":[{"token":"BTY","data":"'"$signedTx"'"}]'
    echo "#request sendTx: $req"
    #    curl -ksd "{$req}" ${MAIN_HTTP}
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    err=$(jq '(.error)' <<< "$resp")
    txhash=$(jq -r ".result" <<< "$resp")
    if [ "$err" == null ];then
        echo "tx hash: $txhash"
        query_tx "$txhash"
    else
        echo "send tx error:$err"
    fi


}
relay_CreateRawRelayOrderTx() {
    req='"method":"relay.CreateRawRelayOrderTx","params":[{"coin":"BTC","amount":299000000,"addr":"1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT","btyAmount":20000000000,"coinWaits":6}]'
    echo "#request: $req"
#    curl -ksd "{$req}" "${MAIN_HTTP}"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    echo "#resp: $resp"
    ok=$(jq '(.error|not) and (.result != "")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    rawtx=$(jq -r ".result" <<<"$resp")
    echo "raw=$rawtx"
    signrawtx "$rawtx"
}

relay_CreateRawRelayAcceptTx() {
    local result="0a0572656c6179122d500112290a01311222384535736161585662396d573877635755555a6a73484a505a7331476d647a755359180220a08d0630da83b6d0e4aaffdf7b3a2131726852677a627a32363465794a753741633633776570736d395473457077584d"
    r1=$(curl -ksd '{"jsonrpc":"2.0","id":0,"method":"relay.CreateRawRelayAcceptTx","params":[{"orderId":"1","coinAddr":"8E5saaXVb9mW8wcWUUZjsHJPZs1GmdzuSY","coinWaits":2}]}'  ${MAIN_HTTP} | jq -r ".result")
    [ "$r1" == "$result" ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

relay_CreateRawRelayRevokeTx() {
    local result="0a0572656c617912d505500532d0050a0131128204303130303030303030316333336562666632613730396631336439663961373536396162313661333237383661663764376532646530393236356534316336316430373832393465636630313030303030303861343733303434303232303033326433306466356565366635376661343663646462356562386430643966653864653662333432643237393432616539306133323331653062613333336530323230336465656538303630666463373032333061376635623461643764376263336536323863626532313961383836623834323639656165623831653236623466653031343130346165333163333162663931323738643939623833373761333562626365356232376439666666313534353638333965393139343533666337623366373231663062613430336666393663396465656236383065356664333431633066633361376239306461343633316565333935363036333964623436326539636238353066666666666666666630323430343230663030303030303030303031393736613931346230646362663937656162663434303465333164393532343737636538323264616462653765313038386163633036306432313130303030303030303139373661393134366231323831656563323561623465316530373933666634653038616231616262333430396364393838616330303030303030301801228101653961363638343565303564356162633061643034656338306637373461376535383563366538646239373539363264303639613532323133376238306331642d636364616662373364386463643031373364356435633363396130373730643062333935336462383839646162393965663035623139303735313863623831352a403030303030303030303030336261323761613230306231636563616164343738643262303034333233343663336631663339383664613161666433336535303620c09a0c30a78dd7a0d9ab9390023a2131726852677a627a32363465794a753741633633776570736d395473457077584d"
    r1=$(curl -ksd '{"jsonrpc":"2.0","id":0,"method":"relay.CreateRawRelayRevokeTx","params":[{"orderId":"1","rawTx":"0100000001c33ebff2a709f13d9f9a7569ab16a32786af7d7e2de09265e41c61d078294ecf010000008a4730440220032d30df5ee6f57fa46cddb5eb8d0d9fe8de6b342d27942ae90a3231e0ba333e02203deee8060fdc70230a7f5b4ad7d7bc3e628cbe219a886b84269eaeb81e26b4fe014104ae31c31bf91278d99b8377a35bbce5b27d9fff15456839e919453fc7b3f721f0ba403ff96c9deeb680e5fd341c0fc3a7b90da4631ee39560639db462e9cb850fffffffff0240420f00000000001976a914b0dcbf97eabf4404e31d952477ce822dadbe7e1088acc060d211000000001976a9146b1281eec25ab4e1e0793ff4e08ab1abb3409cd988ac00000000","txIndex":1,"merkBranch":"e9a66845e05d5abc0ad04ec80f774a7e585c6e8db975962d069a522137b80c1d-ccdafb73d8dcd0173d5d5c3c9a0770d0b3953db889dab99ef05b1907518cb815","blockHash":"000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"}]}'  ${MAIN_HTTP} | jq -r ".result")
    [ "$r1" == "$result" ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

relay_CreateRawRelayConfirmTx() {
    local result="0a0572656c617912d505500532d0050a0131128204303130303030303030316333336562666632613730396631336439663961373536396162313661333237383661663764376532646530393236356534316336316430373832393465636630313030303030303861343733303434303232303033326433306466356565366635376661343663646462356562386430643966653864653662333432643237393432616539306133323331653062613333336530323230336465656538303630666463373032333061376635623461643764376263336536323863626532313961383836623834323639656165623831653236623466653031343130346165333163333162663931323738643939623833373761333562626365356232376439666666313534353638333965393139343533666337623366373231663062613430336666393663396465656236383065356664333431633066633361376239306461343633316565333935363036333964623436326539636238353066666666666666666630323430343230663030303030303030303031393736613931346230646362663937656162663434303465333164393532343737636538323264616462653765313038386163633036306432313130303030303030303139373661393134366231323831656563323561623465316530373933666634653038616231616262333430396364393838616330303030303030301801228101653961363638343565303564356162633061643034656338306637373461376535383563366538646239373539363264303639613532323133376238306331642d636364616662373364386463643031373364356435633363396130373730643062333935336462383839646162393965663035623139303735313863623831352a403030303030303030303030336261323761613230306231636563616164343738643262303034333233343663336631663339383664613161666433336535303620c09a0c30a78dd7a0d9ab9390023a2131726852677a627a32363465794a753741633633776570736d395473457077584d"
    r1=$(curl -ksd '{"jsonrpc":"2.0","id":0,"method":"relay.CreateRawRelayConfirmTx","params":[{"orderId":"1","rawTx":"0100000001c33ebff2a709f13d9f9a7569ab16a32786af7d7e2de09265e41c61d078294ecf010000008a4730440220032d30df5ee6f57fa46cddb5eb8d0d9fe8de6b342d27942ae90a3231e0ba333e02203deee8060fdc70230a7f5b4ad7d7bc3e628cbe219a886b84269eaeb81e26b4fe014104ae31c31bf91278d99b8377a35bbce5b27d9fff15456839e919453fc7b3f721f0ba403ff96c9deeb680e5fd341c0fc3a7b90da4631ee39560639db462e9cb850fffffffff0240420f00000000001976a914b0dcbf97eabf4404e31d952477ce822dadbe7e1088acc060d211000000001976a9146b1281eec25ab4e1e0793ff4e08ab1abb3409cd988ac00000000","txIndex":1,"merkBranch":"e9a66845e05d5abc0ad04ec80f774a7e585c6e8db975962d069a522137b80c1d-ccdafb73d8dcd0173d5d5c3c9a0770d0b3953db889dab99ef05b1907518cb815","blockHash":"000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"}]}'  ${MAIN_HTTP} | jq -r ".result")
    [ "$r1" == "$result" ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

relay_CreateRawRelayVerifyBTCTx() {
    local result="0a0572656c617912d505500532d0050a0131128204303130303030303030316333336562666632613730396631336439663961373536396162313661333237383661663764376532646530393236356534316336316430373832393465636630313030303030303861343733303434303232303033326433306466356565366635376661343663646462356562386430643966653864653662333432643237393432616539306133323331653062613333336530323230336465656538303630666463373032333061376635623461643764376263336536323863626532313961383836623834323639656165623831653236623466653031343130346165333163333162663931323738643939623833373761333562626365356232376439666666313534353638333965393139343533666337623366373231663062613430336666393663396465656236383065356664333431633066633361376239306461343633316565333935363036333964623436326539636238353066666666666666666630323430343230663030303030303030303031393736613931346230646362663937656162663434303465333164393532343737636538323264616462653765313038386163633036306432313130303030303030303139373661393134366231323831656563323561623465316530373933666634653038616231616262333430396364393838616330303030303030301801228101653961363638343565303564356162633061643034656338306637373461376535383563366538646239373539363264303639613532323133376238306331642d636364616662373364386463643031373364356435633363396130373730643062333935336462383839646162393965663035623139303735313863623831352a403030303030303030303030336261323761613230306231636563616164343738643262303034333233343663336631663339383664613161666433336535303620c09a0c30a78dd7a0d9ab9390023a2131726852677a627a32363465794a753741633633776570736d395473457077584d"
    r1=$(curl -ksd '{"jsonrpc":"2.0","id":0,"method":"relay.CreateRawRelayVerifyBTCTx","params":[{"orderId":"1","rawTx":"0100000001c33ebff2a709f13d9f9a7569ab16a32786af7d7e2de09265e41c61d078294ecf010000008a4730440220032d30df5ee6f57fa46cddb5eb8d0d9fe8de6b342d27942ae90a3231e0ba333e02203deee8060fdc70230a7f5b4ad7d7bc3e628cbe219a886b84269eaeb81e26b4fe014104ae31c31bf91278d99b8377a35bbce5b27d9fff15456839e919453fc7b3f721f0ba403ff96c9deeb680e5fd341c0fc3a7b90da4631ee39560639db462e9cb850fffffffff0240420f00000000001976a914b0dcbf97eabf4404e31d952477ce822dadbe7e1088acc060d211000000001976a9146b1281eec25ab4e1e0793ff4e08ab1abb3409cd988ac00000000","txIndex":1,"merkBranch":"e9a66845e05d5abc0ad04ec80f774a7e585c6e8db975962d069a522137b80c1d-ccdafb73d8dcd0173d5d5c3c9a0770d0b3953db889dab99ef05b1907518cb815","blockHash":"000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506"}]}'  ${MAIN_HTTP} | jq -r ".result")
    [ "$r1" == "$result" ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

relay_CreateRawRelaySaveBTCHeadTx() {
    local result="0a0572656c617912d20150063acd010aca010a4033626330656537313263383463353839633639336230396161333638356632363663326332393561366430333139353265636338363361326131656566653435180e2a4030356139383636656461333735626339316134303461613130663366666163653331643334393437313665616436383666653732656263343637393362306261524031366164366435383861656361313262663365376664366232323633393932623434343263393639326634653133346261376366306437393137343633323861600120a08d0630d7e0d4a5ecea8abf5e3a2131726852677a627a32363465794a753741633633776570736d395473457077584d"
    r1=$(curl -ksd '{"jsonrpc":"2.0","id":0,"method":"relay.CreateRawRelaySaveBTCHeadTx","params":[{"hash":"3bc0ee712c84c589c693b09aa3685f266c2c295a6d031952ecc863a2a1eefe45","height":14,"merkleRoot":"05a9866eda375bc91a404aa10f3fface31d3494716ead686fe72ebc46793b0ba","previousHash":"16ad6d588aeca12bf3e7fd6b2263992b4442c9692f4e134ba7cf0d791746328a","isReset":true}]}'  ${MAIN_HTTP} | jq -r ".result")
    [ "$r1" == "$result" ]
    rst=$?
    echo_rst "$FUNCNAME" "$rst"

}

query_GetRelayOrderByStatus(){
    req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetRelayOrderByStatus","payload":{"addr":"","status":"pending","coins":["BTC"],"pageNumber":0,"pageSize":0}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.txDetails|length == 2)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

query_GetSellRelayOrder(){
    req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetSellRelayOrder","payload":{"addr":"1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT","status":"pending","coins":[""],"pageNumber":0,"pageSize":0}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
        echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result[0].address == "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt") and (.result[0].coinoperation == "buy") ' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

query_GetBuyRelayOrder(){
    req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBuyRelayOrder","payload":{"addr":"14KEKbYtKKQm4wMthSK9J4La4nAiidGozt","status":"locking","coins":[""],"pageNumber":0,"pageSize":0}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
        echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.txDetails|length == 2)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

query_GetBTCHeaderList(){
    req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBTCHeaderList","payload":{"reqHeight":"10","counts":10,"direction":0}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.heights|length == 2)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}

query_GetBTCHeaderCurHeight(){
    req='"method":"Chain33.Query", "params":[{"execer":"relay","funcName":"GetBTCHeaderCurHeight","payload":{"baseHeight":"0"}}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" ${MAIN_HTTP})
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.baseHeight == 10) and (.result.curHeight > 10)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"

}
function run_testcases() {
    from="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    exec_relay_addr="1rhRgzbz264eyJu7Ac63wepsm9TsEpwXM"
#    Chain33_SendToAddress "$from" "$exec_relay_addr" 100000000000

    to="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
#    Chain33_SendToAddress "$from" "$to" 50000000000
    block_wait  1

    from="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
#    Chain33_SendToAddress "$from" "$exec_relay_addr" 20000000000
    block_wait  1

#    relay_CreateRawRelaySaveBTCHeadTx
    relay_CreateRawRelayOrderTx
#    query_GetSellRelayOrder
#    relay_CreateRawRelayAcceptTx
#    query_GetBuyRelayOrder
#    query_GetRelayOrderByStatus
#    relay_CreateRawRelayConfirmTx
#    relay_CreateRawRelayVerifyBTCTx
#    relay_CreateRawRelayRevokeTx

#    query_GetBTCHeaderCurHeight
#    query_GetBTCHeaderList


}
function rpc_test() {
    MAIN_HTTP="$1"
    echo "=========== # relay rpc test ============="
    echo "main_ip=$MAIN_HTTP"

    run_testcases

    if [ -n "$CASE_ERR" ]; then
        echo "paracross there some case error"
        exit 1
    fi
}

rpc_test "$1"
