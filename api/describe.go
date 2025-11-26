package api

import (
	"encoding/base64"
	"fmt"
	"io"
	reflect "reflect"
	"strings"
	"unicode"

	"github.com/cockroachdb/errors"
	"github.com/effective-security/x/format"
	"github.com/effective-security/x/print"
	"github.com/effective-security/x/values"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gopkg.in/yaml.v3"
)

// DefaultDescriber is not thread safe.
// It's assumed the registration will be done once at startup
var DefaultDescriber = NewDescriber()

// Describer is an interface that describes a protobuf message in human readable format.
type Describer interface {
	ConvertToMap(msg proto.Message) values.MapAny
	Describe(w io.Writer, msg proto.Message)
	GetEnumDisplayValue(enumDescriptor protoreflect.EnumDescriptor, value int32) string
	GetTabularData(msg proto.Message) (*TabularData, error)
	RegisterEnumNameTypes(enumNameTypes map[string]reflect.Type)
}

type describer struct {
	EnumNameTypes map[string]reflect.Type

	pbDisplayNameExtType protoreflect.ExtensionType
}

func NewDescriber(enumNameTypes ...map[string]reflect.Type) Describer {
	d := &describer{
		EnumNameTypes: make(map[string]reflect.Type),
	}

	// merge all enum name types
	for _, enumNameTypes := range enumNameTypes {
		d.RegisterEnumNameTypes(enumNameTypes)
	}

	return d
}

func (d *describer) RegisterEnumNameTypes(enumNameTypes map[string]reflect.Type) {
	for k, v := range enumNameTypes {
		d.EnumNameTypes[k] = v
	}
}

func (d *describer) protoDisplayValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
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
	} else if fd.IsMap() {
		m := v.Map()
		mapVals := make(values.MapAny)
		m.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
			mapVals[key.String()] = d.protoDisplayValue(fd, value)
			return true
		})
		return mapVals
	}

	return d.protoKindValue(fd, v)
}

func (d *describer) protoKindValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
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
	case protoreflect.BytesKind:
		displayValue = base64.StdEncoding.EncodeToString(value.([]byte))
	case protoreflect.EnumKind:
		enumDesc := fd.Enum()
		enumVal := value.(protoreflect.EnumNumber)
		displayValue = d.GetEnumDisplayValue(enumDesc, int32(enumVal))
	case protoreflect.MessageKind:
		displayValue = d.ConvertToMap(v.Message().Interface())
	default:
		displayValue = fmt.Sprintf("%v", value)
	}
	return displayValue
}

// ConvertToMap converts protobuf message to a human readable dictionary
func ConvertToMap(msg proto.Message) values.MapAny {
	return DefaultDescriber.ConvertToMap(msg)
}

// ConvertToMap converts protobuf message to a human readable dictionary
func (d *describer) ConvertToMap(msg proto.Message) values.MapAny {
	if msg == nil {
		return nil
	}

	if d.pbDisplayNameExtType == nil {
		d.pbDisplayNameExtType, _ = protoregistry.GlobalTypes.FindExtensionByName("es.api.display")
	}

	// Get the message reflection
	msgReflect := msg.ProtoReflect()

	vals := make(values.MapAny)

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
					vals := d.ConvertToMap(item.Message().Interface())
					if len(vals) > 0 {
						listVals = append(listVals, vals)
					}
				}
				if len(listVals) > 0 {
					vals[displayName] = listVals
				}

				return true
			} else if fd.IsMap() {
				m := v.Map()
				if m.Len() == 0 {
					return true
				}
				mapVals := make(values.MapAny)
				fdv := fd.MapValue()
				m.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
					mapVals[key.String()] = d.protoDisplayValue(fdv, value)
					return true
				})
				vals[displayName] = mapVals
			} else {
				// Handle nested messages
				nvals := d.ConvertToMap(v.Message().Interface())
				if len(nvals) > 0 {
					vals[displayName] = nvals
				}
			}
		} else {
			displayValue = d.protoDisplayValue(fd, v)
			if displayValue != "" {
				vals[displayName] = displayValue
			}
		}

		return true
	})
	return vals
}

