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

func isCommitDone(nodes map[string]struct{}, mostSame int) bool {
	return float32(mostSame) > float32(len(nodes))*float32(2)/float32(3)
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

func makeDoneReceipt(execMainHeight int64, commit *pt.ParacrossNodeStatus,
						most, commitCount, totalCount int32) *types.Receipt {

	log := &pt.ReceiptParacrossDone{
		TotalNodes:     totalCount,
		TotalCommit:    commitCount,
		MostSameCommit: most,
		Title:          commit.Title,
		Height:         commit.Height,
		StateHash:      commit.StateHash,
		BlockHash:      commit.BlockHash,
		TxCounts:       commit.TxCounts,
		TxResult:       commit.TxResult,
		TxHashs:        commit.TxHashs,
		CrossTxHashs:   commit.CrossTxHashs,
		CrossTxResult:  commit.CrossTxResult,
		MainBlockHeight: commit.MainBlockHeight,
		MainBlockHash:   commit.MainBlockHash,
	}
	key := calcTitleKey(commit.Title)
	status := &pt.ParacrossStatus{
		Title:     commit.Title,
		Height:    commit.Height,
		BlockHash: commit.BlockHash,
	}
	if execMainHeight >= getDappForkHeight(pt.ForkLoopCheckCommitTxDone) {
		status.MainHeight = commit.MainBlockHeight
		status.MainHash = commit.MainBlockHash
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(status)},
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

//需要在ForkLoopCheckCommitTxDone后使用
func getMostResults(mostHash []byte, stat *pt.ParacrossHeightStatus) ([]byte,[]byte,[]byte,[]byte) {
	for i,hash := range stat.Details.BlockHash{
		if bytes.Equal(mostHash,hash){
			return stat.Details.TxResult[i],stat.Details.TxHashs[i],stat.Details.CrossTxResult[i],stat.Details.CrossTxHashs[i]
		}
	}
	return nil,nil,nil,nil
}


func hasCommited(addrs []string, addr string) (bool, int) {
	for i, a := range addrs {
		if a == addr {
			return true, i
		}
	}
	return false, 0
}

func getDappForkHeight(forkKey string) int64 {
	var forkHeight int64
	if types.IsPara() {
		key := forkKey
		switch forkKey {
		case pt.ForkCommitTx:
			key = pt.MainForkParacrossCommitTx
		case pt.ForkLoopCheckCommitTxDone:
			key = pt.MainLoopCheckCommitTxDoneForkHeight
		}

		forkHeight = types.Conf("config.consensus.sub.para").GInt(key)
		if forkHeight <= 0 {
			forkHeight = types.MaxHeight
		}
	} else {
		forkHeight = types.GetDappFork(pt.ParaX, forkKey)
	}
	return forkHeight
}

func getConfigNodes(db dbm.KV, title string) (map[string]struct{}, []byte, error) {
	key := calcParaNodeGroupAddrsKey(title)
	nodes, _, err := getNodes(db, key)
	if err != nil {
		if errors.Cause(err) != pt.ErrTitleNotExist {
			return nil, nil, errors.Wrapf(err, "getNodes para for title:%s", title)
		}
		key = calcManageConfigNodesKey(title)
		nodes, _, err = getNodes(db, key)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "getNodes manager for title:%s", title)
		}
	}

	return nodes, key, nil
}

func (a *action) getNodesGroup(title string) (map[string]struct{}, error) {
	if a.exec.GetMainHeight() < getDappForkHeight(pt.ForkCommitTx) {
		nodes, _, err := getConfigManageNodes(a.db, title)
		if err != nil {
			return nil, errors.Wrapf(err, "getNodes for title:%s", title)
		}
		return nodes, nil
	}

	nodes, _, err := getConfigNodes(a.db, title)
	return nodes, err

}

