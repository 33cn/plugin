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
	total := 100 * types.DefaultCoinPrecision
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

	// 同态加密
	c1, _ := common.FromHex("010076833cbc35c9b7a87eda9aab7aafec45bd7eb8c0e9306141e7d1e84ecd87a5cb9f80343551bd9b0d04669fd74d020fca8837d9b3b471e3e13c125a06663d820cabb36c243747a11ea2601a9cb32e61c697dcdc846c492d954fa8c3ca2be662e8c0142138b647ce5d9e7c4c4d2ace8ba3c5699fbec98beef24e3c740967ec0b72b1626ecd4a7ce54960ae9bc1e2ea5594b05778f4f772aa213a06421744ed8dd775fb8f212bcc7e0da5fe4949051e6aa09d47db7e8d5028ecc41cea2aed4e23aefd80714519662fae980e3a6ca3defc0dd5596cb7e90da29e1bdd84db82603aed78fb8f98687898cd675da74452052ff1761446bf2ee922fc8c56b7e1beb2d4b91042742837052b340e4afce1836badf4a56ef775bb6a94fd91686a44122543fdab3ee90488160068769fbce2e74ffd5e052bb4c651969c4755f6eb5d7ea0cee3e7fa58f4b38b1722535909bd0f325a0b8d0797c15300ab06095b305f46497bbd75c3682d379387f55f638cd10639db1f050090bfd5d291718cee9fc6894d04ba6ef0acb477184512984f435b13e9bffe630bbc8ceade6d269487bf219e25b3beecc46afb98fb8e1fec1a9ad0af0a16e70611f9a4337af05a2ff82d7c9dcebeea9d8e2070413e49e29ff564e6dcee5696e39dc590ff8f8f553834c66c5ae730d05441e3fcc59de0d6efb12fd8852e19f2e309e4ac4cbf0634655fcaab5a70b4eab49829190ea2862c2568e8b69ff0324055f92275de05f2de1d6a8a78333370da754708f7a827f5e9bbe2e58d58294cafef37898a5a5a4866678b6e07941a3bfcb3c3f2d150f830a12d8e0dd20d756099be48f10d51521b3ba0c49f2ab139724ae3962999d75b88bf572fbfd86b6eef3bad5a446b97949d95f9e742500f8609ecfe189d2b4d5a47ea997164f48b3872ec525f16ec23fa5700d10d3385019edfabfec780f15b639f6863332c69fe6b17895620821fe6aaf94caab29c0fec19ff1bebc59f5ed8f973a3b720257cce803541406ae8e75163cf0049f8fbf14d239e86089cdefba8bcbb03db284283a1ff572320aeb7d4a139d4429e00c8bb196539d7")
	c2, _ := common.FromHex("010076833cbc35c9b7a87eda9aab7aafec45bd7eb8c0e9306141e7d1e84ecd87a5cb9f80343551bd9b0d04669fd74d020fca8837d9b3b471e3e13c125a06663d820cabb36c243747a11ea2601a9cb32e61c697dcdc846c492d954fa8c3ca2be662e8c0142138b647ce5d9e7c4c4d2ace8ba3c5699fbec98beef24e3c740967ec0b72b1626ecd4a7ce54960ae9bc1e2ea5594b05778f4f772aa213a06421744ed8dd775fb8f212bcc7e0da5fe4949051e6aa09d47db7e8d5028ecc41cea2aed4e23aefd80714519662fae980e3a6ca3defc0dd5596cb7e90da29e1bdd84db82603aed78fb8f98687898cd675da74452052ff1761446bf2ee922fc8c56b7e1beb2d4b9338ddba1f37e26ef7e0b9cb3bec18dc4b450d91a1a0901a3aca75000b58dd639635dce66553945a698b186c89443361ab4c1d3525de6bace217cd27fce0c8f6efc9e7c95269139487b5772182d8276ed6edfe0560537d18330aae4191af1dfae0a430b2021401bc17a111730f7114f514b7b84cb09bf717c67bb25c21a31b3a062f32dceb99103cdd622242148ec1799d04f4f3f4fc5af9e18cfaf356388b1413cc95b6f5bc0c293acf09ad0513e15d2ea525f120930e6072e0cf750e4e03bbf65b278ad92476f5e507bb51f01c3d2797931794cad156b5980fb5ec51fc82e99fde99e81dd85e9dba1d549aa8768c609e46171750b7bc636b709e47c92b076f1b7f71bff1fd690f8bfcdbf48559f777017b9ad300cbb3a1089f3eb6fed5632be9edecc33be09da7d43275640a097114f8f9e529289cce15d7b4ba613feda9e4d818743bf2741c5cd2aaaec7cc96ed835ee909ea0e9675f57bad4ad01688cc2a77e6181a0bca9a46d0ca160a1d94771e24db357e861e1f029104782445413c6d4861404df9ec3140b895083ebbb8a92bbc7decc8990353e9347a7933f85ef94ed325a331b3e6e6c752086adc7926abac3dfd8c8f3d7add64c80f327c1c4d74fde7a6c8981f79ede165f0584e5d6317d7272d9836098cd8ef82fbe85770962b709dfe1f2ae3454ac471e6ea24c6545f332a3eb4eecbc09e2276748d484a5216361")
	c3, _ := common.FromHex("010076833cbc35c9b7a87eda9aab7aafec45bd7eb8c0e9306141e7d1e84ecd87a5cb9f80343551bd9b0d04669fd74d020fca8837d9b3b471e3e13c125a06663d820cabb36c243747a11ea2601a9cb32e61c697dcdc846c492d954fa8c3ca2be662e8c0142138b647ce5d9e7c4c4d2ace8ba3c5699fbec98beef24e3c740967ec0b72b1626ecd4a7ce54960ae9bc1e2ea5594b05778f4f772aa213a06421744ed8dd775fb8f212bcc7e0da5fe4949051e6aa09d47db7e8d5028ecc41cea2aed4e23aefd80714519662fae980e3a6ca3defc0dd5596cb7e90da29e1bdd84db82603aed78fb8f98687898cd675da74452052ff1761446bf2ee922fc8c56b7e1beb2d4b90a734625f41c1709ad56e8a801346d70f190e7080c805de1e380198a4ebaf86ca6e5b9dbdc6eacad7c545f78c718a6e0262bc61d68244f99255d0fc8126cfbc0449e935bf5ee81af5e600f7d5b23b4b2d1f48482e037642690035f0289d4d3c72662d63e958edf8d566c765bcbc44cedc618b31db5eebdc6fd5848d53161e3e01c6c73010c32391db709a71cedb4c45fb71ec85de48ace0cba009ef2c0a88386e62b9607faf36c9b219508281267c0df2a182a88ad8fd146101564f5a14c7fd5ed2d4ae547c8741a31c560501230edd5417c622a7aba8d7efb08c08d5b7c45508c766353f2d43bd092834461f0673196abc8ef56b9fd54bdf82c2a392ff8dedcef3d90d3e47370d31083c2d19cefe2ce83936fd881f4a38ef3e5aaf3b0d48842ded5e6a16fc2fb7b5379406c2889e1d9de1fe611645310ba5b5048e817f44c451b5ec557c0e3b8e64049a437b82f901866ba4c1994d3c3caf02f81582f7409d10480df2c353eb63db4ec57671b2ac6c28fce842c605bae902ac5bb649f08199307458b542849d663809c9980a7c939b4f1f6d2eddf3dfe02f70ee4cf8956bf7a6a29a307939cd17eebdbcc80fb830bd545351a12c275375feaf3f690c52cc76dc7521ce87f621163b551a70e06d5bd3274bdc144f61f4ff8f1cdb2fcf97c67df1935057fdadaa23bf675892b4db841a63d49cd788c34498c70fb4ea787f3f8c1")

	tx, err = CreateTx("EncryptStorage", &oty.EncryptNotaryStorage{ContentHash: common.Sha256(c1), EncryptContent: c1, Nonce: ivs[0]}, PrivKeyA, cfg)
	assert.Nil(t, err)
	Exec_Block(t, stateDB, kvdb, env, tx)
	txhash = common.ToHex(tx.Hash())
	reply, err = QueryStorageByKey(stateDB, kvdb, txhash, cfg)
	assert.Nil(t, err)
	assert.Equal(t, common.Sha256(c1), reply.GetEncryptStorage().ContentHash)
	assert.Equal(t, c1, reply.GetEncryptStorage().EncryptContent)

	tx, err = CreateTx("EncryptAdd", &oty.EncryptNotaryAdd{Key: txhash, EncryptAdd: c2}, PrivKeyA, cfg)
	assert.Nil(t, err)
	Exec_Block(t, stateDB, kvdb, env, tx)
	reply, err = QueryStorageByKey(stateDB, kvdb, txhash, cfg)
	assert.Nil(t, err)
	assert.Equal(t, c3, reply.GetEncryptStorage().EncryptContent)

}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName(oty.StorageX, signType), -1)
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
