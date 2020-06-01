package ethereum

import (
	"context"
	"flag"
	"fmt"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/test/setup"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"os"
	"testing"
	"time"

	dbm "github.com/33cn/chain33/common/db"

	relayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	tml "github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
)

var (
	configPath           = flag.String("f", "./../../relayer.toml", "configfile")
	chain33PrivateKeyStr = "0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
	chain33AccountAddr   = "1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"

	passphrase = "123456hzj"

	testKey, _  = crypto.HexToECDSA("8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230")
	testAddr    = crypto.PubkeyToAddress(testKey.PublicKey)
	testBalance = big.NewInt(2e18)
)

type suiteEthRelayer struct {
	suite.Suite
	ethRelayer *Relayer4Ethereum

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

func (r *suiteEthRelayer) Test_3_IsValidatorActive() {
	return
	is, err := r.ethRelayer.IsValidatorActive("0x92c8b16afd6d423652559c6e266cbe1c29bfd84f")
	r.Equal(is, true)
	r.NoError(err)

	is, err = r.ethRelayer.IsValidatorActive("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF")
	r.Equal(is, false)
	r.NoError(err)

	is, err = r.ethRelayer.IsValidatorActive("123")
	r.Error(err)
}

//func (r *suiteEthRelayer) Test_4_DeployContrcts() {
//	bridgeRegistry, err := r.ethRelayer.DeployContrcts()
//	r.NoError(err)
//	r.ethRelayer.bridgeRegistryAddr = common.HexToAddress(bridgeRegistry)
//
//	//time.Sleep(50 * time.Second)
//}

func (r *suiteEthRelayer) Test_4_LockEth() {
	ctx := context.Background()
	bridgeBankBalance, err := r.sim.BalanceAt(ctx, r.x2EthDeployInfo.BridgeBank.Address, nil)
	fmt.Println(bridgeBankBalance, err)
	r.NoError(err)

	userOneAuth, err := ethtxs.PrepareAuth(r.backend, r.para.ValidatorPriKey[0], r.para.InitValidators[0])
	r.NoError(err)
	ethAmount := big.NewInt(50)
	userOneAuth.Value = ethAmount

	//lock 50 eth
	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
	_, err = r.x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethAmount)
	r.NoError(err)
	r.sim.Commit()

	bridgeBankBalance, err = r.sim.BalanceAt(ctx, r.x2EthDeployInfo.BridgeBank.Address, nil)
	r.NoError(err)
	//require.Equal(t, bridgeBankBalance.Int64(), ethAmount.Int64())
	fmt.Println(bridgeBankBalance, err)
	time.Sleep(1 * time.Second)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{r.ethRelayer.bridgeBankAddr},
	}
	logs, err := r.sim.FilterLogs(context.Background(), query)
	if err != nil {
		errinfo := fmt.Sprintf("Failed to filterLogEvents due to:%s", err.Error())
		fmt.Println(errinfo)
	}

	for _, logv := range logs {
		if err := r.ethRelayer.setEthTxEvent(logv); nil != err {
			//	panic(err.Error())
		}
	}

	time.Sleep(5 * time.Second)
}

//func (r *suiteEthRelayer) TestRelayer4Ethereum_ApproveAllowance(t *testing.T) {
//	r.ethRelayer.ApproveAllowance()
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
	backend, chain := newTestBackend(r.T())
	client, _ := backend.Attach()
	defer backend.Stop()
	defer client.Close()

	//r.ethRelayer.backend = r.sim
	r.ethRelayer.clientChainID = new(big.Int)

	ec := ethclient.NewClient(client)
	r.ethRelayer.backend = ec
	_ = chain

	ctx1 := context.Background()
	balance, _ := ec.BalanceAt(ctx1, testAddr, nil)
	fmt.Println("deployer addr,", testAddr, "balance =", balance.String())

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

//func Test3(t *testing.T) {
//	i := 1
//	j := 1
//	for i > 0 {
//		i++
//		fmt.Println("i", i)
//		time.Sleep(time.Second)
//		goto aaa
//	}
//
//	fmt.Println("--00--")
//
//aaa:
//	for j > 0 {
//		j++
//		fmt.Println("j", j)
//		time.Sleep(time.Second)
//		if j > 3 {
//			j = 1
//			break
//		}
//	}
//
//	fmt.Println("--00--")
//}

func (r *suiteEthRelayer) deployContracts() {
	ctx := context.Background()
	//var backend bind.ContractBackend
	r.backend, r.para = setup.PrepareTestEnvironment()
	r.sim = r.backend.(*backends.SimulatedBackend)

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

	r.x2EthContracts, r.x2EthDeployInfo, err = ethtxs.DeployAndInit(r.backend, r.para)
	r.NoError(err)
	r.sim.Commit()
}

