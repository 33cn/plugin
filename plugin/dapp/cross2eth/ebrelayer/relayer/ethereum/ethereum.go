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

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"

	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethtxs"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/events"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

//Relayer4Ethereum ...
type Relayer4Ethereum struct {
	name               string //链的名字，用于区分不同的链
	provider           string
	providerHttp       string
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
	clientWss               ethinterface.EthClientSpec
	bridgeBankAddr          common.Address
	bridgeBankSub           ethereum.Subscription
	bridgeBankLog           chan types.Log
	bridgeBankEventLockSig  string
	bridgeBankEventBurnSig  string
	bridgeBankAbi           abi.ABI
	x2EthDeployInfo         *ethtxs.X2EthDeployInfo
	deployPara              *ethtxs.DeployPara
	operatorInfo            *ethtxs.OperatorInfo
	x2EthContracts          *ethtxs.X2EthContracts
	ethBridgeClaimChan      chan<- *ebTypes.EthBridgeClaim
	chain33MsgChan          <-chan *events.Chain33Msg
	totalTxRelayFromChain33 int64
	symbol2Addr             map[string]common.Address
	symbol2LockAddr         map[string]ebTypes.TokenAddress
	mulSignAddr             string
	withdrawFee             map[string]*WithdrawFeeAndQuota
	Addr2TxNonce            map[common.Address]*ethtxs.NonceMutex
}

var (
	relayerLog = log.New("module", "ethereum_relayer")
)

const (
	DefaultBlockPeriod = 5000
	EthereumChain      = "Ethereum"
	BinanceChain       = "Binance"
	USDT               = "USDT"
)

type EthereumStartPara struct {
	DbHandle           dbm.DB
	EthProvider        string
	EthProviderHttp    string
	BridgeRegistryAddr string
	Degree             int32
	BlockInterval      int32
	EthBridgeClaimChan chan<- *ebTypes.EthBridgeClaim
	Chain33MsgChan     <-chan *events.Chain33Msg
	ProcessWithDraw    bool
	Name               string
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
		provider:                startPara.EthProvider,
		providerHttp:            startPara.EthProviderHttp,
		db:                      startPara.DbHandle,
		unlockchan:              make(chan int, 2),
		bridgeRegistryAddr:      common.HexToAddress(startPara.BridgeRegistryAddr),
		processWithDraw:         startPara.ProcessWithDraw,
		maturityDegree:          startPara.Degree,
		fetchHeightPeriodMs:     startPara.BlockInterval,
		ethBridgeClaimChan:      startPara.EthBridgeClaimChan,
		chain33MsgChan:          startPara.Chain33MsgChan,
		totalTxRelayFromChain33: 0,
		symbol2Addr:             make(map[string]common.Address),
		symbol2LockAddr:         make(map[string]ebTypes.TokenAddress),
		Addr2TxNonce:            make(map[common.Address]*ethtxs.NonceMutex),
	}

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

	// Start clientSpec with infura ropsten provider
	relayerLog.Info("Relayer4Ethereum proc", "Started Ethereum websocket with provider:", ethRelayer.provider, "processWithDraw", ethRelayer.processWithDraw)
	client, err := ethtxs.SetupWebsocketEthClient(ethRelayer.providerHttp)
	if err != nil {
		panic(err)
	}
	ethRelayer.clientSpec = client

	ethRelayer.clientWss, err = ethtxs.SetupWebsocketEthClient(ethRelayer.provider)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	clientChainID, err := client.NetworkID(ctx)
	if err != nil {
		errinfo := fmt.Sprintf("Failed to get NetworkID due to:%s", err.Error())
		panic(errinfo)
	}
	ethRelayer.clientChainID = clientChainID
	ethRelayer.totalTxRelayFromChain33 = ethRelayer.getTotalTxAmount2Eth()
	if 0 == ethRelayer.totalTxRelayFromChain33 {
		statics := &ebTypes.Ethereum2Chain33Statics{}
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
		ethRelayer.x2EthContracts.BridgeBank, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce)
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
	return ethtxs.TransferToken(tokenAddr, fromKey, toAddr, bn, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce)
}