//根据nodes过滤掉可能退出了的addrs
func updateCommitAddrs(stat *pt.ParacrossHeightStatus, nodes map[string]struct{}, execMainHeight int64) {
	details := &pt.ParacrossStatusDetails{}
	for i, addr := range stat.Details.Addrs {
		if _, ok := nodes[addr]; ok {
			details.Addrs = append(details.Addrs, addr)
			details.BlockHash = append(details.BlockHash, stat.Details.BlockHash[i])
			if execMainHeight >= getDappForkHeight(pt.ForkLoopCheckCommitTxDone) {
				details.TxResult = append(details.TxResult, stat.Details.TxResult[i])
				details.TxHashs = append(details.TxHashs, stat.Details.TxHashs[i])
				details.CrossTxResult = append(details.CrossTxResult, stat.Details.CrossTxResult[i])
				details.CrossTxHashs = append(details.CrossTxHashs, stat.Details.CrossTxHashs[i])
			}

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
		if a.exec.GetMainHeight() >= getDappForkHeight(pt.ForkLoopCheckCommitTxDone){
			stat.Details.TxResult = append(stat.Details.TxResult, commit.Status.TxResult)
			stat.Details.TxHashs = append(stat.Details.TxHashs, commit.Status.TxHashs[0])
			stat.Details.CrossTxResult = append(stat.Details.CrossTxResult, commit.Status.CrossTxResult)
			stat.Details.CrossTxHashs = append(stat.Details.CrossTxHashs, commit.Status.CrossTxHashs[0])
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
			if a.exec.GetMainHeight() >= getDappForkHeight(pt.ForkLoopCheckCommitTxDone){
				stat.Details.TxResult[index] = commit.Status.TxResult
				stat.Details.TxHashs[index] = commit.Status.TxHashs[0]
				stat.Details.CrossTxResult[index] = commit.Status.CrossTxResult
				stat.Details.CrossTxHashs[index] = commit.Status.CrossTxHashs[0]
			}
		} else {
			stat.Details.Addrs = append(stat.Details.Addrs, a.fromaddr)
			stat.Details.BlockHash = append(stat.Details.BlockHash, commit.Status.BlockHash)
			if a.exec.GetMainHeight() >= getDappForkHeight(pt.ForkLoopCheckCommitTxDone){
				stat.Details.TxResult = append(stat.Details.TxResult, commit.Status.TxResult)
				stat.Details.TxHashs = append(stat.Details.TxHashs, commit.Status.TxHashs[0])
				stat.Details.CrossTxResult = append(stat.Details.CrossTxResult, commit.Status.CrossTxResult)
				stat.Details.CrossTxHashs = append(stat.Details.CrossTxHashs, commit.Status.CrossTxHashs[0])
			}
		}

		receipt = makeCommitReceipt(a.fromaddr, commit, &copyStat, stat)
	}
	//平行链fork pt.ForkCommitTx=0,主链在ForkCommitTx后支持nodegroup，这里平行链dappFork一定为true
	if types.IsDappFork(commit.Status.MainBlockHeight, pt.ParaX, pt.ForkCommitTx) {
		updateCommitAddrs(stat, nodes,a.exec.GetMainHeight())
	}

	if commit.Status.Height > titleStatus.Height+1 {
		saveTitleHeight(a.db, calcTitleHeightKey(commit.Status.Title, commit.Status.Height), stat)
		//平行链由主链共识无缝切换，即接收第一个收到的高度，可以不从0开始
		paraSwitch, err := isParaSelfConsensSwitch(a.db,stat, titleStatus)
		if err != nil {
			return nil, err
		}
		if !paraSwitch {
			return receipt, nil
		}
	}
	return a.commitTxDone(commit.Status,stat,titleStatus,nodes,receipt)
}

func (a *action)commitTxDone(nodeStatus *pt.ParacrossNodeStatus, stat *pt.ParacrossHeightStatus,titleStatus *pt.ParacrossStatus,
								nodes map[string]struct{},receipt *types.Receipt) (*types.Receipt, error){
	clog.Info("paracross.Commit commit", "stat.title", stat.Title, "stat.height", stat.Height, "notes", len(nodes))
	for i, v := range stat.Details.Addrs {
		clog.Info("paracross.Commit commit detail", "addr", v, "hash", hex.EncodeToString(stat.Details.BlockHash[i]))
	}

	commitCount := len(stat.Details.Addrs)
	most, mostHash := getMostCommit(stat)
	if !isCommitDone(nodes, most) {
		saveTitleHeight(a.db, calcTitleHeightKey(stat.Title, stat.Height), stat)
		return receipt, nil
	}
	clog.Info("paracross.Commit commit ----pass", "most", most, "mostHash", hex.EncodeToString([]byte(mostHash)))
	stat.Status = pt.ParacrossStatusCommitDone
	saveTitleHeight(a.db, calcTitleHeightKey(stat.Title, stat.Height), stat)

	//add commit done receipt
	receiptDone := makeDoneReceipt(a.exec.GetMainHeight(), nodeStatus, int32(most), int32(commitCount), int32(len(nodes)))
	receipt = mergeReceipt(receipt, receiptDone)

	return a.commitTxDoneStep2(nodeStatus,stat,titleStatus,receipt)
}


func (a *action)commitTxDoneStep2(nodeStatus *pt.ParacrossNodeStatus, stat *pt.ParacrossHeightStatus,titleStatus *pt.ParacrossStatus,
									receipt *types.Receipt) (*types.Receipt, error){

	titleStatus.Title = nodeStatus.Title
	titleStatus.Height = nodeStatus.Height
	titleStatus.BlockHash = nodeStatus.BlockHash
	if a.exec.GetMainHeight() >= getDappForkHeight(pt.ForkLoopCheckCommitTxDone) {
		titleStatus.MainHeight = nodeStatus.MainBlockHeight
		titleStatus.MainHash = nodeStatus.MainBlockHash
	}
	saveTitle(a.db, calcTitleKey(titleStatus.Title), titleStatus)

	clog.Info("paracross.Commit commit done", "height", nodeStatus.Height, "statusBlockHash", hex.EncodeToString(nodeStatus.BlockHash))

	//parallel chain not need to process cross commit tx here
	if types.IsPara() {
		//平行连进行奖励分配
		rewardReceipt, err := a.reward(nodeStatus,stat)
		//错误会导致和主链处理的共识结果不一致
		if err != nil {
			clog.Error("paracross mining reward err", "height", nodeStatus.Height,
				"blockhash", hex.EncodeToString(nodeStatus.BlockHash), "err", err)
			return nil, err
		}
		receipt = mergeReceipt(receipt, rewardReceipt)
		return receipt, nil
	}

	//主链，处理跨链交易
	r,err := a.procCrossTxs(nodeStatus)
	if err != nil{
		return nil,err
	}
	receipt = mergeReceipt(receipt,r)
	return receipt, nil
}


func (a *action) procCrossTxs(status *pt.ParacrossNodeStatus)(*types.Receipt, error){
	haveCrossTxs := len(status.CrossTxHashs) > 0
	if status.Height > 0 && types.IsDappFork(status.MainBlockHeight, pt.ParaX, pt.ForkCommitTx) && len(status.CrossTxHashs[0]) == 0 {
		haveCrossTxs = false
	}

	if enableParacrossTransfer && status.Height > 0 && haveCrossTxs {
		clog.Debug("paracross.Commit commitDone", "do cross", "")
		crossTxReceipt, err := a.execCrossTxs(status)
		if err != nil {
			return nil, err
		}
		return crossTxReceipt,nil
	}
	return nil, nil
}


//由于可能对当前块的共识交易进行处理，需要全部数据保存到statedb，通过tx获取数据无法处理当前块的场景
func (a *action) loopCommitTxDone(title string)(*types.Receipt, error){
	receipt := &types.Receipt{}

	nodes, _, err := getParacrossNodes(a.db, title)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", title)
	}
	//从当前共识高度开始遍历
	titleStatus, err := getTitle(a.db, calcTitleKey(title))
	if err != nil {
		return nil, errors.Wrapf(err, "getTitle:%s", title)
	}
	//当前共识高度还未到分叉高度，则不处理
	if titleStatus.GetMainHeight() < getDappForkHeight(pt.ForkLoopCheckCommitTxDone){
		return nil,errors.Wrapf(pt.ErrForkHeightNotReach,
			"titleHeight:%d,forkHeight", titleStatus.MainHeight,getDappForkHeight(pt.ForkLoopCheckCommitTxDone))
	}

	loopHeight := titleStatus.Height

	for{
		loopHeight++

		stat, err := getTitleHeight(a.db, calcTitleHeightKey(title, loopHeight))
		if err != nil {
			clog.Error("paracross.Commit getTitleHeight failed", "err", err)
			return receipt, err
		}
		//防止无限循环
		if stat.MainHeight > a.exec.GetMainHeight(){
			return receipt,nil
		}

		r,err := a.checkCommitTxDone(title,stat,nodes)
		if err != nil{
			clog.Error("paracross.Commit checkExecCommitTxDone", "para title", title,	"height", stat.Height,"error", err, )
			return receipt,nil
		}
		if r == nil{
			return receipt,nil
		}
		receipt = mergeReceipt(receipt,r)

	}

	return receipt,nil

}


