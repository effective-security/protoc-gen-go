// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.27.2
// source: annotations.proto

package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type EnumMeta struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Value         int32                  `protobuf:"varint,1,opt,name=Value,proto3" json:"Value,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	FullName      string                 `protobuf:"bytes,3,opt,name=FullName,proto3" json:"FullName,omitempty"`
	Display       string                 `protobuf:"bytes,4,opt,name=Display,proto3" json:"Display,omitempty"`
	Documentation string                 `protobuf:"bytes,5,opt,name=Documentation,proto3" json:"Documentation,omitempty"`
	Args          []string               `protobuf:"bytes,6,rep,name=Args,proto3" json:"Args,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EnumMeta) Reset() {
	*x = EnumMeta{}
	mi := &file_annotations_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EnumMeta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnumMeta) ProtoMessage() {}

func (x *EnumMeta) ProtoReflect() protoreflect.Message {
	mi := &file_annotations_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnumMeta.ProtoReflect.Descriptor instead.
func (*EnumMeta) Descriptor() ([]byte, []int) {
	return file_annotations_proto_rawDescGZIP(), []int{0}
}

func (x *EnumMeta) GetValue() int32 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *EnumMeta) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *EnumMeta) GetFullName() string {
	if x != nil {
		return x.FullName
	}
	return ""
}

func (x *EnumMeta) GetDisplay() string {
	if x != nil {
		return x.Display
	}
	return ""
}

func (x *EnumMeta) GetDocumentation() string {
	if x != nil {
		return x.Documentation
	}
	return ""
}

func (x *EnumMeta) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

type FieldMeta struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	FullName      string                 `protobuf:"bytes,2,opt,name=FullName,proto3" json:"FullName,omitempty"`
	Display       string                 `protobuf:"bytes,3,opt,name=Display,proto3" json:"Display,omitempty"`
	Documentation string                 `protobuf:"bytes,4,opt,name=Documentation,proto3" json:"Documentation,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FieldMeta) Reset() {
	*x = FieldMeta{}
	mi := &file_annotations_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *FieldMeta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldMeta) ProtoMessage() {}

func (x *FieldMeta) ProtoReflect() protoreflect.Message {
	mi := &file_annotations_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldMeta.ProtoReflect.Descriptor instead.
func (*FieldMeta) Descriptor() ([]byte, []int) {
	return file_annotations_proto_rawDescGZIP(), []int{1}
}

func (x *FieldMeta) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FieldMeta) GetFullName() string {
	if x != nil {
		return x.FullName
	}
	return ""
}

func (x *FieldMeta) GetDisplay() string {
	if x != nil {
		return x.Display
	}
	return ""
}

func (x *FieldMeta) GetDocumentation() string {
	if x != nil {
		return x.Documentation
	}
	return ""
}

type EnumDescription struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Enums         []*EnumMeta            `protobuf:"bytes,2,rep,name=Enums,proto3" json:"Enums,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EnumDescription) Reset() {
	*x = EnumDescription{}
	mi := &file_annotations_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EnumDescription) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnumDescription) ProtoMessage() {}

func (x *EnumDescription) ProtoReflect() protoreflect.Message {
	mi := &file_annotations_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnumDescription.ProtoReflect.Descriptor instead.
func (*EnumDescription) Descriptor() ([]byte, []int) {
	return file_annotations_proto_rawDescGZIP(), []int{2}
}

func (x *EnumDescription) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *EnumDescription) GetEnums() []*EnumMeta {
	if x != nil {
		return x.Enums
	}
	return nil
}

