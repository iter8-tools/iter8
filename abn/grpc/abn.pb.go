// protoc --go_out=. --go_opt=paths=source_relative     --go-grpc_out=. --go-grpc_opt=paths=source_relative     abn/grpc/abn.proto
// python -m grpc_tools.protoc -I../../../iter8-tools/iter8/abn/grpc --python_out=. --grpc_python_out=. ../../../iter8-tools/iter8/abn/grpc/abn.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.19.4
// source: abn/grpc/abn.proto

package grpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Application struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// name of (backend) application or service
	// This value is used to identify the Kubernetes objects that make up the service
	// Kubernetes objects that comprise the service should have the label app.kubernetes.io/name set to name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// User or user session identifier
	User string `protobuf:"bytes,2,opt,name=user,proto3" json:"user,omitempty"`
}

func (x *Application) Reset() {
	*x = Application{}
	if protoimpl.UnsafeEnabled {
		mi := &file_abn_grpc_abn_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Application) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Application) ProtoMessage() {}

func (x *Application) ProtoReflect() protoreflect.Message {
	mi := &file_abn_grpc_abn_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Application.ProtoReflect.Descriptor instead.
func (*Application) Descriptor() ([]byte, []int) {
	return file_abn_grpc_abn_proto_rawDescGZIP(), []int{0}
}

func (x *Application) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Application) GetUser() string {
	if x != nil {
		return x.User
	}
	return ""
}

type Session struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Track or logical name of the application version
	// If this is not available, it will be the version label
	Track string `protobuf:"bytes,1,opt,name=track,proto3" json:"track,omitempty"`
}

func (x *Session) Reset() {
	*x = Session{}
	if protoimpl.UnsafeEnabled {
		mi := &file_abn_grpc_abn_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Session) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Session) ProtoMessage() {}

func (x *Session) ProtoReflect() protoreflect.Message {
	mi := &file_abn_grpc_abn_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Session.ProtoReflect.Descriptor instead.
func (*Session) Descriptor() ([]byte, []int) {
	return file_abn_grpc_abn_proto_rawDescGZIP(), []int{1}
}

func (x *Session) GetTrack() string {
	if x != nil {
		return x.Track
	}
	return ""
}

type MetricValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Metric name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Metric value
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	// name of application
	Application string `protobuf:"bytes,3,opt,name=application,proto3" json:"application,omitempty"`
	// User or user session identifier
	User string `protobuf:"bytes,4,opt,name=user,proto3" json:"user,omitempty"`
}

func (x *MetricValue) Reset() {
	*x = MetricValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_abn_grpc_abn_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricValue) ProtoMessage() {}

func (x *MetricValue) ProtoReflect() protoreflect.Message {
	mi := &file_abn_grpc_abn_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricValue.ProtoReflect.Descriptor instead.
func (*MetricValue) Descriptor() ([]byte, []int) {
	return file_abn_grpc_abn_proto_rawDescGZIP(), []int{2}
}

func (x *MetricValue) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MetricValue) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *MetricValue) GetApplication() string {
	if x != nil {
		return x.Application
	}
	return ""
}

func (x *MetricValue) GetUser() string {
	if x != nil {
		return x.User
	}
	return ""
}

var File_abn_grpc_abn_proto protoreflect.FileDescriptor

var file_abn_grpc_abn_proto_rawDesc = []byte{
	0x0a, 0x12, 0x61, 0x62, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x61, 0x62, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x6d, 0x61, 0x69, 0x6e, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74,
	0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x35, 0x0a, 0x0b, 0x41, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x73,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x22, 0x1f,
	0x0a, 0x07, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x72, 0x61,
	0x63, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x72, 0x61, 0x63, 0x6b, 0x22,
	0x6d, 0x0a, 0x0b, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x61, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x61,
	0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x73,
	0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x32, 0x6f,
	0x0a, 0x03, 0x41, 0x42, 0x4e, 0x12, 0x2c, 0x0a, 0x06, 0x4c, 0x6f, 0x6f, 0x6b, 0x75, 0x70, 0x12,
	0x11, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x1a, 0x0d, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x0b, 0x57, 0x72, 0x69, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x12, 0x11, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x42,
	0x27, 0x5a, 0x25, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69, 0x74,
	0x65, 0x72, 0x38, 0x2d, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x69, 0x74, 0x65, 0x72, 0x38, 0x2f,
	0x61, 0x62, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_abn_grpc_abn_proto_rawDescOnce sync.Once
	file_abn_grpc_abn_proto_rawDescData = file_abn_grpc_abn_proto_rawDesc
)

func file_abn_grpc_abn_proto_rawDescGZIP() []byte {
	file_abn_grpc_abn_proto_rawDescOnce.Do(func() {
		file_abn_grpc_abn_proto_rawDescData = protoimpl.X.CompressGZIP(file_abn_grpc_abn_proto_rawDescData)
	})
	return file_abn_grpc_abn_proto_rawDescData
}

var file_abn_grpc_abn_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_abn_grpc_abn_proto_goTypes = []interface{}{
	(*Application)(nil),   // 0: main.Application
	(*Session)(nil),       // 1: main.Session
	(*MetricValue)(nil),   // 2: main.MetricValue
	(*emptypb.Empty)(nil), // 3: google.protobuf.Empty
}
var file_abn_grpc_abn_proto_depIdxs = []int32{
	0, // 0: main.ABN.Lookup:input_type -> main.Application
	2, // 1: main.ABN.WriteMetric:input_type -> main.MetricValue
	1, // 2: main.ABN.Lookup:output_type -> main.Session
	3, // 3: main.ABN.WriteMetric:output_type -> google.protobuf.Empty
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_abn_grpc_abn_proto_init() }
func file_abn_grpc_abn_proto_init() {
	if File_abn_grpc_abn_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_abn_grpc_abn_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Application); i {
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
		file_abn_grpc_abn_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Session); i {
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
		file_abn_grpc_abn_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetricValue); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_abn_grpc_abn_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_abn_grpc_abn_proto_goTypes,
		DependencyIndexes: file_abn_grpc_abn_proto_depIdxs,
		MessageInfos:      file_abn_grpc_abn_proto_msgTypes,
	}.Build()
	File_abn_grpc_abn_proto = out.File
	file_abn_grpc_abn_proto_rawDesc = nil
	file_abn_grpc_abn_proto_goTypes = nil
	file_abn_grpc_abn_proto_depIdxs = nil
}
