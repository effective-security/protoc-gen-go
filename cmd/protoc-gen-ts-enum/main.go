package main

import (
	"flag"
	"os"
	"strings"

	"github.com/effective-security/protoc-gen-go/internal/enumgen"
	"github.com/effective-security/xlog"
	"google.golang.org/protobuf/compiler/protogen"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "go-enums")

var (
	log        = flag.Bool("logs", false, "output logs")
	out        = flag.String("out", "enums.ts", "output file name")
	importpath = flag.String("import", "", "TS import base path")
)

func main() {
	flag.Parse()
	defer logger.Flush()

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gp *protogen.Plugin) error {
		var formatter xlog.Formatter
		if *log {
			formatter = xlog.NewStringFormatter(os.Stderr).
				Options(xlog.FormatNoCaller, xlog.FormatSkipTime, xlog.FormatSkipLevel)
			xlog.SetGlobalLogLevel(xlog.INFO)
		} else {
			formatter = xlog.NewNilFormatter()
		}
		xlog.SetFormatter(formatter)

		opts := enumgen.TSOpts{
			BaseImportPath: *importpath,
		}

		var fileEnumInfos []enumgen.FileEnumInfo

		for _, name := range gp.Request.FileToGenerate {
			f := gp.FilesByPath[name]
			if len(f.Enums) == 0 && len(f.Messages) == 0 {
				logger.Infof("Skipping %s, no enums", name)
				continue
			}

			enums := enumgen.GetEnums(f.Messages)
			enums = append(enums, f.Enums...)

			tsname := strings.TrimSuffix(name, ".proto")
			if len(enums) == 0 {
				logger.Infof("Skipping %s, no enums", name)
				continue
			}

			fileEnumInfos = append(fileEnumInfos, enumgen.FileEnumInfo{
				FileName: tsname,
				Enums:    enums,
			})
		}

		if len(fileEnumInfos) > 0 {
			logger.Infof("Generating %s\n", *out)

			// TODO: do we need protogen.GoImportPath ?
			f := gp.NewGeneratedFile(*out, protogen.GoImportPath(*importpath))
			err := enumgen.ApplyTemplateTS(f, opts, fileEnumInfos)
			if err != nil {
				gp.Error(err)
			}
		}

		return nil
	})
}
