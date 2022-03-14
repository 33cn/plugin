package executor

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/util"
	bridgevmxgo "github.com/33cn/plugin/plugin/dapp/bridgevmxgo/contracts/generated"
	"github.com/stretchr/testify/mock"

	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
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
	burnAmount      int64 = 200
	execName              = "evmxgo"
	transExecName         = "token"
	mananerexecName       = "manage"
	transToAddr           = "17EVv6tW2HzE73TVB6YXQYThQJxa7kuZb8"
	transAmount     int64 = 100
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
)

var (
	contractAddr       = "1AdSxpZKKbdaFNuydxvGktBtUYHmcuP6C5"
	lockAmt      int64 = 2000
	bridgeToken        = string(Nodes[0])
	recipient          = string(Nodes[0])
	// DefaultFeeRate 默认手续费率
	DefaultFeeRate int64 = 100000
)

const BridgeBankABIBridgevmxgo = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_operatorAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_oracleAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_goAssetBridgeAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_beneficiary\",\"type\":\"address\"}],\"name\":\"LogBridgeTokenMint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_ownerFrom\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_goAssetReceiver\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"LogGoAssetTokenBurn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"LogLock\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"LogNewBridgeToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"LogUnlock\",\"type\":\"event\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"addToken2LockList\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"bridgeTokenCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"bridgeTokenCreated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"bridgeTokenWhitelist\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_goAssetReceiver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_goAssetTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"burnBridgeTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_percents\",\"type\":\"uint8\"}],\"name\":\"configLockedTokenOfflineSave\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_offlineSave\",\"type\":\"address\"}],\"name\":\"configOfflineSaveAccount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"createNewBridgeToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_id\",\"type\":\"bytes32\"}],\"name\":\"getGoAssetDepositStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"getLockedTokenAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"getToken2address\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"getofflineSaveCfg\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"goAssetBridge\",\"outputs\":[{\"internalType\":\"contractGoAssetBridge\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"hasBridgeTokenCreated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"highThreshold\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_recipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"lock\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"lockedFunds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lowThreshold\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_goAssetSender\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_intendedRecipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_bridgeTokenAddress\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"mintBridgeTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"offlineSave\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"offlineSaveCfgs\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_percents\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"operator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"oracle\",\"outputs\":[{\"internalType\":\"contractOracle\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"token2address\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"tokenAllow2Lock\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_recipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"unlock\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_id\",\"type\":\"bytes32\"}],\"name\":\"viewGoAssetDeposit\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

func (e *execEnv) incr() {
	e.blockTime += 1
	e.blockHeight += 1
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
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
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
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		panic(err)
	}
	privto, err := cr.GenKey()
	if err != nil {
		panic(err)
	}
	addrto := address.PubKeyToAddr(address.DefaultID, privto.PubKey().Bytes())
	fmt.Println("addr:", addrto)

	fmt.Println(bridgevmxgo.BridgeBankBin)
	return addrto, privto
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
	c, err := crypto.Load(types.GetSignName(pty.EvmxgoX, signType), -1)
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

func Test_Token(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	Init(pty.EvmxgoX, cfg, nil)
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

	accA, _ := account.NewAccountDB(cfg, pty.EvmxgoX, Symbol, stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ := account.NewAccountDB(cfg, AssetExecPara, Symbol, stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	env := &execEnv{
		10,
		0,
		1539918074,
	}

	// set config key
	item := &types.ConfigItem{
		Key: fmt.Sprintf("mavl-manage-evmxgo-mint-%s", Symbol),
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{fmt.Sprintf("{\"address\":\"%s\",\"precision\":4,\"introduction\":\"介绍\"}", bridgeToken)}},
		},
	}
	stateDB.Set([]byte(item.Key), types.Encode(item))

	itemBridgevmxgoConfig := &types.ConfigItem{
		Key: "mavl-manage-bridgevmxgo-contract-addr",
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{fmt.Sprintf("{\"address\":\"%s\"}", contractAddr)}},
		},
	}
	stateDB.Set([]byte(itemBridgevmxgoConfig.Key), types.Encode(itemBridgevmxgoConfig))

	exec := newEvmxgo()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)

	receipt, err := evmxgo_Exec_Mint(exec, env)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}

	accCheck := accA.LoadAccount(recipient)
	assert.Equal(t, lockAmt, accCheck.Balance)

	set, err := evmxgo_Exec_Mint_Local(exec, receipt)
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	env.incr()
	receipt, err = evmxgo_Exec_Transfer(exec, env)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	set, err = evmxgo_Exec_Transfer_Local(exec, receipt)
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	env.incr()
	receipt, err = evmxgo_Exec_Transfer_Exec(exec, env)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	set, err = evmxgo_Exec_Transfer_Exec_Local(exec, receipt)
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	env.incr()
	receipt, err = evmxgo_Exec_Withdraw(exec, env)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	set, err = evmxgo_Exec_Withdraw_Local(exec, receipt)
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}

	env.incr()
	receipt, err = evmxgo_Exec_Burn(exec, env)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	//t.Log(receipt)
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
	set, err = evmxgo_Exec_Burn_Local(exec, receipt)
	assert.Nil(t, err)
	assert.NotNil(t, set)
	for _, kv := range set.KV {
		kvdb.Set(kv.Key, kv.Value)
	}
}

