// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.9.1
// source: autonomy.proto

package types

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// message for execs.Autonomy
type AutonomyAction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Value:
	//	*AutonomyAction_PropBoard
	//	*AutonomyAction_RvkPropBoard
	//	*AutonomyAction_VotePropBoard
	//	*AutonomyAction_TmintPropBoard
	//	*AutonomyAction_PropProject
	//	*AutonomyAction_RvkPropProject
	//	*AutonomyAction_VotePropProject
	//	*AutonomyAction_PubVotePropProject
	//	*AutonomyAction_TmintPropProject
	//	*AutonomyAction_PropRule
	//	*AutonomyAction_RvkPropRule
	//	*AutonomyAction_VotePropRule
	//	*AutonomyAction_TmintPropRule
	//	*AutonomyAction_Transfer
	//	*AutonomyAction_CommentProp
	//	*AutonomyAction_PropChange
	//	*AutonomyAction_RvkPropChange
	//	*AutonomyAction_VotePropChange
	//	*AutonomyAction_TmintPropChange
	Value isAutonomyAction_Value `protobuf_oneof:"value"`
	Ty    int32                  `protobuf:"varint,20,opt,name=ty,proto3" json:"ty,omitempty"`
}

func (x *AutonomyAction) Reset() {
	*x = AutonomyAction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autonomy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AutonomyAction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AutonomyAction) ProtoMessage() {}

func (x *AutonomyAction) ProtoReflect() protoreflect.Message {
	mi := &file_autonomy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AutonomyAction.ProtoReflect.Descriptor instead.
func (*AutonomyAction) Descriptor() ([]byte, []int) {
	return file_autonomy_proto_rawDescGZIP(), []int{0}
}

func (m *AutonomyAction) GetValue() isAutonomyAction_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *AutonomyAction) GetPropBoard() *ProposalBoard {
	if x, ok := x.GetValue().(*AutonomyAction_PropBoard); ok {
		return x.PropBoard
	}
	return nil
}

func (x *AutonomyAction) GetRvkPropBoard() *RevokeProposalBoard {
	if x, ok := x.GetValue().(*AutonomyAction_RvkPropBoard); ok {
		return x.RvkPropBoard
	}
	return nil
}

func (x *AutonomyAction) GetVotePropBoard() *VoteProposalBoard {
	if x, ok := x.GetValue().(*AutonomyAction_VotePropBoard); ok {
		return x.VotePropBoard
	}
	return nil
}

func (x *AutonomyAction) GetTmintPropBoard() *TerminateProposalBoard {
	if x, ok := x.GetValue().(*AutonomyAction_TmintPropBoard); ok {
		return x.TmintPropBoard
	}
	return nil
}

func (x *AutonomyAction) GetPropProject() *ProposalProject {
	if x, ok := x.GetValue().(*AutonomyAction_PropProject); ok {
		return x.PropProject
	}
	return nil
}

func (x *AutonomyAction) GetRvkPropProject() *RevokeProposalProject {
	if x, ok := x.GetValue().(*AutonomyAction_RvkPropProject); ok {
		return x.RvkPropProject
	}
	return nil
}

func (x *AutonomyAction) GetVotePropProject() *VoteProposalProject {
	if x, ok := x.GetValue().(*AutonomyAction_VotePropProject); ok {
		return x.VotePropProject
	}
	return nil
}

func (x *AutonomyAction) GetPubVotePropProject() *PubVoteProposalProject {
	if x, ok := x.GetValue().(*AutonomyAction_PubVotePropProject); ok {
		return x.PubVotePropProject
	}
	return nil
}

func (x *AutonomyAction) GetTmintPropProject() *TerminateProposalProject {
	if x, ok := x.GetValue().(*AutonomyAction_TmintPropProject); ok {
		return x.TmintPropProject
	}
	return nil
}

func (x *AutonomyAction) GetPropRule() *ProposalRule {
	if x, ok := x.GetValue().(*AutonomyAction_PropRule); ok {
		return x.PropRule
	}
	return nil
}

