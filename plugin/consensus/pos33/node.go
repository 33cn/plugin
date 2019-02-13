package pos33

import (
	"encoding/hex"
	"log"
	"sort"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log/log15"
	pb "github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
	"github.com/golang/protobuf/proto"
)

var plog = log15.New("module", "pos33")

type blockChecker struct {
	b   *pb.Block
	rch chan bool
}

type prand struct {
	pub string
	r   *ty.Pos33Rand
}

type voter struct {
	*committee
	start   int64
	indexs  map[int64]prand // 每个成员的索引
	vs      map[int64]map[string]*ty.Pos33Vote
	voteOk  map[int64]bool
	notVote int64 // 错过的高度
}

func newVoter(c *committee, start int64) *voter {
	if c == nil {
		c = newCommittee(0, nil)
	}
	v := &voter{
		committee: c,
		start:     start,
		indexs:    make(map[int64]prand),
		vs:        make(map[int64]map[string]*ty.Pos33Vote),
		voteOk:    make(map[int64]bool),
		notVote:   -1,
	}
	v.sort()
	plog.Info("@@@@@@@ newVoter", "len(indexs)", len(v.indexs), "start", v.start, "vw", v.vw)
	return v
}

// check vote height
func (v *voter) checkVh(height int64) int {
	if height < v.start {
		return -1
	}
	if height >= v.start+int64(v.vw) {
		return 1
	}
	return 0
}

// check block height
func (v *voter) checkHeight(height int64) int {
	if height <= v.start {
		return -1
	}
	if height > v.start+int64(v.vw) {
		return 1
	}
	return 0
}

// committee is 共识委员会
type committee struct {
	base int64 // base block for sortition
	seed []byte
	vw   int // 总的投票权重 == 区块数量
	rs   map[string]*ty.Pos33Rands
}

func newCommittee(height int64, seed []byte) *committee {
	// seed := getBlockSeed(b)
	// height := int64(0)
	// if b != nil {
	// 	height = b.Height
	// }
	plog.Info("newCommittee", "base", height)
	return &committee{
		seed: seed,
		base: height,
		rs:   make(map[string]*ty.Pos33Rands),
	}
}

func (n *node) checkRands(rs *ty.Pos33Rands, seed []byte, height int64) bool {
	allw := n.allWeight()
	w := n.getWeight(rs.Pub)

	return checkRands(rs.Pub, allw, w, seed, rs, height) == nil
}

func (v *voter) sort() {
	if len(v.rs) == 0 {
		return
	}
	var rs ty.Pos33Rands
	mp := make(map[string]string)
	for pub, r := range v.rs {
		rs.Rands = append(rs.Rands, r.Rands...)
		for _, x := range r.Rands {
			mp[string(x.RandHash)] = pub
		}
	}
	sort.Sort(rs)
	for i, r := range rs.Rands {
		v.indexs[v.start+int64(i)+1] = prand{mp[string(r.RandHash)], r}
	}
}

const errorVote = "nil"

// something error: bp timeout or error or block is error.
func (v *voter) voteError(height int64, n *node) {
	if !v.am(n) {
		return
	}
	v.voteOk[height] = false
	h := height + 1
	lbp := v.indexs[height].pub
	for ; lbp == v.indexs[h].pub; h++ {
	}
	bp := v.indexs[h].pub
	vt := &ty.Pos33Vote{
		BlockHeight: height,
		BlockHash:   []byte(errorVote),
		Weight:      int32(len(v.rs[n.pub].Rands)),
		Bp:          bp,
	}
	plog.Info("vote error", "height", height, "vw", vt.Weight, "bp", hexString([]byte(vt.Bp)))
	vt.Sign(n.priv)
	n.gss.broadcastTCP(n.marshalVoteMsg(vt))
	v.handleVote(vt, n)
}

func (v *voter) am(n *node) bool {
	if v.committee == nil {
		return false
	}
	if _, ok := v.rs[n.pub]; !ok {
		return false // I'm not committee member
	}
	return true
}

