#!/usr/bin/env bash
# shellcheck disable=SC2050
# shellcheck disable=SC2009
# shellcheck source=/dev/null
set -x
set +e

while [ 1 == 1 ]; do
    pid=$(ps -ef | grep "./ebrelayer" | grep -v 'grep' | awk '{print $2}' | xargs)
    while [ "${pid}" == "" ]; do
        time=$(date "+%m-%d-%H:%M:%S")
        nohup "./ebrelayer" >"./ebrelayer${time}.log" 2>&1 &
        sleep 2

        ./ebcli_A unlock -p "$1"
        sleep 2

        ./ebcli_A unlock -p "$1"
        sleep 2

        pid=$(ps -ef | grep "./ebrelayer" | grep -v 'grep' | awk '{print $2}' | xargs)
    done
    sleep 2
done
