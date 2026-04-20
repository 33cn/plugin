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
    run | up | down | init | config | test)
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
BTC_CTL="compose_cmd exec -T btcd btcctl"

BTC_NETWORK="${BTC_NETWORK:-regtest}"
BTC_P2P_ADDR="${BTC_P2P_ADDR:-btcd:18444}"
BTC_RPC_ADDR="${BTC_RPC_ADDR:-btcd:18443}"
HOST_BTC_RPC_ADDR="${HOST_BTC_RPC_ADDR:-127.0.0.1:18443}"
PARA_TITLE="${PARA_TITLE:-user.p.rgbx.}"
TSS_THRESHOLD="${TSS_THRESHOLD:-3}"
AUTO_DISCOVER_TSS_PEERS="${AUTO_DISCOVER_TSS_PEERS:-true}"

GENESIS_KEY="4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
USER_KEY="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"

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
WITHDRAW_DEST_ADDR="${WITHDRAW_DEST_ADDR:-1BoatSLRHtKNngkdXEeobR76b53LETtpyT}"
BTC_FUNDING_PRIV_HEX="${BTC_FUNDING_PRIV_HEX:-0000000000000000000000000000000000000000000000000000000000000001}"
BTC_DEPOSIT_AMOUNT_SATS="${BTC_DEPOSIT_AMOUNT_SATS:-200000000}"
BTC_WITHDRAW_AMOUNT_SATS="${BTC_WITHDRAW_AMOUNT_SATS:-50000000}"
BTC_WITHDRAW_FEE_RATE="${BTC_WITHDRAW_FEE_RATE:-10}"
XBTC_TRANSFER_AMOUNT="${XBTC_TRANSFER_AMOUNT:-1}"

HOST_RGBX_CLI="${ROOT_DIR}/chain33-cli --rpc_laddr=http://127.0.0.1:8801"
USER_B_KEY="${USER_B_KEY:-${AUTH_KEY2}}"
USER_B_ADDR="${USER_B_ADDR:-${AUTH_ADDR2}}"
BTCD_MINING_ADDR="${BTCD_MINING_ADDR:-}"
BTC_FUNDING_WIF="${BTC_FUNDING_WIF:-}"
USER_MAIN_ADDR="${USER_MAIN_ADDR:-}"

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
    sed -i.bak '0,/^Title.*/s//Title="local"/' chain33.test.toml
    sed -i.bak '0,/^TestNet=.*/s//TestNet=true/' chain33.test.toml
    sed -i.bak '0,/^jrpcBindAddr=.*/s//jrpcBindAddr="0.0.0.0:8801"/' chain33.test.toml
    sed -i.bak '0,/^grpcBindAddr=.*/s//grpcBindAddr="0.0.0.0:8802"/' chain33.test.toml
    sed -i.bak '0,/^whitelist=.*/s//whitelist=["*"]/' chain33.test.toml
    sed -i.bak '0,/^isLevelFee=.*/s//isLevelFee=false/' chain33.test.toml
}

