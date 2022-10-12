package ethereum

// -----------------------------------------------------
//      Ethereum relayer
//
//      Initializes the relayer service, which parses,
//      encodes, and packages named events on an Ethereum
//      Smart Contract for validator's to sign and send
//      to the Chain33 bridge.
// -----------------------------------------------------

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	chain33Common "github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	cross2ethErrors "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	"github.com/bitly/go-simplejson"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

//Relayer4Ethereum ...
type Relayer4Ethereum struct {
	name               string //链的名字，用于区分不同的链
	provider           []string
	providerHttp       []string
	clientChainID      *big.Int
	bridgeRegistryAddr common.Address
	db                 dbm.DB
	rwLock             sync.RWMutex

	privateKey4Ethereum *ecdsa.PrivateKey
	ethSender           common.Address
	processWithDraw     bool

	unlockchan              chan int
	maturityDegree          int32
	fetchHeightPeriodMs     int32
	eventLogIndex           ebTypes.EventLogIndex
	clientSpec              ethinterface.EthClientSpec
	clientUrlSelected       string
	oracleInstance          *generated.Oracle
	clientSpecs             []*ethtxs.EthClientWithUrl
	clientBSCRecommendSpecs []*ethtxs.EthClientWithUrl
	clientWss               ethinterface.EthClientSpec
	bridgeBankAddr          common.Address
	bridgeBankSub           ethereum.Subscription
	bridgeBankLog           chan types.Log
	bridgeBankEventLockSig  string
	bridgeBankEventBurnSig  string
	bridgeBankAbi           abi.ABI
	oracleAddr              common.Address
	oracleEventSig          string
	oracleAbi               abi.ABI
	x2EthDeployInfo         *ethtxs.X2EthDeployInfo
	deployPara              *ethtxs.DeployPara
	operatorInfo            *ethtxs.OperatorInfo
	x2EthContracts          *ethtxs.X2EthContracts
	ethBridgeClaimChan      chan<- *ebTypes.EthBridgeClaim
	txRelayAckSendChan      chan<- *ebTypes.TxRelayAck
	txRelayAckRecvChan      <-chan *ebTypes.TxRelayAck
	chain33MsgChan          <-chan *events.Chain33Msg
	totalTxRelayFromChain33 int64
	symbol2Addr             map[string]common.Address
	symbol2LockAddr         map[string]*ebTypes.TokenAddress
	mulSignAddr             string
	withdrawFee             map[string]*WithdrawFeeAndQuota
	Addr2TxNonce            map[common.Address]*ethtxs.NonceMutex
	startListenHeight       int64
	remindUrl               string   // 代理打币地址金额不够时发生提醒短信的 url
	remindClientErrorUrl    string   // BSC or ethereum 节点出错时邮件提醒的 url
	remindEmail             []string // 提醒的邮箱
	delayedSend             bool     // 是否延迟发送ethereum交易, 4个中继器中设置3个为false, 1个为true, 延迟发送burn交易, 过3分钟查看ethereum是否已经执行, 如果已经执行, 就不再发送burn交易, 节约手续费
}

var (
	relayerLog = log.New("module", "ethereum_relayer")
	// BSCRecommendHttp BSC 官方节点
	BSCRecommendHttp = []string{"https://bsc-dataseed.binance.org/", "https://bsc-dataseed1.defibit.io/", "https://bsc-dataseed1.ninicoin.io/"}
)

const (
	DefaultBlockPeriod = 5000
	waitTime           = time.Second * 30
	sleepTime          = time.Second * 10
	//EthereumChain      = "Ethereum"
	//USDT               = "USDT"
)

type EthereumStartPara struct {
	DbHandle             dbm.DB
	EthProvider          []string
	EthProviderHttp      []string
	BridgeRegistryAddr   string
	Degree               int32
	BlockInterval        int32
	EthBridgeClaimChan   chan<- *ebTypes.EthBridgeClaim
	TxRelayAckSendChan   chan<- *ebTypes.TxRelayAck
	TxRelayAckRecvChan   <-chan *ebTypes.TxRelayAck
	Chain33MsgChan       <-chan *events.Chain33Msg
	ProcessWithDraw      bool
	DelayedSend          bool
	Name                 string
	StartListenHeight    int64
	RemindUrl            string
	RemindClientErrorUrl string
	RemindEmail          []string
}

type WithdrawFeeAndQuota struct {
	Fee          *big.Int
	AmountPerDay *big.Int
}

//StartEthereumRelayer ///
func StartEthereumRelayer(startPara *EthereumStartPara) *Relayer4Ethereum {
	if 0 == startPara.BlockInterval {
		startPara.BlockInterval = DefaultBlockPeriod
	}
	ethRelayer := &Relayer4Ethereum{
		name:                    startPara.Name,
		provider:                make([]string, len(startPara.EthProvider)),
		providerHttp:            make([]string, len(startPara.EthProviderHttp)),
		db:                      startPara.DbHandle,
		unlockchan:              make(chan int, 2),
		bridgeRegistryAddr:      common.HexToAddress(startPara.BridgeRegistryAddr),
		processWithDraw:         startPara.ProcessWithDraw,
		delayedSend:             startPara.DelayedSend,
		maturityDegree:          startPara.Degree,
		fetchHeightPeriodMs:     startPara.BlockInterval,
		ethBridgeClaimChan:      startPara.EthBridgeClaimChan,
		txRelayAckSendChan:      startPara.TxRelayAckSendChan,
		txRelayAckRecvChan:      startPara.TxRelayAckRecvChan,
		chain33MsgChan:          startPara.Chain33MsgChan,
		totalTxRelayFromChain33: 0,
		symbol2Addr:             make(map[string]common.Address),
		symbol2LockAddr:         make(map[string]*ebTypes.TokenAddress),
		Addr2TxNonce:            make(map[common.Address]*ethtxs.NonceMutex),
		remindUrl:               startPara.RemindUrl,
		remindClientErrorUrl:    startPara.RemindClientErrorUrl,
		remindEmail:             startPara.RemindEmail,
		startListenHeight:       startPara.StartListenHeight,
	}
	copy(ethRelayer.provider, startPara.EthProvider)
	copy(ethRelayer.providerHttp, startPara.EthProviderHttp)

	registrAddrInDB, err := ethRelayer.getBridgeRegistryAddr()
	//如果输入的registry地址非空，且和数据库保存地址不一致，则直接使用输入注册地址
	if startPara.BridgeRegistryAddr != "" && nil == err && registrAddrInDB != startPara.BridgeRegistryAddr {
		relayerLog.Error("StartEthereumRelayer", "BridgeRegistry is setted already with value", registrAddrInDB, "but now setting to", startPara.BridgeRegistryAddr)
		_ = ethRelayer.setBridgeRegistryAddr(startPara.BridgeRegistryAddr)
	} else if startPara.BridgeRegistryAddr == "" && registrAddrInDB != "" {
		//输入地址为空，且数据库中保存地址不为空，则直接使用数据库中的地址
		ethRelayer.bridgeRegistryAddr = common.HexToAddress(registrAddrInDB)
	}
	ethRelayer.eventLogIndex = ethRelayer.getLastBridgeBankProcessedHeight()
	ethRelayer.initBridgeBankTx()
	ethRelayer.mulSignAddr = ethRelayer.getMultiSignAddress()
	ethRelayer.withdrawFee = ethRelayer.restoreWithdrawFeeInINt()

	ethRelayer.totalTxRelayFromChain33 = ethRelayer.getTotalTxAmount2Eth()
	if 0 == ethRelayer.totalTxRelayFromChain33 {
		statics := &ebTypes.Chain33ToEthereumStatics{}
		data := chain33Types.Encode(statics)
		_ = ethRelayer.setLastestStatics(int32(events.ClaimTypeLock), 0, data)
		_ = ethRelayer.setLastestStatics(int32(events.ClaimTypeBurn), 0, data)
		_ = ethRelayer.setLastestStatics(int32(events.ClaimTypeWithdraw), 0, data)
	}

	go ethRelayer.proc()
	return ethRelayer
}

// ShowBalanceLocked 获取某一个币种的余额
func (ethRelayer *Relayer4Ethereum) ShowBalanceLocked(tokenAddr, bridgeBank string) (string, error) {
	bridgeBankAddrInt := common.HexToAddress(bridgeBank)
	bridgeBankHandle, err := generated.NewBridgeBank(bridgeBankAddrInt, ethRelayer.clientSpec)
	if nil != err {
		return "", errors.New("failed to NewBridgeBank")
	}
	opts := &bind.CallOpts{
		Pending: true,
		From:    common.HexToAddress(bridgeBank),
		Context: context.Background(),
	}
	balance, err := bridgeBankHandle.LockedFunds(opts, common.HexToAddress(tokenAddr))
	if nil != err {
		return "", err
	}

	return balance.String(), nil
}

func (ethRelayer *Relayer4Ethereum) GetBalance(tokenAddr, owner string) (string, error) {
	return ethtxs.GetBalance(ethRelayer.clientSpec, tokenAddr, owner)
}

func (ethRelayer *Relayer4Ethereum) ShowMultiBalance(tokenAddr, owner string) (string, error) {
	relayerLog.Info("ShowMultiBalance", "tokenAddr", tokenAddr, "owner", owner)
	opts := &bind.CallOpts{
		From:    ethRelayer.ethSender,
		Context: context.Background(),
	}

	gnosisSafeAddr := common.HexToAddress(ethRelayer.mulSignAddr)
	gnosisSafeInt, err := gnosis.NewGnosisSafe(gnosisSafeAddr, ethRelayer.clientSpec)
	if nil != err {
		return "", err
	}

	balance, err := gnosisSafeInt.GetSelfBalance(opts)
	if nil != err {
		return "", err
	}
	return balance.String(), nil
}

//ShowBridgeBankAddr ...
func (ethRelayer *Relayer4Ethereum) ShowBridgeBankAddr() (string, error) {
	if nil == ethRelayer.x2EthDeployInfo {
		return "", errors.New("the relayer is not started yes")
	}

	return ethRelayer.x2EthDeployInfo.BridgeBank.Address.String(), nil
}

//ShowBridgeRegistryAddr ...
func (ethRelayer *Relayer4Ethereum) ShowBridgeRegistryAddr() (string, error) {
	if nil == ethRelayer.x2EthDeployInfo {
		return "", errors.New("the relayer is not started yes")
	}

	return ethRelayer.x2EthDeployInfo.BridgeRegistry.Address.String(), nil
}

//ShowLockStatics ...
func (ethRelayer *Relayer4Ethereum) ShowLockStatics(tokenAddr string) (string, error) {
	return ethtxs.GetLockedFunds(ethRelayer.x2EthContracts.BridgeBank, tokenAddr)
}

