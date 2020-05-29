package ethereum

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/test/setup"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	tml "github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
)

var (
	configPath           = flag.String("f", "./../../relayer.toml", "configfile")
	chain33PrivateKeyStr = "0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
	chain33AccountAddr   = "1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"

	passphrase = "123456hzj"
)

type suiteEthRelayer struct {
	suite.Suite
	ethRelayer      *Relayer4Ethereum
	sim             *backends.SimulatedBackend
	x2EthContracts  *ethtxs.X2EthContracts
	x2EthDeployInfo *ethtxs.X2EthDeployInfo
	para            *ethtxs.DeployPara
}

func TestRunSuiteX2Ethereum(t *testing.T) {
	log := new(suiteEthRelayer)
	suite.Run(t, log)
}

func (r *suiteEthRelayer) SetupSuite() {
	r.deployContracts()
	r.ethRelayer = r.newEthRelayer()
}

func (r *suiteEthRelayer) Test_1_ImportPrivateKey() {
	validators, err := r.ethRelayer.GetValidatorAddr()
	r.Error(err)
	r.Empty(validators)

	_, _, err = r.ethRelayer.NewAccount("123")
	r.NoError(err)

	err = r.ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
	r.NoError(err)

	privateKey, addr, err := r.ethRelayer.GetAccount("123")
	r.NoError(err)
	r.NotEqual(privateKey, chain33PrivateKeyStr)

	privateKey, addr, err = r.ethRelayer.GetAccount(passphrase)
	r.NoError(err)
	r.Equal(privateKey, chain33PrivateKeyStr)
	r.Equal(addr, chain33AccountAddr)

	validators, err = r.ethRelayer.GetValidatorAddr()
	r.NoError(err)
	r.Equal(validators.Chain33Validator, chain33AccountAddr)

	time.Sleep(5 * time.Second)
}

func (r *suiteEthRelayer) Test_2_RestorePrivateKeys() {
	// 错误的密码 也不报错
	err := r.ethRelayer.RestorePrivateKeys(passphrase)
	r.NoError(err)

	err = r.ethRelayer.StoreAccountWithNewPassphase(passphrase, passphrase)
	r.NoError(err)
}

func (r *suiteEthRelayer) Test_3_IsValidatorActive() {
	is, err := r.ethRelayer.IsValidatorActive("0x92c8b16afd6d423652559c6e266cbe1c29bfd84f")
	r.Equal(is, true)
	r.NoError(err)

	is, err = r.ethRelayer.IsValidatorActive("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF")
	r.Equal(is, false)
	r.NoError(err)

	/*
		re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
			if !re.MatchString(addr) {
				return false, errors.New("this address is not an ethereum address")
			}
	*/

	is, err = r.ethRelayer.IsValidatorActive("123")
	r.Error(err)
}

func (r *suiteEthRelayer) Test_Relayer4Ethereum_GetAccount() {

}

//func (r *suiteEthRelayer) TestRelayer4Ethereum_ApproveAllowance(t *testing.T) {
//	r.ethRelayer.ApproveAllowance()
//}

