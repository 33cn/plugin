package ethereum

import (
	"context"
	"encoding/hex"
	"flag"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/util/testnode"
	"github.com/stretchr/testify/mock"

	dbm "github.com/33cn/chain33/common/db"
	_ "github.com/33cn/chain33/system"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/test/setup"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	tml "github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"
)

var (
	configPath           = flag.String("f", "./../../relayer.toml", "configfile")
	chain33PrivateKeyStr = "0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
	chain33AccountAddr   = "1GTxrmuWiXavhcvsaH5w9whgVxUrWsUMdV"
	passphrase           = "123456hzj"
	chainTestCfg         = chain33Types.NewChain33Config(chain33Types.GetDefaultCfgstring())

	// 0x8AFDADFC88a1087c9A1D6c0F5Dd04634b87F303a
	deployerPrivateKey = "8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230"
	// 0x92C8b16aFD6d423652559C6E266cBE1c29Bfd84f
	ethValidatorAddrKeyA = "3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
	ethValidatorAddrKeyB = "a5f3063552f4483cfc20ac4f40f45b798791379862219de9e915c64722c1d400"
	ethValidatorAddrKeyC = "bbf5e65539e9af0eb0cfac30bad475111054b09c11d668fc0731d54ea777471e"
	ethValidatorAddrKeyD = "c9fa31d7984edf81b8ef3b40c761f1847f6fcd5711ab2462da97dc458f1f896b"
)

type suiteEthRelayer struct {
	suite.Suite
	ethRelayer      *Relayer4Ethereum
	sim             *ethinterface.SimExtend
	x2EthContracts  *ethtxs.X2EthContracts
	x2EthDeployInfo *ethtxs.X2EthDeployInfo
	para            *ethtxs.DeployPara
	backend         bind.ContractBackend
}

func TestRunSuiteX2Ethereum(t *testing.T) {
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

	log := new(suiteEthRelayer)
	suite.Run(t, log)
}

