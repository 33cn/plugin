package state

import (
	"errors"
)
import "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"

// TxData 本文件用来存储硬分叉中需要用到的数据
type EvmTxData struct {
	blockHeight int64
	testnet     bool
	// 存储多个交易信息 hash-->tx errlog
	txs map[string]string
}

var checkData map[int64]*EvmTxData

// InitEvmCheckData 初始化check数据
func InitEvmCheckData() {
	checkData = make(map[int64]*EvmTxData)
	//27878529 高度一个evm check 引起的状态不一致的bug，此处需要跟随现有主链的check 结果
	txdata := &EvmTxData{blockHeight: 27878529, testnet: false}
	txdata.txs = make(map[string]string)
	txdata.txs["0x93bbffde6c860dbfdb1439b9086caa0a7f6c55b7f909c9fa641aade0b96dcd4b"] = "requires at least 10 percent increase in handling fee,need more:162353"
	checkData[27878529] = txdata

}

// ProcessCheck 处理EVM对交易的检查更正处理
func ProcessCheck(blockHeight int64, txHash []byte) error {
	if txdata, ok := checkData[blockHeight]; ok {
		strHash := common.Bytes2Hex(txHash)
		v, ok := txdata.txs[strHash]
		if ok {
			if len(v) != 0 {
				return errors.New(v)
			}

		}
	}

	return nil
}
