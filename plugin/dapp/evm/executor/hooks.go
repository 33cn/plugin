// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/client"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/state"
)

// CanTransfer 检查合约调用账户是否有充足的金额进行转账交易操作
func CanTransfer(db state.EVMStateDB, sender, recipient common.Address, amount uint64) bool {
	return db.CanTransfer(sender.String(), recipient.String(), amount)
}

// Transfer 在内存数据库中执行转账操作（只修改内存中的金额）
// 从外部账户地址到合约账户地址
func Transfer(db state.EVMStateDB, sender, recipient common.Address, amount uint64) bool {
	return db.Transfer(sender.String(), recipient.String(), amount)
}

// GetHashFn 获取制定高度区块的哈希
func GetHashFn(api client.QueueProtocolAPI) func(blockHeight uint64) common.Hash {
	return func(blockHeight uint64) common.Hash {
		if api != nil {
			reply, err := api.GetBlockHash(&types.ReqInt{Height: int64(blockHeight)})
			if nil != err {
				log.Error("Call GetBlockHash Failed.", err)
			}
			return common.BytesToHash(reply.Hash)
		}
		return common.Hash{}
	}
}
