package executor

import (
	"context"
	"errors"
	"fmt"
	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/mock"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	manageTypes "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

var (
	isMainNetTest bool
	isParaNetTest bool
)

var (
	mainNetgrpcAddr = "localhost:8802"
	ParaNetgrpcAddr = "localhost:8902"

	mainClient types.Chain33Client
	paraClient types.Chain33Client
	r          *rand.Rand

	ErrTest = errors.New("ErrTest")

	addrexec       string
	manageaddrexec string
	privkey        crypto.PrivKey
	privkeySupper  crypto.PrivKey
)

const (
	//defaultAmount = 1e10
	fee = 1e6
)

var (
	mintAmount      int64 = 1000 * types.DefaultCoinPrecision
	burnAmount      int64 = 200 * types.DefaultCoinPrecision
	execName              = "evmxgo"
	transExecName         = "token"
	mananerexecName       = "manage"
	transToAddr           = "17EVv6tW2HzE73TVB6YXQYThQJxa7kuZb8"
	transToExecAddr       = "12hpJBHybh1mSyCijQ2MQJPk7z7kZ7jnQa"
	transAmount     int64 = 100 * types.DefaultCoinPrecision
	walletPass            = "test1234"
)

var (
	Symbol        = "TEST"
	AssetExecPara = "paracross"

	PrivKeyA = "0x4dcb00c7d01a3d377c0d5a14cd7ec91798a74c8b41896c5d21fc8b9bf4b40e42" // 1jQtMFpmEW2J4fgEG3opEoVzfSyMpBMaR
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	Nodes    = [][]byte{
		[]byte("1jQtMFpmEW2J4fgEG3opEoVzfSyMpBMaR"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}

	MintBridgeToken = "BridgeTEST"
	MintRecipient   = string(Nodes[0])
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

func init() {
	fmt.Println("Init start")
	defer fmt.Println("Init end")

	isMainNetTest = true
	if !isMainNetTest && !isParaNetTest {
		return
	}

	conn, err := grpc.Dial(mainNetgrpcAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	mainClient = types.NewChain33Client(conn)

	conn, err = grpc.Dial(ParaNetgrpcAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	paraClient = types.NewChain33Client(conn)

	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	addrexec = address.ExecAddress("")
	manageaddrexec = address.ExecAddress("manage")

	privkey = getprivkey(PrivKeyA)
	privkeySupper = getprivkey(PrivKeyA)
}

func getprivkey(key string) crypto.PrivKey {
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		panic(err)
	}
	bkey, err := common.FromHex(key)
	if err != nil {
		panic(err)
	}
	priv, err := cr.PrivKeyFromBytes(bkey)
	if err != nil {
		panic(err)
	}
	return priv
}

func genaddress() (string, crypto.PrivKey) {
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		panic(err)
	}
	privto, err := cr.GenKey()
	if err != nil {
		panic(err)
	}
	addrto := address.PubKeyToAddress(privto.PubKey().Bytes())
	fmt.Println("addr:", addrto.String())
	return addrto.String(), privto
}

func waitTx(hash []byte) bool {
	i := 0
	for {
		i++
		if i%100 == 0 {
			fmt.Println("wait transaction timeout")
			return false
		}

		var reqHash types.ReqHash
		reqHash.Hash = hash
		res, err := mainClient.QueryTransaction(context.Background(), &reqHash)
		if err != nil {
			time.Sleep(time.Second)
		}
		if res != nil {
			return true
		}
	}
}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.New(types.GetSignName(pty.EvmxgoX, signType))
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

func Test_InitAccount(t *testing.T) {
	if !isMainNetTest {
		return
	}
	fmt.Println("TestInitAccount start")
	defer fmt.Println("TestInitAccount end")

	//need update to fixed addr here
	//addr = ""
	//privkey = ""
	//addr, privkey = genaddress()
	privkey = getprivkey(PrivKeyA)
	label := strconv.Itoa(int(types.Now().UnixNano()))
	params := types.ReqWalletImportPrivkey{Privkey: common.ToHex(privkey.Bytes()), Label: label}

	unlock := types.WalletUnLock{Passwd: walletPass, Timeout: 0, WalletOrTicket: false}
	_, err := mainClient.UnLock(context.Background(), &unlock)
	if err != nil {
		fmt.Println(err)
		t.Error(err)
		return
	}
	time.Sleep(5 * time.Second)

	_, err = mainClient.ImportPrivkey(context.Background(), &params)
	if err != nil && err != types.ErrPrivkeyExist {
		fmt.Println(err)
		t.Error(err)
		return
	}
	time.Sleep(5 * time.Second)
	/*
		txhash, err := sendtoaddress(mainClient, privGenesis, addr, defaultAmount)

		if err != nil {
			t.Error(err)
			return
		}
		if !waitTx(txhash) {
			t.Error(ErrTest)
			return
		}

		time.Sleep(5 * time.Second)
	*/
}

func Test_Token(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	Init(pty.EvmxgoX, cfg, nil)
	tokenTotal := int64(10000 * 1e8)
	tokenBurn := int64(10 * 1e8)
	tokenMint := int64(20 * 1e8)
	total := int64(100000)
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

	execAddr := address.ExecAddress(pty.EvmxgoX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)
	_, _, kvdb := util.CreateTestDB()

	accA, _ := account.NewAccountDB(cfg, AssetExecPara, Symbol, stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ := account.NewAccountDB(cfg, AssetExecPara, Symbol, stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	env := execEnv{
		10,
		0,
		1539918074,
	}

	// set config key
	item := &types.ConfigItem{
		Key: fmt.Sprintf("mavl-manage-evmxgo-mint-%s", Symbol),
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{"{\"address\":\"address1234\",\"precision\":4,\"introduction\":\"介绍\"}"}},
		},
	}
	stateDB.Set([]byte(item.Key), types.Encode(item))

	exec := newEvmxgo()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	accDB, _ := account.NewAccountDB(cfg, pty.EvmxgoX, Symbol, stateDB)

	// evmxgo mint
	// 创建
	pMint := &pty.EvmxgoMint{
		Symbol:      Symbol,
		Amount:      tokenMint,
		BridgeToken: MintBridgeToken,
		Recipient:   MintRecipient,
	}
	//v, _ := types.PBToJSON(p1)
	createTxMint, err := types.CallCreateTransaction(pty.EvmxgoX, "Mint", pMint)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTxMint, err = signTx(createTxMint, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	receipt, err := exec.Exec(createTxMint, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	accCheck := accDB.LoadAccount(MintRecipient)
	assert.Equal(t, tokenTotal+tokenMint, accCheck.Balance)

	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(createTxMint, receiptDate, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	p4 := &pty.EvmxgoBurn{
		Symbol: Symbol,
		Amount: tokenBurn,
	}
	createTx4, err := types.CallCreateTransaction(pty.EvmxgoX, "Burn", p4)
	if err != nil {
		t.Error("RPC_Default_Process", "err", err)
	}
	createTx4, err = signTx(createTx4, PrivKeyA)
	if err != nil {
		t.Error("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	receipt, err = exec.Exec(createTx4, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	accCheck = accDB.LoadAccount(string(Nodes[0]))
	assert.Equal(t, tokenTotal+tokenMint-tokenBurn, accCheck.Balance)

	receiptDate = &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err = exec.ExecLocal(createTx4, receiptDate, int(1))
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
}

func Test_AddConfig(t *testing.T) {
	if !isMainNetTest {
		return
	}
	fmt.Println("Test_AddConfig start")
	defer fmt.Println("Test_AddConfig end")

	v := &manageTypes.ManageAction_Modify{Modify: &types.ModifyConfig{
		Key:   fmt.Sprintf("evmxgo-mint-%s", Symbol),
		Value: "{\"address\":\"address1234\",\"precision\":4,\"introduction\":\"介绍\"}",
		Op:    "add",
		//Op: "delete",
		Addr: "",
	}}
	action := &manageTypes.ManageAction{Value: v, Ty: manageTypes.ManageActionModifyConfig}

	tx := &types.Transaction{Execer: []byte(mananerexecName), Payload: types.Encode(action), Fee: fee, To: manageaddrexec}
	tx.Nonce = r.Int63()

	version, _ := mainClient.Version(context.Background(), nil)
	tx.ChainID = version.GetChainID()
	tx.Sign(types.SECP256K1, privkeySupper)

	reply, err := mainClient.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("err", err)
		t.Error(err)
		return
	}
	if !reply.IsOk {
		fmt.Println("err = ", reply.GetMsg())
		t.Error(ErrTest)
		return
	}
	fmt.Println("err = ", reply.GetMsg())
	if !waitTx(tx.Hash()) {
		t.Error(ErrTest)
		return
	}
}

func Test_EvmxgoMint(t *testing.T) {
	if !isMainNetTest {
		return
	}
	fmt.Println("Test_EvmxgoMint start")
	defer fmt.Println("Test_EvmxgoMint end")

	v := &pty.EvmxgoAction_Mint{Mint: &pty.EvmxgoMint{
		Symbol:      Symbol,
		Amount:      mintAmount,
		BridgeToken: MintBridgeToken,
		Recipient:   MintRecipient,
	}}
	mint := &pty.EvmxgoAction{Value: v, Ty: pty.EvmxgoActionMint}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(mint), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	version, _ := mainClient.Version(context.Background(), nil)
	tx.ChainID = version.GetChainID()
	tx.Sign(types.SECP256K1, privkey)

	reply, err := mainClient.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("err", err)
		t.Error(err)
		return
	}
	if !reply.IsOk {
		fmt.Println("err = ", reply.GetMsg())
		t.Error(ErrTest)
		return
	}
	fmt.Println("err = ", reply.GetMsg())
	if !waitTx(tx.Hash()) {
		t.Error(ErrTest)
		return
	}
}

func Test_EvmxgoBurn(t *testing.T) {
	if !isMainNetTest {
		return
	}
	fmt.Println("Test_EvmxgoBurn start")
	defer fmt.Println("Test_EvmxgoBurn end")

	v := &pty.EvmxgoAction_Burn{Burn: &pty.EvmxgoBurn{
		Symbol: Symbol,
		Amount: burnAmount,
	}}
	burn := &pty.EvmxgoAction{Value: v, Ty: pty.EvmxgoActionBurn}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(burn), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	version, _ := mainClient.Version(context.Background(), nil)
	tx.ChainID = version.GetChainID()
	tx.Sign(types.SECP256K1, privkey)

	reply, err := mainClient.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("err", err)
		t.Error(err)
		return
	}
	if !reply.IsOk {
		fmt.Println("err = ", reply.GetMsg())
		t.Error(ErrTest)
		return
	}

	if !waitTx(tx.Hash()) {
		t.Error(ErrTest)
		return
	}
}

func Test_Transfer(t *testing.T) {
	if !isMainNetTest {
		return
	}
	fmt.Println("Test_Transfer start")
	defer fmt.Println("Test_Transfer end")

	v := &pty.EvmxgoAction_Transfer{Transfer: &types.AssetsTransfer{Cointoken: Symbol, Amount: transAmount, Note: []byte(""), To: transToAddr}}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.ActionTransfer}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	version, _ := mainClient.Version(context.Background(), nil)
	tx.ChainID = version.GetChainID()
	tx.Sign(types.SECP256K1, privkey)

	reply, err := mainClient.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("err", err)
		t.Error(err)
		return
	}
	if !reply.IsOk {
		fmt.Println("err = ", reply.GetMsg())
		t.Error(ErrTest)
		return
	}

	if !waitTx(tx.Hash()) {
		t.Error(ErrTest)
		return
	}
}

func Test_TransferExec(t *testing.T) {
	if !isMainNetTest {
		return
	}
	fmt.Println("Test_TransferExec start")
	defer fmt.Println("Test_TransferExec end")

	v := &pty.EvmxgoAction_TransferToExec{TransferToExec: &types.AssetsTransferToExec{
		Cointoken: Symbol,
		Amount:    transAmount,
		Note:      []byte(""),
		ExecName:  transExecName,
		To:        transToExecAddr},
	}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.EvmxgoActionTransferToExec}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	version, _ := mainClient.Version(context.Background(), nil)
	tx.ChainID = version.GetChainID()
	tx.Sign(types.SECP256K1, privkey)

	reply, err := mainClient.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("err", err)
		t.Error(err)
		return
	}
	if !reply.IsOk {
		fmt.Println("err = ", reply.GetMsg())
		t.Error(ErrTest)
		return
	}

	if !waitTx(tx.Hash()) {
		t.Error(ErrTest)
		return
	}
}

func Test_Withdraw(t *testing.T) {
	if !isMainNetTest {
		return
	}
	fmt.Println("Test_Withdraw start")
	defer fmt.Println("Test_Withdraw end")

	v := &pty.EvmxgoAction_Withdraw{Withdraw: &types.AssetsWithdraw{
		Cointoken: Symbol,
		Amount:    transAmount,
		Note:      []byte(""),
		ExecName:  transExecName,
		To:        transToExecAddr},
	}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.ActionWithdraw}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	version, _ := mainClient.Version(context.Background(), nil)
	tx.ChainID = version.GetChainID()
	tx.Sign(types.SECP256K1, privkey)

	reply, err := mainClient.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("err", err)
		t.Error(err)
		return
	}
	if !reply.IsOk {
		fmt.Println("err = ", reply.GetMsg())
		t.Error(ErrTest)
		return
	}

	if !waitTx(tx.Hash()) {
		t.Error(ErrTest)
		return
	}
}
