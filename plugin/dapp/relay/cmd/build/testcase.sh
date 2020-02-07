#!/usr/bin/env bash

BTCD="${1}_btcd_1"

RELAYD="${1}_relayd_1"

BTC_CTL="docker exec ${BTCD} btcctl"

# shellcheck disable=SC2086,2154
function relay_init() {
    # relayd
    sed -i $sedfix 's/^btcdOrWeb.*/btcdOrWeb = 0/g' relayd.toml
    sed -i $sedfix 's/^Tick33.*/Tick33 = 5/g' relayd.toml
    sed -i $sedfix 's/^TickBTC.*/TickBTC = 5/g' relayd.toml
    sed -i $sedfix 's/^pprof.*/pprof = false/g' relayd.toml
    sed -i $sedfix 's/^watch.*/watch = false/g' relayd.toml

}

function run_relayd_with_btcd() {
    echo "============== run_relayd_with_btcd ==============================="
    docker cp "${BTCD}:/root/rpc.cert" ./rpc.cert
    docker cp ./rpc.cert "${RELAYD}:/root/rpc.cert"
    docker restart "${RELAYD}"
}

function ping_btcd() {
    echo "============== ping_btcd ==============================="
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet listaccounts
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet generate 100
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet generate 1
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet getaddressesbyaccount default
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet listaccounts
}

function relay_config() {
    wait_btcd_up
    run_relayd_with_btcd
    ping_btcd

}

#some times btcwallet bin 18554 server port fail in btcd docker, restart btcd will be ok
# [WRN] BTCW: Can't listen on [::1]:18554: listen tcp6 [::1]:18554: bind: cannot assign requested address
# shellcheck disable=SC2068
function wait_btcd_up() {
    local count=20
    while [ $count -gt 0 ]; do
        status=$(docker-compose ps | grep btcd | awk '{print $5}')
        if [ "${status}" == "Up" ]; then
            break
        fi
        docker-compose logs btcd
        docker-compose restart btcd
        docker-compose ps
        echo "==============btcd fail $count  ================="
        ((count--))
        if [ $count == 0 ]; then
            echo "wait btcd up 20 times"
            exit 1
        fi
        mod=$((count % 4))
        if [ $mod == 0 ]; then
            docker-compose down
            sleep 5
            docker-compose up --build -d
            sleep 60
            continue
        fi
        #btcd restart need wait 30s
        sleep 30
    done
}

function wait_btc_height() {
    if [ "$#" -lt 2 ]; then
        echo "wrong wait_btc_height params"
        exit 1
    fi
    local count=100
    wait_sec=0
    while [ $count -gt 0 ]; do
        cur=$(${1} relay btc_cur_height | jq ".curHeight")
        if [ "${cur}" -ge "${2}" ]; then
            break
        fi
        ((count--))
        wait_sec=$((wait_sec + 1))
        sleep 1
    done
    echo "wait btc blocks ${wait_sec} s"

}

