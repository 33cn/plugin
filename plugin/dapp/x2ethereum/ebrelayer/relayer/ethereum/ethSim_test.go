package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/test/setup"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
)

type suiteEthRelayerSim struct {
	suite.Suite
	ethRelayer      *Relayer4Ethereum
	sim             *backends.SimulatedBackend
	x2EthContracts  *ethtxs.X2EthContracts
	x2EthDeployInfo *ethtxs.X2EthDeployInfo
	para            *ethtxs.DeployPara
	backend         bind.ContractBackend
}

func TestRunSuiteX2EthereumSim(t *testing.T) {
	log := new(suiteEthRelayerSim)
	suite.Run(t, log)
}

func (r *suiteEthRelayerSim) SetupSuite() {
	r.deploySimContracts()
	r.ethRelayer = r.newSimEthRelayer()
}

func (r *suiteEthRelayerSim) TestSim_1_ImportPrivateKey() {
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

func (r *suiteEthRelayerSim) TestSim_2_RestorePrivateKeys() {
	// 错误的密码 也不报错
	err := r.ethRelayer.RestorePrivateKeys(passphrase)
	r.NoError(err)

	err = r.ethRelayer.StoreAccountWithNewPassphase(passphrase, passphrase)
	r.NoError(err)
}

func (r *suiteEthRelayerSim) TestSim_3_IsValidatorActive() {
	is, err := r.ethRelayer.IsValidatorActive("0x92c8b16afd6d423652559c6e266cbe1c29bfd84f")
	r.Equal(is, true)
	r.NoError(err)

	is, err = r.ethRelayer.IsValidatorActive("0x0C05bA5c230fDaA503b53702aF1962e08D0C60BF")
	r.Equal(is, false)
	r.NoError(err)

	is, err = r.ethRelayer.IsValidatorActive("123")
	r.Error(err)
}

func (r *suiteEthRelayerSim) Test_4_DeployContrcts() {
	_, err := r.ethRelayer.DeployContrcts()
	r.Error(err)
}

//func (r *suiteEthRelayerSim) TestSim_4_LockEth() {
//	ctx := context.Background()
//	bridgeBankBalance, err := r.sim.BalanceAt(ctx, r.x2EthDeployInfo.BridgeBank.Address, nil)
//	fmt.Println(bridgeBankBalance, err)
//	r.NoError(err)
//
//	userOneAuth, err := ethtxs.PrepareAuth(r.backend, r.para.ValidatorPriKey[0], r.para.InitValidators[0])
//	r.NoError(err)
//	ethAmount := big.NewInt(50)
//	userOneAuth.Value = ethAmount
//
//	//lock 50 eth
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	_, err = r.x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethAmount)
//	r.NoError(err)
//	r.sim.Commit()
//
//	bridgeBankBalance, err = r.sim.BalanceAt(ctx, r.x2EthDeployInfo.BridgeBank.Address, nil)
//	r.NoError(err)
//	//require.Equal(t, bridgeBankBalance.Int64(), ethAmount.Int64())
//	fmt.Println(bridgeBankBalance, err)
//	//
//	///*	ctx := context.Background()
//	//	auth, err := ethtxs.PrepareAuth(r.backend, r.para.DeployPrivateKey, r.para.Operator)
//	//	r.NoError(err)
//	//
//	//	opts := &bind.CallOpts{
//	//		Pending: true,
//	//		From:    r.para.Operator,
//	//		Context: ctx,
//	//	}
//	//
//	//	tokenCount, err := r.x2EthContracts.BridgeBank.BridgeTokenCount(opts)
//	//	r.Equal(tokenCount.Int64(), int64(0))
//	//	//创建token
//	//	symbol := "bty"
//	//	_, err = r.ethRelayer.x2EthContracts.BridgeBank.BridgeBankTransactor.CreateNewBridgeToken(auth, symbol)
//	//	r.NoError(err)
//	//	r.sim.Commit()
//	//
//	//	eventName := "LogNewBridgeToken"
//	//	bridgeBankABI := ethtxs.LoadABI(ethtxs.BridgeBankABI)
//	//	logNewBridgeTokenSig := bridgeBankABI.Events[eventName].ID().Hex()
//	//
//	//	logEvent := &events.LogNewBridgeToken{}
//	//	select {
//	//	// Handle any errors
//	//	case err := <-r.ethRelayer.bridgeBankSub.Err():
//	//		r.NoError(err)
//	//	// vLog is raw event data
//	//	case vLog := <-r.ethRelayer.bridgeBankLog:
//	//		// Check if the event is a 'LogLock' event
//	//		if vLog.Topics[0].Hex() == logNewBridgeTokenSig {
//	//			//_ = fmt.Sprint("Witnessed new event:%s, Block number:%d, Tx hash:%s", eventName, vLog.BlockNumber, vLog.TxHash.Hex())
//	//
//	//			err = bridgeBankABI.Unpack(logEvent, eventName, vLog.Data)
//	//			r.NoError(err)
//	//			//_ = fmt.Sprint("token addr:%s, symbol:%s", logEvent.Token.String(), logEvent.Symbol)
//	//			r.Equal(symbol, logEvent.Symbol)
//	//
//	//			//tokenCount正确加1
//	//			tokenCount, err = r.x2EthContracts.BridgeBank.BridgeTokenCount(opts)
//	//			r.Equal(tokenCount.Int64(), int64(1))
//	//			break
//	//		}
//	//	}
//	//
//	//	///////////newOracleClaim///////////////////////////
//	//	balance, _ := r.sim.BalanceAt(ctx, r.para.InitValidators[0], nil)
//	//	fmt.Println("InitValidators[0] addr,", r.para.InitValidators[0].String(), "balance =", balance.String())
//	//
//	//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	//	amount := int64(99)
//	//	ethReceiver := r.para.InitValidators[2]
//	//	claimID := crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), logEvent.Token.Bytes(), big.NewInt(amount).Bytes())
//	//
//	//	authOracle, err := ethtxs.PrepareAuth(r.backend, r.para.ValidatorPriKey[0], r.para.InitValidators[0])
//	//	r.NoError(err)
//	//
//	//	signature, err := ethtxs.SignClaim4Eth(claimID, r.para.ValidatorPriKey[0])
//	//	r.NoError(err)
//	//
//	//	bridgeToken, err := generated.NewBridgeToken(logEvent.Token, r.backend)
//	//	r.NoError(err)
//	//	opts = &bind.CallOpts{
//	//		Pending: true,
//	//		Context: ctx,
//	//	}
//	//
//	//	balance, err = bridgeToken.BalanceOf(opts, ethReceiver)
//	//	r.NoError(err)
//	//	r.Equal(balance.Int64(), int64(0))
//	//
//	//	_, err = r.x2EthContracts.Oracle.NewOracleClaim(
//	//		authOracle,
//	//		events.ClaimTypeLock,
//	//		chain33Sender,
//	//		ethReceiver,
//	//		logEvent.Token,
//	//		logEvent.Symbol,
//	//		big.NewInt(amount),
//	//		claimID,
//	//		signature)
//	//	r.NoError(err)
//	//
//	//	r.sim.Commit()
//	//	balance, err = bridgeToken.BalanceOf(opts, ethReceiver)
//	//	r.NoError(err)
//	//	r.Equal(balance.Int64(), amount)
//	//	fmt.Println("the minted amount is:", balance.Int64())
//	//*/
//	time.Sleep(1 * time.Second)
//
//	query := ethereum.FilterQuery{
//		Addresses: []common.Address{r.ethRelayer.bridgeBankAddr},
//	}
//	logs, err := r.sim.FilterLogs(context.Background(), query)
//	if err != nil {
//		errinfo := fmt.Sprintf("Failed to filterLogEvents due to:%s", err.Error())
//		fmt.Println(errinfo)
//	}
//
//	for _, logv := range logs {
//		if err := r.ethRelayer.setEthTxEvent(logv); nil != err {
//			//	panic(err.Error())
//		}
//	}
//
//	time.Sleep(5 * time.Second)
//}

