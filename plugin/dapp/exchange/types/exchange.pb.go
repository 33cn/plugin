// Code generated by protoc-gen-go. DO NOT EDIT.
// source: exchange.proto

/*
Package types is a generated protocol buffer package.

It is generated from these files:
	exchange.proto

It has these top-level messages:
	Exchange
	ExchangeAction
	LimitOrder
	MarketOrder
	RevokeOrder
	Asset
	Order
	OrderPrice
	OrderID
	QueryMarketDepth
	MarketDepth
	MarketDepthList
	QueryCompletedOrderList
	QueryOrder
	QueryOrderList
	OrderList
	ReceiptExchange
*/
package types

import (
	fmt "fmt"

	proto "github.com/golang/protobuf/proto"

	math "math"

	context "golang.org/x/net/context"

	grpc "google.golang.org/grpc"
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

type Exchange struct {
}

func (m *Exchange) Reset()                    { *m = Exchange{} }
func (m *Exchange) String() string            { return proto.CompactTextString(m) }
func (*Exchange) ProtoMessage()               {}
func (*Exchange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type ExchangeAction struct {
	// Types that are valid to be assigned to Value:
	//	*ExchangeAction_LimitOrder
	//	*ExchangeAction_MarketOrder
	//	*ExchangeAction_RevokeOrder
	Value isExchangeAction_Value `protobuf_oneof:"value"`
	Ty    int32                  `protobuf:"varint,6,opt,name=ty" json:"ty,omitempty"`
}

func (m *ExchangeAction) Reset()                    { *m = ExchangeAction{} }
func (m *ExchangeAction) String() string            { return proto.CompactTextString(m) }
func (*ExchangeAction) ProtoMessage()               {}
func (*ExchangeAction) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type isExchangeAction_Value interface {
	isExchangeAction_Value()
}

type ExchangeAction_LimitOrder struct {
	LimitOrder *LimitOrder `protobuf:"bytes,1,opt,name=limitOrder,oneof"`
}
type ExchangeAction_MarketOrder struct {
	MarketOrder *MarketOrder `protobuf:"bytes,2,opt,name=marketOrder,oneof"`
}
type ExchangeAction_RevokeOrder struct {
	RevokeOrder *RevokeOrder `protobuf:"bytes,3,opt,name=revokeOrder,oneof"`
}

func (*ExchangeAction_LimitOrder) isExchangeAction_Value()  {}
func (*ExchangeAction_MarketOrder) isExchangeAction_Value() {}
func (*ExchangeAction_RevokeOrder) isExchangeAction_Value() {}

func (m *ExchangeAction) GetValue() isExchangeAction_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *ExchangeAction) GetLimitOrder() *LimitOrder {
	if x, ok := m.GetValue().(*ExchangeAction_LimitOrder); ok {
		return x.LimitOrder
	}
	return nil
}

func (m *ExchangeAction) GetMarketOrder() *MarketOrder {
	if x, ok := m.GetValue().(*ExchangeAction_MarketOrder); ok {
		return x.MarketOrder
	}
	return nil
}

func (m *ExchangeAction) GetRevokeOrder() *RevokeOrder {
	if x, ok := m.GetValue().(*ExchangeAction_RevokeOrder); ok {
		return x.RevokeOrder
	}
	return nil
}

func (m *ExchangeAction) GetTy() int32 {
	if m != nil {
		return m.Ty
	}
	return 0
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*ExchangeAction) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _ExchangeAction_OneofMarshaler, _ExchangeAction_OneofUnmarshaler, _ExchangeAction_OneofSizer, []interface{}{
		(*ExchangeAction_LimitOrder)(nil),
		(*ExchangeAction_MarketOrder)(nil),
		(*ExchangeAction_RevokeOrder)(nil),
	}
}

func _ExchangeAction_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*ExchangeAction)
	// value
	switch x := m.Value.(type) {
	case *ExchangeAction_LimitOrder:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.LimitOrder); err != nil {
			return err
		}
	case *ExchangeAction_MarketOrder:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.MarketOrder); err != nil {
			return err
		}
	case *ExchangeAction_RevokeOrder:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.RevokeOrder); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("ExchangeAction.Value has unexpected type %T", x)
	}
	return nil
}

