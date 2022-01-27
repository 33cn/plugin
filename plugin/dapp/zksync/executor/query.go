package executor

import (
	"fmt"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
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
	return &leaf, nil
}

// Query_GetAccountByEth  通过eth地址查询account
func (z *zksync) Query_GetAccountByEth(in *zt.ZkQueryReq) (types.Message, error) {
	res := new(zt.ZkQueryResp)
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
	leaves, err := GetLeafByChain33Address(z.GetLocalDB(), in.Chain33Addr)
	if err != nil {
		return nil, err
	}
	res.Leaves = leaves
	return res, nil
}

// Query_GetLastCommitProof 获取最新proof信息
func (z *zksync) Query_GetLastCommitProof(in *types.ReqNil) (types.Message, error) {
	return getLastCommitProofData(z.GetStateDB())
}