package executor

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
)

// Query_GetAccountTree 获取当前的树
func (z *zksync) Query_GetAccountTree(in *types.ReqNil) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	var tree zt.AccountTree
	val, err := z.GetStateDB().Get(GetAccountTreeKey())
	if err != nil {
		return nil, err
	}
	err = types.Decode(val, &tree)
	if err != nil {
		return nil, err
	}
	return &tree, nil
}

// Query_GetNFTStatus 获取nft by id
func (z *zksync) Query_GetNFTStatus(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	var status zt.ZkNFTTokenStatus
	val, err := z.GetStateDB().Get(GetNFTIdPrimaryKey(in.TokenId))
	if err != nil {
		return nil, err
	}
	err = types.Decode(val, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Query_GetNFTId get nft id by content hash
func (z *zksync) Query_GetNFTId(in *types.ReqString) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	var id types.Int64
	val, err := z.GetStateDB().Get(GetNFTHashPrimaryKey(in.Data))
	if err != nil {
		return nil, err
	}
	err = types.Decode(val, &id)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Query_GetAccountById  通过accountId查询account
func (z *zksync) Query_GetAccountById(in *zt.ZkQueryReq) (types.Message, error) {
	var leaf zt.Leaf
	val, err := z.GetStateDB().Get(GetAccountIdPrimaryKey(in.AccountId))
	if err != nil {
		return nil, err
	}

	err = types.Decode(val, &leaf)
	if err != nil {
		return nil, err
	}
	var ok bool
	leaf.EthAddress, ok = zt.DecimalAddr2Hex(leaf.GetEthAddress(), zt.EthAddrLen)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong eth addr format=%s", leaf.GetEthAddress())
	}
	leaf.Chain33Addr, ok = zt.DecimalAddr2Hex(leaf.GetChain33Addr(), zt.BTYAddrLen)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong chain33 addr format=%s", leaf.GetChain33Addr())
	}
	return &leaf, nil
}

// Query_GetAccountByEth  通过eth地址查询account
func (z *zksync) Query_GetAccountByEth(in *zt.ZkQueryReq) (types.Message, error) {
	res := new(zt.ZkQueryResp)
	newEthAddr, ok := zt.HexAddr2Decimal(in.EthAddress)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong eth addr format=%s", in.GetEthAddress())
	}
	leaves, err := GetLeafByEthAddress(z.GetLocalDB(), newEthAddr)
	if err != nil {
		return nil, err
	}
	for _, l := range leaves {
		r, err := z.Query_GetAccountById(&zt.ZkQueryReq{AccountId: l.GetAccountId()})
		if err != nil {
			return nil, errors.Wrapf(err, "id=%d", l.AccountId)
		}
		res.Leaves = append(res.Leaves, r.(*zt.Leaf))
	}
	return res, nil
}

// Query_GetAccountByChain33  通过chain33地址查询account
func (z *zksync) Query_GetAccountByChain33(in *zt.ZkQueryReq) (types.Message, error) {
	res := new(zt.ZkQueryResp)
	addr, ok := zt.HexAddr2Decimal(in.GetChain33Addr())
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "chain33 addr not hex format,%s", in.GetChain33Addr())
	}
	leaves, err := GetLeafByChain33Address(z.GetLocalDB(), addr)
	if err != nil {
		return nil, err
	}
	for _, l := range leaves {
		r, err := z.Query_GetAccountById(&zt.ZkQueryReq{AccountId: l.GetAccountId()})
		if err != nil {
			return nil, errors.Wrapf(err, "id=%d", l.AccountId)
		}
		res.Leaves = append(res.Leaves, r.(*zt.Leaf))
	}
	return res, nil
}

// Query_GetLastCommitProof 获取最新proof信息
func (z *zksync) Query_GetLastCommitProof(in *types.ReqNil) (types.Message, error) {
	return getLastCommitProofData(z.GetStateDB())
}

