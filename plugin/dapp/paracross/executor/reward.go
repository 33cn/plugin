package executor

import (
	"bytes"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

// reward 挖矿奖励， 主要处理挖矿分配逻辑，先实现基本策略，后面根据需求进行重构
func (a *action) reward(nodeStatus *pt.ParacrossNodeStatus, stat *pt.ParacrossHeightStatus) (*types.Receipt, error) {

	//获取挖矿相关配置，这里需注意是共识的高度，而不是交易的高度
	coinReward := types.MGInt("mver.consensus.coinReward", nodeStatus.Height) * types.Coin
	fundReward := types.MGInt("mver.consensus.coinDevFund", nodeStatus.Height) * types.Coin
	fundAddr := types.MGStr("mver.consensus.fundKeyAddr", nodeStatus.Height)

	minerAddrs := getMiners(stat.Details, nodeStatus.BlockHash)
	//分配给矿工的单位奖励
	minerUnit := coinReward / int64(len(minerAddrs))

	receipt := &types.Receipt{Ty: types.ExecOk}
	if minerUnit > 0 {
		//如果不等分转到发展基金
		fundReward += coinReward % minerUnit
		for _, addr := range minerAddrs {
			rep, err := a.coinsAccount.ExecDeposit(addr, a.execaddr, minerUnit)

			if err != nil {
				clog.Error("paracross miner reward deposit err", "height", nodeStatus.Height,
					"execAddr", a.execaddr, "minerAddr", addr, "amount", minerUnit, "err", err)
				return nil, err
			}
			receipt = mergeReceipt(receipt, rep)
		}
	}

	if fundReward > 0 {
		rep, err := a.coinsAccount.ExecDeposit(fundAddr, a.execaddr, fundReward)
		if err != nil {
			clog.Error("paracross fund reward deposit err", "height", nodeStatus.Height,
				"execAddr", a.execaddr, "fundAddr", fundAddr, "amount", fundReward, "err", err)
			return nil, err
		}
		receipt = mergeReceipt(receipt, rep)
	}

	return receipt, nil
}

// getMiners 获取提交共识消息的矿工地址
func getMiners(detail *pt.ParacrossStatusDetails, blockHash []byte) []string {

	addrs := make([]string, 0)
	for i, hash := range detail.BlockHash {
		if bytes.Equal(hash, blockHash) {
			addrs = append(addrs, detail.Addrs[i])
		}
	}
	return addrs
}

//
func mergeReceipt(receipt1, receipt2 *types.Receipt) *types.Receipt {
	if receipt2 != nil {
		receipt1.KV = append(receipt1.KV, receipt2.KV...)
		receipt1.Logs = append(receipt1.Logs, receipt2.Logs...)
	}

	return receipt1
}
