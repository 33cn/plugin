package ethereum

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/33cn/chain33/client/mocks"
	dbm "github.com/33cn/chain33/common/db"
	_ "github.com/33cn/chain33/system"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/test/setup"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	tml "github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	configPath          = flag.String("f", "./../../relayer.toml", "configfile")
	ethPrivateKeyStr    = "0x3fa21584ae2e4fd74db9b58e2386f5481607dfa4d7ba0617aaa7858e5025dc1e"
	ethAccountAddr      = "0x92c8b16afd6d423652559c6e266cbe1c29bfd84f"
	chain33ReceiverAddr = "1BCGLhdcdthNutQowV2YShuuN9fJRRGLxu"
	passphrase          = "123456hzj"
	chainTestCfg        = chain33Types.NewChain33Config(chain33Types.GetDefaultCfgstring())
	ethRelayer          *Relayer4Ethereum
)

func init() {
	fmt.Println("======================= init =======================")
	var tx chain33Types.Transaction
	var ret chain33Types.Reply
	ret.IsOk = true

	mockapi := &mocks.QueueProtocolAPI{}
	// 这里对需要mock的方法打桩,Close是必须的，其它方法根据需要
	mockapi.On("Close").Return()
	mockapi.On("AddPushSubscribe", mock.Anything).Return(&ret, nil)
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

	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
	if err != nil {
		panic(err)
	}
	ethRelayer = newEthRelayer(para, sim, x2EthContracts, x2EthDeployInfo)

	time.Sleep(time.Duration(5000) * time.Millisecond)
	simCommit()
}

func Test_All(t *testing.T) {
	fmt.Println("============= test_ShowAddr begin =============")
	test_ShowAddr(t)
	fmt.Println("============= test_GetValidatorAddr begin =============")
	test_GetValidatorAddr(t)
	fmt.Println("============= test_Lock begin =============")
	test_Lock(t)
	fmt.Println("============= test_IsValidatorActive begin =============")
	test_IsValidatorActive(t)
	fmt.Println("============= test_SetBridgeRegistryAddr begin =============")
	test_SetBridgeRegistryAddr(t)
	fmt.Println("============= test_CreateBridgeToken begin =============")
	test_CreateBridgeToken(t)
	fmt.Println("============= test_BurnBty begin =============")
	test_BurnBty(t)
	fmt.Println("============= test_RestorePrivateKeys begin =============")
	test_RestorePrivateKeys(t)
	fmt.Println("============= test_setWithdrawFee begin =============")
	test_setWithdrawFee(t)
}

func Test_remindBalanceNotEnough(t *testing.T) {
	ethRelayer.remindBalanceNotEnough(ethAccountAddr, "YCC", "0x....")
}

func test_GetValidatorAddr(t *testing.T) {
	_, _, err := NewAccount()
	require.Nil(t, err)

	privateKey, _, err := ethRelayer.GetAccount("123")
	require.Nil(t, err)
	assert.NotEqual(t, privateKey, ethPrivateKeyStr)

	privateKey, addr, err := ethRelayer.GetAccount(passphrase)
	require.Nil(t, err)
	assert.Equal(t, privateKey, ethPrivateKeyStr)
	assert.Equal(t, addr, ethAccountAddr)

	validators, err := ethRelayer.GetValidatorAddr()
	require.Nil(t, err)
	assert.Equal(t, validators.EthereumValidator, ethAccountAddr)
	simCommit()
}

