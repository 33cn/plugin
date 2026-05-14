package types

import (
	"fmt"
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/btcsuite/btcd/txscript"
	"github.com/stretchr/testify/require"
)

func Test_GetActionName(t *testing.T) {

	require.Equal(t, NameMintAction, GetActionName(TyMintAction))
	require.Equal(t, NameTransferAction, GetActionName(TyTransferAction))
	require.Equal(t, NameConfirmAction, GetActionName(TyConfirmAction))

	require.Equal(t, "unknownAction", GetActionName(0))
}

func Test_UtxoAddress(t *testing.T) {

	utxoAddr := "74503993e7c8d4280f6fbb99ae5aaa92231a1981a358e40f97e2b4f4dfbea13c:0"
	require.True(t, IsUtxoAddress(utxoAddr))
	out, err := NewOutPointFromString(utxoAddr)
	require.NoError(t, err)
	require.Equal(t, utxoAddr, out.ToString())
}

func Test_btcCommitment(t *testing.T) {

	tx := &types.Transaction{Execer: []byte("coins")}
	commit := &BtcCommitment{
		Action:  TyConfirmAction,
		Payload: tx.Hash(),
	}
	script, err := txscript.NullDataScript(types.Encode(commit))
	require.NoError(t, err)
	fmt.Println(len(script))
}
