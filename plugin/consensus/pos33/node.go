package pos33

import (
	"encoding/hex"
	"math/big"
	"math/rand"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/difficulty"
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
	cps map[int64]map[int]map[string]*pt.Pos33ElectMsg
	// receive candidate verifers
	cvs map[int64]map[int]map[string]*pt.Pos33VoteMsg
	// receive candidate blocks
	cbs map[int64]map[string]*types.Block

	voteOkHeight int64

	lastBlock *types.Block
	bch       chan *types.Block
	etm       *time.Timer
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
		cps:  make(map[int64]map[int]map[string]*pt.Pos33ElectMsg),
		cvs:  make(map[int64]map[int]map[string]*pt.Pos33VoteMsg),
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

func (n *node) getNotNullBlock(height int64) (*types.Block, error) {
	for i := height; i >= height-10; i-- {
		b, err := n.RequestBlock(i)
		if err != nil {
			return nil, err
		}
		if len(b.Txs) > 0 {
			return b, nil
		}
	}
	return nil, nil
}

func (n *node) genRewordTx(height int64, round int, strHash string) (*types.Transaction, int, error) {
	var vs []*pt.Pos33VoteMsg
	mv := n.cvs[height][round]
	for _, v := range mv {
		if string(v.BlockHash) == strHash {
			vs = append(vs, v)
		}
	}

	vsw := vsWeight(vs)
	plog.Info("genRewordTx", "height", height, "vsw", vsw)
	randHash := n.ips[height+1].Rands.Rands[0].Hash

	data, err := proto.Marshal(&pt.Pos33Action{
		Value: &pt.Pos33Action_Miner{
			Miner: &pt.Pos33MinerAction{
				Votes:    vs,
				RandHash: randHash,
			},
		},
		Ty: pt.Pos33ActionMiner,
	})

	if err != nil {
		panic(err)
	}

	tx := &types.Transaction{
		Execer:  []byte("pos33"),
		To:      address.GetExecAddress("pos33").String(),
		Payload: data,
		Nonce:   rand.Int63(),
		Fee:     pos33MinFee,
		Expire:  time.Now().Unix() + 10,
	}
	return tx, vsw, nil
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

func vmWeight(vm map[string]*pt.Pos33VoteMsg, strHash string) int {
	w := 0
	for _, v := range vm {
		if string(v.BlockHash) == strHash {
			w += v.Weight()
		}
	}
	return w
}

func vsWeight(vs []*pt.Pos33VoteMsg) int {
	w := 0
	for _, v := range vs {
		w += v.Weight()
	}
	return w
}

func diff(w int) uint32 {
	// return types.GetP(0).PowLimitBits
	oldDiff := difficulty.CompactToBig(types.GetP(0).PowLimitBits)
	newDiff := new(big.Int).Add(oldDiff, big.NewInt(int64(w+1)))
	return difficulty.BigToCompact(newDiff)
}

func (n *node) makeBlock(round int) (*types.Block, error) {
	vh := n.lastBlock.Height
	height := vh + 1
	strHash := string(n.lastBlock.Hash())
	tx, w, err := n.genRewordTx(vh, round, strHash)
	if err != nil {
		plog.Error("genRewordTx error", "err", err.Error(), "height", height)
		return nil, err
	}
	tx.Sign(types.ED25519, n.priv)

	nb, err := n.newBlock(n.lastBlock, []*Tx{tx}, height)
	if err != nil {
		plog.Error("makeBlock error", "height", height, "error", err.Error())
		return nil, err
	}

	nb.Difficulty = diff(w)

	n.signBlock(nb)
	plog.Info("@@@@@@@ I make a block: ", "height", height, "hash", hexs(nb.Hash()), "txHash", hexs(nb.TxHash), "diff", nb.Difficulty)
	return nb, nil
}

func (n *node) addBlock(b *types.Block) {
	if !n.IsCaughtUp() {
		return
	}
	plog.Info("node.addBlock", "height", b.Height, "hash", hexs(b.Hash()))

	if n.lastBlock != nil && b.Height <= n.lastBlock.Height {
		plog.Info("addBlock nil", "height", b.Height)
		return
	}

	select {
	case n.bch <- b:
	case <-n.bch:
		n.bch <- b
	}
}

func (n *node) clear(height int64) {
	// clear the caches
	for h := range n.cbs {
		if h+10 <= height {
			delete(n.cbs, h)
		}
	}
	for h := range n.cvs {
		if h+10 <= height {
			delete(n.cvs, h)
		}
	}
	for h := range n.cps {
		if h+10 <= height {
			delete(n.cps, h)
		}
	}
	for h := range n.ips {
		if h+10 <= height {
			delete(n.ips, h)
		}
	}
	for h := range n.ivs {
		if h+10 <= height {
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
	strHash := string(b.Hash())
	plog.Info("node.handleBlock", "height", b.Height, "bp", addr(b.Signature), "hash", hexs(b.Hash()))

	if n.cbs[b.Height] == nil {
		n.cbs[b.Height] = make(map[string]*types.Block)
	}

	n.cbs[b.Height][strHash] = b
}

func (n *node) checkBlock(b, pb *types.Block) error {
	plog.Info("node.checkBlock", "height", b.Height)
	if b.Height < 2 {
		return nil
	}
	if !n.IsCaughtUp() {
		return nil
	}
	if len(b.Txs) == 0 {
		return nil
	}

	act, err := getBlockReword(b)
	if err != nil {
		return err
	}

	// check votes
	for _, v := range act.Votes {
		m := v.Elect
		a := addr(m.Sig)
		allw := n.allWeight(m.Height)
		w := n.getWeight(a, m.Height)
		err = pt.CheckRands(a, allw, w, m.Rands, m.Height, m.Seed, m.Sig, int(m.Round), int(m.Step))
		if err != nil {
			return err
		}
	}

	// check diff
	/*
		vws := vsWeight(act.Votes)
		if vws*3 < pt.Pos33VerifierSize*2 {
			err = errors.New("block votersize error")
			plog.Error(err.Error(), "height", b.Height, "len(votes)", vws)
			return err
		}
			if diff(vws) != b.Difficulty {
				err = errors.New("block difficulty error")
				plog.Error(err.Error(), "height", b.Height)
				return err
			}
	*/

	return nil
}

var zeroHash [32]byte

func (n *node) reSortition(height int64, round int) {
	startHeight := height - pt.Pos33SortitionSize - height%pt.Pos33SortitionSize
	b, err := n.RequestBlock(startHeight)
	if err != nil {
		panic(err)
	}
	seed := zeroHash[:]
	act, err := getBlockReword(b)
	if err != nil {
		plog.Error("getBlockReword err", "error", err, "height", b.Height)
	} else {
		seed = act.RandHash
	}
	const staps = 2
	for s := 0; s < staps; s++ {
		allw := n.allWeight(height)
		w := n.getWeight(n.addr, height)
		rands, sig := pt.GenRands(allw, w, n.priv, height, seed, round, s)
		if rands == nil {
			plog.Info("sortiton nil", "height", height)
			continue
		}
		plog.Info("node.resortition", "height", height, "allw", allw, "w", w, "weight", len(rands.Rands), "round", round)
		if s == 0 {
			n.ips[height] = &pt.Pos33ElectMsg{Rands: rands, Height: height, Seed: seed, Round: int32(round), Step: int32(s), Sig: sig}
		} else {
			n.ivs[height] = &pt.Pos33ElectMsg{Rands: rands, Height: height, Seed: seed, Round: int32(round), Step: int32(s), Sig: sig}
		}
	}
}

func (n *node) sortition(b *types.Block, round int) {
	seed := zeroHash[:]
	startHeight := int64(0)
	if b == nil {
		startHeight = int64(-pt.Pos33SortitionSize)
	} else {
		act, err := getBlockReword(b)
		if err != nil {
			plog.Error("getBlockReword err", "error", err, "height", b.Height)
		} else {
			seed = act.RandHash
			startHeight = b.Height
		}
	}
	startHeight += pt.Pos33SortitionSize + 1
	const staps = 2
	for s := 0; s < staps; s++ {
		for i := 0; i < pt.Pos33SortitionSize; i++ {
			height := startHeight + int64(i)
			allw := n.allWeight(height)
			w := n.getWeight(n.addr, height)
			rands, sig := pt.GenRands(allw, w, n.priv, height, seed, round, s)
			if rands == nil {
				plog.Info("sortiton nil", "height", height)
				continue
			}
			plog.Info("node.sortition", "height", height, "allw", allw, "w", w, "weight", len(rands.Rands))
			if s == 0 {
				n.ips[height] = &pt.Pos33ElectMsg{Rands: rands, Height: height, Seed: seed, Round: int32(round), Step: int32(s), Sig: sig}
			} else {
				n.ivs[height] = &pt.Pos33ElectMsg{Rands: rands, Height: height, Seed: seed, Round: int32(round), Step: int32(s), Sig: sig}
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
		plog.Info("votemsg error", "error", "elect msg too late", "height", m.Height)
		return
	}

	/*
		_, ok := n.ivs[m.Height+1]
		if !ok {
			return
		}
	*/

	a := addr(m.Sig)
	allw := n.allWeight(m.Height)
	w := n.getWeight(a, m.Height)
	plog.Info("handleVoteMsg", "height", m.Height, "voter", a, "weight", len(m.Rands.Rands), "bhash", hexs(vm.BlockHash))

	err := pt.CheckRands(a, allw, w, m.Rands, m.Height, m.Seed, m.Sig, int(m.Round), int(m.Step))
	if err != nil {
		plog.Error("votemsg check rands error", "err", err.Error(), "allw", allw, "w", w)
		return
	}

	height := m.Height
	if n.cvs[height] == nil {
		n.cvs[height] = make(map[int]map[string]*pt.Pos33VoteMsg)
	}
	round := int(m.Round)
	if n.cvs[height][round] == nil {
		mp := make(map[string]*pt.Pos33VoteMsg)
		n.cvs[height][round] = mp
	}
	n.cvs[height][round][a] = vm
	if n.voteOkHeight >= height {
		return
	}
	strHash := string(vm.BlockHash)

	if vmWeight(n.cvs[height][round], strHash)*3 > pt.Pos33VerifierSize*2 {
		b, ok := n.cbs[height][strHash]
		if !ok {
			return
		}
		plog.Info("@@@ set block 2f+1 @@@", "height", height, "bp", addr(b.Signature), "hash", hexs(vm.BlockHash))
		n.voteOkHeight = b.Height

		n.setBlock(b)
	}
}

func (n *node) handleElectMsg(m *pt.Pos33ElectMsg) {
	a := addr(m.Sig)
	height := m.Height
	plog.Info("handleElectMsg", "height", m.Height, "address", addr(m.Sig))
	err := pt.CheckRands(a, n.allWeight(m.Height), n.getWeight(a, m.Height), m.Rands, m.Height, m.Seed, m.Sig, int(m.Round), int(m.Step))
	if err != nil {
		plog.Info("check rand error:", "error", err.Error())
		return
	}
	if n.cps[height] == nil {
		n.cps[height] = make(map[int]map[string]*pt.Pos33ElectMsg)
	}
	round := int(m.Round)
	if n.cps[height][round] == nil {
		n.cps[height][round] = make(map[string]*pt.Pos33ElectMsg)
	}
	n.cps[height][round][a] = m
}

func (n *node) handlePos33Msg(pm *pt.Pos33Msg) bool {
	if pm == nil {
		return false
	}
	switch pm.Ty {
	case pt.Pos33Msg_B:
		var m pt.Pos33BlockMsg
		err := types.Decode(pm.Data, &m)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		//n.handleElectMsg(m.Elect)
		n.handleBlock(m.B)
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
	n.sortition(nil, 0)
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

	n.etm = time.NewTimer(time.Hour)
	ch := make(chan int64, 1)

	if lb.Height == 0 {
		n.firstSortition(lb)
		time.AfterFunc(time.Second, func() { n.addBlock(lb) })
	}
	round := 0

	for {
		select {
		case msg := <-msgch:
			n.handlePos33Msg(msg)
		case height := <-ch:
			if height == n.lastBlock.Height+1 {
				round++
				plog.Info("vote timeout: ", "height", height, "round", round)
				n.reSortition(height, round)
				n.elect(height, round)
				/*
					time.AfterFunc(time.Second*1, func() {
						ch <- height
					})
				*/
			}
		case <-n.etm.C:
			height := n.lastBlock.Height + 1
			plog.Info("elect timeout: ", "height", height)
			n.vote(height, round)

			to := round * 2
			if to == 0 {
				to = 1
			}
			if to > 4 {
				to = 4
			}

			time.AfterFunc(time.Second*time.Duration(to), func() {
				ch <- height
			})
		case b := <-n.bch: // new block add to chain
			if b.Height < n.voteOkHeight {
				break
			}
			if b.Height%pt.Pos33SortitionSize == 0 {
				n.sortition(b, 0)
			}
			round = 0
			n.lastBlock = b
			plog.Info("set last block", "height", b.Height)
			n.elect(b.Height+1, 0)
			n.clear(b.Height - 1)
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

func (n *node) vote(height int64, round int) {
	e, ok := n.ivs[height]
	if !ok {
		plog.Info("I'm not verifer", "height", height)
		return
	}

	var pes []*pt.Pos33ElectMsg
	for _, e := range n.cps[height][round] {
		pes = append(pes, e)
	}
	plog.Info("vote len(pes)", "height", height, "len(pes)", len(pes))

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
		plog.Info("NO block vote out")
		return
	}
	v := &pt.Pos33VoteMsg{Elect: e, BlockHash: vb.Hash()}
	v.Sign(n.priv)
	n.handleVoteMsg(v)
	n.gss.broadcastTCP(marshalVoteMsg(v))
}

func (n *node) elect(height int64, round int) {
	n.etm = time.NewTimer(time.Millisecond * 300)
	pm, ok := n.ips[height]
	if !ok {
		plog.Info("elect: I'm not Proposer", "height", height, "round", round)
		return
	}
	nb, err := n.makeBlock(round)
	if err != nil {
		plog.Error(err.Error(), "height", height)
		return
	}
	n.handleElectMsg(pm)
	n.handleBlock(nb)
	n.gss.broadcastTCP(marshalElectMsg(pm))
	n.gss.broadcastTCP(marshalBlockMsg(&pt.Pos33BlockMsg{B: nb}))
}

func marshalBlockMsg(m *pt.Pos33BlockMsg) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(m),
		Ty:   pt.Pos33Msg_B,
	}
	return types.Encode(pm)
}

func marshalElectMsg(m *pt.Pos33ElectMsg) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(m),
		Ty:   pt.Pos33Msg_E,
	}
	return types.Encode(pm)
}

func marshalVoteMsg(v *pt.Pos33VoteMsg) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(v),
		Ty:   pt.Pos33Msg_V,
	}
	return types.Encode(pm)
}
