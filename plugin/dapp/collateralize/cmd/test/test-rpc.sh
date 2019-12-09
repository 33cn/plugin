#!/usr/bin/env bash
# shellcheck disable=SC2128
set -e
set -o pipefail

# shellcheck source=/dev/null
source ../dapp-test-common.sh

function main() {
    echo "Collateralize cases has integrated in Issuance test"
}

chain33_debug_function main "$1"
