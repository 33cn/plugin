package executor

import (
	"fmt"
	"hash"
	"math/big"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/pkg/errors"
)

//暂时保存到全局变量里面
func getHistoryAccountProofFromDb(targetRootHash string) *zt.HistoryAccountProofInfo {
	if historyProof.RootHash == targetRootHash {
		return &historyProof
	}
	return nil
}

func setHistoryAccountProofToDb(proof *zt.HistoryAccountProofInfo) error {
	historyProof = *proof
	return nil
}

//根据rootHash获取account在该root下的证明
func getAccountProofInHistory(statedb dbm.KV, req *zt.ZkReqExistenceProof) (*zt.ZkProofWitness, error) {
	historyAccountInfo, err := BuildStateDbHistoryAccount(statedb, req.RootHash)
	if err != nil {
		return nil, err
	}
	if len(req.RootHash) > 0 && historyAccountInfo.RootHash != req.RootHash {
		return nil, errors.Wrapf(types.ErrNotFound, "req Root=%s,buildRoot=%s", req.RootHash, historyAccountInfo.GetRootHash())
	}
	return GetHistoryAccountProof(historyAccountInfo, req.AccountId, req.TokenId)
}

func getInitHistoryLeaf(ethFeeAddr, chain33FeeAddr string) []*zt.HistoryLeaf {
	leaves := getInitAccountLeaf(ethFeeAddr, chain33FeeAddr)
	var historyLeaf []*zt.HistoryLeaf
	for _, l := range leaves {
		history := &zt.HistoryLeaf{
			EthAddress:  l.EthAddress,
			AccountId:   l.AccountId,
			Chain33Addr: l.Chain33Addr,
			TokenHash:   zt.Byte2Str(l.TokenHash),
		}
		if len(l.TokenIds) > 0 {
			token := &zt.TokenBalance{
				TokenId: l.TokenIds[0],
				Balance: "0",
			}
			history.Tokens = append(history.Tokens, token)
		}
		historyLeaf = append(historyLeaf, history)
	}

	return historyLeaf
}

//BuildStateDbHistoryAccount 从statedb中构建账户tree，以此构建证明
func BuildStateDbHistoryAccount(db dbm.KV, reqRootHash string) (*zt.HistoryAccountProofInfo, error) {
	//允许reqRootHash为nil，或者和当前相同，则不需要重新构建
	if len(historyProof.RootHash) > 0 && (historyProof.RootHash == reqRootHash || len(reqRootHash) == 0) {
		return &historyProof, nil
	}

	accountMap := make(map[uint64]*zt.HistoryLeaf)
	lastAccountID, err := getLatestAccountID(db)
	if err != nil {
		return nil, errors.Wrapf(err, "getLatestAccountID")
	}
	for id := uint64(zt.SystemDefaultAcctId); id <= uint64(lastAccountID); id++ {
		leaf, err := GetLeafByAccountId(db, id)
		if err != nil {
			return nil, errors.Wrapf(err, "GetLeafByAccountId=%d", id)
		}
		if leaf == nil {
			return nil, errors.Wrapf(types.ErrNotFound, "GetLeafByAccountId=%d leaf=nil", id)
		}

		history := &zt.HistoryLeaf{
			AccountId:    id,
			EthAddress:   leaf.EthAddress,
			Chain33Addr:  leaf.Chain33Addr,
			PubKey:       leaf.PubKey,
			ProxyPubKeys: leaf.ProxyPubKeys,
		}
		for _, tokenId := range leaf.TokenIds {
			token, err := GetTokenByAccountIdAndTokenId(db, id, tokenId)
			if err != nil {
				zklog.Error("BuildStateDbHistoryAccount.getTokenErr", "acctId", id, "tokenId", tokenId, "err", err)
				return nil, errors.Wrapf(err, "GetTokenByAccountId=%d,TokenId=%d", id, tokenId)
			}
			if token == nil {
				zklog.Error("BuildStateDbHistoryAccount.tokenNotFound", "acctId", id, "tokenId", tokenId)
				return nil, errors.Wrapf(types.ErrNotFound, "GetTokenByAccountId=%d,TokenId=%d", id, tokenId)
			}
			history.Tokens = append(history.Tokens, token)
		}
		accountMap[id] = history
	}

	historyAccts, err := getHistoryAccounts(accountMap, uint64(lastAccountID))
	if err != nil {
		return nil, errors.Wrapf(err, "getHistoryAccounts")
	}
	setHistoryAccountProofToDb(historyAccts)
	return historyAccts, nil
}

