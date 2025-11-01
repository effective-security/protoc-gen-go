package api

import (
	"fmt"
	"io"
	reflect "reflect"
	"strings"
	"unicode"

	"github.com/effective-security/x/format"
	"github.com/effective-security/x/print"
	"github.com/effective-security/x/values"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gopkg.in/yaml.v3"
)

// Describer is not thread safe.
// It's assumed the registration will be done once at startup

var DefaultDescriber = NewDescriber()

type Describer struct {
	EnumNameTypes map[string]reflect.Type

	pbDisplayNameExtType protoreflect.ExtensionType
}

func NewDescriber(enumNameTypes ...map[string]reflect.Type) *Describer {
	d := &Describer{
		EnumNameTypes: make(map[string]reflect.Type),
	}

	// merge all enum name types
	for _, enumNameTypes := range enumNameTypes {
		d.RegisterEnumNameTypes(enumNameTypes)
	}

	return d
}

func (d *Describer) RegisterEnumNameTypes(enumNameTypes map[string]reflect.Type) {
	for k, v := range enumNameTypes {
		d.EnumNameTypes[k] = v
	}
}

func (d *Describer) protoDisplayValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	if !v.IsValid() {
		return ""
	}
	if fd.IsList() {
		var values []any
		// Handle repeated fields
		list := v.List()
		count := list.Len()
		if count == 0 {
			return ""
		}
		for i := 0; i < count; i++ {
			item := list.Get(i)
			dv := d.protoKindValue(fd, item)
			if count == 1 {
				return dv
			}
			if dv != "" {
				values = append(values, dv)
			}
			if i >= 8 {
				// Limit the number of displayed items to 8
				break
			}
		}
		return values
	}

	return d.protoKindValue(fd, v)
}

func (d *Describer) protoKindValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	var displayValue any

	value := v.Interface()
	switch fd.Kind() {
	case protoreflect.StringKind:
		displayValue = value.(string)
	case protoreflect.Int32Kind:
		displayValue = value.(int32)
	case protoreflect.Int64Kind:
		displayValue = fmt.Sprintf("%d", value.(int64))
	case protoreflect.Uint32Kind:
		displayValue = value.(uint32)
	case protoreflect.Uint64Kind:
		displayValue = fmt.Sprintf("%d", value.(uint64))
	case protoreflect.Sint32Kind:
		displayValue = value.(int32)
	case protoreflect.Sint64Kind:
		displayValue = fmt.Sprintf("%d", value.(int64))
	case protoreflect.FloatKind:
		displayValue = value.(float32)
	case protoreflect.DoubleKind:
		displayValue = value.(float64)
	case protoreflect.BoolKind:
		displayValue = value.(bool)
	case protoreflect.EnumKind:
		enumDesc := fd.Enum()
		enumVal := value.(protoreflect.EnumNumber)
		displayValue = d.GetEnumDisplayValue(enumDesc, int32(enumVal))
	case protoreflect.MessageKind:
		//skip
	default:
		displayValue = fmt.Sprintf("%v", value)
	}
	return displayValue
}

// DescribeMessage converts protobuf message to a human readable dictionary
func DescribeMessage(msg proto.Message) values.MapAny {
	return DefaultDescriber.DescribeMessage(msg)
}

// DescribeMessage converts protobuf message to a human readable dictionary
func (d *Describer) DescribeMessage(msg proto.Message) values.MapAny {
	if msg == nil {
		return nil
	}

	if d.pbDisplayNameExtType == nil {
		d.pbDisplayNameExtType, _ = protoregistry.GlobalTypes.FindExtensionByName("es.api.display")
	}

	// Get the message reflection
	msgReflect := msg.ProtoReflect()

	values := make(values.MapAny)

	// Iterate over the fields
	msgReflect.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		name := string(fd.Name())
		kind := fd.Kind()
		var displayName string
		var displayValue any

		if d.pbDisplayNameExtType != nil {
			displayName, _ = proto.GetExtension(fd.Options(), d.pbDisplayNameExtType).(string)
		}

		if displayName == "" {
			displayName = format.DisplayName(name)
		}

		if kind == protoreflect.MessageKind && v.IsValid() {
			if fd.IsList() {
				list := v.List()
				count := list.Len()
				if count == 0 {
					return true
				}
				var listVals []any
				for i := 0; i < count; i++ {
					item := list.Get(i)

					// Handle nested messages
					vals := d.DescribeMessage(item.Message().Interface())
					if len(vals) > 0 {
						listVals = append(listVals, vals)
					}
				}
				if len(listVals) > 0 {
					values[displayName] = listVals
				}

				return true
			} else if fd.IsMap() {
				// TODO:
				// m := v.Map()
				// if m.Len() == 0 {
				// 	return true
				// }
			} else {
				// Handle nested messages
				vals := d.DescribeMessage(v.Message().Interface())
				if len(vals) > 0 {
					values[displayName] = vals
				}
			}
		} else {
			displayValue = d.protoDisplayValue(fd, v)
			if displayValue != "" {
				values[displayName] = displayValue
			}
		}

		return true
	})
	return values
}

