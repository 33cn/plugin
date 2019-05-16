#!/usr/bin/env bash
# shellcheck disable=SC2128

MAIN_HTTP=""
CASE_ERR=""
evm_createContract_unsignedTx="0a0365766d129407228405608060405234801561001057600080fd5b50610264806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c8063b8e010de1461003b578063cc80f6f314610045575b600080fd5b6100436100c2565b005b61004d610109565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561008757818101518382015260200161006f565b50505050905090810190601f1680156100b45780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60408051808201909152600d8082527f5468697320697320746573742e000000000000000000000000000000000000006020909201918252610106916000916101a0565b50565b60008054604080516020601f60026000196101006001881615020190951694909404938401819004810282018101909252828152606093909290918301828280156101955780601f1061016a57610100808354040283529160200191610195565b820191906000526020600020905b81548152906001019060200180831161017857829003601f168201915b505050505090505b90565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106101e157805160ff191683800117855561020e565b8280016001018555821561020e579182015b8281111561020e5782518255916020019190600101906101f3565b5061021a92915061021e565b5090565b61019d91905b8082111561021a576000815560010161022456fea165627a7a72305820fec5dd5ca2cb47523ba08c04749bc5c14c435afee039f3047c2b7ea2faca737800293a8a025b7b22636f6e7374616e74223a66616c73652c22696e70757473223a5b5d2c226e616d65223a22736574222c226f757470757473223a5b5d2c2270617961626c65223a66616c73652c2273746174654d75746162696c697479223a226e6f6e70617961626c65222c2274797065223a2266756e6374696f6e227d2c7b22636f6e7374616e74223a747275652c22696e70757473223a5b5d2c226e616d65223a2273686f77222c226f757470757473223a5b7b226e616d65223a22222c2274797065223a22737472696e67227d5d2c2270617961626c65223a66616c73652c2273746174654d75746162696c697479223a2276696577222c2274797065223a2266756e6374696f6e227d5d20c0c7ee04309aedc4bcfba5beca5f3a223139746a5335316b6a7772436f535153313355336f7765376759424c6653666f466d"
evm_preExecContract_unsignedTx="0a4b757365722e65766d2e30786437343339376531363435316639643734316362656264393930653765633830366366356535313431376536343738306633376264333061353662383635373312073a05736574282920e0914330ddf78ef3deeee4d1153a2231466a54374243474c446f4b676571676a7a6d76527671796f537179454c57514c6f"
evm_execContract_unsignedTx="0a4b757365722e65766d2e30786437343339376531363435316639643734316362656264393930653765633830366366356535313431376536343738306633376264333061353662383635373312083a0673686f77282920e0914330efb5ebb5cda7e5a3513a2231466a54374243474c446f4b676571676a7a6d76527671796f537179454c57514c6f"
evm_creatorAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
evm_addr=""


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
        CASE_ERR="FAIL"
    fi
}

chian33_importKey() {
    echo_rst
}

evm_CreateContract() {
    signRawTx "${evm_createContract_unsignedTx}" "${evm_creatorAddr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "CreateContract signRawTx" "$rst"
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "CreateContract sendSignedTx" "$rst"
    fi

    queryTransaction ".receipt.tyName" "ExecOk"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "CreateContract queryExecRes" "$rst"
    fi
}

evm_CallContract() {
    op=$1
    if [ ${op} == "preExec" ]; then
        unsignedTx="${evm_preExecContract_unsignedTx}"
    elif [ ${op} == "Exec" ]; then
        unsignedTx="${evm_execContract_unsignedTxevm}"
    else
        rst=1
        echo_rst "CallContract invalid param" "$rst"
        return
    fi

    signRawTx "${unsignedTx}" "${evm_creatorAddr}"
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "CallContract signRawTx" "$rst"
        return
    fi

    sendSignedTx
    if [ $? -ne 0 ]; then
        rst=$?
        echo_rst "CallContract sendSignedTx" "$rst"
        return
    fi

    queryTransaction ".receipt.tyName" "ExecOk"
    rst=$?
    echo_rst "CallContract queryExecRes" "$rst"
    return
}

evm_abiGet() {
    abiInfo=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.Query","params":[{"execer":"evm","funcName":"QueryABI","payload":{"address":"'${evm_addr}'"}}]}' -H 'content-type:text/plain;' ${MAIN_HTTP})
    rst=$?
    echo_rst "CallContract queryExecRes" "$rst"
    return
}

signRawTx() {
    unsignedTx=$1
    addr=$2
    signedTx=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SignRawTx","params":[{"addr":"'${addr}'","txHex":"'${unsignedTx}'","expire":"120s"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ $signedTx == "null" ]; then
        return 1
    else
        return 0
    fi
}

sendSignedTx() {
    txHash=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"token":"","data":"'${signedTx}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".result")
    if [ $txHash == "null" ]; then
        return 1
    else
        return 0
    fi
}

# 查询交易的执行结果
# 根据传入的规则，校验查询的结果 （参数1: 校验规则 参数2: 预期匹配结果）
queryTransaction() {
    validator=$1
    expectRes=$2
    res=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'${txHash}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r "'${validator}'")
    if [ ${res} != "${expectRes}" ]; then
        return 1
    else
        evm_addr=$(curl -s --data-binary '{"jsonrpc":"2.0","id":2,"method":"Chain33.QueryTransaction","params":[{"hash":"'${txHash}'"}]}' -H 'content-type:text/plain;' ${MAIN_HTTP} | jq -r ".receipt.logs[1].log.contractName")
        return 0
    fi
}

function run_test() {
    local ip=$1
    chain33_importKey
    chain33_unlock
    evm_Transfer
    evm_CreateContract
    evm_CallContract
    evm_abiGet
}

function main() {
    local ip=$1
    MAIN_HTTP="http://$ip:8801"
    echo "=========== # evm rpc test ============="
    echo "main_ip=$MAIN_HTTP"
    run_test "$MAIN_HTTP"

    if [ -n "$CASE_ERR" ]; then
        echo -e "${RED}=============Evm Rpc Test Fail=============${NOC}"
        exit 1
    else
        echo -e "${GRE}=============Evm Rpc Test Pass==============${NOC}"
    fi
}

main "$1"