//BuildHistoryAccountByProof 根据ProofId构建截止到当前proof的account tree，以此账户构建证明，适用于截止到某个proof的证明
func BuildHistoryAccountByProof(db dbm.KV, proofId uint64, reqRootHash string, feeAddrs *zt.ZkFeeAddrs) (*zt.HistoryAccountProofInfo, error) {
	if proofId == 0 {
		return BuildStateDbHistoryAccount(db, "")
	}
	if feeAddrs == nil {
		return nil, errors.Wrap(types.ErrInvalidParam, "feeAddr nil")
	}

	queInfo, err := GetProofId2QueueId(db, proofId)
	if err != nil {
		return nil, errors.Wrapf(err, "GetProofId2QueueId id=%d", proofId)
	}

	//TODO 考虑到queue会很长，将来考虑增加到协程处理，未结束返回等待错误
	var ops []*zt.ZkOperation
	//queueId 从1开始编号
	for i := int64(1); i <= queInfo.LastQueueId; i++ {
		op, err := GetL2QueueIdOp(db, i)
		if err != nil {
			return nil, errors.Wrapf(err, "GetL2QueueIdOp queueId=%d", i)
		}
		ops = append(ops, op)
	}

	return buildHistoryAccountsByOps(ops, reqRootHash, feeAddrs.EthFeeAddr, feeAddrs.L2FeeAddr)
}

//根据某个proof的root恢复所有账户的快照，在资产不会从L2转出到contract时候可以使用
func getHistoryAccountByRoot(localdb dbm.KV, targetRootHash, l1FeeAddr, l2FeeAddr string) (*zt.HistoryAccountProofInfo, error) {
	info := getHistoryAccountProofFromDb(targetRootHash)
	if info != nil {
		return info, nil
	}

	proofTable := NewCommitProofTable(localdb)

	rows, err := proofTable.ListIndex("root", getRootCommitProofKey(targetRootHash), nil, 1, zt.ListASC)
	if err != nil {
		return nil, errors.Wrapf(err, "proofTable.ListIndex")
	}
	proof := rows[0].Data.(*zt.ZkCommitProof)
	if proof == nil {
		return nil, errors.New("proof not exist")
	}

	var ops []*zt.ZkOperation
	for i := uint64(1); i <= proof.ProofId; i++ {
		row, err := proofTable.GetData(getProofIdCommitProofKey(i))
		if err != nil {
			return nil, err
		}
		data := row.Data.(*zt.ZkCommitProof)
		ops = append(ops, transferPubDataToOps(data.PubDatas)...)
	}

	return buildHistoryAccountsByOps(ops, targetRootHash, l1FeeAddr, l2FeeAddr)

}

func buildHistoryAccountsByOps(ops []*zt.ZkOperation, targetRoot string, l1FeeAddr, l2FeeAddr string) (*zt.HistoryAccountProofInfo, error) {
	accountMap := make(map[uint64]*zt.HistoryLeaf)
	maxAccountID := uint64(0)

	//从配置文件获取 feeAddr
	initLeaves := getInitHistoryLeaf(l1FeeAddr, l2FeeAddr)
	for _, l := range initLeaves {
		accountMap[l.AccountId] = l
		if maxAccountID < l.AccountId {
			maxAccountID = l.AccountId
		}
	}

	for _, op := range ops {
		newMaxId, err := getAccountMapByOp(op, accountMap, maxAccountID)
		if err != nil {
			return nil, errors.Wrapf(err, "getAccountMapByOp op=%v", op)
		}
		maxAccountID = newMaxId
	}

	historyAccts, err := getHistoryAccounts(accountMap, maxAccountID)
	if err != nil {
		return nil, errors.Wrapf(err, "getHistoryAccounts")
	}
	if len(targetRoot) != 0 && historyAccts.RootHash != targetRoot {
		return nil, errors.Wrapf(types.ErrInvalidParam, "calc root=%s,expect=%s", historyAccts.RootHash, targetRoot)
	}

	err = setHistoryAccountProofToDb(historyAccts)
	if err != nil {
		zklog.Error("setHistoryAccountProofToDb", "err", err)
	}
	return historyAccts, nil

}

