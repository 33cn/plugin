package init

import (
	_ "github.com/33cn/plugin/plugin/consensus/para"       // register para package
	_ "github.com/33cn/plugin/plugin/consensus/pbft"       // register pbft package
	_ "github.com/33cn/plugin/plugin/consensus/raft"       // register raft package
	_ "github.com/33cn/plugin/plugin/consensus/tendermint" // register tendermint package
	_ "github.com/33cn/plugin/plugin/consensus/ticket"     // register ticket package
)
