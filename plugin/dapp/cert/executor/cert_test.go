package executor

import (
	"fmt"
	"testing"
	"time"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/system/dapp"
	pty "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	_ "github.com/33cn/plugin/plugin/crypto/init"
	"github.com/33cn/plugin/plugin/dapp/cert/authority"
	"github.com/33cn/plugin/plugin/dapp/cert/authority/utils"
	ct "github.com/33cn/plugin/plugin/dapp/cert/types"
	pkt "github.com/33cn/plugin/plugin/dapp/collateralize/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
	kvdb        dbm.KVDB
	api         client.QueueProtocolAPI
	db          dbm.KV
	execAddr    string
	cfg         *types.Chain33Config
	ldb         dbm.DB
	user        *authority.User
}

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	Nodes    = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
	}
	total    = 100 * types.DefaultCoinPrecision
	USERNAME = "user1"
	ORGNAME  = "org1"
	SIGNTYPE = ct.AuthSM2

	transfer1 = &ct.CertAction{Value: &ct.CertAction_Normal{Normal: &ct.CertNormal{Key: "", Value: nil}}, Ty: ct.CertActionNormal}
	tx1       = &types.Transaction{Execer: []byte("cert"), Payload: types.Encode(transfer1), Fee: 100000000, Expire: 0, To: dapp.ExecAddress("cert")}

	transfer2 = &ct.CertAction{Value: &ct.CertAction_New{New: &ct.CertNew{Key: "", Value: nil}}, Ty: ct.CertActionNew}
	tx2       = &types.Transaction{Execer: []byte("cert"), Payload: types.Encode(transfer2), Fee: 100000000, Expire: 0, To: dapp.ExecAddress("cert")}

	transfer3 = &ct.CertAction{Value: &ct.CertAction_Update{Update: &ct.CertUpdate{Key: "", Value: nil}}, Ty: ct.CertActionUpdate}
	tx3       = &types.Transaction{Execer: []byte("cert"), Payload: types.Encode(transfer3), Fee: 100000000, Expire: 0, To: dapp.ExecAddress("cert")}
)

func manageKeySet(key string, value string, db dbm.KV) {
	var item types.ConfigItem
	item.Key = key
	item.Addr = value
	item.Ty = pty.ConfigItemArrayConfig
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, value)

	manageKey := types.ManageKey(key)
	valueSave := types.Encode(&item)
	db.Set([]byte(manageKey), valueSave)
}

func initEnv() (*execEnv, error) {
	cfg := types.NewChain33Config(types.ReadFile("./test/chain33.auth.test.toml"))
	cfg.SetTitleOnlyForTest("chain33")

	sub := cfg.GetSubConfig()
	var subcfg ct.Authority
	if sub.Exec["cert"] != nil {
		types.MustDecode(sub.Exec["cert"], &subcfg)
	}
	Init(ct.CertX, cfg, sub.Exec["cert"])

	userLoader := &authority.UserLoader{}
	err := userLoader.Init(subcfg.CryptoPath, subcfg.SignType)
	if err != nil {
		fmt.Printf("Init user loader falied -> %v", err)
		return nil, err
	}

	user, err := userLoader.Get(USERNAME, ORGNAME)
	if err != nil {
		fmt.Printf("Get user failed")
		return nil, err
	}

	_, ldb, kvdb := util.CreateTestDB()

	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    string(Nodes[0]),
	}

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)

	execAddr := dapp.ExecAddress(ct.CertX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)

	accA := account.NewCoinsAccount(cfg)
	accA.SetDB(stateDB)
	accA.SaveExecAccount(execAddr, &accountA)
	manageKeySet(ct.AdminKey, accountA.Addr, stateDB)

	return &execEnv{
		blockTime:   time.Now().Unix(),
		blockHeight: cfg.GetDappFork(ct.CertX, "Enable"),
		difficulty:  1539918074,
		kvdb:        kvdb,
		api:         api,
		db:          stateDB,
		execAddr:    execAddr,
		cfg:         cfg,
		ldb:         ldb,
		user:        user,
	}, nil
}

func signCertTx(tx *types.Transaction, priv crypto.PrivKey, cert []byte) {
	tx.Sign(int32(SIGNTYPE), priv)
	tx.Signature.Signature = utils.EncodeCertToSignature(tx.Signature.Signature, cert, nil)
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName(pkt.CollateralizeX, signType), -1)
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

func TestCert(t *testing.T) {
	env, err := initEnv()
	if err != nil {
		panic(err)
	}

	signCertTx(tx1, env.user.Key, env.user.Cert)

	// tx1
	exec := newCert()
	exec.SetAPI(env.api)
	exec.SetStateDB(env.db)
	assert.Equal(t, exec.GetCoinsAccount().LoadExecAccount(string(Nodes[0]), env.execAddr).GetBalance(), total)
	exec.SetLocalDB(env.kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err := exec.Exec(tx1, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx1, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	addr := address.PubKeyToAddr(address.DefaultID, env.user.Key.PubKey().Bytes())
	res, err := exec.Query("CertValidSNByAddr", types.Encode(&ct.ReqQueryValidCertSN{Addr: addr}))
	assert.Nil(t, err)
	assert.NotNil(t, res)

	// tx2
	signTx(tx2, PrivKeyA)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err = exec.Exec(tx2, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}

	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx2, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)

	// tx3
	signTx(tx3, PrivKeyA)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err = exec.Exec(tx3, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	t.Log(receipt)
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}

	receiptData = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(tx3, receiptData, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	util.SaveKVList(env.ldb, set.KV)
}
