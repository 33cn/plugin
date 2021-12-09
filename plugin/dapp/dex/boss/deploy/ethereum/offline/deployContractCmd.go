package offline

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-swap-periphery/src/pancakeFactory"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-swap-periphery/src/pancakeRouter"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

//SignFactoryCmd 构造部署factory 合约的交易，并对其签名输出到文件中
type SignCmd struct {
	From        string
	Nonce       uint64
	GasPrice    uint64
	FactoryAddr string
	TxHash      string
	Fee2Addr    string
	Timestamp   string
	SignedTx    string
	Reward      uint64
	StartBlock  uint64
}

func (s *SignCmd) signCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign contract", //first step
		Short: "sign pancake and farm contract to ethereum ",
		Run:   s.signContract, //对要部署的factory合约进行签名
	}
	s.addFlags(cmd)
	return cmd
}

func (s *SignCmd) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("fee2stter", "", "", "fee2stter addr")
	cmd.MarkFlagRequired("fee2stter")
	cmd.Flags().StringP("priv", "p", "", "private key")
	cmd.Flags().Int64P("reward", "", 5, "Set the reward for each block")
	cmd.MarkFlagRequired("reward")
	cmd.Flags().Int64P("start", "", -1, "Set effective height")
	cmd.MarkFlagRequired("start")
	cmd.Flags().Int64P("interval", "i", 5, "interval time for send")
	cmd.Flags().Int64P("nonce", "n", -1, "transaction count")
	cmd.MarkFlagRequired("nonce")
	cmd.Flags().Int64P("gasprice", "g", 1000000000, "gas price") // 1Gwei=1e9wei
	cmd.MarkFlagRequired("gasprice")
}

func (s *SignCmd) signContract(cmd *cobra.Command, args []string) {
	fee2setter, _ := cmd.Flags().GetString("fee2stter")
	key, _ := cmd.Flags().GetString("priv")
	reward, _ := cmd.Flags().GetInt64("reward")
	startBlock, _ := cmd.Flags().GetInt64("start")
	interval, _ := cmd.Flags().GetInt64("interval")
	gasprice, _ := cmd.Flags().GetInt64("gasprice")
	nonce, _ := cmd.Flags().GetInt64("nonce")
	if startBlock <= 0 {
		panic("startBlock  err")
	}

	priv, addr, err := recoverBinancePrivateKey(key)
	if err != nil {
		panic(err)
	}
	fmt.Println("recover addr", addr)

	s.GasPrice = uint64(gasprice)
	s.Nonce = uint64(nonce)
	s.Reward = uint64(reward)
	s.StartBlock = uint64(startBlock)
	gasPrice := big.NewInt(int64(s.GasPrice))
	var timewait time.Duration
	if interval > 0 {
		timewait = time.Duration(interval) * time.Second
	} else {
		timewait = time.Second * 5
	}
	err = s.signContractTx(fee2setter, priv, gasPrice, s.Nonce, timewait)
	if nil != err {
		fmt.Println("Failed to deploy contracts due to:", err.Error())
		return
	}

	fmt.Println("Succeed to signed deploy contracts")
}