//func (r *suiteEthRelayerSim) TestRelayer4Ethereum_ApproveAllowance(t *testing.T) {
//	r.ethRelayer.ApproveAllowance()
//}

func (r *suiteEthRelayerSim) newSimEthRelayer() *Relayer4Ethereum {
	cfg := initCfg(*configPath)
	cfg.SyncTxConfig.Chain33Host = "http://127.0.0.1:8801"
	cfg.BridgeRegistry = r.x2EthDeployInfo.BridgeRegistry.Address.String()
	cfg.SyncTxConfig.PushBind = "127.0.0.1:60000"
	cfg.SyncTxConfig.FetchHeightPeriodMs = 50
	cfg.SyncTxConfig.Dbdriver = "memdb"
	cfg.SyncTxConfig.DbPath = "datadirSim"

	db := dbm.NewDB("relayer_db_service_sim", cfg.SyncTxConfig.Dbdriver, cfg.SyncTxConfig.DbPath, cfg.SyncTxConfig.DbCache)

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
		relayerLog.Error("StartEthereumRelayer", "BridgeRegistry is setted already with value", registrAddrInDB, "but now setting to", cfg.BridgeRegistry)
		_ = relayer.setBridgeRegistryAddr(cfg.BridgeRegistry)
	} else if cfg.BridgeRegistry == "" && registrAddrInDB != "" {
		//输入地址为空，且数据库中保存地址不为空，则直接使用数据库中的地址
		relayer.bridgeRegistryAddr = common.HexToAddress(registrAddrInDB)
	}
	relayer.eventLogIndex = relayer.getLastBridgeBankProcessedHeight()
	relayer.initBridgeBankTx()

	go r.procSim()

	return relayer
}

