package ethtxs

// ------------------------------------------------------------
//	Relay : Builds and encodes EthBridgeClaim Msgs with the
//  	specified variables, before presenting the unsigned
//      transaction to validators for optional signing.
//      Once signed, the data packets are sent as transactions
//      on the chain33 Bridge.
// ------------------------------------------------------------

import (
	"github.com/33cn/chain33/common"
	chain33Crypto "github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	chain33Types "github.com/33cn/chain33/types"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
)

// RelayLockToChain33 : RelayLockToChain33 applies validator's signature to an EthBridgeClaim message
//		containing information about an event on the Ethereum blockchain before relaying to the Bridge
func RelayLockToChain33(privateKey chain33Crypto.PrivKey, claim *ebrelayerTypes.EthBridgeClaim, rpcURL string) (string, error) {
	var res string

	params := &types.Eth2Chain33{
		EthereumChainID:       claim.EthereumChainID,
		BridgeContractAddress: claim.BridgeBrankAddr,
		Nonce:                 claim.Nonce,
		IssuerDotSymbol:       claim.Symbol,
		TokenContractAddress:  claim.TokenAddr,
		EthereumSender:        claim.EthereumSender,
		Chain33Receiver:       claim.Chain33Receiver,
		Amount:                claim.Amount,
		ClaimType:             int64(claim.ClaimType),
		Decimals:              claim.Decimal,
	}

	pm := rpctypes.CreateTxIn{
		Execer:     X2Eth,
		ActionName: types.NameEth2Chain33Action,
		Payload:    chain33Types.MustPBToJSON(params),
	}
	ctx := jsonclient.NewRPCCtx(rpcURL, "Chain33.CreateTransaction", pm, &res)
	_, _ = ctx.RunResult()

	data, err := common.FromHex(res)
	if err != nil {
		return "", err
	}
	var tx chain33Types.Transaction
	err = chain33Types.Decode(data, &tx)
	if err != nil {
		return "", err
	}

	if tx.Fee == 0 {
		tx.Fee, err = tx.GetRealFee(1e5)
		if err != nil {
			return "", err
		}
	}
	//构建交易，验证人validator用来向chain33合约证明自己验证了该笔从以太坊向chain33跨链转账的交易
	tx.Sign(chain33Types.SECP256K1, privateKey)

	txData := chain33Types.Encode(&tx)
	dataStr := common.ToHex(txData)
	pms := rpctypes.RawParm{
		Token: "BTY",
		Data:  dataStr,
	}
	var txhash string

	ctx = jsonclient.NewRPCCtx(rpcURL, "Chain33.SendTransaction", pms, &txhash)
	_, err = ctx.RunResult()
	return txhash, err
}

//RelayBurnToChain33 ...
func RelayBurnToChain33(privateKey chain33Crypto.PrivKey, claim *ebrelayerTypes.EthBridgeClaim, rpcURL string) (string, error) {
	var res string

	params := &types.Eth2Chain33{
		EthereumChainID:       claim.EthereumChainID,
		BridgeContractAddress: claim.BridgeBrankAddr,
		Nonce:                 claim.Nonce,
		IssuerDotSymbol:       claim.Symbol,
		TokenContractAddress:  claim.TokenAddr,
		EthereumSender:        claim.EthereumSender,
		Chain33Receiver:       claim.Chain33Receiver,
		Amount:                claim.Amount,
		ClaimType:             int64(claim.ClaimType),
		Decimals:              claim.Decimal,
	}

	pm := rpctypes.CreateTxIn{
		Execer:     X2Eth,
		ActionName: types.NameWithdrawEthAction,
		Payload:    chain33Types.MustPBToJSON(params),
	}
	ctx := jsonclient.NewRPCCtx(rpcURL, "Chain33.CreateTransaction", pm, &res)
	_, _ = ctx.RunResult()

	data, err := common.FromHex(res)
	if err != nil {
		return "", err
	}
	var tx chain33Types.Transaction
	err = chain33Types.Decode(data, &tx)
	if err != nil {
		return "", err
	}

	if tx.Fee == 0 {
		tx.Fee, err = tx.GetRealFee(1e5)
		if err != nil {
			return "", err
		}
	}
	//构建交易，验证人validator用来向chain33合约证明自己验证了该笔从以太坊向chain33跨链转账的交易
	tx.Sign(chain33Types.SECP256K1, privateKey)

	txData := chain33Types.Encode(&tx)
	dataStr := common.ToHex(txData)
	pms := rpctypes.RawParm{
		Token: "BTY",
		Data:  dataStr,
	}
	var txhash string

	ctx = jsonclient.NewRPCCtx(rpcURL, "Chain33.SendTransaction", pms, &txhash)
	_, err = ctx.RunResult()
	return txhash, err
}