func (a *action) checkCommitTxDone(title string,stat *pt.ParacrossHeightStatus,nodes map[string]struct{})(*types.Receipt, error){
	receipt := &types.Receipt{}
	status, err := getTitle(a.db, calcTitleKey(title))
	if err != nil {
		return nil, errors.Wrapf(err, "getTitle:%s", a.fromaddr)
	}

	//如果是平行链自共识首次切换，可以在正常流程里面再触发
	if stat.Height > status.Height+1 {
		return nil, nil
	}

	updateCommitAddrs(stat, nodes,a.exec.GetMainHeight())
	most, _ := getMostCommit(stat)
	if !isCommitDone(nodes, most) {
		return nil, nil
	}

	return a.commitTxDoneByStat(stat,status,nodes,receipt)

}


//只根据stat的信息在commitDone之后重构一个commitStatus做后续处理
func (a *action) commitTxDoneByStat(stat *pt.ParacrossHeightStatus,titleStatus *pt.ParacrossStatus,
	nodes map[string]struct{},receipt *types.Receipt) (*types.Receipt, error){

	clog.Info("paracross.Commit commit", "stat.title", stat.Title, "stat.height", stat.Height, "notes", len(nodes))
	for i, v := range stat.Details.Addrs {
		clog.Info("paracross.Commit commit detail", "addr", v, "hash", hex.EncodeToString(stat.Details.BlockHash[i]))
	}

	commitCount := len(stat.Details.Addrs)
	most, mostHash := getMostCommit(stat)
	if !isCommitDone(nodes, most) {
		saveTitleHeight(a.db, calcTitleHeightKey(stat.Title, stat.Height), stat)
		return receipt, nil
	}
	clog.Info("paracross.Commit commit ----pass", "most", most, "mostHash", hex.EncodeToString([]byte(mostHash)))
	stat.Status = pt.ParacrossStatusCommitDone
	saveTitleHeight(a.db, calcTitleHeightKey(stat.Title, stat.Height), stat)

	txRst,txHash,crossTxRst,crossTxHash := getMostResults([]byte(mostHash),stat)
	finalStatus := &pt.ParacrossNodeStatus{
		MainBlockHash: 		stat.MainHash,
		MainBlockHeight: 	stat.MainHeight,
		Title: 				stat.Title,
		Height: 			stat.Height,
		BlockHash: 			[]byte(mostHash),
		TxResult: 			txRst,
		TxHashs: 			[][]byte{txHash},
		CrossTxResult: 		crossTxRst,
		CrossTxHashs: 		[][]byte{crossTxHash},
	}
	//stat.ConsensBlockHash = []byte(mostHash)
	//add commit done receipt
	receiptDone := makeDoneReceipt(a.exec.GetMainHeight(), finalStatus,int32(most), int32(commitCount), int32(len(nodes)))
	receipt = mergeReceipt(receipt, receiptDone)
	clog.Info("paracross.Commit commit ----pass", "most", most, "mostHash", hex.EncodeToString([]byte(mostHash)))

	return a.commitTxDoneStep2(finalStatus,stat,titleStatus,receipt)
}

