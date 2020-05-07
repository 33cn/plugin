package executor

import (
	"strconv"

	"github.com/33cn/chain33/types"
	x2ethereumtypes "github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (x *x2ethereum) ExecLocal_Eth2Chain33(payload *x2ethereumtypes.Eth2Chain33, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := x.execLocal(receiptData)
	if err != nil {
		return set, err
	}
	return x.addAutoRollBack(tx, set.KV), nil
}

func (x *x2ethereum) ExecLocal_WithdrawEth(payload *x2ethereumtypes.Eth2Chain33, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := x.execLocal(receiptData)
	if err != nil {
		return set, err
	}
	return x.addAutoRollBack(tx, set.KV), nil
}

func (x *x2ethereum) ExecLocal_WithdrawChain33(payload *x2ethereumtypes.Chain33ToEth, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := x.execLocal(receiptData)
	if err != nil {
		return set, err
	}
	return x.addAutoRollBack(tx, set.KV), nil
}

func (x *x2ethereum) ExecLocal_Chain33ToEth(payload *x2ethereumtypes.Chain33ToEth, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := x.execLocal(receiptData)
	if err != nil {
		return set, err
	}
	return x.addAutoRollBack(tx, set.KV), nil
}

func (x *x2ethereum) ExecLocal_AddValidator(payload *x2ethereumtypes.MsgValidator, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return x.addAutoRollBack(tx, dbSet.KV), nil
}

func (x *x2ethereum) ExecLocal_RemoveValidator(payload *x2ethereumtypes.MsgValidator, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return x.addAutoRollBack(tx, dbSet.KV), nil
}

func (x *x2ethereum) ExecLocal_ModifyPower(payload *x2ethereumtypes.MsgValidator, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return x.addAutoRollBack(tx, dbSet.KV), nil
}

