// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbft

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/qbft/types"
	tmtypes "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
)

// ID is a hex-encoded crypto.Address
type ID string

// Messages in channels are chopped into smaller msgPackets for multiplexing.
type msgPacket struct {
	TypeID byte
	Bytes  []byte
}

// MsgInfo struct
type MsgInfo struct {
	TypeID byte
	Msg    proto.Message
	PeerID ID
	PeerIP string
}

// Peer interface
type Peer interface {
	ID() ID
	RemoteIP() (net.IP, error) // remote IP of the connection
	RemoteAddr() (net.Addr, error)
	IsOutbound() bool
	IsPersistent() bool

	Send(msg MsgInfo) bool
	TrySend(msg MsgInfo) bool

	Stop()

	SetTransferChannel(chan MsgInfo)
}

// PeerConnState struct
type PeerConnState struct {
	mtx sync.Mutex
	ip  net.IP
	ttypes.PeerRoundState
}

type peerConn struct {
	outbound bool

	conn      net.Conn // source connection
	bufReader *bufio.Reader
	bufWriter *bufio.Writer

	persistent bool
	ip         net.IP
	id         ID

	sendQueue     chan MsgInfo
	sendQueueSize int32

	started uint32 //atomic
	stopped uint32 // atomic

	quitUpdate chan struct{}
	quitBeat   chan struct{}

	transferChannel chan MsgInfo

	sendBuffer []byte

	onPeerError func(Peer, interface{})

	myState *ConsensusState

	state            *PeerConnState
	updateStateQueue chan MsgInfo
	heartbeatQueue   chan proto.Message
}

// PeerSet struct
type PeerSet struct {
	mtx    sync.Mutex
	lookup map[ID]*peerSetItem
	list   []Peer
}

type peerSetItem struct {
	peer  Peer
	index int
}

// NewPeerSet method
func NewPeerSet() *PeerSet {
	return &PeerSet{
		lookup: make(map[ID]*peerSetItem),
		list:   make([]Peer, 0, 256),
	}
}

// Add adds the peer to the PeerSet.
// It returns an error carrying the reason, if the peer is already present.
func (ps *PeerSet) Add(peer Peer) error {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.lookup[peer.ID()] != nil {
		return fmt.Errorf("Duplicate peer ID %v", peer.ID())
	}

	index := len(ps.list)
	// Appending is safe even with other goroutines
	// iterating over the ps.list slice.
	ps.list = append(ps.list, peer)
	ps.lookup[peer.ID()] = &peerSetItem{peer, index}
	return nil
}

// Has returns true if the set contains the peer referred to by this
// peerKey, otherwise false.
func (ps *PeerSet) Has(peerKey ID) bool {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	_, ok := ps.lookup[peerKey]
	return ok
}

// HasIP returns true if the set contains the peer referred to by this IP
// address, otherwise false.
func (ps *PeerSet) HasIP(peerIP net.IP) bool {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	return ps.hasIP(peerIP)
}

// hasIP does not acquire a lock so it can be used in public methods which
// already lock.
func (ps *PeerSet) hasIP(peerIP net.IP) bool {
	for _, item := range ps.lookup {
		if ip, err := item.peer.RemoteIP(); err == nil && ip.Equal(peerIP) {
			return true
		}
	}

	return false
}

// GetIP get IP by ID
func (ps *PeerSet) GetIP(peerKey ID) net.IP {
	ps.mtx.Lock()
	ps.mtx.Unlock()

	if item, ok := ps.lookup[peerKey]; ok {
		ip, err := item.peer.RemoteIP()
		if err == nil {
			return ip
		}
	}
	return nil
}

// Size of list
func (ps *PeerSet) Size() int {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return len(ps.list)
}

// List returns the threadsafe list of peers.
func (ps *PeerSet) List() []Peer {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.list
}

// Remove discards peer by its Key, if the peer was previously memoized.
func (ps *PeerSet) Remove(peer Peer) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	item := ps.lookup[peer.ID()]
	if item == nil {
		return
	}

	index := item.index
	// Create a new copy of the list but with one less item.
	// (we must copy because we'll be mutating the list).
	newList := make([]Peer, len(ps.list)-1)
	copy(newList, ps.list)
	// If it's the last peer, that's an easy special case.
	if index == len(ps.list)-1 {
		ps.list = newList
		delete(ps.lookup, peer.ID())
		return
	}

	// Replace the popped item with the last item in the old list.
	lastPeer := ps.list[len(ps.list)-1]
	lastPeerKey := lastPeer.ID()
	lastPeerItem := ps.lookup[lastPeerKey]
	newList[index] = lastPeer
	lastPeerItem.index = index
	ps.list = newList
	delete(ps.lookup, peer.ID())
}

//-------------------------peer connection--------------------------------
func (pc *peerConn) ID() ID {
	if len(pc.id) != 0 {
		return pc.id
	}
	pc.id = GenIDByPubKey(pc.conn.(*SecretConnection).RemotePubKey())
	return pc.id
}

func (pc *peerConn) RemoteIP() (net.IP, error) {
	if pc.ip != nil && len(pc.ip) > 0 {
		return pc.ip, nil
	}

	// In test cases a conn could not be present at all or be an in-memory
	// implementation where we want to return a fake ip.
	if pc.conn == nil || pc.conn.RemoteAddr().String() == "pipe" {
		return nil, errors.New("connect is nil or just pipe")
	}

	host, _, err := net.SplitHostPort(pc.conn.RemoteAddr().String())
	if err != nil {
		panic(err)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		panic(err)
	}

	pc.ip = ips[0]

	return pc.ip, nil
}

func (pc *peerConn) RemoteAddr() (net.Addr, error) {
	if pc.conn == nil || pc.conn.RemoteAddr().String() == "pipe" {
		return nil, errors.New("connect is nil or just pipe")
	}

	return pc.conn.RemoteAddr(), nil
}

func (pc *peerConn) SetTransferChannel(transferChannel chan MsgInfo) {
	pc.transferChannel = transferChannel
}

func (pc *peerConn) String() string {
	return fmt.Sprintf("PeerConn{outbound:%v persistent:%v ip:%s id:%s started:%v stopped:%v}",
		pc.outbound, pc.persistent, pc.ip.String(), pc.id, atomic.LoadUint32(&pc.started), atomic.LoadUint32(&pc.stopped))
}

func (pc *peerConn) CloseConn() {
	err := pc.conn.Close() // nolint: errcheck
	if err != nil {
		qbftlog.Error("peerConn CloseConn failed", "err", err)
	}
}

