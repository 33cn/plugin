// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rule.proto

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type AutonomyProposalRule struct {
	PropRule *ProposalRule `protobuf:"bytes,1,opt,name=propRule" json:"propRule,omitempty"`
	CurRule  *RuleConfig   `protobuf:"bytes,2,opt,name=curRule" json:"curRule,omitempty"`
	// 全体持票人投票结果
	VoteResult *VoteResult `protobuf:"bytes,3,opt,name=voteResult" json:"voteResult,omitempty"`
	// 状态
	Status     int32  `protobuf:"varint,4,opt,name=status" json:"status,omitempty"`
	Address    string `protobuf:"bytes,5,opt,name=address" json:"address,omitempty"`
	Height     int64  `protobuf:"varint,6,opt,name=height" json:"height,omitempty"`
	Index      int32  `protobuf:"varint,7,opt,name=index" json:"index,omitempty"`
	ProposalID string `protobuf:"bytes,8,opt,name=proposalID" json:"proposalID,omitempty"`
}

func (m *AutonomyProposalRule) Reset()                    { *m = AutonomyProposalRule{} }
func (m *AutonomyProposalRule) String() string            { return proto.CompactTextString(m) }
func (*AutonomyProposalRule) ProtoMessage()               {}
func (*AutonomyProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

func (m *AutonomyProposalRule) GetPropRule() *ProposalRule {
	if m != nil {
		return m.PropRule
	}
	return nil
}

func (m *AutonomyProposalRule) GetCurRule() *RuleConfig {
	if m != nil {
		return m.CurRule
	}
	return nil
}

func (m *AutonomyProposalRule) GetVoteResult() *VoteResult {
	if m != nil {
		return m.VoteResult
	}
	return nil
}

func (m *AutonomyProposalRule) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *AutonomyProposalRule) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *AutonomyProposalRule) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *AutonomyProposalRule) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *AutonomyProposalRule) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

type ProposalRule struct {
	// 提案时间
	Year  int32 `protobuf:"varint,1,opt,name=year" json:"year,omitempty"`
	Month int32 `protobuf:"varint,2,opt,name=month" json:"month,omitempty"`
	Day   int32 `protobuf:"varint,3,opt,name=day" json:"day,omitempty"`
	// 规则可修改项,如果某项不修改则置为-1
	RuleCfg *RuleConfig `protobuf:"bytes,4,opt,name=ruleCfg" json:"ruleCfg,omitempty"`
	// 投票相关
	StartBlockHeight   int64 `protobuf:"varint,5,opt,name=startBlockHeight" json:"startBlockHeight,omitempty"`
	EndBlockHeight     int64 `protobuf:"varint,6,opt,name=endBlockHeight" json:"endBlockHeight,omitempty"`
	RealEndBlockHeight int64 `protobuf:"varint,7,opt,name=realEndBlockHeight" json:"realEndBlockHeight,omitempty"`
}

func (m *ProposalRule) Reset()                    { *m = ProposalRule{} }
func (m *ProposalRule) String() string            { return proto.CompactTextString(m) }
func (*ProposalRule) ProtoMessage()               {}
func (*ProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{1} }

func (m *ProposalRule) GetYear() int32 {
	if m != nil {
		return m.Year
	}
	return 0
}

func (m *ProposalRule) GetMonth() int32 {
	if m != nil {
		return m.Month
	}
	return 0
}

func (m *ProposalRule) GetDay() int32 {
	if m != nil {
		return m.Day
	}
	return 0
}

func (m *ProposalRule) GetRuleCfg() *RuleConfig {
	if m != nil {
		return m.RuleCfg
	}
	return nil
}

func (m *ProposalRule) GetStartBlockHeight() int64 {
	if m != nil {
		return m.StartBlockHeight
	}
	return 0
}

func (m *ProposalRule) GetEndBlockHeight() int64 {
	if m != nil {
		return m.EndBlockHeight
	}
	return 0
}

func (m *ProposalRule) GetRealEndBlockHeight() int64 {
	if m != nil {
		return m.RealEndBlockHeight
	}
	return 0
}

type RevokeProposalRule struct {
	ProposalID string `protobuf:"bytes,1,opt,name=proposalID" json:"proposalID,omitempty"`
}