//
//func TestEthRelayerNewRelayerManager(t *testing.T) {
//	ctx := context.Background()
//	println("TEST:BridgeToken creation (Chain33 assets)")
//	//1st部署相关合约
//	backend, para := setup.PrepareTestEnv()
//	sim := backend.(*backends.SimulatedBackend)
//
//	balance, _ := sim.BalanceAt(ctx, para.Deployer, nil)
//	fmt.Println("deployer addr,", para.Deployer.String(), "balance =", balance.String())
//
//	/////////////////////////EstimateGas///////////////////////////
//	callMsg := ethereum.CallMsg{
//		From: para.Deployer,
//		Data: common.FromHex(generated.BridgeBankBin),
//	}
//
//	gas, err := sim.EstimateGas(ctx, callMsg)
//	if nil != err {
//		panic("failed to estimate gas due to:" + err.Error())
//	}
//	fmt.Printf("\nThe estimated gas=%d", gas)
//	////////////////////////////////////////////////////
//
//	x2EthContracts, x2EthDeployInfo, err := ethtxs.DeployAndInit(backend, para)
//	if nil != err {
//		t.Fatalf("DeployAndInit failed due to:%s", err.Error())
//	}
//	sim.Commit()
//	fmt.Println("x2EthDeployInfo.BridgeBank.Address is:", x2EthDeployInfo.BridgeBank.Address.String(), x2EthContracts.BridgeBank)
//	fmt.Println("x2EthDeployInfo.BridgeRegistry.Address is:", x2EthDeployInfo.BridgeRegistry.Address.String())
//	///////////////////////
//
//	//defer func() {
//	fmt.Println("defer remove datadir")
//	err4 := os.RemoveAll("./datadir")
//	if err4 != nil {
//		fmt.Println(err4)
//	}
//	//}()
//
//	var ret = types.ReplySubTxReceipt{IsOk: true}
//	var he = types.Header{Height: 10000}
//
//	mockapi := &mocks.QueueProtocolAPI{}
//	// 这里对需要mock的方法打桩,Close是必须的，其它方法根据需要
//	mockapi.On("Close").Return()
//	mockapi.On("AddSubscribeTxReceipt", mock.Anything).Return(&ret, nil)
//	mockapi.On("GetLastHeader", mock.Anything).Return(&he, nil)
//	mockapi.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)
//
//	mock33 := testnode.New("", mockapi)
//	defer mock33.Close()
//	rpcCfg := mock33.GetCfg().RPC
//	// 这里必须设置监听端口，默认的是无效值
//	rpcCfg.JrpcBindAddr = "127.0.0.1:8801"
//	mock33.GetRPC().Listen()
//
//	if *configPath == "" {
//		*configPath = "./../relayer.toml"
//	}
//	cfg := initCfg(*configPath)
//	db := dbm.NewDB("relayer_db_service", cfg.SyncTxConfig.Dbdriver, cfg.SyncTxConfig.DbPath, cfg.SyncTxConfig.DbCache)
//	cfg.SyncTxConfig.Chain33Host = "http://127.0.0.1:8801"
//	cfg.SyncTxConfig.PushBind = "127.0.0.1:20000"
//	cfg.SyncTxConfig.FetchHeightPeriodMs = 50
//	cfg.BridgeRegistry = "" //x2EthDeployInfo.BridgeRegistry.Address.String()
//
//	//cfg.EthProvider = "http://127.0.0.1:1080/"
//	//_, err = ethtxs.SetupWebsocketEthClient(cfg.EthProvider)
//	//assert.NoError(t, err)
//	//return
//
//	ctx, cancel := context.WithCancel(context.Background())
//
//	chain33RelayerService := ethRelayer.StartChain33Relayer(ctx, cfg.SyncTxConfig, cfg.BridgeRegistry, cfg.EthProvider, db)
//	ethRelayerService := ethRelayer.StartEthereumRelayer(cfg.SyncTxConfig.Chain33Host, db, cfg.EthProvider, cfg.BridgeRegistry, cfg.Deploy, cfg.EthMaturityDegree, cfg.EthBlockFetchPeriod)
//	//ethRelayer.SetClient(ethRelayerService, sim)
//	relayerManager := NewRelayerManager(chain33RelayerService, ethRelayerService, db)
//
//	var result interface{}
//
//	setPasswdReq := relayerTypes.ReqChangePasswd{
//		OldPassphase: "kk",
//		NewPassphase: "123456hzj",
//	}
//
//	err = relayerManager.SetPassphase(setPasswdReq, &result)
//	//assert.NoError(t, err)
//
//	err = relayerManager.Unlock("123456hzj", &result)
//	assert.NoError(t, err)
//	fmt.Println(result)
//
//	err = relayerManager.ImportChain33PrivateKey4EthRelayer("0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68", &result)
//	assert.NoError(t, err)
//
//	time.Sleep(1 * time.Second)
//	ethRelayer.Setx2EthContractsDeployInfo(ethRelayerService, x2EthContracts, x2EthDeployInfo)
//	fmt.Println("***************")
//	// do something
//	{
//		bridgeBankBalance, err := sim.BalanceAt(ctx, x2EthDeployInfo.BridgeBank.Address, nil)
//		require.Nil(t, err)
//		t.Logf("origin eth bridgeBankBalance is:%d", bridgeBankBalance.Int64())
//
//		userOneAuth, err := ethtxs.PrepareAuth(backend, para.ValidatorPriKey[0], para.InitValidators[0])
//		require.Nil(t, err)
//		ethAmount := big.NewInt(50)
//		userOneAuth.Value = ethAmount
//
//		fmt.Println("origin eth bridgeBankBalance is:", bridgeBankBalance.Int64())
//		//lock 50 eth
//		chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//		_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethAmount)
//		require.Nil(t, err)
//		sim.Commit()
//
//		bridgeBankBalance, err = sim.BalanceAt(ctx, x2EthDeployInfo.BridgeBank.Address, nil)
//		require.Nil(t, err)
//		require.Equal(t, bridgeBankBalance.Int64(), ethAmount.Int64())
//		t.Logf("eth bridgeBankBalance changes to:%d", bridgeBankBalance.Int64())
//		fmt.Println("eth bridgeBankBalance is:", bridgeBankBalance.Int64())
//	}
//
//	time.Sleep(5000 * time.Second)
//
//	//os.Exit(0)
//
//	var wg sync.WaitGroup
//
//	ch := make(chan os.Signal, 1)
//	signal.Notify(ch, syscall.SIGTERM)
//	go func() {
//		<-ch
//		cancel()
//		wg.Wait()
//		os.Exit(0)
//	}()
//}

