// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"fmt"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

// Query_GetTitle query paracross title
func (p *Paracross) Query_GetTitle(in *types.ReqString) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return p.paracrossGetHeight(in.GetData())
}

// Query_GetTitleHeight query paracross status with title and height
func (p *Paracross) Query_GetTitleHeight(in *pt.ReqParacrossTitleHeight) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	stat, err := p.paracrossGetStateTitleHeight(in.Title, in.Height)
	if err != nil {
		clog.Error("paracross.GetTitleHeight", "title", title, "height", in.Height, "err", err.Error())
		return nil, err
	}
	status := stat.(*pt.ParacrossHeightStatus)
	res := &pt.ParacrossHeightStatusRsp{
		Status:     status.Status,
		Title:      status.Title,
		Height:     status.Height,
		MainHeight: status.MainHeight,
		MainHash:   common.ToHex(status.MainHash),
	}
	for i, addr := range status.Details.Addrs {
		res.CommitAddrs = append(res.CommitAddrs, addr)
		res.CommitBlockHash = append(res.CommitBlockHash, common.ToHex(status.Details.BlockHash[i]))
	}
	return res, nil
}

// Query_GetTitleByHash query paracross title by block hash
func (p *Paracross) Query_GetTitleByHash(in *pt.ReqParacrossTitleHash) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}

	block, err := p.GetAPI().GetBlockOverview(&types.ReqHash{Hash: in.BlockHash})
	if err != nil || block == nil {
		return nil, types.ErrHashNotExist
	}
	return p.paracrossGetHeight(in.GetTitle())
}

//Query_GetNodeGroupAddrs get node group addrs
func (p *Paracross) Query_GetNodeGroupAddrs(in *pt.ReqParacrossNodeInfo) (types.Message, error) {
	if in == nil || in.GetTitle() == "" {
		return nil, types.ErrInvalidParam
	}

	ret, key, err := getConfigNodes(p.GetStateDB(), in.GetTitle())
	if err != nil {
		return nil, errors.Cause(err)
	}
	var nodes []string
	for k := range ret {
		nodes = append(nodes, k)
	}
	var reply types.ReplyConfig
	reply.Key = string(key)
	reply.Value = fmt.Sprint(nodes)
	return &reply, nil
}

//Query_GetNodeAddrInfo get specific node addr info
func (p *Paracross) Query_GetNodeAddrInfo(in *pt.ReqParacrossNodeInfo) (types.Message, error) {
	if in == nil || in.Title == "" || in.Addr == "" {
		return nil, types.ErrInvalidParam
	}
	stat, err := getNodeAddr(p.GetStateDB(), in.Title, in.Addr)
	if err != nil {
		return nil, err
	}
	mainHeight, err := p.getMainHeight()
	if err != nil {
		return nil, err
	}
	cfg := p.GetAPI().GetConfig()
	if pt.IsParaForkHeight(cfg, mainHeight, pt.ForkLoopCheckCommitTxDone) {
		stat.QuitId = getParaNodeIDSuffix(stat.QuitId)
		stat.ProposalId = getParaNodeIDSuffix(stat.ProposalId)
	}
	return stat, nil
}

func (p *Paracross) getMainHeight() (int64, error) {
	mainHeight := p.GetMainHeight()
	cfg := p.GetAPI().GetConfig()
	if cfg.IsPara() {
		block, err := p.GetAPI().GetBlocks(&types.ReqBlocks{Start: p.GetHeight(), End: p.GetHeight()})
		if err != nil || block == nil || len(block.Items) == 0 {
			return -1, types.ErrBlockExist
		}
		mainHeight = block.Items[0].Block.MainHeight
	}
	return mainHeight, nil
}

//Query_GetNodeIDInfo get specific node addr info
func (p *Paracross) Query_GetNodeIDInfo(in *pt.ReqParacrossNodeInfo) (types.Message, error) {
	if in == nil || in.Title == "" || in.Id == "" {
		return nil, types.ErrInvalidParam
	}
	mainHeight, err := p.getMainHeight()
	if err != nil {
		return nil, err
	}
	cfg := p.GetAPI().GetConfig()
	stat, err := getNodeIDWithFork(cfg, p.GetStateDB(), in.Title, mainHeight, in.Id)
	if err != nil {
		return nil, err
	}

	if pt.IsParaForkHeight(cfg, mainHeight, pt.ForkLoopCheckCommitTxDone) {
		stat.Id = getParaNodeIDSuffix(stat.Id)
	}
	return stat, nil
}