type MessageDescription struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Fields        []*FieldMeta           `protobuf:"bytes,2,rep,name=Fields,proto3" json:"Fields,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *MessageDescription) Reset() {
	*x = MessageDescription{}
	mi := &file_annotations_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MessageDescription) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageDescription) ProtoMessage() {}

func (x *MessageDescription) ProtoReflect() protoreflect.Message {
	mi := &file_annotations_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageDescription.ProtoReflect.Descriptor instead.
func (*MessageDescription) Descriptor() ([]byte, []int) {
	return file_annotations_proto_rawDescGZIP(), []int{3}
}

func (x *MessageDescription) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MessageDescription) GetFields() []*FieldMeta {
	if x != nil {
		return x.Fields
	}
	return nil
}

var file_annotations_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         1071,
		Name:          "es.api.allowed_roles",
		Tag:           "bytes,1071,opt,name=allowed_roles",
		Filename:      "annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         51001,
		Name:          "es.api.search",
		Tag:           "bytes,51001,opt,name=search",
		Filename:      "annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         51002,
		Name:          "es.api.display",
		Tag:           "bytes,51002,opt,name=display",
		Filename:      "annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         51003,
		Name:          "es.api.description",
		Tag:           "bytes,51003,opt,name=description",
		Filename:      "annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         51004,
		Name:          "es.api.csv",
		Tag:           "bytes,51004,opt,name=csv",
		Filename:      "annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.EnumValueOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         52001,
		Name:          "es.api.enum_args",
		Tag:           "bytes,52001,opt,name=enum_args",
		Filename:      "annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.EnumValueOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         52002,
		Name:          "es.api.enum_display",
		Tag:           "bytes,52002,opt,name=enum_display",
		Filename:      "annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.EnumValueOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         52003,
		Name:          "es.api.enum_description",
		Tag:           "bytes,52003,opt,name=enum_description",
		Filename:      "annotations.proto",
	},
}

// Extension fields to descriptorpb.MethodOptions.
var (
	// optional string allowed_roles = 1071;
	E_AllowedRoles = &file_annotations_proto_extTypes[0]
)

// Extension fields to descriptorpb.FieldOptions.
var (
	// search is the option for OpenSearch Index.
	// It can include Index type:
	// keyword|text|nested|object|flat_object|no_index,exclude.
	// - `flat_object` for nested fields that should not be indexed.
	// - `no_index` for fields that should not be indexed.
	// - `exclude` for fields that should not be included in the search results.
	//
	// optional string search = 51001;
	E_Search = &file_annotations_proto_extTypes[1]
	// display is the option for the field's Display Name in the UI.
	//
	// optional string display = 51002;
	E_Display = &file_annotations_proto_extTypes[2]
	// description is the option for the field's description.
	//
	// optional string description = 51003;
	E_Description = &file_annotations_proto_extTypes[3]
	// csv is the option for the field's Name in CSV header.
	//
	// optional string csv = 51004;
	E_Csv = &file_annotations_proto_extTypes[4]
)

// Extension fields to descriptorpb.EnumValueOptions.
var (
	// args is the option for the field's arguments,
	// it can be used to specify the arguments for the enum value, as a string
	// of comma-separated values.
	// For example, "arg1,arg2,arg3" will be parsed as a list of strings
	//
	// optional string enum_args = 52001;
	E_EnumArgs = &file_annotations_proto_extTypes[5]
	// display is the option for the field's Display Name in the UI.
	//
	// optional string enum_display = 52002;
	E_EnumDisplay = &file_annotations_proto_extTypes[6]
	// description is the option for the field's description.
	//
	// optional string enum_description = 52003;
	E_EnumDescription = &file_annotations_proto_extTypes[7]
)

var File_annotations_proto protoreflect.FileDescriptor

const file_annotations_proto_rawDesc = "" +
	"\n" +
	"\x11annotations.proto\x12\x06es.api\x1a google/protobuf/descriptor.proto\"\xa4\x01\n" +
	"\bEnumMeta\x12\x14\n" +
	"\x05Value\x18\x01 \x01(\x05R\x05Value\x12\x12\n" +
	"\x04Name\x18\x02 \x01(\tR\x04Name\x12\x1a\n" +
	"\bFullName\x18\x03 \x01(\tR\bFullName\x12\x18\n" +
	"\aDisplay\x18\x04 \x01(\tR\aDisplay\x12$\n" +
	"\rDocumentation\x18\x05 \x01(\tR\rDocumentation\x12\x12\n" +
	"\x04Args\x18\x06 \x03(\tR\x04Args\"{\n" +
	"\tFieldMeta\x12\x12\n" +
	"\x04Name\x18\x01 \x01(\tR\x04Name\x12\x1a\n" +
	"\bFullName\x18\x02 \x01(\tR\bFullName\x12\x18\n" +
	"\aDisplay\x18\x03 \x01(\tR\aDisplay\x12$\n" +
	"\rDocumentation\x18\x04 \x01(\tR\rDocumentation\"M\n" +
	"\x0fEnumDescription\x12\x12\n" +
	"\x04Name\x18\x01 \x01(\tR\x04Name\x12&\n" +
	"\x05Enums\x18\x02 \x03(\v2\x10.es.api.EnumMetaR\x05Enums\"S\n" +
	"\x12MessageDescription\x12\x12\n" +
	"\x04Name\x18\x01 \x01(\tR\x04Name\x12)\n" +
	"\x06Fields\x18\x02 \x03(\v2\x11.es.api.FieldMetaR\x06Fields:D\n" +
	"\rallowed_roles\x12\x1e.google.protobuf.MethodOptions\x18\xaf\b \x01(\tR\fallowedRoles:7\n" +
	"\x06search\x12\x1d.google.protobuf.FieldOptions\x18\xb9\x8e\x03 \x01(\tR\x06search:9\n" +
	"\adisplay\x12\x1d.google.protobuf.FieldOptions\x18\xba\x8e\x03 \x01(\tR\adisplay:A\n" +
	"\vdescription\x12\x1d.google.protobuf.FieldOptions\x18\xbb\x8e\x03 \x01(\tR\vdescription:1\n" +
	"\x03csv\x12\x1d.google.protobuf.FieldOptions\x18\xbc\x8e\x03 \x01(\tR\x03csv:@\n" +
	"\tenum_args\x12!.google.protobuf.EnumValueOptions\x18\xa1\x96\x03 \x01(\tR\benumArgs:F\n" +
	"\fenum_display\x12!.google.protobuf.EnumValueOptions\x18\xa2\x96\x03 \x01(\tR\venumDisplay:N\n" +
	"\x10enum_description\x12!.google.protobuf.EnumValueOptions\x18\xa3\x96\x03 \x01(\tR\x0fenumDescriptionB1Z/github.com/effective-security/protoc-gen-go/apib\x06proto3"

var (
	file_annotations_proto_rawDescOnce sync.Once
	file_annotations_proto_rawDescData []byte
)

func file_annotations_proto_rawDescGZIP() []byte {
	file_annotations_proto_rawDescOnce.Do(func() {
		file_annotations_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_annotations_proto_rawDesc), len(file_annotations_proto_rawDesc)))
	})
	return file_annotations_proto_rawDescData
}

var file_annotations_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_annotations_proto_goTypes = []any{
	(*EnumMeta)(nil),                      // 0: es.api.EnumMeta
	(*FieldMeta)(nil),                     // 1: es.api.FieldMeta
	(*EnumDescription)(nil),               // 2: es.api.EnumDescription
	(*MessageDescription)(nil),            // 3: es.api.MessageDescription
	(*descriptorpb.MethodOptions)(nil),    // 4: google.protobuf.MethodOptions
	(*descriptorpb.FieldOptions)(nil),     // 5: google.protobuf.FieldOptions
	(*descriptorpb.EnumValueOptions)(nil), // 6: google.protobuf.EnumValueOptions
}
var file_annotations_proto_depIdxs = []int32{
	0,  // 0: es.api.EnumDescription.Enums:type_name -> es.api.EnumMeta
	1,  // 1: es.api.MessageDescription.Fields:type_name -> es.api.FieldMeta
	4,  // 2: es.api.allowed_roles:extendee -> google.protobuf.MethodOptions
	5,  // 3: es.api.search:extendee -> google.protobuf.FieldOptions
	5,  // 4: es.api.display:extendee -> google.protobuf.FieldOptions
	5,  // 5: es.api.description:extendee -> google.protobuf.FieldOptions
	5,  // 6: es.api.csv:extendee -> google.protobuf.FieldOptions
	6,  // 7: es.api.enum_args:extendee -> google.protobuf.EnumValueOptions
	6,  // 8: es.api.enum_display:extendee -> google.protobuf.EnumValueOptions
	6,  // 9: es.api.enum_description:extendee -> google.protobuf.EnumValueOptions
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	2,  // [2:10] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}

func init() { file_annotations_proto_init() }
func file_annotations_proto_init() {
	if File_annotations_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_annotations_proto_rawDesc), len(file_annotations_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 8,
			NumServices:   0,
		},
		GoTypes:           file_annotations_proto_goTypes,
		DependencyIndexes: file_annotations_proto_depIdxs,
		MessageInfos:      file_annotations_proto_msgTypes,
		ExtensionInfos:    file_annotations_proto_extTypes,
	}.Build()
	File_annotations_proto = out.File
	file_annotations_proto_goTypes = nil
	file_annotations_proto_depIdxs = nil
}
