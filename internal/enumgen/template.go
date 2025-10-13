package enumgen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/cockroachdb/errors"
	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/x/enum"
	"github.com/effective-security/xlog"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "enumgen")

// simply add ref
type fakeEnum = enum.Enum

// Opts are the options to set for rendering the template.
type Opts struct {
	// Package provides package name
	Package string
}

// This function is called with a param which contains the entire definition of a method.
func ApplyEnumsTemplate(f *protogen.GeneratedFile, opts Opts, enums []*EnumDescription) error {
	buf := &bytes.Buffer{}
	if err := headerTemplate.Execute(buf, tplHeader{
		Opts: opts,
	}); err != nil {
		return errors.Wrapf(err, "failed to execute template")
	}

	err := ApplyEnums(buf, opts, enums)
	if err != nil {
		return err
	}

	src := buf.Bytes()
	code, err := format.Source(src)
	if err != nil {
		fmt.Printf("failed to format source:\n%s\n", string(src))
		return errors.Wrapf(err, "failed to format source")
	}
	_, err = f.Write(code)
	return err
}

// This function is called with a param which contains the entire definition of a method.
func ApplyMessagesTemplate(f *protogen.GeneratedFile, opts Opts, msgs []*MessageDescription) error {
	buf := &bytes.Buffer{}
	if err := headerTemplate.Execute(buf, tplHeader{
		Opts: opts,
	}); err != nil {
		return errors.Wrapf(err, "failed to execute template")
	}

	err := ApplyMessages(buf, opts, msgs)
	if err != nil {
		return err
	}

	src := buf.Bytes()
	code, err := format.Source(src)
	if err != nil {
		fmt.Printf("failed to format source:\n%s\n", string(src))
		return errors.Wrapf(err, "failed to format source")
	}
	_, err = f.Write(code)
	return err
}

func ApplyEnums(w io.Writer, opts Opts, enums []*EnumDescription) error {
	var descriptions []tplEnum
	for _, en := range enums {
		logger.Infof("Processing %s", en.FullName)
		descriptions = append(descriptions, tplEnum{
			Enum:        en.ProtogenEnum,
			Description: en,
		})
	}

	sort.Slice(descriptions, func(i, j int) bool {
		return descriptions[i].Description.FullName < descriptions[j].Description.FullName
	})

	if err := descrEnumsTemplate.Execute(w, tplEnumDescriptions{
		Data: descriptions,
	}); err != nil {
		return errors.Wrapf(err, "failed to execute template")
	}

	for _, d := range descriptions {
		if err := enumTemplate.Execute(w, d); err != nil {
			return errors.Wrapf(err, "failed to execute template: %s", d.Enum.GoIdent.GoName)
		}
	}

	return nil
}

func ApplyMessages(w io.Writer, opts Opts, msgs []*MessageDescription) error {
	mt := tplMessagesMap{
		Package:      opts.Package,
		Descriptions: msgs,
	}

	sort.Slice(mt.Descriptions, func(i, j int) bool {
		return mt.Descriptions[i].FullName < mt.Descriptions[j].FullName
	})

	// Execute Templates
	for _, md := range mt.Descriptions {
		msg := md.ProtogenMessage
		if err := descrMessageTemplate.Execute(w, tplMessage{
			Message:     msg,
			Description: md,
		}); err != nil {
			return errors.Wrapf(err, "failed to execute template: %s", msg.GoIdent.GoName)
		}
	}

	if err := messagesMapTemplate.Execute(w, mt); err != nil {
		return errors.Wrapf(err, "failed to execute template")
	}
	return nil
}

