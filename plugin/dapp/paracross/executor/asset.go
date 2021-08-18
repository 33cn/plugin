// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	token "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/pkg/errors"
)

//SymbolBty ...
const SymbolBty = "bty"

/*
资产　=　assetExec + assetSymbol 唯一确定一个资产

				exec              		symbol								tx.title=user.p.test1   tx.title=user.p.test2
1. 主链上的资产：
				coins					bty                     	  		transfer                 transfer
				paracross				user.p.test1.coins.fzm    			withdraw                 transfer

2. 平行链上的资产：
				user.p.test1.coins		fzm              					transfer                 NAN
    			user.p.test1.paracross	coins.bty    						withdraw                 NAN
    			user.p.test1.paracross	paracross.user.p.test2.coins.cny	withdraw                 NAN

其中user.p.test1.paracross.paracross.user.p.test2.coins.cny资产解释：
user.p.test1.paracross.是平行链paracross执行器，　paracross.user.p.test2.coins.cny的paracross代表从主链的paracross转移过来的user.p.test2.coins.cny资产
*/
func getCrossAction(transfer *pt.CrossAssetTransfer, txExecer string) (int64, error) {
	paraTitle, ok := types.GetParaExecTitleName(txExecer)
	if !ok {
		return pt.ParacrossNoneTransfer, errors.Wrapf(types.ErrInvalidParam, "asset cross transfer execer:%s should be user.p.xx", txExecer)
	}
	//平行链资产和执行器不一致
	if types.IsParaExecName(transfer.AssetExec) && !strings.Contains(transfer.AssetExec, paraTitle) {
		return pt.ParacrossNoneTransfer, errors.Wrapf(types.ErrInvalidParam, "asset execer=%s not belong to title=%s", transfer.AssetExec, paraTitle)
	}

	//paracross执行器的资产转移都是withdraw回去，除了主链平行链资产转移到另一个平行链
	if types.IsParaExecName(transfer.AssetExec) {
		if strings.Contains(transfer.AssetExec, pt.ParaX) {
			return pt.ParacrossMainAssetWithdraw, nil
		}
		return pt.ParacrossParaAssetTransfer, nil
	}

	if strings.Contains(transfer.AssetExec, pt.ParaX) && strings.Contains(transfer.AssetSymbol, paraTitle) {
		return pt.ParacrossParaAssetWithdraw, nil
	}
	return pt.ParacrossMainAssetTransfer, nil

}

/*
修正原生执行器名字
								      								type			realExec    realSymbol
coins+bty															mainTransfer	coins		bty
paracross+user.p.test1.coins.bty									paraWithdraw	coins		bty
user.p.test1.coins+bty												paraTransfer    coins		bty
user.p.test1.paracross+coins.bty									mainWithdraw	coins		bty
paracross+user.p.test1.coins.bty(->user.p.test2)					mainTransfer 	paracross   user.p.test1.coins.bty
user.p.test2.paracross+paracross.user.p.test1.coins.bty 			mainWithdraw	paracross   user.p.test1.coins.bty
注意:
1. user.p.test1.coins+bty只是对外表示平行链资产，真正执行器也是coins，因为account模型的mavl-coins-bty-　在主链和平行链都一样，平行链模型并不是mavl-user.p.test.coins-bty-
2. paracross执行器下的资产都是外来资产，在withdraw时候，真正的原生执行器是在symbol里面
　　a. 销毁资产　mavl-paracross-coins.bty-exec-addr(user)
　　b. 恢复资产　mavl-coins-bty-exec-addr{paracross}:addr{user}, 在原生coins执行器上恢复资产
*/
func amendTransferParam(transfer *pt.CrossAssetTransfer, act int64) (*pt.CrossAssetTransfer, error) {
	newTransfer := *transfer
	//exec=user.p.test1.coins -> exec=coins
	if types.IsParaExecName(transfer.AssetExec) {
		elements := strings.Split(transfer.AssetExec, ".")
		newTransfer.AssetExec = elements[len(elements)-1]
	}

	//paracross　exec's symbol should contain ".", non-paracross exec should not contain "."
	if newTransfer.AssetExec == pt.ParaX && !strings.Contains(newTransfer.AssetSymbol, ".") {
		return nil, errors.Wrapf(types.ErrInvalidParam, "paracross exec=%s, the symbol=%s should contain '.'", newTransfer.AssetExec, transfer.AssetSymbol)
	}

	if newTransfer.AssetExec != pt.ParaX && strings.Contains(newTransfer.AssetSymbol, ".") {
		return nil, errors.Wrapf(types.ErrInvalidParam, "non-paracross exec=%s, symbol=%s should not contain '.'", newTransfer.AssetExec, transfer.AssetSymbol)
	}

	if act == pt.ParacrossMainAssetWithdraw {
		e := strings.Split(transfer.AssetSymbol, ".")
		if len(e) <= 1 {
			return nil, errors.Wrapf(types.ErrInvalidParam, "main asset withdraw symbol=%s should be exec.symbol", transfer.AssetSymbol)
		}
		newTransfer.AssetExec = e[0]
		newTransfer.AssetSymbol = strings.Join(e[1:], ".")
		return &newTransfer, nil
	}

	//把user.p.{para}.coins.ccny prefix去掉，保留coins.ccny
	if act == pt.ParacrossParaAssetWithdraw {
		e := strings.Split(transfer.AssetSymbol, ".")
		if len(e) <= 1 {
			return nil, errors.Wrapf(types.ErrInvalidParam, "para asset withdraw symbol=%s should be exec.symbol", transfer.AssetSymbol)
		}
		newTransfer.AssetSymbol = e[len(e)-1]
		newTransfer.AssetExec = e[len(e)-2]
		return &newTransfer, nil
	}
	return &newTransfer, nil
}

