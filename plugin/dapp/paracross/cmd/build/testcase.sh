#!/usr/bin/env bash
PARA_CLI="docker exec ${NODE3} /root/chain33-cli --paraName user.p.para. --rpc_laddr http://localhost:8901"

PARA_CLI2="docker exec ${NODE2} /root/chain33-cli --paraName user.p.para. --rpc_laddr http://localhost:8901"
PARA_CLI1="docker exec ${NODE1} /root/chain33-cli --paraName user.p.para. --rpc_laddr http://localhost:8901"
PARA_CLI4="docker exec ${NODE4} /root/chain33-cli --paraName user.p.para. --rpc_laddr http://localhost:8901"
PARA_CLI5="docker exec ${NODE5} /root/chain33-cli --paraName user.p.game. --rpc_laddr http://localhost:8901"
MAIN_CLI="docker exec ${NODE3} /root/chain33-cli"

PARANAME="para"
PARANAME_GAME="game"
PARA_COIN_FROZEN="5.0000"
MainLoopCheckForkHeight="60"
COINSEXEC="coinsx"

BLSPUB_E5="8920442cf306fccd11e7bde3cfffe183a138a941f471df0818edff5580b3ad7df42850a5cec15e09aef0fdd4489f7c12"
BLSPUB_KS="a3d97d4186c80268fe6d3689dd574599e25df2dffdcff03f7d8ef64a3bd483241b7d0985958990de2d373d5604caf805"
BLSPUB_JR="81307df1fdde8f0e846ed1542c859c1e9daba2553e62e48db0877329c5c63fb86e70b9e2e83263da0eb7fcad275857f8"
BLSPUB_NL="ad1d9ff67d790581fa3659c1817985eeec7c65206e8a873147cd5b6bfe1356d5cd4ed1089462bd11e51705e100c95a6b"
BLSPUB_MC="980287e26d4d44f8c57944ffc096f7d98a460c97dadbffaed14ff0de901fa7f8afc59fcb1805a0b031e5eae5601df1c2"

xsedfix=""
if [ "$(uname)" == "Darwin" ]; then
    xsedfix=".bak"
fi

# shellcheck source=/dev/null
#source test-rpc.sh

function para_init() {
    para_set_toml chain33.para33.toml "$PARANAME" "$1"
    para_set_toml chain33.para32.toml "$PARANAME" "$1"
    para_set_toml chain33.para31.toml "$PARANAME" "$1"
    para_set_toml chain33.para30.toml "$PARANAME" "$1"

    sed -i $xsedfix 's/^authAccount=.*/authAccount="1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"/g' chain33.para33.toml
    sed -i $xsedfix 's/^authAccount=.*/authAccount="1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"/g' chain33.para32.toml
    sed -i $xsedfix 's/^authAccount=.*/authAccount="1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"/g' chain33.para31.toml
    sed -i $xsedfix 's/^authAccount=.*/authAccount="1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"/g' chain33.para30.toml

    para_set_toml chain33.para29.toml "$PARANAME_GAME" "$1"
    sed -i $xsedfix 's/^authAccount=.*/authAccount="1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"/g' chain33.para29.toml
}

