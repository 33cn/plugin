package pos33

import (
	"encoding/hex"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
	"github.com/golang/protobuf/proto"
)

var plog = log15.New("module", "pos33")

type committee struct {
	*pt.Pos33Rands
	height int64
	stoped bool
}

type node struct {
	*Client
	addr string
	gss  *gossip
	priv crypto.PrivKey

	// I'm candidate proposer in these blocks
	ips map[int64]*pt.Pos33ElectMsg
	// I'm candidate verifer in these blocks
	ivs map[int64]*pt.Pos33ElectMsg
	// receive candidate proposers
	cps map[int64]map[string]*pt.Pos33ElectMsg
	// receive candidate verifers
	cvs map[int64]map[string][]*pt.Pos33VoteMsg
	// receive candidate blocks
	cbs map[int64]map[string]*types.Block

	lastBlock *types.Block
	bch       chan *types.Block
}

// New create pos33 consensus client
func newNode(conf *subConfig) *node {
	priv := RootPrivKey
	if len(conf.Pos33SecretSeed) != 0 {
		kb, err := hex.DecodeString(conf.Pos33SecretSeed)
		if err != nil {
			plog.Error(err.Error())
		}
		priv, err = myCrypto.PrivKeyFromBytes(kb)
		if err != nil {
			plog.Error(err.Error())
			return nil
		}
	}
	addr := address.PubKeyToAddress(priv.PubKey().Bytes()).String()

	n := &node{
		addr: addr,
		priv: priv,
		ips:  make(map[int64]*pt.Pos33ElectMsg),
		ivs:  make(map[int64]*pt.Pos33ElectMsg),
		cps:  make(map[int64]map[string]*pt.Pos33ElectMsg),
		cvs:  make(map[int64]map[string][]*pt.Pos33VoteMsg),
		cbs:  make(map[int64]map[string]*types.Block),
		bch:  make(chan *types.Block, 1),
	}

	plog.Info("@@@@@@@ node start:", "addr", addr, "listenon", conf.Pos33ListenAddr)
	return n
}

