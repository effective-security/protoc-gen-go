package enumgen

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/effective-security/protoc-gen-go/api"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"
)

// TSOpts are the options to set for rendering the template.
type TSOpts struct {
	// BaseImportPath provides the base import path for TypeScript code
	// example: src/services/foo/protogen
	BaseImportPath string
}

type FileEnumInfo struct {
	FileName string
	Enums    []*protogen.Enum
}

// This function is called with a param which contains the entire definition of a method.
func ApplyTemplateTS(f *protogen.GeneratedFile, opts TSOpts, fileEnumInfos []FileEnumInfo) error {
	buf := &bytes.Buffer{}

	// Generate imports
	if err := tsImportsTemplate.Execute(buf, tplTSImports{
		Opts:          opts,
		FileEnumInfos: fileEnumInfos,
	}); err != nil {
		return errors.Wrapf(err, "failed to execute imports template")
	}

	// Generate interface definition
	if err := tsInterfaceTemplate.Execute(buf, tplTSInterface{}); err != nil {
		return errors.Wrapf(err, "failed to execute interface template")
	}

	// Generate enum definitions for each file
	for _, fileInfo := range fileEnumInfos {
		if len(fileInfo.Enums) > 0 {
			if err := ApplyTSEnums(buf, opts, fileInfo.Enums); err != nil {
				return err
			}
		}
	}

	_, err := f.Write(buf.Bytes())
	return err
}

func ApplyTSEnums(w io.Writer, opts TSOpts, enums []*protogen.Enum) error {
	for _, en := range enums {
		logger.Infof("Processing TypeScript enum %s", en.GoIdent.GoName)
		desc := CreateEnumDescription(en, Opts{})

		if err := tsEnumTemplate.Execute(w, tplTSEnum{
			Opts:        opts,
			Enum:        en,
			Description: desc,
		}); err != nil {
			return errors.Wrapf(err, "failed to execute enum template: %s", en.GoIdent.GoName)
		}
	}
	return nil
}

func tsTempFuncs() template.FuncMap {
	m := sprig.TxtFuncMap()

	m["enum_ts_name"] = func(f *protogen.Enum) string {
		return strings.TrimSuffix(f.GoIdent.GoName, "_Enum")
	}

	m["enum_ts_function_name"] = func(f *protogen.Enum, suffix string) string {
		baseName := strings.TrimSuffix(f.GoIdent.GoName, "_Enum")
		return fmt.Sprintf("%s%s", baseName, suffix)
	}

	m["enum_ts_import_name"] = func(f *protogen.Enum) string {
		return f.GoIdent.GoName
	}

	m["enum_ts_type"] = func(f *protogen.Enum) string {
		return f.GoIdent.GoName + " | string"
	}

	m["enum_ts_parse_type"] = func(f *protogen.Enum) string {
		return "string | " + f.GoIdent.GoName
	}
	return m
}

type tplTSImports struct {
	Opts          TSOpts
	FileEnumInfos []FileEnumInfo
}

type tplTSInterface struct{}

type tplTSEnum struct {
	Opts        TSOpts
	Enum        *protogen.Enum
	Description *api.EnumDescription
}

var (
	tsImportsTemplate = template.Must(template.New("ts_imports").
				Funcs(tsTempFuncs()).
				Parse(`
// Code generated by protoc-gen-go-json. DO NOT EDIT.

{{- range .FileEnumInfos }}
{{- if .Enums }}
import { {{- range $i, $enum := .Enums }}{{ if $i }}, {{ end }}{{ enum_ts_import_name $enum }}{{- end }} } from '{{ $.Opts.BaseImportPath }}/{{ .FileName }}'
{{- end }}
{{- end }}

`))

	tsInterfaceTemplate = template.Must(template.New("ts_interface").
				Funcs(tsTempFuncs()).
				Parse(`
interface ITypeNameInterface {
    [key: number | string]: string
}

interface INameEnumInterface {
    [key: number | string]: number
}

`))

	tsEnumTemplate = template.Must(template.New("ts_enum").
			Funcs(tsTempFuncs()).
			Parse(`
//
// {{ .Enum.GoIdent.GoName }}
//

export const {{ enum_ts_name .Enum }}Name: ITypeNameInterface = {
{{- with .Enum }}
{{- range $.Description.Enums }}
    {{ .Value }}: '{{ .Name }}',
{{- end }}
{{- end }}
}

export const {{ enum_ts_name .Enum }}DisplayName: ITypeNameInterface = {
{{- with .Enum }}
{{- range $.Description.Enums }}
    {{ .Value }}: '{{ .Display }}',
{{- end }}
{{- end }}
}

export const {{ enum_ts_name .Enum }}NameEnum: INameEnumInterface = {
{{- with .Enum }}
{{- range $.Description.Enums }}
    '{{ .Name }}': {{ .Value }},
{{- end }}
{{- end }}
}

export const {{ enum_ts_name .Enum }}DisplayNameEnum: INameEnumInterface = {
{{- with .Enum }}
{{- range $.Description.Enums }}
    '{{ .Display }}': {{ .Value }},
{{- end }}
{{- end }}
}

export function get{{ enum_ts_function_name .Enum "Name" }}(
    opt: {{ enum_ts_type .Enum }},
): string {
    return {{ enum_ts_name .Enum }}Name[opt] || 'Unknown'
}

export function get{{ enum_ts_function_name .Enum "DisplayName" }}(
    opt: {{ enum_ts_type .Enum }},
): string {
    return {{ enum_ts_name .Enum }}DisplayName[opt] || 'Unknown'
}

export function parse{{ enum_ts_name .Enum }}(
    val: {{ enum_ts_parse_type .Enum }},
): {{ .Enum.GoIdent.GoName }} {
    if (typeof val === 'number') {
        return val
    }
    // Try to parse as number first (for string representations like "0", "2", etc.)
    const numVal = parseInt(val, 10)
    if (!isNaN(numVal) && {{ enum_ts_name .Enum }}Name[numVal] !== undefined) {
        return numVal
    }
    return {{ enum_ts_name .Enum }}NameEnum[val] || {{ enum_ts_name .Enum }}DisplayNameEnum[val] || 0
}

`))
)
