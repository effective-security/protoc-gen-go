package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/effective-security/protoc-gen-go/jsongen"
	"github.com/effective-security/xlog"
	"google.golang.org/protobuf/compiler/protogen"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go", "go-json")

var (
	log          = flag.Bool("logs", false, "output logs")
	enumsAsInts  = flag.Bool("enums_as_ints", false, "render enums as integers as opposed to strings")
	emitDefaults = flag.Bool("emit_defaults", false, "render fields with zero values")
	origName     = flag.Bool("orig_name", false, "use original (.proto) name for fields")
	multiline    = flag.Bool("multiline", false, "encode JSON with indent")
	partial      = flag.Bool("partial", false, "allow partial encoding")
	allowUnknown = flag.Bool("allow_unknown", false, "allow messages to contain unknown fields when unmarshaling")
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

		opts := jsongen.Options{
			EnumsAsInts:        *enumsAsInts,
			EmitDefaults:       *emitDefaults,
			OrigName:           *origName,
			AllowUnknownFields: *allowUnknown,
			Partial:            *partial,
			Multiline:          *multiline,
		}

		for _, name := range gp.Request.FileToGenerate {
			f := gp.FilesByPath[name]
			prefix := path.Base(f.GeneratedFilenamePrefix)

			if len(f.Messages) == 0 {
				logger.Infof("Skipping %s, no messages", name)
				continue
			}

			fn := fmt.Sprintf("%s.pb.json.go", prefix)
			logger.Infof("Generating %s\n", fn)

			gf := gp.NewGeneratedFile(fn, f.GoImportPath)

			err := jsongen.ApplyTemplate(gf, f, opts)
			if err != nil {
				gf.Skip()
				gp.Error(err)
				continue
			}
		}

		return nil
	})
}
