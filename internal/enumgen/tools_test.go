package enumgen

import (
	"bytes"
	"os"
	"testing"

	"github.com/effective-security/protoc-gen-go/api"
	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/effective-security/x/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	pluginpb "google.golang.org/protobuf/types/pluginpb"
)

func Test_FormatDisplayName(t *testing.T) {
	assert.Equal(t, "Test", format.DisplayName("Test"))
	assert.Equal(t, "test", format.DisplayName("test"))
	assert.Equal(t, "Test Value", format.DisplayName("test_value"))
	assert.Equal(t, "Test Data", format.DisplayName("TestData"))
	assert.Equal(t, "AWS Name", format.DisplayName("AWSName"))
	assert.Equal(t, "S3 Location", format.DisplayName("S3Location"))
	assert.Equal(t, "EC2 Instance", format.DisplayName("EC2Instance"))
	assert.Equal(t, "Asset IDs", format.DisplayName("AssetIDs"))
}

func Test_EnumDisplayValue(t *testing.T) {
	rt := e2e.ResourceType_EC2Instance | e2e.ResourceType_S3Bucket | e2e.ResourceType_LambdaFunction
	assert.Equal(t, "EC2 Instance,S3 Bucket,Lambda Function", rt.DisplayName())
	assert.Equal(t, []string{"EC2 Instance", "S3 Bucket", "Lambda Function"}, rt.DisplayNames())

	rt = 0
	assert.Equal(t, "Unknown", rt.DisplayName())
	assert.Equal(t, []string{"Unknown"}, rt.DisplayNames())

	rt = e2e.ResourceType_S3Bucket
	assert.Equal(t, "S3 Bucket", rt.DisplayName())
	assert.Equal(t, []string{"S3 Bucket"}, rt.DisplayNames())

	rt = e2e.ResourceType_All
	assert.Equal(t, "EC2 Instance,S3 Bucket,Lambda Function", rt.DisplayName())
	assert.Equal(t, []string{"EC2 Instance", "S3 Bucket", "Lambda Function"}, rt.DisplayNames())
}

func Test_cleanComment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"// This is a comment", "This is a comment"},
		{"// This is a comment\n", "This is a comment"},
		{"// This is a comment\n// Another line", "This is a comment\nAnother line"},
		{"// This is a comment\n\n", "This is a comment"},
		{"// This is a comment\n//TODO: this line should NOT be included in the documentation.\n", "This is a comment"},
		{"// A\n\nB\n\n\n", "A\nB"},
		{"// \n\n\n\n\n", ""},
	}

	for _, test := range tests {
		result := cleanComment(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func Test_GetMessageDescriptions(t *testing.T) {
	descriptions := e2e.GetMessageDescriptions()
	require.NotNil(t, descriptions)
	for fn, md := range descriptions {
		for _, field := range md.Fields {
			if field.Type == "object" || field.Type == "[]object" {
				assert.NotEmpty(t, field.StructName, "Field %s in %s", field.Name, fn)
				//assert.NotEmpty(t, field.Fields, "Field %s in %s", field.Name, fn)
			}
		}
	}
}

func Test_GetEnumsToDescribe(t *testing.T) {
	p := loadPluginFromRequestBin(t, "testdata/code_generator_request.pb.bin")
	opts := Opts{Package: "e2e"}

	allEnums := GetEnumsDescriptions(p, opts)
	assert.Equal(t, 4, len(allEnums))
}

func Test_CreateMessageDescription(t *testing.T) {
	p := loadPluginFromRequestBin(t, "testdata/code_generator_request.pb.bin")

	ops := Opts{Package: "e2e"}

	descriptions := GetMessagesDescriptions(p, ops)
	assert.Equal(t, 27, len(descriptions))
}

func loadPluginFromRequestBin(t *testing.T, path string) *protogen.Plugin {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	req := &pluginpb.CodeGeneratorRequest{}
	require.NoError(t, proto.Unmarshal(data, req))
	opts := protogen.Options{}
	p, err := opts.New(req)
	require.NoError(t, err)
	return p
}

func Test_GoModel(t *testing.T) {
	hestedFields := []*FieldMeta{
		{
			Name:   "NestedField",
			GoName: "NestedField",
			Type:   "string",
		},
		{
			Name:   "NestedField2",
			GoName: "NestedField2",
			Type:   "map",
			Fields: []*FieldMeta{
				{
					Name:   "Key",
					GoName: "Key",
					Type:   "string",
				},
				{
					Name:   "Value",
					GoName: "Value",
					Type:   "string",
				},
			},
		},
	}
	msg := &MessageDescription{
		FullName: "e2e.Test",
		Name:     "Test",
		Package:  "e2e",
		Fields: []*FieldMeta{
			{
				Name:   "str",
				GoName: "Str",
				Type:   "string",
			},
			{
				Name:          "HiddenCounter",
				GoName:        "HiddenCounter",
				Type:          "int32",
				SearchOptions: api.SearchOption_NoIndex,
			},
			{
				Name:          "Counter",
				GoName:        "Counter",
				Type:          "int32",
				SearchOptions: api.SearchOption_None,
			},
			{
				Name:       "Nested",
				GoName:     "Nested",
				Type:       "struct",
				StructName: "Nested",
				Package:    "e2e",
				Fields:     hestedFields,
			},
			{
				Name:       "NestedList",
				GoName:     "NestedList",
				Type:       "[]struct",
				StructName: "Nested",
				Package:    "e2e",
				Fields:     hestedFields,
			},
		},
	}

	opts := Opts{Package: "e2e", ModelPackage: "modelpb"}

	var buf bytes.Buffer
	err := GenerateGoModels(&buf, opts, []*MessageDescription{msg})
	require.NoError(t, err)
	js := buf.String()

	exp, err := os.ReadFile("testdata/gen_model.go.txt")
	require.NoError(t, err)
	assert.Equal(t, string(exp), js)
}