//ShowDepositStatics ...
func (ethRelayer *Relayer4Ethereum) ShowDepositStatics(tokenAddr string) (string, error) {
	return ethtxs.GetDepositFunds(ethRelayer.clientSpec, tokenAddr)
}

//ShowTokenAddrBySymbol ...
func (ethRelayer *Relayer4Ethereum) ShowTokenAddrBySymbol(tokenSymbol string) (string, error) {
	return ethtxs.GetToken2address(ethRelayer.x2EthContracts.BridgeBank, tokenSymbol)
}

func (ethRelayer *Relayer4Ethereum) ShowLockedTokenAddress(tokenSymbol string) (string, error) {
	return ethtxs.GetLockedTokenAddress(ethRelayer.x2EthContracts.BridgeBank, tokenSymbol)
}

//IsProphecyPending ...
func (ethRelayer *Relayer4Ethereum) IsProphecyPending(claimID [32]byte) (bool, error) {
	return ethtxs.IsProphecyPending(claimID, ethRelayer.ethSender, ethRelayer.x2EthContracts.Chain33Bridge)
}

//Burn ...
func (ethRelayer *Relayer4Ethereum) Burn(ownerPrivateKey, tokenAddr, chain33Receiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.Burn(ownerPrivateKey, tokenAddr, chain33Receiver, ethRelayer.x2EthDeployInfo.BridgeBank.Address, bn,
		ethRelayer.x2EthContracts.BridgeBank, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce, ethRelayer.providerHttp[0])
}

//BurnAsync ...
func (ethRelayer *Relayer4Ethereum) BurnAsync(ownerPrivateKey, tokenAddr, chain33Receiver, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.BurnAsync(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce)
}

//TransferToken ...
func (ethRelayer *Relayer4Ethereum) TransferToken(tokenAddr, fromKey, toAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.TransferToken(tokenAddr, fromKey, toAddr, bn, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce, ethRelayer.providerHttp[0])
}

//TransferEth ...
func (ethRelayer *Relayer4Ethereum) TransferEth(fromKey, toAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.TransferEth(fromKey, toAddr, bn, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce, ethRelayer.providerHttp[0])
}

//GetDecimals ...
func (ethRelayer *Relayer4Ethereum) GetDecimals(tokenAddr string) (uint8, error) {
	opts := &bind.CallOpts{
		Pending: true,
		From:    common.HexToAddress(tokenAddr),
		Context: context.Background(),
	}
	bridgeToken, _ := generated.NewBridgeToken(common.HexToAddress(tokenAddr), ethRelayer.clientSpec)
	return bridgeToken.Decimals(opts)
}

//LockEthErc20Asset ...
func (ethRelayer *Relayer4Ethereum) LockEthErc20Asset(ownerPrivateKey, tokenAddr, amount string, chain33Receiver string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.LockEthErc20Asset(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.clientSpec, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.x2EthDeployInfo.BridgeBank.Address, ethRelayer.Addr2TxNonce, ethRelayer.providerHttp[0])
}

//LockEthErc20AssetAsync ...
func (ethRelayer *Relayer4Ethereum) LockEthErc20AssetAsync(ownerPrivateKey, tokenAddr, amount string, chain33Receiver string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.LockEthErc20AssetAsync(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.clientSpec, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.Addr2TxNonce)
}

//ShowTxReceipt ...
func (ethRelayer *Relayer4Ethereum) ShowTxReceipt(hash string) (*types.Receipt, error) {
	txhash := common.HexToHash(hash)
	return ethRelayer.getTransactionReceipt(txhash)
}

func (ethRelayer *Relayer4Ethereum) getClientSpecs() {
	var err error
	bSendEmail := false
	for true {
		ethRelayer.clientSpecs, ethRelayer.clientChainID, err = ethtxs.SetupEthClients(&ethRelayer.providerHttp, ethRelayer.bridgeRegistryAddr)
		if err != nil {
			if !bSendEmail {
				// 节点都不可用 发送邮件
				ethRelayer.remindSetupEthClientError()
				bSendEmail = true
			}
			relayerLog.Error("Failed getClientSpecs SetupEthClients" + err.Error())
		}

		if err == nil {
			relayerLog.Info("Relayer4Ethereum getClientSpecs", "http provider:", ethRelayer.providerHttp, "clientChainID", ethRelayer.clientChainID)
			break
		}

		time.Sleep(sleepTime)
	}
}

func (ethRelayer *Relayer4Ethereum) getClientWss() {
	var err error
	bSendEmail := false
	for true {
		ethRelayer.clientWss, _, err = ethtxs.SetupEthClient(&ethRelayer.provider)
		if err != nil {
			if !bSendEmail {
				// 节点都不可用 发送邮件
				ethRelayer.remindSetupEthClientError()
				bSendEmail = true
			}
			relayerLog.Error("Failed getClientWss SetupEthClients" + err.Error())
		}

		if err == nil {
			relayerLog.Info("Relayer4Ethereum getClientWss", "Started Ethereum websocket with ws provider:", ethRelayer.provider)
			break
		}

		time.Sleep(sleepTime)
	}
}

// 获取同步节点
func (ethRelayer *Relayer4Ethereum) getClientSpec() {
	ethRelayer.clientSpec = ethRelayer.clientSpecs[0].Client
	ethRelayer.clientUrlSelected = ethRelayer.clientSpecs[0].ClientUrl
	ethRelayer.getAvailableClient()
}