func (r *suiteEthRelayer) newEthRelayer() *Relayer4Ethereum {
	cfg := initCfg(*configPath)
	cfg.SyncTxConfig.Chain33Host = "http://127.0.0.1:8801"
	cfg.BridgeRegistry = r.x2EthDeployInfo.BridgeRegistry.Address.String()
	cfg.SyncTxConfig.PushBind = "127.0.0.1:60000"
	cfg.SyncTxConfig.FetchHeightPeriodMs = 50
	cfg.SyncTxConfig.Dbdriver = "memdb"

	db := dbm.NewDB("relayer_db_service", cfg.SyncTxConfig.Dbdriver, cfg.SyncTxConfig.DbPath, cfg.SyncTxConfig.DbCache)

	relayer := &Relayer4Ethereum{
		provider:            cfg.EthProvider,
		db:                  db,
		unlockchan:          make(chan int, 2),
		rpcURL2Chain33:      cfg.SyncTxConfig.Chain33Host,
		bridgeRegistryAddr:  r.x2EthDeployInfo.BridgeRegistry.Address,
		deployInfo:          cfg.Deploy,
		maturityDegree:      cfg.EthMaturityDegree,
		fetchHeightPeriodMs: cfg.EthBlockFetchPeriod,
	}

	registrAddrInDB, err := relayer.getBridgeRegistryAddr()
	//如果输入的registry地址非空，且和数据库保存地址不一致，则直接使用输入注册地址
	if cfg.BridgeRegistry != "" && nil == err && registrAddrInDB != cfg.BridgeRegistry {
		relayerLog.Error("StartEthereumRelayer", "BridgeRegistry is setted already with value", registrAddrInDB,
			"but now setting to", cfg.BridgeRegistry)
		_ = relayer.setBridgeRegistryAddr(cfg.BridgeRegistry)
	} else if cfg.BridgeRegistry == "" && registrAddrInDB != "" {
		//输入地址为空，且数据库中保存地址不为空，则直接使用数据库中的地址
		relayer.bridgeRegistryAddr = common.HexToAddress(registrAddrInDB)
	}
	relayer.eventLogIndex = relayer.getLastBridgeBankProcessedHeight()
	relayer.initBridgeBankTx()

	go r.proc()

	//var wg sync.WaitGroup
	//ch := make(chan os.Signal, 1)
	//signal.Notify(ch, syscall.SIGTERM)
	//go func() {
	//	<-ch
	//	cancel()
	//	wg.Wait()
	//	os.Exit(0)
	//}()

	return relayer
}

