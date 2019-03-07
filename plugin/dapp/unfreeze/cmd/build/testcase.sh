#!/usr/bin/env bash

CLI="docker exec ${NODE3} /root/chain33-cli"
beneficiary=12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
beneficiary_key=0x4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01
#owner=14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
owner_key=CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944
unfreeze_exec_addr=15YsqAuXeEXVHgm6RVx4oJaAAnhtwqnu3H

function unfreeze_test() {
    echo "=========== # unfreeze  test ============="
    echo "=== 1 check exec addr"
    result=$($CLI exec addr -e unfreeze)
    if [ "${result}" != "${unfreeze_exec_addr}" ]; then
        echo "unfreeze exec addr is not right, expect ${unfreeze_exec_addr} result ${result}"
        exit 1
    fi
    block_wait "${CLI}" 2

    echo "=== 2 prepare: transfer bty to unfreeze "
    result=$($CLI send coins transfer -a 5 -n test -t ${unfreeze_exec_addr} -k ${owner_key})
    echo "${result}"
    block_wait "${CLI}" 2

    echo "=== 3 create unfreeze tx"
    tx_hash=$(${CLI} send unfreeze create fix_amount -a 0.01 -e coins -s bty -b ${beneficiary} -p 20 -t 2 -k ${owner_key})
    block_wait "${CLI}" 2
    unfreeze_id=$(${CLI} tx query -s "${tx_hash}" | jq ".receipt.logs[2].log.current.unfreezeID")
    echo "${unfreeze_id}"
    unfreeze_id2=${unfreeze_id#\"mavl-unfreeze-}
    uid=${unfreeze_id2%\"}

    echo "==== 4 check some message "
    sleep 20
    withdraw=$(${CLI} unfreeze show_withdraw --id "${uid}" | jq ".availableAmount")
    if [ "${withdraw}" = "0" ]; then
        echo "create unfreeze failed, expect withdraw shoult >0 "
        exit 1
    fi

    echo "==== 5 withdraw"
    ${CLI} send unfreeze withdraw --id "${uid}" -k "${beneficiary_key}"
    block_wait "${CLI}" 2
    remaining=$(${CLI} unfreeze show --id "${uid}" | jq ".remaining")
    if [ "${remaining}" = '"200000000"' ]; then
        echo "withdraw failed, expect remaining < 200000000, result ${remaining}"
        exit 1
    fi

    echo "==== 6 termenate"
    ${CLI} send unfreeze terminate --id "${uid}" -k "${owner_key}"
    block_wait "${CLI}" 2
    remaining=$(${CLI} unfreeze show --id "${uid}" | jq ".remaining")
    remainingNum=$(echo "$remaining" | awk '{print int($0)}')
    if [ "100000000" -lt "${remainingNum}" ]; then
        echo "terminate failed, expect remaining < 100000000, result ${remaining}"
        exit 1
    fi
    echo "==================== unfreeze test end"
}

function unfreeze() {
    if [ "${2}" == "init" ]; then
        return
    elif [ "${2}" == "config" ]; then
        return
    elif [ "${2}" == "test" ]; then
        unfreeze_test "${1}"
    fi
}