func (x *AutonomyAction) GetRvkPropRule() *RevokeProposalRule {
	if x, ok := x.GetValue().(*AutonomyAction_RvkPropRule); ok {
		return x.RvkPropRule
	}
	return nil
}

func (x *AutonomyAction) GetVotePropRule() *VoteProposalRule {
	if x, ok := x.GetValue().(*AutonomyAction_VotePropRule); ok {
		return x.VotePropRule
	}
	return nil
}

func (x *AutonomyAction) GetTmintPropRule() *TerminateProposalRule {
	if x, ok := x.GetValue().(*AutonomyAction_TmintPropRule); ok {
		return x.TmintPropRule
	}
	return nil
}

func (x *AutonomyAction) GetTransfer() *TransferFund {
	if x, ok := x.GetValue().(*AutonomyAction_Transfer); ok {
		return x.Transfer
	}
	return nil
}

func (x *AutonomyAction) GetCommentProp() *Comment {
	if x, ok := x.GetValue().(*AutonomyAction_CommentProp); ok {
		return x.CommentProp
	}
	return nil
}

func (x *AutonomyAction) GetPropChange() *ProposalChange {
	if x, ok := x.GetValue().(*AutonomyAction_PropChange); ok {
		return x.PropChange
	}
	return nil
}

func (x *AutonomyAction) GetRvkPropChange() *RevokeProposalChange {
	if x, ok := x.GetValue().(*AutonomyAction_RvkPropChange); ok {
		return x.RvkPropChange
	}
	return nil
}

func (x *AutonomyAction) GetVotePropChange() *VoteProposalChange {
	if x, ok := x.GetValue().(*AutonomyAction_VotePropChange); ok {
		return x.VotePropChange
	}
	return nil
}

func (x *AutonomyAction) GetTmintPropChange() *TerminateProposalChange {
	if x, ok := x.GetValue().(*AutonomyAction_TmintPropChange); ok {
		return x.TmintPropChange
	}
	return nil
}

func (x *AutonomyAction) GetTy() int32 {
	if x != nil {
		return x.Ty
	}
	return 0
}

type isAutonomyAction_Value interface {
	isAutonomyAction_Value()
}

type AutonomyAction_PropBoard struct {
	// 提案董事会相关
	PropBoard *ProposalBoard `protobuf:"bytes,1,opt,name=propBoard,proto3,oneof"`
}

type AutonomyAction_RvkPropBoard struct {
	RvkPropBoard *RevokeProposalBoard `protobuf:"bytes,2,opt,name=rvkPropBoard,proto3,oneof"`
}

type AutonomyAction_VotePropBoard struct {
	VotePropBoard *VoteProposalBoard `protobuf:"bytes,3,opt,name=votePropBoard,proto3,oneof"`
}

type AutonomyAction_TmintPropBoard struct {
	TmintPropBoard *TerminateProposalBoard `protobuf:"bytes,4,opt,name=tmintPropBoard,proto3,oneof"`
}

type AutonomyAction_PropProject struct {
	// 提案项目相关
	PropProject *ProposalProject `protobuf:"bytes,5,opt,name=propProject,proto3,oneof"`
}

type AutonomyAction_RvkPropProject struct {
	RvkPropProject *RevokeProposalProject `protobuf:"bytes,6,opt,name=rvkPropProject,proto3,oneof"`
}

type AutonomyAction_VotePropProject struct {
	VotePropProject *VoteProposalProject `protobuf:"bytes,7,opt,name=votePropProject,proto3,oneof"`
}

type AutonomyAction_PubVotePropProject struct {
	PubVotePropProject *PubVoteProposalProject `protobuf:"bytes,8,opt,name=pubVotePropProject,proto3,oneof"`
}

type AutonomyAction_TmintPropProject struct {
	TmintPropProject *TerminateProposalProject `protobuf:"bytes,9,opt,name=tmintPropProject,proto3,oneof"`
}