func (v *voter) vote(b *pb.Block, n *node) {
	if !v.am(n) {
		return
	}
	vt := &ty.Pos33Vote{
		BlockHeight: b.Height,
		BlockHash:   b.Hash(),
		Weight:      int32(len(v.rs[n.pub].Rands)),
		Bp:          v.indexs[b.Height+1].pub,
	}
	plog.Info("vote", "height", b.Height, "vw", vt.Weight, "bp", hexString([]byte(vt.Bp)))
	vt.Sign(n.priv)
	n.gss.broadcastTCP(n.marshalVoteMsg(vt))
	v.handleVote(vt, n)
}

func (n *node) checkBlockHash(height int64, hash []byte) bool {
	b, err := n.RequestBlock(height)
	if err != nil {
		return false
	}
	return string(hash) == string(b.Hash())
}

func (v *voter) handleBlock(b *pb.Block, n *node) {
	if !v.checkBlock(b, n) {
		v.voteError(b.Height, n)
	}
}

func (v *voter) checkBlock(b *pb.Block, n *node) bool {
	// if !b.Verify() {
	// 	return
	// }
	pub := string(b.Signature.Pubkey)
	_, ok := v.rs[pub]
	if !ok {
		return false // not in committee
	}

	pr, ok := v.indexs[b.Height]
	if !ok {
		return false
	}

	if pr.pub != pub {
		return false //the creater of b not the bp
	}

	seed := getBlockSeed(b)
	if string(seed) != string(v.indexs[b.Height].r.RandHash) {
		return false // seed error
	}

	err := n.setBlock(b)
	if err != nil {
		plog.Error("setBlock error", "err", err)
		return false
	}

	return true
}

func (v *voter) checkVote(vt *ty.Pos33Vote, n *node) bool {
	pub := string(vt.Sig.Pubkey)
	if !vt.Verify() {
		// TODO: punish this voter
		plog.Error("handleVote verify error", "height", vt.BlockHeight, "pub", hexString([]byte(pub)))
		return false
	}

	// check vote
	rs, ok := v.rs[pub]
	if !ok {
		plog.Error("handleVote error", "pub", hexString([]byte(pub)))
		return false // not a voter
	}
	if len(rs.Rands) != int(vt.Weight) {
		plog.Error("handleVote error", "rands", len(rs.Rands), "weight", int(vt.Weight))
		return false // weight error
	}
	if v.indexs[vt.BlockHeight+1].pub != vt.Bp {
		plog.Error("handleVote error", "bp", hexString([]byte(vt.Bp)))
		return false // bp error
	}
	/*
		if !n.checkBlockHash(vt.BlockHeight, vt.BlockHash) {
			plog.Error("handleVote error", "block hash", hexString(vt.BlockHash))
			return false// block hash error
		}
	*/
	return true
}

func (v *voter) handleVote(vt *ty.Pos33Vote, n *node) {
	if !v.am(n) {
		return
	}
	height := vt.BlockHeight
	pub := string(vt.Sig.Pubkey)
	plog.Info("handleVote", "height", height, "block hash", hexString(vt.BlockHash), "bp", hexString([]byte(vt.Bp)), "vw", vt.Weight, "pub", hexString([]byte(pub)))

	if !vt.Verify() {
		// TODO: punish this voter
		plog.Error("handleVote verify error", "height", height, "pub", hexString([]byte(pub)))
		return
	}
	rs, ok := v.rs[pub]
	if !ok {
		plog.Error("handleVote error: not a voter", "pub", hexString([]byte(pub)))
		return // not a voter
	}
	if len(rs.Rands) != int(vt.Weight) {
		plog.Error("handleVote error", "rands", len(rs.Rands), "weight", int(vt.Weight))
		return // weight error
	}
	mp, ok := v.vs[height]
	if !ok {
		mp = make(map[string]*ty.Pos33Vote)
		v.vs[height] = mp
	}
	mp[pub] = vt

	var hash, bp string
	vmp := make(map[string]map[string]int)
	voteOK := false
	for _, x := range mp {
		bs := string(x.BlockHash)
		m, ok := vmp[bs]
		if !ok {
			m = make(map[string]int)
			vmp[bs] = m
		}
		m[x.Bp] += int(x.Weight)
		if m[x.Bp]*2 > v.vw { // should 50% or 67%
			hash = bs
			bp = x.Bp
			voteOK = true
		}
	}

	if !voteOK {
		return
	}
	plog.Info("vote ok!", "height", height, "block hash", hexString([]byte(hash)), "bp", hexString([]byte(bp)))
	if v.voteOk[height] {
		plog.Info("vote ok!, @@@@@@@@@@@@@@@@@")
		return
	}
	v.voteOk[height] = true

	if v.notVote == height {
		v.notVote = -1
		defer n.newRound(height)
	}
	if hash == errorVote {
		h := height + 1
		if hexString([]byte(bp)) == "nil" {
			errorBp := v.indexs[height].pub
			plog.Info("error bp", "err bp", hexString([]byte(errorBp)))
			for h = v.start + 1; errorBp == v.indexs[h].pub; h++ {
			}
			bp = v.indexs[h].pub
			plog.Info("error bp", "err bp", hexString([]byte(errorBp)), "new bp", hexString([]byte(bp)))
		} else {
			for ; bp != v.indexs[h].pub; h++ {
			}
		}
		v.indexs[height] = v.indexs[h]
		if bp == n.pub {
			// create null block
			plog.Info("create nil block", "height", height, "bp", hexString([]byte(bp)))
			n.makeBlock(height, v.indexs[height].r.RandHash, v.vs[height], true)
		}
	}
}

