package neutrino

import (
	"testing"

	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/stretchr/testify/require"
)

func TestPendingTxCache(t *testing.T) {

	c := newPendingTxCache(10)
	tx := &rtypes.PendingTx{ActionType: 1}
	c.addTx("test", tx)
	tx1 := c.getTx("test")
	require.Equal(t, tx.ActionType, tx1.ActionType)
	require.Equal(t, 1, len(c.pendingCache))
	tx2 := c.removeTx("test")
	require.Equal(t, tx, tx2)

	require.Equal(t, 0, len(c.pendingCache))

	tx.TxBlockHeight = 1
	c.addTx("test", tx)
	require.Equal(t, int64(1), c.getMinPendingHeight())

	c.addTx("other", &rtypes.PendingTx{TxBlockHeight: 10})
	require.Equal(t, int64(1), c.getMinPendingHeight())

	c.addTx("low", &rtypes.PendingTx{TxBlockHeight: 0})
	require.Equal(t, int64(0), c.getMinPendingHeight())
}
