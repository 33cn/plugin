// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"

	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/db"
	coins "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	token "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/pkg/errors"
)

const SYMBOL_BTY = "bty"

//平行链向主链transfer,先在平行链处理，共识后再在主链铸造
//主链的token转移到user.p.bb.平行链，在平行链上表示为mavl-paracross-token.symbol:
//主链的平行链转移进来的token在主链表示为mavl-paracross-user.p.aa.token.symbol 转移到user.p.bb.平行链，在平行链bb上表示mavl-paracross-paracross.user.p.aa.token.symbol

//在用户看来，平行链的资产是user.p.test.ccny, 主链资产是coins.bty, 平行链往主链转还是主链往平行链转都是这个名字
//如果主链上user.p.test.coins.ccny往user.p.test.平行链转，就是withdraw流程，如果往另一个平行链user.p.xx.转，则是转移流程
//主链转移场景： type=0,tx.exec:user.p.test.
//1. 主链本币转移：	exec:coins/token  symbol:{coins/token}.bty/cny or bty/cny,
//   平行链资产：		paracross-coins.bty
//2. 主链外币转移: 	exec:paracross, symbol: user.p.para.coins.ccny,
//   平行链资产:		paracross-paracross.user.p.para.coins.ccny
//3. 平行链本币提回: 　exec:coins, symbol: user.p.test.coins.ccny
//   平行链资产：		paracross账户coins.ccny资产释放
//平行链转移场景：type=1,tx.exec:user.p.test.
//1.　平行链本币转移:	exec:coins/token, symbol:user.p.test.{coins/token}.ccny
//    主链产生资产:	paracross-user.p.test.{coins}.ccny
//2.  主链外币提取:	exec:paracross, symbol: user.p.para.coins.ccny
//    主链恢复外币资产:	user.p.test.paracross地址释放user.p.para.coins.ccny
//3.  主链本币提取:	exec:coins/token, symbol: coins.bty
//    主链恢复本币资产:	user.p.test.paracross地址释放coin.bty
func getCrossAction(transfer *pt.CrossAssetTransfer, txExecer string) (int64, error) {
	paraTitle, ok := types.GetParaExecTitleName(txExecer)
	if !ok {
		return pt.ParacrossNoneTransfer, errors.Wrapf(types.ErrInvalidParam, "getCrossAction wrong execer:%s", txExecer)
	}
	//主链向平行链转移, 转移主链资产(包括主链本币和平行链转移进来的外币)或平行链资产withdraw
	if transfer.Type == 0 {
		// same prefix for paraChain and Symbol
		if strings.Contains(transfer.AssetSymbol, paraTitle) {
			return pt.ParacrossParaWithdraw, nil
		}
		// different paraChain symbol or mainChain symbol -> main asset transfer
		return pt.ParacrossMainTransfer, nil
	}

	//从平行链向主链转移，平行链资产转移或者主链资产withdraw
	//symbol和paraChain　prefix一致，或者symbol没有"."　-> para asset transfer
	if strings.Contains(transfer.AssetSymbol, paraTitle) {
		return pt.ParacrossParaTransfer, nil
	}
	// different paraChain symbol or mainChain symbol or null symbol -> main asset withdraw
	return pt.ParacrossMainWithdraw, nil
}

