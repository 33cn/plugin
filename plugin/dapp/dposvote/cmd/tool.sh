#! /bin/bash
rcpAddr="http://192.168.0.155:9801"
function impKey() {
    key=$1
    lab=$2
    ./chain33-cli --rpc_laddr="${rcpAddr}" account import_key -k "${key}" -l "${lab}"
}

function trans() {
    src=$1
    dst=$2
    coins=$3

    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" coins transfer -a "${coins}" -t "${dst}")
    echo "${tx}"
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${src}" -d "${tx}")
    echo "${sig}"
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "${sig}"
    sleep 3
    ./chain33-cli --rpc_laddr="${rcpAddr}" account balance -a "${dst}"
}

function init() {
    seed=$(./chain33-cli --rpc_laddr=${rcpAddr} seed generate -l 0)
    echo "${seed}"
    echo "save seed..."
    ./chain33-cli --rpc_laddr="${rcpAddr}" seed save -s "${seed}" -p zzh123456
    sleep 1

    echo "unlock wallet..."
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet unlock -p zzh123456
    sleep 1

    echo "import key..."
    impKey "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944" "genesis"
    impKey "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01" "genesis2"
    impKey "5A6A14DA6F5A42835E529D75D87CC8904544F59EEE5387A37D87EEAD194D7EB2" "test1"
    impKey "754F53FCEA0CB1F528918726A49B3551B7F1284D802A1D6AAF4522E8A8DA1B5B" "test2"
    impKey "85CA38F5FB65E5E13403F0704CA6DC479D8D18FFA5D87CE5A966838C9694EAFE" "test3"
    sleep 1

    echo "transfer coins to test1 account..."
    trans "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" "15LsTP6tkYGZcN7tc1Xo2iYifQfowxot3b" 20000

    echo "transfer coins to test2 account..."
    trans "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" "14w5JaGDrXTZwF5Wv51UAtuGgAupenLAok" 20000

    echo "transfer coins to test3 account..."
    trans "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" "1DQUALqaqPUhJX6FWMCqhvnjrkb6ZfrRmf" 20000
}

function send_exec() {
    addr=$1
    coins=$2
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" coins send_exec -a "${coins}" -e dpos)
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "${sig}"
}

function reg() {
    addr=$1
    ip=$2
    key=$3
    echo "dpos regist -a ${addr} -i ${ip} -k ${key}"
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos regist -a "${addr}" -i "${ip}" -k "${key}")
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "$sig"
    sleep 3
}

function reReg() {
    addr=$1
    ip=$2
    key=$3
    echo "dpos reRegist -a ${addr} -i ${ip} -k ${key}"
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos reRegist -a "${addr}" -i "${ip}" -k "${key}")
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "${sig}"
    sleep 3
}
function cancelReg() {
    addr=$1
    key=$2
    echo "dpos cancelRegist -a ${addr} -k ${key}"
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos cancelRegist -a "${addr}" -k "${key}")
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "${sig}"
    sleep 3
}
function vote() {
    addr=$1
    key=$2
    votes=$3
    echo "dpos vote from addr:${addr} to key:${key} $votes votes"
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos vote -k "${key}" -v "${votes}" -a "${addr}")
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "${sig}"
    sleep 3
}
function cancelVote() {
    addr=$1
    key=$2
    index=$3
    echo "dpos cancel vote from addr:${addr} to key:${key} ${votes} votes"
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos cancelVote -k "${key}" -i "${index}")
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "${sig}"
    sleep 3
}
function regM() {
    addr=$1
    key=$2
    m=$3
    cycle=$4
    echo "dpos reg vrfm for addr:${addr}  key:${key} cycle:${cycle}, m:${m}"
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos vrfMRegist -k "${key}" -c "${cycle}" -m "${m}")
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "${sig}"
    sleep 3
}

function regMn() {
    m=$1
    cycle=$2

    addr="15LsTP6tkYGZcN7tc1Xo2iYifQfowxot3b"
    key="03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"
    regM "${addr}" "${key}" "${m}" "${cycle}"

    addr="14w5JaGDrXTZwF5Wv51UAtuGgAupenLAok"
    key="027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3"
    regM "$addr" "$key" "$m" "$cycle"

    addr="1DQUALqaqPUhJX6FWMCqhvnjrkb6ZfrRmf"
    key="03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE"
    regM "$addr" "$key" "$m" "$cycle"
}

function regRP() {
    addr=$1
    key=$2
    r=$3
    p=$4
    cycle=$5
    echo "dpos reg vrfrp for addr:${addr}  key:${key} cycle:${cycle}, r:${r}, p:${p}"
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos vrfRPRegist -k "${key}" -c "${cycle}" -r "${r}" -p "${p}")
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "${sig}"
    sleep 3
}

function regRPn() {
    r=$1
    p=$2
    cycle=$3

    addr="15LsTP6tkYGZcN7tc1Xo2iYifQfowxot3b"
    key="03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"
    regRP "$addr" "$key" "$r" "$p" "$cycle"

    addr="14w5JaGDrXTZwF5Wv51UAtuGgAupenLAok"
    key="027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3"
    regRP "$addr" "$key" "$r" "$p" "$cycle"

    addr="1DQUALqaqPUhJX6FWMCqhvnjrkb6ZfrRmf"
    key="03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE"
    regRP "$addr" "$key" "$r" "$p" "$cycle"
}
function recordCB() {
    cycle=$1
    height=$2
    hash=$3
    key=$4
    addr=$5

    echo "dpos recordCB for key:${key}  cycle:${cycle} height:${height} hash:${hash}"
    tx=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos cbRecord -k "${key}" -c "${cycle}" -m "${height}" -s "${hash}")
    sig=$(./chain33-cli --rpc_laddr="${rcpAddr}" wallet sign -a "${addr}" -d "${tx}")
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet send -d "$sig"
    sleep 3
}

