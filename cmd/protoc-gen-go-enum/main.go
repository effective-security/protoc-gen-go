package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/effective-security/protoc-gen-go/internal/enumgen"
	"github.com/effective-security/xlog"
	"google.golang.org/protobuf/compiler/protogen"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "go-enums")

var (
	log          = flag.Bool("logs", false, "output logs")
	out          = flag.String("out", "enums", "output file prefix")
	outMsgs      = flag.String("out-msgs", "messages", "output messages")
	outModels    = flag.String("out-models", "models", "output models")
	importpath   = flag.String("import", "", "go import path")
	pkgName      = flag.String("package", "", "go package name")
	modelPkgName = flag.String("model-pkg", "modelpb", "go package name for model types")
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
			Package:      *pkgName,
			ModelPackage: *modelPkgName,
		}

		if dopts.Package == "" {
			return errors.Errorf("package is required")
		}
		if dopts.ModelPackage == "" {
			return errors.Errorf("model package is required")
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

			fn2 := fmt.Sprintf("%s.pb.go", *outModels)
			fullFn := filepath.Join(dopts.ModelPackage, fn2)
			logger.Infof("Generating %s\n", fullFn)

			f2 := gp.NewGeneratedFile(fullFn, protogen.GoImportPath(*importpath))
			err = enumgen.ApplyModelsTemplate(f2, dopts, msgs)
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
