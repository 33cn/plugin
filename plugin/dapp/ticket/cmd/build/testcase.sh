#!/usr/bin/env bash
# shellcheck disable=SC2128
# shellcheck source=/dev/null

# ticket 模块集成测试：主链 [consensus] 本身就是 ticket 共识（chain33.toml
# name="ticket"，genesis 票归 returnAddr），基础链 init_nodes 已导入
# returnAddr 私钥并解锁钱包，这里直接对 NODE3 验证 ticket 生命周期。

RETURN_ADDR="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt" # [consensus] genesis 地址
RETURN_KEY="CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
MINER_ADDR="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv" # init_nodes 导入的挖矿地址
TICKET_PRICE=$((10000 * 100000000))           # 10000 coins

function ticket() {
    if [ "${2}" == "test" ]; then
        echo "========================== ticket test =========================="
        MAIN_HTTP="http://${3}:8801"
        set +e

        # 1. 初始出块正常（minerstart=true 自动挖矿）
        chain33_BlockWait 1 "${MAIN_HTTP}"

        # 2. 节点自动挖矿开关（关->开，验证 ticket 挖矿开关 API）
        req='{"method":"ticket.SetAutoMining","params":[{"flag":false}]}'
        chain33_Http "$req" "${MAIN_HTTP}" '(.error|not) and (.result.isOK == true)' "ticket_SetAutoMiningOff"
        req='{"method":"ticket.SetAutoMining","params":[{"flag":true}]}'
        chain33_Http "$req" "${MAIN_HTTP}" '(.error|not) and (.result.isOK == true)' "ticket_SetAutoMiningOn"

        # 3. 买票并绑定矿工（origin 出钱，票归 bind 矿工）
        req='{"method":"ticket.CreateBindMiner","params":[{"bindAddr":"'"${MINER_ADDR}"'","originAddr":"'"${RETURN_ADDR}"'","amount":'"${TICKET_PRICE}"',"checkBalance":true}]}'
        chain33_Http "$req" "${MAIN_HTTP}" '(.error|not)' "ticket_CreateBindMiner" ".result.txHex"
        chain33_SignAndSendTx "${RETURN_RESP}" "${RETURN_KEY}" "${MAIN_HTTP}"
        chain33_BlockWait 1 "${MAIN_HTTP}"

        # 4. 查询绑定状态的票
        req='{"method":"Chain33.Query","params":[{"execer":"ticket","funcName":"TicketList","payload":{"addr":"'"${MINER_ADDR}"'","status":1}}]}'
        chain33_Http "$req" "${MAIN_HTTP}" '(.error|not) and (.result.tickets|length > 0)' "ticket_TicketList"

        # 5. 票总数 > 0
        req='{"method":"ticket.GetTicketCount","params":[{}]}'
        chain33_Http "$req" "${MAIN_HTTP}" '(.error|not) and (.result > 0)' "ticket_GetTicketCount"

        # 6. 解绑（关闭票）
        req='{"method":"ticket.CloseTickets","params":[{"minerAddress":"'"${MINER_ADDR}"'"}]}'
        chain33_Http "$req" "${MAIN_HTTP}" '(.error|not)' "ticket_CloseTickets"
        chain33_BlockWait 1 "${MAIN_HTTP}"

        # 7. 查询已关闭状态的票
        req='{"method":"Chain33.Query","params":[{"execer":"ticket","funcName":"TicketList","payload":{"addr":"'"${MINER_ADDR}"'","status":2}}]}'
        chain33_Http "$req" "${MAIN_HTTP}" '(.error|not) and (.result.tickets|length > 0)' "ticket_TicketListClosed"

        echo "========================== ticket test end =========================="
    fi
}