//TransferEth ...
func (ethRelayer *Relayer4Ethereum) TransferEth(fromKey, toAddr, amount string) (string, error) {
	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(amount), 10)
	return ethtxs.TransferEth(fromKey, toAddr, bn, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce)
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
	return ethtxs.LockEthErc20Asset(ownerPrivateKey, tokenAddr, chain33Receiver, bn, ethRelayer.clientSpec, ethRelayer.x2EthContracts.BridgeBank, ethRelayer.x2EthDeployInfo.BridgeBank.Address, ethRelayer.Addr2TxNonce)
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
	return ethRelayer.clientSpec.TransactionReceipt(context.Background(), txhash)
}

func (ethRelayer *Relayer4Ethereum) proc() {
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
	ctx := context.Background()
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
			goto burnLockWithdrawProc
		}
	}

burnLockWithdrawProc:
	for {
		select {
		case <-timer.C:
			ethRelayer.procNewHeight(ctx)
		case err := <-ethRelayer.bridgeBankSub.Err():
			relayerLog.Error("proc", "bridgeBankSub err", err.Error())
			ethRelayer.subscribeEvent()
			ethRelayer.filterLogEvents()
		case vLog := <-ethRelayer.bridgeBankLog:
			ethRelayer.storeBridgeBankLogs(vLog, true)
		case chain33Msg := <-ethRelayer.chain33MsgChan:
			ethRelayer.handleChain33Msg(chain33Msg)
		}
	}
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