func (c *committee) addMember(rs *ty.Pos33Rands, n *node, fromSort bool) {
	plog.Info("addMember", "pub", hexString([]byte(rs.Pub)), "base", c.base)
	if rs.Height != c.base {
		plog.Error("addMember", "height", rs.Height)
		return
	}

	if !n.checkRands(rs, c.seed, c.base) {
		plog.Error("addMember: checkRands error", "pub", hexString([]byte(rs.Pub)))
		return
	}

	if rrs, ok := c.rs[rs.Pub]; ok {
		c.vw -= len(rrs.Rands)
	}

	c.rs[rs.Pub] = rs
	c.vw += len(rs.Rands)

	plog.Info("addMember", "vw", c.vw)

	// if _, ok := c.rs[n.pub]; ok && fromSort {
	// 	n.gossipNextCommittee(c, rs)
	// }
}

func (v *voter) newRound(lb *pb.Block, n *node) {
	lastHeight := lb.Height
	if !v.voteOk[lastHeight] {
		v.vote(lb, n)
		if !v.voteOk[lb.Height] {
			plog.Info("vote NOT ok", "height", lb.Height)
			v.notVote = lastHeight
			return
		}
	}
	v.notVote = -1
	height := lastHeight + 1
	bp := v.indexs[height].pub
	if bp == n.pub {
		_, err := n.makeBlock(height, v.indexs[height].r.RandHash, v.vs[lastHeight], v.voteOk[height])
		if err != nil {
			plog.Info("newBlock error", "err", err, "height", height)
			return
		}
	}
	plog.Info("@@@@@ voter.newRound", "height", height, "bp", hexString([]byte(bp)))
}

type node struct {
	*Client

	gss  *gossip
	priv crypto.PrivKey
	pub  string

	bch chan *pb.Block
	// cch chan int64

	bmp  map[int64]*pb.Block
	vmp  map[int64][]*ty.Pos33Vote
	rmp  map[int64][]*ty.Pos33Rands
	comm *committee
	v    *voter
}

// New create pos33 consensus client
func newNode(con *pb.Consensus) *node {
	priv := RootPrivKey
	if len(con.Pos33SecretSeed) != 0 {
		kb, err := hex.DecodeString(con.Pos33SecretSeed)
		if err != nil {
			plog.Error(err.Error())
		}
		priv, err = myCrypto.PrivKeyFromBytes(kb)
		if err != nil {
			plog.Error(err.Error())
			return nil
		}
	}

	log.SetFlags(log.Flags() | log.Lshortfile)

	n := &node{
		priv: priv,
		pub:  string(priv.PubKey().Bytes()),
		bch:  make(chan *pb.Block, 16),
		// cch:  make(chan int64, 1),
		bmp: make(map[int64]*pb.Block),
		vmp: make(map[int64][]*ty.Pos33Vote),
		rmp: make(map[int64][]*ty.Pos33Rands),
	}

	plog.Info("@@@@@@@ node start:", address.PubKeyToAddress([]byte(n.pub)), con.Pos33ListenAddr)
	return n
}

