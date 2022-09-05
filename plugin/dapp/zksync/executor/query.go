package executor

import (
	"fmt"
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
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

//Query_GetNFTId get nft id by content hash
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

// Query_GetTxProof 获取交易证明
func (z *zksync) Query_GetTxProof(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	table := NewZksyncInfoTable(z.GetLocalDB())
	row, err := table.GetData([]byte(fmt.Sprintf("%016d.%016d", in.GetBlockHeight(), in.GetTxIndex())))
	if err != nil {
		return nil, err
	}
	data := row.Data.(*zt.OperationInfo)
	return data, nil
}

// Query_GetTxProof 批量获取交易证明
func (z *zksync) Query_GetTxProofByHeight(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	res := new(zt.ZkQueryResp)
	datas := make([]*zt.OperationInfo, 0)
	table := NewZksyncInfoTable(z.GetLocalDB())
	rows, err := table.ListIndex("height", []byte(fmt.Sprintf("%016d", in.GetBlockHeight())), nil, 1000, zt.ListASC)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		data := row.Data.(*zt.OperationInfo)
		datas = append(datas, data)
	}
	res.OperationInfos = datas
	return res, nil
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
	leaf.EthAddress, ok = zt.DecimalAddr2Hex(leaf.GetEthAddress())
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "wrong eth addr format=%s", leaf.GetEthAddress())
	}
	leaf.Chain33Addr, ok = zt.DecimalAddr2Hex(leaf.GetChain33Addr())
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
	res.Leaves = leaves
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
	res.Leaves = leaves
	return res, nil
}

// Query_GetLastCommitProof 获取最新proof信息
func (z *zksync) Query_GetLastCommitProof(in *zt.ZkChainTitle) (types.Message, error) {
	//平行链缺省是1
	if z.GetAPI().GetConfig().IsPara() && in.GetChainTitleId() == 0 {
		return getLastCommitProofData(z.GetStateDB(), zt.ZkParaChainInnerTitleId)
	}
	//主链需要注明不同的平行链titleId
	if in.GetChainTitleId() <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "req chain id less or equal 0")
	}
	return getLastCommitProofData(z.GetStateDB(), new(big.Int).SetUint64(in.ChainTitleId).String())
}

//Query_GetLastOnChainProof 获取最新的包含OnChainPubData的Proof
func (z *zksync) Query_GetLastOnChainProof(in *zt.ZkChainTitle) (types.Message, error) {
	if in.GetChainTitleId() <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "req chain id less or equal 0")
	}
	return getLastOnChainProofData(z.GetStateDB(), new(big.Int).SetUint64(in.ChainTitleId).String())
}

// Query_GetLastPriorityQueueId 获取最后的eth priority queue id
func (z *zksync) Query_GetLastPriorityQueueId(in *types.ReqNil) (types.Message, error) {
	return getLastEthPriorityQueueID(z.GetStateDB())
}

// Query_GetMaxAccountId 获取当前最大账户id
func (z *zksync) Query_GetMaxAccountId(in *types.ReqNil) (types.Message, error) {
	var tree zt.AccountTree
	val, err := z.GetStateDB().Get(GetAccountTreeKey())
	if err != nil {
		return nil, err
	}
	err = types.Decode(val, &tree)
	if err != nil {
		return nil, err
	}
	var id types.Int64
	id.Data = int64(tree.GetTotalIndex()) - 1
	return &id, nil
}

