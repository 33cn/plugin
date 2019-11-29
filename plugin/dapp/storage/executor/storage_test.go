package executor

import (
	"math/rand"
	"testing"

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
func TestOrace(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	Init(oty.StorageX, cfg, nil)
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

	env := execEnv{
		10,
		cfg.GetDappFork(oty.StorageX, "Enable"),
		1539918074,
	}

	// publish event
	ety := types.LoadExecutorType(oty.StorageX)
	tx, err := ety.Create("ContentStorage", &oty.ContentOnlyNotaryStorage{Content: contents[0]})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.StorageX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	t.Log("tx", tx)
	exec := newStorage()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(1, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	var txhash string
	txhash = common.ToHex(tx.Hash())
	t.Log("txhash:", txhash)
	//根据hash查询存储得明文内容
	msg, err := exec.Query(oty.FuncNameQueryStorage, types.Encode(&oty.QueryStorage{
		TxHash: txhash}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply := msg.(*oty.Storage)
	assert.Equal(t, contents[0], reply.GetContentStorage().Content)

	//根据hash批量查询存储数据
	msg, err = exec.Query(oty.FuncNameBatchQueryStorage, types.Encode(&oty.BatchQueryStorage{
		TxHashs: []string{txhash}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)
	reply2 := msg.(*oty.BatchReplyStorage)
	assert.Equal(t, contents[0], reply2.Storages[0].GetContentStorage().Content)

	tx, err = ety.Create("HashStorage", &oty.HashOnlyNotaryStorage{Hash: common.Sha256(contents[0])})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.StorageX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	t.Log("tx", tx)
	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	txhash = common.ToHex(tx.Hash())
	t.Log("txhash:", txhash)
	//根据hash查询存储得明文内容
	msg, err = exec.Query(oty.FuncNameQueryStorage, types.Encode(&oty.QueryStorage{
		TxHash: txhash}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.Storage)
	assert.Equal(t, common.Sha256(contents[0]), reply.GetHashStorage().Hash)

	//根据hash批量查询存储数据
	msg, err = exec.Query(oty.FuncNameBatchQueryStorage, types.Encode(&oty.BatchQueryStorage{
		TxHashs: []string{txhash}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)
	reply2 = msg.(*oty.BatchReplyStorage)
	assert.Equal(t, common.Sha256(contents[0]), reply2.Storages[0].GetHashStorage().Hash)

	//存储链接地址
	tx, err = ety.Create("LinkStorage", &oty.LinkNotaryStorage{Hash: common.Sha256(contents[0]), Link: contents[0]})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.StorageX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	t.Log("tx", tx)
	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	txhash = common.ToHex(tx.Hash())
	t.Log("txhash:", txhash)
	//根据hash查询存储得明文内容
	msg, err = exec.Query(oty.FuncNameQueryStorage, types.Encode(&oty.QueryStorage{
		TxHash: txhash}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.Storage)
	assert.Equal(t, common.Sha256(contents[0]), reply.GetLinkStorage().Hash)

	//根据hash批量查询存储数据
	msg, err = exec.Query(oty.FuncNameBatchQueryStorage, types.Encode(&oty.BatchQueryStorage{
		TxHashs: []string{txhash}}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)
	reply2 = msg.(*oty.BatchReplyStorage)
	assert.Equal(t, common.Sha256(contents[0]), reply2.Storages[0].GetLinkStorage().Hash)

	//加密存储
	aes := des.NewAES(keys[2], ivs[0])
	crypted, err := aes.Encrypt(contents[0])
	if err != nil {
		t.Error(err)
	}
	tx, err = ety.Create("EncryptStorage", &oty.EncryptNotaryStorage{ContentHash: common.Sha256(contents[0]), EncryptContent: crypted, Nonce: ivs[0]})
	assert.Nil(t, err)
	tx, err = types.FormatTx(cfg, oty.StorageX, tx)
	assert.Nil(t, err)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	t.Log("tx", tx)
	exec.SetEnv(env.blockHeight+1, env.blockTime+20, env.difficulty+1)
	receipt, err = exec.Exec(tx, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx, receiptDate, int(1))
	if err != nil {
		t.Error(err)
	}
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
	txhash = common.ToHex(tx.Hash())
	t.Log("txhash:", txhash)
	//根据hash查询存储得明文内容
	msg, err = exec.Query(oty.FuncNameQueryStorage, types.Encode(&oty.QueryStorage{
		TxHash: txhash}))
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)

	reply = msg.(*oty.Storage)
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

//// golang中标准对称加密库测试
//func TestCryptoDES(t *testing.T){
//	key := []byte("123456")
//	result,err
//   des.NewCipher()
//}
//// golang中AES加密库测试
//func TestCryptoAES(t *testing.T){
//
//}
