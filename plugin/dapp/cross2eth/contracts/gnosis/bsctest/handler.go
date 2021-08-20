package main

//
//import (
//	"context"
//	"crypto/ecdsa"
//	"github.com/ethereum/go-ethereum/accounts/abi"
//	"gitlab.33.cn/pancake/gnosis/bsctest/erc20"
//	"gitlab.33.cn/pancake/gnosis/bsctest/gnosisSafe"
//
//	//"errors"
//	"fmt"
//	"github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/crypto"
//	"github.com/ethereum/go-ethereum/ethclient"
//	"math/big"
//	"strings"
//	"time"
//)
//
//var TestNodeAddr = "https://data-seed-prebsc-1-s1.binance.org:8545/"
//var ethClient *ethclient.Client
//var privateKey *ecdsa.PrivateKey
//var deployerAddr common.Address
//
//var privateKeyStrs = []string{
//	"f934e9171c5cf13b35e6c989e95f5e95fa471515730af147b66d60fbcd664b7c", //addr:0x0f2e821517D4f64a012a04b668a6b1aa3B262e08
//	"2bcf3e23a17d3f3b190a26a098239ad2d20267a673440e0f57a23f44f94b77b9", //addr:0x21B5f4C2F6Ff418fa0067629D9D76AE03fB4a2d2
//	"e5f8caae6468061c17543bc2205c8d910b3c71ad4d055105cde94e88ccb4e650", //addr:0xee760B2E502244016ADeD3491948220B3b1dd789
//}
//
////const ...
//const (
//	// GasLimit : the gas limit in Gwei used for transactions sent with TransactOpts
//	GasLimitTxExec  = uint64(100 * 10000)
//	GasLimit4Deploy = uint64(0) //此处需要设置为0,让交易自行估计,否则将会导致部署失败,TODO:其他解决途径后续调研解决
//	fee2setter      = "0x0f2e821517D4f64a012a04b668a6b1aa3B262e08"
//)
//
//func init() {
//	var err error
//	ethClient, err = SetupWebsocketEthClient(TestNodeAddr)
//	if nil != err {
//		panic(fmt.Sprintf("Failed to SetupWebsocketEthClient with url:%s", TestNodeAddr))
//	}
//	fmt.Println("Succeed to establish connection to bsc")
//}
//
//// SetupWebsocketEthClient : returns boolean indicating if a URL is valid websocket ethclient
//func SetupWebsocketEthClient(ethURL string) (*ethclient.Client, error) {
//	if strings.TrimSpace(ethURL) == "" {
//		return nil, nil
//	}
//
//	client, err := ethclient.Dial(ethURL)
//	if err != nil {
//		return nil, fmt.Errorf("error dialing websocket client %w", err)
//	}
//
//	return client, nil
//}
//
//func GetBalance(tokenAddr, owner string) (string, error) {
//	//查询ETH余额
//	if "" != tokenAddr {
//		token, err := erc20.NewERC20(common.HexToAddress(tokenAddr), ethClient)
//		if nil != err {
//			return "", err
//		}
//		opts := &bind.CallOpts{
//			From:    deployerAddr,
//			Context: context.Background(),
//		}
//		balance, err := token.BalanceOf(opts, common.HexToAddress(owner))
//		if nil != err {
//			return "", err
//		}
//		return balance.String(), nil
//
//	}
//	balance, err := ethClient.BalanceAt(context.Background(), common.HexToAddress(owner), nil)
//	if nil != err {
//		return "", err
//	}
//	return balance.String(), nil
//}
//
//func getNonce(sender common.Address) (*big.Int, error) {
//	nonce, err := ethClient.PendingNonceAt(context.Background(), sender)
//	if nil != err {
//		return nil, err
//	}
//	return big.NewInt(int64(nonce)), nil
//}
//
////PrepareAuth ...
//func PrepareAuth(privateKey *ecdsa.PrivateKey, transactor common.Address, gasLimit uint64) (*bind.TransactOpts, error) {
//	//ctx := context.Background()
//	//gasPrice, err := ethClient.SuggestGasPrice(ctx)
//	//if err != nil {
//	//	return nil, errors.New("failed to get suggest gas price")
//	//}
//	//bind.NewKeyedTransactorWithChainID
//	auth := bind.NewKeyedTransactor(privateKey)
//	auth.Value = big.NewInt(0) // in wei
//	auth.GasLimit = gasLimit
//	//auth.GasPrice = gasPrice
//
//	var err error
//	if auth.Nonce, err = getNonce(transactor); err != nil {
//		return nil, err
//	}
//
//	return auth, nil
//}
//
//func recoverBinancePrivateKey() (err error) {
//	privateKey, err = crypto.ToECDSA(common.FromHex("2bcf3e23a17d3f3b190a26a098239ad2d20267a673440e0f57a23f44f94b77b9"))
//	if nil != err {
//		panic("Failed to recover private key")
//		return err
//	}
//	deployerAddr = crypto.PubkeyToAddress(privateKey.PublicKey)
//	fmt.Println("The address is:", deployerAddr.String())
//
//	return nil
//}
//
////Succeed to deploy DeployMulSign with address = 0x24bA77a96f5BFCC7e8dbBDdA5cA749CEe5b13661
//func DeployMulSign() error {
//	_ = recoverBinancePrivateKey()
//	//1st step to deploy factory
//	auth, err := PrepareAuth(privateKey, deployerAddr, GasLimit4Deploy)
//	if nil != err {
//		return err
//	}
//
//	cakeTokenAddr, deploycakeTokenTx, _, err := gnosisSafe.DeployGnosisSafe(auth, ethClient)
//	if nil != err {
//		panic(fmt.Sprintf("Failed to DeployGnosisSafe with err:%s", err.Error()))
//		return err
//	}
//
//	{
//		fmt.Println("\nDeployGnosisSafe tx hash:", deploycakeTokenTx.Hash().String())
//		timeout := time.NewTimer(300 * time.Second)
//		oneSecondtimeout := time.NewTicker(5 * time.Second)
//		for {
//			select {
//			case <-timeout.C:
//				panic("DeployGnosisSafe timeout")
//			case <-oneSecondtimeout.C:
//				_, err := ethClient.TransactionReceipt(context.Background(), deploycakeTokenTx.Hash())
//				if err == ethereum.NotFound {
//					fmt.Println("\n No receipt received yet for DeployGnosisSafe tx and continue to wait")
//					continue
//				} else if err != nil {
//					panic("DeployGnosisSafe failed due to" + err.Error())
//				}
//				fmt.Println("\n Succeed to deploy DeployGnosisSafe with address =", cakeTokenAddr.String())
//				return nil
//			}
//		}
//	}
//
//	return nil
//}
//
////threshold int, to, paymentToken, paymentReceiver string, payment int64
//func SetupOwnerProc(safe string) error {
//	owners := []string{"0x0f2e821517D4f64a012a04b668a6b1aa3B262e08", "0xee760B2E502244016ADeD3491948220B3b1dd789", "0x21B5f4C2F6Ff418fa0067629D9D76AE03fB4a2d2"}
//	_ = recoverBinancePrivateKey()
//	auth, err := PrepareAuth(privateKey, deployerAddr, GasLimitTxExec)
//	if nil != err {
//		return err
//	}
//
//	gnosisSafeAddr := common.HexToAddress(safe)
//	gnosisSafeInt, err := gnosisSafe.NewGnosisSafe(gnosisSafeAddr, ethClient)
//	if nil != err {
//		return err
//	}
//
//	//_owners []common.Address, _threshold *big.Int, to common.Address, data []byte,
//	// fallbackHandler common.Address, paymentToken common.Address,
//	// payment *big.Int, paymentReceiver common.Address
//	var _owners []common.Address
//	for _, onwer := range owners {
//		_owners = append(_owners, common.HexToAddress(onwer))
//	}
//	AddressZero := common.HexToAddress("0x0000000000000000000000000000000000000000")
//
//	//safe.setup([user1.address, user2.address], 1, AddressZero, "0x", handler.address, AddressZero, 0, AddressZero)
//	setupTx, err := gnosisSafeInt.Setup(auth, _owners, big.NewInt(int64(len(_owners))), AddressZero, []byte{'0', 'x'}, AddressZero, AddressZero, big.NewInt(int64(0)), AddressZero)
//	if nil != err {
//		panic(fmt.Sprintf("Failed to setupTx with err:%s", err.Error()))
//		return err
//	}
//
//	{
//		fmt.Println("\nsetupTx tx hash:", setupTx.Hash().String())
//		timeout := time.NewTimer(300 * time.Second)
//		oneSecondtimeout := time.NewTicker(5 * time.Second)
//		for {
//			select {
//			case <-timeout.C:
//				panic("setupTx timeout")
//			case <-oneSecondtimeout.C:
//				_, err := ethClient.TransactionReceipt(context.Background(), setupTx.Hash())
//				if err == ethereum.NotFound {
//					fmt.Println("\n No receipt received yet for setupTx  and continue to wait")
//					continue
//				} else if err != nil {
//					panic("SetupOwner failed due to" + err.Error())
//				}
//				fmt.Println("\n Succeed to setup Tx")
//				return nil
//			}
//		}
//	}
//
//	return nil
//}
//
//func TransferProc(safe, to, token string, fValue float64) error {
//	owners := []string{"0x0f2e821517D4f64a012a04b668a6b1aa3B262e08",
//		"0xee760B2E502244016ADeD3491948220B3b1dd789",
//		"0x21B5f4C2F6Ff418fa0067629D9D76AE03fB4a2d2"}
//	_ = recoverBinancePrivateKey()
//	auth, err := PrepareAuth(privateKey, deployerAddr, GasLimitTxExec)
//	if nil != err {
//		return err
//	}
//
//	gnosisSafeAddr := common.HexToAddress(safe)
//	gnosisSafeInt, err := gnosisSafe.NewGnosisSafe(gnosisSafeAddr, ethClient)
//	if nil != err {
//		return err
//	}
//	AddressZero := common.HexToAddress("0x0000000000000000000000000000000000000000")
//
//	//_owners []common.Address, _threshold *big.Int, to common.Address, data []byte,
//	// fallbackHandler common.Address, paymentToken common.Address,
//	// payment *big.Int, paymentReceiver common.Address
//	var _owners []common.Address
//	for _, onwer := range owners {
//		_owners = append(_owners, common.HexToAddress(onwer))
//	}
//
//	//opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, signatures []byte
//
//	_to := common.HexToAddress(to)
//	_data := []byte{'0', 'x'}
//	safeTxGas := big.NewInt(10 * 10000)
//	baseGas := big.NewInt(0)
//	gasPrice := big.NewInt(0)
//	var value *big.Int = big.NewInt(int64(fValue * 1e18))
//	opts := &bind.CallOpts{
//		From:    deployerAddr,
//		Context: context.Background(),
//	}
//	//token transfer
//	if token != "" {
//		_to = common.HexToAddress(token)
//
//		erc20Abi, err := abi.JSON(strings.NewReader(erc20.ERC20ABI))
//		if err != nil {
//			return err
//		}
//
//		tokenInstance, err := erc20.NewERC20(_to, ethClient)
//		if err != nil {
//			return err
//		}
//		decimals, err := tokenInstance.Decimals(opts)
//		if err != nil {
//			return err
//		}
//		mul := int64(1)
//		for i:= 0; i < int(decimals); i++ {
//			mul *= 10
//		}
//		value = big.NewInt(int64(fValue * float64(mul)))
//
//		//const data = token.interface.encodeFunctionData("transfer", [address, 500])
//		_data, err = erc20Abi.Pack("transfer", common.HexToAddress(to), value)
//		if err != nil {
//			return err
//		}
//		//对于erc20这种方式 最后需要将其设置为0
//		value = big.NewInt(0)
//	}
//
//	nonce, err := gnosisSafeInt.Nonce(opts)
//	if err != nil {
//		panic("Failed to get Nonce")
//		return err
//	}
//
//	//opts *bind.TransactOpts, to common.Address, value *big.Int, data []byte, operation uint8, safeTxGas *big.Int, baseGas *big.Int, gasPrice *big.Int, gasToken common.Address, refundReceiver common.Address, _nonce *big.Int
//	signContent, err := gnosisSafeInt.GetTransactionHash(opts, _to, value, _data, 0,
//		safeTxGas, baseGas, gasPrice, AddressZero, AddressZero, nonce)
//	if err != nil {
//		panic("Failed to GetTransactionHash")
//		return err
//	}
//	fmt.Println("safe.Nonce =", nonce.String(), "safe.Nonce(int64) =", nonce.Int64())
//	sigs := buildSigs(signContent[:])
//
//	execTx, err := gnosisSafeInt.ExecTransaction(auth, _to, value, _data, 0,
//		safeTxGas, baseGas, gasPrice, AddressZero, AddressZero, sigs)
//	if nil != err {
//		panic(fmt.Sprintf("Failed to ExecTransaction with err:%s", err.Error()))
//		return err
//	}
//
//	{
//		fmt.Println("\nExecTransaction tx hash:", execTx.Hash().String())
//		timeout := time.NewTimer(300 * time.Second)
//		oneSecondtimeout := time.NewTicker(5 * time.Second)
//		for {
//			select {
//			case <-timeout.C:
//				panic("ExecTransaction timeout")
//			case <-oneSecondtimeout.C:
//				_, err := ethClient.TransactionReceipt(context.Background(), execTx.Hash())
//				if err == ethereum.NotFound {
//					fmt.Println("\n No receipt received yet for ExecTransaction  and continue to wait")
//					continue
//				} else if err != nil {
//					panic("ExecTransaction failed due to" + err.Error())
//				}
//				fmt.Println("\n Succeed to ExecTransaction Tx")
//				return nil
//			}
//		}
//	}
//
//	return nil
//}
//
//func buildSigs(data []byte) (sigs []byte) {
//	fmt.Println("\nbuildSigs, data:", common.Bytes2Hex(data))
//
//	for _, privateKeyStr := range privateKeyStrs {
//		privateKey, err := crypto.ToECDSA(common.FromHex(privateKeyStr))
//		if nil != err {
//			panic("Failed to recover private key")
//			return nil
//		}
//
//		signature, err := crypto.Sign(data, privateKey)
//		if err != nil {
//			panic("Failed to sign data due to:" + err.Error())
//			return nil
//		}
//		signature[64] += 27
//		sigs = append(sigs, signature[:]...)
//	}
//	return
//}