// Describe prints protobuf message to a human readable text
func Describe(w io.Writer, msg proto.Message) {
	DefaultDescriber.Describe(w, msg)
}

// Describe prints protobuf message to a human readable text
func (d *describer) Describe(w io.Writer, msg proto.Message) {
	vals := d.ConvertToMap(msg)
	enc := yaml.NewEncoder(w)
	_ = enc.Encode(vals)
	_ = enc.Close()
}

// GetEnumDisplayValue function to dynamically call DisplayName on an enum
func GetEnumDisplayValue(enumDescriptor protoreflect.EnumDescriptor, value int32) string {
	return DefaultDescriber.GetEnumDisplayValue(enumDescriptor, value)
}

// GetEnumDisplayValue function to dynamically call DisplayName on an enum
func (d *describer) GetEnumDisplayValue(enumDescriptor protoreflect.EnumDescriptor, value int32) string {
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

// DocumentMessage prints the message description to a human readable text
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

type HasMessageDescription interface {
	GetMessageDescription() *MessageDescription
}

func GetTabularData(msg proto.Message) (*TabularData, error) {
	return DefaultDescriber.GetTabularData(msg)
}

func (d *describer) GetTabularData(msg proto.Message) (*TabularData, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}

	var md *MessageDescription
	if hms, ok := msg.(HasMessageDescription); ok {
		md = hms.GetMessageDescription()
	}
	if md == nil {
		return nil, errors.New("message description not found")
	}

	// Get the message reflection
	msgReflect := msg.ProtoReflect()

	tabularData := &TabularData{}

	if len(md.ListSources) == 0 {
		t := &Table{
			ID:     md.Display,
			Header: FilterPrintableFields(md.Fields, nil, nil),
		}
		rows := d.createRow(msgReflect, t.Header)
		t.Rows = []*TableRow{rows}

		tabularData.Tables = append(tabularData.Tables, t)
	} else {

		for _, source := range md.ListSources {
			field := md.FindField(source)
			if field == nil {
				return nil, errors.New("field not found")
			}
			if field.ListOption == ListOption_Disable {
				return nil, errors.New("invalid source for disabled field")
			}
			if field.Type != "[]struct" {
				return nil, errors.New("invalid source for non-struct field")
			}
			t := &Table{
				ID:     field.Display,
				Header: FilterPrintableFields(field.Fields, nil, nil),
			}
			mfd := msgReflect.Descriptor().Fields().ByName(protoreflect.Name(field.Name))
			if mfd == nil {
				return nil, errors.Errorf("field not found in message: %s", field.Name)
			}
			fdv := msgReflect.Get(mfd)
			rows, err := d.createRowsFromList(fdv, t.Header)
			if err != nil {
				return nil, err
			}
			t.Rows = rows

			tabularData.Tables = append(tabularData.Tables, t)
		}
	}
	return tabularData, nil
}

func (d *describer) createRowsFromList(sval protoreflect.Value, fields []*FieldMeta) ([]*TableRow, error) {
	list := sval.List()
	var rows []*TableRow

	for i := 0; i < list.Len(); i++ {
		item := list.Get(i)
		rmsg := item.Message()
		if rmsg == nil {
			continue
		}
		row := d.createRow(rmsg, fields)

		rows = append(rows, row)
	}

	return rows, nil
}

func (d *describer) createRow(rmsg protoreflect.Message, fields []*FieldMeta) *TableRow {
	row := &TableRow{
		RawValue: rmsg.Interface(),
		Fields:   fields,
	}

	mfields := rmsg.Descriptor().Fields()
	for _, field := range fields {
		fd := mfields.ByName(protoreflect.Name(field.Name))
		if fd == nil {
			row.Cells = append(row.Cells, "")
		}
		fdv := rmsg.Get(fd)
		val := d.rowValue(fd, fdv)
		row.Cells = append(row.Cells, val)
	}
	return row
}