func test_Lock(t *testing.T) {
	ctx := context.Background()
	bridgeBankBalanceb, err := ethRelayer.clientSpec.BalanceAt(ctx, ethRelayer.x2EthDeployInfo.BridgeBank.Address, nil)
	require.Nil(t, err)
	simCommit()

	//lock 50 eth
	_, err = ethRelayer.LockEthErc20Asset(hexutil.Encode(crypto.FromECDSA(ethRelayer.deployPara.ValidatorPriKey[1])), "", "50", "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
	require.Nil(t, err)
	simCommit()

	bridgeBankBalance, err := ethRelayer.clientSpec.BalanceAt(ctx, ethRelayer.x2EthDeployInfo.BridgeBank.Address, nil)
	require.Nil(t, err)
	assert.Equal(t, bridgeBankBalance.Int64(), bridgeBankBalanceb.Int64()+int64(50))

	for i := 0; i < int(ethRelayer.maturityDegree+1); i++ {
		simCommit()
	}
	time.Sleep(time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)

	//balance, err := ethRelayer.ShowLockStatics("")
	//require.Nil(t, err)
	//assert.Equal(t, balance, "50")
	//simCommit()
}

func test_IsValidatorActive(t *testing.T) {
	is, err := ethRelayer.IsValidatorActive(ethRelayer.deployPara.InitValidators[0].String())
	assert.Equal(t, is, true)
	require.Nil(t, err)

	is, err = ethRelayer.IsValidatorActive("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF")
	assert.Equal(t, is, false)
	require.Nil(t, err)

	_, err = ethRelayer.IsValidatorActive("123")
	require.Error(t, err)
}

func test_ShowAddr(t *testing.T) {
	//ethRelayer.prePareSubscribeEvent()
	addr, err := ethRelayer.ShowBridgeBankAddr()
	require.Nil(t, err)
	assert.Equal(t, addr, ethRelayer.x2EthDeployInfo.BridgeBank.Address.String())

	addr, err = ethRelayer.ShowBridgeRegistryAddr()
	require.Nil(t, err)
	assert.Equal(t, addr, ethRelayer.x2EthDeployInfo.BridgeRegistry.Address.String())

	addr, err = ethRelayer.ShowOperator()
	require.Nil(t, err)
	assert.Equal(t, addr, ethRelayer.deployPara.Operator.String())
	simCommit()
}

func test_SetBridgeRegistryAddr(t *testing.T) {
	_ = ethRelayer.setBridgeRegistryAddr(ethRelayer.x2EthDeployInfo.BridgeRegistry.Address.String())
	registrAddrInDB, err := ethRelayer.getBridgeRegistryAddr()
	require.Nil(t, err)
	assert.Equal(t, registrAddrInDB, ethRelayer.x2EthDeployInfo.BridgeRegistry.Address.String())
}

//func Test_LockEth(t *testing.T) {
//	ctx := context.Background()
//	bridgeBankBalanceb, err := ethRelayer.clientSpec.BalanceAt(ctx, ethRelayer.x2EthDeployInfo.BridgeBank.Address, nil)
//	require.Nil(t, err)
//
//	userOneAuth, err := ethtxs.PrepareAuth4MultiEthereum(ethRelayer.clientSpec, ethRelayer.deployPara.ValidatorPriKey[0], ethRelayer.deployPara.InitValidators[0], ethRelayer.Addr2TxNonce)
//	require.Nil(t, err)
//
//	//lock 50 eth
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	ethAmount := big.NewInt(50)
//	userOneAuth.Value = ethAmount
//	_, err = ethRelayer.x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethAmount)
//	require.Nil(t, err)
//	simCommit()
//
//	bridgeBankBalance, err := ethRelayer.clientSpec.BalanceAt(ctx, ethRelayer.x2EthDeployInfo.BridgeBank.Address, nil)
//	require.Nil(t, err)
//	assert.Equal(t, bridgeBankBalance.Int64(), bridgeBankBalanceb.Int64()+ethAmount.Int64())
//
//	for i := 0; i < int(ethRelayer.maturityDegree+1); i++ {
//		simCommit()
//	}
//
//	time.Sleep(time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//
//	balance, err := ethRelayer.ShowLockStatics("")
//	require.Nil(t, err)
//	assert.Equal(t, balance, "50")
//
//	time.Sleep(time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
//}

func simCommit() {
	sim, isSim := ethRelayer.clientSpec.(*ethinterface.SimExtend)
	if isSim {
		sim.Commit()
	}
}

func CreateBridgeToken(symbol string, client ethinterface.EthClientSpec, para *ethtxs.OperatorInfo, x2EthDeployInfo *ethtxs.X2EthDeployInfo, x2EthContracts *ethtxs.X2EthContracts, addr2TxNonce map[common.Address]*ethtxs.NonceMutex) (string, error) {
	//订阅事件
	eventName := "LogNewBridgeToken"
	bridgeBankABI := ethtxs.LoadABI(ethtxs.BridgeBankABI)
	logNewBridgeTokenSig := bridgeBankABI.Events[eventName].ID.Hex()
	query := ethereum.FilterQuery{
		Addresses: []common.Address{x2EthDeployInfo.BridgeBank.Address},
	}
	// We will check logs for new events
	logs := make(chan types.Log)
	// Filter by contract and event, write results to logs
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if nil != err {
		fmt.Println("CreateBrigeToken", "failed to SubscribeFilterLogs", err.Error())
		return "", err
	}

	//创建token
	auth, err := ethtxs.PrepareAuth4MultiEthereum(client, para.PrivateKey, para.Address, addr2TxNonce)
	if nil != err {
		return "", err
	}

	_, err = x2EthContracts.BridgeBank.BridgeBankTransactor.CreateNewBridgeToken(auth, symbol)
	if nil != err {
		return "", err
	}

	sim, isSim := client.(*ethinterface.SimExtend)
	if isSim {
		fmt.Println("Use the simulator")
		sim.Commit()
	}

	logEvent := &events.LogNewBridgeToken{}
	select {
	// Handle any errors
	case err := <-sub.Err():
		return "", err
	// vLog is raw event data
	case vLog := <-logs:
		// Check if the event is a 'LogLock' event
		if vLog.Topics[0].Hex() == logNewBridgeTokenSig {
			fmt.Println("CreateBrigeToken", "Witnessed new event", eventName, "Block number", vLog.BlockNumber)

			err = bridgeBankABI.UnpackIntoInterface(logEvent, eventName, vLog.Data)
			if nil != err {
				return "", err
			}
			if symbol != logEvent.Symbol {
				fmt.Println("CreateBrigeToken", "symbol", symbol, "logEvent.Symbol", logEvent.Symbol)
			}
			fmt.Println("CreateBrigeToken", "Witnessed new event", eventName, "Block number", vLog.BlockNumber, "token address", logEvent.Token.String())
			break
		}
	}
	return logEvent.Token.String(), nil
}

func test_CreateBridgeToken(t *testing.T) {
	tokenAddrbty, err := CreateBridgeToken("BTY", ethRelayer.clientSpec, ethRelayer.operatorInfo, ethRelayer.x2EthDeployInfo, ethRelayer.x2EthContracts, ethRelayer.Addr2TxNonce)
	require.Nil(t, err)
	require.NotEmpty(t, tokenAddrbty)
	simCommit()

	addr, err := ethRelayer.ShowTokenAddrBySymbol("BTY")
	require.Nil(t, err)
	assert.Equal(t, addr, tokenAddrbty)

	decimals, err := ethRelayer.GetDecimals(tokenAddrbty)
	require.Nil(t, err)
	assert.Equal(t, decimals, uint8(8))

	_, err = ethRelayer.Burn(ethRelayer.deployPara.InitValidators[0].String(), tokenAddrbty, chain33ReceiverAddr, "10")
	require.Error(t, err)
	simCommit()

	_, err = ethRelayer.BurnAsync(ethRelayer.deployPara.InitValidators[0].String(), tokenAddrbty, chain33ReceiverAddr, "10")
	require.Error(t, err)
	simCommit()
}

func test_BurnBty(t *testing.T) {
	tokenAddrbty, err := CreateBridgeToken("bty", ethRelayer.clientSpec, ethRelayer.operatorInfo, ethRelayer.x2EthDeployInfo, ethRelayer.x2EthContracts, ethRelayer.Addr2TxNonce)
	require.Nil(t, err)
	require.NotEmpty(t, tokenAddrbty)
	simCommit()

	symbol := &ebTypes.TokenAddress{Symbol: "bty"}
	token, err := ethRelayer.ShowTokenAddress(symbol)
	require.Nil(t, err)
	require.Equal(t, token.TokenAddress[0].Address, tokenAddrbty)

	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
	amount := int64(100)
	ethReceiver := ethRelayer.deployPara.InitValidators[2]
	claimID := crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), big.NewInt(amount).Bytes())
	authOracle, err := ethtxs.PrepareAuth4MultiEthereum(ethRelayer.clientSpec, ethRelayer.deployPara.ValidatorPriKey[0], ethRelayer.deployPara.InitValidators[0], ethRelayer.Addr2TxNonce)
	require.Nil(t, err)
	signature, err := utils.SignClaim4Evm(claimID, ethRelayer.deployPara.ValidatorPriKey[0])
	require.Nil(t, err)

	_, err = ethRelayer.x2EthContracts.Oracle.NewOracleClaim(
		authOracle,
		uint8(events.ClaimTypeLock),
		chain33Sender,
		ethReceiver,
		common.HexToAddress(tokenAddrbty),
		"bty",
		big.NewInt(amount),
		claimID,
		signature)
	require.Nil(t, err)
	simCommit()

	balanceNew, err := ethRelayer.GetBalance(tokenAddrbty, ethReceiver.String())
	require.Nil(t, err)
	require.Equal(t, balanceNew, "100")

	_, err = ethRelayer.Burn(hexutil.Encode(crypto.FromECDSA(ethRelayer.deployPara.ValidatorPriKey[2])), tokenAddrbty, chain33ReceiverAddr, "10")
	require.NoError(t, err)
	simCommit()

	balanceNew, err = ethRelayer.GetBalance(tokenAddrbty, ethReceiver.String())
	require.Nil(t, err)
	require.Equal(t, balanceNew, "90")

	// ApproveAllowance
	{
		ownerPrivateKey, err := crypto.ToECDSA(common.FromHex(hexutil.Encode(crypto.FromECDSA(ethRelayer.deployPara.ValidatorPriKey[2]))))
		require.Nil(t, err)
		ownerAddr := crypto.PubkeyToAddress(ownerPrivateKey.PublicKey)
		auth, err := ethtxs.PrepareAuth4MultiEthereum(ethRelayer.clientSpec, ownerPrivateKey, ownerAddr, ethRelayer.Addr2TxNonce)
		require.Nil(t, err)

		erc20TokenInstance, err := generated.NewBridgeToken(common.HexToAddress(tokenAddrbty), ethRelayer.clientSpec)
		require.Nil(t, err)

		bn := big.NewInt(1)
		bn, _ = bn.SetString(utils.TrimZeroAndDot("10"), 10)
		_, err = erc20TokenInstance.Approve(auth, ethRelayer.x2EthDeployInfo.BridgeBank.Address, bn)
		require.Nil(t, err)

		simCommit()
	}

	_, err = ethRelayer.BurnAsync(hexutil.Encode(crypto.FromECDSA(ethRelayer.deployPara.ValidatorPriKey[2])), tokenAddrbty, chain33ReceiverAddr, "10")
	require.NoError(t, err)
	simCommit()

	balanceNew, err = ethRelayer.GetBalance(tokenAddrbty, ethReceiver.String())
	require.Nil(t, err)
	require.Equal(t, balanceNew, "80")

	fetchCnt := int32(10)
	logs, err := ethRelayer.getNextValidEthTxEventLogs(ethRelayer.eventLogIndex.Height, ethRelayer.eventLogIndex.Index, fetchCnt)
	require.NoError(t, err)
	fmt.Println("logs", logs)
	simCommit()

	for _, vLog := range logs {
		fmt.Println("*vLog", *vLog)
		ethRelayer.procBridgeBankLogs(*vLog)
	}
	simCommit()
}