func evmxgo_Exec_Mint(exec dapp.Driver, env *execEnv) (*types.Receipt, error) {
	// evmxgo mint
	// lock
	parameter := fmt.Sprintf("lock(%s, %s, %d)", recipient, bridgeToken, lockAmt)
	_, packData, err := evmAbi.Pack(parameter, BridgeBankABIBridgevmxgo, false)
	if nil != err {
		fmt.Println("evmAbi.Pack", "Failed to do abi.Pack: ", err.Error())
		return nil, ErrTest
	}
	evmAction := &evmtypes.EVMContractAction{
		Amount:       0,
		GasLimit:     0,
		GasPrice:     0,
		Para:         packData,
		Note:         "",
		ContractAddr: contractAddr,
	}

	evmTx := &types.Transaction{Execer: []byte(evmtypes.ExecutorName), Payload: types.Encode(evmAction), Fee: fee}

	evmTx.Nonce = r.Int63()

	// 创建
	pMint := &pty.EvmxgoMint{
		Symbol:      Symbol,
		Amount:      lockAmt,
		BridgeToken: bridgeToken,
		Recipient:   recipient,
	}
	createTxMint, err := types.CallCreateTransaction(pty.EvmxgoX, "Mint", pMint)
	if err != nil {
		fmt.Println("RPC_Default_Process", "err", err)
	}

	txGroup := []*types.Transaction{evmTx, createTxMint}
	createTx, err := types.CreateTxGroup(txGroup, DefaultFeeRate)
	if err != nil {
		fmt.Println("RPC_Default_Process", "err", err)
	}
	_ = createTx.SignN(0, 1, privkey)
	_ = createTx.SignN(1, 1, privkey)

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)

	value, ok := exec.(*evmxgo)
	if !ok {
		fmt.Println("type error")
		return nil, ErrTest
	}
	value.SetTxs(txGroup)

	receipt, err := value.Exec_Mint(pMint, createTx.Tx(), 1)
	return receipt, err
}

func evmxgo_Exec_Mint_Local(exec dapp.Driver, receipt *types.Receipt) (*types.LocalDBSet, error) {
	pMint := &pty.EvmxgoMint{
		Symbol:      Symbol,
		Amount:      lockAmt,
		BridgeToken: bridgeToken,
		Recipient:   recipient,
	}
	createTxMint, err := types.CallCreateTransaction(pty.EvmxgoX, "Mint", pMint)
	if err != nil {
		fmt.Println("RPC_Default_Process", "err", err)
		return nil, err
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	return exec.ExecLocal(createTxMint, receiptDate, int(1))
}

func evmxgo_Exec_Transfer(exec dapp.Driver, env *execEnv) (*types.Receipt, error) {
	v := &pty.EvmxgoAction_Transfer{Transfer: &types.AssetsTransfer{Cointoken: Symbol, Amount: transAmount, Note: []byte(""), To: transToAddr}}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.ActionTransfer}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	Tx1, err := signTx(tx, PrivKeyA)
	if err != nil {
		fmt.Println("RPC_Default_Process sign", "err", err)
		return nil, err
	}

	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	return exec.Exec(Tx1, int(1))
}

func evmxgo_Exec_Transfer_Local(exec dapp.Driver, receipt *types.Receipt) (*types.LocalDBSet, error) {
	v := &pty.EvmxgoAction_Transfer{Transfer: &types.AssetsTransfer{Cointoken: Symbol, Amount: transAmount, Note: []byte(""), To: transToAddr}}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.ActionTransfer}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	Tx1, err := signTx(tx, PrivKeyA)
	if err != nil {
		fmt.Println("RPC_Default_Process sign", "err", err)
		return nil, err
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	return exec.ExecLocal(Tx1, receiptDate, int(1))
}