func _ExchangeAction_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*ExchangeAction)
	switch tag {
	case 1: // value.limitOrder
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(LimitOrder)
		err := b.DecodeMessage(msg)
		m.Value = &ExchangeAction_LimitOrder{msg}
		return true, err
	case 2: // value.marketOrder
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(MarketOrder)
		err := b.DecodeMessage(msg)
		m.Value = &ExchangeAction_MarketOrder{msg}
		return true, err
	case 3: // value.revokeOrder
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(RevokeOrder)
		err := b.DecodeMessage(msg)
		m.Value = &ExchangeAction_RevokeOrder{msg}
		return true, err
	default:
		return false, nil
	}
}

func _ExchangeAction_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*ExchangeAction)
	// value
	switch x := m.Value.(type) {
	case *ExchangeAction_LimitOrder:
		s := proto.Size(x.LimitOrder)
		n += proto.SizeVarint(1<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *ExchangeAction_MarketOrder:
		s := proto.Size(x.MarketOrder)
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *ExchangeAction_RevokeOrder:
		s := proto.Size(x.RevokeOrder)
		n += proto.SizeVarint(3<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// 限价订单
type LimitOrder struct {
	// 交易对
	LeftAsset *Asset `protobuf:"bytes,1,opt,name=leftAsset" json:"leftAsset,omitempty"`
	// 交易对
	RightAsset *Asset `protobuf:"bytes,2,opt,name=rightAsset" json:"rightAsset,omitempty"`
	// 价格
	Price float64 `protobuf:"fixed64,3,opt,name=price" json:"price,omitempty"`
	// 总量
	Amount int64 `protobuf:"varint,4,opt,name=amount" json:"amount,omitempty"`
	// 操作， 1为买，2为卖
	Op int32 `protobuf:"varint,5,opt,name=op" json:"op,omitempty"`
}

func (m *LimitOrder) Reset()                    { *m = LimitOrder{} }
func (m *LimitOrder) String() string            { return proto.CompactTextString(m) }
func (*LimitOrder) ProtoMessage()               {}
func (*LimitOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *LimitOrder) GetLeftAsset() *Asset {
	if m != nil {
		return m.LeftAsset
	}
	return nil
}

func (m *LimitOrder) GetRightAsset() *Asset {
	if m != nil {
		return m.RightAsset
	}
	return nil
}

func (m *LimitOrder) GetPrice() float64 {
	if m != nil {
		return m.Price
	}
	return 0
}

func (m *LimitOrder) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *LimitOrder) GetOp() int32 {
	if m != nil {
		return m.Op
	}
	return 0
}

// 市价委托
type MarketOrder struct {
	// 资产1
	LeftAsset *Asset `protobuf:"bytes,1,opt,name=leftAsset" json:"leftAsset,omitempty"`
	// 资产2
	RightAsset *Asset `protobuf:"bytes,2,opt,name=rightAsset" json:"rightAsset,omitempty"`
	// 总量
	Amount int64 `protobuf:"varint,3,opt,name=amount" json:"amount,omitempty"`
	// 操作， 1为买，2为卖
	Op int32 `protobuf:"varint,4,opt,name=op" json:"op,omitempty"`
}

func (m *MarketOrder) Reset()                    { *m = MarketOrder{} }
func (m *MarketOrder) String() string            { return proto.CompactTextString(m) }
func (*MarketOrder) ProtoMessage()               {}
func (*MarketOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *MarketOrder) GetLeftAsset() *Asset {
	if m != nil {
		return m.LeftAsset
	}
	return nil
}

func (m *MarketOrder) GetRightAsset() *Asset {
	if m != nil {
		return m.RightAsset
	}
	return nil
}

func (m *MarketOrder) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *MarketOrder) GetOp() int32 {
	if m != nil {
		return m.Op
	}
	return 0
}

// 撤回订单
type RevokeOrder struct {
	// 订单号
	OrderID string `protobuf:"bytes,1,opt,name=orderID" json:"orderID,omitempty"`
}

func (m *RevokeOrder) Reset()                    { *m = RevokeOrder{} }
func (m *RevokeOrder) String() string            { return proto.CompactTextString(m) }
func (*RevokeOrder) ProtoMessage()               {}
func (*RevokeOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *RevokeOrder) GetOrderID() string {
	if m != nil {
		return m.OrderID
	}
	return ""
}

// 资产类型
type Asset struct {
	Execer string `protobuf:"bytes,1,opt,name=execer" json:"execer,omitempty"`
	Symbol string `protobuf:"bytes,2,opt,name=symbol" json:"symbol,omitempty"`
}

func (m *Asset) Reset()                    { *m = Asset{} }
func (m *Asset) String() string            { return proto.CompactTextString(m) }
func (*Asset) ProtoMessage()               {}
func (*Asset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Asset) GetExecer() string {
	if m != nil {
		return m.Execer
	}
	return ""
}

func (m *Asset) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

// 订单信息
type Order struct {
	OrderID string `protobuf:"bytes,1,opt,name=orderID" json:"orderID,omitempty"`
	// Types that are valid to be assigned to Value:
	//	*Order_LimitOrder
	//	*Order_MarketOrder
	Value isOrder_Value `protobuf_oneof:"value"`
	// 挂单类型
	Ty int32 `protobuf:"varint,4,opt,name=ty" json:"ty,omitempty"`
	// 已经成交的数量
	Executed int64 `protobuf:"varint,5,opt,name=executed" json:"executed,omitempty"`
	// 余额
	Balance int64 `protobuf:"varint,6,opt,name=balance" json:"balance,omitempty"`
	// 状态,0 挂单中ordered， 1 完成completed， 2撤回 revoked
	Status int32 `protobuf:"varint,7,opt,name=status" json:"status,omitempty"`
	// 用户地址
	Addr string `protobuf:"bytes,8,opt,name=addr" json:"addr,omitempty"`
	// 更新时间
	UpdateTime int64 `protobuf:"varint,9,opt,name=updateTime" json:"updateTime,omitempty"`
	// 索引
	Index int64 `protobuf:"varint,10,opt,name=index" json:"index,omitempty"`
}

func (m *Order) Reset()                    { *m = Order{} }
func (m *Order) String() string            { return proto.CompactTextString(m) }
func (*Order) ProtoMessage()               {}
func (*Order) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type isOrder_Value interface {
	isOrder_Value()
}

type Order_LimitOrder struct {
	LimitOrder *LimitOrder `protobuf:"bytes,2,opt,name=limitOrder,oneof"`
}
type Order_MarketOrder struct {
	MarketOrder *MarketOrder `protobuf:"bytes,3,opt,name=marketOrder,oneof"`
}

func (*Order_LimitOrder) isOrder_Value()  {}
func (*Order_MarketOrder) isOrder_Value() {}

func (m *Order) GetValue() isOrder_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Order) GetOrderID() string {
	if m != nil {
		return m.OrderID
	}
	return ""
}