func unmarshal(b []byte) (*pt.Pos33Msg, error) {
	var pm pt.Pos33Msg
	err := proto.Unmarshal(b, &pm)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (n *node) genRewordTx() (*types.Transaction, int, error) {
	var vs []*pt.Pos33VoteMsg
	height := n.lastBlock.Height
	plog.Info("genRewordTx", "height", height, "txs", len(n.lastBlock.Txs))
	if height == 0 || len(n.lastBlock.Txs) == 0 {
		vs = nil
	} else {
		strHash := string(n.lastBlock.Hash())
		vs = n.cvs[height][strHash]
	}
	data, err := proto.Marshal(&pt.Pos33Action{
		Value: &pt.Pos33Action_Reword{
			Reword: &pt.Pos33RewordAction{
				Votes: vs,
			},
		},
		Ty: pt.Pos33ActionReword,
	})

	if err != nil {
		plog.Error(err.Error())
		return nil, 0, err
	}

	tx := &types.Transaction{
		Execer:  []byte("pos33"),
		To:      address.GetExecAddress("pos33").String(),
		Payload: data,
		Fee:     pos33MinFee,
		Expire:  time.Now().Unix() + 10,
	}
	return tx, vsWeight(vs), nil
}

func (n *node) signBlock(b *types.Block) *types.Block {
	sig := n.priv.Sign(b.Hash())
	b.Signature = &types.Signature{Pubkey: n.priv.PubKey().Bytes(), Ty: types.ED25519, Signature: sig.Bytes()}
	return b
}

func vsAccWeight(vs []*pt.Pos33VoteMsg, acc string) (int, int) {
	for i, v := range vs {
		if addr(v.Sig) == acc {
			return v.Weight(), i
		}
	}
	return 0, -1
}

func vsWeight(vs []*pt.Pos33VoteMsg) int {
	w := 0
	for _, v := range vs {
		w += v.Weight()
	}
	return w
}

func (n *node) makeBlock(null bool) (*types.Block, error) {
	var txs []*types.Transaction
	diff := 0
	if !null {
		tx, w, err := n.genRewordTx()
		if err != nil {
			panic(err)
		}
		tx.Sign(types.ED25519, n.priv)
		txs = append(txs, tx)
		diff = w // diff = votes weights
	}
	height := n.lastBlock.Height + 1
	nb, err := n.newBlock(txs, height, null)
	if err != nil {
		plog.Error("makeBlock error", "height", height, "error", err.Error())
		return nil, err
	}
	nb.Difficulty += uint32(diff)

	if null {
		nb.MainHash = crypto.Sha256(n.lastBlock.MainHash)
	} else {
		nb.MainHash = n.ips[height].Rands.Rands[0].Hash
		n.signBlock(nb)
	}
	plog.Info("@@@@@@@ I make a block: ", "height", height, "isNull", null, "hash", hexs(nb.Hash()), "txHash", hexs(nb.TxHash))
	return nb, nil
}

func (n *node) addBlock(b *types.Block) {
	if !n.IsCaughtUp() {
		return
	}
	plog.Info("node.addBlock", "height", b.Height)
	select {
	case n.bch <- b:
	case <-n.bch:
		n.bch <- b
	}

	// clear the caches
	for h := range n.cbs {
		if h+10 <= b.Height {
			delete(n.cbs, h)
		}
	}
	for h := range n.cvs {
		if h+10 <= b.Height {
			delete(n.cvs, h)
		}
	}
	for h := range n.ips {
		if h+10 <= b.Height {
			delete(n.ips, h)
		}
	}
	for h := range n.ivs {
		if h+10 <= b.Height {
			delete(n.ivs, h)
		}
	}
}

func addr(sig *types.Signature) string {
	if sig == nil {
		return ""
	}
	return address.PubKeyToAddress(sig.Pubkey).String()
}

func (n *node) handleBlock(b *types.Block) {
	_, ok := n.ivs[b.Height]
	if !ok {
		return
	}
	strHash := string(b.Hash())
	plog.Info("node.handleBlock", "height", b.Height, "bp", addr(b.Signature), "hash", hexs(b.Hash()))

	if n.cbs[b.Height] == nil {
		n.cbs[b.Height] = make(map[string]*types.Block)
	}

	n.cbs[b.Height][strHash] = b
}

func (n *node) checkBlock(b *types.Block) error {
	if b.Height == 0 {
		return nil
	}

	plog.Info("node.checkBlock", "height", b.Height)
	go n.addBlock(b)
	return nil
}

var zeroHash [32]byte

func (n *node) sortition(seed []byte, startHeight int64) {
	startHeight += pt.Pos33SortitionSize + 1
	const staps = 2
	for s := 0; s < staps; s++ {
		for i := 0; i < pt.Pos33SortitionSize; i++ {
			height := startHeight + int64(i)
			allw := n.allWeight(height)
			w := n.getWeight(n.addr, height)
			rands, sig := pt.GenRands(allw, w, n.priv, height, seed, s)
			if rands == nil {
				plog.Info("sortiton nil", "height", height)
				continue
			}
			plog.Info("node.sortition", "height", height, "allw", allw, "w", w, "weight", len(rands.Rands))
			if s == 0 {
				n.ips[height] = &pt.Pos33ElectMsg{Rands: rands, Height: height, Seed: seed, Stap: int32(s), Sig: sig}
			} else {
				n.ivs[height] = &pt.Pos33ElectMsg{Rands: rands, Height: height, Seed: seed, Stap: int32(s), Sig: sig}
			}
		}
	}
}

func (n *node) handleVoteMsg(vm *pt.Pos33VoteMsg) {
	if n.lastBlock == nil {
		return
	}
	if !vm.Verify() {
		plog.Error("votemsg verify false")
		return
	}
	m := vm.Elect
	if m == nil || m.Height <= n.lastBlock.Height {
		plog.Info("votemsg error", "error", "elect msg too late")
		return
	}

	_, ok := n.ivs[m.Height]
	if !ok {
		return
	}

	a := addr(m.Sig)
	allw := n.allWeight(m.Height)
	w := n.getWeight(a, m.Height)
	plog.Info("handleVoteMsg", "height", m.Height, "voter", a, "weight", len(m.Rands.Rands))

	err := pt.CheckRands(a, allw, w, m.Rands, m.Height, m.Seed, m.Sig, int(m.Stap))
	if err != nil {
		plog.Error("votemsg check rands error", "err", err.Error(), "allw", allw, "w", w)
		return
	}

	if n.cvs[m.Height] == nil {
		n.cvs[m.Height] = make(map[string][]*pt.Pos33VoteMsg)
	}
	strHash := string(vm.BlockHash)
	n.cvs[m.Height][strHash] = append(n.cvs[m.Height][strHash], vm)

	if vsWeight(n.cvs[m.Height][strHash])*3 > pt.Pos33VeriferSize*2 {
		b, ok := n.cbs[m.Height][strHash]
		if !ok {
			return
		}
		plog.Info("@@@ set block 2f+1 @@@", "height", m.Height, "bp", addr(b.Signature))
		n.setBlock(b)
	}
}

func (n *node) voteTimeout(height int64) {
	_, ok := n.ivs[height]
	if !ok {
		return
	}
	max := 0
	maxHash := ""
	for hash, vs := range n.cvs[height] {
		vw := vsWeight(vs)
		if vw > max {
			max = vw
			maxHash = hash
		}
	}
	if max*3 > pt.Pos33VeriferSize {
		b, ok := n.cbs[height][maxHash]
		if !ok {
			panic("can't go here")
		}
		plog.Info("@@@ set block f+1 @@@", "height", height, "hash", hex.EncodeToString([]byte(maxHash)))
		n.setBlock(b)
	} else {
		plog.Error("@@@ vote error, make a null block @@@@ ", "height", height)
		b, err := n.makeBlock(true)
		if err != nil {
			plog.Error("make block error", "error", err, "height", height)
			return
		}
		n.setBlock(b)
	}
}

func (n *node) handleElectMsg(m *pt.Pos33ElectMsg) {
	a := addr(m.Sig)
	err := pt.CheckRands(a, n.allWeight(m.Height), n.getWeight(a, m.Height), m.Rands, m.Height, m.Seed, m.Sig, int(m.Stap))
	if err != nil {
		plog.Info("check rand error:", "error", err.Error())
		return
	}
	if n.cps[m.Height] == nil {
		n.cps[m.Height] = make(map[string]*pt.Pos33ElectMsg)
	}
	n.cps[m.Height][a] = m
}

func (n *node) handlePos33Msg(pm *pt.Pos33Msg) bool {
	if pm == nil {
		return false
	}
	switch pm.Ty {
	case pt.Pos33Msg_B:
		var b types.Block
		err := types.Decode(pm.Data, &b)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleBlock(&b)
	case pt.Pos33Msg_E:
		var m pt.Pos33ElectMsg
		err := types.Decode(pm.Data, &m)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleElectMsg(&m)
	case pt.Pos33Msg_V:
		var vt pt.Pos33VoteMsg
		err := types.Decode(pm.Data, &vt)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleVoteMsg(&vt)
	default:
		panic("not support this message type")
	}

	return true
}

// doGossipMsg multi-goroutine verify pos33 message
func (n *node) doGossipMsg() chan *pt.Pos33Msg {
	num := 4
	ch := make(chan *pt.Pos33Msg, num*16)
	for i := 0; i < num; i++ {
		go func() {
			for {
				pm, err := unmarshal(<-n.gss.C)
				if err != nil {
					plog.Error(err.Error())
					continue
				}
				ch <- pm
			}
		}()
	}
	return ch
}

func reseTm(tm *time.Timer, d time.Duration) {
	if !tm.Stop() {
		select {
		case <-tm.C:
		default:
		}
	}
	tm.Reset(d)
}

func (n *node) firstSortition(firtstBlock *types.Block) {
	n.sortition(zeroHash[:], -pt.Pos33SortitionSize)
}

func (n *node) runLoop() {
	lb, err := n.RequestLastBlock()
	if err != nil {
		plog.Error(err.Error())
		return
	}

	plog.Info("pos33 node runing.......", "last block height", lb.Height)

	n.gss = newGossip(n.priv.PubKey().KeyString(), n.conf.Pos33ListenAddr, n.conf.Pos33AdvertiseAddr, n.conf.Pos33PeerSeed)
	go n.gss.runBroadcast()
	msgch := n.doGossipMsg()

	tm := time.NewTimer(time.Hour)
	ch := make(chan int64, 1)

	if lb.Height == 0 {
		n.firstSortition(lb)
	}
	time.AfterFunc(time.Second, func() { n.addBlock(lb) })

	for {
		select {
		case msg := <-msgch:
			n.handlePos33Msg(msg)
		case height := <-ch:
			n.voteTimeout(height)
		case <-tm.C:
			height := n.lastBlock.Height + 1
			plog.Info("elect timeout: ", "height", height)
			n.vote(height)
			time.AfterFunc(time.Second*3, func() {
				if height == n.lastBlock.Height+1 {
					plog.Info("vote timeout: ", "height", height)
					ch <- height
				}
			})
		case b := <-n.bch: // new block add to chain
			if b.Height%pt.Pos33SortitionSize == 0 {
				n.sortition(b.MainHash, b.Height)
			}
			n.lastBlock = b
			n.elect()
			tm = time.NewTimer(time.Millisecond * 1000)
		}
	}
}

func hexs(b []byte) string {
	s := hex.EncodeToString(b)
	if len(s) <= 16 {
		return s
	}
	return s[:16]
}

func (n *node) vote(height int64) {
	e, ok := n.ivs[height]
	if !ok {
		plog.Info("I'm not verifer", "height", height)
		return
	}

	var pes []*pt.Pos33ElectMsg
	for _, e := range n.cps[height] {
		pes = append(pes, e)
	}

	rs := pt.Sortition(pes, 0)
	if rs == nil {
		plog.Info("sortition nil", "height", height)
		return
	}

	bp := rs.Rands[0].Addr
	plog.Info("vote bp", "height", height, "bp", bp, "len(cbs)", len(n.cbs[height]))
	var vb *types.Block
	for _, b := range n.cbs[height] {
		if addr(b.Signature) == bp {
			vb = b
			break
		}
	}
	if vb == nil {
		panic("NO block vote out")
	}
	v := &pt.Pos33VoteMsg{Elect: e, BlockHash: vb.Hash()}
	v.Sign(n.priv)
	n.handleVoteMsg(v)
	n.gss.broadcastTCP(n.marshalVoteMsg(v))
}

func (n *node) elect() {
	height := n.lastBlock.Height + 1
	pm, ok := n.ips[height]
	if !ok {
		return
	}
	nb, err := n.makeBlock(false)
	if err != nil {
		plog.Error(err.Error(), "height", height)
		return
	}
	n.handleBlock(nb)
	n.handleElectMsg(pm)
	n.gss.broadcastTCP(n.marshalElectMsg(pm))
	n.gss.broadcastTCP(n.marshalBlockMsg(nb))
}

func (n *node) marshalElectMsg(m *pt.Pos33ElectMsg) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(m),
		Ty:   pt.Pos33Msg_E,
	}
	return types.Encode(pm)
}

func (n *node) marshalBlockMsg(b *types.Block) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(b),
		Ty:   pt.Pos33Msg_B,
	}
	return types.Encode(pm)
}

func (n *node) marshalVoteMsg(v *pt.Pos33VoteMsg) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(v),
		Ty:   pt.Pos33Msg_V,
	}
	return types.Encode(pm)
}
