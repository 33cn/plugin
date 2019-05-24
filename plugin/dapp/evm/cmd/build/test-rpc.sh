#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""
evm_createContract_unsignedTx="0a0365766d129407228405608060405234801561001057600080fd5b50610264806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c8063b8e010de1461003b578063cc80f6f314610045575b600080fd5b6100436100c2565b005b61004d610109565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561008757818101518382015260200161006f565b50505050905090810190601f1680156100b45780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60408051808201909152600d8082527f5468697320697320746573742e000000000000000000000000000000000000006020909201918252610106916000916101a0565b50565b60008054604080516020601f60026000196101006001881615020190951694909404938401819004810282018101909252828152606093909290918301828280156101955780601f1061016a57610100808354040283529160200191610195565b820191906000526020600020905b81548152906001019060200180831161017857829003601f168201915b505050505090505b90565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106101e157805160ff191683800117855561020e565b8280016001018555821561020e579182015b8281111561020e5782518255916020019190600101906101f3565b5061021a92915061021e565b5090565b61019d91905b8082111561021a576000815560010161022456fea165627a7a72305820fec5dd5ca2cb47523ba08c04749bc5c14c435afee039f3047c2b7ea2faca737800293a8a025b7b22636f6e7374616e74223a66616c73652c22696e70757473223a5b5d2c226e616d65223a22736574222c226f757470757473223a5b5d2c2270617961626c65223a66616c73652c2273746174654d75746162696c697479223a226e6f6e70617961626c65222c2274797065223a2266756e6374696f6e227d2c7b22636f6e7374616e74223a747275652c22696e70757473223a5b5d2c226e616d65223a2273686f77222c226f757470757473223a5b7b226e616d65223a22222c2274797065223a22737472696e67227d5d2c2270617961626c65223a66616c73652c2273746174654d75746162696c697479223a2276696577222c2274797065223a2266756e6374696f6e227d5d20c0c7ee04309aedc4bcfba5beca5f3a223139746a5335316b6a7772436f535153313355336f7765376759424c6653666f466d"
evm_createContract_para_unsignedTx="0a0f757365722e702e706172612e65766d129407228405608060405234801561001057600080fd5b50610264806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c8063b8e010de1461003b578063cc80f6f314610045575b600080fd5b6100436100c2565b005b61004d610109565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561008757818101518382015260200161006f565b50505050905090810190601f1680156100b45780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60408051808201909152600d8082527f5468697320697320746573742e000000000000000000000000000000000000006020909201918252610106916000916101a0565b50565b60008054604080516020601f60026000196101006001881615020190951694909404938401819004810282018101909252828152606093909290918301828280156101955780601f1061016a57610100808354040283529160200191610195565b820191906000526020600020905b81548152906001019060200180831161017857829003601f168201915b505050505090505b90565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106101e157805160ff191683800117855561020e565b8280016001018555821561020e579182015b8281111561020e5782518255916020019190600101906101f3565b5061021a92915061021e565b5090565b61019d91905b8082111561021a576000815560010161022456fea165627a7a7230582080ff1004de2195e6c08d0d0a65484b3d393c99c280e305cb383dbc89343cdd6a00293a8a025b7b22636f6e7374616e74223a66616c73652c22696e70757473223a5b5d2c226e616d65223a22736574222c226f757470757473223a5b5d2c2270617961626c65223a66616c73652c2273746174654d75746162696c697479223a226e6f6e70617961626c65222c2274797065223a2266756e6374696f6e227d2c7b22636f6e7374616e74223a747275652c22696e70757473223a5b5d2c226e616d65223a2273686f77222c226f757470757473223a5b7b226e616d65223a22222c2274797065223a22737472696e67227d5d2c2270617961626c65223a66616c73652c2273746174654d75746162696c697479223a2276696577222c2274797065223a2266756e6374696f6e227d5d20c0c7ee0430e1c7facdc1f199956c3a2231483969326a67464a594e5167573350695468694337796b7a5663653570764b7478"
evm_creatorAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
evm_contractAddr=""
evm_addr=""

#color
RED='\033[1;31m'
GRE='\033[1;32m'
NOC='\033[0m'

# $2=0 means true, other false
function echo_rst() {
    if [ "$2" -eq 0 ]; then
        echo -e "${GRE}$1 ok${NOC}"
    else
        echo -e "${RED}$1 fail${NOC}"
        CASE_ERR="FAIL"
    fi
}

function chain33_ImportPrivkey() {
    local pri=$2
    local acc=$3
    local req='"method":"Chain33.ImportPrivkey", "params":[{"privkey":"'"$pri"'", "label":"admin"}]'
    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "$1")
    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.label=="admin") and (.result.acc.addr == "'"$acc"'")' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
}