func (m *Order) GetLimitOrder() *LimitOrder {
	if x, ok := m.GetValue().(*Order_LimitOrder); ok {
		return x.LimitOrder
	}
	return nil
}

func (m *Order) GetMarketOrder() *MarketOrder {
	if x, ok := m.GetValue().(*Order_MarketOrder); ok {
		return x.MarketOrder
	}
	return nil
}

func (m *Order) GetTy() int32 {
	if m != nil {
		return m.Ty
	}
	return 0
}

func (m *Order) GetExecuted() int64 {
	if m != nil {
		return m.Executed
	}
	return 0
}

func (m *Order) GetBalance() int64 {
	if m != nil {
		return m.Balance
	}
	return 0
}

func (m *Order) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *Order) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *Order) GetUpdateTime() int64 {
	if m != nil {
		return m.UpdateTime
	}
	return 0
}

func (m *Order) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Order) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Order_OneofMarshaler, _Order_OneofUnmarshaler, _Order_OneofSizer, []interface{}{
		(*Order_LimitOrder)(nil),
		(*Order_MarketOrder)(nil),
	}
}

func _Order_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Order)
	// value
	switch x := m.Value.(type) {
	case *Order_LimitOrder:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.LimitOrder); err != nil {
			return err
		}
	case *Order_MarketOrder:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.MarketOrder); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Order.Value has unexpected type %T", x)
	}
	return nil
}

