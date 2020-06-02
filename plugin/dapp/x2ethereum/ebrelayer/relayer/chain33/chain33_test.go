package chain33

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/test/setup"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
	syncTx "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/relayer/chain33/transceiver/sync"
	ebTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	tml "github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	// 需要显示引用系统插件，以加载系统内置合约
	"github.com/33cn/chain33/client/mocks"
	_ "github.com/33cn/chain33/system"
	"github.com/stretchr/testify/mock"
)

var (
	configPath    = flag.String("f", "./../../relayer.toml", "configfile")
	chainTestCfg  = types.NewChain33Config(types.GetDefaultCfgstring())
	privateKeyStr = "0x3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
	accountAddr   = "0x92c8b16afd6d423652559c6e266cbe1c29bfd84f"
	passphrase    = "123456hzj"
	test          = "0ac3050aa3020a0a7832657468657265756d126d60671a690a2a3078303030303030303030303030303030303030303030303030303030303030303030303030303030301a2a307830633035626135633233306664616135303362353337303261663139363265303864306336306266220831303030303030302a0365746838121a6e080112210320bbac09528e19c55b0f89cb37ab265e7e856b1a8c388780322dbbfd194b52ba1a473045022100c403d9a6e531292336b44d52e4f4dbb9b8ab1e16335383954583728b909478da022031d8a29efcbcea8df648c4054f3c09ab1ab7a330797cf79fd891a3d9336922e920a08d0628e0f193f60530a1d7ad93e5ebc28e253a22314c7538586d537459765777664e716951336e4e4b33345239466648346b5270425612ce0208021a5e0802125a0a2b10c0d59294bb192222313271796f6361794e46374c7636433971573461767873324537553431664b536676122b10a0c88c94bb192222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a55080f12510a291080ade2042222313271796f6361794e46374c7636433971573461767873324537553431664b53667612242222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a92010867128d010a2a3078303030303030303030303030303030303030303030303030303030303030303030303030303030301222313271796f6361794e46374c7636433971573461767873324537553431664b5366761a2a307830633035626135633233306664616135303362353337303261663139363265303864306336306266220831303030303030302a03657468301220c4092a207a38e1da7de4444f2d34c7488293f3a2e01ce2561e720e9bbef355e83755ad833220e68d8418f69d5f18278a53dca53b101f26f76883337a60a5754d5f6d94e42e3c400148c409"
)

type suiteChain33Relayer struct {
	suite.Suite
	chain33Relayer  *Relayer4Chain33
	sim             *backends.SimulatedBackend
	x2EthContracts  *ethtxs.X2EthContracts
	x2EthDeployInfo *ethtxs.X2EthDeployInfo
	para            *ethtxs.DeployPara
}

func TestRunSuiteX2Ethereum(t *testing.T) {
	log := new(suiteChain33Relayer)
	suite.Run(t, log)
}

func (r *suiteChain33Relayer) SetupSuite() {
	r.deployContracts()
	r.chain33Relayer = r.newChain33Relayer()
}

func (r *suiteChain33Relayer) Test_1_ImportPrivateKey() {
	addr, err := r.chain33Relayer.ImportPrivateKey(passphrase, privateKeyStr)
	r.NoError(err)
	r.Equal(addr, accountAddr)

	time.Sleep(50 * time.Millisecond)

	addr, err = r.chain33Relayer.GetAccountAddr()
	r.NoError(err)
	r.Equal(addr, accountAddr)

	key, _, _ := r.chain33Relayer.GetAccount("123")
	r.NotEqual(key, privateKeyStr)

	key, _, _ = r.chain33Relayer.GetAccount(passphrase)
	r.Equal(key, privateKeyStr)
}

func (r *suiteChain33Relayer) Test_2_HandleRequest() {
	body, err := hex.DecodeString(test)
	r.NoError(err)

	r.chain33Relayer.statusCheckedIndex = 1220
	err = syncTx.HandleRequest(body)
	r.NoError(err)

	//time.Sleep(50 * time.Second)
	time.Sleep(50 * time.Millisecond)
}

func (r *suiteChain33Relayer) Test_3_QueryTxhashRelay2Eth() {
	ret := r.chain33Relayer.QueryTxhashRelay2Eth()
	r.NotEmpty(ret)
}

func (r *suiteChain33Relayer) Test_4_StoreAccountWithNewPassphase() {
	err := r.chain33Relayer.StoreAccountWithNewPassphase(passphrase, passphrase)
	r.NoError(err)
}

func (r *suiteChain33Relayer) Test_5_getEthTxhash() {
	txIndex := atomic.LoadInt64(&r.chain33Relayer.totalTx4Chain33ToEth)
	hash, err := r.chain33Relayer.getEthTxhash(txIndex)
	r.NoError(err)
	r.Equal(hash.String(), "0x6fa087c7a2a8a4421f6e269fbc6c0838e99fa59d5760155a71cd7eb1c01aafad")
}