// Query_GetLastOnChainProof 获取最新的包含OnChainPubData的Proof
func (z *zksync) Query_GetLastOnChainProof(in *types.ReqNil) (types.Message, error) {
	return getLastOnChainProofData(z.GetStateDB())
}

// Query_GetMaxAccountId 获取当前最大账户id
func (z *zksync) Query_GetMaxAccountId(in *types.ReqNil) (types.Message, error) {
	lastAccountID, err := getLatestAccountID(z.GetStateDB())
	return &types.Int64{Data: lastAccountID}, err
}

// Query_GetTreeInitRoot 获取系统初始tree root
func (z *zksync) Query_GetTreeInitRoot(in *types.ReqAddrs) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	var eth, chain33 string
	//可以不填addr
	if len(in.Addrs) == 2 {
		addr, ok := zt.HexAddr2Decimal(in.Addrs[0])
		if !ok {
			return nil, errors.Wrapf(types.ErrNotAllow, "addr0=%s not hex format", in.Addrs[0])
		}
		eth = addr

		addr, ok = zt.HexAddr2Decimal(in.Addrs[1])
		if !ok {
			return nil, errors.Wrapf(types.ErrNotAllow, "addr1=%s not hex format", in.Addrs[1])
		}
		chain33 = addr
	}

	root := getInitTreeRoot(z.GetAPI().GetConfig(), eth, chain33)
	return &types.ReplyString{Data: root}, nil
}

// Query_GetCfgFeeAddr 获取系统初始fee addr
func (z *zksync) Query_GetCfgFeeAddr(in *types.ReqNil) (types.Message, error) {
	eth, l2 := getCfgFeeAddr(z.GetAPI().GetConfig())
	return &zt.ZkFeeAddrs{EthFeeAddr: eth, L2FeeAddr: l2}, nil
}

// Query_GetCfgTokenFee 获取系统配置的fee
func (z *zksync) Query_GetCfgTokenFee(in *zt.ZkSetFee) (types.Message, error) {
	amount, err := getDbFeeData(z.GetStateDB(), in.GetActionTy(), in.GetTokenId())
	if err != nil {
		return nil, err
	}
	return &types.ReplyString{Data: amount}, nil
}

// Query_GetVerifiers 获取系统初始fee addr
func (z *zksync) Query_GetVerifiers(in *types.ReqNil) (types.Message, error) {
	return getVerifierData(z.GetStateDB())
}

// Query_GetZkContractAccount 批量获取交易证明
func (z *zksync) Query_GetZkContractAccount(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	accountdb, _ := account.NewAccountDB(z.GetAPI().GetConfig(), zt.Zksync, in.TokenSymbol, z.GetStateDB())
	contractAccount := accountdb.LoadAccount(in.Chain33WalletAddr)
	return contractAccount, nil
}

// 注意：如果val超过1e10会被圆整到1e10格式表示
func decimalVal(val string, decimal uint32) string {
	if decimal == 0 {
		return val
	}
	fbalance := new(big.Float)
	fbalance.SetString(val)
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(int(decimal))))
	return ethValue.String()
}

// Query_GetTokenBalance 根据token和account获取balance
func (z *zksync) Query_GetTokenBalance(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	res := new(zt.ZkQueryResp)
	token, err := GetTokenByAccountIdAndTokenIdInDB(z.GetStateDB(), in.AccountId, in.TokenId)
	if err != nil {
		return nil, errors.Wrap(err, "getTokenInDb")
	}

	if in.Decimal > 0 {
		symbol, err := GetTokenByTokenId(z.GetStateDB(), strconv.Itoa(int(in.TokenId)))
		if err != nil {
			return nil, errors.Wrapf(err, "getTokenSymbol")
		}
		token.Balance = decimalVal(token.Balance, symbol.Decimal)
	}

	res.TokenBalances = append(res.TokenBalances, token)
	return res, nil
}