type AutonomyAction_PropRule struct {
	// 提案规则修改相关
	PropRule *ProposalRule `protobuf:"bytes,10,opt,name=propRule,proto3,oneof"`
}

type AutonomyAction_RvkPropRule struct {
	RvkPropRule *RevokeProposalRule `protobuf:"bytes,11,opt,name=rvkPropRule,proto3,oneof"`
}

type AutonomyAction_VotePropRule struct {
	VotePropRule *VoteProposalRule `protobuf:"bytes,12,opt,name=votePropRule,proto3,oneof"`
}

type AutonomyAction_TmintPropRule struct {
	TmintPropRule *TerminateProposalRule `protobuf:"bytes,13,opt,name=tmintPropRule,proto3,oneof"`
}

type AutonomyAction_Transfer struct {
	// 发展基金转自治系统合约
	Transfer *TransferFund `protobuf:"bytes,14,opt,name=transfer,proto3,oneof"`
}

type AutonomyAction_CommentProp struct {
	CommentProp *Comment `protobuf:"bytes,15,opt,name=commentProp,proto3,oneof"`
}

type AutonomyAction_PropChange struct {
	// 提案改变董事会成员
	PropChange *ProposalChange `protobuf:"bytes,16,opt,name=propChange,proto3,oneof"`
}

type AutonomyAction_RvkPropChange struct {
	RvkPropChange *RevokeProposalChange `protobuf:"bytes,17,opt,name=rvkPropChange,proto3,oneof"`
}

type AutonomyAction_VotePropChange struct {
	VotePropChange *VoteProposalChange `protobuf:"bytes,18,opt,name=votePropChange,proto3,oneof"`
}

type AutonomyAction_TmintPropChange struct {
	TmintPropChange *TerminateProposalChange `protobuf:"bytes,19,opt,name=tmintPropChange,proto3,oneof"`
}

func (*AutonomyAction_PropBoard) isAutonomyAction_Value() {}

func (*AutonomyAction_RvkPropBoard) isAutonomyAction_Value() {}

func (*AutonomyAction_VotePropBoard) isAutonomyAction_Value() {}

func (*AutonomyAction_TmintPropBoard) isAutonomyAction_Value() {}

func (*AutonomyAction_PropProject) isAutonomyAction_Value() {}

func (*AutonomyAction_RvkPropProject) isAutonomyAction_Value() {}

func (*AutonomyAction_VotePropProject) isAutonomyAction_Value() {}

func (*AutonomyAction_PubVotePropProject) isAutonomyAction_Value() {}

func (*AutonomyAction_TmintPropProject) isAutonomyAction_Value() {}

func (*AutonomyAction_PropRule) isAutonomyAction_Value() {}

func (*AutonomyAction_RvkPropRule) isAutonomyAction_Value() {}

func (*AutonomyAction_VotePropRule) isAutonomyAction_Value() {}

func (*AutonomyAction_TmintPropRule) isAutonomyAction_Value() {}

func (*AutonomyAction_Transfer) isAutonomyAction_Value() {}

func (*AutonomyAction_CommentProp) isAutonomyAction_Value() {}

func (*AutonomyAction_PropChange) isAutonomyAction_Value() {}

func (*AutonomyAction_RvkPropChange) isAutonomyAction_Value() {}

func (*AutonomyAction_VotePropChange) isAutonomyAction_Value() {}

func (*AutonomyAction_TmintPropChange) isAutonomyAction_Value() {}

var File_autonomy_proto protoreflect.FileDescriptor

