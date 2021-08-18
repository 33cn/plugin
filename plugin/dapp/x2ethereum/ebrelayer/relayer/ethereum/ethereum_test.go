package ethereum

//
//import (
//	"context"
//	"encoding/hex"
//	"flag"
//	"fmt"
//	"math/big"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/33cn/chain33/client/mocks"
//	dbm "github.com/33cn/chain33/common/db"
//	_ "github.com/33cn/chain33/system"
//	chain33Types "github.com/33cn/chain33/types"
//	"github.com/33cn/chain33/util/testnode"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/test/setup"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethinterface"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
//	ebTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
//	relayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
//	tml "github.com/BurntSushi/toml"
//	"github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/common/hexutil"
//	"github.com/ethereum/go-ethereum/crypto"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//	"github.com/stretchr/testify/require"
//)
//
//var (
//	configPath           = flag.String("f", "./../../relayer.toml", "configfile")
//	chain33PrivateKeyStr = "0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
//	chain33AccountAddr   = "1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"
//	passphrase           = "123456hzj"
//	chainTestCfg         = chain33Types.NewChain33Config(chain33Types.GetDefaultCfgstring())
//
//	// 0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a
//	deployerPrivateKey = "8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
//	// 0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f
//	ethValidatorAddrKeyA = "3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
//	ethValidatorAddrKeyB = "a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
//	ethValidatorAddrKeyC = "bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
//	ethValidatorAddrKeyD = "c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"
//)
//
//func Test_LockAndBurn(t *testing.T) {
//	var tx chain33Types.Transaction
//	var ret chain33Types.Reply
//	ret.IsOk = true
//
//	mockapi := &mocks.QueueProtocolAPI{}
//	// 这里对需要mock的方法打桩,Close是必须的，其它方法根据需要
//	mockapi.On("Close").Return()
//	mockapi.On("AddPushSubscribe", mock.Anything).Return(&ret, nil)
//	mockapi.On("CreateTransaction", mock.Anything).Return(&tx, nil)
//	mockapi.On("SendTx", mock.Anything).Return(&ret, nil)
//	mockapi.On("SendTransaction", mock.Anything).Return(&ret, nil)
//	mockapi.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)
//
//	mock33 := testnode.New("", mockapi)
//	defer mock33.Close()
//	rpcCfg := mock33.GetCfg().RPC
//	// 这里必须设置监听端口，默认的是无效值
//	rpcCfg.JrpcBindAddr = "127.0.0.1:8801"
//	mock33.GetRPC().Listen()
//
//	fmt.Println("======================= testLockEth =======================")
//	testLockEth(t)
//	fmt.Println("======================= testCreateERC20Token =======================")
//	testCreateERC20Token(t)
//	fmt.Println("======================= testBurnBty =======================")
//	testBurnBty(t)
//}
//
//func Test_GetValidatorAddr(t *testing.T) {
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	_, _, err = ethRelayer.NewAccount("123")
//	require.Nil(t, err)
//
//	privateKey, _, err := ethRelayer.GetAccount("123")
//	require.Nil(t, err)
//	assert.NotEqual(t, privateKey, chain33PrivateKeyStr)
//
//	privateKey, addr, err := ethRelayer.GetAccount(passphrase)
//	require.Nil(t, err)
//	assert.Equal(t, privateKey, chain33PrivateKeyStr)
//	assert.Equal(t, addr, chain33AccountAddr)
//
//	validators, err := ethRelayer.GetValidatorAddr()
//	require.Nil(t, err)
//	assert.Equal(t, validators.Chain33Validator, chain33AccountAddr)
//}
//
//func Test_IsValidatorActive(t *testing.T) {
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	is, err := ethRelayer.IsValidatorActive(para.InitValidators[0].String())
//	assert.Equal(t, is, true)
//	require.Nil(t, err)
//
//	is, err = ethRelayer.IsValidatorActive("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF")
//	assert.Equal(t, is, false)
//	require.Nil(t, err)
//
//	_, err = ethRelayer.IsValidatorActive("123")
//	require.Error(t, err)
//}
//
//func Test_ShowAddr(t *testing.T) {
//	{
//		cfg := initCfg(*configPath)
//		relayer := &Relayer4Ethereum{
//			provider:            cfg.EthProvider,
//			unlockchan:          make(chan int, 2),
//			rpcURL2Chain33:      cfg.SyncTxConfig.Chain33Host,
//			maturityDegree:      cfg.EthMaturityDegree,
//			fetchHeightPeriodMs: cfg.EthBlockFetchPeriod,
//		}
//		_, err := relayer.ShowBridgeBankAddr()
//		require.Error(t, err)
//
//		_, err = relayer.ShowBridgeRegistryAddr()
//		require.Error(t, err)
//	}
//
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	ethRelayer.prePareSubscribeEvent()
//
//	addr, err := ethRelayer.ShowBridgeBankAddr()
//	require.Nil(t, err)
//	assert.Equal(t, addr, x2EthDeployInfo.BridgeBank.Address.String())
//
//	addr, err = ethRelayer.ShowBridgeRegistryAddr()
//	require.Nil(t, err)
//	assert.Equal(t, addr, x2EthDeployInfo.BridgeRegistry.Address.String())
//
//	addr, err = ethRelayer.ShowOperator()
//	require.Nil(t, err)
//	assert.Equal(t, addr, para.Operator.String())
//}
//
//func Test_DeployContrcts(t *testing.T) {
//	_, sim, _, _ := deployContracts()
//	cfg := initCfg(*configPath)
//	cfg.SyncTxConfig.Dbdriver = "memdb"
//
//	db := dbm.NewDB("relayer_db_service", cfg.SyncTxConfig.Dbdriver, cfg.SyncTxConfig.DbPath, cfg.SyncTxConfig.DbCache)
//
//	relayer := &Relayer4Ethereum{
//		provider:            cfg.EthProvider,
//		db:                  db,
//		unlockchan:          make(chan int, 2),
//		rpcURL2Chain33:      cfg.SyncTxConfig.Chain33Host,
//		maturityDegree:      cfg.EthMaturityDegree,
//		fetchHeightPeriodMs: cfg.EthBlockFetchPeriod,
//		deployInfo:          cfg.Deploy,
//	}
//	relayer.clientSpec = sim
//	relayer.clientChainID = big.NewInt(1)
//
//	deployPrivateKey, _ := crypto.ToECDSA(common.FromHex(relayer.deployInfo.DeployerPrivateKey))
//	deployerAddr := crypto.PubkeyToAddress(deployPrivateKey.PublicKey)
//	relayer.operatorInfo = &ethtxs.OperatorInfo{
//		PrivateKey: deployPrivateKey,
//		Address:    deployerAddr,
//	}
//
//	_, err := relayer.DeployContrcts()
//	require.NoError(t, err)
//}
//
//func Test_SetBridgeRegistryAddr(t *testing.T) {
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	_ = ethRelayer.setBridgeRegistryAddr(x2EthDeployInfo.BridgeRegistry.Address.String())
//	registrAddrInDB, err := ethRelayer.getBridgeRegistryAddr()
//	require.Nil(t, err)
//	assert.Equal(t, registrAddrInDB, x2EthDeployInfo.BridgeRegistry.Address.String())
//}
//
//func Test_CreateBridgeToken(t *testing.T) {
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	balance, err := ethRelayer.GetBalance("", para.InitValidators[0].String())
//	require.Nil(t, err)
//	assert.Equal(t, balance, "10000000000")
//
//	tokenAddrbty, err := ethRelayer.CreateBridgeToken("BTY")
//	require.Nil(t, err)
//	require.NotEmpty(t, tokenAddrbty)
//	sim.Commit()
//
//	addr, err := ethRelayer.ShowTokenAddrBySymbol("BTY")
//	require.Nil(t, err)
//	assert.Equal(t, addr, tokenAddrbty)
//
//	decimals, err := ethRelayer.GetDecimals(tokenAddrbty)
//	require.Nil(t, err)
//	assert.Equal(t, decimals, uint8(8))
//
//	_, err = ethRelayer.Burn(para.InitValidators[0].String(), tokenAddrbty, chain33AccountAddr, "10")
//	require.Error(t, err)
//
//	_, err = ethRelayer.BurnAsync(para.InitValidators[0].String(), tokenAddrbty, chain33AccountAddr, "10")
//	require.Error(t, err)
//}
//
//func testLockEth(t *testing.T) {
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	ctx := context.Background()
//	bridgeBankBalance, err := sim.BalanceAt(ctx, x2EthDeployInfo.BridgeBank.Address, nil)
//	require.Nil(t, err)
//	assert.Equal(t, bridgeBankBalance.Int64(), int64(0))
//
//	userOneAuth, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	//lock 50 eth
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	ethAmount := big.NewInt(50)
//	userOneAuth.Value = ethAmount
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	bridgeBankBalance, err = sim.BalanceAt(ctx, x2EthDeployInfo.BridgeBank.Address, nil)
//	require.Nil(t, err)
//	assert.Equal(t, bridgeBankBalance.Int64(), ethAmount.Int64())
//
//	for i := 0; i < 11; i++ {
//		sim.Commit()
//	}
//
//	time.Sleep(time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	balance, err := ethRelayer.ShowLockStatics("")
//	require.Nil(t, err)
//	assert.Equal(t, balance, "50")
//
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//}
//
//func testCreateERC20Token(t *testing.T) {
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	tokenErc20Addr, err := ethRelayer.CreateERC20Token("testcc")
//	require.Nil(t, err)
//	require.NotEmpty(t, tokenErc20Addr)
//	sim.Commit()
//
//	_, err = ethRelayer.MintERC20Token(tokenErc20Addr, para.Deployer.String(), "20000000000000")
//	require.Nil(t, err)
//	sim.Commit()
//
//	balance, err := ethRelayer.ShowDepositStatics(tokenErc20Addr)
//	require.Nil(t, err)
//	assert.Equal(t, balance, "20000000000000")
//
//	claimID := crypto.Keccak256Hash(big.NewInt(50).Bytes())
//	bret, err := ethRelayer.IsProphecyPending(claimID)
//	require.Nil(t, err)
//	assert.Equal(t, bret, false)
//
//	txhash, err := ethRelayer.TransferToken(tokenErc20Addr, hexutil.Encode(crypto.FromECDSA(para.DeployPrivateKey)), ethRelayer.deployInfo.ValidatorsAddr[0], "100")
//	require.Nil(t, err)
//	sim.Commit()
//
//	_, err = ethRelayer.ShowTxReceipt(txhash)
//	require.Nil(t, err)
//
//	balance, err = ethRelayer.GetBalance(tokenErc20Addr, ethRelayer.deployInfo.ValidatorsAddr[0])
//	require.Nil(t, err)
//	assert.Equal(t, balance, "100")
//
//	balance, err = ethRelayer.GetBalance(tokenErc20Addr, para.Deployer.String())
//	require.Nil(t, err)
//	assert.Equal(t, balance, "19999999999900")
//
//	tx1 := ethRelayer.QueryTxhashRelay2Eth()
//	require.Empty(t, tx1)
//
//	tx2 := ethRelayer.QueryTxhashRelay2Chain33()
//	require.Empty(t, tx2)
//
//	tokenErc20Addrtestc, err := ethRelayer.CreateERC20Token("testc")
//	require.Nil(t, err)
//	require.NotEmpty(t, tokenErc20Addrtestc)
//	sim.Commit()
//
//	_, err = ethRelayer.MintERC20Token(tokenErc20Addrtestc, para.Deployer.String(), "10000000000000")
//	require.Nil(t, err)
//	sim.Commit()
//
//	balance, err = ethRelayer.ShowDepositStatics(tokenErc20Addrtestc)
//	require.Nil(t, err)
//	assert.Equal(t, balance, "10000000000000")
//
//	chain33Receiver := "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
//	_, err = ethRelayer.LockEthErc20Asset(hexutil.Encode(crypto.FromECDSA(para.DeployPrivateKey)), tokenErc20Addrtestc, "100", chain33Receiver)
//	require.Nil(t, err)
//	sim.Commit()
//
//	balance, err = ethRelayer.GetBalance(tokenErc20Addrtestc, para.Deployer.String())
//	require.Nil(t, err)
//	assert.Equal(t, balance, "9999999999900")
//
//	_, err = ethRelayer.ApproveAllowance(hexutil.Encode(crypto.FromECDSA(para.DeployPrivateKey)), tokenErc20Addrtestc, "500")
//	require.Nil(t, err)
//	sim.Commit()
//
//	_, err = ethRelayer.LockEthErc20AssetAsync(hexutil.Encode(crypto.FromECDSA(para.DeployPrivateKey)), tokenErc20Addrtestc, "100", chain33Receiver)
//	require.Nil(t, err)
//	sim.Commit()
//
//	balance, err = ethRelayer.GetBalance(tokenErc20Addrtestc, para.Deployer.String())
//	require.Nil(t, err)
//	assert.Equal(t, balance, "9999999999800")
//
//	for i := 0; i < 11; i++ {
//		sim.Commit()
//	}
//	time.Sleep(time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//}
//
//func testBurnBty(t *testing.T) {
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	tokenAddrbty, err := ethRelayer.CreateBridgeToken("bty")
//	require.Nil(t, err)
//	require.NotEmpty(t, tokenAddrbty)
//	sim.Commit()
//
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	amount := int64(100)
//	ethReceiver := para.InitValidators[2]
//	claimID := crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), big.NewInt(amount).Bytes())
//	authOracle, err := ethtxs.PrepareAuth(ethRelayer.clientSpec, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//	signature, err := ethtxs.SignClaim4Eth(claimID, para.ValidatorPriKey[0])
//	require.Nil(t, err)
//
//	_, err = x2EthContracts.Oracle.NewOracleClaim(
//		authOracle,
//		events.ClaimTypeLock,
//		chain33Sender,
//		ethReceiver,
//		common.HexToAddress(tokenAddrbty),
//		"bty",
//		big.NewInt(amount),
//		claimID,
//		signature)
//	require.Nil(t, err)
//	sim.Commit()
//
//	balanceNew, err := ethRelayer.GetBalance(tokenAddrbty, ethReceiver.String())
//	require.Nil(t, err)
//	require.Equal(t, balanceNew, "100")
//
//	_, err = ethRelayer.Burn(hexutil.Encode(crypto.FromECDSA(para.ValidatorPriKey[2])), tokenAddrbty, chain33AccountAddr, "10")
//	require.NoError(t, err)
//	sim.Commit()
//
//	balanceNew, err = ethRelayer.GetBalance(tokenAddrbty, ethReceiver.String())
//	require.Nil(t, err)
//	require.Equal(t, balanceNew, "90")
//
//	_, err = ethRelayer.ApproveAllowance(hexutil.Encode(crypto.FromECDSA(para.ValidatorPriKey[2])), tokenAddrbty, "10")
//	require.Nil(t, err)
//	sim.Commit()
//
//	_, err = ethRelayer.BurnAsync(hexutil.Encode(crypto.FromECDSA(para.ValidatorPriKey[2])), tokenAddrbty, chain33AccountAddr, "10")
//	require.NoError(t, err)
//	sim.Commit()
//
//	balanceNew, err = ethRelayer.GetBalance(tokenAddrbty, ethReceiver.String())
//	require.Nil(t, err)
//	require.Equal(t, balanceNew, "80")
//
//	fetchCnt := int32(10)
//	logs, err := ethRelayer.getNextValidEthTxEventLogs(ethRelayer.eventLogIndex.Height, ethRelayer.eventLogIndex.Index, fetchCnt)
//	require.NoError(t, err)
//
//	for _, vLog := range logs {
//		ethRelayer.procBridgeBankLogs(*vLog)
//	}
//
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//}
//
//func Test_RestorePrivateKeys(t *testing.T) {
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	ethRelayer := newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)
//	_ = ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
//	time.Sleep(4 * time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	go func() {
//		for range ethRelayer.unlockchan {
//		}
//	}()
//	ethRelayer.rwLock.RLock()
//	temp := ethRelayer.privateKey4Chain33
//	ethRelayer.rwLock.RUnlock()
//
//	err = ethRelayer.RestorePrivateKeys("123")
//	ethRelayer.rwLock.RLock()
//	assert.NotEqual(t, hex.EncodeToString(temp.Bytes()), hex.EncodeToString(ethRelayer.privateKey4Chain33.Bytes()))
//	ethRelayer.rwLock.RUnlock()
//	require.Nil(t, err)
//
//	err = ethRelayer.RestorePrivateKeys(passphrase)
//	ethRelayer.rwLock.RLock()
//	assert.Equal(t, hex.EncodeToString(temp.Bytes()), hex.EncodeToString(ethRelayer.privateKey4Chain33.Bytes()))
//	ethRelayer.rwLock.RUnlock()
//	require.Nil(t, err)
//
//	err = ethRelayer.StoreAccountWithNewPassphase("new123", passphrase)
//	require.Nil(t, err)
//
//	err = ethRelayer.RestorePrivateKeys("new123")
//	ethRelayer.rwLock.RLock()
//	assert.Equal(t, hex.EncodeToString(temp.Bytes()), hex.EncodeToString(ethRelayer.privateKey4Chain33.Bytes()))
//	ethRelayer.rwLock.RUnlock()
//	require.Nil(t, err)
//}
//
//func newEthRelayer(para *ethtxs.DeployPara, sim *ethinterface.SimExtend, x2EthContracts *ethtxs.X2EthContracts, x2EthDeployInfo *ethtxs.X2EthDeployInfo) *Relayer4Ethereum {
//	cfg := initCfg(*configPath)
//	cfg.SyncTxConfig.Chain33Host = "http://127.0.0.1:8801"
//	cfg.BridgeRegistry = x2EthDeployInfo.BridgeRegistry.Address.String()
//	cfg.SyncTxConfig.PushBind = "127.0.0.1:60000"
//	cfg.SyncTxConfig.FetchHeightPeriodMs = 50
//	cfg.SyncTxConfig.Dbdriver = "memdb"
//	cfg.SyncTxConfig.DbPath = "datadirEth"
//
//	db := dbm.NewDB("relayer_db_service", cfg.SyncTxConfig.Dbdriver, cfg.SyncTxConfig.DbPath, cfg.SyncTxConfig.DbCache)
//
//	relayer := &Relayer4Ethereum{
//		provider:            cfg.EthProvider,
//		db:                  db,
//		unlockchan:          make(chan int, 2),
//		rpcURL2Chain33:      cfg.SyncTxConfig.Chain33Host,
//		bridgeRegistryAddr:  x2EthDeployInfo.BridgeRegistry.Address,
//		maturityDegree:      cfg.EthMaturityDegree,
//		fetchHeightPeriodMs: cfg.EthBlockFetchPeriod,
//		//deployInfo:          cfg.Deploy,
//	}
//
//	relayer.deployInfo = &ebTypes.Deploy{}
//	relayer.deployInfo.DeployerPrivateKey = hexutil.Encode(crypto.FromECDSA(para.DeployPrivateKey))
//	relayer.deployInfo.OperatorAddr = para.Operator.String()
//	for _, v := range para.InitValidators {
//		relayer.deployInfo.ValidatorsAddr = append(relayer.deployInfo.ValidatorsAddr, v.String())
//	}
//	for _, v := range para.InitPowers {
//		relayer.deployInfo.InitPowers = append(relayer.deployInfo.InitPowers, v.Int64())
//	}
//
//	relayer.eventLogIndex = relayer.getLastBridgeBankProcessedHeight()
//	relayer.initBridgeBankTx()
//	relayer.clientSpec = sim
//	relayer.clientChainID = big.NewInt(1)
//
//	deployPrivateKey, _ := crypto.ToECDSA(common.FromHex(relayer.deployInfo.DeployerPrivateKey))
//	deployerAddr := crypto.PubkeyToAddress(deployPrivateKey.PublicKey)
//	relayer.rwLock.Lock()
//	relayer.operatorInfo = &ethtxs.OperatorInfo{
//		PrivateKey: deployPrivateKey,
//		Address:    deployerAddr,
//	}
//	relayer.deployPara = para
//	relayer.x2EthContracts = x2EthContracts
//	relayer.x2EthDeployInfo = x2EthDeployInfo
//	relayer.rwLock.Unlock()
//
//	go relayer.proc()
//	return relayer
//}
//
//func deployContracts() (*ethtxs.DeployPara, *ethinterface.SimExtend, *ethtxs.X2EthContracts, *ethtxs.X2EthDeployInfo) {
//	ethValidatorAddrKeys := make([]string, 0)
//	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyA)
//	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyB)
//	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyC)
//	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyD)
//
//	ctx := context.Background()
//	//var backend bind.ContractBackend
//	backend, para := setup.PrepareTestEnvironment(deployerPrivateKey, ethValidatorAddrKeys)
//	sim := new(ethinterface.SimExtend)
//	sim.SimulatedBackend = backend.(*backends.SimulatedBackend)
//
//	callMsg := ethereum.CallMsg{
//		From: para.Deployer,
//		Data: common.FromHex(generated.BridgeBankBin),
//	}
//
//	_, _ = sim.EstimateGas(ctx, callMsg)
//	x2EthContracts, x2EthDeployInfo, _ := ethtxs.DeployAndInit(sim, para)
//	sim.Commit()
//
//	return para, sim, x2EthContracts, x2EthDeployInfo
//}
//
//func initCfg(path string) *relayerTypes.RelayerConfig {
//	var cfg relayerTypes.RelayerConfig
//	if _, err := tml.DecodeFile(path, &cfg); err != nil {
//		//fmt.Println(err)
//		os.Exit(-1)
//	}
//	return &cfg
//}
