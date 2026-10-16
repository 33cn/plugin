// Copyright Fuzamei Corp. All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticket

import (
	"testing"

	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apimocks "github.com/33cn/chain33/client/mocks"
	_ "github.com/33cn/chain33/system"
	drivers "github.com/33cn/chain33/system/consensus"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
	"github.com/stretchr/testify/mock"
)

// minerBlock builds a block whose tx[0] decodes as a TicketActionMiner but is
// submitted under an arbitrary execer.
func minerBlock(execer string) *types.Block {
	var ta ty.TicketAction
	ta.Value = &ty.TicketAction_Miner{Miner: &ty.TicketMiner{
		Bits: 0x1f00ffff, Reward: 1, TicketId: "any:0xdead:1:seed", Modify: []byte("modify"),
	}}
	ta.Ty = ty.TicketActionMiner
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(&ta)}
	return &types.Block{Height: 100, Txs: []*types.Transaction{tx}}
}

func newTestClient(t *testing.T, cfg *types.Chain33Config) *Client {
	api := new(apimocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	c := &Client{BaseClient: &drivers.BaseClient{}}
	c.SetAPI(api)
	return c
}

// setForkHeight overrides a dapp fork height on the live config.
func setForkHeight(t *testing.T, cfg *types.Chain33Config, height int64) {
	forks, err := cfg.GetForks()
	require.NoError(t, err)
	forks[ty.TicketX+"."+ty.ForkMinerExecerCheck] = height
}

// After the fork, tx[0] must be executed by the ticket execer. Otherwise the
// exec layer would run some other dapp and ticket's Exec_Miner -- the only
// thing that proves ticket ownership -- would never run.
func TestGetMinerTx_RejectsForeignExecerAfterFork(t *testing.T) {
	cfg := types.NewChain33Config(types.ReadFile("testdata/chain33.cfg.toml"))
	setForkHeight(t, cfg, 0)
	c := newTestClient(t, cfg)

	// honest miner tx still accepted
	_, err := c.getMinerTx(minerBlock(ty.TicketX))
	assert.NoError(t, err, "honest ticket miner tx must still be accepted")

	// any other execer rejected, including ones that can carry a dual-decoding payload
	for _, execer := range []string{"coins", "trade", "none", "ticket2", ""} {
		_, err := c.getMinerTx(minerBlock(execer))
		assert.Equal(t, ty.ErrMinerTxExecer, err, "execer %q must be rejected", execer)
	}
}

// The check must stay a real gate: a chain that leaves this fork disabled (for
// example by setting the height to MaxHeight while it schedules an upgrade, or
// a historical/test chain) still has to accept the old-style miner tx.
func TestGetMinerTx_AcceptsForeignExecerBeforeFork(t *testing.T) {
	cfg := types.NewChain33Config(types.ReadFile("testdata/chain33.cfg.toml"))
	setForkHeight(t, cfg, types.MaxHeight)
	c := newTestClient(t, cfg)

	_, err := c.getMinerTx(minerBlock("coins"))
	assert.NoError(t, err, "pre-fork history must still replay")

	// and the payload-shape rules are still enforced independently of the execer
	blk := minerBlock(ty.TicketX)
	var ta ty.TicketAction
	ta.Value = &ty.TicketAction_Topen{Topen: &ty.TicketOpen{Count: 1}}
	ta.Ty = ty.TicketActionOpen
	blk.Txs[0].Payload = types.Encode(&ta)
	_, err = c.getMinerTx(blk)
	assert.Equal(t, types.ErrCoinBaseTxType, err, "non-miner action must still be rejected")
}
