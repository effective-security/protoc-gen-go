package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/effective-security/protoc-gen-go/internal/enumgen"
	"github.com/effective-security/xlog"
	"google.golang.org/protobuf/compiler/protogen"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "go-enums")

var (
	log = flag.Bool("logs", false, "output logs")
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

		for _, name := range gp.Request.FileToGenerate {
			f := gp.FilesByPath[name]
			prefix := path.Base(f.GeneratedFilenamePrefix)

			if len(f.Enums) == 0 && len(f.Messages) == 0 {
				logger.Infof("Skipping %s, no enums", name)
				continue
			}

			enumsCount := 0
			for _, m := range f.Messages {
				enumsCount += len(m.Enums)
			}
			if enumsCount == 0 {
				logger.Infof("Skipping %s, no enums", name)
				continue
			}

			fn := fmt.Sprintf("%s.enum.pb.go", prefix)
			logger.Infof("Generating %s\n", fn)

			gf := gp.NewGeneratedFile(fn, f.GoImportPath)
			err := enumgen.ApplyTemplate(gf, f)
			if err != nil {
				gf.Skip()
				gp.Error(err)
				continue
			}
		}

		return nil
	})
}
