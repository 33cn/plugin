package init

import (
	_ "github.com/33cn/plugin/plugin/consensus/para"
	_ "github.com/33cn/plugin/plugin/consensus/pbft"
	_ "github.com/33cn/plugin/plugin/consensus/raft"
	_ "github.com/33cn/plugin/plugin/consensus/tendermint"
	_ "github.com/33cn/plugin/plugin/consensus/ticket"
)
