// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

var (
	clog                    = log.New("module", "execs.paracross")
	enableParacrossTransfer = true
	driverName              = pt.ParaX
)

// Paracross exec
type Paracross struct {
	cryptoCli crypto.Crypto
	drivers.DriverBase
}

//Init paracross exec register
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), newParacross, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
	setPrefix()
}

//InitExecType ...
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Paracross{}))
}

//GetName return paracross name
func GetName() string {
	return newParacross().GetName()
}

func newParacross() drivers.Driver {
	c := &Paracross{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	cli, err := crypto.Load("bls", -1)
	if err != nil {
		panic("paracross need bls sign register")
	}
	c.cryptoCli = cli
	return c
}

// GetDriverName return paracross driver name
func (c *Paracross) GetDriverName() string {
	return pt.ParaX
}

// CheckTx ...
func (c *Paracross) CheckTx(tx *types.Transaction, index int) error {
	//fork之后不对tx的tx.exec==tx.to做限制，支持transfer2Exec的场景
	if c.GetAPI().GetConfig().IsDappFork(c.GetHeight(), pt.ParaX, pt.ForkParaCheckTx) {
		return nil
	}
	return c.DriverBase.CheckTx(tx, index)
}

func (c *Paracross) checkTxGroup(tx *types.Transaction, index int) ([]*types.Transaction, error) {
	if tx.GroupCount >= 2 {
		txs, err := c.GetTxGroup(index)
		if err != nil {
			clog.Error("ParacrossActionAssetTransfer", "get tx group failed", err, "hash", hex.EncodeToString(tx.Hash()))
			return nil, err
		}
		return txs, nil
	}
	return nil, nil
}

func (c *Paracross) saveLocalParaTxs(tx *types.Transaction, isDel bool) (*types.LocalDBSet, error) {

	var payload pt.ParacrossAction
	err := types.Decode(tx.Payload, &payload)
	if err != nil {
		return nil, err
	}
	if payload.Ty != pt.ParacrossActionCommit || payload.GetCommit() == nil {
		return nil, nil
	}

	commit := payload.GetCommit()
	crossTxs, crossTxResults, err := getCrossTxs(c.GetAPI(), commit.Status)
	if err != nil {
		return nil, err
	}

	return c.updateLocalParaTxs(commit.Status.Title, commit.Status.Height, crossTxs, crossTxResults, isDel)

}

//无法获取到commit tx信息，从commitDone 结构里面构建
func (c *Paracross) saveLocalParaTxsFork(commitDone *pt.ReceiptParacrossDone, isDel bool) (*types.LocalDBSet, error) {
	status := &pt.ParacrossNodeStatus{
		MainBlockHash:   commitDone.MainBlockHash,
		MainBlockHeight: commitDone.MainBlockHeight,
		Title:           commitDone.Title,
		Height:          commitDone.Height,
		BlockHash:       commitDone.BlockHash,
		TxResult:        commitDone.TxResult,
	}

	crossTxs, crossTxResults, err := getCrossTxs(c.GetAPI(), status)
	if err != nil {
		return nil, err
	}

	return c.updateLocalParaTxs(commitDone.Title, commitDone.Height, crossTxs, crossTxResults, isDel)

}

func (c *Paracross) updateLocalParaTxs(paraTitle string, paraHeight int64, crossTxs []*types.Transaction,
	crossTxResults []byte, isDel bool) (*types.LocalDBSet, error) {

	dbSet := &types.LocalDBSet{}
	if len(crossTxs) == 0 {
		return dbSet, nil
	}

	for i, tx := range crossTxs {

		execOK := util.BitMapBit(crossTxResults, uint32(i))
		set, err := c.updateLocalParaTx(paraTitle, paraHeight, tx, execOK, isDel)
		if err != nil {
			clog.Error("updateLocalParaTxs", "paraTitle", paraTitle,
				"paraHeight", paraHeight, "txIndex", i, "txHash", hex.EncodeToString(tx.Hash()),
				"execOK", execOK, "isDel", isDel, "err", err)
			return nil, err
		}

		dbSet.KV = append(dbSet.KV, set.KV...)
	}

	return dbSet, nil
}

func (c *Paracross) updateLocalParaTx(paraTitle string, paraHeight int64,
	crossTx *types.Transaction, success, isDel bool) (*types.LocalDBSet, error) {
	var set types.LocalDBSet

	txHash := hex.EncodeToString(crossTx.Hash())
	var payload pt.ParacrossAction
	err := types.Decode(crossTx.Payload, &payload)
	if err != nil {
		clog.Crit("updateLocalParaTx Decode Tx failed", "para title", paraTitle,
			"para height", paraHeight, "error", err, "txHash", txHash)
		return nil, err
	}
	if payload.Ty == pt.ParacrossActionCrossAssetTransfer {
		act, err := getCrossAction(payload.GetCrossAssetTransfer(), string(crossTx.Execer))
		if err != nil {
			clog.Crit("updateLocalParaTx getCrossAction failed", "error", err)
			return nil, err
		}
		//主链共识后，平行链执行出错的主链资产transfer回滚
		if act == pt.ParacrossMainAssetTransfer || act == pt.ParacrossParaAssetWithdraw {
			kv, err := c.updateLocalAssetTransfer(crossTx, paraHeight, success, isDel)
			if err != nil {
				return nil, err
			}
			set.KV = append(set.KV, kv)
		}
		//主链共识后，平行链执行出错的平行链资产withdraw回滚
		if act == pt.ParacrossMainAssetWithdraw || act == pt.ParacrossParaAssetTransfer {
			asset, err := c.getCrossAssetTransferInfo(payload.GetCrossAssetTransfer(), crossTx, act)
			if err != nil {
				return nil, err
			}
			kv, err := c.initLocalAssetTransferDone(crossTx, asset, paraHeight, success, isDel)
			if err != nil {
				return nil, err
			}
			set.KV = append(set.KV, kv)
		}
	}

	if payload.Ty == pt.ParacrossActionAssetTransfer {
		kv, err := c.updateLocalAssetTransfer(crossTx, paraHeight, success, isDel)
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kv)
	} else if payload.Ty == pt.ParacrossActionAssetWithdraw {
		asset, err := c.getAssetTransferInfo(crossTx, payload.GetAssetWithdraw().Cointoken, true)
		if err != nil {
			return nil, err
		}
		kv, err := c.initLocalAssetTransferDone(crossTx, asset, paraHeight, success, isDel)
		if err != nil {
			return nil, err
		}
		set.KV = append(set.KV, kv)
	}

	return &set, nil
}