func (ethRelayer *Relayer4Ethereum) proc() {
	relayerLog.Info("Relayer4Ethereum proc", "processWithDraw", ethRelayer.processWithDraw, "delayedSend", ethRelayer.delayedSend)
	ethRelayer.getClientSpecs()
	ethRelayer.getClientSpec()
	ethRelayer.getClientWss()
	ethRelayer.clientBSCRecommendSpecs, _ = ethtxs.SetupRecommendClients(&BSCRecommendHttp, ethRelayer.bridgeRegistryAddr)

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
		ethRelayer.oracleInstance = ethRelayer.x2EthContracts.Oracle
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

func (ethRelayer *Relayer4Ethereum) procTxRelayAck(ack *ebTypes.TxRelayAck) {
	//reset with another key to exclude from the check list to resend the same message
	if err := ethRelayer.resetKeyTxRelayedAlready(ethRelayer.name, ack.TxHash); nil != err {
		relayerLog.Error("ethRelayer::procTxRelayAck", "Failed to resetKeyTxRelayedAlready due to:", err.Error())
	}
	//relayEthereum2chain33CheckPonit 4:procTxRelayAck
	relayerLog.Info("relayLockBurnToChain33::relayEthereum2chain33CheckPonit_4::procTxRelayAck", "ethTxhash", ack.TxHash, "ForwardIndex", ack.FdIndex)
}

func (ethRelayer *Relayer4Ethereum) handleChain33Msg(chain33Msg *events.Chain33Msg) {
	if chain33Msg.ClaimType == events.ClaimTypeWithdraw {
		ethRelayer.handleLogWithdraw(chain33Msg)
		return
	}

	ethRelayer.handleLogLockBurn(chain33Msg)
	return
}

func (ethRelayer *Relayer4Ethereum) checkPermissionWithinOneDay(withdrawTx *ebTypes.WithdrawTx) (*big.Int, error) {
	totalAlready, err := ethRelayer.getWithdrawsWithinSameDay(withdrawTx)
	if nil != err {
		relayerLog.Error("checkPermissionWithinOneDay", "Failed to getWithdrawsWithinSameDay due to", err.Error())
		return nil, errors.New("ErrGetWithdrawsWithinSameDay")
	}
	withdrawPara, ok := ethRelayer.withdrawFee[withdrawTx.Symbol]
	if !ok {
		relayerLog.Error("checkPermissionWithinOneDay", "No withdraw parameter configured for symbol ", withdrawTx.Symbol)
		return nil, errors.New("ErrNoWithdrawParaCfged")
	}
	AmountInt, _ := big.NewInt(0).SetString(withdrawTx.Amount, 0)
	totalAlready.Add(totalAlready, AmountInt)
	if totalAlready.Cmp(withdrawPara.AmountPerDay) > 0 {
		relayerLog.Error("checkPermissionWithinOneDay", "No withdraw parameter configured for symbol ", withdrawTx.Symbol)
		return nil, errors.New("ErrWithdrawAmountBigThanQuota")
	}
	relayerLog.Info("checkPermissionWithinOneDay", "total withdraw already", totalAlready, "Chain33Sender", withdrawTx.Chain33Sender,
		"Symbol", withdrawTx.Symbol)
	return withdrawPara.Fee, nil
}

func (ethRelayer *Relayer4Ethereum) SendRemind(url, postData string) {
	res, err := utils.SendToServer(url, strings.NewReader(postData))
	if err != nil {
		relayerLog.Error("SendToServer", "error:", err.Error())
		return
	}
	js, err := simplejson.NewJson(res)
	if err != nil {
		relayerLog.Error("SendToServer", "NewJson error:", err.Error())
		return
	}
	result := js.Get("result").MustBool()
	if result == false {
		reErr := js.Get("error").MustString()
		relayerLog.Error("SendToServer", "send error:", reErr)
		return
	}
	relayerLog.Debug("SendToServer ok")
}

func (ethRelayer *Relayer4Ethereum) remindBalanceNotEnough(addr, symbol, chain33TxHash string) {
	ethName := "以太坊"
	if ethRelayer.GetName() == ethtxs.BinanceChain {
		ethName = "BSC"
	}
	postData := fmt.Sprintf(`{"from":"%s relayer","content":"%s链代理打币地址%s,token:%s 金额不足"}`, ethName, ethName, addr, symbol)
	relayerLog.Debug("SendRemind", "remindUrl", ethRelayer.remindUrl, "postData:", postData, "chain33Txhash", chain33TxHash)
	ethRelayer.SendRemind(ethRelayer.remindUrl, postData)
}

func (ethRelayer *Relayer4Ethereum) handleLogWithdraw(chain33Msg *events.Chain33Msg) {
	//只有通过代理人登录的中继器，才处理提币事件
	var err error
	now := time.Now()
	cstTime := now.UTC().Add(8 * time.Hour)
	year, month, day := cstTime.Date()
	withdrawTx := &ebTypes.WithdrawTx{
		Chain33Sender:    chain33Msg.Chain33Sender.String(),
		EthereumReceiver: chain33Msg.EthereumReceiver.String(),
		Symbol:           chain33Msg.Symbol,
		TxHashOnChain33:  common.Bytes2Hex(chain33Msg.TxHash),
		Nonce:            chain33Msg.Nonce,
		Year:             int32(year),
		Month:            int32(month),
		Day:              int32(day),
	}
	//非代理提币人模式，则不处理代理提币
	if !ethRelayer.processWithDraw {
		relayerLog.Info("handleLogWithdraw", "Needn't process withdraw for this relay validator", ethRelayer.ethSender)
		return
	}

	if ethRelayer.checkIsResendChain33Msg(chain33Msg) {
		return
	}
	chain33TxHash := chain33Common.ToHex(chain33Msg.TxHash)

	defer func() {
		if err != nil {
			withdrawTx.Status = int32(ethtxs.WDError)
			withdrawTx.StatusDescription = ethtxs.WDError.String()
			withdrawTx.ErrorDescription = err.Error()
			relayerLog.Error("handleLogWithdraw", "Failed to withdraw due to:", err.Error(), "chain33Txhash", chain33TxHash)
		}

		err := ethRelayer.setWithdraw(withdrawTx)
		if nil != err {
			relayerLog.Error("handleLogWithdraw", "Failed to setWithdraw due to:", err.Error(), "chain33Txhash", chain33TxHash)
		}

		err = ethRelayer.setWithdrawStatics(withdrawTx, chain33Msg)
		if nil != err {
			relayerLog.Error("handleLogWithdraw", "Failed to setWithdrawStatics due to:", err.Error(), "chain33Txhash", chain33TxHash)
		}
	}()

	relayerLog.Info("handleLogWithdraw", "Received chain33Msg", chain33Msg, "tx hash string", chain33TxHash)
	withdrawFromChain33TokenInfo, exist := ethRelayer.symbol2LockAddr[chain33Msg.Symbol]
	if !exist {
		//因为是withdraw操作，必须从允许lock的token地址中进行查询
		relayerLog.Error("handleLogWithdraw", "Failed to fetch locked Token Info for symbol", chain33Msg.Symbol, "chain33Txhash", chain33TxHash)
		err = errors.New("ErrFetchLockedTokenInfo")
		return
	}

	tokenAddr := common.HexToAddress(withdrawFromChain33TokenInfo.Address)
	//从chain33进行withdraw回来的token需要根据精度进行相应的缩放
	if 8 != withdrawFromChain33TokenInfo.Decimal {
		dist := math.Abs(float64(withdrawFromChain33TokenInfo.Decimal - 8))
		value, exist := utils.Decimal2value[int(dist)]
		if !exist {
			relayerLog.Error("handleLogWithdraw", "does support for decimal, %d", withdrawFromChain33TokenInfo.Decimal, "chain33Txhash", chain33TxHash)
			err = errors.New("ErrDecimalNotSupport")
			return
		}

		if withdrawFromChain33TokenInfo.Decimal > 8 {
			chain33Msg.Amount.Mul(chain33Msg.Amount, big.NewInt(value))
		} else {
			chain33Msg.Amount.Div(chain33Msg.Amount, big.NewInt(value))
		}
	}
	withdrawTx.Amount = chain33Msg.Amount.String()
	relayerLog.Info("handleLogWithdraw", "token address", tokenAddr.String(), "amount", withdrawTx.Amount,
		"Receiver on Ethereum", chain33Msg.EthereumReceiver.String(), "chain33Txhash", chain33TxHash)
	//检查用户提币权限是否得到满足：比如是否超过累计提币额度
	var feeAmount *big.Int
	if feeAmount, err = ethRelayer.checkPermissionWithinOneDay(withdrawTx); nil != err {
		return
	}
	if chain33Msg.Amount.Cmp(feeAmount) < 0 {
		relayerLog.Error("handleLogWithdraw", "ErrWithdrawAmountLessThanFee feeAmount", feeAmount.String(), "Withdraw Amount", chain33Msg.Amount.String(), "chain33Txhash", chain33TxHash)
		err = errors.New("ErrWithdrawAmountCan'tPay4Fee")
		return
	}
	amount2transfer := chain33Msg.Amount.Sub(chain33Msg.Amount, feeAmount)
	value := big.NewInt(0)

	//此处需要完成在以太坊发送以太或者ERC20数字资产的操作
	var intputData []byte // ERC20 or BEP20 token transfer pack data
	var toAddr common.Address
	var senderAddr common.Address
	var balanceOfData []byte // ERC20 or BEP20 token balanceof pack data

	if tokenAddr.String() != ethtxs.EthNullAddr { //判断是否要Pack EVM数据
		toAddr = tokenAddr
		senderAddr = tokenAddr
		intputData, err = ethRelayer.packTransferData(chain33Msg.EthereumReceiver, amount2transfer)
		if err != nil {
			relayerLog.Error("handleLogWithdraw", "CallEvmData err", err, "chain33Txhash", chain33TxHash)
			err = errors.New("ErrPackTransferData")
			return
		}
		//用签名的账户地址作为pack参数，toAddr作为合约地址
		balanceOfData, err = ethRelayer.packBalanceOfData(ethRelayer.ethSender)
		if err != nil {
			relayerLog.Error("handleLogWithdraw", "callEvmBalanceData err", err, "chain33Txhash", chain33TxHash)
			err = errors.New("ErrPackBalanceOfData")
			return
		}
	} else {
		//如果tokenAddr为空，则把toAddr设置为用户指定的地址
		toAddr = chain33Msg.EthereumReceiver
		senderAddr = ethRelayer.ethSender
		value = amount2transfer
	}

	ethRelayer.getAvailableClient()
	//校验余额是否充足
	err = ethRelayer.checkBalanceEnough(senderAddr, amount2transfer, balanceOfData)
	if err != nil {
		relayerLog.Error("handleLogWithdraw", "Failed to checkBalanceEnough:", err.Error(), "chain33Txhash", chain33TxHash)
		err = errors.New("ErrBalanceNotEnough")
		ethRelayer.remindBalanceNotEnough(ethRelayer.ethSender.String(), chain33Msg.Symbol, chain33TxHash)
		return
	}

	// 构建交易并签名
	signedTx, err := ethRelayer.NewTransferSignTx(toAddr, intputData, value, false)
	if err != nil {
		relayerLog.Error("handleLogWithdraw", "NewTransferSignTx err", err, "chain33Txhash", chain33TxHash)
		return
	}
	//交易发送
	err = ethRelayer.sendEthereumTx(signedTx)
	if err != nil {
		// 如果是 nonce 出错导致的错误 再次构建交易发送
		if err.Error() == core.ErrNonceTooLow.Error() || err.Error() == core.ErrNonceTooHigh.Error() {
			relayerLog.Error("handleLogWithdraw", "sendEthereumTx err", err, "出现 nonce 错误重新构建并发送交易, chain33Txhash", chain33TxHash)
			signedTx, err = ethRelayer.NewTransferSignTx(toAddr, intputData, value, true)
			if err != nil {
				relayerLog.Error("handleLogWithdraw", "NewTransferSignTx err", err, "chain33Txhash", chain33TxHash)
				return
			}

			err = ethRelayer.sendEthereumTx(signedTx)
			if err != nil {
				err = errors.New("ErrSendTransaction")
				return
			}
		} else {
			err = errors.New("ErrSendTransaction")
			return
		}
	}
	ethTxhash := signedTx.Hash().Hex()
	relayerLog.Info("handleLogWithdraw", "SendTransaction Hash", ethTxhash, "chain33Txhash", chain33TxHash)

	withdrawTx.Status = int32(ethtxs.WDPending)
	withdrawTx.StatusDescription = ethtxs.WDPending.String()
	withdrawTx.TxHashOnEthereum = ethTxhash

	return
}

func (ethRelayer *Relayer4Ethereum) NewTransferSignTx(toAddr common.Address, intputData []byte, value *big.Int, fromChain bool) (*types.Transaction, error) {
	tx, err := ethtxs.NewTransferTx(ethRelayer.clientSpec, ethRelayer.ethSender, toAddr, intputData, value, ethRelayer.Addr2TxNonce, fromChain)
	if err != nil {
		relayerLog.Error("handleLogWithdraw", "newTx err", err)
		return nil, errors.New("ErrNewTx")
	}

	//交易签名
	signedTx, err := ethRelayer.signTx(tx, ethRelayer.privateKey4Ethereum)
	if err != nil {
		relayerLog.Error("handleLogWithdraw", "SignTx err", err)
		return nil, errors.New("ErrSignTx")
	}

	return signedTx, nil
}

func (ethRelayer *Relayer4Ethereum) checkBalanceEnough(addr common.Address, amount *big.Int, inputdata []byte) error {
	//检测地址余额
	var balance *big.Int
	var err error
	if inputdata == nil {
		balance, err = ethRelayer.getBalanceAt(addr)
		if err != nil {
			return err
		}
	} else {
		var msg ethereum.CallMsg
		msg.To = &addr //合约地址
		msg.Data = inputdata
		result, err := ethRelayer.getCallContract(msg)
		if err != nil {
			return err
		}
		var ok bool
		balance, ok = big.NewInt(1).SetString(common.Bytes2Hex(result), 16)
		if !ok {
			return errors.New(fmt.Sprintf("token balance err:%v", common.Bytes2Hex(result)))
		}
	}

	//与要发动的金额大小进行比较
	if balance.Cmp(amount) > 0 {
		return nil
	}
	relayerLog.Error("Insufficient balance", "balance", balance, "amount", amount)
	return errors.New("insufficient balance")
}

func (ethRelayer *Relayer4Ethereum) signTx(tx *types.Transaction, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	signer := types.NewEIP155Signer(ethRelayer.clientChainID)
	txhash := signer.Hash(tx)
	signature, err := crypto.Sign(txhash.Bytes(), key)
	if err != nil {
		return nil, err
	}
	tx, err = tx.WithSignature(signer, signature)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (ethRelayer *Relayer4Ethereum) packTransferData(_to common.Address, _value *big.Int) ([]byte, error) {
	parsed, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		return nil, err
	}
	abidata, err := parsed.Pack("transfer", _to, _value)
	if err != nil {
		return nil, err
	}
	return abidata, nil
}

func (ethRelayer *Relayer4Ethereum) packBalanceOfData(_to common.Address) ([]byte, error) {
	parsed, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		return nil, err
	}
	abidata, err := parsed.Pack("balanceOf", _to)
	if err != nil {
		return nil, err
	}
	return abidata, nil
}

