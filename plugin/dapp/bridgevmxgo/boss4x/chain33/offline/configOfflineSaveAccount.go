package offline

import (
	"fmt"

	"github.com/33cn/plugin/plugin/dapp/bridgevmxgo/contracts/generated"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/spf13/cobra"
)

/*
./boss4x chain33 offline set_offline_addr -a 16skyHQA4YPPnhrDSSpZnexDzasS8BNx1R -c 1QD5pHMKZ9QWiNb9AsH3G1aG3Hashye83o -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae --chainID 33
./boss4x chain33 offline send -f chain33_set_offline_addr.txt
*/

func ConfigOfflineSaveAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_offline_addr",
		Short: "save config offline account",
		Run:   ConfigMultisignOfflineSaveAccount, //配置账户
	}
	addConfigOfflineSaveAccountFlags(cmd)
	return cmd
}

func addConfigOfflineSaveAccountFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "multisign address")
	_ = cmd.MarkFlagRequired("address")
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func ConfigMultisignOfflineSaveAccount(cmd *cobra.Command, _ []string) {
	multisign, _ := cmd.Flags().GetString("address")
	contract, _ := cmd.Flags().GetString("contract")

	parameter := fmt.Sprintf("configOfflineSaveAccount(%s)", multisign)
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeBankABI, false)
	if nil != err {
		fmt.Println("configOfflineSaveAccount", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	callContractAndSignWrite(cmd, packData, contract, "chain33_set_offline_addr")
}