function qtx() {
    tx=$1
    ./chain33-cli --rpc_laddr="${rcpAddr}" tx query -s "${tx}"
}
function qn() {
    result=$(./chain33-cli --rpc_laddr="${rcpAddr}" dpos candidatorQuery -t topN -n "$1")
    echo "$result"
}

function qk() {
    ./chain33-cli --rpc_laddr="${rcpAddr}" dpos candidatorQuery -t pubkeys -k "$1"
}

function qv() {
    ./chain33-cli --rpc_laddr="${rcpAddr}" dpos voteQuery -a "$1" -k "$2"
}
function qvrf() {
    type=$1
    cycle=$2
    ./chain33-cli --rpc_laddr="${rcpAddr}" dpos vrfQuery -t "$type" -c "$cycle"
}

function qvrfn() {
    cycle=$1
    ./chain33-cli --rpc_laddr="${rcpAddr}" dpos vrfQuery -t "topN" -c "$cycle"
}

function qvrfk() {
    cycle=$2
    keys=$1
    ./chain33-cli --rpc_laddr="${rcpAddr}" dpos vrfQuery -t "pubkeys" -c "$cycle" -k "${keys}"
}

function unlock() {
    ./chain33-cli --rpc_laddr="${rcpAddr}" wallet unlock -p zzh123456
}

function qtopn() {
    version=$1
    ./chain33-cli --rpc_laddr="${rcpAddr}" dpos topNQuery -v "${version}"
}

function qcb() {
    type=$1
    param=$2
    if [ "${type}" == "cycle" ]; then
        ./chain33-cli --rpc_laddr="${rcpAddr}" dpos cbQuery -t "cycle" -c "${param}"
    elif [ "${type}" == "height" ]; then
        ./chain33-cli --rpc_laddr="${rcpAddr}" dpos cbQuery -t "height" -m "${param}"
    elif [ "${type}" == "hash" ]; then
        ./chain33-cli --rpc_laddr="${rcpAddr}" dpos cbQuery -t "hash" -s "${param}"
    fi
}

#main

para="$1"
if [ "$para" == "init" ]; then
    init
    send_exec 15LsTP6tkYGZcN7tc1Xo2iYifQfowxot3b 15000
    send_exec 14w5JaGDrXTZwF5Wv51UAtuGgAupenLAok 15000
    send_exec 1DQUALqaqPUhJX6FWMCqhvnjrkb6ZfrRmf 15000
    send_exec 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt 15000

    reg 15LsTP6tkYGZcN7tc1Xo2iYifQfowxot3b 192.168.0.155 03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4
    reg 14w5JaGDrXTZwF5Wv51UAtuGgAupenLAok 192.168.0.194 027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3
    reg 1DQUALqaqPUhJX6FWMCqhvnjrkb6ZfrRmf 192.168.0.100 03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE

    vote 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt 027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3 30
    vote 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt 03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4 100
    vote 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt 03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE 200

elif [ "$para" == "sendExec" ]; then
    send_exec 15LsTP6tkYGZcN7tc1Xo2iYifQfowxot3b 15000
    send_exec 14w5JaGDrXTZwF5Wv51UAtuGgAupenLAok 15000
    send_exec 1DQUALqaqPUhJX6FWMCqhvnjrkb6ZfrRmf 15000
elif [ "$para" == "reg" ]; then
    reg "$2" "$3" "$4"
elif [ "$para" == "cancelReg" ]; then
    cancelReg "$2" "$3"
elif [ "$para" == "reReg" ]; then
    reReg "$2" "$3" "$4"
elif [ "$para" == "vote" ]; then
    vote "$2" "$3" "$4"
elif [ "$para" == "cancelVote" ]; then
    cancelVote "$2" "$3" "$4"
elif [ "$para" == "regM" ]; then
    regM "$2" "$3" "$4" "$5"
elif [ "$para" == "regMn" ]; then
    regMn "$2" "$3"
elif [ "$para" == "regRP" ]; then
    regRP "$2" "$3" "$4" "$5" "$6"
elif [ "$para" == "regRPn" ]; then
    regRPn "$2" "$3" "$4"
elif [ "$para" == "qtx" ]; then
    qtx "$2"
elif [ "$para" == "sendDpos" ]; then
    send_exec "$2" "$3"
elif [ "$para" == "qn" ]; then
    qn "$2"
elif [ "$para" == "qv" ]; then
    qv "$2" "$3"
elif [ "$para" == "qk" ]; then
    qk "$2"
elif [ "$para" == "qvrf" ]; then
    qvrf "$2" "$3"
elif [ "$para" == "qvrfn" ]; then
    qvrfn "$2"
elif [ "$para" == "qvrfk" ]; then
    qvrfk "$2" "$3"
elif [ "$para" == "unlock" ]; then
    unlock
elif [ "$para" == "qtopn" ]; then
    qtopn "$2"
elif [ "$para" == "recordCB" ]; then
    recordCB "$2" "$3" "$4" "$5" "$6"
elif [ "$para" == "qCB" ]; then
    qcb "$2" "$3"
elif [ "$para" == "trans" ]; then
    trans "$2" "$3" "$4"
fi
