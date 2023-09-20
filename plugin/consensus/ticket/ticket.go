// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticket

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/address/eth"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/difficulty"
	"github.com/33cn/chain33/common/log/log15"
	vrf "github.com/33cn/chain33/common/vrf/secp256k1"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	driver "github.com/33cn/chain33/system/dapp"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/golang/protobuf/proto"
)

var (
	tlog          = log15.New("module", "ticket")
	defaultModify = []byte("modify")
)

const (
	defaultFlushTicketInterval = 3600
)

func init() {
	drivers.Reg("ticket", New)
	drivers.QueryData.Register("ticket", &Client{})
}

// Client export ticket client struct
type Client struct {
	*drivers.BaseClient
	//ticket map for miner
	ticketsMap map[string]*ty.Ticket
	privmap    map[string]crypto.PrivKey
	ticketmu   sync.Mutex
	done       chan struct{}
	subcfg     *subConfig
}

type genesisTicket struct {
	MinerAddr  string `json:"minerAddr"`
	ReturnAddr string `json:"returnAddr"`
	Count      int32  `json:"count"`
}

type subConfig struct {
	GenesisBlockTime int64            `json:"genesisBlockTime"`
	Genesis          []*genesisTicket `json:"genesis"`
	// FlushTicketInterval flush ticket status backend interval
	FlushTicketInterval int64 `json:"flushTicketInterval"`
}

// New  ticket's init env
func New(cfg *types.Consensus, sub []byte) queue.Module {
	c := drivers.NewBaseClient(cfg)
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}
	if subcfg.GenesisBlockTime > 0 {
		cfg.GenesisBlockTime = subcfg.GenesisBlockTime
	}
	if subcfg.FlushTicketInterval <= 0 {
		subcfg.FlushTicketInterval = defaultFlushTicketInterval
	}
	t := &Client{
		BaseClient: c,
		ticketsMap: make(map[string]*ty.Ticket),
		privmap:    nil,
		ticketmu:   sync.Mutex{},
		done:       make(chan struct{}),
		subcfg:     &subcfg}
	c.SetChild(t)
	go t.flushTicketBackend()
	return t
}

func (client *Client) flushTicketBackend() {
	ticket := time.NewTicker(time.Duration(client.subcfg.FlushTicketInterval) * time.Second)
	defer ticket.Stop()
Loop:
	for {
		select {
		case <-ticket.C:
			client.flushTicket()
		case <-client.done:
			break Loop
		}
	}
}

// Close ticket close
func (client *Client) Close() {
	close(client.done)
	client.BaseClient.Close()
	tlog.Info("consensus ticket closed")
}

// CreateGenesisTx ticket create genesis tx
func (client *Client) CreateGenesisTx() (ret []*types.Transaction) {
	cfg := client.GetAPI().GetConfig()
	for _, genesis := range client.subcfg.Genesis {
		tx1 := createTicket(cfg, genesis.MinerAddr, genesis.ReturnAddr, genesis.Count, 0)
		ret = append(ret, tx1...)
	}
	return ret
}

//316190000 coins
func createTicket(cfg *types.Chain33Config, minerAddr, returnAddr string, count int32, height int64) (ret []*types.Transaction) {
	tx1 := types.Transaction{}
	tx1.Execer = []byte(cfg.GetCoinExec())

	//给hotkey 10000 个币，作为miner的手续费
	tx1.To = minerAddr
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{Amount: ty.GetTicketMinerParam(cfg, height).TicketPrice}
	tx1.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx1)

	tx2 := types.Transaction{}

	tx2.Execer = []byte(cfg.GetCoinExec())
	tx2.To = driver.ExecAddress("ticket")
	//gen payload
	g = &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{Amount: int64(count) * ty.GetTicketMinerParam(cfg, height).TicketPrice, ReturnAddress: returnAddr}
	tx2.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx2)

	tx3 := types.Transaction{}
	tx3.Execer = []byte("ticket")
	tx3.To = driver.ExecAddress("ticket")
	gticket := &ty.TicketAction_Genesis{}
	gticket.Genesis = &ty.TicketGenesis{MinerAddress: minerAddr, ReturnAddress: returnAddr, Count: count}
	tx3.Payload = types.Encode(&ty.TicketAction{Value: gticket, Ty: ty.TicketActionGenesis})
	ret = append(ret, &tx3)
	return ret
}

