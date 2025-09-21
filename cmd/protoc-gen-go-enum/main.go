package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
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

		dopts := enumgen.Opts{
			Package: *pkg,
		}

		if dopts.Package == "" {
			return errors.Errorf("package is required")
		}

		allEnums := enumgen.GetEnumsDescriptions(gp, dopts)
		msgs := enumgen.GetMessagesDescriptions(gp, dopts)

		if len(msgs) > 0 {
			fn := fmt.Sprintf("%s.pb.go", *outMsgs)
			logger.Infof("Generating %s\n", fn)

			f := gp.NewGeneratedFile(fn, protogen.GoImportPath(*importpath))
			err := enumgen.ApplyMessagesTemplate(f, dopts, msgs)
			if err != nil {
				gp.Error(err)
			}
		}

		if len(allEnums) > 0 || len(msgs) > 0 {
			fn := fmt.Sprintf("%s.pb.go", *out)
			logger.Infof("Generating %s\n", fn)

			f := gp.NewGeneratedFile(fn, protogen.GoImportPath(*importpath))
			err := enumgen.ApplyEnumsTemplate(f, dopts, allEnums)
			if err != nil {
				gp.Error(err)
			}
		}

		return nil
	})
}