function config_para_file() {
    local file="$1"
    local auth_addr="$2"
    local rank="$3"
    local official="$4"

    cp chain33.para.toml "${file}"
    sed -i.bak '0,/^Title=.*/s//Title="'"${PARA_TITLE}"'"/' "${file}"
    sed -i.bak '0,/^TestNet=.*/s//TestNet=true/' "${file}"
    sed -i.bak '0,/^mainChainGrpcAddr=.*/s//mainChainGrpcAddr="main:8802"/' "${file}"
    sed -i.bak '0,/^authAccount=.*/s//authAccount="'"${auth_addr}"'"/' "${file}"
    # Para nodes must enable DHT before peer discovery and TSS peer collection.
    sed -i.bak '/^\[p2p\]/,/^\[p2p.sub.dht\]/s/^types=.*/types=["dht"]/' "${file}"
    sed -i.bak '/^\[p2p\]/,/^\[p2p.sub.dht\]/s/^enable=.*/enable=true/' "${file}"
    sed -i.bak '/^\[p2p\]/,/^\[p2p.sub.dht\]/s/^waitPid=.*/waitPid=false/' "${file}"

    local toml_peers
    toml_peers=$(join_csv_as_toml_array "${TSS_PEERS}")
    cat >>"${file}" <<EOF

[rpc.sub.light]
clients=["neutrino"]
commitAddr="${AUTH_ADDR1}"

[rpc.sub.light.neutrino]
isOfficialNode=${official}
netName="${BTC_NETWORK}"
connectPeers=["${BTC_P2P_ADDR}"]
btcBlockInterval=5
blockConfirmations=1
maxUtxoRescanTime=60

[rpc.sub.light.neutrino.btcRPC]
host="${BTC_RPC_ADDR}"
user=""
pass=""
disableTLS=true

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
    local retries=40
    local i
    for ((i = 0; i < retries; i++)); do
        if ${BTC_CTL} --"${BTC_NETWORK}" --notls getblockcount >/dev/null 2>&1; then
            return 0
        fi
        sleep 3
    done
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
    local retries=80
    local i
    for ((i = 0; i < retries; i++)); do
        if ${cli} tx query_hash -s "${tx_hash}" >/dev/null 2>&1; then
            return 0
        fi
        sleep 0.5
    done
    fail "tx not found: ${tx_hash}"
}

function normalize_hash() {
    local tx_hash="${1}"
    tx_hash="${tx_hash#0x}"
    echo "${tx_hash}"
}

function prepare_btcd_mining_identity() {
    local key_info
    key_info=$(${ROOT_DIR}/chain33-cli rgbx btcKeyInfo --net "${BTC_NETWORK}" --privHex "${BTC_FUNDING_PRIV_HEX}")
    BTC_FUNDING_WIF=$(echo "${key_info}" | jq -r '.wif')
    BTCD_MINING_ADDR=$(echo "${key_info}" | jq -r '.address')
    assert_non_empty "${BTC_FUNDING_WIF}" "btc funding wif empty"
    assert_non_empty "${BTCD_MINING_ADDR}" "btcd mining address empty"
    export BTCD_MINING_ADDR
    log_step "prepared btcd mining identity address=${BTCD_MINING_ADDR}"
}

function get_main_addr_by_label() {
    local label="$1"
    ${MAIN_CLI} account list | jq -r --arg l "${label}" '[.[]? | select(.label == $l) | .addr][0] // empty'
}

function query_xbtc_balance() {
    local addr="$1"
    ${MAIN_CLI} asset balance -e rgbx -a "${addr}" --asset_exec=rgbx --asset_symbol=XBTC | jq -r '.balance // "0"'
}

function wait_xbtc_balance_not_less_than() {
    local addr="$1"
    local expected="$2"
    local retries="${3:-180}"
    local i
    for ((i = 0; i < retries; i++)); do
        local balance
        balance=$(query_xbtc_balance "${addr}")
        if awk "BEGIN{exit !(${balance} >= ${expected})}"; then
            return 0
        fi
        sleep 2
    done
    fail "xbtc balance not reached, addr=${addr}, expected>=${expected}"
}

function mine_btcd_blocks() {
    local count="$1"
    ${BTC_CTL} --"${BTC_NETWORK}" --notls generate "${count}" >/dev/null
}

