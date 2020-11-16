#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null

source "./dice_guess.sh"

function jvm() {
    if [ "${2}" == "init" ]; then
        return
    elif [ "${2}" == "config" ]; then
        return
    elif [ "${2}" == "test" ]; then
        echo -e "${GRE}=========== $FUNCNAME test begin ===========${NOC}"
        set +e
        set -x
        dice_game_test
        echo -e "${GRE}=========== $FUNCNAME test end ===========${NOC}"
    fi
}
