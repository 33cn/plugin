// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

type action struct {
	coinsAccount *account.DB
	db           dbm.KV
	localdb      dbm.KVDB
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
	api          client.QueueProtocolAPI
	tx           *types.Transaction
	exec         *Paracross
}

func newAction(t *Paracross, tx *types.Transaction) *action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &action{t.GetCoinsAccount(), t.GetStateDB(), t.GetLocalDB(), hash, fromaddr,
		t.GetBlockTime(), t.GetHeight(), dapp.ExecAddress(string(tx.Execer)), t.GetAPI(), tx, t}
}

func getNodes(db dbm.KV, key []byte) (map[string]struct{}, []string, error) {
	item, err := db.Get(key)
	if err != nil {
		clog.Info("getNodes", "get db key", string(key), "failed", err)
		if isNotFound(err) {
			err = pt.ErrTitleNotExist
		}
		return nil, nil, errors.Wrapf(err, "db get key:%s", string(key))
	}
	var config types.ConfigItem
	err = types.Decode(item, &config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "decode config")
	}

	value := config.GetArr()
	if value == nil {
		// 在配置地址后，发现配置错了， 删除会出现这种情况
		return map[string]struct{}{}, nil, nil
	}
	var nodesArray []string
	nodesMap := make(map[string]struct{})
	for _, v := range value.Value {
		if _, exist := nodesMap[v]; !exist {
			nodesMap[v] = struct{}{}
			nodesArray = append(nodesArray, v)
		}
	}

	return nodesMap, nodesArray, nil
}

func getConfigManageNodes(db dbm.KV, title string) (map[string]struct{}, []string, error) {
	key := calcManageConfigNodesKey(title)
	return getNodes(db, key)
}

func getParacrossNodes(db dbm.KV, title string) (map[string]struct{}, []string, error) {
	key := calcParaNodeGroupAddrsKey(title)
	return getNodes(db, key)
}

func validTitle(title string) bool {
	if types.IsPara() {
		return types.GetTitle() == title
	}
	return len(title) > 0
}

func validNode(addr string, nodes map[string]struct{}) bool {
	if _, exist := nodes[addr]; exist {
		return exist
	}
	return false
}

func checkCommitInfo(commit *pt.ParacrossCommitAction) error {
	if commit.Status == nil {
		return types.ErrInvalidParam
	}
	clog.Debug("paracross.Commit check input", "height", commit.Status.Height, "mainHeight", commit.Status.MainBlockHeight,
		"mainHash", hex.EncodeToString(commit.Status.MainBlockHash), "blockHash", hex.EncodeToString(commit.Status.BlockHash),
		"preBlockHash", hex.EncodeToString(commit.Status.PreBlockHash))

	if commit.Status.Height == 0 {
		if len(commit.Status.Title) == 0 || len(commit.Status.BlockHash) == 0 {
			return types.ErrInvalidParam
		}
		return nil
	}
	if len(commit.Status.MainBlockHash) == 0 || len(commit.Status.Title) == 0 || commit.Status.Height < 0 ||
		len(commit.Status.PreBlockHash) == 0 || len(commit.Status.BlockHash) == 0 ||
		len(commit.Status.StateHash) == 0 || len(commit.Status.PreStateHash) == 0 {
		return types.ErrInvalidParam
	}

	return nil
}

func isCommitDone(f interface{}, nodes map[string]struct{}, mostSameHash int) bool {
	return float32(mostSameHash) > float32(len(nodes))*float32(2)/float32(3)
}

func makeCommitReceipt(addr string, commit *pt.ParacrossCommitAction, prev, current *pt.ParacrossHeightStatus) *types.Receipt {
	key := calcTitleHeightKey(commit.Status.Title, commit.Status.Height)
	log := &pt.ReceiptParacrossCommit{
		Addr:    addr,
		Status:  commit.Status,
		Prev:    prev,
		Current: current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParacrossCommit,
				Log: types.Encode(log),
			},
		},
	}
}

