#!/bin/bash
#
# Code coverage generation
set -e -o pipefail

# go test 从 1.20 起支持一次对多个包生成**合并**的覆盖率文件, 并且会自己按 -p 并行跑包。
# 原来用 bash for 循环逐包串行执行: 全量 126 个有测试的包, 单个最慢约 2 分钟,
# 串行合计约 16 分钟(CI 实测 ut coverage 步骤 991s)。改成一次调用后,
# 墙钟时间降到"最慢的包 + 排队"的量级。
JOBS="${COVERAGE_JOBS:-$(nproc 2>/dev/null || echo 4)}"

PKG_LIST=$(go list ./... | grep -v "vendor" | grep -v "chain33/test" | grep -v "mock" | grep -v "mocks")

# PKG_LIST 需要按空白拆分成多个参数, 这里刻意不加引号
# shellcheck disable=SC2086
go test -p "${JOBS}" -parallel=8 -covermode=count -coverprofile=./coverage.cov ${PKG_LIST}

# Display the global code coverage
go tool cover -func=./coverage.cov

# If needed, generate HTML report
if [ "$1" == "html" ]; then
    go tool cover -html=./coverage.cov -o coverage.html
fi
