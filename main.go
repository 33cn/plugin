// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
		fmt.Println(fmt.Sprintf("Build time: %s", version.BuildTime))
		fmt.Println(fmt.Sprintf("System version: %s", version.Platform))
		fmt.Println(fmt.Sprintf("Golang version: %s", version.GoVersion))
		fmt.Println(fmt.Sprintf("plugin version: %s", version.GetVersion()))
		fmt.Println(fmt.Sprintf("chain33 frame version: %s", frameVersion.GetVersion()))
		return
	}
	cli.RunChain33("", "")
}