func makeRecordReceipt(addr string, commit *pt.ParacrossCommitAction) *types.Receipt {
	log := &pt.ReceiptParacrossRecord{
		Addr:   addr,
		Status: commit.Status,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: nil,
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParacrossCommitRecord,
				Log: types.Encode(log),
			},
		},
	}
}

func makeDoneReceipt(addr string, commit *pt.ParacrossCommitAction, current *pt.ParacrossHeightStatus,
	most, commitCount, totalCount int32) *types.Receipt {

	log := &pt.ReceiptParacrossDone{
		TotalNodes:     totalCount,
		TotalCommit:    commitCount,
		MostSameCommit: most,
		Title:          commit.Status.Title,
		Height:         commit.Status.Height,
		StateHash:      commit.Status.StateHash,
		TxCounts:       commit.Status.TxCounts,
		TxResult:       commit.Status.TxResult,
	}
	key := calcTitleKey(commit.Status.Title)
	stat := &pt.ParacrossStatus{
		Title:     commit.Status.Title,
		Height:    commit.Status.Height,
		BlockHash: commit.Status.BlockHash,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(stat)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParacrossCommitDone,
				Log: types.Encode(log),
			},
		},
	}
}

func getMostCommit(stat *pt.ParacrossHeightStatus) (int, string) {
	stats := make(map[string]int)
	n := len(stat.Details.Addrs)
	for i := 0; i < n; i++ {
		if _, ok := stats[string(stat.Details.BlockHash[i])]; ok {
			stats[string(stat.Details.BlockHash[i])]++
		} else {
			stats[string(stat.Details.BlockHash[i])] = 1
		}
	}
	most := -1
	var hash string
	for k, v := range stats {
		if v > most {
			most = v
			hash = k
		}
	}
	return most, hash
}

func hasCommited(addrs []string, addr string) (bool, int) {
	for i, a := range addrs {
		if a == addr {
			return true, i
		}
	}
	return false, 0
}

func getDappForkHeight(fork string) int64 {
	var forkHeight int64
	if types.IsPara() {
		forkHeight = types.Conf("config.consensus.sub.para").GInt("MainForkParacrossCommitTx")
		if forkHeight <= 0 {
			forkHeight = types.MaxHeight
		}
	} else {
		forkHeight = types.GetDappFork(pt.ParaX, fork)
	}
	return forkHeight
}

func (a *action) getNodesGroup(title string) (map[string]struct{}, error) {
	if a.exec.GetMainHeight() < getDappForkHeight(pt.ForkCommitTx) {
		nodes, _, err := getConfigManageNodes(a.db, title)
		if err != nil {
			return nil, errors.Wrapf(err, "getNodes for title:%s", title)
		}
		return nodes, nil
	}

	nodes, _, err := getParacrossNodes(a.db, title)
	if err != nil {
		if errors.Cause(err) != pt.ErrTitleNotExist {
			return nil, errors.Wrapf(err, "getNodes para for title:%s", title)
		}
		nodes, _, err = getConfigManageNodes(a.db, title)
		if err != nil {
			return nil, errors.Wrapf(err, "getNodes manager for title:%s", title)
		}
	}

	return nodes, nil
}

//根据nodes过滤掉可能退出了的addrs
func updateCommitAddrs(stat *pt.ParacrossHeightStatus, nodes map[string]struct{}) {
	details := &pt.ParacrossStatusDetails{}
	for i, addr := range stat.Details.Addrs {
		if _, ok := nodes[addr]; ok {
			details.Addrs = append(details.Addrs, addr)
			details.BlockHash = append(details.BlockHash, stat.Details.BlockHash[i])
		}
	}
	stat.Details = details

}

