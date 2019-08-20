// Code generated by protoc-gen-go. DO NOT EDIT.
// source: game.proto

package types

import (
	fmt "fmt"

	proto "github.com/golang/protobuf/proto"

	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Game struct {
	// 默认是由创建这局游戏的txHash作为gameId
	GameId string `protobuf:"bytes,1,opt,name=gameId,proto3" json:"gameId,omitempty"`
	// create 1 -> Match 2 -> Cancel 3 -> Close 4
	Status int32 `protobuf:"varint,2,opt,name=status,proto3" json:"status,omitempty"`
	// 创建时间
	CreateTime int64 `protobuf:"varint,3,opt,name=createTime,proto3" json:"createTime,omitempty"`
	// 匹配时间(何时参与对赌）
	MatchTime int64 `protobuf:"varint,4,opt,name=matchTime,proto3" json:"matchTime,omitempty"`
	// 状态close的时间（包括cancel）
	Closetime int64 `protobuf:"varint,5,opt,name=closetime,proto3" json:"closetime,omitempty"`
	// 赌注
	Value int64 `protobuf:"varint,6,opt,name=value,proto3" json:"value,omitempty"`
	// 发起者账号地址
	CreateAddress string `protobuf:"bytes,7,opt,name=createAddress,proto3" json:"createAddress,omitempty"`
	// 对赌者账号地址
	MatchAddress string `protobuf:"bytes,8,opt,name=matchAddress,proto3" json:"matchAddress,omitempty"`
	// hash 类型，预留字段
	HashType string `protobuf:"bytes,9,opt,name=hashType,proto3" json:"hashType,omitempty"`
	// 庄家创建游戏时，庄家自己出拳结果加密后的hash值
	HashValue []byte `protobuf:"bytes,10,opt,name=hashValue,proto3" json:"hashValue,omitempty"`
	// 用来公布庄家出拳结果的私钥
	Secret string `protobuf:"bytes,11,opt,name=secret,proto3" json:"secret,omitempty"`
	// 1 平局，2 庄家获胜，3 matcher获胜，4
	// 庄家开奖超时，matcher获胜，并获得本局所有赌资
	Result int32 `protobuf:"varint,12,opt,name=result,proto3" json:"result,omitempty"`
	// matcher 出拳结果
	MatcherGuess int32 `protobuf:"varint,13,opt,name=matcherGuess,proto3" json:"matcherGuess,omitempty"`
	// create txHash
	CreateTxHash string `protobuf:"bytes,14,opt,name=createTxHash,proto3" json:"createTxHash,omitempty"`
	// matche交易hash
	MatchTxHash string `protobuf:"bytes,15,opt,name=matchTxHash,proto3" json:"matchTxHash,omitempty"`
	// close txhash
	CloseTxHash string `protobuf:"bytes,16,opt,name=closeTxHash,proto3" json:"closeTxHash,omitempty"`
	// cancel txhash
	CancelTxHash         string   `protobuf:"bytes,17,opt,name=cancelTxHash,proto3" json:"cancelTxHash,omitempty"`
	Index                int64    `protobuf:"varint,18,opt,name=index,proto3" json:"index,omitempty"`
	PrevIndex            int64    `protobuf:"varint,19,opt,name=prevIndex,proto3" json:"prevIndex,omitempty"`
	CreatorGuess         int32    `protobuf:"varint,20,opt,name=creatorGuess,proto3" json:"creatorGuess,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Game) Reset()         { *m = Game{} }
func (m *Game) String() string { return proto.CompactTextString(m) }
func (*Game) ProtoMessage()    {}
func (*Game) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{0}
}
func (m *Game) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Game.Unmarshal(m, b)
}
func (m *Game) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Game.Marshal(b, m, deterministic)
}
func (dst *Game) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Game.Merge(dst, src)
}
func (m *Game) XXX_Size() int {
	return xxx_messageInfo_Game.Size(m)
}
func (m *Game) XXX_DiscardUnknown() {
	xxx_messageInfo_Game.DiscardUnknown(m)
}

var xxx_messageInfo_Game proto.InternalMessageInfo

func (m *Game) GetGameId() string {
	if m != nil {
		return m.GameId
	}
	return ""
}

func (m *Game) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *Game) GetCreateTime() int64 {
	if m != nil {
		return m.CreateTime
	}
	return 0
}

func (m *Game) GetMatchTime() int64 {
	if m != nil {
		return m.MatchTime
	}
	return 0
}

func (m *Game) GetClosetime() int64 {
	if m != nil {
		return m.Closetime
	}
	return 0
}

func (m *Game) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *Game) GetCreateAddress() string {
	if m != nil {
		return m.CreateAddress
	}
	return ""
}

func (m *Game) GetMatchAddress() string {
	if m != nil {
		return m.MatchAddress
	}
	return ""
}

func (m *Game) GetHashType() string {
	if m != nil {
		return m.HashType
	}
	return ""
}

func (m *Game) GetHashValue() []byte {
	if m != nil {
		return m.HashValue
	}
	return nil
}

func (m *Game) GetSecret() string {
	if m != nil {
		return m.Secret
	}
	return ""
}

func (m *Game) GetResult() int32 {
	if m != nil {
		return m.Result
	}
	return 0
}

func (m *Game) GetMatcherGuess() int32 {
	if m != nil {
		return m.MatcherGuess
	}
	return 0
}

func (m *Game) GetCreateTxHash() string {
	if m != nil {
		return m.CreateTxHash
	}
	return ""
}

func (m *Game) GetMatchTxHash() string {
	if m != nil {
		return m.MatchTxHash
	}
	return ""
}

func (m *Game) GetCloseTxHash() string {
	if m != nil {
		return m.CloseTxHash
	}
	return ""
}

func (m *Game) GetCancelTxHash() string {
	if m != nil {
		return m.CancelTxHash
	}
	return ""
}

func (m *Game) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *Game) GetPrevIndex() int64 {
	if m != nil {
		return m.PrevIndex
	}
	return 0
}

func (m *Game) GetCreatorGuess() int32 {
	if m != nil {
		return m.CreatorGuess
	}
	return 0
}

// message for execs.game
type GameAction struct {
	// Types that are valid to be assigned to Value:
	//	*GameAction_Create
	//	*GameAction_Cancel
	//	*GameAction_Close
	//	*GameAction_Match
	Value                isGameAction_Value `protobuf_oneof:"value"`
	Ty                   int32              `protobuf:"varint,10,opt,name=ty,proto3" json:"ty,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *GameAction) Reset()         { *m = GameAction{} }
func (m *GameAction) String() string { return proto.CompactTextString(m) }
func (*GameAction) ProtoMessage()    {}
func (*GameAction) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{1}
}
func (m *GameAction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameAction.Unmarshal(m, b)
}
func (m *GameAction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameAction.Marshal(b, m, deterministic)
}
func (dst *GameAction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameAction.Merge(dst, src)
}
func (m *GameAction) XXX_Size() int {
	return xxx_messageInfo_GameAction.Size(m)
}
func (m *GameAction) XXX_DiscardUnknown() {
	xxx_messageInfo_GameAction.DiscardUnknown(m)
}

var xxx_messageInfo_GameAction proto.InternalMessageInfo

type isGameAction_Value interface {
	isGameAction_Value()
}

type GameAction_Create struct {
	Create *GameCreate `protobuf:"bytes,1,opt,name=create,proto3,oneof"`
}

type GameAction_Cancel struct {
	Cancel *GameCancel `protobuf:"bytes,2,opt,name=cancel,proto3,oneof"`
}

type GameAction_Close struct {
	Close *GameClose `protobuf:"bytes,3,opt,name=close,proto3,oneof"`
}

type GameAction_Match struct {
	Match *GameMatch `protobuf:"bytes,4,opt,name=match,proto3,oneof"`
}

func (*GameAction_Create) isGameAction_Value() {}

func (*GameAction_Cancel) isGameAction_Value() {}

func (*GameAction_Close) isGameAction_Value() {}

func (*GameAction_Match) isGameAction_Value() {}

func (m *GameAction) GetValue() isGameAction_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *GameAction) GetCreate() *GameCreate {
	if x, ok := m.GetValue().(*GameAction_Create); ok {
		return x.Create
	}
	return nil
}