func (ethRelayer *Relayer4Ethereum) checkIsResendChain33Msg(chain33Msg *events.Chain33Msg) bool {
	if chain33Msg.ForwardTimes <= 1 {
		return false
	}
	chain33TxHash := chain33Common.ToHex(chain33Msg.TxHash)
	relayerLog.Info("checkIsResendChain33Msg", "Received the same chain33Msg more than once with times", chain33Msg.ForwardTimes, "tx hash string", chain33TxHash)
	relayTxDetail, _ := ethRelayer.getChain33TxRelayAlreadyInfo(ethRelayer.name, chain33TxHash)
	if nil == relayTxDetail {
		relayerLog.Info("checkIsResendChain33Msg::haven't relay yet")
		return false
	}

	ethRelayer.txRelayAckSendChan <- &ebTypes.TxRelayAck{
		TxHash:  chain33TxHash,
		FdIndex: chain33Msg.ForwardIndex,
	}
	//relaychain33ToEthereumCheckPonit 2: send ack
	relayerLog.Info("checkIsResendChain33Msg::relaychain33ToEthereumCheckPonit_2::sendBackAck", "chain33TxHash", chain33TxHash, "ForwardIndex", chain33Msg.ForwardIndex, "FdTimes", chain33Msg.ForwardTimes)

	relayerLog.Info("checkIsResendChain33Msg", "have relay already with tx hash:", relayTxDetail.Txhash)
	return true
}

func (ethRelayer *Relayer4Ethereum) handleLogLockBurn(chain33Msg *events.Chain33Msg) {
	//对于通过代理人登录的中继器，不处理lock和burn事件
	if ethRelayer.processWithDraw {
		relayerLog.Info("handleLogLockBurn", "Needn't process lock and burn for this withdraw process specified validator", ethRelayer.ethSender)
		return
	}
	chain33TxHash := chain33Common.ToHex(chain33Msg.TxHash)
	relayerLog.Info("handleLogLockBurn", "Received chain33Msg", chain33Msg, "tx hash string", chain33TxHash)
	if ethRelayer.checkIsResendChain33Msg(chain33Msg) {
		return
	}

	// Parse the Chain33Msg into a ProphecyClaim for relay to Ethereum
	prophecyClaim := ethtxs.Chain33MsgToProphecyClaim(*chain33Msg)
	var tokenAddr common.Address
	exist := false
	operationType := chain33Msg.ClaimType.String()
	if chain33Msg.ClaimType == events.ClaimTypeLock {
		tokenAddr, exist = ethRelayer.symbol2Addr[prophecyClaim.Symbol]
		if !exist {
			relayerLog.Info("handleLogLockBurn", "Query address from ethereum for symbol", prophecyClaim.Symbol)
			//因为是lock操作，则需要从创建的bridgeToken中进行查询
			addr, err := ethRelayer.ShowTokenAddrBySymbol(prophecyClaim.Symbol)
			if err != nil {
				panic(fmt.Sprintf("Pls create bridge token in advance for token:%s", prophecyClaim.Symbol))
			}
			token2set := ebTypes.TokenAddress{
				Address:   addr,
				Symbol:    prophecyClaim.Symbol,
				ChainName: ebTypes.EthereumBlockChainName,
			}
			err = ethRelayer.SetTokenAddress(&token2set)
			if nil != err {
				// 尽管设置数据失败，但是不影响运行，只是relayer启动时，每次从节点远程获取bridge token地址而已
				relayerLog.Error("handleLogLockBurn", "Failed to SetTokenAddress due to", err.Error())
			}
			tokenAddr = common.HexToAddress(addr)
		}
	} else {
		burnFromChain33TokenInfo, exist := ethRelayer.symbol2LockAddr[prophecyClaim.Symbol]
		if !exist {
			//因为是burn操作，必须从允许lock的token地址中进行查询
			relayerLog.Error("handleLogLockBurn", "Failed to fetch locked Token Info for symbol", prophecyClaim.Symbol)
			return
		}

		tokenAddr = common.HexToAddress(burnFromChain33TokenInfo.Address)
		//从chain33进行withdraw回来的token需要根据精度进行相应的缩放
		if 8 != burnFromChain33TokenInfo.Decimal {
			if burnFromChain33TokenInfo.Decimal > 8 {
				dist := burnFromChain33TokenInfo.Decimal - 8
				value, exist := utils.Decimal2value[int(dist)]
				if !exist {
					panic(fmt.Sprintf("does support for decimal, %d", burnFromChain33TokenInfo.Decimal))
				}
				prophecyClaim.Amount.Mul(prophecyClaim.Amount, big.NewInt(value))
			} else {
				dist := 8 - burnFromChain33TokenInfo.Decimal
				value, exist := utils.Decimal2value[int(dist)]
				if !exist {
					panic(fmt.Sprintf("does support for decimal, %d", burnFromChain33TokenInfo.Decimal))
				}
				prophecyClaim.Amount.Div(prophecyClaim.Amount, big.NewInt(value))
			}
		}
	}

	var ethTxhash string
	var err error
	isClaimIDProcessed := false
	if ethRelayer.delayedSend {
		claimID := crypto.Keccak256Hash(prophecyClaim.Chain33TxHash, prophecyClaim.Chain33Sender, prophecyClaim.EthereumReceiver.Bytes(), []byte(prophecyClaim.Symbol), prophecyClaim.Amount.Bytes())
		prophecyProcessed, err := ethRelayer.getClaimIDExecuteAlready(claimID.String())
		if nil != err {
			relayerLog.Info("handleLogLockBurn", "Failed to getClaimIDExecuteAlready due to", err.Error(), "claimID", claimID.String())
		} else {
			if prophecyProcessed.Valid {
				isClaimIDProcessed = true
				ethTxhash = prophecyProcessed.Txhash
				relayerLog.Info("handleLogLockBurn", "prophecyProcessed Valid with tx hash", chain33TxHash)
			}
		}
	}

	if !isClaimIDProcessed {
		ethRelayer.getAvailableClient()
		burnOrLockParameter := &ethtxs.BurnOrLockParameter{
			ClientSpec:              &ethtxs.EthClientWithUrl{ClientUrl: ethRelayer.clientUrlSelected, Client: ethRelayer.clientSpec, OracleInstance: ethRelayer.oracleInstance},
			Clients:                 ethRelayer.clientSpecs,
			ClientBSCRecommendSpecs: ethRelayer.clientBSCRecommendSpecs,
			Sender:                  ethRelayer.ethSender,
			TokenOnEth:              tokenAddr,
			Claim:                   prophecyClaim,
			PrivateKey:              ethRelayer.privateKey4Ethereum,
			Addr2TxNonce:            ethRelayer.Addr2TxNonce,
			ChainId:                 ethRelayer.clientChainID,
			ChainName:               ethRelayer.name,
		}

		// Relay the Chain33Msg to the Ethereum network
		ethTxhash, err = ethtxs.RelayOracleClaimToEthereum(burnOrLockParameter)
		if err != nil {
			//此处收集更多的错误信息
			relayerLog.Error("handleLogLockBurn", "RelayOracleClaimToEthereum failed due to", err.Error())
			panic("RelayOracleClaimToEthereum failed due to" + err.Error())
		}
		relayerLog.Info("handleLogLockBurn", "RelayOracleClaimToEthereum with tx hash", ethTxhash)
	}

	ethRelayer.txRelayAckSendChan <- &ebTypes.TxRelayAck{
		TxHash:  chain33TxHash,
		FdIndex: chain33Msg.ForwardIndex,
	}
	//relaychain33ToEthereumCheckPonit 2: send ack to chain33 relay service
	relayerLog.Info("handleLogLockBurn::relaychain33ToEthereumCheckPonit_2::sendBackAck", "chain33TxHash", chain33TxHash, "ForwardIndex", chain33Msg.ForwardIndex, "FdTimes", chain33Msg.ForwardTimes)

	//保存交易hash，方便查询
	txIndex := atomic.AddInt64(&ethRelayer.totalTxRelayFromChain33, 1)
	if err = ethRelayer.updateTotalTxAmountFromchain33(txIndex); nil != err {
		relayerLog.Error("handleLogLockBurn", "Failed to RelayLockToChain33 due to:", err.Error())
		return
	}
	statics := &ebTypes.Chain33ToEthereumStatics{
		EthTxstatus:      ebTypes.Tx_Status_Pending,
		Chain33Txhash:    chain33TxHash,
		EthereumTxhash:   ethTxhash,
		BurnLockWithdraw: int32(chain33Msg.ClaimType),
		Chain33Sender:    chain33Msg.Chain33Sender.String(),
		EthereumReceiver: chain33Msg.EthereumReceiver.String(),
		Symbol:           chain33Msg.Symbol,
		Amount:           chain33Msg.Amount.String(),
		Nonce:            chain33Msg.Nonce,
		TxIndex:          txIndex,
		OperationType:    operationType,
	}
	data := chain33Types.Encode(statics)
	if err = ethRelayer.setLastestStatics(int32(chain33Msg.ClaimType), txIndex, data); nil != err {
		relayerLog.Error("handleLogLockBurn", "Failed to RelayLockToChain33 due to:", err.Error())
		return
	}

	relayTxDetail := &ebTypes.RelayTxDetail{
		ClaimType:      int32(chain33Msg.ClaimType),
		TxIndexRelayed: txIndex,
		Txhash:         ethTxhash,
	}

	if err = ethRelayer.setChain33TxRelayAlreadyInfo(ethRelayer.name, chain33TxHash, relayTxDetail); nil != err {
		relayerLog.Error("handleLogLockBurn", "Failed to setEthTxRelayAlreadyInfo due to:", err.Error())
		return
	}
	//relaychain33ToEthereumCheckPonit 3: set flag to send relay tx to ethereum node
	relayerLog.Info("handleLogLockBurn::relaychain33ToEthereumCheckPonit_3::setRelayFinishFlag", "chain33TxHash", chain33TxHash, "ForwardIndex", chain33Msg.ForwardIndex, "FdTimes", chain33Msg.ForwardTimes)

	relayerLog.Info("RelayOracleClaimToEthereum::successful",
		"txIndex", txIndex,
		"Chain33Txhash", statics.Chain33Txhash,
		"EthereumTxhash", statics.EthereumTxhash,
		"type", operationType,
		"Symbol", chain33Msg.Symbol,
		"Amount", chain33Msg.Amount,
		"EthereumReceiver", statics.EthereumReceiver,
		"Chain33Sender", statics.Chain33Sender)
}

