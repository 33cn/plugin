package executor

import (
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

// 清单, 做为零知识证明和各个交易所之间沟通的桥梁
// 如1 存款/取款/手续费,
//   1. 在零知识证明存款合法后, 在树上对帐号进行存款
//   2. 零知识证明存款生成存款的清单
//   3. 根据清单, 在交易所中进行等额的存款
// 如2 现货撮合
//   1. 在现货合约中, 进行撮合, 如果进行了2次撮合即 A 用户的Tx0, 和B 用户的Tx1 , C用户的Tx2 进行了撮合 (B, C 提交交易时没有撮合, 变成的市场挂单)
//   2. 在现货合约中, 根据撮合结果, 对用户在现货合约中的帐号进行资产的交换和收取手续费
//   3. 在现货合约中, 生成资产调整的清单
//   4. 根据清单, 对零知识证明上对帐号进行调整

const (
	// 存款
	ListZkDeposit = 1
	// 提款
	ListZkWithdraw = 2
	// 现货撮合
	ListSpotMatch = 1001
)

// 相关的交易为 A签名发送的 tx1(卖bty), B签名发送的tx2(买bty)
// 撮合 100usdt 交易 1bty, 并且收取 A B 各 1usdt的手续费

// 撮合
//          A    B  feesysacc
// BTY     -1    +1
// USDT    +100   -100

// 手续费1
//          A    B  feesysacc
// BTY     0    0     0
// USDT    -1   0     +1

// 手续费1
//          A    B  feesysacc
// BTY     0     0        0
// USDT    0     -1       +1

// 结算后状态
//          A    B  feesysacc
// BTY     -1    +1
// USDT    +99   -101  +2

// 是否需要将清单合并成帐号变化
// BTY-id = 2, USDT-id = 1

// 撮合 包含 1个交换, 和两个手续费
// 币的源头是是从balance/frozen 中转 看balance 的中值是否为frozen
// 币的目的一般到 balance即可, 如果有到frozen的 提供额外的函数或参数

func GetSpotMatch(receipt *types.Receipt) *types.Receipt {
	receipt2 := &types.Receipt{Logs: []*types.ReceiptLog{}}
	for _, l := range receipt.Logs {
		if l.Ty != et.TySpotTradeLog {
			continue
		}
		receipt2.Logs = append(receipt2.Logs, l)
	}
	return receipt2
}

func (a *Action) SpotNftMatch(payload *et.SpotNftTakerOrder, list *types.Receipt) (*types.Receipt, error) {
	receipt := &types.Receipt{}
	for _, tradeRaw := range list.Logs {
		switch tradeRaw.Ty {
		case et.TySpotTradeLog:
			var trade et.ReceiptSpotTrade
			err := types.Decode(tradeRaw.Log, &trade)
			if err != nil {
				return nil, err
			}
			receipt2, err := a.SwapWithNft(payload, &trade)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, receipt2)
		default:
			//
		}
	}
	return receipt, nil
}