function relay_test() {
    local sell_addr="1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3"
    local sell_priv="22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962"
    local acct_addr="1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum"
    local acct_priv="ec9162ea5fc2f473ab8240619a0a0f495ba9e9e5d4d9c434b8794a68280236c4"
    local buy_addr="1BafoGyuC3X6Sx5EhcVuHHfDgMNyjQGc5x"
    local buy_priv="0xd04015639faa5bf740db756d8934003c4134865320e7eae65775be6cf30ff56f"
    echo "================relayd test========================"
    echo "=========== # transfer to acct ============="
    hash=$(${1} send coins transfer -a 1200 -t "$sell_addr" -n "transfer to sell" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    hash=$(${1} send coins transfer -a 600 -t "$acct_addr" -n "transfer to sell" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    hash=$(${1} send coins transfer -a 200 -t "$buy_addr" -n "transfer to buy" -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"

    block_wait "${1}" 2
    acct1=$(${1} account balance -a "$sell_addr" | jq -r ".balance")
    acct2=$(${1} account balance -a "$acct_addr" | jq -r ".balance")
    if [ "${acct1}" == "0.0000" ] || [ "${acct2}" == "0.0000" ]; then
        echo "wrong relay addr balance, should not be zero"
        exit 1
    fi

    times=100
    while true; do
        ${1} relay btc_cur_height
        base_height=$(${1} relay btc_cur_height | jq ".baseHeight")
        btc_cur_height=$(${1} relay btc_cur_height | jq ".curHeight")
        if [ "${btc_cur_height}" == "${base_height}" ]; then
            echo "height not correct, wait 2 block.."
            block_wait "${1}" 2
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "height not correct failed"
                exit 1
            fi
        else
            echo "btc height correct, pass"
            break
        fi
    done

    echo "=========== # get real BTC account ============="
    newacct="relay"
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet walletpassphrase password 100000000
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet createnewaccount "${newacct}"
    btcrcv_addr=$(${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet getaccountaddress "${newacct}")
    echo "btcrcvaddr=${btcrcv_addr}"

    echo "=========== # transfer to relay ============="
    hash=$(${1} send coins transfer -a 1000 -t 1rhRgzbz264eyJu7Ac63wepsm9TsEpwXM -n "transfer to relay" -k "$sell_priv")
    echo "${hash}"
    hash=$(${1} send coins transfer -a 500 -t 1rhRgzbz264eyJu7Ac63wepsm9TsEpwXM -n "send to relay" -k "$acct_priv")
    echo "${hash}"
    hash=$(${1} send coins transfer -a 100 -t 1rhRgzbz264eyJu7Ac63wepsm9TsEpwXM -n "send to relay" -k "${buy_priv}")
    echo "${hash}"

    block_wait "${1}" 1
    before=$(${1} account balance -a "$sell_addr" -e relay | jq -r ".balance")
    if [ "${before}" == "0.0000" ]; then
        echo "wrong relay addr balance, should not be zero"
        exit 1
    fi

    echo "=========== # create buy order ============="
    buy_hash=$(${1} send relay create -m 2.99 -o 0 -c BTC -a 1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT -b 200 -k "$sell_priv")
    echo "${buy_hash}"
    echo "=========== # create sell order ============="
    sell_hash=$(${1} send relay create -m 2.99 -o 1 -c BTC -a 2Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT -b 200 -k "$sell_priv")
    echo "${sell_hash}"
    echo "=========== # create real buy order ============="
    realbuy_hash=$(${1} send relay create -m 10 -o 0 -c BTC -a "${btcrcv_addr}" -b 200 -k "$sell_priv")
    echo "${realbuy_hash}"

    block_wait "${1}" 1

    #    coinaddr=$(${1} tx query -s "${buy_hash}" | jq -r ".receipt.logs[2].log.xAddr")
    #    if [ "${coinaddr}" != "1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT" ]; then
    #        ${1} tx query -s "${buy_hash}"
    #        echo "wrong create order to coinaddr"
    #        exit 1
    #    fi
    #    buy_id=$(${1} tx query -s "${buy_hash}" | jq -r ".receipt.logs[2].log.orderId")
    #    if [ -z "${buy_id}" ]; then
    #        echo "wrong buy id"
    #        exit 1
    #    fi

    status=$(${1} tx query -s "${sell_hash}" | jq -r ".receipt.logs[2].log.curStatus")
    if [ "${status}" != "pending" ]; then
        echo "wrong create sell order status"
        exit 1
    fi

    ${1} relay status -s 1
    num=$(${1} relay status -s 1 | jq -sr '.|length')
    if [ "${num}" != 3 ]; then
        echo "wrong create orders num"
        exit 1
    fi

    id=$(${1} relay status -s 1 | jq -sr '.[] | select(.coinoperation==0)| select(.coinaddr=="1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT") |.orderid')
    if [ "${id}" != "${buy_hash}" ]; then
        echo "wrong relay status buy order id"
        exit 1
    fi
    id=$(${1} relay status -s 1 | jq -sr '.[] | select(.coinoperation==0)| select(.coinamount=="10.0000") |.orderid')
    if [ "${id}" != "${realbuy_hash}" ]; then
        echo "wrong relay status real buy order id"
        exit 1
    fi

    id=$(${1} relay status -s 1 | jq -sr '.[] | select(.coinoperation==1)|.orderid')
    if [ "${id}" != "${sell_hash}" ]; then
        echo "wrong relay status sell order id"
        exit 1
    fi

    echo "=========== # accept buy order ============="
    acct_buy_hash=$(${1} send relay accept -o "${buy_hash}" -a 1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT -k "$acct_priv")
    echo "${acct_buy_hash}"
    echo "=========== # accept real buy order ============="
    acct_realbuy_hash=$(${1} send relay accept -o "${realbuy_hash}" -a 1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT -k "${buy_priv}")
    echo "${acct_realbuy_hash}"
    echo "=========== # accept sell order ============="
    acct_sell_hash=$(${1} send relay accept -o "${sell_hash}" -a 1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT -k "$acct_priv")
    echo "${acct_sell_hash}"
    block_wait "${1}" 1

    ${1} relay status -s 2
    num=$(${1} relay status -s 2 | jq -sr '.|length')
    if [ "${num}" != 3 ]; then
        echo "wrong accept orders num"
        exit 1
    fi

    echo "=========== # btc generate 80 blocks ============="
    ## for unlock order's 36 blocks waiting
    current=$(${1} relay btc_cur_height | jq ".curHeight")
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet generate 80
    wait_btc_height "${1}" $((current + 80))

    echo "=========== # btc tx to real order ============="
    btc_tx_hash=$(${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet sendfrom default "${btcrcv_addr}" 10)
    echo "${btc_tx_hash}"
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet generate 4
    blockhash=$(${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet gettransaction "${btc_tx_hash}" | jq -r ".blockhash")
    blockheight=$(${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet getblockheader "${blockhash}" | jq -r ".height")
    echo "blcockheight=${blockheight}"
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet --wallet getreceivedbyaddress "${btcrcv_addr}"

    wait_btc_height "${1}" $((current + 80 + 4))

    echo "=========== # unlock buy order ==========="
    acceptHeight=$(${1} tx query -s "${acct_buy_hash}" | jq -r ".receipt.logs[2].log.xHeight")
    if [ "${acceptHeight}" -lt "${btc_cur_height}" ]; then
        echo "accept height less previous height"
        exit 1
    fi

    wait_btc_height "${1}" $((acceptHeight + 72))

    revoke_hash=$(${1} send relay revoke -a 0 -t 1 -i "${buy_hash}" -k "$acct_priv")
    echo "${revoke_hash}"
    echo "=========== # confirm real buy order ============="
    confirm_hash=$(${1} send relay confirm -t "${btc_tx_hash}" -o "${realbuy_hash}" -k "${buy_priv}")
    echo "${confirm_hash}"
    echo "=========== # confirm sell order ============="
    confirm_hash=$(${1} send relay confirm -t 6359f0868171b1d194cbee1af2f16ea598ae8fad666d9b012c8ed2b79a236ec4 -o "${sell_hash}" -k "$sell_priv")
    echo "${confirm_hash}"

    block_wait "${1}" 1
    echo "${revoke_hash}"
    ${1} tx query -s "${revoke_hash}"

    id=$(${1} relay status -s 1 | jq -sr '.[] | select(.coinoperation==0)|.orderid')
    if [ "${id}" != "${buy_hash}" ]; then
        echo "wrong relay pending status unlock buy order id"
        exit 1
    fi

    id=$(${1} relay status -s 3 | jq -sr '.[] | select(.coinoperation==0)|.orderid')
    if [ "${id}" != "${realbuy_hash}" ]; then
        echo "wrong relay status confirming real buy order id"
        exit 1
    fi
    id=$(${1} relay status -s 3 | jq -sr '.[] | select(.coinoperation==1)|.orderid')
    if [ "${id}" != "${sell_hash}" ]; then
        echo "wrong relay status confirming sell order id"
        exit 1
    fi

    echo "=========== # btc generate 300 blocks  ==="
    current=$(${1} relay btc_cur_height | jq ".curHeight")
    ${BTC_CTL} --rpcuser=root --rpcpass=1314 --simnet generate 300
    wait_btc_height "${1}" $((current + 300))

    echo "=========== # unlock sell order ==="
    confirmHeight=$(${1} tx query -s "${confirm_hash}" | jq -r ".receipt.logs[1].log.xHeight")
    if [ "${confirmHeight}" -lt "${btc_cur_height}" ]; then
        echo "wrong confirm height"
        exit 1
    fi

    wait_btc_height "${1}" $((confirmHeight + 288))

    revoke_hash=$(${1} send relay revoke -a 0 -t 0 -i "${sell_hash}" -k "$sell_priv")
    echo "${revoke_hash}"
    echo "=========== # test cancel create order ==="
    ${1} account balance -a "$acct_addr"
    cancel_hash=$(${1} send relay create -m 2.99 -o 0 -c BTC -a 1Am9UTGfdnxabvcywYG2hvzr6qK8T3oUZT -b 20 -k "$acct_priv")
    echo "${cancel_hash}"

    block_wait "${1}" 1
    echo "${revoke_hash}"
    ${1} tx query -s "${revoke_hash}"
    echo "${cancel_hash}"
    ${1} tx query -s "${cancel_hash}"

    id=$(${1} relay status -s 1 | jq -sr '.[] | select(.coinoperation==1)| select(.address=="'"$sell_addr"'") | .orderid')
    if [ "${id}" != "${sell_hash}" ]; then
        echo "wrong relay revoke order id "
        exit 1
    fi

    echo "=========== # wait relayd verify order ======="
    ## for relayd verify tick 5s
    block_wait "${1}" 3
    #    sleep 10

    echo "=========== # check finish order ============="
    local count=30
    while true; do
        id=$(${1} relay status -s 4 | jq -sr '.[] | select(.coinoperation==0)|.orderid')
        if [ "${id}" == "${realbuy_hash}" ]; then
            break
        fi
        block_wait "${1}" 1
        count=$((count - 1))
        if [ $count -le 0 ]; then
            echo "wrong relay status finish real buy order id"
            exit 1
        fi
    done

    before=$(${1} account balance -a "${buy_addr}" -e relay | jq -r ".balance")
    if [ "${before}" != "300.0000" ]; then
        echo "wrong relay real buy addr balance, should be 300"
        exit 1
    fi

    echo "=========== # cancel order ============="
    hash=$(${1} send relay revoke -a 1 -t 0 -i "${cancel_hash}" -k "$acct_priv")
    echo "${hash}"
    block_wait "${1}" 1
    ${1} tx query -s "${hash}"

    status=$(${1} relay status -s 5 | jq -r ".status")
    if [ "${status}" != "canceled" ]; then
        echo "wrong relay order pending status"
        exit 1
    fi
    id=$(${1} relay status -s 5 | jq -sr '.[] | select(.coinoperation==0)|.orderid')
    if [ "${id}" != "${cancel_hash}" ]; then
        echo "wrong relay status cancel order id"
        exit 1
    fi

}

function relay() {
    if [ "${2}" == "init" ]; then
        relay_init
    elif [ "${2}" == "config" ]; then
        relay_config
    elif [ "${2}" == "test" ]; then
        relay_test "${1}"
    fi

}