func (ethRelayer *Relayer4Ethereum) getAvailableClient() {
	timeout, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()

	if syncProc, err := ethRelayer.clientSpec.SyncProgress(timeout); nil != syncProc || nil != err {
		relayerLog.Error("getAvailableClient", "Eth node not syncing for address", ethRelayer.clientUrlSelected)
		for {
			var urlSelected string
			ethRelayer.clientSpec, urlSelected, err = ethtxs.SetupEthClient(&ethRelayer.providerHttp)
			if err != nil {
				relayerLog.Error("getAvailableClient", "Failed to SetupEthClient due to", err.Error())
				time.Sleep(5 * time.Second)
				continue
			}

			timeout2, cancel2 := context.WithTimeout(context.Background(), waitTime)
			if syncProc, err := ethRelayer.clientSpec.SyncProgress(timeout2); nil != syncProc || nil != err {
				cancel2()
				relayerLog.Error("getAvailableClient", "Eth node not syncing for address", urlSelected)
				time.Sleep(5 * time.Second)
				continue
			}
			cancel2()

			// 获取新的同步节点后, OracleInstance也重新获取
			oracleInstance, err := ethtxs.GetOracleInstance(ethRelayer.clientSpec, ethRelayer.bridgeRegistryAddr)
			if nil != err {
				if err.Error() != cross2ethErrors.ErrContractNotRegistered.Error() {
					panic("failed to GetOracleInstance" + err.Error())
				}
				relayerLog.Error("getAvailableClient", "GetOracleInstance err", err.Error())
			}
			ethRelayer.oracleInstance = oracleInstance
			ethRelayer.clientUrlSelected = urlSelected
			relayerLog.Info("getAvailableClient", "Eth node is syncing for address", urlSelected)
			break
		}
	}
	return
}

func (ethRelayer *Relayer4Ethereum) getCurrentHeight() (uint64, error) {
	return ethRelayer.getHeaderByNumber()
}

func (ethRelayer *Relayer4Ethereum) ReGetEvent(start, end int64) (string, error) {
	if end < start {
		return "", errors.New("ErrStartEndHeight")
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{ethRelayer.bridgeBankAddr},
	}

	batchCount := int64(10)
	info := ""
	for {
		if batchCount < (end - start + 1) {
			stopHeight := start + batchCount - 1
			query.FromBlock = big.NewInt(start)
			query.ToBlock = big.NewInt(stopHeight)
		} else {
			query.FromBlock = big.NewInt(start)
			query.ToBlock = big.NewInt(end)
		}

		// Filter by contract and event, write results to logs
		logs, err := ethRelayer.getFilterLogs(query)
		if err != nil {
			errinfo := fmt.Sprintf("Failed to filterLogEvents due to:%s", err.Error())
			return "", errors.New(errinfo)
		}

		relayerLog.Info("ReGetEvent", "received logs with number", len(logs),
			"start height", query.FromBlock.String(), "stop height", query.ToBlock.String())
		for _, logv := range logs {
			relayerLog.Info("ReGetEvent", "received log with topics", logv.Topics[0].Hex(), "BlockNumber", logv.BlockNumber)
			if ethRelayer.bridgeBankEventLockSig != logv.Topics[0].Hex() {
				continue
			}

			receipt, err := ethRelayer.getTransactionReceipt(logv.TxHash)
			if nil != err {
				relayerLog.Error("ReGetEvent", "Failed to get tx receipt with hash", logv.TxHash.String())
				return "", err
			}

			//检查当前的交易是否成功执行
			if receipt.Status != types.ReceiptStatusSuccessful {
				relayerLog.Error("ReGetEvent", "tx not successful with status", receipt.Status)
				return "", errors.New("tx not successful")
			}

			eventName := events.LogLockFromETH.String()
			err = ethRelayer.handleLogLockEvent(ethRelayer.clientChainID, ethRelayer.bridgeBankAbi, eventName, logv)
			if nil != err {
				relayerLog.Error("ReGetEvent", "Failed to handleLogLockEvent for tx", logv.TxHash.String())
				return "", err
			}
			info += fmt.Sprintf("Ethereum tx with hash = %s is relayed\n", logv.TxHash.String())

		}

		if query.ToBlock.Int64() >= end {
			relayerLog.Info("ReGetEvent", "Finished FilterLogs to height", end)
			break
		}
		start = query.ToBlock.Int64() + 1
	}

	return info, nil
}

func (ethRelayer *Relayer4Ethereum) ResendLockEvent(height uint64, index uint32) (string, error) {
	relayerLog.Info("Relayer4Ethereum::ResendEvent", "height", height, "index", index)

	logs, err := ethRelayer.getNextValidEthTxEventLogs(height, index, 1)
	if nil != err {
		relayerLog.Error("Failed to get ethereum height", "getNextValidEthTxEventLogs err", err.Error())
		return "", err
	}

	if 0 == len(logs) {
		relayerLog.Info("Relayer4Ethereum::ResendEvent get nil")
		return "No event need to be relayed to chain33", nil
	}
	vLog := *logs[0]

	receipt, err := ethRelayer.getTransactionReceipt(vLog.TxHash)
	if nil != err {
		relayerLog.Error("procBridgeBankLogs", "Failed to get tx receipt with hash", vLog.TxHash.String())
		return "", err
	}

	//检查当前的交易是否成功执行
	if receipt.Status != types.ReceiptStatusSuccessful {
		relayerLog.Error("procBridgeBankLogs", "tx not successful with status", receipt.Status)
		return "", errors.New("tx not successful")
	}

	eventName := events.LogLockFromETH.String()
	err = ethRelayer.handleLogLockEvent(ethRelayer.clientChainID, ethRelayer.bridgeBankAbi, eventName, vLog)
	info := fmt.Sprintf("Ethereum tx with hash = %s is relayed", vLog.TxHash.String())
	return info, err
}

func (ethRelayer *Relayer4Ethereum) checkTxRelay2Chain33() {
	txInfos, err := ethRelayer.getAllTxsUnconfirm()
	if err != nil {
		relayerLog.Error("ethRelayer::checkTxRelay2Chain33", "Failed to getAllTxsUnconfirm due to", err.Error())
		return
	}
	if 0 == len(txInfos) {
		return
	}
	for _, txInfo := range txInfos {
		txHashStr := txInfo.TxHash

		if !txInfo.Resend {
			//为了防止转发出去的消息之后，下一个区块时间马上到来，首次转发的消息需要至少等一个区块间隔之后才会进行转发
			txInfo.Resend = true
			err = ethRelayer.setTxIsRelayedUnconfirm(ethRelayer.name, txHashStr, txInfo.FdIndex, txInfo)
			if nil != err {
				relayerLog.Error("ethRelayer::checkTxRelay2Chain33", "Failed to SetTxIsRelayedconfirm due to", err.Error())
				return
			}
			continue
		}

		ethRelayer.rwLock.RLock()
		event, err := events.UnpackLogLock(ethRelayer.bridgeBankAbi, events.LogLockFromETH.String(), txInfo.Data)
		if nil != err {
			relayerLog.Error("ethRelayer::checkTxRelay2Chain33", "Failed to UnpackLogLock due to", err.Error())
			ethRelayer.rwLock.RUnlock()
			return
		}
		ethRelayer.rwLock.RUnlock()

		tokenLocked, err := ethRelayer.GetLockedTokenAddress(event.Symbol)
		if nil != err {
			relayerLog.Error("ethRelayer::checkTxRelay2Chain33", "Failed to GetLockedTokenAddress due to", err.Error(),
				"symbol", event.Symbol, "chain Name", ethRelayer.name)
			return
		}

		decimal := tokenLocked.Decimal
		ethRelayer.rwLock.RLock()
		prophecyClaim, err := ethtxs.LogLockToEthBridgeClaim(event, ethRelayer.clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), txHashStr, int64(decimal))
		if err != nil {
			ethRelayer.rwLock.RUnlock()
			relayerLog.Error("ethRelayer::checkTxRelay2Chain33", "Failed to LogLockToEthBridgeClaim due to", err.Error())
			return
		}
		ethRelayer.rwLock.RUnlock()
		prophecyClaim.ChainName = ethRelayer.name
		prophecyClaim.ForwardIndex = txInfo.FdIndex
		txInfo.FdTimes += 1
		prophecyClaim.ForwardTimes = txInfo.FdTimes
		ethRelayer.ethBridgeClaimChan <- prophecyClaim

		//relayEthereum2chain33CheckPonit 5:resendClaim
		relayerLog.Info("checkTxRelay2Chain33::relayEthereum2chain33CheckPonit_5::resendClaim", "ethTxhash", txInfo.TxHash, "ForwardIndex", txInfo.FdIndex, "FdTimes", txInfo.FdTimes)
		err = ethRelayer.setTxIsRelayedUnconfirm(ethRelayer.name, txHashStr, txInfo.FdIndex, txInfo)
		if nil != err {
			relayerLog.Error("ethRelayer::checkTxRelay2Chain33", "Failed to setTxIsRelayedconfirm due to", err.Error())
			return
		}
	}
}

func (ethRelayer *Relayer4Ethereum) procNewHeight() {
	currentHeight, _ := ethRelayer.getCurrentHeight()
	ethRelayer.updateTxStatus()
	ethRelayer.checkTxRelay2Chain33()
	relayerLog.Info("procNewHeight", "currentHeight", currentHeight, "ethRelayer.eventLogIndex.Height", ethRelayer.eventLogIndex.Height, "uint64(ethRelayer.maturityDegree)", uint64(ethRelayer.maturityDegree))

	//一次最大只获取10个logEvent进行处理
	fetchCnt := int32(10)
	for ethRelayer.eventLogIndex.Height+uint64(ethRelayer.maturityDegree)+1 <= currentHeight {
		logs, err := ethRelayer.getNextValidEthTxEventLogs(ethRelayer.eventLogIndex.Height, ethRelayer.eventLogIndex.Index, fetchCnt)
		if nil != err {
			relayerLog.Error("Failed to get ethereum height", "getNextValidEthTxEventLogs err", err.Error())
			return
		}

		for i, vLog := range logs {
			if vLog.BlockNumber+uint64(ethRelayer.maturityDegree)+1 > currentHeight {
				logs = logs[:i]
				break
			}
			ethRelayer.procBridgeBankLogs(*vLog)
		}

		cnt := int32(len(logs))
		if len(logs) > 0 {
			//firstHeight := logs[0].BlockNumber
			lastHeight := logs[cnt-1].BlockNumber
			index := logs[cnt-1].TxIndex
			//获取的数量小于批量获取数量，则认为直接
			ethRelayer.setBridgeBankProcessedHeight(lastHeight, uint32(index))
			ethRelayer.eventLogIndex.Height = lastHeight
			ethRelayer.eventLogIndex.Index = uint32(index)
		}

		//当前需要处理的event数量已经少于10个，直接返回
		if cnt < fetchCnt {
			return
		}
	}
}

