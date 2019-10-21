package pos33

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/memberlist"
)

// 创建100个节点, 每个节点发送自己的 ID , 部分节点接收 p2p 的 UDP and TCP message

func createNodes() []*gossip {
	var gss []*gossip
	seeds := []string{"0.0.0.0:20000"}
	for i := 0; i < 100; i++ {
		pub := fmt.Sprintf("@@@+%03d", i)
		// log.Println(pub)
		conf := memberlist.DefaultWANConfig()
		conf.Name = pub
		conf.BindAddr = "0.0.0.0"
		conf.BindPort = 20000 + i
		// conf.DisableTcpPings = true
		// conf.GossipNodes = 100
		// conf.GossipVerifyIncoming = false
		// conf.GossipVerifyOutgoing = false
		// conf.EnableCompression = false
		// conf.ProbeInterval = time.Second * 10
		conf.GossipInterval = time.Millisecond * 10
		// conf.ProbeTimeout = time.Second * 1

		g := createGossip(conf, pub, "localhost:20000")
		seeds = append(seeds, conf.BindAddr+fmt.Sprintf(":%03d", conf.BindPort))
		gss = append(gss, g)
		if i > 0 {
			// g.Join([]string{seeds[i-1]})
			log.Println(g.conf.Name + "joined")
		}
		// time.Sleep(time.Second * 3)
	}

	return gss
}

func iTestGossip(t *testing.T) {
	fmt.Println("@@@@@@@@@@@ gossip test would use 5 s  @@@@@@@@@@@@@@")
	gss := createNodes()
	go func() {
		// 	for {
		for _, g := range gss {
			// g.broadcastCh <- []byte(g.conf.Name + ": hello")
			// g.send(g.conf.Name, []byte(g.conf.Name+"hahaha"))
			g.broadcastUDP([]byte("hello"))
		}
		// }
	}()

	done := make(chan struct{})
	time.AfterFunc(time.Second*15, func() {
		close(done)
	})

	N := 0
	for {
		for _, g := range gss {
			select {
			case s := <-g.C:
				fmt.Println("@@@@@@@@@@@", g.conf.Name+": "+string(s))
				// time.Sleep(time.Second)
				g.broadcastUDP([]byte(g.conf.Name + fmt.Sprintf(":%d", N)))
				N++
				// g.broadcastCh <- []byte(g.conf.Name + "hello tool")
			case <-done:
				fmt.Println("@@@@@@@@@@@ gossip test done @@@@@@@@@@@@@@")
				return
			default:
			}
		}
	}
}
