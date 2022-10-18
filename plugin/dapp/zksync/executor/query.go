package executor

import (
	"fmt"
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
	"math/big"
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
	leaf.EthAddress, _ = zt.DecimalAddr2Hex(leaf.GetEthAddress())
	leaf.Chain33Addr, _ = zt.DecimalAddr2Hex(leaf.GetChain33Addr())
	return &leaf, nil
}

// Query_GetAccountByEth  通过eth地址查询account
func (z *zksync) Query_GetAccountByEth(in *zt.ZkQueryReq) (types.Message, error) {
	res := new(zt.ZkQueryResp)
	in.EthAddress, _ = zt.HexAddr2Decimal(in.EthAddress)
	leaves, err := GetLeafByEthAddress(z.GetLocalDB(), in.EthAddress)
	if err != nil {
		return nil, err
	}
	res.Leaves = leaves
	return res, nil
}

// Query_GetAccountByChain33  通过chain33地址查询account
func (z *zksync) Query_GetAccountByChain33(in *zt.ZkQueryReq) (types.Message, error) {
	res := new(zt.ZkQueryResp)
	in.Chain33Addr, _ = zt.HexAddr2Decimal(in.Chain33Addr)
	leaves, err := GetLeafByChain33Address(z.GetLocalDB(), in.Chain33Addr)
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
func (z *zksync) Query_GetLastPriorityQueueId(in *types.Int64) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return getLastEthPriorityQueueID(z.GetStateDB(), uint32(in.Data))
}

//Query_GetTreeInitRoot 获取系统初始tree root
func (z *zksync) Query_GetTreeInitRoot(in *types.ReqAddrs) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	var eth, chain33 string
	if len(in.Addrs) == 2 {
		eth = in.Addrs[0]
		chain33 = in.Addrs[1]
	}

	root := getInitTreeRoot(z.GetAPI().GetConfig(), eth, chain33)
	return &types.ReplyString{Data: root}, nil
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

func (z *zksync) Query_GetQueueID(in *types.ReqNil) (types.Message, error) {
	ethPriorityQueueID, err := getLastEthPriorityQueueID(z.GetStateDB(), 0)
	return ethPriorityQueueID, err
}
