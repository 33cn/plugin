// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/golang/protobuf/proto"
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
		//clog.Info("getNodes", "get db key", string(key), "failed", err)
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

// manager 合约 分叉前写入 现在不用了
func getConfigManageNodes(db dbm.KV, title string) (map[string]struct{}, []string, error) {
	key := calcManageConfigNodesKey(title)
	return getNodes(db, key)
}

func getParacrossNodes(db dbm.KV, title string) (map[string]struct{}, []string, error) {
	key := calcParaNodeGroupAddrsKey(title)
	return getNodes(db, key)
}

func validTitle(cfg *types.Chain33Config, title string) bool {
	if cfg.IsPara() {
		return cfg.GetTitle() == title
	}
	return len(title) > 0
}

func validNode(addr string, nodes map[string]struct{}) bool {
	if _, exist := nodes[addr]; exist {
		return exist
	}
	return false
}

func checkCommitInfo(cfg *types.Chain33Config, commit *pt.ParacrossNodeStatus) error {
	if commit == nil {
		return types.ErrInvalidParam
	}
	clog.Debug("paracross.Commit check input", "height", commit.Height, "mainHeight", commit.MainBlockHeight,
		"mainHash", common.ToHex(commit.MainBlockHash), "blockHash", common.ToHex(commit.BlockHash))

	if commit.Height == 0 {
		if len(commit.Title) == 0 || len(commit.BlockHash) == 0 {
			clog.Error("checkCommitInfo invalid param", "title", commit.Title, "blochHash", common.ToHex(commit.BlockHash))
			return types.ErrInvalidParam
		}
		return nil
	}

	if !pt.IsParaForkHeight(cfg, commit.MainBlockHeight, pt.ForkLoopCheckCommitTxDone) {
		if len(commit.MainBlockHash) == 0 || len(commit.Title) == 0 || commit.Height < 0 ||
			len(commit.PreBlockHash) == 0 || len(commit.BlockHash) == 0 ||
			len(commit.StateHash) == 0 || len(commit.PreStateHash) == 0 {
			clog.Error("checkCommitInfo invalid param 2", "title", commit.Title, "blochHash", common.ToHex(commit.BlockHash))
			return types.ErrInvalidParam
		}
		return nil
	}

	if len(commit.MainBlockHash) == 0 || len(commit.BlockHash) == 0 ||
		commit.MainBlockHeight < 0 || commit.Height < 0 {
		clog.Error("checkCommitInfo invalid param check", "title", commit.Title, "blochHash", common.ToHex(commit.BlockHash), "height", commit.Height)
		return types.ErrInvalidParam
	}

	if !validTitle(cfg, commit.Title) {
		clog.Error("checkCommitInfo invalid title", "title", commit.Title)
		return pt.ErrInvalidTitle
	}
	return nil
}

//区块链中不能使用float类型
func isCommitDone(nodes, mostSame int) bool {
	return 3*mostSame > 2*nodes
}

