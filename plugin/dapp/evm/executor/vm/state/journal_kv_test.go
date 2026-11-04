package state

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/db"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/require"
)

// recordKV 记录写入底层 StateDB 的 key，模拟 chain33 StateDB.GetSetKeys
type recordKV struct {
	db.KV
	keys []string
}

func (r *recordKV) Set(key, value []byte) error {
	r.keys = append(r.keys, string(key))
	return r.KV.Set(key, value)
}

func newJournalTestDB(t *testing.T) (*MemoryStateDB, *recordKV, *account.DB) {
	cfgStr := ctypes.GetDefaultCfgstring()
	if !strings.Contains(cfgStr, "[exec.sub.evm]") {
		cfgStr += "\n[exec.sub.evm]\nethMapFromExecutor=\"coins\"\nethMapFromSymbol=\"bty\"\n"
	}
	cfg := ctypes.NewChain33Config(cfgStr)
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)

	dbDir, stateDB, localDB := util.CreateTestDB()
	t.Cleanup(func() { util.CloseTestDB(dbDir, stateDB) })

	raw, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), stateDB)
	require.NoError(t, err)
	rec := &recordKV{KV: stateDB}
	coins, err := account.NewAccountDB(cfg, "coins", cfg.GetCoinSymbol(), rec)
	require.NoError(t, err)
	return NewMemoryStateDB(rec, localDB, coins, 1, api), rec, raw
}

// 被回滚的转账：缓冲恢复原值，不写底层 StateDB，也不进回执；
// 写入底层的 key 必须都在回执里（chain33 checkKV 的约束）。
func TestJournalKVRevertedTransfer(t *testing.T) {
	mdb, rec, raw := newJournalTestDB(t)
	a, b, c := "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR", "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k", "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"
	acc := raw.LoadAccount(a)
	acc.Balance = 1000
	raw.SaveAccount(acc)
	rec.keys = nil

	mdb.Snapshot()
	v := mdb.Snapshot()
	require.True(t, mdb.Transfer(a, b, 600))
	require.EqualValues(t, 600, mdb.GetBalance(b))
	mdb.RevertToSnapshot(v)
	require.EqualValues(t, 1000, mdb.GetBalance(a))
	require.EqualValues(t, 0, mdb.GetBalance(b))
	require.Empty(t, rec.keys, "reverted transfer leaked into the underlying StateDB")

	// 回滚后余额已恢复，后续转账不能再花被回滚的钱
	require.False(t, mdb.Transfer(b, c, 600))
	require.True(t, mdb.Transfer(a, c, 300))

	kvs, _ := mdb.GetChangedData(mdb.GetLastSnapshot().GetID())
	inReceipt := map[string]bool{}
	for _, kv := range kvs {
		inReceipt[string(kv.Key)] = true
	}
	require.NotContains(t, inReceipt, string(raw.AccountKey(b)))
	for _, k := range rec.keys {
		require.True(t, inReceipt[k], "key %s written to StateDB but missing from receipt", k)
	}
	require.EqualValues(t, 700, raw.LoadAccount(a).Balance)
	require.EqualValues(t, 0, raw.LoadAccount(b).Balance)
	require.EqualValues(t, 300, raw.LoadAccount(c).Balance)
}

// 嵌套快照：内层回滚只撤销内层写入，外层回滚撤销全部
func TestJournalKVNestedRevert(t *testing.T) {
	mem, err := db.NewGoMemDB("t", "", 0)
	require.NoError(t, err)
	j := newJournalKV(mem)
	k := []byte("k")
	j.snapshot(0)
	require.NoError(t, j.Set(k, []byte("1")))
	j.snapshot(1)
	require.NoError(t, j.Set(k, []byte("2")))
	require.NoError(t, j.Set(k, []byte("3")))
	j.revertTo(1)
	v, err := j.Get(k)
	require.NoError(t, err)
	require.Equal(t, "1", string(v))
	j.revertTo(0)
	_, err = j.Get(k)
	require.Equal(t, ctypes.ErrNotFound, err)
	require.Empty(t, j.changes())
}