func (c *Paracross) getAssetTransferInfo(tx *types.Transaction, coinToken string, isWithdraw bool) (*pt.ParacrossAsset, error) {
	exec := c.GetAPI().GetConfig().GetCoinExec()
	symbol := types.BTY
	if coinToken != "" {
		exec = "token"
		symbol = coinToken
	}
	amount, err := tx.Amount()
	if err != nil {
		return nil, err
	}

	asset := &pt.ParacrossAsset{
		From:       tx.From(),
		To:         tx.To,
		Amount:     amount,
		IsWithdraw: isWithdraw,
		TxHash:     common.ToHex(tx.Hash()),
		Height:     c.GetHeight(),
		Exec:       exec,
		Symbol:     symbol,
	}
	return asset, nil
}

func (c *Paracross) getCrossAssetTransferInfo(payload *pt.CrossAssetTransfer, tx *types.Transaction, act int64) (*pt.ParacrossAsset, error) {
	exec := payload.AssetExec
	symbol := payload.AssetSymbol
	if payload.AssetSymbol == "" {
		symbol = types.BTY
		exec = c.GetAPI().GetConfig().GetCoinExec()
	}

	amount, err := tx.Amount()
	if err != nil {
		return nil, err
	}

	isWithDraw := true
	if act == pt.ParacrossMainAssetTransfer || act == pt.ParacrossParaAssetTransfer {
		isWithDraw = false
	}

	asset := &pt.ParacrossAsset{
		From:       tx.From(),
		To:         tx.To,
		IsWithdraw: isWithDraw,
		Amount:     amount,
		TxHash:     common.ToHex(tx.Hash()),
		Height:     c.GetHeight(),
		Exec:       exec,
		Symbol:     symbol,
	}
	return asset, nil
}

func (c *Paracross) initLocalAssetTransfer(tx *types.Transaction, isDel bool, asset *pt.ParacrossAsset) (*types.KeyValue, error) {
	clog.Debug("para execLocal", "tx hash", hex.EncodeToString(tx.Hash()), "action name", log.Lazy{Fn: tx.ActionName})
	key := calcLocalAssetKey(tx.Hash())
	if isDel {
		c.GetLocalDB().Set(key, nil)
		return &types.KeyValue{Key: key, Value: nil}, nil
	}

	err := c.GetLocalDB().Set(key, types.Encode(asset))
	if err != nil {
		clog.Error("para execLocal", "set", hex.EncodeToString(tx.Hash()), "failed", err)
	}
	return &types.KeyValue{Key: key, Value: types.Encode(asset)}, nil
}