//自动补充一些参数，比如paracross执行器或symbol
func formatTransfer(transfer *pt.CrossAssetTransfer, act int64) *pt.CrossAssetTransfer {
	newTransfer := *transfer
	if act == pt.ParacrossMainTransfer || act == pt.ParacrossMainWithdraw {
		//转移平行链资产到另一个平行链
		if strings.HasPrefix(transfer.AssetSymbol, types.ParaKeyX) {
			newTransfer.AssetExec = pt.ParaX
			return &newTransfer
		}
		//转移资产symbol为bty 或　token.bty场景
		if len(transfer.AssetSymbol) > 0 {
			if strings.Contains(transfer.AssetSymbol, ".") {
				elements := strings.Split(transfer.AssetSymbol, ".")
				newTransfer.AssetExec = elements[len(elements)-2]
				newTransfer.AssetSymbol = elements[len(elements)-1]
				return &newTransfer
			}
			newTransfer.AssetExec = token.TokenX
			return &newTransfer
		}
		//assetSymbol 为null
		newTransfer.AssetExec = coins.CoinsX
		newTransfer.AssetSymbol = SYMBOL_BTY
		return &newTransfer

	}

	//把user.p.{para}.ccny prefix去掉，保留ccny
	if act == pt.ParacrossParaTransfer || act == pt.ParacrossParaWithdraw {
		e := strings.Split(transfer.AssetSymbol, ".")
		newTransfer.AssetSymbol = e[len(e)-1]
		newTransfer.AssetExec = e[len(e)-2]
		//user.p.xx.ccny，没有写coins　执行器场景
		if len(e) == 4 {
			newTransfer.AssetExec = coins.CoinsX
		}
		return &newTransfer
	}

	return transfer

}

func (a *action) crossAssetTransfer(transfer *pt.CrossAssetTransfer, act int64, actTx *types.Transaction) (*types.Receipt, error) {
	newTransfer := formatTransfer(transfer, act)
	clog.Info("paracross.crossAssetTransfer", "action", act, "newExec", newTransfer.AssetExec, "newSymbol", newTransfer.AssetSymbol,
		"ori.symbol", transfer.AssetSymbol, "type", transfer.Type, "txHash", common.ToHex(a.tx.Hash()))
	switch act {
	case pt.ParacrossMainTransfer:
		return a.mainAssetTransfer(newTransfer)
	case pt.ParacrossMainWithdraw:
		return a.mainAssetWithdraw(newTransfer, actTx)
	case pt.ParacrossParaTransfer:
		return a.paraAssetTransfer(newTransfer)
	case pt.ParacrossParaWithdraw:
		return a.paraAssetWithdraw(newTransfer, actTx)
	default:
		return nil, types.ErrNotSupport
	}
}

func (a *action) mainAssetTransfer(transfer *pt.CrossAssetTransfer) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//主链处理分支
	if !isPara {
		return a.execTransfer(transfer)
	}
	return a.execCreateAsset(transfer)
}

func (a *action) mainAssetWithdraw(withdraw *pt.CrossAssetTransfer, withdrawTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//主链处理分支
	if !isPara {
		return a.execWithdraw(withdraw, withdrawTx)
	}
	return a.execDestroyAsset(withdraw)
}

func (a *action) paraAssetTransfer(transfer *pt.CrossAssetTransfer) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//平行链链处理分支
	if isPara {
		return a.execTransfer(transfer)
	}
	return a.execCreateAsset(transfer)
}

//平行链从主链提回，　先在主链处理，然后在平行链处理，　如果平行链执行失败，共识后主链再回滚
func (a *action) paraAssetWithdraw(withdraw *pt.CrossAssetTransfer, withdrawTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//平行链链处理分支
	if isPara {
		return a.execWithdraw(withdraw, withdrawTx)
	}
	return a.execDestroyAsset(withdraw)
}

func (a *action) execTransfer(transfer *pt.CrossAssetTransfer) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	accDB, err := a.createAccount(cfg, a.db, transfer.AssetExec, transfer.AssetSymbol)
	if err != nil {
		return nil, errors.Wrap(err, "execTransfer.createAccount failed")
	}

	//主链上存入toAddr为user.p.xx.paracross地址
	execAddr := address.ExecAddress(pt.ParaX)
	toAddr := address.ExecAddress(string(a.tx.Execer))
	//在平行链上存入toAddr为paracross地址
	if cfg.IsPara() {
		execAddr = address.ExecAddress(string(a.tx.Execer))
		toAddr = address.ExecAddress(pt.ParaX)
	}
	fromAcc := accDB.LoadExecAccount(a.fromaddr, execAddr)
	if fromAcc.Balance < transfer.Amount {
		return nil, errors.Wrapf(types.ErrNoBalance, "execTransfer,fromBalance=%d", fromAcc.Balance)
	}

	clog.Debug("paracross.execTransfer", "execer", string(a.tx.Execer), "assetexec", transfer.AssetExec, "symbol", transfer.AssetSymbol,
		"txHash", hex.EncodeToString(a.tx.Hash()))
	return accDB.ExecTransfer(a.fromaddr, toAddr, execAddr, transfer.Amount)
}