func (a *action) crossAssetTransfer(transfer *pt.CrossAssetTransfer, act int64, actTx *types.Transaction) (*types.Receipt, error) {
	newTransfer, err := amendTransferParam(transfer, act)
	if err != nil {
		return nil, err
	}
	clog.Info("paracross.crossAssetTransfer", "action", act, "newExec", newTransfer.AssetExec, "newSymbol", newTransfer.AssetSymbol,
		"ori.exec", transfer.AssetExec, "ori.symbol", transfer.AssetSymbol, "txHash", common.ToHex(actTx.Hash()))
	switch act {
	case pt.ParacrossMainAssetTransfer:
		return a.mainAssetTransfer(newTransfer, actTx)
	case pt.ParacrossMainAssetWithdraw:
		return a.mainAssetWithdraw(newTransfer, actTx)
	case pt.ParacrossParaAssetTransfer:
		return a.paraAssetTransfer(newTransfer, actTx)
	case pt.ParacrossParaAssetWithdraw:
		return a.paraAssetWithdraw(newTransfer, actTx)
	default:
		return nil, types.ErrNotSupport
	}
}

//主链先transfer, 然后平行链create　asset, 如果平行链失败，主链再rollback
func (a *action) mainAssetTransfer(transfer *pt.CrossAssetTransfer, transferTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//主链处理分支, 先处理
	if !isPara {
		return a.execTransfer(transfer, transferTx)
	}
	return a.execCreateAsset(transfer, transferTx)
}

//平行链先销毁，　共识后主链再withdraw
func (a *action) mainAssetWithdraw(withdraw *pt.CrossAssetTransfer, withdrawTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//主链处理分支，共识后处理，a.tx是共识交易
	if !isPara {
		return a.execWithdraw(withdraw, withdrawTx)
	}
	return a.execDestroyAsset(withdraw, withdrawTx)
}

//平行链先转移，　共识后主链create asset
func (a *action) paraAssetTransfer(transfer *pt.CrossAssetTransfer, transferTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//平行链链处理分支，先处理
	if isPara {
		return a.execTransfer(transfer, transferTx)
	}
	//主链共识后处理
	return a.execCreateAsset(transfer, transferTx)
}

//平行链从主链提回，　先在主链处理，然后在平行链处理，　如果平行链执行失败，共识后主链再回滚
func (a *action) paraAssetWithdraw(withdraw *pt.CrossAssetTransfer, withdrawTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//平行链链处理分支，后处理
	if isPara {
		return a.execWithdraw(withdraw, withdrawTx)
	}
	return a.execDestroyAsset(withdraw, withdrawTx)
}