func (d *describer) rowValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) string {
	var displayValue string

	value := v.Interface()
	switch fd.Kind() {
	case protoreflect.StringKind:
		displayValue = value.(string)
	case protoreflect.Int32Kind:
		displayValue = fmt.Sprintf("%d", value.(int32))
	case protoreflect.Int64Kind:
		displayValue = fmt.Sprintf("%d", value.(int64))
	case protoreflect.Uint32Kind:
		displayValue = fmt.Sprintf("%d", value.(uint32))
	case protoreflect.Uint64Kind:
		displayValue = fmt.Sprintf("%d", value.(uint64))
	case protoreflect.Sint32Kind:
		displayValue = fmt.Sprintf("%d", value.(int32))
	case protoreflect.Sint64Kind:
		displayValue = fmt.Sprintf("%d", value.(int64))
	case protoreflect.FloatKind:
		displayValue = fmt.Sprintf("%f", value.(float32))
	case protoreflect.DoubleKind:
		displayValue = fmt.Sprintf("%f", value.(float64))
	case protoreflect.BoolKind:
		displayValue = fmt.Sprintf("%t", value.(bool))
	case protoreflect.BytesKind:
		displayValue = base64.StdEncoding.EncodeToString(value.([]byte))
	case protoreflect.EnumKind:
		enumDesc := fd.Enum()
		enumVal := value.(protoreflect.EnumNumber)
		displayValue = d.GetEnumDisplayValue(enumDesc, int32(enumVal))
	case protoreflect.MessageKind:
		displayValue = "..."
	default:
		displayValue = fmt.Sprintf("%v", value)
	}
	return displayValue
}

// TableRow is an individual row in a table.
type TableRow struct {
	// Cells will be as wide as the column definitions array and contain string
	// representation of basic types as:
	// strings, numbers (float64 or int64), booleans, simple maps, lists, or
	// null.
	// See the type field of the column definition for a more detailed
	// description.
	Cells []string

	Fields   []*FieldMeta
	RawValue any
}

type Table struct {
	// ID is the identifier for the resource
	ID string
	// Header is the header row of the table.
	Header []*FieldMeta
	// Rows is an array of rows.
	Rows []*TableRow
}

type TabularData struct {
	// Tables is an array of tables.
	Tables []*Table
}

func (r *TabularData) Print(w io.Writer) {
	for _, table := range r.Tables {
		if table.ID != "" {
			_, _ = fmt.Fprintf(w, "%s:\n\n", table.ID)
		}
		table.Print(w)
	}
}

func (r *Table) Print(w io.Writer) {
	rc := len(r.Rows)
	if rc > 1 {
		table := createTable(w)
		var header []string
		for _, field := range r.Header {
			header = append(header, field.Display)
		}
		table.Header(header)
		for _, row := range r.Rows {
			_ = table.Append(row.Cells)
		}
		_ = table.Render()
	} else if rc == 1 {
		table := createTableSimple(w)
		for i, field := range r.Header {
			_ = table.Append([]string{field.Display, r.Rows[0].Cells[i]})
		}
		_ = table.Render()
	}
	_, _ = fmt.Fprintln(w)
}

func createTable(w io.Writer) *tablewriter.Table {
	return tablewriter.NewTable(w,
		tablewriter.WithConfig(
			tablewriter.Config{
				Row: tw.CellConfig{
					Formatting: tw.CellFormatting{
						AutoWrap:  tw.WrapTruncate,
						Alignment: tw.AlignLeft,
					},
					ColMaxWidths: tw.CellWidth{Global: 64},
				},
			},
		))
}

func createTableSimple(w io.Writer) *tablewriter.Table {
	return tablewriter.NewTable(w,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone,
			//Symbols: tw.NewSymbols(tw.StyleASCII),
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.Off},
				Lines:      tw.Lines{ShowFooterLine: tw.On, ShowHeaderLine: tw.On},
			},
		})),
		tablewriter.WithConfig(
			tablewriter.Config{
				Row: tw.CellConfig{
					Formatting: tw.CellFormatting{
						AutoWrap:  tw.WrapTruncate,
						Alignment: tw.AlignLeft,
					},
					ColMaxWidths: tw.CellWidth{Global: 128},
				},
			},
		))
}
