#!/usr/bin/env bash
PARA_CLI="docker exec ${NODE3} /root/chain33-para-cli"

PARA_CLI2="docker exec ${NODE2} /root/chain33-para-cli"
PARA_CLI1="docker exec ${NODE1} /root/chain33-para-cli"
PARA_CLI4="docker exec ${NODE4} /root/chain33-para-cli"
MAIN_CLI="docker exec ${NODE3} /root/chain33-cli"

PARANAME="para"
PARA_COIN_FROZEN="5.0000"

xsedfix=""
if [ "$(uname)" == "Darwin" ]; then
    xsedfix=".bak"
fi

# shellcheck source=/dev/null
source test-rpc.sh

function para_init() {
    para_set_toml chain33.para33.toml
    para_set_toml chain33.para32.toml
    para_set_toml chain33.para31.toml
    para_set_toml chain33.para30.toml

    sed -i $xsedfix 's/^authAccount=.*/authAccount="1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"/g' chain33.para33.toml
    sed -i $xsedfix 's/^authAccount=.*/authAccount="1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"/g' chain33.para32.toml
    sed -i $xsedfix 's/^authAccount=.*/authAccount="1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"/g' chain33.para31.toml
    sed -i $xsedfix 's/^authAccount=.*/authAccount="1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"/g' chain33.para30.toml
}

