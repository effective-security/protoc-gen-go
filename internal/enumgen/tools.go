package enumgen

import (
	"path"
	"sort"
	"strings"

	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/x/format"
	"github.com/effective-security/x/slices"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	enumDescriptions    = make(map[string]*EnumDescription)
	messageDescriptions = make(map[string]*MessageDescription)
)

// CreateEnumDescription convert enum descriptor to EnumMeta message
func CreateEnumDescription(en *protogen.Enum) *EnumDescription {
	fn := string(en.Desc.FullName())
	if _, ok := enumDescriptions[fn]; ok {
		return enumDescriptions[fn]
	}

	opts := en.Desc.Options().ProtoReflect()
	IsBitmask := opts.Get(api.E_IsBitmask.TypeDescriptor()).Bool()

	res := &EnumDescription{
		Name:          string(en.GoIdent.GoName),
		Documentation: cleanComment(en.Comments.Leading.String()),
		IsBitmask:     IsBitmask,
		FullName:      fn,

		ProtogenEnum: en,
		Package:      strings.Split(fn, ".")[0],
		FileName:     en.Location.SourceFile,
	}

	for _, value := range en.Values {
		opts := value.Desc.Options().ProtoReflect()
		display := opts.Get(api.E_EnumDisplay.TypeDescriptor()).String()
		description := opts.Get(api.E_EnumDescription.TypeDescriptor()).String()
		args := opts.Get(api.E_EnumArgs.TypeDescriptor()).String()
		group := opts.Get(api.E_EnumGroup.TypeDescriptor()).String()

		// Fallback to comments if description is empty
		if description == "" {
			description = value.Comments.Leading.String()
		}

		if display == "" {
			display = format.DisplayName(string(value.Desc.Name()))
		}

		meta := &api.EnumMeta{
			Value:         int32(value.Desc.Number()),
			Name:          string(value.Desc.Name()),
			FullName:      string(value.Desc.FullName()),
			Documentation: cleanComment(description),
			Args:          slices.StringsSafeSplit(args, ","),
			Group:         group,
		}
		if display != meta.Name {
			meta.Display = display
		}
		res.Enums = append(res.Enums, meta)
	}

	sort.Slice(res.Enums, func(i, j int) bool {
		return res.Enums[i].Value < res.Enums[j].Value
	})
	enumDescriptions[fn] = res
	return res
}

// CreateMessageDescription convert enum descriptor to EnumMeta message
func CreateMessageDescription(msg *protogen.Message, isInput, isOutput bool, args Opts, queueToDiscover map[string]*protogen.Message) *MessageDescription {
	fn := string(msg.Desc.FullName())
	if _, ok := messageDescriptions[fn]; ok {
		return messageDescriptions[fn]
	}

	opts := msg.Desc.Options().ProtoReflect()

	display := opts.Get(api.E_MessageDisplay.TypeDescriptor()).String()
	description := opts.Get(api.E_MessageDescription.TypeDescriptor()).String()
	generateModel := opts.Get(api.E_GenerateModel.TypeDescriptor()).Bool()
	deprecated := false
	ro := msg.Desc.Options()
	if mo, ok := ro.(*descriptorpb.MethodOptions); ok {
		deprecated = mo.GetDeprecated()
	}

	// Fallback to comments if description is empty
	if description == "" {
		description = strings.TrimSpace(msg.Comments.Leading.String())
	}
	if display == "" {
		display = format.DisplayName(string(msg.Desc.Name()))
	}

	res := &MessageDescription{
		Name:          string(msg.GoIdent.GoName),
		FullName:      fn,
		Documentation: cleanComment(description),

		Deprecated:    deprecated,
		IsInput:       isInput,
		IsOutput:      isOutput,
		GenerateModel: generateModel,

		ProtogenMessage: msg,
		Package:         path.Base(string(msg.GoIdent.GoImportPath)),
	}
	if display != res.Name {
		res.Display = display
	}

	for _, field := range msg.Fields {
		res.Fields = append(res.Fields, fieldMeta(field, args, queueToDiscover))
	}

	messageDescriptions[fn] = res
	delete(queueToDiscover, fn)
	return res
}

