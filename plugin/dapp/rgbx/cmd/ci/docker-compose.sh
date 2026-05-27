#!/usr/bin/env bash

set -e
set -o pipefail

ROOT_DIR=$(cd "$(dirname "$0")" && pwd)
export PATH="${ROOT_DIR}:${PATH}"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/assertions.sh"
# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/bootstrap.sh"

ACTION="run"
PROJECT=""
if [ "$#" -gt 0 ]; then
    case "${1}" in
    run | up | down | init | config | test | native)
        ACTION="${1}"
        PROJECT="${2:-build}"
        ;;
    *)
        PROJECT="${1}"
        ;;
    esac
fi
if [ -z "${PROJECT}" ]; then
    PROJECT="build"
fi

export COMPOSE_PROJECT_NAME="${PROJECT}"

if command -v docker-compose >/dev/null 2>&1; then
    COMPOSE_BIN="docker-compose"
elif docker compose version >/dev/null 2>&1; then
    COMPOSE_BIN="docker compose"
else
    fail "docker compose is required"
fi

function compose_cmd() {
    ${COMPOSE_BIN} "$@"
}

MAIN_CLI="compose_cmd exec -T main /root/chain33-cli --conf=chain33.test.toml"
PARA1_CLI="compose_cmd exec -T para1 /root/chain33-cli --conf=chain33.para1.toml --paraName=user.p.rgbx."
PARA2_CLI="compose_cmd exec -T para2 /root/chain33-cli --conf=chain33.para2.toml --paraName=user.p.rgbx."
PARA3_CLI="compose_cmd exec -T para3 /root/chain33-cli --conf=chain33.para3.toml --paraName=user.p.rgbx."
PARA4_CLI="compose_cmd exec -T para4 /root/chain33-cli --conf=chain33.para4.toml --paraName=user.p.rgbx."
BTCD_RPC_USER="${BTCD_RPC_USER:-root}"
BTCD_RPC_PASS="${BTCD_RPC_PASS:-1314}"
BTC_CTL="compose_cmd exec -T btcd /usr/local/bin/btcctl --configfile=/tmp/btcctl.conf --rpcserver=127.0.0.1:18443 --rpcuser=${BTCD_RPC_USER} --rpcpass=${BTCD_RPC_PASS}"

BTC_NETWORK="${BTC_NETWORK:-regtest}"
BTC_P2P_ADDR="${BTC_P2P_ADDR:-btcd:18444}"
BTC_RPC_ADDR="${BTC_RPC_ADDR:-btcd:18443}"
HOST_BTC_RPC_ADDR="${HOST_BTC_RPC_ADDR:-127.0.0.1:18443}"
PARA_TITLE="${PARA_TITLE:-user.p.rgbx.}"
TSS_THRESHOLD="${TSS_THRESHOLD:-3}"
AUTO_DISCOVER_TSS_PEERS="${AUTO_DISCOVER_TSS_PEERS:-true}"

# 14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
GENESIS_KEY="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"

# 4 para validators: one official TSS node + three validating nodes.
AUTH_ADDR1="1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
AUTH_ADDR2="1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
AUTH_ADDR3="1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
AUTH_ADDR4="1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
AUTH_KEY1="0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
AUTH_KEY2="0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4"
AUTH_KEY3="0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
AUTH_KEY4="0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71"
TSS_PEERS="${TSS_PEERS:-${AUTH_ADDR1},${AUTH_ADDR2},${AUTH_ADDR3},${AUTH_ADDR4}}"

MINT_SYMBOL="${MINT_SYMBOL:-BTC}"
WITHDRAW_DEST_ADDR="${WITHDRAW_DEST_ADDR:-bcrt1qnnwpfpljh5n8m3a8xtf3x5ayvhjjplxmhuexyh}"
BTC_FUNDING_PRIV_HEX="${BTC_FUNDING_PRIV_HEX:-0000000000000000000000000000000000000000000000000000000000000001}"
BTC_DEPOSIT_AMOUNT_SATS="${BTC_DEPOSIT_AMOUNT_SATS:-20000000}"
BTC_WITHDRAW_AMOUNT_SATS="${BTC_WITHDRAW_AMOUNT_SATS:-500000}"
BTC_WITHDRAW_FEE_RATE="${BTC_WITHDRAW_FEE_RATE:-20}"
XBTC_TRANSFER_AMOUNT="${XBTC_TRANSFER_AMOUNT:-500000}"

BTCD_RPC_CERT_IN_CONTAINER="${BTCD_RPC_CERT_IN_CONTAINER:-/btcd/rpc.cert}"
USER_B_KEY="${USER_B_KEY:-${AUTH_KEY2}}"
USER_B_ADDR="${USER_B_ADDR:-${AUTH_ADDR2}}"
# Fixed regtest mining identity (privHex=...0001 -> mrCDr...).
BTCD_MINING_ADDR="mrCDrCybB6J1vRfbwM5hemdJz73FwDBC8r"
BTC_FUNDING_WIF="${BTC_FUNDING_WIF:-}"
USER_MAIN_ADDR="${USER_MAIN_ADDR:-14KEKbYtKKQm4wMthSK9J4La4nAiidGozt}"