func (s *SignCmd) signContractTx(fee2setter string, key *ecdsa.PrivateKey, gasPrice *big.Int, nonce uint64, timewait time.Duration) error {
	fee2setterAddr := common.HexToAddress(fee2setter)
	from := crypto.PubkeyToAddress(key.PublicKey)
	//--------------------
	//sign factory
	//--------------------
	signedTx, txHash, err := s.reWriteDeplopyPancakeFactory(nonce, gasPrice, key, fee2setterAddr)
	if nil != err {
		panic(fmt.Sprintf("Failed to DeployPancakeFactory with err:%s", err.Error()))
	}

	factoryAddr := crypto.CreateAddress(from, nonce)
	var signData = make([]*DeployContract, 0)
	var factData DeployContract
	factData.Interval = timewait
	factData.TxHash = txHash
	factData.RawTx = signedTx
	factData.Nonce = s.Nonce
	factData.ContractAddr = factoryAddr.String()
	factData.ContractName = "factory"
	signData = append(signData, &factData)
	//--------------------
	//sign weth9
	//--------------------
	weth := new(SignWeth9Cmd)
	wsignedTx, hash, err := weth.reWriteDeployWETH9(factData.Nonce+1, gasPrice, key)
	if nil != err {
		panic(fmt.Sprintf("Failed to DeployPancakeFactory with err:%s", err.Error()))
	}

	weth9Addr := crypto.CreateAddress(from, factData.Nonce+1)
	var weth9Data DeployContract
	weth9Data.Interval = timewait
	weth9Data.Nonce = s.Nonce + 1
	weth9Data.TxHash = hash
	weth9Data.RawTx = wsignedTx
	weth9Data.ContractAddr = weth9Addr.String()
	weth9Data.ContractName = "weth9"
	signData = append(signData, &weth9Data)
	//--------------------
	//sign PanCakeRouter
	//--------------------
	panRouter := new(SignPanCakeRout)
	rSignedTx, hash, err := panRouter.reWriteDeployPanCakeRout(weth9Data.Nonce+1, gasPrice, key, factoryAddr, weth9Addr)
	if nil != err {
		panic(fmt.Sprintf("Failed to reWriteDeployPanCakeRout with err:%s", err.Error()))
	}
	panrouterAddr := crypto.CreateAddress(from, weth9Data.Nonce+1)

	var panData DeployContract
	panData.Interval = timewait
	panData.Nonce = weth9Data.Nonce + 1
	panData.RawTx = rSignedTx
	panData.ContractAddr = panrouterAddr.String()
	panData.TxHash = hash
	panData.ContractName = "pancakerouter"
	signData = append(signData, &panData)

	/**************************pancake contract signed completed*************************/
	//--------------------let's begain Farm contract^_^--------------------
	//Sign Farm Contractor
	//--------------------
	farmNonce := panData.Nonce + 1
	var cakeToken = new(SignCakeToken)
	var cakeData = new(DeployContract)
	cakeSignedtx, hash, err := cakeToken.reWriteDeployCakeToken(farmNonce, gasPrice, key)
	if nil != err {
		panic(fmt.Sprintf("Failed to reWriteDeployCakeToken with err:%s", err.Error()))
	}

	cakeContractAddr := crypto.CreateAddress(from, farmNonce)
	cakeData.Interval = timewait
	cakeData.Nonce = farmNonce
	cakeData.RawTx = cakeSignedtx
	cakeData.TxHash = hash
	cakeData.ContractName = "caketoken"
	cakeData.ContractAddr = cakeContractAddr.String()
	signData = append(signData, cakeData)
	//--------------------
	//Sign syrupBar Contractor
	//--------------------
	syrupBarNonce := farmNonce + 1
	var syrupBar = new(signsyrupBar)
	var syrupBarData = new(DeployContract)
	syupSignedTx, hash, err := syrupBar.reWriteDeploysyrupBar(syrupBarNonce, gasPrice, key, cakeContractAddr)
	if err != nil {
		panic(err)
	}

	syupContractAddr := crypto.CreateAddress(from, syrupBarNonce)
	syrupBarData.Interval = timewait
	syrupBarData.Nonce = syrupBarNonce
	syrupBarData.TxHash = hash
	syrupBarData.ContractName = "syrupbar"
	syrupBarData.ContractAddr = syupContractAddr.String()
	syrupBarData.RawTx = syupSignedTx
	signData = append(signData, syrupBarData)
	//--------------------
	//Sign masterChef Contractor
	//--------------------
	masterChefNonce := syrupBarNonce + 1
	var mChefData = new(DeployContract)
	var mChef = new(signMasterChef)
	reward := big.NewInt(int64(s.Reward * 1e18))
	startBlockHeight := big.NewInt(int64(s.StartBlock))
	mchefSignedTx, hash, err := mChef.reWriteDeployMasterChef(masterChefNonce, gasPrice, key, cakeContractAddr, syupContractAddr, from, reward, startBlockHeight)
	if err != nil {
		panic(err)
	}

	mChefContractAddr := crypto.CreateAddress(from, masterChefNonce)
	mChefData.Interval = timewait
	mChefData.Nonce = masterChefNonce
	mChefData.TxHash = hash
	mChefData.ContractName = "masterchef"
	mChefData.RawTx = mchefSignedTx
	mChefData.ContractAddr = mChefContractAddr.String()
	signData = append(signData, mChefData)

	//write signedtx to spec file
	writeToFile("signed.txt", &signData)
	return nil
}