func (r *suiteEthRelayerSim) procSim() {
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
			r.procNewHeight(ctx, &continueFailCount)
		case err := <-r.ethRelayer.bridgeBankSub.Err():
			panic("bridgeBankSub" + err.Error())
		case vLog := <-r.ethRelayer.bridgeBankLog:
			r.ethRelayer.storeBridgeBankLogs(vLog, true)
		}
	}
}

func (r *suiteEthRelayerSim) deploySimContracts() {
	ctx := context.Background()
	r.backend, r.para = setup.PrepareTestEnvironment()
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

func (r *suiteEthRelayerSim) filterLogEvents() {
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

func (r *suiteEthRelayerSim) procNewHeight(ctx context.Context, continueFailCount *int32) {
	*continueFailCount = 0
	currentHeight := uint64(20)
	relayerLog.Info("procNewHeight", "currentHeight", currentHeight)
	//一次最大只获取10个logEvent进行处理
	fetchCnt := int32(10)
	for r.ethRelayer.eventLogIndex.Height+uint64(r.ethRelayer.maturityDegree)+1 <= currentHeight {
		logs, err := r.ethRelayer.getNextValidEthTxEventLogs(r.ethRelayer.eventLogIndex.Height, r.ethRelayer.eventLogIndex.Index, fetchCnt)
		if nil != err {
			relayerLog.Error("Failed to get ethereum height", "getNextValidEthTxEventLogs err", err.Error())
			return
		}

		for i, vLog := range logs {
			if vLog.BlockNumber+uint64(r.ethRelayer.maturityDegree)+1 > currentHeight {
				logs = logs[:i]
				break
			}
			//r.ethRelayer.procBridgeBankLogs(*vLog)
			if r.ethRelayer.checkTxProcessed(vLog.TxHash.Bytes()) {
				relayerLog.Info("procBridgeBankLogs", "Tx has been already Processed with hash:", vLog.TxHash.Hex(),
					"height", vLog.BlockNumber, "index", vLog.Index)
				return
			}

			defer func() {
				if err := r.ethRelayer.setTxProcessed(vLog.TxHash.Bytes()); nil != err {
					panic(err.Error())
				}
			}()
			//lock,用于捕捉 (ETH/ERC20----->chain33) 跨链转移
			if vLog.Topics[0].Hex() == r.ethRelayer.bridgeBankEventLockSig {
				eventName := events.LogLock.String()
				relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
					"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
				err := r.ethRelayer.handleLogLockEvent(r.ethRelayer.clientChainID, r.ethRelayer.bridgeBankAbi, eventName, *vLog)
				if err != nil {
					errinfo := fmt.Sprintf("Failed to handleLogLockEvent due to:%s", err.Error())
					relayerLog.Info("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
					panic(errinfo)
				}
			} else if vLog.Topics[0].Hex() == r.ethRelayer.bridgeBankEventBurnSig {
				//burn,用于捕捉 (chain33 token----->chain33) 实现chain33资产withdraw操作，之后在chain33上实现unlock操作
				eventName := events.LogChain33TokenBurn.String()
				relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
					"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
				err := r.ethRelayer.handleLogBurnEvent(r.ethRelayer.clientChainID, r.ethRelayer.bridgeBankAbi, eventName, *vLog)
				if err != nil {
					errinfo := fmt.Sprintf("Failed to handleLogBurnEvent due to:%s", err.Error())
					relayerLog.Info("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
					panic(errinfo)
				}
			}
		}

		cnt := int32(len(logs))
		if len(logs) > 0 {
			//firstHeight := logs[0].BlockNumber
			lastHeight := logs[cnt-1].BlockNumber
			index := logs[cnt-1].TxIndex
			//获取的数量小于批量获取数量，则认为直接
			r.ethRelayer.setBridgeBankProcessedHeight(lastHeight, uint32(index))
			r.ethRelayer.eventLogIndex.Height = lastHeight
			r.ethRelayer.eventLogIndex.Index = uint32(index)
		}

		//当前需要处理的event数量已经少于10个，直接返回
		if cnt < fetchCnt {
			return
		}
	}
}