// Query_GetTicketCount ticket query ticket count function
func (client *Client) Query_GetTicketCount(req *types.ReqNil) (types.Message, error) {
	var ret types.Int64
	ret.Data = client.getTicketCount()
	return &ret, nil
}

// Query_FlushTicket ticket query flush ticket function
func (client *Client) Query_FlushTicket(req *types.ReqNil) (types.Message, error) {
	err := client.flushTicket()
	if err != nil {
		return nil, err
	}
	return &types.Reply{IsOk: true, Msg: []byte("OK")}, nil
}

// ProcEvent ticket reply not support action err
func (client *Client) ProcEvent(msg *queue.Message) bool {
	msg.ReplyErr("Client", types.ErrActionNotSupport)
	return true
}

func (client *Client) privFromBytes(privkey []byte) (crypto.PrivKey, error) {
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		return nil, err
	}
	return cr.PrivKeyFromBytes(privkey)
}

func (client *Client) getTickets() ([]*ty.Ticket, []crypto.PrivKey, error) {
	resp, err := client.GetAPI().ExecWalletFunc("ticket", "WalletGetTickets", &types.ReqNil{})
	if err != nil {
		return nil, nil, err
	}
	reply := resp.(*ty.ReplyWalletTickets)
	var keys []crypto.PrivKey
	for i := 0; i < len(reply.Privkeys); i++ {
		priv, err := client.privFromBytes(reply.Privkeys[i])
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, priv)
	}
	tlog.Info("getTickets", "ticket n", len(reply.Tickets), "nkey", len(keys))
	return reply.Tickets, keys, nil
}

func (client *Client) getTicketCount() int64 {
	client.ticketmu.Lock()
	defer client.ticketmu.Unlock()
	return int64(len(client.ticketsMap))
}

func (client *Client) setTicket(tlist *ty.ReplyTicketList, privmap map[string]crypto.PrivKey) {
	client.ticketmu.Lock()
	defer client.ticketmu.Unlock()
	client.ticketsMap = make(map[string]*ty.Ticket)
	if tlist == nil || privmap == nil {
		client.ticketsMap = nil
		client.privmap = nil
		return
	}
	for _, ticket := range tlist.Tickets {
		client.ticketsMap[ticket.GetTicketId()] = ticket
	}
	//client.tlist = tlist
	client.privmap = privmap
	tlog.Debug("setTicket", "n", len(tlist.GetTickets()))
}

func (client *Client) flushTicket() error {
	//list accounts
	tickets, privs, err := client.getTickets()
	if err == types.ErrWalletIsLocked || err == ty.ErrNoTicket {
		tlog.Error("flushTicket error", "err", err.Error())
		client.setTicket(nil, nil)
		return nil
	}
	if err != nil {
		tlog.Error("flushTicket error", "err", err)
		return err
	}
	client.setTicket(&ty.ReplyTicketList{Tickets: tickets}, getPrivMap(privs, tickets))
	return nil
}

func getPrivMap(privs []crypto.PrivKey, tickets []*ty.Ticket) map[string]crypto.PrivKey {
	list := make(map[string]crypto.PrivKey)
	tempMinerList := make(map[string]string)
	for _, ts := range tickets {
		addr := strings.Split(ts.GetTicketId(), ":")[0]
		tempMinerList[addr] = ts.GetTicketId()
	}

	for addr := range tempMinerList {
		var addressID int32 = address.DefaultID
		if common.IsHex(addr) {
			addressID = eth.ID
		}

		for _, pk := range privs {
			newAddr := address.PubKeyToAddr(addressID, pk.PubKey().Bytes())
			if newAddr == addr {
				list[addr] = pk
			}

		}
	}

	return list
}