func (a *action) Commit(commit *pt.ParacrossCommitAction) (*types.Receipt, error) {
	err := checkCommitInfo(commit)
	if err != nil {
		return nil, err
	}

	if !validTitle(commit.Status.Title) {
		return nil, pt.ErrInvalidTitle
	}

	nodes, err := a.getNodesGroup(commit.Status.Title)
	if err != nil {
		return nil, err
	}

	if !validNode(a.fromaddr, nodes) {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "not validNode:%s", a.fromaddr)
	}

	titleStatus, err := getTitle(a.db, calcTitleKey(commit.Status.Title))
	if err != nil {
		return nil, errors.Wrapf(err, "getTitle:%s", a.fromaddr)
	}

	if titleStatus.Height+1 == commit.Status.Height && commit.Status.Height > 0 {
		if !bytes.Equal(titleStatus.BlockHash, commit.Status.PreBlockHash) {
			clog.Error("paracross.Commit", "check PreBlockHash", hex.EncodeToString(titleStatus.BlockHash),
				"commit tx", hex.EncodeToString(commit.Status.PreBlockHash), "commitheit", commit.Status.Height,
				"from", a.fromaddr)
			return nil, pt.ErrParaBlockHashNoMatch
		}
	}

	// 极端情况， 分叉后在平行链生成了 commit交易 并发送之后， 但主链完成回滚，时间序如下
	// 主链   （1）Bn1        （3） rollback-Bn1   （4） commit-done in Bn2
	// 平行链         （2）commit                                 （5） 将得到一个错误的块
	// 所以有必要做这个检测
	commitHeight := commit.Status.MainBlockHeight
	commitHash := commit.Status.MainBlockHash
	if types.IsPara() {
		commitHeight = commit.Status.Height
		commitHash = commit.Status.BlockHash
	}
	blockHash, err := getBlockHash(a.api, commitHeight)
	if err != nil {
		clog.Error("paracross.Commit getBlockHash", "err", err,
			"commit tx height", commitHeight, "isMain", !types.IsPara(), "from", a.fromaddr)
		return nil, err
	}
	if !bytes.Equal(blockHash.Hash, commitHash) && commit.Status.Height > 0 {
		clog.Error("paracross.Commit blockHash not match", "isMain", !types.IsPara(), "db", hex.EncodeToString(blockHash.Hash),
			"commit tx", hex.EncodeToString(commitHash), "commitHeight", commit.Status.Height,
			"commitMainHeight", commit.Status.MainBlockHeight, "from", a.fromaddr)
		return nil, types.ErrBlockHashNoMatch
	}

	clog.Debug("paracross.Commit check input done")
	// 在完成共识之后来的， 增加 record log， 只记录不修改已经达成的共识
	if commit.Status.Height <= titleStatus.Height {
		clog.Debug("paracross.Commit record", "node", a.fromaddr, "titile", commit.Status.Title,
			"height", commit.Status.Height)
		return makeRecordReceipt(a.fromaddr, commit), nil
	}

	// 未共识处理， 接受当前高度以及后续高度
	stat, err := getTitleHeight(a.db, calcTitleHeightKey(commit.Status.Title, commit.Status.Height))
	if err != nil && !isNotFound(err) {
		clog.Error("paracross.Commit getTitleHeight failed", "err", err)
		return nil, err
	}

	var receipt *types.Receipt
	if isNotFound(err) {
		stat = &pt.ParacrossHeightStatus{
			Status: pt.ParacrossStatusCommiting,
			Title:  commit.Status.Title,
			Height: commit.Status.Height,
			Details: &pt.ParacrossStatusDetails{
				Addrs:     []string{a.fromaddr},
				BlockHash: [][]byte{commit.Status.BlockHash},
			},
		}
		if a.exec.GetMainHeight() >= getDappForkHeight(pt.ForkCommitTx) {
			stat.MainHeight = commit.Status.MainBlockHeight
			stat.MainHash = commit.Status.MainBlockHash
		}
		receipt = makeCommitReceipt(a.fromaddr, commit, nil, stat)
	} else {
		var copyStat pt.ParacrossHeightStatus
		err = deepCopy(&copyStat, stat)
		if err != nil {
			clog.Error("paracross.Commit deep copy fail", "copy", copyStat, "stat", stat)
			return nil, err
		}
		// 如有分叉， 同一个节点可能再次提交commit交易
		found, index := hasCommited(stat.Details.Addrs, a.fromaddr)
		if found {
			stat.Details.BlockHash[index] = commit.Status.BlockHash
		} else {
			stat.Details.Addrs = append(stat.Details.Addrs, a.fromaddr)
			stat.Details.BlockHash = append(stat.Details.BlockHash, commit.Status.BlockHash)
		}
		receipt = makeCommitReceipt(a.fromaddr, commit, &copyStat, stat)
	}

	if types.IsDappFork(commit.Status.MainBlockHeight, pt.ParaX, pt.ForkCommitTx) {
		updateCommitAddrs(stat, nodes)
	}

	clog.Info("paracross.Commit commit", "stat.title", stat.Title, "stat.height", stat.Height, "notes", len(nodes))
	for i, v := range stat.Details.Addrs {
		clog.Info("paracross.Commit commit detail", "addr", v, "hash", hex.EncodeToString(stat.Details.BlockHash[i]))
	}

	if commit.Status.Height > titleStatus.Height+1 {
		saveTitleHeight(a.db, calcTitleHeightKey(commit.Status.Title, commit.Status.Height), stat)
		//平行链由主链共识无缝切换，即接收第一个收到的高度，可以不从0开始
		if !(types.IsPara() && titleStatus.Height == -1) {
			return receipt, nil
		}
	}

	commitCount := len(stat.Details.Addrs)
	most, mostHash := getMostCommit(stat)
	if !isCommitDone(stat, nodes, most) {
		saveTitleHeight(a.db, calcTitleHeightKey(commit.Status.Title, commit.Status.Height), stat)
		return receipt, nil
	}
	//add commit done receipt
	receiptDone := makeDoneReceipt(a.fromaddr, commit, stat, int32(most), int32(commitCount), int32(len(nodes)))
	receipt = mergeReceipt(receipt, receiptDone)

	//平行连进行奖励分配，考虑可能的失败，需要在保存共识高度等数据之前处理
	if types.IsPara() {

		rewardReceipt, err := a.reward(commit.Status, stat)
		//错误会导致和主链处理的共识结果不一致
		if err != nil {
			clog.Error("paracross mining reward err", "height", titleStatus.Height,
				"blockhash", hex.EncodeToString(titleStatus.BlockHash), "err", err)
			return nil, err
		}
		receipt = mergeReceipt(receipt, rewardReceipt)
	}

	clog.Info("paracross.Commit commit ----pass", "most", most, "mostHash", hex.EncodeToString([]byte(mostHash)))

	stat.Status = pt.ParacrossStatusCommitDone
	saveTitleHeight(a.db, calcTitleHeightKey(commit.Status.Title, commit.Status.Height), stat)

	titleStatus.Title = commit.Status.Title
	titleStatus.Height = commit.Status.Height
	titleStatus.BlockHash = commit.Status.BlockHash
	saveTitle(a.db, calcTitleKey(commit.Status.Title), titleStatus)

	clog.Info("paracross.Commit commit done", "height", commit.Status.Height,
		"cross tx bitmap", string(commit.Status.CrossTxResult), "statusBlockHash", hex.EncodeToString(titleStatus.BlockHash))

	//parallel chain not need to process cross commit tx here
	if types.IsPara() {
		return receipt, nil
	}

	haveCrossTxs := len(commit.Status.CrossTxHashs) > 0
	if commit.Status.Height > 0 && types.IsDappFork(commit.Status.MainBlockHeight, pt.ParaX, pt.ForkCommitTx) && commit.Status.CrossTxHashs[0] == nil {
		haveCrossTxs = false
	}

	if enableParacrossTransfer && commit.Status.Height > 0 && haveCrossTxs {
		clog.Debug("paracross.Commit commitDone", "do cross", "")
		crossTxReceipt, err := a.execCrossTxs(commit)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, crossTxReceipt.KV...)
		receipt.Logs = append(receipt.Logs, crossTxReceipt.Logs...)
	}
	return receipt, nil
}

