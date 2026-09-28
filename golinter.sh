#!/bin/bash
# shellcheck disable=SC2207
set +e

OP="${1}"
path="${2}"

# 启用的 linter 集合。注意：原先还启用了 deadcode/interfacer/varcheck/structcheck/golint，
# 这 5 个在 golangci-lint v1.38~v1.49 间先后废弃，v1.53.3 只打 deprecation 警告并跳过，
# 属于死配置，故移除。如需补回等价能力：deadcode/varcheck/structcheck -> unused，
# golint -> revive，interfacer 无替代（作者已归档）。
LINTERS=(
    --enable=gofmt
    --enable=gosimple
    --enable=unconvert
    --enable=goimports
    --enable=misspell
)

# 必须加引号：不加引号时 bash 把 [..] 当 glob，且 golangci-lint 将 skip-dirs 按正则处理，
# 字符类 [plugin/dapp/...] 会匹配几乎所有路径，导致全仓被跳过、lint 静默空转。
SKIP_DIRS='plugin/dapp/evm/executor/vm/common/crypto'

function filterLinter() {
    # 不用 res=$(...) 捕获后判断长度：那样只收 stdout，golangci-lint 自身的报错走 stderr
    # 会被丢弃，配置错误时会伪装成"检查通过"。这里直接用退出码。
    golangci-lint run --no-config --issues-exit-code=1 --timeout=20m --disable-all \
        "${LINTERS[@]}" \
        --max-issues-per-linter=0 \
        --max-same-issues=0 \
        --skip-dirs="${SKIP_DIRS}" \
        --exclude=underscores \
        --exclude-use-default=false \
        ./...
}

function testLinter() {
    cd "${path}" >/dev/null || exit
    golangci-lint run --no-config --issues-exit-code=1 --timeout=20m --disable-all \
        "${LINTERS[@]}" \
        --enable=nolintlint \
        --max-issues-per-linter=0 \
        --max-same-issues=0 \
        --skip-dirs="${SKIP_DIRS}" \
        --exclude=underscores \
        ./...

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
