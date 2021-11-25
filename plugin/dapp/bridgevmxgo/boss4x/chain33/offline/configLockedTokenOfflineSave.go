package offline

import (
	"fmt"
	"math/big"

	"github.com/33cn/plugin/plugin/dapp/bridgevmxgo/contracts/generated"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	utilsRelayer "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/spf13/cobra"
)

/*
./boss4x chain33 offline set_offline_token -c 1MaP3rrwiLV1wrxPhDwAfHggtei1ByaKrP -s BTY -m 100000000000 -p 50 -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae --chainID 33
./boss4x chain33 offline send -f chain33_set_offline_token.txt
*/

func ConfigLockedTokenOfflineSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_offline_token",
		Short: "set config offline locked token",
		Run:   ConfigMultisignLockedTokenOfflineSave,
	}
	addConfigLockedTokenOfflineSaveFlags(cmd)
	return cmd
}

func addConfigLockedTokenOfflineSaveFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
	cmd.Flags().StringP("token", "t", "", "token addr")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("threshold", "m", "0", "threshold")
	_ = cmd.MarkFlagRequired("threshold")
	cmd.Flags().Uint8P("percents", "p", 50, "percents")
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func ConfigMultisignLockedTokenOfflineSave(cmd *cobra.Command, _ []string) {
	contract, _ := cmd.Flags().GetString("contract")
	token, _ := cmd.Flags().GetString("token")
	symbol, _ := cmd.Flags().GetString("symbol")
	threshold, _ := cmd.Flags().GetString("threshold")
	percents, _ := cmd.Flags().GetUint8("percents")
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utilsRelayer.TrimZeroAndDot(threshold), 10)

	if token == "" || symbol == "BTY" {
		token = ebTypes.BTYAddrChain33
	}

	parameter := fmt.Sprintf("configLockedTokenOfflineSave(%s,%s,%d,%d)", token, symbol, bn, percents)
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeBankABI, false)
	if nil != err {
		fmt.Println("configOfflineSaveAccount", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	callContractAndSignWrite(cmd, packData, contract, "chain33_set_offline_token")
}
