package rollup

import (
	"testing"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/require"
)

func Test_filterParaCrossTx(t *testing.T) {

	tx1 := &types.Transaction{Execer: []byte("user.p.test.paracross"), Payload: []byte("test-tx1")}
	tx2 := &types.Transaction{Execer: []byte("user.p.test.none"), Payload: []byte("test-tx2")}

	crossTxs := filterParaCrossTx([]*types.Transaction{tx1, tx2})
	require.Equal(t, 0, len(crossTxs))
	crossTxs = filterParaCrossTx(nil)
	require.Equal(t, 0, len(crossTxs))

	tx2.Execer = []byte("user.p.test.paracross")
	tx2.Payload = types.Encode(&pt.ParacrossAction{Ty: pt.ParacrossActionCrossAssetTransfer})
	tx2.GroupCount = 2
	crossTxs = filterParaCrossTx([]*types.Transaction{tx1, tx2, tx1, tx2})
	require.Equal(t, 3, len(crossTxs))
	require.Equal(t, tx2.Hash(), crossTxs[0].Hash())
	require.Equal(t, tx1.Hash(), crossTxs[1].Hash())
	require.Equal(t, tx2.Hash(), crossTxs[2].Hash())
	tx1.GroupCount = 2
	crossTxs = filterParaCrossTx([]*types.Transaction{tx1, tx1, tx2})
	require.Equal(t, 1, len(crossTxs))
	require.Equal(t, tx2.Hash(), crossTxs[0].Hash())
	tx1.GroupCount = 3
	crossTxs = filterParaCrossTx([]*types.Transaction{tx1, tx1, tx2, tx1})
	require.Equal(t, 3, len(crossTxs))
	require.Equal(t, tx1.Hash(), crossTxs[0].Hash())
	require.Equal(t, tx1.Hash(), crossTxs[1].Hash())
	require.Equal(t, tx2.Hash(), crossTxs[2].Hash())
}