// A 和 B 交换 = transfer(A,B) + transfer(B,A) + swapfee()
// A 和 A 交换 = transfer(A,A) 0 + transfer(A,A) 0 + swapfee()
func (a *Action) SwapWithNft(payload1 *et.SpotNftTakerOrder, trade *et.ReceiptSpotTrade) (*types.Receipt, error) {
	return nil, nil
	/*
		var logs []*types.ReceiptLog
		var kvs []*types.KeyValue

		zklog.Info("SwapWithNft", "trade-buy", trade.MakerOrder.TokenBuy, "trade-sell", trade.MakerOrder.TokenSell, "detail", trade.Match)
		info, err := getTreeUpdateInfo(a.statedb)
		if err != nil {
			return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
		}

		operationInfo := &et.OperationInfo{
			BlockHeight: uint64(a.height),
			TxIndex:     uint32(a.index),
			TxType:      et.TySwapAction,

			TokenID:     trade.MakerOrder.TokenSell,
			Amount:      trade.MakerOrder.Amount, // nft 不需要乘以 1e18
			FeeAmount:   "0",
			SigData:     payload1.GetOrder().Signature,
			AccountID:   payload1.Order.AccountID,

			SpecialInfo: new(et.OperationSpecialInfo),
		}

		swapSpecialData := genNftSwapSpecialData(payload1, trade)
		operationInfo.SpecialInfo.SpecialDatas = append(operationInfo.SpecialInfo.SpecialDatas, swapSpecialData)

		zklog.Info("swapGenTransfer", "trade-buy", trade.MakerOrder.TokenBuy, "trade-sell", trade.MakerOrder.TokenSell)
		transfers := a.nftSwapGenTransfer(et.OpBuy, payload1.Order, trade)
		zklog.Info("swapGenTransfer", "tokenid0", transfers[0].TokenId, "tokenid1", transfers[1].TokenId)
		// operationInfo, localKvs 通过 zklog 获得
		zklog := &et.ZkReceiptLog{OperationInfo: operationInfo}
		// left asset 不在树上， 所以不用处理
		/ *
			receipt1, err := a.swapByTransfer(transfers[0], trade, info, zklog)
			if err != nil {
				return nil, err
			}
			logs = append(logs, receipt1.Logs...)
			kvs = append(kvs, receipt1.KV...)
		* /
		receipt2, err := a.swapByTransfer(transfers[1], trade, info, zklog)
		if err != nil {
			zlog.Error("swapByTransfer", "err", err)
			return nil, err
		}
		logs = append(logs, receipt2.Logs...)
		kvs = append(kvs, receipt2.KV...)

		receiptLog := &types.ReceiptLog{Ty: et.TySwapLog, Log: types.Encode(zklog)}
		logs = append(logs, receiptLog)

		feelog1, err := a.MakeFeeLog(transfers[2].AmountIn, info, transfers[2].TokenId, transfers[2].Signature)
		if err != nil {
			zlog.Error("MakeFeeLog", "err", err)
			return nil, err
		}

		receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
		receipts = mergeReceipt(receipts, feelog1)
		return receipts, nil
	*/
}

/*
func genNftSwapSpecialData(payload1 *et.SpotNftTakerOrder, trade *et.ReceiptSpotTrade) *et.OperationSpecialInfo {
	left, right := payload1.Order.TokenBuy, payload1.Order.TokenSell

	op := et.OpBuy
	sellRightFee, buyRightFee := trade.Match.FeeMaker, trade.Match.FeeTaker
	if op == et.OpBuy {
		sellRightFee, buyRightFee = trade.Match.FeeTaker, trade.Match.FeeMaker
	}

	sellLeft := &et.OrderPricePair{
		Sell: payload1.Order.Ratio1,
		Buy:  payload1.Order.Ratio2,
	}
	buyLeft := &et.OrderPricePair{
		Sell: trade.MakerOrder.Ratio1,
		Buy:  trade.MakerOrder.Ratio2,
	}

	if op == et.OpBuy {
		sellLeft, buyLeft = buyLeft, sellLeft
	}

	specialData := &et.OperationSpecialInfo{
		TokenID: []uint64{uint64(left), uint64(right)},
		Amount: []string{
			et.AmountToZksync(uint64(trade.Match.LeftBalance)),
			et.AmountToZksync(uint64(trade.Match.RightBalance)),
			et.AmountToZksync(uint64(sellRightFee)),
			et.AmountToZksync(uint64(buyRightFee)),
			payload1.Order.Amount,
			trade.MakerOrder.Amount,
		},
		AccountID:   trade.GetCurrent().Taker.Id,
		RecipientID: trade.GetCurrent().Maker.Id,
		PricePair:   []*et.OrderPricePair{sellLeft, buyLeft},
		SigData:     trade.MakerOrder.Signature,
	}

	return specialData
}
*/

