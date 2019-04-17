package db

import (
	"fmt"
	"time"

	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/plugin/cmd/local-status/es_cli"
	"github.com/olivere/elastic"
	"github.com/pkg/errors"
)

type Sync interface {
	Recover(client *es_cli.ESClient) error
	Sync(client *es_cli.ESClient, headerCh chan int64, blockCh chan interface{})
}

// 保存区块步骤
// 1. 记录 block hash 对应的 keys
// 2. 更新keys
// 3. 更新高度
//
// 重启恢复
// 1. 看高度， 对应高度是已经完成的
// 2. 继续重新下一个高度即可。 重复写， 幂等
// 所以不需要恢复过程， 读出高度即可

// dbStatus/lastHeader/{height} --> hash
// dbStatus/block/hash -> keys[prev,current]
//

type dbStatus struct {
	height int64
	hash   string
}

func NewDBStatus() *dbStatus {
	return &dbStatus{height: -1}
}

func (d *dbStatus) Recover(client *es_cli.ESClient) error {
	return d.lastHeader(client)
}

func (d *dbStatus) Sync(client *es_cli.ESClient, headerCh chan int64, blockCh chan interface{}) {
	nextHeader := d.height + 1

	for {
		headerCh <- nextHeader
		item := <-blockCh
		fmt.Printf("XX %+v\n", item)
		if block, ok := item.(*rpctypes.BlockDetail); ok {
			next, err := d.dealBlock(client, block)
			if err != nil {
				nextHeader = next
				continue
			}
		}
		nextHeader++
	}
}

const statusDB = "dbStatus"
const statusTable = "height"
const lastHeader = "lastHeader" // header, hash

func (d *dbStatus) lastHeader(client *es_cli.ESClient) error {
	result, err := client.Get(statusDB, statusTable, lastHeader)
	if err != nil {
		if elastic.IsNotFound(err) {
			// 第一个连上
			return nil
		}
		return err
	}

	h, err := getInt(result, "header")
	if err != nil {
		return err
	}

	hash, err := getString(result, "hash")
	if err != nil {
		return err
	}
	d.height = h
	d.hash = hash
	return nil
}

func getInt(kvs map[string]interface{}, key string) (int64, error) {
	if v, ok := kvs[key]; ok {
		if i, ok := v.(int64); ok {
			return i, nil
		}
	}
	return 0, errors.New("get value failed")
}

func getString(kvs map[string]interface{}, key string) (string, error) {
	if v, ok := kvs[key]; ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
	}
	return "", errors.New("get value failed")
}

func (d *dbStatus) dealBlock(client *es_cli.ESClient, block *rpctypes.BlockDetail) (int64, error) {
	if block.Block.Height != d.height+1 {
		return d.height + 1, errors.New("height not match")
	}
	if block.Block.Height != 0 && block.Block.ParentHash != d.hash {
		// TODO rollback
		// delete block ${height}
		// set ${height} = d.height - 1
		d.rollback(d.height, client)
		d.height--
		return d.height + 1, errors.New("block hash not match")
	}

	err := d.addBlock(block.Block.Height, client, block)
	if err != nil {
		return d.height + 1, errors.New("add block failed")
	}
	d.height++
	return d.height + 1, nil
}

func (d *dbStatus) addBlock(height int64, client *es_cli.ESClient, block *rpctypes.BlockDetail) error {
	txCount := len(block.Block.Txs)
	fmt.Println("XX BLOCK: len txs:", txCount)
	for i := 0; i < txCount; i++ {
		fmt.Printf("XX BLOCK: tx %d/%d: %s\n", i, txCount, block.Block.Txs[i])
		fmt.Printf("\t\t%s %+v %d\n", block.Receipts[i].TyName, block.Receipts[i].TyName, len(block.Receipts[i].Logs))
		fmt.Println()

		kvc := NewConvert(block.Block.Txs[i].Execer, block)
		if kvc == nil {
			continue
		}
		for _, l := range block.Receipts[i].Logs {
			keys, p, c, err := kvc.Convert(int64(l.Ty), string(l.Log))
			if err != nil {
				continue
			}
			fmt.Println("X(set db):", keys, string(p), string(c))
			client.Update(keys[0], keys[1], keys[2], string(c))
			// try to set es
			time.Sleep(time.Second)
		}
	}

	return nil
}

func (d *dbStatus) rollback(height int64, client *es_cli.ESClient) error {
	return nil
}