func (client *Client) getMinerTx(current *types.Block) (*ty.TicketAction, error) {
	//检查第一个笔交易的execs, 以及执行状态
	if len(current.Txs) == 0 {
		return nil, types.ErrEmptyTx
	}
	baseTx := current.Txs[0]
	//判断交易类型和执行情况
	var ticketAction ty.TicketAction
	err := types.Decode(baseTx.GetPayload(), &ticketAction)
	if err != nil {
		return nil, err
	}
	if ticketAction.GetTy() != ty.TicketActionMiner {
		return nil, types.ErrCoinBaseTxType
	}
	//判断交易执行是否OK
	if ticketAction.GetMiner() == nil {
		return nil, ty.ErrEmptyMinerTx
	}
	return &ticketAction, nil
}

func (client *Client) getMinerModify(block *types.Block) ([]byte, error) {
	ticketAction, err := client.getMinerTx(block)
	if err != nil {
		return defaultModify, err
	}
	return ticketAction.GetMiner().GetModify(), nil
}

func (client *Client) getModify(beg, end int64) ([]byte, error) {
	//通过某个区间计算modify
	timeSource := int64(0)
	total := int64(0)
	newmodify := ""
	for i := beg; i <= end; i++ {
		block, err := client.RequestBlock(i)
		if err != nil {
			return defaultModify, err
		}
		timeSource += block.BlockTime
		if total == 0 {
			total = block.BlockTime
		}
		if timeSource%4 == 0 {
			total += block.BlockTime
		}
		if i == end {
			ticketAction, err := client.getMinerTx(block)
			if err != nil {
				return defaultModify, err
			}
			last := ticketAction.GetMiner().GetModify()
			newmodify = fmt.Sprintf("%s:%d", string(last), total)
		}
	}
	modify := common.ToHex(common.Sha256([]byte(newmodify)))
	return []byte(modify), nil
}