function para_set_toml() {
    cp chain33.para.toml "${1}"
    local paraname="$2"

    sed -i $xsedfix 's/^Title.*/Title="user.p.'''"$paraname"'''."/g' "${1}"
    sed -i $xsedfix 's/^# TestNet=.*/TestNet=true/g' "${1}"
    sed -i $xsedfix 's/^startHeight=.*/startHeight=1/g' "${1}"
    sed -i $xsedfix 's/^emptyBlockInterval=.*/emptyBlockInterval=["0:4"]/g' "${1}"

    sed -i $xsedfix 's/^mainForkParacrossCommitTx=.*/mainForkParacrossCommitTx=10/g' "${1}"
    sed -i $xsedfix 's/^mainLoopCheckCommitTxDoneForkHeight=.*/mainLoopCheckCommitTxDoneForkHeight='''$MainLoopCheckForkHeight'''/g' "${1}"

    sed -i $xsedfix 's/^mainBlockHashForkHeight=.*/mainBlockHashForkHeight=1/g' "${1}"
    sed -i $xsedfix 's/^unBindTime=.*/unBindTime=0/g' "${1}"

    #blsSign case
    if [ -n "$3" ]; then
        echo "${1} blssign=$3"
        sed -i $xsedfix '/types=\["dht"\]/!b;n;cenable=true' "${1}"
        sed -i $xsedfix '/emptyBlockInterval=/!b;n;cblsSign=true' "${1}"
        sed -i $xsedfix '/blsSign=/!b;n;cblsLeaderSwitchIntval=10' "${1}"

    fi

    #blockchain
    sed -i $xsedfix 's/^enableReduceLocaldb=.*/enableReduceLocaldb=false/g' "${1}"
    sed -i $xsedfix 's/^enablePushSubscribe=.*/enablePushSubscribe=true/g' "${1}"
    # rpc
    sed -i $xsedfix 's/^jrpcBindAddr=.*/jrpcBindAddr="0.0.0.0:8901"/g' "${1}"
    sed -i $xsedfix 's/^grpcBindAddr=.*/grpcBindAddr="0.0.0.0:8902"/g' "${1}"
    sed -i $xsedfix 's/^whitelist=.*/whitelist=["localhost","127.0.0.1","0.0.0.0"]/g' "${1}"
    sed -i $xsedfix 's/^ParaRemoteGrpcClient=.*/ParaRemoteGrpcClient="nginx:8803"/g' "${1}"

    sed -i $xsedfix 's/^genesis="1JmFaA6unrCFYEWP.*/genesis="1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3"/g' "${1}"
    # shellcheck disable=SC1004
    sed -i $xsedfix 's/^superManager=.*/superManager=["1Bsg9j6gW83sShoee1fZAt9TkUjcrCgA9S",\
                                                        "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",\
                                                        "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"]/g' "${1}"
    # shellcheck disable=SC1004
    sed -i $xsedfix 's/^tokenApprs=.*/tokenApprs=[	"1Bsg9j6gW83sShoee1fZAt9TkUjcrCgA9S",\
	                                                "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK",\
                                                    "1LY8GFia5EiyoTodMLfkB5PHNNpXRqxhyB",\
                                                    "1GCzJDS6HbgTQ2emade7mEJGGWFfA15pS9",\
                                                    "1JYB8sxi4He5pZWHCd3Zi2nypQ4JMB6AxN",\
	                                                "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",]/g' "${1}"
    #autonomy
    sed -i $xsedfix 's/^useBalance=.*/useBalance=true/g' "${1}"
    sed -i $xsedfix 's/^total="16htvcBNS.*/total="1EZrEKPPC36SLRoLQBwLDjzcheiLRZJg49"/g' "${1}"

    sed -i $xsedfix 's/^ForkRootHash=.*/ForkRootHash=0/g' "${1}"
}

function para_set_wallet() {
    echo "=========== # para set wallet ============="
    para_import_wallet "${PARA_CLI}" "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" "paraAuthAccount"
    para_import_wallet "${PARA_CLI2}" "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" "paraAuthAccount"
    para_import_wallet "${PARA_CLI1}" "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" "paraAuthAccount"
    para_import_wallet "${PARA_CLI4}" "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" "paraAuthAccount"

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
    #1PcGKYYoLn1PLLJJodc1UpgWGeFAQasAkx
    para_import_key "${PARA_CLI}" "9d315182e56fde7fadb94408d360203894e5134216944e858f9b31f70e9ecf40" "rpctestpooladdr"

    #super node behalf test
    #1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj
    para_import_key "${PARA_CLI}" "0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5" "behalfnode"
    #1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH
    para_import_key "${PARA_CLI}" "0xfdf2bbff853ecff2e7b86b2a8b45726c6538ca7d1403dc94e50131ef379bdca0" "othernode1"
    #1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB
    para_import_key "${PARA_CLI}" "0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d" "othernode2"

    #cross_transfer
    #1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    para_import_wallet "${PARA_CLI5}" "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" "paraAuthAccount"
    #1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu
    para_import_key "${PARA_CLI5}" "0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1" "cross_transfer"

}

function para_import_wallet() {
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

function para_import_key() {
    local lable=$3
    echo "=========== # import private key ============="
    echo "key: ${2}"
    result=$(${1} account import_key -k "${2}" -l "$lable" | jq ".label")
    if [ -z "${result}" ]; then
        exit 1
    fi
}

function para_transfer() {
    echo "=========== # para chain transfer ============="
    main_transfer2account "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"
    main_transfer2account "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
    main_transfer2account "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
    main_transfer2account "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
    main_transfer2account "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
    main_transfer2account "1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu"
    # super node test
    main_transfer2account "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
    main_transfer2account "1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj"
    main_transfer2account "1Luh4AziYyaC5zP3hUXtXFZS873xAxm6rH" 10
    main_transfer2account "1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB" 10
    #relay test
    main_transfer2account "1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3" 10
    main_transfer2account "1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum" 10
    #relay rpc test
    para_transfer2account "${PARA_CLI}" "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
    para_transfer2account "${PARA_CLI}" "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
    para_transfer2account "${PARA_CLI}" "1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
    para_transfer2account "${PARA_CLI}" "1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"

    #cross_transfer
    #0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1
    para_transfer2account "${PARA_CLI5}" "1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu"
    #rpc test pool addr
    para_transfer2account "${PARA_CLI}" "1PcGKYYoLn1PLLJJodc1UpgWGeFAQasAkx" 500000
    block_wait "${CLI}" 2

    echo "=========== # main chain send to paracross ============="
    #1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY test
    main_transfer2paracross "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588"
    #1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj
    main_transfer2paracross "0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5" 100

    block_wait "${CLI}" 2

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
    echo "${2}"
    local coins=1000
    if [ "$#" -ge 3 ]; then
        coins="$3"
    fi
    hash1=$(${1} send coins transfer -a "$coins" -n transfer -t "${2}" -k 0xCC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944)
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

function para_configkey() {
    tx=$(${1} config config_tx -o add -c "${2}" -v "${3}")
    sign=$(${CLI} wallet sign -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc -d "${tx}")
    send=$(${CLI} wallet send -d "${sign}")
    echo "${send}"
}

function query_tx() {
    block_wait "${1}" 1

    local times=200
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

    ${1} token precreated
    owner=$(${1} token precreated | jq -r ".owner")
    if [ "${owner}" != "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4" ]; then
        echo "wrong pre create owner"
        exit 1
    fi
    total=$(${1} token precreated | jq -r ".total")
    if [ "${total}" != 10000 ]; then
        echo "wrong pre create total"
        exit 1
    fi

    echo "=========== # 2.token finish ============="
    hash=$(${1} send token finish -f 0.001 -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -s GD -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "${hash}"
    query_tx "${1}" "${hash}"

    ${1} token created
    owner=$(${1} token created | jq -r ".owner")
    if [ "${owner}" != "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4" ]; then
        echo "wrong finish created owner"
        exit 1
    fi
    total=$(${1} token created | jq -r ".total")
    if [ "${total}" != 10000 ]; then
        echo "wrong finish created total"
        exit 1
    fi
    ${1} token balance -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -e token -s GD
    balance=$(${1} token balance -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -e token -s GD | jq -r '.[]|.balance')
    if [ "${balance}" != "10000.0000" ]; then
        echo "wrong para token genesis create, should be 10000.0000"
        exit 1
    fi
}

#1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3 also be used in relay rpc_test
function token_transfer() {
    echo "=========== # 1.token transfer ============="
    hash=$(${1} send token transfer -a 100 -s GD -t 1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    hash2=$(${1} send token transfer -a 100 -s GD -t 1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${hash}"
    echo "${hash2}"
    query_tx "${1}" "${hash}"
    query_tx "${1}" "${hash2}"

    ${1} token balance -a 1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3 -e token -s GD
    balance=$(${1} token balance -a 1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3 -e token -s GD | jq -r '.[]|.balance')
    balance2=$(${1} token balance -a 1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum -e token -s GD | jq -r '.[]|.balance')
    if [ "${balance}" != "100.0000" ] || [ "${balance2}" != "100.0000" ]; then
        echo "wrong para token transfer, should be 100.0000, balance=$balance, balace2=$balance2"
        ${1} token balance -a 1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum -e token -s GD
        exit 1
    fi

    echo "=========== # 2.token send exec ============="
    hash=$(${1} send token send_exec -a 100 -s GD -e relay -k 0x22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962)
    hash2=$(${1} send token send_exec -a 100 -s GD -e relay -k ec9162ea5fc2f473ab8240619a0a0f495ba9e9e5d4d9c434b8794a68280236c4)
    echo "${hash}"
    echo "${hash2}"
    query_tx "${1}" "${hash}"
    query_tx "${1}" "${hash2}"

    #user.p.para.relay addr
    # 1464s4B8HbPdUZNR74EBWSH8QLGYgpjr2q
    ${1} token balance -a 1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3 -e token -s GD
    balance=$(${1} token balance -a 1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3 -e relay -s GD | jq -r '.[]|.balance')
    balance2=$(${1} token balance -a 1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum -e relay -s GD | jq -r '.[]|.balance')
    if [ "${balance}" != "100.0000" ] || [ "${balance2}" != "100.0000" ]; then
        echo "wrong para token send exec, should be 100.0000, balance=$balance,balance1=$balance2"
        ${1} token balance -a 1EZKahMRfoMiKp1BewjWrQWoaJ9kmC4hum -e relay -s GD
        exit 1
    fi

    echo "=========== # 3.token withdraw ============="
    hash=$(${1} send token withdraw -a 20 -s GD -e relay -k 0x22968d29c6de695381a8719ef7bf00e2edb6cce500bb59a4fc73c41887610962)
    echo "${hash}"
    query_tx "${1}" "${hash}"

    ${1} token balance -a 1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3 -e relay -s GD
    balance=$(${1} token balance -a 1G5Cjy8LuQex2fuYv3gzb7B8MxAnxLEqt3 -e relay -s GD | jq -r '.[]|.balance')
    if [ "${balance}" != "80.0000" ]; then
        echo "wrong para token withdraw, should be 80.0000"
        exit 1
    fi
}

function para_cross_transfer_withdraw() {
    echo "=========== # para cross transfer/withdraw test ============="
    paracrossAddr=1HPkPopVe3ERfvaAgedDtJQ792taZFEHCe
    ${CLI} account list
    hash=$(${CLI} send coins transfer -a 10 -n test -t $paracrossAddr -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    hash=$(${CLI} send para asset_transfer --paraName user.p.para. -a 1.4 -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    hash2=$(${CLI} send para asset_withdraw --paraName user.p.para. -a 0.7 -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01)

    local times=200
    while true; do
        acc=$(${CLI} account balance -e paracross -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv | jq -r ".balance")
        acc_para=$(${PARA_CLI} asset balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv --asset_exec paracross --asset_symbol "${COINSEXEC}".bty | jq -r ".balance")
        echo "account balance is ${acc}, expect 9.3, para acct balance is ${acc_para},expect 0.7 "
        if [ "${acc}" != "9.3000" ] || [ "${acc_para}" != "0.7000" ]; then
            block_wait "${CLI}" 2
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                ${CLI} tx query -s "$hash2"
                ${PARA_CLI} tx query -s "$hash2"
                ${PARA_CLI} asset balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv --asset_exec paracross --asset_symbol "${COINSEXEC}".bty
                exit 1
            fi
        else
            echo "para_cross_transfer_withdraw success"
            break
        fi
    done

    echo "check asset transfer tx=$hash"
    res=$(${CLI} para asset_txinfo -s "${hash}")
    echo "$res"
    succ=$(jq -r ".success" <<<"$res")
    if [ "${succ}" != "true" ]; then
        echo "para asset transfer tx report fail"
        exit 1
    fi
    echo "check asset withdraw tx=$hash2"
    res=$(${CLI} para asset_txinfo -s "${hash2}")
    echo "$res"
    succ=$(jq -r ".success" <<<"$res")
    if [ "${succ}" != "true" ]; then
        echo "para asset withdraw tx report fail"
        exit 1
    fi
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

    ${CLI} token precreated
    owner=$(${CLI} token precreated | jq -r ".owner")
    if [ "${owner}" != "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv" ]; then
        echo "wrong pre create owner"
        exit 1
    fi
    total=$(${CLI} token precreated | jq -r ".total")
    if [ "${total}" != 10000 ]; then
        echo "wrong pre create total"
        exit 1
    fi

    echo "=========== # 2.token finish ============="
    hash=$(${CLI} send token finish -f 0.001 -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -s FZM -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "${hash}"
    query_tx "${MAIN_CLI}" "${hash}"

    ${CLI} token created
    owner=$(${CLI} token created | jq -r ".owner")
    if [ "${owner}" != "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv" ]; then
        echo "wrong finish created owner"
        exit 1
    fi
    total=$(${CLI} token created | jq -r ".total")
    if [ "${total}" != 10000 ]; then
        echo "wrong finish created total"
        exit 1
    fi
    ${CLI} token balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e token -s FZM
    balance=$(${CLI} token balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -e token -s FZM | jq -r '.[]|.balance')
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
    hash=$(${CLI} send para asset_transfer --paraName user.p.para. -s FZM -a 220 -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv)
    echo "${hash}"
    query_tx "${MAIN_CLI}" "${hash}"

    echo "=========== # 3.asset_withdraw from parachain ============="
    ${CLI} send para asset_withdraw --paraName user.p.para. -a 111 -s FZM -n test -t 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv

    local times=100
    while true; do
        acc=$(${CLI} asset balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv --asset_symbol FZM --asset_exec token -e paracross | jq -r ".balance")
        acc_para=$(${PARA_CLI} asset balance -a 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv --asset_symbol token.FZM --asset_exec paracross -e paracross | jq -r ".balance")
        echo "account balance is ${acc}, expect 224, para acct balance is ${acc_para}, execpt=109 "
        if [ "${acc}" != "224.0000" ] || [ "${acc_para}" != "109.0000" ]; then
            block_wait "${CLI}" 2
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer_withdraw failed"
                #                exit 1
            fi
        else
            echo "para_cross_transfer_withdraw success"
            break
        fi
    done
}

function para_create_nodegroup_gamechain() {
    echo "=========== # game para chain create node group test ============="
    ##apply
    txhash=$(${CLI} --paraName user.p.game. send para nodegroup apply -a "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4" -p "$BLSPUB_KS" -c 5 -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${PARA_CLI5}" "${txhash}"
    id=$txhash

    echo "=========== # game para chain approve node group ============="
    ##approve
    txhash=$(${CLI} --paraName user.p.game. send para nodegroup approve -i "$id" -c 5 -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "tx=$txhash"
    query_tx "${PARA_CLI5}" "${txhash}"

    status=$(${PARA_CLI5} para nodegroup status | jq -r ".status")
    if [ "$status" != 2 ]; then
        echo "status not approve status=$status"
        exit 1
    fi

    ${PARA_CLI5} para nodegroup addrs

}

function para_cross_transfer_from_parachain() {
    echo "=========== # para cross transfer from parachain test ============="

    balance=$(${PARA_CLI5} account balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -e user.p.game."${COINSEXEC}" | jq -r ".balance")
    if [ "${balance}" != "1000.0000" ]; then
        echo "para account 1BM2xhBk should be 1000, real is $balance"
        exit 1
    fi

    hash=$(${PARA_CLI5} send coins send_exec -e user.p.game.paracross -a 300 -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${PARA_CLI5}" "${hash}"

    balance=$(${PARA_CLI5} account balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -e user.p.game.paracross | jq -r ".balance")
    if [ "${balance}" != "300.0000" ]; then
        echo "para paracross account 1BM2xhBk should be 300, real is $balance"
        exit 1
    fi

    echo "========== #1. user.p.game chain transfer to main chain 300 user.p.game.coins.para, remain=0 ==========="
    hash=$(${PARA_CLI5} send para cross_transfer -a 300 -e user.p.game."${COINSEXEC}" -s para -t 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${PARA_CLI5}" "${hash}"
    check_cross_transfer_game_balance "300.0000" "0.0000" "${hash}"

    echo "========== #2. main transfer 200 user.p.game.coins.para game chain asset to para chain, main remain=100, parachain=200 ===="
    hash=$(${CLI} --paraName=user.p.para. send para cross_transfer -a 200 -e paracross -s user.p.game."${COINSEXEC}".para -t 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    check_cross_transfer_para_balance "100.0000" "200.0000" "${hash}"

    echo "========== #3. withdraw game chain asset to main chain from para chain 50 user.p.game.coins.para,parachain=150,main=150 ===="
    hash=$(${CLI} --paraName=user.p.para. send para cross_transfer -a 50 -e user.p.para.paracross -s paracross.user.p.game."${COINSEXEC}".para -t 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    check_cross_transfer_para_balance "150.0000" "150.0000" "${hash}"

    echo "========== #4. withdraw game chain asset to game chain from main chain 50 user.p.game.coins.para,parachain=150,main=100,game=50 ======"
    hash=$(${CLI} --paraName=user.p.game. send para cross_transfer -a 50 -e paracross -s user.p.game."${COINSEXEC}".para -t 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    check_cross_transfer_game_balance "100.0000" "50.0000" "${hash}"
}

function check_cross_transfer_para_balance() {
    local times=200
    local hash="$3"
    while true; do
        acc=$(${CLI} asset balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu --asset_exec paracross --asset_symbol user.p.game."${COINSEXEC}".para -e paracross | jq -r ".balance")
        acc_para=$(${PARA_CLI} asset balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu --asset_exec paracross --asset_symbol paracross.user.p.game."${COINSEXEC}".para -e paracross | jq -r ".balance")
        res=$(${CLI} para asset_txinfo -s "${hash}")
        echo "$res"
        succ=$(jq -r ".success" <<<"$res")
        echo "main account balance is ${acc}, expect $1, para acct balance is ${acc_para},expect $2, cross rst=$succ, expect=true "
        if [ "${acc}" != "$1" ] || [ "${acc_para}" != "$2" ] || [ "${succ}" != "true" ]; then
            block_wait "${CLI}" 2
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer main chain to para chain failed"
                ${CLI} tx query -s "$hash"
                ${PARA_CLI} tx query -s "$hash"
                exit 1
            fi
        else
            echo "para_cross_transfer main chain to para chain  success"
            break
        fi
    done

}

function check_cross_transfer_game_balance() {
    local times=200
    local hash="$3"
    while true; do
        acc=$(${CLI} asset balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu --asset_exec paracross --asset_symbol user.p.game."${COINSEXEC}".para -e paracross | jq -r ".balance")
        acc_para=$(${PARA_CLI5} account balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -e user.p.game.paracross | jq -r ".balance")
        res=$(${CLI} para asset_txinfo -s "${hash}")
        echo "$res"
        succ=$(jq -r ".success" <<<"$res")
        echo "main account balance is ${acc}, expect $1, para exec acct balance is ${acc_para},expect $2, cross rst=$succ, expect=true "
        if [ "${acc}" != "$1" ] || [ "${acc_para}" != "$2" ] || [ "${succ}" != "true" ]; then
            block_wait "${CLI}" 2
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "para_cross_transfer main chain and game chain failed"
                ${CLI} tx query -s "$3"
                ${PARA_CLI5} tx query -s "$3"
                exit 1
            fi
        else
            echo "para_cross_transfer main chain and game chain  success"
            break
        fi
    done

}

function para_create_nodegroup_test() {
    echo "=========== # para chain create node group test ============="
    ##apply
    txhash=$(${PARA_CLI} send para nodegroup apply -a "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4,1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR,1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k,1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs" -c 5 -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"
    id=$txhash

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "20.0000" ]; then
        echo "apply coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain quit node group ============="
    ##quit
    txhash=$(${PARA_CLI} send para nodegroup quit -i "$id" -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"
    newid=$(${PARA_CLI} para nodegroup list -s 3 | jq -r ".ids[0].id")
    if [ -z "$newid" ]; then
        ${PARA_CLI} para nodegroup list -s 3
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
    local blspubs=$BLSPUB_E5,$BLSPUB_KS,$BLSPUB_JR,$BLSPUB_NL,$BLSPUB_MC
    txhash=$(${PARA_CLI} send para nodegroup apply -a "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY,1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4,1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR,1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k,1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs" -p "$blspubs" -c 6 -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"
    id=$txhash

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "30.0000" ]; then
        echo "apply coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain approve node group ============="
    ##approve
    txhash=$(${PARA_CLI} send para nodegroup approve -i "$id" -c 6 -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

    status=$(${PARA_CLI} para nodegroup status | jq -r ".status")
    if [ "$status" != 2 ]; then
        echo "status not approve status=$status"
        exit 1
    fi

    ${PARA_CLI} para nodegroup addrs

    echo "=========== # para chain quit node group fail ============="
    ##quit fail
    txhash=$(${PARA_CLI} send para nodegroup quit -i "$id" -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${CLI}" "${txhash}"
    status=$(${PARA_CLI} para nodegroup status | jq -r ".status")
    if [ "$status" != 2 ]; then
        echo "status quit not approve status=$status"
        exit 1
    fi
    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "30.0000" ]; then
        echo "quit fail coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain modify node group coin=5 ============="
    txhash=$(${PARA_CLI} send para nodegroup modify -c 5 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"
    modifyid=$(${PARA_CLI} para nodegroup list -s 4 | jq -r ".ids[0].id")
    if [ -z "$modifyid" ]; then
        echo "query modify error "
        ${PARA_CLI} para nodegroup_list -s 4
    fi

    ##approve
    txhash=$(${PARA_CLI} send para nodegroup approve -i "$modifyid" -c 5 -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

    id=$(${PARA_CLI} para nodegroup status | jq -r ".id")
    if [ "$modifyid" != "$id" ]; then
        echo " approve new id wrong"
        ${PARA_CLI} para nodegroup_status
        exit 1
    fi
    coins=$(${PARA_CLI} para nodegroup status | jq -r ".coinsFrozen")
    if [ "$coins" != "500000000" ]; then
        echo " approve new coins wrong"
        ${PARA_CLI} para nodegroup_status
        exit 1
    fi
}

function para_nodegroup_behalf_quit_test() {
    echo "=========== # para chain behalf node quit ============="
    status=$(${PARA_CLI} para node addr_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY | jq -r ".status")
    if [ "${status}" != "10" ]; then
        echo "wrong 1E5 status"
        ${PARA_CLI} para node addr_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    hash=$(${PARA_CLI} send para node quit -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"
    id=$hash

    ${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    ${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588
    hash=$(${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node addr_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY | jq -r ".status")
    if [ "${status}" != "11" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node addr_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    node=$(${PARA_CLI} para nodegroup addrs | jq -r '.value|contains("1E5")')
    if [ "${node}" == "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup addrs
        exit 1
    fi

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "24.0000" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi

}

function para_nodemanage_cancel_test() {
    echo "================# para node manage test ================="
    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".balance")
    if [ "$balance" != "$PARA_COIN_FROZEN" ]; then
        echo "balance coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain new node join ============="
    hash=$(${PARA_CLI} send para node join -c 5 -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"
    id=$hash
    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".frozen")
    if [ "$balance" != "$PARA_COIN_FROZEN" ]; then
        echo "frozen coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain node cancel ============="
    hash=$(${PARA_CLI} send para node cancel -i "$id" -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".frozen")
    if [ "$balance" == "$PARA_COIN_FROZEN" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi

}

function para_nodemanage_test() {

    echo "================# para node manage test ================="
    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".balance")
    if [ "$balance" != "$PARA_COIN_FROZEN" ]; then
        echo "balance coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain new node join reject============="
    hash=$(${PARA_CLI} send para node join -c 5 -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".frozen")
    if [ "$balance" != "$PARA_COIN_FROZEN" ]; then
        echo "frozen coinfrozen error balance=$balance"
        exit 1
    fi
    id=$hash
    echo "=========== # para chain node vote ============="

    ${PARA_CLI} send para node vote -i "$id" -v 2 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node vote -i "$id" -v 2 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    hash=$(${PARA_CLI} send para node vote -i "$id" -v 2 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node addr_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY | jq -r ".status")
    if [ "${status}" == "10" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node addr_status -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY
        exit 1
    fi

    status=$(${PARA_CLI} para node id_status -i "$id" | jq -r ".status")
    if [ "${status}" != "3" ]; then
        echo "wrong cancel status"
        ${PARA_CLI} para node id_status -i "$id"
        exit 1
    fi

    node=$(${PARA_CLI} para nodegroup addrs | jq -r '.value|contains("1E5")')
    if [ "${node}" == "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup addrs
        exit 1
    fi

    balance=$(${CLI} account balance -a 1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY -e paracross | jq -r ".frozen")
    if [ "$balance" != "0.0000" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi
    echo "=========== # para chain node quit reject ============="
    txhash=$(${PARA_CLI} send para node quit -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -k 0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588)
    echo "${txhash}"
    query_tx "${PARA_CLI}" "${txhash}"
    id=$txhash

    echo "=========== # para chain node vote quit ============="
    ${PARA_CLI} send para node vote -i "$id" -v 2 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node vote -i "$id" -v 2 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    hash=$(${PARA_CLI} send para node vote -i "$id" -v 2 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node addr_status -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 | jq -r ".status")
    if [ "${status}" != "10" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node addr_status -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
        exit 1
    fi

    status=$(${PARA_CLI} para node id_status -i "$id" | jq -r ".status")
    if [ "${status}" != "3" ]; then
        echo "wrong close status"
        ${PARA_CLI} para node id_status -i "$id"
        exit 1
    fi
    node=$(${PARA_CLI} para nodegroup addrs | jq -r '.value|contains("1KS")')
    if [ "${node}" != "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup addrs
        exit 1
    fi

}

function para_nodemanage_node_behalf_join() {
    echo "=========== # para chain behalf node vote test ============="
    echo "=========== # para chain new node join 1 ============="
    hash=$(${PARA_CLI} send para node join -c 8 -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"
    node1_id=$hash

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "37.0000" ]; then
        echo "1Ka frozen coinfrozen error balance=$balance"
        exit 1
    fi

    balance=$(${CLI} account balance -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB -e paracross | jq -r ".frozen")
    if [ "$balance" == "$PARA_COIN_FROZEN" ]; then
        echo "1LU frozen coinfrozen error balance=$balance"
        exit 1
    fi

    echo "=========== # para chain new node join 2============="
    hash=$(${PARA_CLI} send para node join -c 9 -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"
    id=$hash

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "46.0000" ]; then
        echo "frozen coinfrozen error balance=$balance"
        exit 1
    fi

    ${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    hash=$(${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    status=$(${PARA_CLI} para node addr_status -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB | jq -r ".status")
    if [ "${status}" != "10" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node addr_status -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB
        exit 1
    fi
    status=$(${PARA_CLI} para node id_status -i "$id" | jq -r ".status")
    if [ "${status}" != "3" ]; then
        echo "wrong close status"
        ${PARA_CLI} para node id_status -i "$id"
        exit 1
    fi

    node=$(${PARA_CLI} para nodegroup addrs | jq -r '.value|contains("1NNa")')
    if [ "${node}" != "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup addrs
        exit 1
    fi

    echo "=========== # para chain same node vote again fail ============="
    ${PARA_CLI} send para node vote -i "$node1_id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node vote -i "$node1_id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    hash=$(${PARA_CLI} send para node vote -i "$node1_id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"

    status=$(${PARA_CLI} para node id_status -i "$node1_id" | jq -r ".status")
    if [ "${status}" == "3" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node id_status -i "$node1_id"
        exit 1
    fi

    echo "=========== # para chain node 1 cancel ============="
    hash=$(${PARA_CLI} send para node cancel -i "$node1_id" -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "38.0000" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi
    status=$(${PARA_CLI} para node id_status -i "$node1_id" | jq -r ".status")
    if [ "${status}" != "4" ]; then
        echo "wrong cancel status"
        ${PARA_CLI} para node id_status -i "$node1_id"
        exit 1
    fi

    echo "=========== # para chain node 2 quit ============="
    hash=$(${PARA_CLI} send para node quit -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB -k 0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"
    id=$hash

    echo "=========== # para chain node2 vote quit ============="
    ${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b
    ${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
    ${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x794443611e7369a57b078881445b93b754cbc9b9b8f526535ab9c6d21d29203d
    hash=$(${PARA_CLI} send para node vote -i "$id" -v 1 -k 0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115)
    echo "${hash}"
    query_tx "${PARA_CLI}" "${hash}"

    balance=$(${CLI} account balance -a 1Ka7EPFRqs3v9yreXG6qA4RQbNmbPJCZPj -e paracross | jq -r ".frozen")
    if [ "$balance" != "29.0000" ]; then
        echo "unfrozen coinfrozen error balance=$balance"
        exit 1
    fi

    status=$(${PARA_CLI} para node addr_status -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB | jq -r ".status")
    if [ "${status}" != "11" ]; then
        echo "wrong vote status"
        ${PARA_CLI} para node addr_status -a 1NNaYHkscJaLJ2wUrFNeh6cQXBS4TrFYeB
        exit 1
    fi
    status=$(${PARA_CLI} para node id_status -i "$id" | jq -r ".status")
    if [ "${status}" != "3" ]; then
        echo "wrong cancel status"
        ${PARA_CLI} para node id_status -i "$id"
        exit 1
    fi
    node=$(${PARA_CLI} para nodegroup addrs | jq -r '.value|contains("1NNa")')
    if [ "${node}" == "true" ]; then
        echo "wrong node group addr"
        ${PARA_CLI} para nodegroup addrs
        exit 1
    fi

}

function check_privacy_utxo() {
    echo '#check utxo balance, addr='"${2}"', assetExec='"${3}"', token='"${4}"', expect='"${5}"
    local times=10
    while true; do
        acc=$(${1} privacy showpai -a "${2}" -e "${3}" -s "${4}" | jq -r ".AvailableAmount")
        echo "utxo avail balance is ${acc} "
        if [[ ${acc} == "${5}" ]]; then
            break
        else
            block_wait "${1}" 1
            times=$((times - 1))
            if [ $times -le 0 ]; then
                echo "check privacy utxo failed"
                ${1} privacy showpai -a "${2}" -e "${3}" -s "${4}"
                exit 1
            fi
        fi
    done
}
function privacy_transfer_test() {
    echo "========= # para privacy test ============="
    echo "#enable privacy"
    ${1} privacy enable -a all

    echo "#transfer to privacy exec" #send to user.p.para.privacy for privacy transfer fee
    ${MAIN_CLI} send coins transfer -a 1 -t 15XvcMYK6H1La7ns4yzJhkyurdpXsjjzfQ -k 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
    ${1} send coins transfer -a 10 -t 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    block_wait "${1}" 2
    ${1} send coins send_exec -a 10 -e privacy -k 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    ${1} send token send_exec -a 10 -s GD -e privacy -k 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    block_wait "${1}" 2

    echo "#privacy pub2priv, to=14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
    ${1} send privacy pub2priv -a 9 -p fcbb75f2b96b6d41f301f2d1abc853d697818427819f412f8e4b4e12cacc0814d2c3914b27bea9151b8968ed1732bd241c8788a332b295b731aee8d39a060388 -e "${COINSEXEC}" -s BTY -k 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    ${1} send privacy pub2priv -a 9 -p fcbb75f2b96b6d41f301f2d1abc853d697818427819f412f8e4b4e12cacc0814d2c3914b27bea9151b8968ed1732bd241c8788a332b295b731aee8d39a060388 -e token -s GD -k 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
    check_privacy_utxo "${1}" 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt token GD 9.0000
    check_privacy_utxo "${1}" 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt "${COINSEXEC}" BTY 9.0000

    echo "#privacy priv2priv, to=1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
    ${1} send privacy priv2priv -a 3 -p 5b0ff936ec2d2825a67a270e34d741d96bf6afe5d4b5692de0a1627f635fd0b3d7b14e44d3f8f7526030a7c59de482084161b441a5d66b483d80316e3b91482b -f 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -e "${COINSEXEC}" -s BTY -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    ${1} send privacy priv2priv -a 3 -p 5b0ff936ec2d2825a67a270e34d741d96bf6afe5d4b5692de0a1627f635fd0b3d7b14e44d3f8f7526030a7c59de482084161b441a5d66b483d80316e3b91482b -f 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -e token -s GD -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    check_privacy_utxo "${1}" 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 token GD 3.0000
    check_privacy_utxo "${1}" 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 "${COINSEXEC}" BTY 3.0000

    echo "#privacy priv2pub, to=1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
    ${1} send privacy priv2pub -a 6 -t 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -f 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -e "${COINSEXEC}" -s BTY -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    ${1} send privacy priv2pub -a 6 -t 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -f 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt -e token -s GD -k 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
    check_privacy_utxo "${1}" 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt token GD 0.0000
    check_privacy_utxo "${1}" 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt "${COINSEXEC}" BTY 0.0000
}

function para_test() {
    echo "=========== # para chain test ============="
    para_create_nodegroup
    para_nodegroup_behalf_quit_test
    para_create_nodegroup_gamechain
    #    para_nodemanage_cancel_test
    #    para_nodemanage_test
    #    para_nodemanage_node_behalf_join
    token_create "${PARA_CLI}"
    token_transfer "${PARA_CLI}"
    para_cross_transfer_withdraw
    para_cross_transfer_withdraw_for_token
    privacy_transfer_test "${PARA_CLI}"
    para_cross_transfer_from_parachain
}

function paracross() {
    if [ "${2}" == "init" ]; then
        para_init "${3}"
    elif [ "${2}" == "config" ]; then
        para_set_wallet
        para_transfer

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
