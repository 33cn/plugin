#!/usr/bin/env bash

PARA_CLI="docker exec ${NODE3} /root/chain33-cli --paraName user.p.para. --rpc_laddr http://localhost:8901"

PARA_CLI2="docker exec ${NODE2} /root/chain33-cli --paraName user.p.para. --rpc_laddr http://localhost:8901"
PARA_CLI1="docker exec ${NODE1} /root/chain33-cli --paraName user.p.para. --rpc_laddr http://localhost:8901"
PARA_CLI4="docker exec ${NODE4} /root/chain33-cli --paraName user.p.para. --rpc_laddr http://localhost:8901"
PARA_CLI5="docker exec ${NODE5} /root/chain33-cli --paraName user.p.game. --rpc_laddr http://localhost:8901"

PARANAME="para"
PARANAME_GAME="game"
MainLoopCheckForkHeight="60"

xsedfix=""
if [ "$(uname)" == "Darwin" ]; then
    xsedfix=".bak"
fi

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
    sed -i $xsedfix 's/^mainChainGrpcAddr=.*/mainChainGrpcAddr="nginx:8803"/g' "${1}"

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
    set +x
    local CLI=${1}

    if [[ $# -lt 2 ]]; then
        echo -e "${RED}wrong check_tx parameters${NOC}"
        exit_cp_file
    fi

    if [[ ${2} == "" ]]; then
        echo -e "${RED}wrong check_tx txHash is empty${NOC}"
        exit_cp_file
    fi

    local count=0
    while true; do
        ty=$(${CLI} tx query -s "${2}" | jq .receipt.ty)
        if [[ ${ty} != "" ]]; then
            break
        fi

        count=$((count + 1))
        sleep 1

        if [[ ${count} -ge 100 ]]; then
            echo "chain33 query tx for too long"
            break
        fi
    done

    set -x

    ty=$(${CLI} tx query -s "${2}" | jq .receipt.ty)
    if [[ ${ty} != 2 ]]; then
        echo -e "${RED}check tx error, hash is ${2}${NOC}"
        exit 1
    fi
}

function para_cross_transfer_from_parachain() {
    echo "=========== # para cross transfer from parachain test ============="

    balance=$(${PARA_CLI5} account balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -e user.p.game.coins | jq -r ".balance")
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
    hash=$(${PARA_CLI5} send para cross_transfer -a 300 -e user.p.game.coins -s para -t 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${PARA_CLI5}" "${hash}"
    check_cross_transfer_game_balance "300.0000" "0.0000" "${hash}"

    echo "========== #2. main transfer 200 user.p.game.coins.para game chain asset to para chain, main remain=100, parachain=200 ===="
    hash=$(${CLI} --paraName=user.p.para. send para cross_transfer -a 200 -e paracross -s user.p.game.coins.para -t 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    check_cross_transfer_para_balance "100.0000" "200.0000" "${hash}"

    echo "========== #3. withdraw game chain asset to main chain from para chain 50 user.p.game.coins.para,parachain=150,main=150 ===="
    hash=$(${CLI} --paraName=user.p.para. send para cross_transfer -a 50 -e user.p.para.paracross -s paracross.user.p.game.coins.para -t 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    check_cross_transfer_para_balance "150.0000" "150.0000" "${hash}"

    echo "========== #4. withdraw game chain asset to game chain from main chain 50 user.p.game.coins.para,parachain=150,main=100,game=50 ======"
    hash=$(${CLI} --paraName=user.p.game. send para cross_transfer -a 50 -e paracross -s user.p.game.coins.para -t 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu -k 0x128de4afa7c061c00d854a1bca51b58e80a2c292583739e5aebf4c0f778959e1)
    echo "${hash}"
    query_tx "${CLI}" "${hash}"
    check_cross_transfer_game_balance "100.0000" "50.0000" "${hash}"
}

function check_cross_transfer_para_balance() {
    local times=200
    local hash="$3"
    while true; do
        acc=$(${CLI} asset balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu --asset_exec paracross --asset_symbol user.p.game.coins.para -e paracross | jq -r ".balance")
        acc_para=$(${PARA_CLI} asset balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu --asset_exec paracross --asset_symbol paracross.user.p.game.coins.para -e paracross | jq -r ".balance")
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
        acc=$(${CLI} asset balance -a 1BM2xhBk95qoae8zKNDWwAVGgBERhb7DQu --asset_exec paracross --asset_symbol user.p.game.coins.para -e paracross | jq -r ".balance")
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

function para_create_nodegroup() {
    echo "=========== # para chain create node group again ============="
    ##apply
    txhash=$(${PARA_CLI} send para nodegroup apply -a "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4,1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR,1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k,1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs" -c 6 -k 0xd165c84ed37c2a427fea487470ee671b7a0495d68d82607cafbc6348bf23bec5)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"
    id=$txhash

    balance=$(${CLI} account balance -a 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4 -e paracross | jq -r ".frozen")

    echo "=========== # para chain approve node group ============="
    ##approve
    txhash=$(${PARA_CLI} send para nodegroup approve -i "$id" -a "" -c 6 -k 0xc34b5d9d44ac7b754806f761d3d4d2c4fe5214f6b074c19f069c4f5c2a29c8cc)
    echo "tx=$txhash"
    query_tx "${PARA_CLI}" "${txhash}"

    status=$(${PARA_CLI} para nodegroup status | jq -r ".status")
    if [ "$status" != 2 ]; then
        echo "status not approve status=$status"
        exit 1
    fi

    ${PARA_CLI} para nodegroup addrs
}
