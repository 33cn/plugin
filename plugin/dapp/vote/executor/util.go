package executor

import (
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

const (
	voteStatusNormal   = iota //非关闭常规状态
	voteStatusPending         //即将开始
	voteStatusOngoing         //正在进行
	voteStatusFinished        //已经结束
	voteStatusClosed          //已经关闭
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

func filterVoteWithStatus(voteList []*vty.VoteInfo, status uint32, currentTime int64) []*vty.VoteInfo {

	var filterList []*vty.VoteInfo
	for _, voteInfo := range voteList {

		if voteInfo.Status == voteStatusClosed {
		} else if voteInfo.BeginTimestamp > currentTime {
			voteInfo.Status = voteStatusPending
		} else if voteInfo.EndTimestamp > currentTime {
			voteInfo.Status = voteStatusOngoing
		} else {
			voteInfo.Status = voteStatusFinished
		}
		//remove vote info with other status
		if status == voteInfo.Status {
			filterList = append(filterList, voteInfo)
		}
	}
	//设置了状态筛选，返回对应的筛选列表
	if status > 0 {
		return filterList
	}
	return voteList
}
