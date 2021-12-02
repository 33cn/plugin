package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/33cn/plugin/plugin/dapp/dex/boss/buildFlags"
	"github.com/33cn/plugin/plugin/dapp/dex/boss/deploy/chain33"
	"github.com/33cn/plugin/plugin/dapp/dex/boss/deploy/chain33/offline"
	"github.com/33cn/plugin/plugin/dapp/dex/boss/deploy/ethereum"
	ethoffline "github.com/33cn/plugin/plugin/dapp/dex/boss/deploy/ethereum/offline"
	"github.com/spf13/cobra"
)

func main() {
	if buildFlags.RPCAddr4Chain33 == "" {
		buildFlags.RPCAddr4Chain33 = "http://localhost:8801"
	}
	buildFlags.RPCAddr4Chain33 = testTLS(buildFlags.RPCAddr4Chain33)

	if buildFlags.RPCAddr4Ethereum == "" {
		buildFlags.RPCAddr4Ethereum = "https://data-seed-prebsc-1-s1.binance.org:8545"
		//buildFlags.RPCAddr4Ethereum = "wss://ws-testnet.huobichain.com"
	}

	rootCmd := RootCmd()
	rootCmd.PersistentFlags().String("rpc_laddr", buildFlags.RPCAddr4Chain33, "http url")
	rootCmd.PersistentFlags().String("rpc_laddr_ethereum", buildFlags.RPCAddr4Ethereum, "http url")
	rootCmd.PersistentFlags().String("paraName", "", "para chain name,Eg:user.p.fzm.")
	rootCmd.PersistentFlags().String("expire", "120m", "transaction expire time (optional)")
	rootCmd.PersistentFlags().Int32("chainID", 0, "chain id, default to 0")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Cmd x2ethereum client command
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "boss",
		Short: "manage create offline tx or deploy contracts(dex) for test",
	}
	cmd.AddCommand(
		DeployCmd(),
		OfflineCmd(),
	)
	return cmd
}

func OfflineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "offline",
		Short: "create and sign offline tx to deploy and set dex contracts to ethereum or chain33",
	}
	cmd.AddCommand(
		offline.Chain33OfflineCmd(),
		ethoffline.EthOfflineCmd(),
	)
	return cmd
}

// Cmd x2ethereum client command
func DeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy dex to ethereum or chain33",
	}
	cmd.AddCommand(
		ethereum.EthCmd(),
		chain33.Chain33Cmd(),
	)
	return cmd
}

func testTLS(RPCAddr string) string {
	rpcaddr := RPCAddr
	if !strings.HasPrefix(rpcaddr, "http://") {
		return RPCAddr
	}
	// if http://
	if rpcaddr[len(rpcaddr)-1] != '/' {
		rpcaddr += "/"
	}
	rpcaddr += "test"
	/* #nosec */
	resp, err := http.Get(rpcaddr)
	if err != nil {
		return "https://" + RPCAddr[7:]
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return RPCAddr
	}
	return "https://" + RPCAddr[7:]
}
