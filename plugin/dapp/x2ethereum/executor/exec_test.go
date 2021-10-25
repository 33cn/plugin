package executor

import (
	"fmt"
	"math"
	"strconv"

	//"github.com/33cn/chain33/client"
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/mocks"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	chain33types "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	types2 "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var chainTestCfg = types.NewChain33Config(types.GetDefaultCfgstring())

func init() {
	Init(types2.X2ethereumX, chainTestCfg, nil)
}

var (
	chain33Receiver       = "1BqP2vHkYNjSgdnTqm7pGbnphLhtEhuJFi"
	bridgeContractAddress = "0xC4cE93a5699c68241fc2fB503Fb0f21724A624BB"
	symbol                = "eth"
	tokenContractAddress  = "0x0000000000000000000000000000000000000000"
	ethereumAddr          = "0x7B95B6EC7EbD73572298cEf32Bb54FA408207359"
	addValidator1         = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
	addValidator2         = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
	privFrom              = getprivkey("4257d8692ef7fe13c68b65d6a52f03933db2fa5ce8faf210b5b8b80c721ced01") // 12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	tx                    = &types.Transaction{}
	sdb                   *db.GoMemDB
	kvdb                  db.KVDB
)

type suiteX2Ethereum struct {
	suite.Suite
	kvdb      *mocks.KVDB
	x2eth     *x2ethereum
	addrX2Eth string
	action    *action
}

func TestRunSuiteX2Ethereum(t *testing.T) {
	log := new(suiteX2Ethereum)
	suite.Run(t, log)
}

func (x *suiteX2Ethereum) SetupSuite() {
	x.kvdb = new(mocks.KVDB)
	x2eth := &x2ethereum{drivers.DriverBase{}}

	_, _, kvdb = util.CreateTestDB()
	x2eth.SetLocalDB(kvdb)
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)
	x2eth.SetAPI(api)
	sdb, _ = db.NewGoMemDB("x2EthereumTestDb", "test", 128)
	x2eth.SetStateDB(sdb)
	x2eth.SetExecutorType(types.LoadExecutorType(driverName))
	x2eth.SetEnv(10, 100, 1)
	x2eth.SetIsFree(false)
	x2eth.SetChild(x2eth)

	tx.Execer = []byte(types2.X2ethereumX)
	tx.To = address.ExecAddress(types2.X2ethereumX)
	tx.Nonce = 1
	tx.ChainID = chainTestCfg.GetChainID()
	tx.Sign(types.SECP256K1, privFrom)

	x.action = newAction(x2eth, tx, 0)
	x.x2eth = x2eth
	x.addrX2Eth = address.ExecAddress(driverName)

	x.Equal("x2ethereum", x.x2eth.GetName())

	x.accountSetup()
}

func (x *suiteX2Ethereum) Test_1_SetConsensus() {
	receipt, err := x.action.procMsgSetConsensusThreshold(&types2.MsgConsensusThreshold{ConsensusThreshold: 80})
	x.NoError(err)
	x.NotEmpty(receipt)
	x.setDb(receipt)

	msg, err := x.x2eth.Query_GetConsensusThreshold(&types2.QueryConsensusThresholdParams{})
	x.NoError(err)

	reply := msg.(*types2.ReceiptQueryConsensusThreshold)
	x.Equal(reply.ConsensusThreshold, int64(80))
}

func (x *suiteX2Ethereum) Test_2_AddValidator() {
	add := &types2.MsgValidator{
		Address: addValidator1,
		Power:   7,
	}

	receipt, err := x.action.procAddValidator(add)
	x.NoError(err)
	x.NotEmpty(receipt)
	x.setDb(receipt)

	_, err = x.action.procAddValidator(add)
	x.Error(err)

	add2 := &types2.MsgValidator{
		Address: addValidator2,
		Power:   6,
	}

	receipt, err = x.action.procAddValidator(add2)
	x.NoError(err)
	x.NotEmpty(receipt)
	x.setDb(receipt)

	msg, err := x.x2eth.Query_GetTotalPower(&types2.QueryTotalPowerParams{})
	x.NoError(err)
	reply := msg.(*types2.ReceiptQueryTotalPower)
	x.Equal(reply.TotalPower, int64(13))

	msg, err = x.x2eth.Query_GetValidators(&types2.QueryValidatorsParams{})
	x.NoError(err)
	reply2 := msg.(*types2.ReceiptQueryValidator)
	x.Equal(reply2.TotalPower, int64(13))
}