func GetEnumsDescriptions(gp *protogen.Plugin, opts Opts) []*EnumDescription {
	seenEnums := make(map[string]*protogen.Enum)
	seenMessages := make(map[string]bool)
	msgsToDiscover := make(map[string]*protogen.Message)

	checkMessage := func(msg *protogen.Message) {
		mfn := string(msg.Desc.FullName())
		if _, ok := seenMessages[mfn]; ok {
			return
		}
		seenMessages[mfn] = true

		for _, field := range msg.Fields {
			kind := field.Desc.Kind()
			switch kind {
			case protoreflect.MessageKind:
				ffn := string(field.Message.Desc.FullName())
				if _, ok := seenMessages[ffn]; !ok {
					msgsToDiscover[ffn] = field.Message
				}
			case protoreflect.EnumKind:
				efn := string(field.Enum.Desc.FullName())
				if _, ok := seenEnums[efn]; !ok {
					seenEnums[efn] = field.Enum
				}
			}
		}
	}

	for _, name := range gp.Request.FileToGenerate {
		f := gp.FilesByPath[name]

		for _, en := range f.Enums {
			fn := string(en.Desc.FullName())
			if _, ok := seenEnums[fn]; !ok {
				seenEnums[fn] = en
			}
		}

		for _, msg := range f.Messages {
			checkMessage(msg)
			for _, en := range msg.Enums {
				efn := string(en.Desc.FullName())
				if _, ok := seenEnums[efn]; !ok {
					seenEnums[efn] = en
				}
			}
		}
		for _, svc := range f.Services {
			for _, m := range svc.Methods {
				checkMessage(m.Input)
				checkMessage(m.Output)
				for _, en := range m.Input.Enums {
					efn := string(en.Desc.FullName())
					if _, ok := seenEnums[efn]; !ok {
						seenEnums[efn] = en
					}
				}
				for _, en := range m.Output.Enums {
					efn := string(en.Desc.FullName())
					if _, ok := seenEnums[efn]; !ok {
						seenEnums[efn] = en
					}
				}
			}
		}

		for i := 0; i < 3 && len(msgsToDiscover) > 0; i++ {
			logger.Infof("[%d] *** Discovering nested messages: %d", i, len(msgsToDiscover))
			prev := msgsToDiscover
			msgsToDiscover = make(map[string]*protogen.Message)
			for fn, msg := range prev {
				checkMessage(msg)
				logger.Infof("*** Discovered nested messages: %s", fn)
			}
		}
	}

	var enums []*protogen.Enum
	pkgPrefix := opts.Package + "."
	for efn, en := range seenEnums {
		if opts.Package == "" || strings.HasPrefix(efn, pkgPrefix) {
			enums = append(enums, en)
		}
	}

	sort.Slice(enums, func(i, j int) bool {
		return enums[i].Desc.FullName() < enums[j].Desc.FullName()
	})

	var res []*EnumDescription
	for _, en := range enums {
		desc := CreateEnumDescription(en)
		res = append(res, desc)
	}

	return res
}

func GetMessagesDescriptions(gp *protogen.Plugin, opts Opts) []*MessageDescription {
	seen := make(map[string]*protogen.Message)

	checkMessage := func(msg *protogen.Message) {
		fn := string(msg.Desc.FullName())
		if _, ok := seen[fn]; !ok {
			seen[fn] = msg

			for _, field := range msg.Fields {
				if field.Desc.Kind() == protoreflect.MessageKind {
					fn := string(field.Message.Desc.FullName())
					if _, ok := seen[fn]; !ok {
						seen[fn] = field.Message
					}
				}
			}
		}
	}

	for _, name := range gp.Request.FileToGenerate {
		f := gp.FilesByPath[name]

		// first add all service requests
		for _, svc := range f.Services {
			for _, m := range svc.Methods {
				checkMessage(m.Input)
				checkMessage(m.Output)
			}
		}

		// then add optionally marked messages
		for _, msg := range f.Messages {
			popts := msg.Desc.Options().ProtoReflect()
			describe := popts.Get(api.E_GenerateMeta.TypeDescriptor()).Bool()
			if !describe {
				continue
			}
			checkMessage(msg)
		}
	}

	var list []*MessageDescription
	msgsToDiscover := make(map[string]*protogen.Message)

	for _, msg := range seen {
		desc := CreateMessageDescription(msg, opts, msgsToDiscover)
		list = append(list, desc)
	}

	for i := 0; i < 10 && len(msgsToDiscover) > 0; i++ {
		logger.Infof("[%d] *** Discovering nested messages: %d", i, len(msgsToDiscover))
		prev := msgsToDiscover
		msgsToDiscover = make(map[string]*protogen.Message)
		for fn, msg := range prev {
			desc := CreateMessageDescription(msg, opts, msgsToDiscover)
			list = append(list, desc)
			logger.Infof("*** Discovered nested messages: %s", fn)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].FullName < list[j].FullName
	})

	return list
}

