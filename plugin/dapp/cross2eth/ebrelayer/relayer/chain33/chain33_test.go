package chain33

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	chain33Common "github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/test/setup"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	tml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/33cn/chain33/system"

	// 需要显示引用系统插件，以加载系统内置合约
	"github.com/33cn/chain33/client/mocks"
	"github.com/stretchr/testify/mock"
)

var (
	configPath    = flag.String("f", "./../../relayer.toml", "configfile")
	privateKeyStr = "0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"
	accountAddr   = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
	passphrase    = "123456hzj"
)

func Test_ImportPrivateKey(t *testing.T) {
	mock33 := newMock33()
	defer mock33.Close()
	_, _, _, x2EthDeployInfo, err := setup.DeployContracts()
	require.NoError(t, err)
	chain33Relayer := newChain33Relayer(x2EthDeployInfo, "127.0.0.1:60000")

	err = chain33Relayer.ImportPrivateKey(passphrase, privateKeyStr)
	assert.NoError(t, err)
	//assert.Equal(t, addr, accountAddr)

	time.Sleep(50 * time.Millisecond)

	addr, err := chain33Relayer.GetAccountAddr()
	assert.NoError(t, err)
	assert.Equal(t, addr, accountAddr)

	key, _, _ := chain33Relayer.GetAccount("123")
	assert.NotEqual(t, key, privateKeyStr)

	key, _, _ = chain33Relayer.GetAccount(passphrase)
	assert.Equal(t, key, privateKeyStr)
}

func Test_RestorePrivateKeys(t *testing.T) {
	mock33 := newMock33()
	defer mock33.Close()
	_, _, _, x2EthDeployInfo, err := setup.DeployContracts()
	require.NoError(t, err)
	chain33Relayer := newChain33Relayer(x2EthDeployInfo, "127.0.0.1:60003")
	err = chain33Relayer.ImportPrivateKey(passphrase, privateKeyStr)
	assert.NoError(t, err)

	go func() {
		for range chain33Relayer.unlockChan {
		}
	}()

	err = chain33Relayer.RestorePrivateKeys("123")
	assert.NotEqual(t, chain33Common.ToHex(chain33Relayer.privateKey4Chain33.Bytes()), privateKeyStr)
	fmt.Println("err", err)
	assert.NoError(t, err)

	err = chain33Relayer.RestorePrivateKeys(passphrase)
	assert.Equal(t, chain33Common.ToHex(chain33Relayer.privateKey4Chain33.Bytes()), privateKeyStr)
	assert.NoError(t, err)

	err = chain33Relayer.StoreAccountWithNewPassphase("new123", passphrase)
	assert.NoError(t, err)

	err = chain33Relayer.RestorePrivateKeys("new123")
	assert.Equal(t, chain33Common.ToHex(chain33Relayer.privateKey4Chain33.Bytes()), privateKeyStr)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
}

func newChain33Relayer(x2EthDeployInfo *ethtxs.X2EthDeployInfo, pushBind string) *Relayer4Chain33 {
	cfg := initCfg(*configPath)
	cfg.SyncTxConfig.Chain33Host = "http://127.0.0.1:8801"
	cfg.BridgeRegistry = x2EthDeployInfo.BridgeRegistry.Address.String()
	cfg.SyncTxConfig.PushBind = pushBind
	cfg.SyncTxConfig.FetchHeightPeriodMs = 50
	cfg.SyncTxConfig.Dbdriver = "memdb"

	db := dbm.NewDB("relayer_db_service", cfg.SyncTxConfig.Dbdriver, cfg.SyncTxConfig.DbPath, cfg.SyncTxConfig.DbCache)
	ctx, cancel := context.WithCancel(context.Background())
	ethBridgeClaimchan := make(chan *relayerTypes.EthBridgeClaim, 100)
	chain33Msgchan := make(chan *events.Chain33Msg, 100)

	var wg sync.WaitGroup

	relayer := &Relayer4Chain33{
		rpcLaddr:             cfg.SyncTxConfig.Chain33Host,
		fetchHeightPeriodMs:  cfg.SyncTxConfig.FetchHeightPeriodMs,
		db:                   db,
		ctx:                  ctx,
		bridgeRegistryAddr:   x2EthDeployInfo.BridgeRegistry.Address.String(),
		chainName:            "",
		chainID:              0,
		unlockChan:           make(chan int),
		deployInfo:           cfg.Deploy,
		ethBridgeClaimChan:   ethBridgeClaimchan,
		chain33MsgChan:       chain33Msgchan,
		totalTx4Chain33ToEth: 0,
		symbol2Addr:          make(map[string]string),
		oracleAddr:           x2EthDeployInfo.Oracle.Address.String(),
		bridgeBankAddr:       x2EthDeployInfo.BridgeBank.Address.String(),
	}

	//err := relayer.setStatusCheckedIndex(1)
	//assert.NoError(t, err)
	//relayer.totalTx4Chain33ToEth = relayer.getTotalTxAmount2Eth()
	//relayer.statusCheckedIndex = relayer.getStatusCheckedIndex()
	//assert.Equal(t, relayer.statusCheckedIndex, int64(1))

	syncCfg := &ebTypes.SyncTxReceiptConfig{
		Chain33Host:       cfg.SyncTxConfig.Chain33Host,
		PushHost:          cfg.SyncTxConfig.PushHost,
		PushName:          cfg.SyncTxConfig.PushName,
		PushBind:          pushBind,
		StartSyncHeight:   cfg.SyncTxConfig.StartSyncHeight,
		StartSyncSequence: cfg.SyncTxConfig.StartSyncSequence,
		StartSyncHash:     cfg.SyncTxConfig.StartSyncHash,
	}
	go relayer.syncProc(syncCfg)

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

func initCfg(path string) *relayerTypes.RelayerConfig {
	var cfg relayerTypes.RelayerConfig
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	return &cfg
}

func newMock33() *testnode.Chain33Mock {
	var ret = types.ReplySubscribePush{IsOk: true, Msg: ""}
	var he = types.Header{Height: 10000}

	mockapi := &mocks.QueueProtocolAPI{}
	// 这里对需要mock的方法打桩,Close是必须的，其它方法根据需要
	mockapi.On("Close").Return()
	mockapi.On("AddPushSubscribe", mock.Anything).Return(&ret, nil)
	mockapi.On("GetLastHeader", mock.Anything).Return(&he, nil)

	mock33 := testnode.New("", mockapi)
	//defer mock33.Close()
	rpcCfg := mock33.GetCfg().RPC
	// 这里必须设置监听端口，默认的是无效值
	rpcCfg.JrpcBindAddr = "127.0.0.1:8801"
	mock33.GetRPC().Listen()

	return mock33
}

func Test_getExecerName(t *testing.T) {
	assert.Equal(t, getExecerName(""), "evm")
	assert.Equal(t, getExecerName("user.p.para."), "user.p.para.evm")
	assert.Equal(t, getExecerName("user.p.para.."), "user.p.para.evm")
	assert.Equal(t, getExecerName("user...p.para.."), "user.p.para.evm")
	assert.Equal(t, getExecerName("user.p...para.."), "user.p.para.evm")
	assert.Equal(t, getExecerName("user.p.para"), "user.p.para.evm")
	assert.Equal(t, getExecerName("user"), "user.evm")
}