func (x *suiteX2Ethereum) Test_3_ModifyAndRemoveValidator() {
	add := &types2.MsgValidator{
		Address: chain33Receiver,
		Power:   7,
	}

	receipt, err := x.action.procAddValidator(add)
	x.NoError(err)
	x.NotEmpty(receipt)
	x.setDb(receipt)

	add.Power = 8
	receipt, err = x.action.procModifyValidator(add)
	x.NoError(err)
	x.NotEmpty(receipt)
	x.setDb(receipt)

	msg, err := x.x2eth.Query_GetValidators(&types2.QueryValidatorsParams{Validator: chain33Receiver})
	x.NoError(err)
	reply := msg.(*types2.ReceiptQueryValidator)
	x.Equal(reply.Validators[0].Power, int64(8))

	receipt, err = x.action.procRemoveValidator(add)
	x.NoError(err)
	x.NotEmpty(receipt)
	x.setDb(receipt)

	_, err = x.x2eth.Query_GetValidators(&types2.QueryValidatorsParams{Validator: chain33Receiver})
	x.Equal(err, types2.ErrInvalidValidator)
}

func (x *suiteX2Ethereum) Test_4_Eth2Chain33() {
	_, err := x.x2eth.Query_GetTotalPower(&types2.QueryTotalPowerParams{})
	if err == types.ErrNotFound {
		x.Test_2_AddValidator()
	}

	payload := &types2.Eth2Chain33{
		EthereumChainID:       0,
		BridgeContractAddress: bridgeContractAddress,
		Nonce:                 0,
		IssuerDotSymbol:       symbol,
		TokenContractAddress:  tokenContractAddress,
		EthereumSender:        ethereumAddr,
		Chain33Receiver:       addValidator1,
		ValidatorAddress:      addValidator1,
		Amount:                "10",
		ClaimType:             int64(types2.LockClaimType),
	}

	receipt, err := x.action.procEth2Chain33_lock(payload)
	x.NoError(err)
	x.setDb(receipt)

	payload.ValidatorAddress = addValidator2
	receipt, err = x.action.procEth2Chain33_lock(payload)
	x.NoError(err)
	x.setDb(receipt)

	_, err = x.x2eth.Query_GetEthProphecy(&types2.QueryEthProphecyParams{ID: "010x7B95B6EC7EbD73572298cEf32Bb54FA408207359"})
	x.Equal(err, types.ErrNotFound)

	ID := strconv.Itoa(0) + strconv.Itoa(0) + ethereumAddr + tokenContractAddress + "lock"
	x.queryGetEthProphecy(ID, types2.EthBridgeStatus_SuccessStatusText)
	x.queryGetSymbolTotalAmountByTxType(1, symbol, "lock", "", "10")

	payload1 := &types2.Chain33ToEth{
		TokenContract:    tokenContractAddress,
		Chain33Sender:    addValidator1,
		EthereumReceiver: ethereumAddr,
		Amount:           "3",
		IssuerDotSymbol:  "eth",
		Decimals:         8,
	}

	receipt, err = x.action.procChain33ToEth_burn(payload1)
	x.NoError(err)
	x.setDb(receipt)

	x.queryGetEthProphecy("000x7B95B6EC7EbD73572298cEf32Bb54FA4082073590x0000000000000000000000000000000000000000lock", types2.EthBridgeStatus_SuccessStatusText)
	x.queryGetSymbolTotalAmount(symbol, "", 7, true)
	x.queryGetSymbolTotalAmountByTxType(2, symbol, "withdraw", "", "3")

	payload.Amount = "1"
	payload.Nonce = 1
	payload.ClaimType = int64(types2.LockClaimType)
	payload.ValidatorAddress = addValidator1
	receipt, err = x.action.procEth2Chain33_lock(payload)
	x.NoError(err)
	x.setDb(receipt)

	payload.ValidatorAddress = addValidator2
	receipt, err = x.action.procEth2Chain33_lock(payload)
	x.NoError(err)
	x.setDb(receipt)

	x.queryGetEthProphecy("010x7B95B6EC7EbD73572298cEf32Bb54FA4082073590x0000000000000000000000000000000000000000lock", types2.EthBridgeStatus_SuccessStatusText)
	x.queryGetSymbolTotalAmountByTxType(1, symbol, "lock", "", "11")
}

