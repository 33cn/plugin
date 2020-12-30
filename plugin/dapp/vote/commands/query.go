package commands

import (
	"fmt"
	"os"

	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
	"github.com/spf13/cobra"
)

func groupInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groupInfo",
		Short: "show group info",
		Run:   groupInfo,
	}
	groupInfoFlags(cmd)
	return cmd
}

func groupInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("groupID", "g", "", "group id")
	markRequired(cmd, "groupID")
}

func groupInfo(cmd *cobra.Command, args []string) {
	groupID, _ := cmd.Flags().GetString("groupID")
	if len(groupID) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilGroupID")
	}

	params := &types.ReqString{
		Data: groupID,
	}
	info := &vty.GroupVoteInfo{}
	sendQueryRPC(cmd, "GetGroup", params, info)
}

func voteInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voteInfo",
		Short: "show vote info",
		Run:   voteInfo,
	}
	voteInfoFlags(cmd)
	return cmd
}

func voteInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("voteID", "v", "", "vote id")
	markRequired(cmd, "voteID")
}

func voteInfo(cmd *cobra.Command, args []string) {
	voteID, _ := cmd.Flags().GetString("voteID")
	if len(voteID) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilVoteID")
	}

	params := &types.ReqString{
		Data: voteID,
	}
	info := &vty.VoteInfo{}
	sendQueryRPC(cmd, "GetVote", params, info)
}

func memberInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memberInfo",
		Short: "show member info",
		Run:   memberInfo,
	}
	memberInfoFlags(cmd)
	return cmd
}

func memberInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "member address")
	markRequired(cmd, "addr")
}

func memberInfo(cmd *cobra.Command, args []string) {
	addr, _ := cmd.Flags().GetString("addr")
	if len(addr) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilAddress")
	}

	params := &types.ReqString{
		Data: addr,
	}
	info := &vty.MemberInfo{}
	sendQueryRPC(cmd, "GetMember", params, info)
}

func listGroupCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listGroup",
		Short: "show group list",
		Run:   listGroup,
	}
	listCmdFlags(cmd)
	return cmd
}

func listGroup(cmd *cobra.Command, args []string) {
	runListCMD(cmd, args, "ListGroup", &vty.GroupVoteInfos{})
}

func listVoteCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listVote",
		Short: "show vote list",
		Run:   listVote,
	}
	listCmdFlags(cmd)
	return cmd
}

func listVote(cmd *cobra.Command, args []string) {
	runListCMD(cmd, args, "ListVote", &vty.VoteInfos{})
}

func listMemberCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listMember",
		Short: "show member list",
		Run:   listMember,
	}
	listCmdFlags(cmd)
	return cmd
}

func listMember(cmd *cobra.Command, args []string) {
	runListCMD(cmd, args, "ListMember", &vty.MemberInfos{})
}

func listCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("startItem", "s", "", "list start item id, default nil value")
	cmd.Flags().Uint32P("count", "c", 10, "list count, default 10")
	cmd.Flags().Uint32P("direction", "d", 0, "list direction, default 1 (Ascending order)")
}

func runListCMD(cmd *cobra.Command, args []string, funcName string, reply types.Message) {
	startID, _ := cmd.Flags().GetString("startItem")
	count, _ := cmd.Flags().GetUint32("count")
	direction, _ := cmd.Flags().GetUint32("direction")
	params := &vty.ReqListItem{
		StartItemID: startID,
		Count:       int32(count),
		Direction:   int32(direction),
	}
	sendQueryRPC(cmd, funcName, params, reply)
}