func unmarshal(b []byte) (*ty.Pos33Msg, error) {
	var pm ty.Pos33Msg
	err := proto.Unmarshal(b, &pm)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func genRewordTx(vs map[string]*ty.Pos33Vote, seed []byte) (*pb.Transaction, int) {
	w := 0
	var votes []*ty.Pos33Vote
	for _, v := range vs {
		votes = append(votes, v)
		w += int(v.Weight)
	}
	data, err := proto.Marshal(&ty.Pos33Action{
		Value: &ty.Pos33Action_Reword{
			Reword: &ty.Pos33RewordAction{
				Votes:    votes,
				RandHash: seed,
			},
		},
	})

	if err != nil {
		plog.Error(err.Error())
		return nil, 0
	}

	tx := &pb.Transaction{
		Execer:  []byte("pos33"),
		To:      address.GetExecAddress("pos33").String(),
		Payload: data,
		Fee:     pb.MinFee * 100,
	}
	return tx, w
}

func (n *node) makeBlock(height int64, newSeed []byte, vs map[string]*ty.Pos33Vote, null bool) (*pb.Block, error) {
	tx, w := genRewordTx(vs, newSeed)
	tx.Sign(pb.ED25519, n.priv)
	txs := []*pb.Transaction{tx}
	nb, err := n.newBlock(txs, height, null)
	if err != nil {
		return nil, err
	}
	nb.Difficulty += uint32(w)
	nb.Sign(n.priv)
	plog.Info("@@@@@@@ I make a block: ", "height", nb.Height, "blockTime", nb.BlockTime, "vs", len(vs))
	n.setBlock(nb)
	n.gss.broadcastTCP(n.marshalBlockMsg(nb))
	return nb, nil
}

// change committee
func (n *node) changeCommittee(c *committee, height int64) {
	plog.Info("node.changeCommittee", "height", height)
	n.v = newVoter(c, height)
	// var err error
	// if b == nil {
	// 	// if block error or bp timeout, should newRound
	// 	n.v.voteOk[height] = true
	// 	b, err = n.RequestBlock(height)
	// 	if err != nil {
	// 		plog.Error("handleVote error", "err", err)
	// 		return
	// 	}
	// 	n.newRound(b.Height)
	// }
	// n.cch <- height
	// n.comm = newCommittee(b)
	// n.sortition(height)

}

func (n *node) newRound(height int64) {
	plog.Info("n.newRound", "height", height)
	if n.v == nil {
		return
	}
	b, err := n.RequestBlock(height)
	if err != nil {
		plog.Error(err.Error())
		return
	}
	n.v.newRound(b, n)
}

// func (n *node) handleBft(m *ty.Pos33Msg) {
// 	n.comm.handleBft(m)
// }

func (n *node) handleRands(m *ty.Pos33Rands) {
	plog.Info("handleRands", "base", m.Height)
	if n.comm == nil {
		return
	}
	if m.Height < n.comm.base {
		return
	}
	if m.Height > n.comm.base {
		n.rmp[m.Height] = append(n.rmp[m.Height], m)
		return
	}
	for h, rs := range n.rmp {
		if h == n.comm.base {
			for _, r := range rs {
				n.comm.addMember(r, n, true)
			}
		}
	}
	delete(n.rmp, n.comm.base)
	if m.Height == n.comm.base {
		n.comm.addMember(m, n, true)
	}
}

func (n *node) addBlock(b *pb.Block) {
	if !n.IsCaughtUp() {
		return
	}
	plog.Info("node.addBlock", "height", b.Height)
	n.bch <- b
}

func (n *node) handleAddBlock(b *pb.Block) bool {
	vs := n.vmp[b.Height]
	for _, v := range vs {
		n.handleVote(v)
	}

	height := b.Height + 1
	nb, ok := n.bmp[height]
	if ok {
		n.handleBlock(nb)
		return false
	}

	n.clear(b.Height)
	return true
}

func (n *node) handleVote(vt *ty.Pos33Vote) {
	plog.Info("n.handleVote", "height", vt.BlockHeight, "pub", hexString(vt.Sig.Pubkey))
	height := vt.BlockHeight
	if n.v != nil {
		c := n.v.checkVh(height)
		if c < 0 {
			return
		}
		_, err := n.RequestBlock(height)
		if c > 0 || err != nil {
			n.vmp[height] = append(n.vmp[height], vt)
			return
		}
		n.v.handleVote(vt, n)
	}
}

func (n *node) handleBlock(b *pb.Block) {
	plog.Info("node.handleBlock", "height", b.Height, "bp", hexString(b.Signature.Pubkey))
	lb, err := n.RequestLastBlock()
	if err != nil {
		plog.Error("node.handleBlock error:", "err", err.Error(), "height", b.Height)
		return
	}
	if lb.Height >= b.Height {
		bb := lb
		if lb.Height != b.Height {
			bb, err = n.RequestBlock(b.Height)
			if err != nil {
				plog.Error("node.handleBlock error:", "err", err.Error(), "height", b.Height)
				return
			}
		}
		if string(bb.Hash()) == string(b.Hash()) {
			return
		}
		// b is fork block
		// TODO:
		return
	} else if lb.Height+1 < b.Height {
		n.bmp[b.Height] = b
		return
	}

	// TODO: check the block
	n.setBlock(b)
}

func (n *node) checkBlock(pb, b *pb.Block) bool {
	if b.Height == 0 {
		return true
	}

	plog.Info("node.checkBlock", "height", b.Height)

	now := time.Now().UnixNano() / 1000000
	if b.BlockTime > now {
		plog.Info("checkBlock false: blockTime too early", "height", b.Height, "blockTime", b.BlockTime, "now", now)
		return false
	}

	if b.BlockTime-pb.BlockTime < n.Cfg.Pos33BlockTime {
		plog.Info("checkBlock false: blockTime too late", "height", b.Height, "pb.BlockTime", pb.BlockTime, "b.BlockTime", b.BlockTime)
		return false
	}

	// TOTO:
	// 1. check the bp hash if bp is not in my committee, if true, add to committee
	// 2. count bp collect votes and check it
	// 3. if b is right, vote it and broadcast my votes
	// tx0 := b.Txs[0]

	return true
}

var zeroHash [32]byte

func getBlockSeed(b *pb.Block) []byte {
	seed := zeroHash[:]
	if b == nil {
		return seed
	}
	if b.Height > 0 {
		ract, err := getBlockReword(b)
		if err != nil {
			plog.Error("getBlockSeed error", "err", err)
		} else {
			seed = ract.RandHash
		}
	}
	return seed
}

// gen and send my rands message
func (n *node) sortition(b *pb.Block) error {
	// b, err := n.RequestBlock(base)
	// if err != nil {
	// 	plog.Error("RequestBlock error", "height", base)
	// 	return err
	// }
	seed := getBlockSeed(b)

	height := b.Height
	n.comm = newCommittee(height, seed)
	m := genRands(seed, n.allWeight(), n.getWeight(n.pub), n.priv, height)
	if m == nil {
		plog.Error("sortiton nil", "height", b.Height)
		return nil
	}
	plog.Info("node.sortition", "height", height, "weight", len(m.Rands))
	n.handleRands(m)
	n.gss.broadcastUDP(n.marshalRandsMsg(m))
	return nil
}

func (n *node) clear(height int64) {
	for h := range n.bmp {
		if height >= h {
			delete(n.bmp, h)
		}
	}

	for h := range n.vmp {
		if h <= height {
			delete(n.vmp, h)
		}
	}
}

func (n *node) handlePos33Msg(pm *ty.Pos33Msg) bool {
	if pm == nil {
		return false
	}
	switch pm.Ty {
	case ty.Pos33Msg_R:
		var rm ty.Pos33Rands
		err := pb.Decode(pm.Data, &rm)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleRands(&rm)
	case ty.Pos33Msg_B:
		var b pb.Block
		err := pb.Decode(pm.Data, &b)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleBlock(&b)
	case ty.Pos33Msg_V:
		var vt ty.Pos33Vote
		err := pb.Decode(pm.Data, &vt)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleVote(&vt)
	case ty.Pos33Msg_C:
		var pc ty.Pos33Committee
		err := pb.Decode(pm.Data, &pc)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleCommittee(&pc)
	case ty.Pos33Msg_NC:
		var pc ty.Pos33Committee
		err := pb.Decode(pm.Data, &pc)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleNextCommittee(&pc)
	default:
		// n.handleBft(pm)
	}

	return true
}

func (n *node) getCommittee() {
}

func (n *node) gossipCommittee(height int64, c *committee, isNext bool) {
	ty := ty.Pos33Msg_C
	if isNext {
		ty = ty.Pos33Msg_NC
	}

	pc := &ty.Pos33Committee{
		Rs:     c.rs,
		Vw:     int32(c.vw),
		Base:   c.base,
		Seed:   c.seed,
		Height: height,
	}

	sig := n.priv.Sign(pb.Encode(pc))
	pc.Sig = &pb.Signature{Ty: pb.ED25519, Pubkey: []byte(n.pub), Signature: sig.Bytes()}
	pm := &ty.Pos33Msg{
		Data: pb.Encode(pc),
		Ty:   ty,
	}
	data := pb.Encode(pm)
	n.gss.broadcastTCP(data)
}

func (n *node) handleCommittee(pc *ty.Pos33Committee) {
	plog.Info("node.handleCommittee", "height", pc.Height)
	if len(pc.Rs) == 0 {
		return
	}
	// TODO: should check rands
	bb, err := n.RequestBlock(pc.Base)
	if err != nil {
		plog.Error("RequestBlock error", "err", err, "height", pc.Base)
		return
	}
	seed := getBlockSeed(bb)
	if string(seed) != string(pc.Seed) {
		plog.Error("base block error", "height", bb.Height)
		return
	}
	if n.v != nil {
		// if n.v.am(n) {
		// 	return
		// }
		if n.v.vw >= int(pc.Vw) {
			if int64(n.v.vw)+n.v.start > pc.Height {
				return
			}
		}
	}

	c := newCommittee(bb.Height, seed)
	for _, rs := range pc.Rs {
		c.addMember(rs, n, false)
	}
	n.v = newVoter(c, pc.Height)

	if n.comm != nil && n.comm.base == pc.Height {
		return
	}
	b, err := n.RequestBlock(pc.Height)
	if err != nil {
		plog.Error(err.Error())
		return
	}
	n.sortition(b)
	//	n.comm = newCommittee(sb)
	// select {
	// case n.cch <- pc.Height:
	// case <-n.cch:
	// 	n.cch <- pc.Height
	// }
}

func (n *node) handleNextCommittee(pc *ty.Pos33Committee) {
	plog.Info("node.handleNextCommittee", "base", pc.Base)
	// TODO: should check rands
	b, err := n.RequestBlock(pc.Base)
	if err != nil {
		plog.Error("RequestBlock error", "height", pc.Base)
		return
	}
	// seed := getBlockSeed(b)
	if n.comm == nil {
		n.sortition(b)
		if n.comm == nil {
			return
		}
		// n.comm = newCommittee(pc.Base, seed)
		for _, rs := range pc.Rs {
			n.comm.addMember(rs, n, false)
		}
	} else {
		if n.comm.base > pc.Base {
			return
		}
		if n.comm.base < pc.Base {
			if n.comm.vw < int(pc.Vw) {
				n.sortition(b)
				if n.comm == nil {
					return
				}
				// n.comm = newCommittee(pc.Base, seed)
				for _, rs := range pc.Rs {
					n.comm.addMember(rs, n, false)
				}
			}
		}
		// if string(n.comm.seed) != string(pc.Seed) {
		// 	return
		// }
	}
}

// doGossipMsg multi-goroutine verify pos33 message
func (n *node) doGossipMsg() chan *ty.Pos33Msg {
	num := 4
	ch := make(chan *ty.Pos33Msg, num*16)
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

func (n *node) test(hch chan int64) {
	time.AfterFunc(time.Second*5, func() {
		hch := make(chan int64, 1)
		runTest(n.Client, true, int(n.Cfg.Pos33TestMaxAccs), int(n.Cfg.Pos33TestMaxTxs), hch)
	})
}

func (n *node) run() {
	lb, err := n.RequestLastBlock()
	if err != nil {
		plog.Error(err.Error())
		return
	}
	plog.Info("pos33 node runing.......", "last block height=", lb.Height)
	height := lb.Height + 1

	if n.Cfg.Pos33Test {
		hch := make(chan int64, 1)
		go n.test(hch)
		go func() {
			for {
				hch <- height
				time.Sleep(time.Millisecond * 1000)
			}
		}()
	}

	n.gss = newGossip(n.priv.PubKey().KeyString(), n.Cfg.Pos33ListenAddr, n.Cfg.Pos33AdvertiseAddr, n.Cfg.Pos33PeerSeed)
	go n.gss.runBroadcast()
	msgch := n.doGossipMsg()

	blockTime := n.Cfg.Pos33BlockTime
	if blockTime == 0 {
		blockTime = 1000
		n.Cfg.Pos33BlockTime = blockTime
	}
	blockTimeout := n.Cfg.Pos33BlockTimeout
	if blockTimeout == 0 {
		blockTimeout = 3300
		n.Cfg.Pos33BlockTimeout = blockTimeout
	}
	sortitionTimeout := 3300

	if n.Cfg.Pos33Test {
		n.sortition(lb)
		n.v = newVoter(n.comm, lb.Height)
		if lb.Height > 0 {
			go func() { n.bch <- lb }()
		}
	}

	lastHeight := int64(0)
	timer := time.NewTimer(time.Hour)
	tmRest := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(time.Millisecond * time.Duration(blockTimeout))
	}

	rch := make(chan int64, 1)
	errTime := 0
	for {
		select {
		case msg := <-msgch:
			n.handlePos33Msg(msg)
		case <-timer.C:
			h := lastHeight
			plog.Error("make block TIMEOUT", "height", h)
			if n.v != nil {
				if n.v.checkHeight(h+1) == 0 {
					n.v.voteError(h+1, n)
				}
			}
			tmRest()
			errTime++
			if errTime == 3 {
				plog.Info("3 TIMEOUT !!!")
				if n.comm != nil && n.comm.vw >= 5 {
					n.changeCommittee(n.comm, lb.Height)
				}
				n.sortition(lb)
				n.newRound(h)
				errTime = 0
			}
		case h := <-rch:
			n.newRound(h)
			lastHeight = h
			tmRest()
			errTime = 0
		case b := <-n.bch:
			lb = b
			plog.Info("node.handleAddBlock", "height", b.Height)
			if n.v != nil {
				if b.Height-n.v.start >= int64(n.v.vw) {
					plog.Info("change commottee", "height", b.Height)
					if n.comm != nil && n.comm.vw >= 5 {
						n.changeCommittee(n.comm, b.Height)
					}
					n.sortition(b)
				}
				r := n.v.checkVh(b.Height)
				if r == 0 {
					n.v.vote(b, n)
				} else if r > 0 {
					time.AfterFunc(time.Millisecond*time.Duration(sortitionTimeout), func() { n.addBlock(b) })
					break
				} else {
					break
				}

				if b.Height-n.v.start == 1 {
					n.gossipCommittee(n.v.start, n.v.committee, false)
				}
				if b.Height-n.v.start == 4 {
					if n.comm != nil && n.comm.vw >= 5 {
						n.gossipCommittee(b.Height, n.comm, true)
					} else {
						n.sortition(b)
					}
				}
			}

			if n.handleAddBlock(b) {
				t := blockTime + b.BlockTime - time.Now().UnixNano()/1000000
				if t < 0 {
					t = 0
				}
				time.AfterFunc(time.Millisecond*time.Duration(t), func() { rch <- b.Height })
			}
		}
	}
}

func (n *node) marshalBlockMsg(b *pb.Block) []byte {
	pm := &ty.Pos33Msg{
		Data: pb.Encode(b),
		Ty:   ty.Pos33Msg_B,
	}
	return pb.Encode(pm)
}

func (n *node) marshalVoteMsg(v *ty.Pos33Vote) []byte {
	pm := &ty.Pos33Msg{
		Data: pb.Encode(v),
		Ty:   ty.Pos33Msg_V,
	}
	return pb.Encode(pm)
}

func (n *node) marshalRandsMsg(m *ty.Pos33Rands) []byte {
	pm := &ty.Pos33Msg{
		Data: pb.Encode(m),
		Ty:   ty.Pos33Msg_R,
	}
	return pb.Encode(pm)
}

func hexString(b []byte) string {
	if len(b) == 0 {
		return "nil"
	}
	if string(b) == errorVote {
		return errorVote
	}
	return hex.EncodeToString(b)[:16]
}
