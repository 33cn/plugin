#!/bin/bash
# shellcheck disable=SC2207
set +e

OP="${1}"
path="${2}"

function filterLinter() {
    res=$(
        golangci-lint run --no-config --issues-exit-code=1 --deadline=2m --disable-all \
            --enable=gofmt \
            --enable=gosimple \
            --enable=deadcode \
            --enable=unconvert \
            --enable=interfacer \
            --enable=varcheck \
            --enable=structcheck \
            --enable=goimports \
            --enable=misspell \
            --enable=golint \
            --exclude=underscores
    )
    if [[ ${#res} -gt "0" ]]; then
        echo -e "${res}"
        exit 1
    fi
}

function testLinter() {
    cd "${path}" >/dev/null || exit
    golangci-lint run --no-config --issues-exit-code=1 --deadline=2m --disable-all \
        --enable=gofmt \
        --enable=gosimple \
        --enable=deadcode \
        --enable=unconvert \
        --enable=interfacer \
        --enable=varcheck \
        --enable=structcheck \
        --enable=goimports \
        --enable=misspell \
        --enable=golint \
        --exclude=underscores

    cd - >/dev/null || exit
}

function main() {
    if [ "${OP}" == "filter" ]; then
        filterLinter
    elif [ "${OP}" == "test" ]; then
        testLinter
    fi
}

# run script
main
