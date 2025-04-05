package enumgen

import (
	"strings"
	"unicode"

	"github.com/effective-security/protoc-gen-go/api"
	"google.golang.org/protobuf/compiler/protogen"
)

// Convert enum descriptor to EnumMeta message
func createEnumDescription(en *protogen.Enum) *api.EnumDescription {
	res := &api.EnumDescription{
		Name: string(en.GoIdent.GoName),
	}
	for _, value := range en.Values {
		opts := value.Desc.Options().ProtoReflect()
		display := opts.Get(api.E_EnumDisplay.TypeDescriptor()).String()
		description := opts.Get(api.E_EnumDescription.TypeDescriptor()).String()
		args := opts.Get(api.E_EnumArgs.TypeDescriptor()).String()

		// Fallback to comments if description is empty
		if description == "" {
			description = strings.TrimSpace(value.Comments.Leading.String())
		}

		if display == "" {
			display = formatDisplayName(string(value.Desc.Name()))
		}

		meta := &api.EnumMeta{
			Value:         int32(value.Desc.Number()),
			Name:          string(value.Desc.Name()),
			Display:       display,
			Documentation: cleanComment(description),
			Args:          nonEmptyStrings(strings.Split(args, ",")),
		}

		res.Enums = append(res.Enums, meta)
	}
	return res
}

func cleanComment(comment string) string {
	lines := strings.Split(comment, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(strings.TrimPrefix(line, "//"))
	}
	return strings.Join(lines, "\n")
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

func formatDisplayName(name string) string {
	// Transform CamelCase to space-separated words
	var result []rune
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, ' ')
		}
		result = append(result, r)
	}
	return string(result)
}
