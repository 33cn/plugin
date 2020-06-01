package ethereum

import (
	"context"
	"flag"
	"github.com/stretchr/testify/assert"
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
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
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

func (r *suiteEthRelayer) Test_2_RestorePrivateKeys() {
	// 错误的密码 也不报错
	err := r.ethRelayer.RestorePrivateKeys(passphrase)
	r.NoError(err)

	err = r.ethRelayer.StoreAccountWithNewPassphase(passphrase, passphrase)
	r.NoError(err)
}

func (r *suiteEthRelayer) Test_3_LockEth() {
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

	query := ethereum.FilterQuery{
		Addresses: []common.Address{r.ethRelayer.bridgeBankAddr},
	}
	logs, err := r.sim.FilterLogs(context.Background(), query)
	r.NoError(err)

	for _, logv := range logs {
		if err := r.ethRelayer.setEthTxEvent(logv); nil != err {
			r.NoError(err)
		}
	}

	time.Sleep(time.Duration(r.ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
	time.Sleep(10 * time.Second)
}

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
	return relayer
}

func (r *suiteEthRelayer) proc() {
	backend, _ := newTestBackend(r.T())
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
	ctx := context.Background()
	//var backend bind.ContractBackend
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

func newTestBackend(t *testing.T) (*node.Node, []*types.Block) {
	// Generate test chain.
	genesis, blocks := generateTestChain()

	// Start Ethereum service.
	var ethservice *eth.Ethereum
	n, err := node.New(&node.Config{})
	err = n.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		config := &eth.Config{Genesis: genesis}
		config.Ethash.PowMode = ethash.ModeFake
		ethservice, err = eth.New(ctx, config)
		return ethservice, err
	})
	assert.NoError(t, err)

	// Import the test chain.
	if err := n.Start(); err != nil {
		t.Fatalf("can't start test node: %v", err)
	}
	if _, err := ethservice.BlockChain().InsertChain(blocks[1:]); err != nil {
		t.Fatalf("can't import test blocks: %v", err)
	}
	return n, blocks
}

func generateTestChain() (*core.Genesis, []*types.Block) {
	db := rawdb.NewMemoryDatabase()
	config := params.AllEthashProtocolChanges
	genesis := &core.Genesis{
		Config:    config,
		Alloc:     core.GenesisAlloc{testEthAddr: {Balance: testBalance}},
		ExtraData: []byte("test genesis"),
		Timestamp: 9000,
	}
	generate := func(i int, g *core.BlockGen) {
		g.OffsetTime(1)
		g.SetExtra([]byte("test"))
	}
	gblock := genesis.ToBlock(db)
	engine := ethash.NewFaker()
	blocks, _ := core.GenerateChain(config, gblock, engine, db, 50, generate)
	blocks = append([]*types.Block{gblock}, blocks...)
	return genesis, blocks
}
