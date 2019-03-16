package price

import (
	"testing"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	drivers "github.com/33cn/chain33/system/mempool"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

var (
	c, _       = crypto.New(types.GetSignName("", types.SECP256K1))
	hex        = "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
	a, _       = common.FromHex(hex)
	privKey, _ = c.PrivKeyFromBytes(a)
	toAddr     = address.PubKeyToAddress(privKey.PubKey().Bytes()).String()
	amount     = int64(1e8)
	v          = &cty.CoinsAction_Transfer{Transfer: &types.AssetsTransfer{Amount: amount}}
	transfer   = &cty.CoinsAction{Value: v, Ty: cty.CoinsActionTransfer}
	tx1        = &types.Transaction{Execer: []byte("coins"), Payload: types.Encode(transfer), Fee: 1000000, Expire: 1, To: toAddr}
	tx2        = &types.Transaction{Execer: []byte("coins"), Payload: types.Encode(transfer), Fee: 1000000, Expire: 2, To: toAddr}
	tx3        = &types.Transaction{Execer: []byte("coins"), Payload: types.Encode(transfer), Fee: 1000000, Expire: 3, To: toAddr}
	tx4        = &types.Transaction{Execer: []byte("coins"), Payload: types.Encode(transfer), Fee: 2000000, Expire: 4, To: toAddr}
	tx5        = &types.Transaction{Execer: []byte("coins"), Payload: types.Encode(transfer), Fee: 1000000, Expire: 5, To: toAddr}
	item1      = &drivers.Item{Value: tx1, Priority: tx1.Fee, EnterTime: types.Now().Unix()}
	item2      = &drivers.Item{Value: tx2, Priority: tx2.Fee, EnterTime: types.Now().Unix()}
	item3      = &drivers.Item{Value: tx3, Priority: tx3.Fee, EnterTime: types.Now().Unix() - 1000}
	item4      = &drivers.Item{Value: tx4, Priority: tx4.Fee, EnterTime: types.Now().Unix() - 1000}
	item5      = &drivers.Item{Value: tx5, Priority: tx5.Fee, EnterTime: types.Now().Unix() - 1000}
)

func initEnv(size int64) *Queue {
	if size == 0 {
		size = 100
	}
	_, sub := types.InitCfg("chain33.test.toml")
	var subcfg subConfig
	types.MustDecode(sub.Mempool["price"], &subcfg)
	subcfg.PoolCacheSize = size
	cache := NewQueue(subcfg)
	return cache
}

func TestMemFull(t *testing.T) {
	cache := initEnv(1)
	hash := string(tx1.Hash())
	err := cache.Push(item1)
	assert.Nil(t, err)
	assert.Equal(t, true, cache.Exist(hash))
	it, err := cache.GetItem(hash)
	assert.Nil(t, err)
	assert.Equal(t, item1, it)

	_, err = cache.GetItem(hash + ":")
	assert.Equal(t, types.ErrNotFound, err)

	err = cache.Push(item1)
	assert.Equal(t, types.ErrTxExist, err)

	err = cache.Push(item2)
	assert.Equal(t, types.ErrMemFull, err)

	cache.Remove(hash)
	assert.Equal(t, 0, cache.Size())
}

func TestWalk(t *testing.T) {
	//push to item
	cache := initEnv(2)
	cache.Push(item1)
	cache.Push(item2)
	assert.Equal(t, 2, cache.Size())
	var data [2]*drivers.Item
	i := 0
	cache.Walk(1, func(value *drivers.Item) bool {
		data[i] = value
		i++
		return true
	})
	assert.Equal(t, 1, i)
	assert.Equal(t, data[0], item1)

	i = 0
	cache.Walk(2, func(value *drivers.Item) bool {
		data[i] = value
		i++
		return true
	})
	assert.Equal(t, 2, i)
	assert.Equal(t, data[0], item1)
	assert.Equal(t, data[1], item2)

	i = 0
	cache.Walk(2, func(value *drivers.Item) bool {
		data[i] = value
		i++
		return false
	})
	assert.Equal(t, 1, i)
}

func TestTimeCompetition(t *testing.T) {
	cache := initEnv(1)
	cache.Push(item1)
	cache.Push(item3)
	assert.Equal(t, false, cache.Exist(string(item1.Value.Hash())))
	assert.Equal(t, true, cache.Exist(string(item3.Value.Hash())))
}

func TestPriceCompetition(t *testing.T) {
	cache := initEnv(1)
	cache.Push(item3)
	cache.Push(item4)
	assert.Equal(t, false, cache.Exist(string(item3.Value.Hash())))
	assert.Equal(t, true, cache.Exist(string(item4.Value.Hash())))
}

func TestAddDuplicateItem(t *testing.T) {
	cache := initEnv(1)
	cache.Push(item1)
	err := cache.Push(item1)
	assert.Equal(t, types.ErrTxExist, err)
}

func TestQueueDirection(t *testing.T) {
	cache := initEnv(0)
	cache.Push(item1)
	cache.Push(item2)
	cache.Push(item3)
	cache.Push(item4)
	cache.Push(item5)
	cache.txList.Print()
	i := 0
	lastScore := cache.txList.GetIterator().First().Score
	var tmpScore int64
	cache.Walk(5, func(value *drivers.Item) bool {
		tmpScore = cache.txMap[string(value.Value.Hash())].Score
		if lastScore < tmpScore {
			return false
		}
		lastScore = tmpScore
		i++
		return true
	})
	assert.Equal(t, 5, i)
	assert.Equal(t, true, lastScore == cache.txList.GetIterator().Last().Score)
}

func TestGetProperFee(t *testing.T) {
	cache := initEnv(0)
	assert.Equal(t, cache.subConfig.ProperFee, cache.GetProperFee())

	cache.Push(item1)
	cache.Push(item4)
	cache.GetProperFee()
	assert.Equal(t, (item1.Value.Fee+item4.Value.Fee)/2, cache.GetProperFee())
}