func (r *suiteEthRelayer) proc() {
	r.ethRelayer.backend = r.sim
	r.ethRelayer.clientChainID = new(big.Int)

	//等待用户导入
	relayerLog.Info("Please unlock or import private key for Ethereum relayer")
	nilAddr := common.Address{}
	if nilAddr != r.ethRelayer.bridgeRegistryAddr {
		r.ethRelayer.x2EthContracts = r.x2EthContracts
		r.ethRelayer.x2EthDeployInfo = r.x2EthDeployInfo

		relayerLog.Info("^-^ ^-^ Succeed to recover corresponding solidity contract handler")
		if nil != r.ethRelayer.recoverDeployPara() {
			panic("Failed to recoverDeployPara")
		}
		r.ethRelayer.unlockchan <- start
	}

	ctx := context.Background()
	var timer *time.Ticker
	//var err error
	continueFailCount := int32(0)
	for range r.ethRelayer.unlockchan {
		relayerLog.Info("Received ethRelayer.unlockchan")
		if nil != r.ethRelayer.privateKey4Chain33 && nilAddr != r.ethRelayer.bridgeRegistryAddr {
			relayerLog.Info("Ethereum relayer starts to run...")
			r.ethRelayer.prePareSubscribeEvent()
			//向bridgeBank订阅事件
			r.ethRelayer.subscribeEvent()
			r.filterLogEvents()
			relayerLog.Info("Ethereum relayer starts to process online log event...")
			timer = time.NewTicker(time.Duration(r.ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
			goto latter
		}
	}

latter:
	for {
		select {
		case <-timer.C:
			r.ethRelayer.procNewHeight(ctx, &continueFailCount)
		case err := <-r.ethRelayer.bridgeBankSub.Err():
			panic("bridgeBankSub" + err.Error())
		case vLog := <-r.ethRelayer.bridgeBankLog:
			r.ethRelayer.storeBridgeBankLogs(vLog, true)
		}
	}
}

func Test3(t *testing.T) {
	i := 1
	j := 1
	for i > 0 {
		i++
		fmt.Println("i", i)
		time.Sleep(time.Second)
		goto aaa
	}

	fmt.Println("--00--")

aaa:
	for j > 0 {
		j++
		fmt.Println("j", j)
		time.Sleep(time.Second)
		if j > 3 {
			j = 1
			break
		}
	}

	fmt.Println("--00--")
}

func (r *suiteEthRelayer) deployContracts() {
	ctx := context.Background()
	var backend bind.ContractBackend
	backend, r.para = setup.PrepareTestEnvironment()
	r.sim = backend.(*backends.SimulatedBackend)

	balance, _ := r.sim.BalanceAt(ctx, r.para.Deployer, nil)
	fmt.Println("deployer addr,", r.para.Deployer.String(), "balance =", balance.String())

	/////////////////////////EstimateGas///////////////////////////
	callMsg := ethereum.CallMsg{
		From: r.para.Deployer,
		Data: common.FromHex(generated.BridgeBankBin),
	}

	gas, err := r.sim.EstimateGas(ctx, callMsg)
	r.NoError(err)
	fmt.Printf("\nThe estimated gas=%d\n", gas)
	////////////////////////////////////////////////////

	r.x2EthContracts, r.x2EthDeployInfo, err = ethtxs.DeployAndInit(backend, r.para)
	r.NoError(err)
	r.sim.Commit()
}

func (r *suiteEthRelayer) filterLogEvents() {
	deployHeight := int64(0)
	height4BridgeBankLogAt := int64(r.ethRelayer.getHeight4BridgeBankLogAt())

	if height4BridgeBankLogAt < deployHeight {
		height4BridgeBankLogAt = deployHeight
	}

	curHeight := int64(0)
	relayerLog.Info("filterLogEvents", "curHeight:", curHeight)

	bridgeBankSig := make(map[string]bool)
	bridgeBankSig[r.ethRelayer.bridgeBankEventLockSig] = true
	bridgeBankSig[r.ethRelayer.bridgeBankEventBurnSig] = true
	bridgeBankLog := make(chan types.Log)
	done := make(chan int)
	go r.ethRelayer.filterLogEventsProc(bridgeBankLog, done, "bridgeBank", curHeight, height4BridgeBankLogAt, r.ethRelayer.bridgeBankAddr, bridgeBankSig)

	for {
		select {
		case vLog := <-bridgeBankLog:
			r.ethRelayer.storeBridgeBankLogs(vLog, true)
		case vLog := <-r.ethRelayer.bridgeBankLog:
			//因为此处是同步保存信息，防止未同步完成出现panic时，直接将其设置为最新高度，中间出现部分信息不同步的情况
			r.ethRelayer.storeBridgeBankLogs(vLog, false)
		case <-done:
			relayerLog.Info("Finshed offline logs processed")
			return
		}
	}
}

func initCfg(path string) *relayerTypes.RelayerConfig {
	var cfg relayerTypes.RelayerConfig
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	return &cfg
}