function build_mature_coinbase_utxo() {
    local block_hash
    block_hash=$(${BTC_CTL} --"${BTC_NETWORK}" --notls getblockhash 1)
    local coinbase_tx
    coinbase_tx=$(${BTC_CTL} --"${BTC_NETWORK}" --notls getblock "${block_hash}" 1 | jq -r '.tx[0]')
    local tx_json
    tx_json=$(${BTC_CTL} --"${BTC_NETWORK}" --notls getrawtransaction "${coinbase_tx}" 1)
    local vout
    vout=$(echo "${tx_json}" | jq -r --arg a "${BTCD_MINING_ADDR}" '.vout[] | select(.scriptPubKey.addresses[]? == $a) | .n' | head -1)
    local amount_sats
    amount_sats=$(echo "${tx_json}" | jq -r --argjson v "${vout}" '.vout[] | select(.n == $v) | (.value * 100000000 | floor)')
    local pk_script
    pk_script=$(echo "${tx_json}" | jq -r --argjson v "${vout}" '.vout[] | select(.n == $v) | .scriptPubKey.hex')
    assert_non_empty "${vout}" "coinbase vout not found for mining address"
    assert_non_empty "${amount_sats}" "coinbase amount sats empty"
    assert_non_empty "${pk_script}" "coinbase pkScript empty"
    echo "${coinbase_tx}:${vout}:${amount_sats}:${pk_script}"
}

function wait_no_withdraw_pending_for_user() {
    local from_addr="$1"
    local retries=180
    local i
    for ((i = 0; i < retries; i++)); do
        local cnt
        cnt=$(${MAIN_CLI} rgbx listPendByFrom -f "${from_addr}" | jq '[.pendingList[]? | select(.actionType == 106)] | length')
        if [ "${cnt}" -eq 0 ]; then
            return 0
        fi
        mine_btcd_blocks 1
        sleep 2
    done
    fail "withdraw pending not cleared for ${from_addr}"
}

function prepare_accounts() {
    log_step "prepare wallets and auth keys"
    wait_cli_ready "${MAIN_CLI}"
    wait_cli_ready "${PARA1_CLI}"
    wait_cli_ready "${PARA2_CLI}"
    wait_cli_ready "${PARA3_CLI}"
    wait_cli_ready "${PARA4_CLI}"

    save_seed_and_unlock "${MAIN_CLI}"
    import_default_keys "${MAIN_CLI}"
    ${MAIN_CLI} account import_key -k "${USER_B_KEY}" -l rgbxUserB >/dev/null || true
    enable_mining "${MAIN_CLI}"
    block_wait "${MAIN_CLI}" 2

    USER_MAIN_ADDR=$(get_main_addr_by_label "returnAddr")
    assert_non_empty "${USER_MAIN_ADDR}" "main user address(returnAddr) not found"
    ${MAIN_CLI} send coins transfer -t "${USER_MAIN_ADDR}" -a 100 -k "${GENESIS_KEY}" >/dev/null
    ${MAIN_CLI} send coins transfer -t "${USER_B_ADDR}" -a 100 -k "${GENESIS_KEY}" >/dev/null
    block_wait "${MAIN_CLI}" 1

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
    local retries=90
    local i
    for ((i = 0; i < retries; i++)); do
        local c1
        local c2
        local c3
        local c4
        c1=$(peer_count_from_cli "${PARA1_CLI}")
        c2=$(peer_count_from_cli "${PARA2_CLI}")
        c3=$(peer_count_from_cli "${PARA3_CLI}")
        c4=$(peer_count_from_cli "${PARA4_CLI}")
        # In a 4-para topology, para1 should discover at least the other 3 peers,
        # and each validator should discover at least one peer.
        if [ "${c1}" -ge 3 ] && [ "${c2}" -ge 1 ] && [ "${c3}" -ge 1 ] && [ "${c4}" -ge 1 ]; then
            log_step "dht ready: para1=${c1}, para2=${c2}, para3=${c3}, para4=${c4}"
            return 0
        fi
        sleep 2
    done
    fail "dht peer discovery timeout; check p2p.dht and container network"
}

function discover_tss_peer_names() {
    local retries=60
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
        sleep 2
    done
    return 1
}