func (ethRelayer *Relayer4Ethereum) storeBridgeBankLogs(vLog types.Log, setBlockNumber bool) {
	//lock,用于捕捉 (ETH/ERC20----->chain33) 跨链转移
	//burn,用于捕捉 (chain33 token----->chain33) 实现chain33资产withdraw操作，之后在chain33上实现unlock操作
	if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventLockSig {
		//先进行数据的持久化，等到一定的高度成熟度之后再进行处理
		relayerLog.Info("Relayer4Ethereum storeBridgeBankLogs", "^_^ ^_^ Received bridgeBankLog for event", "lock",
			"Block number:", vLog.BlockNumber, "tx Index", vLog.TxIndex, "log Index", vLog.Index, "Tx hash:", vLog.TxHash.Hex())
		if err := ethRelayer.setEthTxEvent(vLog); nil != err {
			panic(err.Error())
		}
	} else if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventBurnSig {
		relayerLog.Info("Relayer4Ethereum storeBridgeBankLogs", "^_^ ^_^ Received bridgeBankLog for event", "burn",
			"Block number:", vLog.BlockNumber, "tx Index", vLog.TxIndex, "log Index", vLog.Index, "Tx hash:", vLog.TxHash.Hex())
		if err := ethRelayer.setEthTxEvent(vLog); nil != err {
			panic(err.Error())
		}
	} else if vLog.Topics[0].Hex() == ethRelayer.oracleEventSig {
		relayerLog.Info("Relayer4Ethereum storeBridgeBankLogs", "^_^ ^_^ Received oracleEventLog for event", "LogProphecyProcessed",
			"Block number:", vLog.BlockNumber, "tx Index", vLog.TxIndex, "log Index", vLog.Index, "Tx hash:", vLog.TxHash.Hex())
		if err := ethRelayer.setEthTxEvent(vLog); nil != err {
			panic(err.Error())
		}
	}

	//确定是否需要更新保存同步日志高度
	if setBlockNumber {
		if err := ethRelayer.setHeight4BridgeBankLogAt(vLog.BlockNumber); nil != err {
			panic(err.Error())
		}
	}
}

