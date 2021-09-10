// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.9.1
// source: js.proto

package jsproto

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

// create action
type Create struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code string `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Create) Reset() {
	*x = Create{}
	if protoimpl.UnsafeEnabled {
		mi := &file_js_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Create) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Create) ProtoMessage() {}

func (x *Create) ProtoReflect() protoreflect.Message {
	mi := &file_js_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Create.ProtoReflect.Descriptor instead.
func (*Create) Descriptor() ([]byte, []int) {
	return file_js_proto_rawDescGZIP(), []int{0}
}

func (x *Create) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *Create) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// call action
type Call struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`         // exec name
	Funcname string `protobuf:"bytes,2,opt,name=funcname,proto3" json:"funcname,omitempty"` // call function name
	Args     string `protobuf:"bytes,3,opt,name=args,proto3" json:"args,omitempty"`         // json args
}

func (x *Call) Reset() {
	*x = Call{}
	if protoimpl.UnsafeEnabled {
		mi := &file_js_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Call) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Call) ProtoMessage() {}

func (x *Call) ProtoReflect() protoreflect.Message {
	mi := &file_js_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Call.ProtoReflect.Descriptor instead.
func (*Call) Descriptor() ([]byte, []int) {
	return file_js_proto_rawDescGZIP(), []int{1}
}

func (x *Call) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Call) GetFuncname() string {
	if x != nil {
		return x.Funcname
	}
	return ""
}

func (x *Call) GetArgs() string {
	if x != nil {
		return x.Args
	}
	return ""
}

type JsAction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Value:
	//	*JsAction_Create
	//	*JsAction_Call
	Value isJsAction_Value `protobuf_oneof:"value"`
	Ty    int32            `protobuf:"varint,3,opt,name=ty,proto3" json:"ty,omitempty"`
}

func (x *JsAction) Reset() {
	*x = JsAction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_js_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JsAction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JsAction) ProtoMessage() {}

func (x *JsAction) ProtoReflect() protoreflect.Message {
	mi := &file_js_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JsAction.ProtoReflect.Descriptor instead.
func (*JsAction) Descriptor() ([]byte, []int) {
	return file_js_proto_rawDescGZIP(), []int{2}
}

func (m *JsAction) GetValue() isJsAction_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *JsAction) GetCreate() *Create {
	if x, ok := x.GetValue().(*JsAction_Create); ok {
		return x.Create
	}
	return nil
}

func (x *JsAction) GetCall() *Call {
	if x, ok := x.GetValue().(*JsAction_Call); ok {
		return x.Call
	}
	return nil
}

func (x *JsAction) GetTy() int32 {
	if x != nil {
		return x.Ty
	}
	return 0
}

type isJsAction_Value interface {
	isJsAction_Value()
}

type JsAction_Create struct {
	Create *Create `protobuf:"bytes,1,opt,name=create,proto3,oneof"`
}

type JsAction_Call struct {
	Call *Call `protobuf:"bytes,2,opt,name=call,proto3,oneof"`
}

func (*JsAction_Create) isJsAction_Value() {}

func (*JsAction_Call) isJsAction_Value() {}

type JsLog struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data string `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *JsLog) Reset() {
	*x = JsLog{}
	if protoimpl.UnsafeEnabled {
		mi := &file_js_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JsLog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JsLog) ProtoMessage() {}

func (x *JsLog) ProtoReflect() protoreflect.Message {
	mi := &file_js_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JsLog.ProtoReflect.Descriptor instead.
func (*JsLog) Descriptor() ([]byte, []int) {
	return file_js_proto_rawDescGZIP(), []int{3}
}

func (x *JsLog) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

type QueryResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data string `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *QueryResult) Reset() {
	*x = QueryResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_js_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryResult) ProtoMessage() {}

func (x *QueryResult) ProtoReflect() protoreflect.Message {
	mi := &file_js_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryResult.ProtoReflect.Descriptor instead.
func (*QueryResult) Descriptor() ([]byte, []int) {
	return file_js_proto_rawDescGZIP(), []int{4}
}

func (x *QueryResult) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

var File_js_proto protoreflect.FileDescriptor

var file_js_proto_rawDesc = []byte{
	0x0a, 0x08, 0x6a, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x6a, 0x73, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x30, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x6f, 0x64,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x4a, 0x0a, 0x04, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67,
	0x73, 0x22, 0x73, 0x0a, 0x08, 0x4a, 0x73, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x29, 0x0a,
	0x06, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e,
	0x6a, 0x73, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x48, 0x00,
	0x52, 0x06, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x23, 0x0a, 0x04, 0x63, 0x61, 0x6c, 0x6c,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6a, 0x73, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x43, 0x61, 0x6c, 0x6c, 0x48, 0x00, 0x52, 0x04, 0x63, 0x61, 0x6c, 0x6c, 0x12, 0x0e, 0x0a,
	0x02, 0x74, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x74, 0x79, 0x42, 0x07, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x1b, 0x0a, 0x05, 0x4a, 0x73, 0x4c, 0x6f, 0x67, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x22, 0x21, 0x0a, 0x0b, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2e, 0x2f, 0x6a, 0x73, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_js_proto_rawDescOnce sync.Once
	file_js_proto_rawDescData = file_js_proto_rawDesc
)

func file_js_proto_rawDescGZIP() []byte {
	file_js_proto_rawDescOnce.Do(func() {
		file_js_proto_rawDescData = protoimpl.X.CompressGZIP(file_js_proto_rawDescData)
	})
	return file_js_proto_rawDescData
}

var file_js_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_js_proto_goTypes = []interface{}{
	(*Create)(nil),      // 0: jsproto.Create
	(*Call)(nil),        // 1: jsproto.Call
	(*JsAction)(nil),    // 2: jsproto.JsAction
	(*JsLog)(nil),       // 3: jsproto.JsLog
	(*QueryResult)(nil), // 4: jsproto.QueryResult
}
var file_js_proto_depIdxs = []int32{
	0, // 0: jsproto.JsAction.create:type_name -> jsproto.Create
	1, // 1: jsproto.JsAction.call:type_name -> jsproto.Call
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_js_proto_init() }
func file_js_proto_init() {
	if File_js_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_js_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Create); i {
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
		file_js_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Call); i {
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
		file_js_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JsAction); i {
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
		file_js_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JsLog); i {
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
		file_js_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryResult); i {
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
	file_js_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*JsAction_Create)(nil),
		(*JsAction_Call)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_js_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_js_proto_goTypes,
		DependencyIndexes: file_js_proto_depIdxs,
		MessageInfos:      file_js_proto_msgTypes,
	}.Build()
	File_js_proto = out.File
	file_js_proto_rawDesc = nil
	file_js_proto_goTypes = nil
	file_js_proto_depIdxs = nil
}