func (z *zksync) Query_GetCurrentProof(in *zt.ZkReqExistenceProof) (types.Message, error) {
	info, err := getTreeUpdateInfo(z.GetStateDB())
	if err != nil {
		return nil, errors.Wrapf(err, "db.getTreeUpdateInfo")
	}
	proof, err := calProof(z.GetStateDB(), info, in.GetAccountId(), in.GetTokenId())
	if err != nil {
		return nil, errors.Wrapf(err, "calcProof")
	}
	var witness zt.ZkProofWitness
	witness.AccountWitness = new(zt.AccountWitness)
	witness.TokenWitness = new(zt.TokenWitness)
	witness.TreeRoot = proof.GetTreeProof().RootHash
	witness.AccountWitness.ID = proof.GetLeaf().GetAccountId()
	witness.AccountWitness.EthAddr = proof.GetLeaf().GetEthAddress()
	witness.AccountWitness.Chain33Addr = proof.GetLeaf().GetChain33Addr()
	witness.AccountWitness.TokenTreeRoot = proof.GetTokenProof().GetRootHash()
	witness.AccountWitness.PubKey = proof.GetLeaf().PubKey
	witness.AccountWitness.Sibling = &zt.SiblingPath{
		Path:   proof.GetTreeProof().GetProofSet(),
		Helper: proof.GetTreeProof().GetHelpers(),
	}
	witness.AccountWitness.ProxyPubKeys = proof.GetLeaf().GetProxyPubKeys()
	witness.TokenWitness.ID = proof.GetToken().GetTokenId()
	witness.TokenWitness.Balance = proof.GetToken().GetBalance()
	witness.TokenWitness.Sibling = &zt.SiblingPath{
		Path:   proof.GetTokenProof().GetProofSet(),
		Helper: proof.GetTokenProof().GetHelpers(),
	}
	zklog.Info("Query_GetCurrentProof", "leafTokenRoot", proof.GetLeaf().GetTokenHash(), "tokenroot", proof.GetTokenProof().GetRootHash())
	return &witness, nil
}

//Query_GetExistenceProof 获取指定tree root上某accountId,tokenId对应的存在证明
func (z *zksync) Query_GetExistenceProof(in *zt.ZkReqExistenceProof) (types.Message, error) {
	if len(in.GetRootHash()) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "roothash is nil")
	}

	if in.GetChainTitleId() <= 0 {
		chainTitleId, _ := strconv.Atoi(zt.ZkParaChainInnerTitleId)
		in.ChainTitleId = uint64(chainTitleId)
	}
	return getAccountProofInHistory(z.GetLocalDB(), in)
}

//Query_GetHistoryAccountProofInfo 查询historyAccount leaves, 特别是预先设置historyProof变量(计算比较久)，为证明做准备
func (z *zksync) Query_GetHistoryAccountProofInfo(in *zt.ZkReqExistenceProof) (types.Message, error) {
	if len(in.GetRootHash()) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "roothash is nil")
	}
	chainTitleId, _ := strconv.Atoi(zt.ZkParaChainInnerTitleId)
	if in.GetChainTitleId() > 0 {
		chainTitleId = int(in.GetChainTitleId())
	}
	info, err := getHistoryAccountByRoot(z.GetLocalDB(), uint64(chainTitleId), in.GetRootHash())
	if err != nil {
		return nil, err
	}
	var rsp zt.HistoryAccountProofInfo
	rsp.RootHash = info.RootHash
	if in.AccountId > 0 {
		for _, v := range info.Leaves {
			if v.AccountId == in.AccountId {
				rsp.Leaves = append(rsp.Leaves, v)
			}
		}
		return &rsp, nil
	}
	rsp.Leaves = info.Leaves
	return &rsp, nil
}

//Query_GetTreeInitRoot 获取系统初始tree root
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

//Query_GetCfgFeeAddr 获取系统初始fee addr
func (z *zksync) Query_GetCfgFeeAddr(in *types.ReqNil) (types.Message, error) {
	eth, l2 := getCfgFeeAddr(z.GetAPI().GetConfig())
	return &zt.ZkFeeAddrs{EthFeeAddr: eth, L2FeeAddr: l2}, nil
}

//Query_GetCfgTokenFee 获取系统配置的fee
func (z *zksync) Query_GetCfgTokenFee(in *zt.ZkSetFee) (types.Message, error) {
	amount, err := getDbFeeData(z.GetStateDB(), in.GetActionTy(), in.GetTokenId())
	if err != nil {
		return nil, err
	}
	return &types.ReplyString{Data: amount}, nil
}

//Query_GetVerifiers 获取系统初始fee addr
func (z *zksync) Query_GetVerifiers(in *zt.ZkChainTitle) (types.Message, error) {
	if in.GetChainTitleId() <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "chainTitleId=%d", in.GetChainTitleId())
	}
	return getVerifierData(z.GetStateDB(), new(big.Int).SetUint64(in.GetChainTitleId()).String())
}