// CheckBlock ticket implete checkblock func
func (client *Client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	chain33Cfg := client.GetAPI().GetConfig()
	cfg := ty.GetTicketMinerParam(chain33Cfg, current.Block.Height)
	if current.Block.BlockTime-types.Now().Unix() > cfg.FutureBlockTime {
		return types.ErrFutureBlock
	}
	ticketAction, err := client.getMinerTx(current.Block)
	if err != nil {
		return err
	}
	if parent.Height+1 != current.Block.Height {
		return types.ErrBlockHeight
	}
	//判断exec 是否成功
	if current.Receipts[0].Ty != types.ExecOk {
		return types.ErrCoinBaseExecErr
	}
	//check reward 的值是否正确
	miner := ticketAction.GetMiner()
	if miner.Reward != (cfg.CoinReward + calcTotalFee(current.Block)) {
		return types.ErrCoinbaseReward
	}
	if miner.Bits != current.Block.Difficulty {
		return types.ErrBlockHeaderDifficulty
	}
	//check modify:

	//通过判断区块的难度Difficulty
	//1. target >= currentdiff
	//2.  current bit == target
	target, modify, err := client.getNextTarget(parent, parent.Difficulty)
	if err != nil {
		return err
	}
	if string(modify) != string(miner.Modify) {
		return ty.ErrModify
	}
	currentdiff := client.getCurrentTarget(current.Block.BlockTime, miner.TicketId, miner.Modify, miner.PrivHash)
	if currentdiff.Sign() < 0 {
		return types.ErrCoinBaseTarget
	}
	//当前难度
	currentTarget := difficulty.CompactToBig(current.Block.Difficulty)
	if currentTarget.Cmp(difficulty.CompactToBig(miner.Bits)) != 0 {
		tlog.Error("block error: calc tagget not the same to miner",
			"cacl", printBInt(currentTarget), "current", printBInt(difficulty.CompactToBig(miner.Bits)))
		return types.ErrCoinBaseTarget
	}
	if currentTarget.Cmp(target) != 0 {
		tlog.Error("block error: calc tagget not the same to target",
			"cacl", printBInt(currentTarget), "current", printBInt(target))
		return types.ErrCoinBaseTarget
	}
	if currentdiff.Cmp(currentTarget) > 0 {
		tlog.Error("block error: diff not fit the tagget",
			"current", printBInt(currentdiff), "taget", printBInt(target))
		return types.ErrCoinBaseTarget
	}
	if current.Block.Size() > int(types.MaxBlockSize) {
		return types.ErrBlockSize
	}
	//vrf verify
	if chain33Cfg.IsDappFork(current.Block.Height, ty.TicketX, "ForkTicketVrf") {
		var input []byte
		if current.Block.Height > 1 {
			LastTicketAction, err := client.getMinerTx(parent)
			if err != nil {
				return err
			}
			input = LastTicketAction.GetMiner().GetVrfHash()
		}
		if input == nil {
			input = miner.PrivHash
		}
		minerTx := current.Block.Txs[0]
		if err = vrfVerify(minerTx.Signature.Pubkey, input, miner.VrfProof, miner.VrfHash); err != nil {
			return err
		}
	} else {
		if len(miner.VrfHash) != 0 || len(miner.VrfProof) != 0 {
			tlog.Error("block error: not yet add vrf")
			return ty.ErrNoVrf
		}
	}
	return nil
}

func vrfVerify(pub []byte, input []byte, proof []byte, hash []byte) error {
	pubKey, err := secp256k1.ParsePubKey(pub, secp256k1.S256())
	if err != nil {
		tlog.Error("vrfVerify", "err", err)
		return ty.ErrVrfVerify
	}
	vrfPub := &vrf.PublicKey{PublicKey: (*ecdsa.PublicKey)(pubKey)}
	vrfHash, err := vrfPub.ProofToHash(input, proof)
	if err != nil {
		tlog.Error("vrfVerify", "err", err)
		return ty.ErrVrfVerify
	}
	tlog.Debug("vrf verify", "ProofToHash", fmt.Sprintf("(%x, %x): %x", input, proof, vrfHash), "hash", hex.EncodeToString(hash))
	if !bytes.Equal(vrfHash[:], hash) {
		tlog.Error("vrfVerify", "err", errors.New("invalid VRF hash"))
		return ty.ErrVrfVerify
	}
	return nil
}

func (client *Client) getNextTarget(block *types.Block, bits uint32) (*big.Int, []byte, error) {
	cfg := client.GetAPI().GetConfig()
	if block.Height == 0 {
		powLimit := difficulty.CompactToBig(cfg.GetP(0).PowLimitBits)
		return powLimit, defaultModify, nil
	}
	targetBits, modify, err := client.getNextRequiredDifficulty(block, bits)
	if err != nil {
		return nil, nil, err
	}
	return difficulty.CompactToBig(targetBits), modify, nil
}

func (client *Client) getCurrentTarget(blocktime int64, id string, modify []byte, privHash []byte) *big.Int {
	s := fmt.Sprintf("%d:%s:%x", blocktime, id, modify)
	if len(privHash) != 0 {
		s = s + ":" + string(privHash)
	}
	hash := common.Sha2Sum([]byte(s))
	num := difficulty.HashToBig(hash[:])
	return difficulty.CompactToBig(difficulty.BigToCompact(num))
}