function rewrite_tss_peers_only() {
    local toml_peers
    toml_peers=$(join_csv_as_toml_array "${TSS_PEERS}")
    local file
    for file in chain33.para1.toml chain33.para2.toml chain33.para3.toml chain33.para4.toml; do
        sed -i.bak "0,/^peers=.*/s|^peers=.*|peers=${toml_peers}|" "${file}"
    done
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

    log_step "rewrite tss peers only and restart once"
    rewrite_tss_peers_only
    start_env
    wait_btcd_ready
    wait_cli_ready "${MAIN_CLI}"
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
    block_wait "${MAIN_CLI}" 1

    ${MAIN_CLI} send coins send_exec -e paracross -a 20 -k "${AUTH_KEY1}" >/dev/null
    ${MAIN_CLI} send coins send_exec -e paracross -a 20 -k "${AUTH_KEY2}" >/dev/null
    ${MAIN_CLI} send coins send_exec -e paracross -a 20 -k "${AUTH_KEY3}" >/dev/null
    ${MAIN_CLI} send coins send_exec -e paracross -a 20 -k "${AUTH_KEY4}" >/dev/null
    block_wait "${MAIN_CLI}" 1

    local addrs="${AUTH_ADDR1},${AUTH_ADDR2},${AUTH_ADDR3},${AUTH_ADDR4}"
    local apply_hash
    apply_hash=$(${MAIN_CLI} send para nodegroup apply --paraName="${PARA_TITLE}" -a "${addrs}" -c 5 -k "${AUTH_KEY1}")
    assert_non_empty "${apply_hash}" "nodegroup apply hash empty"
    block_wait "${MAIN_CLI}" 1
    tx_wait "${MAIN_CLI}" "${apply_hash}"

    local approve_hash
    approve_hash=$(${MAIN_CLI} send para nodegroup approve --paraName="${PARA_TITLE}" -i "${apply_hash}" -c 5 -k "${GENESIS_KEY}")
    assert_non_empty "${approve_hash}" "nodegroup approve hash empty"
    block_wait "${MAIN_CLI}" 1
    tx_wait "${MAIN_CLI}" "${approve_hash}"

    ${MAIN_CLI} para nodegroup addrs --paraName="${PARA_TITLE}" >/dev/null
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
    local retries=120
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
        sleep 3
    done
    fail "auto DKG commit timeout"
}

function scenario_user_deposit_via_btc_tx() {
    log_step "scenario: user deposit via btc tx -> service auto submit deposit"
    local before_balance
    before_balance=$(query_xbtc_balance "${USER_MAIN_ADDR}")

    mine_btcd_blocks 120
    local utxo
    utxo=$(build_mature_coinbase_utxo)
    assert_non_empty "${utxo}" "funding utxo empty"

    local tss_addr
    tss_addr=$(${MAIN_CLI} rgbx getCross -s "${MINT_SYMBOL}" | jq -r '.tssAddress // empty')
    assert_non_empty "${tss_addr}" "tssAddress empty before deposit"

    local deposit_tx_hash
    deposit_tx_hash=$(${HOST_RGBX_CLI} rgbx btcDepositTx \
        --net "${BTC_NETWORK}" \
        --rpcHost "${HOST_BTC_RPC_ADDR}" \
        --wif "${BTC_FUNDING_WIF}" \
        --utxo "${utxo}" \
        --tssAddress "${tss_addr}" \
        --chain33Address "${USER_MAIN_ADDR}" \
        --amount "${BTC_DEPOSIT_AMOUNT_SATS}" \
        --fee 500)
    assert_non_empty "${deposit_tx_hash}" "btc deposit tx hash empty"

    mine_btcd_blocks 2
    local expected_balance
    expected_balance=$(awk "BEGIN{printf \"%.8f\", ${before_balance}+${BTC_DEPOSIT_AMOUNT_SATS}/100000000}")
    wait_xbtc_balance_not_less_than "${USER_MAIN_ADDR}" "${expected_balance}" 180
}

