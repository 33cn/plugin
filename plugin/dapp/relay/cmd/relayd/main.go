// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/33cn/chain33/common/limits"
	clog "github.com/33cn/chain33/common/log"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/relay/cmd/relayd/relayd"
)

var (
	cpuNum            = runtime.NumCPU()
	configPath        = flag.String("f", "relayd.toml", "configfile")
	chain33ConfigPath = flag.String("chain33flie", "", "chain33configfile")
)

func main() {
	// TODO this is daemon

	runtime.GOMAXPROCS(cpuNum)
	os.Chdir(pwd())
	d, _ := os.Getwd()
	log.Info("current dir:", "dir", d)

	err := limits.SetLimits()
	if err != nil {
		panic(err)
	}

	flag.Parse()
	cfg := relayd.NewConfig(*configPath)
	clog.SetFileLog(&cfg.Log)
	if *chain33ConfigPath != "" {
		cfg.Chain33Cfg = types.NewChain33Config(types.ReadFile(*chain33ConfigPath))
	}

	if cfg.Watch {
		// set watching
		log.Info("watching 10s")
		t := time.Tick(10 * time.Second)
		go func() {
			for range t {
				watching()
			}
		}()
	}

	r := relayd.NewRelayd(cfg)
	go r.Start()

	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, os.Kill)
	s := <-interrupt
	log.Warn("Got signal:", "signal", s)

	r.Close()
}

func watching() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Info("watching:", "NumGoroutine:", runtime.NumGoroutine())
	log.Info("watching:", "Mem:", m.Sys/(1024*1024))
}

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}
