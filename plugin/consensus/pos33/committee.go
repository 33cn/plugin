package pos33

/*
import pb "github.com/33cn/chain33/types"

type voter struct {
	*committee
	start   int64
	indexs  map[int64]*pt.Pos33Rand // 每个成员的索引
	vs      map[int64]map[string]*pt.Pos33Vote
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
		indexs:    make(map[int64]*pt.Pos33Rand),
		vs:        make(map[int64]map[string]*pt.Pos33Vote),
		voteOk:    make(map[int64]bool),
		notVote:   -1,
	}
	v.sort()
	plog.Info("@@@@@@@ newVoter", "len(indexs)", len(v.indexs), "start", v.start, "vw", v.vw)
	return v
}

// committee is 共识委员会
type committee struct {
	base int64 // base block for sortition
	seed []byte
	vw   int // 总的投票权重 == 区块数量
	rs   []*pt.Pos33Rand
}

func newCommittee(height int64, seed []byte) *committee {
	plog.Info("newCommittee", "base", height)
	return &committee{
		seed: seed,
		base: height,
	}
}

func (v *voter) sort() {
	for i, r := range v.rs {
		v.indexs[v.start+int64(i)+1] = r
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
	vt := &pt.Pos33Vote{
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
	vt := &pt.Pos33Vote{
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

func (v *voter) checkVote(vt *pt.Pos33Vote, n *node) bool {
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

	return true
}

func (v *voter) handleVote(vt *pt.Pos33Vote, n *node) {
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
		mp = make(map[string]*pt.Pos33Vote)
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

func (c *committee) addMember(rs *pt.Pos33Rands, n *node, fromSort bool) {
	plog.Info("addMember", "pub", hexString([]byte(rs.Pub)), "base", c.base)
	if rs.Height != c.base {
		plog.Error("addMember", "height", rs.Height)
		return
	}

	if !n.checkRands(rs, c.seed, c.base) {
		plog.Error("addMember: checkRands error", "pub", hexString([]byte(rs.Pub)))
		return
	}

	if rrs, ok := c.rs[string(rs.Pub)]; ok {
		c.vw -= len(rrs.Rands)
	}

	c.rs[string(rs.Pub)] = rs
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

*/
