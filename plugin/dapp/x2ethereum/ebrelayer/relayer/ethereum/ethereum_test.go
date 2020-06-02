package ethereum

import (
	"context"
	"flag"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/33cn/chain33/client/mocks"
	dbm "github.com/33cn/chain33/common/db"
	_ "github.com/33cn/chain33/system"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/test/setup"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	tml "github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	configPath           = flag.String("f", "./../../relayer.toml", "configfile")
	chain33PrivateKeyStr = "0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
	chain33AccountAddr   = "1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"
	passphrase           = "123456hzj"
	testEthKey, _        = crypto.HexToECDSA("8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230")
	testEthAddr          = crypto.PubkeyToAddress(testEthKey.PublicKey)
	testBalance          = big.NewInt(2e18)
	chainTestCfg         = chain33Types.NewChain33Config(chain33Types.GetDefaultCfgstring())
)

type suiteEthRelayer struct {
	suite.Suite
	ethRelayer      *Relayer4Ethereum
	sim             *backends.SimulatedBackend
	x2EthContracts  *ethtxs.X2EthContracts
	x2EthDeployInfo *ethtxs.X2EthDeployInfo
	para            *ethtxs.DeployPara
	backend         bind.ContractBackend
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
}

func (r *suiteEthRelayer) Test_3_RestorePrivateKeys() {
	// 错误的密码 也不报错
	err := r.ethRelayer.RestorePrivateKeys(passphrase)
	r.NoError(err)

	err = r.ethRelayer.StoreAccountWithNewPassphase(passphrase, passphrase)
	r.NoError(err)

	time.Sleep(1 * time.Second)
}

func (r *suiteEthRelayer) Test_2_LockEth() {
	ctx := context.Background()
	bridgeBankBalance, err := r.sim.BalanceAt(ctx, r.x2EthDeployInfo.BridgeBank.Address, nil)
	r.NoError(err)
	r.Equal(bridgeBankBalance.Int64(), int64(0))

	userOneAuth, err := ethtxs.PrepareAuth(r.backend, r.para.ValidatorPriKey[0], r.para.InitValidators[0])
	r.NoError(err)

	//lock 50 eth
	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
	ethAmount := big.NewInt(50)
	userOneAuth.Value = ethAmount
	_, err = r.x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethAmount)
	r.NoError(err)
	r.sim.Commit()

	bridgeBankBalance, err = r.sim.BalanceAt(ctx, r.x2EthDeployInfo.BridgeBank.Address, nil)
	r.NoError(err)
	r.Equal(bridgeBankBalance.Int64(), ethAmount.Int64())

	time.Sleep(time.Duration(r.ethRelayer.fetchHeightPeriodMs) * time.Millisecond)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{r.ethRelayer.bridgeBankAddr},
	}
	logs, err := r.sim.FilterLogs(context.Background(), query)
	r.NoError(err)

	for _, logv := range logs {
		//err := r.ethRelayer.setEthTxEvent(logv)
		//r.NoError(err)
		r.ethRelayer.storeBridgeBankLogs(logv, true)
	}

	time.Sleep(time.Duration(r.ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
	//time.Sleep(time.Second * 5)
}

func (r *suiteEthRelayer) Test_4_handleLogLockEvent() {
	var tx chain33Types.Transaction
	var ret chain33Types.Reply
	ret.IsOk = true

	mockapi := &mocks.QueueProtocolAPI{}
	// 这里对需要mock的方法打桩,Close是必须的，其它方法根据需要
	mockapi.On("Close").Return()
	mockapi.On("CreateTransaction", mock.Anything).Return(&tx, nil)
	mockapi.On("SendTx", mock.Anything).Return(&ret, nil)
	mockapi.On("SendTransaction", mock.Anything).Return(&ret, nil)
	mockapi.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)

	mock33 := testnode.New("", mockapi)
	defer mock33.Close()
	rpcCfg := mock33.GetCfg().RPC
	// 这里必须设置监听端口，默认的是无效值
	rpcCfg.JrpcBindAddr = "127.0.0.1:8801"
	mock33.GetRPC().Listen()

	query := ethereum.FilterQuery{
		Addresses: []common.Address{r.ethRelayer.bridgeBankAddr},
	}
	logs, err := r.sim.FilterLogs(context.Background(), query)
	r.NoError(err)

	for _, logv := range logs {
		eventName := events.LogLock.String()
		err := r.ethRelayer.handleLogLockEvent(r.ethRelayer.clientChainID, r.ethRelayer.bridgeBankAbi, eventName, logv)
		r.NoError(err)
	}
}

