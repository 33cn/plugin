#!/usr/bin/env bash

set -e
set -o pipefail

function log_step() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*"
}

function fail() {
    log_step "ERROR: $*"
    exit 1
}

function assert_eq() {
    local actual="${1}"
    local expect="${2}"
    local message="${3:-assert_eq failed}"
    if [ "${actual}" != "${expect}" ]; then
        fail "${message}, expect=${expect}, actual=${actual}"
    fi
}

function assert_non_empty() {
    local value="${1}"
    local message="${2:-assert_non_empty failed}"
    if [ -z "${value}" ] || [ "${value}" = "null" ]; then
        fail "${message}"
    fi
}

function assert_true() {
    local expr="${1}"
    local message="${2:-assert_true failed}"
    if [ "${expr}" != "true" ]; then
        fail "${message}, actual=${expr}"
    fi
}

function wait_until() {
    local retries="${1}"
    local interval="${2}"
    local check_cmd="${3}"
    local message="${4:-wait_until timeout}"

    local i
    for ((i = 0; i < retries; i++)); do
        if eval "${check_cmd}" >/dev/null 2>&1; then
            return 0
        fi
        sleep "${interval}"
    done
    fail "${message}"
}