func (m *GameAction) GetCancel() *GameCancel {
	if x, ok := m.GetValue().(*GameAction_Cancel); ok {
		return x.Cancel
	}
	return nil
}

func (m *GameAction) GetClose() *GameClose {
	if x, ok := m.GetValue().(*GameAction_Close); ok {
		return x.Close
	}
	return nil
}

func (m *GameAction) GetMatch() *GameMatch {
	if x, ok := m.GetValue().(*GameAction_Match); ok {
		return x.Match
	}
	return nil
}

func (m *GameAction) GetTy() int32 {
	if m != nil {
		return m.Ty
	}
	return 0
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*GameAction) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _GameAction_OneofMarshaler, _GameAction_OneofUnmarshaler, _GameAction_OneofSizer, []interface{}{
		(*GameAction_Create)(nil),
		(*GameAction_Cancel)(nil),
		(*GameAction_Close)(nil),
		(*GameAction_Match)(nil),
	}
}

func _GameAction_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*GameAction)
	// value
	switch x := m.Value.(type) {
	case *GameAction_Create:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Create); err != nil {
			return err
		}
	case *GameAction_Cancel:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Cancel); err != nil {
			return err
		}
	case *GameAction_Close:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Close); err != nil {
			return err
		}
	case *GameAction_Match:
		b.EncodeVarint(4<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Match); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("GameAction.Value has unexpected type %T", x)
	}
	return nil
}

