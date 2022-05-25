// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"github.com/33cn/chain33/common/address"

	gt "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

var (
	roundPrefix      = "mavl-blackwhite-"
	loopResultPrefix = "LODB-blackwhite-loop-"
)

func calcMavlRoundKey(ID string) []byte {
	return []byte(fmt.Sprintf(roundPrefix+"%s", ID))
}

func calcRoundKey4AddrHeight(addr, heightindex string) []byte {
	key := fmt.Sprintf(loopResultPrefix+"%s-"+"%s", address.FormatAddrKey(addr), heightindex)
	return []byte(key)
}

func calcRoundKey4StatusAddrHeight(status int32, addr, heightindex string) []byte {
	key := fmt.Sprintf(loopResultPrefix+"%d-"+"%s-"+"%s", status, address.FormatAddrKey(addr), heightindex)
	return []byte(key)
}

func calcRoundKey4LoopResult(ID string) []byte {
	return []byte(fmt.Sprintf(loopResultPrefix+"%s", ID))
}

func newRound(create *gt.BlackwhiteCreate, creator string) *gt.BlackwhiteRound {
	t := &gt.BlackwhiteRound{}

	t.Status = gt.BlackwhiteStatusCreate
	t.PlayAmount = create.PlayAmount
	t.PlayerCount = create.PlayerCount
	t.Timeout = create.Timeout
	t.Loop = calcloopNumByPlayer(create.PlayerCount)
	t.CreateAddr = creator
	t.GameName = create.GameName
	return t
}