func getAccountMapByOp(op *zt.ZkOperation, accountMap map[uint64]*zt.HistoryLeaf, maxAccountID uint64) (uint64, error) {
	switch op.Ty {
	case zt.TyDepositAction:
		operation := op.Op.GetDeposit()
		fromLeaf, ok := accountMap[operation.AccountID]
		if !ok {
			fromLeaf = &zt.HistoryLeaf{
				AccountId:   operation.GetAccountID(),
				EthAddress:  operation.GetEthAddress(),
				Chain33Addr: operation.GetLayer2Addr(),
				Tokens: []*zt.TokenBalance{
					{
						TokenId: operation.TokenID,
						Balance: operation.GetAmount(),
					},
				},
			}
		} else {
			var tokenBalance *zt.TokenBalance
			//找到token
			for _, token := range fromLeaf.Tokens {
				if token.TokenId == operation.TokenID {
					tokenBalance = token
				}
			}
			if tokenBalance == nil {
				tokenBalance = &zt.TokenBalance{
					TokenId: operation.TokenID,
					Balance: operation.Amount,
				}
				fromLeaf.Tokens = append(fromLeaf.Tokens, tokenBalance)
			} else {
				balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
				change, _ := new(big.Int).SetString(operation.Amount, 10)
				tokenBalance.Balance = new(big.Int).Add(balance, change).String()
			}
		}
		accountMap[operation.AccountID] = fromLeaf
		if operation.AccountID > maxAccountID {
			maxAccountID = operation.AccountID
		}
	case zt.TyWithdrawAction:
		operation := op.Op.GetWithdraw()
		fromLeaf, ok := accountMap[operation.AccountID]
		if !ok {
			return 0, errors.New(fmt.Sprintf("withdraw account=%d not exist", operation.AccountID))
		}
		var tokenBalance *zt.TokenBalance
		//找到token
		for _, token := range fromLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				tokenBalance = token
			}
		}
		if tokenBalance == nil {
			return 0, errors.New(fmt.Sprintf("withdraw account=%d token=%d not exist", operation.AccountID, operation.TokenID))
		} else {
			balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
			change, _ := new(big.Int).SetString(operation.Amount, 10)
			fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
			//add fee
			change = new(big.Int).Add(change, fee)
			if change.Cmp(balance) > 0 {
				return 0, errors.New(fmt.Sprintf("withdraw account=%d,token=%d,balance=%s,delta=%s",
					operation.AccountID, operation.TokenID, balance.String(), change.String()))
			}
			tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
		}
		accountMap[operation.AccountID] = fromLeaf

	case zt.TyTransferAction:
		operation := op.Op.GetTransfer()
		fromLeaf, ok := accountMap[operation.FromAccountID]
		if !ok {
			return 0, errors.New("account not exist")
		}
		toLeaf, ok := accountMap[operation.ToAccountID]
		if !ok {
			return 0, errors.New("account not exist")
		}

		var fromTokenBalance, toTokenBalance *zt.TokenBalance
		//找到fromToken
		for _, token := range fromLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				fromTokenBalance = token
			}
		}
		if fromTokenBalance == nil {
			return 0, errors.New("token not exist")
		} else {
			balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
			change, _ := new(big.Int).SetString(operation.Amount, 10)
			fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
			//add fee
			change = new(big.Int).Add(change, fee)
			if change.Cmp(balance) > 0 {
				return 0, errors.New(fmt.Sprintf("transfer balance=%s,delta=%s", balance.String(), change.String()))
			}
			fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
		}

		//找到toToken
		for _, token := range toLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				toTokenBalance = token
			}
		}
		change, _ := new(big.Int).SetString(operation.Amount, 10)
		if toTokenBalance == nil {
			toTokenBalance = &zt.TokenBalance{
				TokenId: operation.TokenID,
				Balance: change.String(),
			}
			toLeaf.Tokens = append(toLeaf.Tokens, toTokenBalance)
		} else {
			balance, _ := new(big.Int).SetString(toTokenBalance.GetBalance(), 10)
			toTokenBalance.Balance = new(big.Int).Add(balance, change).String()
		}
		accountMap[operation.FromAccountID] = fromLeaf
		accountMap[operation.ToAccountID] = toLeaf
	case zt.TyTransferToNewAction:
		operation := op.Op.GetTransferToNew()
		fromLeaf, ok := accountMap[operation.FromAccountID]
		if !ok {
			return 0, errors.New("account not exist")
		}

		var fromTokenBalance *zt.TokenBalance
		//找到fromToken
		for _, token := range fromLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				fromTokenBalance = token
			}
		}
		if fromTokenBalance == nil {
			return 0, errors.New("token not exist")
		} else {
			balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
			change, _ := new(big.Int).SetString(operation.Amount, 10)
			fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
			//add fee
			change = new(big.Int).Add(change, fee)
			if change.Cmp(balance) > 0 {
				return 0, errors.New("balance not enough")
			}
			fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
		}

		change, _ := new(big.Int).SetString(operation.Amount, 10)
		toLeaf := &zt.HistoryLeaf{
			AccountId:   operation.GetToAccountID(),
			EthAddress:  operation.GetEthAddress(),
			Chain33Addr: operation.GetLayer2Addr(),
			Tokens: []*zt.TokenBalance{
				{
					TokenId: operation.TokenID,
					Balance: change.String(),
				},
			},
		}

		accountMap[operation.FromAccountID] = fromLeaf
		accountMap[operation.ToAccountID] = toLeaf
		if operation.ToAccountID > maxAccountID {
			maxAccountID = operation.ToAccountID
		}
	case zt.TyProxyExitAction:
		operation := op.Op.GetProxyExit()
		//proxy
		fromLeaf, ok := accountMap[operation.ProxyID]
		if !ok {
			return 0, errors.New(fmt.Sprintf("account=%d not exist", operation.ProxyID))
		}

		var tokenBalance *zt.TokenBalance
		//找到token
		for _, token := range fromLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				tokenBalance = token
			}
		}
		if tokenBalance == nil {
			return 0, errors.New(fmt.Sprintf("token=%d not exist", operation.TokenID))
		} else {
			balance, _ := new(big.Int).SetString(tokenBalance.Balance, 10)
			fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
			tokenBalance.Balance = new(big.Int).Sub(balance, fee).String()
		}
		accountMap[operation.ProxyID] = fromLeaf

		//target account
		targetLeaf, ok := accountMap[operation.TargetID]
		if !ok {
			return 0, errors.New(fmt.Sprintf("proxy account=%d not exist", operation.TargetID))
		}
		//找到token
		for _, token := range targetLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				tokenBalance = token
			}
		}
		if tokenBalance == nil {
			return 0, errors.New(fmt.Sprintf("proxy target token=%d not exist", operation.TokenID))
		} else {
			if tokenBalance.Balance != operation.Amount {
				return 0, errors.New(fmt.Sprintf("proxy target tokenBalance different"))
			}
			tokenBalance.Balance = "0"
		}
		accountMap[operation.TargetID] = targetLeaf
	case zt.TySetPubKeyAction:
		operation := op.Op.GetSetPubKey()
		fromLeaf, ok := accountMap[operation.AccountID]
		if !ok {
			return 0, errors.New("account not exist")
		}

		pubKey := &zt.ZkPubKey{
			X: operation.PubKey.X,
			Y: operation.PubKey.Y,
		}
		if fromLeaf.ProxyPubKeys == nil {
			fromLeaf.ProxyPubKeys = new(zt.AccountProxyPubKeys)
		}
		switch operation.PubKeyTy {
		case 0:
			fromLeaf.PubKey = pubKey
		case zt.NormalProxyPubKey:
			fromLeaf.ProxyPubKeys.Normal = pubKey
		case zt.SystemProxyPubKey:
			fromLeaf.ProxyPubKeys.System = pubKey
		case zt.SuperProxyPubKey:
			fromLeaf.ProxyPubKeys.Super = pubKey
		default:
			return 0, errors.New(fmt.Sprintf("setPubKey ty=%d not support", operation.PubKeyTy))
		}
		accountMap[operation.AccountID] = fromLeaf

	//case zt.TyFullExitAction:
	//	operation := op.Op.GetFullExit()
	//	fromLeaf, ok := accountMap[operation.AccountID]
	//	if !ok {
	//		return nil, errors.New(fmt.Sprintf("account=%d not exist", operation.AccountID))
	//	}
	//	var tokenBalance *zt.TokenBalance
	//	//找到token
	//	for _, token := range fromLeaf.Tokens {
	//		if token.TokenId == operation.TokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		return nil, errors.New(fmt.Sprintf("token=%d not exist", operation.TokenID))
	//	} else {
	//		tokenBalance.Balance = "0"
	//	}
	//	accountMap[operation.AccountID] = fromLeaf

	//case zt.TySwapAction:
	//	operation := op.Op.GetSwap()
	//	//token: left asset, right asset 顺序
	//	//电路顺序为：sell-leftAsset, buy+leftAsset, sell-rightAsset-fee, buy+rightAsset-2ndFee
	//	//这里考虑leaf获取方便，顺序调整为 sell-leftAsset buy+rightAsset-2ndfee, sell-rightAsset-fee,buy+leftAsset
	//	leftLeaf, ok := accountMap[operation.Left.AccountID]
	//	if !ok {
	//		return nil, errors.New("account not exist")
	//	}
	//	var tokenBalance *zt.TokenBalance
	//	//找到sell token
	//	for _, token := range leftLeaf.Tokens {
	//		if token.TokenId == operation.LeftTokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		return nil, errors.New("taker token not exist")
	//	}
	//	balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//	change, _ := new(big.Int).SetString(operation.LeftDealAmount, 10)
	//	if change.Cmp(balance) > 0 {
	//		return nil, errors.New("balance not enough")
	//	}
	//	tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
	//
	//	//buy token
	//	tokenBalance = nil
	//	for _, token := range leftLeaf.Tokens {
	//		if token.TokenId == operation.RightTokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	//taker 2nd fee
	//	change, _ = new(big.Int).SetString(operation.RightDealAmount, 10)
	//	fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
	//	if fee.Cmp(change) > 0 {
	//		return nil, errors.New("change not enough to fee to taker")
	//	}
	//	//sub fee
	//	change = new(big.Int).Sub(change, fee)
	//	if tokenBalance == nil {
	//		newToken := &zt.TokenBalance{
	//			TokenId: operation.RightTokenID,
	//			Balance: change.String(),
	//		}
	//		leftLeaf.Tokens = append(leftLeaf.Tokens, newToken)
	//	} else {
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		tokenBalance.Balance = new(big.Int).Add(balance, change).String()
	//	}
	//	accountMap[operation.Left.AccountID] = leftLeaf
	//
	//	//toAccount leaf
	//	rightLeaf, ok := accountMap[operation.Right.AccountID]
	//	if !ok {
	//		return nil, errors.New(fmt.Sprintf("right account=%d not exist", operation.Right.AccountID))
	//	}
	//
	//	//找到right asset
	//	tokenBalance = nil
	//	for _, token := range rightLeaf.Tokens {
	//		if token.TokenId == operation.RightTokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		return nil, errors.New(fmt.Sprintf("right sell token=%d not exist", operation.RightTokenID))
	//	}
	//	balance, _ = new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//	change, _ = new(big.Int).SetString(operation.RightDealAmount, 10)
	//	if change.Cmp(balance) > 0 {
	//		return nil, errors.New("maker token balance not enough")
	//	}
	//	newBalance := new(big.Int).Sub(balance, change)
	//	//1st fee
	//	fee, _ = new(big.Int).SetString(operation.Fee.Fee, 10)
	//	if fee.Cmp(newBalance) > 0 {
	//		return nil, errors.New("change not enough to fee to taker")
	//	}
	//	tokenBalance.Balance = new(big.Int).Sub(newBalance, fee).String()
	//
	//	//buy token
	//	tokenBalance = nil
	//	for _, token := range rightLeaf.Tokens {
	//		if token.TokenId == operation.RightTokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		newToken := &zt.TokenBalance{
	//			TokenId: operation.RightTokenID,
	//			Balance: operation.RightDealAmount,
	//		}
	//		rightLeaf.Tokens = append(rightLeaf.Tokens, newToken)
	//	} else {
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		change, _ := new(big.Int).SetString(operation.RightDealAmount, 10)
	//		tokenBalance.Balance = new(big.Int).Add(balance, change).String()
	//	}
	//	accountMap[operation.Right.AccountID] = rightLeaf

	case zt.TyContractToTreeAction:
		operation := op.Op.GetContractToTree()
		fromLeaf, ok := accountMap[zt.SystemTree2ContractAcctId]
		if !ok {
			return 0, errors.New("account not exist")
		}
		var tokenBalance *zt.TokenBalance
		//找到token
		for _, token := range fromLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				tokenBalance = token
			}
		}
		if tokenBalance == nil {
			return 0, errors.New(fmt.Sprintf("contract2tree system acct  token=%d not exist", operation.TokenID))
		} else {
			balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
			change, _ := new(big.Int).SetString(operation.Amount, 10)
			tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
		}
		accountMap[zt.SystemTree2ContractAcctId] = fromLeaf

		//toAccount
		toLeaf, ok := accountMap[operation.AccountID]
		if !ok {
			return 0, errors.Wrapf(types.ErrAccountNotExist, "ty=%d,toAccountID=%d not exist", zt.TyContractToTreeAction, operation.AccountID)
		}
		//找到toToken
		var toTokenBalance *zt.TokenBalance
		for _, token := range toLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				toTokenBalance = token
			}
		}
		change, _ := new(big.Int).SetString(operation.Amount, 10)
		if toTokenBalance == nil {
			toTokenBalance = &zt.TokenBalance{
				TokenId: operation.TokenID,
				Balance: change.String(),
			}
			toLeaf.Tokens = append(toLeaf.Tokens, toTokenBalance)
		} else {
			balance, _ := new(big.Int).SetString(toTokenBalance.GetBalance(), 10)
			toTokenBalance.Balance = new(big.Int).Add(balance, change).String()
		}
		accountMap[operation.AccountID] = toLeaf

	case zt.TyContractToTreeNewAction:
		operation := op.Op.GetContract2TreeNew()
		fromLeaf, ok := accountMap[zt.SystemTree2ContractAcctId]
		if !ok {
			return 0, errors.New("account not exist")
		}
		var tokenBalance *zt.TokenBalance
		//找到token
		for _, token := range fromLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				tokenBalance = token
			}
		}
		if tokenBalance == nil {
			errors.New(fmt.Sprintf("contract2treeNew system acct  token=%d not exist", operation.TokenID))
		} else {
			balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
			change, _ := new(big.Int).SetString(operation.Amount, 10)
			tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
		}
		accountMap[zt.SystemTree2ContractAcctId] = fromLeaf

		//to leaf
		change, _ := new(big.Int).SetString(operation.Amount, 10)
		toLeaf := &zt.HistoryLeaf{
			AccountId:   operation.ToAccountID,
			EthAddress:  operation.GetEthAddress(),
			Chain33Addr: operation.GetLayer2Addr(),
			Tokens: []*zt.TokenBalance{
				{
					TokenId: operation.TokenID,
					Balance: change.String(),
				},
			},
		}
		accountMap[operation.ToAccountID] = toLeaf
		if operation.ToAccountID > maxAccountID {
			maxAccountID = operation.ToAccountID
		}

	case zt.TyTreeToContractAction:
		operation := op.Op.GetTreeToContract()
		fromLeaf, ok := accountMap[operation.AccountID]
		if !ok {
			return 0, errors.New("account not exist")
		}
		var tokenBalance *zt.TokenBalance
		//找到token
		for _, token := range fromLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				tokenBalance = token
			}
		}
		if tokenBalance == nil {
			return 0, errors.New("token not exist")
		} else {
			balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
			change, _ := new(big.Int).SetString(operation.Amount, 10)
			if change.Cmp(balance) > 0 {
				return 0, errors.New("balance not enough")
			}
			tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
		}
		accountMap[operation.AccountID] = fromLeaf

	case zt.TyFeeAction:
		operation := op.Op.GetFee()
		fromLeaf, ok := accountMap[operation.AccountID]
		if !ok {
			return 0, errors.New(fmt.Sprintf("fee AccountID=%d not exist", operation.AccountID))
		}
		if operation.AccountID != zt.SystemFeeAccountId {
			return 0, errors.New(fmt.Sprintf("fee AccountID=%d not systmeId=%d", operation.AccountID, zt.SystemFeeAccountId))
		}
		var tokenBalance *zt.TokenBalance
		//找到token
		for _, token := range fromLeaf.Tokens {
			if token.TokenId == operation.TokenID {
				tokenBalance = token
			}
		}
		if tokenBalance == nil {
			newToken := &zt.TokenBalance{
				TokenId: operation.TokenID,
				Balance: operation.Amount,
			}
			fromLeaf.Tokens = append(fromLeaf.Tokens, newToken)
		} else {
			balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
			change, _ := new(big.Int).SetString(operation.Amount, 10)

			tokenBalance.Balance = new(big.Int).Add(balance, change).String()
		}
		accountMap[operation.AccountID] = fromLeaf

	//case zt.TyMintNFTAction:
	//	operation := op.Op.GetMintNFT()
	//	fromLeaf, ok := accountMap[operation.MintAcctID]
	//	if !ok {
	//		return nil, errors.New("account not exist")
	//	}
	//	var tokenBalance *zt.TokenBalance
	//	//1. fee
	//	for _, token := range fromLeaf.Tokens {
	//		if token.TokenId == operation.Fee.TokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		return nil, errors.New("mint nft token not exist")
	//	} else {
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		change, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
	//		if change.Cmp(balance) > 0 {
	//			return nil, errors.New("mint nft fee balance not enough")
	//		}
	//		tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
	//	}
	//
	//	//2. creator systemNFTTokenID+1 get serialId
	//	tokenBalance = nil
	//	serialId := "0"
	//	for _, token := range fromLeaf.Tokens {
	//		if token.TokenId == zt.SystemNFTTokenId {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		newToken := &zt.TokenBalance{
	//			TokenId: zt.SystemNFTTokenId,
	//			Balance: "1",
	//		}
	//		fromLeaf.Tokens = append(fromLeaf.Tokens, newToken)
	//	} else {
	//		//before balance as the serialId
	//		serialId = tokenBalance.GetBalance()
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		tokenBalance.Balance = new(big.Int).Add(balance, big.NewInt(1)).String()
	//	}
	//
	//	//3. systemNFT AccountID's systemNFTTokenID balance+1 get newTokenID
	//	newNFTTokenID := uint64(0)
	//	systemNFTLeaf, ok := accountMap[zt.SystemNFTAccountId]
	//	if !ok {
	//		return nil, errors.New("SystemNFTAccountID not found")
	//	}
	//	tokenBalance = nil
	//	for _, token := range systemNFTLeaf.Tokens {
	//		if token.TokenId == zt.SystemNFTTokenId {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		newNFTTokenID = new(big.Int).Add(big.NewInt(zt.SystemNFTTokenId), big.NewInt(1)).Uint64()
	//		newToken := &zt.TokenBalance{
	//			TokenId: zt.SystemNFTTokenId,
	//			Balance: new(big.Int).Add(big.NewInt(zt.SystemNFTTokenId), big.NewInt(2)).String(),
	//		}
	//		systemNFTLeaf.Tokens = append(systemNFTLeaf.Tokens, newToken)
	//	} else {
	//		//before balance as the new token Id
	//		b, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		newNFTTokenID = b.Uint64()
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		tokenBalance.Balance = new(big.Int).Add(balance, big.NewInt(1)).String()
	//	}
	//
	//	//4. set new system NFT TokenID balance
	//	//检查systemNFTAccount 没有此newNFTTokenID
	//	tokenBalance = nil
	//	for _, token := range systemNFTLeaf.Tokens {
	//		if token.TokenId == newNFTTokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance != nil {
	//		return nil, errors.New(fmt.Sprintf("systemNFTAccount has the newNFT id=%d", newNFTTokenID))
	//	}
	//	newSysNFTTokenBalance, err := getNewNFTTokenBalance(operation.MintAcctID, serialId, operation.ErcProtocol, operation.Amount,
	//		operation.ContentHash[0], operation.ContentHash[1])
	//	if err != nil {
	//		return nil, errors.Wrapf(err, "newNFTTokenBalance")
	//	}
	//	newToken := &zt.TokenBalance{
	//		TokenId: newNFTTokenID,
	//		Balance: newSysNFTTokenBalance,
	//	}
	//	systemNFTLeaf.Tokens = append(systemNFTLeaf.Tokens, newToken)
	//	accountMap[zt.SystemNFTAccountId] = systemNFTLeaf
	//
	//	//5. recipient id
	//	toLeaf, ok := accountMap[operation.RecipientID]
	//	if !ok {
	//		return nil, errors.New("account not exist")
	//	}
	//	tokenBalance = nil
	//	for _, token := range fromLeaf.Tokens {
	//		if token.TokenId == operation.Fee.TokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance != nil {
	//		return nil, errors.New("nft recipient nft token  existed")
	//	}
	//	newToken = &zt.TokenBalance{
	//		TokenId: newNFTTokenID,
	//		Balance: new(big.Int).SetUint64(operation.Amount).String(),
	//	}
	//	toLeaf.Tokens = append(toLeaf.Tokens, newToken)

	//case zt.TyWithdrawNFTAction:
	//	operation := op.Op.GetWithdrawNFT()
	//	fromLeaf, ok := accountMap[operation.FromAcctID]
	//	if !ok {
	//		return nil, errors.New("account not exist")
	//	}
	//	var tokenBalance *zt.TokenBalance
	//	//1. fee
	//	for _, token := range fromLeaf.Tokens {
	//		if token.TokenId == operation.Fee.TokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		return nil, errors.New("withdraw nft fee token not exist")
	//	} else {
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		change, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
	//		if change.Cmp(balance) > 0 {
	//			return nil, errors.New("withdraw nft fee balance not enough")
	//		}
	//		tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
	//	}
	//	//2. NFT token balance-amount
	//	tokenBalance = nil
	//	for _, token := range fromLeaf.Tokens {
	//		if token.TokenId == operation.NFTTokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		return nil, errors.New("withdraw nft token not exist")
	//	} else {
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		change := new(big.Int).SetUint64(operation.WithdrawAmount)
	//		if change.Cmp(balance) > 0 {
	//			return nil, errors.New("withdraw nft  balance not enough")
	//		}
	//		tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
	//	}
	//case zt.TyTransferNFTAction:
	//	operation := op.Op.GetTransferNFT()
	//	fromLeaf, ok := accountMap[operation.FromAccountID]
	//	if !ok {
	//		return nil, errors.New("account not exist")
	//	}
	//	var tokenBalance *zt.TokenBalance
	//	//1. fee
	//	for _, token := range fromLeaf.Tokens {
	//		if token.TokenId == operation.Fee.TokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		return nil, errors.New("transfer nft fee token not exist")
	//	} else {
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		change, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
	//		if change.Cmp(balance) > 0 {
	//			return nil, errors.New("withdraw nft fee balance not enough")
	//		}
	//		tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
	//	}
	//	//2. NFT token balance-amount
	//	tokenBalance = nil
	//	for _, token := range fromLeaf.Tokens {
	//		if token.TokenId == operation.NFTTokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		return nil, errors.New("transfer nft token not exist")
	//	} else {
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		change := new(big.Int).SetUint64(operation.Amount)
	//		if change.Cmp(balance) > 0 {
	//			return nil, errors.New("transfer nft  balance not enough")
	//		}
	//		tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
	//	}
	//
	//	toLeaf, ok := accountMap[operation.RecipientID]
	//	if !ok {
	//		return nil, errors.New("account not exist")
	//	}
	//
	//	// NFT token balance+amount
	//	tokenBalance = nil
	//	for _, token := range toLeaf.Tokens {
	//		if token.TokenId == operation.NFTTokenID {
	//			tokenBalance = token
	//		}
	//	}
	//	if tokenBalance == nil {
	//		newToken := &zt.TokenBalance{
	//			TokenId: operation.NFTTokenID,
	//			Balance: new(big.Int).SetUint64(operation.Amount).String(),
	//		}
	//		toLeaf.Tokens = append(toLeaf.Tokens, newToken)
	//	} else {
	//		balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
	//		change := new(big.Int).SetUint64(operation.Amount)
	//		tokenBalance.Balance = new(big.Int).Add(balance, change).String()
	//	}
	default:
		return 0, errors.New(fmt.Sprintf("not support op ty=%d", op.Ty))
	}

	return maxAccountID, nil
}