func evmxgo_Exec_Transfer_Exec(exec dapp.Driver, env *execEnv) (*types.Receipt, error) {
	v := &pty.EvmxgoAction_TransferToExec{TransferToExec: &types.AssetsTransferToExec{
		Cointoken: Symbol,
		Amount:    transAmount,
		Note:      []byte(""),
		ExecName:  transExecName,
		To:        address.ExecAddress(transExecName)},
	}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.EvmxgoActionTransferToExec}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	Tx1, err := signTx(tx, PrivKeyA)
	if err != nil {
		fmt.Println("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	return exec.Exec(Tx1, int(1))
}

func evmxgo_Exec_Transfer_Exec_Local(exec dapp.Driver, receipt *types.Receipt) (*types.LocalDBSet, error) {
	v := &pty.EvmxgoAction_TransferToExec{TransferToExec: &types.AssetsTransferToExec{
		Cointoken: Symbol,
		Amount:    transAmount,
		Note:      []byte(""),
		ExecName:  transExecName,
		To:        address.ExecAddress(transExecName)},
	}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.EvmxgoActionTransferToExec}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	Tx1, err := signTx(tx, PrivKeyA)
	if err != nil {
		fmt.Println("RPC_Default_Process sign", "err", err)
		return nil, err
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	return exec.ExecLocal(Tx1, receiptDate, int(1))
}

func evmxgo_Exec_Withdraw(exec dapp.Driver, env *execEnv) (*types.Receipt, error) {
	v := &pty.EvmxgoAction_Withdraw{Withdraw: &types.AssetsWithdraw{
		Cointoken: Symbol,
		Amount:    transAmount,
		Note:      []byte(""),
		ExecName:  transExecName,
		To:        address.ExecAddress(transExecName)},
	}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.ActionWithdraw}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	Tx1, err := signTx(tx, PrivKeyA)
	if err != nil {
		fmt.Println("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	return exec.Exec(Tx1, int(1))
}

func evmxgo_Exec_Withdraw_Local(exec dapp.Driver, receipt *types.Receipt) (*types.LocalDBSet, error) {
	v := &pty.EvmxgoAction_Withdraw{Withdraw: &types.AssetsWithdraw{
		Cointoken: Symbol,
		Amount:    transAmount,
		Note:      []byte(""),
		ExecName:  transExecName,
		To:        address.ExecAddress(transExecName)},
	}
	transfer := &pty.EvmxgoAction{Value: v, Ty: pty.ActionWithdraw}

	tx := &types.Transaction{Execer: []byte(execName), Payload: types.Encode(transfer), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	Tx1, err := signTx(tx, PrivKeyA)
	if err != nil {
		fmt.Println("RPC_Default_Process sign", "err", err)
		return nil, err
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	return exec.ExecLocal(Tx1, receiptDate, int(1))
}

func evmxgo_Exec_Burn(exec dapp.Driver, env *execEnv) (*types.Receipt, error) {
	p4 := &pty.EvmxgoBurn{
		Symbol: Symbol,
		Amount: burnAmount,
	}
	tx, err := types.CallCreateTransaction(pty.EvmxgoX, "Burn", p4)
	if err != nil {
		fmt.Println("RPC_Default_Process CallCreateTransaction", "err", err)
		return nil, err
	}

	Tx1, err := signTx(tx, PrivKeyA)
	if err != nil {
		fmt.Println("RPC_Default_Process sign", "err", err)
	}

	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	return exec.Exec(Tx1, int(1))
}

func evmxgo_Exec_Burn_Local(exec dapp.Driver, receipt *types.Receipt) (*types.LocalDBSet, error) {
	p4 := &pty.EvmxgoBurn{
		Symbol: Symbol,
		Amount: burnAmount,
	}
	tx, err := types.CallCreateTransaction(pty.EvmxgoX, "Burn", p4)
	if err != nil {
		fmt.Println("RPC_Default_Process CallCreateTransaction", "err", err)
		return nil, err
	}

	Tx1, err := signTx(tx, PrivKeyA)
	if err != nil {
		fmt.Println("RPC_Default_Process sign  test", "err", err)
		return nil, err
	}
	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	return exec.ExecLocal(Tx1, receiptDate, int(1))
}