func (x *suiteX2Ethereum) Test_5_Chain33ToEth() {
	_, err := x.x2eth.Query_GetTotalPower(&types2.QueryTotalPowerParams{})
	if err == types.ErrNotFound {
		x.Test_2_AddValidator()
	}
	msgLock := &types2.Chain33ToEth{
		TokenContract:    tokenContractAddress,
		Chain33Sender:    addValidator1,
		EthereumReceiver: ethereumAddr,
		Amount:           "5",
		IssuerDotSymbol:  "coins.bty",
	}

	receipt, err := x.action.procChain33ToEth_lock(msgLock)
	x.NoError(err)
	x.setDb(receipt)

	x.queryGetSymbolTotalAmount("coins.bty", tokenContractAddress, 5, false)
	x.queryGetSymbolTotalAmountByTxType(2, "coins.bty", "lock", tokenContractAddress, "5")

	payload := &types2.Eth2Chain33{
		EthereumChainID:       0,
		BridgeContractAddress: bridgeContractAddress,
		Nonce:                 2,
		IssuerDotSymbol:       "coins.bty",
		TokenContractAddress:  tokenContractAddress,
		EthereumSender:        ethereumAddr,
		Chain33Receiver:       addValidator1,
		ValidatorAddress:      addValidator1,
		Amount:                "4",
		ClaimType:             int64(types2.BurnClaimType),
	}
	receipt1, err := x.action.procEth2Chain33_burn(payload)
	x.NoError(err)
	x.setDb(receipt1)

	payload.ValidatorAddress = addValidator2
	receipt1, err = x.action.procEth2Chain33_burn(payload)
	x.NoError(err)
	x.setDb(receipt1)

	x.queryGetSymbolTotalAmount("coins.bty", tokenContractAddress, 1, false)
	x.queryGetSymbolTotalAmountByTxType(1, "coins.bty", "withdraw", tokenContractAddress, "4")
	x.queryGetEthProphecy("020x7B95B6EC7EbD73572298cEf32Bb54FA4082073590x0000000000000000000000000000000000000000burn", types2.EthBridgeStatus_SuccessStatusText)
}

func (x *suiteX2Ethereum) accountSetup() {
	acc := x.x2eth.GetCoinsAccount()

	account := &types.Account{
		Balance: 1000 * 1e8,
		Addr:    addValidator1,
	}
	acc.SaveAccount(account)
	account = acc.LoadAccount(addValidator1)
	x.Equal(int64(1000*1e8), account.Balance)
	_, err := acc.TransferToExec(addValidator1, x.addrX2Eth, 200*1e8)
	x.Nil(err)
	account = acc.LoadExecAccount(addValidator1, x.addrX2Eth)
	x.Equal(int64(200*1e8), account.Balance)
	account = &types.Account{
		Balance: 1000 * 1e8,
		Addr:    addValidator2,
	}
	acc.SaveAccount(account)
	account = acc.LoadAccount(addValidator2)
	x.Equal(int64(1000*1e8), account.Balance)
	_, err = acc.TransferToExec(addValidator2, x.addrX2Eth, 200*1e8)
	x.Nil(err)
	account = acc.LoadExecAccount(addValidator2, x.addrX2Eth)
	x.Equal(int64(200*1e8), account.Balance)
}

