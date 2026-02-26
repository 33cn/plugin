package lighttypes

import (
	"github.com/btcsuite/btcd/chaincfg"
)

// GetBtcChainParams 获取比特币链参数
func GetBtcChainParams(netName string) *chaincfg.Params {
	switch netName {
	case "mainnet", "":
		return &chaincfg.MainNetParams
	case "testnet3":
		return &chaincfg.TestNet3Params
	case "testnet4":
		return &chaincfg.TestNet4Params
	case "regtest":
		return &chaincfg.RegressionNetParams
	case "simnet":
		return &chaincfg.SimNetParams
	case "signet":
		return &chaincfg.SigNetParams
	default:
		return &chaincfg.MainNetParams
	}
}