//构造交易，签名交易 factory
func (s *SignCmd) reWriteDeplopyPancakeFactory(nonce uint64, gasPrice *big.Int, key *ecdsa.PrivateKey, fee2addr common.Address, param ...interface{}) (signedTx, hash string, err error) {
	parsed, err := abi.JSON(strings.NewReader(pancakeFactory.PancakeFactoryABI))
	if err != nil {
		return
	}
	input, err := parsed.Pack("", fee2addr)
	if err != nil {
		return
	}
	abiBin := pancakeFactory.PancakeFactoryBin
	data := append(common.FromHex(abiBin), input...)
	var amount = new(big.Int)
	ntx := types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data)
	//ntx := types.NewTransaction(nonce, common.Address{}, amount, gasLimit, gasPrice, data)
	return SignTx(key, ntx)
}

type SignWeth9Cmd struct {
}

//only sign Weth9
func (s *SignWeth9Cmd) reWriteDeployWETH9(nonce uint64, gasPrice *big.Int, key *ecdsa.PrivateKey, param ...interface{}) (signedTx, hash string, err error) {
	parsed, err := abi.JSON(strings.NewReader(pancakeRouter.WETH9ABI))
	if err != nil {
		return "", "", err
	}
	input, err := parsed.Pack("", param...)
	abiBin := pancakeRouter.WETH9Bin
	data := append(common.FromHex(abiBin), input...)
	var amount = new(big.Int)
	ntx := types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data)
	return SignTx(key, ntx)
}

type SignPanCakeRout struct {
}

func (s *SignPanCakeRout) reWriteDeployPanCakeRout(nonce uint64, gasPrice *big.Int, key *ecdsa.PrivateKey, factoryAddr, Weth9 common.Address) (signedTx, hash string, err error) {
	parsed, err := abi.JSON(strings.NewReader(pancakeRouter.PancakeRouterABI))
	if err != nil {
		return
	}
	input, err := parsed.Pack("", factoryAddr, Weth9)
	if err != nil {
		return
	}
	abiBin := pancakeRouter.PancakeRouterBin
	data := append(common.FromHex(abiBin), input...)
	var amount = new(big.Int)
	ntx := types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data)
	return SignTx(key, ntx)
}

func SignTx(key *ecdsa.PrivateKey, tx *types.Transaction) (signedTx, hash string, err error) {
	signer := types.HomesteadSigner{}
	txhash := signer.Hash(tx)
	signature, err := crypto.Sign(txhash.Bytes(), key)
	if err != nil {
		return
	}
	tx, err = tx.WithSignature(signer, signature)
	if err != nil {
		return
	}
	txBinary, err := tx.MarshalBinary()
	if err != nil {
		return
	}
	hash = tx.Hash().String()
	signedTx = common.Bytes2Hex(txBinary[:])
	return
}

func SignEIP155Tx(key *ecdsa.PrivateKey, tx *types.Transaction, chainEthId int64) (signedTx, hash string, err error) {
	signer := types.NewEIP155Signer(big.NewInt(chainEthId))
	txhash := signer.Hash(tx)
	signature, err := crypto.Sign(txhash.Bytes(), key)
	if err != nil {
		return
	}
	tx, err = tx.WithSignature(signer, signature)
	if err != nil {
		return
	}
	txBinary, err := tx.MarshalBinary()
	if err != nil {
		return
	}
	hash = tx.Hash().String()
	signedTx = common.Bytes2Hex(txBinary[:])
	return
}

func recoverBinancePrivateKey(key string) (priv *ecdsa.PrivateKey, address common.Address, err error) {
	//louyuqi: f726c7c704e57ec5d59815dda23ddd794f71ae15f7e0141f00f73eff35334ac6
	//hzj: 2bcf3e23a17d3f3b190a26a098239ad2d20267a673440e0f57a23f44f94b77b9
	priv, err = crypto.ToECDSA(common.FromHex(key))
	if err != nil {
		panic("Failed to recover private key")
	}
	address = crypto.PubkeyToAddress(priv.PublicKey)
	fmt.Println("The address is:", address.String())
	return
}