func (ethRelayer *Relayer4Ethereum) procBridgeBankLogs(vLog types.Log) {
	txHashStr := vLog.TxHash.String()
	if ethRelayer.checkTxProcessed(txHashStr) {
		relayerLog.Info("procBridgeBankLogs", "Tx has been already Processed with hash:", txHashStr,
			"height", vLog.BlockNumber, "index", vLog.Index)
		return
	}

	//检查当前交易是否因为区块回退而导致交易丢失
	receipt, err := ethRelayer.getTransactionReceipt(vLog.TxHash)
	if nil != err {
		relayerLog.Error("procBridgeBankLogs", "Failed to get tx receipt with hash", txHashStr)
		return
	}

	//检查当前的交易是否成功执行
	if receipt.Status != types.ReceiptStatusSuccessful {
		relayerLog.Error("procBridgeBankLogs", "tx not successful with status", receipt.Status)
		return
	}

	//lock,用于捕捉 (ETH/ERC20----->chain33) 跨链转移
	if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventLockSig {
		eventName := events.LogLockFromETH.String()
		relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
			"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
		err := ethRelayer.handleLogLockEvent(ethRelayer.clientChainID, ethRelayer.bridgeBankAbi, eventName, vLog)
		if err != nil {
			errinfo := fmt.Sprintf("Failed to handleLogLockEvent due to:%s", err.Error())
			relayerLog.Error("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
			panic(errinfo)
		}
	} else if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventBurnSig {
		//burn,用于捕捉 (chain33 token----->chain33) 实现chain33资产withdraw操作，之后在chain33上实现unlock操作
		//代理提币节点不转发burn信息
		if ethRelayer.processWithDraw {
			return
		}
		eventName := events.LogBurnFromETH.String()
		relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
			"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
		err := ethRelayer.handleLogBurnEvent(ethRelayer.clientChainID, ethRelayer.bridgeBankAbi, eventName, vLog)
		if err != nil {
			errinfo := fmt.Sprintf("Failed to handleLogBurnEvent due to:%s", err.Error())
			relayerLog.Error("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
			panic(errinfo)
		}
	} else if vLog.Topics[0].Hex() == ethRelayer.oracleEventSig {
		eventName := events.LogProphecyProcessed.String()
		event, err := events.UnpackLogProphecyProcessed(ethRelayer.oracleAbi, eventName, vLog.Data)
		if nil != err {
			errinfo := fmt.Sprintf("Failed to LogProphecyProcessed due to:%s", err.Error())
			relayerLog.Error("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
			panic(errinfo)
		}

		//claimID := crypto.Keccak256Hash(event.ClaimID[:])
		claimID := hexutil.Encode(event.ClaimID[:])
		relayerLog.Info("Relayer4Ethereum ProphecyProcessedLogs", "claimID", claimID)

		info := &ebTypes.ProphecyProcessed{
			ClaimID: claimID,
			Valid:   true,
			Txhash:  vLog.TxHash.String(),
		}
		err = ethRelayer.setClaimIDExecuteAlready(claimID, info)
		if nil != err {
			relayerLog.Info("Relayer4Ethereum setClaimIDExecuteAlready", "errinfo", err)
		}
	}
}

func (ethRelayer *Relayer4Ethereum) getcurHeight() (int64, int64) {
	curHeightUint64, _ := ethRelayer.getCurrentHeight()
	curHeight := int64(curHeightUint64)
	relayerLog.Info("filterLogEvents", "curHeight:", curHeight)

	//获取部署高度
	deployHeight, _ := ethtxs.GetDeployHeight(ethRelayer.clientSpec, ethRelayer.x2EthDeployInfo.BridgeRegistry.Address, ethRelayer.x2EthDeployInfo.BridgeRegistry.Address)
	//获取上次处理过的高度
	height4BridgeBankLogAt := int64(ethRelayer.getHeight4BridgeBankLogAt())

	//2者取其大，以为处理高度开始为0
	if height4BridgeBankLogAt < deployHeight {
		height4BridgeBankLogAt = deployHeight
	}
	//确认配置信息是否配置了起始侦听高度，如果是，且大于保存的起始高度，则直接使用较大的高度
	if height4BridgeBankLogAt < ethRelayer.startListenHeight {
		height4BridgeBankLogAt = ethRelayer.startListenHeight
	}

	return curHeight, height4BridgeBankLogAt
}

//因为订阅事件的功能只会推送在订阅生效的高度之后的事件，之前订阅停止～当前订阅生效高度的这一段只能通过FilterLogs来获取事件信息，否则就会遗漏
func (ethRelayer *Relayer4Ethereum) filterLogEvents() {
	curHeight, height4BridgeBankLogAt := ethRelayer.getcurHeight()
	if height4BridgeBankLogAt >= curHeight {
		relayerLog.Error("filterLogEvents height4BridgeBankLogAt > curHeight", "height4BridgeBankLogAt", height4BridgeBankLogAt, "curHeight:", curHeight)
		return
	}

	contractAddrs := []common.Address{ethRelayer.bridgeBankAddr}
	bridgeBankSig := make(map[string]bool)
	ethRelayer.rwLock.RLock()
	bridgeBankSig[ethRelayer.bridgeBankEventLockSig] = true
	bridgeBankSig[ethRelayer.bridgeBankEventBurnSig] = true
	if ethRelayer.delayedSend {
		bridgeBankSig[ethRelayer.oracleEventSig] = true
		contractAddrs = append(contractAddrs, ethRelayer.oracleAddr)
	}
	ethRelayer.rwLock.RUnlock()
	bridgeBankLog := make(chan types.Log)
	done := make(chan int)

	go ethRelayer.filterLogEventsProc(bridgeBankLog, done, "bridgeBank", curHeight, height4BridgeBankLogAt, contractAddrs, bridgeBankSig)

	for {
		select {
		case vLog := <-bridgeBankLog:
			ethRelayer.storeBridgeBankLogs(vLog, true)
		case vLog := <-ethRelayer.bridgeBankLog:
			//因为此处是同步保存信息，防止未同步完成出现panic时，直接将其设置为最新高度，中间出现部分信息不同步的情况
			ethRelayer.storeBridgeBankLogs(vLog, false)
		case <-done:
			relayerLog.Info("Finshed offline logs processed")
			return
		}
	}
}

//因为订阅事件的功能只会推送在订阅生效的高度之后的事件，之前订阅停止～当前订阅生效高度的这一段只能通过FilterLogs来获取事件信息，否则就会遗漏
func (ethRelayer *Relayer4Ethereum) filterLogEventsProc(logchan chan<- types.Log, done chan<- int, title string, curHeight, heightLogProcAt int64, contractAddrs []common.Address, eventSig map[string]bool) {
	relayerLog.Info(title, "eventSig", eventSig, "heightLogProcAt", heightLogProcAt, "curHeight", curHeight)

	startHeight := heightLogProcAt
	batchCount := int64(10)
	query := ethereum.FilterQuery{
		Addresses: contractAddrs,
	}

	for {
		if batchCount < (curHeight - startHeight + 1) {
			stopHeight := startHeight + batchCount - 1
			query.FromBlock = big.NewInt(startHeight)
			query.ToBlock = big.NewInt(stopHeight)
		} else {
			query.FromBlock = big.NewInt(startHeight)
			query.ToBlock = big.NewInt(curHeight)
		}

		// Filter by contract and event, write results to logs
		logs, err := ethRelayer.getFilterLogs(query)
		if err != nil {
			errinfo := fmt.Sprintf("Failed to filterLogEvents due to:%s", err.Error())
			panic(errinfo)
		}

		relayerLog.Info(title, "received logs with number", len(logs),
			"start height", query.FromBlock.String(), "stop height", query.ToBlock.String())
		for _, logv := range logs {
			relayerLog.Info(title, "received log with topics", logv.Topics[0].Hex(), "BlockNumber", logv.BlockNumber)
			if _, exist := eventSig[logv.Topics[0].Hex()]; !exist {
				continue
			}
			logchan <- logv
			relayerLog.Info(title, "get unprocessed log with topic:", logv.Topics[0].String(),
				"BlockNumber", logv.BlockNumber)
		}
		//更新
		if err := ethRelayer.setHeight4BridgeBankLogAt(query.ToBlock.Uint64()); nil != err {
			panic(err.Error())
		}

		if query.ToBlock.Int64() == curHeight {
			relayerLog.Info(title, "Finished FilterLogs to height", curHeight)
			done <- 1
			break
		}
		startHeight = query.ToBlock.Int64() + 1
	}
}

func (ethRelayer *Relayer4Ethereum) prePareSubscribeEvent() {
	var eventName string
	//bridgeBank处理
	contactAbi := ethtxs.LoadABI(ethtxs.BridgeBankABI)
	contactOracleAbi := ethtxs.LoadABI(ethtxs.OracleABI)
	ethRelayer.rwLock.Lock()
	ethRelayer.bridgeBankAbi = contactAbi
	eventName = events.LogLockFromETH.String()
	ethRelayer.bridgeBankEventLockSig = contactAbi.Events[eventName].ID.Hex()
	eventName = events.LogBurnFromETH.String()
	ethRelayer.bridgeBankEventBurnSig = contactAbi.Events[eventName].ID.Hex()
	ethRelayer.bridgeBankAddr = ethRelayer.x2EthDeployInfo.BridgeBank.Address

	ethRelayer.oracleAbi = contactOracleAbi
	eventName = events.LogProphecyProcessed.String()
	ethRelayer.oracleEventSig = contactOracleAbi.Events[eventName].ID.Hex()
	ethRelayer.oracleAddr = ethRelayer.x2EthDeployInfo.Oracle.Address
	ethRelayer.rwLock.Unlock()
}

func (ethRelayer *Relayer4Ethereum) subscribeEvent() {
	ethRelayer.rwLock.RLock()
	targetAddress := ethRelayer.bridgeBankAddr
	targetOracleAddress := ethRelayer.oracleAddr
	ethRelayer.rwLock.RUnlock()

	// We need the target address in bytes[] for the query
	//因为订阅事件的功能只会推送在订阅生效的高度之后的事件，所以FromBlock只需要填写１就可以了
	query := ethereum.FilterQuery{
		Addresses: []common.Address{targetAddress},
		FromBlock: big.NewInt(int64(1)),
	}

	if ethRelayer.delayedSend {
		query.Addresses = append(query.Addresses, targetOracleAddress)
	}

	// We will check logs for new events
	logs := make(chan types.Log, 10)

	for true {
		// Filter by contract and event, write results to logs
		sub, err := ethRelayer.clientWss.SubscribeFilterLogs(context.Background(), query, logs)
		if err != nil {
			relayerLog.Error("subscribeEvent", "Failed to SubscribeFilterLogs due to:", err.Error(), "bridgeBankAddr:", targetAddress.String())
			ethRelayer.getClientWss()
		}

		if err == nil {
			relayerLog.Info("subscribeEvent", "Subscribed to contract at address:", targetAddress.Hex())
			ethRelayer.bridgeBankSub = sub
			break
		}

		time.Sleep(sleepTime)
	}

	ethRelayer.bridgeBankLog = logs
}

//IsValidatorActive ...
func (ethRelayer *Relayer4Ethereum) IsValidatorActive(addr string) (bool, error) {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	if !re.MatchString(addr) {
		return false, errors.New("this address is not an ethereum address")
	}
	ethRelayer.rwLock.RLock()
	active, err := ethtxs.IsActiveValidator(common.HexToAddress(addr), ethRelayer.x2EthContracts.Valset)
	ethRelayer.rwLock.RUnlock()
	return active, err
}

//ShowOperator ...
func (ethRelayer *Relayer4Ethereum) ShowOperator() (string, error) {
	operator, err := ethtxs.GetOperator(ethRelayer.clientSpec, ethRelayer.ethSender, ethRelayer.bridgeBankAddr)
	if nil != err {
		return "", err
	}
	return operator.String(), nil
}

// handleLogLockEvent : unpacks a LogLock event, converts it to a ProphecyClaim, and relays a tx to chain33
func (ethRelayer *Relayer4Ethereum) handleLogLockEvent(clientChainID *big.Int, contractABI abi.ABI, eventName string, log types.Log) error {
	// Unpack the LogLock event using its unique event signature from the contract's ABI
	event, err := events.UnpackLogLock(contractABI, eventName, log.Data)
	if nil != err {
		return err
	}

	var decimal uint8
	tokenLocked, err := ethRelayer.GetLockedTokenAddress(event.Symbol)
	if nil == tokenLocked {
		//如果在本地没有找到该币种，则进行信息的收集和保存
		if event.Token.String() == "" || event.Token.String() == "0x0000000000000000000000000000000000000000" {
			decimal = 18
		} else {
			opts := &bind.CallOpts{
				Pending: true,
				From:    common.HexToAddress(event.Token.String()),
				Context: context.Background(),
			}
			bridgeToken, _ := generated.NewBridgeToken(common.HexToAddress(event.Token.String()), ethRelayer.clientSpec)
			decimal, err = bridgeToken.Decimals(opts)
			if err != nil {
				relayerLog.Error("handleLogLockEvent", "Failed to get Decimals due to", err.Error())
				return err
			}
		}

		token2set := &ebTypes.TokenAddress{
			Address:   event.Token.String(),
			Symbol:    event.Symbol,
			ChainName: ethRelayer.name,
			Decimal:   int32(decimal),
		}
		err = ethRelayer.SetLockedTokenAddress(token2set)
		if nil != err {
			relayerLog.Error("handleLogLockEvent", "Failed to SetLockedTokenAddress due to", err.Error())
			return errors.New("Failed ")
		}
	} else {
		decimal = uint8(tokenLocked.Decimal)
	}

	// Parse the LogLock event's payload into a struct
	ethTxhash := log.TxHash.String()
	prophecyClaim, err := ethtxs.LogLockToEthBridgeClaim(event, clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), ethTxhash, int64(decimal))
	if err != nil {
		return err
	}
	prophecyClaim.ChainName = ethRelayer.name

	fdIndex := int64(0)
	if !ethRelayer.processWithDraw {
		fdIndex = ethRelayer.getFdTx2Chain33TotalAmount() + 1
		prophecyClaim.ForwardIndex = fdIndex
		prophecyClaim.ForwardTimes = 1
	}
	ethRelayer.ethBridgeClaimChan <- prophecyClaim

	if ethRelayer.processWithDraw {
		//代理提币节点不需要记录标志
		return nil
	}
	//relayEthereum2chain33CheckPonit 1:send prophecyClaim to chain33 relay service
	relayerLog.Info("handleLogLockEvent::relayEthereum2chain33CheckPonit_1::sendClaim2Chain33", "ethTxhash", ethTxhash, "ForwardIndex", prophecyClaim.ForwardIndex, "FdTimes", prophecyClaim.ForwardTimes)
	_ = ethRelayer.updateFdTx2EthTotalAmount(fdIndex)
	txRelayConfirm4Chain33 := &ebTypes.TxRelayConfirm4Ethereum{
		EventType: int32(events.LogLockFromETH),
		Data:      log.Data,
		FdTimes:   1,
		FdIndex:   fdIndex,
		TxHash:    ethTxhash,
		Resend:    false,
	}
	return ethRelayer.setTxIsRelayedUnconfirm(ethRelayer.name, ethTxhash, fdIndex, txRelayConfirm4Chain33)
}

func (ethRelayer *Relayer4Ethereum) CreateLockEventManually(event *events.LockEvent) error {
	var decimal uint8
	tokenLocked, err := ethRelayer.GetLockedTokenAddress(event.Symbol)
	if nil == tokenLocked {
		//如果在本地没有找到该币种，则进行信息的收集和保存
		if event.Token.String() == "" || event.Token.String() == "0x0000000000000000000000000000000000000000" {
			decimal = 18
		} else {
			opts := &bind.CallOpts{
				Pending: true,
				From:    common.HexToAddress(event.Token.String()),
				Context: context.Background(),
			}
			bridgeToken, _ := generated.NewBridgeToken(common.HexToAddress(event.Token.String()), ethRelayer.clientSpec)
			decimal, err = bridgeToken.Decimals(opts)
			if err != nil {
				return err
			}
		}

		token2set := &ebTypes.TokenAddress{
			Address:   event.Token.String(),
			Symbol:    event.Symbol,
			ChainName: ethRelayer.name,
			Decimal:   int32(decimal),
		}
		err = ethRelayer.SetLockedTokenAddress(token2set)
		if nil != err {
			relayerLog.Error("CreateLockEventManually", "Failed to SetLockedTokenAddress due to", err.Error())
			return errors.New("Failed ")
		}
	} else {
		decimal = uint8(tokenLocked.Decimal)
	}

	prophecyClaim, err := ethtxs.LogLockToEthBridgeClaim(event, ethRelayer.clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), "0x1111111111111111111111111111111111111111111111111111111111111111", int64(decimal))
	if err != nil {
		return err
	}
	prophecyClaim.ChainName = ethRelayer.name
	ethRelayer.ethBridgeClaimChan <- prophecyClaim

	return nil
}

// handleLogBurnEvent : unpacks a burn event, converts it to a ProphecyClaim, and relays a tx to chain33
func (ethRelayer *Relayer4Ethereum) handleLogBurnEvent(clientChainID *big.Int, contractABI abi.ABI, eventName string, log types.Log) error {
	if ethRelayer.processWithDraw {
		//如果是代理提币中继器，则不进行消息的转发
		return nil
	}
	event, err := events.UnpackLogBurn(contractABI, eventName, log.Data)
	if nil != err {
		return err
	}

	var decimal uint8
	if event.Token.String() == "" || event.Token.String() == "0x0000000000000000000000000000000000000000" {
		decimal = 18
	} else {
		opts := &bind.CallOpts{
			Pending: true,
			From:    common.HexToAddress(event.Token.String()),
			Context: context.Background(),
		}
		bridgeToken, _ := generated.NewBridgeToken(common.HexToAddress(event.Token.String()), ethRelayer.clientSpec)

		decimal, err = bridgeToken.Decimals(opts)
		if err != nil {
			return err
		}
	}

	// Parse the LogLock event's payload into a struct
	prophecyClaim, err := ethtxs.LogBurnToEthBridgeClaim(event, clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), log.TxHash.String(), int64(decimal))
	if err != nil {
		return err
	}

	ethRelayer.ethBridgeClaimChan <- prophecyClaim

	return nil
}

func (ethRelayer *Relayer4Ethereum) ShowStatics(request *ebTypes.TokenStaticsRequest) (*ebTypes.TokenStaticsResponse, error) {
	res := &ebTypes.TokenStaticsResponse{}

	datas, err := ethRelayer.getStatics(request.Operation, request.TxIndex, request.Count)
	if nil != err {
		return nil, err
	}

	for _, data := range datas {
		var statics ebTypes.Chain33ToEthereumStatics
		_ = chain33Types.Decode(data, &statics)
		if request.Status != 0 && ebTypes.Tx_Status_Map[request.Status] != statics.EthTxstatus {
			continue
		}
		if len(request.Symbol) > 0 && request.Symbol != statics.Symbol {
			continue
		}
		res.C2Estatics = append(res.C2Estatics, &statics)
	}
	return res, nil
}

func (ethRelayer *Relayer4Ethereum) updateTxStatus() {
	if ethRelayer.processWithDraw {
		ethRelayer.updateSingleTxStatus(events.ClaimTypeWithdraw)
		return
	}
	ethRelayer.updateSingleTxStatus(events.ClaimTypeBurn)
	ethRelayer.updateSingleTxStatus(events.ClaimTypeLock)

}