func (pc *peerConn) HandshakeTimeout(
	ourNodeInfo NodeInfo,
	timeout time.Duration,
) (peerNodeInfo NodeInfo, err error) {
	peerNodeInfo = NodeInfo{}
	// Set deadline for handshake so we don't block forever on conn.ReadFull
	if err := pc.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return peerNodeInfo, errors.Wrap(err, "Error setting deadline")
	}

	var err1 error
	var err2 error
	Parallel(
		func() {
			info, err1 := json.Marshal(ourNodeInfo)
			if err1 != nil {
				qbftlog.Error("Peer handshake Marshal ourNodeInfo failed", "err", err1)
				return
			}
			frame := make([]byte, 4)
			binary.BigEndian.PutUint32(frame, uint32(len(info)))
			_, err1 = pc.conn.Write(frame)
			if err1 != nil {
				qbftlog.Error("Peer handshake write info size failed", "err", err1)
				return
			}
			_, err1 = pc.conn.Write(info[:])
			if err1 != nil {
				qbftlog.Error("Peer handshake write info failed", "err", err1)
				return
			}
		},
		func() {
			readBuffer := make([]byte, 4)
			_, err2 = io.ReadFull(pc.conn, readBuffer[:])
			if err2 != nil {
				qbftlog.Error("Peer handshake read info size failed", "err", err1)
				return
			}
			len := binary.BigEndian.Uint32(readBuffer)
			readBuffer = make([]byte, len)
			_, err2 = io.ReadFull(pc.conn, readBuffer[:])
			if err2 != nil {
				qbftlog.Error("Peer handshake read info failed", "err", err1)
				return
			}
			err2 = json.Unmarshal(readBuffer, &peerNodeInfo)
			if err2 != nil {
				qbftlog.Error("Peer handshake Unmarshal failed", "err", err1)
				return
			}
			qbftlog.Info("Peer handshake", "peerNodeInfo", peerNodeInfo)
		},
	)
	if err1 != nil {
		return peerNodeInfo, errors.Wrap(err1, "Error during handshake/write")
	}
	if err2 != nil {
		return peerNodeInfo, errors.Wrap(err2, "Error during handshake/read")
	}

	// Remove deadline
	if err := pc.conn.SetDeadline(time.Time{}); err != nil {
		return peerNodeInfo, errors.Wrap(err, "Error removing deadline")
	}

	return peerNodeInfo, nil
}

func (pc *peerConn) IsOutbound() bool {
	return pc.outbound
}

func (pc *peerConn) IsPersistent() bool {
	return pc.persistent
}

func (pc *peerConn) Send(msg MsgInfo) bool {
	if !pc.IsRunning() {
		return false
	}
	select {
	case pc.sendQueue <- msg:
		atomic.AddInt32(&pc.sendQueueSize, 1)
		return true
	case <-time.After(defaultSendTimeout):
		if msg.TypeID == ttypes.ProposalBlockID {
			if propBlock, ok := msg.Msg.(*tmtypes.QbftBlock); ok {
				msg.Msg = &tmtypes.QbftBlock{
					Header:     propBlock.Header,
					LastCommit: propBlock.LastCommit,
				}
			}
		}
		qbftlog.Error("send msg timeout", "peerip", msg.PeerIP, "msg", msg)
		return false
	}
}

func (pc *peerConn) TrySend(msg MsgInfo) bool {
	if !pc.IsRunning() {
		return false
	}
	select {
	case pc.sendQueue <- msg:
		atomic.AddInt32(&pc.sendQueueSize, 1)
		return true
	default:
		return false
	}
}

// PickSendVote picks a vote and sends it to the peer.
// Returns true if vote was sent.
func (pc *peerConn) PickSendVote(votes ttypes.VoteSetReader) bool {
	if UseAggSig() {
		if votes.GetAggVote() != nil {
			if aggVote, ok := pc.state.PickAggVoteToSend(votes); ok {
				msg := MsgInfo{TypeID: ttypes.AggVoteID, Msg: aggVote.QbftAggVote, PeerID: pc.id, PeerIP: pc.ip.String()}
				qbftlog.Debug("Sending aggregate vote message", "msg", msg)
				if pc.Send(msg) {
					pc.state.SetHasAggVote(aggVote)
					return true
				}
			}
			return false
		}
	}
	if vote, ok := pc.state.PickVoteToSend(votes); ok {
		if vote == nil {
			qbftlog.Warn("Pick nil vote", "aggVote", votes.GetAggVote(), "vote", votes)
			return false
		}
		msg := MsgInfo{TypeID: ttypes.VoteID, Msg: vote.QbftVote, PeerID: pc.id, PeerIP: pc.ip.String()}
		qbftlog.Debug("Sending vote message", "msg", msg)
		if pc.Send(msg) {
			pc.state.SetHasVote(vote)
			return true
		}
	}
	return false
}

func (pc *peerConn) IsRunning() bool {
	return atomic.LoadUint32(&pc.started) == 1 && atomic.LoadUint32(&pc.stopped) == 0
}

func (pc *peerConn) Start() error {
	if atomic.CompareAndSwapUint32(&pc.started, 0, 1) {
		if atomic.LoadUint32(&pc.stopped) == 1 {
			qbftlog.Error("peerConn already stopped", "peerIP", pc.ip.String())
			return nil
		}
		pc.bufReader = bufio.NewReaderSize(pc.conn, minReadBufferSize)
		pc.bufWriter = bufio.NewWriterSize(pc.conn, minWriteBufferSize)
		pc.sendQueue = make(chan MsgInfo, maxSendQueueSize)
		pc.sendBuffer = make([]byte, 0, MaxMsgPacketPayloadSize)
		pc.quitUpdate = make(chan struct{})
		pc.quitBeat = make(chan struct{})
		pc.state = &PeerConnState{ip: pc.ip, PeerRoundState: ttypes.PeerRoundState{
			Round:              -1,
			ProposalPOLRound:   -1,
			LastCommitRound:    -1,
			CatchupCommitRound: -1,
		}}
		pc.updateStateQueue = make(chan MsgInfo, maxSendQueueSize)
		pc.heartbeatQueue = make(chan proto.Message, 100)

		go pc.sendRoutine()
		go pc.recvRoutine()
		go pc.updateStateRoutine()
		go pc.heartbeatRoutine()

		go pc.gossipDataRoutine()
		go pc.gossipVotesRoutine()
		go pc.queryMaj23Routine()

	}
	return nil
}

func (pc *peerConn) Stop() {
	if atomic.CompareAndSwapUint32(&pc.stopped, 0, 1) {
		pc.CloseConn()
		qbftlog.Info("peerConn close connection", "peerIP", pc.ip.String())
	}
}

func (pc *peerConn) stopForError(r interface{}) {
	qbftlog.Error("peerConn recovered panic", "error", r, "peer", pc.ip.String())
	if pc.onPeerError != nil {
		pc.onPeerError(pc, r)
	} else {
		pc.Stop()
	}
}

