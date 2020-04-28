package pos33

import (
	"fmt"
	"testing"
	"time"

	ccrypto "github.com/33cn/chain33/common/crypto"
)

func TestGossip2(t *testing.T) {
	c, err := ccrypto.New("secp256k1")
	if err != nil {
		t.Error(err)
		return
	}
	priv1, err := c.GenKey()
	if err != nil {
		t.Error(err)
		return
	}
	priv2, err := c.GenKey()
	if err != nil {
		t.Error(err)
		return
	}
	g1 := newGossip2(priv1, "10001", "gosssipTest", "bar")
	g2 := newGossip2(priv2, "10002", "gosssipTest", "bar")

	g2.bootstrap(peerAddr(g1.h).String())
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		msg := []byte(fmt.Sprintf("%d ----------------- %d", i, i))
		g1.gossip("bar", msg)
		data := <-g2.C
		fmt.Println(string(data))
		time.Sleep(time.Millisecond * 100)
	}

	t.Log("go here")
}
