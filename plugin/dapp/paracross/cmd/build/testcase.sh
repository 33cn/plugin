#!/usr/bin/env bash
PARA_CLI="docker exec ${NODE3} /root/chain33-para-cli"

PARA_CLI2="docker exec ${NODE2} /root/chain33-para-cli"
PARA_CLI1="docker exec ${NODE1} /root/chain33-para-cli"
PARA_CLI4="docker exec ${NODE4} /root/chain33-para-cli"
MAIN_CLI="docker exec ${NODE3} /root/chain33-cli"

PARANAME="para"

xsedfix=""
if [ "$(uname)" == "Darwin" ]; then
    xsedfix=".bak"
fi

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

    #测试使用，主链也要替换ForkParacrossCommitTx 为300
    # sed -i $xsedfix '/^emptyBlockInterval=.*/a MainParaSelfConsensusForkHeight=300' "${1}"

    # rpc
    sed -i $xsedfix 's/^jrpcBindAddr=.*/jrpcBindAddr="0.0.0.0:8901"/g' "${1}"
    sed -i $xsedfix 's/^grpcBindAddr=.*/grpcBindAddr="0.0.0.0:8902"/g' "${1}"
    sed -i $xsedfix 's/^whitelist=.*/whitelist=["localhost","127.0.0.1","0.0.0.0"]/g' "${1}"
    sed -i $xsedfix 's/^ParaRemoteGrpcClient=.*/ParaRemoteGrpcClient="nginx:8803"/g' "${1}"
}

function para_set_wallet() {
    echo "=========== # para set wallet ============="
    para_import_key "${PARA_CLI}" "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
    para_import_key "${PARA_CLI2}" "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
    para_import_key "${PARA_CLI1}" "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
    para_import_key "${PARA_CLI4}" "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71"
}

function para_import_key() {
    echo "=========== # save seed to wallet ============="
    result=$(${1} seed save -p 1314 -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib" | jq ".isok")
    if [ "${result}" = "false" ]; then
        echo "save seed to wallet error seed, result: ${result}"
        exit 1
    fi

    echo "=========== # unlock wallet ============="
    result=$(${1} wallet unlock -p 1314 -t 0 | jq ".isok")
    if [ "${result}" = "false" ]; then
        exit 1
    fi

    echo "=========== # import private key ============="
    echo "key: ${2}"
    result=$(${1} account import_key -k "${2}" -l paraAuthAccount | jq ".label")
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
    para_transfer2account "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"
    para_transfer2account "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
    para_transfer2account "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
    para_transfer2account "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
    para_transfer2account "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
    block_wait "${CLI}" 1

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

    txhash=$(para_configkey "${PARA_CLI}" "token-blacklist" "BTY")
    echo "txhash=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

    echo "=========== # para chain takeover node group ============="
    txhash=$(${PARA_CLI} send para node -o takeover -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

}

function para_transfer2account() {
    echo "${1}"
    hash1=$(${CLI} send coins transfer -a 10 -n test -t "${1}" -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash1}"
}

function para_configkey() {
    tx=$(${1} config config_tx -o add -c "${2}" -v "${3}")
    sign=$(${CLI} wallet sign -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc -d "${tx}")
    send=$(${CLI} wallet send -d "${sign}")
    echo "${send}"
}

function query_tx() {
    block_wait "${1}" 3

    local times=100
    while true; do
        ret=$(${1} tx query -s "${2}" | jq -r ".tx.hash")
        echo "query hash is ${2}, return ${ret} "
        if [ "${ret}" != "${2}" ]; then
            block_wait "${1}" 2
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
    ${CLI} send bty transfer -a 10 -n test -t $paracrossAddr -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01
    hash=$(${CLI} send para asset_transfer --title user.p.para. -a 1.4 -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"

    sleep 15
    ${CLI} send para asset_withdraw --title user.p.para. -a 0.7 -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01

    local times=100
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

    echo "=========== # send bty to token ============="
    hash=$(${CLI} send bty send_exec -a 500 -e token -n send2exec -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
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

function para_nodemanage_test() {
    echo "=========== # para chain new node join ============="
    hash=$(${PARA_CLI} send para node -o join -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_list -t user.p.para. -s 1 | jq -r ".addrs[0].applyAddr")
    if [ "${status}" != "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" ]; then
        echo "wrong join status"
        ${PARA_CLI} para node_list -t user.p.para. -s 1
        exit 1
    fi

    echo "=========== # para chain node vote ============="
    ${PARA_CLI} send para node -o vote -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -v yes -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node -o vote -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -v yes -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    hash=$(${PARA_CLI} send para node -o vote -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -v yes -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt | jq -r ".status")
    if [ "${status}" != "2" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
        exit 1
    fi

    node=$(${PARA_CLI} para node_group -t user.p.para. | jq -r '.value|contains("14K")')
    if [ "${node}" != "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para node_group -t user.p.para.
        exit 1
    fi

    echo "=========== # para chain node quit ============="
    hash=$(${PARA_CLI} send para node -o quit -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_list -t user.p.para. -s 3 | jq -r ".addrs[0].applyAddr")
    if [ "${status}" != "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" ]; then
        echo "wrong join status"
        ${PARA_CLI} para node_list -t user.p.para. -s 3
        exit 1
    fi

    echo "=========== # para chain node vote quit ============="
    ${PARA_CLI} send para node -o vote -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -v yes -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node -o vote -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -v yes -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    hash=$(${PARA_CLI} send para node -o vote -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -v yes -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node_status -t user.p.para. -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt | jq -r ".status")
    if [ "${status}" != "4" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node_status -t user.p.para. -a 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
        exit 1
    fi

    node=$(${PARA_CLI} para node_group -t user.p.para. | jq -r '.value|contains("14K")')
    if [ "${node}" == "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para node_group -t user.p.para.
        exit 1
    fi

}

function para_test() {
    echo "=========== # para chain test ============="
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
        para_transfer
        para_set_wallet
    elif [ "${2}" == "test" ]; then
        para_test "${1}"
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