// calcNextRequiredDifficulty calculates the required difficulty for the block
// after the passed previous block node based on the difficulty retarget rules.
// This function differs from the exported CalcNextRequiredDifficulty in that
// the exported version uses the current best chain as the previous block node
// while this function accepts any block node.
func (client *Client) getNextRequiredDifficulty(block *types.Block, bits uint32) (uint32, []byte, error) {
	chain33Cfg := client.GetAPI().GetConfig()
	// Genesis block.
	if block == nil {
		return chain33Cfg.GetP(0).PowLimitBits, defaultModify, nil
	}
	powLimitBits := chain33Cfg.GetP(block.Height).PowLimitBits
	cfg := ty.GetTicketMinerParam(chain33Cfg, block.Height)
	blocksPerRetarget := int64(cfg.TargetTimespan / cfg.TargetTimePerBlock)
	// Return the previous block's difficulty requirements if this block
	// is not at a difficulty retarget interval.
	if (block.Height+1) <= blocksPerRetarget || (block.Height+1)%blocksPerRetarget != 0 {
		// For the main network (or any unrecognized networks), simply
		// return the previous block's difficulty requirements.
		modify, err := client.getMinerModify(block)
		if err != nil {
			return bits, defaultModify, err
		}
		return bits, modify, nil
	}

	// Get the block node at the previous retarget (targetTimespan days
	// worth of blocks).
	firstBlock, err := client.RequestBlock(block.Height + 1 - blocksPerRetarget)
	if err != nil {
		return powLimitBits, defaultModify, err
	}
	if firstBlock == nil {
		return powLimitBits, defaultModify, types.ErrBlockNotFound
	}

	modify, err := client.getModify(block.Height+1-blocksPerRetarget, block.Height)
	if err != nil {
		return powLimitBits, defaultModify, err
	}
	// Limit the amount of adjustment that can occur to the previous
	// difficulty.
	actualTimespan := block.BlockTime - firstBlock.BlockTime
	adjustedTimespan := actualTimespan
	targetTimespan := int64(cfg.TargetTimespan / time.Second)

	minRetargetTimespan := targetTimespan / (cfg.RetargetAdjustmentFactor)
	maxRetargetTimespan := targetTimespan * cfg.RetargetAdjustmentFactor
	if actualTimespan < minRetargetTimespan {
		adjustedTimespan = minRetargetTimespan
	} else if actualTimespan > maxRetargetTimespan {
		adjustedTimespan = maxRetargetTimespan
	}

	// Calculate new target difficulty as:
	//  currentDifficulty * (adjustedTimespan / targetTimespan)
	// The result uses integer division which means it will be slightly
	// rounded down.  Bitcoind also uses integer division to calculate this
	// result.
	oldTarget := difficulty.CompactToBig(bits)
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	newTarget.Div(newTarget, big.NewInt(targetTimespan))

	// Limit new value to the proof of work limit.
	powLimit := difficulty.CompactToBig(powLimitBits)
	if newTarget.Cmp(powLimit) > 0 {
		newTarget.Set(powLimit)
	}

	// Log new target difficulty and return it.  The new target logging is
	// intentionally converting the bits back to a number instead of using
	// newTarget since conversion to the compact representation loses
	// precision.
	newTargetBits := difficulty.BigToCompact(newTarget)
	tlog.Info(fmt.Sprintf("Difficulty retarget at block height %d", block.Height+1))
	tlog.Info(fmt.Sprintf("Old target %08x, (%064x)", bits, oldTarget))
	tlog.Info(fmt.Sprintf("New target %08x, (%064x)", newTargetBits, difficulty.CompactToBig(newTargetBits)))
	tlog.Info("Timespan", "Actual timespan", time.Duration(actualTimespan)*time.Second,
		"adjusted timespan", time.Duration(adjustedTimespan)*time.Second,
		"target timespan", cfg.TargetTimespan)
	prevmodify, err := client.getMinerModify(block)
	if err != nil {
		panic(err)
	}
	tlog.Info("UpdateModify", "prev", string(prevmodify), "current", string(modify))
	return newTargetBits, modify, nil
}

