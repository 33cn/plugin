package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/33cn/plugin/plugin/dapp/bridgevmxgo/boss4x/buildFlags"
	"github.com/33cn/plugin/plugin/dapp/bridgevmxgo/boss4x/chain33"
	"github.com/spf13/cobra"
)

func main() {
	if buildFlags.RPCAddr4Chain33 == "" {
		buildFlags.RPCAddr4Chain33 = "http://localhost:8801"
	}
	buildFlags.RPCAddr4Chain33 = testTLS(buildFlags.RPCAddr4Chain33)

	rootCmd := RootCmd()
	rootCmd.PersistentFlags().String("rpc_laddr", buildFlags.RPCAddr4Chain33, "http url")
	rootCmd.PersistentFlags().String("paraName", "", "para chain name,Eg:user.p.fzm.")
	rootCmd.PersistentFlags().String("expire", "120m", "transaction expire time (optional)")
	rootCmd.PersistentFlags().Int32("chainID", 0, "chain id, default to 0")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// RootCmd Cmd x2ethereum client command
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "boss for bridgevmxgo",
		Short: "manage create offline tx or deploy contracts(bridgevmxgo) for test",
	}
	cmd.AddCommand(
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