func getHistoryAccounts(accountMap map[uint64]*zt.HistoryLeaf, maxAccountId uint64) (*zt.HistoryAccountProofInfo, error) {
	h := mimc.NewMiMC(zt.ZkMimcHashSeed)
	historyAccounts := &zt.HistoryAccountProofInfo{}
	for i := uint64(zt.SystemDefaultAcctId); i <= maxAccountId; i++ {
		if _, ok := accountMap[i]; !ok {
			return nil, errors.Wrapf(types.ErrNotFound, "AccountID=%d not exist", i)
		}
		historyAccounts.Leaves = append(historyAccounts.Leaves, accountMap[i])
		historyAccounts.LeafHashes = append(historyAccounts.LeafHashes, getHistoryLeafHash(accountMap[i], h))
	}

	//验证leafHash和rootHash是否匹配
	accountMerkleProof, err := getMerkleTreeProof(zt.SystemFeeAccountId, historyAccounts.LeafHashes, h)
	if err != nil {
		return nil, errors.Wrapf(err, "account.getMerkleTreeProof")
	}
	historyAccounts.RootHash = accountMerkleProof.RootHash
	return historyAccounts, nil
}

func getHistoryLeafHash(leaf *zt.HistoryLeaf, h hash.Hash) []byte {
	h.Reset()
	tokenHash := getHistoryTokenHash(leaf.AccountId, leaf.Tokens, h)

	h.Reset()
	accountIdBytes := new(fr.Element).SetUint64(leaf.GetAccountId()).Bytes()
	h.Write(accountIdBytes[:])
	h.Write(zt.Str2Byte(leaf.GetEthAddress()))
	h.Write(zt.Str2Byte(leaf.GetChain33Addr()))

	getLeafPubKeyHash(h, leaf.GetPubKey())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetNormal())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSystem())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSuper())

	h.Write(zt.Str2Byte(tokenHash))
	sum := h.Sum(nil)
	h.Reset()
	return sum
}