func (a *action) execTransfer(transfer *pt.CrossAssetTransfer, transferTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	accDB, err := a.createAccount(cfg, a.db, transfer.AssetExec, transfer.AssetSymbol)
	if err != nil {
		return nil, errors.Wrapf(err, "execTransfer.createAccount,exec=%s,symbol=%s", transfer.AssetExec, transfer.AssetSymbol)
	}

	//主链上存入toAddr为user.p.xx.paracross地址
	execAddr := address.ExecAddress(pt.ParaX)
	toAddr := address.ExecAddress(string(transferTx.Execer))
	//在平行链上存入toAddr为paracross地址
	if cfg.IsPara() {
		execAddr = address.ExecAddress(string(transferTx.Execer))
		toAddr = address.ExecAddress(pt.ParaX)
	}

	clog.Debug("paracross.execTransfer", "execer", string(transferTx.Execer), "assetexec", transfer.AssetExec, "symbol", transfer.AssetSymbol,
		"txHash", common.ToHex(transferTx.Hash()))

	//对于paracross合约下的资产直接转账，不需要通过存到paracross合约下再转账，这里只有主链的A平行链资产转移到另一个B平行链场景
	if transfer.AssetExec == pt.ParaX {
		r, err := accDB.Transfer(transferTx.From(), toAddr, transfer.Amount)
		if err != nil {
			return nil, errors.Wrapf(err, "assetTransfer,assetExec=%s,assetSym=%s", transfer.AssetExec, transfer.AssetSymbol)
		}
		return r, nil
	}

	fromAcc := accDB.LoadExecAccount(transferTx.From(), execAddr)
	if fromAcc.Balance < transfer.Amount {
		return nil, errors.Wrapf(types.ErrNoBalance, "execTransfer,acctBalance=%d,assetExec=%s,assetSym=%s", fromAcc.Balance, transfer.AssetExec, transfer.AssetSymbol)
	}
	r, err := accDB.ExecTransfer(transferTx.From(), toAddr, execAddr, transfer.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "assetTransfer,assetExec=%s,assetSym=%s", transfer.AssetExec, transfer.AssetSymbol)
	}
	return r, nil
}

//withdraw是共识交易触发的，a.tx是共识交易，　withdrawTx是最初提交的withdraw的tx
func (a *action) execWithdraw(withdraw *pt.CrossAssetTransfer, withdrawTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	accDB, err := a.createAccount(cfg, a.db, withdraw.AssetExec, withdraw.AssetSymbol)
	if err != nil {
		return nil, errors.Wrapf(err, "execWithdraw.createAccount,exec=%s,symbol=%s", withdraw.AssetExec, withdraw.AssetSymbol)
	}
	execAddr := address.ExecAddress(pt.ParaX)
	fromAddr := address.ExecAddress(string(withdrawTx.Execer))
	if cfg.IsPara() {
		execAddr = address.ExecAddress(string(withdrawTx.Execer))
		fromAddr = address.ExecAddress(pt.ParaX)
	}

	clog.Debug("Paracross.execWithdraw", "amount", withdraw.Amount, "from", fromAddr,
		"assetExec", withdraw.AssetExec, "symbol", withdraw.AssetSymbol, "execAddr", execAddr, "txHash", common.ToHex(withdrawTx.Hash()))

	//对于paracross合约下的资产直接转账，不需要通过存到paracross合约下再转账，这里只有主链的A平行链资产转移到另一个B平行链场景
	if withdraw.AssetExec == pt.ParaX {
		r, err := accDB.Transfer(fromAddr, withdraw.ToAddr, withdraw.Amount)
		if err != nil {
			return nil, errors.Wrapf(err, "assetWithdraw,assetExec=%s,assetSym=%s", withdraw.AssetExec, withdraw.AssetSymbol)
		}
		return r, nil
	}

	r, err := accDB.ExecTransfer(fromAddr, withdraw.ToAddr, execAddr, withdraw.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "assetWithdraw,assetExec=%s,assetSym=%s", withdraw.AssetExec, withdraw.AssetSymbol)
	}
	return r, nil
}

//主链Alice的token转移到user.p.bb.平行链，在平行链上表示为mavl-paracross-token.symbol-Addr(Alice),这里并没有放在Addr(user.p.bb.paracross)子账号下
//平行链转移到主链的token在主链表示为mavl-paracross-user.p.aa.token.symbol-exec-Addr(Alice)，再转移到另一个user.p.bb.平行链，需要先transfer到paracross执行器下
//在平行链bb上铸造新币，表示为mavl-paracross-paracross.user.p.aa.token.symbol-exec-Addr(Alice)，第二个paracross代表在主链原生执行器为paracross
func (a *action) createParaAccount(cross *pt.CrossAssetTransfer, crossTx *types.Transaction) (*account.DB, error) {
	cfg := a.api.GetConfig()
	paraTitle, err := getTitleFrom(crossTx.Execer)
	if err != nil {
		return nil, errors.Wrapf(err, "createParaAccount call getTitleFrom failed,exec=%s", string(crossTx.Execer))
	}

	assetExec := cross.AssetExec
	assetSymbol := cross.AssetSymbol
	if !cfg.IsPara() {
		assetExec = string(paraTitle) + assetExec
	}
	paraAcc, err := NewParaAccount(cfg, string(paraTitle), assetExec, assetSymbol, a.db)
	clog.Debug("createParaAccount", "assetExec", assetExec, "symbol", assetSymbol, "txHash", common.ToHex(crossTx.Hash()))
	if err != nil {
		return nil, errors.Wrapf(err, "createParaAccount,exec=%s,symbol=%s,title=%s", assetExec, assetSymbol, paraTitle)
	}
	return paraAcc, nil
}