/*
func (a *Action) nftSwapGenTransfer(op int32, takerOrder *et.ZkOrder, trade *et.ReceiptSpotTrade) []*et.ZkTransferWithFee {
	// 挂单/摘单模式， 不会和自己交易
	var transfers []*et.ZkTransferWithFee
	if false && trade.Current.Maker.Id == trade.Current.Taker.Id {
		return a.selfSwapGenTransfer(op, takerOrder, trade)
	}

	// A 和 B 交易
	// Sell
	takerTokenID, makerTokenID := takerOrder.TokenSell, takerOrder.TokenBuy
	feeTokenId := takerOrder.TokenBuy
	takerPay, makerRcv := trade.Match.LeftBalance, trade.Match.LeftBalance

	rightBalance := trade.Match.RightBalance
	takerRcv := rightBalance - trade.Match.FeeTaker
	makerPay := rightBalance + trade.Match.FeeMaker
	fee := trade.Match.FeeMaker + trade.Match.FeeTaker

	if op == et.OpBuy {
		takerTokenID, makerTokenID = makerTokenID, takerTokenID
		feeTokenId = takerOrder.TokenSell
		takerRcv, makerPay = trade.Match.LeftBalance, trade.Match.LeftBalance

		takerPay, makerRcv = rightBalance+trade.Match.FeeTaker, rightBalance-trade.Match.FeeMaker
	}

	taker1 := &et.ZkTransferWithFee{
		TokenId:       uint64(takerTokenID),
		AmountOut:     et.AmountToZksync(uint64(takerPay)),
		FromAccountId: takerOrder.AccountID,
		ToAccountId:   trade.Current.Maker.Id,
		Signature:     takerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(makerRcv)),
	}
	maker1 := &et.ZkTransferWithFee{
		TokenId:       uint64(makerTokenID),
		AmountOut:     et.AmountToZksync(uint64(makerPay)),
		FromAccountId: trade.Current.Maker.Id,
		ToAccountId:   takerOrder.AccountID,
		Signature:     trade.MakerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(takerRcv)),
	}
	fee1 := &et.ZkTransferWithFee{
		TokenId:       feeTokenId,
		AmountOut:     et.AmountToZksync(0),
		FromAccountId: takerOrder.AccountID,
		ToAccountId:   trade.Current.Fee.Id,
		Signature:     takerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(fee)),
	}

	// 对先后顺序有要求: 先处理LeftAsset, 在处理RightAsset
	if op == et.OpSell {
		transfers = append(transfers, taker1)
		transfers = append(transfers, maker1)
	} else {
		transfers = append(transfers, maker1)
		transfers = append(transfers, taker1)
	}
	transfers = append(transfers, fee1)
	zlog.Error("swapGenTransfer", "takerPay", takerPay, "takerRcv", takerRcv,
		"makerPay", makerPay, "makerRcv", makerRcv)
	return transfers
}

*/
// 处理撮合结果
//  如果是 按不同的交易类型来处理的话, 零知识证明部分的代码, 会随交易的多样化, 需要也写很多函数来支持.
//  所以结果最好以 结算的形式作为参数.
//  不同交易的结果, 转化为有限的几种结算
//    主动结算: (用户地址发起的交易)    如: 撮合
//    被动结算: (系统特定帐号发起的交易) 如: 永续中暴仓, 和资金费
//  结算的列表以结果的形式体现帐号的变化, 和具体的业务无关
func (a *Action) SpotMatch(payload *et.SpotLimitOrder, list *types.Receipt) (*types.Receipt, error) {
	receipt := &types.Receipt{}
	for _, tradeRaw := range list.Logs {
		switch tradeRaw.Ty {
		case et.TySpotTradeLog:
			var trade et.ReceiptSpotTrade
			err := types.Decode(tradeRaw.Log, &trade)
			if err != nil {
				return nil, err
			}
			receipt2, err := a.Swap(payload, &trade)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, receipt2)
		default:
			//
		}
	}
	return receipt, nil
}

func (a *Action) AssetMatch(payload *et.SpotAssetLimitOrder, list *types.Receipt) (*types.Receipt, error) {
	receipt := &types.Receipt{}
	for _, tradeRaw := range list.Logs {
		switch tradeRaw.Ty {
		case et.TySpotTradeLog:
			var trade et.ReceiptSpotTrade
			err := types.Decode(tradeRaw.Log, &trade)
			if err != nil {
				return nil, err
			}
			receipt2, err := a.Swap(nil /* payload */, &trade)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, receipt2)
		default:
			//
		}
	}
	return receipt, nil
}

