package ethtxs

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	txslog = log15.New("ethereum relayer", "ethtxs")
)

//const ...
const (
	// GasLimit : the gas limit in Gwei used for transactions sent with TransactOpts
	GasLimit         = uint64(100 * 10000)
	GasLimit4RelayTx = uint64(40 * 10000)
	GasLimit4Deploy  = uint64(0) //此处需要设置为0,让交易自行估计,否则将会导致部署失败,TODO:其他解决途径后续调研解决
)

type BurnOrLockParameter struct {
	Clients                 []ethinterface.EthClientSpec
	ClientBSCRecommendSpecs []ethinterface.EthClientSpec
	Sender                  common.Address
	TokenOnEth              common.Address
	Claim                   ProphecyClaim
	PrivateKey              *ecdsa.PrivateKey
	Addr2TxNonce            map[common.Address]*NonceMutex
	ChainId                 *big.Int
	ChainName               string
	Registry                common.Address
}

// RelayOracleClaimToEthereum : relays the provided burn or lock to Chain33Bridge contract on the Ethereum network
func RelayOracleClaimToEthereum(burnOrLockParameter *BurnOrLockParameter) (string, error) {
	client := burnOrLockParameter.Clients[0]
	sender := burnOrLockParameter.Sender
	tokenOnEth := burnOrLockParameter.TokenOnEth
	claim := burnOrLockParameter.Claim
	privateKey := burnOrLockParameter.PrivateKey
	addr2TxNonce := burnOrLockParameter.Addr2TxNonce
	chainId := burnOrLockParameter.ChainId

	txslog.Info("RelayProphecyClaimToEthereum", "sender", sender.String(), "chain33Sender", hexutil.Encode(claim.Chain33Sender), "ethereumReceiver", claim.EthereumReceiver.String(),
		"TokenAddress", claim.TokenContractAddress.String(), "symbol", claim.Symbol, "Amount", claim.Amount.String(), "claimType", claim.ClaimType, "tokenOnEth", tokenOnEth.String())

	auth, err := PrepareAuth4MultiEthereumOpt(client, privateKey, sender, addr2TxNonce, chainId)
	if nil != err {
		txslog.Error("RelayProphecyClaimToEthereum", "PrepareAuth err", err.Error())
		return "", ErrNodeNetwork
	}
	defer func() {
		if nil != err {
			_, _ = revokeNonce4MultiEth(sender, addr2TxNonce)
		}
	}()

	auth.GasLimit = GasLimit4RelayTx

	claimID := crypto.Keccak256Hash(claim.Chain33TxHash, claim.Chain33Sender, claim.EthereumReceiver.Bytes(), []byte(claim.Symbol), claim.Amount.Bytes())

	// Sign the hash using the active validator's private key
	signature, err := utils.SignClaim4Evm(claimID, privateKey)
	if nil != err {
		return "", err
	}

	txslog.Info("RelayProphecyClaimToEthereum", "sender", sender.String(), "nonce", auth.Nonce, "claim.chain33TxHash", chain33Common.ToHex(claim.Chain33TxHash), "claimID", claimID.String())

	var txhash string
	bSendSucceed := false
	for true {
		for i := 0; i < len(burnOrLockParameter.Clients); i++ {
			tx, err := NewOracleClaimSend(burnOrLockParameter.Clients[i], auth, burnOrLockParameter, claimID, signature)
			if err != nil {
				// 如果第一次发生交易, 不是金额不够的错误,, 直接 return 报错退出
				if i == 0 && err.Error() != core.ErrInsufficientFunds.Error() {
					txslog.Error("RelayProphecyClaimToEthereum", "NewOracleClaim failed due to:", err.Error())
					return "", ErrNodeNetwork
				}

				if err.Error() != core.ErrAlreadyKnown.Error() && err.Error() != core.ErrNonceTooLow.Error() && err.Error() != core.ErrNonceTooHigh.Error() {
					txslog.Error("RelayProphecyClaimToEthereum", "PrepareAuth err", err.Error())
				}
			} else {
				if !bSendSucceed {
					bSendSucceed = true
					txhash = tx
				}
			}
		}

		// 交易同时发送到 BSC 官方节点
		if burnOrLockParameter.ChainName == BinanceChain {
			for i := 0; i < len(burnOrLockParameter.ClientBSCRecommendSpecs); i++ {
				tx, err := NewOracleClaimSend(burnOrLockParameter.ClientBSCRecommendSpecs[i], auth, burnOrLockParameter, claimID, signature)
				if err != nil {
					if err.Error() != core.ErrAlreadyKnown.Error() && err.Error() != core.ErrNonceTooLow.Error() && err.Error() != core.ErrNonceTooHigh.Error() {
						txslog.Error("RelayProphecyClaimToEthereum", "PrepareAuth err", err.Error())
					}
				} else {
					if !bSendSucceed {
						bSendSucceed = true
						txhash = tx
					}
				}
			}
		}

		if bSendSucceed {
			break
		}

		time.Sleep(time.Second * 10)
	}

	txslog.Info("RelayProphecyClaimToEthereum", "NewOracleClaim tx hash:", txhash)
	return txhash, nil
}

func NewOracleClaimSend(client ethinterface.EthClientSpec, transactOpts *bind.TransactOpts, burnOrLockParameter *BurnOrLockParameter, claimID [32]byte, signature []byte) (string, error) {
	registry := burnOrLockParameter.Registry
	tokenOnEth := burnOrLockParameter.TokenOnEth
	claim := burnOrLockParameter.Claim

	oracleAddr, err := GetAddressFromBridgeRegistry(client, registry, registry, Oracle)
	if nil != err {
		return "", errors.New("failed to get addr for bridgeBank from registry " + err.Error())
	}

	oracleInstance, err := generated.NewOracle(*oracleAddr, client)
	if nil != err {
		return "", errors.New("failed to NewOracle " + err.Error())
	}

	tx, err := oracleInstance.NewOracleClaim(transactOpts, uint8(claim.ClaimType), claim.Chain33Sender, claim.EthereumReceiver, tokenOnEth, claim.Symbol, claim.Amount, claimID, signature)
	if nil != err {
		return "", err
	}
	txslog.Info("RelayProphecyClaimToEthereum", "tx", tx.Hash().Hex(), "claim.chain33TxHash", chain33Common.ToHex(claim.Chain33TxHash))
	return tx.Hash().Hex(), nil
}
