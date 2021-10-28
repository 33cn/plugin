package gossip

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/common/pubsub"
	"google.golang.org/grpc/credentials"

	"github.com/33cn/chain33/p2p"

	"github.com/33cn/chain33/p2p/utils"

	"os"
	"strings"
	"testing"

	"github.com/33cn/chain33/client"

	l "github.com/33cn/chain33/common/log"

	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/wallet"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	testChannel = int32(119)
)

func init() {
	l.SetLogLevel("err")
}

func processMsg(q queue.Queue) {
	go func() {
		cfg := q.GetConfig()
		wcli := wallet.New(cfg)
		client := q.Client()
		wcli.SetQueueClient(client)
		defer func(path string) {
			_ = os.RemoveAll(path)
		}(cfg.GetModuleConfig().Wallet.DbPath)
		//导入种子，解锁钱包
		password := "a12345678"
		seed := "cushion canal bitter result harvest sentence ability time steel basket useful ask depth sorry area course purpose search exile chapter mountain project ranch buffalo"
		saveSeedByPw := &types.SaveSeedByPw{Seed: seed, Passwd: password}
		_, err := wcli.GetAPI().ExecWalletFunc("wallet", "SaveSeed", saveSeedByPw)
		if err != nil {
			return
		}
		walletUnLock := &types.WalletUnLock{
			Passwd:         password,
			Timeout:        0,
			WalletOrTicket: false,
		}

		_, err = wcli.GetAPI().ExecWalletFunc("wallet", "WalletUnLock", walletUnLock)
		if err != nil {
			return
		}
	}()

	go func() {
		blockchainKey := "blockchain"
		client := q.Client()
		client.Sub(blockchainKey)
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventGetBlocks:
				if req, ok := msg.GetData().(*types.ReqBlocks); ok {
					if req.Start == 1 {
						msg.Reply(client.NewMessage(blockchainKey, types.EventBlocks, &types.Transaction{}))
					} else {
						msg.Reply(client.NewMessage(blockchainKey, types.EventBlocks, &types.BlockDetails{}))
					}
				} else {
					msg.ReplyErr("Do not support", types.ErrInvalidParam)
				}

			case types.EventGetHeaders:
				if req, ok := msg.GetData().(*types.ReqBlocks); ok {
					if req.Start == 10 {
						msg.Reply(client.NewMessage(blockchainKey, types.EventHeaders, &types.Transaction{}))
					} else {
						msg.Reply(client.NewMessage(blockchainKey, types.EventHeaders, &types.Headers{}))
					}
				} else {
					msg.ReplyErr("Do not support", types.ErrInvalidParam)
				}

			case types.EventGetLastHeader:
				msg.Reply(client.NewMessage("p2p", types.EventHeader, &types.Header{Height: 2019}))
			case types.EventGetBlockHeight:

				msg.Reply(client.NewMessage("p2p", types.EventReplyBlockHeight, &types.ReplyBlockHeight{Height: 2019}))

			}

		}

	}()

	go func() {
		mempoolKey := "mempool"
		client := q.Client()
		client.Sub(mempoolKey)
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventGetMempoolSize:
				msg.Reply(client.NewMessage("p2p", types.EventMempoolSize, &types.MempoolSize{Size: 0}))
			}
		}
	}()
}

//new p2p
func newP2p(cfg *types.Chain33Config, port int32, dbpath string, q queue.Queue) *P2p {
	p2pCfg := cfg.GetModuleConfig().P2P
	p2pCfg.Enable = true
	p2pCfg.DbPath = dbpath
	p2pCfg.DbCache = 4
	p2pCfg.Driver = "leveldb"
	p2pMgr := p2p.NewP2PMgr(cfg)
	p2pMgr.Client = q.Client()
	p2pMgr.SysAPI, _ = client.New(p2pMgr.Client, nil)

	pcfg := &subConfig{}
	types.MustDecode(cfg.GetSubConfig().P2P[P2PTypeName], pcfg)
	pcfg.Port = port
	pcfg.Channel = testChannel
	pcfg.ServerStart = true
	subCfgBytes, _ := json.Marshal(pcfg)
	p2pcli := New(p2pMgr, subCfgBytes).(*P2p)

	p2pcli.node.nodeInfo.addrBook.initKey()
	privkey, _ := p2pcli.node.nodeInfo.addrBook.GetPrivPubKey()
	p2pcli.node.nodeInfo.addrBook.bookDb.Set([]byte(privKeyTag), []byte(privkey))
	p2pcli.node.nodeInfo.SetServiceTy(7)
	p2pcli.StartP2P()
	return p2pcli
}