// Query_GetTokenSymbol 根据id获取当前symbol，根据symbol获取对应token id
func (z *zksync) Query_GetTokenSymbol(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	//symbol非空，查询id
	if len(in.TokenSymbol) > 0 {
		return GetTokenBySymbol(z.GetStateDB(), in.TokenSymbol)
	}
	//根据id查询symbol
	idStr := new(big.Int).SetUint64(in.TokenId).String()
	return GetTokenByTokenId(z.GetStateDB(), idStr)
}

// Query_GetLastPriorityQueueId 获取最后的 l1 priority  id
func (z *zksync) Query_GetLastPriorityQueueId(in *types.ReqNil) (types.Message, error) {
	return getLastEthPriorityQueueID(z.GetStateDB())
}

// Query_GetPriorityOpInfo 根据priorityId获取operation信息
func (z *zksync) Query_GetPriorityOpInfo(in *types.Int64) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return GetPriorityDepositData(z.GetStateDB(), in.Data)
}

// Query_GetBatchPriorityOpInfo 根据priorityId获取operation信息
func (z *zksync) Query_GetBatchPriorityOpInfo(in *types.ReqBlocks) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	var batch zt.ZkBatchOperation
	for i := in.Start; i <= in.End; i++ {
		deposit, err := GetPriorityDepositData(z.GetStateDB(), i)
		if err != nil {
			zklog.Error("Query_GetBatchPriorityOpInfo", "priorityid", i, "err", err)
			return nil, err
		}
		batch.Ops = append(batch.Ops, &zt.ZkOperation{Ty: zt.TyDepositAction, Op: &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: deposit}}})
	}
	return &batch, nil

}

// Query_GetL2QueueOpInfo 根据l2 queue id获取operation信息
func (z *zksync) Query_GetL2QueueOpInfo(in *types.Int64) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return GetL2QueueIdOp(z.GetStateDB(), in.Data)
}

// Query_GetL2BatchQueueOpInfo 批量获取l2 queue id的operation信息
func (z *zksync) Query_GetL2BatchQueueOpInfo(in *types.ReqBlocks) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	if in.End < in.Start {
		return nil, errors.Wrapf(types.ErrInvalidParam, "end=%d < start=%d", in.End, in.Start)
	}
	var batch zt.ZkBatchOperation
	for i := in.Start; i <= in.End; i++ {
		op, err := GetL2QueueIdOp(z.GetStateDB(), i)
		if err != nil {
			return nil, errors.Wrapf(err, "id=%d", i)
		}
		batch.Ops = append(batch.Ops, op)
	}
	return &batch, nil
}

// Query_GetL2LastQueueId get l2 last queue id
func (z *zksync) Query_GetL2LastQueueId(in *types.ReqNil) (types.Message, error) {
	lastId, err := GetL2LastQueueId(z.GetStateDB())
	return &types.Int64{Data: lastId}, err
}

// Query_GetProofId2QueueId get proof 's queue info
func (z *zksync) Query_GetProofId2QueueId(in *types.Int64) (types.Message, error) {
	if in == nil {
		return nil, errors.Wrapf(types.ErrInvalidParam, "id nil")
	}
	return GetProofId2QueueId(z.GetStateDB(), uint64(in.Data))
}

// Query_GetCommitProofById 根据proofId获取commitProof信息
func (z *zksync) Query_GetCommitProofById(in *zt.ZkQueryReq) (types.Message, error) {
	var proofInfo zt.QueryProofInfo
	table := NewCommitProofTable(z.GetLocalDB())
	row, err := table.GetData(getProofIdCommitProofKey(in.ProofId))
	if err != nil {
		return nil, err
	}
	proofInfo.Proof = row.Data.(*zt.ZkCommitProof)
	proofInfo.Queues, err = GetProofId2QueueId(z.GetStateDB(), in.ProofId)
	if err != nil {
		return nil, errors.Wrapf(err, "GetProofId2QueueId id=%d", in.ProofId)
	}

	return &proofInfo, nil
}

