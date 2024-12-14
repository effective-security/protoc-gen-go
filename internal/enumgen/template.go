package enumgen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"path"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/effective-security/x/enum"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "enumgen")

// simply add ref
type fakeEnum = enum.Enum

// This function is called with a param which contains the entire definition of a method.
func ApplyTemplate(w io.Writer, f *protogen.File) error {
	buf := &bytes.Buffer{}
	if err := headerTemplate.Execute(buf, tplHeader{
		File: f,
	}); err != nil {
		return errors.Wrapf(err, "failed to execute template: %s", f.GeneratedFilenamePrefix)
	}

	if err := applyEnums(buf, f.Enums); err != nil {
		return err
	}
	if err := applyMessages(buf, f.Messages); err != nil {
		return err
	}
	src := buf.Bytes()
	code, err := format.Source(src)
	if err != nil {
		fmt.Printf("failed to format source: %s\n%s\n", f.GeneratedFilenamePrefix, string(src))
		return errors.Wrapf(err, "failed to format source: %s", f.GeneratedFilenamePrefix)
	}
	_, err = w.Write(code)
	return err
}

func applyEnums(w io.Writer, enums []*protogen.Enum) error {
	for _, en := range enums {
		logger.Infof("Processing %s", en.GoIdent.GoName)

		if err := enumTemplate.Execute(w, tplEnum{
			Enum: en,
		}); err != nil {
			return errors.Wrapf(err, "failed to execute template: %s", en.GoIdent.GoName)
		}
	}

	return nil
}

func applyMessages(w io.Writer, msgs []*protogen.Message) error {
	for _, msg := range msgs {
		if len(msg.Enums) == 0 {
			continue
		}

		for _, en := range msg.Enums {
			logger.Infof("Processing %s_%s", msg.GoIdent.GoName, en.GoIdent.GoName)

			if err := enumTemplate.Execute(w, tplEnum{
				Enum: en,
			}); err != nil {
				return errors.Wrapf(err, "failed to execute template: %s", en.GoIdent.GoName)
			}
		}
	}

	return nil
}

func tempFuncs() template.FuncMap {
	m := sprig.TxtFuncMap()
	m["type"] = func(f *protogen.Message) string {
		return path.Base(string(f.GoIdent.GoImportPath)) + "." + f.GoIdent.GoName
	}
	m["supported"] = func(f *protogen.Enum) string {
		var names []string
		for _, v := range f.Values {
			names = append(names, string(v.Desc.Name()))
		}
		return strings.Join(names, ",")
	}
	return m
}

type tplHeader struct {
	*protogen.File
}

type tplEnum struct {
	Enum    *protogen.Enum
	Message *protogen.Message
}

var (
	headerTemplate = template.Must(template.New("header").
			Funcs(tempFuncs()).
			Parse(`
// Code generated by protoc-gen-go-json. DO NOT EDIT.
// source: {{.Proto.Name}}

package {{.GoPackageName}}

import (
	{{.GoImportPath}}
	"google.golang.org/protobuf/proto"
	"github.com/effective-security/x/enum"
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

`))
)
