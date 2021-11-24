package tendermint

import (
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/tendermint/types"
	tmtypes "github.com/33cn/plugin/plugin/dapp/valnode/types"
	"github.com/stretchr/testify/assert"
)

var (
	secureConnCrypto crypto.Crypto
	sum              = 0
	mutx             sync.Mutex
	privKey          = "23278EA4CFE8B00360EBB376F2BBFAC345136EE5BC4549532C394C0AF2B80DFE8D80E15927EF2854C78D981015BD2AD469867957081357D0FADD88871752A7E1"
	expectAddress    = "07FE011CE6F4C458FD9D417ED38CB262A4364FA1"
)

func init() {
	cr2, err := crypto.Load(types.GetSignName("", types.ED25519), -1)
	if err != nil {
		fmt.Println("crypto.Load failed for types.ED25519")
		return
	}
	secureConnCrypto = cr2
}

func TestParallel(t *testing.T) {
	Parallel(
		func() {
			mutx.Lock()
			sum++
			mutx.Unlock()
		},
		func() {
			mutx.Lock()
			sum += 2
			mutx.Unlock()
		},
		func() {
			mutx.Lock()
			sum += 3
			mutx.Unlock()
		},
	)

	fmt.Println("TestParallel ok")
	assert.Equal(t, 6, sum)
}

func TestGenIDByPubKey(t *testing.T) {
	tmp, err := hex.DecodeString(privKey)
	assert.Nil(t, err)

	priv, err := secureConnCrypto.PrivKeyFromBytes(tmp)
	assert.Nil(t, err)

	id := GenIDByPubKey(priv.PubKey())
	addr, err := hex.DecodeString(string(id))
	assert.Nil(t, err)
	strAddr := fmt.Sprintf("%X", addr)
	assert.Equal(t, expectAddress, strAddr)
	fmt.Println("TestGenIDByPubKey ok")
}

func TestIP2IPPort(t *testing.T) {
	testMap := NewMutexMap()
	assert.Equal(t, false, testMap.Has("1.1.1.1"))

	testMap.Set("1.1.1.1", "1.1.1.1:80")
	assert.Equal(t, true, testMap.Has("1.1.1.1"))

	testMap.Set("1.1.1.2", "1.1.1.2:80")
	assert.Equal(t, true, testMap.Has("1.1.1.2"))

	testMap.Delete("1.1.1.1")
	assert.Equal(t, false, testMap.Has("1.1.1.1"))
	fmt.Println("TestIP2IPPort ok")
}

func TestNodeFunc(t *testing.T) {
	node := &Node{Version: "1.1.1", Network: "net1"}
	assert.NotNil(t, node.CompatibleWith(NodeInfo{Version: "1.1", Network: "net1"}))
	assert.NotNil(t, node.CompatibleWith(NodeInfo{Version: "2.1.1", Network: "net1"}))
	assert.NotNil(t, node.CompatibleWith(NodeInfo{Version: "1.1.1", Network: "net2"}))
	assert.Nil(t, node.CompatibleWith(NodeInfo{Version: "1.2.3", Network: "net1"}))

	assert.False(t, isIpv6(net.IP{127, 0, 0, 1}))
	assert.True(t, isIpv6(net.IP{0xff, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}))
	fmt.Println("TestNodeFunc ok")
}

func TestPeerSet(t *testing.T) {
	testSet := NewPeerSet()
	assert.Equal(t, false, testSet.Has("1"))

	peer1 := &peerConn{id: "1", ip: net.IP{127, 0, 0, 1}}
	testSet.Add(peer1)
	assert.Equal(t, true, testSet.Has("1"))
	assert.Equal(t, true, testSet.HasIP(net.IP{127, 0, 0, 1}))

	err := testSet.Add(peer1)
	assert.NotNil(t, err)

	peer2 := &peerConn{id: "2", ip: net.IP{127, 0, 0, 2}}
	testSet.Add(peer2)
	assert.Equal(t, true, testSet.Has("2"))
	assert.Equal(t, 2, testSet.Size())

	testSet.Remove(peer1)
	assert.Equal(t, 1, testSet.Size())
	assert.Equal(t, false, testSet.Has("1"))
	assert.Equal(t, false, testSet.HasIP(net.IP{127, 0, 0, 1}))

	fmt.Println("TestPeerSet ok")
}

func TestPeerConn(t *testing.T) {
	pc := &peerConn{id: "3", ip: net.IP{127, 0, 0, 3}, outbound: true, persistent: false}
	_, err := pc.RemoteAddr()
	assert.NotNil(t, err)
	assert.True(t, pc.IsOutbound())
	assert.False(t, pc.IsPersistent())

	pc.sendQueue = make(chan MsgInfo, maxSendQueueSize)
	assert.False(t, pc.Send(MsgInfo{}))
	assert.False(t, pc.TrySend(MsgInfo{}))
	pc.started = 1
	assert.True(t, pc.Send(MsgInfo{}))
	assert.True(t, pc.TrySend(MsgInfo{}))

	testUpdateStateRoutine(t, pc)

	fmt.Println("TestPeerConn ok")
}