func makeCommitReceipt(addr string, commit *pt.ParacrossNodeStatus, prev, current *pt.ParacrossHeightStatus) *types.Receipt {
	key := calcTitleHeightKey(commit.Title, commit.Height)
	log := &pt.ReceiptParacrossCommit{
		Addr:    addr,
		Status:  commit,
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

func makeCommitStatReceipt(current *pt.ParacrossHeightStatus) *types.Receipt {
	key := calcTitleHeightKey(current.Title, current.Height)
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: nil,
	}
}

func makeRecordReceipt(addr string, commit *pt.ParacrossNodeStatus) *types.Receipt {
	log := &pt.ReceiptParacrossRecord{
		Addr:   addr,
		Status: commit,
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

func makeDoneReceipt(cfg *types.Chain33Config, execMainHeight, execHeight int64, commit *pt.ParacrossNodeStatus,
	most, commitCount, totalCount, mostSupervisionCount, totalSupervisionCommit, totalSupervisionNodes int32) *types.Receipt {

	log := &pt.ReceiptParacrossDone{
		TotalNodes:             totalCount,
		TotalCommit:            commitCount,
		MostSameCommit:         most,
		Title:                  commit.Title,
		Height:                 commit.Height,
		BlockHash:              commit.BlockHash,
		TxResult:               commit.TxResult,
		MainBlockHeight:        commit.MainBlockHeight,
		MainBlockHash:          commit.MainBlockHash,
		ChainExecHeight:        execHeight,
		TotalSupervisionNodes:  totalSupervisionNodes,
		TotalSupervisionCommit: totalSupervisionCommit,
		MostSupervisionCommit:  mostSupervisionCount,
	}
	key := calcTitleKey(commit.Title)
	status := &pt.ParacrossStatus{
		Title:     commit.Title,
		Height:    commit.Height,
		BlockHash: commit.BlockHash,
	}
	if execMainHeight >= pt.GetDappForkHeight(cfg, pt.ForkLoopCheckCommitTxDone) {
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

//GetMostCommit ...
func GetMostCommit(commits [][]byte) (int, string) {
	stats := make(map[string]int)
	n := len(commits)
	for i := 0; i < n; i++ {
		if _, ok := stats[string(commits[i])]; ok {
			stats[string(commits[i])]++
		} else {
			stats[string(commits[i])] = 1
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
func getMostResults(mostHash []byte, stat *pt.ParacrossHeightStatus) []byte {
	for i, hash := range stat.BlockDetails.BlockHashs {
		if bytes.Equal(mostHash, hash) {
			return stat.BlockDetails.TxResults[i]
		}
	}
	return nil
}

func hasCommited(addrs []string, addr string) (bool, int) {
	for i, a := range addrs {
		if a == addr {
			return true, i
		}
	}
	return false, 0
}

func getConfigNodes(db dbm.KV, title string) (map[string]struct{}, []string, []byte, error) {
	key := calcParaNodeGroupAddrsKey(title)
	nodes, nodesArray, err := getNodes(db, key)
	if err != nil {
		if errors.Cause(err) != pt.ErrTitleNotExist {
			return nil, nil, nil, errors.Wrapf(err, "getNodes para for title:%s", title)
		}
		nodes, nodesArray, err = getConfigManageNodes(db, title)
		if err != nil {
			return nil, nil, nil, errors.Wrapf(err, "getNodes manager for title:%s", title)
		}
	}

	return nodes, nodesArray, key, nil
}

func (a *action) getNodesGroup(title string) (map[string]struct{}, []string, error) {
	cfg := a.api.GetConfig()
	// 如果高度是分叉前，获取老的Nodes
	if a.exec.GetMainHeight() < pt.GetDappForkHeight(cfg, pt.ForkCommitTx) {
		nodes, nodesArray, err := getConfigManageNodes(a.db, title)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "getNodes for title:%s", title)
		}
		return nodes, nodesArray, nil
	}

	nodes, nodesArray, _, err := getConfigNodes(a.db, title)
	return nodes, nodesArray, err
}

func getSupervisionNodeGroupAddrs(db dbm.KV, title string) (map[string]struct{}, []string, []byte, error) {
	key := calcParaSupervisionNodeGroupAddrsKey(title)
	nodes, nodesArray, err := getNodes(db, key)
	if err != nil {
		if errors.Cause(err) != pt.ErrTitleNotExist {
			return nil, nil, nil, errors.Wrapf(err, "getSupervisionNodeGroupAddrs para for title:%s", title)
		}
	}

	return nodes, nodesArray, key, nil
}

func (a *action) isValidSuperNode(addr string) error {
	cfg := a.api.GetConfig()
	nodes, _, err := getParacrossNodes(a.db, cfg.GetTitle())
	if err != nil {
		return errors.Wrapf(err, "getNodes for title:%s", cfg.GetTitle())
	}
	if !validNode(addr, nodes) {
		return errors.Wrapf(pt.ErrParaNodeAddrNotExisted, "invalid node=%s", addr)
	}
	return nil
}

//相同的BlockHash，只保留一份数据
func updateCommitBlockHashs(stat *pt.ParacrossHeightStatus, commit *pt.ParacrossNodeStatus) {
	if stat.BlockDetails == nil {
		stat.BlockDetails = &pt.ParacrossStatusBlockDetails{}
	}
	for _, blockHash := range stat.BlockDetails.BlockHashs {
		if bytes.Equal(blockHash, commit.BlockHash) {
			return
		}
	}
	stat.BlockDetails.BlockHashs = append(stat.BlockDetails.BlockHashs, commit.BlockHash)
	stat.BlockDetails.TxResults = append(stat.BlockDetails.TxResults, commit.TxResult)
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

func updateSupervisionDetailsCommitAddrs(stat *pt.ParacrossHeightStatus, nodes map[string]struct{}) {
	supervisionDetailsDetails := &pt.ParacrossStatusDetails{}
	for i, addr := range stat.SupervisionDetails.Addrs {
		if _, ok := nodes[addr]; ok {
			supervisionDetailsDetails.Addrs = append(supervisionDetailsDetails.Addrs, addr)
			supervisionDetailsDetails.BlockHash = append(supervisionDetailsDetails.BlockHash, stat.SupervisionDetails.BlockHash[i])
		}
	}
	stat.SupervisionDetails = supervisionDetailsDetails
}

//自共识分阶段使能，综合考虑挖矿奖励和共识分配奖励，判断是否自共识使能需要采用共识的高度，而不能采用当前区块高度a.height
//考虑自共识使能区块高度100，如果采用区块高度判断，则在100高度可能收到80~99的20条共识交易，这20条交易在100高度参与共识，则无奖励可分配，而且共识高度将是80而不是100
//采用共识高度commit.Status.Height判断，则严格执行了产生奖励和分配奖励，且共识高度从100开始
func paraCheckSelfConsOn(cfg *types.Chain33Config, db dbm.KV, commit *pt.ParacrossNodeStatus) (bool, *types.Receipt, error) {
	if !cfg.IsDappFork(commit.Height, pt.ParaX, pt.ForkParaSelfConsStages) {
		return true, nil, nil
	}

	//分叉之后，key不存在，自共识没配置也认为不支持自共识
	isSelfConsOn, err := isSelfConsOn(db, commit.Height)
	if err != nil && errors.Cause(err) != pt.ErrKeyNotExist {
		return false, nil, err
	}
	if !isSelfConsOn {
		clog.Debug("paracross.Commit self consens off", "height", commit.Height)
		return false, &types.Receipt{Ty: types.ExecOk}, nil
	}
	return true, nil, nil
}

func (a *action) preCheckCommitInfo(commit *pt.ParacrossNodeStatus, commitAddrs []string) error {
	cfg := a.api.GetConfig()
	err := checkCommitInfo(cfg, commit)
	if err != nil {
		return err
	}

	titleStatus, err := getTitle(a.db, calcTitleKey(commit.Title))
	if err != nil {
		return errors.Wrapf(err, "getTitle=%s", commit.Title)
	}
	if titleStatus.Height+1 == commit.Height && commit.Height > 0 && !pt.IsParaForkHeight(cfg, commit.MainBlockHeight, pt.ForkLoopCheckCommitTxDone) {
		if !bytes.Equal(titleStatus.BlockHash, commit.PreBlockHash) {
			clog.Error("paracross.Commit", "check PreBlockHash", common.ToHex(titleStatus.BlockHash),
				"commit tx", common.ToHex(commit.PreBlockHash), "commitheit", commit.Height, "from", commitAddrs)
			return pt.ErrParaBlockHashNoMatch
		}
	}

	// 极端情况， 分叉后在平行链生成了 commit交易 并发送之后， 但主链完成回滚，时间序如下
	// 主链   （1）Bn1        （3） rollback-Bn1   （4） commit-done in Bn2
	// 平行链         （2）commit                                 （5） 将得到一个错误的块
	// 所以有必要做这个检测
	var dbMainHash []byte
	if !cfg.IsPara() {
		blockHash, err := getBlockHash(a.api, commit.MainBlockHeight)
		if err != nil {
			clog.Error("paracross.Commit getBlockHash", "err", err, "commit tx height", commit.MainBlockHeight, "from", commitAddrs)
			return err
		}
		dbMainHash = blockHash.Hash

	} else {
		block, err := getBlockInfo(a.api, commit.Height)
		if err != nil {
			clog.Error("paracross.Commit getBlockInfo", "err", err, "height", commit.Height, "from", commitAddrs)
			return err
		}
		dbMainHash = block.MainHash
	}

	//对于主链，校验的是主链高度对应的blockhash是否和commit的一致
	//对于平行链， 校验的是commit信息的平行链height block对应的mainHash是否和本地相同高度对应的mainHash一致， 在主链hash一致的时候看平行链共识blockhash是否一致
	if !bytes.Equal(dbMainHash, commit.MainBlockHash) && commit.Height > 0 {
		clog.Error("paracross.Commit blockHash not match", "isMain", !cfg.IsPara(), "db", common.ToHex(dbMainHash),
			"commit", common.ToHex(commit.MainBlockHash), "commitHeight", commit.Height,
			"commitMainHeight", commit.MainBlockHeight, "from", commitAddrs)
		return types.ErrBlockHashNoMatch
	}

	return nil
}

func getValidAddrs(nodes map[string]struct{}, addrs []string) []string {
	var ret []string
	for _, addr := range addrs {
		if !validNode(addr, nodes) {
			clog.Error("paracross.Commit getValidAddrs not valid", "addr", addr)
			continue
		}
		ret = append(ret, addr)
	}
	return ret
}

//get secp256 addr's bls pubkey
func getAddrBlsPubKey(db dbm.KV, title, addr string) (string, error) {
	addrStat, err := getNodeAddr(db, title, addr)
	if err != nil {
		return "", errors.Wrapf(err, "nodeAddr:%s-%s get error", title, addr)
	}
	return addrStat.BlsPubKey, nil
}

//bls签名共识交易验证 大约平均耗时3ms (2~4ms)
func (a *action) procBlsSign(nodesArry []string, commit *pt.ParacrossCommitAction) ([]string, error) {
	signAddrs := util.GetAddrsByBitMap(nodesArry, commit.Bls.AddrsMap)
	var pubs []string
	for _, addr := range signAddrs {
		pub, err := getAddrBlsPubKey(a.db, commit.Status.Title, addr /*, commitNodeType*/)
		if err != nil {
			return nil, errors.Wrapf(err, "pubkey not exist to addr=%s", addr)
		}
		pubs = append(pubs, pub)
	}
	err := verifyBlsSign(a.exec.cryptoCli, pubs, commit)
	if err != nil {
		clog.Error("paracross.Commit bls sign verify", "addr", signAddrs, "nodes", nodesArry, "from", a.fromaddr)
		return nil, err
	}
	return signAddrs, nil
}

func verifyBlsSign(cryptoCli crypto.Crypto, pubs []string, commit *pt.ParacrossCommitAction) error {
	t1 := types.Now()
	//1. 获取addr对应的bls 公钥
	pubKeys := make([]crypto.PubKey, 0)
	for _, p := range pubs {
		k, err := common.FromHex(p)
		if err != nil {
			return errors.Wrapf(err, "pub FromHex=%s", p)
		}
		pub, err := cryptoCli.PubKeyFromBytes(k)
		if err != nil {
			return errors.Wrapf(err, "DeserializePublicKey=%s", p)
		}
		pubKeys = append(pubKeys, pub)

	}

	//2.　获取聚合的签名, deserial 300us
	sign, err := cryptoCli.SignatureFromBytes(commit.Bls.Sign)
	if err != nil {
		return errors.Wrapf(err, "DeserializeSignature,key=%s", common.ToHex(commit.Bls.Sign))
	}
	//3. 获取签名前原始msg
	msg := types.Encode(commit.Status)

	//4. verify 1ms, total 2ms
	agg, err := crypto.ToAggregate(cryptoCli)
	if err != nil {
		return errors.Wrap(err, "ToAggregate")
	}
	err = agg.VerifyAggregatedOne(pubKeys, msg, sign)
	if err != nil {
		clog.Error("paracross.Commit bls sign verify", "title", commit.Status.Title, "height", commit.Status.Height,
			"addrsMap", common.ToHex(commit.Bls.AddrsMap), "sign", common.ToHex(commit.Bls.Sign), "data", common.ToHex(msg))
		return pt.ErrBlsSignVerify
	}
	clog.Debug("paracross procBlsSign success", "title", commit.Status.Title, "height", commit.Status.Height, "time", types.Since(t1))
	return nil
}

func (a *action) getValidCommitAddrs(commit *pt.ParacrossCommitAction, nodesMap map[string]struct{}, nodesArry []string) ([]string, error) {
	//获取commitAddrs, bls sign 包含多个账户的聚合签名
	commitAddrs := []string{a.fromaddr}
	if commit.Bls != nil {
		addrs, err := a.procBlsSign(nodesArry, commit)
		if err != nil {
			return nil, errors.Wrap(err, "procBlsSign")
		}
		commitAddrs = addrs
	}

	validAddrs := getValidAddrs(nodesMap, commitAddrs)
	if len(validAddrs) <= 0 {
		return nil, errors.Wrapf(errors.New("getValidAddrs error"), "getValidAddrs nil commitAddrs=%s ", strings.Join(commitAddrs, ","))
	}

	return validAddrs, nil
}

//共识commit　msg　处理
func (a *action) Commit(commit *pt.ParacrossCommitAction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	//平行链侧，自共识未使能则不处理
	if cfg.IsPara() {
		isSelfConsOn, receipt, err := paraCheckSelfConsOn(cfg, a.db, commit.Status)
		if !isSelfConsOn {
			return receipt, errors.Wrap(err, "checkSelfConsOn")
		}
	}

	nodesMap, nodesArry, err := a.getNodesGroup(commit.Status.Title)
	if err != nil {
		return nil, errors.Wrap(err, "getNodesGroup")
	}

	var validAddrs, supervisionValidAddrs []string

	bIsCommitSuperNode := false
	bIsCommitSupervisionNode := false
	if _, exist := nodesMap[a.fromaddr]; exist {
		validAddrs, err = a.getValidCommitAddrs(commit, nodesMap, nodesArry)
		if err != nil {
			return nil, errors.Wrap(err, "getValidCommitAddrs")
		}

		bIsCommitSuperNode = true
	}

	// 获取监督节点的数据
	supervisionNodesMap, supervisionNodesArry, _, err := getSupervisionNodeGroupAddrs(a.db, commit.Status.Title)
	if err != nil && errors.Cause(err) != pt.ErrTitleNotExist {
		return nil, errors.Wrap(err, "getSupervisionNodeGroupAddrs")
	}

	if !bIsCommitSuperNode {
		if _, exist := supervisionNodesMap[a.fromaddr]; exist {
			supervisionValidAddrs, err = a.getValidCommitAddrs(commit, supervisionNodesMap, supervisionNodesArry)
			if err != nil {
				return nil, errors.Wrap(err, "getValidCommitAddrs")
			}
			bIsCommitSupervisionNode = true
		}
	}

	if !bIsCommitSuperNode && !bIsCommitSupervisionNode {
		return nil, errors.Wrapf(errors.New("from addr error"), "form addr %s not in SuperNodesGroup, not in SupervisionNodesGroup", a.fromaddr)
	}

	return a.proCommitMsg(commit.Status, nodesMap, validAddrs, supervisionNodesMap, supervisionValidAddrs)
}

func (a *action) proCommitMsg(commit *pt.ParacrossNodeStatus, nodes map[string]struct{}, commitAddrs []string, supervisionNodes map[string]struct{}, supervisionValidAddrs []string) (*types.Receipt, error) {
	cfg := a.api.GetConfig()

	var err error
	if len(commitAddrs) > 0 {
		err = a.preCheckCommitInfo(commit, commitAddrs)
	} else {
		err = a.preCheckCommitInfo(commit, supervisionValidAddrs)
	}
	if err != nil {
		return nil, err
	}

	titleStatus, err := getTitle(a.db, calcTitleKey(commit.Title))
	if err != nil {
		return nil, errors.Wrapf(err, "getTitle:%s", commit.Title)
	}

	// 在完成共识之后来的， 增加 record log， 只记录不修改已经达成的共识
	if commit.Height <= titleStatus.Height {
		clog.Debug("paracross.Commit record", "node", commitAddrs, "titile", commit.Title, "height", commit.Height)
		addr := strings.Join(commitAddrs, ",") + strings.Join(supervisionValidAddrs, ",")
		return makeRecordReceipt(addr, commit), nil
	}

	// 未共识处理， 接受当前高度以及后续高度
	stat, err := getTitleHeight(a.db, calcTitleHeightKey(commit.Title, commit.Height))
	if err != nil && !isNotFound(err) {
		clog.Error("paracross.Commit getTitleHeight failed", "err", err)
		return nil, err
	}

	var receipt *types.Receipt
	var copyStat *pt.ParacrossHeightStatus
	if isNotFound(err) {
		stat = &pt.ParacrossHeightStatus{
			Status:  pt.ParacrossStatusCommiting,
			Title:   commit.Title,
			Height:  commit.Height,
			Details: &pt.ParacrossStatusDetails{},
		}
		if pt.IsParaForkHeight(cfg, a.exec.GetMainHeight(), pt.ForkCommitTx) {
			stat.MainHeight = commit.MainBlockHeight
			stat.MainHash = commit.MainBlockHash
		}
	} else {
		copyStat = proto.Clone(stat).(*pt.ParacrossHeightStatus)
	}

	for _, addr := range commitAddrs {
		// 如有分叉， 同一个节点可能再次提交commit交易
		found, index := hasCommited(stat.Details.Addrs, addr)
		if found {
			stat.Details.BlockHash[index] = commit.BlockHash
		} else {
			stat.Details.Addrs = append(stat.Details.Addrs, addr)
			stat.Details.BlockHash = append(stat.Details.BlockHash, commit.BlockHash)
		}
	}

	for _, addr := range supervisionValidAddrs {
		if stat.SupervisionDetails == nil {
			stat.SupervisionDetails = &pt.ParacrossStatusDetails{}
		}
		// 如有分叉， 同一个节点可能再次提交commit交易
		found, index := hasCommited(stat.SupervisionDetails.Addrs, addr)
		if found {
			stat.SupervisionDetails.BlockHash[index] = commit.BlockHash
		} else {
			stat.SupervisionDetails.Addrs = append(stat.SupervisionDetails.Addrs, addr)
			stat.SupervisionDetails.BlockHash = append(stat.SupervisionDetails.BlockHash, commit.BlockHash)
		}
	}

	//用commit.MainBlockHeight 判断更准确，如果用a.exec.MainHeight也可以，但是可能收到MainHeight之前的高度共识tx，
	// 后面loopCommitTxDone时候也是用当前共识高度大于分叉高度判断
	if pt.IsParaForkHeight(cfg, commit.MainBlockHeight, pt.ForkLoopCheckCommitTxDone) {
		updateCommitBlockHashs(stat, commit)
	}
	receipt = makeCommitReceipt(strings.Join(commitAddrs, ","), commit, copyStat, stat)

	//平行链fork pt.ForkCommitTx=0,主链在ForkCommitTx后支持nodegroup，这里平行链dappFork一定为true
	if cfg.IsDappFork(commit.MainBlockHeight, pt.ParaX, pt.ForkCommitTx) {
		updateCommitAddrs(stat, nodes)
	}
	if stat.SupervisionDetails != nil {
		updateSupervisionDetailsCommitAddrs(stat, supervisionNodes)
	}

	_ = saveTitleHeight(a.db, calcTitleHeightKey(stat.Title, stat.Height), stat)
	//fork之前记录的stat 没有根据nodes更新而更新
	if pt.IsParaForkHeight(cfg, stat.MainHeight, pt.ForkLoopCheckCommitTxDone) {
		r := makeCommitStatReceipt(stat)
		receipt = mergeReceipt(receipt, r)
	}

	if commit.Height > titleStatus.Height+1 {
		_ = saveTitleHeight(a.db, calcTitleHeightKey(commit.Title, commit.Height), stat)
		//平行链由主链共识无缝切换，即接收第一个收到的高度，可以不从0开始
		allow, err := a.isAllowConsensJump(commit, titleStatus)
		if err != nil {
			clog.Error("paracross.Commit allowJump", "err", err)
			return nil, err
		}
		if !allow {
			return receipt, nil
		}
	}
	r, err := a.commitTxDone(commit, stat, titleStatus, nodes, supervisionNodes)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)
	return receipt, nil
}

//分叉以前stat里面只记录了blockhash的信息，没有crossTxHash等信息，无法通过stat直接重构出mostCommitStatus
func (a *action) commitTxDone(nodeStatus *pt.ParacrossNodeStatus, stat *pt.ParacrossHeightStatus, titleStatus *pt.ParacrossStatus,
	nodes map[string]struct{}, supervisionNodes map[string]struct{}) (*types.Receipt, error) {
	receipt := &types.Receipt{}

	clog.Debug("paracross.Commit commit", "stat.title", stat.Title, "stat.height", stat.Height, "notes", len(nodes))
	for i, v := range stat.Details.Addrs {
		clog.Debug("paracross.Commit commit detail", "addr", v, "hash", common.ToHex(stat.Details.BlockHash[i]))
	}

	mostCount, mostHash := GetMostCommit(stat.Details.BlockHash)
	if !isCommitDone(len(nodes), mostCount) {
		return receipt, nil
	}

	clog.Debug("paracross.Commit commit ----pass", "most", mostCount, "mostHash", common.ToHex([]byte(mostHash)))
	// 如果已经有监督节点
	mostSupervisionCount := 0
	if len(supervisionNodes) > 0 && stat.SupervisionDetails != nil {
		for i, v := range stat.SupervisionDetails.Addrs {
			clog.Debug("paracross.Commit commit SupervisionDetails", "addr", v, "hash", common.ToHex(stat.SupervisionDetails.BlockHash[i]))
		}
		mostSupervisionCount, mostSupervisionHash := GetMostCommit(stat.SupervisionDetails.BlockHash)
		if !isCommitDone(len(supervisionNodes), mostSupervisionCount) {
			return receipt, nil
		}
		clog.Debug("paracross.Commit commit SupervisionDetails ----pass", "mostSupervisionCount", mostSupervisionCount, "mostSupervisionHash", common.ToHex([]byte(mostSupervisionHash)))

		if mostHash != mostSupervisionHash {
			clog.Error("paracross.Commit commit mostSupervisionHash mostHash not equal", "mostHash: ", common.ToHex([]byte(mostHash)), "mostSupervisionHash: ", common.ToHex([]byte(mostSupervisionHash)))
			return receipt, nil
		}
	}

	stat.Status = pt.ParacrossStatusCommitDone
	_ = saveTitleHeight(a.db, calcTitleHeightKey(stat.Title, stat.Height), stat)

	//之前记录的stat 状态没更新
	cfg := a.api.GetConfig()
	if pt.IsParaForkHeight(cfg, stat.MainHeight, pt.ForkLoopCheckCommitTxDone) {
		r := makeCommitStatReceipt(stat)
		receipt = mergeReceipt(receipt, r)
	}

	supervisionDetailsAddrsLen := 0
	if stat.SupervisionDetails != nil {
		supervisionDetailsAddrsLen = len(stat.SupervisionDetails.Addrs)
	}

	//add commit done receipt
	receiptDone := makeDoneReceipt(cfg, a.exec.GetMainHeight(), a.height, nodeStatus,
		int32(mostCount), int32(len(stat.Details.Addrs)), int32(len(nodes)),
		int32(mostSupervisionCount), int32(supervisionDetailsAddrsLen), int32(len(supervisionNodes)))
	receipt = mergeReceipt(receipt, receiptDone)

	r, err := a.commitTxDoneStep2(nodeStatus, stat, titleStatus)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)
	return receipt, nil
}

func (a *action) commitTxDoneStep2(nodeStatus *pt.ParacrossNodeStatus, stat *pt.ParacrossHeightStatus, titleStatus *pt.ParacrossStatus) (*types.Receipt, error) {
	receipt := &types.Receipt{}
	titleStatus.Title = nodeStatus.Title
	titleStatus.Height = nodeStatus.Height
	titleStatus.BlockHash = nodeStatus.BlockHash
	cfg := a.api.GetConfig()
	if pt.IsParaForkHeight(cfg, a.exec.GetMainHeight(), pt.ForkLoopCheckCommitTxDone) {
		titleStatus.MainHeight = nodeStatus.MainBlockHeight
		titleStatus.MainHash = nodeStatus.MainBlockHash
	}
	_ = saveTitle(a.db, calcTitleKey(titleStatus.Title), titleStatus)

	clog.Debug("paracross.Commit commit done", "height", nodeStatus.Height, "statusBlockHash", common.ToHex(nodeStatus.BlockHash))

	//parallel chain not need to process cross commit tx here
	if cfg.IsPara() {
		confPara := types.ConfSub(cfg, pt.ParaX)
		//如果关掉则不进行自共识
		if !confPara.IsEnable("closeSelfConsensus") {
			//平行链自共识校验
			selfBlockHash, err := getBlockHash(a.api, nodeStatus.Height)
			if err != nil {
				clog.Error("paracross.CommitDone getBlockHash", "err", err, "commit tx height", nodeStatus.Height, "tx", common.ToHex(a.txhash))
				return nil, err
			}
			//说明本节点blockhash和共识hash不一致，需要停止本节点执行
			if !bytes.Equal(selfBlockHash.Hash, nodeStatus.BlockHash) {
				clog.Error("paracross.CommitDone mosthash not match", "height", nodeStatus.Height,
					"blockHash", common.ToHex(selfBlockHash.Hash), "mosthash", common.ToHex(nodeStatus.BlockHash))
				return nil, types.ErrConsensusHashErr
			}
		}

		//平行连进行奖励分配
		rewardReceipt, err := a.reward(nodeStatus, stat)
		//错误会导致和主链处理的共识结果不一致
		if err != nil {
			clog.Error("paracross mining reward err", "height", nodeStatus.Height,
				"blockhash", common.ToHex(nodeStatus.BlockHash), "err", err)
			return nil, err
		}
		receipt = mergeReceipt(receipt, rewardReceipt)
		return receipt, nil
	}

	//主链，处理跨链交易
	r, err := a.procCrossTxs(nodeStatus)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)
	return receipt, nil
}

func isHaveCrossTxs(cfg *types.Chain33Config, status *pt.ParacrossNodeStatus) bool {
	//ForkLoopCheckCommitTxDone分叉后只返回全部txResult的结果，要实际过滤出来后才能确定有没有跨链tx
	if pt.IsParaForkHeight(cfg, status.MainBlockHeight, pt.ForkLoopCheckCommitTxDone) {
		return true
	}

	haveCrossTxs := len(status.CrossTxHashs) > 0
	//ForkCommitTx后，CrossTxHashs[][] 所有跨链交易做成一个校验hash，如果没有则[0]为nil
	if status.Height > 0 && pt.IsParaForkHeight(cfg, status.MainBlockHeight, pt.ForkCommitTx) && len(status.CrossTxHashs[0]) == 0 {
		haveCrossTxs = false
	}
	return haveCrossTxs

}

func (a *action) procCrossTxs(status *pt.ParacrossNodeStatus) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if enableParacrossTransfer && status.Height > 0 && isHaveCrossTxs(cfg, status) {
		clog.Debug("paracross.Commit commitDone do cross", "height", status.Height)

		crossTxs, crossTxResult, err := getCrossTxs(a.api, status)
		if err != nil {
			clog.Error("paracross.Commit getCrossTxs", "err", err.Error())
			return nil, err
		}
		crossTxReceipt, err := a.execCrossTxs(status.Title, status.Height, crossTxs, crossTxResult)
		if err != nil {
			return nil, err
		}
		return crossTxReceipt, nil
	}
	return nil, nil
}

//由于可能对当前块的共识交易进行处理，需要全部数据保存到statedb，通过tx获取数据无法处理当前块的场景
func (a *action) loopCommitTxDone(title string) (*types.Receipt, error) {
	receipt := &types.Receipt{}

	nodes, _, err := getParacrossNodes(a.db, title)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", title)
	}
	// 获取监督节点的数据
	supervisionNodes, _, _, err := getSupervisionNodeGroupAddrs(a.db, title)
	if err != nil && errors.Cause(err) != pt.ErrTitleNotExist {
		return nil, errors.Wrap(err, "getSupervisionNodeGroupAddrs loopCommitTxDone")
	}
	//从当前共识高度开始遍历
	titleStatus, err := getTitle(a.db, calcTitleKey(title))
	if err != nil {
		return nil, errors.Wrapf(err, "getTitle:%s", title)
	}
	//当前共识高度还未到分叉高度，则不处理
	cfg := a.api.GetConfig()
	if !pt.IsParaForkHeight(cfg, titleStatus.GetMainHeight(), pt.ForkLoopCheckCommitTxDone) {
		return nil, errors.Wrapf(pt.ErrForkHeightNotReach,
			"titleHeight:%d,forkHeight:%d", titleStatus.MainHeight, pt.GetDappForkHeight(cfg, pt.ForkLoopCheckCommitTxDone))
	}

	loopHeight := titleStatus.Height

	for {
		loopHeight++

		stat, err := getTitleHeight(a.db, calcTitleHeightKey(title, loopHeight))
		if err != nil {
			clog.Error("paracross.loopCommitTxDone getTitleHeight failed", "title", title, "height", loopHeight, "err", err.Error())
			return receipt, err
		}
		//防止无限循环
		if stat.MainHeight > a.exec.GetMainHeight() {
			return receipt, nil
		}

		r, err := a.checkCommitTxDone(stat, nodes, supervisionNodes)
		if err != nil {
			clog.Error("paracropara_cross_transfer main chain and game chain failedss.loopCommitTxDone checkExecCommitTxDone", "para title", title, "height", stat.Height, "error", err)
			return receipt, nil
		}
		if r == nil {
			return receipt, nil
		}
		receipt = mergeReceipt(receipt, r)
	}
}

func (a *action) checkCommitTxDone(stat *pt.ParacrossHeightStatus, nodes, supervisionNodes map[string]struct{}) (*types.Receipt, error) {
	status, err := getTitle(a.db, calcTitleKey(stat.Title))
	if err != nil {
		return nil, errors.Wrapf(err, "getTitle:%s", stat.Title)
	}

	//待共识的stat的高度大于当前status高度+1，说明共识不连续，退出，如果是平行链自共识首次切换场景，可以在正常流程里面再触发
	if stat.Height > status.Height+1 {
		return nil, nil
	}

	return a.commitTxDoneByStat(stat, status, nodes, supervisionNodes)
}

//只根据stat的信息在commitDone之后重构一个commitMostStatus做后续处理
func (a *action) commitTxDoneByStat(stat *pt.ParacrossHeightStatus, titleStatus *pt.ParacrossStatus, nodes, supervisionNodes map[string]struct{}) (*types.Receipt, error) {
	clog.Debug("paracross.commitTxDoneByStat", "stat.title", stat.Title, "stat.height", stat.Height, "notes", len(nodes))
	for i, v := range stat.Details.Addrs {
		clog.Debug("paracross.commitTxDoneByStat detail", "addr", v, "hash", common.ToHex(stat.Details.BlockHash[i]))
	}

	updateCommitAddrs(stat, nodes)
	most, mostHash := GetMostCommit(stat.Details.BlockHash)
	if !isCommitDone(len(nodes), most) {
		return nil, nil
	}
	clog.Debug("paracross.commitTxDoneByStat ----pass", "most", most, "mostHash", common.ToHex([]byte(mostHash)))

	mostSupervisionCount := 0
	if len(supervisionNodes) > 0 {
		for i, v := range stat.SupervisionDetails.Addrs {
			clog.Debug("paracross.commitTxDoneByStat SupervisionDetails", "addr", v, "hash", common.ToHex(stat.SupervisionDetails.BlockHash[i]))
		}

		updateSupervisionDetailsCommitAddrs(stat, supervisionNodes)
		mostSupervisionCount, mostSupervisionHash := GetMostCommit(stat.SupervisionDetails.BlockHash)
		if !isCommitDone(len(supervisionNodes), mostSupervisionCount) {
			return nil, nil
		}
		clog.Debug("paracross.commitTxDoneByStat SupervisionDetails ----pass", "mostSupervisionCount", mostSupervisionCount, "mostSupervisionHash", common.ToHex([]byte(mostSupervisionHash)))

		if mostHash != mostSupervisionHash {
			clog.Error("paracross.commitTxDoneByStat mostSupervisionHash mostHash not equal", "mostHash: ", common.ToHex([]byte(mostHash)), "mostSupervisionHash: ", common.ToHex([]byte(mostSupervisionHash)))
			return nil, nil
		}
	}

	stat.Status = pt.ParacrossStatusCommitDone
	_ = saveTitleHeight(a.db, calcTitleHeightKey(stat.Title, stat.Height), stat)
	r := makeCommitStatReceipt(stat)
	receipt := &types.Receipt{}
	receipt = mergeReceipt(receipt, r)

	txRst := getMostResults([]byte(mostHash), stat)
	mostStatus := &pt.ParacrossNodeStatus{
		MainBlockHash:   stat.MainHash,
		MainBlockHeight: stat.MainHeight,
		Title:           stat.Title,
		Height:          stat.Height,
		BlockHash:       []byte(mostHash),
		TxResult:        txRst,
	}

	supervisionDetailsAddrsLen := 0
	if stat.SupervisionDetails != nil {
		supervisionDetailsAddrsLen = len(stat.SupervisionDetails.Addrs)
	}

	//add commit done receipt
	cfg := a.api.GetConfig()
	receiptDone := makeDoneReceipt(cfg, a.exec.GetMainHeight(), a.height, mostStatus,
		int32(most), int32(len(stat.Details.Addrs)), int32(len(nodes)),
		int32(mostSupervisionCount), int32(supervisionDetailsAddrsLen), int32(len(supervisionNodes)))
	receipt = mergeReceipt(receipt, receiptDone)

	r, err := a.commitTxDoneStep2(mostStatus, stat, titleStatus)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)
	return receipt, nil
}

//主链共识跳跃条件： 仅支持主链共识初始高度为-1，也就是没有共识过，共识过不允许再跳跃
func (a *action) isAllowMainConsensJump(commit *pt.ParacrossNodeStatus, titleStatus *pt.ParacrossStatus) bool {
	cfg := a.api.GetConfig()
	if cfg.IsDappFork(a.exec.GetMainHeight(), pt.ParaX, pt.ForkLoopCheckCommitTxDone) {
		if titleStatus.Height == -1 {
			return true
		}
	}

	return false
}

//平行链自共识无缝切换条件：1，平行链没有共识过，2：commit高度是大于自共识分叉高度且上一次共识的主链高度小于自共识分叉高度，保证只运行一次，
// 1. 分叉之前，开启过共识的平行链需要从１跳跃，没开启过的将使用新版本，从0开始发送，不用考虑从１跳跃的问题
// 2. 分叉之后，只有stage.blockHeight== commit.height，也就是stage起始高度时候允许跳跃
func (a *action) isAllowParaConsensJump(commit *pt.ParacrossNodeStatus, titleStatus *pt.ParacrossStatus) (bool, error) {
	cfg := a.api.GetConfig()
	if cfg.IsDappFork(a.height, pt.ParaX, pt.ForkParaSelfConsStages) {
		stage, err := getSelfConsOneStage(a.db, commit.Height)
		if err != nil && errors.Cause(err) != pt.ErrKeyNotExist {
			return false, err
		}
		if stage == nil {
			return false, nil
		}
		return stage.StartHeight == commit.Height, nil
	}

	//兼容分叉之前从１跳跃场景
	return titleStatus.Height == -1, nil
}

func (a *action) isAllowConsensJump(commit *pt.ParacrossNodeStatus, titleStatus *pt.ParacrossStatus) (bool, error) {
	cfg := a.api.GetConfig()
	if cfg.IsPara() {
		return a.isAllowParaConsensJump(commit, titleStatus)
	}
	return a.isAllowMainConsensJump(commit, titleStatus), nil
}

func execCrossTxNew(a *action, cross *types.Transaction, crossTxHash []byte) (*types.Receipt, error) {
	if !bytes.HasSuffix(cross.Execer, []byte(pt.ParaX)) {
		return nil, nil
	}
	var payload pt.ParacrossAction
	err := types.Decode(cross.Payload, &payload)
	if err != nil {
		clog.Crit("paracross.Commit Decode Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
		return nil, err
	}

	if payload.Ty == pt.ParacrossActionCrossAssetTransfer {
		act, err := getCrossAction(payload.GetCrossAssetTransfer(), string(cross.Execer))
		if err != nil {
			clog.Crit("paracross.Commit getCrossAction Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
			return nil, err
		}
		if act == pt.ParacrossMainAssetWithdraw || act == pt.ParacrossParaAssetTransfer {
			receipt, err := a.crossAssetTransfer(payload.GetCrossAssetTransfer(), act, cross)
			if err != nil {
				clog.Crit("paracross.Commit crossAssetTransfer Tx failed", "error", err, "act", act, "txHash", common.ToHex(crossTxHash))
				return nil, err
			}
			clog.Debug("paracross.Commit crossAssetTransfer done", "act", act, "txHash", common.ToHex(crossTxHash))
			return receipt, nil
		}

	}

	//主链共识后，执行主链资产withdraw, 在支持CrossAssetTransfer之前使用此action
	if payload.Ty == pt.ParacrossActionAssetWithdraw {
		receiptWithdraw, err := a.assetWithdraw(payload.GetAssetWithdraw(), cross)
		if err != nil {
			clog.Crit("paracross.Commit withdraw Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
			return nil, errors.Cause(err)
		}

		clog.Debug("paracross.Commit WithdrawCoins", "txHash", common.ToHex(crossTxHash))
		return receiptWithdraw, nil
	}
	return nil, nil
}

//func execCrossTx(a *action, cross *types.TransactionDetail, crossTxHash []byte) (*types.Receipt, error) {
//	if !bytes.HasSuffix(cross.Tx.Execer, []byte(pt.ParaX)) {
//		return nil, nil
//	}
//	var payload pt.ParacrossAction
//	err := types.Decode(cross.Tx.Payload, &payload)
//	if err != nil {
//		clog.Crit("paracross.Commit Decode Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
//		return nil, err
//	}
//
//	if payload.Ty == pt.ParacrossActionCrossAssetTransfer {
//		act, err := getCrossAction(payload.GetCrossAssetTransfer(), string(cross.Tx.Execer))
//		if err != nil {
//			clog.Crit("paracross.Commit getCrossAction Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
//			return nil, err
//		}
//		if act == pt.ParacrossMainAssetWithdraw || act == pt.ParacrossParaAssetTransfer {
//			receipt, err := a.crossAssetTransfer(payload.GetCrossAssetTransfer(), act, cross.Tx)
//			if err != nil {
//				clog.Crit("paracross.Commit crossAssetTransfer Tx failed", "error", err, "act", act, "txHash", common.ToHex(crossTxHash))
//				return nil, err
//			}
//			clog.Debug("paracross.Commit crossAssetTransfer done", "act", act, "txHash", common.ToHex(crossTxHash))
//			return receipt, nil
//		}
//
//	}
//
//	//主链共识后，执行主链资产withdraw, 在支持CrossAssetTransfer之前使用此action
//	if payload.Ty == pt.ParacrossActionAssetWithdraw {
//		receiptWithdraw, err := a.assetWithdraw(payload.GetAssetWithdraw(), cross.Tx)
//		if err != nil {
//			clog.Crit("paracross.Commit withdraw Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
//			return nil, errors.Cause(err)
//		}
//
//		clog.Debug("paracross.Commit WithdrawCoins", "txHash", common.ToHex(crossTxHash))
//		return receiptWithdraw, nil
//	}
//	return nil, nil
//}

func rollbackCrossTxNew(a *action, cross *types.Transaction, crossTxHash []byte) (*types.Receipt, error) {
	if !bytes.HasSuffix(cross.Execer, []byte(pt.ParaX)) {
		return nil, nil
	}
	var payload pt.ParacrossAction
	err := types.Decode(cross.Payload, &payload)
	if err != nil {
		clog.Crit("paracross.Commit.rollbackCrossTx Decode Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
		return nil, err
	}

	if payload.Ty == pt.ParacrossActionCrossAssetTransfer {
		act, err := getCrossAction(payload.GetCrossAssetTransfer(), string(cross.Execer))
		if err != nil {
			clog.Crit("paracross.Commit.rollbackCrossTx getCrossAction failed", "error", err, "txHash", common.ToHex(crossTxHash))
			return nil, err
		}
		//主链共识后，平行链执行出错的主链资产transfer回滚
		if act == pt.ParacrossMainAssetTransfer {
			receipt, err := a.assetTransferRollback(payload.GetCrossAssetTransfer(), cross)
			if err != nil {
				clog.Crit("paracross.Commit crossAssetTransfer rbk failed", "error", err, "txHash", common.ToHex(crossTxHash))
				return nil, errors.Cause(err)
			}

			clog.Debug("paracross.Commit crossAssetTransfer rollbackCrossTx", "txHash", common.ToHex(crossTxHash), "mainHeight", a.height)
			return receipt, nil
		}
		//主链共识后，平行链执行出错的平行链资产withdraw回滚
		if act == pt.ParacrossParaAssetWithdraw {
			receipt, err := a.paraAssetWithdrawRollback(payload.GetCrossAssetTransfer(), cross)
			if err != nil {
				clog.Crit("paracross.Commit rbk paraAssetWithdraw Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
				return nil, errors.Cause(err)
			}

			clog.Debug("paracross.Commit paraAssetWithdraw rollbackCrossTx", "txHash", common.ToHex(crossTxHash), "mainHeight", a.height)
			return receipt, nil
		}
	}

	//主链共识后，平行链执行出错的主链资产transfer回滚
	if payload.Ty == pt.ParacrossActionAssetTransfer {
		cfg := payload.GetAssetTransfer()
		transfer := &pt.CrossAssetTransfer{
			AssetSymbol: cfg.Cointoken,
			Amount:      cfg.Amount,
			Note:        string(cfg.Note),
			ToAddr:      cfg.To,
		}

		receipt, err := a.assetTransferRollback(transfer, cross)
		if err != nil {
			clog.Crit("paracross.Commit rbk asset transfer Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
			return nil, errors.Cause(err)
		}

		clog.Debug("paracross.Commit assetTransfer rollbackCrossTx", "txHash", common.ToHex(crossTxHash), "mainHeight", a.height)
		return receipt, nil
	}
	return nil, nil

}

//func rollbackCrossTx(a *action, cross *types.TransactionDetail, crossTxHash []byte) (*types.Receipt, error) {
//	if !bytes.HasSuffix(cross.Tx.Execer, []byte(pt.ParaX)) {
//		return nil, nil
//	}
//	var payload pt.ParacrossAction
//	err := types.Decode(cross.Tx.Payload, &payload)
//	if err != nil {
//		clog.Crit("paracross.Commit.rollbackCrossTx Decode Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
//		return nil, err
//	}
//
//	if payload.Ty == pt.ParacrossActionCrossAssetTransfer {
//		act, err := getCrossAction(payload.GetCrossAssetTransfer(), string(cross.Tx.Execer))
//		if err != nil {
//			clog.Crit("paracross.Commit.rollbackCrossTx getCrossAction failed", "error", err, "txHash", common.ToHex(crossTxHash))
//			return nil, err
//		}
//		//主链共识后，平行链执行出错的主链资产transfer回滚
//		if act == pt.ParacrossMainAssetTransfer {
//			receipt, err := a.assetTransferRollback(payload.GetCrossAssetTransfer(), cross.Tx)
//			if err != nil {
//				clog.Crit("paracross.Commit crossAssetTransfer rbk failed", "error", err, "txHash", common.ToHex(crossTxHash))
//				return nil, errors.Cause(err)
//			}
//
//			clog.Debug("paracross.Commit crossAssetTransfer rollbackCrossTx", "txHash", common.ToHex(crossTxHash), "mainHeight", a.height)
//			return receipt, nil
//		}
//		//主链共识后，平行链执行出错的平行链资产withdraw回滚
//		if act == pt.ParacrossParaAssetWithdraw {
//			receipt, err := a.paraAssetWithdrawRollback(payload.GetCrossAssetTransfer(), cross.Tx)
//			if err != nil {
//				clog.Crit("paracross.Commit rbk paraAssetWithdraw Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
//				return nil, errors.Cause(err)
//			}
//
//			clog.Debug("paracross.Commit paraAssetWithdraw rollbackCrossTx", "txHash", common.ToHex(crossTxHash), "mainHeight", a.height)
//			return receipt, nil
//		}
//	}
//
//	//主链共识后，平行链执行出错的主链资产transfer回滚
//	if payload.Ty == pt.ParacrossActionAssetTransfer {
//		cfg := payload.GetAssetTransfer()
//		transfer := &pt.CrossAssetTransfer{
//			AssetSymbol: cfg.Cointoken,
//			Amount:      cfg.Amount,
//			Note:        string(cfg.Note),
//			ToAddr:      cfg.To,
//		}
//
//		receipt, err := a.assetTransferRollback(transfer, cross.Tx)
//		if err != nil {
//			clog.Crit("paracross.Commit rbk asset transfer Tx failed", "error", err, "txHash", common.ToHex(crossTxHash))
//			return nil, errors.Cause(err)
//		}
//
//		clog.Debug("paracross.Commit assetTransfer rollbackCrossTx", "txHash", common.ToHex(crossTxHash), "mainHeight", a.height)
//		return receipt, nil
//	}
//	return nil, nil
//
//}

//无跨链交易高度列表是人为配置的，是确认的历史高度，是一种特殊处理，不会影响区块状态hash
//para.ignore.10-100.200-300
func isInIgnoreHeightList(str string, status *pt.ParacrossNodeStatus) (bool, error) {
	if len(str) <= 0 {
		return false, nil
	}
	e := strings.Split(str, ".")
	if len(e) <= 2 {
		return false, errors.Wrapf(types.ErrInvalidParam, "wrong config str=%s,title=%s", str, status.Title)
	}
	if strings.ToLower(pt.ParaPrefix+e[0]+".") != strings.ToLower(status.Title) {
		return false, errors.Wrapf(types.ErrInvalidParam, "wrong title str=%s,title=%s", str, status.Title)
	}

	if e[1] != pt.ParaCrossAssetTxIgnoreKey {
		return false, errors.Wrapf(types.ErrInvalidParam, "wrong ignore str=%s,title=%s", str, status.Title)
	}

	for _, h := range e[2:] {
		p := strings.Split(h, "-")
		if len(p) != 2 {
			return false, errors.Wrapf(types.ErrInvalidParam, "check NoHeightCrossAsseList title=%s,height=%s", status.Title, h)
		}
		s, err := strconv.Atoi(p[0])
		if err != nil {
			return false, errors.Wrapf(err, "check NoHeightCrossAsseList title=%s,height=%s", status.Title, h)
		}
		e, err := strconv.Atoi(p[1])
		if err != nil {
			return false, errors.Wrapf(err, "check NoHeightCrossAsseList title=%s,height=%s", status.Title, h)
		}
		if s > e {
			return false, errors.Wrapf(types.ErrInvalidParam, "check NoHeightCrossAsseList title=%s,height=%s", status.Title, h)
		}

		//共识的平行链高度(不是主链高度）落在范围内，说明此高度没有跨链资产交易，可以忽略
		if status.Height >= int64(s) && status.Height <= int64(e) {
			return true, nil
		}
	}
	return false, nil
}

// "para.hit.10.100.200"
func isInHitHeightList(str string, status *pt.ParacrossNodeStatus) (bool, error) {
	if len(str) <= 0 {
		return false, nil
	}

	e := strings.Split(str, ".")
	if len(e) <= 2 {
		return false, errors.Wrapf(types.ErrInvalidParam, "wrong config str=%s,title=%s", str, status.Title)
	}
	if strings.ToLower(pt.ParaPrefix+e[0]+".") != strings.ToLower(status.Title) {
		return false, errors.Wrapf(types.ErrInvalidParam, "wrong title str=%s,title=%s", str, status.Title)
	}
	if e[1] != pt.ParaCrossAssetTxHitKey {
		return false, errors.Wrapf(types.ErrInvalidParam, "wrong hit str=%s,title=%s", str, status.Title)
	}

	for _, hStr := range e[2:] {
		h, err := strconv.Atoi(hStr)
		if err != nil {
			return false, errors.Wrapf(types.ErrInvalidParam, "wrong config str=%s in %s,title=%s", str, hStr, status.Title)
		}
		//高度命中
		if status.Height == int64(h) {
			return true, nil
		}
	}
	return false, nil
}

//命中高度
//s: para.hit.10.100, title=user.p.para.
//在有设置ParaCrossStatusBitMapVerLen的老版本之前，有些共识tx对应的区块没有跨链tx，分片查询很浪费时间，这里统计出来后，
//增加到配置文件里直接忽略相应高度。采用新的BitMap版本后，主链可以根据版本号判断有没有跨链tx，没有则直接退出，不需要再去区块中检查了。
func checkIsIgnoreHeight(heightList []string, status *pt.ParacrossNodeStatus) (bool, error) {
	if len(heightList) <= 0 {
		return false, nil
	}

	var hitStr, ignoreStr string
	hitPrefix := strings.ToLower(status.Title + pt.ParaCrossAssetTxHitKey)[len(pt.ParaPrefix):]
	ignorePrefix := strings.ToLower(status.Title + pt.ParaCrossAssetTxIgnoreKey)[len(pt.ParaPrefix):]

	for _, s := range heightList {
		desStr := strings.ToLower(s)
		if strings.HasPrefix(desStr, hitPrefix) {
			if len(hitStr) > 0 {
				return false, errors.Wrapf(types.ErrInvalidParam, "checkIsInIgnoreHeightList repeate=%s", hitPrefix)
			}
			hitStr = s
		}
		if strings.HasPrefix(desStr, ignorePrefix) {
			if len(ignoreStr) > 0 {
				return false, errors.Wrapf(types.ErrInvalidParam, "checkIsInIgnoreHeightList repeate=%s", ignorePrefix)
			}
			ignoreStr = s
		}
		if len(hitStr) > 0 && len(ignoreStr) > 0 {
			break
		}

	}

	in, err := isInHitHeightList(hitStr, status)
	if err != nil {
		return false, err
	}
	//如果在hit 列表中，不忽略
	if in {
		return false, nil
	}

	return isInIgnoreHeightList(ignoreStr, status)

}

func getCrossTxsByRst(api client.QueueProtocolAPI, status *pt.ParacrossNodeStatus) ([]*types.Transaction, []byte, error) {
	//支持带版本号的跨链交易bitmap
	//1.如果等于0，是老版本的平行链，按老的方式处理. 2. 如果大于0等于ver，新版本且没有跨链交易，不需要处理. 3. 大于ver，说明有跨链交易按老的方式处理
	if len(string(status.CrossTxResult)) == pt.ParaCrossStatusBitMapVerLen {
		return nil, nil, nil
	}

	rst, err := hex.DecodeString(string(status.TxResult))
	if err != nil {
		clog.Error("getCrossTxHashs decode rst", "CrossTxResult", string(status.TxResult), "paraHeight", status.Height)
		return nil, nil, types.ErrInvalidParam
	}

	cfg := api.GetConfig()
	if !cfg.IsDappFork(status.MainBlockHeight, pt.ParaX, pt.ForkParaAssetTransferRbk) {
		if len(rst) == 0 {
			return nil, nil, nil
		}
	}

	//在有设置ParaCrossStatusBitMapVerLen的老版本之前，有些共识tx对应的区块没有跨链tx，分片查询很浪费时间，这里统计出来后，
	//增加到配置文件里直接忽略相应高度。采用新的BitMap版本后，主链可以根据版本号判断有没有跨链tx，没有则直接退出，不需要再去区块中检查了。
	//para.hit.6.8, para.ignore.1-10, 比如高度7， 如果命中则继续处理，如果没命中，检查是否在ignore列表，如果在直接退出，否则继续处理
	//零散的命中列表可以减少忽略高度列表的范围
	//此平行链高度在忽略检查跨链交易列表中,则直接退出
	conf := types.ConfSub(api.GetConfig(), pt.ParaX)
	heightList := conf.GStrList("paraCrossAssetTxHeightList")
	ignore, err := checkIsIgnoreHeight(heightList, status)
	if err != nil {
		return nil, nil, err
	}
	if ignore {
		return nil, nil, nil
	}

	blockDetail, err := GetBlock(api, status.MainBlockHash)
	if err != nil {
		return nil, nil, err
	}

	//抽取平行链交易和跨链交易
	paraAllTxs := FilterTxsForPara(cfg, blockDetail.FilterParaTxsByTitle(cfg, status.Title))
	var baseHashs [][]byte
	for _, tx := range paraAllTxs {
		baseHashs = append(baseHashs, tx.Hash())
	}
	paraCrossTxs := FilterParaCrossTxs(paraAllTxs)
	var paraCrossHashs [][]byte
	for _, tx := range paraCrossTxs {
		paraCrossHashs = append(paraCrossHashs, tx.Hash())
	}
	crossRst := util.CalcBitMapByBitMap(paraCrossHashs, baseHashs, rst)
	return paraCrossTxs, crossRst, nil
}

func getCrossTxs(api client.QueueProtocolAPI, status *pt.ParacrossNodeStatus) ([]*types.Transaction, []byte, error) {
	cfg := api.GetConfig()
	if pt.IsParaForkHeight(cfg, status.MainBlockHeight, pt.ForkLoopCheckCommitTxDone) {
		return getCrossTxsByRst(api, status)
	}

	if len(status.CrossTxHashs) == 0 {
		clog.Error("getCrossTxHashs len=0", "paraHeight", status.Height,
			"mainHeight", status.MainBlockHeight, "mainHash", common.ToHex(status.MainBlockHash))
		return nil, nil, types.ErrCheckTxHash
	}

	//para.hit.6.8, para.ignore.1-10, 比如高度7， 如果命中则继续处理，如果没命中，检查是否在ignore列表，如果在直接退出，否则继续处理
	//零散的命中列表可以减少忽略高度列表的范围
	//比如高度6，命中，则继续处理，高度7，未命中，但是在ignore scope,退出，高度11，未命中，也不在ignore scope,继续处理
	conf := types.ConfSub(api.GetConfig(), pt.ParaX)
	heightList := conf.GStrList("paraCrossAssetTxHeightList")
	ignore, err := checkIsIgnoreHeight(heightList, status)
	if err != nil {
		return nil, nil, err
	}
	if ignore {
		return nil, nil, nil
	}

	blockDetail, err := GetBlock(api, status.MainBlockHash)
	if err != nil {
		return nil, nil, err
	}
	if !pt.IsParaForkHeight(cfg, status.MainBlockHeight, pt.ForkCommitTx) {
		// 直接按照status.CrossTxHashs指定的哈希列表查找交易
		paraCrossTxs := make([]*types.Transaction, len(status.CrossTxHashs))
		txs := blockDetail.Block.Txs
		for i, hash := range status.CrossTxHashs {
			for j, tx := range txs {
				if bytes.Equal(hash, tx.Hash()) {
					paraCrossTxs[i] = tx
					txs = txs[j:]
					break
				}
			}
		}
		if len(paraCrossTxs) != len(status.CrossTxHashs) {
			return nil, nil, types.ErrTxNotExist
		}
		return paraCrossTxs, status.CrossTxResult, nil
	}
	//校验
	paraBaseTxs := FilterTxsForPara(cfg, blockDetail.FilterParaTxsByTitle(cfg, status.Title))
	paraCrossTxs := FilterParaCrossTxs(paraBaseTxs)
	var baseHashs [][]byte
	for _, tx := range paraBaseTxs {
		baseHashs = append(baseHashs, tx.Hash())
	}
	var paraCrossHashs [][]byte
	for _, tx := range paraCrossTxs {
		paraCrossHashs = append(paraCrossHashs, tx.Hash())
	}
	baseCheckTxHash := CalcTxHashsHash(baseHashs)
	crossCheckHash := CalcTxHashsHash(paraCrossHashs)
	if !bytes.Equal(status.CrossTxHashs[0], crossCheckHash) {
		clog.Error("getCrossTxHashs para hash not equal", "paraHeight", status.Height,
			"mainHeight", status.MainBlockHeight, "mainHash", common.ToHex(status.MainBlockHash),
			"main.crossHash", common.ToHex(crossCheckHash), "commit.crossHash", common.ToHex(status.CrossTxHashs[0]),
			"main.baseHash", common.ToHex(baseCheckTxHash), "commit.baseHash", common.ToHex(status.TxHashs[0]))
		for _, hash := range baseHashs {
			clog.Error("getCrossTxHashs base tx hash", "txhash", common.ToHex(hash))
		}
		for _, hash := range paraCrossHashs {
			clog.Error("getCrossTxHashs paracross tx hash", "txhash", common.ToHex(hash))
		}
		return nil, nil, types.ErrCheckTxHash
	}

	//只获取跨链tx
	rst, err := hex.DecodeString(string(status.CrossTxResult))
	if err != nil {
		clog.Error("getCrossTxHashs decode rst", "CrossTxResult", string(status.CrossTxResult), "paraHeight", status.Height)
		return nil, nil, types.ErrInvalidParam
	}

	return paraCrossTxs, rst, nil
}

func (a *action) execCrossTxs(title string, height int64, crossTxs []*types.Transaction, crossTxResult []byte) (*types.Receipt, error) {
	var receipt types.Receipt

	for i := 0; i < len(crossTxs); i++ {
		clog.Debug("paracross.Commit commitDone", "do cross number", i, "hash", common.ToHex(crossTxs[i].Hash()),
			"res", util.BitMapBit(crossTxResult, uint32(i)))
		if util.BitMapBit(crossTxResult, uint32(i)) {
			//receiptCross, err := crossTxProc(a, crossTxHashs[i], execCrossTx)
			receiptCross, err := execCrossTxNew(a, crossTxs[i], crossTxs[i].Hash())
			if err != nil {
				clog.Error("paracross.Commit execCrossTx", "para title", title, "para height", height,
					"para tx index", i, "error", err)
				return nil, errors.Cause(err)
			}
			if receiptCross == nil {
				continue
			}
			clog.Debug("paracross.Commit commitDone.title ok ", "title", title, "height", height, "main", a.height, "i", i, "hash", common.ToHex(crossTxs[i].Hash()))
			receipt.KV = append(receipt.KV, receiptCross.KV...)
			receipt.Logs = append(receipt.Logs, receiptCross.Logs...)
		} else {
			clog.Error("paracross.Commit commitDone", "do cross number", i, "hash",
				common.ToHex(crossTxs[i].Hash()), "para res", util.BitMapBit(crossTxResult, uint32(i)))
			cfg := a.api.GetConfig()
			if cfg.IsDappFork(a.height, pt.ParaX, pt.ForkParaAssetTransferRbk) {
				//receiptCross, err := crossTxProc(a, crossTxHashs[i], rollbackCrossTx)
				receiptCross, err := rollbackCrossTxNew(a, crossTxs[i], crossTxs[i].Hash())
				if err != nil {
					clog.Error("paracross.Commit rollbackCrossTx", "para title", title, "para height", height,
						"para tx index", i, "error", err)
					return nil, errors.Cause(err)
				}
				if receiptCross == nil {
					continue
				}
				clog.Debug("paracross.Commit commitDone.title rbk", "title", title, "height", height, "main", a.height, "i", i, "hash", common.ToHex(crossTxs[i].Hash()))
				receipt.KV = append(receipt.KV, receiptCross.KV...)
				receipt.Logs = append(receipt.Logs, receiptCross.Logs...)
			}
		}
	}

	return &receipt, nil

}

func (a *action) assetTransferMainCheck(cfg *types.Chain33Config, transfer *types.AssetsTransfer) error {
	//主链如果没有nodegroup配置，也不允许跨链,直接返回错误，平行链也不会执行
	if cfg.IsDappFork(a.height, pt.ParaX, pt.ForkParaAssetTransferRbk) {
		if len(transfer.To) == 0 {
			return errors.Wrap(types.ErrInvalidParam, "toAddr should not be null")
		}
		return a.isAllowTransfer()
	}
	return nil
}

func (a *action) AssetTransfer(transfer *types.AssetsTransfer) (*types.Receipt, error) {
	clog.Debug("Paracross.Exec", "AssetTransfer", transfer.Cointoken, "transfer", "")
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()

	//主链如果没有nodegroup配置，也不允许跨链,直接返回错误，平行链也不会执行
	if !isPara {
		err := a.assetTransferMainCheck(cfg, transfer)
		if err != nil {
			return nil, errors.Wrap(err, "AssetTransfer check")
		}
	}

	receipt, err := a.assetTransfer(transfer)
	if err != nil {
		return nil, errors.Wrap(err, "AssetTransfer")
	}
	return receipt, nil
}

func (a *action) assetWithdrawMainCheck(cfg *types.Chain33Config, withdraw *types.AssetsWithdraw) error {
	if !cfg.IsDappFork(a.height, pt.ParaX, "ForkParacrossWithdrawFromParachain") {
		if withdraw.Cointoken != "" {
			return errors.Wrapf(types.ErrNotSupport, "not support,token=%s", withdraw.Cointoken)
		}
	}

	//rbk fork后　如果没有nodegroup　conf，也不允许跨链
	if cfg.IsDappFork(a.height, pt.ParaX, pt.ForkParaAssetTransferRbk) {
		if len(withdraw.To) == 0 {
			return errors.Wrap(types.ErrInvalidParam, "toAddr should not be null")
		}
		err := a.isAllowTransfer()
		if err != nil {
			return errors.Wrap(err, "AssetWithdraw not allow")
		}
	}
	return nil
}

func (a *action) AssetWithdraw(withdraw *types.AssetsWithdraw) (*types.Receipt, error) {
	//分叉高度之后，支持从平行链提取资产
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	if !isPara {
		err := a.assetWithdrawMainCheck(cfg, withdraw)
		if err != nil {
			return nil, err
		}
	}

	if !isPara {
		// 需要平行链先执行， 达成共识时，继续执行
		return nil, nil
	}
	clog.Debug("paracross.AssetWithdraw isPara", "execer", string(a.tx.Execer),
		"txHash", common.ToHex(a.tx.Hash()), "token name", withdraw.Cointoken)
	receipt, err := a.assetWithdraw(withdraw, a.tx)
	if err != nil {
		return nil, errors.Wrap(err, "AssetWithdraw failed")
	}
	return receipt, nil
}

func (a *action) crossAssetTransferMainCheck(transfer *pt.CrossAssetTransfer) error {
	err := a.isAllowTransfer()
	if err != nil {
		return errors.Wrap(err, "not Allow")
	}

	if len(transfer.AssetExec) == 0 || len(transfer.AssetSymbol) == 0 || transfer.Amount == 0 || len(transfer.ToAddr) == 0 {
		return errors.Wrapf(types.ErrInvalidParam, "exec=%s, symbol=%s, amount=%d,toAddr=%s should not be null",
			transfer.AssetExec, transfer.AssetSymbol, transfer.Amount, transfer.ToAddr)
	}
	return nil
}

func (a *action) CrossAssetTransfer(transfer *pt.CrossAssetTransfer) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()

	//主链检查参数
	if !isPara {
		err := a.crossAssetTransferMainCheck(transfer)
		if err != nil {
			return nil, err
		}
	}

	//平行链在ForkRootHash之前不支持crossAssetTransfer
	if isPara && !cfg.IsFork(a.exec.GetMainHeight(), "ForkRootHash") {
		return nil, errors.Wrap(types.ErrNotSupport, "not Allow before ForkRootHash")
	}

	act, err := getCrossAction(transfer, string(a.tx.Execer))
	if act == pt.ParacrossNoneTransfer {
		return nil, errors.Wrap(err, "non action")
	}
	// 需要平行链先执行， 达成共识时，继续执行
	if !isPara && (act == pt.ParacrossMainAssetWithdraw || act == pt.ParacrossParaAssetTransfer) {
		return nil, nil
	}
	receipt, err := a.crossAssetTransfer(transfer, act, a.tx)
	if err != nil {
		return nil, errors.Wrap(err, "CrossAssetTransfer failed")
	}
	return receipt, nil
}

func getTitleFrom(exec []byte) ([]byte, error) {
	last := bytes.LastIndex(exec, []byte("."))
	if last == -1 {
		return nil, types.ErrNotFound
	}
	// 现在配置是包含 .的， 所有取title 是也把 `.` 取出来
	return exec[:last+1], nil
}

func (a *action) isAllowTransfer() error {
	//1. 没有配置nodegroup　不允许
	tempTitle, err := getTitleFrom(a.tx.Execer)
	if err != nil {
		return errors.Wrapf(types.ErrInvalidParam, "not para chain exec=%s", string(a.tx.Execer))
	}
	nodes, _, err := a.getNodesGroup(string(tempTitle))
	if err != nil {
		return errors.Wrapf(err, "nodegroup not config,title=%s", tempTitle)
	}
	if len(nodes) == 0 {
		return errors.Wrapf(types.ErrNotSupport, "nodegroup not create,title=%s", tempTitle)
	}
	//2. 非跨链执行器不允许
	if !types.IsParaExecName(string(a.tx.Execer)) {
		return errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
	}

	return nil
}

func (a *action) Transfer(transfer *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	clog.Debug("Paracross.Exec Transfer", "symbol", transfer.Cointoken, "amount",
		transfer.Amount, "to", tx.To)
	from := tx.From()

	cfg := a.api.GetConfig()
	acc, err := account.NewAccountDB(cfg, pt.ParaX, transfer.Cointoken, a.db)
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
	cfg := a.api.GetConfig()
	acc, err := account.NewAccountDB(cfg, pt.ParaX, withdraw.Cointoken, a.db)
	if err != nil {
		clog.Error("Withdraw failed", "err", err)
		return nil, err
	}
	if dapp.IsDriverAddress(tx.GetRealToAddr(), a.height) || dapp.ExecAddress(withdraw.ExecName) == tx.GetRealToAddr() {
		return acc.TransferWithdraw(tx.From(), tx.GetRealToAddr(), withdraw.Amount)
	}
	return nil, types.ErrToAddrNotSameToExecAddr
}

func (a *action) TransferToExec(transfer *types.AssetsTransferToExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	clog.Debug("Paracross.Exec TransferToExec", "symbol", transfer.Cointoken, "amount",
		transfer.Amount, "to", tx.To)
	from := tx.From()

	cfg := a.api.GetConfig()
	acc, err := account.NewAccountDB(cfg, pt.ParaX, transfer.Cointoken, a.db)
	if err != nil {
		clog.Error("TransferToExec failed", "err", err)
		return nil, err
	}
	//to 是 execs 合约地址
	if dapp.IsDriverAddress(tx.GetRealToAddr(), a.height) || dapp.ExecAddress(transfer.ExecName) == tx.GetRealToAddr() {
		return acc.TransferToExec(from, tx.GetRealToAddr(), transfer.Amount)
	}
	return nil, types.ErrToAddrNotSameToExecAddr
}
