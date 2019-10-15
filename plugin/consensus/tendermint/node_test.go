package tendermint

import (
	"encoding/hex"
	"fmt"
	"github.com/33cn/chain33/types"
	"sync"
	"testing"

	"github.com/33cn/chain33/common/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	secureConnCrypto crypto.Crypto
	sum              = 0
	mutx             sync.Mutex
	privKey          = "B3DC4C0725884EBB7264B92F1D8D37584A64ADE1799D997EC64B4FE3973E08DE220ACBE680DF2473A0CB48987A00FCC1812F106A7390BE6B8E2D31122C992A19"
	expectAddress    = "02A13174B92727C4902DB099E51A3339F48BD45E"
)

func init() {
	cr2, err := crypto.New(types.GetSignName("", types.ED25519))
	if err != nil {
		fmt.Println("crypto.New failed for types.ED25519")
		return
	}
	secureConnCrypto = cr2
}

func TestParallel(t *testing.T) {
	Parallel(
		func() {
			mutx.Lock()
			sum++
			mutx.Unlock()
		},
		func() {
			mutx.Lock()
			sum += 2
			mutx.Unlock()
		},
		func() {
			mutx.Lock()
			sum += 3
			mutx.Unlock()
		},
	)

	fmt.Println("TestParallel ok")
	assert.Equal(t, 6, sum)
}

func TestGenAddressByPubKey(t *testing.T) {
	tmp, err := hex.DecodeString(privKey)
	assert.Nil(t, err)

	priv, err := secureConnCrypto.PrivKeyFromBytes(tmp)
	assert.Nil(t, err)

	addr := GenAddressByPubKey(priv.PubKey())
	strAddr := fmt.Sprintf("%X", addr)
	assert.Equal(t, expectAddress, strAddr)
	fmt.Println("TestGenAddressByPubKey ok")
}

func TestIP2IPPort(t *testing.T) {
	testMap := NewMutexMap()
	assert.Equal(t, false, testMap.Has("1.1.1.1"))

	testMap.Set("1.1.1.1", "1.1.1.1:80")
	assert.Equal(t, true, testMap.Has("1.1.1.1"))

	testMap.Set("1.1.1.2", "1.1.1.2:80")
	assert.Equal(t, true, testMap.Has("1.1.1.2"))

	testMap.Delete("1.1.1.1")
	assert.Equal(t, false, testMap.Has("1.1.1.1"))
	fmt.Println("TestIP2IPPort ok")
}

func TestPeerSet(t *testing.T) {
	testSet := NewPeerSet()
	assert.Equal(t, false, testSet.Has("1"))

	peer1 := &peerConn{id: "1", ip: []byte("1.1.1.1")}
	testSet.Add(peer1)
	assert.Equal(t, true, testSet.Has("1"))
	assert.Equal(t, true, testSet.HasIP([]byte("1.1.1.1")))

	err := testSet.Add(peer1)
	assert.NotNil(t, err)

	peer2 := &peerConn{id: "2", ip: []byte("1.1.1.2")}
	testSet.Add(peer2)
	assert.Equal(t, true, testSet.Has("2"))
	assert.Equal(t, 2, testSet.Size())

	testSet.Remove(peer1)
	assert.Equal(t, 1, testSet.Size())
	assert.Equal(t, false, testSet.Has("1"))
	assert.Equal(t, false, testSet.HasIP([]byte("1.1.1.1")))

	fmt.Println("TestPeerSet ok")
}