func (a *action) execCrossTx(tx *types.TransactionDetail, commit *pt.ParacrossCommitAction, crossTxHash []byte) (*types.Receipt, error) {
	if !bytes.HasSuffix(tx.Tx.Execer, []byte(pt.ParaX)) {
		return nil, nil
	}
	var payload pt.ParacrossAction
	err := types.Decode(tx.Tx.Payload, &payload)
	if err != nil {
		clog.Crit("paracross.Commit Decode Tx failed", "para title", commit.Status.Title,
			"para height", commit.Status.Height, "error", err, "txHash", hex.EncodeToString(crossTxHash))
		return nil, err
	}

	if payload.Ty == pt.ParacrossActionAssetWithdraw {
		receiptWithdraw, err := a.assetWithdraw(payload.GetAssetWithdraw(), tx.Tx)
		if err != nil {
			clog.Crit("paracross.Commit Decode Tx failed", "para title", commit.Status.Title,
				"para height", commit.Status.Height, "error", err, "txHash", hex.EncodeToString(crossTxHash))
			return nil, errors.Cause(err)
		}

		clog.Info("paracross.Commit WithdrawCoins", "para title", commit.Status.Title,
			"para height", commit.Status.Height, "error", err, "txHash", hex.EncodeToString(crossTxHash))
		return receiptWithdraw, nil
	} //else if tx.ActionName == pt.ParacrossActionAssetTransferStr {
	return nil, nil
	//}
}

