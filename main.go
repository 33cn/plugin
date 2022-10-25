// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.9
// +build go1.9

/*
每个系统的功能通过插件完成，插件分成4类：
共识 加密 dapp 存储
这个go 包提供了 官方提供的 插件。
*/
package main

import (
	"flag"
	"fmt"
	"os"

	frameVersion "github.com/33cn/chain33/common/version"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/util/cli"
	_ "github.com/33cn/plugin/plugin"
	"github.com/33cn/plugin/version"
)

var (
	versionCmd = flag.Bool("version", false, "detail version")
)

func main() {
	flag.Parse()
	if *versionCmd {
		fmt.Fprintln(os.Stdout, "Build time:", version.BuildTime)
		fmt.Fprintln(os.Stdout, "System version:", version.Platform)
		fmt.Fprintln(os.Stdout, "Golang version:", version.GoVersion)
		fmt.Fprintln(os.Stdout, "plugin version:", version.GetVersion())
		fmt.Fprintln(os.Stdout, "chain33 frame version:", frameVersion.GetVersion())
		return
	}
	cli.RunChain33("", "")
}