func goName(importPath, name, thisPkg string) string {
	goPkg := path.Base(importPath)
	// If message is in the same package:
	if importPath == thisPkg || goPkg == thisPkg {
		return name
	}

	return fmt.Sprintf("%s.%s", goPkg, name)
}

func tempFuncs() template.FuncMap {
	m := sprig.TxtFuncMap()

	m["allocation_func"] = func(thisPkg string, md *MessageDescription) string {
		if md.ProtogenMessage.Desc.IsMapEntry() {
			keyField := md.Fields[0]
			valField := md.Fields[1]
			valType := valField.Type

			switch valField.Type {
			case "object", "struct":
				msg := messageDescriptions[valField.StructName].ProtogenMessage
				valType = "*" + goName(string(msg.GoIdent.GoImportPath), msg.GoIdent.GoName, thisPkg)
			case "[]object", "[]struct":
				msg := messageDescriptions[valField.StructName].ProtogenMessage
				valType = "[]*" + goName(string(msg.GoIdent.GoImportPath), msg.GoIdent.GoName, thisPkg)
			case "int32":
				if valField.EnumDescription != nil {
					valType = valField.EnumDescription.Name
				}
			}
			return "make(map[" + keyField.Type + "]" + valType + ")"
		}

		msg := md.ProtogenMessage
		return "new(" + goName(string(msg.GoIdent.GoImportPath), msg.GoIdent.GoName, thisPkg) + ")"
	}
	m["trim_package"] = TrimLocalPackageName
	m["package_name"] = ExternalPackageName
	m["supported"] = func(f *protogen.Enum) string {
		var names []string
		for _, v := range f.Values {
			names = append(names, string(v.Desc.Name()))
		}
		return strings.Join(names, ",")
	}
	m["search_enum"] = func(val api.SearchOption_Enum) string {
		var names []string
		exclude := false
		if val&api.SearchOption_NoIndex != 0 {
			names = append(names, "api.SearchOption_NoIndex")
			exclude = true
		}
		if val&api.SearchOption_Exclude != 0 {
			names = append(names, "api.SearchOption_Exclude")
			exclude = true
		}

		if !exclude {
			if val&api.SearchOption_Facet != 0 {
				names = append(names, "api.SearchOption_Facet")
			}
			if val&api.SearchOption_Sortable != 0 {
				names = append(names, "api.SearchOption_Sortable")
			}
			if val&api.SearchOption_Store != 0 {
				names = append(names, "api.SearchOption_Store")
			}
			if val&api.SearchOption_Hidden != 0 {
				names = append(names, "api.SearchOption_Hidden")
			}
			if val&api.SearchOption_WithKeyword != 0 {
				names = append(names, "api.SearchOption_WithKeyword")
			}
			if val&api.SearchOption_WithText != 0 {
				names = append(names, "api.SearchOption_WithText")
			}
		}
		if len(names) == 0 {
			return "api.SearchOption_None"
		}

		return strings.Join(names, "|")
	}
	m["enum_name"] = func(f *protogen.Enum, name string) string {
		return strings.TrimSuffix(f.GoIdent.GoName, "_Enum") + "_" + name
	}
	m["enum_dot_name"] = func(f *protogen.Enum) string {
		if strings.HasSuffix(f.GoIdent.GoName, "_Enum") {
			return strings.TrimSuffix(f.GoIdent.GoName, "_Enum") + ".Enum"
		}
		return f.GoIdent.GoName
	}
	m["list"] = func(vals []string) string {
		if len(vals) == 0 {
			return "nil"
		}
		var result []string
		for _, v := range vals {
			result = append(result, fmt.Sprintf("\"%s\"", v))
		}
		return "[]string{" + strings.Join(result, ",") + "}"
	}
	return m
}

