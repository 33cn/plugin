// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"errors"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

func (client *client) setLocalDb(set *types.LocalDBSet) error {
	//如果追赶上主链了，则落盘
	if client.isCaughtUp() {
		set.Txid = 1
		client.blockSyncClient.handleLocalCaughtUpMsg()
	}

	msg := client.GetQueueClient().NewMessage("blockchain", types.EventSetValueByKey, set)
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return err
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}
	if resp.GetData().(*types.Reply).IsOk {
		return nil
	}
	return errors.New(string(resp.GetData().(*types.Reply).GetMsg()))
}

func (client *client) getLocalDb(set *types.LocalDBGet, count int) ([][]byte, error) {
	msg := client.GetQueueClient().NewMessage("blockchain", types.EventGetValueByKey, set)
	err := client.GetQueueClient().Send(msg, true)
	if err != nil {
		return nil, err
	}
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return nil, err
	}

	reply := resp.GetData().(*types.LocalReplyValue)
	if len(reply.Values) != count {
		plog.Error("Parachain getLocalDb count not match", "expert", count, "real", len(reply.Values))
		return nil, types.ErrInvalidParam
	}

	return reply.Values, nil
}

func (client *client) addLocalBlock(block *pt.ParaLocalDbBlock) error {
	cfg := client.GetAPI().GetConfig()
	set := &types.LocalDBSet{}

	key := calcTitleHeightKey(cfg.GetTitle(), block.Height)
	kv := &types.KeyValue{Key: key, Value: types.Encode(block)}
	set.KV = append(set.KV, kv)

	//两个key原子操作
	key = calcTitleLastHeightKey(cfg.GetTitle())
	kv = &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: block.Height})}
	set.KV = append(set.KV, kv)

	return client.setLocalDb(set)
}

func (client *client) saveBatchLocalBlocks(blocks []*pt.ParaLocalDbBlock) error {
	cfg := client.GetAPI().GetConfig()
	set := &types.LocalDBSet{}

	for _, block := range blocks {
		key := calcTitleHeightKey(cfg.GetTitle(), block.Height)
		kv := &types.KeyValue{Key: key, Value: types.Encode(block)}
		set.KV = append(set.KV, kv)
	}
	//save lastHeight,两个key原子操作
	key := calcTitleLastHeightKey(cfg.GetTitle())
	kv := &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: blocks[len(blocks)-1].Height})}
	set.KV = append(set.KV, kv)

	return client.setLocalDb(set)
}

func (client *client) delLocalBlock(height int64) error {
	cfg := client.GetAPI().GetConfig()
	set := &types.LocalDBSet{}
	key := calcTitleHeightKey(cfg.GetTitle(), height)
	kv := &types.KeyValue{Key: key, Value: nil}
	set.KV = append(set.KV, kv)

	//两个key原子操作
	key = calcTitleLastHeightKey(cfg.GetTitle())
	kv = &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: height - 1})}
	set.KV = append(set.KV, kv)

	return client.setLocalDb(set)
}

// localblock 设置到当前高度，当前高度后面block会被新的区块覆盖
func (client *client) removeLocalBlocks(curHeight int64) error {
	cfg := client.GetAPI().GetConfig()
	set := &types.LocalDBSet{}

	key := calcTitleLastHeightKey(cfg.GetTitle())
	kv := &types.KeyValue{Key: key, Value: types.Encode(&types.Int64{Data: curHeight})}
	set.KV = append(set.KV, kv)

	return client.setLocalDb(set)
}

func (client *client) getLastLocalHeight() (int64, error) {
	cfg := client.GetAPI().GetConfig()
	key := calcTitleLastHeightKey(cfg.GetTitle())
	set := &types.LocalDBGet{Keys: [][]byte{key}}
	value, err := client.getLocalDb(set, len(set.Keys))
	if err != nil {
		return -1, err
	}
	if len(value) == 0 || value[0] == nil {
		return -1, types.ErrNotFound
	}

	height := &types.Int64{}
	err = types.Decode(value[0], height)
	if err != nil {
		return -1, err
	}
	return height.Data, nil

}

func (client *client) getLocalBlockByHeight(height int64) (*pt.ParaLocalDbBlock, error) {
	cfg := client.GetAPI().GetConfig()
	key := calcTitleHeightKey(cfg.GetTitle(), height)
	set := &types.LocalDBGet{Keys: [][]byte{key}}

	value, err := client.getLocalDb(set, len(set.Keys))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 || value[0] == nil {
		return nil, types.ErrNotFound
	}

	var block pt.ParaLocalDbBlock
	err = types.Decode(value[0], &block)
	if err != nil {
		return nil, err
	}
	return &block, nil

}

func (client *client) saveMainBlock(height int64, block *types.ParaTxDetail) error {
	cfg := client.GetAPI().GetConfig()
	set := &types.LocalDBSet{}

	key := calcTitleMainHeightKey(cfg.GetTitle(), height)
	kv := &types.KeyValue{Key: key, Value: types.Encode(block)}
	set.KV = append(set.KV, kv)

	return client.setLocalDb(set)
}

func (client *client) saveBatchMainBlocks(txs *types.ParaTxDetails) error {
	cfg := client.GetAPI().GetConfig()
	set := &types.LocalDBSet{}

	for _, block := range txs.Items {
		key := calcTitleMainHeightKey(cfg.GetTitle(), block.Header.Height)
		kv := &types.KeyValue{Key: key, Value: types.Encode(block)}
		set.KV = append(set.KV, kv)
	}

	return client.setLocalDb(set)
}

func (client *client) rmvBatchMainBlocks(start, end int64) error {
	cfg := client.GetAPI().GetConfig()
	set := &types.LocalDBSet{}

	for i := start; i < end; i++ {
		key := calcTitleMainHeightKey(cfg.GetTitle(), i)
		kv := &types.KeyValue{Key: key, Value: nil}
		set.KV = append(set.KV, kv)
	}

	return client.setLocalDb(set)
}

func (client *client) getMainBlockFromDb(height int64) (*types.ParaTxDetail, error) {
	cfg := client.GetAPI().GetConfig()
	key := calcTitleMainHeightKey(cfg.GetTitle(), height)
	set := &types.LocalDBGet{Keys: [][]byte{key}}

	value, err := client.getLocalDb(set, len(set.Keys))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 || value[0] == nil {
		return nil, types.ErrNotFound
	}

	var tx types.ParaTxDetail
	err = types.Decode(value[0], &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
