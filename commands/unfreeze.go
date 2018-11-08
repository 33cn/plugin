package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "unfreeze",
		Short:                      "Unfreeze construct management",
		Args:                       cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(createCmd())
	cmd.AddCommand(withdrawCmd())
	cmd.AddCommand(terminateCmd())
	cmd.AddCommand(showCmd())
	return cmd
}

func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create",
		Short: "create unfreeze construct",
	}

	cmd.AddCommand(fixAmountCmd())
	cmd.AddCommand(leftCmd())
	return cmd
}

func createFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringP("beneficiary", "b", "", "address of beneficiary")
	cmd.MarkFlagRequired("beneficiary")

	cmd.PersistentFlags().StringP("asset_exec", "e", "", "asset exec")
	cmd.MarkFlagRequired("asset_exec")

	cmd.PersistentFlags().StringP("asset_symbol", "s", "", "asset symbol")
	cmd.MarkFlagRequired("asset_symbol")

	cmd.PersistentFlags().Float64P("total", "t", 0, "total count of asset")
	cmd.MarkFlagRequired("total")

	cmd.PersistentFlags().Int64P("start_ts", "", 0, "effect, UTC timestamp")
	//cmd.MarkFlagRequired("start_ts")

	return cmd
}

func fixAmountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fix_amount",
		Short: "create fix amount means unfreeze construct",
		Run: fixAmount,
	}
	cmd = createFlag(cmd)
	cmd.Flags().Int64P("amount", "a", 0, "amount every period")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().Int64P("period", "p", 0, "period in second")
	cmd.MarkFlagRequired("period")
	return cmd
}

func fixAmount(cmd *cobra.Command, args []string) {
	beneficiary, _ := cmd.Flags().GetString("beneficiary")
	fmt.Printf("%s\n", beneficiary)
}

func leftCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "left_proportion",
		Short: "create left proportion means unfreeze construct",
		Run: left,
	}
	cmd = createFlag(cmd)
	cmd.Flags().Int64P("ten_thousandth", "", 0, "input/10000 of total")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().Int64P("period", "p", 0, "period in second")
	cmd.MarkFlagRequired("period")
	return cmd
}

func left(cmd *cobra.Command, args []string) {
	beneficiary, _ := cmd.Flags().GetString("beneficiary")
	fmt.Printf("%s\n", beneficiary)
}

func withdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "withdraw",
		Short: "withdraw asset from construct",
		Run: withdraw,
	}
	cmd.Flags().StringP("id", "", "", "unfreeze construct id")
	cmd.MarkFlagRequired("id")

	return cmd
}


func terminateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "terminate",
		Short: "terminate construct",
		Run: terminate,
	}
	cmd.Flags().StringP("id", "", "", "unfreeze construct id")
	cmd.MarkFlagRequired("id")

	return cmd
}

func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "show",
		Short: "show construct",
		Run: show,
	}
	cmd.Flags().StringP("id", "", "", "unfreeze construct id")
	cmd.MarkFlagRequired("id")

	return cmd
}

func queryWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "show",
		Short: "show available withdraw amount of one unfreeze construct",
		Run: queryWithdraw,
	}
	cmd.Flags().StringP("id", "", "", "unfreeze construct id")
	cmd.MarkFlagRequired("id")

	return cmd
}

func queryWithdraw(cmd *cobra.Command, args []string) {

}

func show(cmd *cobra.Command, args []string) {

}

func withdraw(cmd *cobra.Command, args []string) {

}

func terminate(cmd *cobra.Command, args []string) {

}
