package score

import (
	"time"

	"github.com/33cn/chain33/common/skiplist"
	"github.com/33cn/chain33/system/mempool"
	"github.com/golang/protobuf/proto"
)

// Queue 分数队列模式(分数=定量a*常量b*手续费/交易字节数-常量c*时间,按分数排队,高的优先,定量a和常量b,c可配置)
type Queue struct {
	*skiplist.Queue
	subConfig subConfig
}

type scoreScore struct {
	*mempool.Item
	subConfig subConfig
}

func (item *scoreScore) GetScore() int64 {
	size := proto.Size(item.Value)
	score := item.subConfig.PriceConstant*(item.Value.Fee/int64(size))*
		item.subConfig.PricePower - item.subConfig.TimeParam*item.EnterTime
	return score
}

func (item *scoreScore) Hash() []byte {
	return item.Value.Hash()
}

func (item *scoreScore) Compare(cmp skiplist.Scorer) int {
	it := cmp.(*scoreScore)
	//时间越小，权重越高
	if item.EnterTime < it.EnterTime {
		return skiplist.Big
	}
	if item.EnterTime == it.EnterTime {
		return skiplist.Equal
	}
	return skiplist.Small
}

func (item *scoreScore) ByteSize() int64 {
	return int64(proto.Size(item.Value))
}

// NewQueue 创建队列
func NewQueue(subcfg subConfig) *Queue {
	return &Queue{
		Queue:     skiplist.NewQueue(subcfg.PoolCacheSize),
		subConfig: subcfg,
	}
}

//func (cache *Queue) newSkipValue(item *mempool.Item) (*skiplist.SkipValue, error) {
//	size := proto.Size(item.Value)
//	return &skiplist.SkipValue{Score: cache.subConfig.PriceConstant*(item.Value.Fee/int64(size))*
//		cache.subConfig.PricePower - cache.subConfig.TimeParam*item.EnterTime, Value: item}, nil
//}

//GetItem 获取数据通过 key
func (cache *Queue) GetItem(hash string) (*mempool.Item, error) {
	item, err := cache.Queue.GetItem(hash)
	if err != nil {
		return nil, err
	}
	return item.(*scoreScore).Item, nil
}

// Push 把给定tx添加到Queue；如果tx已经存在Queue中或Mempool已满则返回对应error
func (cache *Queue) Push(item *mempool.Item) error {
	return cache.Queue.Push(&scoreScore{Item: item, subConfig: cache.subConfig})
}

// Walk 遍历整个队列
func (cache *Queue) Walk(count int, cb func(value *mempool.Item) bool) {
	cache.Queue.Walk(count, func(item skiplist.Scorer) bool {
		return cb(item.(*scoreScore).Item)
	})
}

// GetProperFee 获取合适的手续费
func (cache *Queue) GetProperFee() int64 {
	var sumScore int64
	var properFeerate int64
	if cache.Size() == 0 {
		return cache.subConfig.ProperFee
	}
	i := 0
	cache.Queue.Walk(0, func(score skiplist.Scorer) bool {
		if i == 100 {
			return false
		}
		sumScore += score.GetScore()
		i++
		return true
	})
	//这里的int64(100)是一般交易的大小
	properFeerate = (sumScore/int64(i) + cache.subConfig.TimeParam*time.Now().Unix()) * int64(100) /
		(cache.subConfig.PriceConstant * cache.subConfig.PricePower)
	return properFeerate
}
