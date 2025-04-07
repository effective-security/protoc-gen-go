package enumgen

import (
	"testing"

	"github.com/effective-security/protoc-gen-go/e2e"
	"github.com/stretchr/testify/assert"
)

func Test_FormatDisplayName(t *testing.T) {
	assert.Equal(t, "Test", FormatDisplayName("Test"))
	assert.Equal(t, "test", FormatDisplayName("test"))
	assert.Equal(t, "Test Data", FormatDisplayName("TestData"))
	assert.Equal(t, "AWS Name", FormatDisplayName("AWSName"))
	assert.Equal(t, "S3 Location", FormatDisplayName("S3Location"))
	assert.Equal(t, "EC2 Instance", FormatDisplayName("EC2Instance"))
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
