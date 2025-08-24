package enumgen

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/effective-security/protoc-gen-go/api"
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
			display = FormatDisplayName(string(value.Desc.Name()))
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
	// Fallback to comments if description is empty
	if description == "" {
		description = strings.TrimSpace(msg.Comments.Leading.String())
	}
	if display == "" {
		display = FormatDisplayName(string(msg.Desc.Name()))
	}

	res := &api.MessageDescription{
		Name:          string(msg.GoIdent.GoName),
		FullName:      string(msg.Desc.FullName()),
		Documentation: cleanComment(description),
		Display:       display,
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

	// Fallback to comments if description is empty
	if description == "" {
		description = field.Comments.Leading.String()
	}

	if display == "" {
		display = FormatDisplayName(string(field.Desc.Name()))
	}

	fm := &api.FieldMeta{
		Name:          string(field.Desc.Name()),
		FullName:      string(field.Desc.FullName()),
		Documentation: cleanComment(description),
		Display:       display,
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

// FormatDisplayName fixes display name conversion to preserve common acronyms
func FormatDisplayName(name string) string {
	s := Split(name)
	if len(s) == 0 {
		return name
	}
	return strings.Join(s, " ")
}

// Split splits the camelcase word and returns a list of words. It also
// supports digits. Both lower camel case and upper camel case are supported.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
//
// Examples
//
//	"" =>                     [""]
//	"lowercase" =>            ["lowercase"]
//	"Class" =>                ["Class"]
//	"MyClass" =>              ["My", "Class"]
//	"MyC" =>                  ["My", "C"]
//	"HTML" =>                 ["HTML"]
//	"PDFLoader" =>            ["PDF", "Loader"]
//	"AString" =>              ["A", "String"]
//	"SimpleXMLParser" =>      ["Simple", "XML", "Parser"]
//	"vimRPCPlugin" =>         ["vim", "RPC", "Plugin"]
//	"GL11Version" =>          ["GL", "11", "Version"]
//	"99Bottles" =>            ["99", "Bottles"]
//	"May5" =>                 ["May", "5"]
//	"BFG9000" =>              ["BFG", "9000"]
//	"BöseÜberraschung" =>     ["Böse", "Überraschung"]
//	"Two  spaces" =>          ["Two", "  ", "spaces"]
//	"BadUTF8\xe2\xe2\xa1" =>  ["BadUTF8\xe2\xe2\xa1"]
//
// Splitting rules
//
//  1. If string is not valid UTF-8, return it without splitting as
//     single item array.
//  2. Assign all unicode characters into one of 4 sets: lower case
//     letters, upper case letters, numbers, and all other characters.
//  3. Iterate through characters of string, introducing splits
//     between adjacent characters that belong to different sets.
//  4. Iterate through array of split strings, and if a given string
//     is upper case:
//     if subsequent string is lower case:
//     move last character of upper case string to beginning of
//     lower case string
func Split(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		case r == '_':
			class = 4
		default:
			class = 10
		}
		if class == lastClass || (lastClass == 2 && class == 3) {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else if class != 4 {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return
}
