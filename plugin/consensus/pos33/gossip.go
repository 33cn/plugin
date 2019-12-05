package pos33

import (
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

type gossipDelegate struct {
	// broadcastCh chan []byte
	recvCh chan []byte
	pub    string
	gss    *gossip
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (gd *gossipDelegate) NodeMeta(limit int) []byte {
	return []byte(gd.pub)
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed
func (gd *gossipDelegate) NotifyMsg(b []byte) {
	p := make([]byte, len(b))
	copy(p, b)
	gd.recvCh <- p
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit. Care should be taken that this method does not block,
// since doing so would block the entire UDP packet receive loop.
func (gd *gossipDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil //gd.gss.broadcasts.GetBroadcasts(overhead, limit)
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (gd *gossipDelegate) LocalState(join bool) []byte {
	return nil
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (gd *gossipDelegate) MergeRemoteState(buf []byte, join bool) {
}

type gossip struct {
	conf *memberlist.Config
	*memberlist.Memberlist
	//broadcasts *memberlist.TransmitLimitedQueue

	tcpCh, udpCh chan []byte
	recvCh       chan []byte
	eventCh      chan memberlist.NodeEvent
	C            chan []byte
}

func ipPort(addr string) (string, int) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	iport, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	return host, iport
}

func createGossip(conf *memberlist.Config, publicKey, seedAddr string) *gossip {
	recvCh := make(chan []byte, 1024)

	d := &gossipDelegate{recvCh: recvCh, pub: publicKey}
	conf.Delegate = d
	ml, err := memberlist.Create(conf)
	if err != nil {
		plog.Crit(err.Error())
	}

	if len(seedAddr) > 0 {
		_, err = ml.Join([]string{seedAddr})
		if err != nil {
			plog.Crit(err.Error())
		}
		plog.Info("@@@@@@@: memberlist", "members", ml.Members())
	}

	/*
		broadcasts := &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				n := ml.NumMembers()
				return n
			},
			RetransmitMult: 1,
		}
	*/

	gss := &gossip{
		Memberlist: ml,
		//broadcasts: broadcasts,
		conf:  conf,
		C:     recvCh,
		tcpCh: make(chan []byte, 16),
		udpCh: make(chan []byte, 16),
	}
	d.gss = gss
	return gss
}

func newGossip(publicKey, listenAddr, advertiseAddr, bootAddr string) *gossip {
	conf := memberlist.DefaultWANConfig()
	conf.Name = publicKey
	addr, port := ipPort(listenAddr)
	conf.BindAddr = addr
	conf.BindPort = port
	if advertiseAddr != "" {
		addr, port = ipPort(advertiseAddr)
		conf.AdvertiseAddr = addr
		conf.AdvertisePort = port
		plog.Info("@@@@@@@ newGossip", "advertiseAddr", addr, "advertisePort", port)
	}
	conf.ProtocolVersion = memberlist.ProtocolVersion2Compatible

	return createGossip(conf, publicKey, bootAddr)
}

type brInfo struct {
	mb  *memberlist.Node
	msg []byte
}

func (g *gossip) runBroadcast() {
	N := 4
	tcpCh := make(chan brInfo, N)
	udpCh := make(chan brInfo, N)
	for i := 0; i < N; i++ {
		go func() {
			for {
				br := <-tcpCh
				g.SendReliable(br.mb, br.msg)
			}
		}()
		go func() {
			for {
				br := <-udpCh
				g.SendBestEffort(br.mb, br.msg)
			}
		}()
	}

	tch := time.NewTicker(time.Second * 60).C
	for {
		select {
		case <-tch:
			plog.Info("@@@@@@@: memberlist", "members", g.Members())
		case msg := <-g.tcpCh:
			for _, mb := range g.Members() {
				if mb.Name != g.conf.Name {
					tcpCh <- brInfo{mb, msg}
				}
			}
		case msg := <-g.udpCh:
			for _, mb := range g.Members() {
				if mb.Name != g.conf.Name {
					udpCh <- brInfo{mb, msg}
				}
			}
		}
	}
}

func (g *gossip) broadcastTCP(msg []byte) {
	// for _, mb := range g.Members() {
	// 	if mb.Name != g.conf.Name {
	// 		mb := mb
	// 		go g.SendReliable(mb, msg)
	// 	}
	// }
	g.tcpCh <- msg
}

func (g *gossip) broadcastUDP(msg []byte) {
	// for _, mb := range g.Members() {
	// 	if mb.Name != g.conf.Name {
	// 		mb := mb
	// 		go g.SendBestEffort(mb, msg)
	// 	}
	// }
	g.udpCh <- msg
}

func (g *gossip) gossip(msg []byte) {
	g.tcpCh <- msg
	/*
		g.broadcasts.QueueBroadcast(&broadcast{
			msg:    msg,
			notify: nil,
		})
	*/
}

func (g *gossip) send(public string, msg []byte) error {
	ms := g.Members()
	for _, m := range ms {
		if m.Name == public {
			// use UDP send
			return g.SendBestEffort(m, msg)
		}
	}
	return errors.New(public + "peer is not online")
}
