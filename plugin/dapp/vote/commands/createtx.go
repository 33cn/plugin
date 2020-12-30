package commands

import (
	"fmt"
	"os"

	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
	"github.com/spf13/cobra"
)

func createGroupCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "createGroup",
		Short:   "create tx(create vote group)",
		Run:     createGroup,
		Example: "createGroup -n=group1 -a=admin1 -m=member1 -m=member2",
	}
	createGroupFlags(cmd)
	return cmd
}

func createGroupFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("name", "n", "", "group name")
	cmd.Flags().StringArrayP("admins", "a", nil, "group admin address array")
	cmd.Flags().StringArrayP("members", "m", nil, "group member address array")
	cmd.Flags().UintSliceP("weights", "w", nil, "member vote weight array")
	markRequired(cmd, "name")
}

func createGroup(cmd *cobra.Command, args []string) {

	name, _ := cmd.Flags().GetString("name")
	admins, _ := cmd.Flags().GetStringArray("admins")
	memberAddrs, _ := cmd.Flags().GetStringArray("members")
	weights, _ := cmd.Flags().GetUintSlice("weights")

	if name == "" {
		fmt.Fprintf(os.Stderr, "ErrNilGroupName")
	}
	if len(weights) == 0 {
		weights = make([]uint, len(memberAddrs))
	}
	if len(weights) != len(memberAddrs) {
		fmt.Fprintf(os.Stderr, "member address array length should equal with vote weight array length")
	}

	members := make([]*vty.GroupMember, 0)
	for i, addr := range memberAddrs {
		members = append(members, &vty.GroupMember{Addr: addr, VoteWeight: uint32(weights[i])})
	}

	params := &vty.CreateGroup{
		Name:    name,
		Admins:  admins,
		Members: members,
	}
	sendCreateTxRPC(cmd, vty.NameCreateGroupAction, params)
}

func updateMemberCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "updateMember",
		Short:   "create tx(update group members)",
		Run:     updateMember,
		Example: "updateMember -g=id -a=addMember1 -a=addMember2 -r=removeMember1 ...",
	}
	updateMemberFlags(cmd)
	return cmd
}

func updateMemberFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("groupID", "g", "", "group id")
	cmd.Flags().StringArrayP("addMembers", "a", nil, "group member address array for adding")
	cmd.Flags().UintSliceP("weights", "w", nil, "member vote weight array for adding")
	cmd.Flags().StringArrayP("removeMembers", "r", nil, "group member address array for removing")
	markRequired(cmd, "groupID")
}

func updateMember(cmd *cobra.Command, args []string) {

	groupID, _ := cmd.Flags().GetString("groupID")
	addAddrs, _ := cmd.Flags().GetStringArray("addMembers")
	weights, _ := cmd.Flags().GetUintSlice("weights")
	removeAddrs, _ := cmd.Flags().GetStringArray("removeMembers")

	if groupID == "" {
		fmt.Fprintf(os.Stderr, "ErrNilGroupID")
	}
	if len(weights) == 0 {
		weights = make([]uint, len(addAddrs))
	}
	if len(weights) != len(addAddrs) {
		fmt.Fprintf(os.Stderr, "member address array length should equal with vote weight array length")
	}

	addMembers := make([]*vty.GroupMember, 0)
	for i, addr := range addAddrs {
		addMembers = append(addMembers, &vty.GroupMember{Addr: addr, VoteWeight: uint32(weights[i])})
	}

	params := &vty.UpdateMember{
		GroupID:           groupID,
		RemoveMemberAddrs: removeAddrs,
		AddMembers:        addMembers,
	}
	sendCreateTxRPC(cmd, vty.NameUpdateMemberAction, params)
}

func createVoteCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createVote",
		Short: "create tx(create vote)",
		Run:   createVote,
	}
	createVoteFlags(cmd)
	return cmd
}

func createVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("name", "n", "", "vote name")
	cmd.Flags().StringArrayP("groupIDs", "g", nil, "related group id array")
	cmd.Flags().StringArrayP("options", "o", nil, "vote option array")
	cmd.Flags().Int64P("beginTime", "b", 0, "vote begin unix timestamp, default set now time")
	cmd.Flags().Int64P("endTime", "e", 0, "vote end unix timestamp, default set beginTime + 300 seconds")

	markRequired(cmd, "name", "groupIDs", "options")
}

func createVote(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	groupIDs, _ := cmd.Flags().GetStringArray("groupIDs")
	options, _ := cmd.Flags().GetStringArray("options")
	beginTime, _ := cmd.Flags().GetInt64("beginTime")
	endTime, _ := cmd.Flags().GetInt64("endTime")
	if name == "" {
		fmt.Fprintf(os.Stderr, "ErrNilVoteName")
	}
	if len(groupIDs) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilGroupIDs")
	}

	if len(options) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilOptions")
	}
	if beginTime == 0 {
		beginTime = types.Now().Unix()
	}
	if endTime == 0 {
		endTime = beginTime + 300
	}

	params := &vty.CreateVote{
		Name:           name,
		VoteGroups:     groupIDs,
		VoteOptions:    options,
		BeginTimestamp: beginTime,
		EndTimestamp:   endTime,
	}
	sendCreateTxRPC(cmd, vty.NameCreateVoteAction, params)
}

func commitVoteCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commitVote",
		Short: "create tx(commit vote)",
		Run:   commitVote,
	}
	commitVoteFlags(cmd)
	return cmd
}

func commitVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("voteID", "v", "", "vote id")
	cmd.Flags().StringP("groupID", "g", "", "belonging group id")
	cmd.Flags().Uint32P("optionIndex", "o", 0, "voting option index in option array")
	markRequired(cmd, "voteID", "groupID", "optionIndex")
}

func commitVote(cmd *cobra.Command, args []string) {
	voteID, _ := cmd.Flags().GetString("voteID")
	groupID, _ := cmd.Flags().GetString("groupID")
	optionIndex, _ := cmd.Flags().GetUint32("optionIndex")
	if voteID == "" {
		fmt.Fprintf(os.Stderr, "ErrNilVoteID")
	}
	if len(groupID) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilGroupID")
	}

	params := &vty.CommitVote{
		VoteID:      voteID,
		GroupID:     groupID,
		OptionIndex: optionIndex,
	}
	sendCreateTxRPC(cmd, vty.NameCommitVoteAction, params)
}
