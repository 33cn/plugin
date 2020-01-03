package pos33

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
	"github.com/golang/protobuf/proto"
)

var plog = log15.New("module", "pos33")

type node struct {
	*Client
	gss *gossip

	// I'm candidate proposer in these blocks
	ips map[int64]map[int]*pt.Pos33SortMsg
	// I'm candidate verifer in these blocks
	ivs map[int64]map[int][]*pt.Pos33SortMsg
	// receive candidate proposers
	cps map[int64]map[int]map[string]*pt.Pos33SortMsg
	// receive candidate verifers
	cvs map[int64]map[int]map[string][]*pt.Pos33VoteMsg
	css map[int64]map[int][]*pt.Pos33SortMsg
	// receive candidate blocks
	cbs    map[int64]map[int]map[string]*types.Block
	bsLock sync.Mutex

	bch      chan *types.Block
	round    int
	lb       *types.Block
	unbMap   map[int64][]*pt.Pos33BlockMsg // unchecked block map
	votedMap map[int64]map[int]bool        // already voted map
	nevMap   map[int64]map[int]string      // not enough vote map
}

// New create pos33 consensus client
func newNode(conf *subConfig) *node {
	n := &node{
		ips:      make(map[int64]map[int]*pt.Pos33SortMsg),
		ivs:      make(map[int64]map[int][]*pt.Pos33SortMsg),
		cps:      make(map[int64]map[int]map[string]*pt.Pos33SortMsg),
		cvs:      make(map[int64]map[int]map[string][]*pt.Pos33VoteMsg),
		css:      make(map[int64]map[int][]*pt.Pos33SortMsg),
		cbs:      make(map[int64]map[int]map[string]*types.Block),
		bch:      make(chan *types.Block, 1),
		unbMap:   make(map[int64][]*pt.Pos33BlockMsg),
		votedMap: make(map[int64]map[int]bool),
		nevMap:   make(map[int64]map[int]string),
	}

	plog.Info("@@@@@@@ node start:", "addr", addr, "conf", conf)
	return n
}

func (n *node) setLastBlock(height int64) {
	b := n.GetCurrentBlock()
	if b.Height == height {
		n.lb = b
		return
	}
	b, err := n.RequestBlock(height)
	if err != nil {
		panic(err)
	}
	n.lb = b
}

func (n *node) lastBlock() *types.Block {
	if n.lb != nil {
		return n.lb
	}
	return n.GetCurrentBlock()
}