func (ethRelayer *Relayer4Ethereum) updateSingleTxStatus(claimType events.ClaimType) {
	txIndex := ethRelayer.getEthLockTxUpdateTxIndex(claimType)
	datas, _ := ethRelayer.getStatics(int32(claimType), txIndex, 0)
	if nil == datas {
		relayerLog.Debug("ethRelayer::updateSingleTxStatus", "no new tx need to be update status for claimType", claimType, "from tx index", txIndex)
		return
	}
	for _, data := range datas {
		var statics ebTypes.Chain33ToEthereumStatics
		_ = chain33Types.Decode(data, &statics)
		receipt, _ := ethRelayer.getTransactionReceipt(common.HexToHash(statics.EthereumTxhash))
		//当前处理机制比较简单，如果发现该笔交易未执行，就不再产寻后续交易的回执
		if nil == receipt {
			break
		}
		status := ebTypes.Tx_Status_Success
		if receipt.Status != types.ReceiptStatusSuccessful {
			status = ebTypes.Tx_Status_Failed
		}
		statics.EthTxstatus = status
		dataNew := chain33Types.Encode(&statics)
		_ = ethRelayer.setLastestStatics(int32(claimType), statics.TxIndex, dataNew)
		_ = ethRelayer.setEthLockTxUpdateTxIndex(statics.TxIndex, claimType)
		relayerLog.Info("updateSingleTxStatus", "txHash", statics.EthereumTxhash, "updated status", status)
	}
}

func (ethRelayer *Relayer4Ethereum) SafeTransfer(para *ebTypes.SafeTransfer) (string, error) {
	if "" == ethRelayer.mulSignAddr {
		return "", ebTypes.ErrMulSignNotDeployed
	}

	return ethtxs.SafeTransfer(para.OperatorPrivateKey, ethRelayer.mulSignAddr, para.To, para.Token, para.OwnerPrivateKeys, para.Amount, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce, ethRelayer.providerHttp[0])
}

func (ethRelayer *Relayer4Ethereum) SetMultiSignAddr(address string) {
	ethRelayer.rwLock.Lock()
	ethRelayer.mulSignAddr = address
	ethRelayer.rwLock.Unlock()

	ethRelayer.setMultiSignAddress(address)
}

func (ethRelayer *Relayer4Ethereum) GetMultiSignAddr() string {
	return ethRelayer.getMultiSignAddress()
}

func (ethRelayer *Relayer4Ethereum) CfgWithdraw(symbol, feeAmount, amountPerDay string) error {
	fee, _ := big.NewInt(0).SetString(feeAmount, 10)
	amountPerDayInt, _ := big.NewInt(0).SetString(amountPerDay, 10)
	withdrawPara := &WithdrawFeeAndQuota{
		Fee:          fee,
		AmountPerDay: amountPerDayInt,
	}
	ethRelayer.rwLock.Lock()
	ethRelayer.withdrawFee[symbol] = withdrawPara
	ethRelayer.rwLock.Unlock()

	WithdrawPara := ethRelayer.restoreWithdrawFee()
	WithdrawPara[symbol] = &ebTypes.WithdrawPara{
		Fee:          feeAmount,
		AmountPerDay: amountPerDay,
	}

	return ethRelayer.setWithdrawFee(WithdrawPara)
}

func (ethRelayer *Relayer4Ethereum) GetCfgWithdraw(symbol string) *ebTypes.WithdrawPara {
	WithdrawPara := ethRelayer.restoreWithdrawFee()
	return WithdrawPara[symbol]
}

func (ethRelayer *Relayer4Ethereum) GetName() string {
	return ethRelayer.name
}

func (ethRelayer *Relayer4Ethereum) GeneralQuery(param, abiData, contract, owner string) (string, error) {
	return utils.QueryResult(param, abiData, contract, owner, ethRelayer.clientSpec)
}

func (ethRelayer *Relayer4Ethereum) sendEthereumTx(signedTx *types.Transaction) error {
	bSuccess := false
	var err error
	for i := 0; i < len(ethRelayer.clientSpecs); i++ {
		timeout, cancel := context.WithTimeout(context.Background(), waitTime)
		err = ethRelayer.clientSpecs[i].Client.SendTransaction(timeout, signedTx)
		cancel()
		if err == nil {
			bSuccess = true
		} else {
			if err.Error() != core.ErrAlreadyKnown.Error() {
				relayerLog.Error("handleLogWithdraw", "SendTransaction err", err)
			}
		}
	}

	// 交易同时发送到 BSC 官方节点
	if ethRelayer.name == ethtxs.BinanceChain {
		for i := 0; i < len(ethRelayer.clientBSCRecommendSpecs); i++ {
			timeout, cancel := context.WithTimeout(context.Background(), waitTime)
			err = ethRelayer.clientBSCRecommendSpecs[i].Client.SendTransaction(timeout, signedTx)
			cancel()
			if err == nil {
				bSuccess = true
			} else {
				if err.Error() != core.ErrAlreadyKnown.Error() {
					relayerLog.Error("handleLogWithdraw", "SendTransaction err", err)
				}
			}
		}
	}

	if bSuccess == false {
		return err
	} else {
		return nil
	}
}

func (ethRelayer *Relayer4Ethereum) remindSetupEthClientError() {
	ethName := "以太坊"
	if ethRelayer.name == ethtxs.BinanceChain {
		ethName = "BSC"
	}

	var remindEmail string
	for i := 0; i < len(ethRelayer.remindEmail); i++ {
		remindEmail += "\"" + ethRelayer.remindEmail[i] + "\""
		if i < len(ethRelayer.remindEmail)-1 {
			remindEmail += ","
		}
	}
	postData := fmt.Sprintf(`{"subject":"%s节点出错","receiver":[%s],"content":"节点 %s 连接失败"}`, ethName, remindEmail, ethRelayer.providerHttp)
	relayerLog.Info("SendRemind", "remindClientErrorUrl", ethRelayer.remindClientErrorUrl, "remindEmail", ethRelayer.remindEmail, "postData:", postData)
	ethRelayer.SendRemind(ethRelayer.remindClientErrorUrl, postData)
}

func (ethRelayer *Relayer4Ethereum) regainClient(isSendEmail *bool) {
	// 重新获取 client
	var err error
	ethRelayer.clientSpecs, _, err = ethtxs.SetupEthClients(&ethRelayer.providerHttp, ethRelayer.bridgeRegistryAddr)
	if err != nil {
		relayerLog.Error("regainClient", "SetupEthClient err", err)
	}

	if len(ethRelayer.clientSpecs) == 0 && *isSendEmail == false {
		// 节点都不可用 发送邮件
		ethRelayer.remindSetupEthClientError()
		*isSendEmail = true
	}
}

func (ethRelayer *Relayer4Ethereum) getFilterLogs(query ethereum.FilterQuery) ([]types.Log, error) {
	isSendEmail := false
	for {
		for i := 0; i < len(ethRelayer.clientSpecs); i++ {
			timeout, cancel := context.WithTimeout(context.Background(), waitTime)
			logs, err := ethRelayer.clientSpecs[i].Client.FilterLogs(timeout, query)
			cancel()
			if err == nil {
				return logs, nil
			} else {
				relayerLog.Error("getFilterLogs", "FilterLogs err", err)
				ethRelayer.clientSpecs = append(ethRelayer.clientSpecs[:0], ethRelayer.clientSpecs[1:]...)
				i--
			}
		}

		time.Sleep(time.Second)
		ethRelayer.regainClient(&isSendEmail)
	}
}

func (ethRelayer *Relayer4Ethereum) getTransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	for i := 0; i < len(ethRelayer.clientSpecs); i++ {
		timeout, cancel := context.WithTimeout(context.Background(), waitTime)
		receipt, err := ethRelayer.clientSpecs[i].Client.TransactionReceipt(timeout, txHash)
		cancel()
		if err == nil {
			return receipt, nil
		} else {
			if err.Error() == "not found" {
				return nil, err
			}
			ethRelayer.clientSpecs = append(ethRelayer.clientSpecs[:0], ethRelayer.clientSpecs[1:]...)
			i--
			relayerLog.Error("getTransactionReceipt", "TransactionReceipt err", err, "txhash", txHash.String())
		}
	}

	return nil, errors.New("TransactionReceipt err")
}

func (ethRelayer *Relayer4Ethereum) getHeaderByNumber() (uint64, error) {
	isSendEmail := false
	for {
		for i := 0; i < len(ethRelayer.clientSpecs); i++ {
			timeout, cancel := context.WithTimeout(context.Background(), waitTime)
			head, err := ethRelayer.clientSpecs[i].Client.HeaderByNumber(timeout, nil)
			cancel()
			if err == nil {
				return head.Number.Uint64(), nil
			} else {
				ethRelayer.clientSpecs = append(ethRelayer.clientSpecs[:0], ethRelayer.clientSpecs[1:]...)
				i--
				relayerLog.Error("getHeaderByNumber", "getHeaderByNumber err", err)
			}
		}

		time.Sleep(time.Second)
		ethRelayer.regainClient(&isSendEmail)
	}
}

func (ethRelayer *Relayer4Ethereum) getBalanceAt(addr common.Address) (*big.Int, error) {
	isSendEmail := false
	for j := 0; j < 2; j++ {
		for i := 0; i < len(ethRelayer.clientSpecs); i++ {
			timeout, cancel := context.WithTimeout(context.Background(), waitTime)
			balance, err := ethRelayer.clientSpecs[i].Client.BalanceAt(timeout, addr, nil)
			cancel()
			if err == nil {
				return balance, nil
			} else {
				ethRelayer.clientSpecs = append(ethRelayer.clientSpecs[:0], ethRelayer.clientSpecs[1:]...)
				i--
				relayerLog.Error("getBalanceAt", "getBalanceAt err", err)
			}
		}

		time.Sleep(time.Second)
		ethRelayer.regainClient(&isSendEmail)
	}

	return nil, errors.New("getBalanceAt err")
}

func (ethRelayer *Relayer4Ethereum) getCallContract(call ethereum.CallMsg) ([]byte, error) {
	isSendEmail := false
	for j := 0; j < 2; j++ {
		for i := 0; i < len(ethRelayer.clientSpecs); i++ {
			timeout, cancel := context.WithTimeout(context.Background(), waitTime)
			result, err := ethRelayer.clientSpecs[i].Client.CallContract(timeout, call, nil)
			cancel()
			if err == nil {
				return result, nil
			} else {
				ethRelayer.clientSpecs = append(ethRelayer.clientSpecs[:0], ethRelayer.clientSpecs[1:]...)
				i--
				relayerLog.Error("getCallContract", "getCallContract err", err)
			}
		}

		time.Sleep(time.Second)
		ethRelayer.regainClient(&isSendEmail)
	}
	return nil, errors.New("getCallContract err")
}