// Query_GetProofList 根据proofId fetch 后续证明
func (z *zksync) Query_GetProofList(in *zt.ZkFetchProofList) (types.Message, error) {

	table := NewCommitProofTable(z.GetLocalDB())

	if in.GetReqOnChainProof() {
		//升序
		rows, err := table.ListIndex("onChainId", []byte(fmt.Sprintf("%016d", in.OnChainProofId)), nil, 1, zt.ListASC)
		if err != nil {
			zklog.Error("Query_GetProofList.getOnChainSubId", "id", in.OnChainProofId, "err", err.Error())
			return nil, err
		}
		return rows[0].Data.(*zt.ZkCommitProof), nil
	}

	//按截止高度获取最新proof
	if in.GetReqLatestProof() {
		//降序获取到第一个小于等于endHeight的commitHeight proof
		rows, err := table.ListIndex("commitHeight", []byte(fmt.Sprintf("%016d", in.GetEndHeight())), nil, 1, zt.ListDESC)
		if err != nil {
			zklog.Error("Query_GetProofList.listCommitHeight", "endHeight", in.GetEndHeight())
			return nil, err
		}
		if len(rows) <= 0 {
			zklog.Error("Query_GetProofList.listCommitHeight not found", "endHeight", in.GetEndHeight())
			return nil, types.ErrNotFound
		}
		//如果获得的最新proofId大于请求的ProofId则返回，否则按ProofId获取下一个proof
		if rows[0].Data.(*zt.ZkCommitProof).ProofId > in.ProofId {
			return rows[0].Data.(*zt.ZkCommitProof), nil
		}
	}
	// 按序获取proofId
	rows, err := table.GetData(getProofIdCommitProofKey(in.ProofId))
	if err != nil {
		zklog.Error("Query_GetProofList.getProofId", "currentProofId", in.ProofId, "err", err.Error())
		return nil, err
	}
	return rows.Data.(*zt.ZkCommitProof), nil
}

func (z *zksync) Query_GetCurrentExodusMode(in *types.ReqNil) (types.Message, error) {
	var mode types.Int64
	data, err := getExodusMode(z.GetStateDB())
	if err != nil {
		return nil, err
	}
	mode.Data = data
	return &mode, nil
}

func (z *zksync) Query_GetExistenceProof(in *zt.ZkReqExistenceProof) (types.Message, error) {
	return getAccountProofInHistory(z.GetStateDB(), in)
}

// Query_BuildHistoryAccounts 获取statedb中的tree账户信息构建merkel tree，返回tree roothash
func (z *zksync) Query_BuildHistoryAccounts(in *zt.CommitProofState) (types.Message, error) {
	if in == nil || in.ProofId == 0 {
		accts, err := BuildStateDbHistoryAccount(z.GetStateDB(), "")
		if err != nil {
			return nil, err
		}
		var resp types.ReplyString
		resp.Data = accts.RootHash
		return &resp, nil
	}
	feeAddrs, _ := z.Query_GetCfgFeeAddr(nil)

	accts, err := BuildHistoryAccountByProof(z.GetStateDB(), in.ProofId, in.NewTreeRoot, feeAddrs.(*zt.ZkFeeAddrs))
	if err != nil {
		return nil, err
	}
	var resp types.ReplyString
	resp.Data = accts.RootHash
	return &resp, nil

}

// Query_GetTotalDeposit 根据id获取当前symbol l2 全部存款
func (z *zksync) Query_GetTotalDeposit(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	totalBalance := new(big.Int)
	lastAccountID, err := getLatestAccountID(z.GetStateDB())
	if err != nil {
		return nil, errors.Wrapf(err, "getLatestAccountID")
	}
	for id := uint64(zt.SystemDefaultAcctId); id <= uint64(lastAccountID); id++ {
		token, _ := GetTokenByAccountIdAndTokenId(z.GetStateDB(), id, in.TokenId)
		if token != nil {
			balance, _ := new(big.Int).SetString(token.Balance, 10)
			totalBalance = new(big.Int).Add(totalBalance, balance)
		}
	}
	return &zt.TokenBalance{TokenId: in.TokenId, Balance: decimalVal(totalBalance.String(), in.Decimal)}, nil
}
