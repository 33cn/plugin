package init

import (
	_ "gitlab.33.cn/chain33/plugin/plugin/consensus/para"
	_ "gitlab.33.cn/chain33/plugin/plugin/consensus/pbft"
	_ "gitlab.33.cn/chain33/plugin/plugin/consensus/raft"
	_ "gitlab.33.cn/chain33/plugin/plugin/consensus/tendermint"
	_ "gitlab.33.cn/chain33/plugin/plugin/consensus/ticket"
)
