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

// GetRollupStatus get rollup status
func GetRollupStatus(kv db.KV, title string) (*rolluptypes.RollupStatus, error) {

	status := &rolluptypes.RollupStatus{}
	err := readStateDB(kv, formatRollupStatusKey(title), status)
	if err == types.ErrNotFound {
		return status, nil
	}
	return status, err
}

// GetRoundInfo get round info
func GetRoundInfo(kv db.KV, title string, round int64) (*rolluptypes.CommitRoundInfo, error) {

	info := &rolluptypes.CommitRoundInfo{}
	err := readStateDB(kv, formatCommitRoundInfoKey(title, round), info)
	return info, err
}

func sha256Hash(h *types.Header) []byte{
	return common.Sha256(types.Encode(h))
}

func calcBlockHash(h *types.Header) string {
	return common.ToHex(sha256Hash(h))
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