//Query_ListNodeStatusInfo list node info by status
func (p *Paracross) Query_ListNodeStatusInfo(in *pt.ReqParacrossNodeInfo) (types.Message, error) {
	if in == nil || in.Title == "" {
		return nil, types.ErrInvalidParam
	}
	resp, err := listLocalNodeStatus(p.GetLocalDB(), in.Title, in.Status)
	if err != nil {
		return resp, err
	}
	mainHeight, err := p.getMainHeight()
	if err != nil {
		return nil, err
	}
	cfg := p.GetAPI().GetConfig()
	if !pt.IsParaForkHeight(cfg, mainHeight, pt.ForkLoopCheckCommitTxDone) {
		return resp, err
	}
	addrs := resp.(*pt.RespParacrossNodeAddrs)
	for _, id := range addrs.Ids {
		id.Id = getParaNodeIDSuffix(id.Id)
	}
	return resp, nil
}

//Query_GetNodeGroupStatus get specific node addr info
func (p *Paracross) Query_GetNodeGroupStatus(in *pt.ReqParacrossNodeInfo) (types.Message, error) {
	if in == nil || in.Title == "" {
		return nil, types.ErrInvalidParam
	}
	stat, err := getNodeGroupStatus(p.GetStateDB(), in.Title)
	if err != nil {
		return stat, err
	}
	mainHeight, err := p.getMainHeight()
	if err != nil {
		return nil, err
	}
	cfg := p.GetAPI().GetConfig()
	if pt.IsParaForkHeight(cfg, mainHeight, pt.ForkLoopCheckCommitTxDone) {
		stat.Id = getParaNodeIDSuffix(stat.Id)
	}
	return stat, nil
}

//Query_ListNodeGroupStatus list node info by status
func (p *Paracross) Query_ListNodeGroupStatus(in *pt.ReqParacrossNodeInfo) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	resp, err := listLocalNodeGroupStatus(p.GetLocalDB(), in.Status)
	if err != nil {
		return resp, err
	}
	mainHeight, err := p.getMainHeight()
	if err != nil {
		return nil, err
	}
	cfg := p.GetAPI().GetConfig()
	if pt.IsParaForkHeight(cfg, mainHeight, pt.ForkLoopCheckCommitTxDone) {
		addrs := resp.(*pt.RespParacrossNodeGroups)
		for _, id := range addrs.Ids {
			id.Id = getParaNodeIDSuffix(id.Id)
		}
	}

	return resp, nil
}

//Query_ListTitles query paracross titles list
func (p *Paracross) Query_ListTitles(in *types.ReqNil) (types.Message, error) {
	return p.paracrossListTitles()
}

// Query_GetDoneTitleHeight query title height
func (p *Paracross) Query_GetDoneTitleHeight(in *pt.ReqParacrossTitleHeight) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return p.paracrossGetTitleHeight(in.Title, in.Height)
}

// Query_GetAssetTxResult query get asset tx reseult
func (p *Paracross) Query_GetAssetTxResult(in *types.ReqString) (types.Message, error) {
	if in == nil || in.Data == "" {
		return nil, types.ErrInvalidParam
	}
	hash, err := common.FromHex(in.Data)
	if err != nil {
		return nil, errors.Wrap(err, "fromHex")
	}
	return p.paracrossGetAssetTxResult(hash)
}

// Query_GetMainBlockHash query get mainblockHash by tx
func (p *Paracross) Query_GetMainBlockHash(in *types.Transaction) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return p.paracrossGetMainBlockHash(in)
}

func (p *Paracross) paracrossGetMainBlockHash(tx *types.Transaction) (types.Message, error) {
	var paraAction pt.ParacrossAction
	err := types.Decode(tx.GetPayload(), &paraAction)
	if err != nil {
		return nil, err
	}
	if paraAction.GetTy() != pt.ParacrossActionMiner {
		return nil, types.ErrCoinBaseTxType
	}

	if paraAction.GetMiner() == nil {
		return nil, pt.ErrParaEmptyMinerTx
	}

	paraNodeStatus := paraAction.GetMiner().GetStatus()
	if paraNodeStatus == nil {
		return nil, types.ErrCoinBaseTxType
	}

	mainHashFromNode := paraNodeStatus.MainBlockHash

	return &types.ReplyHash{Hash: mainHashFromNode}, nil
}

func (p *Paracross) paracrossGetHeight(title string) (types.Message, error) {
	ret, err := getTitle(p.GetStateDB(), calcTitleKey(title))
	if err != nil {
		return nil, errors.Cause(err)
	}
	return ret, nil
}

func (p *Paracross) paracrossGetStateTitleHeight(title string, height int64) (types.Message, error) {
	ret, err := getTitleHeight(p.GetStateDB(), calcTitleHeightKey(title, height))
	if err != nil {
		return nil, errors.Cause(err)
	}
	return ret, nil
}

func (p *Paracross) paracrossListTitles() (types.Message, error) {
	return listLocalTitles(p.GetLocalDB())
}

