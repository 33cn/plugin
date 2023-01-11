package executor

import (
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	paratypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

func readStateDB(stateDB db.KV, key []byte, result types.Message) error {

	val, err := stateDB.Get(key)
	if err != nil {
		return err
	}
	return types.Decode(val, result)
}

func (r *rollup) getRollupStatus(title string) (*rolluptypes.RollupStatus, error) {

	status := &rolluptypes.RollupStatus{}
	err := readStateDB(r.GetStateDB(), formatRollupStatusKey(title), status)
	if err == types.ErrNotFound {
		return status, nil
	}
	return status, err
}

func (r *rollup) getRoundInfo(title string, round int64) (*rolluptypes.CommitRoundInfo, error) {

	info := &rolluptypes.CommitRoundInfo{}
	err := readStateDB(r.GetStateDB(), formatCommitRoundInfoKey(title, round), info)
	return info, err
}

func calcBlockHash(header *types.Header) string {
	return common.ToHex(common.Sha256(types.Encode(header)))
}

// 基于平行链质押逻辑
func (r *rollup) getValidatorNodesBlsPubs(title string) ([]string, error) {

	params := &paratypes.ReqParacrossNodeInfo{Title: title}
	resp, err := r.GetAPI().Query(paratypes.ParaX, "GetNodeGroupStatus", params)
	if err != nil {
		elog.Error("getValidatorNodesBlsPubs", "title", title, "err", err)
		return nil, err
	}

	status := resp.(*paratypes.ParaNodeGroupStatus)
	return strings.Split(status.BlsPubKeys, ","), nil
}