func (m *RevokeProposalRule) Reset()                    { *m = RevokeProposalRule{} }
func (m *RevokeProposalRule) String() string            { return proto.CompactTextString(m) }
func (*RevokeProposalRule) ProtoMessage()               {}
func (*RevokeProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{2} }

func (m *RevokeProposalRule) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

type VoteProposalRule struct {
	ProposalID string `protobuf:"bytes,1,opt,name=proposalID" json:"proposalID,omitempty"`
	Approve    bool   `protobuf:"varint,2,opt,name=approve" json:"approve,omitempty"`
}

func (m *VoteProposalRule) Reset()                    { *m = VoteProposalRule{} }
func (m *VoteProposalRule) String() string            { return proto.CompactTextString(m) }
func (*VoteProposalRule) ProtoMessage()               {}
func (*VoteProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{3} }

func (m *VoteProposalRule) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

func (m *VoteProposalRule) GetApprove() bool {
	if m != nil {
		return m.Approve
	}
	return false
}

type TerminateProposalRule struct {
	ProposalID string `protobuf:"bytes,1,opt,name=proposalID" json:"proposalID,omitempty"`
}

func (m *TerminateProposalRule) Reset()                    { *m = TerminateProposalRule{} }
func (m *TerminateProposalRule) String() string            { return proto.CompactTextString(m) }
func (*TerminateProposalRule) ProtoMessage()               {}
func (*TerminateProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{4} }

func (m *TerminateProposalRule) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

// receipt
type ReceiptProposalRule struct {
	Prev    *AutonomyProposalRule `protobuf:"bytes,1,opt,name=prev" json:"prev,omitempty"`
	Current *AutonomyProposalRule `protobuf:"bytes,2,opt,name=current" json:"current,omitempty"`
}

func (m *ReceiptProposalRule) Reset()                    { *m = ReceiptProposalRule{} }
func (m *ReceiptProposalRule) String() string            { return proto.CompactTextString(m) }
func (*ReceiptProposalRule) ProtoMessage()               {}
func (*ReceiptProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{5} }

func (m *ReceiptProposalRule) GetPrev() *AutonomyProposalRule {
	if m != nil {
		return m.Prev
	}
	return nil
}

func (m *ReceiptProposalRule) GetCurrent() *AutonomyProposalRule {
	if m != nil {
		return m.Current
	}
	return nil
}

type LocalProposalRule struct {
	PropRule *AutonomyProposalRule `protobuf:"bytes,1,opt,name=propRule" json:"propRule,omitempty"`
	Comments []string              `protobuf:"bytes,2,rep,name=comments" json:"comments,omitempty"`
}

func (m *LocalProposalRule) Reset()                    { *m = LocalProposalRule{} }
func (m *LocalProposalRule) String() string            { return proto.CompactTextString(m) }
func (*LocalProposalRule) ProtoMessage()               {}
func (*LocalProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{6} }

func (m *LocalProposalRule) GetPropRule() *AutonomyProposalRule {
	if m != nil {
		return m.PropRule
	}
	return nil
}

func (m *LocalProposalRule) GetComments() []string {
	if m != nil {
		return m.Comments
	}
	return nil
}

// query
type ReqQueryProposalRule struct {
	// 优先根据status查询
	Status    int32 `protobuf:"varint,1,opt,name=status" json:"status,omitempty"`
	Count     int32 `protobuf:"varint,2,opt,name=count" json:"count,omitempty"`
	Direction int32 `protobuf:"varint,3,opt,name=direction" json:"direction,omitempty"`
	Index     int64 `protobuf:"varint,4,opt,name=index" json:"index,omitempty"`
}

func (m *ReqQueryProposalRule) Reset()                    { *m = ReqQueryProposalRule{} }
func (m *ReqQueryProposalRule) String() string            { return proto.CompactTextString(m) }
func (*ReqQueryProposalRule) ProtoMessage()               {}
func (*ReqQueryProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{7} }

func (m *ReqQueryProposalRule) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *ReqQueryProposalRule) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *ReqQueryProposalRule) GetDirection() int32 {
	if m != nil {
		return m.Direction
	}
	return 0
}

