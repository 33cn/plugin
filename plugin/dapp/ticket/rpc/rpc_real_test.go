// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

//only load all plugin and system
import (
	"testing"

	rpctypes "github.com/33cn/chain33/rpc/types"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/stretchr/testify/assert"
)

func TestNewTicket(t *testing.T) {
	//选票(可以用hotwallet 关闭选票)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.GetModuleConfig().Consensus.Name = "ticket"
	mocker := testnode.NewWithConfig(cfg, nil)
	mocker.Listen()
	defer mocker.Close()

	in := &ty.TicketClose{MinerAddress: mocker.GetHotAddress()}
	var res rpctypes.ReplyHashes
	err := mocker.GetJSONC().Call("ticket.CloseTickets", in, &res)
	assert.Nil(t, err)
}
