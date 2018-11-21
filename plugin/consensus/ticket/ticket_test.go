// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticket

import (
	"testing"

	"github.com/33cn/plugin/plugin/dapp/ticket/types"

	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
	"github.com/stretchr/testify/assert"
)

// 执行： go test -cover
func TestTicket(t *testing.T) {
	cfg, sub := testnode.GetDefaultConfig()
	cfg.Consensus.Name = "ticket"
	mock33 := testnode.NewWithConfig(cfg, sub, nil)
	defer mock33.Close()
	err := mock33.WaitHeight(100)
	assert.Nil(t, err)
}

func TestTicketMap(t *testing.T) {
	c := Client{}
	ticketList := &types.ReplyTicketList{}
	ticketList.Tickets = []*types.Ticket{
		{TicketId: "1111"},
		{TicketId: "2222"},
		{TicketId: "3333"},
		{TicketId: "4444"},
	}
	assert.Equal(t, c.getTicketCount(), int64(0))
	c.setTicket(ticketList, nil)
	assert.Equal(t, c.getTicketCount(), int64(4))
	c.delTicket("3333")
	assert.Equal(t, c.getTicketCount(), int64(3))

}
