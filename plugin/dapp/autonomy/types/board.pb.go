// Code generated by protoc-gen-go. DO NOT EDIT.
// source: board.proto

package types

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type AutonomyProposalBoard struct {
	PropBoard *ProposalBoard `protobuf:"bytes,1,opt,name=propBoard,proto3" json:"propBoard,omitempty"`
	// 投票该提案的规则
	CurRule *RuleConfig `protobuf:"bytes,2,opt,name=curRule,proto3" json:"curRule,omitempty"`
	// 投票董事会
	Board *ActiveBoard `protobuf:"bytes,3,opt,name=board,proto3" json:"board,omitempty"`
	// 全体持票人投票结果
	VoteResult *VoteResult `protobuf:"bytes,4,opt,name=voteResult,proto3" json:"voteResult,omitempty"`
	// 状态
	Status               int32    `protobuf:"varint,5,opt,name=status,proto3" json:"status,omitempty"`
	Address              string   `protobuf:"bytes,6,opt,name=address,proto3" json:"address,omitempty"`
	Height               int64    `protobuf:"varint,7,opt,name=height,proto3" json:"height,omitempty"`
	Index                int32    `protobuf:"varint,8,opt,name=index,proto3" json:"index,omitempty"`
	ProposalID           string   `protobuf:"bytes,9,opt,name=proposalID,proto3" json:"proposalID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AutonomyProposalBoard) Reset()         { *m = AutonomyProposalBoard{} }
func (m *AutonomyProposalBoard) String() string { return proto.CompactTextString(m) }
func (*AutonomyProposalBoard) ProtoMessage()    {}
func (*AutonomyProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{0}
}

func (m *AutonomyProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AutonomyProposalBoard.Unmarshal(m, b)
}
func (m *AutonomyProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AutonomyProposalBoard.Marshal(b, m, deterministic)
}
func (m *AutonomyProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AutonomyProposalBoard.Merge(m, src)
}
func (m *AutonomyProposalBoard) XXX_Size() int {
	return xxx_messageInfo_AutonomyProposalBoard.Size(m)
}
func (m *AutonomyProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_AutonomyProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_AutonomyProposalBoard proto.InternalMessageInfo

func (m *AutonomyProposalBoard) GetPropBoard() *ProposalBoard {
	if m != nil {
		return m.PropBoard
	}
	return nil
}

func (m *AutonomyProposalBoard) GetCurRule() *RuleConfig {
	if m != nil {
		return m.CurRule
	}
	return nil
}

func (m *AutonomyProposalBoard) GetBoard() *ActiveBoard {
	if m != nil {
		return m.Board
	}
	return nil
}

func (m *AutonomyProposalBoard) GetVoteResult() *VoteResult {
	if m != nil {
		return m.VoteResult
	}
	return nil
}

func (m *AutonomyProposalBoard) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *AutonomyProposalBoard) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *AutonomyProposalBoard) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *AutonomyProposalBoard) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *AutonomyProposalBoard) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

// action
type ProposalBoard struct {
	// 提案时间
	Year  int32 `protobuf:"varint,1,opt,name=year,proto3" json:"year,omitempty"`
	Month int32 `protobuf:"varint,2,opt,name=month,proto3" json:"month,omitempty"`
	Day   int32 `protobuf:"varint,3,opt,name=day,proto3" json:"day,omitempty"`
	// 是否更新
	Update bool `protobuf:"varint,4,opt,name=update,proto3" json:"update,omitempty"`
	// 提案董事会成员
	Boards []string `protobuf:"bytes,5,rep,name=boards,proto3" json:"boards,omitempty"`
	// 投票相关
	StartBlockHeight     int64    `protobuf:"varint,6,opt,name=startBlockHeight,proto3" json:"startBlockHeight,omitempty"`
	EndBlockHeight       int64    `protobuf:"varint,7,opt,name=endBlockHeight,proto3" json:"endBlockHeight,omitempty"`
	RealEndBlockHeight   int64    `protobuf:"varint,8,opt,name=realEndBlockHeight,proto3" json:"realEndBlockHeight,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProposalBoard) Reset()         { *m = ProposalBoard{} }
func (m *ProposalBoard) String() string { return proto.CompactTextString(m) }
func (*ProposalBoard) ProtoMessage()    {}
func (*ProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{1}
}

func (m *ProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProposalBoard.Unmarshal(m, b)
}
func (m *ProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProposalBoard.Marshal(b, m, deterministic)
}
func (m *ProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProposalBoard.Merge(m, src)
}
func (m *ProposalBoard) XXX_Size() int {
	return xxx_messageInfo_ProposalBoard.Size(m)
}
func (m *ProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_ProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_ProposalBoard proto.InternalMessageInfo

func (m *ProposalBoard) GetYear() int32 {
	if m != nil {
		return m.Year
	}
	return 0
}

func (m *ProposalBoard) GetMonth() int32 {
	if m != nil {
		return m.Month
	}
	return 0
}

func (m *ProposalBoard) GetDay() int32 {
	if m != nil {
		return m.Day
	}
	return 0
}

func (m *ProposalBoard) GetUpdate() bool {
	if m != nil {
		return m.Update
	}
	return false
}

func (m *ProposalBoard) GetBoards() []string {
	if m != nil {
		return m.Boards
	}
	return nil
}

func (m *ProposalBoard) GetStartBlockHeight() int64 {
	if m != nil {
		return m.StartBlockHeight
	}
	return 0
}

func (m *ProposalBoard) GetEndBlockHeight() int64 {
	if m != nil {
		return m.EndBlockHeight
	}
	return 0
}

func (m *ProposalBoard) GetRealEndBlockHeight() int64 {
	if m != nil {
		return m.RealEndBlockHeight
	}
	return 0
}

type RevokeProposalBoard struct {
	ProposalID           string   `protobuf:"bytes,1,opt,name=proposalID,proto3" json:"proposalID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RevokeProposalBoard) Reset()         { *m = RevokeProposalBoard{} }
func (m *RevokeProposalBoard) String() string { return proto.CompactTextString(m) }
func (*RevokeProposalBoard) ProtoMessage()    {}
func (*RevokeProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{2}
}

func (m *RevokeProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RevokeProposalBoard.Unmarshal(m, b)
}
func (m *RevokeProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RevokeProposalBoard.Marshal(b, m, deterministic)
}
func (m *RevokeProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RevokeProposalBoard.Merge(m, src)
}
func (m *RevokeProposalBoard) XXX_Size() int {
	return xxx_messageInfo_RevokeProposalBoard.Size(m)
}
func (m *RevokeProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_RevokeProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_RevokeProposalBoard proto.InternalMessageInfo

func (m *RevokeProposalBoard) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

type VoteProposalBoard struct {
	ProposalID           string   `protobuf:"bytes,1,opt,name=proposalID,proto3" json:"proposalID,omitempty"`
	Approve              bool     `protobuf:"varint,2,opt,name=approve,proto3" json:"approve,omitempty"`
	OriginAddr           []string `protobuf:"bytes,3,rep,name=originAddr,proto3" json:"originAddr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *VoteProposalBoard) Reset()         { *m = VoteProposalBoard{} }
func (m *VoteProposalBoard) String() string { return proto.CompactTextString(m) }
func (*VoteProposalBoard) ProtoMessage()    {}
func (*VoteProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{3}
}

func (m *VoteProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VoteProposalBoard.Unmarshal(m, b)
}
func (m *VoteProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VoteProposalBoard.Marshal(b, m, deterministic)
}
func (m *VoteProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VoteProposalBoard.Merge(m, src)
}
func (m *VoteProposalBoard) XXX_Size() int {
	return xxx_messageInfo_VoteProposalBoard.Size(m)
}
func (m *VoteProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_VoteProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_VoteProposalBoard proto.InternalMessageInfo

func (m *VoteProposalBoard) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

func (m *VoteProposalBoard) GetApprove() bool {
	if m != nil {
		return m.Approve
	}
	return false
}

func (m *VoteProposalBoard) GetOriginAddr() []string {
	if m != nil {
		return m.OriginAddr
	}
	return nil
}

type TerminateProposalBoard struct {
	ProposalID           string   `protobuf:"bytes,1,opt,name=proposalID,proto3" json:"proposalID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TerminateProposalBoard) Reset()         { *m = TerminateProposalBoard{} }
func (m *TerminateProposalBoard) String() string { return proto.CompactTextString(m) }
func (*TerminateProposalBoard) ProtoMessage()    {}
func (*TerminateProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{4}
}

func (m *TerminateProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TerminateProposalBoard.Unmarshal(m, b)
}
func (m *TerminateProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TerminateProposalBoard.Marshal(b, m, deterministic)
}
func (m *TerminateProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TerminateProposalBoard.Merge(m, src)
}
func (m *TerminateProposalBoard) XXX_Size() int {
	return xxx_messageInfo_TerminateProposalBoard.Size(m)
}
func (m *TerminateProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_TerminateProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_TerminateProposalBoard proto.InternalMessageInfo

func (m *TerminateProposalBoard) GetProposalID() string {
	if m != nil {
		return m.ProposalID
	}
	return ""
}

// receipt
type ReceiptProposalBoard struct {
	Prev                 *AutonomyProposalBoard `protobuf:"bytes,1,opt,name=prev,proto3" json:"prev,omitempty"`
	Current              *AutonomyProposalBoard `protobuf:"bytes,2,opt,name=current,proto3" json:"current,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *ReceiptProposalBoard) Reset()         { *m = ReceiptProposalBoard{} }
func (m *ReceiptProposalBoard) String() string { return proto.CompactTextString(m) }
func (*ReceiptProposalBoard) ProtoMessage()    {}
func (*ReceiptProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{5}
}

func (m *ReceiptProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReceiptProposalBoard.Unmarshal(m, b)
}
func (m *ReceiptProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReceiptProposalBoard.Marshal(b, m, deterministic)
}
func (m *ReceiptProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReceiptProposalBoard.Merge(m, src)
}
func (m *ReceiptProposalBoard) XXX_Size() int {
	return xxx_messageInfo_ReceiptProposalBoard.Size(m)
}
func (m *ReceiptProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_ReceiptProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_ReceiptProposalBoard proto.InternalMessageInfo

func (m *ReceiptProposalBoard) GetPrev() *AutonomyProposalBoard {
	if m != nil {
		return m.Prev
	}
	return nil
}

func (m *ReceiptProposalBoard) GetCurrent() *AutonomyProposalBoard {
	if m != nil {
		return m.Current
	}
	return nil
}

type LocalProposalBoard struct {
	PropBd               *AutonomyProposalBoard `protobuf:"bytes,1,opt,name=propBd,proto3" json:"propBd,omitempty"`
	Comments             []string               `protobuf:"bytes,2,rep,name=comments,proto3" json:"comments,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *LocalProposalBoard) Reset()         { *m = LocalProposalBoard{} }
func (m *LocalProposalBoard) String() string { return proto.CompactTextString(m) }
func (*LocalProposalBoard) ProtoMessage()    {}
func (*LocalProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{6}
}

func (m *LocalProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalProposalBoard.Unmarshal(m, b)
}
func (m *LocalProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalProposalBoard.Marshal(b, m, deterministic)
}
func (m *LocalProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalProposalBoard.Merge(m, src)
}
func (m *LocalProposalBoard) XXX_Size() int {
	return xxx_messageInfo_LocalProposalBoard.Size(m)
}
func (m *LocalProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_LocalProposalBoard proto.InternalMessageInfo

func (m *LocalProposalBoard) GetPropBd() *AutonomyProposalBoard {
	if m != nil {
		return m.PropBd
	}
	return nil
}

func (m *LocalProposalBoard) GetComments() []string {
	if m != nil {
		return m.Comments
	}
	return nil
}

// query
type ReqQueryProposalBoard struct {
	Status               int32    `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	Addr                 string   `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
	Count                int32    `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	Direction            int32    `protobuf:"varint,4,opt,name=direction,proto3" json:"direction,omitempty"`
	Height               int64    `protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty"`
	Index                int32    `protobuf:"varint,6,opt,name=index,proto3" json:"index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReqQueryProposalBoard) Reset()         { *m = ReqQueryProposalBoard{} }
func (m *ReqQueryProposalBoard) String() string { return proto.CompactTextString(m) }
func (*ReqQueryProposalBoard) ProtoMessage()    {}
func (*ReqQueryProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{7}
}

func (m *ReqQueryProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReqQueryProposalBoard.Unmarshal(m, b)
}
func (m *ReqQueryProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReqQueryProposalBoard.Marshal(b, m, deterministic)
}
func (m *ReqQueryProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReqQueryProposalBoard.Merge(m, src)
}
func (m *ReqQueryProposalBoard) XXX_Size() int {
	return xxx_messageInfo_ReqQueryProposalBoard.Size(m)
}
func (m *ReqQueryProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_ReqQueryProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_ReqQueryProposalBoard proto.InternalMessageInfo

func (m *ReqQueryProposalBoard) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *ReqQueryProposalBoard) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *ReqQueryProposalBoard) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *ReqQueryProposalBoard) GetDirection() int32 {
	if m != nil {
		return m.Direction
	}
	return 0
}

func (m *ReqQueryProposalBoard) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *ReqQueryProposalBoard) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}

type ReplyQueryProposalBoard struct {
	PropBoards           []*AutonomyProposalBoard `protobuf:"bytes,1,rep,name=propBoards,proto3" json:"propBoards,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *ReplyQueryProposalBoard) Reset()         { *m = ReplyQueryProposalBoard{} }
func (m *ReplyQueryProposalBoard) String() string { return proto.CompactTextString(m) }
func (*ReplyQueryProposalBoard) ProtoMessage()    {}
func (*ReplyQueryProposalBoard) Descriptor() ([]byte, []int) {
	return fileDescriptor_937f74b042f92c0f, []int{8}
}

func (m *ReplyQueryProposalBoard) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReplyQueryProposalBoard.Unmarshal(m, b)
}
func (m *ReplyQueryProposalBoard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReplyQueryProposalBoard.Marshal(b, m, deterministic)
}
func (m *ReplyQueryProposalBoard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReplyQueryProposalBoard.Merge(m, src)
}
func (m *ReplyQueryProposalBoard) XXX_Size() int {
	return xxx_messageInfo_ReplyQueryProposalBoard.Size(m)
}
func (m *ReplyQueryProposalBoard) XXX_DiscardUnknown() {
	xxx_messageInfo_ReplyQueryProposalBoard.DiscardUnknown(m)
}

var xxx_messageInfo_ReplyQueryProposalBoard proto.InternalMessageInfo

func (m *ReplyQueryProposalBoard) GetPropBoards() []*AutonomyProposalBoard {
	if m != nil {
		return m.PropBoards
	}
	return nil
}

func init() {
	proto.RegisterType((*AutonomyProposalBoard)(nil), "types.AutonomyProposalBoard")
	proto.RegisterType((*ProposalBoard)(nil), "types.ProposalBoard")
	proto.RegisterType((*RevokeProposalBoard)(nil), "types.RevokeProposalBoard")
	proto.RegisterType((*VoteProposalBoard)(nil), "types.VoteProposalBoard")
	proto.RegisterType((*TerminateProposalBoard)(nil), "types.TerminateProposalBoard")
	proto.RegisterType((*ReceiptProposalBoard)(nil), "types.ReceiptProposalBoard")
	proto.RegisterType((*LocalProposalBoard)(nil), "types.LocalProposalBoard")
	proto.RegisterType((*ReqQueryProposalBoard)(nil), "types.ReqQueryProposalBoard")
	proto.RegisterType((*ReplyQueryProposalBoard)(nil), "types.ReplyQueryProposalBoard")
}

func init() {
	proto.RegisterFile("board.proto", fileDescriptor_937f74b042f92c0f)
}

var fileDescriptor_937f74b042f92c0f = []byte{
	// 587 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0x4d, 0x6f, 0xd4, 0x30,
	0x10, 0x55, 0x9a, 0x4d, 0xba, 0x99, 0xaa, 0xa8, 0x35, 0x6d, 0xb1, 0xaa, 0x0a, 0xad, 0x72, 0x40,
	0x11, 0x48, 0x2b, 0x28, 0x1f, 0xe2, 0xc0, 0xa5, 0x05, 0x24, 0x90, 0x38, 0x80, 0x85, 0xe0, 0x9c,
	0x26, 0xd3, 0x6e, 0xd4, 0xc4, 0x36, 0x8e, 0xb3, 0x62, 0x6f, 0xfc, 0x19, 0x7e, 0x12, 0xbf, 0x07,
	0xe4, 0x49, 0xb6, 0x9b, 0xb4, 0x2b, 0xa0, 0x37, 0xbf, 0xc9, 0x9b, 0xc9, 0xcc, 0xf3, 0x3c, 0xc3,
	0xd6, 0x99, 0x4a, 0x4d, 0x3e, 0xd5, 0x46, 0x59, 0xc5, 0x02, 0xbb, 0xd0, 0x58, 0x1f, 0x6e, 0x97,
	0x99, 0xaa, 0x2a, 0x25, 0xdb, 0x68, 0xfc, 0x6b, 0x03, 0xf6, 0x4f, 0x1a, 0xab, 0xa4, 0xaa, 0x16,
	0x1f, 0x8d, 0xd2, 0xaa, 0x4e, 0xcb, 0x53, 0x97, 0xc5, 0x8e, 0x21, 0xd2, 0x46, 0x69, 0x02, 0xdc,
	0x9b, 0x78, 0xc9, 0xd6, 0xf1, 0xde, 0x94, 0x6a, 0x4c, 0x07, 0x44, 0xb1, 0xa2, 0xb1, 0x47, 0xb0,
	0x99, 0x35, 0x46, 0x34, 0x25, 0xf2, 0x0d, 0xca, 0xd8, 0xed, 0x32, 0x5c, 0xe8, 0xb5, 0x92, 0xe7,
	0xc5, 0x85, 0x58, 0x32, 0x58, 0x02, 0x01, 0xf5, 0xc7, 0x7d, 0xa2, 0xb2, 0x8e, 0x7a, 0x92, 0xd9,
	0x62, 0x8e, 0x6d, 0xe9, 0x96, 0xc0, 0x9e, 0x00, 0xcc, 0x95, 0x45, 0x81, 0x75, 0x53, 0x5a, 0x3e,
	0x1a, 0x54, 0xfe, 0x72, 0xf5, 0x41, 0xf4, 0x48, 0xec, 0x00, 0xc2, 0xda, 0xa6, 0xb6, 0xa9, 0x79,
	0x30, 0xf1, 0x92, 0x40, 0x74, 0x88, 0x71, 0xd8, 0x4c, 0xf3, 0xdc, 0x60, 0x5d, 0xf3, 0x70, 0xe2,
	0x25, 0x91, 0x58, 0x42, 0x97, 0x31, 0xc3, 0xe2, 0x62, 0x66, 0xf9, 0xe6, 0xc4, 0x4b, 0x7c, 0xd1,
	0x21, 0xb6, 0x07, 0x41, 0x21, 0x73, 0xfc, 0xce, 0xc7, 0x54, 0xa8, 0x05, 0xec, 0x3e, 0x80, 0xee,
	0x54, 0x78, 0xff, 0x86, 0x47, 0x54, 0xaa, 0x17, 0x89, 0x7f, 0x7b, 0xb0, 0x3d, 0xd4, 0x93, 0xc1,
	0x68, 0x81, 0xa9, 0x21, 0x29, 0x03, 0x41, 0x67, 0x57, 0xbb, 0x52, 0xd2, 0xce, 0x48, 0xad, 0x40,
	0xb4, 0x80, 0xed, 0x80, 0x9f, 0xa7, 0x0b, 0x92, 0x25, 0x10, 0xee, 0xe8, 0x7a, 0x6b, 0x74, 0x9e,
	0x5a, 0xa4, 0xe1, 0xc7, 0xa2, 0x43, 0x2e, 0x4e, 0x0a, 0xb9, 0x29, 0xfd, 0x24, 0x12, 0x1d, 0x62,
	0x0f, 0x61, 0xa7, 0xb6, 0xa9, 0xb1, 0xa7, 0xa5, 0xca, 0x2e, 0xdf, 0xb5, 0x53, 0x85, 0x34, 0xd5,
	0x8d, 0x38, 0x7b, 0x00, 0x77, 0x50, 0xe6, 0x7d, 0x66, 0x3b, 0xff, 0xb5, 0x28, 0x9b, 0x02, 0x33,
	0x98, 0x96, 0x6f, 0x87, 0xdc, 0x31, 0x71, 0xd7, 0x7c, 0x89, 0x9f, 0xc3, 0x5d, 0x81, 0x73, 0x75,
	0x89, 0x43, 0x19, 0x86, 0xc2, 0x79, 0x37, 0x84, 0xab, 0x60, 0xd7, 0x5d, 0xe9, 0xad, 0x92, 0xe8,
	0x56, 0xb5, 0x36, 0x6a, 0xde, 0xee, 0xdd, 0x58, 0x2c, 0xa1, 0xcb, 0x54, 0xa6, 0xb8, 0x28, 0xe4,
	0x49, 0x9e, 0x1b, 0xee, 0x93, 0x4a, 0xbd, 0x48, 0xfc, 0x12, 0x0e, 0x3e, 0xa3, 0xa9, 0x0a, 0x99,
	0xde, 0xf2, 0x9f, 0xf1, 0x0f, 0x0f, 0xf6, 0x04, 0x66, 0x58, 0x68, 0x3b, 0x4c, 0x7c, 0x0c, 0x23,
	0x6d, 0x70, 0xde, 0x79, 0xe6, 0x68, 0xb9, 0xd6, 0xeb, 0x4c, 0x26, 0x88, 0xc9, 0x5e, 0x90, 0x6d,
	0x0c, 0x4a, 0xdb, 0xd9, 0xe6, 0xef, 0x49, 0x4b, 0x72, 0x7c, 0x0e, 0xec, 0x83, 0xca, 0xd2, 0x72,
	0xf8, 0xff, 0x67, 0x10, 0x92, 0x23, 0xf3, 0xff, 0xea, 0xa0, 0xe3, 0xb2, 0x43, 0x18, 0xbb, 0x87,
	0x01, 0xa5, 0xad, 0xf9, 0x06, 0xc9, 0x74, 0x85, 0xe3, 0x9f, 0x1e, 0xec, 0x0b, 0xfc, 0xf6, 0xa9,
	0x41, 0x73, 0xed, 0x91, 0x58, 0xd9, 0xcc, 0x1b, 0xd8, 0x8c, 0xc1, 0xc8, 0xf9, 0x8a, 0xc6, 0x89,
	0x04, 0x9d, 0xdd, 0xb2, 0x67, 0xaa, 0x91, 0xb6, 0x5b, 0xec, 0x16, 0xb0, 0x23, 0x88, 0xf2, 0xc2,
	0x60, 0x66, 0x0b, 0x25, 0x69, 0xbb, 0x03, 0xb1, 0x0a, 0xf4, 0x4c, 0x19, 0xac, 0x37, 0x65, 0xd8,
	0x33, 0x65, 0xfc, 0x15, 0xee, 0x09, 0xd4, 0xe5, 0x62, 0x4d, 0xa3, 0xaf, 0xda, 0xdb, 0x3c, 0x6d,
	0xdd, 0xe2, 0x4d, 0xfc, 0x7f, 0x0a, 0xd3, 0xe3, 0x9f, 0x85, 0xf4, 0x58, 0x3e, 0xfd, 0x13, 0x00,
	0x00, 0xff, 0xff, 0xa1, 0x53, 0x3b, 0x4b, 0x51, 0x05, 0x00, 0x00,
}