func (a *action) execCreateAsset(transfer *pt.CrossAssetTransfer, transferTx *types.Transaction) (*types.Receipt, error) {
	paraAcc, err := a.createParaAccount(transfer, transferTx)
	if err != nil {
		return nil, errors.Wrapf(err, "createAsset")
	}
	clog.Debug("paracross.execCreateAsset", "assetExec", transfer.AssetExec, "symbol", transfer.AssetSymbol,
		"txHash", common.ToHex(transferTx.Hash()))

	r, err := assetDepositBalance(paraAcc, transfer.ToAddr, transfer.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "createParaAsset,assetExec=%s,assetSym=%s", transfer.AssetExec, transfer.AssetSymbol)
	}
	return r, nil
}

func (a *action) execDestroyAsset(withdraw *pt.CrossAssetTransfer, withdrawTx *types.Transaction) (*types.Receipt, error) {
	paraAcc, err := a.createParaAccount(withdraw, withdrawTx)
	if err != nil {
		return nil, errors.Wrapf(err, "destroyAsset")
	}
	clog.Debug("paracross.execDestroyAsset", "assetExec", withdraw.AssetExec, "symbol", withdraw.AssetSymbol,
		"txHash", common.ToHex(withdrawTx.Hash()), "from", withdrawTx.From(), "amount", withdraw.Amount)
	r, err := assetWithdrawBalance(paraAcc, withdrawTx.From(), withdraw.Amount)
	if err != nil {
		return nil, errors.Wrapf(err, "destroyAsset,assetExec=%s,assetSym=%s", withdraw.AssetExec, withdraw.AssetSymbol)
	}
	return r, nil
}

//旧的接口，只有主链向平行链转移
func (a *action) assetTransfer(transfer *types.AssetsTransfer) (*types.Receipt, error) {
	tr := &pt.CrossAssetTransfer{
		AssetSymbol: transfer.Cointoken,
		Amount:      transfer.Amount,
		Note:        string(transfer.Note),
		ToAddr:      transfer.To,
	}
	adaptNullAssetExec(tr)
	return a.mainAssetTransfer(tr, a.tx)
}

//旧的接口，只有主链从平行链转移
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
	adaptNullAssetExec(tr)
	return a.mainAssetWithdraw(tr, withdrawTx)
}

func (a *action) assetTransferRollback(transfer *pt.CrossAssetTransfer, transferTx *types.Transaction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	isPara := cfg.IsPara()
	//主链处理分支
	if !isPara {
		accDB, err := a.createAccount(cfg, a.db, transfer.AssetExec, transfer.AssetSymbol)
		if err != nil {
			return nil, errors.Wrap(err, "assetTransferRollback.createAccount failed")
		}
		execAddr := address.ExecAddress(pt.ParaX)
		fromAcc := address.ExecAddress(string(transferTx.Execer))
		clog.Debug("paracross.AssetTransferRbk ", "exec", transfer.AssetExec, "sym", transfer.AssetSymbol,
			"transfer.txHash", common.ToHex(transferTx.Hash()), "curTx", common.ToHex(a.tx.Hash()))
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
		withdraw, err := amendTransferParam(wtw, pt.ParacrossParaAssetWithdraw)
		if err != nil {
			return nil, errors.Wrapf(err, "paraAssetWithdrawRollback amend param")
		}
		paraAcc, err := a.createParaAccount(withdraw, withdrawTx)
		if err != nil {
			return nil, errors.Wrapf(err, "createAsset")
		}
		clog.Debug("paracross.paraAssetWithdrawRollback", "exec", withdraw.AssetExec, "sym", withdraw.AssetSymbol,
			"transfer.txHash", common.ToHex(withdrawTx.Hash()), "curTx", common.ToHex(a.tx.Hash()))
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

//对于旧的assetTransfer,assetWithdraw资产转移接口，没有填AssetExec，这里只支持coins，如果主链是coinsx执行器，则旧接口会失败，需要用新接口。
//新接口必须填AssetExec,主链和平行链从跨链交易资产参数上会保持一致，目的资产不依赖目标的toml配置文件
//比如主链是coinsx 平行链也会铸造coinsx.bty资产，withdraw到主链也是coinsx
func adaptNullAssetExec(transfer *pt.CrossAssetTransfer) {
	if transfer.AssetSymbol == "" {
		transfer.AssetExec = "coins"
		transfer.AssetSymbol = SymbolBty
		return
	}
	if transfer.AssetExec == "" {
		transfer.AssetExec = token.TokenX
	}
}
