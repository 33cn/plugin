package executor

import (
	"fmt"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
)

// Query_GetAccountTree 获取当前的树
func (z *zksync) Query_GetAccountTree(in *types.ReqNil) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return getAccountTree(z.GetStateDB(), nil)
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

	return &resp, nil
}

// Query_GetAccountByEth  通过eth地址查询account
func (z *zksync) Query_GetAccountByEth(in *zt.ZkQueryReq) (types.Message, error) {

}

// Query_GetAccountByChain33  通过chain33地址查询account
func (z *zksync) Query_GetAccountByChain33(in *zt.ZkQueryReq) (types.Message, error) {

}

// Query_GetLastCommitProof 获取最新proof信息
func (z *zksync) Query_GetLastCommitProof(in *types.ReqNil) (types.Message, error) {
	return getLastCommitProofData(z.GetStateDB())
}