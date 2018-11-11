// +build go1.8

package main

import (
	_ "gitlab.33.cn/chain33/chain33/system"
	"gitlab.33.cn/chain33/chain33/util/cli"
	"gitlab.33.cn/chain33/plugin/cli/buildflags"
	_ "gitlab.33.cn/chain33/plugin/plugin"
)

func main() {
	if buildflags.RPCAddr == "" {
		buildflags.RPCAddr = "http://localhost:8801"
	}
	cli.Run(buildflags.RPCAddr, buildflags.ParaName)

}