func getHistoryTokenHash(accountId uint64, tokens []*zt.TokenBalance, h hash.Hash) string {
	if (accountId <= zt.SystemTree2ContractAcctId) && len(tokens) <= 0 {
		return "0"
	}
	var tokenHashes [][]byte
	for _, token := range tokens {
		tokenHashes = append(tokenHashes, getTokenBalanceHash(h, token))
	}

	tokenTree := getNewTreeWithHash(h)
	for _, v := range tokenHashes {
		tokenTree.Push(v)
	}
	root := zt.Byte2Str(tokenTree.Root())
	h.Reset()
	return root
}

func getTokenBalanceHash(h hash.Hash, token *zt.TokenBalance) []byte {
	h.Reset()
	tokenIdBytes := new(fr.Element).SetUint64(token.GetTokenId()).Bytes()
	h.Write(tokenIdBytes[:])
	h.Write(zt.Str2Byte(token.Balance))
	sum := h.Sum(nil)
	h.Reset()
	return sum
}

func getMerkleTreeProof(index uint64, hashes [][]byte, h hash.Hash) (*zt.MerkleTreeProof, error) {
	tree := getNewTreeWithHash(h)
	err := tree.SetIndex(index)
	if err != nil {
		return nil, errors.Wrapf(err, "tree.SetIndex")
	}
	for _, h := range hashes {
		tree.Push(h)
	}
	rootHash, proofSet, proofIndex, numLeaves := tree.Prove()
	helpers := make([]string, 0)
	proofStringSet := make([]string, 0)
	for _, v := range merkletree.GenerateProofHelper(proofSet, proofIndex, numLeaves) {
		helpers = append(helpers, big.NewInt(int64(v)).String())
	}
	for _, v := range proofSet {
		proofStringSet = append(proofStringSet, zt.Byte2Str(v))
	}
	h.Reset()
	return &zt.MerkleTreeProof{RootHash: zt.Byte2Str(rootHash), ProofSet: proofStringSet, Helpers: helpers}, nil
}

