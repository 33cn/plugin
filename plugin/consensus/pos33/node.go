package pos33

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
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
	// receive candidate votes
	cvs map[int64]map[int]map[string][]*pt.Pos33VoteMsg
	// receive candidate verifers
	css map[int64]map[int][]*pt.Pos33SortMsg

	// for new block incoming to add
	bch chan *types.Block
	// the last block
	lb *types.Block
}

// New create pos33 consensus client
func newNode(conf *subConfig) *node {
	n := &node{
		ips: make(map[int64]map[int]*pt.Pos33SortMsg),
		ivs: make(map[int64]map[int][]*pt.Pos33SortMsg),
		cps: make(map[int64]map[int]map[string]*pt.Pos33SortMsg),
		cvs: make(map[int64]map[int]map[string][]*pt.Pos33VoteMsg),
		css: make(map[int64]map[int][]*pt.Pos33SortMsg),
		bch: make(chan *types.Block, 32),
	}

	plog.Info("@@@@@@@ node start:", "addr", addr, "conf", conf)
	return n
}

// func (n *node) setLastBlock(height int64) {
// 	b := n.GetCurrentBlock()
// 	if b.Height == height {
// 		n.lb = b
// 		return
// 	}
// 	b, err := n.RequestBlock(height)
// 	if err != nil {
// 		panic(err)
// 	}
// 	n.lb = b
// }

func (n *node) lastBlock() *types.Block {
	// if n.lb != nil {
	// 	return n.lb
	// }
	b, err := n.RequestLastBlock()
	if err != nil {
		panic(err)
	}
	return b
}

var errNotEnoughVotes = errors.New("NOT enough votes for the last block")

