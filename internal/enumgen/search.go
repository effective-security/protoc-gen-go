package enumgen

import (
	"strings"

	"github.com/effective-security/protoc-gen-go/api"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var protoTypeToOpenSearchType = map[protoreflect.Kind]string{
	protoreflect.StringKind:   "keyword",
	protoreflect.BytesKind:    "keyword",
	protoreflect.EnumKind:     "integer",
	protoreflect.Int32Kind:    "integer",
	protoreflect.Int64Kind:    "integer",
	protoreflect.Uint32Kind:   "integer",
	protoreflect.Uint64Kind:   "integer",
	protoreflect.Sint32Kind:   "integer",
	protoreflect.Sint64Kind:   "integer",
	protoreflect.Fixed32Kind:  "integer",
	protoreflect.Fixed64Kind:  "integer",
	protoreflect.Sfixed32Kind: "integer",
	protoreflect.Sfixed64Kind: "integer",
	protoreflect.FloatKind:    "float",
	protoreflect.DoubleKind:   "float",
	protoreflect.BoolKind:     "boolean",
	protoreflect.MessageKind:  "flat_object",
}

var sortableType = map[string]bool{
	"keyword":       true,
	"text":          false,
	"integer":       true,
	"long":          true,
	"unsigned_long": true,
	"short":         true,
	"float":         true,
	"double":        true,
	"date":          true,
	"boolean":       true,
}

func parseSearchOptions(searchOpts string, field *protogen.Field) (opts api.SearchOption_Enum, typ string) {
	kind := field.Desc.Kind()
	typ = protoTypeToOpenSearchType[kind]

	tokens := strings.Split(searchOpts, ",")
	for _, token := range tokens {
		switch token {
		case "no_index":
			opts |= api.SearchOption_NoIndex
		case "exclude":
			opts |= api.SearchOption_Exclude
		case "facet":
			opts |= api.SearchOption_Facet
		case "store":
			opts |= api.SearchOption_Store
		case "object":
			typ = "object"
		case "nested":
			typ = "nested"
		case "flat_object":
			typ = "flat_object"
		default:
			if _, ok := sortableType[token]; ok {
				// override the default type
				typ = token
			}
		}
	}

	if sortableType[typ] {
		opts |= api.SearchOption_Sortable
	}

	return
}