func printBInt(data *big.Int) string {
	txt := data.Text(16)
	return strings.Repeat("0", 64-len(txt)) + txt
}

func (client *Client) searchTargetTicket(parent, block *types.Block) (*ty.Ticket, crypto.PrivKey, *big.Int, []byte, string, error) {
	cfg := client.GetAPI().GetConfig()
	bits := parent.Difficulty
	diff, modify, err := client.getNextTarget(parent, bits)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	client.ticketmu.Lock()
	defer client.ticketmu.Unlock()
	for ticketID, ticket := range client.ticketsMap {
		if client.IsClosed() {
			return nil, nil, nil, nil, "", nil
		}
		if ticket == nil {
			tlog.Warn("Client searchTargetTicket ticket is nil", "ticketID", ticketID)
			continue
		}
		//已经到成熟期
		if !ticket.GetIsGenesis() && (block.BlockTime-ticket.GetCreateTime() <= ty.GetTicketMinerParam(cfg, block.Height).TicketFrozenTime) {
			continue
		}
		// 查找私钥
		priv, ok := client.privmap[ticket.MinerAddress]
		if !ok {
			tlog.Error("Client searchTargetTicket can't find private key", "MinerAddress", ticket.MinerAddress)
			continue
		}

		privHash, err := genPrivHash(priv, ticketID)
		if err != nil {
			tlog.Error("Client searchTargetTicket genPrivHash ", "error", err)
			continue
		}
		currentdiff := client.getCurrentTarget(block.BlockTime, ticket.TicketId, modify, privHash)
		if currentdiff.Cmp(diff) >= 0 { //难度要大于前一个，注意数字越小难度越大
			continue
		}
		tlog.Info("currentdiff", "hex", printBInt(currentdiff))
		tlog.Info("FindBlock", "height------->", block.Height, "ntx", len(block.Txs))
		return ticket, priv, diff, modify, ticketID, nil
	}
	return nil, nil, nil, nil, "", nil
}

func (client *Client) delTicket(ticketID string) {
	client.ticketmu.Lock()
	defer client.ticketmu.Unlock()
	if client.ticketsMap == nil || len(ticketID) == 0 {
		return
	}
	if _, ok := client.ticketsMap[ticketID]; ok {
		delete(client.ticketsMap, ticketID)
	}
}

// Miner ticket miner function
func (client *Client) Miner(block *types.Block) error {
	//add miner address
	parentBlock := client.GetCurrentBlock()
	ticket, priv, diff, modify, ticketID, err := client.searchTargetTicket(parentBlock, block)
	if err != nil {
		tlog.Error("Miner", "err", err)
		lastBlock, err := client.RequestLastBlock()
		if err != nil {
			tlog.Error("Miner.RequestLastBlock", "err", err)
		}
		client.SetCurrentBlock(lastBlock)
		return err
	}
	if ticket == nil {
		return errors.New("ticket is nil")
	}
	newBlock := *block
	err = client.addMinerTx(parentBlock, &newBlock, diff, priv, ticket.TicketId, modify)
	if err != nil {
		return err
	}
	//需要首先对交易进行排序
	cfg := client.GetAPI().GetConfig()
	if cfg.IsFork(newBlock.Height, "ForkRootHash") {
		newBlock.Txs = types.TransactionSort(newBlock.Txs)
	}

	err = client.WriteBlock(parentBlock.StateHash, &newBlock)
	if err != nil {
		return err
	}
	client.delTicket(ticketID)
	return nil
}

//gas 直接燃烧
func calcTotalFee(block *types.Block) (total int64) {
	return 0
}