func test_RestorePrivateKeys(t *testing.T) {
	_, err := ethRelayer.ImportPrivateKey(passphrase, ethPrivateKeyStr)
	require.Nil(t, err)
	time.Sleep(time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
	simCommit()

	go func() {
		for range ethRelayer.unlockchan {
		}
	}()
	ethRelayer.rwLock.RLock()
	temp := ethRelayer.privateKey4Ethereum
	ethRelayer.rwLock.RUnlock()

	err = ethRelayer.RestorePrivateKeys("123")
	ethRelayer.rwLock.RLock()
	assert.NotEqual(t, common.Bytes2Hex(crypto.FromECDSA(temp)), common.Bytes2Hex(crypto.FromECDSA(ethRelayer.privateKey4Ethereum)))
	ethRelayer.rwLock.RUnlock()
	require.Nil(t, err)

	err = ethRelayer.RestorePrivateKeys(passphrase)
	ethRelayer.rwLock.RLock()
	assert.Equal(t, common.Bytes2Hex(crypto.FromECDSA(temp)), common.Bytes2Hex(crypto.FromECDSA(ethRelayer.privateKey4Ethereum)))
	ethRelayer.rwLock.RUnlock()
	require.Nil(t, err)

	err = ethRelayer.StoreAccountWithNewPassphase("new123", passphrase)
	require.Nil(t, err)

	err = ethRelayer.RestorePrivateKeys("new123")
	ethRelayer.rwLock.RLock()
	assert.Equal(t, common.Bytes2Hex(crypto.FromECDSA(temp)), common.Bytes2Hex(crypto.FromECDSA(ethRelayer.privateKey4Ethereum)))
	ethRelayer.rwLock.RUnlock()
	require.Nil(t, err)
	simCommit()
}

func test_setWithdrawFee(t *testing.T) {
	WithdrawPara := make(map[string]*ebTypes.WithdrawPara)
	WithdrawPara["ETH"] = &ebTypes.WithdrawPara{
		Fee:          "0.5",
		AmountPerDay: "100",
	}
	err := ethRelayer.setWithdrawFee(WithdrawPara)
	require.Nil(t, err)

	WithdrawPara = ethRelayer.restoreWithdrawFee()
	assert.Equal(t, WithdrawPara["ETH"].Fee, "0.5")
	assert.Equal(t, WithdrawPara["ETH"].AmountPerDay, "100")
	simCommit()
}

func newEthRelayer(para *ethtxs.DeployPara, sim *ethinterface.SimExtend, x2EthContracts *ethtxs.X2EthContracts, x2EthDeployInfo *ethtxs.X2EthDeployInfo) *Relayer4Ethereum {
	cfg := initCfg(*configPath)
	cfg.Chain33RelayerCfg.SyncTxConfig.Chain33Host = "http://127.0.0.1:8801"
	cfg.EthRelayerCfg[0].BridgeRegistry = x2EthDeployInfo.BridgeRegistry.Address.String()
	cfg.Chain33RelayerCfg.SyncTxConfig.PushBind = "127.0.0.1:60000"
	cfg.Chain33RelayerCfg.SyncTxConfig.FetchHeightPeriodMs = 50
	cfg.Dbdriver = "memdb"
	cfg.DbPath = "datadirEth"

	db := dbm.NewDB("relayer_db_service", cfg.Dbdriver, cfg.DbPath, cfg.DbCache)
	ethBridgeClaimchan := make(chan *relayerTypes.EthBridgeClaim, 100)
	chain33Msgchan := make(chan *events.Chain33Msg, 100)

	relayer := &Relayer4Ethereum{
		name:                    cfg.EthRelayerCfg[0].EthChainName,
		provider:                cfg.EthRelayerCfg[0].EthProvider,
		providerHttp:            cfg.EthRelayerCfg[0].EthProviderCli,
		db:                      db,
		unlockchan:              make(chan int, 2),
		bridgeRegistryAddr:      x2EthDeployInfo.BridgeRegistry.Address,
		maturityDegree:          1,
		fetchHeightPeriodMs:     1,
		totalTxRelayFromChain33: 0,
		symbol2Addr:             make(map[string]common.Address),
		symbol2LockAddr:         make(map[string]*ebTypes.TokenAddress),
		ethBridgeClaimChan:      ethBridgeClaimchan,
		chain33MsgChan:          chain33Msgchan,
		Addr2TxNonce:            make(map[common.Address]*ethtxs.NonceMutex),
		//remindUrl:               cfg.RemindUrl,
	}

	relayer.eventLogIndex = relayer.getLastBridgeBankProcessedHeight()
	relayer.initBridgeBankTx()
	relayer.clientSpec = sim
	relayer.clientWss = sim
	relayer.clientChainID = big.NewInt(1337)

	relayer.rwLock.Lock()
	relayer.operatorInfo = &ethtxs.OperatorInfo{
		PrivateKey: para.DeployPrivateKey,
		Address:    para.Deployer,
	}
	relayer.deployPara = para
	relayer.x2EthContracts = x2EthContracts
	relayer.x2EthDeployInfo = x2EthDeployInfo
	relayer.rwLock.Unlock()

	relayer.totalTxRelayFromChain33 = relayer.getTotalTxAmount2Eth()

	relayer.rwLock.Lock()
	_, err := relayer.ImportPrivateKey(passphrase, ethPrivateKeyStr)
	relayer.rwLock.Unlock()
	if err != nil {
		panic(err)
	}
	//go relayer.proc()
	go proc(relayer)
	return relayer
}

func proc(ethRelayer *Relayer4Ethereum) {
	//等待用户导入
	relayerLog.Info("Please unlock or import private key for Ethereum relayer")
	if err := ethRelayer.RestoreTokenAddress(); nil != err {
		relayerLog.Info("Failed to RestoreTokenAddress")
		return
	}

	nilAddr := common.Address{}
	if nilAddr != ethRelayer.bridgeRegistryAddr {
		relayerLog.Info("proc", "Going to recover corresponding solidity contract handler with bridgeRegistryAddr", ethRelayer.bridgeRegistryAddr.String())
		var err error
		ethRelayer.rwLock.Lock()
		ethRelayer.x2EthContracts, ethRelayer.x2EthDeployInfo, err = ethtxs.RecoverContractHandler(ethRelayer.clientSpec, ethRelayer.bridgeRegistryAddr, ethRelayer.bridgeRegistryAddr)
		if nil != err {
			panic("Failed to recover corresponding solidity contract handler due to:" + err.Error())
		}
		ethRelayer.rwLock.Unlock()
		relayerLog.Info("^-^ ^-^ Succeed to recover corresponding solidity contract handler")

		ethRelayer.unlockchan <- start
	}

	var timer *time.Ticker
	for range ethRelayer.unlockchan {
		relayerLog.Info("Received ethRelayer.unlockchan")
		ethRelayer.rwLock.RLock()
		privateKey4Ethereum := ethRelayer.privateKey4Ethereum
		ethRelayer.rwLock.RUnlock()
		if nil != privateKey4Ethereum && nilAddr != ethRelayer.bridgeRegistryAddr {
			relayerLog.Info("Ethereum relayer starts to run...")
			ethRelayer.prePareSubscribeEvent()
			//向bridgeBank订阅事件
			ethRelayer.subscribeEvent()
			ethRelayer.filterLogEvents()
			relayerLog.Info("Ethereum relayer starts to process online log event...")
			timer = time.NewTicker(time.Duration(ethRelayer.fetchHeightPeriodMs) * time.Millisecond)
			break
		}
	}

	for {
		select {
		case <-timer.C:
			ethRelayer.procNewHeight()
		case err := <-ethRelayer.bridgeBankSub.Err():
			relayerLog.Error("proc", "Need to subscribeEvent again due to bridgeBankSub err", err.Error())
			ethRelayer.subscribeEvent()
			ethRelayer.filterLogEvents()
		case vLog := <-ethRelayer.bridgeBankLog:
			ethRelayer.storeBridgeBankLogs(vLog, true)
		case chain33Msg := <-ethRelayer.chain33MsgChan:
			ethRelayer.handleChain33Msg(chain33Msg)
		case txRelayAck := <-ethRelayer.txRelayAckRecvChan:
			ethRelayer.procTxRelayAck(txRelayAck)
		}
	}
}

func initCfg(path string) *relayerTypes.RelayerConfig {
	var cfg relayerTypes.RelayerConfig
	if _, err := tml.DecodeFile(path, &cfg); err != nil {
		os.Exit(-1)
	}
	return &cfg
}

func Test_UnpackLogProphecyProcessed(t *testing.T) {
	eventData := []byte{121, 110, 255, 239, 36, 105, 91, 194, 159, 116, 120, 172, 247, 183, 65, 84, 137, 248, 222, 154, 153, 21, 31, 44, 217, 190, 53, 63, 120, 15, 86, 234, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 225, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 200, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 217, 218, 176, 33, 231, 78, 207, 71, 87, 136, 237, 123, 97, 53, 96, 86, 178, 9, 88, 48}

	log, err := events.UnpackLogProphecyProcessed(ethtxs.LoadABI(ethtxs.OracleABI), events.LogProphecyProcessed.String(), eventData)
	require.Nil(t, err)

	claimID := hexutil.Encode(log.ClaimID[:])
	require.NotEmpty(t, claimID)
}