var file_autonomy_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x61, 0x75, 0x74, 0x6f, 0x6e, 0x6f, 0x6d, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x05, 0x74, 0x79, 0x70, 0x65, 0x73, 0x1a, 0x0b, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x72, 0x75, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0c, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x85, 0x0a,
	0x0a, 0x0e, 0x41, 0x75, 0x74, 0x6f, 0x6e, 0x6f, 0x6d, 0x79, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x34, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x70, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x70,
	0x6f, 0x73, 0x61, 0x6c, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x48, 0x00, 0x52, 0x09, 0x70, 0x72, 0x6f,
	0x70, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x12, 0x40, 0x0a, 0x0c, 0x72, 0x76, 0x6b, 0x50, 0x72, 0x6f,
	0x70, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x76, 0x6f, 0x6b, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f,
	0x73, 0x61, 0x6c, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x48, 0x00, 0x52, 0x0c, 0x72, 0x76, 0x6b, 0x50,
	0x72, 0x6f, 0x70, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x12, 0x40, 0x0a, 0x0d, 0x76, 0x6f, 0x74, 0x65,
	0x50, 0x72, 0x6f, 0x70, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x18, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x56, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70,
	0x6f, 0x73, 0x61, 0x6c, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x48, 0x00, 0x52, 0x0d, 0x76, 0x6f, 0x74,
	0x65, 0x50, 0x72, 0x6f, 0x70, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x12, 0x47, 0x0a, 0x0e, 0x74, 0x6d,
	0x69, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x54, 0x65, 0x72, 0x6d, 0x69,
	0x6e, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x42, 0x6f, 0x61, 0x72,
	0x64, 0x48, 0x00, 0x52, 0x0e, 0x74, 0x6d, 0x69, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x42, 0x6f,
	0x61, 0x72, 0x64, 0x12, 0x3a, 0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x2e, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x48, 0x00, 0x52, 0x0b, 0x70, 0x72, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12,
	0x46, 0x0a, 0x0e, 0x72, 0x76, 0x6b, 0x50, 0x72, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x52, 0x65, 0x76, 0x6f, 0x6b, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x50, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x48, 0x00, 0x52, 0x0e, 0x72, 0x76, 0x6b, 0x50, 0x72, 0x6f, 0x70,
	0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x46, 0x0a, 0x0f, 0x76, 0x6f, 0x74, 0x65, 0x50,
	0x72, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x56, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f,
	0x70, 0x6f, 0x73, 0x61, 0x6c, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x48, 0x00, 0x52, 0x0f,
	0x76, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12,
	0x4f, 0x0a, 0x12, 0x70, 0x75, 0x62, 0x56, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x50, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x50, 0x75, 0x62, 0x56, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f,
	0x73, 0x61, 0x6c, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x48, 0x00, 0x52, 0x12, 0x70, 0x75,
	0x62, 0x56, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x12, 0x4d, 0x0a, 0x10, 0x74, 0x6d, 0x69, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x50, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2e, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70,
	0x6f, 0x73, 0x61, 0x6c, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x48, 0x00, 0x52, 0x10, 0x74,
	0x6d, 0x69, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12,
	0x31, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x70, 0x52, 0x75, 0x6c, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x13, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73,
	0x61, 0x6c, 0x52, 0x75, 0x6c, 0x65, 0x48, 0x00, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x70, 0x52, 0x75,
	0x6c, 0x65, 0x12, 0x3d, 0x0a, 0x0b, 0x72, 0x76, 0x6b, 0x50, 0x72, 0x6f, 0x70, 0x52, 0x75, 0x6c,
	0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x52, 0x65, 0x76, 0x6f, 0x6b, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52, 0x75,
	0x6c, 0x65, 0x48, 0x00, 0x52, 0x0b, 0x72, 0x76, 0x6b, 0x50, 0x72, 0x6f, 0x70, 0x52, 0x75, 0x6c,
	0x65, 0x12, 0x3d, 0x0a, 0x0c, 0x76, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x52, 0x75, 0x6c,
	0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x56, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52, 0x75, 0x6c, 0x65,
	0x48, 0x00, 0x52, 0x0c, 0x76, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x52, 0x75, 0x6c, 0x65,
	0x12, 0x44, 0x0a, 0x0d, 0x74, 0x6d, 0x69, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x52, 0x75, 0x6c,
	0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61,
	0x6c, 0x52, 0x75, 0x6c, 0x65, 0x48, 0x00, 0x52, 0x0d, 0x74, 0x6d, 0x69, 0x6e, 0x74, 0x50, 0x72,
	0x6f, 0x70, 0x52, 0x75, 0x6c, 0x65, 0x12, 0x31, 0x0a, 0x08, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66,
	0x65, 0x72, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x46, 0x75, 0x6e, 0x64, 0x48, 0x00, 0x52,
	0x08, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x12, 0x32, 0x0a, 0x0b, 0x63, 0x6f, 0x6d,
	0x6d, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x48, 0x00,
	0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x12, 0x37, 0x0a,
	0x0a, 0x70, 0x72, 0x6f, 0x70, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x10, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x15, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73,
	0x61, 0x6c, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x48, 0x00, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70,
	0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x43, 0x0a, 0x0d, 0x72, 0x76, 0x6b, 0x50, 0x72, 0x6f,
	0x70, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x52, 0x65, 0x76, 0x6f, 0x6b, 0x65, 0x50, 0x72, 0x6f, 0x70,
	0x6f, 0x73, 0x61, 0x6c, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x48, 0x00, 0x52, 0x0d, 0x72, 0x76,
	0x6b, 0x50, 0x72, 0x6f, 0x70, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x43, 0x0a, 0x0e, 0x76,
	0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x12, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x56, 0x6f, 0x74, 0x65,
	0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x48, 0x00,
	0x52, 0x0e, 0x76, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65,
	0x12, 0x4a, 0x0a, 0x0f, 0x74, 0x6d, 0x69, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x43, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x18, 0x13, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f,
	0x73, 0x61, 0x6c, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x48, 0x00, 0x52, 0x0f, 0x74, 0x6d, 0x69,
	0x6e, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x74, 0x79, 0x18, 0x14, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x74, 0x79, 0x42, 0x07, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x0a, 0x5a, 0x08, 0x2e, 0x2e, 0x2f, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_autonomy_proto_rawDescOnce sync.Once
	file_autonomy_proto_rawDescData = file_autonomy_proto_rawDesc
)