// A 和 B 交换 = transfer(A,B) + transfer(B,A) + swapfee()
// A 和 A 交换 = transfer(A,A) 0 + transfer(A,A) 0 + swapfee()
func (a *Action) Swap(payload1 *et.SpotLimitOrder, trade *et.ReceiptSpotTrade) (*types.Receipt, error) {
	return nil, nil
	/*
		var logs []*types.ReceiptLog
		var kvs []*types.KeyValue

		info, err := getTreeUpdateInfo(a.statedb)
		if err != nil {
			return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
		}
		zlog.Info("swap", "match", trade.Match)
		operationInfo := &et.OperationInfo{
			BlockHeight: uint64(a.height),
			TxIndex:     uint32(a.index),
			TxType:      et.TySwapAction,
			TokenID:     uint64(payload1.LeftAsset),
			Amount:      et.AmountToZksync(uint64(payload1.Amount)),
			FeeAmount:   "0",
			SigData:     payload1.GetOrder().Signature,
			AccountID:   payload1.Order.AccountID,
			SpecialInfo: new(et.OperationSpecialInfo),
		}

		swapSpecialData := genSwapSpecialData(payload1, trade)
		operationInfo.SpecialInfo.SpecialDatas = append(operationInfo.SpecialInfo.SpecialDatas, swapSpecialData)

		zlog.Debug("swapGenTransfer", "trade-buy", trade.MakerOrder.TokenBuy, "trade-sell", trade.MakerOrder.TokenSell)
		transfers := a.swapGenTransfer(payload1.Op, payload1.Order, trade)
		zlog.Debug("swapGenTransfer", "tokenid0", transfers[0].TokenId, "tokenid1", transfers[1].TokenId)
		// operationInfo, localKvs 通过 zklog 获得
		zklog := &et.ZkReceiptLog{OperationInfo: operationInfo}
		//for _, transfer1 := range transfers {
		receipt1, err := a.swapByTransfer(transfers[0], trade, info, zklog)
		if err != nil {
			return nil, err
		}
		logs = append(logs, receipt1.Logs...)
		kvs = append(kvs, receipt1.KV...)
		receipt2, err := a.swapByTransfer(transfers[1], trade, info, zklog)
		if err != nil {
			return nil, err
		}
		logs = append(logs, receipt2.Logs...)
		kvs = append(kvs, receipt2.KV...)

		receiptLog := &types.ReceiptLog{Ty: et.TySwapLog, Log: types.Encode(zklog)}
		logs = append(logs, receiptLog)

		feelog1, err := a.MakeFeeLog(transfers[2].AmountIn, info, transfers[2].TokenId, transfers[2].Signature)
		if err != nil {
			return nil, err
		}

		receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
		receipts = mergeReceipt(receipts, feelog1)
		return receipts, nil
	*/
}

/*
func toString(i int64) string {
	return new(big.Int).SetInt64(i).String()
}
*/

/*
func genSwapSpecialData(payload1 *et.SpotLimitOrder, trade *et.ReceiptSpotTrade) *et.OperationSpecialData {
	left, right := payload1.LeftAsset, payload1.RightAsset

	sellRightFee, buyRightFee := trade.Match.FeeMaker, trade.Match.FeeTaker
	if payload1.Op == et.OpBuy {
		sellRightFee, buyRightFee = trade.Match.FeeTaker, trade.Match.FeeMaker
	}

	sellLeft := &et.OrderPricePair{
		Sell: payload1.Order.Ratio1,
		Buy:  payload1.Order.Ratio2,
	}
	buyLeft := &et.OrderPricePair{
		Sell: trade.MakerOrder.Ratio1,
		Buy:  trade.MakerOrder.Ratio2,
	}

	if payload1.Op == et.OpBuy {
		sellLeft, buyLeft = buyLeft, sellLeft
	}

	specialData := &et.OperationSpecialData{
		TokenID: []uint64{uint64(left), uint64(right)},
		Amount: []string{
			et.AmountToZksync(uint64(trade.Match.LeftBalance)),
			et.AmountToZksync(uint64(trade.Match.RightBalance)),
			et.AmountToZksync(uint64(sellRightFee)),
			et.AmountToZksync(uint64(buyRightFee)),
			payload1.Order.Amount,
			trade.MakerOrder.Amount,
		},
		AccountID:   trade.GetCurrent().Taker.Id,
		RecipientID: trade.GetCurrent().Maker.Id,
		PricePair:   []*et.OrderPricePair{sellLeft, buyLeft},
		SigData:     trade.MakerOrder.Signature,
	}

	return specialData
}

*/
// 将参加放到 ZkTransfer, 可以方便的修改 Transfer的实现

