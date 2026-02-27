package executor

import (
	"bytes"
	"testing"
	"time"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

func TestLightclientCheckTxBasic(t *testing.T) {
	cli := newLightclient().(*lightclient)
	addr, priv := util.Genaddress()
	lightCfg.CommitAddress = addr
	lightCfg.BtcNetName = "regtest"

	tx := &types.Transaction{Payload: []byte("invalid")}
	require.Equal(t, ErrDecodeAction, cli.CheckTx(tx, 0))

	action := &ltypes.LightClientAction{Ty: 999}
	tx.Payload = types.Encode(action)
	tx.Sign(types.SECP256K1, priv)
	require.Equal(t, types.ErrActionNotSupport, cli.CheckTx(tx, 0))
}

func TestLightclientCheckBtcHeaders(t *testing.T) {
	dir, stateDB, localDB := util.CreateTestDB()
	defer util.CloseTestDB(dir, stateDB)

	cli := newLightclient().(*lightclient)
	setupTestDriver(t, cli)
	cli.SetStateDB(stateDB)
	cli.SetLocalDB(localDB)

	commitAddr, commitPriv := util.Genaddress()
	_, illegalPriv := util.Genaddress()
	lightCfg.CommitAddress = commitAddr
	lightCfg.BtcNetName = "regtest"

	regtest := &chaincfg.RegressionNetParams
	ts := types.Now().Add(-time.Hour)
	prev := mineBtcHeader(t, nil, 100, regtest.PowLimitBits, ts)
	next := mineBtcHeader(t, prev, 101, regtest.PowLimitBits, ts.Add(time.Minute))

	require.NoError(t, stateDB.Set(btcLastHeaderKey(), types.Encode(prev)))
	require.NoError(t, localDB.Set(btcHeaderKey(prev.Height), types.Encode(prev)))

	t.Run("illegal commit address", func(t *testing.T) {
		tx := buildCheckTx(t, &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{next}}, illegalPriv)
		require.Equal(t, ErrIllegalCommitAddress, cli.CheckTx(tx, 0))
	})

	t.Run("empty headers", func(t *testing.T) {
		tx := buildCheckTx(t, &ltypes.BtcHeaders{}, commitPriv)
		require.Equal(t, types.ErrInvalidParam, cli.CheckTx(tx, 0))
	})

	t.Run("nil header", func(t *testing.T) {
		tx := buildCheckTx(t, &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{nil}}, commitPriv)
		require.Equal(t, ErrBtcHeaderDisorder, cli.CheckTx(tx, 0))
	})

	t.Run("header disorder", func(t *testing.T) {
		bad := cloneHeader(next)
		bad.PreviousHash = chainhash.Hash{}.String()
		tx := buildCheckTx(t, &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{bad}}, commitPriv)
		require.Equal(t, ErrBtcHeaderDisorder, cli.CheckTx(tx, 0))
	})

	t.Run("invalid wire header", func(t *testing.T) {
		bad := cloneHeader(next)
		bad.MerkleRoot = "not-a-hash"
		tx := buildCheckTx(t, &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{bad}}, commitPriv)
		require.Equal(t, ErrToBtcWireHeader, cli.CheckTx(tx, 0))
	})

	t.Run("invalid header hash", func(t *testing.T) {
		bad := cloneHeader(next)
		bad.Hash = prev.Hash
		tx := buildCheckTx(t, &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{bad}}, commitPriv)
		require.Equal(t, ErrInvalidBtcBlockHash, cli.CheckTx(tx, 0))
	})

	t.Run("target bits mismatch", func(t *testing.T) {
		badBitsHeader := mineBtcHeader(t, prev, 101, regtest.PowLimitBits-1, ts.Add(2*time.Minute))
		tx := buildCheckTx(t, &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{badBitsHeader}}, commitPriv)
		require.Equal(t, ErrBtcTargetBits, cli.CheckTx(tx, 0))
	})

	t.Run("success", func(t *testing.T) {
		tx := buildCheckTx(t, &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{next}}, commitPriv)
		require.NoError(t, cli.CheckTx(tx, 0))
	})
}

