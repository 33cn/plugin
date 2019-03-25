// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"

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

// Query_GetTitleByHash query paracross title by block hash
func (p *Paracross) Query_GetTitleByHash(in *pt.ReqParacrossTitleHash) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}

	if !types.IsDappFork(p.GetMainHeight(), pt.ParaX, pt.ForkCommitTx) {
		block, err := p.GetAPI().GetBlockOverview(&types.ReqHash{Hash: in.BlockHash})
		if err != nil || block == nil {
			return nil, types.ErrHashNotExist
		}
		return p.paracrossGetHeight(in.GetTitle())
	}

	return p.paracrossGetHeightByHash(in)

}

//Query_ListTitles query paracross titles list
func (p *Paracross) Query_ListTitles(in *types.ReqNil) (types.Message, error) {
	return p.paracrossListTitles()
}

// Query_GetTitleHeight query title height
func (p *Paracross) Query_GetTitleHeight(in *pt.ReqParacrossTitleHeight) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return p.paracrossGetTitleHeight(in.Title, in.Height)
}

// Query_GetAssetTxResult query get asset tx reseult
func (p *Paracross) Query_GetAssetTxResult(in *types.ReqHash) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return p.paracrossGetAssetTxResult(in.Hash)
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

func (p *Paracross) paracrossGetHeightByHash(in *pt.ReqParacrossTitleHash) (types.Message, error) {
	ret, err := getTitle(p.GetStateDB(), calcTitleHashKey(in.GetTitle(), hex.EncodeToString(in.GetBlockHash())))
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
		resp.Titles = append(resp.Titles, &st)
	}
	return &resp, nil
}

func loadLocalTitle(db dbm.KV, title string, height int64) (types.Message, error) {
	key := calcLocalTitleHeightKey(title, height)
	res, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	var resp pt.ReceiptParacrossDone
	err = types.Decode(res, &resp)
	if err != nil {
		panic(err)
	}
	return &resp, nil
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

	var result pt.ParacrossAsset
	err = types.Decode(value, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
