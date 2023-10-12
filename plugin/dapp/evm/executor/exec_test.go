package executor

import (
	"testing"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/system/crypto/secp256k1eth"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

func newEvmTestTx(nonce int64, priv crypto.PrivKey) *types.Transaction {

	tx := &types.Transaction{Execer: []byte("evm")}
	tx.Nonce = nonce
	tx.Payload = types.Encode(&evmtypes.EVMContractAction{})
	tx.Signature = &types.Signature{Ty: types.EncodeSignID(secp256k1eth.ID, 2)}
	tx.Signature.Pubkey = priv.PubKey().Bytes()
	return tx
}

func TestExecNonce(t *testing.T) {

	testCfg := types.NewChain33Config(types.GetDefaultCfgstring())
	util.ResetDatadir(testCfg.GetModuleConfig(), "$TEMP/")
	dbDir, stateDB, localDB := util.CreateTestDB()
	defer util.CloseTestDB(dbDir, stateDB)
	q := queue.New("channel")
	q.SetConfig(testCfg)
	qapi, _ := client.New(q.Client(), nil)

	exec := newEVMDriver()
	exec.SetAPI(qapi)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(localDB)
	exec.SetEnv(0, 1539918074, 1539918074)
	exec = exec.(*EVMExecutor)

	_, priv := util.Genaddress()
	recp := &types.ReceiptData{Ty: types.ExecOk}
	addr := address.PubKeyToAddr(2, priv.PubKey().Bytes())
	_, err := localDB.Get(secp256k1eth.CaculCoinsEvmAccountKey(addr))
	require.Equal(t, types.ErrNotFound, err)

	tx := newEvmTestTx(int64(3), priv)
	set, err := exec.ExecLocal(tx, recp, 3)
	require.Equal(t, nil, err)
	for _, kv := range set.GetKV() {
		err = localDB.Set(kv.Key, kv.Value)
		require.Nil(t, err)
	}
	nonceV, err := localDB.Get(secp256k1eth.CaculCoinsEvmAccountKey(addr))
	require.Equal(t, nil, err)
	evmNonce := &types.EvmAccountNonce{}
	_ = types.Decode(nonceV, evmNonce)
	require.Equal(t, int64(1), evmNonce.GetNonce())
	set, err = exec.ExecDelLocal(tx, recp, 0)
	require.Nil(t, err)
	for _, kv := range set.GetKV() {
		err = localDB.Set(kv.Key, kv.Value)
		require.Nil(t, err)
	}

	nonceV, err = localDB.Get(secp256k1eth.CaculCoinsEvmAccountKey(addr))
	require.Equal(t, nil, err)
	_ = types.Decode(nonceV, evmNonce)
	require.Equal(t, int64(0), evmNonce.GetNonce())

	// exec local
	count := 10
	for i := 0; i < count; i++ {
		//execpack execok
		recp.Ty = int32(i%2) + 1
		tx := newEvmTestTx(int64(i), priv)
		set, err := exec.ExecLocal(tx, recp, i)
		require.Nil(t, err)

		for _, kv := range set.GetKV() {
			err = localDB.Set(kv.Key, kv.Value)
			require.Nil(t, err)
		}
	}

	evmNonce = &types.EvmAccountNonce{}
	nonceV, err = localDB.Get(secp256k1eth.CaculCoinsEvmAccountKey(addr))
	require.Nil(t, err)
	_ = types.Decode(nonceV, evmNonce)
	require.Equal(t, int64(count), evmNonce.GetNonce())
	require.Equal(t, addr, evmNonce.Addr)
	// exec del local
	delSet := &types.LocalDBSet{}
	for i := count - 1; i >= 5; i-- {
		//execpack execok
		recp.Ty = int32(i%2) + 1
		tx = newEvmTestTx(int64(i), priv)
		set, err := exec.ExecDelLocal(tx, recp, i)
		require.Nil(t, err)
		delSet.KV = append(delSet.KV, set.GetKV()...)
	}
	for _, kv := range delSet.GetKV() {
		err = localDB.Set(kv.Key, kv.Value)
		require.Nil(t, err)
	}
	// nonce rollback to 5
	tx = newEvmTestTx(5, priv)
	_, err = exec.ExecLocal(tx, recp, 0)
	require.Nil(t, err)

}
