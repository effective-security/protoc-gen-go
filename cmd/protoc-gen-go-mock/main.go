package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/effective-security/protoc-gen-go/internal/mockgen"
	"github.com/effective-security/xlog"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "go-mock")

var (
	log     = flag.Bool("logs", true, "output logs")
	pkgName = flag.String("pkg", "mockpb", "go package name")
)

func main() {
	flag.Parse()
	defer logger.Flush()

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gp *protogen.Plugin) error {
		gp.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		var formatter xlog.Formatter
		if *log {
			formatter = xlog.NewStringFormatter(os.Stderr).
				Options(xlog.FormatWithCaller(false), xlog.FormatSkipTime(true), xlog.FormatSkipLevel(true))
			xlog.SetGlobalLogLevel(xlog.INFO)
		} else {
			formatter = xlog.NewNilFormatter()
		}
		xlog.SetFormatter(formatter)

		pkg := *pkgName
		if pkg == "" {
			return errors.Errorf("mocks should be generated in a separate package: use -pkg flag")
		}

		opts := mockgen.Options{
			Package: pkg,
		}

		for _, name := range gp.Request.FileToGenerate {
			f := gp.FilesByPath[name]

			if len(f.Services) == 0 {
				logger.Infof("Skipping %s, no services", name)
				continue
			}

			logger.Infof("Processing: %s", name)

			prefix := filepath.Base(f.GeneratedFilenamePrefix)
			fn := fmt.Sprintf("%s.mock.pb.go", prefix)
			fullFn := filepath.Join(pkg, fn)
			logger.Infof("Generating %s\n", fullFn)

			gf := gp.NewGeneratedFile(fullFn, f.GoImportPath)

			err := mockgen.ApplyTemplate(gf, f, opts)
			if err != nil {
				gf.Skip()
				gp.Error(err)
				continue
			}
		}

		return nil
	})
}
