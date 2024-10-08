// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.0
// source: logger.proto

package logger

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

type Log struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Raw     *Log_Raw `protobuf:"bytes,1,opt,name=raw,proto3" json:"raw,omitempty"`
	Address []byte   `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	Topics  [][]byte `protobuf:"bytes,3,rep,name=topics,proto3" json:"topics,omitempty"`
	Data    []byte   `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	Removed bool     `protobuf:"varint,5,opt,name=removed,proto3" json:"removed,omitempty"`
}

func (x *Log) Reset() {
	*x = Log{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Log) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Log) ProtoMessage() {}

func (x *Log) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Log.ProtoReflect.Descriptor instead.
func (*Log) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{0}
}

func (x *Log) GetRaw() *Log_Raw {
	if x != nil {
		return x.Raw
	}
	return nil
}

func (x *Log) GetAddress() []byte {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *Log) GetTopics() [][]byte {
	if x != nil {
		return x.Topics
	}
	return nil
}

func (x *Log) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Log) GetRemoved() bool {
	if x != nil {
		return x.Removed
	}
	return false
}

type InfoResMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address [][]byte `protobuf:"bytes,1,rep,name=address,proto3" json:"address,omitempty"`
}

func (x *InfoResMessage) Reset() {
	*x = InfoResMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InfoResMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfoResMessage) ProtoMessage() {}

func (x *InfoResMessage) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfoResMessage.ProtoReflect.Descriptor instead.
func (*InfoResMessage) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{1}
}

func (x *InfoResMessage) GetAddress() [][]byte {
	if x != nil {
		return x.Address
	}
	return nil
}

type ConnectReqMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FromBlock uint64 `protobuf:"varint,1,opt,name=fromBlock,proto3" json:"fromBlock,omitempty"`
	Address   []byte `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *ConnectReqMessage) Reset() {
	*x = ConnectReqMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConnectReqMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectReqMessage) ProtoMessage() {}

func (x *ConnectReqMessage) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConnectReqMessage.ProtoReflect.Descriptor instead.
func (*ConnectReqMessage) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{2}
}

func (x *ConnectReqMessage) GetFromBlock() uint64 {
	if x != nil {
		return x.FromBlock
	}
	return 0
}

func (x *ConnectReqMessage) GetAddress() []byte {
	if x != nil {
		return x.Address
	}
	return nil
}

type BlockNumberMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockNumber uint64 `protobuf:"varint,1,opt,name=blockNumber,proto3" json:"blockNumber,omitempty"`
}

func (x *BlockNumberMessage) Reset() {
	*x = BlockNumberMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockNumberMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockNumberMessage) ProtoMessage() {}

func (x *BlockNumberMessage) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockNumberMessage.ProtoReflect.Descriptor instead.
func (*BlockNumberMessage) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{3}
}

func (x *BlockNumberMessage) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

type AddressReqMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address []byte `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *AddressReqMessage) Reset() {
	*x = AddressReqMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddressReqMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddressReqMessage) ProtoMessage() {}

func (x *AddressReqMessage) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddressReqMessage.ProtoReflect.Descriptor instead.
func (*AddressReqMessage) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{4}
}

func (x *AddressReqMessage) GetAddress() []byte {
	if x != nil {
		return x.Address
	}
	return nil
}

type Log_Raw struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockNumber uint64 `protobuf:"varint,1,opt,name=blockNumber,proto3" json:"blockNumber,omitempty"`
	BlockHash   []byte `protobuf:"bytes,2,opt,name=blockHash,proto3" json:"blockHash,omitempty"`
	Index       uint32 `protobuf:"varint,3,opt,name=index,proto3" json:"index,omitempty"`
	TxHash      []byte `protobuf:"bytes,4,opt,name=txHash,proto3" json:"txHash,omitempty"`
	TxIndex     uint32 `protobuf:"varint,5,opt,name=txIndex,proto3" json:"txIndex,omitempty"`
}

func (x *Log_Raw) Reset() {
	*x = Log_Raw{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Log_Raw) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Log_Raw) ProtoMessage() {}

func (x *Log_Raw) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Log_Raw.ProtoReflect.Descriptor instead.
func (*Log_Raw) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Log_Raw) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *Log_Raw) GetBlockHash() []byte {
	if x != nil {
		return x.BlockHash
	}
	return nil
}

func (x *Log_Raw) GetIndex() uint32 {
	if x != nil {
		return x.Index
	}
	return 0
}

func (x *Log_Raw) GetTxHash() []byte {
	if x != nil {
		return x.TxHash
	}
	return nil
}

func (x *Log_Raw) GetTxIndex() uint32 {
	if x != nil {
		return x.TxIndex
	}
	return 0
}

var File_logger_proto protoreflect.FileDescriptor