//free P2p
func freeP2p(p2p *P2p) {
	p2p.CloseP2P()
	if err := os.RemoveAll(p2p.p2pCfg.DbPath); err != nil {
		log.Error("removeTestDbErr", "err", err)
	}
}

func testP2PEvent(t *testing.T, p2p *P2p) {
	msgs := make([]*queue.Message, 0)
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventBlockBroadcast, &types.Block{}))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventTxBroadcast, &types.Transaction{}))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventFetchBlocks, &types.ReqBlocks{}))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventGetMempool, nil))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventPeerInfo, nil))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventGetNetInfo, nil))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventFetchBlockHeaders, &types.ReqBlocks{}))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventAddBlacklist, &types.BlackPeer{PeerAddr: "192.168.1.1:13802"}))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventDelBlacklist, &types.BlackPeer{PeerAddr: "192.168.1.1:13802"}))
	msgs = append(msgs, p2p.client.NewMessage("p2p", types.EventShowBlacklist, &types.ReqNil{}))

	for _, msg := range msgs {
		p2p.mgr.PubSub.Pub(msg, P2PTypeName)
	}

}
func testNetInfo(t *testing.T, p2p *P2p) {
	p2p.node.nodeInfo.IsNatDone()
	p2p.node.nodeInfo.SetNatDone()
	p2p.node.nodeInfo.Get()
	p2p.node.nodeInfo.Set(p2p.node.nodeInfo)
	assert.NotNil(t, p2p.node.nodeInfo.GetListenAddr())
	assert.NotNil(t, p2p.node.nodeInfo.GetExternalAddr())
}

