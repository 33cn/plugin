#!/usr/bin/env bash
# shellcheck disable=SC2128

RPC_TESTFILE=test-rpc.sh
DAPP_TEST_COMMON=dapp-test-common.sh

function dapp_test_rpc() {
    local ip=$1
    local dockerNamePrefix=$2
    echo "============ # dapp rpc test begin ============="
    if [ -d dapptest ]; then
        cp "$DAPP_TEST_COMMON" dapptest/
        cd dapptest || return
        rm -f "jobs.log"
        rm -f "jobsTicket.log"
        rm -rf "outdir"

        dapps=$(find . -maxdepth 1 -type d ! -name dapptest ! -name ticket ! -name . | sed 's/^\.\///' | sort)
        echo "dapps list: $dapps"
        set +e
        parallel -k --jobs 40 --results outdir --joblog ./jobs.log ./{}/"${RPC_TESTFILE}" "$ip" "$dockerNamePrefix" ::: "$dapps"
        local ret=$?
        parallel -k --jobs 1 --results outdir --joblog ./jobsTicket.log ./{}/"${RPC_TESTFILE}" "$ip" ::: "ticket"
        local retTicket=$?

        if [ $ret -ne 0 ]; then
            wrongdapps=$(awk '{print $7,$9 }' jobs.log | grep -a 1 | awk -F '/' '{print $2}')
            parallel -k 'cat ./outdir/1/{}/stderr; cat ./outdir/1/{}/stdout' ::: "$wrongdapps"
        fi

        if [ $retTicket -ne 0 ]; then
            wrongdapps=$(awk '{print $7,$9 }' jobsTicket.log | grep -a 1 | awk -F '/' '{print $2}')
            parallel -k 'cat ./outdir/1/{}/stderr; cat ./outdir/1/{}/stdout' ::: "$wrongdapps"
        fi

        echo "============ # check dapps test log: ============="
        cat ./jobs.log
        cat ./jobsTicket.log
        set -e
        if [ $ret -ne 0 ] || [ $retTicket -ne 0 ]; then
            exit 1
        fi
    fi
    echo "============ # dapp rpc test end ============="
}