func (c *Paracross) initLocalAssetTransferDone(tx *types.Transaction, asset *pt.ParacrossAsset, paraHeight int64, success, isDel bool) (*types.KeyValue, error) {
	key := calcLocalAssetKey(tx.Hash())
	if isDel {
		c.GetLocalDB().Set(key, nil)
		return &types.KeyValue{Key: key, Value: nil}, nil
	}

	asset.ParaHeight = paraHeight
	asset.CommitDoneHeight = c.GetHeight()
	asset.Success = success

	err := c.GetLocalDB().Set(key, types.Encode(asset))
	if err != nil {
		clog.Error("para execLocal", "set", "", "failed", err)
	}
	return &types.KeyValue{Key: key, Value: types.Encode(asset)}, nil
}

func (c *Paracross) updateLocalAssetTransfer(tx *types.Transaction, paraHeight int64, success, isDel bool) (*types.KeyValue, error) {
	clog.Debug("para execLocal", "tx hash", hex.EncodeToString(tx.Hash()))
	key := calcLocalAssetKey(tx.Hash())

	var asset pt.ParacrossAsset
	v, err := c.GetLocalDB().Get(key)
	if err != nil {
		return nil, err
	}
	err = types.Decode(v, &asset)
	if err != nil {
		panic(err)
	}
	if !isDel {
		asset.ParaHeight = paraHeight

		asset.CommitDoneHeight = c.GetHeight()
		asset.Success = success
	} else {
		asset.ParaHeight = 0
		asset.CommitDoneHeight = 0
		asset.Success = false
	}
	c.GetLocalDB().Set(key, types.Encode(&asset))
	return &types.KeyValue{Key: key, Value: types.Encode(&asset)}, nil
}

//IsFriend call exec is same seariase exec
func (c *Paracross) IsFriend(myexec, writekey []byte, tx *types.Transaction) bool {
	//不允许平行链
	cfg := c.GetAPI().GetConfig()
	if cfg.IsPara() {
		return false
	}
	//friend 调用必须是自己在调用
	if string(myexec) != c.GetDriverName() {
		return false
	}
	//只允许同系列的执行器（tx 也必须是 paracross）
	if string(types.GetRealExecName(tx.Execer)) != c.GetDriverName() {
		return false
	}
	//只允许跨链交易
	return c.allow(tx, 0) == nil
}

func (c *Paracross) allow(tx *types.Transaction, index int) error {
	// 增加新的规则: 在主链执行器带着title的 asset-transfer/asset-withdraw 交易允许执行
	// 1. user.p.${tilte}.${paraX}
	// 1. payload 的 actionType = t/w
	cfg := c.GetAPI().GetConfig()
	if !cfg.IsPara() && c.allowIsParaTx(tx.Execer) {
		var payload pt.ParacrossAction
		err := types.Decode(tx.Payload, &payload)
		if err != nil {
			return err
		}
		if payload.Ty == pt.ParacrossActionAssetTransfer || payload.Ty == pt.ParacrossActionAssetWithdraw {
			return nil
		}
		//对一些跨链的新feature，主链分叉之前不允许执行，但会走none执行器，因为分叉之前的版本就是走none执行器
		//然后会被过滤到平行链，最好在分叉前控制主链不发送相关交易，要么设置系统统一的fork比如ForkRootHash，主链和平行链都可以阻止执行，
		if cfg.IsDappFork(c.GetHeight(), pt.ParaX, pt.ForkCommitTx) {
			if payload.Ty == pt.ParacrossActionCommit || payload.Ty == pt.ParacrossActionNodeConfig ||
				payload.Ty == pt.ParacrossActionNodeGroupApply {
				return nil
			}
		}
		if cfg.IsDappFork(c.GetHeight(), pt.ParaX, pt.ForkParaAssetTransferRbk) {
			if payload.Ty == pt.ParacrossActionCrossAssetTransfer {
				return nil
			}
		}
		if cfg.IsDappFork(c.GetHeight(), pt.ParaX, pt.ForkParaSupervision) {
			if payload.Ty == pt.ParacrossActionSupervisionNodeConfig {
				return nil
			}
		}
	}
	return types.ErrNotAllow
}

// Allow add paracross allow rule
func (c *Paracross) Allow(tx *types.Transaction, index int) error {
	//默认规则
	err := c.DriverBase.Allow(tx, index)
	if err == nil {
		return nil
	}
	//paracross 添加的规则
	return c.allow(tx, index)
}

func (c *Paracross) allowIsParaTx(execer []byte) bool {
	if !bytes.HasPrefix(execer, types.ParaKey) {
		return false
	}
	count := 0
	index := 0
	s := len(types.ParaKey)
	for i := s; i < len(execer); i++ {
		if execer[i] == '.' {
			count++
			index = i
		}
	}
	if count == 1 && c.AllowIsSame(execer[index+1:]) {
		return true
	}
	return false
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Paracross) CheckReceiptExecOk() bool {
	return true
}
