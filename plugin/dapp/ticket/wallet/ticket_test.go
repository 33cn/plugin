// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet_test

import (
	"testing"

	"github.com/33cn/chain33/util/testnode"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	ticketwallet "github.com/33cn/plugin/plugin/dapp/ticket/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin"
)

func Test_WalletTicket(t *testing.T) {
	minerAddr := "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
	t.Log("Begin wallet ticket test")

	cfg, sub := testnode.GetDefaultConfig()
	cfg.Consensus.Name = "ticket"
	mock33 := testnode.NewWithConfig(cfg, sub, nil)
	defer mock33.Close()
	err := mock33.WaitHeight(0)
	assert.Nil(t, err)
	msg, err := mock33.GetAPI().Query(ty.TicketX, "TicketList", &ty.TicketList{Addr: minerAddr, Status: 1})
	assert.Nil(t, err)
	ticketList := msg.(*ty.ReplyTicketList)
	assert.NotNil(t, ticketList)
	//return
	ticketwallet.FlushTicket(mock33.GetAPI())
	err = mock33.WaitHeight(2)
	assert.Nil(t, err)
	header, err := mock33.GetAPI().GetLastHeader()
	require.Equal(t, err, nil)
	require.Equal(t, header.Height >= 2, true)

	in := &ty.TicketClose{MinerAddress: minerAddr}
	msg, err = mock33.GetAPI().ExecWalletFunc(ty.TicketX, "CloseTickets", in)
	assert.Nil(t, err)
	hashes := msg.(*types.ReplyHashes)
	assert.NotNil(t, hashes)
	t.Log("End wallet ticket test")
}
