package types

import "strings"

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

var (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-x2ethereum-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-x2ethereum-"
)

func CalProphecyPrefix(id string) []byte {
	return []byte(KeyPrefixStateDB + string(ProphecyKey) + id)
}

func CalEth2Chain33Prefix() []byte {
	return []byte(KeyPrefixStateDB + string(Eth2Chain33Key))
}

func CalWithdrawEthPrefix() []byte {
	return []byte(KeyPrefixStateDB + string(WithdrawEthKey))
}

func CalChain33ToEthPrefix() []byte {
	return []byte(KeyPrefixStateDB + string(Chain33ToEthKey))
}

func CalWithdrawChain33Prefix() []byte {
	return []byte(KeyPrefixStateDB + string(WithdrawChain33Key))
}

func CalValidatorMapsPrefix() []byte {
	return []byte(KeyPrefixStateDB + string(ValidatorMapsKey))
}

func CalLastTotalPowerPrefix() []byte {
	return []byte(KeyPrefixStateDB + string(LastTotalPowerKey))
}

func CalConsensusThresholdPrefix() []byte {
	return []byte(KeyPrefixStateDB + string(ConsensusThresholdKey))
}

func CalTokenSymbolTotalLockOrBurnAmount(symbol, tokenAddress, direction, txType string) []byte {
	return []byte(KeyPrefixLocalDB + string(TokenSymbolTotalLockOrBurnAmountKey) + direction + "-" + txType + "-" + strings.ToLower(symbol) + "-" + strings.ToLower(tokenAddress))
}

func CalTokenSymbolToTokenAddress(symbol string) []byte {
	return []byte(KeyPrefixLocalDB + string(TokenSymbolToTokenAddressKey) + strings.ToLower(symbol))
}
