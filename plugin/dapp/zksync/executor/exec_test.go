package executor

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

var (
	isMainNetTest bool
	isParaNetTest bool

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

	burnAmount    int64 = 200
	transExecName       = "token"

	Symbol        = "TEST"
	AssetExecPara = "paracross"

	tokenId0 uint64 = 0
	tokenId1 uint64 = 1
	symbol0         = "ETH"
	symbol1         = "USDT"
	queueId  int64  = 0

	managementKey = "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	//managementKey = "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
	managementAddr = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"

	accountId4 uint64 = 4
	addr0             = "1LLa1hkV94eWqtFwQvTYdQYrrWe2MNhzNZ"
	key0              = "0xe7581f5b251b8d70aa019f897bdf46097af82a26d5f51dd2e33bb51bcae24326"
	l2addr0           = "27fd765bd3c8d09582c7dd3a0845513f77334b7a29eabd536d1354e17d4c3de9"
	accountId5 uint64 = 5
	addr1             = "1BTjxDZbZk6bsUavyxwvmm9LxYnxCZ6mFk"
	key1              = "0x9dc469bcb2e049c102a7fa5ed3242e2abdc55457923a608f9adec39758231ddf"
	l2addr1           = "16b0d05a930e62729c6f41090113ec8b168422eb2c9631b712ec52ef1275bfe9"
	accountId6 uint64 = 6
	addr2             = "1NKYDBrSmEKzmv3yttHfaxsnum9tazQSMV"
	key2              = "0xd4182bc90f9e10846f1af0d890d08d3680773884ffbd5195e18513da3c053696"
	l2addr2           = "23fc093f84848d1c3a975ca2fd8d3d02ac690303ab67d62881536f0d60522c8f"

	ethTestAddr    = "0xe51a1c7f7C704D8FcC192b242bE656fFB34A70e4"
	ethTestAddrKey = "0xd3064a91f01a60b0e3d92d08fc8be144a61a2a4a7780827727cd3804a66d31bd"

	withdrawFee  = 10000000000
	transferFee  = 10000000000
	proxyExitFee = 10000000000
	contractFee  = 10000000000

	configZksync = "\n[exec.sub.zksync]\n#管理员列表\nmanager=[\n    \"14KEKbYtKKQm4wMthSK9J4La4nAiidGozt\",\n    \"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv\",\n    \"0x991fb09dc31a44b3177673f330c582ac2ea168e0\"\n]\nethFeeAddr=\"0x832367164346888E248bd58b9A5f480299F1e88d\"\n \nlayer2FeeAddr=\"06140f1bf242cf182b6d1288f6d5d4d7f45aa0e7fdad7ffa99bffdfc2e66c770\""
)

const (
	//defaultAmount = 1e10
	fee = 1e6
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

func Test_All(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1) + configZksync)
	Init(zt.Zksync, cfg, nil)
	total := int64(100000)
	account0 := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    addr0,
	}
	//accountA := types.Account{
	//	Balance: total,
	//	Frozen:  0,
	//	Addr:    string(Nodes[0]),
	//}
	//accountB := types.Account{
	//	Balance: total,
	//	Frozen:  0,
	//	Addr:    string(Nodes[1]),
	//}

	execAddr := address.ExecAddress(zt.Zksync)
	stateDB, _ := dbm.NewGoMemDB("zkSync", "./", 100)
	_, _, kvdb := util.CreateTestDB()

	acc0, _ := account.NewAccountDB(cfg, zt.Zksync, Symbol, stateDB)
	acc0.SaveExecAccount(execAddr, &account0)

	//accA, _ := account.NewAccountDB(cfg, zt.Zksync, Symbol, stateDB)
	//accA.SaveExecAccount(execAddr, &accountA)
	//
	//accB, _ := account.NewAccountDB(cfg, AssetExecPara, Symbol, stateDB)
	//accB.SaveExecAccount(execAddr, &accountB)

	//env := &execEnv{
	//	10,
	//	0,
	//	1539918074,
	//}

	exec := NewZksync()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	reply, err := exec.Query("GetMaxAccountId", nil)
	fmt.Println(reply, err)

	test_Deposit(exec, stateDB, t)

	reply2, err := exec.Query("GetAccountById", types.Encode(&zt.ZkQueryReq{AccountId: accountId4}))
	fmt.Println(reply2, err)
	reply, err = exec.Query("GetMaxAccountId", nil)
	fmt.Println(reply, err)

	//accCheck := accA.LoadAccount(recipient)
	//assert.Equal(t, lockAmt, accCheck.Balance)

	//env.incr()
}

func checkBalance(exec dapp.Driver, accountId, tokenId uint64, balance string) bool {
	reply, err := exec.Query("GetTokenBalance", types.Encode(&zt.ZkQueryReq{AccountId: accountId, TokenId: tokenId}))
	if err != nil {
		return false
	}

	if len(reply.(*zt.ZkQueryResp).TokenBalances) < int(tokenId) || len(reply.(*zt.ZkQueryResp).TokenBalances) < 0 {
		return false
	}

	fmt.Println("Query balance result:", reply.(*zt.ZkQueryResp).TokenBalances[tokenId].Balance)
	return reply.(*zt.ZkQueryResp).TokenBalances[tokenId].Balance == balance
}

func test_Deposit(exec dapp.Driver, stateDB dbm.KV, t *testing.T) {
	param := &zt.ZkDeposit{
		TokenId:            tokenId0,
		Amount:             "8000000000000000000",
		EthAddress:         ethTestAddr,
		Chain33Addr:        l2addr0,
		EthPriorityQueueId: queueId,
	}

	v := &zt.ZksyncAction_Deposit{Deposit: param}
	deposit := &zt.ZksyncAction{Value: v, Ty: zt.TyDepositAction}

	tx := &types.Transaction{Execer: []byte(zt.Zksync), Payload: types.Encode(deposit), Fee: fee, To: addrexec}
	tx.Nonce = r.Int63()

	Tx1, err := signTx(tx, managementKey)
	assert.Nil(t, err)

	receipt, err := exec.Exec(Tx1, 1)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	for _, kv := range receipt.KV {
		_ = stateDB.Set(kv.Key, kv.Value)
	}

	equal := checkBalance(exec, accountId4, tokenId0, "8000000000000000000")
	assert.True(t, equal)
	queueId++
}

//func evmxgo_Exec_Mint_Local(exec dapp.Driver, receipt *types.Receipt) (*types.LocalDBSet, error) {
//pMint := &pty.EvmxgoMint{
//	Symbol:      Symbol,
//	Amount:      lockAmt,
//	BridgeToken: bridgeToken,
//	Recipient:   recipient,
//}
//createTxMint, err := types.CallCreateTransaction(zt.Zksync, "Mint", pMint)
//if err != nil {
//	fmt.Println("RPC_Default_Process", "err", err)
//	return nil, err
//}
//receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
//return exec.ExecLocal(createTxMint, receiptDate, int(1))
//}

//func (e *execEnv) incr() {
//	e.blockTime += 1
//	e.blockHeight += 1
//}

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

	//privkey = getprivkey(PrivKeyA)
	//privkeySupper = getprivkey(PrivKeyA)
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

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName(zt.Zksync, signType), -1)
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
