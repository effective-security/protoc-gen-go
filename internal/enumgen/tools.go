package enumgen

import (
	"strings"

	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/x/format"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// CreateEnumDescription convert enum descriptor to EnumMeta message
func CreateEnumDescription(en *protogen.Enum, args Opts) *api.EnumDescription {
	opts := en.Desc.Options().ProtoReflect()
	IsBitmask := opts.Get(api.E_IsBitmask.TypeDescriptor()).Bool()

	res := &api.EnumDescription{
		Name:          string(en.GoIdent.GoName),
		Documentation: cleanComment(en.Comments.Leading.String()),
		IsBitmask:     IsBitmask,
		FullName:      string(en.Desc.FullName()),
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
	return res
}

// CreateMessageDescription convert enum descriptor to EnumMeta message
func CreateMessageDescription(msg *protogen.Message, args Opts) *api.MessageDescription {
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

	res := &api.MessageDescription{
		Name:          string(msg.GoIdent.GoName),
		FullName:      string(msg.Desc.FullName()),
		Documentation: cleanComment(description),
		Display:       display,
		TableSource:   tableSource,
		TableHeader:   nonEmptyStrings(strings.Split(tableHeader, ",")),
	}

	for _, field := range msg.Fields {
		res.Fields = append(res.Fields, fieldMeta(field, args))
	}

	return res
}

func fieldMeta(field *protogen.Field, args Opts) *api.FieldMeta {
	opts := field.Desc.Options().ProtoReflect()

	display := opts.Get(api.E_Display.TypeDescriptor()).String()
	description := opts.Get(api.E_Description.TypeDescriptor()).String()
	search := opts.Get(api.E_Search.TypeDescriptor()).String()
	required := opts.Get(api.E_Required.TypeDescriptor()).Bool()
	requiredOr := opts.Get(api.E_RequiredOr.TypeDescriptor()).String()
	min := opts.Get(api.E_Min.TypeDescriptor()).Uint()
	max := opts.Get(api.E_Max.TypeDescriptor()).Uint()

	// Fallback to comments if description is empty
	if description == "" {
		description = field.Comments.Leading.String()
	}

	if display == "" {
		display = format.DisplayName(string(field.Desc.Name()))
	}

	fm := &api.FieldMeta{
		Name:          string(field.Desc.Name()),
		FullName:      string(field.Desc.FullName()),
		Documentation: cleanComment(description),
		Display:       display,
		Required:      required,
		RequiredOr:    nonEmptyStrings(strings.Split(requiredOr, ",")),
		Min:           int32(min),
		Max:           int32(max),
	}

	fm.SearchOptions, fm.SearchType = parseSearchOptions(search, field)

	goTyp, llmTyp := mapScalarToTypes(field.Desc.Kind())
	fm.GoType = goTyp
	fm.Type = llmTyp

	switch {
	case field.Desc.Kind() == protoreflect.MessageKind:
		if field.Desc.IsMap() {
			fm.Type = "object"
			// TODO: map
			fm.GoType = "map"
		} else {
			if fm.SearchType == "object" || fm.SearchType == "nested" {
				msg := field.Message
				for _, field := range msg.Fields {
					fm.Fields = append(fm.Fields, fieldMeta(field, args))
				}
				fm.GoType = args.Package + "." + msg.GoIdent.GoName
			}
		}
	case field.Desc.Kind() == protoreflect.EnumKind:
		enumDescr := CreateEnumDescription(field.Enum, args)
		//fm.GoType = enumDescr.Name
		fm.EnumDescription = enumDescr
	case field.Desc.IsList():
		fm.GoType = "[]" + goTyp
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
