package pos33

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/difficulty"
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
	ips map[int64]*pt.Pos33SortitionMsg
	// I'm candidate verifer in these blocks
	ivs map[int64][]*pt.Pos33SortitionMsg
	// receive candidate proposers
	cps map[int64]map[int]map[string]*pt.Pos33SortitionMsg
	// receive candidate verifers
	cvs map[int64]map[int]map[string][]*pt.Pos33VoteMsg
	// receive candidate blocks
	cbs map[int64]map[string]*types.Block

	voteOkHeight int64
	lastBlock    *types.Block
	bch          chan *types.Block
	etm          *time.Timer
}

// New create pos33 consensus client
func newNode(conf *subConfig) *node {
	n := &node{
		ips: make(map[int64]*pt.Pos33SortitionMsg),
		ivs: make(map[int64][]*pt.Pos33SortitionMsg),
		cps: make(map[int64]map[int]map[string]*pt.Pos33SortitionMsg),
		cvs: make(map[int64]map[int]map[string][]*pt.Pos33VoteMsg),
		cbs: make(map[int64]map[string]*types.Block),
		bch: make(chan *types.Block, 1),
	}

	plog.Info("@@@@@@@ node start:", "addr", addr, "conf", conf)
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

func (n *node) genMinerTx(height int64, round int, bp string) (*types.Transaction, int, error) {
	vs := n.cvs[height][round][bp]

	plog.Info("genRewordTx", "height", height, "vsw", len(vs))
	if height > 0 && len(vs)*3 <= pt.Pos33VoterSize*2 {
		return nil, 0, fmt.Errorf("NOT enouph votes")
	}

	if len(vs) > pt.Pos33VoterSize {
		sort.Sort(pt.Votes(vs))
		vs = vs[:pt.Pos33VoterSize]
	}

	sort := n.ips[height]
	act := &pt.Pos33TicketAction{
		Value: &pt.Pos33TicketAction_Pminer{
			Pminer: &pt.Pos33Miner{
				Votes: vs,
				Sort:  sort,
			},
		},
		Ty: pt.Pos33TicketActionMiner,
	}

	cfg := n.GetAPI().GetConfig()
	tx, err := types.CreateFormatTx(cfg, "pos33", types.Encode(act))
	if err != nil {
		return nil, 0, err
	}

	priv := n.privmap[n.ticketsMap[sort.Input.TicketId].MinerAddress]
	tx.Sign(types.SECP256K1, priv)
	return tx, len(vs), nil
}

func (n *node) diff(w int) uint32 {
	// return types.GetP(0).PowLimitBits
	cfg := n.GetAPI().GetConfig()
	oldDiff := difficulty.CompactToBig(cfg.GetP(0).PowLimitBits)
	newDiff := new(big.Int).Add(oldDiff, big.NewInt(int64(w+1)))
	return difficulty.BigToCompact(newDiff)
}

func (n *node) makeBlock(height int64, round int, bp string) error {
	pm, ok := n.ips[height]
	if !ok {
		err := fmt.Errorf("makeBlock: I'm not Proposer height=%d, round=%d", height, round)
		plog.Error("makeBlock error", "err", err)
		return err
	}
	minerAddr := address.PubKeyToAddr(pm.Pubkey)
	if minerAddr != bp {
		err := fmt.Errorf("I'm not Bp %d, %d", height, round)
		plog.Error("makeBlock error", "err", err)
		return err
	}

	tx, w, err := n.genMinerTx(height, round, bp)
	if err != nil {
		plog.Error("genRewordTx error", "err", err.Error(), "height", height, "round", round, "bp", bp)
		return err
	}

	nb, err := n.newBlock(n.lastBlock, []*Tx{tx}, height)
	if err != nil {
		plog.Error("makeBlock error", "height", height, "round", round, "error", err.Error())
		return err
	}

	nb.Difficulty = n.diff(w)

	plog.Info("@@@@@@@ I make a block: ", "height", height, "hash", hexs(nb.HashOld()), "txCount", len(nb.Txs), "diff", nb.Difficulty)
	return n.setBlock(nb)
}

func (n *node) addBlock(b *types.Block) {
	if !n.IsCaughtUp() {
		return
	}
	plog.Info("node.addBlock", "height", b.Height, "hash", hexs(b.HashOld()))

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

func (n *node) checkBlock(b, pb *types.Block) error {
	plog.Info("node.checkBlock", "height", b.Height)
	if b.Height < 2 {
		return nil
	}
	if !n.IsCaughtUp() {
		return nil
	}
	if len(b.Txs) == 0 {
		return fmt.Errorf("nil block error")
	}

	if string(b.ParentHash) != string(pb.Hash(n.GetAPI().GetConfig())) {
		return fmt.Errorf("block parent hash error")
	}

	act, err := getMiner(b)
	if err != nil {
		return err
	}
	seed, err := n.getMinerSeed(b.Height)
	if err != nil {
		return nil
	}
	err = n.verifySort(b.Height, 0, n.allWeight(b.Height), seed, act.GetSort())
	if err != nil {
		return err
	}

	round := act.GetSort().Input.Round
	// check votes
	if len(act.Votes)*3 <= pt.Pos33VoterSize*2 {
		return fmt.Errorf("the block NOT enouph vote, height=%d", b.Height)
	}
	if len(act.Votes) > pt.Pos33VoterSize {
		return fmt.Errorf("the block vote too much, height=%d", b.Height)
	}
	allw := n.allWeight(b.Height)
	for _, v := range act.Votes {
		if !v.Verify() {
			return fmt.Errorf("vote signature error")
		}
		m := v.Sort
		// check height, height of vote is the pre-block height
		if m.Input.Height != b.Height {
			return fmt.Errorf("height error")
		}

		// check round, all round of vote must be same
		if round != m.Input.Round {
			return fmt.Errorf("round error")
		}

		// check vrf
		err = n.verifySort(b.Height, 1, allw, seed, m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *node) getMinerSeed(height int64) ([]byte, error) {
	startHeight := height - height%pt.Pos33SortitionSize
	if startHeight == height {
		startHeight -= pt.Pos33SortitionSize
	}
	b, err := n.RequestBlock(startHeight)
	if err != nil {
		plog.Info("should't go here. do nothing")
		return nil, err
	}
	seed := zeroHash[:]
	if b.Height > 0 {
		act, err := getMiner(b)
		if err != nil {
			plog.Error("getBlockMiner err", "error", err, "height", b.Height)
			return nil, err
		}
		seed = act.Sort.GetHash()
	}
	return seed, nil
}

var zeroHash [32]byte

func (n *node) reSortition(height int64, round int) {
	seed, err := n.getMinerSeed(height)
	if err != nil {
		return
	}
	const staps = 2
	allw := n.allWeight(height)
	for s := 0; s < staps; s++ {
		sms := n.sort(seed, height, round, s, allw)
		if sms == nil {
			plog.Info("node.sortition nil")
			continue
		}
		plog.Info("node.reSortition", "height", height, "round", round, "step", s, "weight", len(sms))
		if s == 0 {
			n.ips[height] = sms[0]
		} else {
			n.ivs[height] = sms
		}
	}
}

func (n *node) sortition(b *types.Block, round int) {
	seed := zeroHash[:]
	startHeight := int64(0)
	plog.Info("sortition", "height", b.Height, "round", round)
	if b.Height > 0 {
		act, err := getMiner(b)
		if err != nil {
			plog.Error("getBlockMiner err", "error", err, "height", b.Height)
		} else {
			seed = act.Sort.Hash
			startHeight = b.Height
		}
	}
	startHeight++
	const steps = 2
	allw := n.allWeight(startHeight)
	for s := 0; s < steps; s++ {
		for i := 0; i < pt.Pos33SortitionSize; i++ {
			height := startHeight + int64(i)
			sms := n.sort(seed, height, round, s, allw)
			if sms == nil {
				plog.Info("node.sortition nil")
				continue
			}
			plog.Info("node.sortition", "height", height, "weight", len(sms))
			if s == 0 {
				n.ips[height] = sms[0]
			} else {
				n.ivs[height] = sms
			}
		}
	}
}

func (n *node) handleVote(vm *pt.Pos33VoteMsg, height int64, round, allw int, bp string) error {
	if !vm.Verify() {
		return fmt.Errorf("votemsg verify false")
	}
	if vm.Sort == nil && vm.Sort.Input == nil {
		return fmt.Errorf("votemsg error")
	}
	if height != vm.Sort.Input.Height {
		return fmt.Errorf("vote height is NOT consistent")
	}
	if round != int(vm.Sort.GetInput().Round) {
		return fmt.Errorf("vote round is NOT consistent")
	}

	if bp != vm.Bp {
		return fmt.Errorf("vote Bp is NOT consistent")
	}

	if vm == nil || height <= n.lastBlock.Height {
		return fmt.Errorf("votemsg too late height = %d", height)
	}

	if string(vm.Sig.Pubkey) != string(vm.Sort.Pubkey) {
		return fmt.Errorf("vote pubkey is NOT consistent")
	}

	seed, err := n.getMinerSeed(height)
	if err != nil {
		return err
	}
	err = n.verifySort(height, 1, allw, seed, vm.Sort)
	if err != nil {
		return err
	}

	if n.cvs[height] == nil {
		n.cvs[height] = make(map[int]map[string][]*pt.Pos33VoteMsg)
	}

	if n.cvs[height][round] == nil {
		mp := make(map[string][]*pt.Pos33VoteMsg)
		n.cvs[height][round] = mp
	}

	vs := n.cvs[height][round][vm.Bp]
	for _, v := range vs {
		if v.Equal(vm) {
			return fmt.Errorf("votemsg repeated")
		}
	}
	vs = append(vs, vm)
	n.cvs[height][round][vm.Bp] = vs

	return nil
}

func (n *node) handleVotesMsg(vms *pt.Pos33Votes) {
	if n.lastBlock == nil {
		return
	}
	if len(vms.Vs) == 0 {
		plog.Error("votemsg sortition is 0")
		return
	}

	vm := vms.Vs[0]
	if vm.Sort == nil && vm.Sort.Input == nil {
		return
	}
	height := vm.Sort.Input.Height
	round := int(vm.Sort.Input.Round)
	bp := vm.Bp

	if n.voteOkHeight >= height {
		return
	}
	allw := n.allWeight(height)
	for _, vm := range vms.Vs {
		err := n.handleVote(vm, height, round, allw, bp)
		if err != nil {
			plog.Error("handleVote error", "err", err)
			return
		}
	}
	if len(n.cvs[height][round][bp])*3 > pt.Pos33VoterSize*2 {
		plog.Info("@@@ set block 2f+1 @@@", "height", height, "bp", bp)

		//n.setBlock(b)
		n.makeBlock(height, round, bp)
	}

}

func (n *node) handleSortitionMsg(m *pt.Pos33SortitionMsg) {
	height := m.Input.Height
	plog.Info("handleElectMsg", "height", height)
	seed, err := n.getMinerSeed(height)
	n.verifySort(height, 0, n.allWeight(height), seed, m)
	if err != nil {
		plog.Info("check rand error:", "error", err.Error())
		return
	}
	if n.cps[height] == nil {
		n.cps[height] = make(map[int]map[string]*pt.Pos33SortitionMsg)
	}
	round := int(m.Input.Round)
	if n.cps[height][round] == nil {
		n.cps[height][round] = make(map[string]*pt.Pos33SortitionMsg)
	}
	n.cps[height][round][m.Input.TicketId] = m
}

func (n *node) handlePos33Msg(pm *pt.Pos33Msg) bool {
	if pm == nil {
		return false
	}
	switch pm.Ty {
	case pt.Pos33Msg_S:
		var m pt.Pos33SortitionMsg
		err := types.Decode(pm.Data, &m)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleSortitionMsg(&m)
	case pt.Pos33Msg_V:
		var vt pt.Pos33Votes
		err := types.Decode(pm.Data, &vt)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		n.handleVotesMsg(&vt)
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

func (n *node) runLoop() {
	lb, err := n.RequestLastBlock()
	if err != nil {
		plog.Error(err.Error())
		return
	}

	plog.Info("pos33 node runing.......", "last block height", lb.Height)

	n.gss = newGossip(n.nodeID(), n.conf.ListenAddr, n.conf.AdvertiseAddr, n.conf.BootPeerAddr)
	go n.gss.runBroadcast()
	msgch := n.doGossipMsg()

	n.etm = time.NewTimer(time.Hour)
	ch := make(chan int64, 1)

	time.AfterFunc(time.Second, func() { n.addBlock(lb) })
	round := 0

	for {
		if !n.miningOK() {
			time.Sleep(time.Second)
			continue
		}
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

			to := (round + 1) * 2
			if to > 16 {
				to = 16
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
				n.flushTicket()
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
	ivs, ok := n.ivs[height]
	if !ok {
		plog.Info("I'm not verifer", "height", height)
		return
	}

	var pss []*pt.Pos33SortitionMsg
	for _, s := range n.cps[height][round] {
		pss = append(pss, s)
	}
	if len(pss) == 0 {
		plog.Info("no block maker", "height", height, "round", round)
		return
	}
	plog.Info("vote len(pss)", "height", height, "round", round, "len(pss)", len(pss))

	sort.Sort(pt.Sorts(pss))
	bp := address.PubKeyToAddress(pss[0].Pubkey).String()

	plog.Info("vote bp", "height", height, "round", round, "bp", bp)
	var vs []*pt.Pos33VoteMsg
	for _, s := range ivs {
		v := &pt.Pos33VoteMsg{Sort: s, Bp: bp}
		t := n.ticketsMap[s.Input.TicketId]
		priv := n.privmap[t.MinerAddress]
		v.Sign(priv)
		vs = append(vs, v)
	}
	v := &pt.Pos33Votes{Vs: vs}
	n.handleVotesMsg(v)
	n.gss.gossip(marshalVoteMsg(v))
}

func (n *node) elect(height int64, round int) {
	n.etm = time.NewTimer(time.Millisecond * 1000)
	pm, ok := n.ips[height]
	if !ok {
		plog.Info("elect: I'm not Proposer", "height", height, "round", round)
		return
	}
	n.handleSortitionMsg(pm)
	n.gss.gossip(marshalSortMsg(pm))
}

func marshalSortMsg(m *pt.Pos33SortitionMsg) []byte {
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
