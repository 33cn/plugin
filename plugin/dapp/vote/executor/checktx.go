package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/system/dapp"

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
	} else if action.Ty == vty.TyUpdateGroupAction {
		err = v.checkUpdateGroup(action.GetUpdateGroup(), tx, index)
	} else if action.Ty == vty.TyCreateVoteAction {
		err = v.checkCreateVote(action.GetCreateVote(), tx, index)
	} else if action.Ty == vty.TyCommitVoteAction {
		err = v.checkCommitVote(action.GetCommitVote(), tx, index)
	} else if action.Ty == vty.TyCloseVoteAction {
		err = v.checkCloseVote(action.GetCloseVote(), tx, index)
	} else if action.Ty == vty.TyUpdateMemberAction {
		err = v.checkUpdateMember(action.GetUpdateMember())
	} else {
		err = types.ErrActionNotSupport
	}

	if err != nil {
		elog.Error("vote CheckTx", "txHash", txHash, "actionName", tx.ActionName(), "err", err, "actionData", action.String())
	}
	return err
}

func checkMemberValidity(members []*vty.GroupMember) error {
	filter := make(map[string]struct{}, len(members))
	for _, member := range members {
		if member.GetAddr() == "" {
			return types.ErrInvalidAddress
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

func (v *vote) checkUpdateGroup(update *vty.UpdateGroup, tx *types.Transaction, index int) error {

	action := newAction(v, tx, index)
	groupInfo, err := action.getGroupInfo(update.GetGroupID())
	if err != nil {
		return err
	}
	if !checkSliceItemExist(action.fromAddr, groupInfo.GetAdmins()) {
		return errAddrPermissionDenied
	}

	//防止将管理员全部删除
	if len(update.RemoveAdmins) >= len(groupInfo.GetAdmins()) && len(update.AddAdmins) == 0 {
		return errAddrPermissionDenied
	}

	addrs := make([]string, 0, len(update.RemoveMembers)+len(update.AddAdmins)+len(update.RemoveAdmins))
	addrs = append(addrs, update.RemoveMembers...)
	addrs = append(addrs, update.AddAdmins...)
	addrs = append(addrs, update.RemoveAdmins...)
	for _, member := range update.AddMembers {
		if len(member.Addr) != addrLen {
			return types.ErrInvalidAddress
		}
	}

	for _, addr := range addrs {
		if len(addr) != addrLen {
			elog.Error("checkUpdateGroup", "invalid addr", addr)
			return types.ErrInvalidAddress
		}
	}

	//保证管理员地址合法性
	for _, addr := range update.GetAddAdmins() {
		if err := dapp.CheckAddress(v.GetAPI().GetConfig(), addr, v.GetHeight()); err != nil {
			elog.Error("checkUpdateGroup", "addr", addr, "CheckAddress err", err)
			return types.ErrInvalidAddress
		}
	}

	return nil
}

func (v *vote) checkCreateVote(create *vty.CreateVote, tx *types.Transaction, index int) error {

	if create.GetName() == "" {
		return errEmptyName
	}

	action := newAction(v, tx, index)
	groupInfo, err := action.getGroupInfo(create.GetGroupID())
	if err != nil {
		return err
	}
	if !checkSliceItemExist(action.fromAddr, groupInfo.GetAdmins()) {
		return errAddrPermissionDenied
	}

	if create.GetEndTimestamp() <= v.GetBlockTime() || create.EndTimestamp <= create.BeginTimestamp {
		return errInvalidVoteTime
	}

	if len(create.VoteOptions) < 2 {
		return errInvalidVoteOption
	}

	return nil
}

func (v *vote) checkCommitVote(commit *vty.CommitVote, tx *types.Transaction, index int) error {

	action := newAction(v, tx, index)
	voteInfo, err := action.getVoteInfo(commit.GetVoteID())
	if err != nil {
		return err
	}

	if voteInfo.Status == voteStatusClosed {
		return errVoteAlreadyClosed
	}

	if commit.OptionIndex >= uint32(len(voteInfo.VoteOptions)) {
		return errInvalidOptionIndex
	}

	groupInfo, err := action.getGroupInfo(voteInfo.GroupID)
	if err != nil {
		return err
	}

	// check if exist in group members
	if !checkMemberExist(action.fromAddr, groupInfo.Members) {
		return errAddrPermissionDenied
	}
	// check if already vote
	for _, info := range voteInfo.GetCommitInfos() {
		if action.fromAddr == info.Addr {
			return errAddrAlreadyVoted
		}
	}

	return nil
}

func (v *vote) checkCloseVote(close *vty.CloseVote, tx *types.Transaction, index int) error {

	action := newAction(v, tx, index)
	voteInfo, err := action.getVoteInfo(close.GetVoteID())
	if err != nil {
		return err
	}
	if voteInfo.Creator != action.fromAddr {
		return errAddrPermissionDenied
	}
	if voteInfo.Status == voteStatusClosed {
		return errVoteAlreadyClosed
	}
	return nil
}

func (v *vote) checkUpdateMember(update *vty.UpdateMember) error {
	if update.GetName() == "" {
		return errEmptyName
	}
	return nil
}