func getCrossTxHashs(api client.QueueProtocolAPI, commit *pt.ParacrossCommitAction) ([][]byte, []byte, error) {
	if types.IsDappFork(commit.Status.MainBlockHeight, pt.ParaX, pt.ForkCommitTx) {
		if len(commit.Status.CrossTxHashs) == 0 {
			clog.Error("getCrossTxHashs len=0", "paraHeight", commit.Status.Height,
				"mainHeight", commit.Status.MainBlockHeight, "mainHash", hex.EncodeToString(commit.Status.MainBlockHash))
			return nil, nil, types.ErrCheckTxHash
		}
		blockDetail, err := GetBlock(api, commit.Status.MainBlockHash)
		if err != nil {
			return nil, nil, err
		}
		//校验
		paraBaseTxs := FilterTxsForPara(commit.Status.Title, blockDetail)
		paraCrossHashs := FilterParaCrossTxHashes(commit.Status.Title, paraBaseTxs)
		var baseHashs [][]byte
		for _, tx := range paraBaseTxs {
			baseHashs = append(baseHashs, tx.Hash())
		}
		baseCheckTxHash := CalcTxHashsHash(baseHashs)
		crossCheckHash := CalcTxHashsHash(paraCrossHashs)
		if !bytes.Equal(commit.Status.CrossTxHashs[0], crossCheckHash) {
			clog.Error("getCrossTxHashs para hash not equal", "paraHeight", commit.Status.Height,
				"mainHeight", commit.Status.MainBlockHeight, "mainHash", hex.EncodeToString(commit.Status.MainBlockHash),
				"main.crossHash", hex.EncodeToString(crossCheckHash),
				"commit.crossHash", hex.EncodeToString(commit.Status.CrossTxHashs[0]),
				"main.baseHash", hex.EncodeToString(baseCheckTxHash), "commit.baseHash", hex.EncodeToString(commit.Status.TxHashs[0]))
			for _, hash := range baseHashs {
				clog.Error("getCrossTxHashs base tx hash", "txhash", hex.EncodeToString(hash))
			}
			for _, hash := range paraCrossHashs {
				clog.Error("getCrossTxHashs paracross tx hash", "txhash", hex.EncodeToString(hash))
			}
			return nil, nil, types.ErrCheckTxHash
		}

		//只获取跨链tx
		rst, err := hex.DecodeString(string(commit.Status.CrossTxResult))
		if err != nil {
			clog.Error("getCrossTxHashs decode string", "CrossTxResult", string(commit.Status.CrossTxResult),
				"commit.height", commit.Status.Height)
			return nil, nil, types.ErrInvalidParam
		}

		return paraCrossHashs, rst, nil
	}

	return commit.Status.CrossTxHashs, commit.Status.CrossTxResult, nil
}