type tplHeader struct {
	Opts
}

type tplEnum struct {
	Enum        *protogen.Enum
	Description *EnumDescription
}

type tplEnumDescriptions struct {
	Data []tplEnum
}

type tplMessage struct {
	Message     *protogen.Message
	Description *MessageDescription
}

type tplMessagesMap struct {
	Package      string
	Descriptions []*MessageDescription
}

var (
	headerTemplate = template.Must(template.New("header").
			Funcs(tempFuncs()).
			Parse(`
// Code generated by protoc-gen-go-json. DO NOT EDIT.

package {{.Package}}

import (
	"google.golang.org/protobuf/proto"
	"github.com/effective-security/x/enum"
	"github.com/effective-security/protoc-gen-go/api"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)
`))

	enumTemplate = template.Must(template.New("enum").
			Funcs(tempFuncs()).
			Parse(`
//
// {{.Enum.GoIdent.GoName}}
//

const {{.Enum.GoIdent.GoName}}_SupportedNamesHelp = "{{supported .Enum}}"

// ValuesMap returns a map of enum values
func (s {{.Enum.GoIdent.GoName}}) ValuesMap() map[string]int32 {
	return {{.Enum.GoIdent.GoName}}_value
}

// NamesMap returns map of enum names 
func (s {{.Enum.GoIdent.GoName}}) NamesMap() map[int32]string {
	return {{.Enum.GoIdent.GoName}}_name
}

// DisplayNamesMap returns a map of enum display names	
func (s {{.Enum.GoIdent.GoName}}) DisplayNamesMap() map[int32]string {
	return {{.Enum.GoIdent.GoName}}_displayValue
}

// SupportedNames returns string of supported Enum name concatenated by ","	
func (s {{.Enum.GoIdent.GoName}}) SupportedNames() string {
	return enum.SupportedNames[{{.Enum.GoIdent.GoName}}]()
}

// ValueNames returns list of Enum value names
func (s {{.Enum.GoIdent.GoName}}) ValueNames() []string {
	return enum.FlagNames(s)
}

// ValueString returns string of Enum value names concatenated by ","
func (s {{.Enum.GoIdent.GoName}}) ValueString() string {
	return strings.Join(s.ValueNames(), ",")
}

// Flags returns list of Enum values
func (s {{.Enum.GoIdent.GoName}}) Flags() []{{.Enum.GoIdent.GoName}} {
	return enum.Flags(s)
}

// FlagsInt returns list of Enum values as int32
func (s {{.Enum.GoIdent.GoName}}) FlagsInt() []int32 {
	return enum.FlagsInt(s)
}

// UnmarshalYAML unmarshals Enum from YAML
func (s *{{.Enum.GoIdent.GoName}}) UnmarshalYAML(unmarshal func(any) error) error {
	// Try to unmarshal as an integer
	var valInt int32
	if err := unmarshal(&valInt); err == nil {
		*s = {{.Enum.GoIdent.GoName}}(valInt)
		return nil
	}

	// Try to unmarshal as a string
	var valStr string
	if err := unmarshal(&valStr); err == nil {
		*s = enum.Parse[{{.Enum.GoIdent.GoName}}](valStr)
		return nil
	}

	// If both attempts fail, set to default
	*s = 0
	return nil
}

// DisplayValues returns display names of Enum bitflag value
func (s {{.Enum.GoIdent.GoName}}) DisplayValues() []string {
	flags := enum.Flags(s)
	count := len(flags)
	if count == 0 {
		return []string{s.String()}
	}
	if count == 1 {
		return []string{ {{.Enum.GoIdent.GoName}}_DisplayValue[flags[0]] }
	}
	var names []string
	for _, flag := range flags {
		names = append(names, {{.Enum.GoIdent.GoName}}_DisplayValue[flag])
	}
	return names
}

// DisplayValue returns display name of Enum value
func (s {{.Enum.GoIdent.GoName}}) DisplayValue() string {
	{{- if .Description.IsBitmask }}
	flags := enum.Flags(s)
	count := len(flags)
	if count == 0 {
		return s.String()
	}
	if count == 1 {
		return {{.Enum.GoIdent.GoName}}_DisplayValue[flags[0]]
	}
	var names []string
	for _, flag := range flags {
		names = append(names, {{.Enum.GoIdent.GoName}}_DisplayValue[flag])
	}
	return strings.Join(names, ",")
	{{- else }}
	if val, ok := {{.Enum.GoIdent.GoName}}_DisplayValue[s]; ok {
		return val
	}
	return s.String()
	{{- end }}
}

// Meta returns Enum meta information
func (s {{.Enum.GoIdent.GoName}}) Meta() *api.EnumMeta {
	return {{.Enum.GoIdent.GoName}}_Meta[s]
}

// Describe returns Enum meta information for all values
func (s {{.Enum.GoIdent.GoName}}) Describe() map[{{.Enum.GoIdent.GoName}}]*api.EnumMeta {
	return {{.Enum.GoIdent.GoName}}_Meta
}

var {{.Enum.GoIdent.GoName}}_Name = map[{{.Enum.GoIdent.GoName}}]string {
{{- with .Enum }}
{{- range $.Description.Enums }}
	{{enum_name $.Enum .Name}}: "{{.Name}}",
{{- end }}
{{- end }}
}

var {{.Enum.GoIdent.GoName}}_Value = map[string]{{.Enum.GoIdent.GoName}} {
{{- with .Enum }}
{{- range $.Description.Enums }}
	"{{.Name}}":{{enum_name $.Enum .Name}},
{{- end }}
{{- end }}
}

var {{.Enum.GoIdent.GoName}}_DisplayValue = map[{{.Enum.GoIdent.GoName}}]string {
{{- with .Enum }}
{{- range $.Description.Enums }}
	{{enum_name $.Enum .Name}}: Display_{{enum_name $.Enum .Name}},
{{- end }}
{{- end }}
}

var {{.Enum.GoIdent.GoName}}_displayValue = map[int32]string {
{{- with .Enum }}
{{- range $.Description.Enums }}
	{{.Value}}: Display_{{enum_name $.Enum .Name}},
{{- end }}
{{- end }}
}

var {{.Enum.GoIdent.GoName}}_EnumDescription = &api.EnumDescription {
	Name: "{{.Description.Name}}",
	FullName: "{{.Description.FullName}}",
	IsBitmask: {{.Description.IsBitmask}},
	Enums: []*api.EnumMeta {
	{{- with .Enum }}
	{{- range $.Description.Enums }}
		{
			Value: {{.Value}},
			Name: "{{.Name}}",
			FullName: "{{.FullName}}",
			Display: Display_{{enum_name $.Enum .Name}},
			{{- if .Args }}
			Args: {{list .Args}},
			{{- end }}
			{{- if .Documentation }}
			Documentation: ` + "`{{.Documentation}}`" + `,
			{{- end }}
		},
	{{- end }}
	{{- end }}
	},
	{{- if .Description.Documentation }}
	Documentation: ` + "`{{.Description.Documentation}}`" + `,
	{{- end }}
}

var {{.Enum.GoIdent.GoName}}_Meta = map[{{.Enum.GoIdent.GoName}}]*api.EnumMeta {
	{{- with .Enum }}
	{{- range $i, $value := $.Description.Enums }}
	{{enum_name $.Enum $value.Name}}: {{$.Enum.GoIdent.GoName}}_EnumDescription.Enums[{{$i}}],
	{{- end }}
	{{- end }}
}

`))

	descrEnumsTemplate = template.Must(template.New("enum_descriptions").
				Funcs(tempFuncs()).
				Parse(`

const (
{{- range .Data }}
{{- $root := . }}
{{- with .Enum }}
  {{- range $root.Description.Enums }}
    Display_{{enum_name $root.Enum .Name}} = "{{.Display}}"
  {{- end }}
{{- end }}
{{- end }}
)

var EnumNameTypes = map[string]reflect.Type{
{{- range .Data }}
{{- $root := . }}
    "{{.Description.Package}}.{{enum_dot_name .Enum}}": reflect.TypeOf({{.Enum.GoIdent.GoName}}(0)),
{{- end }}
}
`))

	descrMessageTemplate = template.Must(template.New("message_descriptions").
				Funcs(tempFuncs()).
				Parse(`
{{- $root := . }}				
var {{.Description.Name}}_MessageDescription = &api.MessageDescription {
	Name: "{{.Description.Name}}",
	Display: "{{.Description.Display}}",
	FullName: "{{.Description.FullName}}",
	{{- if .Description.Documentation }}
	Documentation: ` + "`{{.Description.Documentation}}`" + `,
	{{- end }}
	{{- if .Description.TableSource }}
	TableSource: "{{.Description.TableSource}}",
	{{- end }}
	{{- if .Description.TableHeader }}
	TableHeader: {{list .Description.TableHeader}},
	{{- end }}
	{{- if .Description.Deprecated }}
	Deprecated: true,
	{{- end }}
	Fields: []*api.FieldMeta {
	{{- range .Description.Fields }}
		{
			Name: "{{.Name}}",
			FullName: "{{.FullName}}",
			Display: "{{.Display}}",
			Type: "{{.Type}}",
			{{- if .StructName }}
			StructName: "{{.StructName}}",
			{{- end }}
			{{- if .SearchType }}
			SearchType: "{{.SearchType}}",
			{{- end }}
			{{- if .SearchOptions }}
			SearchOptions: {{search_enum .SearchOptions}},
			{{- end }}
			{{- if .FieldsDescriptionName }}
			Fields: {{ .FieldsDescriptionName }},
			{{- end }}
			{{- if .EnumDescriptionName }}
			EnumDescription: {{ .EnumDescriptionName }},
			{{- end }}
			{{- if .Required }}
			Required: true,
			{{- end }}
			{{- if .RequiredOr }}
			RequiredOr: {{list .RequiredOr}},
			{{- end }}
			{{- if .Min }}
			Min: {{.Min}},
			{{- end }}
			{{- if .Max }}
			Max: {{.Max}},
			{{- end }}
			{{- if .MinCount }}
			MinCount: {{.MinCount}},
			{{- end }}
			{{- if .MaxCount }}
			MaxCount: {{.MaxCount}},
			{{- end }}
			{{- if .Deprecated }}
			Deprecated: true,
			{{- end }}
			{{- if .Documentation }}
			Documentation: ` + "`{{.Documentation}}`" + `,
			{{- end }}
		},
	{{- end }}
	},
}
`))

	messagesMapTemplate = template.Must(template.New("messages_map").
				Funcs(tempFuncs()).
				Parse(`


// MessageAllocator defines constructor to allocate Protobuf message
type MessageAllocator func() any

var (
	initMessageDescriptionOnce sync.Once

{{- $root := . }}				
	messageDescriptions = map[string]*api.MessageDescription {
	{{- range .Descriptions }}
	"{{.FullName}}": {{.Name}}_MessageDescription,
	{{- end }}
  	}

	messageAllocators = map[string]MessageAllocator {
	{{- range .Descriptions }}
	"{{.FullName}}": func() any { return {{allocation_func $root.Package . }} },
	{{- end }}
	}
)

func GetMessageDescriptions() map[string]*api.MessageDescription {
	// Update the message Fields with the nested messages
	initMessageDescriptionOnce.Do(func() {
		for _, md := range messageDescriptions {
			for _, field := range md.Fields {
				if field.Fields == nil && (field.Type == "struct" || field.Type == "[]struct" || field.Type == "object" || field.Type == "[]object") {
					if msgDescr, ok := messageDescriptions[field.StructName]; ok {
						field.Fields = msgDescr.Fields
					}
				}
			}	
		}
	})
	return messageDescriptions
}

func CreateMessage(fullname string) any {
	allocator := messageAllocators[fullname]
	if allocator == nil {
		panic(fmt.Sprintf("allocator for %s not found", fullname))
	}
	return allocator()
}


func GetMessageDescription(fullname string) *api.MessageDescription {
	return GetMessageDescriptions()[fullname]
}

func init() {
    _ = GetMessageDescriptions()
}
`))
)
