package types

//Version4Relayer ...
const Version4Relayer = "0.1.4"

const (
	Chain33BlockChainName    = "Chain33-mainchain"
	EthereumBlockChainName   = "Ethereum-mainchain"
	BinanceChainName         = "Binance"
	EthereumChainName        = "Ethereum"
	BTYAddrChain33           = "1111111111111111111114oLvT2"
	NilAddrChain33           = "1111111111111111111114oLvT2"
	EthNilAddr               = "0x0000000000000000000000000000000000000000"
	SYMBOL_BTY               = "BTY"
	Tx_Status_Pending        = "pending"
	Tx_Status_Success        = "Successful"
	Tx_Status_Failed         = "Failed"
	Source_Chain_Ethereum    = int32(0)
	Source_Chain_Chain33     = int32(1)
	Invalid_Tx_Index         = int64(0)
	Invalid_Chain33Tx_Status = int32(-1)
)

var Tx_Status_Map = map[int32]string{
	1: Tx_Status_Pending,
	2: Tx_Status_Success,
	3: Tx_Status_Failed,
}

var DecimalsPrefix = map[uint8]int64{
	1:  1e1,
	2:  1e2,
	3:  1e3,
	4:  1e4,
	5:  1e5,
	6:  1e6,
	7:  1e7,
	8:  1e8,
	9:  1e9,
	10: 1e10,
	11: 1e11,
	12: 1e12,
	13: 1e13,
	14: 1e14,
	15: 1e15,
	16: 1e16,
	17: 1e17,
	18: 1e18,
}