// 数据压缩
func encodeMsg(msg types.Message, typeID byte) []byte {
	raw := types.Encode(msg)
	cmp := byte(0)
	if len(raw) > MaxMsgPacketPayloadSize {
		buf := make([]byte, 0)
		buf = snappy.Encode(buf, raw)
		cmp = byte(1)
		qbftlog.Info("compress large message", "old", len(raw), "new", len(buf))
		raw = buf
	}
	ebuf := make([]byte, len(raw)+6)
	ebuf[0] = typeID
	ebuf[1] = cmp
	bytelen := make([]byte, 4)
	binary.BigEndian.PutUint32(bytelen, uint32(len(raw)))
	copy(ebuf[2:6], bytelen)
	copy(ebuf[6:], raw)
	return ebuf
}

// 数据解压
func decodeMsg(msg []byte, cmp byte) ([]byte, error) {
	if cmp == byte(0) {
		return msg, nil
	}
	buf := make([]byte, 0)
	buf, err := snappy.Decode(buf, msg)
	if err != nil {
		return nil, err
	}
	qbftlog.Info("uncompress large message", "old", len(msg), "new", len(buf))
	return buf, nil
}

func (pc *peerConn) sendRoutine() {
FOR_LOOP:
	for {
		select {
		case msg := <-pc.sendQueue:
			raw := encodeMsg(msg.Msg, msg.TypeID)
			_, err := pc.bufWriter.Write(raw)
			if err != nil {
				qbftlog.Error("sendRoutine buffer write fail", "peer", pc, "err", err)
				pc.stopForError(err)
				break FOR_LOOP
			}
			err = pc.bufWriter.Flush()
			if err != nil {
				qbftlog.Error("sendRoutine buffer flush fail", "peer", pc, "err", err)
				pc.stopForError(err)
				break FOR_LOOP
			}
		}
	}
	qbftlog.Info("peerConn stop sendRoutine", "peerIP", pc.ip.String())
}

func (pc *peerConn) recvRoutine() {
FOR_LOOP:
	for {
		//typeID+cmp+msgLen+msg
		var buf [6]byte
		_, err := io.ReadFull(pc.bufReader, buf[:])
		if err != nil {
			qbftlog.Error("recvRoutine read byte fail", "peer", pc, "err", err)
			pc.stopForError(err)
			break FOR_LOOP
		}

		pkt := msgPacket{}
		pkt.TypeID = buf[0]
		cmp := buf[1]
		msgLen := binary.BigEndian.Uint32(buf[2:6])
		if msgLen <= 0 {
			qbftlog.Error("recvRoutine read invalid data", "msgLen", msgLen, "cmp", cmp, "peerIP", pc.ip.String())
			continue
		}
		buf2 := make([]byte, msgLen)
		_, err = io.ReadFull(pc.bufReader, buf2)
		if err != nil {
			qbftlog.Error("recvRoutine read data fail", "err", err, "peerIP", pc.ip.String())
			continue
		}
		buf3, err := decodeMsg(buf2, cmp)
		if err != nil {
			qbftlog.Error("recvRoutine decode msg fail", "err", err, "peerIP", pc.ip.String())
			continue
		}
		pkt.Bytes = buf3

		if v, ok := ttypes.MsgMap[pkt.TypeID]; ok {
			realMsg := reflect.New(v).Interface()
			err := proto.Unmarshal(pkt.Bytes, realMsg.(proto.Message))
			if err != nil {
				qbftlog.Error("recvRoutine Unmarshal data fail", "msgTy", pkt.TypeID, "msgLen", len(pkt.Bytes), "err", err, "peerIP", pc.ip.String())
				continue
			}
			if pc.transferChannel != nil && (pkt.TypeID == ttypes.ProposalID || pkt.TypeID == ttypes.VoteID ||
				pkt.TypeID == ttypes.ProposalBlockID || pkt.TypeID == ttypes.AggVoteID) {
				pc.transferChannel <- MsgInfo{pkt.TypeID, realMsg.(proto.Message), pc.ID(), pc.ip.String()}
				if pkt.TypeID == ttypes.ProposalID {
					proposal := realMsg.(*tmtypes.QbftProposal)
					qbftlog.Debug("Receiving proposal", "proposal-height", proposal.Height, "peerip", pc.ip.String())
					pc.state.SetHasProposal(proposal)
				} else if pkt.TypeID == ttypes.VoteID {
					vote := &ttypes.Vote{QbftVote: realMsg.(*tmtypes.QbftVote)}
					qbftlog.Debug("Receiving vote", "vote-height", vote.Height, "peerip", pc.ip.String())
					pc.state.SetHasVote(vote)
				} else if pkt.TypeID == ttypes.ProposalBlockID {
					block := &ttypes.QbftBlock{QbftBlock: realMsg.(*tmtypes.QbftBlock)}
					qbftlog.Debug("Receiving proposal block", "block-height", block.Header.Height, "peerip", pc.ip.String())
					pc.state.SetHasProposalBlock(block)
				} else if pkt.TypeID == ttypes.AggVoteID {
					aggVote := &ttypes.AggVote{QbftAggVote: realMsg.(*tmtypes.QbftAggVote)}
					qbftlog.Debug("Receiving aggregate vote", "aggVote-height", aggVote.Height, "peerip", pc.ip.String())
					pc.state.SetHasAggVote(aggVote)
				}
			} else if pkt.TypeID == ttypes.ProposalHeartbeatID {
				pc.heartbeatQueue <- realMsg.(*tmtypes.QbftHeartbeat)
			} else {
				pc.updateStateQueue <- MsgInfo{pkt.TypeID, realMsg.(proto.Message), pc.ID(), pc.ip.String()}
			}
		} else {
			qbftlog.Error("receive unknown message type", "type", pkt.TypeID, "peerIP", pc.ip.String())
		}
	}
	pc.quitUpdate <- struct{}{}
	pc.quitBeat <- struct{}{}
	qbftlog.Info("peerConn stop recvRoutine", "peerIP", pc.ip.String())
}

