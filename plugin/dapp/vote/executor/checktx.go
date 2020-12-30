package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
)

// CheckTx 实现自定义检验交易接口，供框架调用
func (v *vote) CheckTx(tx *types.Transaction, index int) error {
	// implement code

	txHash := hex.EncodeToString(tx.Hash())
	var action vty.VoteAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		elog.Error("vote CheckTx", "txHash", txHash, "Decode payload error", err)
		return types.ErrActionNotSupport
	}
	if action.Ty == vty.TyCreateGroupAction {
		err = v.checkCreateGroup(action.GetCreateGroup())
	} else if action.Ty == vty.TyUpdateMemberAction {
		err = v.checkUpdateMember(action.GetUpdateMember())
	} else if action.Ty == vty.TyCreateVoteAction {
		err = v.checkCreateVote(action.GetCreateVote())
	} else if action.Ty == vty.TyCommitVoteAction {
		err = v.checkCommitVote(action.GetCommitVote(), tx, index)
	} else {
		err = types.ErrActionNotSupport
	}

	if err != nil {
		elog.Error("vote CheckTx", "txHash", txHash, "actionName", tx.ActionName(), "err", err, "actionData", action)
	}
	return err
}

func checkMemberValidity(members []*vty.GroupMember) error {
	filter := make(map[string]struct{}, len(members))
	for _, member := range members {
		if member.GetAddr() == "" {
			return errNilMember
		}
		if _, ok := filter[member.Addr]; ok {
			return errDuplicateMember
		}
		filter[member.Addr] = struct{}{}
	}
	return nil
}

func (v *vote) checkCreateGroup(create *vty.CreateGroup) error {

	if create.GetName() == "" {
		return errEmptyName
	}

	//检测组成员是否有重复
	if err := checkMemberValidity(create.GetMembers()); err != nil {
		return err
	}
	//检测管理员是否有重复
	if checkSliceItemDuplicate(create.GetAdmins()) {
		return errDuplicateAdmin
	}

	return nil
}

func (v *vote) checkUpdateMember(update *vty.UpdateMember) error {

	if len(update.GetGroupID()) != IDLen {
		return errInvalidGroupID
	}
	for _, member := range update.AddMembers {
		if len(member.Addr) != addrLen {
			return types.ErrInvalidAddress
		}
	}

	for _, addr := range update.RemoveMemberAddrs {
		if len(addr) != addrLen {
			return types.ErrInvalidAddress
		}
	}

	return nil
}

func (v *vote) checkCreateVote(create *vty.CreateVote) error {

	if create.GetName() == "" {
		return errEmptyName
	}

	if create.GetEndTimestamp() <= v.GetBlockTime() || create.EndTimestamp <= create.BeginTimestamp {
		return errInvalidVoteTime
	}

	if len(create.VoteOptions) < 2 {
		return errInvalidVoteOption
	}

	if len(create.VoteGroups) == 0 {
		return errEmptyVoteGroup
	}

	if checkSliceItemDuplicate(create.VoteGroups) {
		return errDuplicateGroup
	}

	return nil
}

func (v *vote) checkCommitVote(commit *vty.CommitVote, tx *types.Transaction, index int) error {

	voteID := commit.GetVoteID()
	groupID := commit.GetGroupID()
	if len(voteID) != IDLen {
		return errInvalidVoteID
	}
	if len(groupID) != IDLen {
		return errInvalidGroupID
	}
	action := newAction(v, tx, index)
	voteInfo, err := action.getVoteInfo(voteID)
	if err != nil {
		return err
	}

	if voteInfo.EndTimestamp <= action.blockTime {
		return errVoteAlreadyFinished
	}

	if commit.OptionIndex >= uint32(len(voteInfo.VoteOptions)) {
		return errInvalidOptionIndex
	}

	//check group validity
	if !checkSliceItemExist(commit.GroupID, voteInfo.VoteGroups) {
		return errInvalidGroupID
	}

	// check if already vote
	if checkSliceItemExist(action.fromAddr, voteInfo.VotedMembers) {
		return errAlreadyVoted
	}

	groupInfo, err := action.getGroupInfo(commit.GroupID)
	if err != nil {
		return err
	}

	// check if exist in group members
	if !checkMemberExist(action.fromAddr, groupInfo.Members) {
		return errInvalidGroupMember
	}

	return nil
}
