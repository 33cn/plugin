package offline

import (
	"fmt"
	"strings"

	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/spf13/cobra"
)

/*
./boss4x chain33 offline multisign_setup -m 1GrhufvPtnBCtfxDrFGcCoihmYMHJafuPn -o 168Sn1DXnLrZHTcAM9stD6t2P49fNuJfJ9,13KTf57aCkVVJYNJBXBBveiA5V811SrLcT,1JQwQWsShTHC4zxHzbUfYQK4kRBriUQdEe,1NHuKqoKe3hyv52PF8XBAyaTmJWAqA2Jbb -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae --chainID 33
./boss4x chain33 offline send -f multisign_setup.txt
*/

func SetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign_setup",
		Short: "Setup owners to contract",
		Run:   SetupOwner,
	}
	SetupOwnerFlags(cmd)
	return cmd
}

func SetupOwnerFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner", "o", "", "owners's address, separated by ','")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("key", "k", "", "operator private key")
	_ = cmd.MarkFlagRequired("operator")
	cmd.Flags().StringP("multisign", "m", "", "multisign contract address")
	_ = cmd.MarkFlagRequired("multisign")
}

func SetupOwner(cmd *cobra.Command, _ []string) {
	multisign, _ := cmd.Flags().GetString("multisign")
	owner, _ := cmd.Flags().GetString("owner")
	owners := strings.Split(owner, ",")

	BTYAddrChain33 := ebTypes.BTYAddrChain33
	parameter := "setup(["
	parameter += fmt.Sprintf("%s", owners[0])
	for _, owner := range owners[1:] {
		parameter += fmt.Sprintf(",%s", owner)
	}
	parameter += "], "
	parameter += fmt.Sprintf("%d, %s, 0102, %s, %s, 0, %s)", len(owners), BTYAddrChain33, BTYAddrChain33, BTYAddrChain33, BTYAddrChain33)
	_, packData, err := evmAbi.Pack(parameter, gnosis.GnosisSafeABI, false)
	if nil != err {
		fmt.Println("multisign_setup", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	callContractAndSignWrite(cmd, packData, multisign, "multisign_setup")
}