func (pc *peerConn) updateStateRoutine() {
FOR_LOOP:
	for {
		select {
		case <-pc.quitUpdate:
			break FOR_LOOP
		case msg := <-pc.updateStateQueue:
			typeID := msg.TypeID
			if typeID == ttypes.NewRoundStepID {
				pc.state.ApplyNewRoundStepMessage(msg.Msg.(*tmtypes.QbftNewRoundStepMsg))
			} else if typeID == ttypes.ValidBlockID {
				pc.state.ApplyValidBlockMessage(msg.Msg.(*tmtypes.QbftValidBlockMsg))
			} else if typeID == ttypes.HasVoteID {
				pc.state.ApplyHasVoteMessage(msg.Msg.(*tmtypes.QbftHasVoteMsg))
			} else if typeID == ttypes.VoteSetMaj23ID {
				tmp := msg.Msg.(*tmtypes.QbftVoteSetMaj23Msg)
				pc.myState.SetPeerMaj23(tmp.Height, int(tmp.Round), byte(tmp.Type), pc.id, tmp.BlockID)
				var myVotes *ttypes.BitArray
				switch byte(tmp.Type) {
				case ttypes.VoteTypePrevote:
					myVotes = pc.myState.GetPrevotesState(tmp.Height, int(tmp.Round), tmp.BlockID)
				case ttypes.VoteTypePrecommit:
					myVotes = pc.myState.GetPrecommitsState(tmp.Height, int(tmp.Round), tmp.BlockID)
				default:
					qbftlog.Error("Bad VoteSetBitsMessage field Type", "type", byte(tmp.Type))
					return
				}
				if myVotes != nil && myVotes.QbftBitArray != nil {
					voteSetBitMsg := &tmtypes.QbftVoteSetBitsMsg{
						Height:  tmp.Height,
						Round:   tmp.Round,
						Type:    tmp.Type,
						BlockID: tmp.BlockID,
						Votes:   myVotes.QbftBitArray,
					}
					pc.sendQueue <- MsgInfo{TypeID: ttypes.VoteSetBitsID, Msg: voteSetBitMsg, PeerID: pc.id, PeerIP: pc.ip.String()}
				}

			} else if typeID == ttypes.ProposalPOLID {
				pc.state.ApplyProposalPOLMessage(msg.Msg.(*tmtypes.QbftProposalPOLMsg))
			} else if typeID == ttypes.VoteSetBitsID {
				tmp := msg.Msg.(*tmtypes.QbftVoteSetBitsMsg)
				if pc.myState.Height == tmp.Height {
					var myVotes *ttypes.BitArray
					switch byte(tmp.Type) {
					case ttypes.VoteTypePrevote:
						myVotes = pc.myState.GetPrevotesState(tmp.Height, int(tmp.Round), tmp.BlockID)
					case ttypes.VoteTypePrecommit:
						myVotes = pc.myState.GetPrecommitsState(tmp.Height, int(tmp.Round), tmp.BlockID)
					default:
						qbftlog.Error("Bad VoteSetBitsMessage field Type", "type", byte(tmp.Type))
						return
					}
					pc.state.ApplyVoteSetBitsMessage(tmp, myVotes)
				} else {
					pc.state.ApplyVoteSetBitsMessage(tmp, nil)
				}
			} else {
				qbftlog.Error("Unknown message type in updateStateRoutine", "msg", msg)
			}
		}
	}
	close(pc.updateStateQueue)
	qbftlog.Info("peerConn stop updateStateRoutine", "peerIP", pc.ip.String())
}

func (pc *peerConn) heartbeatRoutine() {
FOR_LOOP:
	for {
		select {
		case <-pc.quitBeat:
			break FOR_LOOP
		case heartbeat := <-pc.heartbeatQueue:
			msg, ok := heartbeat.(*tmtypes.QbftHeartbeat)
			if ok {
				qbftlog.Debug("Received proposal heartbeat message",
					"height", msg.Height, "round", msg.Round, "sequence", msg.Sequence,
					"valIdx", msg.ValidatorIndex, "valAddr", fmt.Sprintf("%X", msg.ValidatorAddress))
			}
		}
	}
	close(pc.heartbeatQueue)
	qbftlog.Info("peerConn stop heartbeatRoutine", "peerIP", pc.ip.String())
}

func (pc *peerConn) gossipDataRoutine() {
OUTER_LOOP:
	for {
		// Manage disconnects from self or peer.
		if !pc.IsRunning() {
			qbftlog.Info("peerConn stop gossipDataRoutine", "peerIP", pc.ip.String())
			return
		}

		rs := pc.myState.GetRoundState()
		prs := pc.state.GetRoundState()

		// If the peer is on a previous height, help catch up.
		if (0 < prs.Height) && (prs.Height < rs.Height) {
			if prs.ProposalBlockHash != nil && !prs.ProposalBlock {
				proposalBlock := pc.myState.client.LoadProposalBlock(prs.Height)
				if proposalBlock == nil {
					qbftlog.Error("load proposal block fail", "selfHeight", rs.Height,
						"blockHeight", pc.myState.client.GetCurrentHeight())
					time.Sleep(pc.myState.PeerGossipSleep())
					continue OUTER_LOOP
				}
				newBlock := &ttypes.QbftBlock{QbftBlock: proposalBlock}
				if !newBlock.HashesTo(prs.ProposalBlockHash) {
					qbftlog.Error(fmt.Sprintf("Wrong proposal block hash. Expected %X, got %X", prs.ProposalBlockHash,
						newBlock.Hash()), "height", prs.Height)
					time.Sleep(pc.myState.PeerGossipSleep())
					continue OUTER_LOOP
				}
				msg := MsgInfo{TypeID: ttypes.ProposalBlockID, Msg: proposalBlock, PeerID: pc.id, PeerIP: pc.ip.String()}
				qbftlog.Info("Sending block for catchup", "peerIP", pc.ip.String(),
					"selfHeight", rs.Height, "peer(H/R/S)", fmt.Sprintf("%v/%v/%v", prs.Height, prs.Round, prs.Step),
					"block(H/R/hash)", fmt.Sprintf("%v/%v/%X", proposalBlock.Header.Height, proposalBlock.Header.Round, newBlock.Hash()))
				if pc.Send(msg) {
					pc.state.SetHasProposalBlock(newBlock)
				}
				continue OUTER_LOOP
			}
		}

		// If height and round don't match, sleep.
		if (rs.Height != prs.Height) || (rs.Round != prs.Round) {
			time.Sleep(pc.myState.PeerGossipSleep())
			continue OUTER_LOOP
		}

		// By here, height and round match.
		// Proposal block parts were already matched and sent if any were wanted.
		// (These can match on hash so the round doesn't matter)
		// Now consider sending other things, like the Proposal itself.

		// Send Proposal && ProposalPOL BitArray?
		if rs.Proposal != nil && !prs.Proposal {
			// Proposal: share the proposal metadata with peer.
			{
				msg := MsgInfo{TypeID: ttypes.ProposalID, Msg: rs.Proposal, PeerID: pc.id, PeerIP: pc.ip.String()}
				qbftlog.Debug(fmt.Sprintf("Sending proposal. Self state: %v/%v/%v", rs.Height, rs.Round, rs.Step),
					"peerip", pc.ip.String(), "proposal-height", rs.Proposal.Height, "proposal-round", rs.Proposal.Round)
				if pc.Send(msg) {
					pc.state.SetHasProposal(rs.Proposal)
				}
			}
			// ProposalPOL: lets peer know which POL votes we have so far.
			// Peer must receive ttypes.ProposalMessage first.
			// rs.Proposal was validated, so rs.Proposal.POLRound <= rs.Round,
			// so we definitely have rs.Votes.Prevotes(rs.Proposal.POLRound).
			if 0 <= rs.Proposal.POLRound {
				msg := MsgInfo{TypeID: ttypes.ProposalPOLID, Msg: &tmtypes.QbftProposalPOLMsg{
					Height:           rs.Height,
					ProposalPOLRound: rs.Proposal.POLRound,
					ProposalPOL:      rs.Votes.Prevotes(int(rs.Proposal.POLRound)).BitArray().QbftBitArray,
				}, PeerID: pc.id, PeerIP: pc.ip.String()}
				qbftlog.Debug("Sending POL", "height", prs.Height, "round", prs.Round)
				pc.Send(msg)
			}
			continue OUTER_LOOP
		}

		// Send proposal block
		if rs.Proposal != nil && prs.ProposalBlockHash != nil && bytes.Equal(rs.Proposal.Blockhash, prs.ProposalBlockHash) {
			if rs.ProposalBlock != nil && !prs.ProposalBlock {
				msg := MsgInfo{TypeID: ttypes.ProposalBlockID, Msg: rs.ProposalBlock.QbftBlock, PeerID: pc.id, PeerIP: pc.ip.String()}
				qbftlog.Debug(fmt.Sprintf("Sending proposal block. Self state: %v/%v/%v", rs.Height, rs.Round, rs.Step),
					"peerip", pc.ip.String(), "block-height", rs.ProposalBlock.Header.Height, "block-round", rs.ProposalBlock.Header.Round)
				if pc.Send(msg) {
					pc.state.SetHasProposalBlock(rs.ProposalBlock)
				}
				continue OUTER_LOOP
			}
		}

		// Nothing to do. Sleep.
		time.Sleep(pc.myState.PeerGossipSleep())
		continue OUTER_LOOP
	}
}

