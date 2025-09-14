package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/effective-security/protoc-gen-go/internal/enumgen"
	"github.com/effective-security/xlog"
	"google.golang.org/protobuf/compiler/protogen"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "go-enums")

var (
	log        = flag.Bool("logs", false, "output logs")
	out        = flag.String("out", "enums", "output file prefix")
	outMsgs    = flag.String("out-msgs", "messages", "output messages")
	importpath = flag.String("import", "", "go import path")
	pkg        = flag.String("package", "", "go package name")
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

		opts := enumgen.Opts{
			Package: *pkg,
		}

		var allEnums []*protogen.Enum
		var msgs []*protogen.Message

		for _, name := range gp.Request.FileToGenerate {
			f := gp.FilesByPath[name]
			if len(f.Enums) == 0 && len(f.Messages) == 0 {
				logger.Infof("Skipping %s, no enums", name)
				continue
			}

			if opts.Package == "" {
				opts.Package = string(f.GoPackageName)
			}

			allEnums = append(allEnums, f.Enums...)
			allEnums = append(allEnums, enumgen.GetEnums(f.Messages)...)

			msgs = append(msgs, enumgen.GetMessagesToDescribe(opts, f.Messages, f.Services)...)
		}

		if len(msgs) > 0 {
			fn := fmt.Sprintf("%s.pb.go", *outMsgs)
			logger.Infof("Generating %s\n", fn)

			f := gp.NewGeneratedFile(fn, protogen.GoImportPath(*importpath))
			err := enumgen.ApplyMessagesTemplate(f, opts, msgs)
			if err != nil {
				gp.Error(err)
			}
		}

		if len(allEnums) > 0 || len(msgs) > 0 {
			fn := fmt.Sprintf("%s.pb.go", *out)
			logger.Infof("Generating %s\n", fn)

			f := gp.NewGeneratedFile(fn, protogen.GoImportPath(*importpath))
			err := enumgen.ApplyEnumsTemplate(f, opts, allEnums)
			if err != nil {
				gp.Error(err)
			}
		}

		return nil
	})
}
