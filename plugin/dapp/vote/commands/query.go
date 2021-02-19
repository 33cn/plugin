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
		Use:     "groupInfo",
		Aliases: []string{"gf"},
		Short:   "get group infos",
		Run:     groupInfo,
		Example: "groupInfo -g=id1 -g=id2...",
	}
	groupInfoFlags(cmd)
	return cmd
}

func groupInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("groupIDs", "g", nil, "group id array")
	markRequired(cmd, "groupIDs")
}

func groupInfo(cmd *cobra.Command, args []string) {
	groupIDs, _ := cmd.Flags().GetStringArray("groupIDs")
	if len(groupIDs) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilGroupIDs")
		return
	}

	params := &vty.ReqStrings{
		Items: groupIDs,
	}
	info := &vty.GroupInfos{}
	sendQueryRPC(cmd, "GetGroups", params, info)
}

func voteInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "voteInfo",
		Aliases: []string{"vf"},
		Short:   "get vote info",
		Run:     voteInfo,
	}
	voteInfoFlags(cmd)
	return cmd
}

func voteInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("voteIDs", "v", nil, "vote id array")
	markRequired(cmd, "voteID")
}

func voteInfo(cmd *cobra.Command, args []string) {
	voteIDs, _ := cmd.Flags().GetStringArray("voteIDs")
	if len(voteIDs) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilVoteID")
		return
	}

	params := &vty.ReqStrings{
		Items: voteIDs,
	}
	info := &vty.ReplyVoteList{}
	sendQueryRPC(cmd, "GetVotes", params, info)
}

func memberInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memberInfo",
		Aliases: []string{"mf"},
		Short:   "get member info",
		Run:     memberInfo,
	}
	memberInfoFlags(cmd)
	return cmd
}

func memberInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("addrs", "a", nil, "member address array")
	markRequired(cmd, "addr")
}

func memberInfo(cmd *cobra.Command, args []string) {
	addrs, _ := cmd.Flags().GetStringArray("addrs")
	if len(addrs) == 0 {
		fmt.Fprintf(os.Stderr, "ErrNilAddress")
		return
	}

	params := &vty.ReqStrings{
		Items: addrs,
	}
	info := &vty.MemberInfos{}
	sendQueryRPC(cmd, "GetMembers", params, info)
}

func listGroupCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "listGroup",
		Aliases: []string{"lg"},
		Short:   "show group list",
		Run:     listGroup,
	}
	listCmdFlags(cmd)
	return cmd
}

func listGroup(cmd *cobra.Command, args []string) {
	runListCMD(cmd, "ListGroup", &vty.GroupInfos{})
}

func listVoteCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "listVote",
		Aliases: []string{"lv"},
		Short:   "show vote list",
		Run:     listVote,
	}
	listVoteFlags(cmd)
	return cmd
}
func listVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("groupID", "g", "", "list vote belongs to specified group, list all if not set")
	cmd.Flags().Uint32P("status", "t", 0, "vote status")
	listCmdFlags(cmd)
}

func listVote(cmd *cobra.Command, args []string) {
	groupID, _ := cmd.Flags().GetString("groupID")
	status, _ := cmd.Flags().GetUint32("status")
	listReq := getListReq(cmd)
	req := &vty.ReqListVote{
		GroupID: groupID,
		ListReq: listReq,
		Status:  status,
	}
	sendQueryRPC(cmd, "ListVote", req, &vty.ReplyVoteList{})
}

func listMemberCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "listMember",
		Aliases: []string{"lm"},
		Short:   "show member list",
		Run:     listMember,
	}
	listCmdFlags(cmd)
	return cmd
}

func listMember(cmd *cobra.Command, args []string) {
	runListCMD(cmd, "ListMember", &vty.MemberInfos{})
}

func listCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("startItem", "s", "", "list start item id, default nil value")
	cmd.Flags().Uint32P("count", "c", 5, "list count, default 5")
	cmd.Flags().Uint32P("direction", "d", 1, "list direction, default 1 (Ascending order)")
}

func runListCMD(cmd *cobra.Command, funcName string, reply types.Message) {
	req := getListReq(cmd)
	sendQueryRPC(cmd, funcName, req, reply)
}

func getListReq(cmd *cobra.Command) *vty.ReqListItem {
	startID, _ := cmd.Flags().GetString("startItem")
	count, _ := cmd.Flags().GetUint32("count")
	direction, _ := cmd.Flags().GetUint32("direction")
	req := &vty.ReqListItem{
		StartItemID: startID,
		Count:       int32(count),
		Direction:   int32(direction),
	}
	return req
}