func (r *suiteEthRelayer) Test_5_Show() {

	addr, err := r.ethRelayer.ShowBridgeBankAddr()
	r.NoError(err)
	r.Equal(addr, r.x2EthDeployInfo.BridgeBank.Address.String())

	addr, err = r.ethRelayer.ShowBridgeRegistryAddr()
	r.NoError(err)
	r.Equal(addr, r.x2EthDeployInfo.BridgeRegistry.Address.String())

	balance, err := r.ethRelayer.GetBalance("", testEthAddr.String())
	r.NoError(err)
	r.Equal(balance, "2000000000000000000")

	_, err = r.ethRelayer.GetBalance("0x0000000000000000000000000000000000000000", testEthAddr.String())
	r.Error(err)

	balance, err = r.ethRelayer.ShowLockStatics("")
	r.NoError(err)
	r.Equal(balance, "50")

	_, err = r.ethRelayer.ShowDepositStatics("")
	r.Error(err)

	_, err = r.ethRelayer.ShowTokenAddrBySymbol("bty")
	r.Error(err)

	claimID := crypto.Keccak256Hash(big.NewInt(50).Bytes())
	ret, err := r.ethRelayer.IsProphecyPending(claimID)
	r.NoError(err)
	r.Equal(ret, false)

	//addr, err = r.ethRelayer.CreateBridgeToken("bty")
	//r.NoError(err)

	//tokenAddr, err := r.ethRelayer.CreateERC20Token("testc")
	//r.Error(err)

	//addr, err = r.ethRelayer.MintERC20Token(tokenAddr, r.ethRelayer.ethValidator.String(), "20000000000000")
	//r.Error(err)

	_, err = r.ethRelayer.ShowOperator()
	r.Error(err)

	tx1 := r.ethRelayer.QueryTxhashRelay2Eth()
	r.Empty(tx1)

	tx2 := r.ethRelayer.QueryTxhashRelay2Chain33()
	r.Empty(tx2)
}