func GetHistoryAccountProof(historyAccountInfo *zt.HistoryAccountProofInfo, targetAccountID, targetTokenID uint64) (*zt.ZkProofWitness, error) {
	if targetAccountID >= uint64(len(historyAccountInfo.Leaves)) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "targetAccountID=%d not exist,maxAcctId=%d", targetAccountID, len(historyAccountInfo.Leaves)-1)
	}
	targetLeaf := historyAccountInfo.Leaves[targetAccountID]

	var tokenFound bool
	var tokenIndex int
	for i, t := range targetLeaf.Tokens {
		if t.TokenId == targetTokenID {
			tokenIndex = i
			tokenFound = true
			break
		}
	}
	if !tokenFound {
		return nil, errors.Wrapf(types.ErrInvalidParam, "AccountID=%d has no asset TokenID=%d", targetAccountID, targetTokenID)
	}
	h := mimc.NewMiMC(zt.ZkMimcHashSeed)
	accountMerkleProof, err := getMerkleTreeProof(targetAccountID, historyAccountInfo.LeafHashes, h)
	if err != nil {
		return nil, errors.Wrapf(err, "account.getMerkleTreeProof")
	}
	if accountMerkleProof.RootHash != historyAccountInfo.RootHash {
		return nil, errors.Wrapf(types.ErrInvalidParam, "calc root=%s,expect=%s", accountMerkleProof.RootHash, historyAccountInfo.RootHash)
	}

	//token proof
	var tokenHashes [][]byte
	for _, token := range targetLeaf.Tokens {
		tokenHashes = append(tokenHashes, getTokenBalanceHash(h, token))
	}
	tokenMerkleProof, err := getMerkleTreeProof(uint64(tokenIndex), tokenHashes, h)
	if err != nil {
		return nil, errors.Wrapf(err, "token.getMerkleProof")
	}

	accTreePath := &zt.SiblingPath{
		Path:   accountMerkleProof.ProofSet,
		Helper: accountMerkleProof.Helpers,
	}
	accountW := &zt.AccountWitness{
		ID:            targetAccountID,
		EthAddr:       targetLeaf.EthAddress,
		Chain33Addr:   targetLeaf.Chain33Addr,
		TokenTreeRoot: tokenMerkleProof.RootHash,
		PubKey:        targetLeaf.PubKey,
		ProxyPubKeys:  targetLeaf.ProxyPubKeys,
		Sibling:       accTreePath,
	}

	tokenTreePath := &zt.SiblingPath{
		Path:   tokenMerkleProof.ProofSet,
		Helper: tokenMerkleProof.Helpers,
	}
	tokenW := &zt.TokenWitness{
		ID:      targetTokenID,
		Balance: targetLeaf.Tokens[tokenIndex].Balance,
		Sibling: tokenTreePath,
	}
	var witness zt.ZkProofWitness
	witness.AccountWitness = accountW
	witness.TokenWitness = tokenW
	witness.TreeRoot = historyProof.RootHash

	return &witness, nil
}