func unmarshal(b []byte) (*pt.Pos33Msg, error) {
	var pm pt.Pos33Msg
	err := proto.Unmarshal(b, &pm)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (n *node) voteLast() {
	act, err := getMiner(n.lastBlock())
	if err != nil {
		return
	}
	height := n.lastBlock().Height
	round := int(act.Sort.Proof.Input.Round)
	for !n.reSortition(height, round) {
		time.Sleep(time.Millisecond * 300)
	}
	tid := act.Sort.SortHash.Tid
	n.vote(height, round, tid)
}

var errNotEnoughVotes = errors.New("NOT enough votes for the last block")

func (n *node) getBlockVotes(b *types.Block) ([]*pt.Pos33VoteMsg, error) {
	height := b.Height
	if height == 0 {
		return nil, nil
	}

	lastAct, err := getMiner(b)
	if err != nil {
		return nil, err
	}

	lround := int(lastAct.GetSort().Proof.Input.Round)
	ltid := lastAct.GetSort().SortHash.Tid
	mp, ok := n.nevMap[height]
	if !ok {
		mp = make(map[int]string, 1)
	}
	n.nevMap[height] = mp
	if n.cvs[height] == nil || n.cvs[height][lround] == nil {
		mp[n.round] = ltid
		return nil, errNotEnoughVotes
	}
	vs := n.cvs[height][lround][ltid]
	if len(vs) < pt.Pos33MustVotes {
		mp[n.round] = ltid
		return nil, errNotEnoughVotes
	}
	if len(vs) > pt.Pos33VoterSize {
		sort.Sort(pt.Votes(vs))
		vs = vs[:pt.Pos33VoterSize]
	}
	return vs, nil
}

func (n *node) minerTx(sm *pt.Pos33SortMsg, vs []*pt.Pos33VoteMsg, priv crypto.PrivKey) (*types.Transaction, error) {
	plog.Info("genRewordTx", "vsw", len(vs))
	act := &pt.Pos33TicketAction{
		Value: &pt.Pos33TicketAction_Pminer{
			Pminer: &pt.Pos33Miner{
				Votes: vs,
				Sort:  sm,
			},
		},
		Ty: pt.Pos33TicketActionMiner,
	}

	cfg := n.GetAPI().GetConfig()
	tx, err := types.CreateFormatTx(cfg, "pos33", types.Encode(act))
	if err != nil {
		return nil, err
	}

	tx.Sign(types.SECP256K1, priv)
	return tx, nil
}

func (n *node) blockDiff(lb *types.Block, w int) uint32 {
	powLimitBits := n.GetAPI().GetConfig().GetP(lb.Height).PowLimitBits
	return powLimitBits
	// oldTarget := difficulty.CompactToBig(lb.Difficulty)
	// newTarget := new(big.Int).Mul(oldTarget, big.NewInt(11)) // pt.Pos33MustVotes))
	// newTarget.Div(newTarget, big.NewInt(int64(w+1)))

	// powLimit := difficulty.CompactToBig(powLimitBits)
	// if newTarget.Cmp(powLimit) > 0 {
	// 	newTarget.Set(powLimit)
	// }
	// return difficulty.BigToCompact(newTarget)
}

func (n *node) myVotes(height int64, round int) []*pt.Pos33SortMsg {
	mp, ok := n.ivs[height]
	if !ok {
		return nil
	}

	vs, ok := mp[round]
	if !ok {
		return nil
	}

	return vs
}

func (n *node) mySort(height int64, round int) *pt.Pos33SortMsg {
	mp, ok := n.ips[height]
	if !ok {
		return nil
	}

	sort, ok := mp[round]
	if !ok {
		return nil
	}

	return sort
}

func (n *node) getPrivByTid(tid string) (crypto.PrivKey, error) {
	t := n.getTicket(tid)
	if t == nil {
		return nil, fmt.Errorf("getTicket error: %s", tid)
	}
	priv := n.getPriv(t.MinerAddress)
	if priv == nil {
		return nil, fmt.Errorf("getPriv error: %s", t.MinerAddress)
	}
	return priv, nil
}

func (n *node) makeBlock(height int64, round int) (*pt.Pos33BlockMsg, error) {
	lb := n.lastBlock()
	if height != lb.Height+1 {
		return nil, fmt.Errorf("lastHeight error")
	}
	vs, err := n.getBlockVotes(lb)
	if err != nil {
		return nil, err
	}
	sort := n.mySort(height, round)
	if sort == nil {
		err = fmt.Errorf("I'm not bp")
		return nil, err
	}
	if n.getCacheBlock(height, round, sort.SortHash.Tid) != nil {
		return nil, fmt.Errorf("I have maked height=%d round=%d block", height, round)
	}

	priv, err := n.getPrivByTid(sort.SortHash.Tid)
	if err != nil {
		return nil, err
	}

	tx, err := n.minerTx(sort, vs, priv)
	if err != nil {
		return nil, err
	}

	nb, err := n.newBlock(lb, []*Tx{tx}, height)
	if err != nil {
		return nil, err
	}

	nb.Difficulty = n.blockDiff(lb, len(vs))
	plog.Info("@@@@@@@ I make a block: ", "height", height, "round", round, "ntx", len(nb.Txs), "nvs", len(vs), "diff", nb.Difficulty)
	bm := &pt.Pos33BlockMsg{B: nb}
	bm.Sign(priv)
	return bm, nil
}

func (n *node) addBlock(b *types.Block) {
	if !n.miningOK() {
		return
	}

	if b.Height < 0 && n.lastBlock() != nil && b.Height <= n.lastBlock().Height {
		plog.Info("addBlock nil", "height", b.Height)
		return
	}

	fn := func(nb *types.Block) {
		select {
		case n.bch <- nb:
		case <-n.bch:
			n.bch <- nb
		}
	}

	t := b.BlockTime - n.lastBlock().BlockTime
	plog.Info("node.addBlock", "height", b.Height, "delta", t, "hash", common.ToHex(b.Hash(n.GetAPI().GetConfig())))
	if t < 1 {
		time.AfterFunc(time.Millisecond*300, func() {
			fn(b)
		})
	} else {
		fn(b)
	}
}

func (n *node) clear(height int64) {
	// clear the caches
	n.bsLock.Lock()
	for h := range n.cbs {
		if h+10 <= height {
			delete(n.cbs, h)
		}
	}
	n.bsLock.Unlock()
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
	for h := range n.unbMap {
		if h+10 <= height {
			delete(n.unbMap, h)
		}
	}
	for h := range n.nevMap {
		if h+10 <= height {
			delete(n.nevMap, h)
		}
	}
	for h := range n.votedMap {
		if h+10 <= height {
			delete(n.votedMap, h)
		}
	}
}

func addr(sig *types.Signature) string {
	if sig == nil {
		return ""
	}
	return address.PubKeyToAddress(sig.Pubkey).String()
}

func sumR(r int) int {
	sum := 0
	for i := 1; i <= r; i++ {
		sum += i + 1
	}
	return sum
}

func (n *node) checkBlock(b, pb *types.Block) error {
	if b.Height <= pb.Height {
		return fmt.Errorf("")
	}
	if b.Height < 2 {
		return nil
	}
	if !n.IsCaughtUp() {
		return nil
	}
	if len(b.Txs) == 0 {
		return fmt.Errorf("nil block error")
	}

	plog.Info("node.checkBlock", "height", b.Height, "pbheight", pb.Height)
	act, err := getMiner(b)
	if err != nil {
		return err
	}
	if act.Sort == nil || act.Sort.Proof == nil || act.Sort.Proof.Input == nil {
		return fmt.Errorf("miner tx error")
	}
	round := int(act.Sort.Proof.Input.Round)
	if n.getCacheBlock(b.Height, round, act.Sort.SortHash.Tid) != nil {
		// already checked
		return nil
	}

	seed, err := n.getMinerSeed(b.Height)
	if err != nil {
		return err
	}
	err = n.verifySort(b.Height, 0, n.allw(b.Height, true), seed, act.GetSort())
	if err != nil {
		return err
	}

	// check votes
	if len(act.Votes) < pt.Pos33MustVotes {
		return fmt.Errorf("the block NOT enouph vote, height=%d", b.Height)
	}
	if len(act.Votes) > pt.Pos33VoterSize {
		return fmt.Errorf("the block vote too much, height=%d", b.Height)
	}
	// vote is last height vote
	lact, err := getMiner(pb)
	if err != nil {
		return err
	}
	round = int(lact.Sort.Proof.Input.Round)
	height := pb.Height
	tid := lact.Sort.SortHash.Tid
	allw := n.allw(height, true)
	seed, err = n.getMinerSeed(height)
	if err != nil {
		return err
	}
	for _, v := range act.Votes {
		if n.cvs[height] != nil && n.cvs[height][round] != nil {
			for _, mv := range n.cvs[height][round][tid] {
				if mv.Sort.SortHash.Tid == v.Sort.SortHash.Tid {
					continue // already checked
				}
			}
		}

		err = n.checkVote(v, height, round, seed, allw, tid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *node) getMinerSort(height int64) (*pt.Pos33SortMsg, error) {
	/*
		startHeight := height - height%pt.Pos33SortitionSize
		if startHeight == height {
			startHeight -= pt.Pos33SortitionSize
		}
	*/
	startHeight := height - pt.Pos33SortitionSize
	b, err := n.RequestBlock(startHeight)
	if err != nil {
		plog.Info("should't go here. do nothing")
		return nil, err
	}
	if b.Height == 0 {
		return nil, nil
	}
	act, err := getMiner(b)
	if err != nil {
		plog.Error("getBlockMiner err", "error", err, "height", b.Height)
		return nil, err
	}
	return act.Sort, nil
}

func (n *node) getMinerSeed(height int64) ([]byte, error) {
	seed := zeroHash[:]
	if height > pt.Pos33SortitionSize {
		sort, err := n.getMinerSort(height)
		if err != nil {
			return nil, err
		}
		if sort != nil {
			seed = sort.SortHash.Hash
		}
	}
	return seed, nil
}

var zeroHash [32]byte

func (n *node) reSortition(height int64, round int) bool {
	seed, err := n.getMinerSeed(height)
	if err != nil {
		plog.Error("reSortition error", "height", height, "round", round, "err", err)
		return false
	}
	const staps = 2
	allw := n.allw(height, false)
	for s := 0; s < staps; s++ {
		sms := n.sort(seed, height, round, s, allw)
		if sms == nil {
			plog.Info("node.sortition nil", "height", height, "round", round)
			continue
		}
		plog.Info("node.reSortition", "height", height, "round", round, "step", s, "weight", len(sms))
		if s == 0 {
			mp := n.ips[height]
			if mp == nil {
				n.ips[height] = make(map[int]*pt.Pos33SortMsg)
			}
			n.ips[height][round] = sms[0]
		} else {
			mp := n.ivs[height]
			if mp == nil {
				n.ivs[height] = make(map[int][]*pt.Pos33SortMsg)
			}
			n.ivs[height][round] = sms
		}
	}
	n.sendVoteSorts(height, round)
	return len(n.ivs[height][round]) >= pt.Pos33MustVotes
}

func (n *node) sortition(b *types.Block, round int) {
	plog.Info("sortition", "height", b.Height, "round", round)

	startHeight := b.Height
	seed := zeroHash[:]
	const steps = 2
	loop := 1
	allw := n.allw(startHeight, false)
	if b.Height == 0 {
		loop = pt.Pos33SortitionSize
		startHeight++
	} else {
		startHeight += pt.Pos33SortitionSize
		act, err := getMiner(b)
		if err != nil {
			panic(err)
		}
		seed = act.Sort.SortHash.Hash
	}
	for s := 0; s < steps; s++ {
		for i := 0; i < loop; i++ {
			height := startHeight + int64(i)
			sms := n.sort(seed, height, round, s, allw)
			if sms == nil {
				plog.Info("node.sortition nil", "height", height, "round", round)
				continue
			}
			plog.Info("node.sortition", "height", height, "round", round, "weight", len(sms))
			if s == 0 {
				mp := n.ips[height]
				if mp == nil {
					n.ips[height] = make(map[int]*pt.Pos33SortMsg)
				}
				n.ips[height][round] = sms[0]
			} else {
				mp := n.ivs[height]
				if mp == nil {
					n.ivs[height] = make(map[int][]*pt.Pos33SortMsg)
				}
				n.ivs[height][round] = sms
				n.sendVoteSorts(height, 0)
			}
		}
	}
}

func (n *node) checkVote(vm *pt.Pos33VoteMsg, height int64, round int, seed []byte, allw int, tid string) error {
	if !vm.Verify() {
		return fmt.Errorf("votemsg verify false")
	}
	if vm.Sort == nil || vm.Sort.Proof == nil || vm.Sort.Proof.Input == nil || vm.Sort.SortHash == nil {
		return fmt.Errorf("votemsg error, vm.Sort==nil or vm.Sort.Input==nil")
	}
	if height != vm.Sort.Proof.Input.Height {
		return fmt.Errorf("vote height is NOT consistent")
	}
	if round != int(vm.Sort.Proof.Input.Round) {
		return fmt.Errorf("vote round is NOT consistent: %d != %d", round, vm.Sort.Proof.Input.Round)
	}
	if tid != vm.Tid {
		return fmt.Errorf("vote Tid is NOT consistent: %s != %s", tid, vm.Tid)
	}
	if string(vm.Sig.Pubkey) != string(vm.Sort.Proof.Pubkey) {
		return fmt.Errorf("vote pubkey is NOT consistent")
	}

	err := n.verifySort(height, 1, allw, seed, vm.Sort)
	if err != nil {
		return err
	}
	return nil
}

func (n *node) addVote(vm *pt.Pos33VoteMsg, height int64, round int, tid string) {
	if n.cvs[height] == nil {
		n.cvs[height] = make(map[int]map[string][]*pt.Pos33VoteMsg)
	}

	if n.cvs[height][round] == nil {
		mp := make(map[string][]*pt.Pos33VoteMsg)
		n.cvs[height][round] = mp
	}

	vs := n.cvs[height][round][tid]
	for i, v := range vs {
		if v.Equal(vm) {
			// delete previous
			vs[i] = vs[len(vs)-1]
			vs = vs[:len(vs)-1]
			break
		}
	}
	vs = append(vs, vm)
	n.cvs[height][round][tid] = vs
}

func (n *node) handleVotesMsg(vms *pt.Pos33Votes, myself bool) {
	if n.lastBlock() == nil {
		return
	}
	if len(vms.Vs) == 0 {
		plog.Error("votemsg sortition is 0")
		return
	}

	vm := vms.Vs[0]
	if vm.Sort == nil || vm.Sort.Proof == nil || vm.Sort.Proof.Input == nil {
		return
	}
	height := vm.Sort.Proof.Input.Height
	round := int(vm.Sort.Proof.Input.Round)
	tid := vm.Tid

	seed, err := n.getMinerSeed(height)
	if err != nil {
		plog.Error("getMinerSeed error", "err", err, "height", height)
		return
	}
	allw := n.allw(height, true)

	for _, vm := range vms.Vs {
		if !myself {
			err := n.checkVote(vm, height, round, seed, allw, tid)
			if err != nil {
				plog.Error("check error", "height", height, "round", round, "err", err)
				if err == errDiff {
					continue
				}
				return
			}
		}
		n.addVote(vm, height, round, tid)
	}
	plog.Info("handleVotesMsg", "height", height, "round", round, "tid", tid[:16], "voter", addr(vm.GetSig()), "votes", len(n.cvs[height][round][tid]))
	if n.checkVotesEnough(n.cvs[height][round][tid], height, round) {
		if height <= n.lastBlock().Height && n.cvs[height] != nil {
			plog.Info("handleVoteMsg vote too late", "height", height, "round", round)
			mp, ok := n.nevMap[height]
			if !ok {
				return
			}
			if mp[round] == tid {
				n.elect(height+1, n.round)
			}
			return
		}

		plog.Info("@@@ set block 2f+1 @@@", "height", height, "tid", tid[:16])
		if b := n.getCacheBlock(height, round, tid); b != nil {
			n.setBlock(b)
		}
	}
}

func (n *node) checkVotesEnough(vs []*pt.Pos33VoteMsg, height int64, round int) bool {
	mp, ok := n.css[height]
	if !ok {
		return false
	}
	ss, ok := mp[round]
	if !ok {
		return false
	}
	if len(ss) < pt.Pos33MustVotes {
		return false
	}
	ss = ss[:pt.Pos33MustVotes]
	count := 0
	for _, v := range vs {
		found := false
		for _, s := range ss {
			if string(v.Sort.SortHash.Hash) == string(s.SortHash.Hash) {
				found = true
				break
			}
		}
		if found {
			count++
		}
	}
	return count >= pt.Pos33MustVotes
}

func (n *node) getCacheBlock(height int64, round int, tid string) *types.Block {
	n.bsLock.Lock()
	defer n.bsLock.Unlock()

	if n.cbs[height] != nil && n.cbs[height][round] != nil {
		b, ok := n.cbs[height][round][tid]
		if ok {
			return b
		}
	}
	return nil
}

func (n *node) allw(height int64, check bool) int {
	if height > 10 {
		if check {
			height -= pt.Pos33SortitionSize
		}
		return n.allWeight(height)
	}
	return n.allWeight(1)
}

func (n *node) handleSortitionMsg(m *pt.Pos33SortMsg) {
	if m == nil || m.Proof == nil || m.Proof.Input == nil || m.SortHash == nil {
		plog.Error("handleSortitionMsg error, input msg is nil")
		return
	}
	height := m.Proof.Input.Height
	if n.cps[height] == nil {
		n.cps[height] = make(map[int]map[string]*pt.Pos33SortMsg)
	}
	round := int(m.Proof.Input.Round)
	if n.cps[height][round] == nil {
		n.cps[height][round] = make(map[string]*pt.Pos33SortMsg)
	}
	oldBp := n.bp(height, round)
	n.cps[height][round][m.SortHash.Tid] = m
	bp := n.bp(height, round)
	plog.Info("handleSortitionMsg", "height", height, "round", round, "sorter", bp[:16], "size", len(n.cps[height][round]))

	// if recv last round block, and is right, then revote?
	if oldBp != bp {
		mp, ok := n.votedMap[height]
		if !ok {
			return
		}
		voted, ok := mp[round]
		if !ok {
			return
		}
		if voted {
			plog.Info("!!! revote !!!", "height", height, "round", round, "newBp", bp[:16])
			n.vote(height, round, bp)
		}
	}
}

func (n *node) checkSort(s *pt.Pos33SortMsg) error {
	if s == nil {
		return fmt.Errorf("sortMsg error")
	}
	if s.Proof == nil || s.Proof.Input == nil || s.SortHash == nil {
		return fmt.Errorf("sortMsg error")
	}

	height := s.Proof.Input.Height
	seed, err := n.getMinerSeed(height)
	if err != nil {
		return err
	}
	return n.verifySort(height, int(s.Proof.Input.Step), n.allw(height, true), seed, s)
}

func (n *node) handleSortsMsg(m *pt.Pos33Sorts, myself bool) {
	if len(m.Sorts) == 0 {
		return
	}
	plog.Info("handleSortsMsg", "height", m.Sorts[0].Proof.Input.Height)
	for _, s := range m.Sorts {
		if !myself {
			err := n.checkSort(s)
			if err != nil {
				plog.Error("checkSort error", "err", err)
				return
			}
		}
		height := s.Proof.Input.Height
		round := int(s.Proof.Input.Round)
		if n.css[height] == nil {
			n.css[height] = make(map[int][]*pt.Pos33SortMsg)
		}
		ss := n.css[height][round]
		ss = append(ss, s)
		sort.Sort(pt.Sorts(ss))
		n.css[height][round] = ss
	}
}

func (n *node) handleBlockMsg(bm *pt.Pos33BlockMsg, myself bool) {
	b := bm.B
	if !myself {
		if b == nil {
			plog.Error("handleBlockMsg error: block is nil")
			return
		}
		lb := n.lastBlock()
		if lb.Height >= b.Height {
			plog.Info("handleBlockMsg: block too late", "height", b.Height)
			return
		}
		if b.Height > lb.Height+1 {
			n.unbMap[b.Height] = append(n.unbMap[b.Height], bm)
			plog.Info("handleBlockMsg block too high", "height", b.Height, "preHeight", lb.Height)
			return
		}
		if !bm.Verify() {
			plog.Error("handleBlockMsg error, sig error")
			return
		}
		plog.Info("handleBlockMsg", "height", b.Height, "maker", addr(bm.Sig))
		err := n.checkBlock(b, lb)
		if err != nil {
			plog.Error("handleBlockMsg err: checkBlock error", "err", err, "height", b.Height)
			return
		}
	}

	act, err := getMiner(b)
	if err != nil {
		plog.Error("handleBlockMsg error", "err", err)
		return
	}
	n.addBlockCache(b, act)
	n.handleSortitionMsg(act.GetSort())
}

func (n *node) addBlockCache(b *types.Block, act *pt.Pos33Miner) {
	n.bsLock.Lock()
	defer n.bsLock.Unlock()

	mp, ok := n.cbs[b.Height]
	if !ok {
		mp = make(map[int]map[string]*types.Block)
	}

	round := int(act.Sort.Proof.Input.Round)
	bmp, ok := n.cbs[b.Height][round]
	if !ok {
		bmp = make(map[string]*types.Block)
	}

	tid := act.Sort.SortHash.Tid
	plog.Info("handleBlockMsg", "height", b.Height, "round", round, "tid", tid)
	bmp[tid] = b
	mp[round] = bmp
	n.cbs[b.Height] = mp
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
		n.handleBlockMsg(&m, false)
	case pt.Pos33Msg_S:
		var m pt.Pos33Sorts
		err := types.Decode(pm.Data, &m)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleSortsMsg(&m, false)
	case pt.Pos33Msg_V:
		var vt pt.Pos33Votes
		err := types.Decode(pm.Data, &vt)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleVotesMsg(&vt, false)
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
	plog.Info("firstSortition")
	n.sortition(firtstBlock, 0)
}

func (n *node) checkUnB(lb *types.Block) {
	for _, bm := range n.unbMap[lb.Height+1] {
		if bm != nil {
			n.handleBlockMsg(bm, false)
		}
	}
	n.unbMap[lb.Height] = nil
}

func (n *node) runLoop() {
	lb, err := n.RequestLastBlock()
	if err != nil {
		plog.Error(err.Error())
		return
	}

	n.gss = newGossip(n.nodeID(), n.conf.ListenAddr, n.conf.AdvertiseAddr, n.conf.BootPeerAddr)
	go n.gss.runBroadcast()
	msgch := n.doGossipMsg()

	etm := time.NewTimer(time.Hour)
	ch := make(chan int64, 1)

	time.AfterFunc(time.Second, func() {
		if lb.Height > 0 && n.conf.SingleNode {
			lastAct, err := getMiner(lb)
			if err != nil {
				panic(err)
			}
			lround := int(lastAct.Sort.Proof.Input.Round)
			if n.cvs[lb.Height] == nil || n.cvs[lb.Height][lround] == nil {
				n.voteLast()
			}
		}
		n.addBlock(lb)
	})

	blockTimeout := time.Millisecond * time.Duration(n.conf.BlockTimeout)
	if blockTimeout < time.Millisecond*500 || blockTimeout > time.Millisecond*30000 {
		blockTimeout = time.Millisecond * 1000
	}
	baseST := blockTimeout * 2 / 3
	deltaST := blockTimeout / 3
	var rs []int
	plog.Info("@@@@@@@@ pos33 node runing.......", "last block height", lb.Height, "baseST", baseST)

	for {
		if !n.miningOK() {
			time.Sleep(time.Second)
			continue
		}
		select {
		case msg := <-msgch:
			n.handlePos33Msg(msg)
		case height := <-ch:
			if height == n.lastBlock().Height+1 {
				n.round++
				plog.Info("@@@ vote timeout: ", "height", height, "round", n.round)
				n.reSortition(height, n.round)
				n.elect(height, n.round)
				etm = time.NewTimer(baseST + deltaST*time.Duration(n.round))
			}
		case <-etm.C:
			height := n.lastBlock().Height + 1
			plog.Info("elect timeout: ", "height", height)
			tid := n.bp(height, n.round)
			n.vote(height, n.round, tid)
			// du := baseST + deltaST*time.Duration(n.round+1)
			time.AfterFunc(time.Second*3, func() {
				ch <- height
			})
		case b := <-n.bch: // new block add to chain
			n.setLastBlock(b.Height)
			n.checkUnB(b)
			if b.Height%pt.Pos33SortitionSize == 0 {
				n.flushTicket()
			}
			n.sortition(b, 0)
			baseST, rs = changeBaseST(rs, b, baseST, deltaST, blockTimeout)
			n.round = 0
			n.elect(b.Height+1, 0)
			n.clear(b.Height - 1)
			plog.Info("elect timer", "t", baseST)
			etm = time.NewTimer(baseST)
		}
	}
}

func (n *node) sendVoteSorts(height int64, round int) {
	mp, ok := n.ivs[height]
	if !ok {
		return
	}
	ss, ok := mp[round]
	if !ok {
		return
	}
	if len(ss) == 0 {
		return
	}

	m := &pt.Pos33Sorts{Sorts: ss}
	b := marshalSortsMsg(m)
	n.gss.gossip(b)
	n.handleSortsMsg(m, true)
}

func changeBaseST(rs []int, b *types.Block, baseST, deltaST, bt time.Duration) (time.Duration, []int) {
	rs = append(rs, getBlockRound(b))
	if len(rs) > pt.Pos33SortitionSize {
		rs = rs[1:]
	}

	if b.Height%pt.Pos33SortitionSize != 0 {
		return baseST, rs
	}

	if avgRound(rs) > 1 {
		baseST += deltaST
	} else {
		baseST -= deltaST
		if baseST < bt*2/3 {
			baseST = bt * 2 / 3
		}
	}
	return baseST, rs
}

func avgRound(rs []int) int {
	if len(rs) < pt.Pos33SortitionSize {
		return 0
	}
	sum := 0
	for _, r := range rs {
		sum += r
	}
	return sum / len(rs)
}

func hexs(b []byte) string {
	s := hex.EncodeToString(b)
	if len(s) <= 16 {
		return s
	}
	return s[:16]
}

func (n *node) bp(height int64, round int) string {
	var pss []*pt.Pos33SortMsg
	for _, s := range n.cps[height][round] {
		pss = append(pss, s)
	}
	if len(pss) == 0 {
		return ""
	}

	sort.Sort(pt.Sorts(pss))

	return pss[0].SortHash.Tid
}

func (n *node) vote(height int64, round int, tid string) {
	if tid == "" {
		return
	}
	ss := n.myVotes(height, round)
	if ss == nil {
		plog.Info("I'm not verifer", "height", height)
		return
	}
	if len(ss) > pt.Pos33VoterSize {
		ss = ss[:pt.Pos33VoterSize]
	}
	plog.Info("vote bp", "height", height, "round", round, "tid", tid[:16])
	var vs []*pt.Pos33VoteMsg
	for _, s := range ss {
		v := &pt.Pos33VoteMsg{Sort: s, Tid: tid}
		t := n.getTicket(s.SortHash.Tid)
		if t == nil {
			plog.Info("vote error: my ticket is gone", "ticketID", s.SortHash.Tid)
			continue
		}
		priv := n.getPriv(t.MinerAddress)
		if priv == nil {
			plog.Info("vote error: my miner address is gone", "mineaddr", t.MinerAddress)
			continue
		}
		v.Sign(priv)
		vs = append(vs, v)
	}
	v := &pt.Pos33Votes{Vs: vs}
	n.gss.gossip(marshalVoteMsg(v))
	n.handleVotesMsg(v, true)
	mp, ok := n.votedMap[height]
	if !ok {
		mp = make(map[int]bool)
	}
	mp[round] = true
	n.votedMap[height] = mp
}

func getBlockRound(b *types.Block) int {
	if b.Height == 0 {
		return 0
	}
	act, err := getMiner(b)
	if err != nil {
		plog.Error("getBlockRound error", "err", err, "height", b.Height)
		return 0
	}
	return int(act.Sort.Proof.Input.Round)
}

func (n *node) elect(height int64, round int) error {
	plog.Info("@@@ elect @@@", "height", height, "round", round)
	bm, err := n.makeBlock(height, round)
	if err != nil {
		plog.Error("make block error", "err", err, "height", height, "round", round)
		//return err
		if err != errNotEnoughVotes || round <= 3 {
			return err
		}
		plog.Info("aha, the last block error, we remake it")
		lb := n.lastBlock()
		lr := getBlockRound(lb)
		n.setLastBlock(height - 2)
		bm, err = n.makeBlock(height-1, lr+1)
		if err != nil {
			plog.Error("make block error", "err", err, "height", height, "round", round)
			return err
		}
	}
	n.gss.gossip(marshalBlockMsg(bm))
	n.handleBlockMsg(bm, true)
	return nil
}

func marshalSortsMsg(m *pt.Pos33Sorts) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(m),
		Ty:   pt.Pos33Msg_S,
	}
	return types.Encode(pm)
}

func marshalBlockMsg(m *pt.Pos33BlockMsg) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(m),
		Ty:   pt.Pos33Msg_B,
	}
	return types.Encode(pm)
}

/*
func marshalSortMsg(m *pt.Pos33SortMsg) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(m),
		Ty:   pt.Pos33Msg_S,
	}
	return types.Encode(pm)
}
*/

func marshalVoteMsg(v *pt.Pos33Votes) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(v),
		Ty:   pt.Pos33Msg_V,
	}
	return types.Encode(pm)
}
