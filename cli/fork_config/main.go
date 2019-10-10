package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin"
)

var (
	configPath = flag.String("f", "", "configfile")
)

func main() {
	flag.Parse()
	if *configPath == "" {
		*configPath = "chain33.toml"
	}
	cfg, _ := types.InitCfg(*configPath)
	types.Init(cfg.Title, cfg)
	forks, err := types.CloneFork("chain33")
	if err != nil {
		fmt.Printf("clone fork failed: %v", err)
		return
	}

	fmtForks(forks)
}

/*
	两个规则：
	key 有 ".", Part1.Part2 为 [fork.sub.Part1] Part2=value
	key 没有 "." [fork.system] key=value
	把相同段的fork打印到一起
		[fork.system]
		ForkChainParamV1= 0  # ForkBlockCheck=1560000

		[fork.sub.ticket]
		Enable=0  # manage.ForkManageExec=400000

		[fork.sub.store-kvmvccmavl]
		ForkKvmvccmavl=2270000 # store-kvmvccmavl.ForkKvmvccmavl=1870000
*/
func fmtForks(forks map[string]int64) {
	systemFork := make(map[string]int64)
	subFork := make(map[string]map[string]int64)
	for k, v := range forks {
		if strings.Contains(k, ".") {
			str2 := strings.SplitN(k, ".", 2)
			if len(str2) != 2 {
				fmt.Fprintf(os.Stderr, "can't deal key=%s ", k)
				continue
			}
			_, ok := subFork[str2[0]]
			if !ok {
				subFork[str2[0]] = make(map[string]int64)
			}
			subFork[str2[0]][str2[1]] = v
		} else {
			systemFork[k] = v
		}

	}

	fmt.Println("[fork.system]")
	for k, v := range systemFork {
		fmt.Printf("%s=%d\n", k, v)
	}
	fmt.Println("")

	plugins := make([]string, 0)
	for plugin := range subFork {
		plugins = append(plugins, plugin)
	}
	sort.Strings(plugins)

	for _, plugin := range plugins {
		fmt.Printf("[fork.sub.%s]\n", plugin)
		forks := subFork[plugin]
		for k, v := range forks {
			fmt.Printf("%s=%d\n", k, v)
		}
		fmt.Println("")
	}

}