func genPrivHash(priv crypto.PrivKey, tid string) ([]byte, error) {
	var privHash []byte
	parts := strings.Split(tid, ":")
	if len(parts) > ty.TicketOldParts {
		count := parts[ty.TicketOldParts-1]
		seed := parts[len(parts)-1]

		var countNum int
		countNum, err := strconv.Atoi(count)
		if err != nil {
			return nil, err
		}
		privStr := fmt.Sprintf("%x:%d:%s", priv.Bytes(), countNum, seed)
		privHash = common.Sha256([]byte(privStr))
	}
	return privHash, nil
}

func (client *Client) addMinerTx(parent, block *types.Block, diff *big.Int, priv crypto.PrivKey, tid string, modify []byte) error {
	//return 0 always
	cfg := client.GetAPI().GetConfig()
	fee := calcTotalFee(block)

	var ticketAction ty.TicketAction
	miner := &ty.TicketMiner{}
	miner.TicketId = tid
	miner.Bits = difficulty.BigToCompact(diff)
	miner.Modify = modify
	miner.Reward = ty.GetTicketMinerParam(cfg, block.Height).CoinReward + fee
	privHash, err := genPrivHash(priv, tid)
	if err != nil {
		return err
	}
	miner.PrivHash = privHash
	//add vrf
	if cfg.IsDappFork(block.Height, ty.TicketX, "ForkTicketVrf") {
		var input []byte
		if block.Height > 1 {
			LastTicketAction, err := client.getMinerTx(parent)
			if err != nil {
				return err
			}
			input = LastTicketAction.GetMiner().GetVrfHash()
		}
		if input == nil {
			input = miner.PrivHash
		}
		privKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), priv.Bytes())
		vrfPriv := &vrf.PrivateKey{PrivateKey: (*ecdsa.PrivateKey)(privKey)}
		vrfHash, vrfProof := vrfPriv.Evaluate(input)
		miner.VrfHash = vrfHash[:]
		miner.VrfProof = vrfProof
	}

	ticketAction.Value = &ty.TicketAction_Miner{Miner: miner}
	ticketAction.Ty = ty.TicketActionMiner
	tickerMinerAddr := strings.Split(tid, ":")[0]
	var addressID = address.DefaultID
	if common.IsHex(tickerMinerAddr) {
		addressID = eth.ID
	}
	//构造transaction
	tx := client.createMinerTx(&ticketAction, priv, int32(addressID))
	//unshift
	if tx == nil {
		return ty.ErrEmptyMinerTx
	}
	block.Difficulty = miner.Bits
	//判断是替换还是append
	_, err = client.getMinerTx(block)
	if err != nil {
		block.Txs = append([]*types.Transaction{tx}, block.Txs...)
	} else {
		//ticket miner 交易已经存在
		block.Txs[0] = tx
	}
	return nil
}

func (client *Client) createMinerTx(ticketAction proto.Message, priv crypto.PrivKey, addressID int32) *types.Transaction {
	cfg := client.GetAPI().GetConfig()
	tx, err := types.CreateFormatTx(cfg, "ticket", types.Encode(ticketAction))
	if err != nil {
		return nil
	}
	tx.Sign(types.EncodeSignID(types.SECP256K1, addressID), priv)
	return tx
}

func (client *Client) createBlock() *types.Block {
	cfg := client.GetAPI().GetConfig()
	lastBlock := client.GetCurrentBlock()
	var newblock types.Block
	newblock.ParentHash = lastBlock.Hash(cfg)
	newblock.Height = lastBlock.Height + 1
	newblock.BlockTime = types.Now().Unix()
	if lastBlock.BlockTime >= newblock.BlockTime {
		newblock.BlockTime = lastBlock.BlockTime
	}
	txs := client.RequestTx(int(cfg.GetP(newblock.Height).MaxTxNumber)-1, nil)
	client.AddTxsToBlock(&newblock, txs)
	return &newblock
}