func _Order_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Order)
	switch tag {
	case 2: // value.limitOrder
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(LimitOrder)
		err := b.DecodeMessage(msg)
		m.Value = &Order_LimitOrder{msg}
		return true, err
	case 3: // value.marketOrder
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(MarketOrder)
		err := b.DecodeMessage(msg)
		m.Value = &Order_MarketOrder{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Order_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Order)
	// value
	switch x := m.Value.(type) {
	case *Order_LimitOrder:
		s := proto.Size(x.LimitOrder)
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Order_MarketOrder:
		s := proto.Size(x.MarketOrder)
		n += proto.SizeVarint(3<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// 挂单价
type OrderPrice struct {
	Price float64 `protobuf:"fixed64,1,opt,name=price" json:"price,omitempty"`
	Index int64   `protobuf:"varint,2,opt,name=index" json:"index,omitempty"`
}

func (m *OrderPrice) Reset()                    { *m = OrderPrice{} }
func (m *OrderPrice) String() string            { return proto.CompactTextString(m) }
func (*OrderPrice) ProtoMessage()               {}
func (*OrderPrice) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *OrderPrice) GetPrice() float64 {
	if m != nil {
		return m.Price
	}
	return 0
}

func (m *OrderPrice) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

// 单号
type OrderID struct {
	ID    string `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	Index int64  `protobuf:"varint,2,opt,name=index" json:"index,omitempty"`
}

func (m *OrderID) Reset()                    { *m = OrderID{} }
func (m *OrderID) String() string            { return proto.CompactTextString(m) }
func (*OrderID) ProtoMessage()               {}
func (*OrderID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *OrderID) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *OrderID) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

// 查询接口
type QueryMarketDepth struct {
	// 资产1
	LeftAsset *Asset `protobuf:"bytes,1,opt,name=leftAsset" json:"leftAsset,omitempty"`
	// 资产2
	RightAsset *Asset `protobuf:"bytes,2,opt,name=rightAsset" json:"rightAsset,omitempty"`
	// 操作， 1为买，2为卖
	Op int32 `protobuf:"varint,3,opt,name=op" json:"op,omitempty"`
	// 这里用价格作为索引值
	Price float64 `protobuf:"fixed64,4,opt,name=price" json:"price,omitempty"`
	// 单页返回多少条记录，默认返回10条,为了系统安全最多单次只能返回20条
	Count int32 `protobuf:"varint,5,opt,name=count" json:"count,omitempty"`
}

func (m *QueryMarketDepth) Reset()                    { *m = QueryMarketDepth{} }
func (m *QueryMarketDepth) String() string            { return proto.CompactTextString(m) }
func (*QueryMarketDepth) ProtoMessage()               {}
func (*QueryMarketDepth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *QueryMarketDepth) GetLeftAsset() *Asset {
	if m != nil {
		return m.LeftAsset
	}
	return nil
}

func (m *QueryMarketDepth) GetRightAsset() *Asset {
	if m != nil {
		return m.RightAsset
	}
	return nil
}

func (m *QueryMarketDepth) GetOp() int32 {
	if m != nil {
		return m.Op
	}
	return 0
}

func (m *QueryMarketDepth) GetPrice() float64 {
	if m != nil {
		return m.Price
	}
	return 0
}

func (m *QueryMarketDepth) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

// 市场深度
type MarketDepth struct {
	// 资产1
	LeftAsset *Asset `protobuf:"bytes,1,opt,name=leftAsset" json:"leftAsset,omitempty"`
	// 资产2
	RightAsset *Asset `protobuf:"bytes,2,opt,name=rightAsset" json:"rightAsset,omitempty"`
	// 价格
	Price float64 `protobuf:"fixed64,3,opt,name=price" json:"price,omitempty"`
	// 总量
	Amount int64 `protobuf:"varint,4,opt,name=amount" json:"amount,omitempty"`
	// 操作， 1为买，2为卖
	Op int32 `protobuf:"varint,5,opt,name=op" json:"op,omitempty"`
}

func (m *MarketDepth) Reset()                    { *m = MarketDepth{} }
func (m *MarketDepth) String() string            { return proto.CompactTextString(m) }
func (*MarketDepth) ProtoMessage()               {}
func (*MarketDepth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *MarketDepth) GetLeftAsset() *Asset {
	if m != nil {
		return m.LeftAsset
	}
	return nil
}

func (m *MarketDepth) GetRightAsset() *Asset {
	if m != nil {
		return m.RightAsset
	}
	return nil
}

func (m *MarketDepth) GetPrice() float64 {
	if m != nil {
		return m.Price
	}
	return 0
}

func (m *MarketDepth) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *MarketDepth) GetOp() int32 {
	if m != nil {
		return m.Op
	}
	return 0
}

// 查询接口返回的市场深度列表
type MarketDepthList struct {
	List []*MarketDepth `protobuf:"bytes,1,rep,name=list" json:"list,omitempty"`
}

func (m *MarketDepthList) Reset()                    { *m = MarketDepthList{} }
func (m *MarketDepthList) String() string            { return proto.CompactTextString(m) }
func (*MarketDepthList) ProtoMessage()               {}
func (*MarketDepthList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *MarketDepthList) GetList() []*MarketDepth {
	if m != nil {
		return m.List
	}
	return nil
}

// 查询最新得成交信息,外部接口
type QueryCompletedOrderList struct {
	// 资产1
	LeftAsset *Asset `protobuf:"bytes,1,opt,name=leftAsset" json:"leftAsset,omitempty"`
	// 资产2
	RightAsset *Asset `protobuf:"bytes,2,opt,name=rightAsset" json:"rightAsset,omitempty"`
	// 索引值
	Index int64 `protobuf:"varint,3,opt,name=index" json:"index,omitempty"`
	// 单页返回多少条记录，默认返回10条,为了系统安全最多单次只能返回20条
	Count int32 `protobuf:"varint,4,opt,name=count" json:"count,omitempty"`
	// 0降序，1升序，默认降序
	Direction int32 `protobuf:"varint,5,opt,name=direction" json:"direction,omitempty"`
}

func (m *QueryCompletedOrderList) Reset()                    { *m = QueryCompletedOrderList{} }
func (m *QueryCompletedOrderList) String() string            { return proto.CompactTextString(m) }
func (*QueryCompletedOrderList) ProtoMessage()               {}
func (*QueryCompletedOrderList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *QueryCompletedOrderList) GetLeftAsset() *Asset {
	if m != nil {
		return m.LeftAsset
	}
	return nil
}

func (m *QueryCompletedOrderList) GetRightAsset() *Asset {
	if m != nil {
		return m.RightAsset
	}
	return nil
}

func (m *QueryCompletedOrderList) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *QueryCompletedOrderList) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *QueryCompletedOrderList) GetDirection() int32 {
	if m != nil {
		return m.Direction
	}
	return 0
}

// 根据orderID去查询订单信息
type QueryOrder struct {
	OrderID string `protobuf:"bytes,1,opt,name=orderID" json:"orderID,omitempty"`
}

func (m *QueryOrder) Reset()                    { *m = QueryOrder{} }
func (m *QueryOrder) String() string            { return proto.CompactTextString(m) }
func (*QueryOrder) ProtoMessage()               {}
func (*QueryOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *QueryOrder) GetOrderID() string {
	if m != nil {
		return m.OrderID
	}
	return ""
}

// 根据地址，状态查询用户自己的挂单信息
type QueryOrderList struct {
	// 挂单状态必填(默认是0,只查询ordered挂单中的)
	Status int32 `protobuf:"varint,1,opt,name=status" json:"status,omitempty"`
	// 用户地址信息，必填
	Address string `protobuf:"bytes,2,opt,name=address" json:"address,omitempty"`
	// 索引值
	Index int64 `protobuf:"varint,3,opt,name=index" json:"index,omitempty"`
	// 单页返回多少条记录，默认返回10条,为了系统安全最多单次只能返回20条
	Count int32 `protobuf:"varint,4,opt,name=count" json:"count,omitempty"`
	// 0降序，1升序，默认降序
	Direction int32 `protobuf:"varint,5,opt,name=direction" json:"direction,omitempty"`
}

func (m *QueryOrderList) Reset()                    { *m = QueryOrderList{} }
func (m *QueryOrderList) String() string            { return proto.CompactTextString(m) }
func (*QueryOrderList) ProtoMessage()               {}
func (*QueryOrderList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *QueryOrderList) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *QueryOrderList) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *QueryOrderList) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *QueryOrderList) GetCount() int32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *QueryOrderList) GetDirection() int32 {
	if m != nil {
		return m.Direction
	}
	return 0
}

// 订单列表
type OrderList struct {
	List []*Order `protobuf:"bytes,1,rep,name=list" json:"list,omitempty"`
}

func (m *OrderList) Reset()                    { *m = OrderList{} }
func (m *OrderList) String() string            { return proto.CompactTextString(m) }
func (*OrderList) ProtoMessage()               {}
func (*OrderList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

func (m *OrderList) GetList() []*Order {
	if m != nil {
		return m.List
	}
	return nil
}

// exchange执行票据日志
type ReceiptExchange struct {
	Order       *Order   `protobuf:"bytes,1,opt,name=order" json:"order,omitempty"`
	MatchOrders []*Order `protobuf:"bytes,2,rep,name=matchOrders" json:"matchOrders,omitempty"`
	Index       int64    `protobuf:"varint,3,opt,name=index" json:"index,omitempty"`
}

func (m *ReceiptExchange) Reset()                    { *m = ReceiptExchange{} }
func (m *ReceiptExchange) String() string            { return proto.CompactTextString(m) }
func (*ReceiptExchange) ProtoMessage()               {}
func (*ReceiptExchange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *ReceiptExchange) GetOrder() *Order {
	if m != nil {
		return m.Order
	}
	return nil
}

func (m *ReceiptExchange) GetMatchOrders() []*Order {
	if m != nil {
		return m.MatchOrders
	}
	return nil
}

func (m *ReceiptExchange) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func init() {
	proto.RegisterType((*Exchange)(nil), "types.Exchange")
	proto.RegisterType((*ExchangeAction)(nil), "types.ExchangeAction")
	proto.RegisterType((*LimitOrder)(nil), "types.LimitOrder")
	proto.RegisterType((*MarketOrder)(nil), "types.MarketOrder")
	proto.RegisterType((*RevokeOrder)(nil), "types.RevokeOrder")
	proto.RegisterType((*Asset)(nil), "types.asset")
	proto.RegisterType((*Order)(nil), "types.Order")
	proto.RegisterType((*OrderPrice)(nil), "types.OrderPrice")
	proto.RegisterType((*OrderID)(nil), "types.OrderID")
	proto.RegisterType((*QueryMarketDepth)(nil), "types.QueryMarketDepth")
	proto.RegisterType((*MarketDepth)(nil), "types.MarketDepth")
	proto.RegisterType((*MarketDepthList)(nil), "types.MarketDepthList")
	proto.RegisterType((*QueryCompletedOrderList)(nil), "types.QueryCompletedOrderList")
	proto.RegisterType((*QueryOrder)(nil), "types.QueryOrder")
	proto.RegisterType((*QueryOrderList)(nil), "types.QueryOrderList")
	proto.RegisterType((*OrderList)(nil), "types.OrderList")
	proto.RegisterType((*ReceiptExchange)(nil), "types.ReceiptExchange")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Exchange service

type ExchangeClient interface {
}

type exchangeClient struct {
	cc *grpc.ClientConn
}

func NewExchangeClient(cc *grpc.ClientConn) ExchangeClient {
	return &exchangeClient{cc}
}

// Server API for Exchange service

type ExchangeServer interface {
}

func RegisterExchangeServer(s *grpc.Server, srv ExchangeServer) {
	s.RegisterService(&_Exchange_serviceDesc, srv)
}

var _Exchange_serviceDesc = grpc.ServiceDesc{
	ServiceName: "types.exchange",
	HandlerType: (*ExchangeServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "exchange.proto",
}

func init() { proto.RegisterFile("exchange.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 659 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x55, 0xed, 0x6a, 0xd4, 0x4c,
	0x14, 0xee, 0xe4, 0xa3, 0xdb, 0x9c, 0x7d, 0xd9, 0xbe, 0x0e, 0xa2, 0x41, 0x44, 0x96, 0xfc, 0xa8,
	0x45, 0x74, 0x85, 0x16, 0xfc, 0xf8, 0x59, 0xad, 0xd0, 0x42, 0xa5, 0x3a, 0x78, 0x03, 0x69, 0x72,
	0xec, 0x86, 0x66, 0x37, 0x21, 0x99, 0x2d, 0x5d, 0xbc, 0x05, 0xc1, 0x9b, 0x50, 0xf0, 0x26, 0xc4,
	0x3b, 0xf0, 0x9a, 0x64, 0x4e, 0x26, 0x99, 0xd9, 0x76, 0xa5, 0xa2, 0xee, 0xbf, 0x3c, 0x73, 0x3e,
	0xf2, 0x9c, 0x73, 0x9e, 0x33, 0x03, 0x03, 0xbc, 0x48, 0xc6, 0xf1, 0xf4, 0x14, 0x47, 0x65, 0x55,
	0xc8, 0x82, 0xfb, 0x72, 0x5e, 0x62, 0x1d, 0x01, 0x6c, 0xbc, 0xd2, 0x86, 0xe8, 0x07, 0x83, 0x41,
	0x0b, 0xf6, 0x12, 0x99, 0x15, 0x53, 0xbe, 0x0b, 0x90, 0x67, 0x93, 0x4c, 0x1e, 0x57, 0x29, 0x56,
	0x21, 0x1b, 0xb2, 0xed, 0xfe, 0xce, 0x8d, 0x11, 0x85, 0x8e, 0x8e, 0x3a, 0xc3, 0xc1, 0x9a, 0xb0,
	0xdc, 0xf8, 0x13, 0xe8, 0x4f, 0xe2, 0xea, 0x0c, 0x75, 0x94, 0x43, 0x51, 0x5c, 0x47, 0xbd, 0x36,
	0x96, 0x83, 0x35, 0x61, 0x3b, 0xaa, 0xb8, 0x0a, 0xcf, 0x8b, 0x33, 0x6c, 0xe2, 0xdc, 0x85, 0x38,
	0x61, 0x2c, 0x2a, 0xce, 0x72, 0xe4, 0x03, 0x70, 0xe4, 0x3c, 0x5c, 0x1f, 0xb2, 0x6d, 0x5f, 0x38,
	0x72, 0xfe, 0xa2, 0x07, 0xfe, 0x79, 0x9c, 0xcf, 0x30, 0xfa, 0xcc, 0x00, 0x0c, 0x4b, 0xfe, 0x00,
	0x82, 0x1c, 0xdf, 0xcb, 0xbd, 0xba, 0x46, 0xa9, 0x6b, 0xf9, 0x4f, 0x67, 0x8f, 0xd5, 0x99, 0x30,
	0x66, 0xfe, 0x10, 0xa0, 0xca, 0x4e, 0xc7, 0xda, 0xd9, 0x59, 0xe2, 0x6c, 0xd9, 0xf9, 0x4d, 0xf0,
	0xcb, 0x2a, 0x4b, 0x90, 0x38, 0x33, 0xd1, 0x00, 0x7e, 0x0b, 0xd6, 0xe3, 0x49, 0x31, 0x9b, 0xca,
	0xd0, 0x1b, 0xb2, 0x6d, 0x57, 0x68, 0xa4, 0xf8, 0x16, 0x65, 0xe8, 0x37, 0x7c, 0x8b, 0x32, 0xfa,
	0xc4, 0xa0, 0x6f, 0xb5, 0x65, 0x85, 0x3c, 0x0d, 0x23, 0x77, 0x09, 0x23, 0xaf, 0x63, 0x74, 0x1f,
	0xfa, 0x56, 0xbf, 0x79, 0x08, 0xbd, 0x42, 0x7d, 0x1c, 0xee, 0x13, 0x9d, 0x40, 0xb4, 0x30, 0x7a,
	0x0a, 0x7e, 0xdc, 0x66, 0xc6, 0x0b, 0x4c, 0xb4, 0x48, 0x02, 0xa1, 0x91, 0x3a, 0xaf, 0xe7, 0x93,
	0x93, 0x22, 0x27, 0x6e, 0x81, 0xd0, 0x28, 0xfa, 0xee, 0x80, 0x7f, 0x4d, 0xf2, 0x4b, 0xe2, 0x73,
	0xfe, 0x48, 0x7c, 0xee, 0xef, 0x8a, 0xaf, 0x11, 0x91, 0xd7, 0x8a, 0x88, 0xdf, 0x81, 0x0d, 0x55,
	0xc2, 0x4c, 0x62, 0x4a, 0xa3, 0x72, 0x45, 0x87, 0x15, 0xe5, 0x93, 0x38, 0x8f, 0xa7, 0x09, 0x92,
	0xea, 0x5c, 0xd1, 0x42, 0x2a, 0x57, 0xc6, 0x72, 0x56, 0x87, 0x3d, 0xca, 0xa4, 0x11, 0xe7, 0xe0,
	0xc5, 0x69, 0x5a, 0x85, 0x1b, 0x54, 0x21, 0x7d, 0xf3, 0x7b, 0x00, 0xb3, 0x32, 0x8d, 0x25, 0xbe,
	0xcb, 0x26, 0x18, 0x06, 0x94, 0xc8, 0x3a, 0x51, 0xa2, 0xca, 0xa6, 0x29, 0x5e, 0x84, 0x40, 0xa6,
	0x06, 0x18, 0x71, 0x3f, 0x03, 0x20, 0xe6, 0x6f, 0x48, 0x6b, 0x9d, 0x02, 0x99, 0xad, 0xc0, 0x2e,
	0x85, 0x63, 0xa5, 0x88, 0x1e, 0x43, 0xef, 0x58, 0xb7, 0x78, 0x00, 0x4e, 0xd7, 0x77, 0xe7, 0x70,
	0xff, 0x17, 0x01, 0x5f, 0x19, 0xfc, 0xff, 0x76, 0x86, 0xd5, 0xbc, 0xe9, 0xdf, 0x3e, 0x96, 0x72,
	0xbc, 0x42, 0x95, 0x36, 0x6a, 0x74, 0x5b, 0x35, 0x9a, 0xda, 0xbc, 0x4b, 0xb5, 0x25, 0x24, 0xe5,
	0x66, 0x91, 0x1a, 0x10, 0x7d, 0xe9, 0x76, 0x69, 0xd5, 0x2c, 0xff, 0x6e, 0xe7, 0x9f, 0xc3, 0xa6,
	0x45, 0xf3, 0x28, 0xab, 0x25, 0xdf, 0x02, 0x2f, 0xcf, 0x6a, 0xc5, 0xd2, 0xbd, 0x22, 0x59, 0xf2,
	0x12, 0x64, 0x8f, 0xbe, 0x31, 0xb8, 0x4d, 0xd3, 0x78, 0x59, 0x4c, 0xca, 0x1c, 0x25, 0xa6, 0x34,
	0x4d, 0xca, 0xb1, 0xd2, 0x72, 0x1b, 0x65, 0xb8, 0x96, 0x32, 0xcc, 0x10, 0x3c, 0x6b, 0x08, 0xfc,
	0x2e, 0x04, 0x69, 0x56, 0x21, 0x3d, 0x21, 0xba, 0x66, 0x73, 0x10, 0x6d, 0x01, 0x10, 0xfd, 0xeb,
	0xee, 0x96, 0x8f, 0x0c, 0x06, 0xc6, 0x91, 0xca, 0x33, 0xeb, 0xc5, 0x16, 0xd6, 0x2b, 0x84, 0x9e,
	0x5a, 0x29, 0xac, 0x6b, 0x7d, 0xcd, 0xb4, 0xf0, 0x1f, 0xd2, 0x7e, 0x04, 0x81, 0x21, 0x32, 0x5c,
	0x98, 0x55, 0xdb, 0x35, 0xb2, 0xeb, 0x29, 0x7d, 0x80, 0x4d, 0x81, 0x09, 0x66, 0xa5, 0x6c, 0x9f,
	0x54, 0x1e, 0x81, 0x5f, 0x58, 0xef, 0xe8, 0x62, 0x54, 0x63, 0xe2, 0x23, 0x75, 0x7d, 0xc9, 0x64,
	0x4c, 0x87, 0xaa, 0x9a, 0xab, 0xf9, 0x6d, 0x87, 0xe5, 0xf5, 0xed, 0x80, 0xba, 0xbc, 0x9a, 0xbf,
	0x9e, 0xac, 0xd3, 0x7b, 0xbf, 0xfb, 0x33, 0x00, 0x00, 0xff, 0xff, 0xd8, 0xc6, 0x87, 0x60, 0x01,
	0x08, 0x00, 0x00,
}
