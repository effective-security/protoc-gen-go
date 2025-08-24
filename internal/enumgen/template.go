package enumgen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/x/enum"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"
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
func ApplyEnumsTemplate(f *protogen.GeneratedFile, opts Opts, enums []*protogen.Enum) error {
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
func ApplyMessagesTemplate(f *protogen.GeneratedFile, opts Opts, msgs []*protogen.Message) error {
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

func ApplyEnums(w io.Writer, opts Opts, enums []*protogen.Enum) error {
	var descriptions []tplEnum
	for _, en := range enums {
		logger.Infof("Processing %s", en.GoIdent.GoName)
		desc := CreateEnumDescription(en, opts)
		descriptions = append(descriptions, tplEnum{
			Enum:        en,
			Description: desc,
		})
	}

	if err := descrEnumsTemplate.Execute(w, tplEnumDescriptions{
		Opts: opts,
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

func ApplyMessages(w io.Writer, opts Opts, msgs []*protogen.Message) error {
	seen := make(map[string]bool)
	mt := tplMessagesMap{
		Opts:         opts,
		Descriptions: make([]*api.MessageDescription, 0, len(msgs)),
	}
	for _, msg := range msgs {
		fn := string(msg.Desc.FullName())
		if _, ok := seen[fn]; ok {
			continue
		}
		seen[fn] = true
		logger.Infof("Processing %s", msg.GoIdent.GoName)
		md := CreateMessageDescription(msg, opts)
		mt.Descriptions = append(mt.Descriptions, md)
		if err := descrMessageTemplate.Execute(w, tplMessage{
			Opts:        opts,
			Message:     msg,
			Description: md,
		}); err != nil {
			return errors.Wrapf(err, "failed to execute template: %s", msg.GoIdent.GoName)
		}
	}

	sort.Slice(mt.Descriptions, func(i, j int) bool {
		return mt.Descriptions[i].FullName < mt.Descriptions[j].FullName
	})

	if err := messagesMapTemplate.Execute(w, mt); err != nil {
		return errors.Wrapf(err, "failed to execute template")
	}
	return nil
}

func GetEnums(msgs []*protogen.Message) []*protogen.Enum {
	var enums []*protogen.Enum
	for _, msg := range msgs {
		if len(msg.Enums) == 0 {
			continue
		}
		for _, en := range msg.Enums {
			enums = append(enums, en)
		}
	}
	return enums
}

func GetMessagesToDescribe(msgs []*protogen.Message, services []*protogen.Service) []*protogen.Message {
	var res []*protogen.Message
	seen := make(map[string]bool)

	// first add all service requests
	for _, svc := range services {
		for _, m := range svc.Methods {
			if _, ok := seen[m.Input.GoIdent.GoName]; !ok {
				res = append(res, m.Input)
				seen[m.Input.GoIdent.GoName] = true
			}
		}
	}

	// then add optionally marked messages
	for _, msg := range msgs {
		opts := msg.Desc.Options().ProtoReflect()
		describe := opts.Get(api.E_GenerateMeta.TypeDescriptor()).Bool()
		if _, ok := seen[msg.GoIdent.GoName]; describe && !ok {
			res = append(res, msg)
			seen[msg.GoIdent.GoName] = true
		}
	}

	return res
}

func tempFuncs() template.FuncMap {
	m := sprig.TxtFuncMap()
	// m["type"] = func(f *protogen.Message) string {
	// 	return path.Base(string(f.GoIdent.GoImportPath)) + "." + f.GoIdent.GoName
	// }
	m["trim_package"] = func(val, pack string) string {
		if strings.HasPrefix(val, pack) {
			return val[len(pack)+1:]
		}
		return val
	}
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
	Description *api.EnumDescription
}

type tplEnumDescriptions struct {
	Opts
	Data []tplEnum
}

type tplMessage struct {
	Opts
	Message     *protogen.Message
	Description *api.MessageDescription
}

type tplMessagesMap struct {
	Opts
	Descriptions []*api.MessageDescription
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
)
`))

	enumTemplate = template.Must(template.New("enum").
			Funcs(tempFuncs()).
			Parse(`
//
// {{.Enum.GoIdent.GoName}}
//

const {{.Enum.GoIdent.GoName}}_SupportedNamesHelp = "{{supported .Enum}}"

// ValuesMap returns map of enum values
func (s {{.Enum.GoIdent.GoName}}) ValuesMap() map[string]int32 {
	return {{.Enum.GoIdent.GoName}}_value
}

// NamesMap returns map of enum names	
func (s {{.Enum.GoIdent.GoName}}) NamesMap() map[int32]string {
	return {{.Enum.GoIdent.GoName}}_name
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
    "{{$.Package}}.{{enum_dot_name .Enum}}": reflect.TypeOf({{.Enum.GoIdent.GoName}}(0)),
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
	{{- if .Description.Documentation }}
	Documentation: ` + "`{{.Description.Documentation}}`" + `,
	{{- end }}
	Fields: []*api.FieldMeta {
	{{- range .Description.Fields }}
		{
			Name: "{{.Name}}",
			FullName: "{{.FullName}}",
			Display: "{{.Display}}",
			Type: "{{.Type}}",
			GoType: "{{.GoType}}",
			{{- if .SearchOptions }}
			SearchOptions: {{search_enum .SearchOptions}},
			{{- end }}
			{{- if .SearchType }}
			SearchType: "{{.SearchType}}",
			{{- end }}
			{{- if .Documentation }}
			Documentation: ` + "`{{.Documentation}}`" + `,
			{{- end }}
			{{- if .Fields }}
			Fields: {{trim_package .GoType $root.Package }}_MessageDescription.Fields,
			{{- end }}
			{{- if .EnumDescription }}
			EnumDescription: {{.EnumDescription.Name}}_EnumDescription,
			{{- end }}
		},
	{{- end }}
	},
}
`))

	messagesMapTemplate = template.Must(template.New("messages_map").
				Funcs(tempFuncs()).
				Parse(`
{{- $root := . }}				
var MessageDescriptions = map[string]*api.MessageDescription {
	{{- range .Descriptions }}
	"{{.FullName}}": {{.Name}}_MessageDescription,
	{{- end }}
}

`))
)