func (client *Client) updateBlock(block *types.Block, txHashList [][]byte) (txList [][]byte) {
	chain33Cfg := client.GetAPI().GetConfig()
	lastBlock := client.GetCurrentBlock()
	block.BlockTime = types.Now().Unix()

	//需要去重复tx并删除过期tx交易
	if lastBlock.Height != block.Height-1 {
		block.Txs = client.CheckTxDup(block.Txs)
		block.Txs = client.CheckTxExpire(block.Txs, lastBlock.Height+1, block.BlockTime)
	}
	block.ParentHash = lastBlock.Hash(chain33Cfg)
	block.Height = lastBlock.Height + 1
	cfg := chain33Cfg.GetP(block.Height)
	var txs []*types.Transaction
	if len(block.Txs) < int(cfg.MaxTxNumber-1) {
		txs = client.RequestTx(int(cfg.MaxTxNumber)-1-len(block.Txs), txHashList)
	}
	//tx 有更新
	if len(txs) > 0 {
		//防止区块过大
		txs = client.AddTxsToBlock(block, txs)
		if len(txs) > 0 {
			txHashList = append(txHashList, getTxHashes(txs)...)
		}
	}
	if lastBlock.BlockTime >= block.BlockTime {
		block.BlockTime = lastBlock.BlockTime + 1
	}
	return txHashList
}

// CreateBlock ticket create block func
func (client *Client) CreateBlock() {
	for {
		if client.IsClosed() {
			tlog.Info("create block stop")
			break
		}
		if !client.IsMining() || !(client.IsCaughtUp() || client.Cfg.ForceMining) {
			tlog.Debug("createblock.ismining is disable or client is caughtup is false")
			time.Sleep(time.Second)
			continue
		}
		if client.getTicketCount() == 0 {
			tlog.Debug("createblock.getticketcount = 0")
			time.Sleep(time.Second)
			continue
		}
		block := client.createBlock()
		txList := getTxHashes(block.Txs)
		for err := client.Miner(block); err != nil; err = client.Miner(block) {
			if err == queue.ErrIsQueueClosed {
				break
			}
			//加入新的txs, 继续挖矿
			lasttime := block.BlockTime
			//只有时间增加了1s影响，影响难度计算了，才会去更新区块
			for lasttime >= types.Now().Unix() {
				time.Sleep(time.Second / 10)
			}
			txList = client.updateBlock(block, txList)
		}
	}
}

func getTxHashes(txs []*types.Transaction) (hashes [][]byte) {
	hashes = make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		hashes[i] = txs[i].Hash()
	}
	return hashes
}

//CmpBestBlock 比较newBlock是不是最优区块，目前ticket主要是比较挖矿交易的难度系数
func (client *Client) CmpBestBlock(newBlock *types.Block, cmpBlock *types.Block) bool {
	cfg := client.GetAPI().GetConfig()

	//newblock挖矿交易的难度系数
	newBlockTicket, err := client.getMinerTx(newBlock)
	if err != nil {
		tlog.Error("CmpBestBlock:getMinerTx", "newBlockHash", common.ToHex(newBlock.Hash(cfg)))
		return false
	}
	newBlockMiner := newBlockTicket.GetMiner()
	newBlockDiff := client.getCurrentTarget(newBlock.BlockTime, newBlockMiner.TicketId, newBlockMiner.Modify, newBlockMiner.PrivHash)

	//cmpBlock挖矿交易的难度系数
	cmpBlockTicket, err := client.getMinerTx(cmpBlock)
	if err != nil {
		tlog.Error("CmpBestBlock:getMinerTx", "cmpBlockHash", common.ToHex(cmpBlock.Hash(cfg)))
		return false
	}
	cmpBlockMiner := cmpBlockTicket.GetMiner()
	cmpBlockDiff := client.getCurrentTarget(cmpBlock.BlockTime, cmpBlockMiner.TicketId, cmpBlockMiner.Modify, cmpBlockMiner.PrivHash)

	//数字越小难度越大
	return newBlockDiff.Cmp(cmpBlockDiff) < 0
}