/*
func (a *Action) swapByTransfer(payload *et.ZkTransferWithFee, trade *et.ReceiptSpotTrade, info *TreeUpdateInfo, zklog *et.ZkReceiptLog) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue

	operationInfo := zklog.OperationInfo
	//加上手续费
	amountOutTmp, _ := new(big.Int).SetString(payload.AmountOut, 10)
	amountOut := amountOutTmp.String()

	amountInTmp, _ := new(big.Int).SetString(payload.AmountIn, 10)
	amountIn := amountInTmp.String()

	err := checkParam(amountOut)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}
	err = checkParam(amountIn)
	if err != nil {
		return nil, errors.Wrapf(err, "checkParam")
	}

	fromLeaf, err := GetLeafByAccountId(a.statedb, payload.GetFromAccountId(), info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if fromLeaf == nil {
		return nil, errors.New("account not exist")
	}

	err = authVerification(payload.Signature.PubKey, fromLeaf.PubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "authVerification")
	}

	fromToken, err := GetTokenByAccountIdAndTokenId(a.statedb, payload.FromAccountId, payload.TokenId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	err = checkAmount(fromToken, amountOut)
	if err != nil {
		return nil, errors.Wrapf(err, "db.checkAmount")
	}

	//更新之前先计算证明
	receipt, err := calProof(a.statedb, info, payload.FromAccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}

	before := getBranchByReceipt(receipt, operationInfo, fromLeaf.EthAddress, fromLeaf.Chain33Addr, fromLeaf.PubKey, fromLeaf.ProxyPubKeys, payload.FromAccountId, payload.TokenId, receipt.Token.Balance)
	// after
	//更新fromLeaf
	fromKvs, fromLocal, err := UpdateLeaf(a.statedb, a.localDB, info, fromLeaf.GetAccountId(), payload.TokenId, amountOut, zt.Sub)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	kvs = append(kvs, fromKvs...)
	localKvs = append(localKvs, fromLocal...)
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.FromAccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}

	after := getBranchByReceipt(receipt, operationInfo, fromLeaf.EthAddress, fromLeaf.Chain33Addr, fromLeaf.PubKey, fromLeaf.ProxyPubKeys, payload.FromAccountId, payload.TokenId, receipt.Token.Balance)

	branch := &zt.OperationPairBranch{
		Before: before,
		After:  after,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch)
	// 2before
	toLeaf, err := GetLeafByAccountId(a.statedb, payload.ToAccountId, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if toLeaf == nil {
		return nil, errors.New("account not exist")
	}

	//更新之前先计算证明
	receipt, err = calProof(a.statedb, info, payload.ToAccountId, payload.TokenId)
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	var balance string
	if receipt.Token == nil {
		balance = "0"
	} else {
		balance = receipt.Token.Balance
	}
	before2 := getBranchByReceipt(receipt, operationInfo, toLeaf.EthAddress, toLeaf.Chain33Addr, toLeaf.PubKey, toLeaf.ProxyPubKeys, payload.ToAccountId, payload.TokenId, balance)
	// 2after
	//更新toLeaf
	tokvs, toLocal, err := UpdateLeaf(a.statedb, a.localDB, info, toLeaf.GetAccountId(), payload.GetTokenId(), amountIn, zt.Add)
	if err != nil {
		return nil, errors.Wrapf(err, "db.UpdateLeaf")
	}
	kvs = append(kvs, tokvs...)
	localKvs = append(localKvs, toLocal...)
	//更新之后计算证明
	receipt, err = calProof(a.statedb, info, payload.GetToAccountId(), payload.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "calProof")
	}
	after2 := getBranchByReceipt(receipt, operationInfo, toLeaf.EthAddress, toLeaf.Chain33Addr, toLeaf.PubKey, toLeaf.ProxyPubKeys, payload.ToAccountId, payload.TokenId, receipt.Token.Balance)
	rootHash := et.Str2Byte(receipt.TreeProof.RootHash)
	kv := &types.KeyValue{
		Key:   getHeightKey(a.height),
		Value: rootHash,
	}
	kvs = append(kvs, kv)

	branch2 := &et.OperationPairBranch{
		Before: before2,
		After:  after2,
	}
	operationInfo.OperationBranches = append(operationInfo.GetOperationBranches(), branch2)

	// 返回 operationInfo (本来就是引用zklog) localKvs
	zklog.LocalKvs = append(zklog.LocalKvs, localKvs...)

	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func (a *Action) swapGenTransfer(op int32, takerOrder *et.ZkOrder, trade *et.ReceiptSpotTrade) []*et.ZkTransferWithFee {
	// A 和 A 交易
	var transfers []*et.ZkTransferWithFee
	if trade.Current.Maker.Id == trade.Current.Taker.Id {
		return a.selfSwapGenTransfer(op, takerOrder, trade)
	}

	// A 和 B 交易
	// Sell
	takerTokenID, makerTokenID := takerOrder.TokenSell, takerOrder.TokenBuy
	feeTokenId := takerOrder.TokenBuy
	takerPay, makerRcv := trade.Match.LeftBalance, trade.Match.LeftBalance

	rightBalance := trade.Match.RightBalance
	takerRcv := rightBalance - trade.Match.FeeTaker
	makerPay := rightBalance + trade.Match.FeeMaker
	fee := trade.Match.FeeMaker + trade.Match.FeeTaker

	if op == et.OpBuy {
		takerTokenID, makerTokenID = makerTokenID, takerTokenID
		feeTokenId = takerOrder.TokenSell
		takerRcv, makerPay = trade.Match.LeftBalance, trade.Match.LeftBalance

		takerPay, makerRcv = rightBalance+trade.Match.FeeTaker, rightBalance-trade.Match.FeeMaker
	}

	taker1 := &et.ZkTransferWithFee{
		TokenId:       uint64(takerTokenID),
		AmountOut:     et.AmountToZksync(uint64(takerPay)),
		FromAccountId: takerOrder.AccountID,
		ToAccountId:   trade.Current.Maker.Id,
		Signature:     takerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(makerRcv)),
	}
	maker1 := &et.ZkTransferWithFee{
		TokenId:       uint64(makerTokenID),
		AmountOut:     et.AmountToZksync(uint64(makerPay)),
		FromAccountId: trade.Current.Maker.Id,
		ToAccountId:   takerOrder.AccountID,
		Signature:     trade.MakerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(takerRcv)),
	}
	fee1 := &et.ZkTransferWithFee{
		TokenId:       feeTokenId,
		AmountOut:     et.AmountToZksync(0),
		FromAccountId: takerOrder.AccountID,
		ToAccountId:   trade.Current.Fee.Id,
		Signature:     takerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(fee)),
	}

	// 对先后顺序有要求: 先处理LeftAsset, 在处理RightAsset
	if op == et.OpSell {
		transfers = append(transfers, taker1)
		transfers = append(transfers, maker1)
	} else {
		transfers = append(transfers, maker1)
		transfers = append(transfers, taker1)
	}
	transfers = append(transfers, fee1)
	zlog.Error("swapGenTransfer", "takerPay", takerPay, "takerRcv", takerRcv,
		"makerPay", makerPay, "makerRcv", makerRcv)
	return transfers
}

// A 和 A 交易时, 也需要构造4个transfer
// maker/taker 由于是同一个帐号, 所以takerPay makerPay 为0
func (a *Action) selfSwapGenTransfer(op int32, takerOrder *et.ZkOrder, trade *et.ReceiptSpotTrade) []*et.ZkTransferWithFee {
	// A 和 A 交易, 构造3个transfer, 使用transfer 实现
	var transfers []*et.ZkTransferWithFee
	leftPay, rightPay := int64(0), trade.Match.FeeTaker+trade.Match.FeeMaker
	feeTokenId := takerOrder.TokenBuy
	if op == et.OpBuy {
		feeTokenId = takerOrder.TokenSell
	}

	left := &et.ZkTransferWithFee{
		TokenId:       takerOrder.TokenSell,
		AmountOut:     et.AmountToZksync(uint64(leftPay)),
		FromAccountId: takerOrder.AccountID,
		ToAccountId:   trade.Current.Maker.Id,
		Signature:     takerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(leftPay)),
	}
	right := &et.ZkTransferWithFee{
		TokenId:       takerOrder.TokenBuy,
		AmountOut:     et.AmountToZksync(uint64(rightPay)),
		FromAccountId: trade.Current.Maker.Id,
		ToAccountId:   takerOrder.AccountID,
		Signature:     trade.MakerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(0)),
	}
	fee := &et.ZkTransferWithFee{
		TokenId:       feeTokenId,
		AmountOut:     et.AmountToZksync(uint64(0)),
		FromAccountId: takerOrder.AccountID,
		ToAccountId:   trade.Current.Fee.Id,
		Signature:     takerOrder.Signature,
		AmountIn:      et.AmountToZksync(uint64(rightPay)),
	}

	transfers = append(transfers, left)
	transfers = append(transfers, right)
	transfers = append(transfers, fee)
	return transfers
}
*/