func TestLightclientCheckBtcHeadersBootstrap(t *testing.T) {
	dir, stateDB, localDB := util.CreateTestDB()
	defer util.CloseTestDB(dir, stateDB)

	cli := newLightclient().(*lightclient)
	setupTestDriver(t, cli)
	cli.SetStateDB(stateDB)
	cli.SetLocalDB(localDB)

	commitAddr, commitPriv := util.Genaddress()
	lightCfg.CommitAddress = commitAddr
	lightCfg.BtcNetName = "regtest"

	regtest := &chaincfg.RegressionNetParams
	ts := types.Now().Add(-time.Hour)
	h1 := mineBtcHeader(t, nil, 1, regtest.PowLimitBits, ts)
	h2 := mineBtcHeader(t, h1, 2, regtest.PowLimitBits, ts.Add(time.Minute))

	tx := buildCheckTx(t, &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{h1, h2}}, commitPriv)
	require.NoError(t, cli.CheckTx(tx, 0))
}

func TestLightclientExecBtcHeaders(t *testing.T) {
	t.Run("get last header error", func(t *testing.T) {
		dir, stateDB, _ := util.CreateTestDB()
		defer util.CloseTestDB(dir, stateDB)
		cli := newLightclient().(*lightclient)
		setupTestDriver(t, cli)
		cli.SetStateDB(stateDB)
		// Corrupt data to force decode error.
		require.NoError(t, stateDB.Set(btcLastHeaderKey(), []byte("bad-data")))

		headers := &ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{{Height: 1, Hash: chainhash.Hash{}.String()}}}
		recp, err := cli.Exec_BtcHeaders(headers, &types.Transaction{}, 0)
		require.Nil(t, recp)
		require.Equal(t, ErrBtcGetLastHeader, err)
	})

	t.Run("success", func(t *testing.T) {
		dir, stateDB, _ := util.CreateTestDB()
		defer util.CloseTestDB(dir, stateDB)
		cli := newLightclient().(*lightclient)
		setupTestDriver(t, cli)
		cli.SetStateDB(stateDB)

		prev := &ltypes.BtcHeader{Height: 100, Hash: "prevhash"}
		require.NoError(t, stateDB.Set(btcLastHeaderKey(), types.Encode(prev)))
		commit := &ltypes.BtcHeader{Height: 101, Hash: "commithash", Confirmations: 6}

		recp, err := cli.Exec_BtcHeaders(&ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{commit}}, &types.Transaction{}, 0)
		require.NoError(t, err)
		require.EqualValues(t, types.ExecOk, recp.Ty)
		require.Len(t, recp.KV, 1)
		require.Equal(t, btcLastHeaderKey(), recp.KV[0].Key)

		saved := &ltypes.BtcHeader{}
		require.NoError(t, types.Decode(recp.KV[0].Value, saved))
		require.Equal(t, commit.Height, saved.Height)
		require.Equal(t, commit.Hash, saved.Hash)

		require.Len(t, recp.Logs, 1)
		log := &ltypes.BtcHeadersLog{}
		require.NoError(t, types.Decode(recp.Logs[0].Log, log))
		require.Equal(t, prev.Height, log.LastHeight)
		require.Equal(t, prev.Hash, log.LastHash)
		require.Equal(t, commit.Height, log.CommitHeight)
		require.Equal(t, commit.Hash, log.CommitHash)
		require.Equal(t, commit.Confirmations, log.Confirmations)
	})
}