func (r *suiteEthRelayer) newEthRelayer() *Relayer4Ethereum {
	cfg := initCfg(*configPath)
	cfg.SyncTxConfig.Chain33Host = "http://127.0.0.1:8801"
	cfg.BridgeRegistry = r.x2EthDeployInfo.BridgeRegistry.Address.String()
	cfg.SyncTxConfig.PushBind = "127.0.0.1:60000"
	cfg.SyncTxConfig.FetchHeightPeriodMs = 50
	cfg.SyncTxConfig.Dbdriver = "memdb"
	cfg.SyncTxConfig.DbPath = "datadirEth"

	db := dbm.NewDB("relayer_db_service", cfg.SyncTxConfig.Dbdriver, cfg.SyncTxConfig.DbPath, cfg.SyncTxConfig.DbCache)

	relayer := &Relayer4Ethereum{
		provider:            cfg.EthProvider,
		db:                  db,
		unlockchan:          make(chan int, 2),
		rpcURL2Chain33:      cfg.SyncTxConfig.Chain33Host,
		bridgeRegistryAddr:  r.x2EthDeployInfo.BridgeRegistry.Address,
		maturityDegree:      cfg.EthMaturityDegree,
		fetchHeightPeriodMs: cfg.EthBlockFetchPeriod,
	}

	relayer.deployInfo = &ebTypes.Deploy{}
	//relayer.deployInfo.DeployerPrivateKey = "0x9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"
	//relayer.deployInfo.OperatorAddr = "0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF"
	//relayer.deployInfo.ValidatorsAddr = append(relayer.deployInfo.ValidatorsAddr, "0xA4Ea64a583F6e51C3799335b28a8F0529570A635")
	//relayer.deployInfo.ValidatorsAddr = append(relayer.deployInfo.ValidatorsAddr, "0x1919203bA8b325278d28Fb8fFeac49F2CD881A4e")
	//relayer.deployInfo.ValidatorsAddr = append(relayer.deployInfo.ValidatorsAddr, "0x9cBA1fF8D0b0c9Bc95d5762533F8CddBE795f687")
	//relayer.deployInfo.ValidatorsAddr = append(relayer.deployInfo.ValidatorsAddr, "0xdb15E7327aDc83F2878624bBD6307f5Af1B477b4")
	//InitPowers := []int64{int64(80), int64(10), int64(10), int64(10)}
	//relayer.deployInfo.InitPowers = InitPowers
	relayer.deployInfo.DeployerPrivateKey = common.ToHex(crypto.FromECDSA(r.para.DeployPrivateKey))
	relayer.deployInfo.OperatorAddr = r.para.Operator.String()
	for _, v := range r.para.InitValidators {
		relayer.deployInfo.ValidatorsAddr = append(relayer.deployInfo.ValidatorsAddr, v.String())
	}
	for _, v := range r.para.InitPowers {
		relayer.deployInfo.InitPowers = append(relayer.deployInfo.InitPowers, v.Int64())
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
	return relayer
}

func (r *suiteEthRelayer) proc() {
	backend, _ := r.newTestBackend()
	client, _ := backend.Attach()
	defer backend.Stop()
	defer client.Close()

	r.ethRelayer.clientChainID = new(big.Int)
	r.ethRelayer.backend = ethclient.NewClient(client)

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
			r.ethRelayer.filterLogEvents()
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

func (r *suiteEthRelayer) deployContracts() {
	//// 0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF
	//var deployerPrivateKey = "9dc6df3a8ab139a54d8a984f54958ae0661f880229bf3bdbb886b87d58b56a08"
	//// 0xA4Ea64a583F6e51C3799335b28a8F0529570A635
	//var ethValidatorAddrKeyA = "355b876d7cbcb930d5dfab767f66336ce327e082cbaa1877210c1bae89b1df71"
	//var ethValidatorAddrKeyB = "62ca4122aac0e6f35bed02fc15c7ddbdaa07f2f2a1821c8b8210b891051e3ee9"
	//var ethValidatorAddrKeyC = "4ae589fe3837dcfc90d1c85b8423dc30841525cbebc41dfb537868b0f8376bbf"
	//var ethValidatorAddrKeyD = "1385016736f7379884763f4a39811d1391fa156a7ca017be6afffa52bb327695"
	//ethValidatorAddrKeys := make([]string, 0)
	//ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyA)
	//ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyB)
	//ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyC)
	//ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyD)

	ctx := context.Background()
	r.backend, r.para = setup.PrepareTestEnv()
	r.sim = r.backend.(*backends.SimulatedBackend)

	balance, _ := r.sim.BalanceAt(ctx, r.para.Deployer, nil)
	r.Equal(balance.Int64(), int64(10000000000*10000))

	callMsg := ethereum.CallMsg{
		From: r.para.Deployer,
		Data: common.FromHex(generated.BridgeBankBin),
	}

	_, err := r.sim.EstimateGas(ctx, callMsg)
	r.NoError(err)

	r.x2EthContracts, r.x2EthDeployInfo, err = ethtxs.DeployAndInit(r.backend, r.para)
	r.NoError(err)
	r.sim.Commit()
}

func initCfg(path string) *relayerTypes.RelayerConfig {
	var cfg relayerTypes.RelayerConfig
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		//fmt.Println(err)
		os.Exit(-1)
	}
	return &cfg
}

func (r *suiteEthRelayer) newTestBackend() (*node.Node, []*types.Block) {
	// Generate test chain.
	genesis, blocks := r.generateTestChain()

	// Start Ethereum service.
	var ethservice *eth.Ethereum
	n, err := node.New(&node.Config{})
	err = n.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		config := &eth.Config{Genesis: genesis}
		config.Ethash.PowMode = ethash.ModeFake
		ethservice, err = eth.New(ctx, config)
		return ethservice, err
	})
	assert.NoError(r.T(), err)

	// Import the test chain.
	if err := n.Start(); err != nil {
		r.T().Fatalf("can't start test node: %v", err)
	}
	if _, err := ethservice.BlockChain().InsertChain(blocks[1:]); err != nil {
		r.T().Fatalf("can't import test blocks: %v", err)
	}
	return n, blocks
}

func (r *suiteEthRelayer) generateTestChain() (*core.Genesis, []*types.Block) {
	db := rawdb.NewMemoryDatabase()
	config := params.AllEthashProtocolChanges
	alloc := make(core.GenesisAlloc)
	alloc[testEthAddr] = core.GenesisAccount{Balance: testBalance}
	alloc[r.para.Operator] = core.GenesisAccount{Balance: testBalance}
	for _, v := range r.para.InitValidators {
		alloc[v] = core.GenesisAccount{Balance: testBalance}
	}
	genesis := &core.Genesis{
		Config: config,
		//Alloc:     core.GenesisAlloc{testEthAddr: {Balance: testBalance}},
		Alloc:     alloc,
		ExtraData: []byte("test genesis"),
		Timestamp: 9000,
	}
	generate := func(i int, g *core.BlockGen) {
		g.OffsetTime(1)
		g.SetExtra([]byte("test"))
	}
	gblock := genesis.ToBlock(db)
	engine := ethash.NewFaker()
	blocks, _ := core.GenerateChain(config, gblock, engine, db, 20, generate)
	blocks = append([]*types.Block{gblock}, blocks...)
	return genesis, blocks
}
