package executor

import (
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	des "github.com/33cn/plugin/plugin/dapp/storage/crypto"
	oty "github.com/33cn/plugin/plugin/dapp/storage/types"
	"github.com/stretchr/testify/assert"

	"strings"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4

	Nodes = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
	contents = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
	keys = [][]byte{
		[]byte("123456ab"),
		[]byte("G2F4ED5m123456abx6vDrScs"),
		[]byte("G2F4ED5m123456abx6vDrScsHD3psX7k"),
	}
	ivs = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
)
var (
	r *rand.Rand
)

func init() {
	r = rand.New(rand.NewSource(types.Now().UnixNano()))
}
func TestStorage(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	Init(oty.StorageX, cfg, nil)
	cfg.RegisterDappFork(oty.StorageX, oty.ForkStorageLocalDB, 0)
	total := 100 * types.Coin
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[0]),
	}
	accountB := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[1]),
	}

	accountC := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[2]),
	}
	accountD := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[3]),
	}
	execAddr := address.ExecAddress(oty.StorageX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 1000)
	_, _, kvdb := util.CreateTestDB()

	accA, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	accC, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accC.SaveExecAccount(execAddr, &accountC)

	accD, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accD.SaveExecAccount(execAddr, &accountD)

	env := &execEnv{
		10,
		cfg.GetDappFork(oty.StorageX, "Enable"),
		1539918074,
	}

	tx, err := CreateTx("ContentStorage", &oty.ContentOnlyNotaryStorage{Content: contents[0], Op: 0}, PrivKeyA, cfg)
	assert.Nil(t, err)
	Exec_Block(t, stateDB, kvdb, env, tx)
	txhash := common.ToHex(tx.Hash())
	//根据hash查询存储得明文内容
	reply, err := QueryStorageByKey(stateDB, kvdb, txhash, cfg)
	assert.Nil(t, err)
	assert.Equal(t, contents[0], reply.GetContentStorage().Content)

	//根据hash批量查询存储数据
	reply2, err := QueryBatchStorageByKey(stateDB, kvdb, &oty.BatchQueryStorage{TxHashs: []string{txhash}}, cfg)
	assert.Nil(t, err)
	assert.Equal(t, contents[0], reply2.Storages[0].GetContentStorage().Content)

	tx, err = CreateTx("ContentStorage", &oty.ContentOnlyNotaryStorage{Content: contents[1], Op: 1, Key: txhash}, PrivKeyA, cfg)
	assert.Nil(t, err)
	Exec_Block(t, stateDB, kvdb, env, tx)

	reply, err = QueryStorageByKey(stateDB, kvdb, txhash, cfg)
	assert.Nil(t, err)
	assert.Equal(t, append(append(contents[0], []byte(",")...), contents[1]...), reply.GetContentStorage().Content)

	tx, err = CreateTx("HashStorage", &oty.HashOnlyNotaryStorage{Hash: common.Sha256(contents[0])}, PrivKeyA, cfg)
	assert.Nil(t, err)
	Exec_Block(t, stateDB, kvdb, env, tx)
	txhash = common.ToHex(tx.Hash())
	//根据hash查询存储得明文内容
	reply, err = QueryStorageByKey(stateDB, kvdb, txhash, cfg)
	assert.Nil(t, err)
	assert.Equal(t, common.Sha256(contents[0]), reply.GetHashStorage().Hash)

	//存储链接地址
	tx, err = CreateTx("LinkStorage", &oty.LinkNotaryStorage{Hash: common.Sha256(contents[0]), Link: contents[0]}, PrivKeyA, cfg)
	assert.Nil(t, err)
	Exec_Block(t, stateDB, kvdb, env, tx)
	txhash = common.ToHex(tx.Hash())
	//根据hash查询存储得明文内容
	reply, err = QueryStorageByKey(stateDB, kvdb, txhash, cfg)
	assert.Nil(t, err)
	assert.Equal(t, common.Sha256(contents[0]), reply.GetLinkStorage().Hash)

	//加密存储
	aes := des.NewAES(keys[2], ivs[0])
	crypted, err := aes.Encrypt(contents[0])
	assert.Nil(t, err)
	tx, err = CreateTx("EncryptStorage", &oty.EncryptNotaryStorage{ContentHash: common.Sha256(contents[0]), EncryptContent: crypted, Nonce: ivs[0]}, PrivKeyA, cfg)
	assert.Nil(t, err)
	Exec_Block(t, stateDB, kvdb, env, tx)
	txhash = common.ToHex(tx.Hash())
	reply, err = QueryStorageByKey(stateDB, kvdb, txhash, cfg)
	assert.Nil(t, err)
	assert.Equal(t, common.Sha256(contents[0]), reply.GetEncryptStorage().ContentHash)
	assert.Equal(t, crypted, reply.GetEncryptStorage().EncryptContent)
	assert.Equal(t, ivs[0], reply.GetEncryptStorage().Nonce)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.New(types.GetSignName(oty.StorageX, signType))
	if err != nil {
		return tx, err
	}

	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return tx, err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return tx, err
	}

	tx.Sign(int32(signType), privKey)
	return tx, nil
}
func QueryStorageByKey(stateDB dbm.KV, kvdb dbm.KVDB, key string, cfg *types.Chain33Config) (*oty.Storage, error) {
	exec := newStorage()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	//根据hash查询存储得明文内容
	msg, err := exec.Query(oty.FuncNameQueryStorage, types.Encode(&oty.QueryStorage{
		TxHash: key}))
	if err != nil {
		return nil, err
	}
	return msg.(*oty.Storage), nil
}
func QueryBatchStorageByKey(stateDB dbm.KV, kvdb dbm.KVDB, para proto.Message, cfg *types.Chain33Config) (*oty.BatchReplyStorage, error) {
	exec := newStorage()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	//根据hash查询存储得明文内容
	msg, err := exec.Query(oty.FuncNameBatchQueryStorage, types.Encode(para))
	if err != nil {
		return nil, err
	}
	return msg.(*oty.BatchReplyStorage), nil
}
func CreateTx(action string, message types.Message, priv string, cfg *types.Chain33Config) (*types.Transaction, error) {
	ety := types.LoadExecutorType(oty.StorageX)
	tx, err := ety.Create(action, message)
	if err != nil {
		return nil, err
	}
	tx, err = types.FormatTx(cfg, oty.StorageX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, priv)
	return tx, err
}

//模拟区块中交易得执行过程
func Exec_Block(t *testing.T, stateDB dbm.DB, kvdb dbm.KVDB, env *execEnv, txs ...*types.Transaction) error {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.RegisterDappFork(oty.StorageX, oty.ForkStorageLocalDB, 0)
	cfg.SetTitleOnlyForTest("chain33")
	exec := newStorage()
	e := exec.(*storage)
	for index, tx := range txs {
		err := e.CheckTx(tx, index)
		if err != nil {
			t.Log(err.Error())
			return err
		}
	}
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 20
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	for index, tx := range txs {
		receipt, err := exec.Exec(tx, index)
		if err != nil {
			t.Log(err.Error())
			return err
		}
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
		receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
		set, err := exec.ExecLocal(tx, receiptData, index)
		if err != nil {
			t.Log(err.Error())
			return err
		}
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
		//save to database
		util.SaveKVList(stateDB, set.KV)
		assert.Equal(t, types.ExecOk, int(receipt.Ty))
	}
	return nil
}