func (r *suiteEthRelayer) SetupSuite() {
	r.deployContracts()
	r.ethRelayer = r.newEthRelayer()

	_, err := r.ethRelayer.DeployContrcts()
	r.Error(err)
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

func (r *suiteEthRelayer) Test_2_IsValidatorActive() {
	r.ethRelayer.x2EthContracts = r.x2EthContracts
	r.ethRelayer.x2EthDeployInfo = r.x2EthDeployInfo

	is, err := r.ethRelayer.IsValidatorActive("0x92c8b16afd6d423652559c6e266cbe1c29bfd84f")
	r.Equal(is, true)
	r.NoError(err)

	is, err = r.ethRelayer.IsValidatorActive("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF")
	r.Equal(is, false)
	r.NoError(err)

	is, err = r.ethRelayer.IsValidatorActive("123")
	r.Error(err)
}

func (r *suiteEthRelayer) Test_3_Token() {
	r.isChain33KeyImport()
	r.ethRelayer.prePareSubscribeEvent()

	addr, err := r.ethRelayer.ShowBridgeBankAddr()
	r.NoError(err)
	r.Equal(addr, r.x2EthDeployInfo.BridgeBank.Address.String())

	addr, err = r.ethRelayer.ShowBridgeRegistryAddr()
	r.NoError(err)
	r.Equal(addr, r.x2EthDeployInfo.BridgeRegistry.Address.String())

	addr, err = r.ethRelayer.ShowOperator()
	r.NoError(err)
	r.Equal(addr, r.para.Operator.String())

	balance, err := r.ethRelayer.GetBalance("", r.para.InitValidators[0].String())
	r.NoError(err)
	r.Equal(balance, "10000000000")

	tokenAddrbty, err := r.ethRelayer.CreateBridgeToken("bty")
	r.NoError(err)
	r.NotEmpty(tokenAddrbty)
	r.sim.Commit()

	addr, err = r.ethRelayer.ShowTokenAddrBySymbol("bty")
	r.NoError(err)
	r.Equal(addr, tokenAddrbty)

	tokenErc20Addr, err := r.ethRelayer.CreateERC20Token("testcc")
	r.NoError(err)
	r.NotEmpty(tokenErc20Addr)
	r.sim.Commit()

	_, err = r.ethRelayer.MintERC20Token(tokenErc20Addr, r.para.Deployer.String(), "20000000000000")
	r.NoError(err)
	r.sim.Commit()

	balance, err = r.ethRelayer.ShowDepositStatics(tokenErc20Addr)
	r.NoError(err)
	r.Equal(balance, "20000000000000")

	//claimID := crypto.Keccak256Hash(big.NewInt(50).Bytes())
	//ret, err = r.ethRelayer.IsProphecyPending(claimID)
	//r.NoError(err)
	//r.Equal(ret, false)

	decimals, err := r.ethRelayer.GetDecimals(tokenAddrbty)
	r.NoError(err)
	r.Equal(decimals, uint8(8))

	_, err = r.ethRelayer.Burn(ethValidatorAddrKeyA, tokenAddrbty, chain33AccountAddr, "10")
	r.Error(err)

	_, err = r.ethRelayer.BurnAsync(ethValidatorAddrKeyA, tokenAddrbty, chain33AccountAddr, "10")
	r.Error(err)

	txhash, err := r.ethRelayer.TransferToken(tokenErc20Addr, hexutil.Encode(crypto.FromECDSA(r.para.DeployPrivateKey)), r.ethRelayer.deployInfo.ValidatorsAddr[0], "100")
	r.NoError(err)
	r.sim.Commit()

	_, err = r.ethRelayer.ShowTxReceipt(txhash)
	r.NoError(err)

	balance, err = r.ethRelayer.GetBalance(tokenErc20Addr, r.ethRelayer.deployInfo.ValidatorsAddr[0])
	r.NoError(err)
	r.Equal(balance, "100")

	balance, err = r.ethRelayer.GetBalance(tokenErc20Addr, r.para.Deployer.String())
	r.NoError(err)
	r.Equal(balance, "19999999999900")

	tx1 := r.ethRelayer.QueryTxhashRelay2Eth()
	r.Empty(tx1)

	tx2 := r.ethRelayer.QueryTxhashRelay2Chain33()
	r.Empty(tx2)
}

func (r *suiteEthRelayer) Test_4_LockEth() {
	r.isChain33KeyImport()

	ctx := context.Background()
	bridgeBankBalance, err := r.sim.BalanceAt(ctx, r.x2EthDeployInfo.BridgeBank.Address, nil)
	r.NoError(err)
	r.Equal(bridgeBankBalance.Int64(), int64(0))

	userOneAuth, err := ethtxs.PrepareAuth(r.sim, r.para.ValidatorPriKey[0], r.para.InitValidators[0])
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

	for i := 0; i < 11; i++ {
		r.sim.Commit()
	}

	time.Sleep(time.Duration(r.ethRelayer.fetchHeightPeriodMs) * time.Millisecond)

	balance, err := r.ethRelayer.ShowLockStatics("")
	r.NoError(err)
	r.Equal(balance, "50")
}

//func (r *suiteEthRelayer) Test_5_LockErc20() {
//	r.isChain33KeyImport()
//
//	tokenErc20Addr, err := r.ethRelayer.CreateERC20Token("testc")
//	r.NoError(err)
//	r.NotEmpty(tokenErc20Addr)
//	r.sim.Commit()
//
//	_, err = r.ethRelayer.MintERC20Token(tokenErc20Addr, r.para.Deployer.String(), "10000000000000")
//	r.NoError(err)
//	r.sim.Commit()
//
//	balance, err := r.ethRelayer.ShowDepositStatics(tokenErc20Addr)
//	r.NoError(err)
//	r.Equal(balance, "10000000000000")
//
//	chain33Receiver := "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
//	_, err = r.ethRelayer.LockEthErc20Asset(hexutil.Encode(crypto.FromECDSA(r.para.DeployPrivateKey)), tokenErc20Addr, "100", chain33Receiver)
//	r.NoError(err)
//	r.sim.Commit()
//
//	balance, err = r.ethRelayer.GetBalance(tokenErc20Addr, r.para.Deployer.String())
//	r.NoError(err)
//	r.Equal(balance, "9999999999900")
//
//	_, err = r.ethRelayer.ApproveAllowance(hexutil.Encode(crypto.FromECDSA(r.para.DeployPrivateKey)), tokenErc20Addr, "500")
//	r.NoError(err)
//	r.sim.Commit()
//
//	_, err = r.ethRelayer.LockEthErc20AssetAsync(hexutil.Encode(crypto.FromECDSA(r.para.DeployPrivateKey)), tokenErc20Addr, "100", chain33Receiver)
//	r.NoError(err)
//	r.sim.Commit()
//
//	balance, err = r.ethRelayer.GetBalance(tokenErc20Addr, r.para.Deployer.String())
//	r.NoError(err)
//	r.Equal(balance, "9999999999800")
//
//	//ret, err := r.ethRelayer.ApproveAllowance(hexutil.Encode(crypto.FromECDSA(r.para.DeployPrivateKey)), tokenErc20Addr, "20000000000000")
//	//fmt.Println("?????", ret, err)
//	//r.Error(err)
//
//	//ctx := context.Background()
//	//bridgeBankBalance, err := r.sim.BalanceAt(ctx, r.x2EthDeployInfo.BridgeBank.Address, nil)
//	//r.NoError(err)
//	//r.Equal(bridgeBankBalance.Int64(), int64(0))
//	//
//	//userOneAuth, err := ethtxs.PrepareAuth(r.sim, r.para.DeployPrivateKey, r.para.Deployer)
//	//r.NoError(err)
//	//
//	//allowAmount := int64(100)
//	//erc20TokenInstance, err := generated.NewBridgeToken(common.HexToAddress(tokenErc20Addr), r.ethRelayer.clientSpec)
//	//r.NoError(err)
//	//
//	//_, err = erc20TokenInstance.Approve(userOneAuth, r.ethRelayer.x2EthDeployInfo.BridgeBank.Address, big.NewInt(allowAmount))
//	//r.NoError(err)
//	//r.sim.Commit()
//	//
//	////lock 100 testc
//	//chain33Receiver := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	//lockAmount := big.NewInt(100)
//	//userOneAuth, err = ethtxs.PrepareAuth(r.sim, r.para.DeployPrivateKey, r.para.Deployer)
//	//r.NoError(err)
//	//
//	//_, err = r.x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Receiver, common.HexToAddress(tokenErc20Addr), lockAmount)
//	//r.NoError(err)
//	//r.sim.Commit()
//	//
//	//balance, err = r.ethRelayer.ShowLockStatics(tokenErc20Addr)
//	//r.NoError(err)
//	//r.Equal(balance, "100")
//
//	for i := 0; i < 11; i++ {
//		r.sim.Commit()
//	}
//	time.Sleep(time.Duration(r.ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//}

func (r *suiteEthRelayer) Test_6_RestorePrivateKeys() {
	go func() {
		for range r.ethRelayer.unlockchan {
		}
	}()
	temp := r.ethRelayer.privateKey4Chain33

	err := r.ethRelayer.RestorePrivateKeys("123")
	r.NotEqual(hex.EncodeToString(temp.Bytes()), hex.EncodeToString(r.ethRelayer.privateKey4Chain33.Bytes()))
	r.NoError(err)

	err = r.ethRelayer.RestorePrivateKeys(passphrase)
	r.Equal(hex.EncodeToString(temp.Bytes()), hex.EncodeToString(r.ethRelayer.privateKey4Chain33.Bytes()))
	r.NoError(err)

	err = r.ethRelayer.StoreAccountWithNewPassphase("new123", passphrase)
	r.NoError(err)

	err = r.ethRelayer.RestorePrivateKeys("new123")
	r.Equal(hex.EncodeToString(temp.Bytes()), hex.EncodeToString(r.ethRelayer.privateKey4Chain33.Bytes()))
	r.NoError(err)

	time.Sleep(time.Second)
}

func (r *suiteEthRelayer) isChain33KeyImport() {
	_, err := r.ethRelayer.GetValidatorAddr()
	if err != nil {
		err = r.ethRelayer.ImportChain33PrivateKey(passphrase, chain33PrivateKeyStr)
		r.NoError(err)
	}
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
		deployInfo:          cfg.Deploy,
	}

	_ = relayer.setBridgeRegistryAddr(cfg.BridgeRegistry)
	registrAddrInDB, err := relayer.getBridgeRegistryAddr()
	r.NoError(err)
	r.Equal(registrAddrInDB, cfg.BridgeRegistry)
	relayer.eventLogIndex = relayer.getLastBridgeBankProcessedHeight()
	relayer.initBridgeBankTx()

	relayer.clientSpec = r.sim
	relayer.clientChainID = big.NewInt(1)

	deployPrivateKey, err := crypto.ToECDSA(common.FromHex(relayer.deployInfo.DeployerPrivateKey))
	deployerAddr := crypto.PubkeyToAddress(deployPrivateKey.PublicKey)
	relayer.operatorInfo = &ethtxs.OperatorInfo{
		PrivateKey: deployPrivateKey,
		Address:    deployerAddr,
	}
	relayer.deployPara = r.para
	relayer.x2EthContracts = r.x2EthContracts
	relayer.x2EthDeployInfo = r.x2EthDeployInfo

	go relayer.proc()

	var wg sync.WaitGroup
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)
	go func() {
		<-ch
		//cancel()
		wg.Wait()
		os.Exit(0)
	}()

	return relayer
}

func (r *suiteEthRelayer) deployContracts() {
	ethValidatorAddrKeys := make([]string, 0)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyA)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyB)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyC)
	ethValidatorAddrKeys = append(ethValidatorAddrKeys, ethValidatorAddrKeyD)

	ctx := context.Background()
	r.backend, r.para = setup.PrepareTestEnvironment(deployerPrivateKey, ethValidatorAddrKeys)
	r.sim = new(ethinterface.SimExtend)
	r.sim.SimulatedBackend = r.backend.(*backends.SimulatedBackend)

	balance, _ := r.sim.BalanceAt(ctx, r.para.Deployer, nil)
	r.Equal(balance.Int64(), int64(10000000000*10000))

	callMsg := ethereum.CallMsg{
		From: r.para.Deployer,
		Data: common.FromHex(generated.BridgeBankBin),
	}

	_, err := r.sim.EstimateGas(ctx, callMsg)
	r.NoError(err)

	r.x2EthContracts, r.x2EthDeployInfo, err = ethtxs.DeployAndInit(r.sim, r.para)
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
