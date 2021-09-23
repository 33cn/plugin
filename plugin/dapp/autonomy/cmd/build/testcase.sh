#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null

source "./autonomyTest.sh"

function autonomy() {
    if [ "${2}" == "test" ]; then
        echo "========================== autonomy test =========================="
        set +e
        set -x
        mainTest
        echo "========================== autonomy test end =========================="
    fi
}