func testUpdateStateRoutine(t *testing.T, pc *peerConn) {
	pc.quitUpdate = make(chan struct{})
	pc.updateStateQueue = make(chan MsgInfo)
	pc.state = &PeerConnState{
		ip: pc.ip,
		PeerRoundState: ttypes.PeerRoundState{
			Height:             int64(2),
			Round:              0,
			Step:               ttypes.RoundStepCommit,
			Proposal:           true,
			ProposalBlockHash:  []byte("ProposalBlockHash@2"),
			LastCommitRound:    0,
			CatchupCommitRound: 0,
		},
	}
	ps := pc.state
	go pc.updateStateRoutine()

	//NewRoundStepID msg
	rsMsg := &tmtypes.NewRoundStepMsg{
		Height:                int64(3),
		Round:                 int32(1),
		Step:                  int32(3),
		SecondsSinceStartTime: int32(1),
		LastCommitRound:       int32(1),
	}
	pc.updateStateQueue <- MsgInfo{ttypes.NewRoundStepID, rsMsg, ID("TEST"), pc.ip.String()}
	pc.updateStateQueue <- MsgInfo{TypeID: byte(0x00)}
	assert.Equal(t, int64(3), ps.Height)
	assert.Equal(t, 1, ps.Round)
	assert.Equal(t, ttypes.RoundStepPropose, ps.Step)
	assert.Equal(t, false, ps.Proposal)
	assert.Equal(t, 1, ps.LastCommitRound)
	assert.Equal(t, -1, ps.CatchupCommitRound)
	//SetHasProposal
	proposal := &tmtypes.Proposal{
		Height:    int64(3),
		Round:     int32(1),
		POLRound:  int32(1),
		Blockhash: []byte("ProposalBlockHash@3"),
	}
	ps.SetHasProposal(proposal)
	assert.True(t, ps.Proposal)
	assert.Equal(t, 1, ps.ProposalPOLRound)
	assert.Equal(t, []byte("ProposalBlockHash@3"), ps.ProposalBlockHash)
	//SetHasProposalBlock
	block := &ttypes.TendermintBlock{
		TendermintBlock: &tmtypes.TendermintBlock{
			Header: &tmtypes.TendermintBlockHeader{
				Height: int64(3),
				Round:  int64(1),
			},
		},
	}
	ps.SetHasProposalBlock(block)
	assert.True(t, ps.ProposalBlock)
	//ValidBlockID msg
	validBlockMsg := &tmtypes.ValidBlockMsg{
		Height:    int64(3),
		Round:     int32(1),
		Blockhash: []byte("ValidBlockHash@3"),
		IsCommit:  false,
	}
	pc.updateStateQueue <- MsgInfo{ttypes.ValidBlockID, validBlockMsg, ID("TEST"), pc.ip.String()}
	pc.updateStateQueue <- MsgInfo{TypeID: byte(0x00)}
	assert.Equal(t, []byte("ValidBlockHash@3"), ps.ProposalBlockHash)
	//HasVoteID msg
	hasVoteMsg := &tmtypes.HasVoteMsg{
		Height: int64(3),
		Round:  int32(1),
		Type:   int32(ttypes.VoteTypePrevote),
		Index:  int32(1),
	}
	ps.EnsureVoteBitArrays(int64(3), 2)
	ps.EnsureVoteBitArrays(int64(2), 2)
	assert.False(t, ps.Prevotes.GetIndex(1))
	pc.updateStateQueue <- MsgInfo{ttypes.HasVoteID, hasVoteMsg, ID("TEST"), pc.ip.String()}
	pc.updateStateQueue <- MsgInfo{TypeID: byte(0x00)}
	assert.True(t, ps.Prevotes.GetIndex(1))
	//ProposalPOLID msg
	proposalPOL := ps.Prevotes.TendermintBitArray
	proposalPOLMsg := &tmtypes.ProposalPOLMsg{
		Height:           int64(3),
		ProposalPOLRound: int32(1),
		ProposalPOL:      proposalPOL,
	}
	pc.updateStateQueue <- MsgInfo{ttypes.ProposalPOLID, proposalPOLMsg, ID("TEST"), pc.ip.String()}
	pc.updateStateQueue <- MsgInfo{TypeID: byte(0x00)}
	assert.EqualValues(t, proposalPOL, ps.ProposalPOL.TendermintBitArray)

	//PickSendVote
	ttypes.Init()
	vals := make([]*ttypes.Validator, 2)
	votes := ttypes.NewVoteSet("TEST", 3, 1, ttypes.VoteTypePrevote, &ttypes.ValidatorSet{Validators: vals})
	assert.False(t, pc.PickSendVote(votes))

	assert.Equal(t, int64(3), ps.GetHeight())
	assert.NotNil(t, ps.GetRoundState())
	assert.Nil(t, ps.getVoteBitArray(3, 1, byte(0x03)))
	assert.NotNil(t, ps.getVoteBitArray(3, 1, ttypes.VoteTypePrecommit))
	assert.Nil(t, ps.getVoteBitArray(2, 1, ttypes.VoteTypePrevote))
	assert.NotNil(t, ps.getVoteBitArray(2, 1, ttypes.VoteTypePrecommit))

	ps.ensureCatchupCommitRound(3, 2, 2)
	assert.Equal(t, 2, ps.CatchupCommitRound)
	assert.NotNil(t, ps.CatchupCommit)
	assert.Nil(t, ps.getVoteBitArray(3, 2, ttypes.VoteTypePrevote))
	assert.NotNil(t, ps.getVoteBitArray(3, 2, ttypes.VoteTypePrecommit))

	pc.quitUpdate <- struct{}{}

	fmt.Println("testUpdateStateRoutine ok")
}
