package enumgen

import (
	"testing"

	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/effective-security/x/format"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "EC2 Instance,S3 Bucket,Lambda Function", rt.DisplayValue())
	assert.Equal(t, []string{"EC2 Instance", "S3 Bucket", "Lambda Function"}, rt.DisplayValues())

	rt = 0
	assert.Equal(t, "Unknown", rt.DisplayValue())
	assert.Equal(t, []string{"Unknown"}, rt.DisplayValues())

	rt = e2e.ResourceType_S3Bucket
	assert.Equal(t, "S3 Bucket", rt.DisplayValue())
	assert.Equal(t, []string{"S3 Bucket"}, rt.DisplayValues())

	rt = e2e.ResourceType_All
	assert.Equal(t, "EC2 Instance,S3 Bucket,Lambda Function", rt.DisplayValue())
	assert.Equal(t, []string{"EC2 Instance", "S3 Bucket", "Lambda Function"}, rt.DisplayValues())
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
		{"// A\n\nB\n\n\n", "A\n\nB"},
		{"// \n\n\n\n\n", ""},
	}

	for _, test := range tests {
		result := cleanComment(test.input)
		assert.Equal(t, test.expected, result)
	}
}