func (a *action) execCrossTxs(commit *pt.ParacrossCommitAction) (*types.Receipt, error) {
	var receipt types.Receipt

	crossTxHashs, crossTxResult, err := getCrossTxHashs(a.api, commit)
	if err != nil {
		clog.Error("paracross.Commit getCrossTxHashs", "err", err.Error())
		return nil, err
	}
	if len(crossTxHashs) == 0 {
		return &receipt, nil
	}

	for i := 0; i < len(crossTxHashs); i++ {
		clog.Info("paracross.Commit commitDone", "do cross number", i, "hash",
			hex.EncodeToString(crossTxHashs[i]),
			"res", util.BitMapBit(crossTxResult, uint32(i)))
		if util.BitMapBit(crossTxResult, uint32(i)) {
			tx, err := GetTx(a.api, crossTxHashs[i])
			if err != nil {
				clog.Crit("paracross.Commit Load Tx failed", "para title", commit.Status.Title,
					"para height", commit.Status.Height, "para tx index", i, "error", err, "txHash",
					hex.EncodeToString(crossTxHashs[i]))
				return nil, err
			}
			if tx == nil {
				clog.Error("paracross.Commit Load Tx failed", "para title", commit.Status.Title,
					"para height", commit.Status.Height, "para tx index", i, "error", err, "txHash",
					hex.EncodeToString(crossTxHashs[i]))
				return nil, types.ErrHashNotExist
			}
			receiptCross, err := a.execCrossTx(tx, commit, crossTxHashs[i])
			if err != nil {
				return nil, errors.Cause(err)
			}
			if receiptCross == nil {
				continue
			}
			receipt.KV = append(receipt.KV, receiptCross.KV...)
			receipt.Logs = append(receipt.Logs, receiptCross.Logs...)
		} else {
			clog.Error("paracross.Commit commitDone", "do cross number", i, "hash",
				hex.EncodeToString(crossTxHashs[i]),
				"para res", util.BitMapBit(crossTxResult, uint32(i)))
		}
	}

	return &receipt, nil
}

func (a *action) AssetTransfer(transfer *types.AssetsTransfer) (*types.Receipt, error) {
	clog.Debug("Paracross.Exec", "AssetTransfer", transfer.Cointoken, "transfer", "")
	receipt, err := a.assetTransfer(transfer)
	if err != nil {
		clog.Error("AssetTransfer failed", "err", err)
		return nil, errors.Cause(err)
	}
	return receipt, nil
}

func (a *action) AssetWithdraw(withdraw *types.AssetsWithdraw) (*types.Receipt, error) {
	//分叉高度之后，支持从平行链提取资产
	if !types.IsDappFork(a.height, pt.ParaX, "ForkParacrossWithdrawFromParachain") {
		if withdraw.Cointoken != "" {
			return nil, types.ErrNotSupport
		}
	}

	isPara := types.IsPara()
	if !isPara {
		// 需要平行链先执行， 达成共识时，继续执行
		return nil, nil
	}
	clog.Debug("paracross.AssetWithdraw isPara", "execer", string(a.tx.Execer),
		"txHash", hex.EncodeToString(a.tx.Hash()), "token name", withdraw.Cointoken)
	receipt, err := a.assetWithdraw(withdraw, a.tx)
	if err != nil {
		clog.Error("AssetWithdraw failed", "err", err)
		return nil, errors.Cause(err)
	}
	return receipt, nil
}