//func (r *suiteEthRelayer) filterLogEvents() {
//	deployHeight := int64(0)
//	height4BridgeBankLogAt := int64(r.ethRelayer.getHeight4BridgeBankLogAt())
//
//	if height4BridgeBankLogAt < deployHeight {
//		height4BridgeBankLogAt = deployHeight
//	}
//
//	curHeight := int64(0)
//	relayerLog.Info("filterLogEvents", "curHeight:", curHeight)
//
//	bridgeBankSig := make(map[string]bool)
//	bridgeBankSig[r.ethRelayer.bridgeBankEventLockSig] = true
//	bridgeBankSig[r.ethRelayer.bridgeBankEventBurnSig] = true
//	bridgeBankLog := make(chan types.Log)
//	done := make(chan int)
//	go r.ethRelayer.filterLogEventsProc(bridgeBankLog, done, "bridgeBank", curHeight, height4BridgeBankLogAt, r.ethRelayer.bridgeBankAddr, bridgeBankSig)
//
//	for {
//		select {
//		case vLog := <-bridgeBankLog:
//			r.ethRelayer.storeBridgeBankLogs(vLog, true)
//		case vLog := <-r.ethRelayer.bridgeBankLog:
//			//因为此处是同步保存信息，防止未同步完成出现panic时，直接将其设置为最新高度，中间出现部分信息不同步的情况
//			r.ethRelayer.storeBridgeBankLogs(vLog, false)
//		case <-done:
//			relayerLog.Info("Finshed offline logs processed")
//			return
//		}
//	}
//}

//func (r *suiteEthRelayer) procNewHeight(ctx context.Context, continueFailCount *int32) {
//	*continueFailCount = 0
//
//	currentHeight := uint64(20)
//	relayerLog.Info("procNewHeight", "currentHeight", currentHeight)
//	//一次最大只获取10个logEvent进行处理
//	fetchCnt := int32(10)
//	for r.ethRelayer.eventLogIndex.Height+uint64(r.ethRelayer.maturityDegree)+1 <= currentHeight {
//		logs, err := r.ethRelayer.getNextValidEthTxEventLogs(r.ethRelayer.eventLogIndex.Height, r.ethRelayer.eventLogIndex.Index, fetchCnt)
//		if nil != err {
//			relayerLog.Error("Failed to get ethereum height", "getNextValidEthTxEventLogs err", err.Error())
//			return
//		}
//
//		for i, vLog := range logs {
//			if vLog.BlockNumber+uint64(r.ethRelayer.maturityDegree)+1 > currentHeight {
//				logs = logs[:i]
//				break
//			}
//			//r.ethRelayer.procBridgeBankLogs(*vLog)
//
//			if r.ethRelayer.checkTxProcessed(vLog.TxHash.Bytes()) {
//				relayerLog.Info("procBridgeBankLogs", "Tx has been already Processed with hash:", vLog.TxHash.Hex(),
//					"height", vLog.BlockNumber, "index", vLog.Index)
//				return
//			}
//
//			defer func() {
//				if err := r.ethRelayer.setTxProcessed(vLog.TxHash.Bytes()); nil != err {
//					panic(err.Error())
//				}
//			}()
//			//lock,用于捕捉 (ETH/ERC20----->chain33) 跨链转移
//			if vLog.Topics[0].Hex() == r.ethRelayer.bridgeBankEventLockSig {
//				eventName := events.LogLock.String()
//				relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
//					"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
//				err := r.ethRelayer.handleLogLockEvent(r.ethRelayer.clientChainID, r.ethRelayer.bridgeBankAbi, eventName, vLog)
//				if err != nil {
//					errinfo := fmt.Sprintf("Failed to handleLogLockEvent due to:%s", err.Error())
//					relayerLog.Info("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
//					panic(errinfo)
//				}
//			} else if vLog.Topics[0].Hex() == r.ethRelayer.bridgeBankEventBurnSig {
//				//burn,用于捕捉 (chain33 token----->chain33) 实现chain33资产withdraw操作，之后在chain33上实现unlock操作
//				eventName := events.LogChain33TokenBurn.String()
//				relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
//					"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
//				err := r.ethRelayer.handleLogBurnEvent(r.ethRelayer.clientChainID, r.ethRelayer.bridgeBankAbi, eventName, vLog)
//				if err != nil {
//					errinfo := fmt.Sprintf("Failed to handleLogBurnEvent due to:%s", err.Error())
//					relayerLog.Info("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
//					panic(errinfo)
//				}
//			}
//		}
//
//		cnt := int32(len(logs))
//		if len(logs) > 0 {
//			//firstHeight := logs[0].BlockNumber
//			lastHeight := logs[cnt-1].BlockNumber
//			index := logs[cnt-1].TxIndex
//			//获取的数量小于批量获取数量，则认为直接
//			r.ethRelayer.setBridgeBankProcessedHeight(lastHeight, uint32(index))
//			r.ethRelayer.eventLogIndex.Height = lastHeight
//			r.ethRelayer.eventLogIndex.Index = uint32(index)
//		}
//
//		//当前需要处理的event数量已经少于10个，直接返回
//		if cnt < fetchCnt {
//			return
//		}
//	}
//}

func initCfg(path string) *relayerTypes.RelayerConfig {
	var cfg relayerTypes.RelayerConfig
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		fmt.Println(err)
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
	n.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		config := &eth.Config{Genesis: genesis}
		config.Ethash.PowMode = ethash.ModeFake
		ethservice, err = eth.New(ctx, config)
		return ethservice, err
	})

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
		Alloc:     core.GenesisAlloc{testAddr: {Balance: testBalance}},
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
