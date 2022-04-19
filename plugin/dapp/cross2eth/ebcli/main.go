// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/pluginmgr"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebcli/buildflags"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/version"
	pluginVersion "github.com/33cn/plugin/version"
	tml "github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "chain33xEth-relayer" + "-cli",
		Short: "chain33xEth-relayer" + "client tools",
	}
	configPath = flag.String("f", "", "configfile")
)

func init() {
	rootCmd.AddCommand(
		SetPwdCmd(),
		ChangePwdCmd(),
		LockCmd(),
		UnlockCmd(),
		Chain33RelayerCmd(),
		EthereumRelayerCmd(),
		StaticsCmd(),
		VersionCmd(),
	)
}

func testTLS(RPCAddr string) string {
	rpcaddr := RPCAddr
	if strings.HasPrefix(rpcaddr, "https://") {
		return RPCAddr
	}
	if !strings.HasPrefix(rpcaddr, "http://") {
		return RPCAddr
	}
	//test tls ok
	if rpcaddr[len(rpcaddr)-1] != '/' {
		rpcaddr += "/"
	}
	rpcaddr += "test"
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

//run :
func run(RPCAddr, NodeAddr string) {
	//test tls is enable
	RPCAddr = testTLS(RPCAddr)
	pluginmgr.AddCmd(rootCmd)
	log.SetLogLevel("error")
	rootCmd.PersistentFlags().String("rpc_laddr", RPCAddr, "http url")
	rootCmd.PersistentFlags().String("node_addr", NodeAddr, "bsc node url")
	rootCmd.PersistentFlags().String("eth_chain_name", "Binance", "bsc chain name")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initCfg(path string) *relayerTypes.RelayerConfig {
	var cfg relayerTypes.RelayerConfig
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	return &cfg
}

func main() {
	if *configPath == "" {
		*configPath = "relayer.toml"
	}

	cfg := initCfg(*configPath)

	if buildflags.RPCAddr == "" {
		buildflags.RPCAddr = "http://localhost:9901"
	}
	if buildflags.NodeAddr == "" {
		buildflags.NodeAddr = cfg.EthRelayerCfg[0].EthProviderCli[0]
	}

	run(buildflags.RPCAddr, buildflags.NodeAddr)
}

func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run:   showVersion,
	}
	return cmd
}

func showVersion(_ *cobra.Command, _ []string) {
	fmt.Println("plugin version  :", pluginVersion.GetVersion())
	fmt.Println("relayer version :", version.GetVersion())
	fmt.Println("commit          :", version.GitCommit)
	fmt.Println("buildTime       :", version.BuildTime)
	fmt.Println("goVersion       :", version.GoVersion)
	fmt.Println("platform        :", version.Platform)

}