var file_logger_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x98, 0x02, 0x0a, 0x03, 0x4c, 0x6f, 0x67, 0x12, 0x21, 0x0a, 0x03, 0x72,
	0x61, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65,
	0x72, 0x2e, 0x4c, 0x6f, 0x67, 0x2e, 0x52, 0x61, 0x77, 0x52, 0x03, 0x72, 0x61, 0x77, 0x12, 0x18,
	0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x6f, 0x70, 0x69,
	0x63, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x73,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x1a, 0x8d,
	0x01, 0x0a, 0x03, 0x52, 0x61, 0x77, 0x12, 0x20, 0x0a, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x48, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x16, 0x0a, 0x06,
	0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x74, 0x78,
	0x48, 0x61, 0x73, 0x68, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x78, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x74, 0x78, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x22, 0x2a,
	0x0a, 0x0e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0c, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x4b, 0x0a, 0x11, 0x43, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x66, 0x72, 0x6f, 0x6d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x09, 0x66, 0x72, 0x6f, 0x6d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x18, 0x0a,
	0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x36, 0x0a, 0x12, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x20, 0x0a,
	0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22,
	0x2d, 0x0a, 0x11, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x32, 0x79,
	0x0a, 0x06, 0x4c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x12, 0x38, 0x0a, 0x04, 0x49, 0x6e, 0x66, 0x6f,
	0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65,
	0x72, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x22, 0x00, 0x12, 0x35, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x12, 0x19, 0x2e,
	0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x52, 0x65,
	0x71, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x0b, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65,
	0x72, 0x2e, 0x4c, 0x6f, 0x67, 0x22, 0x00, 0x30, 0x01, 0x32, 0x87, 0x02, 0x0a, 0x05, 0x41, 0x64,
	0x6d, 0x69, 0x6e, 0x12, 0x3e, 0x0a, 0x03, 0x41, 0x64, 0x64, 0x12, 0x19, 0x2e, 0x6c, 0x6f, 0x67,
	0x67, 0x65, 0x72, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x1a, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x22, 0x00, 0x12, 0x41, 0x0a, 0x06, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x12, 0x19, 0x2e,
	0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65,
	0x71, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x1a, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65,
	0x72, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x22, 0x00, 0x12, 0x3d, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x72, 0x74, 0x12,
	0x1a, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75,
	0x6d, 0x62, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3c, 0x0a, 0x04, 0x53, 0x74, 0x6f, 0x70, 0x12, 0x16, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1a, 0x2e, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x22, 0x00, 0x42, 0x0b, 0x5a, 0x09, 0x2e, 0x2e, 0x2f, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_logger_proto_rawDescOnce sync.Once
	file_logger_proto_rawDescData = file_logger_proto_rawDesc
)

func file_logger_proto_rawDescGZIP() []byte {
	file_logger_proto_rawDescOnce.Do(func() {
		file_logger_proto_rawDescData = protoimpl.X.CompressGZIP(file_logger_proto_rawDescData)
	})
	return file_logger_proto_rawDescData
}

var file_logger_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_logger_proto_goTypes = []any{
	(*Log)(nil),                // 0: logger.Log
	(*InfoResMessage)(nil),     // 1: logger.InfoResMessage
	(*ConnectReqMessage)(nil),  // 2: logger.ConnectReqMessage
	(*BlockNumberMessage)(nil), // 3: logger.BlockNumberMessage
	(*AddressReqMessage)(nil),  // 4: logger.AddressReqMessage
	(*Log_Raw)(nil),            // 5: logger.Log.Raw
	(*emptypb.Empty)(nil),      // 6: google.protobuf.Empty
}
var file_logger_proto_depIdxs = []int32{
	5, // 0: logger.Log.raw:type_name -> logger.Log.Raw
	6, // 1: logger.Logger.Info:input_type -> google.protobuf.Empty
	2, // 2: logger.Logger.Connect:input_type -> logger.ConnectReqMessage
	4, // 3: logger.Admin.Add:input_type -> logger.AddressReqMessage
	4, // 4: logger.Admin.Remove:input_type -> logger.AddressReqMessage
	3, // 5: logger.Admin.Start:input_type -> logger.BlockNumberMessage
	6, // 6: logger.Admin.Stop:input_type -> google.protobuf.Empty
	1, // 7: logger.Logger.Info:output_type -> logger.InfoResMessage
	0, // 8: logger.Logger.Connect:output_type -> logger.Log
	3, // 9: logger.Admin.Add:output_type -> logger.BlockNumberMessage
	3, // 10: logger.Admin.Remove:output_type -> logger.BlockNumberMessage
	6, // 11: logger.Admin.Start:output_type -> google.protobuf.Empty
	3, // 12: logger.Admin.Stop:output_type -> logger.BlockNumberMessage
	7, // [7:13] is the sub-list for method output_type
	1, // [1:7] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_logger_proto_init() }
func file_logger_proto_init() {
	if File_logger_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_logger_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Log); i {
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
		file_logger_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*InfoResMessage); i {
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
		file_logger_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ConnectReqMessage); i {
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
		file_logger_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*BlockNumberMessage); i {
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
		file_logger_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*AddressReqMessage); i {
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
		file_logger_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*Log_Raw); i {
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
			RawDescriptor: file_logger_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_logger_proto_goTypes,
		DependencyIndexes: file_logger_proto_depIdxs,
		MessageInfos:      file_logger_proto_msgTypes,
	}.Build()
	File_logger_proto = out.File
	file_logger_proto_rawDesc = nil
	file_logger_proto_goTypes = nil
	file_logger_proto_depIdxs = nil
}