func (pc *peerConn) gossipVotesRoutine() {
	// Simple hack to throttle logs upon sleep.
	var sleeping = 0

OUTER_LOOP:
	for {
		// Manage disconnects from self or peer.
		if !pc.IsRunning() {
			qbftlog.Info("peerConn stop gossipVotesRoutine", "peerIP", pc.ip.String())
			return
		}

		rs := pc.myState.GetRoundState()
		prs := pc.state.GetRoundState()

		switch sleeping {
		case 1: // First sleep
			sleeping = 2
		case 2: // No more sleep
			sleeping = 0
		}

		// If height matches, then send LastCommit, Prevotes, Precommits.
		if rs.Height == prs.Height {
			if !UseAggSig() || gossipVotes.Load().(bool) {
				if pc.gossipVotesForHeight(rs, prs) {
					continue OUTER_LOOP
				}
			}
		}

		// Special catchup logic.
		// If peer is lagging by height 1, send LastCommit.
		if prs.Height != 0 && rs.Height == prs.Height+1 {
			if pc.PickSendVote(rs.LastCommit) {
				qbftlog.Debug("Picked rs.LastCommit to send", "peerip", pc.ip.String(), "height", prs.Height)
				continue OUTER_LOOP
			}
		}

		// Catchup logic
		// If peer is lagging by more than 1, send Commit.
		if prs.Height != 0 && rs.Height >= prs.Height+2 {
			// Load the block commit for prs.Height,
			// which contains precommit signatures for prs.Height.
			commit := pc.myState.client.LoadBlockCommit(prs.Height + 1)
			commitObj := &ttypes.Commit{QbftCommit: commit}
			if pc.PickSendVote(commitObj) {
				qbftlog.Info("Picked Catchup commit to send",
					"commit(H/R)", fmt.Sprintf("%v/%v", commitObj.Height(), commitObj.Round()),
					"BitArray", commitObj.BitArray().String(),
					"peerip", pc.ip.String(), "peer(H/R/S)", fmt.Sprintf("%v/%v/%v", prs.Height, prs.Round, prs.Step))
				continue OUTER_LOOP
			}
		}

		if sleeping == 0 {
			// We sent nothing. Sleep...
			sleeping = 1
			qbftlog.Debug("No votes to send, sleeping", "peerip", pc.ip.String(), "rs.Height", rs.Height, "prs.Height", prs.Height,
				"localPV", rs.Votes.Prevotes(rs.Round).BitArray(), "peerPV", prs.Prevotes,
				"localPC", rs.Votes.Precommits(rs.Round).BitArray(), "peerPC", prs.Precommits)
		} else if sleeping == 2 {
			// Continued sleep...
			sleeping = 1
		}

		time.Sleep(pc.myState.PeerGossipSleep())
		continue OUTER_LOOP
	}
}

func (pc *peerConn) gossipVotesForHeight(rs *ttypes.RoundState, prs *ttypes.PeerRoundState) bool {
	// If there are lastCommits to send...
	if prs.Step == ttypes.RoundStepNewHeight {
		if pc.PickSendVote(rs.LastCommit) {
			qbftlog.Debug("Picked rs.LastCommit to send", "peerip", pc.ip.String(),
				"peer(H/R)", fmt.Sprintf("%v/%v", prs.Height, prs.Round))
			return true
		}
	}
	// If there are POL prevotes to send...
	if prs.Step <= ttypes.RoundStepPropose && prs.Round != -1 && prs.Round <= rs.Round && prs.ProposalPOLRound != -1 {
		if polPrevotes := rs.Votes.Prevotes(prs.ProposalPOLRound); polPrevotes != nil {
			if pc.PickSendVote(polPrevotes) {
				qbftlog.Debug("Picked rs.Prevotes(prs.ProposalPOLRound) to send",
					"peerip", pc.ip.String(), "peer(H/R)", fmt.Sprintf("%v/%v", prs.Height, prs.Round),
					"POLRound", prs.ProposalPOLRound)
				return true
			}
		}
	}
	// If there are prevotes to send...
	if prs.Step <= ttypes.RoundStepPrevoteWait && prs.Round != -1 && prs.Round <= rs.Round {
		if pc.PickSendVote(rs.Votes.Prevotes(prs.Round)) {
			qbftlog.Debug("Picked rs.Prevotes(prs.Round) to send",
				"peerip", pc.ip.String(), "peer(H/R)", fmt.Sprintf("%v/%v", prs.Height, prs.Round))
			return true
		}
	}
	// If there are precommits to send...
	if prs.Step <= ttypes.RoundStepPrecommitWait && prs.Round != -1 && prs.Round <= rs.Round {
		if pc.PickSendVote(rs.Votes.Precommits(prs.Round)) {
			qbftlog.Debug("Picked rs.Precommits(prs.Round) to send",
				"peerip", pc.ip.String(), "peer(H/R)", fmt.Sprintf("%v/%v", prs.Height, prs.Round))
			return true
		}
	}
	// If there are prevotes to send...Needed because of validBlock mechanism
	if prs.Round != -1 && prs.Round <= rs.Round {
		if pc.PickSendVote(rs.Votes.Prevotes(prs.Round)) {
			qbftlog.Debug("Picked rs.Prevotes(prs.Round) to send",
				"peerip", pc.ip.String(), "peer(H/R)", fmt.Sprintf("%v/%v", prs.Height, prs.Round))
			return true
		}
	}
	// If there are POLPrevotes to send...
	if prs.ProposalPOLRound != -1 {
		if polPrevotes := rs.Votes.Prevotes(prs.ProposalPOLRound); polPrevotes != nil {
			if pc.PickSendVote(polPrevotes) {
				qbftlog.Debug("Picked rs.Prevotes(prs.ProposalPOLRound) to send",
					"peerip", pc.ip.String(), "round", prs.ProposalPOLRound)
				return true
			}
		}
	}
	return false
}

