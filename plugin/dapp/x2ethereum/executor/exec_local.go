package executor

import (
	"strconv"

	"github.com/33cn/chain33/types"
	x2eTy "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (x *x2ethereum) ExecLocal_Eth2Chain33Lock(payload *x2eTy.Eth2Chain33, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := x.execLocal(receiptData)
	if err != nil {
		return set, err
	}
	return x.addAutoRollBack(tx, set.KV), nil
}

func (x *x2ethereum) ExecLocal_Eth2Chain33Burn(payload *x2eTy.Eth2Chain33, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := x.execLocal(receiptData)
	if err != nil {
		return set, err
	}
	return x.addAutoRollBack(tx, set.KV), nil
}

func (x *x2ethereum) ExecLocal_Chain33ToEthBurn(payload *x2eTy.Chain33ToEth, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := x.execLocal(receiptData)
	if err != nil {
		return set, err
	}
	return x.addAutoRollBack(tx, set.KV), nil
}

func (x *x2ethereum) ExecLocal_Chain33ToEthLock(payload *x2eTy.Chain33ToEth, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := x.execLocal(receiptData)
	if err != nil {
		return set, err
	}
	return x.addAutoRollBack(tx, set.KV), nil
}

func (x *x2ethereum) ExecLocal_AddValidator(payload *x2eTy.MsgValidator, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return x.addAutoRollBack(tx, dbSet.KV), nil
}

func (x *x2ethereum) ExecLocal_RemoveValidator(payload *x2eTy.MsgValidator, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return x.addAutoRollBack(tx, dbSet.KV), nil
}

func (x *x2ethereum) ExecLocal_ModifyPower(payload *x2eTy.MsgValidator, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return x.addAutoRollBack(tx, dbSet.KV), nil
}

func (x *x2ethereum) ExecLocal_SetConsensusThreshold(payload *x2eTy.MsgConsensusThreshold, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return x.addAutoRollBack(tx, dbSet.KV), nil
}