// Describe prints protobuf message to a human readable text
func Describe(w io.Writer, msg proto.Message) {
	DefaultDescriber.Describe(w, msg)
}

// Describe prints protobuf message to a human readable text
func (d *Describer) Describe(w io.Writer, msg proto.Message) {
	vals := d.DescribeMessage(msg)
	enc := yaml.NewEncoder(w)
	_ = enc.Encode(vals)
	_ = enc.Close()
}

// GetEnumDisplayValue function to dynamically call DisplayName on an enum
func GetEnumDisplayValue(enumDescriptor protoreflect.EnumDescriptor, value int32) string {
	return DefaultDescriber.GetEnumDisplayValue(enumDescriptor, value)
}

// GetEnumDisplayValue function to dynamically call DisplayName on an enum
func (d *Describer) GetEnumDisplayValue(enumDescriptor protoreflect.EnumDescriptor, value int32) string {
	// Get the enum full name to locate the concrete Go type
	enumFullName := enumDescriptor.FullName()

	if d.EnumNameTypes != nil {
		// Map enum full name to the actual Go enum type (you need to implement this map)
		goEnumType, ok := d.EnumNameTypes[string(enumFullName)]
		if ok {
			// Use reflection to create a new instance of the enum type and set its value
			enumValue := reflect.New(goEnumType).Elem()
			enumValue.SetInt(int64(value))

			// Try to call DisplayName() if the method exists
			method := enumValue.MethodByName("DisplayName")
			if method.IsValid() {
				result := method.Call(nil)
				if len(result) == 1 && result[0].Kind() == reflect.String {
					return result[0].String()
				}
			}
		}
	}

	// Fallback to the enum name
	enumNumber := protoreflect.EnumNumber(value)
	enumValueDesc := enumDescriptor.Values().ByNumber(enumNumber)
	if enumValueDesc == nil {
		return "Unknown"
	}

	return string(enumValueDesc.Name())
}

func DocumentMessage(w io.Writer, dscr *MessageDescription, indent string) {
	if dscr == nil {
		return
	}

	_, _ = fmt.Fprintf(w, "%s:\n", dscr.Display)
	print.Text(w, dscr.Documentation, indent, false)
	nextIndent := indent + indent
	fieldDocIndent := nextIndent + indent

	_, _ = fmt.Fprint(w, indent)
	_, _ = fmt.Fprint(w, "Fields:\n")
	for _, field := range dscr.Fields {
		_, _ = fmt.Fprint(w, nextIndent)
		_, _ = fmt.Fprintf(w, "- Field: %s\n", field.Name)
		_, _ = fmt.Fprint(w, nextIndent)

		_, _ = fmt.Fprintf(w, "  Type: %s\n", field.SearchType)
		if field.EnumDescription != nil {
			_, _ = fmt.Fprint(w, nextIndent)
			_, _ = fmt.Fprint(w, "  Enum values: ")
			for idx, enum := range field.EnumDescription.Enums {
				if idx > 0 {
					_, _ = fmt.Fprint(w, ", ")
				}
				_, _ = fmt.Fprintf(w, "%s (%d)", enum.Display, enum.Value)
			}
			_, _ = fmt.Fprintln(w)
		}
		if field.Documentation != "" {
			_, _ = fmt.Fprint(w, nextIndent)
			_, _ = fmt.Fprint(w, "  Documentation: ")
			print.Text(w, field.Documentation, fieldDocIndent, true)
		}
	}
	_, _ = fmt.Fprintln(w)
}

func Documentation(w io.Writer, doc string, indent string, noFirstIndent bool) {
	if doc == "" {
		return
	}
	parts := strings.Split(doc, "\n")
	lines := 0
	last := len(parts) - 1
	for last > 0 && (parts[last] == "" || parts[last] == "\n") {
		last--
	}
	for idx, part := range parts {
		if idx > last {
			break
		}
		if lines > 0 || !noFirstIndent {
			_, _ = fmt.Fprint(w, indent)
		}
		_, _ = fmt.Fprintln(w, part)
		lines++
	}
}

func DocumentationOneLine(w io.Writer, doc string) {
	if doc == "" {
		return
	}
	parts := strings.Split(doc, "\n")
	lines := 0
	prevPartDot := false
	for _, part := range parts {
		part = strings.TrimSpace(part)
		size := len(part)
		if size > 0 {
			if lines > 0 {
				if !prevPartDot && unicode.IsUpper(rune(part[0])) {
					_, _ = fmt.Fprint(w, ".")
				}
				_, _ = fmt.Fprint(w, " ")
			}
			_, _ = fmt.Fprint(w, part)
			prevPartDot = part[size-1] == '.'
			lines++
		}
	}
	if lines > 0 && !prevPartDot {
		_, _ = fmt.Fprint(w, ".")
	}
}
