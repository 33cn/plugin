// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	"github.com/golang/protobuf/proto"
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
	//Set(string, interface{})
	//Get(string) interface{}
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
	pongChannel   chan struct{}

	started uint32 //atomic
	stopped uint32 // atomic

	quit     chan struct{}
	waitQuit sync.WaitGroup

	transferChannel chan MsgInfo

	sendBuffer []byte

	onPeerError func(Peer, interface{})

	myState *ConsensusState
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
	_, ok := ps.lookup[peerKey]
	ps.mtx.Unlock()
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
	address := GenAddressByPubKey(pc.conn.(*SecretConnection).RemotePubKey())
	pc.id = ID(hex.EncodeToString(address))
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

func (pc *peerConn) CloseConn() {
	err := pc.conn.Close() // nolint: errcheck
	if err != nil {
		dposlog.Error("peerConn CloseConn failed", "err", err)
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
				dposlog.Error("Peer handshake Marshal ourNodeInfo failed", "err", err1)
				return
			}
			frame := make([]byte, 4)
			binary.BigEndian.PutUint32(frame, uint32(len(info)))
			_, err1 = pc.conn.Write(frame)
			if err1 != nil {
				dposlog.Error("Peer handshake write info size failed", "err", err1)
				return
			}
			_, err1 = pc.conn.Write(info[:])
			if err1 != nil {
				dposlog.Error("Peer handshake write info failed", "err", err1)
				return
			}
		},
		func() {
			readBuffer := make([]byte, 4)
			_, err2 = io.ReadFull(pc.conn, readBuffer[:])
			if err2 != nil {
				dposlog.Error("Peer handshake read info size failed", "err", err1)
				return
			}
			len := binary.BigEndian.Uint32(readBuffer)
			readBuffer = make([]byte, len)
			_, err2 = io.ReadFull(pc.conn, readBuffer[:])
			if err2 != nil {
				dposlog.Error("Peer handshake read info failed", "err", err1)
				return
			}
			err2 = json.Unmarshal(readBuffer, &peerNodeInfo)
			if err2 != nil {
				dposlog.Error("Peer handshake Unmarshal failed", "err", err1)
				return
			}
			dposlog.Info("Peer handshake", "peerNodeInfo", peerNodeInfo)
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
		dposlog.Error("send msg timeout", "peerip", msg.PeerIP, "msg", msg.Msg)
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

func (pc *peerConn) IsRunning() bool {
	return atomic.LoadUint32(&pc.started) == 1 && atomic.LoadUint32(&pc.stopped) == 0
}

func (pc *peerConn) Start() error {
	if atomic.CompareAndSwapUint32(&pc.started, 0, 1) {
		if atomic.LoadUint32(&pc.stopped) == 1 {
			dposlog.Error("peerConn already stoped can not start", "peerIP", pc.ip.String())
			return nil
		}
		pc.bufReader = bufio.NewReaderSize(pc.conn, minReadBufferSize)
		pc.bufWriter = bufio.NewWriterSize(pc.conn, minWriteBufferSize)
		pc.pongChannel = make(chan struct{})
		pc.sendQueue = make(chan MsgInfo, maxSendQueueSize)
		pc.sendBuffer = make([]byte, 0, MaxMsgPacketPayloadSize)
		pc.quit = make(chan struct{})
		pc.waitQuit.Add(5) //sendRoutine, updateStateRoutine,gossipDataRoutine,gossipVotesRoutine,queryMaj23Routine

		go pc.sendRoutine()
		go pc.recvRoutine()
	}
	return nil
}

func (pc *peerConn) Stop() {
	if atomic.CompareAndSwapUint32(&pc.stopped, 0, 1) {
		if pc.quit != nil {
			close(pc.quit)
			dposlog.Info("peerConn stop quit wait", "peerIP", pc.ip.String())
			pc.waitQuit.Wait()
			dposlog.Info("peerConn stop quit wait finish", "peerIP", pc.ip.String())
			pc.quit = nil
		}
		close(pc.sendQueue)
		pc.transferChannel = nil
		pc.CloseConn()
	}
}

// Catch panics, usually caused by remote disconnects.
func (pc *peerConn) _recover() {
	if r := recover(); r != nil {
		stack := debug.Stack()
		err := StackError{r, stack}
		pc.stopForError(err)
	}
}

func (pc *peerConn) stopForError(r interface{}) {
	dposlog.Error("peerConn recovered panic", "error", r, "peer", pc.ip.String())
	if pc.onPeerError != nil {
		pc.onPeerError(pc, r)
	}
	pc.Stop()
}

func (pc *peerConn) sendRoutine() {
	defer pc._recover()
FOR_LOOP:
	for {
		select {
		case <-pc.quit:
			pc.waitQuit.Done()
			break FOR_LOOP
		case msg := <-pc.sendQueue:
			bytes, err := proto.Marshal(msg.Msg)
			if err != nil {
				dposlog.Error("peerConn sendroutine marshal data failed", "error", err)
				pc.stopForError(err)
				break FOR_LOOP
			}

			len := len(bytes)
			bytelen := make([]byte, 4)
			binary.BigEndian.PutUint32(bytelen, uint32(len))
			pc.sendBuffer = pc.sendBuffer[:0]
			pc.sendBuffer = append(pc.sendBuffer, msg.TypeID)
			pc.sendBuffer = append(pc.sendBuffer, bytelen...)

			pc.sendBuffer = append(pc.sendBuffer, bytes...)
			if len+5 > MaxMsgPacketPayloadSize {
				pc.sendBuffer = append(pc.sendBuffer, bytes[MaxMsgPacketPayloadSize-5:]...)
			}
			_, err = pc.bufWriter.Write(pc.sendBuffer[:len+5])
			if err != nil {
				dposlog.Error("peerConn sendroutine write data failed", "error", err)
				pc.stopForError(err)
				break FOR_LOOP
			}
			err = pc.bufWriter.Flush()
			if err != nil {
				dposlog.Error("peerConn sendroutine flush buffer failed", "error", err)
				pc.stopForError(err)
				break FOR_LOOP
			}
		case _, ok := <-pc.pongChannel:
			if ok {
				dposlog.Debug("Send Pong")
				var pong [5]byte
				pong[0] = ttypes.PacketTypePong
				_, err := pc.bufWriter.Write(pong[:])
				if err != nil {
					dposlog.Error("peerConn sendroutine write pong failed", "error", err)
					pc.stopForError(err)
					break FOR_LOOP
				}
			} else {
				pc.pongChannel = nil
			}
		}
	}
}

func (pc *peerConn) recvRoutine() {
	defer pc._recover()
FOR_LOOP:
	for {
		//typeID+msgLen+msg
		var buf [5]byte
		_, err := io.ReadFull(pc.bufReader, buf[:])
		if err != nil {
			dposlog.Error("Connection failed @ recvRoutine (reading byte)", "conn", pc, "err", err)
			pc.stopForError(err)
			break FOR_LOOP
		}
		pkt := msgPacket{}
		pkt.TypeID = buf[0]
		len := binary.BigEndian.Uint32(buf[1:])
		if len > 0 {
			buf2 := make([]byte, len)
			_, err = io.ReadFull(pc.bufReader, buf2)
			if err != nil {
				dposlog.Error("Connection failed @ recvRoutine", "conn", pc, "err", err)
				pc.stopForError(err)
				panic(fmt.Sprintf("peerConn recvRoutine packetTypeMsg failed :%v", err))
			}
			pkt.Bytes = buf2
		}

		if pkt.TypeID == ttypes.PacketTypePong {
			dposlog.Debug("Receive Pong")
		} else if pkt.TypeID == ttypes.PacketTypePing {
			dposlog.Debug("Receive Ping")
			pc.pongChannel <- struct{}{}
		} else {
			if v, ok := ttypes.MsgMap[pkt.TypeID]; ok {
				realMsg := reflect.New(v).Interface()
				err := proto.Unmarshal(pkt.Bytes, realMsg.(proto.Message))
				if err != nil {
					dposlog.Error("peerConn recvRoutine Unmarshal data failed", "err", err)
					continue
				}
				if pc.transferChannel != nil && (pkt.TypeID == ttypes.VoteID || pkt.TypeID == ttypes.VoteReplyID || pkt.TypeID == ttypes.NotifyID || pkt.TypeID == ttypes.CBInfoID) {
					pc.transferChannel <- MsgInfo{pkt.TypeID, realMsg.(proto.Message), pc.ID(), pc.ip.String()}
				}
			} else {
				err := fmt.Errorf("Unknown message type %v", pkt.TypeID)
				dposlog.Error("Connection failed @ recvRoutine", "conn", pc, "err", err)
				pc.stopForError(err)
				break FOR_LOOP
			}
		}
	}

	close(pc.pongChannel)
	for range pc.pongChannel {
		// Drain
	}
}

// StackError struct
type StackError struct {
	Err   interface{}
	Stack []byte
}

func (se StackError) String() string {
	return fmt.Sprintf("Error: %v\nStack: %s", se.Err, se.Stack)
}

func (se StackError) Error() string {
	return se.String()
}