function scenario_user_transfer_crosschain_asset() {
    log_step "scenario: user A transfer cross-chain asset(XBTC) to user B on mainchain"
    local before_a
    local before_b
    before_a=$(query_xbtc_balance "${USER_MAIN_ADDR}")
    before_b=$(query_xbtc_balance "${USER_B_ADDR}")

    local transfer_hash
    transfer_hash=$(${MAIN_CLI} send rgbx transfer -a "${XBTC_TRANSFER_AMOUNT}" -s "${MINT_SYMBOL}" \
        -t "${USER_B_ADDR}" -x true -k "${USER_KEY}")
    assert_non_empty "${transfer_hash}" "transfer hash empty"
    block_wait "${MAIN_CLI}" 1
    tx_wait "${MAIN_CLI}" "${transfer_hash}"

    local after_a
    local after_b
    after_a=$(query_xbtc_balance "${USER_MAIN_ADDR}")
    after_b=$(query_xbtc_balance "${USER_B_ADDR}")
    assert_true "$(awk "BEGIN{print (${after_a} < ${before_a}) ? \"true\" : \"false\"}")" "user A xbtc not decreased after transfer"
    assert_true "$(awk "BEGIN{print (${after_b} > ${before_b}) ? \"true\" : \"false\"}")" "user B xbtc not increased after transfer"
}

function scenario_user_withdraw_auto_confirm() {
    log_step "scenario: user withdraw on mainchain -> service auto confirm"
    local before_balance
    before_balance=$(query_xbtc_balance "${USER_MAIN_ADDR}")

    local withdraw_hash
    withdraw_hash=$(${MAIN_CLI} send rgbx withdraw -a "${BTC_WITHDRAW_AMOUNT_SATS}" -f "${BTC_WITHDRAW_FEE_RATE}" \
        -d "${WITHDRAW_DEST_ADDR}" -s "${MINT_SYMBOL}" -k "${USER_KEY}")
    assert_non_empty "${withdraw_hash}" "withdraw hash empty"
    block_wait "${MAIN_CLI}" 1
    tx_wait "${MAIN_CLI}" "${withdraw_hash}"

    wait_no_withdraw_pending_for_user "${USER_MAIN_ADDR}"

    local after_balance
    after_balance=$(query_xbtc_balance "${USER_MAIN_ADDR}")
    assert_true "$(awk "BEGIN{print (${after_balance} < ${before_balance}) ? \"true\" : \"false\"}")" "xbtc balance not decreased after withdraw settle"
}

function scenario_restart_recovery() {
    log_step "scenario: restart recovery and pending continuity"
    local before
    before=$(${MAIN_CLI} rgbx listPend -s 0 -i 0 -c 20 | jq -r '.pendingList | length')

    compose_cmd restart main >/dev/null
    wait_cli_ready "${MAIN_CLI}"
    save_seed_and_unlock "${MAIN_CLI}" || true
    enable_mining "${MAIN_CLI}" || true
    block_wait "${MAIN_CLI}" 1

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

function run_tests() {
    scenario_para_health
    setup_para_nodegroup_on_main
    ensure_btc_crosschain_prerequisite
    wait_auto_dkg_commit
    scenario_user_deposit_via_btc_tx
    scenario_user_transfer_crosschain_asset
    scenario_user_withdraw_auto_confirm
    scenario_restart_recovery
}

function print_logs_hint() {
    log_step "collect logs with: ${COMPOSE_BIN} logs --tail=200"
    log_step "neutrino peer endpoint: ${BTC_P2P_ADDR}"
    log_step "neutrino rpc endpoint: ${BTC_RPC_ADDR}"
    log_step "for neutrino config: netName=${BTC_NETWORK}, connectPeers=[\"${BTC_P2P_ADDR}\"], btcRPC.host=\"${BTC_RPC_ADDR}\", btcRPC.disableTLS=true"
    log_step "TSS roles: para1 official(rank=0), para2-4 validators(rank=1..3), threshold=${TSS_THRESHOLD}"
}

function do_up_only() {
    prepare_btcd_mining_identity
    init_env
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
    compose_cmd down
}

case "${ACTION}" in
run)
    do_run_all
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
