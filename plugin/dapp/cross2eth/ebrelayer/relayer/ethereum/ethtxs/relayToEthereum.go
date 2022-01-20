package ethtxs

import (
	"crypto/ecdsa"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethinterface"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/ethereum/go-ethereum/common"
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

// RelayOracleClaimToEthereum : relays the provided burn or lock to Chain33Bridge contract on the Ethereum network
func RelayOracleClaimToEthereum(oracleInstance *generated.Oracle, client ethinterface.EthClientSpec, sender, tokenOnEth common.Address, claim ProphecyClaim, privateKey *ecdsa.PrivateKey, addr2TxNonce map[common.Address]*NonceMutex) (txhash string, err error) {
	txslog.Info("RelayProphecyClaimToEthereum", "sender", sender.String(), "chain33Sender", hexutil.Encode(claim.Chain33Sender), "ethereumReceiver", claim.EthereumReceiver.String(),
		"TokenAddress", claim.TokenContractAddress.String(), "symbol", claim.Symbol, "Amount", claim.Amount.String(), "claimType", claim.ClaimType, "tokenOnEth", tokenOnEth.String())

	auth, err := PrepareAuth4MultiEthereum(client, privateKey, sender, addr2TxNonce)
	if nil != err {
		txslog.Error("RelayProphecyClaimToEthereum", "PrepareAuth err", err.Error())
		return "", err
	}
	defer func() {
		if nil != err {
			_, _ = revokeNonce4MultiEth(sender, addr2TxNonce)
		}
	}()

	auth.GasLimit = GasLimit4RelayTx

	claimID := crypto.Keccak256Hash(claim.chain33TxHash, claim.Chain33Sender, claim.EthereumReceiver.Bytes(), []byte(claim.Symbol), claim.Amount.Bytes())

	// Sign the hash using the active validator's private key
	signature, err := utils.SignClaim4Evm(claimID, privateKey)
	if nil != err {
		return "", err
	}

	txslog.Info("RelayProphecyClaimToEthereum", "sender", sender.String(), "nonce", auth.Nonce, "claim.chain33TxHash", common.Bytes2Hex(claim.chain33TxHash))

	tx, err := oracleInstance.NewOracleClaim(auth, uint8(claim.ClaimType), claim.Chain33Sender, claim.EthereumReceiver, tokenOnEth, claim.Symbol, claim.Amount, claimID, signature)
	if nil != err {
		txslog.Error("RelayProphecyClaimToEthereum", "NewOracleClaim failed due to:", err.Error())
		return "", err
	}

	txhash = tx.Hash().Hex()
	txslog.Info("RelayProphecyClaimToEthereum", "NewOracleClaim tx hash:", txhash)
	return txhash, nil
}
