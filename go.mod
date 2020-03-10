module github.com/33cn/plugin

go 1.12

require (
	github.com/33cn/chain33 v0.0.0-20200306092706-a15573bb187f
	github.com/BurntSushi/toml v0.3.1
	github.com/NebulousLabs/Sia v1.3.7
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/protobuf v1.3.2
	github.com/hashicorp/golang-lru v0.5.3
	github.com/hashicorp/memberlist v0.1.5
	github.com/libp2p/go-conn-security v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.5.0
	github.com/libp2p/go-libp2p-autonat-svc v0.1.0
	github.com/libp2p/go-libp2p-core v0.3.0
	github.com/libp2p/go-libp2p-host v0.1.0 // indirect
	github.com/libp2p/go-libp2p-interface-connmgr v0.1.0 // indirect
	github.com/libp2p/go-libp2p-interface-pnet v0.1.0 // indirect
	github.com/libp2p/go-libp2p-kad-dht v0.5.0
	github.com/libp2p/go-libp2p-metrics v0.1.0 // indirect
	github.com/libp2p/go-libp2p-mplex v0.2.1
	github.com/libp2p/go-libp2p-net v0.1.0 // indirect
	github.com/libp2p/go-libp2p-peerstore v0.1.4
	github.com/libp2p/go-libp2p-protocol v0.1.0 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.2.5
	github.com/libp2p/go-libp2p-quic-transport v0.2.2
	github.com/libp2p/go-libp2p-routing v0.1.0
	github.com/libp2p/go-libp2p-secio v0.2.1
	github.com/libp2p/go-libp2p-tls v0.1.2
	github.com/libp2p/go-libp2p-transport v0.1.0 // indirect
	github.com/libp2p/go-libp2p-yamux v0.2.1
	github.com/multiformats/go-multiaddr v0.2.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/rs/cors v1.6.0
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tjfoc/gmsm v0.0.0-20171124023159-98aa888b79d8
	github.com/valyala/fasthttp v1.5.0
	github.com/whyrusleeping/go-smux-multiplex v3.0.16+incompatible // indirect
	github.com/whyrusleeping/go-smux-multistream v2.0.2+incompatible // indirect
	github.com/whyrusleeping/go-smux-yamux v2.0.9+incompatible // indirect
	github.com/whyrusleeping/yamux v1.2.0 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.2.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20200109152110-61a87790db17
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	google.golang.org/grpc v1.22.1
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)

replace github.com/33cn/chain33 v0.0.0-20200205062829-bb33acc5e2e8 => /Users/w/go/src/github.com/33cn/chain33

replace github.com/dgraph-io/badger v1.6.0 => github.com/dgraph-io/badger v1.5.5
