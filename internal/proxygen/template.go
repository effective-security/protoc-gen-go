package proxygen

import (
	"bytes"
	"go/format"
	"io"
	"path"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "go-proxy")

// simply add ref
var _ = httperror.Error{}

// Options are the options to set for rendering the template.
type Options struct {
	// Package provides package name for the proxy
	Package string
	// Prefix specifies prefix to be added to message types:
	// {{.Prefix}}{{.Message.GoName}}
	// If not provided, the the package name of the process file will be used.
	Prefix string
}

// This function is called with a param which contains the entire definition of a method.
func ApplyTemplate(w io.Writer, f *protogen.File, opts Options) error {
	if opts.Prefix == "" {
		opts.Prefix = string(f.GoPackageName) + "."
	}
	if !strings.HasSuffix(opts.Prefix, ".") {
		opts.Prefix = opts.Prefix + "."
	}

	buf := &bytes.Buffer{}
	if err := headerTemplate.Execute(buf, tplHeader{
		File:    f,
		Options: opts,
	}); err != nil {
		return errors.Wrapf(err, "failed to execute template: %s", f.GeneratedFilenamePrefix)
	}

	if err := applyServices(buf, f.Services, opts); err != nil {
		return err
	}
	code, err := format.Source(buf.Bytes())
	if err != nil {
		return errors.Wrapf(err, "failed to format source: %s", f.GeneratedFilenamePrefix)
	}
	_, err = w.Write(code)
	return err
}

func applyServices(w io.Writer, svcs []*protogen.Service, opts Options) error {
	for _, svc := range svcs {
		logger.Infof("Processing %s", svc.GoName)

		proxyName := "proxy" + svc.GoName + "Server"
		clientName := "proxy" + svc.GoName + "Client"
		if err := serviceTemplate.Execute(w, tplService{
			Service:          svc,
			Options:          opts,
			ProxyStructName:  proxyName,
			ClientStructName: clientName,
			ServerName:       svc.GoName + "Server",
			ClientName:       svc.GoName + "Client",
		}); err != nil {
			return errors.Wrapf(err, "failed to execute template: %s", svc.GoName)
		}

		for _, met := range svc.Methods {
			if err := methodTemplate.Execute(w, tplMethod{
				Service:          svc,
				Method:           met,
				Options:          opts,
				ProxyStructName:  proxyName,
				ClientStructName: clientName,
				Namespace:        string(svc.Desc.FullName()),
			}); err != nil {
				return errors.Wrapf(err, "failed to execute template: %s", met.GoName)
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
	return m
}

type tplHeader struct {
	Options

	File *protogen.File
}

type tplService struct {
	Options

	Service          *protogen.Service
	ProxyStructName  string
	ClientStructName string
	ServerName       string
	ClientName       string
}

type tplMethod struct {
	Options

	Service          *protogen.Service
	Method           *protogen.Method
	ProxyStructName  string
	ClientStructName string
	Namespace        string
}

var (
	headerTemplate = template.Must(template.New("header").
			Funcs(tempFuncs()).
			Parse(`
// Code generated by protoc-gen-go-proxy. DO NOT EDIT.
// source: {{.File.Proto.Name}}

package {{.Package}}

import (
	"context"
	"net/http"

	{{.File.GoImportPath}}
	"google.golang.org/protobuf/proto"
	"github.com/effective-security/porto/xhttp/correlation"
	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/porto/pkg/retriable"
)
`))

	serviceTemplate = template.Must(template.New("service").
			Funcs(tempFuncs()).
			Parse(`
			
type {{.ProxyStructName}} struct {
	srv {{.Prefix}}{{.ServerName}}
}

type {{.ClientStructName}} struct {
	remote   {{.Prefix}}{{.ClientName}}
	callOpts []grpc.CallOption
}

type post{{.ClientStructName}} struct {
	client   retriable.PostRequester
}

// {{.Service.GoName}}ServerToClient returns {{.Prefix}}{{.ClientName}}
func {{.Service.GoName}}ServerToClient(srv {{.Prefix}}{{.ServerName}}) {{.Prefix}}{{.ClientName}} {
	return &{{.ProxyStructName}}{srv}
}

// New{{.ClientName}} returns instance of the {{.ClientName}}
func New{{.ClientName}}(conn *grpc.ClientConn, callOpts []grpc.CallOption) {{.Prefix}}{{.ServerName}} {
	return &{{.ClientStructName}}{
		remote:   {{.Prefix}}New{{.ClientName}}(conn),
		callOpts: callOpts,
	}
}

// New{{.ClientName}}FromProxy returns instance of {{.ClientName}}
func New{{.ClientName}}FromProxy(proxy {{.Prefix}}{{.ClientName}}) {{.Prefix}}{{.ServerName}} {
	return &{{.ClientStructName}}{
		remote: proxy,
	}
}

// New{{.ClientName}}FromProxy returns instance of {{.ClientName}}
func NewHTTP{{.ClientName}}(client retriable.PostRequester) {{.Prefix}}{{.ServerName}} {
	return &post{{.ClientStructName}}{
		client: client,
	}
}

`))
	methodTemplate = template.Must(template.New("method").
			Funcs(tempFuncs()).
			Parse(`

{{- .Method.Comments.Leading -}}
func (s *{{.ProxyStructName}}) {{.Method.GoName}}(ctx context.Context, req *{{type .Method.Input}}, opts ...grpc.CallOption) (*{{type .Method.Output}}, error) {
	// add correlation ID to outgoing RPC calls
	ctx = correlation.WithMetaFromContext(ctx)
	res, err := s.srv.{{.Method.GoName}}(ctx, req)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

{{ .Method.Comments.Leading -}}
func (s *{{.ClientStructName}}) {{.Method.GoName}}(ctx context.Context, req *{{type .Method.Input}}) (*{{type .Method.Output}}, error) {
	// add correlation ID to outgoing RPC calls
	ctx = correlation.WithMetaFromContext(ctx)
	res, err := s.remote.{{.Method.GoName}}(ctx, req, s.callOpts...)
	if err != nil {
		return nil, httperror.NewFromPb(err)
	}
	return res, nil
}

{{ .Method.Comments.Leading -}}
func (s *post{{.ClientStructName}}) {{.Method.GoName}}(ctx context.Context, req *{{type .Method.Input}}) (*{{type .Method.Output}}, error) {
	var res {{type .Method.Output}}
	path := {{.Namespace}}_{{.Method.GoName}}_FullMethodName
	_, _, err := s.client.Post(ctx, path, req, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

`))
)
