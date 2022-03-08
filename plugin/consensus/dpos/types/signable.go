// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/33cn/chain33/common/address"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log/log15"
)

// error defines
var (
	ErrNotifyInvalidValidatorAddress = errors.New("Invalid validator address for notify")
	ErrNotifyInvalidValidatorIndex   = errors.New("Invalid validator index for notify")
	ErrNotifyInvalidSignature        = errors.New("Invalid notify signature")

	ErrVoteInvalidValidatorIndex   = errors.New("Invalid validator index for vote")
	ErrVoteInvalidValidatorAddress = errors.New("Invalid validator address for vote")
	ErrVoteInvalidSignature        = errors.New("Invalid vote signature")
	ErrVoteNil                     = errors.New("Nil vote")

	votelog = log15.New("module", "tendermint-vote")

	ConsensusCrypto  crypto.Crypto
	SecureConnCrypto crypto.Crypto
)

// Signable is an interface for all signable things.
// It typically removes signatures before serializing.
type Signable interface {
	WriteSignBytes(chainID string, w io.Writer, n *int, err *error)
}

// SignBytes is a convenience method for getting the bytes to sign of a Signable.
func SignBytes(chainID string, o Signable) []byte {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	o.WriteSignBytes(chainID, buf, n, err)
	if *err != nil {
		PanicCrisis(err)
	}
	return buf.Bytes()
}

// Vote Represents a vote from validators for consensus.
type Vote struct {
	*DPosVote
}

// WriteSignBytes ...
func (vote *Vote) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	if *err != nil {
		return
	}
	canonical := CanonicalJSONOnceVote{
		chainID,
		CanonicalVote(vote),
	}
	byteVote, e := json.Marshal(&canonical)
	if e != nil {
		*err = e
		votelog.Error("vote WriteSignBytes marshal failed", "err", e)
		return
	}
	number, writeErr := w.Write(byteVote)
	*n = number
	*err = writeErr
}

// Copy ...
func (vote *Vote) Copy() *Vote {
	voteCopy := *vote
	return &voteCopy
}

func (vote *Vote) String() string {
	if vote == nil {
		return "nil-Vote"
	}

	return fmt.Sprintf("Vote{VotedNodeIndex:%v, VotedNodeAddr:%X,Cycle[%v,%v],Period[%v,%v],StartHeight:%v,VoteId:%X,VoteTimeStamp:%v,VoteNodeIndex:%v,VoteNodeAddr:%X,Sig:%X}",
		vote.VoteItem.VotedNodeIndex,
		Fingerprint(vote.VoteItem.VotedNodeAddress),
		vote.VoteItem.CycleStart,
		vote.VoteItem.CycleStop,
		vote.VoteItem.PeriodStart,
		vote.VoteItem.PeriodStop,
		vote.VoteItem.Height,
		Fingerprint(vote.VoteItem.VoteID),
		CanonicalTime(time.Unix(0, vote.VoteTimestamp)),
		vote.VoterNodeIndex,
		Fingerprint(vote.VoterNodeAddress),
		Fingerprint(vote.Signature),
	)
}

// Verify ...
func (vote *Vote) Verify(chainID string, pubKey crypto.PubKey) error {
	addr := address.BytesToBtcAddress(address.NormalVer, pubKey.Bytes()).Hash160[:]
	if !bytes.Equal(addr, vote.VoterNodeAddress) {
		return ErrVoteInvalidValidatorAddress
	}

	sig, err := ConsensusCrypto.SignatureFromBytes(vote.Signature)
	if err != nil {
		votelog.Error("vote Verify failed", "err", err)
		return err
	}

	if !pubKey.VerifyBytes(SignBytes(chainID, vote), sig) {
		return ErrVoteInvalidSignature
	}
	return nil
}

// Hash ...
func (vote *Vote) Hash() []byte {
	if vote == nil {
		//votelog.Error("vote hash is nil")
		return nil
	}
	bytes, err := json.Marshal(vote)
	if err != nil {
		votelog.Error("vote hash marshal failed", "err", err)
		return nil
	}

	return crypto.Ripemd160(bytes)
}

// Notify Represents a notify from validators for consensus.
type Notify struct {
	*DPosNotify
}

// WriteSignBytes ...
func (notify *Notify) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	if *err != nil {
		return
	}
	canonical := CanonicalJSONOnceNotify{
		chainID,
		CanonicalNotify(notify),
	}
	byteVote, e := json.Marshal(&canonical)
	if e != nil {
		*err = e
		votelog.Error("vote WriteSignBytes marshal failed", "err", e)
		return
	}
	number, writeErr := w.Write(byteVote)
	*n = number
	*err = writeErr
}

// Copy ...
func (notify *Notify) Copy() *Notify {
	notifyCopy := *notify
	return &notifyCopy
}

func (notify *Notify) String() string {
	if notify == nil {
		return "nil-notify"
	}

	return fmt.Sprintf("Notify{VotedNodeIndex:%v, VotedNodeAddr:%X,Cycle[%v,%v],Period[%v,%v],StartHeight:%v,VoteId:%X,NotifyTimeStamp:%v,HeightStop:%v,NotifyNodeIndex:%v,NotifyNodeAddr:%X,Sig:%X}",
		notify.Vote.VotedNodeIndex,
		Fingerprint(notify.Vote.VotedNodeAddress),
		notify.Vote.CycleStart,
		notify.Vote.CycleStop,
		notify.Vote.PeriodStart,
		notify.Vote.PeriodStop,
		notify.Vote.Height,
		Fingerprint(notify.Vote.VoteID),
		CanonicalTime(time.Unix(0, notify.NotifyTimestamp)),
		notify.HeightStop,
		notify.NotifyNodeIndex,
		Fingerprint(notify.NotifyNodeAddress),
		Fingerprint(notify.Signature),
	)
}

// Verify ...
func (notify *Notify) Verify(chainID string, pubKey crypto.PubKey) error {
	addr := address.BytesToBtcAddress(address.NormalVer, pubKey.Bytes()).Hash160[:]
	if !bytes.Equal(addr, notify.NotifyNodeAddress) {
		return ErrNotifyInvalidValidatorAddress
	}

	sig, err := ConsensusCrypto.SignatureFromBytes(notify.Signature)
	if err != nil {
		votelog.Error("Notify Verify failed", "err", err)
		return err
	}

	if !pubKey.VerifyBytes(SignBytes(chainID, notify), sig) {
		return ErrNotifyInvalidSignature
	}
	return nil
}

// Hash ...
func (notify *Notify) Hash() []byte {
	if notify == nil {
		//votelog.Error("vote hash is nil")
		return nil
	}
	bytes, err := json.Marshal(notify)
	if err != nil {
		votelog.Error("vote hash marshal failed", "err", err)
		return nil
	}

	return crypto.Ripemd160(bytes)
}