func (x *x2ethereum) ExecLocal_SetConsensusThreshold(payload *x2ethereumtypes.MsgConsensusThreshold, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
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
		case x2ethereumtypes.TyEth2Chain33Log:
			var receiptEth2Chain33 x2ethereumtypes.ReceiptEth2Chain33
			err := types.Decode(log.Log, &receiptEth2Chain33)
			if err != nil {
				return nil, err
			}

			nb, err := x.GetLocalDB().Get(x2ethereumtypes.CalTokenSymbolTotalLockOrBurnAmount(receiptEth2Chain33.LocalCoinSymbol, receiptEth2Chain33.TokenContractAddress, x2ethereumtypes.DirEth2Chain33, "lock"))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var now x2ethereumtypes.ReceiptQuerySymbolAssetsByTxType
			err = types.Decode(nb, &now)
			if err != nil {
				return nil, err
			}
			preAmount, _ := strconv.ParseFloat(x2ethereumtypes.TrimZeroAndDot(now.TotalAmount), 64)
			nowAmount, _ := strconv.ParseFloat(x2ethereumtypes.TrimZeroAndDot(receiptEth2Chain33.Amount), 64)
			TokenAssetsByTxTypeBytes := types.Encode(&x2ethereumtypes.ReceiptQuerySymbolAssetsByTxType{
				TokenSymbol: receiptEth2Chain33.LocalCoinSymbol,
				TxType:      "lock",
				TotalAmount: strconv.FormatFloat(preAmount+nowAmount, 'f', 4, 64),
				Direction:   1,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2ethereumtypes.CalTokenSymbolTotalLockOrBurnAmount(receiptEth2Chain33.LocalCoinSymbol, receiptEth2Chain33.TokenContractAddress, x2ethereumtypes.DirEth2Chain33, "lock"),
				Value: TokenAssetsByTxTypeBytes,
			})

			nb, err = x.GetLocalDB().Get(x2ethereumtypes.CalTokenSymbolToTokenAddress(receiptEth2Chain33.LocalCoinSymbol))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var t x2ethereumtypes.ReceiptTokenToTokenAddress
			err = types.Decode(nb, &t)
			if err != nil {
				return nil, err
			}
			var exist bool
			for _, addr := range t.TokenAddress {
				if addr == receiptEth2Chain33.TokenContractAddress {
					exist = true
				}
			}
			if !exist {
				t.TokenAddress = append(t.TokenAddress, receiptEth2Chain33.TokenContractAddress)
			}
			TokenToTokenAddressBytes := types.Encode(&x2ethereumtypes.ReceiptTokenToTokenAddress{
				TokenAddress: t.TokenAddress,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2ethereumtypes.CalTokenSymbolToTokenAddress(receiptEth2Chain33.LocalCoinSymbol),
				Value: TokenToTokenAddressBytes,
			})
		case x2ethereumtypes.TyWithdrawEthLog:
			var receiptEth2Chain33 x2ethereumtypes.ReceiptEth2Chain33
			err := types.Decode(log.Log, &receiptEth2Chain33)
			if err != nil {
				return nil, err
			}

			nb, err := x.GetLocalDB().Get(x2ethereumtypes.CalTokenSymbolTotalLockOrBurnAmount(receiptEth2Chain33.LocalCoinSymbol, receiptEth2Chain33.TokenContractAddress, x2ethereumtypes.DirEth2Chain33, "withdraw"))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var now x2ethereumtypes.ReceiptQuerySymbolAssetsByTxType
			err = types.Decode(nb, &now)
			if err != nil {
				return nil, err
			}

			preAmount, _ := strconv.ParseFloat(x2ethereumtypes.TrimZeroAndDot(now.TotalAmount), 64)
			nowAmount, _ := strconv.ParseFloat(x2ethereumtypes.TrimZeroAndDot(receiptEth2Chain33.Amount), 64)
			TokenAssetsByTxTypeBytes := types.Encode(&x2ethereumtypes.ReceiptQuerySymbolAssetsByTxType{
				TokenSymbol: receiptEth2Chain33.LocalCoinSymbol,
				TxType:      "withdraw",
				TotalAmount: strconv.FormatFloat(preAmount+nowAmount, 'f', 4, 64),
				Direction:   2,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2ethereumtypes.CalTokenSymbolTotalLockOrBurnAmount(receiptEth2Chain33.LocalCoinSymbol, receiptEth2Chain33.TokenContractAddress, x2ethereumtypes.DirEth2Chain33, "withdraw"),
				Value: TokenAssetsByTxTypeBytes,
			})
		case x2ethereumtypes.TyChain33ToEthLog:
			var receiptChain33ToEth x2ethereumtypes.ReceiptChain33ToEth
			err := types.Decode(log.Log, &receiptChain33ToEth)
			if err != nil {
				return nil, err
			}

			nb, err := x.GetLocalDB().Get(x2ethereumtypes.CalTokenSymbolTotalLockOrBurnAmount(receiptChain33ToEth.EthSymbol, receiptChain33ToEth.TokenContract, x2ethereumtypes.DirChain33ToEth, "lock"))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var now x2ethereumtypes.ReceiptQuerySymbolAssetsByTxType
			err = types.Decode(nb, &now)
			if err != nil {
				return nil, err
			}

			preAmount, _ := strconv.ParseFloat(x2ethereumtypes.TrimZeroAndDot(now.TotalAmount), 64)
			nowAmount, _ := strconv.ParseFloat(x2ethereumtypes.TrimZeroAndDot(receiptChain33ToEth.Amount), 64)
			TokenAssetsByTxTypeBytes := types.Encode(&x2ethereumtypes.ReceiptQuerySymbolAssetsByTxType{
				TokenSymbol: receiptChain33ToEth.EthSymbol,
				TxType:      "lock",
				TotalAmount: strconv.FormatFloat(preAmount+nowAmount, 'f', 4, 64),
				Direction:   1,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2ethereumtypes.CalTokenSymbolTotalLockOrBurnAmount(receiptChain33ToEth.EthSymbol, receiptChain33ToEth.TokenContract, x2ethereumtypes.DirChain33ToEth, "lock"),
				Value: TokenAssetsByTxTypeBytes,
			})
		case x2ethereumtypes.TyWithdrawChain33Log:
			var receiptChain33ToEth x2ethereumtypes.ReceiptChain33ToEth
			err := types.Decode(log.Log, &receiptChain33ToEth)
			if err != nil {
				return nil, err
			}

			nb, err := x.GetLocalDB().Get(x2ethereumtypes.CalTokenSymbolTotalLockOrBurnAmount(receiptChain33ToEth.EthSymbol, receiptChain33ToEth.TokenContract, x2ethereumtypes.DirChain33ToEth, "withdraw"))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			var now x2ethereumtypes.ReceiptQuerySymbolAssetsByTxType
			err = types.Decode(nb, &now)
			if err != nil {
				return nil, err
			}

			preAmount, _ := strconv.ParseFloat(x2ethereumtypes.TrimZeroAndDot(now.TotalAmount), 64)
			nowAmount, _ := strconv.ParseFloat(x2ethereumtypes.TrimZeroAndDot(receiptChain33ToEth.Amount), 64)
			TokenAssetsByTxTypeBytes := types.Encode(&x2ethereumtypes.ReceiptQuerySymbolAssetsByTxType{
				TokenSymbol: receiptChain33ToEth.EthSymbol,
				TxType:      "withdraw",
				TotalAmount: strconv.FormatFloat(preAmount+nowAmount, 'f', 4, 64),
				Direction:   2,
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   x2ethereumtypes.CalTokenSymbolTotalLockOrBurnAmount(receiptChain33ToEth.EthSymbol, receiptChain33ToEth.TokenContract, x2ethereumtypes.DirChain33ToEth, "withdraw"),
				Value: TokenAssetsByTxTypeBytes,
			})
		default:
			continue
		}
	}
	return dbSet, nil
}
