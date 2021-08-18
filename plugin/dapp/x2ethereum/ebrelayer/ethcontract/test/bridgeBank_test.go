package test

//
//import (
//	"context"
//	"fmt"
//	"math/big"
//	"testing"
//	"time"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/generated"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethcontract/test/setup"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/ethtxs"
//	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/events"
//	"github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/ethereum/go-ethereum/crypto"
//	"github.com/stretchr/testify/require"
//)
//
////"BridgeToken creation (Chain33 assets)"
//func TestBrigeTokenCreat(t *testing.T) {
//	ctx := context.Background()
//	println("TEST:BridgeToken creation (Chain33 assets)")
//	//1st部署相关合约
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//
//	//2nd：订阅事件
//	eventName := "LogNewBridgeToken"
//	bridgeBankABI := ethtxs.LoadABI(ethtxs.BridgeBankABI)
//	logNewBridgeTokenSig := bridgeBankABI.Events[eventName].ID.Hex()
//	query := ethereum.FilterQuery{
//		Addresses: []common.Address{x2EthDeployInfo.BridgeBank.Address},
//	}
//	// We will check logs for new events
//	logs := make(chan types.Log)
//	// Filter by contract and event, write results to logs
//	sub, err := sim.SubscribeFilterLogs(ctx, query, logs)
//	require.Nil(t, err)
//
//	t.Logf("x2EthDeployInfo.BridgeBank.Address is:%s", x2EthDeployInfo.BridgeBank.Address.String())
//	bridgeBank, err := generated.NewBridgeBank(x2EthDeployInfo.BridgeBank.Address, sim)
//	require.Nil(t, err)
//
//	opts := &bind.CallOpts{
//		Pending: true,
//		From:    para.Operator,
//		Context: ctx,
//	}
//	BridgeBankAddr, err := x2EthContracts.BridgeRegistry.BridgeBank(opts)
//	require.Nil(t, err)
//	t.Logf("BridgeBankAddr is:%s", BridgeBankAddr.String())
//
//	tokenCount, err := bridgeBank.BridgeBankCaller.BridgeTokenCount(opts)
//	require.Nil(t, err)
//	require.Equal(t, tokenCount.Int64(), int64(0))
//
//	//3rd：创建token
//	auth, err := ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	if nil != err {
//		t.Fatalf("PrepareAuth failed due to:%s", err.Error())
//	}
//	symbol := "BTY"
//	_, err = x2EthContracts.BridgeBank.BridgeBankTransactor.CreateNewBridgeToken(auth, symbol)
//	if nil != err {
//		t.Fatalf("CreateNewBridgeToken failed due to:%s", err.Error())
//	}
//	sim.Commit()
//
//	timer := time.NewTimer(30 * time.Second)
//	for {
//		select {
//		case <-timer.C:
//			t.Fatal("failed due to timeout")
//		// Handle any errors
//		case err := <-sub.Err():
//			t.Fatalf("sub error:%s", err.Error())
//		// vLog is raw event data
//		case vLog := <-logs:
//			// Check if the event is a 'LogLock' event
//			if vLog.Topics[0].Hex() == logNewBridgeTokenSig {
//				t.Logf("Witnessed new event:%s, Block number:%d, Tx hash:%s", eventName,
//					vLog.BlockNumber, vLog.TxHash.Hex())
//				logEvent := &events.LogNewBridgeToken{}
//				//err = bridgeBankABI.Unpack(logEvent, eventName, vLog.Data)
//				_, err = bridgeBankABI.Unpack(eventName, vLog.Data)
//				require.Nil(t, err)
//				t.Logf("token addr:%s, symbol:%s", logEvent.Token.String(), logEvent.Symbol)
//				require.Equal(t, symbol, logEvent.Symbol)
//
//				//tokenCount正确加1
//				tokenCount, err := x2EthContracts.BridgeBank.BridgeTokenCount(opts)
//				require.Nil(t, err)
//				require.Equal(t, tokenCount.Int64(), int64(1))
//
//				return
//			}
//		}
//	}
//}
//
////测试在chain33上锁定资产,然后在以太坊上铸币
////发行token="BTY"
////铸币NewOracleClaim
////铸币成功
////Bridge token minting (for locked chain33 assets)
//func TestBrigeTokenMint(t *testing.T) {
//	ctx := context.Background()
//	println("TEST:BridgeToken creation (Chain33 assets)")
//	//1st部署相关合约
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//
//	//2nd：订阅事件
//	eventName := "LogNewBridgeToken"
//	bridgeBankABI := ethtxs.LoadABI(ethtxs.BridgeBankABI)
//	logNewBridgeTokenSig := bridgeBankABI.Events[eventName].ID.Hex()
//	query := ethereum.FilterQuery{
//		Addresses: []common.Address{x2EthDeployInfo.BridgeBank.Address},
//	}
//	// We will check logs for new events
//	logs := make(chan types.Log)
//	// Filter by contract and event, write results to logs
//	sub, err := sim.SubscribeFilterLogs(ctx, query, logs)
//	require.Nil(t, err)
//
//	opts := &bind.CallOpts{
//		Pending: true,
//		From:    para.Operator,
//		Context: ctx,
//	}
//
//	tokenCount, err := x2EthContracts.BridgeBank.BridgeTokenCount(opts)
//	require.Nil(t, err)
//	require.Equal(t, tokenCount.Int64(), int64(0))
//
//	//3rd：创建token
//	symbol := "BTY"
//	auth, err := ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	if nil != err {
//		t.Fatalf("PrepareAuth failed due to:%s", err.Error())
//	}
//	_, err = x2EthContracts.BridgeBank.BridgeBankTransactor.CreateNewBridgeToken(auth, symbol)
//	if nil != err {
//		t.Fatalf("CreateNewBridgeToken failed due to:%s", err.Error())
//	}
//	sim.Commit()
//
//	logEvent := &events.LogNewBridgeToken{}
//	select {
//	// Handle any errors
//	case err := <-sub.Err():
//		t.Fatalf("sub error:%s", err.Error())
//	// vLog is raw event data
//	case vLog := <-logs:
//		// Check if the event is a 'LogLock' event
//		if vLog.Topics[0].Hex() == logNewBridgeTokenSig {
//			t.Logf("Witnessed new event:%s, Block number:%d, Tx hash:%s", eventName,
//				vLog.BlockNumber, vLog.TxHash.Hex())
//
//			_, err = bridgeBankABI.Unpack(eventName, vLog.Data)
//			require.Nil(t, err)
//			t.Logf("token addr:%s, symbol:%s", logEvent.Token.String(), logEvent.Symbol)
//			require.Equal(t, symbol, logEvent.Symbol)
//
//			//tokenCount正确加1
//			tokenCount, err = x2EthContracts.BridgeBank.BridgeTokenCount(opts)
//			require.Nil(t, err)
//			require.Equal(t, tokenCount.Int64(), int64(1))
//			break
//		}
//	}
//
//	///////////newOracleClaim///////////////////////////
//	balance, _ := sim.BalanceAt(ctx, para.InitValidators[0], nil)
//	fmt.Println("InitValidators[0] addr,", para.InitValidators[0].String(), "balance =", balance.String())
//
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	amount := int64(99)
//	ethReceiver := para.InitValidators[2]
//	claimID := crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), logEvent.Token.Bytes(), big.NewInt(amount).Bytes())
//
//	authOracle, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	signature, err := ethtxs.SignClaim4Eth(claimID, para.ValidatorPriKey[0])
//	require.Nil(t, err)
//
//	bridgeToken, err := generated.NewBridgeToken(logEvent.Token, sim)
//	require.Nil(t, err)
//	opts = &bind.CallOpts{
//		Pending: true,
//		Context: ctx,
//	}
//
//	balance, err = bridgeToken.BalanceOf(opts, ethReceiver)
//	require.Nil(t, err)
//	require.Equal(t, balance.Int64(), int64(0))
//
//	_, err = x2EthContracts.Oracle.NewOracleClaim(
//		authOracle,
//		events.ClaimTypeLock,
//		chain33Sender,
//		ethReceiver,
//		logEvent.Token,
//		logEvent.Symbol,
//		big.NewInt(amount),
//		claimID,
//		signature)
//	require.Nil(t, err)
//
//	sim.Commit()
//	balance, err = bridgeToken.BalanceOf(opts, ethReceiver)
//	require.Nil(t, err)
//	require.Equal(t, balance.Int64(), amount)
//	t.Logf("The minted amount is:%d", balance.Int64())
//}
//
////测试在以太坊上lock数字资产,包括Eth和Erc20
////Bridge deposit locking (deposit erc20/eth assets)
//func TestBridgeDepositLock(t *testing.T) {
//	ctx := context.Background()
//	println("TEST:Bridge deposit locking (Erc20/Eth assets)")
//	//1st部署相关合约
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//
//	//创建token
//	operatorAuth, err := ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	assert.Nil(t, err)
//	symbol := "USDT"
//	bridgeTokenAddr, _, bridgeTokenInstance, err := generated.DeployBridgeToken(operatorAuth, sim, symbol)
//	require.Nil(t, err)
//	sim.Commit()
//	t.Logf("The new creaded symbol:%s, address:%s", symbol, bridgeTokenAddr.String())
//
//	//创建实例 为userOne铸币 userOne为bridgebank允许allowance设置数额
//	userOne := para.InitValidators[0]
//	callopts := &bind.CallOpts{
//		Pending: true,
//		From:    userOne,
//		Context: ctx,
//	}
//	symQuery, err := bridgeTokenInstance.Symbol(callopts)
//	assert.Nil(t, err)
//	require.Equal(t, symQuery, symbol)
//	t.Logf("symQuery = %s", symQuery)
//
//	isMiner, err := bridgeTokenInstance.IsMinter(callopts, para.Operator)
//	require.Nil(t, err)
//	require.Equal(t, isMiner, true)
//
//	operatorAuth, err = ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	require.Nil(t, err)
//
//	mintAmount := int64(1000)
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	_, err = bridgeTokenInstance.Mint(operatorAuth, userOne, big.NewInt(mintAmount))
//	require.Nil(t, err)
//	sim.Commit()
//
//	userOneAuth, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//	allowAmount := int64(100)
//	_, err = bridgeTokenInstance.Approve(userOneAuth, x2EthDeployInfo.BridgeBank.Address, big.NewInt(allowAmount))
//	require.Nil(t, err)
//	sim.Commit()
//
//	userOneBalance, err := bridgeTokenInstance.BalanceOf(callopts, userOne)
//	require.Nil(t, err)
//	t.Logf("userOneBalance:%d", userOneBalance.Int64())
//	require.Equal(t, userOneBalance.Int64(), mintAmount)
//
//	// 测试子项目:should allow users to lock ERC20 tokens
//	userOneAuth, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	//lock 100
//	lockAmount := big.NewInt(100)
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, bridgeTokenAddr, lockAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	//balance减少到900
//	userOneBalance, err = bridgeTokenInstance.BalanceOf(callopts, userOne)
//	require.Nil(t, err)
//	expectAmount := int64(900)
//	require.Equal(t, userOneBalance.Int64(), expectAmount)
//	t.Logf("userOneBalance changes to:%d", userOneBalance.Int64())
//
//	//bridgebank增加了100
//	bridgeBankBalance, err := bridgeTokenInstance.BalanceOf(callopts, x2EthDeployInfo.BridgeBank.Address)
//	require.Nil(t, err)
//	expectAmount = int64(100)
//	require.Equal(t, bridgeBankBalance.Int64(), expectAmount)
//	t.Logf("bridgeBankBalance changes to:%d", bridgeBankBalance.Int64())
//
//	// 测试子项目:锁定ETH，should allow users to lock Ethereum
//	bridgeBankBalance, err = sim.BalanceAt(ctx, x2EthDeployInfo.BridgeBank.Address, nil)
//	require.Nil(t, err)
//	t.Logf("origin eth bridgeBankBalance is:%d", bridgeBankBalance.Int64())
//
//	userOneAuth, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//	ethAmount := big.NewInt(50)
//	userOneAuth.Value = ethAmount
//
//	//lock 50 eth
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	bridgeBankBalance, err = sim.BalanceAt(ctx, x2EthDeployInfo.BridgeBank.Address, nil)
//	require.Nil(t, err)
//	require.Equal(t, bridgeBankBalance.Int64(), ethAmount.Int64())
//	t.Logf("eth bridgeBankBalance changes to:%d", bridgeBankBalance.Int64())
//}
//
////测试在以太坊上unlock数字资产,包括Eth和Erc20,
////即从chain33取回在eth上发行的的ETH或ERC20数字资产，之前通过lock操作发送到了chain33
////现在则通过NewProphecyClaim 的burn操作将数字资产取回
////Ethereum/ERC20 token unlocking (for burned chain33 assets)
//func TestBridgeBankUnlock(t *testing.T) {
//	ctx := context.Background()
//	println("TEST:Ethereum/ERC20 token unlocking (for burned chain33 assets)")
//	//1st部署相关合约
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//	//1.lockEth资产
//	ethAddr := common.Address{}
//	userOneAuth, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	ethLockAmount := big.NewInt(150)
//	userOneAuth.Value = ethLockAmount
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	//lock 150 eth
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethLockAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	//2.lockErc20资产
//	//创建token
//	operatorAuth, err := ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	assert.Nil(t, err)
//	symbolUsdt := "USDT"
//	bridgeTokenAddr, _, bridgeTokenInstance, err := generated.DeployBridgeToken(operatorAuth, sim, symbolUsdt)
//	require.Nil(t, err)
//	sim.Commit()
//	t.Logf("The new creaded symbolUsdt:%s, address:%s", symbolUsdt, bridgeTokenAddr.String())
//
//	//创建实例
//	//为userOne铸币
//	//userOne为bridgebank允许allowance设置数额
//	userOne := para.InitValidators[0]
//	callopts := &bind.CallOpts{
//		Pending: true,
//		From:    userOne,
//		Context: ctx,
//	}
//	symQuery, err := bridgeTokenInstance.Symbol(callopts)
//	assert.Nil(t, err)
//	require.Equal(t, symQuery, symbolUsdt)
//	t.Logf("symQuery = %s", symQuery)
//
//	isMiner, err := bridgeTokenInstance.IsMinter(callopts, para.Operator)
//	require.Nil(t, err)
//	require.Equal(t, isMiner, true)
//
//	operatorAuth, err = ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	assert.Nil(t, err)
//	mintAmount := int64(1000)
//	_, err = bridgeTokenInstance.Mint(operatorAuth, userOne, big.NewInt(mintAmount))
//	require.Nil(t, err)
//	sim.Commit()
//
//	userOneAuth, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	assert.Nil(t, err)
//	allowAmount := int64(100)
//	_, err = bridgeTokenInstance.Approve(userOneAuth, x2EthDeployInfo.BridgeBank.Address, big.NewInt(allowAmount))
//	require.Nil(t, err)
//	sim.Commit()
//
//	userOneBalance, err := bridgeTokenInstance.BalanceOf(callopts, userOne)
//	require.Nil(t, err)
//	t.Logf("userOneBalance:%d", userOneBalance.Int64())
//	require.Equal(t, userOneBalance.Int64(), mintAmount)
//
//	//***测试子项目:should allow users to lock ERC20 tokens
//	userOneAuth, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	//lock 100
//	lockAmount := big.NewInt(100)
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, bridgeTokenAddr, lockAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	// newOracleClaim
//	newProphecyAmount := int64(55)
//	ethReceiver := para.InitValidators[2]
//	ethSym := "eth"
//	claimID := crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), ethAddr.Bytes(), big.NewInt(newProphecyAmount).Bytes())
//
//	authOracle, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	signature, err := ethtxs.SignClaim4Eth(claimID, para.ValidatorPriKey[0])
//	require.Nil(t, err)
//
//	_, err = x2EthContracts.Oracle.NewOracleClaim(
//		authOracle,
//		events.ClaimTypeBurn,
//		chain33Sender,
//		ethReceiver,
//		ethAddr,
//		ethSym,
//		big.NewInt(newProphecyAmount),
//		claimID,
//		signature)
//	require.Nil(t, err)
//
//	userEthbalance, _ := sim.BalanceAt(ctx, ethReceiver, nil)
//	t.Logf("userEthbalance for addr:%s balance=%d", ethReceiver.String(), userEthbalance.Int64())
//
//	sim.Commit()
//	userEthbalanceAfter, _ := sim.BalanceAt(ctx, ethReceiver, nil)
//	t.Logf("userEthbalance after ProcessBridgeProphecy for addr:%s balance=%d", ethReceiver.String(), userEthbalanceAfter.Int64())
//	require.Equal(t, userEthbalance.Int64()+newProphecyAmount, userEthbalanceAfter.Int64())
//
//	//////////////////////////////////////////////////////////////////
//	///////should unlock and transfer ERC20 tokens upon the processing of a burn prophecy//////
//	//////////////////////////////////////////////////////////////////
//	// newOracleClaim
//	newProphecyAmount = int64(100)
//	ethReceiver = para.InitValidators[2]
//	claimID = crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), bridgeTokenAddr.Bytes(), big.NewInt(newProphecyAmount).Bytes())
//
//	authOracle, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	signature, err = ethtxs.SignClaim4Eth(claimID, para.ValidatorPriKey[0])
//	require.Nil(t, err)
//
//	_, err = x2EthContracts.Oracle.NewOracleClaim(
//		authOracle,
//		events.ClaimTypeBurn,
//		chain33Sender,
//		ethReceiver,
//		bridgeTokenAddr,
//		symbolUsdt,
//		big.NewInt(newProphecyAmount),
//		claimID,
//		signature)
//	require.Nil(t, err)
//
//	userUSDTbalance, err := bridgeTokenInstance.BalanceOf(callopts, ethReceiver)
//	require.Nil(t, err)
//	t.Logf("userEthbalance for addr:%s balance=%d", ethReceiver.String(), userUSDTbalance.Int64())
//	require.Equal(t, userUSDTbalance.Int64(), newProphecyAmount)
//}
//
////测试在以太坊上多次unlock数字资产Eth
////Ethereum/ERC20 token second unlocking (for burned chain33 assets)
//func TestBridgeBankSecondUnlockEth(t *testing.T) {
//	ctx := context.Background()
//	println("TEST:to be unlocked incrementally by successive burn prophecies (for burned chain33 assets)")
//	//1st部署相关合约
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//
//	//1.lockEth资产
//	ethAddr := common.Address{}
//	userOneAuth, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	ethLockAmount := big.NewInt(150)
//	userOneAuth.Value = ethLockAmount
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	//lock 150 eth
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethLockAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	//2.lockErc20资产
//	//创建token
//	operatorAuth, err := ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	assert.Nil(t, err)
//	symbolUsdt := "USDT"
//	bridgeTokenAddr, _, bridgeTokenInstance, err := generated.DeployBridgeToken(operatorAuth, sim, symbolUsdt)
//	require.Nil(t, err)
//	sim.Commit()
//	t.Logf("The new creaded symbolUsdt:%s, address:%s", symbolUsdt, bridgeTokenAddr.String())
//
//	//创建实例
//	//为userOne铸币
//	//userOne为bridgebank允许allowance设置数额
//	userOne := para.InitValidators[0]
//	callopts := &bind.CallOpts{
//		Pending: true,
//		From:    userOne,
//		Context: ctx,
//	}
//	symQuery, err := bridgeTokenInstance.Symbol(callopts)
//	assert.Nil(t, err)
//	require.Equal(t, symQuery, symbolUsdt)
//	t.Logf("symQuery = %s", symQuery)
//
//	isMiner, err := bridgeTokenInstance.IsMinter(callopts, para.Operator)
//	require.Nil(t, err)
//	require.Equal(t, isMiner, true)
//
//	operatorAuth, err = ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	require.Nil(t, err)
//	mintAmount := int64(1000)
//	_, err = bridgeTokenInstance.Mint(operatorAuth, userOne, big.NewInt(mintAmount))
//	require.Nil(t, err)
//	sim.Commit()
//
//	userOneAuth, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	assert.Nil(t, err)
//	allowAmount := int64(100)
//	_, err = bridgeTokenInstance.Approve(userOneAuth, x2EthDeployInfo.BridgeBank.Address, big.NewInt(allowAmount))
//	require.Nil(t, err)
//	sim.Commit()
//
//	userOneBalance, err := bridgeTokenInstance.BalanceOf(callopts, userOne)
//	require.Nil(t, err)
//	t.Logf("userOneBalance:%d", userOneBalance.Int64())
//	require.Equal(t, userOneBalance.Int64(), mintAmount)
//
//	//***测试子项目:should allow users to lock ERC20 tokens
//	userOneAuth, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	//lock 100
//	lockAmount := big.NewInt(100)
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, bridgeTokenAddr, lockAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	// newOracleClaim
//	newProphecyAmount := int64(44)
//	ethReceiver := para.InitValidators[2]
//	ethSym := "eth"
//	claimID := crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), ethAddr.Bytes(), big.NewInt(newProphecyAmount).Bytes())
//
//	authOracle, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	signature, err := ethtxs.SignClaim4Eth(claimID, para.ValidatorPriKey[0])
//	require.Nil(t, err)
//
//	_, err = x2EthContracts.Oracle.NewOracleClaim(
//		authOracle,
//		events.ClaimTypeBurn,
//		chain33Sender,
//		ethReceiver,
//		ethAddr,
//		ethSym,
//		big.NewInt(newProphecyAmount),
//		claimID,
//		signature)
//	require.Nil(t, err)
//
//	userEthbalance, _ := sim.BalanceAt(ctx, ethReceiver, nil)
//	t.Logf("userEthbalance for addr:%s balance=%d", ethReceiver.String(), userEthbalance.Int64())
//
//	sim.Commit()
//
//	userEthbalanceAfter, _ := sim.BalanceAt(ctx, ethReceiver, nil)
//	t.Logf("userEthbalance after ProcessBridgeProphecy for addr:%s balance=%d", ethReceiver.String(), userEthbalanceAfter.Int64())
//	require.Equal(t, userEthbalance.Int64()+newProphecyAmount, userEthbalanceAfter.Int64())
//
//	//第二次 newOracleClaim
//	newProphecyAmountSecond := int64(33)
//	claimID = crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), ethAddr.Bytes(), big.NewInt(newProphecyAmountSecond).Bytes())
//	authOracle, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	signature, err = ethtxs.SignClaim4Eth(claimID, para.ValidatorPriKey[0])
//	require.Nil(t, err)
//
//	_, err = x2EthContracts.Oracle.NewOracleClaim(
//		authOracle,
//		events.ClaimTypeBurn,
//		chain33Sender,
//		ethReceiver,
//		ethAddr,
//		ethSym,
//		big.NewInt(newProphecyAmountSecond),
//		claimID,
//		signature)
//	require.Nil(t, err)
//
//	userEthbalance, _ = sim.BalanceAt(ctx, ethReceiver, nil)
//	t.Logf("userEthbalance for addr:%s balance=%d", ethReceiver.String(), userEthbalance.Int64())
//
//	sim.Commit()
//	userEthbalanceAfter, _ = sim.BalanceAt(ctx, ethReceiver, nil)
//	t.Logf("userEthbalance after ProcessBridgeProphecy for addr:%s balance=%d", ethReceiver.String(), userEthbalanceAfter.Int64())
//	require.Equal(t, userEthbalance.Int64()+newProphecyAmountSecond, userEthbalanceAfter.Int64())
//}
//
////测试在以太坊上多次unlock数字资产Erc20
////Ethereum/ERC20 token unlocking (for burned chain33 assets)
//func TestBridgeBankSedondUnlockErc20(t *testing.T) {
//	ctx := context.Background()
//	println("TEST:ERC20 to be unlocked incrementally by successive burn prophecies (for burned chain33 assets))")
//	//1st部署相关合约
//	para, sim, x2EthContracts, x2EthDeployInfo, err := setup.DeployContracts()
//	require.NoError(t, err)
//
//	//1.lockEth资产
//	userOneAuth, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//	ethLockAmount := big.NewInt(150)
//	userOneAuth.Value = ethLockAmount
//	chain33Sender := []byte("14KEKbYtKKQm4wMthSK9J4La4nAiidGozt")
//	//lock 150 eth
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, common.Address{}, ethLockAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	//2.lockErc20资产
//	//创建token
//	operatorAuth, err := ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	assert.Nil(t, err)
//	symbolUsdt := "USDT"
//	bridgeTokenAddr, _, bridgeTokenInstance, err := generated.DeployBridgeToken(operatorAuth, sim, symbolUsdt)
//	require.Nil(t, err)
//	sim.Commit()
//	t.Logf("The new creaded symbolUsdt:%s, address:%s", symbolUsdt, bridgeTokenAddr.String())
//
//	//创建实例 为userOne铸币 userOne为bridgebank允许allowance设置数额
//	userOne := para.InitValidators[0]
//	callopts := &bind.CallOpts{
//		Pending: true,
//		From:    userOne,
//		Context: ctx,
//	}
//	symQuery, err := bridgeTokenInstance.Symbol(callopts)
//	assert.Nil(t, err)
//	require.Equal(t, symQuery, symbolUsdt)
//	t.Logf("symQuery = %s", symQuery)
//	isMiner, err := bridgeTokenInstance.IsMinter(callopts, para.Operator)
//	require.Nil(t, err)
//	require.Equal(t, isMiner, true)
//
//	operatorAuth, err = ethtxs.PrepareAuth(sim, para.DeployPrivateKey, para.Operator)
//	require.Nil(t, err)
//	mintAmount := int64(1000)
//	_, err = bridgeTokenInstance.Mint(operatorAuth, userOne, big.NewInt(mintAmount))
//	require.Nil(t, err)
//	sim.Commit()
//
//	userOneAuth, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//	allowAmount := int64(100)
//	_, err = bridgeTokenInstance.Approve(userOneAuth, x2EthDeployInfo.BridgeBank.Address, big.NewInt(allowAmount))
//	require.Nil(t, err)
//	sim.Commit()
//
//	userOneBalance, err := bridgeTokenInstance.BalanceOf(callopts, userOne)
//	require.Nil(t, err)
//	t.Logf("userOneBalance:%d", userOneBalance.Int64())
//	require.Equal(t, userOneBalance.Int64(), mintAmount)
//
//	//测试子项目:should allow users to lock ERC20 tokens
//	userOneAuth, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//	//lock 100
//	lockAmount := big.NewInt(100)
//	_, err = x2EthContracts.BridgeBank.Lock(userOneAuth, chain33Sender, bridgeTokenAddr, lockAmount)
//	require.Nil(t, err)
//	sim.Commit()
//
//	// newOracleClaim
//	newProphecyAmount := int64(33)
//	ethReceiver := para.InitValidators[2]
//	claimID := crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), bridgeTokenAddr.Bytes(), big.NewInt(newProphecyAmount).Bytes())
//
//	userUSDTbalance0, err := bridgeTokenInstance.BalanceOf(callopts, ethReceiver)
//	require.Nil(t, err)
//	t.Logf("userEthbalance for addr:%s balance=%d", ethReceiver.String(), userUSDTbalance0.Int64())
//	require.Equal(t, userUSDTbalance0.Int64(), int64(0))
//
//	///////////newOracleClaim///////////////////////////
//	authOracle, err := ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	signature, err := ethtxs.SignClaim4Eth(claimID, para.ValidatorPriKey[0])
//	require.Nil(t, err)
//	_, err = x2EthContracts.Oracle.NewOracleClaim(
//		authOracle,
//		events.ClaimTypeBurn,
//		chain33Sender,
//		ethReceiver,
//		bridgeTokenAddr,
//		symbolUsdt,
//		big.NewInt(newProphecyAmount),
//		claimID,
//		signature)
//	require.Nil(t, err)
//
//	userUSDTbalance1, err := bridgeTokenInstance.BalanceOf(callopts, ethReceiver)
//	require.Nil(t, err)
//	t.Logf("userEthbalance for addr:%s balance=%d", ethReceiver.String(), userUSDTbalance1.Int64())
//	require.Equal(t, userUSDTbalance1.Int64(), userUSDTbalance0.Int64()+newProphecyAmount)
//
//	// newOracleClaim
//	newProphecyAmountSecond := int64(66)
//	claimID = crypto.Keccak256Hash(chain33Sender, ethReceiver.Bytes(), bridgeTokenAddr.Bytes(), big.NewInt(newProphecyAmountSecond).Bytes())
//	authOracle, err = ethtxs.PrepareAuth(sim, para.ValidatorPriKey[0], para.InitValidators[0])
//	require.Nil(t, err)
//
//	signature, err = ethtxs.SignClaim4Eth(claimID, para.ValidatorPriKey[0])
//	require.Nil(t, err)
//	_, err = x2EthContracts.Oracle.NewOracleClaim(
//		authOracle,
//		events.ClaimTypeBurn,
//		chain33Sender,
//		ethReceiver,
//		bridgeTokenAddr,
//		symbolUsdt,
//		big.NewInt(newProphecyAmountSecond),
//		claimID,
//		signature)
//	require.Nil(t, err)
//
//	userUSDTbalance2, err := bridgeTokenInstance.BalanceOf(callopts, ethReceiver)
//	require.Nil(t, err)
//	t.Logf("userEthbalance for addr:%s balance=%d", ethReceiver.String(), userUSDTbalance2.Int64())
//	require.Equal(t, userUSDTbalance2.Int64(), userUSDTbalance1.Int64()+newProphecyAmountSecond)
//}