func fieldMeta(field *protogen.Field, args Opts, queueToDiscover map[string]*protogen.Message) *FieldMeta {
	opts := field.Desc.Options().ProtoReflect()

	alias := opts.Get(api.E_Alias.TypeDescriptor()).String()
	display := opts.Get(api.E_Display.TypeDescriptor()).String()
	description := opts.Get(api.E_Description.TypeDescriptor()).String()
	search := opts.Get(api.E_Search.TypeDescriptor()).String()
	required := opts.Get(api.E_Required.TypeDescriptor()).Bool()
	requiredOr := opts.Get(api.E_RequiredOr.TypeDescriptor()).String()
	min := opts.Get(api.E_Min.TypeDescriptor()).Int()
	max := opts.Get(api.E_Max.TypeDescriptor()).Int()
	minCount := opts.Get(api.E_MinCount.TypeDescriptor()).Int()
	maxCount := opts.Get(api.E_MaxCount.TypeDescriptor()).Int()
	deprecated := false
	if fo, ok := field.Desc.Options().(*descriptorpb.FieldOptions); ok {
		deprecated = fo.GetDeprecated()
	}

	// Fallback to comments if description is empty
	if description == "" {
		description = field.Comments.Leading.String()
	}

	name := string(field.Desc.Name())
	fullname := string(field.Desc.FullName())

	// Handle map fields, key and value
	// we set required to true and Name to Key and Value to make it Exported
	switch name {
	case "key":
		name = "Key"
		required = true
		if s, ok := strings.CutSuffix(fullname, ".key"); ok {
			fullname = s + ".Key"
		}
	case "value":
		required = true
		name = "Value"
		if s, ok := strings.CutSuffix(fullname, ".value"); ok {
			fullname = s + ".Value"
		}
	}
	if display == "" {
		display = format.DisplayName(name)
	}

	fm := &FieldMeta{
		Name:          name,
		FullName:      fullname,
		GoName:        field.GoName,
		Alias:         alias,
		Documentation: cleanComment(description),
		Required:      required,
		RequiredOr:    slices.StringsSafeSplit(requiredOr, ","),
		Min:           int32(min),
		Max:           int32(max),
		MinCount:      int32(minCount),
		MaxCount:      int32(maxCount),
		Deprecated:    deprecated,

		ProtogenField: field,
		Package:       path.Base(string(field.GoIdent.GoImportPath)),
	}
	if display != fm.Name {
		fm.Display = display
	}
	fm.SearchOptions, fm.SearchType = parseSearchOptions(search, field)

	kind := field.Desc.Kind()
	isList := field.Desc.IsList()
	isMap := field.Desc.IsMap()

	goType, _ := mapScalarToTypes(kind)
	//fm.GoType = goTyp
	fm.Type = goType

	switch kind {
	case protoreflect.MessageKind:
		fm.Type = "struct"
		if fm.SearchType == "object" || fm.SearchType == "flat_object" || fm.SearchType == "nested" {
			fm.StructName = string(field.Message.Desc.FullName())
			if msgDescr, ok := messageDescriptions[fm.StructName]; ok {
				fm.Fields = msgDescr.Fields
			} else {
				logger.Infof("*** Adding nested message to discover: %s", fm.StructName)
				queueToDiscover[fm.StructName] = field.Message
			}
		}
	case protoreflect.EnumKind:
		enumDescr := CreateEnumDescription(field.Enum)
		fm.EnumDescription = enumDescr
		if !strings.HasPrefix(enumDescr.FullName, "google.") {
			fm.EnumDescriptionName = ExternalPackageName(enumDescr.FullName, args.Package) + enumDescr.Name + "_EnumDescription"
		}
	}

	if isMap {
		fm.Type = "map"
	}
	if isList {
		fm.Type = "[]" + goType
	}
	return fm
}

func cleanComment(comment string) string {
	lines := strings.Split(comment, "\n")

	newLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(line, "//")), "`", "'")
		if line == "" || strings.HasPrefix(line, "TODO") {
			continue
		}
		newLines = append(newLines, line)
	}
	return strings.Join(newLines, "\n")
}

func mapScalarToTypes(kind protoreflect.Kind) (goType string, llmType string) {
	switch kind {
	case protoreflect.BoolKind:
		return "bool", "boolean"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32", "integer"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64", "integer"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32", "integer"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64", "integer"
	case protoreflect.FloatKind:
		return "float32", "number"
	case protoreflect.DoubleKind:
		return "float64", "number"
	case protoreflect.StringKind:
		return "string", "string"
	case protoreflect.BytesKind:
		return "[]byte", "string"
	case protoreflect.EnumKind:
		return "int32", "integer"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return "struct", "object"
	default:
		return "unknown", "unknown"
	}
}

func TrimLocalPackageName(name, pkg string) string {
	if strings.HasPrefix(name, pkg+".") {
		return name[len(pkg)+1:]
	}
	return name
}

func ExternalPackageName(fullname, pkg string) string {
	pn := strings.Split(fullname, ".")[0]
	if pn == "google" || pn == pkg || len(pn) == 1 {
		return ""
	}
	return pn + "."
}

type EnumDescription struct {
	Name          string
	Enums         []*api.EnumMeta
	Documentation string
	IsBitmask     bool
	FullName      string

	ProtogenEnum *protogen.Enum
	Package      string
	FileName     string
}

type MessageDescription struct {
	Name            string
	Display         string
	Fields          []*FieldMeta
	Documentation   string
	FullName        string
	Deprecated      bool
	RefreshInterval uint32
	// IsInput is true if the message is an input message
	IsInput bool
	// IsOutput is true if the message is an output message
	IsOutput bool

	GenerateModel bool

	// message is the original message descriptor
	ProtogenMessage *protogen.Message
	Package         string
}

type FieldMeta struct {
	Name            string
	FullName        string
	Display         string
	Documentation   string
	Type            string
	SearchType      string
	SearchOptions   api.SearchOption_Enum
	Required        bool
	RequiredOr      []string
	GoName          string
	StructName      string
	Fields          []*FieldMeta
	EnumDescription *EnumDescription
	Min             int32
	Max             int32
	MinCount        int32
	MaxCount        int32
	Deprecated      bool
	Alias           string

	// field is the original field descriptor
	ProtogenField         *protogen.Field
	Package               string
	EnumDescriptionName   string
	FieldsDescriptionName string
}