func file_autonomy_proto_rawDescGZIP() []byte {
	file_autonomy_proto_rawDescOnce.Do(func() {
		file_autonomy_proto_rawDescData = protoimpl.X.CompressGZIP(file_autonomy_proto_rawDescData)
	})
	return file_autonomy_proto_rawDescData
}

var file_autonomy_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_autonomy_proto_goTypes = []interface{}{
	(*AutonomyAction)(nil),           // 0: types.AutonomyAction
	(*ProposalBoard)(nil),            // 1: types.ProposalBoard
	(*RevokeProposalBoard)(nil),      // 2: types.RevokeProposalBoard
	(*VoteProposalBoard)(nil),        // 3: types.VoteProposalBoard
	(*TerminateProposalBoard)(nil),   // 4: types.TerminateProposalBoard
	(*ProposalProject)(nil),          // 5: types.ProposalProject
	(*RevokeProposalProject)(nil),    // 6: types.RevokeProposalProject
	(*VoteProposalProject)(nil),      // 7: types.VoteProposalProject
	(*PubVoteProposalProject)(nil),   // 8: types.PubVoteProposalProject
	(*TerminateProposalProject)(nil), // 9: types.TerminateProposalProject
	(*ProposalRule)(nil),             // 10: types.ProposalRule
	(*RevokeProposalRule)(nil),       // 11: types.RevokeProposalRule
	(*VoteProposalRule)(nil),         // 12: types.VoteProposalRule
	(*TerminateProposalRule)(nil),    // 13: types.TerminateProposalRule
	(*TransferFund)(nil),             // 14: types.TransferFund
	(*Comment)(nil),                  // 15: types.Comment
	(*ProposalChange)(nil),           // 16: types.ProposalChange
	(*RevokeProposalChange)(nil),     // 17: types.RevokeProposalChange
	(*VoteProposalChange)(nil),       // 18: types.VoteProposalChange
	(*TerminateProposalChange)(nil),  // 19: types.TerminateProposalChange
}
var file_autonomy_proto_depIdxs = []int32{
	1,  // 0: types.AutonomyAction.propBoard:type_name -> types.ProposalBoard
	2,  // 1: types.AutonomyAction.rvkPropBoard:type_name -> types.RevokeProposalBoard
	3,  // 2: types.AutonomyAction.votePropBoard:type_name -> types.VoteProposalBoard
	4,  // 3: types.AutonomyAction.tmintPropBoard:type_name -> types.TerminateProposalBoard
	5,  // 4: types.AutonomyAction.propProject:type_name -> types.ProposalProject
	6,  // 5: types.AutonomyAction.rvkPropProject:type_name -> types.RevokeProposalProject
	7,  // 6: types.AutonomyAction.votePropProject:type_name -> types.VoteProposalProject
	8,  // 7: types.AutonomyAction.pubVotePropProject:type_name -> types.PubVoteProposalProject
	9,  // 8: types.AutonomyAction.tmintPropProject:type_name -> types.TerminateProposalProject
	10, // 9: types.AutonomyAction.propRule:type_name -> types.ProposalRule
	11, // 10: types.AutonomyAction.rvkPropRule:type_name -> types.RevokeProposalRule
	12, // 11: types.AutonomyAction.votePropRule:type_name -> types.VoteProposalRule
	13, // 12: types.AutonomyAction.tmintPropRule:type_name -> types.TerminateProposalRule
	14, // 13: types.AutonomyAction.transfer:type_name -> types.TransferFund
	15, // 14: types.AutonomyAction.commentProp:type_name -> types.Comment
	16, // 15: types.AutonomyAction.propChange:type_name -> types.ProposalChange
	17, // 16: types.AutonomyAction.rvkPropChange:type_name -> types.RevokeProposalChange
	18, // 17: types.AutonomyAction.votePropChange:type_name -> types.VoteProposalChange
	19, // 18: types.AutonomyAction.tmintPropChange:type_name -> types.TerminateProposalChange
	19, // [19:19] is the sub-list for method output_type
	19, // [19:19] is the sub-list for method input_type
	19, // [19:19] is the sub-list for extension type_name
	19, // [19:19] is the sub-list for extension extendee
	0,  // [0:19] is the sub-list for field type_name
}

