// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"
	"time"
)

var (
	// MsgMap define
	MsgMap map[byte]reflect.Type
)

// step and message id define
const (
	VoteID      = byte(0x06)
	VoteReplyID = byte(0x07)
	NotifyID    = byte(0x08)

	PacketTypePing = byte(0xff)
	PacketTypePong = byte(0xfe)
)

// InitMessageMap ...
func InitMessageMap() {
	MsgMap = map[byte]reflect.Type{
		VoteID:      reflect.TypeOf(DPosVote{}),
		VoteReplyID: reflect.TypeOf(DPosVoteReply{}),
		NotifyID:    reflect.TypeOf(DPosNotify{}),
	}
}

// CanonicalJSONVoteItem ...
type CanonicalJSONVoteItem struct {
	VotedNodeIndex   int32  `json:"votedNodeIndex,omitempty"`
	VotedNodeAddress []byte `json:"votedNodeAddress,omitempty"`
	CycleStart       int64  `json:"cycleStart,omitempty"`
	CycleStop        int64  `json:"cycleStop,omitempty"`
	PeriodStart      int64  `json:"periodStart,omitempty"`
	PeriodStop       int64  `json:"periodStop,omitempty"`
	Height           int64  `json:"height,omitempty"`
	VoteID           []byte `json:"voteID,omitempty"`
}

// CanonicalJSONVote ...
type CanonicalJSONVote struct {
	VoteItem         *CanonicalJSONVoteItem `json:"vote,omitempty"`
	VoteTimestamp    int64                  `json:"voteTimestamp,omitempty"`
	VoterNodeIndex   int32                  `json:"voterNodeIndex,omitempty"`
	VoterNodeAddress []byte                 `json:"voterNodeAddress,omitempty"`
}

// CanonicalJSONOnceVote ...
type CanonicalJSONOnceVote struct {
	ChainID string            `json:"chain_id"`
	Vote    CanonicalJSONVote `json:"vote"`
}

// CanonicalVote ...
func CanonicalVote(vote *Vote) CanonicalJSONVote {
	return CanonicalJSONVote{
		VoteItem: &CanonicalJSONVoteItem{
			VotedNodeIndex:   vote.VoteItem.VotedNodeIndex,
			VotedNodeAddress: vote.VoteItem.VotedNodeAddress,
			CycleStart:       vote.VoteItem.CycleStart,
			CycleStop:        vote.VoteItem.CycleStop,
			PeriodStart:      vote.VoteItem.PeriodStart,
			PeriodStop:       vote.VoteItem.PeriodStop,
			Height:           vote.VoteItem.Height,
			VoteID:           vote.VoteItem.VoteID,
		},
		VoteTimestamp:    vote.VoteTimestamp,
		VoterNodeIndex:   vote.VoterNodeIndex,
		VoterNodeAddress: vote.VoterNodeAddress,
	}
}

// CanonicalJSONNotify ...
type CanonicalJSONNotify struct {
	VoteItem *CanonicalJSONVoteItem `json:"vote,omitempty"`

	HeightStop      int64 `json:"heightStop,omitempty"`
	NotifyTimestamp int64 `json:"notifyTimestamp,omitempty"`
}

// CanonicalJSONOnceNotify ...
type CanonicalJSONOnceNotify struct {
	ChainID string              `json:"chain_id"`
	Notify  CanonicalJSONNotify `json:"vote"`
}

// CanonicalNotify ...
func CanonicalNotify(notify *Notify) CanonicalJSONNotify {
	return CanonicalJSONNotify{
		VoteItem: &CanonicalJSONVoteItem{
			VotedNodeIndex:   notify.Vote.VotedNodeIndex,
			VotedNodeAddress: notify.Vote.VotedNodeAddress,
			CycleStart:       notify.Vote.CycleStart,
			CycleStop:        notify.Vote.CycleStop,
			PeriodStart:      notify.Vote.PeriodStart,
			PeriodStop:       notify.Vote.PeriodStop,
			Height:           notify.Vote.Height,
			VoteID:           notify.Vote.VoteID,
		},
		HeightStop:      notify.HeightStop,
		NotifyTimestamp: notify.NotifyTimestamp,
	}
}

// CanonicalTime ...
func CanonicalTime(t time.Time) string {
	// note that sending time over go-wire resets it to
	// local time, we need to force UTC here, so the
	// signatures match
	return t.UTC().Format(timeFormat)
}
