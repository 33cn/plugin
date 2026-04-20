#!/usr/bin/env bash

set -e
set -o pipefail

function save_seed_and_unlock() {
    local cli="${1}"
    ${cli} seed save -p 1314fuzamei -s "tortoise main civil member grace happy century convince father cage beach hip maid merry rib" >/dev/null
    ${cli} wallet unlock -p 1314fuzamei -t 0 >/dev/null
}

function import_default_keys() {
    local cli="${1}"
    ${cli} account import_key -k 4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01 -l minerAddr >/dev/null
    ${cli} account import_key -k CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944 -l returnAddr >/dev/null
}

function enable_mining() {
    local cli="${1}"
    ${cli} wallet auto_mine -f 1 >/dev/null
}
