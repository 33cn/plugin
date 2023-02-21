package commands

import (
	"fmt"
	"os"

	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/spf13/cobra"
)

func validatorCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator",
		Short: "show validators bls pubkey",
		Run:   getValidator,
	}
	addTitleFlags(cmd)
	return cmd
}

func addTitleFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("paratitle", "t", "", "para chain title")
	markRequired(cmd, "paratitle")
}

func getValidator(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("paratitle")
	if title == "" {
		fmt.Fprintf(os.Stderr, "Err empty parachain title")
		return
	}

	params := &rtypes.ChainTitle{
		Value: title,
	}
	info := &rtypes.ValidatorPubs{}
	sendQueryRPC(cmd, "GetValidatorPubs", params, info)
}

func rollupStatusCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show rollup status",
		Run:   getRollupStatus,
	}
	addTitleFlags(cmd)
	return cmd
}

func getRollupStatus(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("paratitle")
	if title == "" {
		fmt.Fprintf(os.Stderr, "Err empty parachain title")
		return
	}

	params := &rtypes.ChainTitle{
		Value: title,
	}
	info := &rtypes.RollupStatus{}
	sendQueryRPC(cmd, "GetRollupStatus", params, info)
}

func roundInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "round",
		Short: "show rollup commit round info",
		Run:   getRoundInfo,
	}
	addTitleFlags(cmd)
	cmd.Flags().Int64P("round", "r", 0, "commit round")
	markRequired(cmd, "round")
	return cmd
}

func getRoundInfo(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("paratitle")
	if title == "" {
		fmt.Fprintf(os.Stderr, "Err empty parachain title")
		return
	}

	round, _ := cmd.Flags().GetInt64("round")

	params := &rtypes.ReqGetCommitRound{
		ChainTitle:  title,
		CommitRound: round,
	}
	info := &rtypes.CommitRoundInfo{}
	sendQueryRPC(cmd, "GetCommitRoundInfo", params, info)
}