func (pc *peerConn) queryMaj23Routine() {
OUTER_LOOP:
	for {
		// Manage disconnects from self or peer.
		if !pc.IsRunning() {
			qbftlog.Info("peerConn stop queryMaj23Routine", "peerIP", pc.ip.String())
			return
		}

		// Maybe send Height/Round/Prevotes
		{
			rs := pc.myState.GetRoundState()
			prs := pc.state.GetRoundState()
			if rs.Height == prs.Height {
				if maj23, ok := rs.Votes.Prevotes(prs.Round).TwoThirdsMajority(); ok {
					msg := MsgInfo{TypeID: ttypes.VoteSetMaj23ID, Msg: &tmtypes.QbftVoteSetMaj23Msg{
						Height:  prs.Height,
						Round:   int32(prs.Round),
						Type:    int32(ttypes.VoteTypePrevote),
						BlockID: &maj23,
					}, PeerID: pc.id, PeerIP: pc.ip.String(),
					}
					pc.TrySend(msg)
					time.Sleep(pc.myState.PeerQueryMaj23Sleep())
				}
			}
		}

		// Maybe send Height/Round/Precommits
		{
			rs := pc.myState.GetRoundState()
			prs := pc.state.GetRoundState()
			if rs.Height == prs.Height {
				if maj23, ok := rs.Votes.Precommits(prs.Round).TwoThirdsMajority(); ok {
					msg := MsgInfo{TypeID: ttypes.VoteSetMaj23ID, Msg: &tmtypes.QbftVoteSetMaj23Msg{
						Height:  prs.Height,
						Round:   int32(prs.Round),
						Type:    int32(ttypes.VoteTypePrecommit),
						BlockID: &maj23,
					}, PeerID: pc.id, PeerIP: pc.ip.String(),
					}
					pc.TrySend(msg)
					time.Sleep(pc.myState.PeerQueryMaj23Sleep())
				}
			}
		}

		// Maybe send Height/Round/ProposalPOL
		{
			rs := pc.myState.GetRoundState()
			prs := pc.state.GetRoundState()
			if rs.Height == prs.Height && prs.ProposalPOLRound >= 0 {
				if maj23, ok := rs.Votes.Prevotes(prs.ProposalPOLRound).TwoThirdsMajority(); ok {
					msg := MsgInfo{TypeID: ttypes.VoteSetMaj23ID, Msg: &tmtypes.QbftVoteSetMaj23Msg{
						Height:  prs.Height,
						Round:   int32(prs.ProposalPOLRound),
						Type:    int32(ttypes.VoteTypePrevote),
						BlockID: &maj23,
					}, PeerID: pc.id, PeerIP: pc.ip.String(),
					}
					pc.TrySend(msg)
					time.Sleep(pc.myState.PeerQueryMaj23Sleep())
				}
			}
		}

		// Little point sending LastCommitRound/LastCommit,
		// These are fleeting and non-blocking.

		// Maybe send Height/CatchupCommitRound/CatchupCommit.
		{
			prs := pc.state.GetRoundState()
			if prs.CatchupCommitRound != -1 && 0 < prs.Height && prs.Height <= pc.myState.client.csStore.LoadStateHeight() {
				commit := pc.myState.LoadCommit(prs.Height)
				if commit.VoteType == uint32(ttypes.VoteTypePrecommit) {
					commitTmp := ttypes.Commit{QbftCommit: commit}
					msg := MsgInfo{TypeID: ttypes.VoteSetMaj23ID, Msg: &tmtypes.QbftVoteSetMaj23Msg{
						Height:  prs.Height,
						Round:   int32(commitTmp.Round()),
						Type:    int32(ttypes.VoteTypePrecommit),
						BlockID: commit.BlockID,
					}, PeerID: pc.id, PeerIP: pc.ip.String(),
					}
					pc.TrySend(msg)
					time.Sleep(pc.myState.PeerQueryMaj23Sleep())
				}
			}
		}

		time.Sleep(pc.myState.PeerQueryMaj23Sleep())

		continue OUTER_LOOP
	}
}

// GetRoundState returns an atomic snapshot of the PeerRoundState.
// There's no point in mutating it since it won't change PeerState.
func (ps *PeerConnState) GetRoundState() *ttypes.PeerRoundState {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	prs := ps.PeerRoundState // copy
	return &prs
}

// GetHeight returns an atomic snapshot of the PeerRoundState's height
// used by the mempool to ensure peers are caught up before broadcasting new txs
func (ps *PeerConnState) GetHeight() int64 {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.PeerRoundState.Height
}

// SetHasProposal sets the given proposal as known for the peer.
func (ps *PeerConnState) SetHasProposal(proposal *tmtypes.QbftProposal) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != proposal.Height || ps.Round != int(proposal.Round) {
		return
	}
	if ps.Proposal {
		return
	}
	qbftlog.Debug("Peer set proposal", "peerIP", ps.ip.String(),
		"peer-state", fmt.Sprintf("%v/%v/%v", ps.Height, ps.Round, ps.Step),
		"proposal(H/R/Hash)", fmt.Sprintf("%v/%v/%X", proposal.Height, proposal.Round, proposal.Blockhash))
	ps.Proposal = true

	ps.ProposalBlockHash = proposal.Blockhash
	ps.ProposalPOLRound = int(proposal.POLRound)
	ps.ProposalPOL = nil // Nil until ttypes.ProposalPOLMessage received.
}

// SetHasProposalBlock sets the given proposal block as known for the peer.
func (ps *PeerConnState) SetHasProposalBlock(block *ttypes.QbftBlock) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != block.Header.Height ||
		(ps.Round != int(block.Header.Round) && ps.CatchupCommitRound != int(block.Header.Round)) {
		return
	}
	if ps.ProposalBlock {
		return
	}
	qbftlog.Debug("Peer set proposal block", "peerIP", ps.ip.String(),
		"peer-state", fmt.Sprintf("%v/%v(%v)/%v", ps.Height, ps.Round, ps.CatchupCommitRound, ps.Step),
		"block(H/R)", fmt.Sprintf("%v/%v", block.Header.Height, block.Header.Round))
	ps.ProposalBlock = true
}

