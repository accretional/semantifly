// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.3
// source: index.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Roughly corresponding to file extension, how to parse/encode the file.
type DataType int32

const (
	DataType_TEXT DataType = 0
)

// Enum value maps for DataType.
var (
	DataType_name = map[int32]string{
		0: "TEXT",
	}
	DataType_value = map[string]int32{
		"TEXT": 0,
	}
)

func (x DataType) Enum() *DataType {
	p := new(DataType)
	*p = x
	return p
}

func (x DataType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DataType) Descriptor() protoreflect.EnumDescriptor {
	return file_index_proto_enumTypes[0].Descriptor()
}

func (DataType) Type() protoreflect.EnumType {
	return &file_index_proto_enumTypes[0]
}

func (x DataType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DataType.Descriptor instead.
func (DataType) EnumDescriptor() ([]byte, []int) {
	return file_index_proto_rawDescGZIP(), []int{0}
}

// How to *access* the content. Eg locally as a file, remotely as a web page, etc.
type SourceType int32

const (
	SourceType_LOCAL_FILE SourceType = 0
	SourceType_WEBPAGE    SourceType = 1
)

// Enum value maps for SourceType.
var (
	SourceType_name = map[int32]string{
		0: "LOCAL_FILE",
		1: "WEBPAGE",
	}
	SourceType_value = map[string]int32{
		"LOCAL_FILE": 0,
		"WEBPAGE":    1,
	}
)

func (x SourceType) Enum() *SourceType {
	p := new(SourceType)
	*p = x
	return p
}

func (x SourceType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SourceType) Descriptor() protoreflect.EnumDescriptor {
	return file_index_proto_enumTypes[1].Descriptor()
}

func (SourceType) Type() protoreflect.EnumType {
	return &file_index_proto_enumTypes[1]
}

func (x SourceType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SourceType.Descriptor instead.
func (SourceType) EnumDescriptor() ([]byte, []int) {
	return file_index_proto_rawDescGZIP(), []int{1}
}

type Index struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entries []*IndexListEntry `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
}

func (x *Index) Reset() {
	*x = Index{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Index) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Index) ProtoMessage() {}

func (x *Index) ProtoReflect() protoreflect.Message {
	mi := &file_index_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Index.ProtoReflect.Descriptor instead.
func (*Index) Descriptor() ([]byte, []int) {
	return file_index_proto_rawDescGZIP(), []int{0}
}

func (x *Index) GetEntries() []*IndexListEntry {
	if x != nil {
		return x.Entries
	}
	return nil
}

type IndexListEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name              string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	URI               string                 `protobuf:"bytes,2,opt,name=URI,proto3" json:"URI,omitempty"`
	DataType          DataType               `protobuf:"varint,3,opt,name=dataType,proto3,enum=semantifly.DataType" json:"dataType,omitempty"`
	SourceType        SourceType             `protobuf:"varint,4,opt,name=sourceType,proto3,enum=semantifly.SourceType" json:"sourceType,omitempty"`
	FirstAddedTime    *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=first_added_time,json=firstAddedTime,proto3" json:"first_added_time,omitempty"`
	LastRefreshedTime *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=last_refreshed_time,json=lastRefreshedTime,proto3" json:"last_refreshed_time,omitempty"`
	// Possibly to be changed to byte or some more complex representation based on DataType
	Content string `protobuf:"bytes,7,opt,name=content,proto3" json:"content,omitempty"`
	// Map that stores the count of each word
	WordOccurrences map[string]int32 `protobuf:"bytes,8,rep,name=word_occurrences,json=wordOccurrences,proto3" json:"word_occurrences,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

func (x *IndexListEntry) Reset() {
	*x = IndexListEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IndexListEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IndexListEntry) ProtoMessage() {}

func (x *IndexListEntry) ProtoReflect() protoreflect.Message {
	mi := &file_index_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IndexListEntry.ProtoReflect.Descriptor instead.
func (*IndexListEntry) Descriptor() ([]byte, []int) {
	return file_index_proto_rawDescGZIP(), []int{1}
}

func (x *IndexListEntry) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *IndexListEntry) GetURI() string {
	if x != nil {
		return x.URI
	}
	return ""
}

func (x *IndexListEntry) GetDataType() DataType {
	if x != nil {
		return x.DataType
	}
	return DataType_TEXT
}

func (x *IndexListEntry) GetSourceType() SourceType {
	if x != nil {
		return x.SourceType
	}
	return SourceType_LOCAL_FILE
}

func (x *IndexListEntry) GetFirstAddedTime() *timestamppb.Timestamp {
	if x != nil {
		return x.FirstAddedTime
	}
	return nil
}

func (x *IndexListEntry) GetLastRefreshedTime() *timestamppb.Timestamp {
	if x != nil {
		return x.LastRefreshedTime
	}
	return nil
}

func (x *IndexListEntry) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *IndexListEntry) GetWordOccurrences() map[string]int32 {
	if x != nil {
		return x.WordOccurrences
	}
	return nil
}

var File_index_proto protoreflect.FileDescriptor

var file_index_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x73,
	0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x66, 0x6c, 0x79, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3d, 0x0a, 0x05, 0x49, 0x6e,
	0x64, 0x65, 0x78, 0x12, 0x34, 0x0a, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x73, 0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x66, 0x6c,
	0x79, 0x2e, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x07, 0x65, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x22, 0xec, 0x03, 0x0a, 0x0e, 0x49, 0x6e,
	0x64, 0x65, 0x78, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x10, 0x0a, 0x03, 0x55, 0x52, 0x49, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x55,
	0x52, 0x49, 0x12, 0x30, 0x0a, 0x08, 0x64, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x73, 0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x66, 0x6c,
	0x79, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x36, 0x0a, 0x0a, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x16, 0x2e, 0x73, 0x65, 0x6d, 0x61, 0x6e,
	0x74, 0x69, 0x66, 0x6c, 0x79, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x0a, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x44, 0x0a, 0x10,
	0x66, 0x69, 0x72, 0x73, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x65, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x0e, 0x66, 0x69, 0x72, 0x73, 0x74, 0x41, 0x64, 0x64, 0x65, 0x64, 0x54, 0x69,
	0x6d, 0x65, 0x12, 0x4a, 0x0a, 0x13, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x72, 0x65, 0x66, 0x72, 0x65,
	0x73, 0x68, 0x65, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x11, 0x6c, 0x61, 0x73,
	0x74, 0x52, 0x65, 0x66, 0x72, 0x65, 0x73, 0x68, 0x65, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x5a, 0x0a, 0x10, 0x77, 0x6f, 0x72, 0x64,
	0x5f, 0x6f, 0x63, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x18, 0x08, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x73, 0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x66, 0x6c, 0x79, 0x2e,
	0x49, 0x6e, 0x64, 0x65, 0x78, 0x4c, 0x69, 0x73, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x2e, 0x57,
	0x6f, 0x72, 0x64, 0x4f, 0x63, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x0f, 0x77, 0x6f, 0x72, 0x64, 0x4f, 0x63, 0x63, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x63, 0x65, 0x73, 0x1a, 0x42, 0x0a, 0x14, 0x57, 0x6f, 0x72, 0x64, 0x4f, 0x63, 0x63, 0x75,
	0x72, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x2a, 0x14, 0x0a, 0x08, 0x44, 0x61, 0x74, 0x61,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x54, 0x45, 0x58, 0x54, 0x10, 0x00, 0x2a, 0x1c,
	0x0a, 0x0a, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x0a,
	0x4c, 0x4f, 0x43, 0x41, 0x4c, 0x5f, 0x46, 0x49, 0x4c, 0x45, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07,
	0x57, 0x45, 0x42, 0x50, 0x41, 0x47, 0x45, 0x10, 0x01, 0x42, 0x22, 0x5a, 0x20, 0x61, 0x63, 0x63,
	0x72, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x65, 0x6d,
	0x61, 0x6e, 0x74, 0x69, 0x66, 0x6c, 0x79, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_index_proto_rawDescOnce sync.Once
	file_index_proto_rawDescData = file_index_proto_rawDesc
)

func file_index_proto_rawDescGZIP() []byte {
	file_index_proto_rawDescOnce.Do(func() {
		file_index_proto_rawDescData = protoimpl.X.CompressGZIP(file_index_proto_rawDescData)
	})
	return file_index_proto_rawDescData
}

var file_index_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_index_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_index_proto_goTypes = []any{
	(DataType)(0),                 // 0: semantifly.DataType
	(SourceType)(0),               // 1: semantifly.SourceType
	(*Index)(nil),                 // 2: semantifly.Index
	(*IndexListEntry)(nil),        // 3: semantifly.IndexListEntry
	nil,                           // 4: semantifly.IndexListEntry.WordOccurrencesEntry
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_index_proto_depIdxs = []int32{
	3, // 0: semantifly.Index.entries:type_name -> semantifly.IndexListEntry
	0, // 1: semantifly.IndexListEntry.dataType:type_name -> semantifly.DataType
	1, // 2: semantifly.IndexListEntry.sourceType:type_name -> semantifly.SourceType
	5, // 3: semantifly.IndexListEntry.first_added_time:type_name -> google.protobuf.Timestamp
	5, // 4: semantifly.IndexListEntry.last_refreshed_time:type_name -> google.protobuf.Timestamp
	4, // 5: semantifly.IndexListEntry.word_occurrences:type_name -> semantifly.IndexListEntry.WordOccurrencesEntry
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_index_proto_init() }
func file_index_proto_init() {
	if File_index_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_index_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Index); i {
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
		file_index_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*IndexListEntry); i {
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
			RawDescriptor: file_index_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_index_proto_goTypes,
		DependencyIndexes: file_index_proto_depIdxs,
		EnumInfos:         file_index_proto_enumTypes,
		MessageInfos:      file_index_proto_msgTypes,
	}.Build()
	File_index_proto = out.File
	file_index_proto_rawDesc = nil
	file_index_proto_goTypes = nil
	file_index_proto_depIdxs = nil
}
