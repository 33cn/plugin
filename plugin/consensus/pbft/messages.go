// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pbft

import (
	"bytes"
	"encoding/base64"
	"io"
	"net"

	"github.com/33cn/plugin/plugin/dapp/pbft/types"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/sha3"
)

// EQ 主要用于比较两个[]byte是否是相同的(内容完全一样)
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

// ToCheckpoint 用于产生Checkpoint的types，非Request
func ToCheckpoint(sequence uint64, digest string) *types.Checkpoint {
	return &types.Checkpoint{Sequence: sequence, Digest: digest}
}

//=====================================================
// Request生成操作
//=====================================================

// 生成不同的Request

// ToRequestClient 用于产生<Client> Request
func ToRequestClient(op *types.BlockData, timestamp, client string) *types.Request {
	return &types.Request{
		Value: &types.Request_Client{
			Client: &types.RequestClient{Op: op, Timestamp: timestamp, Client: client}},
	}
}

// ToRequestPreprepare 用于产生<Pre-prepare> Request
func ToRequestPreprepare(view, sequence uint64, digest string, request *types.RequestClient, replica uint64) *types.Request {
	return &types.Request{
		Value: &types.Request_Preprepare{
			Preprepare: &types.RequestPrePrepare{View: view, Sequence: sequence, Digest: digest, Request: request, Replica: replica}},
	}
}

// ToRequestPrepare 用于产生<Prepare> Request
func ToRequestPrepare(view, sequence uint64, digest string, replica uint64) *types.Request {
	return &types.Request{
		Value: &types.Request_Prepare{
			Prepare: &types.RequestPrepare{View: view, Sequence: sequence, Digest: digest, Replica: replica}},
	}
}

// ToRequestCommit 用于产生<Commit> Request
func ToRequestCommit(view, sequence uint64, digest string, replica uint64) *types.Request {
	return &types.Request{
		Value: &types.Request_Commit{
			Commit: &types.RequestCommit{View: view, Sequence: sequence, Digest: digest, Replica: replica}},
	}
}

// ToRequestCheckpoint 用于产生<Checkpoint> Request
func ToRequestCheckpoint(sequence uint64, digest string, replica uint64) *types.Request {
	return &types.Request{
		Value: &types.Request_Checkpoint{
			Checkpoint: &types.RequestCheckpoint{Sequence: sequence, Digest: digest, Replica: replica}},
	}
}

// ToRequestViewChange 用于产生<View-change> Request
func ToRequestViewChange(view, h uint64, checkpoints []*types.RequestViewChange_C,
	preps []*types.RequestViewChange_PQ, prePreps []*types.RequestViewChange_PQ,
	replica uint64) *types.Request {
	return &types.Request{
		Value: &types.Request_Viewchange{
			Viewchange: &types.RequestViewChange{View: view, H: h, Cset: checkpoints, Pset: preps, Qset: prePreps, Replica: replica}},
	}
}

// ToRequestAck 用于产生<Ack> Request
func ToRequestAck(view, replica, viewchanger uint64, digest string) *types.Request {
	return &types.Request{
		Value: &types.Request_Ack{
			Ack: &types.RequestAck{View: view, Replica: replica, ViewchangeSender: viewchanger, Digest: digest}},
	}
}

// ToRequestNewView 用于产生<New-view> Request
func ToRequestNewView(view uint64, viewChanges []*types.RequestViewChange, summaries map[uint64]string, replica uint64) *types.Request {
	return &types.Request{
		Value: &types.Request_Newview{
			Newview: &types.RequestNewView{View: view, Vset: viewChanges, Xset: summaries, Replica: replica}},
	}
}

// ToRequestReply 用于产生一个返回给客户端的回复，该回复没有确认
func ToRequestReply(view uint64, timestamp, client string, replica uint64, block *types.BlockData) *types.Request {
	return &types.Request{
		Value: &types.Request_Reply{
			Reply: &types.ClientReply{View: view, Timestamp: timestamp, Client: client, Replica: replica, Result: block}},
	}
}

//=====================================================
// 消息Digest操作
//=====================================================

// ComputeCryptoHash 用于执行加密hash获得hash 数组
func ComputeCryptoHash(data []byte) (hash []byte) {
	hash = make([]byte, 64)
	sha3.ShakeSum256(hash, data)
	return
}

// Hash 用于处理加密Request
func Hash(REQ *types.Request) string {
	var raw []byte
	var err error
	switch REQ.Value.(type) {
	case *types.Request_Client:
		raw, err = proto.Marshal(REQ.GetClient())
	case *types.Request_Preprepare:
		raw, err = proto.Marshal(REQ.GetPreprepare())
	case *types.Request_Prepare:
		raw, err = proto.Marshal(REQ.GetPrepare())
	case *types.Request_Commit:
		raw, err = proto.Marshal(REQ.GetCommit())
	case *types.Request_Checkpoint:
		raw, err = proto.Marshal(REQ.GetCheckpoint())
	case *types.Request_Viewchange:
		raw, err = proto.Marshal(REQ.GetViewchange())
	case *types.Request_Ack:
		raw, err = proto.Marshal(REQ.GetAck())
	case *types.Request_Newview:
		raw, err = proto.Marshal(REQ.GetNewview())
	default:
		plog.Error("Asked to hash non-supported message type, ignoring")
		return ""
	}
	if err != nil {
		plog.Error("Hash() Marshal failed", "type", REQ.Value, "err", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(ComputeCryptoHash(raw))
}

// DigestClientRequest 用于计算客户端消息的Hash值
func DigestClientRequest(REQ *types.RequestClient) string {
	if REQ == nil {
		return ""
	}
	raw, err := proto.Marshal(REQ)
	if err != nil {
		plog.Error("DigestClientRequest() Marshal failed", "RequestClient", REQ, "err", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(ComputeCryptoHash(raw))
}

// DigestReply 用于计算回复消息的Hash值
func DigestReply(reply *types.ClientReply) string {
	if reply == nil {
		return ""
	}
	raw, err := proto.Marshal(reply)
	if err != nil {
		plog.Error("DigestReply() Marshal failed", "ClientReply", reply, "err", err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(ComputeCryptoHash(raw))
}

// DigestViewchange 用于计算视图变更消息的Hash值
func DigestViewchange(vc *types.RequestViewChange) string {
	if vc == nil {
		return ""
	}
	raw, err := proto.Marshal(vc)
	if err != nil {
		plog.Error("DigestViewchange() Marshal failed", "RequestViewChange", vc, "err", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(ComputeCryptoHash(raw))
}

// Digest
/*func (req *Request) LowWaterMark() uint64 {
	// only for requestViewChange
	reqViewChange := req.GetViewchange()
	checkpoints := reqViewChange.GetCheckpoints()
	lastStable := checkpoints[len(checkpoints)-1]
	lwm := lastStable.Sequence
	return lwm
}*/

//=====================================================
// 消息传输操作
//=====================================================

// WriteMessage 用于向地址addr写proto类的消息
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
	_, err = conn.Write(bz)
	return err
}

// ReadMessage 用于读取写入的proto类的消息
func ReadMessage(conn io.Reader, msg proto.Message) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, conn)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(buf.Bytes(), msg)
	return err
}
