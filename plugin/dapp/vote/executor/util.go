package executor

import (
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

const (
	// IDLen length of groupID or voteID
	IDLen   = 65
	addrLen = 34
)

func formatGroupID(txHash string) string {
	return "g" + txHash
}

func formatVoteID(txHash string) string {
	return "v" + txHash
}

func checkMemberExist(addr string, members []*vty.GroupMember) bool {
	for _, item := range members {
		if addr == item.Addr {
			return true
		}
	}
	return false
}

func checkSliceItemExist(target string, items []string) bool {
	for _, item := range items {
		if target == item {
			return true
		}
	}
	return false
}

func checkSliceItemDuplicate(items []string) bool {

	filter := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, ok := filter[item]; ok {
			return true
		}
		filter[item] = struct{}{}
	}
	return false
}

func readStateDB(stateDB db.KV, key []byte, result types.Message) error {

	val, err := stateDB.Get(key)
	if err != nil {
		return err
	}
	return types.Decode(val, result)
}

func mustDecodeProto(data []byte, msg types.Message) {
	if err := types.Decode(data, msg); err != nil {
		panic(err.Error())
	}
}

func decodeGroupInfo(data []byte) *vty.GroupInfo {
	info := &vty.GroupInfo{}
	mustDecodeProto(data, info)
	return info
}

func decodeVoteInfo(data []byte) *vty.VoteInfo {
	info := &vty.VoteInfo{}
	mustDecodeProto(data, info)
	return info
}
