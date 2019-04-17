// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package main 挖矿监控
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	l "github.com/33cn/chain33/common/log/log15"
	tml "github.com/BurntSushi/toml"

	"github.com/33cn/plugin/cmd/local-status/chain33rpc"
	"github.com/33cn/plugin/cmd/local-status/db"
	"github.com/33cn/plugin/cmd/local-status/es_cli"
	"github.com/33cn/plugin/cmd/local-status/proto"
)

var (
	log        = l.New("module", "main")
	configPath = flag.String("f", "local-status.toml", "config file")
)

func main() {
	d, _ := os.Getwd()
	log.Debug("current dir:", "dir", d)
	os.Chdir(pwd())
	d, _ = os.Getwd()
	log.Debug("current dir:", "dir", d)
	flag.Parse()
	cfg := InitCfg(*configPath)
	log.Debug("load config", "cfgPath", *configPath, "h1", cfg.Chain33Host, "h2", cfg.EsHost)

	// recover from ES
	es, err := es_cli.NewESClient(cfg.EsHost)
	if err != nil {
		panic(err)
	}

	st, err := recover(es)
	if err != nil {
		panic(err)
	}
	fmt.Println("started")

	// sync and save block to ES
	heightCh := make(chan int64)
	blockCh := make(chan interface{})
	go chain33rpc.SyncBlock(cfg.Chain33Host, heightCh, blockCh)
	for {
		st.Sync(es, heightCh, blockCh)
	}

}

func recover(es *es_cli.ESClient) (db.Sync, error) {
	st := db.NewDBStatus()
	err := st.Recover(es)
	fmt.Println(st)
	if err != nil {
		fmt.Println(err)
	}
	return st, err
}

//InitCfg 初始化cfg
func InitCfg(path string) *proto.Config {
	var cfg proto.Config
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	fmt.Println(cfg)
	return &cfg
}

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}