//当前miner tx不需要校验上一个区块的衔接性，因为tx就是本节点发出，高度，preHash等都在本区块里面的blockchain做了校验
func (a *action) Miner(miner *pt.ParacrossMinerAction) (*types.Receipt, error) {
	if miner.Status.Title != types.GetTitle() || miner.Status.PreBlockHash == nil || miner.Status.MainBlockHash == nil {
		return nil, pt.ErrParaMinerExecErr
	}

	var logs []*types.ReceiptLog
	var receipt = &pt.ReceiptParacrossMiner{}

	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogParacrossMiner
	receipt.Status = miner.Status

	log.Log = types.Encode(receipt)
	logs = append(logs, log)

	minerReceipt := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: logs}

	//自共识后才挖矿
	if miner.IsSelfConsensus {
		//增发coins到paracross合约中，只处理发放，不做分配
		totalReward := int64(0)
		coinReward := types.MGInt("mver.consensus.coinReward", a.height)
		fundReward := types.MGInt("mver.consensus.coinDevFund", a.height)

		if coinReward > 0 {
			totalReward += coinReward
		}
		if fundReward > 0 {
			totalReward += fundReward
		}
		totalReward *= types.Coin

		if totalReward > 0 {
			issueReceipt, err := a.coinsAccount.ExecIssueCoins(a.execaddr, totalReward)

			if err != nil {
				clog.Error("paracross miner issue err", "height", miner.Status.Height,
					"execAddr", a.execaddr, "amount", totalReward/types.Coin)
				return nil, err
			}
			minerReceipt = mergeReceipt(minerReceipt, issueReceipt)
		}
	}

	return minerReceipt, nil
}

func getTitleFrom(exec []byte) ([]byte, error) {
	last := bytes.LastIndex(exec, []byte("."))
	if last == -1 {
		return nil, types.ErrNotFound
	}
	// 现在配置是包含 .的， 所有取title 是也把 `.` 取出来
	return exec[:last+1], nil
}

/*
func (a *Paracross) CrossLimits(tx *types.Transaction, index int) bool {
	if tx.GroupCount < 2 {
		return true
	}

	txs, err := a.GetTxGroup(index)
	if err != nil {
		clog.Error("crossLimits", "get tx group failed", err, "hash", hex.EncodeToString(tx.Hash()))
		return false
	}

	titles := make(map[string] struct{})
	for _, txTmp := range txs {
		title, err := getTitleFrom(txTmp.Execer)
		if err == nil {
			titles[string(title)] = struct{}{}
		}
	}
	return len(titles) <= 1
}
*/

func (a *action) Transfer(transfer *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	clog.Debug("Paracross.Exec Transfer", "symbol", transfer.Cointoken, "amount",
		transfer.Amount, "to", tx.To)
	from := tx.From()

	acc, err := account.NewAccountDB(pt.ParaX, transfer.Cointoken, a.db)
	if err != nil {
		clog.Error("Transfer failed", "err", err)
		return nil, err
	}
	//to 是 execs 合约地址
	if dapp.IsDriverAddress(tx.GetRealToAddr(), a.height) {
		return acc.TransferToExec(from, tx.GetRealToAddr(), transfer.Amount)
	}
	return acc.Transfer(from, tx.GetRealToAddr(), transfer.Amount)
}

func (a *action) Withdraw(withdraw *types.AssetsWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	clog.Debug("Paracross.Exec Withdraw", "symbol", withdraw.Cointoken, "amount",
		withdraw.Amount, "to", tx.To)
	acc, err := account.NewAccountDB(pt.ParaX, withdraw.Cointoken, a.db)
	if err != nil {
		clog.Error("Withdraw failed", "err", err)
		return nil, err
	}
	if dapp.IsDriverAddress(tx.GetRealToAddr(), a.height) || dapp.ExecAddress(withdraw.ExecName) == tx.GetRealToAddr() {
		return acc.TransferWithdraw(tx.From(), tx.GetRealToAddr(), withdraw.Amount)
	}
	return nil, types.ErrActionNotSupport
}

func (a *action) TransferToExec(transfer *types.AssetsTransferToExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	clog.Debug("Paracross.Exec TransferToExec", "symbol", transfer.Cointoken, "amount",
		transfer.Amount, "to", tx.To)
	from := tx.From()

	acc, err := account.NewAccountDB(pt.ParaX, transfer.Cointoken, a.db)
	if err != nil {
		clog.Error("TransferToExec failed", "err", err)
		return nil, err
	}
	//to 是 execs 合约地址
	if dapp.IsDriverAddress(tx.GetRealToAddr(), a.height) || dapp.ExecAddress(transfer.ExecName) == tx.GetRealToAddr() {
		return acc.TransferToExec(from, tx.GetRealToAddr(), transfer.Amount)
	}
	return nil, types.ErrActionNotSupport
}