func (r *suiteChain33Relayer) Test_7_RestorePrivateKeys() {
	//err := r.chain33Relayer.RestorePrivateKeys("123") // 不会报错
	//r.Error(err)

	go func() {
		time.Sleep(1 * time.Millisecond)
		<-r.chain33Relayer.unlock
	}()
	err := r.chain33Relayer.RestorePrivateKeys(passphrase)
	r.NoError(err)
}

func (r *suiteChain33Relayer) newChain33Relayer() *Relayer4Chain33 {
	cfg := initCfg(*configPath)
	cfg.SyncTxConfig.Chain33Host = "http://127.0.0.1:8801"
	cfg.BridgeRegistry = r.x2EthDeployInfo.BridgeRegistry.Address.String()
	cfg.SyncTxConfig.PushBind = "127.0.0.1:60000"
	cfg.SyncTxConfig.FetchHeightPeriodMs = 50
	cfg.SyncTxConfig.Dbdriver = "memdb"

	db := dbm.NewDB("relayer_db_service", cfg.SyncTxConfig.Dbdriver, cfg.SyncTxConfig.DbPath, cfg.SyncTxConfig.DbCache)
	ctx, cancel := context.WithCancel(context.Background())

	relayer := &Relayer4Chain33{
		rpcLaddr:            cfg.SyncTxConfig.Chain33Host,
		fetchHeightPeriodMs: cfg.SyncTxConfig.FetchHeightPeriodMs,
		unlock:              make(chan int),
		db:                  db,
		ctx:                 ctx,
	}
	err := relayer.setStatusCheckedIndex(1)
	r.NoError(err)

	relayer.ethBackend = r.sim
	relayer.bridgeRegistryAddr = r.para.Deployer
	relayer.totalTx4Chain33ToEth = relayer.getTotalTxAmount2Eth()
	relayer.statusCheckedIndex = relayer.getStatusCheckedIndex()
	r.Equal(relayer.statusCheckedIndex, int64(1))

	syncCfg := &ebTypes.SyncTxReceiptConfig{
		Chain33Host:       cfg.SyncTxConfig.Chain33Host,
		PushHost:          cfg.SyncTxConfig.PushHost,
		PushName:          cfg.SyncTxConfig.PushName,
		PushBind:          cfg.SyncTxConfig.PushBind,
		StartSyncHeight:   cfg.SyncTxConfig.StartSyncHeight,
		StartSyncSequence: cfg.SyncTxConfig.StartSyncSequence,
		StartSyncHash:     cfg.SyncTxConfig.StartSyncHash,
	}
	_ = syncCfg
	go r.syncProc(syncCfg)

	var wg sync.WaitGroup
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
		wg.Wait()
		os.Exit(0)
	}()

	return relayer
}

func (r *suiteChain33Relayer) deployContracts() {
	// 0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a
	var deployerPrivateKey = "8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
	// 0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f
	var ethValidatorAddrKeyA = "3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
	var ethValidatorAddrKeyB = "a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
	var ethValidatorAddrKeyC = "bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
	var ethValidatorAddrKeyD = "c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"
	ethValidatorAddrKeys := make([]string, 0)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyA)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyB)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyC)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyD)

	ctx := context.Background()
	var backend bind.ContractBackend
	backend, r.para = setup.PrepareTestEnvironment(deployerPrivateKey, ethValidatorAddrKeys)
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

func (r *suiteChain33Relayer) syncProc(syncCfg *ebTypes.SyncTxReceiptConfig) {
	var ret = types.ReplySubscribePush{IsOk: true}
	var he = types.Header{Height: 10000}

	mockapi := &mocks.QueueProtocolAPI{}
	// 这里对需要mock的方法打桩,Close是必须的，其它方法根据需要
	mockapi.On("Close").Return()
	mockapi.On("AddPushSubscribe", mock.Anything).Return(&ret, nil)
	mockapi.On("GetLastHeader", mock.Anything).Return(&he, nil)
	mockapi.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)

	mock33 := testnode.New("", mockapi)
	defer mock33.Close()
	rpcCfg := mock33.GetCfg().RPC
	// 这里必须设置监听端口，默认的是无效值
	rpcCfg.JrpcBindAddr = "127.0.0.1:8801"
	mock33.GetRPC().Listen()

	fmt.Println("Pls unlock or import private key for Chain33 relayer")
	<-r.chain33Relayer.unlock
	fmt.Println("Chain33 relayer starts to run...")

	r.chain33Relayer.syncTxReceipts = syncTx.StartSyncTxReceipt(syncCfg, r.chain33Relayer.db)
	r.chain33Relayer.lastHeight4Tx = r.chain33Relayer.loadLastSyncHeight()
	r.chain33Relayer.oracleInstance = r.x2EthContracts.Oracle

	timer := time.NewTicker(time.Duration(r.chain33Relayer.fetchHeightPeriodMs) * time.Millisecond)
	for {
		select {
		case <-timer.C:
			height := r.chain33Relayer.getCurrentHeight()
			relayerLog.Debug("syncProc", "getCurrentHeight", height)
			r.chain33Relayer.onNewHeightProc(height)

		case <-r.chain33Relayer.ctx.Done():
			timer.Stop()
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
