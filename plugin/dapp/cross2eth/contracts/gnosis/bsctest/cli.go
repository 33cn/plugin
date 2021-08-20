package main

//
//import (
//	"fmt"
//	"github.com/spf13/cobra"
//	"os"
//)
//
//func main() {
//	rootCmd := Cmd();
//
//	if err := rootCmd.Execute(); err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//}
//
//// Cmd x2ethereum client command
//func Cmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "pancake test",
//		Short: "pancake test command",
//		Args:  cobra.MinimumNArgs(1),
//	}
//	cmd.AddCommand(
//		GetBalanceCmd(),
//		DeployMulSignCmd(),
//		SetupCmd(),
//		TransferCmd(),
//
//	)
//	return cmd
//}
//
//
////GetBalanceCmd ...
//func GetBalanceCmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "balance",
//		Short: "get owner's balance for ETH or ERC20",
//		Run:   ShowBalance,
//	}
//	GetBalanceFlags(cmd)
//	return cmd
//}
//
////GetBalanceFlags ...
//func GetBalanceFlags(cmd *cobra.Command) {
//	cmd.Flags().StringP("owner", "o", "", "owner address")
//	_ = cmd.MarkFlagRequired("owner")
//	cmd.Flags().StringP("tokenAddr", "t", "", "token address, optional, nil for Eth")
//}
//
////GetBalance ...
//func ShowBalance(cmd *cobra.Command, args []string) {
//	owner, _ := cmd.Flags().GetString("owner")
//	tokenAddr, _ := cmd.Flags().GetString("tokenAddr")
//	balance, err := GetBalance(tokenAddr, owner)
//	if nil != err {
//		fmt.Println("err:",err.Error())
//	}
//	fmt.Println("balance =", balance)
//}
//
//func DeployMulSignCmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "deploy MulSign",
//		Short: "deploy MulSign to bsc ",
//		Run:   DeployContracts,
//	}
//	return cmd
//}
//
//func DeployContracts(cmd *cobra.Command, args []string) {
//	err := DeployMulSign()
//	if nil != err {
//		fmt.Println("Failed to deploy contracts due to:", err.Error())
//		return
//	}
//	fmt.Println("Succeed to deploy contracts")
//}
//
//func SetupCmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "Setup",
//		Short: "Setup owners to contract",
//		Run:   SetupOwner,
//	}
//	SetupOwnerFlags(cmd)
//	return cmd
//}
//
//func SetupOwnerFlags(cmd *cobra.Command) {
//	cmd.Flags().StringP("safe", "a", "", "safe address")
//	_ = cmd.MarkFlagRequired("safe")
//}
//
//func SetupOwner(cmd *cobra.Command, args []string) {
//	safeAddr, _ := cmd.Flags().GetString("safe")
//	err := SetupOwnerProc(safeAddr)
//	if nil != err {
//		fmt.Println("Failed to Setup Owner Proc due to:", err.Error())
//		return
//	}
//	fmt.Println("Succeed to Setup Owner contracts")
//}
//
//func TransferCmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "transfer",
//		Short: "transfer via safe",
//		Run:   Transfer,
//	}
//	TransferFlags(cmd)
//	return cmd
//}
//
//func TransferFlags(cmd *cobra.Command) {
//	cmd.Flags().StringP("safe", "c", "", "safe contract address")
//	_ = cmd.MarkFlagRequired("safe")
//	cmd.Flags().StringP("to", "t", "", "receive address")
//	_ = cmd.MarkFlagRequired("to")
//
//	cmd.Flags().Float64P("amount", "a", 0, "amount to transfer")
//	_ = cmd.MarkFlagRequired("amount")
//
//	cmd.Flags().StringP("token", "k", "", "erc20 address")
//}
//
//func Transfer(cmd *cobra.Command, args []string) {
//	safeAddr, _ := cmd.Flags().GetString("safe")
//	toAddr, _ := cmd.Flags().GetString("to")
//	tokenAddr, _ := cmd.Flags().GetString("token")
//	amount, _ := cmd.Flags().GetFloat64("amount")
//
//	err := TransferProc(safeAddr, toAddr, tokenAddr, amount)
//	if nil != err {
//		fmt.Println("Failed to Transfer due to:", err.Error())
//		return
//	}
//	fmt.Println("Succeed to Transfer via safe")
//}
//