//平行链自共识无缝切换条件：1，平行链没有共识过，2：commit高度是大于自共识分叉高度且上一次共识的主链高度小于自共识分叉高度，保证只运行一次，
// 这样在主链没有共识空洞前提下，平行链允许有条件的共识跳跃
func (a *action) isParaSelfConsensSwitch(stat *pt.ParacrossHeightStatus, titleStatus *pt.ParacrossStatus) (bool, error) {
	if !types.IsPara() {
		return false, nil
	}

	if titleStatus.Height == -1 {
		return true, nil
	}

	selfConsensForkHeight := getDappForkHeight(pt.ParaSelfConsensForkHeight)
	lastStatusMainHeight := int64(-1)
	if titleStatus.Height > -1 {
		s, err := getTitleHeight(a.db, calcTitleHeightKey(stat.Title, titleStatus.Height))
		if err != nil {
			clog.Error("paracross.Commit isParaSelfConsensSwitch getTitleHeight failed", "err", err.Error())
			return false, err
		}
		lastStatusMainHeight = s.MainHeight
	}

	return stat.MainHeight > selfConsensForkHeight && lastStatusMainHeight < selfConsensForkHeight, nil

}

func (a *action) execCrossTx(tx *types.TransactionDetail, crossTxHash []byte) (*types.Receipt, error) {
	if !bytes.HasSuffix(tx.Tx.Execer, []byte(pt.ParaX)) {
		return nil, nil
	}
	var payload pt.ParacrossAction
	err := types.Decode(tx.Tx.Payload, &payload)
	if err != nil {
		clog.Crit("paracross.Commit Decode Tx failed", "error", err, "txHash", hex.EncodeToString(crossTxHash))
		return nil, err
	}

	if payload.Ty == pt.ParacrossActionAssetWithdraw {
		receiptWithdraw, err := a.assetWithdraw(payload.GetAssetWithdraw(), tx.Tx)
		if err != nil {
			clog.Crit("paracross.Commit Decode Tx failed", "error", err, "txHash", hex.EncodeToString(crossTxHash))
			return nil, errors.Cause(err)
		}

		clog.Info("paracross.Commit WithdrawCoins",  "txHash", hex.EncodeToString(crossTxHash))
		return receiptWithdraw, nil
	}
	return nil, nil

}


