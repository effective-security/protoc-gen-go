package main

import (
	"flag"
	"os"
	"sort"
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

		enums := enumgen.GetEnumsDescriptions(gp, enumgen.Opts{})

		grouped := map[string][]*enumgen.EnumDescription{}
		for _, en := range enums {
			fn := strings.TrimSuffix(en.FileName, ".proto")
			grouped[fn] = append(grouped[fn], en)
		}

		var fileEnumInfos []enumgen.FileEnumInfo
		for fn, enums := range grouped {

			sort.Slice(enums, func(i, j int) bool {
				return strings.ToLower(enums[i].FullName) < strings.ToLower(enums[j].FullName)
			})

			fileEnumInfos = append(fileEnumInfos, enumgen.FileEnumInfo{
				FileName: fn,
				Enums:    enums,
			})
		}

		sort.Slice(fileEnumInfos, func(i, j int) bool {
			return fileEnumInfos[i].FileName < fileEnumInfos[j].FileName
		})

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
