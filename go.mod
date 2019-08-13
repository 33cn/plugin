module github.com/33cn/plugin

go 1.12

replace (
	github.com/coreos/etcd => github.com/etcd-io/etcd v3.3.13+incompatible
	go.etcd.io/bbolt => github.com/etcd-io/bbolt v1.3.1-etcd.8
	go.etcd.io/etcd => github.com/etcd-io/etcd v3.3.13+incompatible
)

require (
	github.com/33cn/chain33 v6.1.1-0.20190812064448-5dd036921d46+incompatible
	github.com/AndreasBriese/bbloom v0.0.0-20180913140656-343706a395b7 // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/NebulousLabs/Sia v1.3.7
	github.com/NebulousLabs/entropy-mnemonics v0.0.0-20170316012907-7b01a644a636 // indirect
	github.com/NebulousLabs/errors v0.0.0-20171229012116-7ead97ef90b8 // indirect
	github.com/NebulousLabs/fastrand v0.0.0-20180208210444-3cf7173006a0 // indirect
	github.com/NebulousLabs/merkletree v0.0.0-20181025040823-2a1d1d1dc33c // indirect
	github.com/XiaoMi/pegasus-go-client v0.0.0-20181029071519-9400942c5d1c // indirect
	github.com/apache/thrift v0.0.0-20171203172758-327ebb6c2b6d // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/btcsuite/btcd v0.0.0-20181013004428-67e573d211ac
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d // indirect
	github.com/btcsuite/go-socks v0.0.0-20170105172521-4720035b7bfd // indirect
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792 // indirect
	github.com/coreos/bbolt v1.3.0 // indirect
	github.com/coreos/etcd v0.0.0-00010101000000-000000000000
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/blake256 v1.0.0 // indirect
	github.com/decred/base58 v1.0.0 // indirect
	github.com/dgraph-io/badger v1.5.4 // indirect
	github.com/dgryski/go-farm v0.0.0-20180109070241-2de33835d102 // indirect
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/haltingstate/secp256k1-go v0.0.0-20151224084235-572209b26df6 // indirect
	github.com/hashicorp/golang-lru v0.5.0
	github.com/huin/goupnp v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mr-tron/base58 v1.1.0 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pkg/errors v0.8.0
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/robertkrimen/otto v0.0.0-20180617131154-15f95af6e78d
	github.com/rs/cors v1.6.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v0.0.0-20181105012736-f9080354173f // indirect
	github.com/tjfoc/gmsm v0.0.0-20171124023159-98aa888b79d8
	github.com/valyala/fasthttp v1.4.0
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/grpc v1.22.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0-20170531160350-a96e63847dc3 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)
