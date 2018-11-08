package init

import (
	_ "./consensus/para"
	_ "./consensus/pbft"
	_ "./consensus/raft"
	_ "./consensus/tendermint"
	_ "./consensus/ticket"
)
