package ethereum

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Cmd x2ethereum client command
func CakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cake",
		Short: "cake command",
	}
	cmd.AddCommand(
		GetBalanceCmd(),
		DeployPancakeCmd(),
		AddAllowance4LPCmd(),
		CheckAllowance4LPCmd(),
		showPairInitCodeHashCmd(),
		setFeeToCmd(),
	)
	return cmd
}

func setFeeToCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setFeeTo",
		Short: "set FeeTo address to factory",
		Run:   setFeeTo,
	}
	AddSetFeeToFlags(cmd)
	return cmd
}

func AddSetFeeToFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("factory", "f", "", "factory Addr ")
	_ = cmd.MarkFlagRequired("factory")

	cmd.Flags().StringP("feeTo", "t", "", "feeTo Addr ")
	_ = cmd.MarkFlagRequired("feeTo")
	cmd.Flags().Uint64P("gas", "g", 80*10000, "gaslimit")
	cmd.Flags().StringP("key", "k", "f934e9171c5cf13b35e6c989e95f5e95fa471515730af147b66d60fbcd664b7c", "private key of feetoSetter")
}

func setFeeTo(cmd *cobra.Command, args []string) {
	ethNodeAddr, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	factroy, _ := cmd.Flags().GetString("factory")
	feeTo, _ := cmd.Flags().GetString("feeTo")
	key, _ := cmd.Flags().GetString("key")
	gasLimit, _ := cmd.Flags().GetUint64("gas")

	setupWebsocketEthClient(ethNodeAddr)
	err := setFeeToHandle(factroy, feeTo, key, gasLimit)
	if nil != err {
		fmt.Println("Failed to deploy contracts due to:", err.Error())
		return
	}
	fmt.Println("Succeed to deploy contracts")
}

func DeployPancakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy pancake router to ethereum ",
		Run:   DeployContractsCake,
	}
	cmd.Flags().StringP("key", "k", "", "private key for deploy")
	return cmd
}

func DeployContractsCake(cmd *cobra.Command, args []string) {
	ethNodeAddr, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	key,_:=cmd.Flags().GetString("key")
	setupWebsocketEthClient(ethNodeAddr)
	err := DeployPancake(key)
	if nil != err {
		fmt.Println("Failed to deploy contracts due to:", err.Error())
		return
	}
	fmt.Println("Succeed to deploy contracts")
}

func AddAllowance4LPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allowance",
		Short: "approve allowance for add lp to pool",
		Run:   AddAllowance4LP,
	}

	AddAllowance4LPFlags(cmd)

	return cmd
}

func AddAllowance4LPFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("masterchef", "m", "", "master Chef Addr ")
	_ = cmd.MarkFlagRequired("masterchef")

	cmd.Flags().StringP("lptoken", "l", "", "lp Addr ")
	_ = cmd.MarkFlagRequired("lptoken")

	cmd.Flags().Int64P("amount", "p", 0, "amount to approve")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("key","k","","private key")

}

func AddAllowance4LP(cmd *cobra.Command, args []string) {
	masterChefAddrStr, _ := cmd.Flags().GetString("masterchef")
	amount, _ := cmd.Flags().GetInt64("amount")
	lpToken, _ := cmd.Flags().GetString("lptoken")
	ethNodeAddr, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	privkey,_:=cmd.Flags().GetString("key")
	setupWebsocketEthClient(ethNodeAddr)

	//owner string, spender string, amount int64
	err := AddAllowance4LPHandle(lpToken, masterChefAddrStr,privkey, amount)
	if nil != err {
		fmt.Println("Failed to AddPool2Farm due to:", err.Error())
		return
	}
	fmt.Println("Succeed to AddPool2Farm")
}

func CheckAllowance4LPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-allowance",
		Short: "check allowance for add lp to pool",
		Run:   CheckAllowance4LP,
	}

	CheckAllowance4LPFlags(cmd)

	return cmd
}

func CheckAllowance4LPFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("masterchef", "m", "", "master Chef Addr ")
	_ = cmd.MarkFlagRequired("masterchef")

	cmd.Flags().StringP("lptoken", "l", "", "lp Addr ")
	_ = cmd.MarkFlagRequired("lptoken")

	cmd.Flags().StringP("key","k","","private key")


}

func CheckAllowance4LP(cmd *cobra.Command, args []string) {
	masterChefAddrStr, _ := cmd.Flags().GetString("masterchef")
	lpToken, _ := cmd.Flags().GetString("lptoken")
	privkey,_:=cmd.Flags().GetString("key")
	ethNodeAddr, _ := cmd.Flags().GetString("rpc_laddr_ethereum")

	setupWebsocketEthClient(ethNodeAddr)

	//owner string, spender string, amount int64
	err := CheckAllowance4LPHandle(lpToken, masterChefAddrStr,privkey)
	if nil != err {
		fmt.Println("Failed to CheckAllowance4LP due to:", err.Error())
		return
	}
	fmt.Println("Succeed to CheckAllowance4LP")
}

func showPairInitCodeHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showInitHash",
		Short: "show pair's init code hash",
		Run:   showPairInitCodeHash,
	}

	showPairInitCodeHashFlags(cmd)

	return cmd
}

func showPairInitCodeHashFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("factory", "f", "", "factory address")
	_ = cmd.MarkFlagRequired("factory")
	cmd.Flags().StringP("key","k","","private key")
}

func showPairInitCodeHash(cmd *cobra.Command, args []string) {
	factory, _ := cmd.Flags().GetString("factory")

	ethNodeAddr, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	privkey,_:=cmd.Flags().GetString("key")
	setupWebsocketEthClient(ethNodeAddr)

	//owner string, spender string, amount int64
	err := showPairInitCodeHashHandle(factory,privkey)
	if nil != err {
		fmt.Println("Failed to showPairInitCodeHash due to:", err.Error())
		return
	}
}