function para_set_toml() {
    cp chain33.para.toml "${1}"

    sed -i $xsedfix 's/^Title.*/Title="user.p.'''$PARANAME'''."/g' "${1}"
    sed -i $xsedfix 's/^# TestNet=.*/TestNet=true/g' "${1}"
    sed -i $xsedfix 's/^startHeight=.*/startHeight=0/g' "${1}"
    sed -i $xsedfix 's/^emptyBlockInterval=.*/emptyBlockInterval=4/g' "${1}"
    sed -i $xsedfix '/^emptyBlockInterval=.*/a MainBlockHashForkHeight=1' "${1}"

    sed -i $xsedfix 's/^MainParaSelfConsensusForkHeight=.*/MainParaSelfConsensusForkHeight=50/g' "${1}"
    sed -i $xsedfix 's/^MainForkParacrossCommitTx=.*/MainForkParacrossCommitTx=1/g' "${1}"

    # rpc
    sed -i $xsedfix 's/^jrpcBindAddr=.*/jrpcBindAddr="0.0.0.0:8901"/g' "${1}"
    sed -i $xsedfix 's/^grpcBindAddr=.*/grpcBindAddr="0.0.0.0:8902"/g' "${1}"
    sed -i $xsedfix 's/^whitelist=.*/whitelist=["localhost","127.0.0.1","0.0.0.0"]/g' "${1}"
    sed -i $xsedfix 's/^ParaRemoteGrpcClient=.*/ParaRemoteGrpcClient="nginx:8803"/g' "${1}"
}

function para_set_wallet() {
    echo "=========== # para set wallet ============="
    para_import_key "${PARA_CLI}" "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" "paraAuthAccount"
    para_import_key "${PARA_CLI2}" "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" "paraAuthAccount"
    para_import_key "${PARA_CLI1}" "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" "paraAuthAccount"
    para_import_key "${PARA_CLI4}" "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" "paraAuthAccount"

    #14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    para_import_key "${PARA_CLI}" "0xCC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" "genesis"
    #12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
    para_import_key "${PARA_CLI}" "0x4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" "test"
    #1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
    para_import_key "${PARA_CLI}" "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588" "relay"
    #1PUiGcbsccfxW3zuvHXZBJfznziph5miAo
    para_import_key "${PARA_CLI}" "0x56942AD84CCF4788ED6DACBC005A1D0C4F91B63BCF0C99A02BE03C8DEAE71138" "dapptest1"
    #1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX
    para_import_key "${PARA_CLI}" "0x2116459C0EC8ED01AA0EEAE35CAC5C96F94473F7816F114873291217303F6989" "dapptest2"

    #super node behalf test
    #1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj
    para_import_key "${PARA_CLI}" "0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5" "behalfnode"
    #1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH
    para_import_key "${PARA_CLI}" "0xfdf2bbff853ecff2e7b86b2a8b45726c6538ca7d1403dc94e50131ef379bdca0" "othernode1"
    #1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB
    para_import_key "${PARA_CLI}" "0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d" "othernode2"
}

function para_import_key() {
    local lable=$3
    echo "=========== # save seed to wallet ============="
    result=$(${1} seed save -p 1314fuzamei -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib" | jq ".isok")
    if [ "${result}" = "false" ]; then
        echo "save seed to wallet error seed, result: ${result}"
        exit 1
    fi

    echo "=========== # unlock wallet ============="
    result=$(${1} wallet unlock -p 1314fuzamei -t 0 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi

    echo "=========== # import private key ============="
    echo "key: ${2}"
    result=$(${1} account import_key -k "${2}" -l "$lable" | jq ".label")
    if [ -z "${result}" ]; then
        exit 1
    fi

    echo "=========== # close auto mining ============="
    result=$(${1} wallet auto_mine -f 0 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi
    echo "=========== # wallet status ============="
    ${1} wallet status
}

function para_transfer() {
    echo "=========== # para chain transfer ============="
    main_transfer2account "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"
    main_transfer2account "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
    main_transfer2account "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
    main_transfer2account "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
    main_transfer2account "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
    # super node test
    main_transfer2account "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
    main_transfer2account "1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj"
    main_transfer2account "1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH" 10
    main_transfer2account "1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB" 10
    #relay rpc test
    para_transfer2account "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    para_transfer2account "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
    para_transfer2account "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    para_transfer2account "1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
    block_wait "${CLI}" 2

    echo "=========== # main chain send to paracross ============="
    #1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY test
    main_transfer2paracross "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588"
    #1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj
    main_transfer2paracross "0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5" 100

    #relay rpc test
    para_transfer2exec "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588" "relay"
    para_transfer2exec "0x4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" "relay"
    block_wait "${CLI}" 2

    #    para_create_manage_nodegroup

    echo "=========== # config token blacklist ============="
    #token precreate
    txhash=$(para_configkey "${PARA_CLI}" "token-blacklist" "BTY")
    echo "txhash=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

}

function main_transfer2account() {
    echo "${1}"
    local coins=200
    if [ "$#" -ge 2 ]; then
        coins="$2"
    fi
    hash1=$(${CLI} send coins transfer -a "$coins" -n test -t "${1}" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash1}"
}

function para_transfer2account() {
    echo "${1}"
    hash1=$(${PARA_CLI} send coins transfer -a 1000 -n transfer -t "${1}" -k 0xCC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    echo "${hash1}"
}

function main_transfer2paracross() {
    echo "addr=${1}"
    local coins=5
    if [ "$#" -ge 2 ]; then
        coins="$2"
    fi
    hash1=$(${CLI} send coins send_exec -a "$coins" -e paracross -k "${1}")
    echo "${hash1}"
}

function para_transfer2exec() {
    echo "exec=$2,addr=${1}"
    hash1=$(${PARA_CLI} send coins send_exec -a 500 -e "$2" -k "${1}")
    echo "${hash1}"
}

function para_create_nodegroup_test() {
    echo "=========== # para chain create node group ============="
    ##apply
    txhash=$(${PARA_CLI} send para nodegroup -o 1 -a "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4,1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR,1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k,1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs" -c 5 -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

    id=$(${PARA_CLI} tx query -s "${txhash}" | jq -r ".receipt.logs[0].log.current.id")
    if [ -z "$id" ]; then
        ${PARA_CLI} tx query -s "${txhash}"
        echo "group id not getted"
        exit 1
    fi

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "20.0000" ]; then
        echo "apply coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain quit node group ============="
    ##quit
    txhash=$(${PARA_CLI} send para nodegroup -o 3 -i "$id" -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"
    newid=$(${PARA_CLI} para nodegroup_list -s 3 | jq -r ".addrs[0].id")
    if [ -z "$newid" ]; then
        ${PARA_CLI} para nodegroup_list -s 3
        echo "quit status error "
        exit 1
    fi
    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".balance")
    if [ "$balance" != "100.0000" ]; then
        echo "quit coinfrozen error balance=$balance"
        exit 1
    fi

}

function para_create_nodegroup() {
    para_create_nodegroup_test

    echo "=========== # para chain create node group again ============="
    ##apply
    txhash=$(${PARA_CLI} send para nodegroup -o 1 -a "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY,1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4,1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR,1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k,1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs" -c 5 -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

    id=$(${PARA_CLI} tx query -s "${txhash}" | jq -r ".receipt.logs[0].log.current.id")
    if [ -z "$id" ]; then
        ${PARA_CLI} tx query -s "${txhash}"
        echo "group id not getted"
        exit 1
    fi

    echo "=========== # para chain approve node group ============="
    ##approve
    txhash=$(${PARA_CLI} send para nodegroup -o 2 -i "$id" -c 5 -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

    status=$(${PARA_CLI} para nodegroup_status -t user.p.para. | jq -r ".status")
    if [ "$status" != 2 ]; then
        echo "status not approve status=$status"
        exit 1
    fi

    ${PARA_CLI} para nodegroup_addrs -t user.p.para.

    echo "=========== # para chain quit node group fail ============="
    ##quit fail
    txhash=$(${PARA_CLI} send para nodegroup -o 3 -i "$id" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "tx=$txhash"
    query_tx "${CLI}" "${txhash}"
    status=$(${CLI} para nodegroup_status -t user.p.para. | jq -r ".status")
    if [ "$status" != 2 ]; then
        echo "status quit not approve status=$status"
        exit 1
    fi
    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "25.0000" ]; then
        echo "quit fail coinfrozen error balance=$balance"
        exit 1
    fi

}

function para_create_manage_nodegroup() {
    echo "=========== # para chain send config ============="
    para_configkey "${CLI}" "paracross-nodes-user.p.${PARANAME}." "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
    para_configkey "${CLI}" "paracross-nodes-user.p.${PARANAME}." "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
    para_configkey "${CLI}" "paracross-nodes-user.p.${PARANAME}." "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
    para_configkey "${CLI}" "paracross-nodes-user.p.${PARANAME}." "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"

    para_configkey "${PARA_CLI}" "paracross-nodes-user.p.${PARANAME}." "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
    para_configkey "${PARA_CLI}" "paracross-nodes-user.p.${PARANAME}." "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
    para_configkey "${PARA_CLI}" "paracross-nodes-user.p.${PARANAME}." "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
    para_configkey "${PARA_CLI}" "paracross-nodes-user.p.${PARANAME}." "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
    block_wait "${CLI}" 1
}

function para_configkey() {
    tx=$(${1} config config_tx -o add -c "${2}" -v "${3}")
    sign=$(${CLI} wallet sign -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc -d "${tx}")
    send=$(${CLI} wallet send -d "${sign}")
    echo "${send}"
}

function query_tx() {
    block_wait "${1}" 1

    local times=100
    while true; do
        ret=$(${1} tx query -s "${2}" | jq -r ".tx.hash")
        echo "query hash is ${2}, return ${ret} "
        if [ "${ret}" != "${2}" ]; then
            block_wait "${1}" 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "query tx=$2 failed"
                exit 1
            fi
        else
            echo "query tx=$2  success"
            break
        fi
    done
}

function token_create() {
    echo "=========== # para token test ============="
    echo "=========== # 1.token precreate ============="
    hash=$(${1} send token precreate -f 0.001 -i test -n guodunjifen -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -p 0 -s GD -t 10000 -k 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4)
    echo "${hash}"
    query_tx "${1}" "${hash}"

    ${1} token get_precreated
    owner=$(${1} token get_precreated | jq -r ".owner")
    if [ "${owner}" != "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4" ]; then
        echo "wrong pre create owner"
        exit 1
    fi
    total=$(${1} token get_precreated | jq -r ".total")
    if [ "${total}" != 10000 ]; then
        echo "wrong pre create total"
        exit 1
    fi

    echo "=========== # 2.token finish ============="
    hash=$(${1} send token finish -f 0.001 -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -s GD -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "${hash}"
    query_tx "${1}" "${hash}"

    ${1} token get_finish_created
    owner=$(${1} token get_finish_created | jq -r ".owner")
    if [ "${owner}" != "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4" ]; then
        echo "wrong finish created owner"
        exit 1
    fi
    total=$(${1} token get_finish_created | jq -r ".total")
    if [ "${total}" != 10000 ]; then
        echo "wrong finish created total"
        exit 1
    fi
    ${1} token token_balance -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -e token -s GD
    balance=$(${1} token token_balance -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -e token -s GD | jq -r '.[]|.balance')
    if [ "${balance}" != "10000.0000" ]; then
        echo "wrong para token genesis create, should be 10000.0000"
        exit 1
    fi
}

function token_transfer() {
    echo "=========== # 2.token transfer ============="
    hash=$(${1} send token transfer -a 11 -s GD -t 1GGF8toZd96wCnfJngTwXZnWCBdWHYYvjw -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${hash}"
    query_tx "${1}" "${hash}"

    ${1} token token_balance -a 1GGF8toZd96wCnfJngTwXZnWCBdWHYYvjw -e token -s GD
    balance=$(${1} token token_balance -a 1GGF8toZd96wCnfJngTwXZnWCBdWHYYvjw -e token -s GD | jq -r '.[]|.balance')
    if [ "${balance}" != "11.0000" ]; then
        echo "wrong para token transfer, should be 11.0000"
        exit 1
    fi

    echo "=========== # 3.token send exec ============="
    hash=$(${1} send token send_exec -a 11 -s GD -e paracross -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${hash}"
    query_tx "${1}" "${hash}"

    # $ ./build/chain33-cli   exec addr  -e user.p.para.paracross
    # 19WJJv96nKAU4sHFWqGmsqfjxd37jazqii
    ${1} token token_balance -a 19WJJv96nKAU4sHFWqGmsqfjxd37jazqii -e token -s GD
    balance=$(${1} token token_balance -a 19WJJv96nKAU4sHFWqGmsqfjxd37jazqii -e token -s GD | jq -r '.[]|.balance')
    if [ "${balance}" != "11.0000" ]; then
        echo "wrong para token send exec, should be 11.0000"
        exit 1
    fi

    echo "=========== # 4.token withdraw ============="
    hash=$(${1} send token withdraw -a 11 -s GD -e paracross -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${hash}"
    query_tx "${1}" "${hash}"

    ${1} token token_balance -a 19WJJv96nKAU4sHFWqGmsqfjxd37jazqii -e token -s GD
    balance=$(${1} token token_balance -a 19WJJv96nKAU4sHFWqGmsqfjxd37jazqii -e token -s GD | jq -r '.[]|.balance')
    if [ "${balance}" != "0.0000" ]; then
        echo "wrong para token withdraw, should be 0.0000"
        exit 1
    fi
}

function para_cross_transfer_withdraw() {
    echo "=========== # para cross transfer/withdraw test ============="
    paracrossAddr=1HPkPopVe3ERfvaAgedDtJQ792taZFEHCe
    ${CLI} account list
    ${CLI} send coins transfer -a 10 -n test -t $paracrossAddr -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    hash=$(${CLI} send para asset_transfer --title user.p.para. -a 1.4 -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"

    sleep 15
    ${CLI} send para asset_withdraw --title user.p.para. -a 0.7 -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

    local times=200
    while true; do
        acc=$(${CLI} account balance -e paracross -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq -r ".balance")
        echo "account balance is ${acc}, expect 9.3 "
        if [ "${acc}" != "9.3000" ]; then
            block_wait "${CLI}" 2
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                exit 1
            fi
        else
            echo "para_cross_transfer_withdraw success"
            break
        fi
    done
}

function token_create_on_mainChain() {
    echo "=========== # main chain token test ============="
    echo "=========== # 0.config token-blacklist ============="
    hash=$(${CLI} send config config_tx -c token-blacklist -o add -v BTY -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "${hash}"
    query_tx "${MAIN_CLI}" "${hash}"

    echo "=========== # send coins to token ============="
    hash=$(${CLI} send coins send_exec -a 500 -e token -n send2exec -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    query_tx "${MAIN_CLI}" "${hash}"

    echo "=========== # 1.token precreate token FZM ============="
    hash=$(${CLI} send token precreate -f 0.001 -i test -n guodunjifen -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -p 0 -s FZM -t 10000 -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "tx hash for token precreate is:" "${hash}"
    echo "MAIN_CLI is:" "${MAIN_CLI}"
    query_tx "${MAIN_CLI}" "${hash}"

    ${CLI} token get_precreated
    owner=$(${CLI} token get_precreated | jq -r ".owner")
    if [ "${owner}" != "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv" ]; then
        echo "wrong pre create owner"
        exit 1
    fi
    total=$(${CLI} token get_precreated | jq -r ".total")
    if [ "${total}" != 10000 ]; then
        echo "wrong pre create total"
        exit 1
    fi

    echo "=========== # 2.token finish ============="
    hash=$(${CLI} send token finish -f 0.001 -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -s FZM -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "${hash}"
    query_tx "${MAIN_CLI}" "${hash}"

    ${CLI} token get_finish_created
    owner=$(${CLI} token get_finish_created | jq -r ".owner")
    if [ "${owner}" != "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv" ]; then
        echo "wrong finish created owner"
        exit 1
    fi
    total=$(${CLI} token get_finish_created | jq -r ".total")
    if [ "${total}" != 10000 ]; then
        echo "wrong finish created total"
        exit 1
    fi
    ${CLI} token token_balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e token -s FZM
    balance=$(${CLI} token token_balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e token -s FZM | jq -r '.[]|.balance')
    if [ "${balance}" != "10000.0000" ]; then
        echo "wrong para token genesis create, should be 10000.0000"
        exit 1
    fi
}

function para_cross_transfer_withdraw_for_token() {
    token_create_on_mainChain

    echo "=========== # 1.transfer token:FZM to paracross ============="
    hash=$(${CLI} send token send_exec -a 333 -s FZM -e paracross -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    query_tx "${MAIN_CLI}" "${hash}"

    echo "=========== # 2.transfer asset to para chain ============="
    hash=$(${CLI} send para asset_transfer --title user.p.para. -s FZM -a 220 -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    query_tx "${MAIN_CLI}" "${hash}"

    echo "=========== # 3.asset_withdraw from parachain ============="
    ${CLI} send para asset_withdraw --title user.p.para. -a 111 -s FZM -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv

    local times=100
    while true; do
        acc=$(${CLI} asset balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv --asset_symbol FZM --asset_exec token -e paracross | jq -r ".balance")
        echo "account balance is ${acc}, expect 224 "
        if [ "${acc}" != "224.0000" ]; then
            block_wait "${CLI}" 2
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                exit 1
            fi
        else
            echo "para_cross_transfer_withdraw success"
            break
        fi
    done
}

function para_nodemanage_node_join() {
    echo "================# para node manage test ================="
    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".balance")
    if [ "$balance" != "$PARA_COIN_FROZEN" ]; then
        echo "balance coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain new node join ============="
    hash=$(${PARA_CLI} send para node -o 1 -c 5 -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".frozen")
    if [ "$balance" != "$PARA_COIN_FROZEN" ]; then
        echo "frozen coinfrozen error balance=$balance"
        exit 1
    fi

}

function para_nodemanage_node_behalf_join() {
    echo "=========== # para chain new node join 1 ============="
    hash=$(${PARA_CLI} send para node -o 1 -c 8 -a 1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "28.0000" ]; then
        echo "1Ka frozen coinfrozen error balance=$balance"
        exit 1
    fi

    balance=$(${CLI} account balance -a 1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH -e paracross | jq -r ".frozen")
    if [ "$balance" == "$PARA_COIN_FROZEN" ]; then
        echo "1LU frozen coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain new node join 2============="
    hash=$(${PARA_CLI} send para node -o 1 -c 9 -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "37.0000" ]; then
        echo "frozen coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain node 1 quit ============="
    id=$(${PARA_CLI} para node_status -a 1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH -t user.p.para. | jq -r ".id")
    if [ -z "${id}" ]; then
        echo "wrong id "
        ${PARA_CLI} para node_status -t user.p.para. -a 1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH
        exit 1
    fi
    hash=$(${PARA_CLI} send para node -o 3 -i "$id" -k 0xfdf2bbff853ecff2e7b86b2a8b45726c6538ca7d1403dc94e50131ef379bdca0)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "29.0000" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi

}

function para_nodemanage_quit_test() {
    para_nodemanage_node_join

    echo "=========== # para chain node quit ============="
    id=$(${PARA_CLI} para node_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -t user.p.para. | jq -r ".id")
    if [ -z "${id}" ]; then
        echo "wrong id "
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi
    hash=$(${PARA_CLI} send para node -o 3 -i "$id" -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".frozen")
    if [ "$balance" == "$PARA_COIN_FROZEN" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY | jq -r ".status")
    if [ "${status}" != "4" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

}

function para_nodegroup_behalf_quit_test() {
    echo "=========== # para chain behalf node quit ============="
    id=$(${PARA_CLI} para node_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -t user.p.para. | jq -r ".id")
    if [ -z "${id}" ]; then
        echo "wrong id "
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    hash=$(${PARA_CLI} send para node -o 3 -i "$id" -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY | jq -r ".status")
    if [ "${status}" != "3" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588
    hash=$(${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY | jq -r ".status")
    if [ "${status}" != "4" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    node=$(${PARA_CLI} para nodegroup_addrs -t user.p.para. | jq -r '.value|contains("1E5")')
    if [ "${node}" == "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup_addrs -t user.p.para.
        exit 1
    fi

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "20.0000" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi

}

function para_nodemanage_test() {
    para_nodemanage_quit_test
    para_nodemanage_node_join

    echo "=========== # para chain node vote ============="
    id=$(${PARA_CLI} para node_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -t user.p.para. | jq -r ".id")
    if [ -z "${id}" ]; then
        echo "wrong id "
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    hash=$(${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY | jq -r ".status")
    if [ "${status}" != "2" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    node=$(${PARA_CLI} para nodegroup_addrs -t user.p.para. | jq -r '.value|contains("1E5")')
    if [ "${node}" != "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup_addrs -t user.p.para.
        exit 1
    fi

    echo "=========== # para chain node quit ============="
    hash=$(${PARA_CLI} send para node -o 3 -i "$id" -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_list -t user.p.para. -s 3 | jq -r ".addrs[0].targetAddr")
    if [ "${status}" != "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY" ]; then
        echo "wrong join status"
        ${PARA_CLI} para node_list -t user.p.para. -s 3
        exit 1
    fi

    echo "=========== # para chain node vote quit ============="
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588
    hash=$(${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".frozen")
    if [ "$balance" != "0.0000" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY | jq -r ".status")
    if [ "${status}" != "4" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    node=$(${PARA_CLI} para nodegroup_addrs -t user.p.para. | jq -r '.value|contains("1E5")')
    if [ "${node}" == "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup_addrs -t user.p.para.
        exit 1
    fi

    echo "=========== # para chain behalf node vote test ============="
    para_nodemanage_node_behalf_join
    id=$(${PARA_CLI} para node_status -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB -t user.p.para. | jq -r ".id")
    if [ -z "${id}" ]; then
        echo "wrong id "
        ${PARA_CLI} para node_status -t user.p.para. -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB
        exit 1
    fi
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    hash=$(${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB | jq -r ".status")
    if [ "${status}" != "2" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    node=$(${PARA_CLI} para nodegroup_addrs -t user.p.para. | jq -r '.value|contains("1NNa")')
    if [ "${node}" != "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup_addrs -t user.p.para.
        exit 1
    fi

    echo "=========== # para chain node quit ============="
    hash=$(${PARA_CLI} send para node -o 3 -i "$id" -k 0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    echo "=========== # para chain node vote quit ============="
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    ${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d
    hash=$(${PARA_CLI} send para node -o 2 -i "$id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "20.0000" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB | jq -r ".status")
    if [ "${status}" != "4" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    node=$(${PARA_CLI} para nodegroup_addrs -t user.p.para. | jq -r '.value|contains("1NNa")')
    if [ "${node}" == "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup_addrs -t user.p.para.
        exit 1
    fi
}

function para_test() {
    echo "=========== # para chain test ============="
    para_create_nodegroup
    para_nodegroup_behalf_quit_test
    para_nodemanage_test
    token_create "${PARA_CLI}"
    token_transfer "${PARA_CLI}"
    para_cross_transfer_withdraw
    para_cross_transfer_withdraw_for_token
}

function paracross() {
    if [ "${2}" == "init" ]; then
        para_init
    elif [ "${2}" == "config" ]; then
        para_set_wallet
        para_transfer

    elif [ "${2}" == "test" ]; then
        para_test "${1}"
        #        dapp_rpc_test "${3}"

    fi

    if [ "${2}" == "forkInit" ]; then
        para_init
    elif [ "${2}" == "forkConfig" ]; then
        para_transfer
        para_set_wallet
    elif [ "${2}" == "forkCheckRst" ]; then
        checkParaBlockHashfun 30
    fi

    if [ "${2}" == "fork2Init" ]; then
        para_init
    elif [ "${2}" == "fork2Config" ]; then
        para_transfer
        para_set_wallet
    elif [ "${2}" == "fork2CheckRst" ]; then
        checkParaBlockHashfun 30
    fi

}
