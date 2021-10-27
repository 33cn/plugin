package executor

import (
	"bytes"

	"github.com/33cn/plugin/plugin/dapp/paracross/executor/minerrewards"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

func (a *action) rewardSuperNode(coinReward int64, miners []string, statusHeight int64) (*types.Receipt, int64, error) {
	cfg := a.api.GetConfig()
	receipt := &types.Receipt{Ty: types.ExecOk}

	mode := cfg.MGStr("mver.consensus.paracross.minerMode", a.height)

	rewards, change := minerrewards.MinerRewards[mode].RewardMiners(cfg, coinReward, miners, statusHeight)
	resp, err := a.rewardDeposit(rewards, statusHeight)
	if err != nil {
		return nil, 0, err
	}
	receipt = mergeReceipt(receipt, resp)
	return receipt, change, nil
}

func (a *action) rewardDeposit(rewards []*pt.ParaMinerReward, statusHeight int64) (*types.Receipt, error) {
	receipt := &types.Receipt{}
	for _, v := range rewards {
		rep, err := a.coinsAccount.ExecDeposit(v.Addr, a.execaddr, v.Amount)

		if err != nil {
			clog.Error("paracross super node reward deposit err", "height", statusHeight,
				"execAddr", a.execaddr, "minerAddr", v.Addr, "amount", v.Amount, "err", err)
			return nil, err
		}
		receipt = mergeReceipt(receipt, rep)
	}
	return receipt, nil
}

//奖励委托挖矿账户
func (a *action) rewardBindAddr(coinReward int64, bindAddrList []*pt.ParaBindMinerInfo, statusHeight int64) (*types.Receipt, int64, error) {
	if coinReward <= 0 {
		return nil, 0, nil
	}

	var totalCoins int64
	for _, addr := range bindAddrList {
		totalCoins += addr.BindCoins
	}

	//分配给矿工的单位奖励
	minerUnit := coinReward / totalCoins
	var change int64
	receipt := &types.Receipt{Ty: types.ExecOk}
	if minerUnit > 0 {
		//如果不等分转到发展基金
		change = coinReward % minerUnit
		for _, miner := range bindAddrList {
			rep, err := a.coinsAccount.ExecDeposit(miner.Addr, a.execaddr, minerUnit*miner.BindCoins)
			if err != nil {
				clog.Error("paracross bind miner reward deposit err", "height", statusHeight,
					"execAddr", a.execaddr, "minerAddr", miner.Addr, "amount", minerUnit*miner.BindCoins, "err", err)
				return nil, 0, err
			}
			receipt = mergeReceipt(receipt, rep)
		}
	}
	return receipt, change, nil
}

// reward 挖矿奖励， 主要处理挖矿分配逻辑，先实现基本策略，后面根据需求进行重构
func (a *action) reward(nodeStatus *pt.ParacrossNodeStatus, stat *pt.ParacrossHeightStatus) (*types.Receipt, error) {
	//获取挖矿相关配置，这里需注意是共识的高度，而不是交易的高度
	cfg := a.api.GetConfig()
	//此分叉后 0高度不产生挖矿奖励，也就是以后的新版本默认0高度不产生挖矿奖励
	if nodeStatus.Height == 0 && cfg.IsDappFork(nodeStatus.Height, pt.ParaX, pt.ForkParaFullMinerHeight) {
		return nil, nil
	}

	mode := cfg.MGStr("mver.consensus.paracross.minerMode", a.height)
	if _, ok := minerrewards.MinerRewards[mode]; !ok {
		panic("getReward not be set depend on consensus.paracross.minerMode")
	}
	coinReward, fundReward, coinBaseReward := minerrewards.MinerRewards[mode].GetConfigReward(cfg, nodeStatus.Height)

	fundAddr := cfg.MGStr("mver.consensus.fundKeyAddr", nodeStatus.Height)
	//超级节点地址
	nodeAddrs := getSuperNodes(stat.Details, nodeStatus.BlockHash)
	//委托地址
	bindAddrs, err := a.getBindAddrs(nodeAddrs, nodeStatus.Height)
	if err != nil {
		return nil, err
	}

	// 监督节点地址
	supervisionAddrs := make([]string, 0)
	if stat.SupervisionDetails != nil {
		supervisionAddrs = getSuperNodes(stat.SupervisionDetails, nodeStatus.BlockHash)
	}

	//奖励超级节点
	minderRewards := coinReward
	//如果有委托挖矿地址，则超级节点分baseReward部分，否则全部
	if len(bindAddrs) > 0 {
		minderRewards = coinBaseReward
	}
	receipt := &types.Receipt{Ty: types.ExecOk}

	miners := nodeAddrs
	for _, addr := range supervisionAddrs {
		miners = append(miners, addr)
	}
	r, change, err := a.rewardSuperNode(minderRewards, miners, nodeStatus.Height)
	if err != nil {
		return nil, err
	}
	fundReward += change
	mergeReceipt(receipt, r)

	//奖励委托挖矿地址
	r, change, err = a.rewardBindAddr(coinReward-minderRewards, bindAddrs, nodeStatus.Height)
	if err != nil {
		return nil, err
	}
	fundReward += change
	mergeReceipt(receipt, r)

	//奖励发展基金
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

// getSuperNodes 获取提交共识消息的矿工地址
func getSuperNodes(detail *pt.ParacrossStatusDetails, blockHash []byte) []string {
	addrs := make([]string, 0)
	for i, hash := range detail.BlockHash {
		if bytes.Equal(hash, blockHash) {
			addrs = append(addrs, detail.Addrs[i])
		}
	}
	return addrs
}
