package pos33

import (
	"fmt"
	"log"
	"testing"
	"time"

	ccrypto "github.com/33cn/chain33/common/crypto"
)

// 创建100个节点, 每个节点发送自己的 ID , 部分节点接收 p2p 的 UDP and TCP message

func createNodes() []*gossip {
	var gss []*gossip
	seeds := []string{"localhost:20000"}
	boot := newGossip("xxx", seeds[0], "", "")
	gss = append(gss, boot)
	for i := 0; i < 5; i++ {
		pub := fmt.Sprintf("@@@+%03d", i)
		g := newGossip(pub, fmt.Sprintf("localhost:%d", 20001+i), "", seeds[0])
		gss = append(gss, g)
	}

	return gss
}

func TestGossip(t *testing.T) {
	fmt.Println("@@@@@@@@@@@ gossip test would use 5 s  @@@@@@@@@@@@@@")
	gss := createNodes()
	log.Println("haha go here")
	time.Sleep(3 * time.Second)
	gss[4].gossip([]byte("ahaha, hello"))
	/*
		go func() {
			for i, g := range gss {
				if i != 0 {
					g.gossip([]byte(fmt.Sprintf("hello%d", i)))
				}
			}
		}()
	*/

	done := make(chan struct{})
	time.AfterFunc(time.Second*10, func() {
		close(done)
	})

	N := 0
	for {
		for _, g := range gss {
			select {
			case s := <-g.C:
				//	g.gossip([]byte(fmt.Sprintf("hello:%d", N)))
				g.gossip([]byte(fmt.Sprintf("hello:%d", N)))

				N++
				fmt.Println(string(s))
			case <-done:
				t.Error("go here end")
				log.Println("@@@@@@@@@@@ gossip test done @@@@@@@@@@@@@@")
				return
			default:
			}
		}
	}
}

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
	g1 := newGossip2(priv1, "10001", "bar")
	g2 := newGossip2(priv2, "10002", "bar")

	g2.bootstrap(peerAddr(g1.h).String())
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		msg := []byte(fmt.Sprintf("%d ----------------- %d", i, i))
		g1.gossip("bar", msg)
		data := <-g2.C
		fmt.Println(string(data))
		time.Sleep(time.Millisecond * 100)
	}

	t.Error("go here")

}