func (a *action) execWithdraw(withdraw *pt.CrossAssetTransfer, withdrawTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	accDB, err := a.createAccount(cfg, a.db, withdraw.AssetExec, withdraw.AssetSymbol)
	if err != nil {
		return nil, errors.Wrap(err, "execWithdraw.createAccount failed")
	}
	execAddr := address.ExecAddress(pt.ParaX)
	fromAddr := address.ExecAddress(string(withdrawTx.Execer))
	if cfg.IsPara() {
		execAddr = address.ExecAddress(string(withdrawTx.Execer))
		fromAddr = address.ExecAddress(pt.ParaX)
	}

	clog.Debug("Paracross.execWithdraw", "amount", withdraw.Amount, "from", fromAddr,
		"assetExec", withdraw.AssetExec, "symbol", withdraw.AssetSymbol, "execAddr", execAddr, "txHash", hex.EncodeToString(a.tx.Hash()))
	return accDB.ExecTransfer(fromAddr, withdraw.ToAddr, execAddr, withdraw.Amount)
}

func (a *action) execCreateAsset(transfer *pt.CrossAssetTransfer) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	paraTitle, err := getTitleFrom(a.tx.Execer)
	if err != nil {
		return nil, errors.Wrapf(err, "execCreateAsset call getTitleFrom failed,exec=%s", string(a.tx.Execer))
	}

	assetExec := transfer.AssetExec
	assetSymbol := transfer.AssetSymbol
	if assetSymbol == "" {
		assetExec = coins.CoinsX
		assetSymbol = SYMBOL_BTY
	} else if assetExec == "" {
		assetExec = token.TokenX
	}
	if !cfg.IsPara() {
		assetExec = string(paraTitle) + assetExec
	}
	paraAcc, err := NewParaAccount(cfg, string(paraTitle), assetExec, assetSymbol, a.db)

	if err != nil {
		return nil, errors.Wrapf(err, "execCreateAsset call NewParaAccount failed,exec=%s,symbol=%s", assetExec, assetSymbol)
	}
	clog.Debug("paracross.execCreateAsset", "execer", string(a.tx.Execer), "assetExec", assetExec, "symbol", assetSymbol,
		"txHash", hex.EncodeToString(a.tx.Hash()))
	return assetDepositBalance(paraAcc, transfer.ToAddr, transfer.Amount)
}

func (a *action) execDestroyAsset(withdraw *pt.CrossAssetTransfer) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	paraTitle, err := getTitleFrom(a.tx.Execer)
	if err != nil {
		return nil, errors.Wrap(err, "execDestroyAsset call getTitleFrom failed")
	}

	assetExec := withdraw.AssetExec
	assetSymbol := withdraw.AssetSymbol
	if assetSymbol == "" {
		assetExec = coins.CoinsX
		assetSymbol = SYMBOL_BTY
	} else if assetExec == "" {
		assetExec = token.TokenX
	}
	if !cfg.IsPara() {
		assetExec = string(paraTitle) + assetExec
	}
	paraAcc, err := NewParaAccount(cfg, string(paraTitle), assetExec, assetSymbol, a.db)
	if err != nil {
		return nil, errors.Wrapf(err, "execDestroyAsset call NewParaAccount failed,exec=%s,symbol=%s", assetExec, assetSymbol)
	}
	clog.Debug("paracross.execDestroyAsset", "execer", string(a.tx.Execer), "assetExec", assetExec, "symbol", assetSymbol,
		"txHash", hex.EncodeToString(a.tx.Hash()), "from", a.fromaddr, "amount", withdraw.Amount)
	return assetWithdrawBalance(paraAcc, a.fromaddr, withdraw.Amount)
}

