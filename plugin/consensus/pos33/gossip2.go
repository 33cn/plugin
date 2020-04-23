package pos33

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	ccrypto "github.com/33cn/chain33/common/crypto"

	"github.com/libp2p/go-libp2p"
	autonat "github.com/libp2p/go-libp2p-autonat-svc"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	routing "github.com/libp2p/go-libp2p-routing"
	secio "github.com/libp2p/go-libp2p-secio"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/multiformats/go-multiaddr"
)

type mdnsNotifee struct {
	h   host.Host
	ctx context.Context
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if m.h.Network().Connectedness(pi.ID) != network.Connected {
		plog.Info("peer mdns found", "pid", pi.ID.String())
		m.h.Connect(m.ctx, pi)
	}
}

type gossip2 struct {
	C      chan []byte
	h      host.Host
	tmap   map[string]*pubsub.Topic
	ctx    context.Context
	cancel context.CancelFunc
}

func (g *gossip2) bootstrap(addrs ...string) error {
	for _, addr := range addrs {
		targetAddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			plog.Error("bootstrap error", "err", err)
			return err
		}

		targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
		if err != nil {
			plog.Error("bootstrap error", "err", err)
			return err
		}

		err = g.h.Connect(g.ctx, *targetInfo)
		if err != nil {
			plog.Error("bootstrap error", "err", err)
			return err
		}
		plog.Info("@@@@@@@ connect boot peer", "bootpeer", targetAddr.String())
	}
	return nil
}

func newGossip2(priv ccrypto.PrivKey, port, stag string, topics ...string) *gossip2 {
	ctx, cancel := context.WithCancel(context.Background())
	pr, err := crypto.UnmarshalSecp256k1PrivateKey(priv.Bytes())
	if err != nil {
		panic(err)
	}
	h := newHost(ctx, pr, port, stag)
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}
	go func() {
		for range time.NewTicker(time.Minute).C {
			plog.Info("@@@@@@@ ", "peers", ps.ListPeers(topics[0]))
		}
	}()
	tmap := make(map[string]*pubsub.Topic)
	for _, t := range topics {
		topic, err := ps.Join(t)
		if err != nil {
			panic(err)
		}
		tmap[t] = topic
	}
	g := &gossip2{C: make(chan []byte, 16), h: h, tmap: tmap, ctx: ctx, cancel: cancel}
	g.run()
	return g
}

func (g *gossip2) run() {
	smap := make(map[string]*pubsub.Subscription)
	for t, topic := range g.tmap {
		s, err := topic.Subscribe()
		if err != nil {
			panic(err)
		}
		smap[t] = s
	}
	for _, sb := range smap {
		go func(s *pubsub.Subscription) {
			for {
				m, err := s.Next(g.ctx)
				if err != nil {
					panic(err)
				}
				id, err := peer.IDFromBytes(m.From)
				if err != nil {
					panic(err)
				}

				if g.h.ID().String() == id.String() {
					continue
				}
				g.C <- m.Data
			}
		}(sb)
	}
}

func (g *gossip2) gossip(topic string, data []byte) error {
	t, ok := g.tmap[topic]
	if !ok {
		return fmt.Errorf("%s topic NOT match", topic)
	}
	return t.Publish(g.ctx, data)
}

func newHost(ctx context.Context, priv crypto.PrivKey, port, stag string) host.Host {
	var idht *dht.IpfsDHT
	h, err := libp2p.New(ctx,
		// Use the keypair we generated
		libp2p.Identity(priv),
		// Multiple listen addresses
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/"+port, // regular tcp connections
		),
		// libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		// support secio connections
		libp2p.Security(secio.ID, secio.New),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			dht, err := dht.New(ctx, h)
			idht = dht
			return idht, err
		}),
		// Let this host use relays and advertise itself on relays if
		// it finds it is behind NAT. Use libp2p.Relay(options...) to
		// enable active relays and more.
		libp2p.EnableRelay(),
	)
	if err != nil {
		panic(err)
	}

	paddr := peerAddr(h)
	err = ioutil.WriteFile("yccpeeraddr.txt", []byte(paddr.String()), 0644)
	if err != nil {
		panic(err)
	}
	plog.Info("@@@@@@@ host inited", "host", paddr)

	// If you want to help other peers to figure out if they are behind
	// NATs, you can launch the server-side of AutoNAT too (AutoRelay
	// already runs the client)
	_, err = autonat.NewAutoNATService(ctx, h,
		// Support same non default security and transport options as
		// original host.
		libp2p.Security(secio.ID, secio.New),
		libp2p.DefaultTransports,
	)
	if err != nil {
		panic(err)
	}

	err = idht.Bootstrap(ctx)
	if err != nil {
		panic(err)
	}

	mdns, err := discovery.NewMdnsService(ctx, h, time.Second*10, stag)
	if err != nil {
		panic(err)
	}

	mn := &mdnsNotifee{h: h, ctx: ctx}
	mdns.RegisterNotifee(mn)

	// routingDiscovery := disc.NewRoutingDiscovery(idht)
	// disc.Advertise(ctx, routingDiscovery, string("ycc"))
	// peers, err := disc.FindPeers(ctx, routingDiscovery, string("ycc"))
	// for _, peer := range peers {
	// 	mn.HandlePeerFound(peer)
	// }

	// The last step to get fully up and running would be to connect to
	// bootstrap peers (or any other peers). We leave this commented as
	// this is an example and the peer will die as soon as it finishes, so
	// it is unnecessary to put strain on the network.

	/*
		// This connects to public bootstrappers
		for _, addr := range dht.DefaultBootstrapPeers {
			pi, _ := peer.AddrInfoFromP2pAddr(addr)
			// We ignore errors as some bootstrap peers may be down
			// and that is fine.
			h.Connect(ctx, *pi)
		}
	*/

	return h
}

func peerAddr(h host.Host) multiaddr.Multiaddr {
	peerInfo := &peerstore.PeerInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	addrs, err := peerstore.InfoToP2pAddrs(peerInfo)
	if err != nil {
		panic(err)
	}
	return addrs[0]
}