//测试Peer
func testPeer(t *testing.T, p2p *P2p, q queue.Queue) {
	cfg := types.NewChain33Config(types.ReadFile("../../../chain33.toml"))
	conn, err := grpc.Dial("localhost:53802", grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")))
	assert.Nil(t, err)
	defer conn.Close()

	remote, err := NewNetAddressString(fmt.Sprintf("127.0.0.1:%d", p2p.node.listenPort))
	assert.Nil(t, err)

	localP2P := newP2p(cfg, 43802, "testPeer", q)
	defer freeP2p(localP2P)

	t.Log(localP2P.node.CacheBoundsSize())
	t.Log(localP2P.node.GetCacheBounds())
	_, localPeerName := localP2P.node.nodeInfo.addrBook.GetPrivPubKey()
	localP2P.node.RemoveCachePeer("localhost:12345")
	assert.False(t, localP2P.node.HasCacheBound("localhost:12345"))
	peer, err := P2pComm.dialPeer(remote, localP2P.node)
	t.Log("peerName", peer.GetPeerName(), "self peerName", localPeerName)
	assert.Nil(t, err)
	defer peer.Close()
	peer.MakePersistent()
	localP2P.node.addPeer(peer)

	var info *innerpeer
	t.Log("WaitRegisterPeerStart...")
	trytime := 0
	for info == nil || info.p2pversion == 0 {
		trytime++
		time.Sleep(time.Millisecond * 100)
		info = p2p.node.server.p2pserver.getInBoundPeerInfo(localPeerName)
		if trytime > 100 {
			return
		}
	}
	exist, _ := p2p.node.isInBoundPeer(localPeerName)
	assert.True(t, exist)
	t.Log("WaitRegisterPeerStop...")
	p2pcli := NewNormalP2PCli()
	num, err := p2pcli.GetInPeersNum(peer)
	assert.Equal(t, 1, num)
	assert.Nil(t, err)
	tx1 := &types.Transaction{Execer: []byte("testTx1")}
	tx2 := &types.Transaction{Execer: []byte("testTx2")}
	localP2P.node.pubToPeer(&types.P2PTx{Tx: tx1}, peer.GetPeerName())
	p2p.node.server.p2pserver.pubToStream(&types.P2PTx{Tx: tx2}, info.name)
	t.Log("WaitRegisterTxFilterStart...")
	for !(txHashFilter.Contains(hex.EncodeToString(tx1.Hash())) &&
		txHashFilter.Contains(hex.EncodeToString(tx1.Hash()))) {
		time.Sleep(time.Millisecond * 10)
	}
	t.Log("WaitRegisterTxFilterStop")

	localP2P.node.AddCachePeer(peer)
	assert.Equal(t, localP2P.node.CacheBoundsSize(), len(localP2P.node.GetCacheBounds()))
	peer.GetRunning()
	localP2P.node.nodeInfo.FetchPeerInfo(localP2P.node)
	peers, infos := localP2P.node.GetActivePeers()
	assert.Equal(t, len(peers), len(infos))
	localP2P.node.flushNodePort(43803, 43802)

	localP2P.node.nodeInfo.peerInfos.SetPeerInfo(nil)
	localP2P.node.nodeInfo.peerInfos.GetPeerInfo("1222")
	t.Log(p2p.node.GetRegisterPeer(localPeerName))
	//测试发送Ping消息
	err = p2pcli.SendPing(peer, localP2P.node.nodeInfo)
	assert.Nil(t, err)

	//获取peer节点的被连接数
	pnum, err := p2pcli.GetInPeersNum(peer)
	assert.Nil(t, err)
	assert.Equal(t, 1, pnum)

	_, err = peer.GetPeerInfo()
	assert.Nil(t, err)
	//获取节点列表
	_, err = p2pcli.GetAddrList(peer)
	assert.Nil(t, err)

	_, err = p2pcli.SendVersion(peer, localP2P.node.nodeInfo)
	assert.Nil(t, err)
	t.Log("nodeinfo", localP2P.node.nodeInfo)
	t.Log(p2pcli.CheckPeerNatOk("localhost:53802", localP2P.node.nodeInfo))
	t.Log("checkself:", p2pcli.CheckSelf("loadhost:43803", localP2P.node.nodeInfo))
	_, err = p2pcli.GetAddr(peer)
	assert.Nil(t, err)

	localP2P.node.pubsub.FIFOPub(&types.P2PTx{Tx: &types.Transaction{}, Route: &types.P2PRoute{}}, "tx")
	localP2P.node.pubsub.FIFOPub(&types.P2PBlock{Block: &types.Block{}}, "block")
	//	//测试获取高度
	height, err := p2pcli.GetBlockHeight(localP2P.node.nodeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int(height), 2019)
	assert.Equal(t, false, p2pcli.CheckSelf("localhost:53802", localP2P.node.nodeInfo))
	//测试下载
	job := NewDownloadJob(NewP2PCli(localP2P).(*Cli), []*Peer{peer})

	job.GetFreePeer(1)

	var ins []*types.Inventory
	var inv types.Inventory
	inv.Ty = msgBlock
	inv.Height = 2
	ins = append(ins, &inv)
	var bChan = make(chan *types.BlockPid, 256)
	job.syncDownloadBlock(peer, ins[0], bChan)
	respIns := job.DownloadBlock(ins, bChan)
	t.Log(respIns)
	job.ResetDownloadPeers([]*Peer{peer})
	t.Log(job.avalidPeersNum())
	job.setBusyPeer(peer.GetPeerName())
	job.setFreePeer(peer.GetPeerName())
	job.removePeer(peer.GetPeerName())
	job.CancelJob()

	localP2P.node.addPeer(peer)
	assert.True(t, localP2P.node.needMore())
	peer.Close()
	localP2P.node.remove(peer.GetPeerName())
}

//测试grpc 多连接
func testGrpcConns(t *testing.T) {
	var conns []*grpc.ClientConn

	for i := 0; i < maxSamIPNum; i++ {
		conn, err := grpc.Dial("localhost:53802", grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")))
		assert.Nil(t, err)

		cli := types.NewP2PgserviceClient(conn)
		_, err = cli.GetHeaders(context.Background(), &types.P2PGetHeaders{
			StartHeight: 0, EndHeight: 0, Version: 1002}, grpc.FailFast(true))
		assert.Equal(t, false, strings.Contains(err.Error(), "no authorized"))
		conns = append(conns, conn)
	}

	conn, err := grpc.Dial("localhost:53802", grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")))
	assert.Nil(t, err)
	cli := types.NewP2PgserviceClient(conn)
	_, err = cli.GetHeaders(context.Background(), &types.P2PGetHeaders{
		StartHeight: 0, EndHeight: 0, Version: 1002}, grpc.FailFast(true))
	assert.Equal(t, true, strings.Contains(err.Error(), "no authorized"))

	conn.Close()
	for _, conn := range conns {
		conn.Close()
	}

}

//测试grpc 流多连接
func testGrpcStreamConns(t *testing.T, p2p *P2p) {

	conn, err := grpc.Dial("localhost:53802", grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")))
	assert.Nil(t, err)
	cli := types.NewP2PgserviceClient(conn)
	var p2pdata types.P2PGetData
	resp, err := cli.GetData(context.Background(), &p2pdata)
	assert.Nil(t, err)
	_, err = resp.Recv()
	assert.Equal(t, true, strings.Contains(err.Error(), "no authorized"))

	ping, err := P2pComm.NewPingData(p2p.node.nodeInfo)
	assert.Nil(t, err)

	_, err = cli.ServerStreamSend(context.Background(), ping)
	assert.Nil(t, err)

	_, err = cli.ServerStreamRead(context.Background())
	assert.Nil(t, err)
	var emptyBlock types.P2PBlock

	_, err = cli.BroadCastBlock(context.Background(), &emptyBlock)
	assert.Equal(t, true, strings.Contains(err.Error(), "no authorized"))

	conn.Close()

}

func testP2pComm(t *testing.T, p2p *P2p) {

	addrs := P2pComm.AddrRouteble([]string{"localhost:53802"}, utils.CalcChannelVersion(testChannel, VERSION), nil, nil)
	t.Log(addrs)
	i32 := P2pComm.BytesToInt32([]byte{0xff})
	t.Log(i32)
	_, _, err := P2pComm.GenPrivPubkey()
	assert.Nil(t, err)
	ping, err := P2pComm.NewPingData(p2p.node.nodeInfo)
	assert.Nil(t, err)
	assert.Equal(t, true, P2pComm.CheckSign(ping))
	assert.IsType(t, "string", P2pComm.GetLocalAddr())
	assert.Equal(t, 5, len(P2pComm.RandStr(5)))
}

func testAddrBook(t *testing.T, p2p *P2p) {

	prv, pub, err := P2pComm.GenPrivPubkey()
	if err != nil {
		t.Log(err.Error())
		return
	}

	t.Log("priv:", hex.EncodeToString(prv), "pub:", hex.EncodeToString(pub))

	pubstr, err := P2pComm.Pubkey(hex.EncodeToString(prv))
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("GenPubkey:", pubstr)

	addrBook := p2p.node.nodeInfo.addrBook
	addrBook.AddOurAddress(NewNetAddressIPPort(net.ParseIP("127.0.0.1"), 1234))
	addrBook.AddAddress(nil, nil)
	addrBook.AddAddress(NewNetAddressIPPort(net.ParseIP("127.0.0.1"), 1234), nil)
	addrBook.AddAddress(NewNetAddressIPPort(net.ParseIP("127.0.0.2"), 1234), nil)
	assert.True(t, addrBook.ISOurAddress(NewNetAddressIPPort(net.ParseIP("127.0.0.1"), 1234)))
	assert.True(t, addrBook.IsOurStringAddress("127.0.0.1:1234"))
	assert.Equal(t, addrBook.Size(), len(addrBook.GetPeers()))
	addrBook.setAddrStat("127.0.0.2:1234", true)
	addrBook.setAddrStat("127.0.0.2:1234", false)
	addrBook.saveToDb()
	addrBook.GetPeerStat("locolhost:43802")
	addrBook.genPubkey(hex.EncodeToString(prv))
	assert.Equal(t, addrBook.genPubkey(hex.EncodeToString(prv)), pubstr)
	addrBook.Save()
	addrBook.GetPeers()
	addrBook.GetAddrs()
	addrBook.ResetPeerkey("", "")
	privkey, _ := addrBook.GetPrivPubKey()
	assert.NotEmpty(t, privkey)
	addrBook.ResetPeerkey(hex.EncodeToString(prv), pubstr)
	resetkey, _ := addrBook.GetPrivPubKey()
	assert.NotEqual(t, resetkey, privkey)
}

func testRestart(t *testing.T, p2p *P2p) {
	client := p2p.client
	assert.False(t, p2p.isRestart())
	p2p.txFactory <- struct{}{}
	p2p.processEvent(client.NewMessage("p2p", types.EventTxBroadcast, &types.Transaction{}), 128, p2p.p2pCli.BroadCastTx)
	atomic.StoreInt32(&p2p.restart, 1)
	p2p.ReStart()
	atomic.StoreInt32(&p2p.restart, 0)
	p2p.ReStart()
}

func Test_p2p(t *testing.T) {
	cfg := types.NewChain33Config(types.ReadFile("../../../chain33.toml"))
	q := queue.New("channel")
	q.SetConfig(cfg)
	go q.Start()
	processMsg(q)
	p2p := newP2p(cfg, 53802, "testP2p", q)
	p2p.Wait()
	defer freeP2p(p2p)
	defer q.Close()
	testP2PEvent(t, p2p)
	testNetInfo(t, p2p)
	testPeer(t, p2p, q)
	testGrpcConns(t)
	testGrpcStreamConns(t, p2p)
	testP2pComm(t, p2p)
	testAddrBook(t, p2p)
	testRestart(t, p2p)
}

func Test_AddDelStream(t *testing.T) {

	s := NewP2pServer()
	peerName := "testpeer"
	delChan := s.addStreamHandler(peerName)
	//replace
	dataChan := s.addStreamHandler(peerName)

	_, ok := <-delChan
	assert.False(t, ok)

	//del old
	s.deleteStream(peerName, delChan)
	_, ok = s.streams[peerName]
	assert.True(t, ok)
	//del nil
	s.deleteStream("", delChan)
	//del exist
	s.deleteStream(peerName, dataChan)

	_, ok = s.streams[peerName]
	assert.False(t, ok)
}

func TestRandStr(t *testing.T) {
	t.Log(P2pComm.RandStr(5))
}

func TestBytesToInt32(t *testing.T) {

	t.Log(P2pComm.BytesToInt32([]byte{0xff}))
	t.Log(P2pComm.Int32ToBytes(255))
}

func TestComm_CheckNetAddr(t *testing.T) {
	_, _, err := P2pComm.ParaseNetAddr("192.16666.0.1")
	assert.NotNil(t, err)
	assert.Equal(t, "invalid ip", err.Error())
	_, _, err = P2pComm.ParaseNetAddr("192.169.0.1:899999")
	assert.NotNil(t, err)
	assert.Equal(t, "invalid port", err.Error())
	_, _, err = P2pComm.ParaseNetAddr("192.169.257.1:899")
	assert.NotNil(t, err)
	assert.Equal(t, "invalid ip", err.Error())
	_, _, err = P2pComm.ParaseNetAddr("192.169.1.1")
	assert.Nil(t, err)
	_, _, err = P2pComm.ParaseNetAddr("192.169.1.1:123")
	assert.Nil(t, err)

}

func TestBlackList_Add(t *testing.T) {
	bl := &BlackList{badPeers: make(map[string]int64)}
	bl.Add("192.168.1.1:13802", 3600)
	assert.True(t, bl.Has("192.168.1.1:13802"))
	pid := "0306c47d6b4e2abbacbf2285b083a1218b89ec70092ed0f0232577d111e9d94d6c"
	bl.addPeerStore(pid, "192.168.1.1:13802")
	_, ok := bl.getpeerStore(pid)
	assert.True(t, ok)
}

func TestBlackList_Delete(t *testing.T) {
	bl := &BlackList{badPeers: make(map[string]int64)}
	bl.Add("192.168.1.1:13802", 3600)
	pid := "0306c47d6b4e2abbacbf2285b083a1218b89ec70092ed0f0232577d111e9d94d6c"
	bl.addPeerStore(pid, "192.168.1.1:13802")
	bl.Delete("192.168.2.1")
	assert.True(t, bl.Has("192.168.1.1:13802"))
	assert.Equal(t, 1, len(bl.GetBadPeers()))
	bl.Delete("192.168.1.1:13802")
	assert.Equal(t, 0, len(bl.GetBadPeers()))
	bl.deletePeerStore(pid)
}

func TestSortArr(t *testing.T) {
	var Inventorys = make(Invs, 0)
	for i := 100; i >= 0; i-- {
		var inv types.Inventory
		inv.Ty = 111
		inv.Height = int64(i)
		Inventorys = append(Inventorys, &inv)
	}
	sort.Sort(Inventorys)
}

func TestCreds(t *testing.T) {
	cert := `-----BEGIN CERTIFICATE-----
MIIDdTCCAl2gAwIBAgIJAJ1Z/S9L51/5MA0GCSqGSIb3DQEBCwUAMFExCzAJBgNV
BAYTAkNOMQswCQYDVQQIDAJaSjELMAkGA1UEBwwCSFoxDDAKBgNVBAoMA0ZaTTEM
MAoGA1UECwwDRlpNMQwwCgYDVQQDDANMQlowHhcNMTgwNjI5MDMxNzEzWhcNMjgw
NjI2MDMxNzEzWjBRMQswCQYDVQQGEwJDTjELMAkGA1UECAwCWkoxCzAJBgNVBAcM
AkhaMQwwCgYDVQQKDANGWk0xDDAKBgNVBAsMA0ZaTTEMMAoGA1UEAwwDTEJaMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2OkNfozvtf5td2qgnDya9q+c
R+wjD69ZuWe3DkPeOI2H/wRqyeasCj51qDDd6kQEVoyVfVtNMQgMQUxvHxSt1QU9
rMp4zsm/aJaoeiYhJJH7l/FXLL4hYQ7LUSr2ee4at8fV9CCRh33DMpQ+50xGiWLt
IfRtzAqiKV7P6RO+jz3iCtedWLb71lGUfAQ89NlOJT6b0819hMd5wZpvrc1ZXfdm
copIHq6FsjwocoZ6cm2tY3L3NSk2WA8QY5Zej51aphAv6ZvhUBS0FEwPGX95AQpw
T209Gy/GW965dp6oR7LLLgXfWiCST49NH3Q6gP6j1r3KxTEk2g9aBhs9QQOksQID
AQABo1AwTjAdBgNVHQ4EFgQUiW78+xheZX7bdjFjCibo+3q2ZxMwHwYDVR0jBBgw
FoAUiW78+xheZX7bdjFjCibo+3q2ZxMwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0B
AQsFAAOCAQEAYkDwYepsJ734ytpfZY3D9HxR6fM2XdM0o35kQu1+lb2Ik+7oJKvT
SprSkL+l+1B3dYa4rLS8dztngR57js3BI6qgXavl3EeLf1gRSpAGul1uf+jkupOK
BgQ76TIlY88crbQw6Fkgrf9X9kfCbEDwoRZuX3aIWVpQtb+hkMoNI4wa8beWGWZK
EVaSxR1/QJIZIVxi5xcUQW2qdR/T4KvG3QVVcxJm2nZg2jexc5XopPNRLUfWZeXy
u8/Svlv5uH+2EqDGtYiDqmWlyGFJ3Q6lOGwCqRvhty7SYaHDZpV+10M32UuMBOOz
aHJJceqATq0U4NdzjbR0ygkApyDfv/5yfw==
-----END CERTIFICATE-----
`
	key := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2OkNfozvtf5td2qgnDya9q+cR+wjD69ZuWe3DkPeOI2H/wRq
yeasCj51qDDd6kQEVoyVfVtNMQgMQUxvHxSt1QU9rMp4zsm/aJaoeiYhJJH7l/FX
LL4hYQ7LUSr2ee4at8fV9CCRh33DMpQ+50xGiWLtIfRtzAqiKV7P6RO+jz3iCted
WLb71lGUfAQ89NlOJT6b0819hMd5wZpvrc1ZXfdmcopIHq6FsjwocoZ6cm2tY3L3
NSk2WA8QY5Zej51aphAv6ZvhUBS0FEwPGX95AQpwT209Gy/GW965dp6oR7LLLgXf
WiCST49NH3Q6gP6j1r3KxTEk2g9aBhs9QQOksQIDAQABAoIBAFgMilDDjw62X+Mo
olepdlIKyQqc/UKBsI9FXZZp0Evuj7TiKyMYAuAJqKMEayCMSaKEYK5JIJV1qdvb
1gOs1j9xaC665b0zQgKHnY4v6iY5KALuka904oHOipPBN9oY4DmX4e6+RbTVRSZD
7SDg4oUkZhPxb5xy+I6IkScQv7rO6aGiC7Z55K1X75/S3ga8mK2KaEKHat3mHDTo
h32D4jt9u0KzctXsUc6zpBE8gODZJ1DBN64TsX+6ZEov123PBounNanxxkn362cv
BIhD9iblOdWShxhNew1o7wCaD6ID24a/Y/dqSLWjWvYdAvs8s1KYijrlSVw6psBW
18esx9ECgYEA7NFUXVU9KQ8oY1wq8+vC41HYiXXpnU3xNOm2+qPIz2rSU35OUovp
icd706fhcoewQJ/cu25QWOwqHpCQl5Yo++cDSHtnRyMBlE6Z7NbxTFhfUzYV5UxW
fFcQGzZ17I93QoYleWaR66DeJOCnBKgePzNpTSjq69MN6HySwsL8Kq0CgYEA6nrs
VGFhcmbPTdKK7UZNjY3EG//+BITGEBSS0ouPAZudVbHggck+Xu/slHLgrEsQy/cJ
KBDpXN+rXE577BWATJdAjCe0DmoaArW6Lm4pdNUh7l8r79y9I3Xn85T+oWji65Vw
kolYlokpa1xuViYr3FhPhuH/VmRH8Q0mg/TRxpUCgYBifF/AfO11gOdEAxWd4WNo
VCZgbFgeYka4waWmMK0XjY4wyOtbqvIRqZNWn4/DqKhlB9atYCAsCvMtSOPJFtqu
gBE+eIun6ugCPHoJJA6vuGTUXz7V4FxrU23QU2LRYYywbsdw6HYw7vLTlVYAOsZx
dDkLrMOeFWTIVd5W/u4N9QKBgF5+eT0sHWRIMGTxY1Fp0pkoN479JDZX96XFVMIK
we/o8Yf2bj5/hmYmFFZi0U491iAMhyEhZ5oo/VruuhwTMigrkDSrT3G7qo3LBKPv
ez99IPZ6Xi+E6qgevQI52j/cEA7Wo446UXwg/JMqpcCME4LyB+KYsxjywtdO8GWf
RObdAoGBALP9HK7KuX7xl0cKBzOiXqnAyoMUfxvO30CsMI3DS0SrPc1p95OHswdu
/q1W3bMgctjEkgFljjDxDcdyYrPA2ZdXdY1An8nqZ9C48Hyvnj60Gfb3b6ycyZcb
/gd2v1Fb6oM82QBmWOFWaRSZj0UHZf8GvT09bs/SCSW4/hY/m4uC
-----END RSA PRIVATE KEY-----`

	certificate, err := tls.X509KeyPair([]byte(cert), []byte(key))
	assert.Nil(t, err)
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM([]byte(cert)) {
		return
	}
	var node Node
	node.nodeInfo = &NodeInfo{}
	servCreds := credentials.NewServerTLSFromCert(&certificate)
	cliCreds := credentials.NewClientTLSFromCert(cp, "")

	node.listenPort = 3331
	node.nodeInfo.servCreds = servCreds
	newListener("tcp", &node)
	netAddr, err := NewNetAddressString("localhost:3331")
	assert.Nil(t, err)

	conn, err := grpc.Dial(netAddr.String(), grpc.WithTransportCredentials(cliCreds))
	assert.Nil(t, err)
	assert.NotNil(t, conn)
	conn.Close()

	conn, err = grpc.Dial(netAddr.String())
	assert.NotNil(t, err)
	t.Log("without creds", err)
	assert.Nil(t, conn)
	conn, err = grpc.Dial(netAddr.String(), grpc.WithInsecure())
	assert.Nil(t, err)
	assert.NotNil(t, conn)

}

func TestCaCreds(t *testing.T) {

	ca := `-----BEGIN CERTIFICATE-----
MIIDKDCCAhCgAwIBAgIQMKlTasMav0IcCFxNKBlKlzANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTIxMTAxMTA3MzkwN1oXDTIyMTAxMTA3Mzkw
N1owEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBAOYK2OA6jsIWGK1faMZHdCMGKcc2SqErBcU/Sqis455B+9DCfZjesnut
5YgopQmvPKHF4ZROAJYtaLaodnEK7uMH2nYDU8Cy6+zXHG0c4FCnZxTiNlplYlrP
qSeDX/Ms2b1XmHAl8i289+4BbxWIj6JbMwPX7iQ68o4xo/D/FG+yfRs3xFEdwB6p
tC2TUNMBzaY/f1e43fC71AFd3xk5iUWRr2FPCqdQHpi5tHRYZ3SMxc630B/ISaDg
/DMCYzUdU7XfgehpeUrfVszMrIggwN3SM6bKGI7Zkt+mHMngAT5v0VdI3W8c6lI7
WFEsPq2n55XXDfzt9enbGQEIsv7mZC8CAwEAAaN6MHgwDgYDVR0PAQH/BAQDAgKk
MBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYE
FOfRwVXYMI6PvtWOxoLVI5OZCC4NMCEGA1UdEQQaMBiHBMCoAFiHBMCoAKSHBMCo
AHmHBMCoAKIwDQYJKoZIhvcNAQELBQADggEBAFGTcaltmP0pTPSZD/bvfnO2fUDp
OPG5ka0zeg6wNtUNQK43CACHSXtdlKtuaqZAJ1c8S/Ocsac9LJsXnI8NX75uCxf4
sdaEJN7mEN4lrqqrfihqbdeuDysbwUcoFjg7dzYIGZtMm2BR4kMaSqOHHWHoiUup
ylt2x864WHRvfHx53L8l2u3ZgnxHNZ+rk4VODGcpsnun1poHmfW+xJhkhc9U/lGw
GctxUtk6NUse9nZNxZG6ieSOD2+o5NSwUXliksPXzPkGQSx7VVXfG+4szBeXD+9x
mtQaeUpsIJdxsGcc0Zmu6v5XrBZ5xsZbCt8nMVA6rsGPYhczSXuBnVY6zu8=
-----END CERTIFICATE-----`

	cert := `-----BEGIN CERTIFICATE-----
MIIBzTCCAXSgAwIBAgIRAKA1R7bK7YPXBjHgoYqi+J0wCgYIKoZIzj0EAwIwQzEL
MAkGA1UEBhMCQ04xCzAJBgNVBAgTAlpKMQswCQYDVQQHEwJIWjEaMBgGA1UEAxMR
Y2hhaW4zMy1jYS1zZXJ2ZXIwHhcNMjExMDIyMDgwMTUyWhcNMjIwMTMwMDgwMTUy
WjBDMQswCQYDVQQGEwJDTjELMAkGA1UECBMCWkoxCzAJBgNVBAcTAkhaMRowGAYD
VQQDExFjaGFpbjMzLWNhLXNlcnZlcjBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BMJSLzYghkU4SpHvguL2pzwzg8GOcBG5n4QX10e7ScQFx1kUmcB0xZ/oyFMIdFBH
3BJ/0zwInlNAo0ekgUtRYlSjSTBHMA4GA1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8E
AjAAMCcGA1UdEQQgMB6HBMCoAFiHBMCoAHmHBMCoAKKHBMCoAKSHBMCoADswCgYI
KoZIzj0EAwIDRwAwRAIgBulQxbARTa9q6nA2ypZ5mX20dTactlPmLamI2xvaTU4C
ICQov1WBMv+P/pEL/CR8yKaVqggLa0B4KzDMji5u0zXd
-----END CERTIFICATE-----`

	key := `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgBabS0GvOURbOoP+u
mErJlKF2YVZfEwb2rjObA1q/hxqhRANCAATCUi82IIZFOEqR74Li9qc8M4PBjnAR
uZ+EF9dHu0nEBcdZFJnAdMWf6MhTCHRQR9wSf9M8CJ5TQKNHpIFLUWJU
-----END PRIVATE KEY-----`

	certificate, err := tls.X509KeyPair([]byte(cert), []byte(key))
	assert.Nil(t, err)
	cp := x509.NewCertPool()
	var node Node
	node.nodeInfo = &NodeInfo{}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM([]byte(ca)); !ok {
		assert.True(t, ok)
	}

	servCreds := newTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert, //校验客户端证书,用ca.pem校验
		ClientCAs:    certPool,
	})
	cliCreds := newTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName:   "",
		RootCAs:      certPool,
	})

	node.listenPort = 13332
	node.nodeInfo.servCreds = servCreds
	node.pubsub = pubsub.NewPubSub(10200)
	l := newListener("tcp", &node)
	assert.NotNil(t, l)
	go l.Start()
	defer l.Close()
	netAddr, err := NewNetAddressString(fmt.Sprintf("127.0.0.1:%v", node.listenPort))
	assert.Nil(t, err)

	conn, err := grpc.Dial(netAddr.String(), grpc.WithTransportCredentials(cliCreds))
	assert.Nil(t, err)
	assert.NotNil(t, conn)
	conn.Close()

	conn, err = grpc.Dial(netAddr.String())
	assert.NotNil(t, err)
	t.Log("without creds", err)
	assert.Nil(t, conn)
	conn, err = grpc.Dial(netAddr.String(), grpc.WithInsecure())
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	_, err = netAddr.DialTimeout(0, cliCreds, nil)
	assert.NotNil(t, err)
	t.Log(err.Error())

	cp = x509.NewCertPool()
	if !cp.AppendCertsFromPEM([]byte(cert)) {
		return
	}
	cliCreds = credentials.NewClientTLSFromCert(cp, "")
	_, err = netAddr.DialTimeout(0, cliCreds, nil)
	assert.NotNil(t, err)

}