func (ethRelayer *Relayer4Ethereum) handleLogWithdraw(chain33Msg *events.Chain33Msg) {
	//只有通过代理人登录的中继器，才处理提币事件
	var err error
	now := time.Now()
	year, month, day := now.Date()
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
	defer func() {
		if err != nil {
			withdrawTx.Status = int32(ethtxs.WDError)
			withdrawTx.StatusDescription = ethtxs.WDError.String()
			withdrawTx.ErrorDescription = err.Error()
			relayerLog.Error("handleLogWithdraw", "Failed to withdraw due to:", err.Error())
		}

		err := ethRelayer.setWithdraw(withdrawTx)
		if nil != err {
			relayerLog.Error("handleLogWithdraw", "Failed to setWithdraw due to:", err.Error())
		}

		err = ethRelayer.setWithdrawStatics(withdrawTx, chain33Msg)
		if nil != err {
			relayerLog.Error("handleLogWithdraw", "Failed to setWithdrawStatics due to:", err.Error())
		}
	}()

	relayerLog.Info("handleLogWithdraw", "Received chain33Msg", chain33Msg, "tx hash string", common.Bytes2Hex(chain33Msg.TxHash))
	withdrawFromChain33TokenInfo, exist := ethRelayer.symbol2LockAddr[chain33Msg.Symbol]
	if !exist {
		//因为是withdraw操作，必须从允许lock的token地址中进行查询
		relayerLog.Error("handleLogWithdraw", "Failed to fetch locked Token Info for symbol", chain33Msg.Symbol)
		err = errors.New("ErrFetchLockedTokenInfo")
		return
	}

	tokenAddr := common.HexToAddress(withdrawFromChain33TokenInfo.Address)
	//从chain33进行withdraw回来的token需要根据精度进行相应的缩放
	if 8 != withdrawFromChain33TokenInfo.Decimal {
		dist := math.Abs(float64(withdrawFromChain33TokenInfo.Decimal - 8))
		value, exist := utils.Decimal2value[int(dist)]
		if !exist {
			relayerLog.Error("handleLogWithdraw", "does support for decimal, %d", withdrawFromChain33TokenInfo.Decimal)
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
		"Receiver on Ethereum", chain33Msg.EthereumReceiver.String())
	//检查用户提币权限是否得到满足：比如是否超过累计提币额度
	var feeAmount *big.Int
	if feeAmount, err = ethRelayer.checkPermissionWithinOneDay(withdrawTx); nil != err {
		return
	}
	if chain33Msg.Amount.Cmp(feeAmount) < 0 {
		relayerLog.Error("handleLogWithdraw", "ErrWithdrawAmountLessThanFee feeAmount", feeAmount.String(), "Withdraw Amount", chain33Msg.Amount.String())
		err = errors.New("ErrWithdrawAmountCan'tPay4Fee")
		return
	}
	amount2transfer := chain33Msg.Amount.Sub(chain33Msg.Amount, feeAmount)
	value := big.NewInt(0)

	//此处需要完成在以太坊发送以太或者ERC20数字资产的操作
	ctx := context.Background()
	timeout, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	var intputData []byte // ERC20 or BEP20 token transfer pack data
	var toAddr common.Address
	var balanceOfData []byte // ERC20 or BEP20 token balanceof pack data

	if tokenAddr.String() != ethtxs.EthNullAddr { //判断是否要Pack EVM数据
		toAddr = tokenAddr
		intputData, err = ethRelayer.packTransferData(chain33Msg.EthereumReceiver, amount2transfer)
		if err != nil {
			relayerLog.Error("handleLogWithdraw", "CallEvmData err", err)
			err = errors.New("ErrPackTransferData")
			return
		}
		//用签名的账户地址作为pack参数，toAddr作为合约地址
		balanceOfData, err = ethRelayer.packBalanceOfData(ethRelayer.ethSender)
		if err != nil {
			relayerLog.Error("handleLogWithdraw", "callEvmBalanceData err", err)
			err = errors.New("ErrPackBalanceOfData")
			return
		}
	} else {
		//如果tokenAddr为空，则把toAddr设置为用户指定的地址
		toAddr = chain33Msg.EthereumReceiver
		value = amount2transfer
	}

	//校验余额是否充足
	err = ethRelayer.checkBalanceEnough(toAddr, amount2transfer, balanceOfData)
	if err != nil {
		relayerLog.Error("handleLogWithdraw", "Failed to checkBalanceEnough:", err.Error())
		err = errors.New("ErrBalanceNotEnough")
		return
	}
	//param: from,to,evm-packdata,amount
	//交易构造
	tx, err := ethtxs.NewTransferTx(ethRelayer.clientSpec, ethRelayer.ethSender, toAddr, intputData, value, ethRelayer.Addr2TxNonce)
	if err != nil {
		relayerLog.Error("handleLogWithdraw", "newTx err", err)
		err = errors.New("ErrNewTx")
		return
	}

	//交易签名
	signedTx, err := ethRelayer.signTx(tx, ethRelayer.privateKey4Ethereum)
	if err != nil {
		relayerLog.Error("handleLogWithdraw", "SignTx err", err)
		err = errors.New("ErrSignTx")
		return
	}
	//交易发送
	err = ethRelayer.clientSpec.SendTransaction(timeout, signedTx)
	if err != nil {
		relayerLog.Error("handleLogWithdraw", "SendTransaction err", err)
		err = errors.New("ErrSendTransaction")
		return
	}
	relayerLog.Info("handleLogWithdraw", "SendTransaction Hash", signedTx.Hash())

	withdrawTx.Status = int32(ethtxs.WDPending)
	withdrawTx.StatusDescription = ethtxs.WDPending.String()
	withdrawTx.TxHashOnEthereum = signedTx.Hash().String()

	return
}

func (ethRelayer *Relayer4Ethereum) checkBalanceEnough(addr common.Address, amount *big.Int, inputdata []byte) error {
	//检测地址余额
	var balance *big.Int
	var err error
	if inputdata == nil {
		balance, err = ethRelayer.clientSpec.BalanceAt(context.Background(), addr, nil)
		if err != nil {
			//retry
			balance, err = ethRelayer.clientSpec.BalanceAt(context.Background(), addr, nil)
			if err != nil {
				return err
			}
		}
	} else {
		var msg ethereum.CallMsg
		msg.To = &addr //合约地址
		msg.Data = inputdata
		result, err := ethRelayer.clientSpec.CallContract(context.Background(), msg, nil)
		if err != nil {
			//retry
			result, err = ethRelayer.clientSpec.CallContract(context.Background(), msg, nil)
			if err != nil {
				return err
			}
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

func (ethRelayer *Relayer4Ethereum) handleLogLockBurn(chain33Msg *events.Chain33Msg) {
	//对于通过代理人登录的中继器，不处理lock和burn事件
	if ethRelayer.processWithDraw {
		relayerLog.Info("handleLogLockBurn", "Needn't process lock and burn for this withdraw process specified validator", ethRelayer.ethSender)
		return
	}
	relayerLog.Info("handleLogLockBurn", "Received chain33Msg", chain33Msg, "tx hash string", common.Bytes2Hex(chain33Msg.TxHash))

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

	// Relay the Chain33Msg to the Ethereum network
	txhash, err := ethtxs.RelayOracleClaimToEthereum(ethRelayer.x2EthContracts.Oracle, ethRelayer.clientSpec, ethRelayer.ethSender, tokenAddr, prophecyClaim, ethRelayer.privateKey4Ethereum, ethRelayer.Addr2TxNonce)
	if nil != err {
		panic("RelayOracleClaimToEthereum failed due to" + err.Error())
	}
	relayerLog.Info("handleLogLockBurn", "RelayOracleClaimToEthereum with tx hash", txhash)

	//保存交易hash，方便查询
	txIndex := atomic.AddInt64(&ethRelayer.totalTxRelayFromChain33, 1)
	if err = ethRelayer.updateTotalTxAmount2chain33(txIndex); nil != err {
		relayerLog.Error("handleLogLockBurn", "Failed to RelayLockToChain33 due to:", err.Error())
		return
	}
	statics := &ebTypes.Chain33ToEthereumStatics{
		EthTxstatus:      ebTypes.Tx_Status_Pending,
		Chain33Txhash:    common.Bytes2Hex(chain33Msg.TxHash),
		EthereumTxhash:   txhash,
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

func (ethRelayer *Relayer4Ethereum) getCurrentHeight(ctx context.Context) (uint64, error) {
	head, err := ethRelayer.clientSpec.HeaderByNumber(ctx, nil)
	if nil == err {
		return head.Number.Uint64(), nil
	}

	//TODO: 需要在下面添加报警处理
	for {
		time.Sleep(5 * time.Second)
		ethRelayer.clientSpec, err = ethtxs.SetupWebsocketEthClient(ethRelayer.providerHttp)
		if err != nil {
			relayerLog.Error("getCurrentHeight", "Failed to SetupWebsocketEthClient due to:", err.Error())
			continue
		}
		head, err := ethRelayer.clientSpec.HeaderByNumber(ctx, nil)
		if nil != err {
			relayerLog.Error("getCurrentHeight", "Failed to HeaderByNumber due to:", err.Error())
			continue
		}
		relayerLog.Debug("getCurrentHeight", "clientSpec SetupWebsocketEthClient:", ethRelayer.providerHttp)
		return head.Number.Uint64(), nil
	}
}

func (ethRelayer *Relayer4Ethereum) procNewHeight4Withdraw(ctx context.Context) {
	currentHeight, _ := ethRelayer.getCurrentHeight(ctx)
	relayerLog.Info("procNewHeight4Withdraw", "currentHeight", currentHeight)
}

func (ethRelayer *Relayer4Ethereum) procNewHeight(ctx context.Context) {
	currentHeight, _ := ethRelayer.getCurrentHeight(ctx)
	ethRelayer.updateTxStatus()
	//currentHeight := head.Number.Uint64()
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
	}

	//确定是否需要更新保存同步日志高度
	if setBlockNumber {
		if err := ethRelayer.setHeight4BridgeBankLogAt(vLog.BlockNumber); nil != err {
			panic(err.Error())
		}
	}
}

func (ethRelayer *Relayer4Ethereum) procBridgeBankLogs(vLog types.Log) {
	if ethRelayer.checkTxProcessed(vLog.TxHash.Bytes()) {
		relayerLog.Info("procBridgeBankLogs", "Tx has been already Processed with hash:", vLog.TxHash.Hex(),
			"height", vLog.BlockNumber, "index", vLog.Index)
		return
	}

	defer func() {
		if err := ethRelayer.setTxProcessed(vLog.TxHash.Bytes()); nil != err {
			panic(err.Error())
		}
	}()

	//检查当前交易是否因为区块回退而导致交易丢失
	receipt, err := ethRelayer.clientSpec.TransactionReceipt(context.Background(), vLog.TxHash)
	if nil != err {
		relayerLog.Error("procBridgeBankLogs", "Failed to get tx receipt with hash", vLog.TxHash.String())
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
			relayerLog.Info("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
			panic(errinfo)
		}
	} else if vLog.Topics[0].Hex() == ethRelayer.bridgeBankEventBurnSig {
		//burn,用于捕捉 (chain33 token----->chain33) 实现chain33资产withdraw操作，之后在chain33上实现unlock操作
		eventName := events.LogBurnFromETH.String()
		relayerLog.Info("Relayer4Ethereum proc", "Going to process", eventName,
			"Block number:", vLog.BlockNumber, "Tx hash:", vLog.TxHash.Hex())
		err := ethRelayer.handleLogBurnEvent(ethRelayer.clientChainID, ethRelayer.bridgeBankAbi, eventName, vLog)
		if err != nil {
			errinfo := fmt.Sprintf("Failed to handleLogBurnEvent due to:%s", err.Error())
			relayerLog.Info("Relayer4Ethereum procBridgeBankLogs", "errinfo", errinfo)
			panic(errinfo)
		}
	}
}

//因为订阅事件的功能只会推送在订阅生效的高度之后的事件，之前订阅停止～当前订阅生效高度的这一段只能通过FilterLogs来获取事件信息，否则就会遗漏
func (ethRelayer *Relayer4Ethereum) filterLogEvents() {
	deployHeight, _ := ethtxs.GetDeployHeight(ethRelayer.clientSpec, ethRelayer.x2EthDeployInfo.BridgeRegistry.Address, ethRelayer.x2EthDeployInfo.BridgeRegistry.Address)
	height4BridgeBankLogAt := int64(ethRelayer.getHeight4BridgeBankLogAt())

	if height4BridgeBankLogAt < deployHeight {
		height4BridgeBankLogAt = deployHeight
	}

	curHeightUint64, _ := ethRelayer.getCurrentHeight(context.Background())
	curHeight := int64(curHeightUint64)
	relayerLog.Info("filterLogEvents", "curHeight:", curHeight)

	bridgeBankSig := make(map[string]bool)
	bridgeBankSig[ethRelayer.bridgeBankEventLockSig] = true
	bridgeBankSig[ethRelayer.bridgeBankEventBurnSig] = true
	bridgeBankLog := make(chan types.Log)
	done := make(chan int)

	go ethRelayer.filterLogEventsProc(bridgeBankLog, done, "bridgeBank", curHeight, height4BridgeBankLogAt, ethRelayer.bridgeBankAddr, bridgeBankSig)

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
func (ethRelayer *Relayer4Ethereum) filterLogEventsProc(logchan chan<- types.Log, done chan<- int, title string, curHeight, heightLogProcAt int64, contractAddr common.Address, eventSig map[string]bool) {
	relayerLog.Info(title, "eventSig", eventSig, "heightLogProcAt", heightLogProcAt, "curHeight", curHeight)

	startHeight := heightLogProcAt
	batchCount := int64(10)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
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
		logs, err := ethRelayer.clientSpec.FilterLogs(context.Background(), query)
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
	ethRelayer.bridgeBankAbi = contactAbi
	eventName = events.LogLockFromETH.String()
	ethRelayer.bridgeBankEventLockSig = contactAbi.Events[eventName].ID.Hex()
	eventName = events.LogBurnFromETH.String()
	ethRelayer.bridgeBankEventBurnSig = contactAbi.Events[eventName].ID.Hex()
	ethRelayer.bridgeBankAddr = ethRelayer.x2EthDeployInfo.BridgeBank.Address
}

func (ethRelayer *Relayer4Ethereum) subscribeEvent() {
	targetAddress := ethRelayer.bridgeBankAddr

	// We need the target address in bytes[] for the query
	//因为订阅事件的功能只会推送在订阅生效的高度之后的事件，所以FromBlock只需要填写１就可以了
	query := ethereum.FilterQuery{
		Addresses: []common.Address{targetAddress},
		FromBlock: big.NewInt(int64(1)),
	}
	// We will check logs for new events
	logs := make(chan types.Log, 10)
	// Filter by contract and event, write results to logs
	sub, err := ethRelayer.clientWss.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		errinfo := fmt.Sprintf("Failed to SubscribeFilterLogs due to:%s, bridgeBankAddr:%s", err.Error(), ethRelayer.bridgeBankAddr)
		panic(errinfo)
	}
	relayerLog.Info("subscribeEvent", "Subscribed to contract at address:", targetAddress.Hex())
	ethRelayer.bridgeBankLog = logs
	ethRelayer.bridgeBankSub = sub
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
				return err
			}
		}

		token2set := ebTypes.TokenAddress{
			Address:   event.Token.String(),
			Symbol:    event.Symbol,
			ChainName: ebTypes.EthereumBlockChainName,
			Decimal:   int32(decimal),
		}
		err = ethRelayer.SetLockedTokenAddress(token2set)
		if nil != err {
			relayerLog.Error("handleChain33Msg", "Failed to SetLockedTokenAddress due to", err.Error())
			return errors.New("Failed ")
		}
	} else {
		decimal = uint8(tokenLocked.Decimal)
	}

	// Parse the LogLock event's payload into a struct
	prophecyClaim, err := ethtxs.LogLockToEthBridgeClaim(event, clientChainID.Int64(), ethRelayer.bridgeBankAddr.String(), log.TxHash.String(), int64(decimal))
	if err != nil {
		return err
	}
	prophecyClaim.ChainName = ethRelayer.name
	//如果不是以太坊的USDT,则需要将其铸币为XUSD,如Binance的USDT，则铸币为BUSD
	if prophecyClaim.Symbol == "USDT" && EthereumChain != ethRelayer.name {
		prophecyClaim.Symbol = ethRelayer.name[0:1] + "USDT"
		prophecyClaim.Symbol = strings.ToUpper(prophecyClaim.Symbol)
	}

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
	ethRelayer.updateSingleTxStatus(events.ClaimTypeBurn)
	ethRelayer.updateSingleTxStatus(events.ClaimTypeLock)
	ethRelayer.updateSingleTxStatus(events.ClaimTypeWithdraw)
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
		receipt, _ := ethRelayer.clientSpec.TransactionReceipt(context.Background(), common.HexToHash(statics.EthereumTxhash))
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

	return ethtxs.SafeTransfer(para.OperatorPrivateKey, ethRelayer.mulSignAddr, para.To, para.Token, para.OwnerPrivateKeys, para.Amount, ethRelayer.clientSpec, ethRelayer.Addr2TxNonce)
}

func (ethRelayer *Relayer4Ethereum) SetMultiSignAddr(address string) {
	ethRelayer.rwLock.Lock()
	ethRelayer.mulSignAddr = address
	ethRelayer.rwLock.Unlock()

	ethRelayer.setMultiSignAddress(address)
}

func (ethRelayer *Relayer4Ethereum) CfgWithdraw(symbol string, feeAmount, amountPerDay string) error {
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

func (ethRelayer *Relayer4Ethereum) GetName() string {
	return ethRelayer.name
}