function join_csv_as_toml_array() {
    local csv="${1}"
    local out=""
    IFS=',' read -r -a items <<<"${csv}"
    local i
    for ((i = 0; i < ${#items[@]}; i++)); do
        local item
        item=$(echo "${items[$i]}" | xargs)
        if [ -n "${item}" ]; then
            if [ -n "${out}" ]; then
                out="${out},"
            fi
            out="${out}\"${item}\""
        fi
    done
    echo "[${out}]"
}

function config_main() {
    log_step "configure chain33.test.toml"
    local main_cfg="${ROOT_DIR}/chain33.test.toml"
    if [ -f "${main_cfg}" ]; then
        chmod u+w "${main_cfg}" 2>/dev/null || true
    fi
    perl -i -pe 's/^Title=.*/Title="local"/' "${main_cfg}"
    perl -i -pe 's/^TestNet=.*/TestNet=true/' "${main_cfg}"
    perl -i -pe 's/^jrpcBindAddr=.*/jrpcBindAddr="0.0.0.0:8801"/' "${main_cfg}"
    perl -i -pe 's/^grpcBindAddr=.*/grpcBindAddr="0.0.0.0:8802"/' "${main_cfg}"
    perl -i -pe 's/^whitelist=.*/whitelist=["*"]/' "${main_cfg}"
    perl -i -pe 's/^isLevelFee=.*/isLevelFee=false/' "${main_cfg}"

    if ! grep -q '^\[exec.sub.rgbx\]' "${main_cfg}" 2>/dev/null; then
        cat >>"${main_cfg}" <<EOF

[exec.sub.lightclient]
btcNetName="regtest"
commitAddress="${AUTH_ADDR1}"
allowRegtestTimeWarp=true

[exec.sub.rgbx]
commitAddress="${AUTH_ADDR1}"
crossChainAssetPrefix="X"
guardianParachainTitle="${PARA_TITLE}"
EOF
    fi
}

function config_para_file() {
    local file="$1"
    local auth_addr="$2"
    local rank="$3"
    local official="$4"
    local path="${ROOT_DIR}/${file}"

    if [ -f "${path}" ]; then
        chmod u+w "${path}" 2>/dev/null || true
    fi
    cp "${ROOT_DIR}/chain33.para.toml" "${path}"
    chmod u+w "${path}" 2>/dev/null || true
    perl -i -pe 's/^Title=.*/Title="'"${PARA_TITLE}"'"/' "${path}"
    perl -i -pe 's/^TestNet=.*/TestNet=true/' "${path}"
    perl -i -pe 's/^mainChainGrpcAddr=.*/mainChainGrpcAddr="main:8802"/' "${path}"
    perl -i -pe 's/^jrpcBindAddr=.*/jrpcBindAddr="0.0.0.0:8801"/' "${path}"
    perl -i -pe 's/^authAccount=.*/authAccount="'"${auth_addr}"'"/' "${path}"
    perl -i -pe 's/^startHeight=.*/startHeight=1/' "${path}"
    perl -i -pe 's/^enableTSS=.*/enableTSS=true/' "${path}"
    if ! grep -q '^enableTSS=' "${path}" 2>/dev/null; then
        perl -i -0pe 's/\[crypto\]\n/[crypto]\nenableTSS=true\n/' "${path}"
    fi
    # Para nodes must enable DHT before peer discovery and TSS peer collection.
    perl -i -pe 's/^types=.*/types=["dht"]/' "${path}"
    perl -i -pe 's/^enable=.*/enable=true/' "${path}"
    perl -i -pe 's/^waitPid=.*/waitPid=false/' "${path}"

    local toml_peers
    toml_peers=$(join_csv_as_toml_array "${TSS_PEERS}")
    cat >>"${path}" <<EOF

[rpc.sub.light]
clients=["neutrino"]
commitAddr="${auth_addr}"

[rpc.sub.light.neutrino]
isOfficialNode=${official}
netName="${BTC_NETWORK}"
connectPeers=["${BTC_P2P_ADDR}"]
btcBlockInterval=2
blockConfirmations=1
maxUtxoRescanTime=60

[rpc.sub.light.neutrino.btcRPC]
host="${BTC_RPC_ADDR}"
user="${BTCD_RPC_USER}"
pass="${BTCD_RPC_PASS}"
disableTLS=false
certFile="${BTCD_RPC_CERT_IN_CONTAINER}"

[rpc.sub.light.neutrino.tss]
peers=${toml_peers}
threshold=${TSS_THRESHOLD}
rank=${rank}
EOF
}

function init_env() {
    log_step "init realistic topology config (main + 4 para + btcd)"
    config_main
    config_para_file chain33.para1.toml "${AUTH_ADDR1}" 0 true
    config_para_file chain33.para2.toml "${AUTH_ADDR2}" 1 false
    config_para_file chain33.para3.toml "${AUTH_ADDR3}" 1 false
    config_para_file chain33.para4.toml "${AUTH_ADDR4}" 1 false
}

function start_env() {
    log_step "start docker-compose services"
    compose_cmd down
    compose_cmd up --build -d
    sleep 8
    compose_cmd ps
}

function wait_cli_ready() {
    local cli="$1"
    local retries=120
    local i
    for ((i = 0; i < retries; i++)); do
        if ${cli} block last_header >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
    done
    fail "cli not ready: ${cli}"
}

function wait_btcd_ready() {
    log_step "wait btcd ready"
    # Ensure btcd home/config paths exist under bind-mounted volume.
    compose_cmd exec -T btcd sh -c 'mkdir -p /root/.btcd /tmp && [ -f /root/.btcd/btcd.conf ] || : > /root/.btcd/btcd.conf' >/dev/null 2>&1 || true
    local retries=10
    local i
    local last_err=""
    for ((i = 0; i < retries; i++)); do
        local out
        if out=$(${BTC_CTL} --"${BTC_NETWORK}" getblockcount 2>&1); then
            return 0
        fi
        last_err="${out}"
        sleep 1
    done
    if [ -n "${last_err}" ]; then
        log_step "btcd probe last error: ${last_err}"
    fi
    compose_cmd ps btcd || true
    compose_cmd logs --tail=120 btcd || true
    fail "btcd not ready"
}

function block_wait() {
    local cli="$1"
    local delta="$2"
    local cur
    local expect
    cur=$(${cli} block last_header | jq ".height")
    expect=$((cur + delta))
    local count=300
    while [ "${count}" -gt 0 ]; do
        local now
        now=$(${cli} block last_header | jq ".height")
        if [ "${now}" -ge "${expect}" ]; then
            return 0
        fi
        count=$((count - 1))
        sleep 0.2
    done
    fail "wait block timeout for ${cli}, expect=${expect}"
}

function tx_wait() {
    local cli="$1"
    local tx_hash="$2"
    local retries=20
    local i
    for ((i = 0; i < retries; i++)); do
        if ${cli} tx query_hash -s "${tx_hash}" >/dev/null 2>&1; then
            return 0
        fi
        sleep 0.5
    done
    fail "tx not found: ${tx_hash}"
}

function prepare_btcd_mining_identity() {
    local raw_output
    local key_info
    raw_output=$(compose_cmd run --rm --no-deps -T main /root/chain33-cli rgbx btcKeyInfo --net "${BTC_NETWORK}" --privHex "${BTC_FUNDING_PRIV_HEX}")
    # docker compose may print progress lines before command output; keep JSON payload only.
    key_info=$(echo "${raw_output}" | awk 'BEGIN{keep=0} /^\{/ {keep=1} keep==1 {print}')
    assert_non_empty "${key_info}" "btcKeyInfo json output empty"
    BTC_FUNDING_WIF=$(echo "${key_info}" | jq -r '.wif')
    local derived_addr
    derived_addr=$(echo "${key_info}" | jq -r '.address')
    assert_non_empty "${BTC_FUNDING_WIF}" "btc funding wif empty"
    assert_non_empty "${derived_addr}" "btcd mining address empty"
    assert_eq "${derived_addr}" "${BTCD_MINING_ADDR}" "btc funding key does not match fixed BTCD_MINING_ADDR"
    log_step "prepared btcd mining identity address=${BTCD_MINING_ADDR}"
}

function get_main_addr_by_label() {
    local label="$1"
    ${MAIN_CLI} account list | jq -r --arg l "${label}" '[.[]? | select(.label == $l) | .addr][0] // empty'
}

function query_xbtc_balance() {
    local addr="$1"
    ${MAIN_CLI} asset balance -a "${addr}" --asset_exec=rgbx --asset_symbol=XBTC | jq -r '.balance // "0"'
}

function wait_xbtc_balance_not_less_than() {
    local addr="$1"
    local expected="$2"
    local retries="${3:-30}"
    local i
    for ((i = 0; i < retries; i++)); do
        local balance
        balance=$(query_xbtc_balance "${addr}")
        if awk "BEGIN{exit !(${balance} >= ${expected})}"; then
            return 0
        fi
        sleep 1
    done
    fail "xbtc balance not reached, addr=${addr}, expected>=${expected}"
}

function mine_btcd_blocks() {
    local count="$1"
    ${BTC_CTL} --"${BTC_NETWORK}" generate "${count}" >/dev/null
}

function build_mature_coinbase_utxo() {
    assert_non_empty "${BTCD_MINING_ADDR}" "BTCD_MINING_ADDR empty"
    local best_height
    best_height=$(${BTC_CTL} --"${BTC_NETWORK}" getblockcount)
    local mature_height=$((best_height - 100))
    if [ "${mature_height}" -lt 1 ]; then
        fail "no mature coinbase yet: bestHeight=${best_height}"
    fi

    local block_hash
    block_hash=$(${BTC_CTL} --"${BTC_NETWORK}" getblockhash "${mature_height}")
    local coinbase_tx
    coinbase_tx=$(${BTC_CTL} --"${BTC_NETWORK}" getblock "${block_hash}" 1 | jq -r '.tx[0] // empty')
    assert_non_empty "${coinbase_tx}" "coinbase tx empty at height=${mature_height}"
    local tx_json
    tx_json=$(${BTC_CTL} --"${BTC_NETWORK}" getrawtransaction "${coinbase_tx}" 1)
    local vout
    vout=$(echo "${tx_json}" | jq -r --arg a "${BTCD_MINING_ADDR}" '.vout[] | select(.scriptPubKey.addresses[]? == $a) | .n' | head -1)
    assert_non_empty "${vout}" "coinbase vout not found for mining address"
    local amount_sats
    amount_sats=$(echo "${tx_json}" | jq -r --argjson v "${vout}" '.vout[] | select(.n == $v) | (.value * 100000000 | floor)')
    local pk_script
    pk_script=$(echo "${tx_json}" | jq -r --argjson v "${vout}" '.vout[] | select(.n == $v) | .scriptPubKey.hex')
    assert_non_empty "${amount_sats}" "coinbase amount sats empty"
    assert_non_empty "${pk_script}" "coinbase pkScript empty"
    echo "${coinbase_tx}:${vout}:${amount_sats}:${pk_script}"
}

function wait_no_withdraw_pending_for_user() {
    local from_addr="$1"
    local retries=30
    local i
    local cnt=0
    for ((i = 0; i < retries; i++)); do
        cnt=$(${MAIN_CLI} rgbx listPendByFrom -f "${from_addr}" | jq '[.pendingList[]? | select(.actionType == 106)] | length')
        if [ "${cnt}" -eq 0 ]; then
            return 0
        fi
        mine_btcd_blocks 1
        sleep 2
    done
    fail "withdraw pending not cleared for ${from_addr}， cnt=${cnt}"
}

function query_latest_received_sats() {
    local addr="$1"
    local txs
    if ! txs=$(${BTC_CTL} --"${BTC_NETWORK}" searchrawtransactions "${addr}" 1 0 1 0 true 2>/dev/null); then
        echo 0
        return 0
    fi
    echo "${txs}" | jq -r --arg addr "${addr}" '
        [
            .[0]?.vout[]?
            | select((((.scriptPubKey.addresses? // []) | index($addr)) != null) or (.scriptPubKey.address? == $addr))
            | (.value * 100000000 | floor)
        ] | add // 0'
}

function wait_btc_address_received_not_less_than() {
    local addr="$1"
    local expected_sats="$2"
    local retries=3
    local i
    for ((i = 0; i < retries; i++)); do
        local received_sats
        received_sats=$(query_btc_address_received_sats "${addr}")
        if awk "BEGIN{exit !(${received_sats} >= ${expected_sats})}"; then
            log_step "btc address received: addr=${addr}, sats=${received_sats}"
            return 0
        fi
        mine_btcd_blocks 1
        sleep 1
    done
    fail "btc withdraw destination balance not received, addr=${addr}, expect>=${expected_sats}"
}

function prepare_accounts() {
    log_step "prepare wallets and auth keys"

    save_seed_and_unlock "${MAIN_CLI}"
    import_default_keys "${MAIN_CLI}"
    ${MAIN_CLI} account import_key -k "${USER_B_KEY}" -l rgbxUserB >/dev/null || true
    ${MAIN_CLI} send coins transfer -t "${USER_B_ADDR}" -a 100 -k "${GENESIS_KEY}" >/dev/null

    save_seed_and_unlock "${PARA1_CLI}" || true
    save_seed_and_unlock "${PARA2_CLI}" || true
    save_seed_and_unlock "${PARA3_CLI}" || true
    save_seed_and_unlock "${PARA4_CLI}" || true

    ${PARA1_CLI} account import_key -k "${AUTH_KEY1}" -l paraAuth1 >/dev/null || true
    ${PARA2_CLI} account import_key -k "${AUTH_KEY2}" -l paraAuth2 >/dev/null || true
    ${PARA3_CLI} account import_key -k "${AUTH_KEY3}" -l paraAuth3 >/dev/null || true
    ${PARA4_CLI} account import_key -k "${AUTH_KEY4}" -l paraAuth4 >/dev/null || true
}

function collect_peer_names_from_cli() {
    local cli="$1"
    ${cli} net peer | jq -r '[.. | objects | .name? | strings] | .[]' 2>/dev/null || true
}

function peer_count_from_cli() {
    local cli="$1"
    local peers
    peers=$(collect_peer_names_from_cli "${cli}" | sort | uniq | sed '/^$/d')
    echo "${peers}" | sed '/^$/d' | wc -l | xargs
}

function wait_para_dht_discovery() {
    log_step "wait para dht discovery"
    local retries=30
    local i
    for ((i = 0; i < retries; i++)); do
        local count
        count=$(peer_count_from_cli "${PARA1_CLI}")
        # In a 4-para topology, para1 should discover at least the other 3 peers,
        if [ "${count}" -ge 4 ]; then
            log_step "dht ready: para1=${count}"
            return 0
        fi
        sleep 1
    done
    fail "dht peer discovery timeout; check p2p.dht and container network"
}

function discover_tss_peer_names() {
    local retries=30
    local i
    for ((i = 0; i < retries; i++)); do
        # All para nodes use the same tss peers config. Query from one node is enough.
        local peers
        peers=$(collect_peer_names_from_cli "${PARA1_CLI}" | sort | uniq | sed '/^$/d')
        local count
        count=$(echo "${peers}" | sed '/^$/d' | wc -l | xargs)
        if [ "${count}" -ge 4 ]; then
            echo "${peers}" | head -4 | paste -sd, -
            return 0
        fi
        sleep 1
    done
    return 1
}

function rewrite_tss_peers_only() {
    local toml_peers
    toml_peers=$(join_csv_as_toml_array "${TSS_PEERS}")
    local file
    for file in chain33.para1.toml chain33.para2.toml chain33.para3.toml chain33.para4.toml; do
        perl -i -pe "s/^peers=.*/peers=${toml_peers}/" "${ROOT_DIR}/${file}"
    done
}

function copy_para_toml_into_container() {
    local svc="$1"
    local f="chain33.${svc}.toml"
    local cid
    cid=$(compose_cmd ps -q "${svc}" 2>/dev/null | head -1)
    assert_non_empty "${cid}" "no running container for ${svc}; copy ${f} skipped"
    docker cp "${ROOT_DIR}/${f}" "${cid}:/root/${f}"
}

function restart_para_nodes_with_new_toml() {
    log_step "push updated para toml into para1-4 and restart those services (main/btcd unchanged)"
    local svc
    for svc in para1 para2 para3 para4; do
        copy_para_toml_into_container "${svc}"
    done
    compose_cmd restart para1 para2 para3 para4
    sleep 3
}

function apply_dynamic_tss_peers_and_restart() {
    if [ "${AUTO_DISCOVER_TSS_PEERS}" != "true" ]; then
        log_step "skip dynamic peer discovery; use static TSS_PEERS=${TSS_PEERS}"
        return 0
    fi

    wait_para_dht_discovery

    local discovered
    discovered=$(discover_tss_peer_names) || fail "failed to discover 4 peer names from net peer"
    TSS_PEERS="${discovered}"
    log_step "discovered TSS_PEERS=${TSS_PEERS}"

    log_step "rewrite TSS peers in para toml, copy into containers, restart para only"
    rewrite_tss_peers_only
    restart_para_nodes_with_new_toml
    wait_cli_ready "${PARA1_CLI}"
    wait_cli_ready "${PARA2_CLI}"
    wait_cli_ready "${PARA3_CLI}"
    wait_cli_ready "${PARA4_CLI}"
}

function setup_para_nodegroup_on_main() {
    log_step "setup para nodegroup on main chain"
    ${MAIN_CLI} send coins transfer -t "${AUTH_ADDR1}" -a 100 -k "${GENESIS_KEY}" >/dev/null
    ${MAIN_CLI} send coins transfer -t "${AUTH_ADDR2}" -a 100 -k "${GENESIS_KEY}" >/dev/null
    ${MAIN_CLI} send coins transfer -t "${AUTH_ADDR3}" -a 100 -k "${GENESIS_KEY}" >/dev/null
    ${MAIN_CLI} send coins transfer -t "${AUTH_ADDR4}" -a 100 -k "${GENESIS_KEY}" >/dev/null

    ${MAIN_CLI} send coins send_exec -e paracross -a 20 -k "${AUTH_KEY1}" >/dev/null
    ${MAIN_CLI} send coins send_exec -e paracross -a 20 -k "${AUTH_KEY2}" >/dev/null
    ${MAIN_CLI} send coins send_exec -e paracross -a 20 -k "${AUTH_KEY3}" >/dev/null
    ${MAIN_CLI} send coins send_exec -e paracross -a 20 -k "${AUTH_KEY4}" >/dev/null

    local addrs="${AUTH_ADDR1},${AUTH_ADDR2},${AUTH_ADDR3},${AUTH_ADDR4}"
    local apply_hash
    apply_hash=$(${MAIN_CLI} send para nodegroup apply --paraName="${PARA_TITLE}" -a "${addrs}" -c 5 -k "${AUTH_KEY1}")
    assert_length "${apply_hash}" 66
    tx_wait "${MAIN_CLI}" "${apply_hash}"

    local approve_hash
    approve_hash=$(${MAIN_CLI} send para nodegroup approve --paraName="${PARA_TITLE}" -i "${apply_hash}" -c 5 -k "${AUTH_KEY1}")
    assert_length "${approve_hash}" 66
    tx_wait "${MAIN_CLI}" "${approve_hash}"
    ${MAIN_CLI} para nodegroup addrs --paraName="${PARA_TITLE}"
}

function ensure_btc_crosschain_prerequisite() {
    log_step "check BTC cross-chain prerequisite only (no mint bootstrap)"
    set +e
    local info
    info=$(${MAIN_CLI} rgbx getCross -s "${MINT_SYMBOL}" 2>/dev/null)
    local rc=$?
    set -e
    if [ "${rc}" -ne 0 ]; then
        fail "BTC cross-chain info not ready; please pre-configure rgbx BTC asset and cross-chain metadata before running this CI"
    fi
    local symbol
    symbol=$(echo "${info}" | jq -r '.assetSymbol // empty')
    if [ -z "${symbol}" ]; then
        fail "BTC cross-chain info missing; this test does not create asset via mint"
    fi
}

function wait_auto_dkg_commit() {
    log_step "wait auto DKG commit by neutrino+tss"
    local retries=60
    local i
    for ((i = 0; i < retries; i++)); do
        set +e
        local info
        info=$(${MAIN_CLI} rgbx getCross -s "${MINT_SYMBOL}" 2>/dev/null)
        local rc=$?
        set -e
        if [ "${rc}" -eq 0 ]; then
            local tss_addr
            tss_addr=$(echo "${info}" | jq -r '.tssAddress // empty')
            if [ -n "${tss_addr}" ]; then
                log_step "auto DKG done, tssAddress=${tss_addr}"
                return 0
            fi
        fi
        sleep 1
    done
    fail "auto DKG commit timeout"
}

function scenario_user_deposit_via_btc_tx() {
    log_step "scenario: user deposit via btc tx -> service auto submit deposit"
    local before_balance
    before_balance=$(query_xbtc_balance "${USER_MAIN_ADDR}")
    # make sure segwit is activated
    mine_btcd_blocks 450
    local utxo
    utxo=$(build_mature_coinbase_utxo)
    assert_non_empty "${utxo}" "funding utxo empty"

    local tss_addr
    tss_addr=$(${MAIN_CLI} rgbx getCross -s "${MINT_SYMBOL}" | jq -r '.tssAddress // empty')
    assert_non_empty "${tss_addr}" "tssAddress empty before deposit"

    local deposit_tx_hash
    deposit_tx_hash=$(compose_cmd exec -T main /root/chain33-cli rgbx btcDepositTx \
        --net "${BTC_NETWORK}" \
        --rpcHost "${BTC_RPC_ADDR}" \
        --rpcUser "${BTCD_RPC_USER}" \
        --rpcPass "${BTCD_RPC_PASS}" \
        --disableTLS=false \
        --rpcCertFile "${BTCD_RPC_CERT_IN_CONTAINER}" \
        --wif "${BTC_FUNDING_WIF}" \
        --utxo "${utxo}" \
        --tssAddress "${tss_addr}" \
        --chain33Address "${USER_MAIN_ADDR}" \
        --amount "${BTC_DEPOSIT_AMOUNT_SATS}" \
        --fee 500)
    assert_length "${deposit_tx_hash}" 64 "btc deposit tx hash length mismatch"

    mine_btcd_blocks 2
    local expected_balance
    expected_balance=$(awk "BEGIN{printf \"%.8f\", ${before_balance}+${BTC_DEPOSIT_AMOUNT_SATS}/100000000}")
    wait_xbtc_balance_not_less_than "${USER_MAIN_ADDR}" "${expected_balance}"
}

function scenario_user_transfer_crosschain_asset() {
    log_step "scenario: user A transfer cross-chain asset(XBTC) to user B on mainchain"
    local before_a
    local before_b
    before_a=$(query_xbtc_balance "${USER_MAIN_ADDR}")
    before_b=$(query_xbtc_balance "${USER_B_ADDR}")

    local transfer_hash
    local xbtc_transfer_amount
    xbtc_transfer_amount=$(awk "BEGIN{printf \"%.8f\", ${XBTC_TRANSFER_AMOUNT}/100000000}")
    transfer_hash=$(${MAIN_CLI} send rgbx transfer -a "${xbtc_transfer_amount}" -s XBTC \
        -t "${USER_B_ADDR}" -k "${GENESIS_KEY}")
    assert_length "${transfer_hash}" 66 "transfer tx hash"
    # tx_wait "${MAIN_CLI}" "${transfer_hash}"

    local after_a
    local after_b
    after_a=$(query_xbtc_balance "${USER_MAIN_ADDR}")
    after_b=$(query_xbtc_balance "${USER_B_ADDR}")
    expected_a=$(awk "BEGIN{printf \"%.4f\", ${before_a} - ${xbtc_transfer_amount}}")
    expected_b=$(awk "BEGIN{printf \"%.4f\", ${before_b} + ${xbtc_transfer_amount}}")
    assert_balance "${after_a}" "${expected_a}" "user A xbtc not decreased after transfer"
    assert_balance "${after_b}" "${expected_b}" "user B xbtc not increased after transfer"
}

function scenario_user_withdraw_auto_confirm() {
    log_step "scenario: user withdraw on mainchain -> service auto confirm"
    local before_balance
    before_balance=$(query_xbtc_balance "${USER_MAIN_ADDR}")

    local withdraw_hash
    local btc_withdraw_amount
    btc_withdraw_amount=$(awk "BEGIN{printf \"%.8f\", ${BTC_WITHDRAW_AMOUNT_SATS}/100000000}")
    withdraw_hash=$(${MAIN_CLI} send rgbx withdraw -a "${btc_withdraw_amount}" -f "${BTC_WITHDRAW_FEE_RATE}" \
        -d "${WITHDRAW_DEST_ADDR}" -s "${MINT_SYMBOL}" -k "${GENESIS_KEY}")
    assert_length "${withdraw_hash}" 66
    # tx_wait "${MAIN_CLI}" "${withdraw_hash}"
    sleep 10 # wait for withdraw tx to be committed
    wait_no_withdraw_pending_for_user "${USER_MAIN_ADDR}"
    received_sats=$(query_latest_received_sats "${WITHDRAW_DEST_ADDR}")
    local expected_received=$((BTC_WITHDRAW_AMOUNT_SATS - 5000))
    # 允许 ±1000 sats 的误差
    diff=$((received_sats - expected_received))
    if ((diff < 0)); then diff=$((-diff)); fi
    if ((diff >= 1000)); then
        fail "btc withdraw amount mismatch, expect≈${expected_received}, actual=${received_sats}"
    fi

    local after_balance
    after_balance=$(query_xbtc_balance "${USER_MAIN_ADDR}")
    expected_balance=$(awk "BEGIN{printf \"%.4f\", ${before_balance} - ${btc_withdraw_amount}}")
    assert_balance "${after_balance}" "${expected_balance}" "xbtc balance not decreased after withdraw settle"
}

function scenario_restart_recovery() {
    log_step "scenario: restart recovery and pending continuity"
    local before
    before=$(${MAIN_CLI} rgbx listPend -s 0 -i 0 -c 20 | jq -r '.pendingList | length')

    compose_cmd restart main >/dev/null
    wait_cli_ready "${MAIN_CLI}"
    save_seed_and_unlock "${MAIN_CLI}" || true

    local after
    after=$(${MAIN_CLI} rgbx listPend -s 0 -i 0 -c 20 | jq -r '.pendingList | length')
    assert_true "$([ "${after}" -ge 0 ] && echo true || echo false)" "pending list query failed after restart"
    log_step "pending continuity check before=${before}, after=${after}"
}

function scenario_para_health() {
    log_step "scenario: 4 para nodes sync and network health"
    ${PARA1_CLI} net is_sync >/dev/null
    ${PARA2_CLI} net is_sync >/dev/null
    ${PARA3_CLI} net is_sync >/dev/null
    ${PARA4_CLI} net is_sync >/dev/null
}

function scenario_native_asset_mint() {
    log_step "scenario: native asset mint with btc spending confirm"

    # 1. Get a mature coinbase UTXO
    local utxo
    utxo=$(build_mature_coinbase_utxo)
    assert_non_empty "${utxo}" "native mint funding utxo empty"

    local utxo_txid
    local utxo_vout
    local utxo_amount
    local utxo_pkscript
    utxo_txid=$(echo "${utxo}" | cut -d: -f1)
    utxo_vout=$(echo "${utxo}" | cut -d: -f2)
    utxo_amount=$(echo "${utxo}" | cut -d: -f3)
    utxo_pkscript=$(echo "${utxo}" | cut -d: -f4)
    assert_non_empty "${utxo_txid}" "native mint utxo txid empty"
    assert_non_empty "${utxo_pkscript}" "native mint utxo pkscript empty"

    # genesis_out format for mint: hash:index:pkScript
    local genesis_out="${utxo_txid}:${utxo_vout}:${utxo_pkscript}"

    # 2. Mint a native asset on chain33 with the genesis UTXO
    local mint_hash
    mint_hash=$(${MAIN_CLI} send rgbx mint -s NATIVE1 -a 10000 -o "${genesis_out}" -m "6e6174697665316d657461" -k "${GENESIS_KEY}" 2>/dev/null)
    if [ -z "${mint_hash}" ] || [ ${#mint_hash} -lt 64 ]; then
        fail "native mint send failed, hash=${mint_hash}"
    fi

    # Wait for mint tx to be confirmed on chain33
    block_wait "${MAIN_CLI}" 1

    # 3. Get the full mint tx hash from chain33 for OP_RETURN commitment
    # The send output gives us the chain33 tx hash hex (may contain 0x prefix)
    local mint_tx_hash_hex="${mint_hash}"
    mint_tx_hash_hex="${mint_tx_hash_hex#0x}"

    # 4. Construct, sign, and broadcast BTC spending transaction using btcMintSpend
    local fee=1000
    local spend_amount=$((utxo_amount - fee))
    if [ "${spend_amount}" -le 0 ]; then
        fail "native mint insufficient utxo amount=${utxo_amount}, fee=${fee}"
    fi

    local spend_txid
    spend_txid=$(compose_cmd exec -T main /root/chain33-cli rgbx btcMintSpend \
        --net "${BTC_NETWORK}" \
        --rpcHost "${BTC_RPC_ADDR}" \
        --rpcUser "${BTCD_RPC_USER}" \
        --rpcPass "${BTCD_RPC_PASS}" \
        --disableTLS=false \
        --rpcCertFile "${BTCD_RPC_CERT_IN_CONTAINER}" \
        --wif "${BTC_FUNDING_WIF}" \
        --utxo "${utxo}" \
        --destAddress "${BTCD_MINING_ADDR}" \
        --opReturnData "${mint_tx_hash_hex}" \
        --fee "${fee}")
    assert_non_empty "${spend_txid}" "native mint btcMintSpend failed"

    # Mine BTC blocks for confirmations (neutrino uses blockConfirmations=1)
    mine_btcd_blocks 2

    # 5. Wait for the neutrino service to detect the UTXO spend and submit a confirm tx
    log_step "wait for neutrino to confirm native asset mint (NATIVE1)"
    local retries=60
    local i
    for ((i = 0; i < retries; i++)); do
        set +e
        local asset_info
        asset_info=$(${MAIN_CLI} rgbx getAsset -s NATIVE1 2>/dev/null)
        local rc=$?
        set -e
        if [ "${rc}" -eq 0 ]; then
            local symbol
            symbol=$(echo "${asset_info}" | jq -r '.symbol // empty')
            if [ "${symbol}" = "NATIVE1" ]; then
                local total_amount
                total_amount=$(echo "${asset_info}" | jq -r '.totalAmount // 0')
                log_step "native asset NATIVE1 created, totalAmount=${total_amount}"
                return 0
            fi
        fi
        mine_btcd_blocks 1
        sleep 2
    done

    fail "native asset NATIVE1 not created after timeout"
}

function ensure_btcd_network_consistency() {
    if [ "${BTC_NETWORK}" != "regtest" ]; then
        fail "unsupported BTC_NETWORK=${BTC_NETWORK}; this CI uses btcd --regtest only"
    fi
}

function run_tests() {
    ensure_btcd_network_consistency
    prepare_btcd_mining_identity
    wait_btcd_ready
    scenario_para_health
    setup_para_nodegroup_on_main
    # ensure_btc_crosschain_prerequisite
    wait_auto_dkg_commit
    scenario_user_deposit_via_btc_tx
    scenario_user_transfer_crosschain_asset
    scenario_user_withdraw_auto_confirm
    scenario_restart_recovery
    scenario_native_asset_mint
}

function run_native_tests() {
    ensure_btcd_network_consistency
    prepare_btcd_mining_identity
    wait_btcd_ready
    scenario_para_health
    setup_para_nodegroup_on_main
    wait_auto_dkg_commit

    # Mine BTC blocks so build_mature_coinbase_utxo can find a mature coinbase
    mine_btcd_blocks 101
    scenario_native_asset_mint
}

function run_all_tests_wrapper() {
    run_tests
}

function print_logs_hint() {
    log_step "collect logs with: ${COMPOSE_BIN} logs --tail=200"
    log_step "neutrino peer endpoint: ${BTC_P2P_ADDR}"
    log_step "neutrino rpc endpoint: ${BTC_RPC_ADDR}"
    log_step "for neutrino config: netName=${BTC_NETWORK}, connectPeers=[\"${BTC_P2P_ADDR}\"], btcRPC.host=\"${BTC_RPC_ADDR}\", btcRPC.disableTLS=false"
    log_step "TSS roles: para1 official(rank=0), para2-4 validators(rank=1), threshold=${TSS_THRESHOLD}"
}

function do_up_only() {
    ensure_btcd_network_consistency
    mkdir -p "${ROOT_DIR}/btcd-data"
    init_env
    prepare_btcd_mining_identity
    start_env
    wait_btcd_ready
    apply_dynamic_tss_peers_and_restart
    prepare_accounts
}

function do_run_all() {
    do_up_only
    run_tests
    print_logs_hint
}

function do_down() {
    compose_cmd down --remove-orphans
}

case "${ACTION}" in
run)
    do_run_all
    ;;
native)
    do_up_only
    run_native_tests
    print_logs_hint
    ;;
all)
    do_up_only
    run_all_tests_wrapper
    print_logs_hint
    ;;
up)
    do_up_only
    ;;
init)
    init_env
    ;;
config)
    prepare_accounts
    ;;
test)
    run_tests
    ;;
down)
    do_down
    ;;
*)
    fail "unknown action: ${ACTION}"
    ;;
esac