// SetHasAggVote sets the given aggregate precommit as known for the peer.
func (ps *PeerConnState) SetHasAggVote(aggVote *ttypes.AggVote) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != aggVote.Height ||
		(ps.Round != int(aggVote.Round) && ps.CatchupCommitRound != int(aggVote.Round)) {
		return
	}
	if (aggVote.Type == uint32(ttypes.VoteTypePrevote) && ps.AggPrevote) ||
		(aggVote.Type == uint32(ttypes.VoteTypePrecommit) && ps.AggPrecommit) {
		return
	}
	qbftlog.Debug("Peer set aggregate vote", "peerIP", ps.ip.String(),
		"peer-state", fmt.Sprintf("%v/%v(%v)/%v", ps.Height, ps.Round, ps.CatchupCommitRound, ps.Step),
		"aggVote(H/R/T)", fmt.Sprintf("%v/%v/%v", aggVote.Height, aggVote.Round, aggVote.Type))
	if aggVote.Type == uint32(ttypes.VoteTypePrevote) {
		ps.AggPrevote = true
	} else if aggVote.Type == uint32(ttypes.VoteTypePrecommit) {
		ps.AggPrecommit = true
	}
}

// PickAggVoteToSend picks aggregate vote to send to the peer.
// Returns true if a vote was picked.
func (ps *PeerConnState) PickAggVoteToSend(votes ttypes.VoteSetReader) (vote *ttypes.AggVote, ok bool) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	aggVote := votes.GetAggVote()
	if aggVote == nil {
		return nil, false
	}
	if votes.IsCommit() {
		ps.ensureCatchupCommitRound(votes.Height(), votes.Round(), votes.Size())
	}
	if ps.Height != aggVote.Height ||
		(ps.Round != int(aggVote.Round) && ps.CatchupCommitRound != int(aggVote.Round)) {
		return nil, false
	}
	if (aggVote.Type == uint32(ttypes.VoteTypePrevote) && ps.AggPrevote) ||
		(aggVote.Type == uint32(ttypes.VoteTypePrecommit) && ps.AggPrecommit) {
		return nil, false
	}
	return aggVote, true
}

// PickVoteToSend picks a vote to send to the peer.
// Returns true if a vote was picked.
// NOTE: `votes` must be the correct Size() for the Height().
func (ps *PeerConnState) PickVoteToSend(votes ttypes.VoteSetReader) (vote *ttypes.Vote, ok bool) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if votes.Size() == 0 {
		return nil, false
	}

	height, round, voteType, size := votes.Height(), votes.Round(), votes.Type(), votes.Size()

	// Lazily set data using 'votes'.
	if votes.IsCommit() {
		ps.ensureCatchupCommitRound(height, round, size)
	}
	ps.ensureVoteBitArrays(height, size)

	psVotes := ps.getVoteBitArray(height, round, voteType)
	if psVotes == nil {
		return nil, false // Not something worth sending
	}

	if index, ok := votes.BitArray().Sub(psVotes).PickRandom(); ok {
		qbftlog.Debug("PickVoteToSend", "peer(H/R)", fmt.Sprintf("%v/%v", ps.Height, ps.Round),
			"vote(H/R)", fmt.Sprintf("%v/%v", height, round), "type", voteType, "selfVotes", votes.BitArray().String(),
			"peerVotes", psVotes.String(), "peerIP", ps.ip.String())
		return votes.GetByIndex(index), true
	}
	return nil, false
}

func (ps *PeerConnState) getVoteBitArray(height int64, round int, voteType byte) *ttypes.BitArray {
	if !ttypes.IsVoteTypeValid(voteType) {
		return nil
	}

	if ps.Height == height {
		if ps.Round == round {
			switch voteType {
			case ttypes.VoteTypePrevote:
				return ps.Prevotes
			case ttypes.VoteTypePrecommit:
				return ps.Precommits
			}
		}
		if ps.CatchupCommitRound == round {
			switch voteType {
			case ttypes.VoteTypePrevote:
				return nil
			case ttypes.VoteTypePrecommit:
				return ps.CatchupCommit
			}
		}
		if ps.ProposalPOLRound == round {
			switch voteType {
			case ttypes.VoteTypePrevote:
				return ps.ProposalPOL
			case ttypes.VoteTypePrecommit:
				return nil
			}
		}
		return nil
	}
	if ps.Height == height+1 {
		if ps.LastCommitRound == round {
			switch voteType {
			case ttypes.VoteTypePrevote:
				return nil
			case ttypes.VoteTypePrecommit:
				return ps.LastCommit
			}
		}
		return nil
	}
	return nil
}

// 'round': A round for which we have a +2/3 commit.
func (ps *PeerConnState) ensureCatchupCommitRound(height int64, round int, numValidators int) {
	if ps.Height != height {
		return
	}
	/*
		NOTE: This is wrong, 'round' could change.
		e.g. if orig round is not the same as block LastCommit round.
		if ps.CatchupCommitRound != -1 && ps.CatchupCommitRound != round {
			ttypes.PanicSanity(ttypes.Fmt("Conflicting CatchupCommitRound. Height: %v, Orig: %v, New: %v", height, ps.CatchupCommitRound, round))
		}
	*/
	if ps.CatchupCommitRound == round {
		return // Nothing to do!
	}
	qbftlog.Debug("ensureCatchupCommitRound", "height", height, "round", round, "ps.CatchupCommitRound", ps.CatchupCommitRound,
		"ps.Round", ps.Round, "peerip", ps.ip.String())
	ps.CatchupCommitRound = round
	if round == ps.Round {
		ps.CatchupCommit = ps.Precommits
	} else {
		ps.CatchupCommit = ttypes.NewBitArray(numValidators)
	}
}

// EnsureVoteBitArrays ensures the bit-arrays have been allocated for tracking
// what votes this peer has received.
// NOTE: It's important to make sure that numValidators actually matches
// what the node sees as the number of validators for height.
func (ps *PeerConnState) EnsureVoteBitArrays(height int64, numValidators int) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	ps.ensureVoteBitArrays(height, numValidators)
}

func (ps *PeerConnState) ensureVoteBitArrays(height int64, numValidators int) {
	if ps.Height == height {
		if ps.Prevotes == nil {
			ps.Prevotes = ttypes.NewBitArray(numValidators)
		}
		if ps.Precommits == nil {
			ps.Precommits = ttypes.NewBitArray(numValidators)
		}
		if ps.CatchupCommit == nil {
			ps.CatchupCommit = ttypes.NewBitArray(numValidators)
		}
		if ps.ProposalPOL == nil {
			ps.ProposalPOL = ttypes.NewBitArray(numValidators)
		}
	} else if ps.Height == height+1 {
		if ps.LastCommit == nil {
			ps.LastCommit = ttypes.NewBitArray(numValidators)
		}
	}
}

// SetHasVote sets the given vote as known by the peer
func (ps *PeerConnState) SetHasVote(vote *ttypes.Vote) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	ps.setHasVote(vote.Height, int(vote.Round), byte(vote.Type), int(vote.ValidatorIndex))
}