// Query_GetTxOperationByOffSetOrCount 根据起始高度批量获取交易证明
// 1、指定count，实现快速与精确偏移，获取交易快照操作。
// 2、指定blockOffSet，高度差偏移，实现高度交易操作打包，获取交易快照操作。
func (z *zksync) Query_GetTxOperationByOffSetOrCount(in *zt.ZkQueryTxOperationReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	startHeight := in.GetStartBlockHeight()
	if z.GetHeight() <= int64(startHeight) || z.GetHeight() <= int64(in.Maturity) {
		return new(zt.ZkQueryProofResp), nil
	}
	endBlockHeight := z.GetHeight() - int64(in.Maturity)

	if int64(startHeight) >= endBlockHeight {
		return new(zt.ZkQueryProofResp), nil
	}

	startIndexTx := in.GetStartIndex()
	startOpIndex := in.OpIndex

	res := new(zt.ZkQueryProofResp)
	ops := make([]*zt.OperationInfo, 0)
	table := NewZksyncInfoTable(z.GetLocalDB())

	if in.Count > 0 {
		totalCount := int(in.Count)
	END:
		for i := startHeight; i < uint64(endBlockHeight); i++ {
			if totalCount <= 0 {
				break END
			}
			var primaryKey []byte
			if i == startHeight && startIndexTx != 0 {
				primaryKey = []byte(fmt.Sprintf("%016d.%016d.%016d", i, startIndexTx, startOpIndex))
			} else {
				primaryKey = nil
			}

			rows, err := table.ListIndex("height", []byte(fmt.Sprintf("%016d", i)), primaryKey, types.MaxTxsPerBlock, zt.ListASC)
			if err != nil {
				if isNotFound(err) {
					continue
				} else {
					return nil, err
				}
			}
			for _, row := range rows {
				data := row.Data.(*zt.OperationInfo)
				chunk, err := zt.GetOpChunkNum(data.TxType)
				if err != nil {
					zklog.Error("GetTxOperationByOffSetOrCount.GetOpChunk", "ty", data.TxType)
					return nil, types.ErrInvalidParam
				}
				if totalCount < chunk {
					break END
				}
				ops = append(ops, data)
				totalCount -= chunk
			}
		}
		res.OperationInfos = ops
		return res, nil
	}

	//2. 按高度offset查找
	if endBlockHeight > int64(startHeight)+int64(in.BlockOffset) {
		endBlockHeight = int64(startHeight) + int64(in.BlockOffset)
	} else {
		if int64(startHeight) > endBlockHeight {
			endBlockHeight = int64(startHeight)
		}
	}

	return z.Query_GetTxProofByHeights(&zt.ZkQueryProofReq{
		NeedDetail:       false,
		StartBlockHeight: startHeight,
		EndBlockHeight:   uint64(endBlockHeight),
		StartIndex:       startHeight,
		OpIndex:          startOpIndex,
	})
}

// Query_GetTxProofByHeights 根据多个高度批量获取交易证明
func (z *zksync) Query_GetTxProofByHeights(in *zt.ZkQueryProofReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	res := new(zt.ZkQueryProofResp)
	datas := make([]*zt.OperationInfo, 0)
	table := NewZksyncInfoTable(z.GetLocalDB())
	for i := in.GetStartBlockHeight(); i <= in.GetEndBlockHeight(); i++ {
		var primaryKey []byte
		if i == in.GetStartBlockHeight() && in.GetStartIndex() != 0 {
			primaryKey = []byte(fmt.Sprintf("%016d.%016d.%016d", i, in.GetStartIndex(), in.OpIndex))
		} else {
			primaryKey = nil
		}
		rows, err := table.ListIndex("height", []byte(fmt.Sprintf("%016d", i)), primaryKey, types.MaxTxsPerBlock, zt.ListASC)
		if err != nil {
			if isNotFound(err) {
				continue
			} else {
				return nil, err
			}
		}
		for _, row := range rows {
			data := row.Data.(*zt.OperationInfo)
			if in.GetNeedDetail() {
				datas = append(datas, data)
			} else {
				info := new(zt.OperationInfo)
				info.BlockHeight = i
				info.TxType = data.TxType
				datas = append(datas, info)
			}
		}
	}
	res.OperationInfos = datas
	return res, nil
}