func _GameAction_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*GameAction)
	switch tag {
	case 1: // value.create
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(GameCreate)
		err := b.DecodeMessage(msg)
		m.Value = &GameAction_Create{msg}
		return true, err
	case 2: // value.cancel
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(GameCancel)
		err := b.DecodeMessage(msg)
		m.Value = &GameAction_Cancel{msg}
		return true, err
	case 3: // value.close
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(GameClose)
		err := b.DecodeMessage(msg)
		m.Value = &GameAction_Close{msg}
		return true, err
	case 4: // value.match
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(GameMatch)
		err := b.DecodeMessage(msg)
		m.Value = &GameAction_Match{msg}
		return true, err
	default:
		return false, nil
	}
}

func _GameAction_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*GameAction)
	// value
	switch x := m.Value.(type) {
	case *GameAction_Create:
		s := proto.Size(x.Create)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GameAction_Cancel:
		s := proto.Size(x.Cancel)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GameAction_Close:
		s := proto.Size(x.Close)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GameAction_Match:
		s := proto.Size(x.Match)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type GameMatch struct {
	GameId               string   `protobuf:"bytes,1,opt,name=gameId,proto3" json:"gameId,omitempty"`
	Guess                int32    `protobuf:"varint,2,opt,name=guess,proto3" json:"guess,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameMatch) Reset()         { *m = GameMatch{} }
func (m *GameMatch) String() string { return proto.CompactTextString(m) }
func (*GameMatch) ProtoMessage()    {}
func (*GameMatch) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{2}
}
func (m *GameMatch) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameMatch.Unmarshal(m, b)
}
func (m *GameMatch) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameMatch.Marshal(b, m, deterministic)
}
func (dst *GameMatch) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameMatch.Merge(dst, src)
}
func (m *GameMatch) XXX_Size() int {
	return xxx_messageInfo_GameMatch.Size(m)
}
func (m *GameMatch) XXX_DiscardUnknown() {
	xxx_messageInfo_GameMatch.DiscardUnknown(m)
}

var xxx_messageInfo_GameMatch proto.InternalMessageInfo

func (m *GameMatch) GetGameId() string {
	if m != nil {
		return m.GameId
	}
	return ""
}

func (m *GameMatch) GetGuess() int32 {
	if m != nil {
		return m.Guess
	}
	return 0
}

type GameCancel struct {
	GameId               string   `protobuf:"bytes,1,opt,name=gameId,proto3" json:"gameId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameCancel) Reset()         { *m = GameCancel{} }
func (m *GameCancel) String() string { return proto.CompactTextString(m) }
func (*GameCancel) ProtoMessage()    {}
func (*GameCancel) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{3}
}
func (m *GameCancel) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameCancel.Unmarshal(m, b)
}
func (m *GameCancel) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameCancel.Marshal(b, m, deterministic)
}
func (dst *GameCancel) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameCancel.Merge(dst, src)
}
func (m *GameCancel) XXX_Size() int {
	return xxx_messageInfo_GameCancel.Size(m)
}
func (m *GameCancel) XXX_DiscardUnknown() {
	xxx_messageInfo_GameCancel.DiscardUnknown(m)
}

var xxx_messageInfo_GameCancel proto.InternalMessageInfo

func (m *GameCancel) GetGameId() string {
	if m != nil {
		return m.GameId
	}
	return ""
}

type GameClose struct {
	GameId               string   `protobuf:"bytes,1,opt,name=gameId,proto3" json:"gameId,omitempty"`
	Secret               string   `protobuf:"bytes,2,opt,name=secret,proto3" json:"secret,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameClose) Reset()         { *m = GameClose{} }
func (m *GameClose) String() string { return proto.CompactTextString(m) }
func (*GameClose) ProtoMessage()    {}
func (*GameClose) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{4}
}
func (m *GameClose) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameClose.Unmarshal(m, b)
}
func (m *GameClose) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameClose.Marshal(b, m, deterministic)
}
func (dst *GameClose) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameClose.Merge(dst, src)
}
func (m *GameClose) XXX_Size() int {
	return xxx_messageInfo_GameClose.Size(m)
}
func (m *GameClose) XXX_DiscardUnknown() {
	xxx_messageInfo_GameClose.DiscardUnknown(m)
}

var xxx_messageInfo_GameClose proto.InternalMessageInfo

func (m *GameClose) GetGameId() string {
	if m != nil {
		return m.GameId
	}
	return ""
}

func (m *GameClose) GetSecret() string {
	if m != nil {
		return m.Secret
	}
	return ""
}

type GameCreate struct {
	Value int64 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
	// 加密用的算法
	HashType string `protobuf:"bytes,2,opt,name=hashType,proto3" json:"hashType,omitempty"`
	// 加密后的值
	HashValue            []byte   `protobuf:"bytes,3,opt,name=hashValue,proto3" json:"hashValue,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameCreate) Reset()         { *m = GameCreate{} }
func (m *GameCreate) String() string { return proto.CompactTextString(m) }
func (*GameCreate) ProtoMessage()    {}
func (*GameCreate) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{5}
}
func (m *GameCreate) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameCreate.Unmarshal(m, b)
}
func (m *GameCreate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameCreate.Marshal(b, m, deterministic)
}
func (dst *GameCreate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameCreate.Merge(dst, src)
}
func (m *GameCreate) XXX_Size() int {
	return xxx_messageInfo_GameCreate.Size(m)
}
func (m *GameCreate) XXX_DiscardUnknown() {
	xxx_messageInfo_GameCreate.DiscardUnknown(m)
}