func init() { file_autonomy_proto_init() }
func file_autonomy_proto_init() {
	if File_autonomy_proto != nil {
		return
	}
	file_board_proto_init()
	file_project_proto_init()
	file_rule_proto_init()
	file_change_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_autonomy_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AutonomyAction); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_autonomy_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*AutonomyAction_PropBoard)(nil),
		(*AutonomyAction_RvkPropBoard)(nil),
		(*AutonomyAction_VotePropBoard)(nil),
		(*AutonomyAction_TmintPropBoard)(nil),
		(*AutonomyAction_PropProject)(nil),
		(*AutonomyAction_RvkPropProject)(nil),
		(*AutonomyAction_VotePropProject)(nil),
		(*AutonomyAction_PubVotePropProject)(nil),
		(*AutonomyAction_TmintPropProject)(nil),
		(*AutonomyAction_PropRule)(nil),
		(*AutonomyAction_RvkPropRule)(nil),
		(*AutonomyAction_VotePropRule)(nil),
		(*AutonomyAction_TmintPropRule)(nil),
		(*AutonomyAction_Transfer)(nil),
		(*AutonomyAction_CommentProp)(nil),
		(*AutonomyAction_PropChange)(nil),
		(*AutonomyAction_RvkPropChange)(nil),
		(*AutonomyAction_VotePropChange)(nil),
		(*AutonomyAction_TmintPropChange)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_autonomy_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_autonomy_proto_goTypes,
		DependencyIndexes: file_autonomy_proto_depIdxs,
		MessageInfos:      file_autonomy_proto_msgTypes,
	}.Build()
	File_autonomy_proto = out.File
	file_autonomy_proto_rawDesc = nil
	file_autonomy_proto_goTypes = nil
	file_autonomy_proto_depIdxs = nil
}
