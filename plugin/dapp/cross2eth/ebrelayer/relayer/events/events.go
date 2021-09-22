package events

import (
	log "github.com/33cn/chain33/common/log/log15"
)

type ClaimType int32
type Event int

var eventsLog = log.New("module", "cross2eth_relayer")

const (
	ClaimTypeUnknown = ClaimType(0)
	ClaimTypeBurn    = ClaimType(1)
	ClaimTypeLock    = ClaimType(2)
)

const (
	// Unsupported : unsupported Chain33 or Ethereum event
	Unsupported Event = iota
	// LogLockFromETH : Ethereum event 'LogLock'
	LogLockFromETH
	// LogBurnFromETH : Ethereum event 'LogChain33TokenBurn'
	LogBurnFromETH
)

// 此处的名字命令不能随意改动，需要与合约event中的命名完全一致
func (d Event) String() string {
	return [...]string{"unknown-LOG", "LogLock", "LogChain33TokenBurn"}[d]
}

func (d ClaimType) String() string {
	return [...]string{"unknown-LOG", "burn", "lock"}[d]
}