// Query_GetFirstOnChainOp 根据proofId获取在此proof之后的第一个onChainTx,为提供无效交易做准备
func (z *zksync) Query_GetFirstOnChainOp(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil || in.GetProofId() <= 0 {
		return nil, types.ErrInvalidParam
	}
	chainTitleId, _ := strconv.Atoi(zt.ZkParaChainInnerTitleId)
	if in.GetChainTitleId() > 0 {
		chainTitleId = int(in.GetChainTitleId())
	}
	reqProof := &zt.ZkFetchProofList{
		ProofId:      in.GetProofId(),
		ChainTitleId: uint64(chainTitleId),
	}
	rep, err := z.Query_GetProofList(reqProof)
	if err != nil {
		return nil, errors.Wrapf(err, "getProofId")
	}
	proof := rep.(*zt.ZkCommitProof)
	onChainOpTy := []uint32{zt.TyDepositAction, zt.TyWithdrawAction, zt.TyProxyExitAction, zt.TyFullExitAction, zt.TyWithdrawNFTAction}
	txTypeMap := make(map[uint32]bool)
	//如果参数提供了OpType,就采用提供的
	if in.GetOpType() > 0 {
		txTypeMap[in.GetOpType()] = true
	} else {
		for _, t := range onChainOpTy {
			txTypeMap[t] = true
		}
	}
	var txProofResp zt.ZkTxProofResp
	table := NewZksyncInfoTable(z.GetLocalDB())
OuterLoop:
	for i := proof.BlockEnd; i <= uint64(z.GetHeight()); i++ {
		var primaryKey []byte
		if i == proof.GetBlockEnd() && proof.GetIndexEnd() != 0 {
			//获取起始搜索的 height,txIndex,opIndex
			primaryKey = []byte(fmt.Sprintf("%016d.%016d.%016d", i, proof.GetIndexEnd(), proof.GetOpIndex()))
		} else {
			primaryKey = nil
		}
		rows, err := table.ListIndex("height", []byte(fmt.Sprintf("%016d", i)), primaryKey, types.MaxTxsPerBlock, zt.ListASC)
		if err != nil {
			if isNotFound(err) {
				continue
			} else {
				return nil, err
			}
		}
		for _, row := range rows {
			data := row.Data.(*zt.OperationInfo)
			if txTypeMap[data.TxType] {
				txProofResp.BlockHeight = data.BlockHeight
				txProofResp.TxIndex = data.TxIndex
				txProofResp.OpIndex = data.OpIndex
				txProofResp.TxType = data.TxType
				txProofResp.TxHash = data.TxHash
				break OuterLoop
			}
		}
	}

	lastProof, err := getLastCommitProofData(z.GetStateDB(), strconv.Itoa(chainTitleId))
	if err != nil {
		zlog.Error("GetFirstOnChainOp.getLastProof", "err", err)
		return &txProofResp, nil
	}
	for i := in.GetProofId() + 1; i <= lastProof.ProofId; i++ {

		reqProof = &zt.ZkFetchProofList{
			ProofId:      i,
			ChainTitleId: uint64(chainTitleId),
		}
		rep, err = z.Query_GetProofList(reqProof)
		if err != nil {
			return &txProofResp, errors.Wrapf(err, "fetchProofId")
		}
		proof = rep.(*zt.ZkCommitProof)
		//只有包含onChainTx的proof才会有对应OnChain交易
		if proof.GetOnChainProofId() == 0 {
			continue
		}
		//txOp 需要在proof的证明范围内
		//proof 的blockStart==blockEnd时候，需要检查txIndex
		if proof.BlockStart == proof.BlockEnd && txProofResp.BlockHeight == proof.BlockEnd &&
			uint32(proof.IndexStart) <= txProofResp.TxIndex && txProofResp.TxIndex <= uint32(proof.IndexEnd) {
			txProofResp.ProofId = proof.ProofId
			txProofResp.ProofNewRoot = proof.NewTreeRoot
			return &txProofResp, nil
		}
		//proof blockStart!=blockEnd, 只需要检查tx的blockHeight是否在高度区间内
		if proof.BlockStart != proof.BlockEnd && proof.BlockStart <= txProofResp.BlockHeight && txProofResp.BlockHeight <= proof.BlockEnd {
			txProofResp.ProofId = proof.ProofId
			txProofResp.ProofNewRoot = proof.NewTreeRoot
			return &txProofResp, nil
		}
	}

	return &txProofResp, nil
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

// Query_GetTokenBalance 根据token和account获取balance
func (z *zksync) Query_GetTokenBalance(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	res := new(zt.ZkQueryResp)
	token, err := GetTokenByAccountIdAndTokenIdInDB(z.GetStateDB(), in.AccountId, in.TokenId)
	if err != nil {
		return nil, err
	}
	res.TokenBalances = append(res.TokenBalances, token)
	return res, nil
}

// Query_GetPriorityOpInfo 根据priorityId获取operation信息
func (z *zksync) Query_GetPriorityOpInfo(in *zt.EthPriorityQueueID) (types.Message, error) {
	if len(in.GetID()) == 0 {
		return nil, types.ErrInvalidParam
	}
	table := NewZksyncInfoTable(z.GetLocalDB())
	rows, err := table.ListIndex("priorityId", []byte(fmt.Sprintf("%s", in.GetID())), nil, 1, zt.ListASC)
	if err != nil {
		return nil, errors.Wrapf(err, "listIndex")
	}
	if len(rows) < 1 {
		return nil, types.ErrNotFound
	}
	return rows[0].Data.(*zt.OperationInfo), nil
}

// Query_GetProofByTxHash 根据txhash获取proof信息
func (z *zksync) Query_GetProofByTxHash(in *zt.ZkQueryReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	res := new(zt.ZkQueryResp)
	datas := make([]*zt.OperationInfo, 0)
	table := NewZksyncInfoTable(z.GetLocalDB())
	rows, err := table.ListIndex("txHash", []byte(fmt.Sprintf("%s", in.GetTxHash())), nil, 1, zt.ListASC)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		data := row.Data.(*zt.OperationInfo)
		datas = append(datas, data)
	}
	res.OperationInfos = datas
	return res, nil
}