func getCrossTxHashs(api client.QueueProtocolAPI, status *pt.ParacrossNodeStatus) ([][]byte, []byte, error) {
	if !types.IsDappFork(status.MainBlockHeight, pt.ParaX, pt.ForkCommitTx){
		return status.CrossTxHashs, status.CrossTxResult, nil
	}

	if len(status.CrossTxHashs) == 0 {
		clog.Error("getCrossTxHashs len=0", "paraHeight", status.Height,
			"mainHeight", status.MainBlockHeight, "mainHash", hex.EncodeToString(status.MainBlockHash))
		return nil, nil, types.ErrCheckTxHash
	}


	blockDetail, err := GetBlock(api, status.MainBlockHash)
	if err != nil {
		return nil, nil, err
	}
	//校验
	paraBaseTxs := FilterTxsForPara(status.Title, blockDetail)
	paraCrossHashs := FilterParaCrossTxHashes(status.Title, paraBaseTxs)
	var baseHashs [][]byte
	for _, tx := range paraBaseTxs {
		baseHashs = append(baseHashs, tx.Hash())
	}
	baseCheckTxHash := CalcTxHashsHash(baseHashs)
	crossCheckHash := CalcTxHashsHash(paraCrossHashs)
	if !bytes.Equal(status.CrossTxHashs[0], crossCheckHash) {
		clog.Error("getCrossTxHashs para hash not equal", "paraHeight", status.Height,
			"mainHeight", status.MainBlockHeight, "mainHash", hex.EncodeToString(status.MainBlockHash),
			"main.crossHash", hex.EncodeToString(crossCheckHash),
			"commit.crossHash", hex.EncodeToString(status.CrossTxHashs[0]),
			"main.baseHash", hex.EncodeToString(baseCheckTxHash), "commit.baseHash", hex.EncodeToString(status.TxHashs[0]))
		for _, hash := range baseHashs {
			clog.Error("getCrossTxHashs base tx hash", "txhash", hex.EncodeToString(hash))
		}
		for _, hash := range paraCrossHashs {
			clog.Error("getCrossTxHashs paracross tx hash", "txhash", hex.EncodeToString(hash))
		}
		return nil, nil, types.ErrCheckTxHash
	}

	//只获取跨链tx
	rst, err := hex.DecodeString(string(status.CrossTxResult))
	if err != nil {
		clog.Error("getCrossTxHashs decode string", "CrossTxResult", string(status.CrossTxResult),
			"commit.height", status.Height)
		return nil, nil, types.ErrInvalidParam
	}

	return paraCrossHashs, rst, nil

}


func (a *action) execCrossTxs(status *pt.ParacrossNodeStatus) (*types.Receipt, error) {
	var receipt types.Receipt

	crossTxHashs, crossTxResult, err := getCrossTxHashs(a.api, status)
	if err != nil {
		clog.Error("paracross.Commit getCrossTxHashs", "err", err.Error())
		return nil, err
	}
	if len(crossTxHashs) == 0 {
		return &receipt, nil
	}

	for i := 0; i < len(crossTxHashs); i++ {
		clog.Info("paracross.Commit commitDone", "do cross number", i, "hash",hex.EncodeToString(crossTxHashs[i]),
			"res", util.BitMapBit(crossTxResult, uint32(i)))
		if util.BitMapBit(crossTxResult, uint32(i)) {
			tx, err := GetTx(a.api, crossTxHashs[i])
			if err != nil {
				clog.Crit("paracross.Commit Load Tx failed", "para title", title, "para height", status.Height,
					"para tx index", i, "error", err, "txHash",	hex.EncodeToString(crossTxHashs[i]))
				return nil, err
			}
			if tx == nil {
				clog.Error("paracross.Commit Load Tx failed", "para title", title,	"para height", status.Height,
					"para tx index", i, "error", err, "txHash",	hex.EncodeToString(crossTxHashs[i]))
				return nil, types.ErrHashNotExist
			}
			receiptCross, err := a.execCrossTx(tx, crossTxHashs[i])
			if err != nil {
				clog.Error("paracross.Commit execCrossTx", "para title", title,"para height", status.Height,
					"para tx index", i, "error", err)
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