//设置自动回滚
func (x *x2ethereum) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {
	dbSet := &types.LocalDBSet{}
	dbSet.KV = x.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

func (x *x2ethereum) execLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case x2eTy.TyEth2Chain33Log:
			var receiptEth2Chain33 x2eTy.ReceiptEth2Chain33
			err := types.Decode(log.Log, &receiptEth2Chain33)
			if err != nil {
				return nil, err
			}

			nb, err := x.GetLocalDB().Get(x2eTy.CalTokenSymbolTotalLockOrBurnAmount(receiptEth2Chain33.IssuerDotSymbol, receiptEth2Chain33.TokenAddress, x2eTy.DirEth2Chain33, "lock"))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var now x2eTy.ReceiptQuerySymbolAssetsByTxType
			err = types.Decode(nb, &now)
			if err != nil {
				return nil, err
			}
			preAmount, _ := strconv.ParseFloat(x2eTy.TrimZeroAndDot(now.TotalAmount), 64)
			nowAmount, _ := strconv.ParseFloat(x2eTy.TrimZeroAndDot(receiptEth2Chain33.Amount), 64)
			TokenAssetsByTxTypeBytes := types.Encode(&x2eTy.ReceiptQuerySymbolAssetsByTxType{
				TokenSymbol: receiptEth2Chain33.IssuerDotSymbol,
				TxType:      "lock",
				TotalAmount: strconv.FormatFloat(preAmount+nowAmount, 'f', 4, 64),
				Direction:   1,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2eTy.CalTokenSymbolTotalLockOrBurnAmount(receiptEth2Chain33.IssuerDotSymbol, receiptEth2Chain33.TokenAddress, x2eTy.DirEth2Chain33, "lock"),
				Value: TokenAssetsByTxTypeBytes,
			})

			nb, err = x.GetLocalDB().Get(x2eTy.CalTokenSymbolToTokenAddress(receiptEth2Chain33.IssuerDotSymbol))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var t x2eTy.ReceiptTokenToTokenAddress
			err = types.Decode(nb, &t)
			if err != nil {
				return nil, err
			}
			var exist bool
			for _, addr := range t.TokenAddress {
				if addr == receiptEth2Chain33.TokenAddress {
					exist = true
				}
			}
			if !exist {
				t.TokenAddress = append(t.TokenAddress, receiptEth2Chain33.TokenAddress)
			}
			TokenToTokenAddressBytes := types.Encode(&x2eTy.ReceiptTokenToTokenAddress{
				TokenAddress: t.TokenAddress,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2eTy.CalTokenSymbolToTokenAddress(receiptEth2Chain33.IssuerDotSymbol),
				Value: TokenToTokenAddressBytes,
			})
		case x2eTy.TyWithdrawEthLog:
			var receiptEth2Chain33 x2eTy.ReceiptEth2Chain33
			err := types.Decode(log.Log, &receiptEth2Chain33)
			if err != nil {
				return nil, err
			}

			nb, err := x.GetLocalDB().Get(x2eTy.CalTokenSymbolTotalLockOrBurnAmount(receiptEth2Chain33.IssuerDotSymbol, receiptEth2Chain33.TokenAddress, x2eTy.DirEth2Chain33, "withdraw"))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var now x2eTy.ReceiptQuerySymbolAssetsByTxType
			err = types.Decode(nb, &now)
			if err != nil {
				return nil, err
			}

			preAmount, _ := strconv.ParseFloat(x2eTy.TrimZeroAndDot(now.TotalAmount), 64)
			nowAmount, _ := strconv.ParseFloat(x2eTy.TrimZeroAndDot(receiptEth2Chain33.Amount), 64)
			TokenAssetsByTxTypeBytes := types.Encode(&x2eTy.ReceiptQuerySymbolAssetsByTxType{
				TokenSymbol: receiptEth2Chain33.IssuerDotSymbol,
				TxType:      "withdraw",
				TotalAmount: strconv.FormatFloat(preAmount+nowAmount, 'f', 4, 64),
				Direction:   2,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2eTy.CalTokenSymbolTotalLockOrBurnAmount(receiptEth2Chain33.IssuerDotSymbol, receiptEth2Chain33.TokenAddress, x2eTy.DirEth2Chain33, "withdraw"),
				Value: TokenAssetsByTxTypeBytes,
			})
		case x2eTy.TyChain33ToEthLog:
			var receiptChain33ToEth x2eTy.ReceiptChain33ToEth
			err := types.Decode(log.Log, &receiptChain33ToEth)
			if err != nil {
				return nil, err
			}

			nb, err := x.GetLocalDB().Get(x2eTy.CalTokenSymbolTotalLockOrBurnAmount(receiptChain33ToEth.IssuerDotSymbol, receiptChain33ToEth.TokenContract, x2eTy.DirChain33ToEth, "lock"))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var now x2eTy.ReceiptQuerySymbolAssetsByTxType
			err = types.Decode(nb, &now)
			if err != nil {
				return nil, err
			}

			preAmount, _ := strconv.ParseFloat(x2eTy.TrimZeroAndDot(now.TotalAmount), 64)
			nowAmount, _ := strconv.ParseFloat(x2eTy.TrimZeroAndDot(receiptChain33ToEth.Amount), 64)
			TokenAssetsByTxTypeBytes := types.Encode(&x2eTy.ReceiptQuerySymbolAssetsByTxType{
				TokenSymbol: receiptChain33ToEth.IssuerDotSymbol,
				TxType:      "lock",
				TotalAmount: strconv.FormatFloat(preAmount+nowAmount, 'f', 4, 64),
				Direction:   1,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2eTy.CalTokenSymbolTotalLockOrBurnAmount(receiptChain33ToEth.IssuerDotSymbol, receiptChain33ToEth.TokenContract, x2eTy.DirChain33ToEth, "lock"),
				Value: TokenAssetsByTxTypeBytes,
			})
		case x2eTy.TyWithdrawChain33Log:
			var receiptChain33ToEth x2eTy.ReceiptChain33ToEth
			err := types.Decode(log.Log, &receiptChain33ToEth)
			if err != nil {
				return nil, err
			}

			nb, err := x.GetLocalDB().Get(x2eTy.CalTokenSymbolTotalLockOrBurnAmount(receiptChain33ToEth.IssuerDotSymbol, receiptChain33ToEth.TokenContract, x2eTy.DirChain33ToEth, ""))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var now x2eTy.ReceiptQuerySymbolAssetsByTxType
			err = types.Decode(nb, &now)
			if err != nil {
				return nil, err
			}

			preAmount, _ := strconv.ParseFloat(x2eTy.TrimZeroAndDot(now.TotalAmount), 64)
			nowAmount, _ := strconv.ParseFloat(x2eTy.TrimZeroAndDot(receiptChain33ToEth.Amount), 64)
			TokenAssetsByTxTypeBytes := types.Encode(&x2eTy.ReceiptQuerySymbolAssetsByTxType{
				TokenSymbol: receiptChain33ToEth.IssuerDotSymbol,
				TxType:      "withdraw",
				TotalAmount: strconv.FormatFloat(preAmount+nowAmount, 'f', 4, 64),
				Direction:   2,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2eTy.CalTokenSymbolTotalLockOrBurnAmount(receiptChain33ToEth.IssuerDotSymbol, receiptChain33ToEth.TokenContract, x2eTy.DirChain33ToEth, "withdraw"),
				Value: TokenAssetsByTxTypeBytes,
			})
		default:
			continue
		}
	}
	return dbSet, nil
}
