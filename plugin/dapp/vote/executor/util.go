package executor

import (
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

const (
	voteStatusNormal = iota
	voteStatusClosed
)

const (
	// IDLen length of groupID or voteID
	IDLen   = 19
	addrLen = 34
)

func formatGroupID(id string) string {
	return "g" + id
}

func formatVoteID(id string) string {
	return "v" + id
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

func decodeCommitInfo(data []byte) *vty.CommitInfo {
	info := &vty.CommitInfo{}
	mustDecodeProto(data, info)
	return info
}

func classifyVoteList(infos *vty.VoteInfos) *vty.ReplyVoteList {

	reply := &vty.ReplyVoteList{}
	currentTime := types.Now().Unix()
	for _, voteInfo := range infos.GetVoteList() {

		if voteInfo.Status == voteStatusClosed {
			reply.ClosedList = append(reply.ClosedList, voteInfo)
		} else if voteInfo.BeginTimestamp > currentTime {
			reply.PendingList = append(reply.PendingList, voteInfo)
		} else if voteInfo.EndTimestamp > currentTime {
			reply.OngoingList = append(reply.OngoingList, voteInfo)
		} else {
			reply.FinishedList = append(reply.FinishedList, voteInfo)
		}
	}
	reply.CurrentTimestamp = currentTime
	return reply
}
