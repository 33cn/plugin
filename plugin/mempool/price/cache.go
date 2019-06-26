package price

import (
	"github.com/33cn/chain33/common/skiplist"
	"github.com/33cn/chain33/system/mempool"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

// Queue 价格队列模式(价格=手续费/交易字节数,价格高者优先,同价则时间早优先)
type Queue struct {
	*skiplist.Queue
	subConfig subConfig
}

type priceScore struct {
	*mempool.Item
}

func (item *priceScore) GetScore() int64 {
	txSize := proto.Size(item.Value)
	return item.Value.Fee / int64(txSize)
}

// NewQueue 创建队列
func NewQueue(subcfg subConfig) *Queue {
	return &Queue{
		Queue:     skiplist.NewQueue(),
		subConfig: subcfg,
	}
}

//GetItem 获取数据通过 key
func (cache *Queue) GetItem(hash string) (*mempool.Item, error) {
	item, err := cache.Queue.GetItem(hash)
	if err != nil {
		return nil, err
	}
	return item.(*priceScore).Item, nil
}

//Walk 获取数据通过 key
func (cache *Queue) Walk(count int, cb func(tx *mempool.Item) bool) {
	cache.Queue.Walk(count, func(item skiplist.Scorer) bool {
		return cb(item.(*priceScore).Item)
	})
}

// Push 把给定tx添加到Queue；如果tx已经存在Queue中或Mempool已满则返回对应error
func (cache *Queue) Push(item *mempool.Item) error {
	hash := item.Value.Hash()
	if cache.Exist(string(hash)) {
		return types.ErrTxExist
	}
	sv := cache.CreateSkipValue(&priceScore{Item: item})
	if int64(cache.Size()) >= cache.subConfig.PoolCacheSize {
		tail := cache.Last().(*priceScore)
		lasthash := string(tail.Value.Hash())
		//价格高存留
		switch sv.Compare(cache.CreateSkipValue(tail)) {
		case -1:
			cache.Queue.Remove(lasthash)
		case 0:
			if item.EnterTime < tail.EnterTime {
				cache.Queue.Remove(lasthash)
				break
			}
			return types.ErrMemFull
		case 1:
			return types.ErrMemFull
		default:
			return types.ErrMemFull
		}
	}
	cache.Queue.Push(string(hash), &priceScore{Item: item})
	return nil
}

// GetProperFee 获取合适的手续费率,取前100的平均手续费率
func (cache *Queue) GetProperFee() int64 {
	var sumFeeRate int64
	var properFeeRate int64
	if cache.Size() == 0 {
		return cache.subConfig.ProperFee
	}
	i := 0
	var txSize int
	var feeRate int64
	cache.Walk(100, func(item *mempool.Item) bool {
		txSize = proto.Size(item.Value)
		feeRate = item.Value.Fee / int64(txSize/1000+1)
		sumFeeRate += feeRate
		i++
		return true
	})
	properFeeRate = sumFeeRate / int64(i)
	return properFeeRate
}
