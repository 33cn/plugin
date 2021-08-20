#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null
set -x
set +e

source "./publicTest.sh"
source "./relayerPublic.sh"

# shellcheck disable=SC2120
function mainTest() {
    kill_ebrelayer "chain33 -f"
    sleep 2
    # delete chain33 datadir
    rm ../../datadir ../../logs -rf

    local ganacheName=ganachetest
    # shellcheck disable=SC2155
    local isExit=$(docker inspect ${ganacheName} | jq ".[]" | jq ".Id")
    if [[ ${isExit} != "" ]]; then
        docker stop ${ganacheName}
        docker rm ${ganacheName}
    fi

    kill_all_ebrelayer

    cp ../../../plugin/dapp/cross2eth/ebrelayer/relayer.toml ./relayer.toml
}

mainTest "${1}"

