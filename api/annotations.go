package api

import (
	"fmt"
	"strings"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

type EnumLike interface {
	Number() protoreflect.EnumNumber
}

func (e *EnumDescription) Parse(val any) int32 {
	var res int32

	if v, ok := val.(EnumLike); ok {
		return int32(v.Number())
	}

	switch v := val.(type) {
	case int32:
		res = v
	case int:
		res = int32(v)
	case int64:
		res = int32(v)
	case uint64:
		res = int32(v)
	case uint32:
		res = int32(v)
	case string:
		var tokens []string
		if strings.Contains(v, ",") {
			tokens = strings.Split(v, ",")
		} else if strings.Contains(v, "|") {
			tokens = strings.Split(v, "|")
		} else {
			tokens = []string{v}
		}
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			for _, enum := range e.Enums {
				if enum.Name == token || enum.FullName == token || enum.Display == token {
					res |= enum.Value
					if !e.IsBitmask {
						break
					}
				}
			}
		}
	case []string:
		for _, token := range v {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			for _, enum := range e.Enums {
				if enum.Name == token || enum.FullName == token || enum.Display == token {
					res |= enum.Value
					if !e.IsBitmask {
						break
					}
				}
			}
		}
	case []int32:
		for _, token := range v {
			res |= token
		}
	case []int:
		for _, token := range v {
			res |= int32(token)
		}
	default:
		panic(fmt.Sprintf("unsupported type: %T", v))
	}

	return res
}

func (m *MessageDescription) FindField(name string) *FieldMeta {
	for _, field := range m.Fields {
		if field.Name == name || field.FullName == name {
			return field
		}
	}
	return nil
}

func FindFieldMeta(fields []*FieldMeta, name string) *FieldMeta {
	for _, field := range fields {
		if field.Name == name || field.FullName == name {
			return field
		}
	}
	return nil
}
