package price

import (
	"container/list"
	"fmt"

	"github.com/33cn/chain33/common/skiplist"
	"github.com/33cn/chain33/system/mempool"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

// Queue 价格队列模式(价格=手续费/交易字节数,价格高者优先,同价则时间早优先)
type Queue struct {
	txMap     map[string]*list.Element
	txList    *skiplist.SkipList
	subConfig subConfig
}

// NewQueue 创建队列
func NewQueue(subcfg subConfig) *Queue {
	return &Queue{
		make(map[string]*list.Element),
		skiplist.NewSkipList(&skiplist.SkipValue{Score: -1, Value: nil}),
		subcfg,
	}
}

/*
为了处理相同 Score 的问题，需要一个队列保存相同 Score 下面的交易
*/
func (cache *Queue) insertSkipValue(item *mempool.Item) *list.Element {
	txSize := proto.Size(item.Value)
	skvalue := &skiplist.SkipValue{Score: item.Value.Fee / int64(txSize)}
	value := cache.txList.Find(skvalue)
	var orderlist *list.List
	if value == nil { //new OrderList
		orderlist = list.New()
		skvalue.Value = orderlist
		cache.txList.Insert(skvalue)
	} else {
		orderlist = value.Value.(*list.List)
	}
	return orderlist.PushBack(item)
}

func (cache *Queue) newSkipValue(item *mempool.Item) *skiplist.SkipValue {
	txSize := proto.Size(item.Value)
	skvalue := &skiplist.SkipValue{Score: item.Value.Fee / int64(txSize)}
	return skvalue
}

//Exist 是否存在
func (cache *Queue) Exist(hash string) bool {
	_, exists := cache.txMap[hash]
	return exists
}

//GetItem 获取数据通过 key
func (cache *Queue) GetItem(hash string) (*mempool.Item, error) {
	if k, exist := cache.txMap[hash]; exist {
		return k.Value.(*mempool.Item), nil
	}
	return nil, types.ErrNotFound
}

// Push 把给定tx添加到Queue；如果tx已经存在Queue中或Mempool已满则返回对应error
func (cache *Queue) Push(item *mempool.Item) error {
	hash := item.Value.Hash()
	if cache.Exist(string(hash)) {
		return types.ErrTxExist
	}
	it := &mempool.Item{Value: item.Value, Priority: item.Value.Fee, EnterTime: item.EnterTime}
	sv := cache.newSkipValue(it)
	if int64(cache.txList.Len()) >= cache.subConfig.PoolCacheSize {
		tail := cache.txList.GetIterator().Last()
		lasthash := string(tail.Value.(*mempool.Item).Value.Hash())
		printhash("remove tail", []byte(lasthash))
		printhash("push hash", hash)
		fmt.Println("compare", sv.Compare(tail))
		//价格高存留
		switch sv.Compare(tail) {
		case -1:
			cache.Remove(lasthash)
		case 0:
			if sv.Value.(*mempool.Item).EnterTime < tail.Value.(*mempool.Item).EnterTime {
				cache.Remove(lasthash)
				break
			}
			return types.ErrMemFull
		case 1:
			return types.ErrMemFull
		default:
			return types.ErrMemFull
		}
	}
	cache.add(string(hash), sv)
	return nil
}

func (cache *Queue) add(hash string, item *list.Element) error {
	cache.txMap[string(hash)] = cache.insertSkipValue(item)
	return nil
}

// Remove 删除数据
func (cache *Queue) Remove(hash string) error {
	retcode := cache.txList.Delete()
	if retcode == 0 { //not found
		printhash("remove error", []byte(hash))
	}
	delete(cache.txMap, hash)
	return nil
}

// Size 数据总数
func (cache *Queue) Size() int {
	return cache.txList.Len()
}

// Walk 遍历整个队列
func (cache *Queue) Walk(count int, cb func(value *mempool.Item) bool) {
	i := 0
	cache.txList.Walk(func(item interface{}) bool {
		if !cb(item.(*mempool.Item)) {
			return false
		}
		i++
		return i != count
	})
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
	cache.txList.Walk(func(tx interface{}) bool {
		if i == 100 {
			return false
		}
		txSize = proto.Size(tx.(*mempool.Item).Value)
		feeRate = tx.(*mempool.Item).Value.Fee / int64(txSize/1000+1)
		sumFeeRate += feeRate
		i++
		return true
	})
	properFeeRate = sumFeeRate / int64(i)
	return properFeeRate
}

func printhash(title string, hash []byte) {
	fmt.Printf(title+" %x \n", hash)
}