var xxx_messageInfo_GameCreate proto.InternalMessageInfo

func (m *GameCreate) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *GameCreate) GetHashType() string {
	if m != nil {
		return m.HashType
	}
	return ""
}

func (m *GameCreate) GetHashValue() []byte {
	if m != nil {
		return m.HashValue
	}
	return nil
}

// queryByAddr 和 queryByStatus共用同一个结构体
type QueryGameListByStatusAndAddr struct {
	// 优先根据status查询,status不可为空
	Status int32 `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	// 二级搜索，如果要查询一个地址下的所有game信息，可以根据status，分多次查询，这样规避存储数据时的臃余情况
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	// 索引值
	Index int64 `protobuf:"varint,3,opt,name=index,proto3" json:"index,omitempty"`
	// 单页返回多少条记录，默认返回20条，单次最多返回100条
	Count int32 `protobuf:"varint,4,opt,name=count,proto3" json:"count,omitempty"`
	// 0降序，1升序，默认降序
	Direction            int32    `protobuf:"varint,5,opt,name=direction,proto3" json:"direction,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryGameListByStatusAndAddr) Reset()         { *m = QueryGameListByStatusAndAddr{} }
func (m *QueryGameListByStatusAndAddr) String() string { return proto.CompactTextString(m) }
func (*QueryGameListByStatusAndAddr) ProtoMessage()    {}
func (*QueryGameListByStatusAndAddr) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{6}
}
func (m *QueryGameListByStatusAndAddr) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryGameListByStatusAndAddr.Unmarshal(m, b)
}
func (m *QueryGameListByStatusAndAddr) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryGameListByStatusAndAddr.Marshal(b, m, deterministic)
}
func (dst *QueryGameListByStatusAndAddr) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryGameListByStatusAndAddr.Merge(dst, src)
}
func (m *QueryGameListByStatusAndAddr) XXX_Size() int {
	return xxx_messageInfo_QueryGameListByStatusAndAddr.Size(m)
}
func (m *QueryGameListByStatusAndAddr) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryGameListByStatusAndAddr.DiscardUnknown(m)
}

var xxx_messageInfo_QueryGameListByStatusAndAddr proto.InternalMessageInfo

func (m *QueryGameListByStatusAndAddr) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *QueryGameListByStatusAndAddr) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *QueryGameListByStatusAndAddr) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *QueryGameListByStatusAndAddr) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *QueryGameListByStatusAndAddr) GetDirection() int32 {
	if m != nil {
		return m.Direction
	}
	return 0
}

// 统计数量
type QueryGameListCount struct {
	// 优先根据status查询,status不可为空
	Status int32 `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	// 二级搜索，如果要查询一个地址下的所有game信息，可以根据status，分多次查询，这样规避存储数据时的臃余情况
	Address              string   `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryGameListCount) Reset()         { *m = QueryGameListCount{} }
func (m *QueryGameListCount) String() string { return proto.CompactTextString(m) }
func (*QueryGameListCount) ProtoMessage()    {}
func (*QueryGameListCount) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{7}
}
func (m *QueryGameListCount) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryGameListCount.Unmarshal(m, b)
}
func (m *QueryGameListCount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryGameListCount.Marshal(b, m, deterministic)
}
func (dst *QueryGameListCount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryGameListCount.Merge(dst, src)
}
func (m *QueryGameListCount) XXX_Size() int {
	return xxx_messageInfo_QueryGameListCount.Size(m)
}
func (m *QueryGameListCount) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryGameListCount.DiscardUnknown(m)
}

var xxx_messageInfo_QueryGameListCount proto.InternalMessageInfo

func (m *QueryGameListCount) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *QueryGameListCount) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