func (x *suiteX2Ethereum) setDb(receipt *chain33types.Receipt) {
	for _, kv := range receipt.KV {
		_ = sdb.Set(kv.Key, kv.Value)
	}

	receiptDate := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := x.x2eth.execLocal(receiptDate)
	x.NoError(err)

	for _, kv := range set.KV {
		_ = kvdb.Set(kv.Key, kv.Value)
	}
}

func (x *suiteX2Ethereum) queryGetSymbolTotalAmountByTxType(direction int64, tokenSymbol, txType, tokenAddr, equal string) {
	params := &types2.QuerySymbolAssetsByTxTypeParams{
		TokenSymbol: tokenSymbol,
		Direction:   direction,
		TxType:      txType,
		TokenAddr:   tokenAddr,
	}
	msg, err := x.x2eth.Query_GetSymbolTotalAmountByTxType(params)
	x.NoError(err)

	fmt.Println("***", msg, err)

	symbolAmount := msg.(*types2.ReceiptQuerySymbolAssets)
	x.Equal(symbolAmount.Res[0].TotalAmount, equal)
}

func (x *suiteX2Ethereum) queryGetSymbolTotalAmount(tokenSymbol, tokenAddress string, equal int64, eth2chain33 bool) {
	var lock, burn *types2.ReceiptQuerySymbolAssets
	if eth2chain33 {
		msg, _ := x.x2eth.Query_GetSymbolTotalAmountByTxType(&types2.QuerySymbolAssetsByTxTypeParams{TokenSymbol: tokenSymbol, TokenAddr: tokenAddress, Direction: 1, TxType: types2.LockClaim})
		lock = msg.(*types2.ReceiptQuerySymbolAssets)

		msg, _ = x.x2eth.Query_GetSymbolTotalAmountByTxType(&types2.QuerySymbolAssetsByTxTypeParams{TokenSymbol: tokenSymbol, TokenAddr: tokenAddress, Direction: 2, TxType: "withdraw"})
		burn = msg.(*types2.ReceiptQuerySymbolAssets)
	} else {
		msg, _ := x.x2eth.Query_GetSymbolTotalAmountByTxType(&types2.QuerySymbolAssetsByTxTypeParams{TokenSymbol: tokenSymbol, TokenAddr: tokenAddress, Direction: 2, TxType: types2.LockClaim})
		lock = msg.(*types2.ReceiptQuerySymbolAssets)

		msg, err := x.x2eth.Query_GetSymbolTotalAmountByTxType(&types2.QuerySymbolAssetsByTxTypeParams{TokenSymbol: tokenSymbol, TokenAddr: tokenAddress, Direction: 1, TxType: "withdraw"})
		if err != nil {
			burn = new(types2.ReceiptQuerySymbolAssets)
		} else {
			burn = msg.(*types2.ReceiptQuerySymbolAssets)
		}

	}

	var lockAmount, burnAmount float64
	if lock.Res != nil {
		lockAmount, _ = strconv.ParseFloat(lock.Res[0].TotalAmount, 64)
	} else {
		lockAmount = 0
	}

	if burn.Res != nil {
		burnAmount, _ = strconv.ParseFloat(burn.Res[0].TotalAmount, 64)
	} else {
		burnAmount = 0
	}

	x.Equal(math.Abs(lockAmount-burnAmount), float64(equal))
}

func (x *suiteX2Ethereum) queryGetEthProphecy(id string, statusTest types2.EthBridgeStatus) {
	msg, err := x.x2eth.Query_GetEthProphecy(&types2.QueryEthProphecyParams{ID: id})
	x.NoError(err)
	reply := msg.(*types2.ReceiptEthProphecy)
	x.Equal(reply.Status.Text, statusTest)
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
