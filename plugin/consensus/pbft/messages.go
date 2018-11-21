// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pbft

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net"

	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

// EQ Digest
func EQ(d1 []byte, d2 []byte) bool {
	if len(d1) != len(d2) {
		return false
	}
	for idx, b := range d1 {
		if b != d2[idx] {
			return false
		}
	}
	return true
}

// ToCheckpoint method
func ToCheckpoint(sequence uint32, digest []byte) *types.Checkpoint {
	return &types.Checkpoint{Sequence: sequence, Digest: digest}
}

// ToEntry method
func ToEntry(sequence uint32, digest []byte, view uint32) *types.Entry {
	return &types.Entry{Sequence: sequence, Digest: digest, View: view}
}

// ToViewChange method
func ToViewChange(viewchanger uint32, digest []byte) *types.ViewChange {
	return &types.ViewChange{Viewchanger: viewchanger, Digest: digest}
}

// ToSummary method
func ToSummary(sequence uint32, digest []byte) *types.Summary {
	return &types.Summary{Sequence: sequence, Digest: digest}
}

// ToRequestClient method
func ToRequestClient(op *types.Operation, timestamp, client string) *types.Request {
	return &types.Request{
		Value: &types.Request_Client{
			Client: &types.RequestClient{Op: op, Timestamp: timestamp, Client: client}},
	}
}

// ToRequestPreprepare method
func ToRequestPreprepare(view, sequence uint32, digest []byte, replica uint32) *types.Request {
	return &types.Request{
		Value: &types.Request_Preprepare{
			Preprepare: &types.RequestPrePrepare{View: view, Sequence: sequence, Digest: digest, Replica: replica}},
	}
}

// ToRequestPrepare method
func ToRequestPrepare(view, sequence uint32, digest []byte, replica uint32) *types.Request {
	return &types.Request{
		Value: &types.Request_Prepare{
			Prepare: &types.RequestPrepare{View: view, Sequence: sequence, Digest: digest, Replica: replica}},
	}
}

// ToRequestCommit method
func ToRequestCommit(view, sequence, replica uint32) *types.Request {
	return &types.Request{
		Value: &types.Request_Commit{
			Commit: &types.RequestCommit{View: view, Sequence: sequence, Replica: replica}},
	}
}

// ToRequestCheckpoint method
func ToRequestCheckpoint(sequence uint32, digest []byte, replica uint32) *types.Request {
	return &types.Request{
		Value: &types.Request_Checkpoint{
			Checkpoint: &types.RequestCheckpoint{Sequence: sequence, Digest: digest, Replica: replica}},
	}
}

// ToRequestViewChange method
func ToRequestViewChange(view, sequence uint32, checkpoints []*types.Checkpoint, preps, prePreps []*types.Entry, replica uint32) *types.Request {
	return &types.Request{
		Value: &types.Request_Viewchange{
			Viewchange: &types.RequestViewChange{View: view, Sequence: sequence, Checkpoints: checkpoints, Preps: preps, Prepreps: prePreps, Replica: replica}},
	}
}

// ToRequestAck method
func ToRequestAck(view, replica, viewchanger uint32, digest []byte) *types.Request {
	return &types.Request{
		Value: &types.Request_Ack{
			Ack: &types.RequestAck{View: view, Replica: replica, Viewchanger: viewchanger, Digest: digest}},
	}
}

// ToRequestNewView method
func ToRequestNewView(view uint32, viewChanges []*types.ViewChange, summaries []*types.Summary, replica uint32) *types.Request {
	return &types.Request{
		Value: &types.Request_Newview{
			Newview: &types.RequestNewView{View: view, Viewchanges: viewChanges, Summaries: summaries, Replica: replica}},
	}
}

// ReqDigest method
func ReqDigest(req *types.Request) []byte {
	if req == nil {
		return nil
	}
	bytes := md5.Sum([]byte(req.String()))
	return bytes[:]
}

/*func (req *Request) LowWaterMark() uint32 {
	// only for requestViewChange
	reqViewChange := req.GetViewchange()
	checkpoints := reqViewChange.GetCheckpoints()
	lastStable := checkpoints[len(checkpoints)-1]
	lwm := lastStable.Sequence
	return lwm
}*/

// ToReply method
func ToReply(view uint32, timestamp, client string, replica uint32, result *types.Result) *types.ClientReply {
	return &types.ClientReply{View: view, Timestamp: timestamp, Client: client, Replica: replica, Result: result}
}

// RepDigest method
func RepDigest(reply fmt.Stringer) []byte {
	if reply == nil {
		return nil
	}
	bytes := md5.Sum([]byte(reply.String()))
	return bytes[:]
}

// WriteMessage write proto message
func WriteMessage(addr string, msg proto.Message) error {
	conn, err := net.Dial("tcp", addr)
	defer conn.Close()
	if err != nil {
		return err
	}
	bz, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	n, err := conn.Write(bz)
	plog.Debug("size of byte is", "", n)
	return err
}

// ReadMessage read proto message
func ReadMessage(conn io.Reader, msg proto.Message) error {
	var buf bytes.Buffer
	n, err := io.Copy(&buf, conn)
	plog.Debug("size of byte is", "", n)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(buf.Bytes(), msg)
	return err
}