func (m *ReqQueryProposalRule) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

type ReplyQueryProposalRule struct {
	PropRules []*AutonomyProposalRule `protobuf:"bytes,1,rep,name=propRules" json:"propRules,omitempty"`
}

func (m *ReplyQueryProposalRule) Reset()                    { *m = ReplyQueryProposalRule{} }
func (m *ReplyQueryProposalRule) String() string            { return proto.CompactTextString(m) }
func (*ReplyQueryProposalRule) ProtoMessage()               {}
func (*ReplyQueryProposalRule) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{8} }

func (m *ReplyQueryProposalRule) GetPropRules() []*AutonomyProposalRule {
	if m != nil {
		return m.PropRules
	}
	return nil
}

// TransferFund action
type TransferFund struct {
	Amount int64  `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Note   string `protobuf:"bytes,2,opt,name=note" json:"note,omitempty"`
}

func (m *TransferFund) Reset()                    { *m = TransferFund{} }
func (m *TransferFund) String() string            { return proto.CompactTextString(m) }
func (*TransferFund) ProtoMessage()               {}
func (*TransferFund) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{9} }

func (m *TransferFund) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *TransferFund) GetNote() string {
	if m != nil {
		return m.Note
	}
	return ""
}

// Comment action
type Comment struct {
	ProposalID string `protobuf:"bytes,1,opt,name=proposalID" json:"proposalID,omitempty"`
	RepCmtHash string `protobuf:"bytes,2,opt,name=repCmtHash" json:"repCmtHash,omitempty"`
	Comment    string `protobuf:"bytes,3,opt,name=comment" json:"comment,omitempty"`
}

func (m *Comment) Reset()                    { *m = Comment{} }
func (m *Comment) String() string            { return proto.CompactTextString(m) }
func (*Comment) ProtoMessage()               {}
func (*Comment) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{10} }

func (m *Comment) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

func (m *Comment) GetRepCmtHash() string {
	if m != nil {
		return m.RepCmtHash
	}
	return ""
}

func (m *Comment) GetComment() string {
	if m != nil {
		return m.Comment
	}
	return ""
}

type ReceiptProposalComment struct {
	Cmt    *Comment `protobuf:"bytes,1,opt,name=cmt" json:"cmt,omitempty"`
	Height int64    `protobuf:"varint,2,opt,name=height" json:"height,omitempty"`
	Index  int32    `protobuf:"varint,3,opt,name=index" json:"index,omitempty"`
}

func (m *ReceiptProposalComment) Reset()                    { *m = ReceiptProposalComment{} }
func (m *ReceiptProposalComment) String() string            { return proto.CompactTextString(m) }
func (*ReceiptProposalComment) ProtoMessage()               {}
func (*ReceiptProposalComment) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{11} }

func (m *ReceiptProposalComment) GetCmt() *Comment {
	if m != nil {
		return m.Cmt
	}
	return nil
}

func (m *ReceiptProposalComment) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *ReceiptProposalComment) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}

// query
type ReqQueryProposalComment struct {
	ProposalID string `protobuf:"bytes,1,opt,name=proposalID" json:"proposalID,omitempty"`
	Count      int32  `protobuf:"varint,2,opt,name=count" json:"count,omitempty"`
	Direction  int32  `protobuf:"varint,3,opt,name=direction" json:"direction,omitempty"`
	Index      int64  `protobuf:"varint,4,opt,name=index" json:"index,omitempty"`
}

func (m *ReqQueryProposalComment) Reset()                    { *m = ReqQueryProposalComment{} }
func (m *ReqQueryProposalComment) String() string            { return proto.CompactTextString(m) }
func (*ReqQueryProposalComment) ProtoMessage()               {}
func (*ReqQueryProposalComment) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{12} }

func (m *ReqQueryProposalComment) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

func (m *ReqQueryProposalComment) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *ReqQueryProposalComment) GetDirection() int32 {
	if m != nil {
		return m.Direction
	}
	return 0
}

func (m *ReqQueryProposalComment) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

type RelationCmt struct {
	RepCmtHash string `protobuf:"bytes,1,opt,name=repCmtHash" json:"repCmtHash,omitempty"`
	Comment    string `protobuf:"bytes,2,opt,name=comment" json:"comment,omitempty"`
	Height     int64  `protobuf:"varint,3,opt,name=height" json:"height,omitempty"`
	Index      int32  `protobuf:"varint,4,opt,name=index" json:"index,omitempty"`
}

func (m *RelationCmt) Reset()                    { *m = RelationCmt{} }
func (m *RelationCmt) String() string            { return proto.CompactTextString(m) }
func (*RelationCmt) ProtoMessage()               {}
func (*RelationCmt) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{13} }

func (m *RelationCmt) GetRepCmtHash() string {
	if m != nil {
		return m.RepCmtHash
	}
	return ""
}

func (m *RelationCmt) GetComment() string {
	if m != nil {
		return m.Comment
	}
	return ""
}

func (m *RelationCmt) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *RelationCmt) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}

type ReplyQueryProposalComment struct {
	RltCmt []*RelationCmt `protobuf:"bytes,1,rep,name=rltCmt" json:"rltCmt,omitempty"`
}

func (m *ReplyQueryProposalComment) Reset()                    { *m = ReplyQueryProposalComment{} }
func (m *ReplyQueryProposalComment) String() string            { return proto.CompactTextString(m) }
func (*ReplyQueryProposalComment) ProtoMessage()               {}
func (*ReplyQueryProposalComment) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{14} }

func (m *ReplyQueryProposalComment) GetRltCmt() []*RelationCmt {
	if m != nil {
		return m.RltCmt
	}
	return nil
}

func init() {
	proto.RegisterType((*AutonomyProposalRule)(nil), "types.AutonomyProposalRule")
	proto.RegisterType((*ProposalRule)(nil), "types.ProposalRule")
	proto.RegisterType((*RevokeProposalRule)(nil), "types.RevokeProposalRule")
	proto.RegisterType((*VoteProposalRule)(nil), "types.VoteProposalRule")
	proto.RegisterType((*TerminateProposalRule)(nil), "types.TerminateProposalRule")
	proto.RegisterType((*ReceiptProposalRule)(nil), "types.ReceiptProposalRule")
	proto.RegisterType((*LocalProposalRule)(nil), "types.LocalProposalRule")
	proto.RegisterType((*ReqQueryProposalRule)(nil), "types.ReqQueryProposalRule")
	proto.RegisterType((*ReplyQueryProposalRule)(nil), "types.ReplyQueryProposalRule")
	proto.RegisterType((*TransferFund)(nil), "types.TransferFund")
	proto.RegisterType((*Comment)(nil), "types.Comment")
	proto.RegisterType((*ReceiptProposalComment)(nil), "types.ReceiptProposalComment")
	proto.RegisterType((*ReqQueryProposalComment)(nil), "types.ReqQueryProposalComment")
	proto.RegisterType((*RelationCmt)(nil), "types.RelationCmt")
	proto.RegisterType((*ReplyQueryProposalComment)(nil), "types.ReplyQueryProposalComment")
}

func init() { proto.RegisterFile("rule.proto", fileDescriptor4) }

var fileDescriptor4 = []byte{
	// 683 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x55, 0x5d, 0x6f, 0xd3, 0x30,
	0x14, 0x55, 0x9a, 0xa6, 0x1f, 0x77, 0x63, 0xda, 0xbc, 0x32, 0xc2, 0x40, 0x53, 0x95, 0x07, 0x54,
	0x0d, 0xa9, 0x88, 0x2f, 0x4d, 0xf0, 0x06, 0xe5, 0x63, 0x48, 0x7b, 0x00, 0x33, 0xf1, 0x6e, 0x92,
	0xbb, 0x35, 0x5a, 0x62, 0x07, 0xc7, 0xa9, 0xd6, 0x07, 0x9e, 0xf8, 0x31, 0xfc, 0x48, 0x5e, 0x90,
	0x1d, 0x67, 0x4d, 0xda, 0x6a, 0x63, 0x12, 0x6f, 0xbe, 0xd7, 0x27, 0xd7, 0x3e, 0xe7, 0x9e, 0xeb,
	0x00, 0xc8, 0x22, 0xc1, 0x71, 0x26, 0x85, 0x12, 0xc4, 0x53, 0xf3, 0x0c, 0xf3, 0xfd, 0x3b, 0x49,
	0x28, 0xd2, 0x54, 0xf0, 0x32, 0x1b, 0xfc, 0x6e, 0xc1, 0xe0, 0x4d, 0xa1, 0x04, 0x17, 0xe9, 0xfc,
	0xb3, 0x14, 0x99, 0xc8, 0x59, 0x42, 0x8b, 0x04, 0xc9, 0x13, 0xe8, 0x65, 0x52, 0x64, 0x7a, 0xed,
	0x3b, 0x43, 0x67, 0xb4, 0xf1, 0x6c, 0x77, 0x6c, 0x2a, 0x8c, 0xeb, 0x30, 0x7a, 0x05, 0x22, 0x8f,
	0xa1, 0x1b, 0x16, 0xd2, 0xe0, 0x5b, 0x06, 0xbf, 0x63, 0xf1, 0x3a, 0x35, 0x11, 0xfc, 0x2c, 0x3e,
	0xa7, 0x15, 0x82, 0x3c, 0x05, 0x98, 0x09, 0x85, 0x14, 0xf3, 0x22, 0x51, 0xbe, 0xdb, 0xc0, 0x7f,
	0xbb, 0xda, 0xa0, 0x35, 0x10, 0xd9, 0x83, 0x4e, 0xae, 0x98, 0x2a, 0x72, 0xbf, 0x3d, 0x74, 0x46,
	0x1e, 0xb5, 0x11, 0xf1, 0xa1, 0xcb, 0xa2, 0x48, 0x62, 0x9e, 0xfb, 0xde, 0xd0, 0x19, 0xf5, 0x69,
	0x15, 0xea, 0x2f, 0xa6, 0x18, 0x9f, 0x4f, 0x95, 0xdf, 0x19, 0x3a, 0x23, 0x97, 0xda, 0x88, 0x0c,
	0xc0, 0x8b, 0x79, 0x84, 0x97, 0x7e, 0xd7, 0x14, 0x2a, 0x03, 0x72, 0x00, 0x90, 0x59, 0x66, 0x9f,
	0xde, 0xf9, 0x3d, 0x53, 0xaa, 0x96, 0x09, 0xfe, 0x38, 0xb0, 0xd9, 0x50, 0x88, 0x40, 0x7b, 0x8e,
	0x4c, 0x1a, 0x75, 0x3c, 0x6a, 0xd6, 0xba, 0x74, 0x2a, 0xb8, 0x9a, 0x1a, 0x09, 0x3c, 0x5a, 0x06,
	0x64, 0x1b, 0xdc, 0x88, 0xcd, 0x0d, 0x4d, 0x8f, 0xea, 0xa5, 0x16, 0x4b, 0xb7, 0x66, 0x72, 0x76,
	0x6e, 0xd8, 0xac, 0x17, 0xcb, 0x22, 0xc8, 0x21, 0x6c, 0xe7, 0x8a, 0x49, 0xf5, 0x36, 0x11, 0xe1,
	0xc5, 0x71, 0xc9, 0xc8, 0x33, 0x8c, 0x56, 0xf2, 0xe4, 0x11, 0x6c, 0x21, 0x8f, 0xea, 0xc8, 0x92,
	0xfb, 0x52, 0x96, 0x8c, 0x81, 0x48, 0x64, 0xc9, 0xfb, 0x26, 0xb6, 0x6b, 0xb0, 0x6b, 0x76, 0x82,
	0x17, 0x40, 0x28, 0xce, 0xc4, 0x05, 0x36, 0x24, 0x68, 0x6a, 0xe6, 0xac, 0x68, 0x76, 0x02, 0xdb,
	0xba, 0x9b, 0xb7, 0xf9, 0xc6, 0xf4, 0x33, 0xcb, 0xa4, 0x98, 0x95, 0x3e, 0xea, 0xd1, 0x2a, 0x0c,
	0x8e, 0xe0, 0xee, 0x29, 0xca, 0x34, 0xe6, 0xec, 0x76, 0x25, 0x83, 0x9f, 0xb0, 0x4b, 0x31, 0xc4,
	0x38, 0x53, 0x4b, 0x16, 0x6f, 0x67, 0x12, 0x67, 0xd6, 0xde, 0x0f, 0x6c, 0x07, 0xd6, 0x4d, 0x03,
	0x35, 0x40, 0xf2, 0xd2, 0x58, 0x5c, 0x22, 0x57, 0xd6, 0xe2, 0xd7, 0x7e, 0x53, 0x61, 0x83, 0x29,
	0xec, 0x9c, 0x88, 0x90, 0x25, 0x8d, 0xc3, 0x8f, 0x56, 0xe6, 0xeb, 0xda, 0x62, 0x8b, 0x39, 0xdb,
	0x87, 0x9e, 0x9e, 0x60, 0xe4, 0x2a, 0xf7, 0x5b, 0x43, 0x77, 0xd4, 0xa7, 0x57, 0x71, 0x70, 0x09,
	0x03, 0x8a, 0x3f, 0xbe, 0x14, 0x28, 0x9b, 0xc3, 0xbc, 0x98, 0x1d, 0xa7, 0x31, 0x3b, 0x03, 0xf0,
	0x42, 0x51, 0x58, 0x3a, 0x1e, 0x2d, 0x03, 0xf2, 0x10, 0xfa, 0x51, 0x2c, 0x31, 0x54, 0xb1, 0xe0,
	0xd6, 0xb4, 0x8b, 0xc4, 0x62, 0x7a, 0xda, 0xc6, 0x2c, 0x65, 0x10, 0x7c, 0x85, 0x3d, 0x8a, 0x59,
	0x32, 0x5f, 0x3d, 0xfb, 0x15, 0xf4, 0xab, 0xbb, 0xeb, 0xe3, 0xdd, 0x9b, 0x98, 0x2e, 0xd0, 0xc1,
	0x6b, 0xd8, 0x3c, 0x95, 0x8c, 0xe7, 0x67, 0x28, 0x3f, 0x14, 0x3c, 0xd2, 0x34, 0x58, 0x6a, 0xee,
	0xeb, 0x94, 0x03, 0x5d, 0x46, 0x7a, 0x12, 0xb9, 0x50, 0xa5, 0x5f, 0xfa, 0xd4, 0xac, 0x83, 0x10,
	0xba, 0x93, 0x52, 0x96, 0x1b, 0x1d, 0x77, 0x00, 0x20, 0x31, 0x9b, 0xa4, 0xea, 0x98, 0xe5, 0x53,
	0x5b, 0xa4, 0x96, 0xd1, 0x8e, 0xb4, 0x0a, 0x1b, 0x35, 0xfa, 0xb4, 0x0a, 0x83, 0xa9, 0x66, 0xdd,
	0x30, 0x56, 0x75, 0xe6, 0x10, 0xdc, 0x30, 0x55, 0xb6, 0xb3, 0x5b, 0x96, 0xaf, 0xdd, 0xa4, 0x7a,
	0xab, 0xf6, 0x3a, 0xb5, 0xd6, 0xbf, 0x4e, 0x6e, 0xed, 0x75, 0x0a, 0x7e, 0x39, 0x70, 0x6f, 0xb9,
	0xb5, 0xff, 0xca, 0xef, 0xff, 0x75, 0xb9, 0x80, 0x0d, 0x8a, 0x09, 0xd3, 0x88, 0x49, 0xaa, 0x96,
	0x84, 0x73, 0xae, 0x13, 0xae, 0xd5, 0x10, 0xae, 0x46, 0xde, 0x5d, 0x4f, 0xbe, 0x5d, 0x27, 0xff,
	0x11, 0xee, 0xaf, 0x9a, 0xab, 0x62, 0x7f, 0x08, 0x1d, 0x99, 0xa8, 0x89, 0x11, 0x5b, 0x9b, 0x8b,
	0x54, 0x2f, 0xe9, 0xe2, 0xa2, 0xd4, 0x22, 0xbe, 0x77, 0xcc, 0x4f, 0xef, 0xf9, 0xdf, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x34, 0x4a, 0x95, 0xdd, 0x18, 0x07, 0x00, 0x00,
}