func listLocalTitles(db dbm.KVDB) (types.Message, error) {
	prefix := calcLocalTitlePrefix()
	res, err := db.List(prefix, []byte(""), 0, 1)
	if err != nil {
		return nil, err
	}
	var resp pt.RespParacrossTitles
	for _, r := range res {
		var st pt.ReceiptParacrossDone
		err = types.Decode(r, &st)
		if err != nil {
			panic(err)
		}
		rst := &pt.RespParacrossDone{
			TotalNodes:     st.TotalNodes,
			TotalCommit:    st.TotalCommit,
			MostSameCommit: st.MostSameCommit,
			Title:          st.Title,
			Height:         st.Height,
			TxResult:       string(st.TxResult),
		}

		resp.Titles = append(resp.Titles, rst)
	}
	return &resp, nil
}

func listNodeStatus(db dbm.KVDB, prefix []byte) (types.Message, error) {
	res, err := db.List(prefix, []byte(""), 0, 1)
	if err != nil {
		return nil, err
	}

	var resp pt.RespParacrossNodeAddrs
	for _, r := range res {
		var st pt.ParaNodeIdStatus
		err = types.Decode(r, &st)
		if err != nil {
			panic(err)
		}
		resp.Ids = append(resp.Ids, &st)
	}
	return &resp, nil
}

func listNodeGroupStatus(db dbm.KVDB, prefix []byte) (types.Message, error) {
	res, err := db.List(prefix, []byte(""), 0, 1)
	if err != nil {
		return nil, err
	}

	var resp pt.RespParacrossNodeGroups
	for _, r := range res {
		var st pt.ParaNodeGroupStatus
		err = types.Decode(r, &st)
		if err != nil {
			panic(err)
		}
		resp.Ids = append(resp.Ids, &st)
	}
	return &resp, nil
}

//按状态遍历
func listLocalNodeStatus(db dbm.KVDB, title string, status int32) (types.Message, error) {
	var prefix []byte
	if status == 0 {
		prefix = calcLocalNodeTitlePrefix(title)
	} else {
		prefix = calcLocalNodeStatusPrefix(title, status)
	}

	return listNodeStatus(db, prefix)
}

func listLocalNodeGroupStatus(db dbm.KVDB, status int32) (types.Message, error) {
	var prefix []byte
	if status == 0 {
		prefix = calcLocalNodeGroupAllPrefix()
	} else {
		prefix = calcLocalNodeGroupStatusPrefix(status)
	}

	return listNodeGroupStatus(db, prefix)
}

func loadLocalTitle(db dbm.KV, title string, height int64) (types.Message, error) {
	key := calcLocalHeightKey(title, height)
	res, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	var st pt.ReceiptParacrossDone
	err = types.Decode(res, &st)
	if err != nil {
		panic(err)
	}

	return &pt.RespParacrossDone{
		TotalNodes:      st.TotalNodes,
		TotalCommit:     st.TotalCommit,
		MostSameCommit:  st.MostSameCommit,
		Title:           st.Title,
		Height:          st.Height,
		ChainExecHeight: st.ChainExecHeight,
		TxResult:        hex.EncodeToString(st.TxResult),
	}, nil
}

func (p *Paracross) paracrossGetTitleHeight(title string, height int64) (types.Message, error) {
	return loadLocalTitle(p.GetLocalDB(), title, height)
}

func (p *Paracross) paracrossGetAssetTxResult(hash []byte) (types.Message, error) {
	if len(hash) == 0 {
		return nil, types.ErrInvalidParam
	}

	key := calcLocalAssetKey(hash)
	value, err := p.GetLocalDB().Get(key)
	if err != nil {
		return nil, err
	}

	var rst pt.ParacrossAsset
	err = types.Decode(value, &rst)
	if err != nil {
		return nil, err
	}

	rsp := &pt.ParacrossAssetRsp{
		From:             rst.From,
		To:               rst.To,
		Amount:           rst.Amount,
		Exec:             rst.Exec,
		Symbol:           rst.Symbol,
		Height:           rst.Height,
		CommitDoneHeight: rst.CommitDoneHeight,
		ParaHeight:       rst.ParaHeight,
	}

	rsp.TxHash = common.ToHex(rst.TxHash)
	rsp.IsWithdraw = "false"
	if rst.IsWithdraw {
		rsp.IsWithdraw = "true"
	}
	rsp.Success = "false"
	if rst.Success {
		rsp.Success = "true"
	}

	return rsp, nil
}

//Query_GetSelfConsStages get self consensus stages configed
func (p *Paracross) Query_GetSelfConsStages(in *types.ReqNil) (types.Message, error) {
	stages, err := getSelfConsensStages(p.GetStateDB())
	if err != nil {
		return nil, errors.Cause(err)
	}

	return stages, nil
}

//Query_GetSelfConsOneStage get self consensus one stage
func (p *Paracross) Query_GetSelfConsOneStage(in *types.Int64) (types.Message, error) {
	stage, err := getSelfConsOneStage(p.GetStateDB(), in.Data)
	if err != nil {
		return nil, err
	}
	return stage, nil
}

// Query_ListSelfStages 批量查询
func (p *Paracross) Query_ListSelfStages(in *pt.ReqQuerySelfStages) (types.Message, error) {
	return p.listSelfStages(in)
}