func (n *node) minerTx(sm *pt.Pos33SortMsg, vs []*pt.Pos33VoteMsg, priv crypto.PrivKey) (*types.Transaction, error) {
	plog.Info("genRewordTx", "vsw", len(vs))
	if len(vs) > pt.Pos33VoterSize {
		sort.Sort(pt.Votes(vs))
		vs = vs[:pt.Pos33VoterSize]
	}
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

func (n *node) makeBlock(height int64, round int, tid string, vs []*pt.Pos33VoteMsg) error {
	sort := n.mySort(height, round)
	if sort == nil {
		err := fmt.Errorf("I'm not bp")
		return err
	}
	if sort.SortHash.Tid != tid {
		err := fmt.Errorf("I'm not bp2")
		return err
	}

	priv, err := n.getPrivByTid(sort.SortHash.Tid)
	if err != nil {
		return err
	}

	tx, err := n.minerTx(sort, vs, priv)
	if err != nil {
		return err
	}

	lb := n.lastBlock()
	nb, err := n.newBlock(lb, []*Tx{tx}, height)
	if err != nil {
		return err
	}

	nb.Difficulty = n.blockDiff(lb, len(vs))
	plog.Info("@@@@@@@ I make a block: ", "height", height, "round", round, "ntx", len(nb.Txs), "nvs", len(vs), "diff", nb.Difficulty)
	if nb.BlockTime-lb.BlockTime >= 1 {
		return n.setBlock(nb)
	}
	time.AfterFunc(time.Millisecond*500, func() { n.setBlock(nb) })
	return nil
}

func (n *node) addBlock(b *types.Block) {
	if !n.miningOK() {
		return
	}

	lastHeight := n.lastBlock().Height
	if b.Height != lastHeight {
		plog.Error("addBlock height error", "height", b.Height, "lastHeight", lastHeight)
		return
	}

	fn := func(nb *types.Block) {
		select {
		case n.bch <- nb:
		default:
			<-n.bch
			n.bch <- nb
		}
	}

	plog.Info("node.addBlock", "height", b.Height, "hash", common.ToHex(b.Hash(n.GetAPI().GetConfig())))
	fn(b)
}

func (n *node) clear(height int64) {
	for h := range n.cvs {
		if h+10 <= height {
			delete(n.cvs, h)
		}
	}
	for h := range n.css {
		if h+10 <= height {
			delete(n.css, h)
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

func (n *node) checkBlock(b, pb *types.Block) error {
	plog.Info("node.checkBlock", "height", b.Height, "pbheight", pb.Height)
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
	if !n.miningOK() {
		return nil
	}

	act, err := getMiner(b)
	if err != nil {
		return err
	}
	if act.Sort == nil || act.Sort.Proof == nil || act.Sort.Proof.Input == nil {
		return fmt.Errorf("miner tx error")
	}
	round := int(act.Sort.Proof.Input.Round)
	if string(act.Sort.Proof.Pubkey) == string(n.getPriv("").PubKey().Bytes()) {
		return nil
	}

	seed, err := n.getMinerSeed(b.Height)
	if err != nil {
		return err
	}
	allw := n.allw(b.Height, true)
	err = n.verifySort(b.Height, 0, allw, seed, act.GetSort())
	if err != nil {
		return err
	}

	// check votes
	height := b.Height
	rvs, ok := n.checkVotesEnough(act.Votes, height, round)
	if len(rvs) >= len(act.Votes) && !ok {
		return fmt.Errorf("the block NOT enouph vote, height=%d", b.Height)
	}
	if len(act.Votes) > pt.Pos33VoterSize {
		return fmt.Errorf("the block vote too much, height=%d", b.Height)
	}
	tid := act.Sort.SortHash.Tid
	for _, v := range act.Votes {
		err = n.checkVote(v, height, round, seed, allw, tid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *node) getMinerSort(height int64) (*pt.Pos33SortMsg, error) {
	startHeight := height - pt.Pos33SortitionSize
	b, err := n.RequestBlock(startHeight)
	if err != nil {
		plog.Error("RequstBlock error", "err", err, "height", startHeight)
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
			}
			n.sendVoteSorts(height, 0)
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
	vs := n.cvs[height][round][tid]
	plog.Info("handleVotesMsg", "height", height, "round", round, "tid", tid, "voter", addr(vm.GetSig()), "votes", len(vs))
	rvs, ok := n.checkVotesEnough(vs, height, round)
	if ok {
		if height <= n.lastBlock().Height {
			plog.Info("vote too late")
			return
		}

		err := n.makeBlock(height, round, tid, rvs)
		if err != nil {
			plog.Error("makeBlock error", "err", err)
		}
	}
}

func (n *node) getSorts(height int64, round int) []*pt.Pos33SortMsg {
	mp, ok := n.css[height]
	if !ok {
		return nil
	}
	ss, ok := mp[round]
	if !ok {
		return nil
	}
	return ss
}

func (n *node) checkVotesEnough(vs []*pt.Pos33VoteMsg, height int64, round int) ([]*pt.Pos33VoteMsg, bool) {
	ss := n.getSorts(height, round)
	plog.Info("checkVotesEnough", "ss len", len(ss), "vs len", len(vs))
	if len(ss) < pt.Pos33MinVotes {
		return nil, false
	}
	if len(vs)*3 <= len(ss)*2 {
		return nil, false
	}
	var rvs []*pt.Pos33VoteMsg
	for _, v := range vs {
		for _, s := range ss {
			if string(v.Sort.SortHash.Hash) == string(s.SortHash.Hash) {
				rvs = append(rvs, v)
				break
			}
		}
	}
	return rvs, len(rvs)*3 > len(ss)*2
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
	n.cps[height][round][m.SortHash.Tid] = m
	plog.Info("handleSortitionMsg", "height", height, "round", round, "size", len(n.cps[height][round]))
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
	height := m.Sorts[0].Proof.Input.Height
	plog.Info("handleSortsMsg", "height", height)
	if n.lastBlock().Height >= height {
		plog.Info("the sortsmsg height too low!!!", "lastBlockHeight", n.lastBlock().Height)
		return
	}
	if m.S != nil {
		n.handleSortitionMsg(m.S)
	}
	for _, s := range m.Sorts {
		if s == nil || s.Proof == nil || s.Proof.Input == nil || s.SortHash == nil {
			plog.Error("handleSortsMsg error, input msg is nil")
			return
		}
		/*
			if !myself {
				err := n.checkSort(s)
				if err != nil {
					plog.Error("checkSort error", "err", err)
					return
				}
			}
		*/
		height := s.Proof.Input.Height
		round := int(s.Proof.Input.Round)
		if n.css[height] == nil {
			n.css[height] = make(map[int][]*pt.Pos33SortMsg)
		}
		ss := n.css[height][round]
		ss = append(ss, s)
		// sort.Sort(pt.Sorts(ss))
		n.css[height][round] = ss
	}
}

func unmarshal(b []byte) (*pt.Pos33Msg, error) {
	var pm pt.Pos33Msg
	err := proto.Unmarshal(b, &pm)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (n *node) handlePos33Msg(pm *pt.Pos33Msg) bool {
	if pm == nil {
		return false
	}
	switch pm.Ty {
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

// handleGossipMsg multi-goroutine verify pos33 message
func (n *node) handleGossipMsg() chan *pt.Pos33Msg {
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

func (n *node) runLoop() {
	lb, err := n.RequestLastBlock()
	if err != nil {
		plog.Error(err.Error())
		return
	}

	n.gss = newGossip(n.nodeID(), n.conf.ListenAddr, n.conf.AdvertiseAddr, n.conf.BootPeerAddr)
	go n.gss.runBroadcast()
	msgch := n.handleGossipMsg()

	time.AfterFunc(time.Second, func() {
		n.addBlock(lb)
	})

	plog.Info("@@@@@@@@ pos33 node runing.......", "last block height", lb.Height)
	isSync := false
	syncTick := time.NewTicker(time.Microsecond * 100)
	etm := time.NewTimer(time.Hour)
	vtm := time.NewTimer(time.Hour)
	ch := make(chan int64, 1)
	// height := int64(0)
	round := 0

	for {
		select {
		case msg := <-msgch:
			n.handlePos33Msg(msg)
		case <-syncTick.C:
			isSync = n.IsCaughtUp()
		default:
		}
		if !isSync {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		select {
		case <-vtm.C:
		case height := <-ch:
			if height == n.lastBlock().Height+1 {
				round++
				plog.Info("@@@ vote timeout: ", "height", height, "round", round)
				n.reSortition(height, round)
				etm = time.NewTimer(time.Second * 3)
			}
		case <-etm.C:
			height := n.lastBlock().Height + 1
			n.vote(height, round)
			vtm = time.NewTimer(time.Second * 3)
			time.AfterFunc(time.Second*3, func() {
				ch <- height
			})
		case b := <-n.bch: // new block add to chain
			plog.Info("new block added", "height", b.Height)
			if b.Height%pt.Pos33SortitionSize == 0 {
				n.flushTicket()
			}
			n.sortition(b, 0)
			round = 0
			n.clear(b.Height - 1)
			etm = time.NewTimer(time.Millisecond * 10)
		default:
		}
	}
}

func (n *node) sendVoteSorts(height int64, round int) {
	ss := n.myVotes(height, round)
	s := n.mySort(height, round)

	m := &pt.Pos33Sorts{Sorts: ss, S: s}
	b := marshalSortsMsg(m)
	n.gss.gossip(b)
	n.handleSortsMsg(m, true)
}

func hexs(b []byte) string {
	s := hex.EncodeToString(b)
	if len(s) <= 16 {
		return s
	}
	return s[:16]
}

func (n *node) vote(height int64, round int) {
	tid := n.bp(height, round)
	if tid == "" {
		plog.Info("vote bp is nil")
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
	plog.Info("vote bp", "height", height, "round", round, "tid", tid)
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

func marshalSortsMsg(m *pt.Pos33Sorts) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(m),
		Ty:   pt.Pos33Msg_S,
	}
	return types.Encode(pm)
}

func marshalVoteMsg(v *pt.Pos33Votes) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(v),
		Ty:   pt.Pos33Msg_V,
	}
	return types.Encode(pm)
}