function Chain33_SendToAddress() {
    local from="$1"
    local to="$2"
    local amount=$3
    local req='"method":"Chain33.SendToAddress", "params":[{"from":"'"$from"'","to":"'"$to"'", "amount":'"$amount"', "note":"test\n"}]'
    #    echo "#request: $req"
    resp=$(curl -ksd "{$req}" "${MAIN_HTTP}")
    #    echo "#response: $resp"
    ok=$(jq '(.error|not) and (.result.hash|length==66)' <<<"$resp")
    [ "$ok" == true ]
    echo_rst "$FUNCNAME" "$?"
    hash=$(jq '(.result.hash)' <<<"$resp")
    echo "hash=$hash"
    #    query_tx "$hash"
}

function chain33_unlock() {
    ok=$(curl -k -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.UnLock","params":[{"passwd":"1314fuzamei","timeout":0}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result.isOK")
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

function evm_createContract() {
    validator=$1
    expectRes=$2
    if [ "$ispara" == "true" ]; then
        paraName="user.p.para."
        signRawTx "${evm_createContract_para_unsignedTx}" "${evm_creatorAddr}"
    else
        signRawTx "${evm_createContract_unsignedTx}" "${evm_creatorAddr}"
    fi
    echo_rst "CreateContract signRawTx" "$?"

    sendSignedTx
    echo_rst "CreateContract sendSignedTx" "$?"

    block_wait 2

    queryTransaction "${validator}" "${expectRes}"
    echo_rst "CreateContract queryExecRes" "$?"
}

function evm_addressCheck() {
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"evm","funcName":"CheckAddrExists","payload":{"addr":"'${evm_contractAddr}'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    bContract=$(echo "${res}" | jq -r ".result.contract")
    contractAddr=$(echo "${res}" | jq -r ".result.contractAddr")
    if [ "${bContract}" == "true" ] && [ "${contractAddr}" == "${evm_contractAddr}" ]; then
        echo_rst "evm address check" 0
    else
        echo_rst "evm address check" 1
    fi

    return
}
function evm_callContract() {
    op=$1
    validator=$2
    expectRes=$3
    if [ "${op}" == "preExec" ]; then
        unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"evm.EvmCallTx","params":[{"fee":1,"caller":"'${evm_creatorAddr}'", "expire":"120s", "exec":"'${evm_addr}'", "abi": "set()"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    elif [ "${op}" == "Exec" ]; then
        unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"evm.EvmCallTx","params":[{"fee":1,"caller":"'${evm_creatorAddr}'", "expire":"120s", "exec":"'${evm_addr}'", "abi": "show()"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    else
        rst=1
        echo_rst "CallContract invalid param" "$rst"
        return
    fi

    signRawTx "${unsignedTx}" "${evm_creatorAddr}"
    rst=$?
    echo_rst "CallContract signRawTx" "$rst"
    if [ ${rst} == 1 ]; then
        return
    fi

    sendSignedTx
    echo_rst "CallContract sendSignedTx" "$?"

    block_wait 2

    queryTransaction "${validator}" "${expectRes}"
    echo_rst "CallContract queryExecRes" "$?"
}

function evm_abiGet() {
    abiInfo=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"evm","funcName":"QueryABI","payload":{"address":"'${evm_contractAddr}'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    res=$(echo "${abiInfo}" | jq -r ".result" | jq -r 'has("abi")')
    if [ "${res}" == "true" ]; then
        echo_rst "CallContract queryExecRes" 0
    else
        echo_rst "CallContract queryExecRes" 1
    fi
    return
}

function evm_transfer() {
    validator=$1
    expectRes=$2
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"evm.EvmTransferTx","params":[{"amount":1,"caller":"'${evm_creatorAddr}'","expire":"", "exec":"'${evm_addr}'", "paraName": "'${paraName}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        rst=1
        echo_rst "evm transfer create tx" "$rst"
        return
    fi

    signRawTx "${unsignedTx}" "${evm_creatorAddr}"
    echo_rst "evm transfer signRawTx" "$?"

    sendSignedTx
    echo_rst "evm transfer sendSignedTx" "$?"

    block_wait 2

    queryTransaction "${validator}" "${expectRes}"
    echo_rst "evm transfer queryExecRes" "$?"
}

function evm_getBalance() {
    expectBalance=$1
    echo "This is evm get balance test."
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.GetBalance","params":[{"addresses":["'${evm_creatorAddr}'"],"execer":"'${evm_addr}'", "paraName": "'${paraName}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    balance=$(echo "${res}" | jq -r ".result[0].balance")
    addr=$(echo "${res}" | jq -r ".result[0].addr")

    if [ "${balance}" == "${expectBalance}" ] && [ "${addr}" == "${evm_creatorAddr}" ]; then
        echo_rst "evm getBalance" 0
    else
        echo_rst "evm getBalance" 1
    fi
}

function evm_withDraw() {
    validator=$1
    expectRes=$2
    unsignedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"evm.EvmWithdrawTx","params":[{"amount":1,"caller":"'${evm_creatorAddr}'","expire":"", "exec":"'${evm_addr}'", "paraName":"'${paraName}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "${unsignedTx}" == "" ]; then
        rst=1
        echo_rst "evm withdraw create tx" "$rst"
        return
    fi

    signRawTx "${unsignedTx}" "${evm_creatorAddr}"
    echo_rst "evm withdraw signRawTx" "$?"

    sendSignedTx
    echo_rst "evm withdraw sendSignedTx" "$?"

    block_wait 2

    queryTransaction "${validator}" "${expectRes}"
    echo_rst "evm withdraw queryExecRes" "$?"
}
function signRawTx() {
    unsignedTx=$1
    addr=$2
    signedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SignRawTx","params":[{"addr":"'"${addr}"'","txHex":"'"${unsignedTx}"'","expire":"120s"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "$signedTx" == "null" ]; then
        return 1
    else
        return 0
    fi
}

function sendSignedTx() {
    txHash=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"token":"","data":"'"${signedTx}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ "$txHash" == "null" ]; then
        return 1
    else
        return 0
    fi
}

# 查询交易的执行结果
# 根据传入的规则，校验查询的结果 （参数1: 校验规则 参数2: 预期匹配结果）
function queryTransaction() {
    validators=$1
    expectRes=$2

    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${txHash}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})

    times=$(echo "${validators}" | awk -F '|' '{print NF}')
    for ((i = 1; i <= times; i++)); do
        validator=$(echo "${validators}" | awk -F '|' '{print $'$i'}')
        res=$(echo "${res}" | ${validator})
    done

    if [ "${res}" != "${expectRes}" ]; then
        return 1
    else
        res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'"${txHash}"'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
        if [ "${evm_addr}" == "" ]; then
            if [ "$ispara" == "true" ]; then
                evm_addr=$(echo "${res}" | jq -r ".result.receipt.logs[0].log.contractName")
            else
                evm_addr=$(echo "${res}" | jq -r ".result.receipt.logs[1].log.contractName")
            fi
        fi

        if [ "${evm_contractAddr}" == "" ]; then
            if [ "$ispara" == "true" ]; then
                evm_contractAddr=$(echo "${res}" | jq -r ".result.receipt.logs[0].log.contractAddr")
            else
                evm_contractAddr=$(echo "${res}" | jq -r ".result.receipt.logs[1].log.contractAddr")
            fi

        fi
        return 0
    fi
}

function init() {
    ispara=$(echo '"'"${MAIN_HTTP}"'"' | jq '.|contains("8901")')
    echo "ipara=$ispara"
    from="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
    local evm_addr=""
    if [ "$ispara" == "true" ]; then
        evm_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"user.p.para.evm"}]}' ${MAIN_HTTP} | jq -r ".result")
        Chain33_SendToAddress "$from" "$evm_addr" 10000000000
    else
        evm_addr=$(curl -ksd '{"method":"Chain33.ConvertExectoAddr","params":[{"execname":"evm"}]}' ${MAIN_HTTP} | jq -r ".result")
    fi
    echo "evm=$evm_addr"

    Chain33_SendToAddress "$from" "$evm_addr" 10000000000
    block_wait 2
}
function run_test() {
    local ip=$1
    evm_createContract "jq -r .result.receipt.tyName" "ExecOk"
    evm_addressCheck

    if [ "$ispara" == "true" ]; then
        evm_callContract preExec "jq -r .result.receipt.logs[0].tyName" "LogEVMStateChangeItem"
        evm_callContract Exec "jq -r .result.receipt.logs[0].log.jsonRet | jq -r .[0].value" "This is test."
    else
        evm_callContract preExec "jq -r .result.receipt.logs[1].tyName" "LogEVMStateChangeItem"
        evm_callContract Exec "jq -r .result.receipt.logs[1].log.jsonRet | jq -r .[0].value" "This is test."
    fi

    evm_abiGet
    evm_transfer "jq -r .result.receipt.tyName" "ExecOk"
    evm_getBalance 100000000
    evm_withDraw "jq -r .result.receipt.tyName" "ExecOk"
    evm_getBalance 0
}

function main() {
    local ip=$1
    MAIN_HTTP=$ip
    echo "=========== # evm rpc test ============="
    echo "main_ip=$MAIN_HTTP"

    init
    run_test "$MAIN_HTTP"

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Evm Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Evm Rpc Test Pass==============${NOC}"
    fi
}

main "$1"
