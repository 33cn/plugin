package pos33

import (
	"encoding/hex"
	"errors"
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

	bch            chan *types.Block
	comm, lastComm *committee // current committee and next committee
	myWeight       int

	bmp map[int64]*types.Block    // cache blocks
	vmp map[int64][]*pt.Pos33Vote // cache votes
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
		bch:  make(chan *types.Block, 16),
		bmp:  make(map[int64]*types.Block),
		vmp:  make(map[int64][]*pt.Pos33Vote),
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

func (n *node) myHash(height int64) []byte {
	x, bp := n.commIndex(height)
	r := n.comm.Rands[x]
	if r.Addr != bp {
		panic("can't go here")
	}
	return r.RandHash
}

func (n *node) genRewordTx(height int64, vs []*pt.Pos33Vote) (*types.Transaction, error) {
	data, err := proto.Marshal(&pt.Pos33Action{
		Value: &pt.Pos33Action_Reword{
			Reword: &pt.Pos33RewordAction{
				Votes:    vs,
				RandHash: n.myHash(height),
			},
		},
		Ty: pt.Pos33ActionReword,
	})

	if err != nil {
		plog.Error(err.Error())
		return nil, err
	}

	tx := &types.Transaction{
		Execer:  []byte("pos33"),
		To:      address.GetExecAddress("pos33").String(),
		Payload: data,
		Fee:     pos33MinFee,
		Expire:  time.Now().Unix() + 10,
	}
	return tx, nil
}

func (n *node) signBlock(b *types.Block) *types.Block {
	sig := n.priv.Sign(b.Hash())
	b.Signature = &types.Signature{Pubkey: n.priv.PubKey().Bytes(), Ty: types.ED25519, Signature: sig.Bytes()}
	return b
}

func vsAccWeight(vs []*pt.Pos33Vote, acc string) (int, int) {
	for i, v := range vs {
		if addr(v.Sig) == acc {
			return int(v.Weight), i
		}
	}
	return 0, -1
}

func vsWeight(vs []*pt.Pos33Vote) int {
	w := 0
	for _, v := range vs {
		w += int(v.Weight)
	}
	return w
}

func (n *node) makeBlock(height int64, vs []*pt.Pos33Vote, null bool) (*types.Block, error) {
	tx, err := n.genRewordTx(height, vs)
	if err != nil {
		panic(err)
	}
	tx.Sign(types.ED25519, n.priv)
	txs := []*types.Transaction{tx}
	plog.Info("@@@@@@@ I make a block: ", "height", height, "isNull", null)
	nb, err := n.newBlock(txs, height, null)
	if err != nil {
		plog.Error("makeBlock error", "height", height, "error", err.Error())
		return nil, err
	}

	nb.Difficulty += uint32(vsWeight(vs))
	snb := n.signBlock(nb)
	n.setBlock(snb)
	// n.gss.broadcastTCP(n.marshalBlockMsg(nb))
	return nb, nil
}

func (n *node) addBlock(b *types.Block) {
	if !n.IsCaughtUp() {
		return
	}
	plog.Info("node.addBlock", "height", b.Height)
	n.bch <- b
}

func getWeight(comm *committee, u string) int {
	w := 0
	for _, r := range comm.Rands {
		if u == r.Addr {
			w++
		}
	}
	return w
}

// TODO:
func (n *node) checkVote(vt *pt.Pos33Vote) bool {
	who := addr(vt.Sig)
	cw := getWeight(n.comm, who)
	if int(vt.Weight) != cw {
		plog.Error("vote weight error", "addr", who, "vtw", vt.Weight, "comm_weight", cw)
		return false
	}
	return true
}

func (n *node) countVote(height int64) (int64, string, []*pt.Pos33Vote, bool) {
	plog.Info("countVote ", "height", height)
	vts := n.vmp[height]
	hmp := make(map[string][]*pt.Pos33Vote)
	for _, v := range vts {
		hmp[string(v.BlockHash)] = append(hmp[string(v.BlockHash)], v)
	}
	max := 0
	maxHash := ""
	for k, vs := range hmp {
		w := 0
		for _, v := range vs {
			w += int(v.Weight)
		}
		if w > max {
			max = w
			maxHash = k
		}
	}

	// plog.Info("countVote ", "max", max, "maxHash", hex.EncodeToString([]byte(maxHash)))
	if !n.comm.stoped && max*3 < pt.Pos33CommitteeSize*2 {
		return -1, "", nil, false
	}

	bmp := make(map[string][]*pt.Pos33Vote)
	for _, v := range hmp[maxHash] {
		bmp[v.Bp] = append(bmp[v.Bp], v)
	}

	max = 0
	maxBp := ""
	for k, vs := range bmp {
		w := 0
		for _, v := range vs {
			w += int(v.Weight)
		}
		if w > max {
			max = w
			maxBp = k
		}
	}

	// plog.Info("countVote ", "max", max, "maxBp", maxBp)

	if !n.comm.stoped && max*3 < pt.Pos33CommitteeSize*2 {
		return -1, "", nil, false
	}

	if maxHash == "nil" { // block error or timeout
		// p, bp := n.commIndex(height) // height 高度的 bp和位置
		// x := n.findIndex(maxBp, p)   // 投票选择的bp的位置
		// n.comm.Rands[p] = n.comm.Rands[x] // 使用正确节点代替错误的
		// plog.Info("use maxBp replace bp", "height", height, "bp", bp, "maxbp", maxBp, "bp_pos", p, "maxbp_pos", x)
		return height, maxBp, bmp[maxBp], true
	}
	return height + 1, maxBp, bmp[maxBp], false
}

func (n *node) findIndex(who string, p int) int {
	for i := p; i < len(n.comm.Rands); i++ {
		if n.comm.Rands[i].Addr == who {
			return i
		}
	}
	return -1
}

func (n *node) handleVote(vt *pt.Pos33Vote) {
	plog.Info("n.handleVote", "height", vt.BlockHeight, "addr", addr(vt.Sig), "vt.hash", hex.EncodeToString(vt.BlockHash), "bp", vt.Bp)
	lastB, err := n.RequestLastBlock()
	if err != nil {
		panic("can't go here")
	}
	if lastB.Height > vt.BlockHeight || vt.BlockHeight < n.comm.height {
		plog.Info("vote too late", "lastHeight", lastB.Height)
		return
	}

	who := addr(vt.Sig)
	w, i := vsAccWeight(n.vmp[vt.BlockHeight], who)
	if w > 0 {
		plog.Info("vote repeated", "addr", who)
		n.vmp[vt.BlockHeight][i] = vt
	}

	if n.comm == nil {
		return
	}

	if vt.BlockHeight >= n.comm.height+pt.Pos33CommitteeSize {
		n.vmp[vt.BlockHeight] = append(n.vmp[vt.BlockHeight], vt)
		plog.Info("vote too early", "lastHeight", lastB.Height)
		return
	}

	if !n.checkVote(vt) {
		plog.Info("chechVote failed", "addr", addr(vt.Sig))
		return
	}

	n.vmp[vt.BlockHeight] = append(n.vmp[vt.BlockHeight], vt)
	dt := time.Now().Unix() - lastB.BlockTime
	if vt.BlockHeight == lastB.Height && dt > 1 && dt < 3 {
		n.newRound(vt.BlockHeight)
	}
}

func addr(sig *types.Signature) string {
	return address.PubKeyToAddress(sig.Pubkey).String()
}

func (n *node) handleBlock(b *types.Block) {
	plog.Info("node.handleBlock", "height", b.Height, "bp", addr(b.Signature))
	lb, err := n.RequestLastBlock()
	if err != nil {
		plog.Error("node.handleBlock error:", "err", err.Error(), "height", b.Height)
		return
	}
	if lb.Height >= b.Height {
		return
	} else if lb.Height+1 < b.Height {
		n.bmp[b.Height] = b
		return
	}

	// TODO: check the block
	n.setBlock(b)
}

func (n *node) checkBlock(b *types.Block) error {
	if b.Height == 0 {
		return nil
	}

	plog.Info("node.checkBlock", "height", b.Height)

	// check first Tx
	tx := b.Txs[0]
	var act pt.Pos33Action
	err := types.Decode(tx.GetPayload(), &act)
	if err != nil {
		return err
	}
	if act.Ty != pt.Pos33ActionReword {
		return errors.New("first tx must include reword action")
	}
	rewordAct := act.GetReword()

	// must enought votes
	w := vsWeight(rewordAct.Votes)
	if w*3 < pt.Pos33CommitteeSize*2 {
		return errors.New("block vote weight too low")
	}

	comm := n.comm
	if comm != nil {
		// block maker must be committee
		bp := addr(b.Signature)
		ok := false
		for _, r := range comm.Rands {
			if r.Addr == bp {
				ok = true
				break
			}
		}
		if !ok {
			return errors.New("block maker is NOT in commmittee")
		}
	}

	// ok
	n.addBlock(b)
	return nil
}

var zeroHash [32]byte

func getBlockSeed(b *types.Block) []byte {
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
func (n *node) sortition(b *types.Block) error {
	seed := getBlockSeed(b)

	height := b.Height
	rands, sig := pt.GenRands(n.allWeight(), n.getWeight(n.addr), n.priv, height, seed)
	if rands == nil {
		plog.Info("sortiton nil", "height", b.Height)
		return nil
	}
	plog.Info("node.sortition", "height", height, "weight", len(rands))

	tx, err := pt.NewElecteTx(rands, seed, height, sig)
	if err != nil {
		return err
	}

	tx.Sign(types.ED25519, n.priv)
	return n.sendTx(tx)
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
	case pt.Pos33Msg_V:
		var vt pt.Pos33Vote
		err := types.Decode(pm.Data, &vt)
		if err != nil {
			plog.Error(err.Error())
			return false
		}
		if n.myWeight > 0 {
			n.handleVote(&vt)
		}
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

func printCommittee(comm *committee) {
	for i, r := range comm.Rands {
		plog.Info("current committee", "index", i, "addr", r.Addr, "hash", hex.EncodeToString(r.RandHash))
	}
}

func (n *node) changeCommittee(b *types.Block) {
	err := n.sortition(b)
	if err != nil {
		plog.Error("sortition error", "err", err)
		return
	}

	if b.Height > 0 {
		comm, err := n.getCurrentCommittee(b.Height)
		if err != nil {
			if err != nil {
				plog.Error("getCurrentCommittee error", "err", err)
				return
			}
		}
		plog.Info("@@@@@@@ current committee", "height", b.Height)
		n.comm = &committee{Pos33Rands: comm, height: b.Height}
		printCommittee(n.comm)
		n.lastComm = n.comm
	}
	if len(n.comm.Rands) != pt.Pos33CommitteeSize {
		panic("can't go here")
	}
	n.myWeight = 0
	for _, r := range n.comm.Rands {
		if n.addr == r.Addr {
			n.myWeight++
		}
	}
}

func (n *node) commIndex(height int64) (int, string) {
	x := height%int64(pt.Pos33CommitteeSize) - 1
	if x < 0 { // last
		x = pt.Pos33CommitteeSize - 1
	}
	return int(x), n.comm.Rands[x].Addr
}

func (n *node) voteBlock(height int64, hash []byte) {
	x, bp := n.commIndex(height)
	nbp := bp
	if string(hash) != "nil" {
		x++
		nbp = n.comm.Rands[x%pt.Pos33CommitteeSize].Addr
	} else {
		for bp == nbp {
			x++
			nbp = n.comm.Rands[x%pt.Pos33CommitteeSize].Addr
		}
	}
	vt := &pt.Pos33Vote{
		BlockHeight: height,
		BlockHash:   hash,
		Bp:          nbp,
		Weight:      int32(n.myWeight),
	}
	vt.Sign(n.priv)
	n.gss.broadcastUDP(n.marshalVoteMsg(vt))
	n.handleVote(vt)
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

func (n *node) firstCommittee() error {
	height := int64(-1)
	seed := zeroHash[:]
	rands, sig := pt.GenRands(n.allWeight(), n.getWeight(n.addr), n.priv, height, seed)
	if rands == nil {
		plog.Info("sortiton nil", "height", height)
		return nil
	}
	plog.Info("node.sortition", "height", height, "weight", len(rands))

	act := &pt.Pos33ElecteAction{Rands: rands, Hash: seed, Height: height, Sig: sig}
	comm := pt.Sortition([]*pt.Pos33ElecteAction{act})
	n.comm = &committee{Pos33Rands: comm, height: 0}
	n.myWeight = len(n.comm.Rands)
	return nil
}

func (n *node) newRound(height int64) {
	if n.myWeight == 0 {
		plog.Info("I'm not a committee", "addr", n.addr, "height", height)
		return
	}
	changed := false
	if height < n.comm.height {
		plog.Info("committee changed, this height must use lastComm", "height", height)
		n.comm, n.lastComm = n.lastComm, n.comm
		changed = true
	}
	defer func() {
		if changed {
			n.comm, n.lastComm = n.lastComm, n.comm
		}
	}()
	if !n.comm.stoped {
		lastB, err := n.RequestLastBlock()
		if err != nil {
			plog.Crit("should't go here")
			return
		}
		if time.Now().Unix()-lastB.BlockTime >= int64(pt.Pos33MaxCommittee) {
			n.comm.stoped = true
		}
	}
	newHeight, bp, vs, null := n.countVote(height)
	if newHeight < 0 {
		plog.Error("vote NOT enought", "addr", n.addr, "height", height)
		return
	}
	if bp == n.addr {
		n.makeBlock(newHeight, vs, null)
	}
	n.clear(height)
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

	timeoutTm := time.NewTimer(time.Hour)

	ch := make(chan int64, 1)

	if lb.Height == 0 {
		n.firstCommittee()
	}
	time.AfterFunc(time.Second, func() { n.addBlock(lb) })

	for {
		select {
		case msg := <-msgch:
			n.handlePos33Msg(msg)
		case <-timeoutTm.C:
			height := lb.Height + 1
			plog.Info("timeout......", "height", height)
			reseTm(timeoutTm, time.Second*5)
			if n.myWeight > 0 {
				n.voteBlock(height, []byte("nil"))
			}
			time.AfterFunc(time.Second*2, func() { ch <- height })
		case height := <-ch:
			n.newRound(height)
		case b := <-n.bch: // new block add to chain
			lb = b
			reseTm(timeoutTm, time.Second*3)

			if b.Height%pt.Pos33CommitteeSize == 0 {
				n.changeCommittee(b)
			}
			if n.myWeight > 0 {
				n.voteBlock(b.Height, b.Hash())
				h := b.Height
				time.AfterFunc(time.Second, func() { ch <- h })
			}
		}
	}
}

func (n *node) marshalBlockMsg(b *types.Block) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(b),
		Ty:   pt.Pos33Msg_B,
	}
	return types.Encode(pm)
}

func (n *node) marshalVoteMsg(v *pt.Pos33Vote) []byte {
	pm := &pt.Pos33Msg{
		Data: types.Encode(v),
		Ty:   pt.Pos33Msg_V,
	}
	return types.Encode(pm)
}