// 索引value值
type GameRecord struct {
	GameId               string   `protobuf:"bytes,1,opt,name=gameId,proto3" json:"gameId,omitempty"`
	Index                int64    `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GameRecord) Reset()         { *m = GameRecord{} }
func (m *GameRecord) String() string { return proto.CompactTextString(m) }
func (*GameRecord) ProtoMessage()    {}
func (*GameRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{8}
}
func (m *GameRecord) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GameRecord.Unmarshal(m, b)
}
func (m *GameRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GameRecord.Marshal(b, m, deterministic)
}
func (dst *GameRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GameRecord.Merge(dst, src)
}
func (m *GameRecord) XXX_Size() int {
	return xxx_messageInfo_GameRecord.Size(m)
}
func (m *GameRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_GameRecord.DiscardUnknown(m)
}

var xxx_messageInfo_GameRecord proto.InternalMessageInfo

func (m *GameRecord) GetGameId() string {
	if m != nil {
		return m.GameId
	}
	return ""
}

func (m *GameRecord) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

type QueryGameInfo struct {
	GameId               string   `protobuf:"bytes,1,opt,name=gameId,proto3" json:"gameId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryGameInfo) Reset()         { *m = QueryGameInfo{} }
func (m *QueryGameInfo) String() string { return proto.CompactTextString(m) }
func (*QueryGameInfo) ProtoMessage()    {}
func (*QueryGameInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{9}
}
func (m *QueryGameInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryGameInfo.Unmarshal(m, b)
}
func (m *QueryGameInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryGameInfo.Marshal(b, m, deterministic)
}
func (dst *QueryGameInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryGameInfo.Merge(dst, src)
}
func (m *QueryGameInfo) XXX_Size() int {
	return xxx_messageInfo_QueryGameInfo.Size(m)
}
func (m *QueryGameInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryGameInfo.DiscardUnknown(m)
}

var xxx_messageInfo_QueryGameInfo proto.InternalMessageInfo

func (m *QueryGameInfo) GetGameId() string {
	if m != nil {
		return m.GameId
	}
	return ""
}

type QueryGameInfos struct {
	GameIds              []string `protobuf:"bytes,1,rep,name=gameIds,proto3" json:"gameIds,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryGameInfos) Reset()         { *m = QueryGameInfos{} }
func (m *QueryGameInfos) String() string { return proto.CompactTextString(m) }
func (*QueryGameInfos) ProtoMessage()    {}
func (*QueryGameInfos) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{10}
}
func (m *QueryGameInfos) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryGameInfos.Unmarshal(m, b)
}
func (m *QueryGameInfos) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryGameInfos.Marshal(b, m, deterministic)
}
func (dst *QueryGameInfos) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryGameInfos.Merge(dst, src)
}
func (m *QueryGameInfos) XXX_Size() int {
	return xxx_messageInfo_QueryGameInfos.Size(m)
}
func (m *QueryGameInfos) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryGameInfos.DiscardUnknown(m)
}

var xxx_messageInfo_QueryGameInfos proto.InternalMessageInfo

func (m *QueryGameInfos) GetGameIds() []string {
	if m != nil {
		return m.GameIds
	}
	return nil
}

type ReplyGameList struct {
	Games                []*Game  `protobuf:"bytes,1,rep,name=games,proto3" json:"games,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReplyGameList) Reset()         { *m = ReplyGameList{} }
func (m *ReplyGameList) String() string { return proto.CompactTextString(m) }
func (*ReplyGameList) ProtoMessage()    {}
func (*ReplyGameList) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{11}
}
func (m *ReplyGameList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReplyGameList.Unmarshal(m, b)
}
func (m *ReplyGameList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReplyGameList.Marshal(b, m, deterministic)
}
func (dst *ReplyGameList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReplyGameList.Merge(dst, src)
}
func (m *ReplyGameList) XXX_Size() int {
	return xxx_messageInfo_ReplyGameList.Size(m)
}
func (m *ReplyGameList) XXX_DiscardUnknown() {
	xxx_messageInfo_ReplyGameList.DiscardUnknown(m)
}

var xxx_messageInfo_ReplyGameList proto.InternalMessageInfo

func (m *ReplyGameList) GetGames() []*Game {
	if m != nil {
		return m.Games
	}
	return nil
}

type ReplyGameListCount struct {
	Count                int64    `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReplyGameListCount) Reset()         { *m = ReplyGameListCount{} }
func (m *ReplyGameListCount) String() string { return proto.CompactTextString(m) }
func (*ReplyGameListCount) ProtoMessage()    {}
func (*ReplyGameListCount) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{12}
}
func (m *ReplyGameListCount) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReplyGameListCount.Unmarshal(m, b)
}
func (m *ReplyGameListCount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReplyGameListCount.Marshal(b, m, deterministic)
}
func (dst *ReplyGameListCount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReplyGameListCount.Merge(dst, src)
}
func (m *ReplyGameListCount) XXX_Size() int {
	return xxx_messageInfo_ReplyGameListCount.Size(m)
}
func (m *ReplyGameListCount) XXX_DiscardUnknown() {
	xxx_messageInfo_ReplyGameListCount.DiscardUnknown(m)
}

var xxx_messageInfo_ReplyGameListCount proto.InternalMessageInfo

func (m *ReplyGameListCount) GetCount() int64 {
	if m != nil {
		return m.Count
	}
	return 0
}

type ReplyGame struct {
	Game                 *Game    `protobuf:"bytes,1,opt,name=game,proto3" json:"game,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReplyGame) Reset()         { *m = ReplyGame{} }
func (m *ReplyGame) String() string { return proto.CompactTextString(m) }
func (*ReplyGame) ProtoMessage()    {}
func (*ReplyGame) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{13}
}
func (m *ReplyGame) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReplyGame.Unmarshal(m, b)
}
func (m *ReplyGame) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReplyGame.Marshal(b, m, deterministic)
}
func (dst *ReplyGame) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReplyGame.Merge(dst, src)
}
func (m *ReplyGame) XXX_Size() int {
	return xxx_messageInfo_ReplyGame.Size(m)
}
func (m *ReplyGame) XXX_DiscardUnknown() {
	xxx_messageInfo_ReplyGame.DiscardUnknown(m)
}

