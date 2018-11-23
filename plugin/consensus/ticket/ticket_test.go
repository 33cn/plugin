// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticket

import (
	"testing"

	"github.com/33cn/chain33/types"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/queue"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
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
	ticketList := &ty.ReplyTicketList{}
	ticketList.Tickets = []*ty.Ticket{
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

func TestProcEvent(t *testing.T) {
	c := Client{}
	ret := c.ProcEvent(queue.Message{})
	assert.Equal(t, ret, true)

}

func Test_genPrivHash(t *testing.T) {
	c, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.NoError(t, err)
	priv, err := c.GenKey()

	bt, err := genPrivHash(priv, "AA:BB:CC:DD")
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(bt))

	bt, err = genPrivHash(priv, "111:222:333:444")
	assert.NoError(t, err)
	assert.Equal(t, 32, len(bt))
}

func Test_getNextRequiredDifficulty(t *testing.T) {
	c := &Client{}

	bits, bt, err := c.getNextRequiredDifficulty(nil, 1)
	assert.NoError(t, err)
	assert.Equal(t, bt, defaultModify)
	assert.Equal(t, bits, types.GetP(0).PowLimitBits)
}