//旧的接口，只有主链向平行链转移
func (a *action) assetTransfer(transfer *types.AssetsTransfer) (*types.Receipt, error) {
	tr := &pt.CrossAssetTransfer{
		AssetSymbol: transfer.Cointoken,
		Amount:      transfer.Amount,
		Note:        string(transfer.Note),
		ToAddr:      transfer.To,
	}
	return a.mainAssetTransfer(tr)
}

func (a *action) assetWithdraw(withdraw *types.AssetsWithdraw, withdrawTx *types.Transaction) (*types.Receipt, error) {
	tr := &pt.CrossAssetTransfer{
		AssetExec:   withdraw.ExecName,
		AssetSymbol: withdraw.Cointoken,
		Amount:      withdraw.Amount,
		Note:        string(withdraw.Note),
		ToAddr:      withdraw.To,
	}
	//旧的只有主链向平行链转移和withdraw操作，如果cointoken非空，执行器就是token，不会是其他的
	if withdraw.Cointoken != "" {
		tr.AssetExec = token.TokenX
	}
	return a.mainAssetWithdraw(tr, withdrawTx)
}

func (a *action) assetTransferRollback(tr *pt.CrossAssetTransfer, transferTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//主链处理分支
	if !isPara {
		transfer := formatTransfer(tr, pt.ParacrossMainTransfer)
		accDB, err := a.createAccount(cfg, a.db, transfer.AssetExec, transfer.AssetSymbol)
		if err != nil {
			return nil, errors.Wrap(err, "assetTransferRollback.createAccount failed")
		}
		execAddr := address.ExecAddress(pt.ParaX)
		fromAcc := address.ExecAddress(string(transferTx.Execer))
		clog.Debug("paracross.AssetTransferRbk ", "execer", string(transferTx.Execer),
			"transfer.txHash", hex.EncodeToString(transferTx.Hash()), "curTx", hex.EncodeToString(a.tx.Hash()))
		return accDB.ExecTransfer(fromAcc, transferTx.From(), execAddr, transfer.Amount)
	}
	return nil, nil
}

//平行链从主链withdraw在平行链执行失败，主链恢复数据，主链执行
func (a *action) paraAssetWithdrawRollback(wtw *pt.CrossAssetTransfer, withdrawTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//主链处理分支
	if !isPara {
		withdraw := formatTransfer(wtw, pt.ParacrossParaWithdraw)
		paraTitle, err := getTitleFrom(a.tx.Execer)
		if err != nil {
			return nil, errors.Wrap(err, "paraAssetWithdrawRollback call getTitleFrom failed")
		}
		var paraAcc *account.DB
		paraAcc, err = NewParaAccount(cfg, string(paraTitle), string(paraTitle)+withdraw.AssetExec, withdraw.AssetSymbol, a.db)
		if err != nil {
			return nil, errors.Wrap(err, "paraAssetWithdrawRollback call NewParaAccount failed")
		}
		clog.Debug("paracross.paraAssetWithdrawRollback", "execer", string(a.tx.Execer), "txHash", hex.EncodeToString(a.tx.Hash()))
		return assetDepositBalance(paraAcc, withdrawTx.From(), withdraw.Amount)
	}
	return nil, nil
}

func (a *action) createAccount(cfg *types.Chain33Config, db db.KV, exec, symbol string) (*account.DB, error) {
	var accDB *account.DB
	if symbol == "" {
		accDB = account.NewCoinsAccount(cfg)
		accDB.SetDB(db)
		return accDB, nil
	}
	if exec == "" {
		exec = token.TokenX
	}
	return account.NewAccountDB(cfg, exec, symbol, db)
}