var xxx_messageInfo_ReplyGame proto.InternalMessageInfo

func (m *ReplyGame) GetGame() *Game {
	if m != nil {
		return m.Game
	}
	return nil
}

type ReceiptGame struct {
	GameId string `protobuf:"bytes,1,opt,name=gameId,proto3" json:"gameId,omitempty"`
	Status int32  `protobuf:"varint,2,opt,name=status,proto3" json:"status,omitempty"`
	// 记录上一次状态
	PrevStatus           int32    `protobuf:"varint,3,opt,name=prevStatus,proto3" json:"prevStatus,omitempty"`
	Addr                 string   `protobuf:"bytes,4,opt,name=addr,proto3" json:"addr,omitempty"`
	CreateAddr           string   `protobuf:"bytes,5,opt,name=createAddr,proto3" json:"createAddr,omitempty"`
	MatchAddr            string   `protobuf:"bytes,6,opt,name=matchAddr,proto3" json:"matchAddr,omitempty"`
	Index                int64    `protobuf:"varint,7,opt,name=index,proto3" json:"index,omitempty"`
	PrevIndex            int64    `protobuf:"varint,8,opt,name=prevIndex,proto3" json:"prevIndex,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReceiptGame) Reset()         { *m = ReceiptGame{} }
func (m *ReceiptGame) String() string { return proto.CompactTextString(m) }
func (*ReceiptGame) ProtoMessage()    {}
func (*ReceiptGame) Descriptor() ([]byte, []int) {
	return fileDescriptor_game_10cc1411e24e5934, []int{14}
}
func (m *ReceiptGame) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReceiptGame.Unmarshal(m, b)
}
func (m *ReceiptGame) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReceiptGame.Marshal(b, m, deterministic)
}
func (dst *ReceiptGame) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReceiptGame.Merge(dst, src)
}
func (m *ReceiptGame) XXX_Size() int {
	return xxx_messageInfo_ReceiptGame.Size(m)
}
func (m *ReceiptGame) XXX_DiscardUnknown() {
	xxx_messageInfo_ReceiptGame.DiscardUnknown(m)
}

var xxx_messageInfo_ReceiptGame proto.InternalMessageInfo

func (m *ReceiptGame) GetGameId() string {
	if m != nil {
		return m.GameId
	}
	return ""
}

func (m *ReceiptGame) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *ReceiptGame) GetPrevStatus() int32 {
	if m != nil {
		return m.PrevStatus
	}
	return 0
}

func (m *ReceiptGame) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *ReceiptGame) GetCreateAddr() string {
	if m != nil {
		return m.CreateAddr
	}
	return ""
}

func (m *ReceiptGame) GetMatchAddr() string {
	if m != nil {
		return m.MatchAddr
	}
	return ""
}

func (m *ReceiptGame) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *ReceiptGame) GetPrevIndex() int64 {
	if m != nil {
		return m.PrevIndex
	}
	return 0
}

func init() {
	proto.RegisterType((*Game)(nil), "types.Game")
	proto.RegisterType((*GameAction)(nil), "types.GameAction")
	proto.RegisterType((*GameMatch)(nil), "types.GameMatch")
	proto.RegisterType((*GameCancel)(nil), "types.GameCancel")
	proto.RegisterType((*GameClose)(nil), "types.GameClose")
	proto.RegisterType((*GameCreate)(nil), "types.GameCreate")
	proto.RegisterType((*QueryGameListByStatusAndAddr)(nil), "types.QueryGameListByStatusAndAddr")
	proto.RegisterType((*QueryGameListCount)(nil), "types.QueryGameListCount")
	proto.RegisterType((*GameRecord)(nil), "types.GameRecord")
	proto.RegisterType((*QueryGameInfo)(nil), "types.QueryGameInfo")
	proto.RegisterType((*QueryGameInfos)(nil), "types.QueryGameInfos")
	proto.RegisterType((*ReplyGameList)(nil), "types.ReplyGameList")
	proto.RegisterType((*ReplyGameListCount)(nil), "types.ReplyGameListCount")
	proto.RegisterType((*ReplyGame)(nil), "types.ReplyGame")
	proto.RegisterType((*ReceiptGame)(nil), "types.ReceiptGame")
}

func init() { proto.RegisterFile("game.proto", fileDescriptor_game_10cc1411e24e5934) }

var fileDescriptor_game_10cc1411e24e5934 = []byte{
	// 712 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x55, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0xad, 0x93, 0x38, 0xa9, 0xc7, 0x4d, 0x69, 0x97, 0x08, 0x59, 0xa8, 0x82, 0xb0, 0xaa, 0x44,
	0x54, 0x50, 0x0f, 0xe1, 0x04, 0x9c, 0xda, 0x4a, 0xb4, 0x95, 0xe0, 0xc0, 0x52, 0x71, 0xe2, 0x62,
	0xec, 0xa5, 0x89, 0x94, 0xc4, 0x91, 0xbd, 0xa9, 0x9a, 0x5f, 0xe1, 0xb7, 0xf8, 0x04, 0xf8, 0x10,
	0x34, 0x33, 0xdb, 0xf5, 0xba, 0x52, 0x2a, 0xc1, 0xcd, 0xf3, 0xde, 0xb3, 0xf7, 0xed, 0xec, 0x9b,
	0x35, 0xc0, 0x75, 0x3a, 0xd7, 0xc7, 0xcb, 0xb2, 0x30, 0x85, 0x08, 0xcd, 0x7a, 0xa9, 0x2b, 0xf9,
	0xa7, 0x03, 0x9d, 0xf3, 0x74, 0xae, 0xc5, 0x13, 0xe8, 0x22, 0x7b, 0x99, 0x27, 0xc1, 0x30, 0x18,
	0x45, 0xca, 0x56, 0x88, 0x57, 0x26, 0x35, 0xab, 0x2a, 0x69, 0x0d, 0x83, 0x51, 0xa8, 0x6c, 0x25,
	0x9e, 0x01, 0x64, 0xa5, 0x4e, 0x8d, 0xbe, 0x9a, 0xce, 0x75, 0xd2, 0x1e, 0x06, 0xa3, 0xb6, 0xf2,
	0x10, 0x71, 0x00, 0xd1, 0x3c, 0x35, 0xd9, 0x84, 0xe8, 0x0e, 0xd1, 0x35, 0x80, 0x6c, 0x36, 0x2b,
	0x2a, 0x6d, 0x90, 0x0d, 0x99, 0x75, 0x80, 0x18, 0x40, 0x78, 0x93, 0xce, 0x56, 0x3a, 0xe9, 0x12,
	0xc3, 0x85, 0x38, 0x84, 0x3e, 0x7f, 0xff, 0x24, 0xcf, 0x4b, 0x5d, 0x55, 0x49, 0x8f, 0x8c, 0x36,
	0x41, 0x21, 0x61, 0x87, 0x96, 0xb9, 0x13, 0x6d, 0x93, 0xa8, 0x81, 0x89, 0xa7, 0xb0, 0x3d, 0x49,
	0xab, 0xc9, 0xd5, 0x7a, 0xa9, 0x93, 0x88, 0x78, 0x57, 0xa3, 0x33, 0x7c, 0xfe, 0x4a, 0xeb, 0xc3,
	0x30, 0x18, 0xed, 0xa8, 0x1a, 0xa0, 0x6e, 0xe8, 0xac, 0xd4, 0x26, 0x89, 0xb9, 0x4b, 0x5c, 0x21,
	0x5e, 0xea, 0x6a, 0x35, 0x33, 0xc9, 0x0e, 0x77, 0x89, 0x2b, 0xe7, 0x46, 0x97, 0xe7, 0x2b, 0x74,
	0xd3, 0x27, 0xb6, 0x81, 0xa1, 0xc6, 0xf6, 0xed, 0xf6, 0x22, 0xad, 0x26, 0xc9, 0x2e, 0x3b, 0xf6,
	0x31, 0x31, 0x84, 0x98, 0x9b, 0xc7, 0x92, 0x47, 0x24, 0xf1, 0x21, 0x54, 0x50, 0x03, 0xad, 0x62,
	0x8f, 0x15, 0x1e, 0x44, 0xeb, 0xa4, 0x8b, 0x4c, 0xcf, 0xac, 0x64, 0xdf, 0xae, 0xe3, 0x61, 0xd8,
	0xf9, 0xe9, 0x22, 0xd7, 0xb7, 0x89, 0xe0, 0xce, 0x53, 0x81, 0x3d, 0x59, 0x96, 0xfa, 0xe6, 0x92,
	0x98, 0xc7, 0x7c, 0x5a, 0x0e, 0x70, 0xfe, 0x0b, 0xbb, 0xc7, 0x01, 0xef, 0xd1, 0xc7, 0xe4, 0xaf,
	0x00, 0x00, 0x63, 0x76, 0x92, 0x99, 0x69, 0xb1, 0x10, 0xaf, 0xa0, 0xcb, 0xdb, 0xa3, 0xb0, 0xc5,
	0xe3, 0xfd, 0x63, 0x4a, 0xe3, 0x31, 0x4a, 0xce, 0x88, 0xb8, 0xd8, 0x52, 0x56, 0x42, 0x62, 0xf2,
	0x48, 0x09, 0xbc, 0x27, 0x26, 0x82, 0xc4, 0xf4, 0x24, 0x46, 0x10, 0xd2, 0x9e, 0x29, 0x91, 0xf1,
	0x78, 0xcf, 0xd7, 0x22, 0x7e, 0xb1, 0xa5, 0x58, 0x80, 0x4a, 0xea, 0x1f, 0x85, 0xb3, 0xa9, 0xfc,
	0x84, 0x38, 0x2a, 0x49, 0x20, 0x76, 0xa1, 0x65, 0xd6, 0x94, 0x85, 0x50, 0xb5, 0xcc, 0xfa, 0xb4,
	0x67, 0xe3, 0x29, 0xdf, 0x42, 0xe4, 0xe4, 0x1b, 0x07, 0x68, 0x00, 0xe1, 0x35, 0xf5, 0x85, 0xe7,
	0x87, 0x0b, 0x79, 0xc8, 0xfd, 0x60, 0xff, 0x9b, 0xde, 0x95, 0xef, 0x79, 0x01, 0x72, 0xfe, 0xe0,
	0x84, 0x72, 0x26, 0x5b, 0x7e, 0x26, 0xe5, 0x37, 0xbb, 0x04, 0x77, 0xd1, 0xcd, 0x54, 0xe0, 0xcf,
	0x94, 0x3f, 0x09, 0xad, 0x87, 0x26, 0xa1, 0x7d, 0x6f, 0x12, 0xe4, 0xcf, 0x00, 0x0e, 0x3e, 0xaf,
	0x74, 0xb9, 0xc6, 0x35, 0x3e, 0x4e, 0x2b, 0x73, 0xba, 0xfe, 0x42, 0x37, 0xc3, 0xc9, 0x22, 0xc7,
	0x39, 0xf3, 0x2e, 0x8e, 0xa0, 0x71, 0x71, 0x24, 0xd0, 0x4b, 0xed, 0x6c, 0xf2, 0x8a, 0x77, 0x65,
	0x1d, 0xbe, 0xb6, 0x1f, 0xbe, 0x01, 0x84, 0x59, 0xb1, 0x5a, 0x18, 0x3a, 0xa7, 0x50, 0x71, 0x81,
	0xe6, 0xf2, 0x69, 0xa9, 0x29, 0x4e, 0x74, 0x81, 0x84, 0xaa, 0x06, 0xe4, 0x07, 0x10, 0x0d, 0x6f,
	0x67, 0xf4, 0xce, 0x3f, 0x3b, 0x92, 0xef, 0xb8, 0x85, 0x4a, 0x67, 0x45, 0x99, 0x3f, 0x74, 0xc2,
	0xec, 0xbb, 0xe5, 0xf9, 0x96, 0x2f, 0xa1, 0xef, 0x3c, 0x5c, 0x2e, 0x7e, 0x14, 0x1b, 0x0f, 0xf9,
	0x08, 0x76, 0x1b, 0x42, 0x32, 0xc4, 0x1c, 0x3a, 0x6d, 0xa3, 0x21, 0x5b, 0xca, 0x31, 0xf4, 0x95,
	0x5e, 0xce, 0xdc, 0xc6, 0xc4, 0x0b, 0x08, 0x91, 0x63, 0x61, 0x3c, 0x8e, 0xbd, 0x14, 0x2b, 0x66,
	0xe4, 0x11, 0x88, 0xc6, 0x3b, 0xdc, 0x0c, 0xd7, 0x56, 0x9b, 0x07, 0x2a, 0xe4, 0x6b, 0x88, 0x9c,
	0x56, 0x3c, 0x87, 0x0e, 0x7e, 0xc1, 0xce, 0x68, 0xe3, 0xd3, 0x44, 0xc8, 0xdf, 0x01, 0xc4, 0x4a,
	0x67, 0x7a, 0xba, 0x34, 0xff, 0xfb, 0x0f, 0xc1, 0x6b, 0x84, 0x73, 0x43, 0xa7, 0x1e, 0x2a, 0x0f,
	0x11, 0x02, 0x3a, 0x78, 0x12, 0x74, 0xf2, 0x91, 0xa2, 0xe7, 0xfa, 0xbf, 0x83, 0x21, 0xa3, 0x93,
	0x8f, 0x94, 0x87, 0xb8, 0xff, 0x0e, 0xd1, 0x5d, 0xa2, 0x6b, 0xa0, 0x3e, 0xaa, 0xde, 0xc6, 0xfb,
	0x6d, 0xfb, 0xde, 0xfd, 0xf6, 0xbd, 0x4b, 0x3f, 0xcc, 0x37, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff,
	0xa4, 0x8c, 0x93, 0xd4, 0x3e, 0x07, 0x00, 0x00,
}
