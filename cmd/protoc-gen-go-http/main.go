package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/cockroachdb/errors"
	"github.com/effective-security/protoc-gen-go/internal/httpgen"
	"github.com/effective-security/xlog"
	"google.golang.org/protobuf/compiler/protogen"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "go-http")

var (
	log       = flag.Bool("logs", true, "output logs")
	pkgName   = flag.String("pkg", "httppb", "go package name")
	pbPkgName = flag.String("pbpkg", "pb", "go package name for main protobuf types")
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

		pkg := *pkgName
		if pkg == "" {
			return errors.Errorf("HTTP handler should be generated in a separage package. Use -pkg flag.")
		}
		// pbpkg := *pbPkgName
		// if pbpkg == "" {
		// 	pbpkg = "pb"
		// }

		opts := httpgen.Options{
			Package:   pkg,
			PbPackage: *pbPkgName,
		}

		for _, name := range gp.Request.FileToGenerate {
			f := gp.FilesByPath[name]

			if len(f.Services) == 0 {
				logger.Infof("Skipping %s, no services", name)
				continue
			}

			logger.Infof("Processing: %s", name)

			prefix := path.Base(f.GeneratedFilenamePrefix)
			fn := fmt.Sprintf("%s.pb.go", prefix)
			fullFn := path.Join(pkg, fn)
			logger.Infof("Generating %s\n", fullFn)

			gf := gp.NewGeneratedFile(fullFn, f.GoImportPath)

			err := httpgen.ApplyTemplate(gf, f, opts)
			if err != nil {
				gf.Skip()
				gp.Error(err)
				continue
			}
		}

		return nil
	})
}