func TestLightclientExecLocalBtcHeaders(t *testing.T) {
	dir, stateDB, localDB := util.CreateTestDB()
	defer util.CloseTestDB(dir, stateDB)

	cli := newLightclient().(*lightclient)
	setupTestDriver(t, cli)
	cli.SetStateDB(stateDB)
	cli.SetLocalDB(localDB)

	h1 := &ltypes.BtcHeader{Height: 11, Hash: "h11"}
	h2 := &ltypes.BtcHeader{Height: 12, Hash: "h12"}
	tx := &types.Transaction{Execer: []byte(ltypes.LightclientX)}
	dbSet, err := cli.ExecLocal_BtcHeaders(&ltypes.BtcHeaders{Headers: []*ltypes.BtcHeader{h1, h2}}, tx, nil, 0)
	require.NoError(t, err)
	require.NotNil(t, dbSet)
	require.GreaterOrEqual(t, len(dbSet.KV), 4)

	require.True(t, hasKV(dbSet.KV, btcHeaderKey(h1.Height)))
	require.True(t, hasKV(dbSet.KV, btcHeaderHashHeightKey(h1.Hash)))
	require.True(t, hasKV(dbSet.KV, btcHeaderKey(h2.Height)))
	require.True(t, hasKV(dbSet.KV, btcHeaderHashHeightKey(h2.Hash)))
}

func buildCheckTx(t *testing.T, headers *ltypes.BtcHeaders, priv crypto.PrivKey) *types.Transaction {
	t.Helper()
	action := &ltypes.LightClientAction{
		Ty: ltypes.TyBtcHeadersAction,
		Value: &ltypes.LightClientAction_BtcHeaders{
			BtcHeaders: headers,
		},
	}
	tx := &types.Transaction{Payload: types.Encode(action)}
	tx.Sign(types.SECP256K1, priv)
	return tx
}

func mineBtcHeader(t *testing.T, prev *ltypes.BtcHeader, height uint64, bits uint32, ts time.Time) *ltypes.BtcHeader {
	t.Helper()

	prevHash := chainhash.Hash{}.String()
	if prev != nil {
		prevHash = prev.Hash
	}
	pre, err := chainhash.NewHashFromStr(prevHash)
	require.NoError(t, err)
	merkle := chainhash.DoubleHashH([]byte{byte(height), byte(height >> 8)})

	head := &wire.BlockHeader{
		Version:    1,
		PrevBlock:  *pre,
		MerkleRoot: merkle,
		Timestamp:  ts,
		Bits:       bits,
	}
	target := blockchain.CompactToBig(bits)
	for {
		hash := head.BlockHash()
		if blockchain.HashToBig(&hash).Cmp(target) <= 0 {
			return &ltypes.BtcHeader{
				Hash:          hash.String(),
				Height:        height,
				Version:       uint32(head.Version),
				MerkleRoot:    head.MerkleRoot.String(),
				Time:          head.Timestamp.Unix(),
				Nonce:         uint64(head.Nonce),
				Bits:          int64(head.Bits),
				PreviousHash:  head.PrevBlock.String(),
				Confirmations: 0,
			}
		}
		head.Nonce++
		if head.Nonce == 0 {
			head.Timestamp = head.Timestamp.Add(time.Second)
		}
	}
}

func hasKV(kvs []*types.KeyValue, key []byte) bool {
	for _, kv := range kvs {
		if bytes.Equal(kv.Key, key) {
			return true
		}
	}
	return false
}

func cloneHeader(h *ltypes.BtcHeader) *ltypes.BtcHeader {
	return &ltypes.BtcHeader{
		Hash:          h.GetHash(),
		Confirmations: h.GetConfirmations(),
		Height:        h.GetHeight(),
		Version:       h.GetVersion(),
		MerkleRoot:    h.GetMerkleRoot(),
		Time:          h.GetTime(),
		Nonce:         h.GetNonce(),
		Bits:          h.GetBits(),
		Difficulty:    h.GetDifficulty(),
		PreviousHash:  h.GetPreviousHash(),
		NextHash:      h.GetNextHash(),
	}
}

func setupTestDriver(t *testing.T, cli *lightclient) {
	t.Helper()
	api := &mocks.QueueProtocolAPI{}
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	cli.SetAPI(api)
}
