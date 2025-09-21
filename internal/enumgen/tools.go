package enumgen

import (
	"sort"
	"strings"

	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/x/format"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
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
			Display:       display,
			Documentation: cleanComment(description),
			Args:          nonEmptyStrings(strings.Split(args, ",")),
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
func CreateMessageDescription(msg *protogen.Message, args Opts, queueToDiscover map[string]*protogen.Message) *MessageDescription {
	fn := string(msg.Desc.FullName())
	if _, ok := messageDescriptions[fn]; ok {
		return messageDescriptions[fn]
	}

	opts := msg.Desc.Options().ProtoReflect()

	display := opts.Get(api.E_MessageDisplay.TypeDescriptor()).String()
	description := opts.Get(api.E_MessageDescription.TypeDescriptor()).String()

	tableSource := opts.Get(api.E_TableSource.TypeDescriptor()).String()
	tableHeader := opts.Get(api.E_TableHeader.TypeDescriptor()).String()

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
		Display:       display,
		TableSource:   tableSource,
		TableHeader:   nonEmptyStrings(strings.Split(tableHeader, ",")),

		ProtogenMessage: msg,
		Package:         strings.Split(fn, ".")[0],
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

	display := opts.Get(api.E_Display.TypeDescriptor()).String()
	description := opts.Get(api.E_Description.TypeDescriptor()).String()
	search := opts.Get(api.E_Search.TypeDescriptor()).String()
	required := opts.Get(api.E_Required.TypeDescriptor()).Bool()
	requiredOr := opts.Get(api.E_RequiredOr.TypeDescriptor()).String()
	min := opts.Get(api.E_Min.TypeDescriptor()).Int()
	max := opts.Get(api.E_Max.TypeDescriptor()).Int()
	minCount := opts.Get(api.E_MinCount.TypeDescriptor()).Int()
	maxCount := opts.Get(api.E_MaxCount.TypeDescriptor()).Int()

	// Fallback to comments if description is empty
	if description == "" {
		description = field.Comments.Leading.String()
	}

	if display == "" {
		display = format.DisplayName(string(field.Desc.Name()))
	}

	fm := &FieldMeta{
		Name:          string(field.Desc.Name()),
		FullName:      string(field.Desc.FullName()),
		Documentation: cleanComment(description),
		Display:       display,
		Required:      required,
		RequiredOr:    nonEmptyStrings(strings.Split(requiredOr, ",")),
		Min:           int32(min),
		Max:           int32(max),
		MinCount:      int32(minCount),
		MaxCount:      int32(maxCount),

		ProtogenField: field,
	}

	fm.SearchOptions, fm.SearchType = parseSearchOptions(search, field)

	kind := field.Desc.Kind()
	isList := field.Desc.IsList()
	isMap := field.Desc.IsMap()

	_, llmTyp := mapScalarToTypes(kind)
	//fm.GoType = goTyp
	fm.Type = llmTyp

	switch kind {
	case protoreflect.MessageKind:
		fm.Type = "object"
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

	if isList || isMap {
		fm.Type = "[]" + llmTyp
	}
	return fm
}

func cleanComment(comment string) string {
	lines := strings.Split(comment, "\n")
	for i, line := range lines {
		lines[i] = strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(line, "//")), "`", "'")
	}
	i := len(lines)
	for i > 0 {
		if lines[i-1] != "" {
			break
		}
		i--
	}

	return strings.Join(lines[:i], "\n")
}

func nonEmptyStrings(items []string) []string {
	var res []string
	for _, item := range items {
		if s := strings.TrimSpace(item); s != "" {
			res = append(res, s)
		}
	}
	return res
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

func TrimLocalPackageName(val, pack string) string {
	if strings.HasPrefix(val, pack+".") {
		return val[len(pack)+1:]
	}
	return val
}

func ExternalPackageName(fullname, pack string) string {
	pn := strings.Split(fullname, ".")[0]
	if pn == "google" || pn == pack || len(pn) == 1 {
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
	Name          string
	Display       string
	Fields        []*FieldMeta
	Documentation string
	FullName      string
	TableSource   string
	TableHeader   []string

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
	StructName      string
	Fields          []*FieldMeta
	EnumDescription *EnumDescription
	Min             int32
	Max             int32
	MinCount        int32
	MaxCount        int32

	// field is the original field descriptor
	ProtogenField         *protogen.Field
	EnumDescriptionName   string
	FieldsDescriptionName string
}