// Query_GetCommitProofById 根据proofId获取commitProof信息
func (z *zksync) Query_GetCommitProofById(in *zt.ZkQueryReq) (types.Message, error) {
	if in.GetChainTitleId() == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "chain title not set")
	}

	table := NewCommitProofTable(z.GetLocalDB())
	row, err := table.GetData(getProofIdCommitProofKey(new(big.Int).SetUint64(in.GetChainTitleId()).String(), in.ProofId))
	if err != nil {
		return nil, err
	}
	data := row.Data.(*zt.ZkCommitProof)

	return data, nil
}

// Query_GetProofChainTitleList 获取所有chainTitle信息
func (z *zksync) Query_GetProofChainTitleList(in *types.ReqNil) (types.Message, error) {

	table := NewCommitProofTable(z.GetLocalDB())
	//只查找有proofId=1的记录，再统计
	rows, err := table.ListIndex("proofId", []byte(fmt.Sprintf("%016d", 1)), nil, 0, zt.ListASC)
	if err != nil {
		zklog.Error("Query_GetProofChainTitleList", "err", err.Error())
		return nil, err
	}
	var chains zt.ZkChainTitleList
	for _, r := range rows {
		chain := &zt.ZkChainTitle{
			ChainTitleId: r.Data.(*zt.ZkCommitProof).GetChainTitleId(),
			ChainTitle:   r.Data.(*zt.ZkCommitProof).GetChainTitle(),
		}
		chains.Chains = append(chains.Chains, chain)
	}
	return &chains, nil

}

// Query_GetProofList 根据proofId fetch 后续证明
func (z *zksync) Query_GetProofList(in *zt.ZkFetchProofList) (types.Message, error) {
	if in.GetChainTitleId() <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "chain title not set")
	}

	table := NewCommitProofTable(z.GetLocalDB())

	if in.GetReqOnChainProof() {
		//升序
		rows, err := table.ListIndex("onChainId", []byte(fmt.Sprintf("%s-%016d", new(big.Int).SetUint64(in.ChainTitleId).String(), in.OnChainProofId)), nil, 1, zt.ListASC)
		if err != nil {
			zklog.Error("Query_GetProofList.getOnChainSubId", "id", in.OnChainProofId, "err", err.Error())
			return nil, err
		}
		return rows[0].Data.(*zt.ZkCommitProof), nil
	}

	//按截止高度获取最新proof
	if in.GetReqLatestProof() {
		//降序获取到第一个小于等于endHeight的commitHeight proof
		rows, err := table.ListIndex("commitHeight", []byte(fmt.Sprintf("%s-%016d", new(big.Int).SetUint64(in.ChainTitleId).String(), in.GetEndHeight())), nil, 1, zt.ListDESC)
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
	rows, err := table.GetData(getProofIdCommitProofKey(new(big.Int).SetUint64(in.ChainTitleId).String(), in.ProofId))
	if err != nil {
		zklog.Error("Query_GetProofList.getProofId", "currentProofId", in.ProofId, "err", err.Error())
		return nil, err
	}
	return rows.Data.(*zt.ZkCommitProof), nil
}