func (ps *PeerConnState) setHasVote(height int64, round int, voteType byte, index int) {
	// NOTE: some may be nil BitArrays -> no side effects.
	psVotes := ps.getVoteBitArray(height, round, voteType)
	qbftlog.Debug("setHasVote before", "height", height, "psVotes", psVotes.String(), "peerip", ps.ip.String())
	if psVotes != nil {
		psVotes.SetIndex(index, true)
	}
	qbftlog.Debug("setHasVote after", "height", height, "index", index, "type", voteType,
		"peerVotes", psVotes.String(), "peerip", ps.ip.String())
}

// ApplyNewRoundStepMessage updates the peer state for the new round.
func (ps *PeerConnState) ApplyNewRoundStepMessage(msg *tmtypes.QbftNewRoundStepMsg) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	// Ignore duplicates or decreases
	if CompareHRS(msg.Height, int(msg.Round), ttypes.RoundStepType(msg.Step), ps.Height, ps.Round, ps.Step) <= 0 {
		return
	}

	// Just remember these values.
	psHeight := ps.Height
	psRound := ps.Round
	//psStep := ps.Step
	psCatchupCommitRound := ps.CatchupCommitRound
	psCatchupCommit := ps.CatchupCommit
	psPrecommits := ps.Precommits

	startTime := time.Now().Add(-1 * time.Duration(msg.SecondsSinceStartTime) * time.Second)
	ps.Height = msg.Height
	ps.Round = int(msg.Round)
	ps.Step = ttypes.RoundStepType(msg.Step)
	ps.StartTime = startTime

	qbftlog.Debug("ApplyNewRoundStepMessage", "peerip", ps.ip.String(),
		"peer(H/R)", fmt.Sprintf("%v/%v", psHeight, psRound),
		"msg(H/R/S)", fmt.Sprintf("%v/%v/%v", msg.Height, msg.Round, ps.Step))

	if psHeight != msg.Height || psRound != int(msg.Round) {
		qbftlog.Debug("Reset QbftProposal, Prevotes, Precommits", "peerip", ps.ip.String(),
			"peer(H/R)", fmt.Sprintf("%v/%v", psHeight, psRound))
		ps.Proposal = false
		ps.ProposalBlock = false
		ps.ProposalBlockHash = nil
		ps.ProposalPOLRound = -1
		ps.ProposalPOL = nil
		// We'll update the BitArray capacity later.
		ps.Prevotes = nil
		ps.Precommits = nil
		ps.AggPrevote = false
		ps.AggPrecommit = false
	}
	if psHeight == msg.Height && psRound != int(msg.Round) && int(msg.Round) == psCatchupCommitRound {
		// Peer caught up to CatchupCommitRound.
		// Preserve psCatchupCommit!
		// NOTE: We prefer to use prs.Precommits if
		// pr.Round matches pr.CatchupCommitRound.
		qbftlog.Debug("Reset Precommits to CatchupCommit", "peerip", ps.ip.String(),
			"peer(H/R)", fmt.Sprintf("%v/%v", psHeight, psRound))
		ps.Precommits = psCatchupCommit
	}
	if psHeight != msg.Height {
		qbftlog.Debug("Reset LastCommit, CatchupCommit", "peerip", ps.ip.String(),
			"peer(H/R)", fmt.Sprintf("%v/%v", psHeight, psRound))
		// Shift Precommits to LastCommit.
		if psHeight+1 == msg.Height && psRound == int(msg.LastCommitRound) {
			ps.LastCommitRound = int(msg.LastCommitRound)
			ps.LastCommit = psPrecommits
		} else {
			ps.LastCommitRound = int(msg.LastCommitRound)
			ps.LastCommit = nil
		}
		// We'll update the BitArray capacity later.
		ps.CatchupCommitRound = -1
		ps.CatchupCommit = nil
	}
}

// ApplyValidBlockMessage updates the peer state for the new valid block.
func (ps *PeerConnState) ApplyValidBlockMessage(msg *tmtypes.QbftValidBlockMsg) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != msg.Height {
		return
	}
	if ps.Round != int(msg.Round) && !msg.IsCommit {
		return
	}
	qbftlog.Debug("ApplyValidBlockMessage", "peerIP", ps.ip.String(),
		"peer(H/R/S)", fmt.Sprintf("%v/%v/%v", ps.Height, ps.Round, ps.Step),
		"blockhash", fmt.Sprintf("%X", msg.Blockhash))

	ps.ProposalBlockHash = msg.Blockhash
	if ps.CatchupCommitRound == int(msg.Round) && msg.IsCommit {
		qbftlog.Info("Set ProposalBlockHash for catchup", "peerIP", ps.ip.String(),
			"peer(H/R/S)", fmt.Sprintf("%v/%v/%v", ps.Height, ps.Round, ps.Step),
			"CommitRound", ps.CatchupCommitRound,
			"ProposalBlockHash", fmt.Sprintf("%X", ps.ProposalBlockHash))
	}
}

// ApplyProposalPOLMessage updates the peer state for the new proposal POL.
func (ps *PeerConnState) ApplyProposalPOLMessage(msg *tmtypes.QbftProposalPOLMsg) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != msg.Height {
		return
	}
	if ps.ProposalPOLRound != int(msg.ProposalPOLRound) {
		return
	}

	// TODO: Merge onto existing ps.ProposalPOL?
	// We might have sent some prevotes in the meantime.
	ps.ProposalPOL = &ttypes.BitArray{QbftBitArray: msg.ProposalPOL}
}

// ApplyHasVoteMessage updates the peer state for the new vote.
func (ps *PeerConnState) ApplyHasVoteMessage(msg *tmtypes.QbftHasVoteMsg) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.Height != msg.Height {
		return
	}

	qbftlog.Debug("ApplyHasVoteMessage", "msg(H/R)", fmt.Sprintf("%v/%v", msg.Height, msg.Round),
		"peerip", ps.ip.String())
	ps.setHasVote(msg.Height, int(msg.Round), byte(msg.Type), int(msg.Index))
}

// ApplyVoteSetBitsMessage updates the peer state for the bit-array of votes
// it claims to have for the corresponding BlockID.
// `ourVotes` is a BitArray of votes we have for msg.BlockID
// NOTE: if ourVotes is nil (e.g. msg.Height < rs.Height),
// we conservatively overwrite ps's votes w/ msg.Votes.
func (ps *PeerConnState) ApplyVoteSetBitsMessage(msg *tmtypes.QbftVoteSetBitsMsg, ourVotes *ttypes.BitArray) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	votes := ps.getVoteBitArray(msg.Height, int(msg.Round), byte(msg.Type))
	if votes != nil {
		if ourVotes == nil {
			bitarray := &ttypes.BitArray{QbftBitArray: msg.Votes}
			votes.Update(bitarray)
		} else {
			otherVotes := votes.Sub(ourVotes)
			bitarray := &ttypes.BitArray{QbftBitArray: msg.Votes}
			hasVotes := otherVotes.Or(bitarray)
			votes.Update(hasVotes)
		}
	}
}
